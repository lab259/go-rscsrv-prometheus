package promsql

import (
	"github.com/prometheus/client_golang/prometheus"
)

type NamedQuery struct {
	parent            *QueryCollector
	name              string
	TotalCalls        prometheus.Counter
	TotalDuration     prometheus.Counter
	TotalSuccess      prometheus.Counter
	TotalFailures     prometheus.Counter
	TotalRowsAffected prometheus.Counter
}
