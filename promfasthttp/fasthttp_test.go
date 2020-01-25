package promfasthttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

type errorCollector struct{}

func (e errorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("invalid_metric", "not helpful", nil, nil)
}

func (e errorCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.NewInvalidMetric(
		prometheus.NewDesc("invalid_metric", "not helpful", nil, nil),
		errors.New("collect error"),
	)
}

type blockingCollector struct {
	CollectStarted, Block chan struct{}
}

func (b blockingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy_desc", "not helpful", nil, nil)
}

func (b blockingCollector) Collect(ch chan<- prometheus.Metric) {
	select {
	case b.CollectStarted <- struct{}{}:
	default:
	}
	// Collects nothing, just waits for a channel receive.
	<-b.Block
}

var _ = Describe("Handler", func() {
	When("using custom error handling", func() {

		var (
			reg             *prometheus.Registry
			cnt             prometheus.Counter
			cntVec          *prometheus.CounterVec
			logBuf          *bytes.Buffer
			errorHandler    fasthttp.RequestHandler
			continueHandler fasthttp.RequestHandler
			panicHandler    fasthttp.RequestHandler
		)

		BeforeEach(func() {
			// Create a registry that collects a MetricFamily with two elements,
			// another with one, and reports an error. Further down, we'll use the
			// same registry in the HandlerOpts.
			reg = prometheus.NewRegistry()

			cnt = prometheus.NewCounter(prometheus.CounterOpts{
				Name: "the_count",
				Help: "Ah-ah-ah! Thunder and lightning!",
			})
			reg.MustRegister(cnt)

			cntVec = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name:        "name",
					Help:        "docstring",
					ConstLabels: prometheus.Labels{"constname": "constvalue"},
				},
				[]string{"labelname"},
			)
			cntVec.WithLabelValues("val1").Inc()
			cntVec.WithLabelValues("val2").Inc()
			reg.MustRegister(cntVec)

			reg.MustRegister(errorCollector{})

			logBuf = &bytes.Buffer{}
			logger := log.New(logBuf, "", 0)

			errorHandler = HandlerFor(reg, HandlerOpts{
				ErrorLog:      logger,
				ErrorHandling: HTTPErrorOnError,
				Registry:      reg,
			})
			continueHandler = HandlerFor(reg, HandlerOpts{
				ErrorLog:      logger,
				ErrorHandling: ContinueOnError,
				Registry:      reg,
			})
			panicHandler = HandlerFor(reg, HandlerOpts{
				ErrorLog:      logger,
				ErrorHandling: PanicOnError,
				Registry:      reg,
			})
		})

		wantMsg := `error gathering metrics: error collecting metric Desc{fqName: "invalid_metric", help: "not helpful", constLabels: {}, variableLabels: []}: collect error
`
		wantErrorBody := `An error has occurred while serving metrics:

error collecting metric Desc{fqName: "invalid_metric", help: "not helpful", constLabels: {}, variableLabels: []}: collect error`
		wantOKBody1 := `# HELP name docstring
# TYPE name counter
name{constname="constvalue",labelname="val1"} 1
name{constname="constvalue",labelname="val2"} 1
# HELP promfasthttp_metric_handler_errors_total Total number of internal errors encountered by the promfasthttp metric handler.
# TYPE promfasthttp_metric_handler_errors_total counter
promfasthttp_metric_handler_errors_total{cause="encoding"} 0
promfasthttp_metric_handler_errors_total{cause="gathering"} 1
# HELP the_count Ah-ah-ah! Thunder and lightning!
# TYPE the_count counter
the_count 0
`
		// It might happen that counting the gathering error makes it to the
		// promfasthttp_metric_handler_errors_total counter before it is gathered
		// itself. Thus, we have to bodies that are acceptable for the test.
		wantOKBody2 := `# HELP name docstring
# TYPE name counter
name{constname="constvalue",labelname="val1"} 1
name{constname="constvalue",labelname="val2"} 1
# HELP promfasthttp_metric_handler_errors_total Total number of internal errors encountered by the promfasthttp metric handler.
# TYPE promfasthttp_metric_handler_errors_total counter
promfasthttp_metric_handler_errors_total{cause="encoding"} 0
promfasthttp_metric_handler_errors_total{cause="gathering"} 2
# HELP the_count Ah-ah-ah! Thunder and lightning!
# TYPE the_count counter
the_count 0
`

		It("should return a http error on error", func() {
			ctx := createRequestCtx("GET", "/http")
			ctx.Request.Header.Add("Accept", "text/plain")
			errorHandler(ctx)

			got, want := ctx.Response.StatusCode(), fasthttp.StatusInternalServerError
			Expect(got).To(Equal(want))

			Expect(logBuf.String()).To(Equal(wantMsg))
			Expect(string(ctx.Response.Body())).To(Equal(wantErrorBody))
		})

		It("should continue on error", func() {
			ctx := createRequestCtx("GET", "/continue")
			ctx.Request.Header.Add("Accept", "text/plain")
			continueHandler(ctx)

			Expect(ctx.Response.StatusCode()).To(Equal(fasthttp.StatusOK))
			Expect(logBuf.String()).To(Equal(wantMsg))
			Expect(string(ctx.Response.Body())).To(Or(Equal(wantOKBody1), Equal(wantOKBody2)))
		})

		It("should panic on error", func() {
			defer func() {
				if err := recover(); err == nil {
					Fail("expected panic from panicHandler")
				}
			}()

			ctx := createRequestCtx("GET", "/panic")
			ctx.Request.Header.Add("Accept", "text/plain")
			panicHandler(ctx)
		})
	})

	When("using InstrumentMetricHandler", func() {
		It("should be idempotency", func() {
			reg := prometheus.NewRegistry()
			h := InstrumentMetricHandler(reg, HandlerFor(reg, HandlerOpts{}))
			// Do it again to test idempotency.
			InstrumentMetricHandler(reg, HandlerFor(reg, HandlerOpts{}))

			reqCtx1 := createRequestCtx("GET", "/")
			reqCtx1.Request.Header.Add("Accept", "test/plain")

			h(reqCtx1)
			Expect(reqCtx1.Response.StatusCode()).To(Equal(fasthttp.StatusOK))

			want := "promfasthttp_metric_handler_requests_in_flight 1\n"
			Expect(string(reqCtx1.Response.Body())).To(ContainSubstring(want))
			want = "promfasthttp_metric_handler_requests_total{code=\"200\"} 0\n"
			Expect(string(reqCtx1.Response.Body())).To(ContainSubstring(want))

			reqCtx2 := createRequestCtx("GET", "/")
			reqCtx2.Request.Header.Add("Accept", "test/plain")

			h(reqCtx2)
			Expect(reqCtx2.Response.StatusCode()).To(Equal(fasthttp.StatusOK))

			want = "promfasthttp_metric_handler_requests_in_flight 1\n"
			Expect(string(reqCtx2.Response.Body())).To(ContainSubstring(want))

			want = "promfasthttp_metric_handler_requests_total{code=\"200\"} 1\n"
			Expect(string(reqCtx2.Response.Body())).To(ContainSubstring(want))
		})
	})

	When("using MaxRequestsInFlight", func() {
		It("should fail when exceeded", func() {
			reg := prometheus.NewRegistry()

			handler := HandlerFor(reg, HandlerOpts{MaxRequestsInFlight: 1})

			reqCtx1 := createRequestCtx("GET", "/")
			reqCtx1.Request.Header.Add("Accept", "test/plain")

			reqCtx2 := createRequestCtx("GET", "/")
			reqCtx2.Request.Header.Add("Accept", "test/plain")

			reqCtx3 := createRequestCtx("GET", "/")
			reqCtx3.Request.Header.Add("Accept", "test/plain")

			c := blockingCollector{Block: make(chan struct{}), CollectStarted: make(chan struct{}, 1)}
			reg.MustRegister(c)

			rq1Done := make(chan struct{})
			go func() {
				handler(reqCtx1)
				close(rq1Done)
			}()
			<-c.CollectStarted

			handler(reqCtx2)

			Expect(reqCtx2.Response.StatusCode()).To(Equal(fasthttp.StatusServiceUnavailable))
			Expect(string(reqCtx2.Response.Body())).To(ContainSubstring("Limit of concurrent requests reached (1), try again later."))

			close(c.Block)
			<-rq1Done

			handler(reqCtx3)
			Expect(reqCtx3.Response.StatusCode()).To(Equal(fasthttp.StatusOK))
		})
	})

	When("using Timeout", func() {
		It("should return error when exceeded", func() {
			reg := prometheus.NewRegistry()
			handler := HandlerFor(reg, HandlerOpts{Timeout: time.Millisecond})

			c := blockingCollector{Block: make(chan struct{}), CollectStarted: make(chan struct{}, 1)}
			reg.MustRegister(c)

			r, err := http.NewRequest("GET", "http://test/", nil)
			Expect(err).ToNot(HaveOccurred())

			res, err := serve(handler, r)
			Expect(err).ToNot(HaveOccurred())

			body, err := ioutil.ReadAll(res.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(res.StatusCode).To(Equal(fasthttp.StatusRequestTimeout))
			Expect(string(body)).To(ContainSubstring("Exceeded configured timeout of 1ms."))
			close(c.Block) // To not leak a goroutine.
		})
	})
})

func serve(handler fasthttp.RequestHandler, req *http.Request) (*http.Response, error) {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		err := fasthttp.Serve(ln, handler)
		if err != nil {
			panic(fmt.Errorf("failed to serve: %v", err))
		}
	}()

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return ln.Dial()
			},
		},
	}

	return client.Do(req)
}
