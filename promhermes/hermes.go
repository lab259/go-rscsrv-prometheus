// Package promhermes provides tooling around HTTP servers.
//
// First, the package allows the creation of hermes.Handler instances to expose
// Prometheus metrics via HTTP. promhermes.Handler acts on the
// prometheus.DefaultGatherer. With HandlerFor, you can create a handler for a
// custom registry or anything that implements the Gatherer interface. It also
// allows the creation of handlers that act differently on errors or allow to
// log errors.
//
// Second, the package provides tooling to instrument instances of hermes.Handler
// via middleware. Middleware wrappers follow the naming scheme
// InstrumentHandlerX, where X describes the intended use of the middleware.
// See each function's doc comment for specific details.
package promhermes

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/lab259/errors/v2"
	"github.com/lab259/hermes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/valyala/fasthttp"
)

const (
	contentTypeHeader     = "Content-Type"
	contentEncodingHeader = "Content-Encoding"
	acceptEncodingHeader  = "Accept-Encoding"
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

var errMaxConcurrentRequests = errors.New("limit of concurrent requests reached")
var errGzipFailedToClose = errors.New("unable to close gzip.Writer")
var errTimeout = errors.New("configured timeout exceeded")

// DefaultHandler returns an hermes.Handler for the prometheus.DefaultGatherer, using
// default HandlerOpts, i.e. it reports the first error as an HTTP error, it has
// no error logging, and it applies compression if requested by the client.
//
// The returned hermes.Handler is already instrumented using the
// InstrumentMetricHandler function and the prometheus.DefaultRegisterer. If you
// create multiple hermes.Handlers by separate calls of the Handler function, the
// metrics used for instrumentation will be shared between them, providing
// global scrape counts.
//
// This function is meant to cover the bulk of basic use cases. If you are doing
// anything that requires more customization (including using a non-default
// Gatherer, different instrumentation, and non-default HandlerOpts), use the
// HandlerFor function. See there for details.
func DefaultHandler() hermes.Handler {
	return InstrumentMetricHandler(
		prometheus.DefaultRegisterer, HandlerFor(prometheus.DefaultGatherer, HandlerOpts{}),
	)
}

// Handler returns an hermes.Handler for the promsrv.Service, using
// default HandlerOpts, i.e. it reports the first error as an HTTP error, it has
// no error logging, and it applies compression if requested by the client.
//
// The returned hermes.Handler is already instrumented using the
// InstrumentMetricHandler function and the prometheus.DefaultRegisterer. If you
// create multiple hermes.Handlers by separate calls of the Handler function, the
// metrics used for instrumentation will be shared between them, providing
// global scrape counts.
//
// This function is meant to cover the bulk of basic use cases. If you are doing
// anything that requires more customization (including using a non-default
// Gatherer, different instrumentation, and non-default HandlerOpts), use the
// HandlerFor function. See there for details.
func Handler(srv interface {
	prometheus.Gatherer
	prometheus.Registerer
}) hermes.Handler {
	return InstrumentMetricHandler(
		srv, HandlerFor(srv, HandlerOpts{}),
	)
}

// HandlerFor returns an uninstrumented hermes.Handler for the provided
// Gatherer. The behavior of the Handler is defined by the provided
// HandlerOpts. Thus, HandlerFor is useful to create hermes.Handlers for custom
// Gatherers, with non-default HandlerOpts, and/or with custom (or no)
// instrumentation. Use the InstrumentMetricHandler function to apply the same
// kind of instrumentation as it is used by the Handler function.
func HandlerFor(reg prometheus.Gatherer, opts HandlerOpts) hermes.Handler {
	var (
		inFlightSem chan struct{}
		errCnt      = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "promhermes_metric_handler_errors_total",
				Help: "Total number of internal errors encountered by the promhermes metric handler.",
			},
			[]string{"cause"},
		)
	)

	if opts.MaxRequestsInFlight > 0 {
		inFlightSem = make(chan struct{}, opts.MaxRequestsInFlight)
	}
	if opts.Registry != nil {
		// Initialize all possibilites that can occur below.
		errCnt.WithLabelValues("gathering")
		errCnt.WithLabelValues("encoding")
		if err := opts.Registry.Register(errCnt); err != nil {
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
				errCnt = are.ExistingCollector.(*prometheus.CounterVec)
			} else {
				panic(err)
			}
		}
	}

	h := hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		if inFlightSem != nil {
			select {
			case inFlightSem <- struct{}{}: // All good, carry on.
				defer func() { <-inFlightSem }()
			default:
				return res.Error(errMaxConcurrentRequests, fmt.Sprintf(
					"Limit of concurrent requests reached (%d), try again later.", opts.MaxRequestsInFlight,
				), hermes.StatusServiceUnavailable, errors.Code("max-concurrent-request"), errors.Module("promhermes"))
			}
		}
		mfs, err := reg.Gather()
		if err != nil {
			if opts.ErrorLog != nil {
				opts.ErrorLog.Println("error gathering metrics:", err)
			}
			errCnt.WithLabelValues("gathering").Inc()
			switch opts.ErrorHandling {
			case PanicOnError:
				panic(err)
			case ContinueOnError:
				if len(mfs) == 0 {
					// Still report the error if no metrics have been gathered.
					return httpError(req, res, err)
				}
			case HTTPErrorOnError:
				return httpError(req, res, err)
			}
		}

		headers := parseHeaders(req)
		contentType := expfmt.Negotiate(headers)
		res.Header(contentTypeHeader, string(contentType))

		buf := bytes.NewBuffer(nil)
		var w io.Writer

		if !opts.DisableCompression && gzipAccepted(req) {
			res.Header(contentEncodingHeader, "gzip")
			gz := gzipPool.Get().(*gzip.Writer)
			defer gzipPool.Put(gz)

			gz.Reset(buf)
			w = gz
		} else {
			w = buf
		}

		enc := expfmt.NewEncoder(w, contentType)

		var lastErr error
		for _, mf := range mfs {
			if err := enc.Encode(mf); err != nil {
				lastErr = err
				if opts.ErrorLog != nil {
					opts.ErrorLog.Println("error encoding and sending metric family:", err)
				}
				errCnt.WithLabelValues("encoding").Inc()
				switch opts.ErrorHandling {
				case PanicOnError:
					panic(err)
				case ContinueOnError:
					// Handled later.
				case HTTPErrorOnError:
					return httpError(req, res, err)
				}
			}
		}

		if lastErr != nil {
			return httpError(req, res, lastErr)
		}

		if closer, ok := w.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return httpError(req, res, errGzipFailedToClose)
			}
		}

		return res.Data(buf.Bytes())
	})

	if opts.Timeout <= 0 {
		return h
	}

	return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		var r hermes.Result

		ctx, cancelCtx := context.WithTimeout(req.Context(), opts.Timeout)
		defer cancelCtx()

		req = req.WithContext(ctx)

		done := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		newRawReqCtx := &fasthttp.RequestCtx{}
		req.Raw().Request.CopyTo(&newRawReqCtx.Request)

		newReq := hermes.AcquireRequest(ctx, newRawReqCtx)
		newRes := hermes.AcquireResponse(newReq.Raw())

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			defer func() {
				hermes.ReleaseRequest(newReq)
				hermes.ReleaseResponse(newRes)
			}()

			r = h(newReq, newRes)

			close(done)
		}()

		select {
		case p := <-panicChan:
			panic(p)
		case <-done:
			return r
		case <-ctx.Done():
			return res.Error(errTimeout, fmt.Sprintf(
				"Exceeded configured timeout of %v.",
				opts.Timeout,
			), hermes.StatusRequestTimeout, errors.Code("timeout"), errors.Module("promhermes"))
		}
	})
}

