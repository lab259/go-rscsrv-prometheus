package promhermes

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/lab259/hermes"
	dto "github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"
)

// magicString is used for the hacky label test in checkLabels. Remove once fixed.
const magicString = "zZgWfBxLqvG8kc8IMv3POi2Bb0tZI3vAnBx+gBaFi9FyPzB/CzKUer1yufDa"

// InstrumentHandlerInFlight is a middleware that wraps the provided
// hermes.Handler. It sets the provided prometheus.Gauge to the number of
// requests currently handled by the wrapped hermes.Handler.
//
// See the example for InstrumentHandlerDuration for example usage.
func InstrumentHandlerInFlight(g prometheus.Gauge, next hermes.Handler) hermes.Handler {
	return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		g.Inc()
		defer g.Dec()
		return next(req, res)
	})
}

// InstrumentHandlerDuration is a middleware that wraps the provided
// hermes.Handler to observe the request duration with the provided ObserverVec.
// The ObserverVec must have zero, one, or two non-const non-curried labels. For
// those, the only allowed label names are "code" and "method". The function
// panics otherwise. The Observe method of the Observer in the ObserverVec is
// called with the request duration in seconds. Partitioning happens by HTTP
// status code and/or HTTP method if the respective instance label names are
// present in the ObserverVec. For unpartitioned observations, use an
// ObserverVec with zero labels. Note that partitioning of Histograms is
// expensive and should be used judiciously.
//
// If the wrapped Handler does not set a status code, a status code of 200 is assumed.
//
// If the wrapped Handler panics, no values are reported.
//
// Note that this method is only guaranteed to never observe negative durations
// if used with Go1.9+.
func InstrumentHandlerDuration(obs prometheus.ObserverVec, next hermes.Handler) hermes.Handler {
	code, method := checkLabels(obs)

	if code {
		return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
			now := time.Now()
			r := next(req, res)
			obs.With(labels(code, method, string(req.Method()), req.Raw().Response.StatusCode())).Observe(time.Since(now).Seconds())
			return r
		})
	}

	return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		now := time.Now()
		r := next(req, res)
		obs.With(labels(code, method, string(req.Method()), 0)).Observe(time.Since(now).Seconds())
		return r
	})
}

// InstrumentHandlerCounter is a middleware that wraps the provided hermes.Handler
// to observe the request result with the provided CounterVec.  The CounterVec
// must have zero, one, or two non-const non-curried labels. For those, the only
// allowed label names are "code" and "method". The function panics
// otherwise. Partitioning of the CounterVec happens by HTTP status code and/or
// HTTP method if the respective instance label names are present in the
// CounterVec. For unpartitioned counting, use a CounterVec with zero labels.
//
// If the wrapped Handler does not set a status code, a status code of 200 is assumed.
//
// If the wrapped Handler panics, the Counter is not incremented.
//
// See the example for InstrumentHandlerDuration for example usage.
func InstrumentHandlerCounter(counter *prometheus.CounterVec, next hermes.Handler) hermes.Handler {
	code, method := checkLabels(counter)

	if code {
		return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
			r := next(req, res)
			counter.With(labels(code, method, string(req.Method()), req.Raw().Response.StatusCode())).Inc()
			return r
		})
	}

	return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		r := next(req, res)
		counter.With(labels(code, method, string(req.Method()), 0)).Inc()
		return r
	})
}

// InstrumentHandlerTimeToWriteHeader is a middleware that wraps the provided
// hermes.Handler to observe with the provided ObserverVec the request duration
// until the response headers are written. The ObserverVec must have zero, one,
// or two non-const non-curried labels. For those, the only allowed label names
// are "code" and "method". The function panics otherwise. The Observe method of
// the Observer in the ObserverVec is called with the request duration in
// seconds. Partitioning happens by HTTP status code and/or HTTP method if the
// respective instance label names are present in the ObserverVec. For
// unpartitioned observations, use an ObserverVec with zero labels. Note that
// partitioning of Histograms is expensive and should be used judiciously.
//
// If the wrapped Handler panics before calling WriteHeader, no value is
// reported.
//
// Note that this method is only guaranteed to never observe negative durations
// if used with Go1.9+.
//
// See the example for InstrumentHandlerDuration for example usage.
// TODO: try to implement this one using hermes/fasthttp
// func InstrumentHandlerTimeToWriteHeader(obs prometheus.ObserverVec, next hermes.Handler) hermes.Handler {
// 	code, method := checkLabels(obs)

