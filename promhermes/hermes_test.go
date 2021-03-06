package promhermes

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/lab259/hermes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
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
			reg     *prometheus.Registry
			cnt     prometheus.Counter
			cntVec  *prometheus.CounterVec
			logBuf  *bytes.Buffer
			handler fasthttp.RequestHandler
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

			router := hermes.DefaultRouter()
			router.Get("/http", HandlerFor(reg, HandlerOpts{
				ErrorLog:      logger,
				ErrorHandling: HTTPErrorOnError,
				Registry:      reg,
			}))
			router.Get("/continue", HandlerFor(reg, HandlerOpts{
				ErrorLog:      logger,
				ErrorHandling: ContinueOnError,
				Registry:      reg,
			}))
			router.Get("/panic", HandlerFor(reg, HandlerOpts{
				ErrorLog:      logger,
				ErrorHandling: PanicOnError,
				Registry:      reg,
			}))

			handler = router.Handler()
		})

		wantMsg := `error gathering metrics: error collecting metric Desc{fqName: "invalid_metric", help: "not helpful", constLabels: {}, variableLabels: []}: collect error
`
		wantErrorBody := `An error has occurred while serving metrics: error collecting metric Desc{fqName: "invalid_metric", help: "not helpful", constLabels: {}, variableLabels: []}: collect error`
		wantOKBody1 := `# HELP name docstring
# TYPE name counter
name{constname="constvalue",labelname="val1"} 1
name{constname="constvalue",labelname="val2"} 1
# HELP promhermes_metric_handler_errors_total Total number of internal errors encountered by the promhermes metric handler.
# TYPE promhermes_metric_handler_errors_total counter
promhermes_metric_handler_errors_total{cause="encoding"} 0
promhermes_metric_handler_errors_total{cause="gathering"} 0
# HELP the_count Ah-ah-ah! Thunder and lightning!
# TYPE the_count counter
the_count 0
`
		// It might happen that counting the gathering error makes it to the
		// promhermes_metric_handler_errors_total counter before it is gathered
		// itself. Thus, we have to bodies that are acceptable for the test.
		wantOKBody2 := `# HELP name docstring
# TYPE name counter
name{constname="constvalue",labelname="val1"} 1
name{constname="constvalue",labelname="val2"} 1
# HELP promhermes_metric_handler_errors_total Total number of internal errors encountered by the promhermes metric handler.
# TYPE promhermes_metric_handler_errors_total counter
promhermes_metric_handler_errors_total{cause="encoding"} 0
promhermes_metric_handler_errors_total{cause="gathering"} 1
# HELP the_count Ah-ah-ah! Thunder and lightning!
# TYPE the_count counter
the_count 0
`

		It("should return a http error on error", func() {
			ctx := createRequestCtx("GET", "/http")
			ctx.Request.Header.Add("Accept", "text/plain")
			handler(ctx)
			got, want := ctx.Response.StatusCode(), hermes.StatusInternalServerError
			Expect(got).To(Equal(want))

			Expect(logBuf.String()).To(Equal(wantMsg))
			var httperr map[string]interface{}
			Expect(json.Unmarshal(ctx.Response.Body(), &httperr)).To(Succeed())
			Expect(httperr).To(HaveKeyWithValue("message", wantErrorBody))
			Expect(httperr).To(HaveKeyWithValue("module", "promhermes"))
			Expect(httperr).To(HaveKeyWithValue("code", "prometheus-failed"))

		})

		It("should continue on error", func() {
			ctx := createRequestCtx("GET", "/continue")
			ctx.Request.Header.Add("Accept", "text/plain")

			handler(ctx)
			Expect(ctx.Response.StatusCode()).To(Equal(hermes.StatusOK))
			Expect(logBuf.String()).To(Equal(wantMsg))
			Expect(string(ctx.Response.Body())).To(Equal(wantOKBody1))

			logBuf.Reset()
			ctx.Response.Reset()
			handler(ctx)

			Expect(ctx.Response.StatusCode()).To(Equal(hermes.StatusOK))
			Expect(logBuf.String()).To(Equal(wantMsg))
			Expect(string(ctx.Response.Body())).To(Equal(wantOKBody2))
		})

		It("should panic on error", func() {
			defer func() {
				if err := recover(); err == nil {
					Fail("expected panic from panicHandler")
				}
			}()

			ctx := createRequestCtx("GET", "/panic")
			ctx.Request.Header.Add("Accept", "text/plain")
			handler(ctx)
		})
	})

	When("using InstrumentMetricHandler", func() {
		It("should be idempotency", func() {
			reg := prometheus.NewRegistry()
			h := InstrumentMetricHandler(reg, HandlerFor(reg, HandlerOpts{}))
			// Do it again to test idempotency.
			InstrumentMetricHandler(reg, HandlerFor(reg, HandlerOpts{}))

			router := hermes.DefaultRouter()
			router.Get("/", h)

			handler := router.Handler()

			reqCtx1 := createRequestCtx("GET", "/")
			reqCtx1.Request.Header.Add("Accept", "test/plain")

			handler(reqCtx1)
			Expect(reqCtx1.Response.StatusCode()).To(Equal(hermes.StatusOK))

			want := "promhermes_metric_handler_requests_in_flight 1\n"
			Expect(string(reqCtx1.Response.Body())).To(ContainSubstring(want))
			want = "promhermes_metric_handler_requests_total{code=\"200\"} 0\n"
			Expect(string(reqCtx1.Response.Body())).To(ContainSubstring(want))

			reqCtx2 := createRequestCtx("GET", "/")
			reqCtx2.Request.Header.Add("Accept", "test/plain")

			handler(reqCtx2)
			Expect(reqCtx2.Response.StatusCode()).To(Equal(hermes.StatusOK))

			want = "promhermes_metric_handler_requests_in_flight 1\n"
			Expect(string(reqCtx2.Response.Body())).To(ContainSubstring(want))

			want = "promhermes_metric_handler_requests_total{code=\"200\"} 1\n"
			Expect(string(reqCtx2.Response.Body())).To(ContainSubstring(want))
		})
	})

	When("using MaxRequestsInFlight", func() {
		It("should fail when exceeded", func() {
			reg := prometheus.NewRegistry()

			router := hermes.DefaultRouter()
			router.Get("/", HandlerFor(reg, HandlerOpts{MaxRequestsInFlight: 1}))

			handler := router.Handler()

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

			Expect(reqCtx2.Response.StatusCode()).To(Equal(hermes.StatusServiceUnavailable))
			Expect(string(reqCtx2.Response.Body())).To(ContainSubstring("Limit of concurrent requests reached (1), try again later."))

			close(c.Block)
			<-rq1Done

			handler(reqCtx3)
			Expect(reqCtx3.Response.StatusCode()).To(Equal(hermes.StatusOK))
		})
	})

	When("using Timeout", func() {
		It("should return error when exceeded", func() {
			reg := prometheus.NewRegistry()
			handler := HandlerFor(reg, HandlerOpts{Timeout: time.Millisecond})

			c := blockingCollector{Block: make(chan struct{}), CollectStarted: make(chan struct{}, 1)}
			reg.MustRegister(c)

			router := hermes.DefaultRouter()
			router.Get("/", handler)

			ctx := createRequestCtx("GET", "/")
			router.Handler()(ctx)

			Expect(ctx.Response.StatusCode()).To(Equal(hermes.StatusRequestTimeout))
			Expect(string(ctx.Response.Body())).To(ContainSubstring("Exceeded configured timeout of 1ms."))
			close(c.Block) // To not leak a goroutine.
		})
	})
})
