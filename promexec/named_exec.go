package promexec

import (
	"github.com/prometheus/client_golang/prometheus"
)

type NamedExecCollector struct {
	parent            *ExecCollector
	name              string
	TotalCalls        prometheus.Counter
	TotalDuration     prometheus.Counter
	TotalSuccess      prometheus.Counter
	TotalFailures     prometheus.Counter
	TotalRowsAffected prometheus.Counter
}
