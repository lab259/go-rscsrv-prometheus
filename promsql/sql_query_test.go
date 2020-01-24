package promsql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"

	"github.com/lab259/go-rscsrv-prometheus/ginkgotest"
	"github.com/lab259/go-rscsrv-prometheus/promquery"
	_ "github.com/lib/pq"
)

func TestPromsqlSQLQuery(t *testing.T) {
	ginkgotest.Init("Promsql PromSQLQuery Test Suite", t)
}

var _ = Describe("PromSQLQuery", func() {

	var dbInfo = fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
	var dbDriverName = "postgres"

	BeforeEach(func() {
		db, err := sql.Open(dbDriverName, dbInfo)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(db.Ping()).ShouldNot(HaveOccurred())

		defer db.Close()

		_, err = db.Exec(`
		CREATE TABLE users (
			id int NOT NULL PRIMARY KEY,
			name text NOT NULL
		);

		INSERT INTO users (id, name)
		VALUES (1, 'john')
		`)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		db, err := sql.Open(dbDriverName, dbInfo)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(db.Ping()).ShouldNot(HaveOccurred())
		defer db.Close()

		_, err = db.Exec(`
		DROP TABLE users`)
		Expect(err).ShouldNot(HaveOccurred())
	})

	When("using Query", func() {

		var parentQueryCollector *promquery.QueryCollector
		var usersNamedQuery *promquery.NamedQueryCollector
		BeforeEach(func() {
			parentQueryCollector = promquery.NewQueryCollector(&promquery.QueryCollectorOpts{})
			usersNamedQuery = parentQueryCollector.NewNamedQuery("users")
		})

		It("should increase success to 1", func() {
			db, err := sql.Open("postgres", dbInfo)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			promSQL := NewSQLQuery(usersNamedQuery, db)

			rows, err := promSQL.Query("select name from users where id = 1")
			Expect(err).ShouldNot(HaveOccurred())
			defer rows.Close()

			var metric dto.Metric
			Expect(usersNamedQuery.TotalCalls.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(usersNamedQuery.TotalSuccess.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(usersNamedQuery.TotalFailures.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			Expect(usersNamedQuery.TotalDuration.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
		})

		It("should increase failures to 1", func() {
			db, err := sql.Open("postgres", dbInfo)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			promSQL := NewSQLQuery(usersNamedQuery, db)

			_, err = promSQL.Query("wrong query")
			Expect(err).Should(HaveOccurred())

			var metric dto.Metric
			Expect(usersNamedQuery.TotalCalls.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(usersNamedQuery.TotalSuccess.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			Expect(usersNamedQuery.TotalFailures.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(usersNamedQuery.TotalDuration.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
		})
	})

	When("using QueryContext", func() {

		var parentQueryCollector *promquery.QueryCollector
		var usersNamedQuery *promquery.NamedQueryCollector
		BeforeEach(func() {
			parentQueryCollector = promquery.NewQueryCollector(&promquery.QueryCollectorOpts{})
			usersNamedQuery = parentQueryCollector.NewNamedQuery("users")
		})

		It("should increase success to 1", func() {
			db, err := sql.Open("postgres", dbInfo)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			promSQL := NewSQLQuery(usersNamedQuery, db)

			rows, err := promSQL.QueryContext(context.Background(), "select name from users where id = 1")
			Expect(err).ShouldNot(HaveOccurred())
			defer rows.Close()

			var metric dto.Metric
			Expect(usersNamedQuery.TotalCalls.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(usersNamedQuery.TotalSuccess.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(usersNamedQuery.TotalFailures.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			Expect(usersNamedQuery.TotalDuration.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
		})

		It("should increase failures to 1", func() {
			db, err := sql.Open("postgres", dbInfo)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			promSQL := NewSQLQuery(usersNamedQuery, db)

			_, err = promSQL.QueryContext(context.Background(), "wrong query")
			Expect(err).Should(HaveOccurred())

			var metric dto.Metric
			Expect(usersNamedQuery.TotalCalls.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(usersNamedQuery.TotalSuccess.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			Expect(usersNamedQuery.TotalFailures.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(usersNamedQuery.TotalDuration.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
		})
	})
})
