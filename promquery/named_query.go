package promquery

import (
	"github.com/prometheus/client_golang/prometheus"
)

type NamedQueryCollector struct {
	parent        *QueryCollector
	name          string
	TotalCalls    prometheus.Counter
	TotalDuration prometheus.Counter
	TotalSuccess  prometheus.Counter
	TotalFailures prometheus.Counter
}
