package promsql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type DatabaseCollectorOpts struct {
	Prefix string
}

type dbStats interface {
	Stats() sql.DBStats
}

type databaseCollector struct {
	db                      dbStats
	descMaxOpenConnections  *prometheus.Desc
	descPoolOpenConnections *prometheus.Desc
	descPoolInUse           *prometheus.Desc
	descPoolIdle            *prometheus.Desc
	descWaitCount           *prometheus.Desc
	descWaitDuration        *prometheus.Desc
	descMaxIdleClosed       *prometheus.Desc
	descMaxLifetimeClosed   *prometheus.Desc
}

func NewDatabaseCollector(db dbStats, opts DatabaseCollectorOpts) *databaseCollector {
	prefix := opts.Prefix
	if prefix != "" && !strings.HasSuffix(opts.Prefix, "_") {
		prefix += "_"
	}
	return &databaseCollector{
		db:                      db,
		descMaxOpenConnections:  prometheus.NewDesc(fmt.Sprintf("db_%smax_open_connections", prefix), "Maximum number of open connections to the database.", nil, nil),
		descPoolOpenConnections: prometheus.NewDesc(fmt.Sprintf("db_%spool_open_connections", prefix), "The number of established connections both in use and idle.", nil, nil),
		descPoolInUse:           prometheus.NewDesc(fmt.Sprintf("db_%spool_in_use", prefix), "The number of connections currently in use.", nil, nil),
		descPoolIdle:            prometheus.NewDesc(fmt.Sprintf("db_%spool_idle", prefix), "The number of idle connections.", nil, nil),
		descWaitCount:           prometheus.NewDesc(fmt.Sprintf("db_%swait_count", prefix), "The total number of connections waited for.", nil, nil),
		descWaitDuration:        prometheus.NewDesc(fmt.Sprintf("db_%swait_duration", prefix), "The total time blocked waiting for a new connection.", nil, nil),
		descMaxIdleClosed:       prometheus.NewDesc(fmt.Sprintf("db_%smax_idle_closed", prefix), "The total number of connections closed due to SetMaxIdleConns.", nil, nil),
		descMaxLifetimeClosed:   prometheus.NewDesc(fmt.Sprintf("db_%smax_lifetime_closed", prefix), "The total number of connections closed due to SetConnMaxLifetime.", nil, nil),
	}
}

// Describe returns the description of metrics colllected by this collector.
func (collector *databaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.descMaxOpenConnections
	ch <- collector.descPoolOpenConnections
	ch <- collector.descPoolInUse
	ch <- collector.descPoolIdle
	ch <- collector.descWaitCount
	ch <- collector.descWaitDuration
	ch <- collector.descMaxIdleClosed
	ch <- collector.descMaxLifetimeClosed
}

// Collect gets the database stats information and provides it to the prometheus.
func (collector *databaseCollector) Collect(ch chan<- prometheus.Metric) {
	stats := collector.db.Stats()
	ch <- prometheus.MustNewConstMetric(collector.descMaxOpenConnections, prometheus.GaugeValue, float64(stats.MaxOpenConnections))
	ch <- prometheus.MustNewConstMetric(collector.descPoolOpenConnections, prometheus.GaugeValue, float64(stats.OpenConnections))
	ch <- prometheus.MustNewConstMetric(collector.descPoolInUse, prometheus.GaugeValue, float64(stats.InUse))
	ch <- prometheus.MustNewConstMetric(collector.descPoolIdle, prometheus.GaugeValue, float64(stats.Idle))
	ch <- prometheus.MustNewConstMetric(collector.descWaitCount, prometheus.CounterValue, float64(stats.WaitCount))
	ch <- prometheus.MustNewConstMetric(collector.descWaitDuration, prometheus.CounterValue, float64(stats.WaitDuration))
	ch <- prometheus.MustNewConstMetric(collector.descMaxIdleClosed, prometheus.CounterValue, float64(stats.MaxIdleClosed))
	ch <- prometheus.MustNewConstMetric(collector.descMaxLifetimeClosed, prometheus.CounterValue, float64(stats.MaxLifetimeClosed))
}