// 	return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
// 		now := time.Now()
// 		d := newDelegator(w, func(status int) {
// 			obs.With(labels(code, method, r.Method, status)).Observe(time.Since(now).Seconds())
// 		})
// 		next.ServeHTTP(d, r)
// 	})
// }

// InstrumentHandlerRequestSize is a middleware that wraps the provided
// hermes.Handler to observe the request size with the provided ObserverVec.  The
// ObserverVec must have zero, one, or two non-const non-curried labels. For
// those, the only allowed label names are "code" and "method". The function
// panics otherwise. The Observe method of the Observer in the ObserverVec is
// called with the request size in bytes. Partitioning happens by HTTP status
// code and/or HTTP method if the respective instance label names are present in
// the ObserverVec. For unpartitioned observations, use an ObserverVec with zero
// labels. Note that partitioning of Histograms is expensive and should be used
// judiciously.
//
// If the wrapped Handler does not set a status code, a status code of 200 is assumed.
//
// If the wrapped Handler panics, no values are reported.
//
// See the example for InstrumentHandlerDuration for example usage.
func InstrumentHandlerRequestSize(obs prometheus.ObserverVec, next hermes.Handler) hermes.Handler {
	code, method := checkLabels(obs)

	if code {
		return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
			r := next(req, res)
			size := computeApproximateRequestSize(req)
			obs.With(labels(code, method, string(req.Method()), req.Raw().Response.StatusCode())).Observe(float64(size))
			return r
		})
	}

	return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		r := next(req, res)
		size := computeApproximateRequestSize(req)
		obs.With(labels(code, method, string(req.Method()), 0)).Observe(float64(size))
		return r
	})
}

// InstrumentHandlerResponseSize is a middleware that wraps the provided
// hermes.Handler to observe the response size with the provided ObserverVec.  The
// ObserverVec must have zero, one, or two non-const non-curried labels. For
// those, the only allowed label names are "code" and "method". The function
// panics otherwise. The Observe method of the Observer in the ObserverVec is
// called with the response size in bytes. Partitioning happens by HTTP status
// code and/or HTTP method if the respective instance label names are present in
// the ObserverVec. For unpartitioned observations, use an ObserverVec with zero
// labels. Note that partitioning of Histograms is expensive and should be used
// judiciously.
//
// If the wrapped Handler does not set a status code, a status code of 200 is assumed.
//
// If the wrapped Handler panics, no values are reported.
//
// See the example for InstrumentHandlerDuration for example usage.
func InstrumentHandlerResponseSize(obs prometheus.ObserverVec, next hermes.Handler) hermes.Handler {
	code, method := checkLabels(obs)
	return hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		r := next(req, res)
		obs.With(labels(code, method, string(req.Method()), req.Raw().Response.StatusCode())).Observe(float64(len(req.Raw().Response.Body())))
		return r
	})
}

func parseHeaders(req hermes.Request) map[string][]string {
	header := make(map[string][]string)
	req.Raw().Request.Header.VisitAll(func(k, v []byte) {
		sk, sv := string(k), string(v)
		if headerValue, ok := header[sk]; ok {
			header[sk] = append(headerValue, sv)
		} else {
			header[sk] = []string{sv}
		}
	})
	return header

}

func checkLabels(c prometheus.Collector) (code bool, method bool) {
	// TODO(beorn7): Remove this hacky way to check for instance labels
	// once Descriptors can have their dimensionality queried.
	var (
		desc *prometheus.Desc
		m    prometheus.Metric
		pm   dto.Metric
		lvs  []string
	)

	// Get the Desc from the Collector.
	descc := make(chan *prometheus.Desc, 1)
	c.Describe(descc)

	select {
	case desc = <-descc:
	default:
		panic("no description provided by collector")
	}
	select {
	case <-descc:
		panic("more than one description provided by collector")
	default:
	}

	close(descc)

	// Create a ConstMetric with the Desc. Since we don't know how many
	// variable labels there are, try for as long as it needs.
	for err := errors.New("dummy"); err != nil; lvs = append(lvs, magicString) {
		m, err = prometheus.NewConstMetric(desc, prometheus.UntypedValue, 0, lvs...)
	}

	// Write out the metric into a proto message and look at the labels.
	// If the value is not the magicString, it is a constLabel, which doesn't interest us.
	// If the label is curried, it doesn't interest us.
	// In all other cases, only "code" or "method" is allowed.
	if err := m.Write(&pm); err != nil {
		panic("error checking metric for labels")
	}
	for _, label := range pm.Label {
		name, value := label.GetName(), label.GetValue()
		if value != magicString || isLabelCurried(c, name) {
			continue
		}
		switch name {
		case "code":
			code = true
		case "method":
			method = true
		default:
			panic("metric partitioned with non-supported labels")
		}
	}
	return
}