// InstrumentMetricHandler is usually used with an hermes.Handler returned by the
// HandlerFor function. It instruments the provided hermes.Handler with two
// metrics: A counter vector "promhermes_metric_handler_requests_total" to count
// scrapes partitioned by HTTP status code, and a gauge
// "promhermes_metric_handler_requests_in_flight" to track the number of
// simultaneous scrapes. This function idempotently registers collectors for
// both metrics with the provided Registerer. It panics if the registration
// fails. The provided metrics are useful to see how many scrapes hit the
// monitored target (which could be from different Prometheus servers or other
// scrapers), and how often they overlap (which would result in more than one
// scrape in flight at the same time). Note that the scrapes-in-flight gauge
// will contain the scrape by which it is exposed, while the scrape counter will
// only get incremented after the scrape is complete (as only then the status
// code is known). For tracking scrape durations, use the
// "scrape_duration_seconds" gauge created by the Prometheus server upon each
// scrape.
func InstrumentMetricHandler(reg prometheus.Registerer, handler hermes.Handler) hermes.Handler {
	cnt := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "promhermes_metric_handler_requests_total",
			Help: "Total number of scrapes by HTTP status code.",
		},
		[]string{"code"},
	)
	// Initialize the most likely HTTP status codes.
	cnt.WithLabelValues("200")
	cnt.WithLabelValues("500")
	cnt.WithLabelValues("503")
	if err := reg.Register(cnt); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			cnt = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}

	gge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "promhermes_metric_handler_requests_in_flight",
		Help: "Current number of scrapes being served.",
	})
	if err := reg.Register(gge); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			gge = are.ExistingCollector.(prometheus.Gauge)
		} else {
			panic(err)
		}
	}

	return InstrumentHandlerCounter(cnt, InstrumentHandlerInFlight(gge, handler))
}

