package promsql

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// QueryCollector is the collector responsible to initialize and control all
// available metrics. Those metrics are reused by `NamedQuery` with the query
// name as a label value.
type QueryCollector struct {
	totalCalls        *prometheus.CounterVec
	totalDuration     *prometheus.CounterVec
	totalSuccesses    *prometheus.CounterVec
	totalFailures     *prometheus.CounterVec
	totalRowsAffected *prometheus.CounterVec
}

// QueryHandler is returned by the `QueryCollector.NamedQuery` helper method for
// returning the `Query` proxy instance.
type QueryHandler func(DBQueryProxy) *Query

// QueryCollectorOpts is the input option for QueryCollector.
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
				Name: fmt.Sprintf("namedqry_%stotal_duration_seconds", prefix),
				Help: "The total duration (in seconds) from a query processed",
			},
			queryCollectorLabels,
		),
		totalSuccesses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("namedqry_%stotal_successes", prefix),
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
				Name: fmt.Sprintf("namedqry_%stotal_rows_affected", prefix),
				Help: "The total number of rows affected by a query",
			},
			queryCollectorLabels,
		),
	}
}

// NewNamedQuery returns a new instance of `NamedQuery` with its metrics
// initalized with the query name as a label.
func (collector *QueryCollector) NewNamedQuery(name string) *NamedQuery {
	return &NamedQuery{
		parent:            collector,
		name:              name,
		TotalCalls:        collector.totalCalls.WithLabelValues(name),
		TotalDuration:     collector.totalDuration.WithLabelValues(name),
		TotalSuccess:      collector.totalSuccesses.WithLabelValues(name),
		TotalFailures:     collector.totalFailures.WithLabelValues(name),
		TotalRowsAffected: collector.totalRowsAffected.WithLabelValues(name),
	}
}

// NamedQuery creates internally a `NamedQuery` and then returns a 2nd order
// function that uses the NamedQuery to create mew `Query` instances.
//
// Example:
//
// ```
// nqFUH := qryCollector.NamedQuery("fetch_users_history")
// // ...
// rs, err := nqFUH(db).Query("SELECT id, ... FROM  ... INNER JOIN ... INNER JOIN ... WHERE ...")
// ```
//
// In the example above, a new `Query` is created everytime `nqFUH` is
// called. Also, a `sql.TX` can be used instead of using a `sql.DB` reference.
func (collector *QueryCollector) NamedQuery(name string) QueryHandler {
	nqry := collector.NewNamedQuery(name)
	return func(db DBQueryProxy) *Query {
		return NewQuery(nqry, db)
	}
}

// Describe forwards all descriptions to the metrics.
func (collector *QueryCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.totalCalls.Describe(ch)
	collector.totalDuration.Describe(ch)
	collector.totalSuccesses.Describe(ch)
	collector.totalFailures.Describe(ch)
	collector.totalRowsAffected.Describe(ch)
}

// Collect forwards all collect calls to all metrics.
func (collector *QueryCollector) Collect(metrics chan<- prometheus.Metric) {
	collector.totalCalls.Collect(metrics)
	collector.totalDuration.Collect(metrics)
	collector.totalSuccesses.Collect(metrics)
	collector.totalFailures.Collect(metrics)
	collector.totalRowsAffected.Collect(metrics)
}