func isLabelCurried(c prometheus.Collector, label string) bool {
	// This is even hackier than the label test above.
	// We essentially try to curry again and see if it works.
	// But for that, we need to type-convert to the two
	// types we use here, ObserverVec or *CounterVec.
	switch v := c.(type) {
	case *prometheus.CounterVec:
		if _, err := v.CurryWith(prometheus.Labels{label: "dummy"}); err == nil {
			return false
		}
	case prometheus.ObserverVec:
		if _, err := v.CurryWith(prometheus.Labels{label: "dummy"}); err == nil {
			return false
		}
	default:
		panic("unsupported metric vec type")
	}
	return true
}

// emptyLabels is a one-time allocation for non-partitioned metrics to avoid
// unnecessary allocations on each request.
var emptyLabels = prometheus.Labels{}

func labels(code, method bool, reqMethod string, status int) prometheus.Labels {
	if !(code || method) {
		return emptyLabels
	}
	labels := prometheus.Labels{}

	if code {
		labels["code"] = sanitizeCode(status)
	}
	if method {
		labels["method"] = sanitizeMethod(reqMethod)
	}

	return labels
}

func computeApproximateRequestSize(req hermes.Request) int {
	ctx := req.Raw()

	s := 0
	if ctx.URI() != nil {
		s += len(ctx.URI().String())
	}

	s += len(ctx.Method())
	header := parseHeaders(req)
	for name, values := range header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}

	s += int(len(ctx.Request.Body()))
	return s
}

func sanitizeMethod(m string) string {
	switch m {
	case "GET", "get":
		return "get"
	case "PUT", "put":
		return "put"
	case "HEAD", "head":
		return "head"
	case "POST", "post":
		return "post"
	case "DELETE", "delete":
		return "delete"
	case "CONNECT", "connect":
		return "connect"
	case "OPTIONS", "options":
		return "options"
	case "NOTIFY", "notify":
		return "notify"
	default:
		return strings.ToLower(m)
	}
}

// If the wrapped hermes.Handler has not set a status code, i.e. the value is
// currently 0, santizeCode will return 200, for consistency with behavior in
// the stdlib.
func sanitizeCode(s int) string {
	switch s {
	case 100:
		return "100"
	case 101:
		return "101"

	case 200, 0:
		return "200"
	case 201:
		return "201"
	case 202:
		return "202"
	case 203:
		return "203"
	case 204:
		return "204"
	case 205:
		return "205"
	case 206:
		return "206"

	case 300:
		return "300"
	case 301:
		return "301"
	case 302:
		return "302"
	case 304:
		return "304"
	case 305:
		return "305"
	case 307:
		return "307"

	case 400:
		return "400"
	case 401:
		return "401"
	case 402:
		return "402"
	case 403:
		return "403"
	case 404:
		return "404"
	case 405:
		return "405"
	case 406:
		return "406"
	case 407:
		return "407"
	case 408:
		return "408"
	case 409:
		return "409"
	case 410:
		return "410"
	case 411:
		return "411"
	case 412:
		return "412"
	case 413:
		return "413"
	case 414:
		return "414"
	case 415:
		return "415"
	case 416:
		return "416"
	case 417:
		return "417"
	case 418:
		return "418"

	case 500:
		return "500"
	case 501:
		return "501"
	case 502:
		return "502"
	case 503:
		return "503"
	case 504:
		return "504"
	case 505:
		return "505"

	case 428:
		return "428"
	case 429:
		return "429"
	case 431:
		return "431"
	case 511:
		return "511"

	default:
		return strconv.Itoa(s)
	}
}
