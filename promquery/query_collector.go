package promquery

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// QueryCollector is responsible for concentrate all metrics avaiable
type QueryCollector struct {
	totalCalls        *prometheus.CounterVec
	totalDuration     *prometheus.CounterVec
	totalSuccess      *prometheus.CounterVec
	totalFailures     *prometheus.CounterVec
	totalRowsAffected *prometheus.CounterVec
}

// QueryCollectorOpts is the input option for QueryCollector
type QueryCollectorOpts struct {
	// Prefix: responsible for all counters descs prefix
	Prefix string
}

var queryCollectorLabels = []string{"name"}

// NewQueryCollector returns a new QueryCollector pointer
func NewQueryCollector(opts *QueryCollectorOpts) *QueryCollector {
	prefix := opts.Prefix
	if prefix != "" && !strings.HasSuffix(opts.Prefix, "_") {
		prefix += "_"
	}

	// TODO: Add prefix name and descriptions
	return &QueryCollector{
		totalCalls: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("namedqry_%stotal_calls", prefix),
				Help: "The total number of calls from a query",
			},
			queryCollectorLabels,
		),
		totalDuration: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("namedqry_%stotal_duration", prefix),
				Help: "The total duration (in seconds) from a query processed",
			},
			queryCollectorLabels,
		),
		totalSuccess: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("namedqry_%stotal_success", prefix),
				Help: "The total number of a query processed with success",
			},
			queryCollectorLabels,
		),
		totalFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("namedqry_%stotal_failures", prefix),
				Help: "The total number of a query processed with failure",
			},
			queryCollectorLabels,
		),
		totalRowsAffected: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("db_%stotal_rows_affected", prefix),
				Help: "The total number of rows affected by a query",
			},
			queryCollectorLabels,
		),
	}
}

// NewNamedQuery returns a new NamedQueryCollector pointer
func (collector *QueryCollector) NewNamedQuery(name string) *NamedQueryCollector {
	return &NamedQueryCollector{
		parent:            collector,
		name:              name,
		TotalCalls:        collector.totalCalls.WithLabelValues(name),
		TotalDuration:     collector.totalDuration.WithLabelValues(name),
		TotalSuccess:      collector.totalSuccess.WithLabelValues(name),
		TotalFailures:     collector.totalFailures.WithLabelValues(name),
		TotalRowsAffected: collector.totalRowsAffected.WithLabelValues(name),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent. The sent descriptors fulfill the
// consistency and uniqueness requirements described in the Desc
// documentation.
//
// It is valid if one and the same Collector sends duplicate
// descriptors. Those duplicates are simply ignored. However, two
// different Collectors must not send duplicate descriptors.
//
// Sending no descriptor at all marks the Collector as “unchecked”,
// i.e. no checks will be performed at registration time, and the
// Collector may yield any Metric it sees fit in its Collect method.
//
// This method idempotently sends the same descriptors throughout the
// lifetime of the Collector. It may be called concurrently and
// therefore must be implemented in a concurrency safe way.
//
// If a Collector encounters an error while executing this method, it
// must send an invalid descriptor (created with NewInvalidDesc) to
// signal the error to the registry.
func (collector *QueryCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.totalCalls.Describe(ch)
	collector.totalDuration.Describe(ch)
	collector.totalSuccess.Describe(ch)
	collector.totalFailures.Describe(ch)
	collector.totalRowsAffected.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent. The
// descriptor of each sent metric is one of those returned by Describe
// (unless the Collector is unchecked, see above). Returned metrics that
// share the same descriptor must differ in their variable label
// values.
//
// This method may be called concurrently and must therefore be
// implemented in a concurrency safe way. Blocking occurs at the expense
// of total performance of rendering all registered metrics. Ideally,
// Collector implementations support concurrent readers.
func (collector *QueryCollector) Collect(metrics chan<- prometheus.Metric) {
	collector.totalCalls.Collect(metrics)
	collector.totalDuration.Collect(metrics)
	collector.totalSuccess.Collect(metrics)
	collector.totalFailures.Collect(metrics)
	collector.totalRowsAffected.Collect(metrics)
}
