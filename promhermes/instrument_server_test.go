package promhermes

import (
	"fmt"
	"log"

	"github.com/valyala/fasthttp"

	"github.com/lab259/hermes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

func createRequestCtx(method, path string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(method)
	ctx.Request.URI().SetPath(path)
	return ctx
}

var _ = Describe("Instrument Server", func() {
	When("validating labels", func() {
		scenarios := map[string]struct {
			varLabels     []string
			constLabels   []string
			curriedLabels []string
			ok            bool
		}{
			"empty": {
				varLabels:     []string{},
				constLabels:   []string{},
				curriedLabels: []string{},
				ok:            true,
			},
			"code as single var label": {
				varLabels:     []string{"code"},
				constLabels:   []string{},
				curriedLabels: []string{},
				ok:            true,
			},
			"method as single var label": {
				varLabels:     []string{"method"},
				constLabels:   []string{},
				curriedLabels: []string{},
				ok:            true,
			},
			"cade and method as var labels": {
				varLabels:     []string{"method", "code"},
				constLabels:   []string{},
				curriedLabels: []string{},
				ok:            true,
			},
			"valid case with all labels used": {
				varLabels:     []string{"code", "method"},
				constLabels:   []string{"foo", "bar"},
				curriedLabels: []string{"dings", "bums"},
				ok:            true,
			},
			"unsupported var label": {
				varLabels:     []string{"foo"},
				constLabels:   []string{},
				curriedLabels: []string{},
				ok:            false,
			},
			"mixed var labels": {
				varLabels:     []string{"method", "foo", "code"},
				constLabels:   []string{},
				curriedLabels: []string{},
				ok:            false,
			},
			"unsupported var label but curried": {
				varLabels:     []string{},
				constLabels:   []string{},
				curriedLabels: []string{"foo"},
				ok:            true,
			},
			"mixed var labels but unsupported curried": {
				varLabels:     []string{"code", "method"},
				constLabels:   []string{},
				curriedLabels: []string{"foo"},
				ok:            true,
			},
			"supported label as const and curry": {
				varLabels:     []string{},
				constLabels:   []string{"code"},
				curriedLabels: []string{"method"},
				ok:            true,
			},
			"supported label as const and curry with unsupported as var": {
				varLabels:     []string{"foo"},
				constLabels:   []string{"code"},
				curriedLabels: []string{"method"},
				ok:            false,
			},
		}

		for name, sc := range scenarios {
			It(name, func() {
				constLabels := prometheus.Labels{}
				for _, l := range sc.constLabels {
					constLabels[l] = "dummy"
				}
				c := prometheus.NewCounterVec(
					prometheus.CounterOpts{
						Name:        "c",
						Help:        "c help",
						ConstLabels: constLabels,
					},
					append(sc.varLabels, sc.curriedLabels...),
				)
				o := prometheus.ObserverVec(prometheus.NewHistogramVec(
					prometheus.HistogramOpts{
						Name:        "c",
						Help:        "c help",
						ConstLabels: constLabels,
					},
					append(sc.varLabels, sc.curriedLabels...),
				))
				for _, l := range sc.curriedLabels {
					c = c.MustCurryWith(prometheus.Labels{l: "dummy"})
					o = o.MustCurryWith(prometheus.Labels{l: "dummy"})
				}

				func() {
					defer func() {
						if err := recover(); err != nil {
							if sc.ok {
								Fail(fmt.Sprintf("unexpected panic: %s", err))
							}
						} else if !sc.ok {
							Fail("expected panic")
						}
					}()
					InstrumentHandlerCounter(c, nil)
				}()
				func() {
					defer func() {
						if err := recover(); err != nil {
							if sc.ok {
								Fail(fmt.Sprintf("unexpected panic: %s", err))
							}
						} else if !sc.ok {
							Fail("expected panic")
						}
					}()
					InstrumentHandlerDuration(o, nil)
				}()
				if sc.ok {
					// Test if wantCode and wantMethod were detected correctly.
					var wantCode, wantMethod bool
					for _, l := range sc.varLabels {
						if l == "code" {
							wantCode = true
						}
						if l == "method" {
							wantMethod = true
						}
					}
					gotCode, gotMethod := checkLabels(c)
					Expect(gotCode).To(Equal(wantCode))
					Expect(gotMethod).To(Equal(wantMethod))

					gotCode, gotMethod = checkLabels(o)
					Expect(gotCode).To(Equal(wantCode))
					Expect(gotMethod).To(Equal(wantMethod))
				}
			})
		}
	})

	PIt("should use middleware")

	It("should chain handlers", func() {
		reg := prometheus.NewRegistry()

		inFlightGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "in_flight_requests",
			Help: "A gauge of requests currently being served by the wrapped handler.",
		})

		counter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_requests_total",
				Help: "A counter for requests to the wrapped handler.",
			},
			[]string{"code", "method"},
		)

		histVec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "response_duration_seconds",
				Help:        "A histogram of request latencies.",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: prometheus.Labels{"handler": "api"},
			},
			[]string{"method"},
		)

		writeHeaderVec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "write_header_duration_seconds",
				Help:        "A histogram of time to first write latencies.",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: prometheus.Labels{"handler": "api"},
			},
			[]string{},
		)

		responseSize := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "push_request_size_bytes",
				Help:    "A histogram of request sizes for requests.",
				Buckets: []float64{200, 500, 900, 1500},
			},
			[]string{},
		)

		handler := hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
			return res.Data([]byte("OK"))
		})

		reg.MustRegister(inFlightGauge, counter, histVec, responseSize, writeHeaderVec)

		chain := InstrumentHandlerInFlight(inFlightGauge,
			InstrumentHandlerCounter(counter,
				InstrumentHandlerDuration(histVec,
					// InstrumentHandlerTimeToWriteHeader(writeHeaderVec,
					InstrumentHandlerResponseSize(responseSize, handler),
					// ),
				),
			),
		)

		router := hermes.DefaultRouter()
		router.Get("/", chain)

		ctx := createRequestCtx("GET", "/")
		router.Handler()(ctx)
	})
})

func ExampleInstrumentHandlerDuration() {
	inFlightGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "method"},
	)

	// responseSize has no labels, making it a zero-dimensional
	// ObserverVec.
	responseSize := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{},
	)

	// Create the handlers that will be wrapped by the middleware.
	pushHandler := hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		return res.Data([]byte("Push"))
	})
	pullHandler := hermes.Handler(func(req hermes.Request, res hermes.Response) hermes.Result {
		return res.Data([]byte("Pull"))
	})

	// Register all of the metrics in the standard registry.
	prometheus.MustRegister(inFlightGauge, counter, duration, responseSize)

	// Instrument the handlers with all the metrics, injecting the "handler"
	// label by currying.
	pushChain := InstrumentHandlerInFlight(inFlightGauge,
		InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": "push"}),
			InstrumentHandlerCounter(counter,
				InstrumentHandlerResponseSize(responseSize, pushHandler),
			),
		),
	)
	pullChain := InstrumentHandlerInFlight(inFlightGauge,
		InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": "pull"}),
			InstrumentHandlerCounter(counter,
				InstrumentHandlerResponseSize(responseSize, pullHandler),
			),
		),
	)

	router := hermes.DefaultRouter()
	router.Get("/metrics", DefaultHandler())
	router.Get("/push", pushChain)
	router.Get("/pull", pullChain)

	app := hermes.NewApplication(hermes.ApplicationConfig{
		HTTP: hermes.FasthttpServiceConfiguration{
			Bind: ":3000",
		},
	}, router)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
