package promsql_test

import (
	"database/sql"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/lab259/go-rscsrv-prometheus/ginkgotest"
	"github.com/lab259/go-rscsrv-prometheus/promsql"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestPromsqlCollector(t *testing.T) {
	ginkgotest.Init("PromsqlCollector Test Suite", t)
}

type fakeDB struct {
}

func (*fakeDB) Stats() sql.DBStats {
	return sql.DBStats{
		MaxOpenConnections: 1,
		OpenConnections:    2,
		InUse:              3,
		Idle:               4,
		WaitCount:          5,
		WaitDuration:       6,
		MaxIdleClosed:      7,
		MaxLifetimeClosed:  8,
	}
}

var _ = Describe("Database Collector", func() {
	It("should generate default description names", func() {
		collector := promsql.NewDatabaseCollector(nil, promsql.DatabaseCollectorOpts{})
		ch := make(chan *prometheus.Desc)
		go func() {
			collector.Describe(ch)
			close(ch)
		}()

		Expect((<-ch).String()).To(ContainSubstring("db_max_open_connections"))
		Expect((<-ch).String()).To(ContainSubstring("db_pool_open_connections"))
		Expect((<-ch).String()).To(ContainSubstring("db_pool_in_use"))
		Expect((<-ch).String()).To(ContainSubstring("db_pool_idle"))
		Expect((<-ch).String()).To(ContainSubstring("db_wait_count"))
		Expect((<-ch).String()).To(ContainSubstring("db_wait_duration"))
		Expect((<-ch).String()).To(ContainSubstring("db_max_idle_closed"))
		Expect((<-ch).String()).To(ContainSubstring("db_max_lifetime_closed"))
	})

	It("should generate description names with prefix", func() {
		collector := promsql.NewDatabaseCollector(nil, promsql.DatabaseCollectorOpts{
			Prefix: "test",
		})
		ch := make(chan *prometheus.Desc)
		go func() {
			collector.Describe(ch)
			close(ch)
		}()

		Expect((<-ch).String()).To(ContainSubstring("db_test_max_open_connections"))
		Expect((<-ch).String()).To(ContainSubstring("db_test_pool_open_connections"))
		Expect((<-ch).String()).To(ContainSubstring("db_test_pool_in_use"))
		Expect((<-ch).String()).To(ContainSubstring("db_test_pool_idle"))
		Expect((<-ch).String()).To(ContainSubstring("db_test_wait_count"))
		Expect((<-ch).String()).To(ContainSubstring("db_test_wait_duration"))
		Expect((<-ch).String()).To(ContainSubstring("db_test_max_idle_closed"))
		Expect((<-ch).String()).To(ContainSubstring("db_test_max_lifetime_closed"))
	})

	It("should generate default metric values", func() {
		collector := promsql.NewDatabaseCollector(&fakeDB{}, promsql.DatabaseCollectorOpts{})
		ch := make(chan prometheus.Metric)
		go func() {
			collector.Collect(ch)
			close(ch)
		}()

		var metric dto.Metric
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetGauge().GetValue()).To(BeEquivalentTo(1.0))
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetGauge().GetValue()).To(BeEquivalentTo(2.0))
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetGauge().GetValue()).To(BeEquivalentTo(3.0))
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetGauge().GetValue()).To(BeEquivalentTo(4.0))
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(5.0))
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(6.0))
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(7.0))
		Expect((<-ch).Write(&metric)).To(Succeed())
		Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(8))
	})
})