// HandlerErrorHandling defines how a Handler serving metrics will handle
// errors.
type HandlerErrorHandling int

// These constants cause handlers serving metrics to behave as described if
// errors are encountered.
const (
	// Serve an HTTP status code 500 upon the first error
	// encountered. Report the error message in the body.
	HTTPErrorOnError HandlerErrorHandling = iota
	// Ignore errors and try to serve as many metrics as possible.  However,
	// if no metrics can be served, serve an HTTP status code 500 and the
	// last error message in the body. Only use this in deliberate "best
	// effort" metrics collection scenarios. In this case, it is highly
	// recommended to provide other means of detecting errors: By setting an
	// ErrorLog in HandlerOpts, the errors are logged. By providing a
	// Registry in HandlerOpts, the exposed metrics include an error counter
	// "promhermes_metric_handler_errors_total", which can be used for
	// alerts.
	ContinueOnError
	// Panic upon the first error encountered (useful for "crash only" apps).
	PanicOnError
)

// Logger is the minimal interface HandlerOpts needs for logging. Note that
// log.Logger from the standard library implements this interface, and it is
// easy to implement by custom loggers, if they don't do so already anyway.
type Logger interface {
	Println(v ...interface{})
}

// HandlerOpts specifies options how to serve metrics via an hermes.Handler. The
// zero value of HandlerOpts is a reasonable default.
type HandlerOpts struct {
	// ErrorLog specifies an optional logger for errors collecting and
	// serving metrics. If nil, errors are not logged at all.
	ErrorLog Logger
	// ErrorHandling defines how errors are handled. Note that errors are
	// logged regardless of the configured ErrorHandling provided ErrorLog
	// is not nil.
	ErrorHandling HandlerErrorHandling
	// If Registry is not nil, it is used to register a metric
	// "promhermes_metric_handler_errors_total", partitioned by "cause". A
	// failed registration causes a panic. Note that this error counter is
	// different from the instrumentation you get from the various
	// InstrumentHandler... helpers. It counts errors that don't necessarily
	// result in a non-2xx HTTP status code. There are two typical cases:
	// (1) Encoding errors that only happen after streaming of the HTTP body
	// has already started (and the status code 200 has been sent). This
	// should only happen with custom collectors. (2) Collection errors with
	// no effect on the HTTP status code because ErrorHandling is set to
	// ContinueOnError.
	Registry prometheus.Registerer
	// If DisableCompression is true, the handler will never compress the
	// response, even if requested by the client.
	DisableCompression bool
	// The number of concurrent HTTP requests is limited to
	// MaxRequestsInFlight. Additional requests are responded to with 503
	// Service Unavailable and a suitable message in the body. If
	// MaxRequestsInFlight is 0 or negative, no limit is applied.
	MaxRequestsInFlight int
	// If handling a request takes longer than Timeout, it is responded to
	// with 503 ServiceUnavailable and a suitable Message. No timeout is
	// applied if Timeout is 0 or negative. Note that with the current
	// implementation, reaching the timeout simply ends the HTTP requests as
	// described above (and even that only if sending of the body hasn't
	// started yet), while the bulk work of gathering all the metrics keeps
	// running in the background (with the eventual result to be thrown
	// away). Until the implementation is improved, it is recommended to
	// implement a separate timeout in potentially slow Collectors.
	Timeout time.Duration
}

// gzipAccepted returns whether the client will accept gzip-encoded content.
func gzipAccepted(req hermes.Request) bool {
	a := req.Header(acceptEncodingHeader)
	parts := strings.Split(string(a), ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "gzip" || strings.HasPrefix(part, "gzip;") {
			return true
		}
	}
	return false
}

// httpError removes any content-encoding header and then calls hermes.Error with
// the provided error and hermes.StatusInternalServerErrer. Error contents is
// supposed to be uncompressed plain text. However, same as with a plain
// hermes.Error, any header settings will be void if the header has already been
// sent. The error message will still be written to the writer, but it will
// probably be of limited use.
func httpError(req hermes.Request, res hermes.Response, err error) hermes.Result {
	req.Raw().Response.Header.Del(contentEncodingHeader)

	return res.Error(
		err,
		"An error has occurred while serving metrics: "+err.Error(),
		hermes.StatusInternalServerError,
		errors.Code("prometheus-failed"),
		errors.Module("promhermes"),
	)
}
