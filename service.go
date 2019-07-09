package promsrv

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Service represents a Prometheus service.
type Service struct {
	r *prometheus.Registry
}

func (service *Service) registry() *prometheus.Registry {
	if service.r == nil {
		service.r = prometheus.NewRegistry()
	}
	return service.r
}

// Restart restarts the Prometheus service.
func (service *Service) Restart() error {
	if err := service.Stop(); err != nil {
		return err
	}
	return service.Start()
}

// Start starts the Prometheus service.
func (service *Service) Start() error {
	return nil
}

// Stop stops the Prometheus service.
func (service *Service) Stop() error {
	return nil
}

// Gather implements prometheus.Gatherer.
func (service *Service) Gather() ([]*dto.MetricFamily, error) {
	return service.registry().Gather()
}

// Register implements prometheus.Registerer.
func (service *Service) Register(c prometheus.Collector) error {
	return service.registry().Register(c)
}

// MustRegister implements prometheus.Registerer.
func (service *Service) MustRegister(cs ...prometheus.Collector) {
	service.MustRegister(cs...)
}

// Unregister implements prometheus.Registerer.
func (service *Service) Unregister(c prometheus.Collector) bool {
	return service.registry().Unregister(c)
}

// NewCounter works like the function of the same name in the prometheus package
// but it automatically registers the Counter with the
// service's internal registry. If the registration fails, NewCounter panics.
func (service *Service) NewCounter(opts prometheus.CounterOpts) prometheus.Counter {
	c := prometheus.NewCounter(opts)
	service.MustRegister(c)
	return c
}

// NewCounterVec works like the function of the same name in the prometheus
// package but it automatically registers the CounterVec with the
// service's internal registry. If the registration fails, NewCounterVec
// panics.
func (service *Service) NewCounterVec(opts prometheus.CounterOpts, labelNames []string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(opts, labelNames)
	service.MustRegister(c)
	return c
}

// NewCounterFunc works like the function of the same name in the prometheus
// package but it automatically registers the CounterFunc with the
// service's internal registry. If the registration fails, NewCounterFunc
// panics.
func (service *Service) NewCounterFunc(opts prometheus.CounterOpts, function func() float64) prometheus.CounterFunc {
	g := prometheus.NewCounterFunc(opts, function)
	service.MustRegister(g)
	return g
}

// NewGauge works like the function of the same name in the prometheus package
// but it automatically registers the Gauge with the
// service's internal registry. If the registration fails, NewGauge panics.
func (service *Service) NewGauge(opts prometheus.GaugeOpts) prometheus.Gauge {
	g := prometheus.NewGauge(opts)
	service.MustRegister(g)
	return g
}

// NewGaugeVec works like the function of the same name in the prometheus
// package but it automatically registers the GaugeVec with the
// service's internal registry. If the registration fails, NewGaugeVec panics.
func (service *Service) NewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	g := prometheus.NewGaugeVec(opts, labelNames)
	service.MustRegister(g)
	return g
}

// NewGaugeFunc works like the function of the same name in the prometheus
// package but it automatically registers the GaugeFunc with the
// service's internal registry. If the registration fails, NewGaugeFunc panics.
func (service *Service) NewGaugeFunc(opts prometheus.GaugeOpts, function func() float64) prometheus.GaugeFunc {
	g := prometheus.NewGaugeFunc(opts, function)
	service.MustRegister(g)
	return g
}

// NewSummary works like the function of the same name in the prometheus package
// but it automatically registers the Summary with the
// service's internal registry. If the registration fails, NewSummary panics.
func (service *Service) NewSummary(opts prometheus.SummaryOpts) prometheus.Summary {
	s := prometheus.NewSummary(opts)
	service.MustRegister(s)
	return s
}

// NewSummaryVec works like the function of the same name in the prometheus
// package but it automatically registers the SummaryVec with the
// service's internal registry. If the registration fails, NewSummaryVec
// panics.
func (service *Service) NewSummaryVec(opts prometheus.SummaryOpts, labelNames []string) *prometheus.SummaryVec {
	s := prometheus.NewSummaryVec(opts, labelNames)
	service.MustRegister(s)
	return s
}

// NewHistogram works like the function of the same name in the prometheus
// package but it automatically registers the Histogram with the
// service's internal registry. If the registration fails, NewHistogram panics.
func (service *Service) NewHistogram(opts prometheus.HistogramOpts) prometheus.Histogram {
	h := prometheus.NewHistogram(opts)
	service.MustRegister(h)
	return h
}

// NewHistogramVec works like the function of the same name in the prometheus
// package but it automatically registers the HistogramVec with the
// service's internal registry. If the registration fails, NewHistogramVec
// panics.
func (service *Service) NewHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	h := prometheus.NewHistogramVec(opts, labelNames)
	service.MustRegister(h)
	return h
}
