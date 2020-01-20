package promsql_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"

	"github.com/lab259/go-rscsrv-prometheus/ginkgotest"
	"github.com/lab259/go-rscsrv-prometheus/promsql"
	_ "github.com/lib/pq"
)

func TestPromsqlDriverCollector(t *testing.T) {
	ginkgotest.Init("PromsqlCollector Driver Test Suite", t)
}

var _ = Describe("Driver Collector", func() {

	BeforeEach(func() {
		driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
			DriverName: "postgres",
		})
		Expect(err).ShouldNot(HaveOccurred())

		psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
		db, err := sql.Open(driverCollector.DriverName, psqlInfo)
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
		driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
			DriverName: "postgres",
		})
		Expect(err).ShouldNot(HaveOccurred())

		psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
		db, err := sql.Open(driverCollector.DriverName, psqlInfo)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(db.Ping()).ShouldNot(HaveOccurred())
		defer db.Close()

		_, err = db.Exec(`
		DROP TABLE users`)
		Expect(err).ShouldNot(HaveOccurred())
	})

	When("using queries", func() {
		It("should increase amount to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			rs, err := db.Query("select name from users where id = 1")
			Expect(err).ShouldNot(HaveOccurred())
			defer rs.Close()

			var metric dto.Metric
			Expect(driverCollector.QueryTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase amount to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			rs, err := db.Query("select name from users where id = 1")
			Expect(err).ShouldNot(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.QueryTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))

			stmt, err := db.Prepare("select name from users where id = 1")
			Expect(err).ShouldNot(HaveOccurred())

			rs, err = stmt.QueryContext(context.Background())
			Expect(err).ShouldNot(HaveOccurred())
			defer rs.Close()

			Expect(driverCollector.QueryTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})

		It("should increase successful to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			rs, err := db.Query("select name from users where id = 1")
			Expect(err).ShouldNot(HaveOccurred())
			defer rs.Close()

			var metric dto.Metric
			Expect(driverCollector.QuerySuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase successful to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			rs, err := db.Query("select name from users where id = 1")
			Expect(err).ShouldNot(HaveOccurred())
			defer rs.Close()

			var metric dto.Metric
			Expect(driverCollector.QuerySuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))

			stmt, err := db.Prepare("select name from users where id = 1")
			Expect(err).ShouldNot(HaveOccurred())

			rs, err = stmt.Query()
			Expect(err).ShouldNot(HaveOccurred())

			Expect(driverCollector.QuerySuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})

		It("should increase failed to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			_, err = db.Query("wrong example query")
			Expect(err).Should(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.QueryFailedCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase failed to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			_, err = db.Query("wrong example query")
			Expect(err).Should(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.QueryFailedCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))

			_, err = db.Query("wrong example query")
			Expect(err).Should(HaveOccurred())

			Expect(driverCollector.QueryFailedCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})
	})

	When("using transactions commit", func() {
		It("should increase total amount to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			tx, err := db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			stmt, err := tx.Prepare(`INSERT INTO users (id, name)
				VALUES(2, 'LeBron James');`)
			Expect(err).ShouldNot(HaveOccurred())

			defer stmt.Close()

			rs, err := stmt.Exec()
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err := rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			Expect(tx.Commit()).ShouldNot(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.TransactionCommitTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase total amount to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			// First transaction
			tx, err := db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			rs, err := tx.Exec(`INSERT INTO users (id, name)
				VALUES(2, 'LeBron James');`)
			Expect(err).ShouldNot(HaveOccurred())

			rowsCounter, err := rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			Expect(tx.Commit()).ShouldNot(HaveOccurred())

			// Second transaction
			tx, err = db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			stmt, err := tx.Prepare(`INSERT INTO users (id, name)
				VALUES(3, 'Chistian Bale');`)
			Expect(err).ShouldNot(HaveOccurred())
			defer stmt.Close()

			rs, err = stmt.Exec()
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err = rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			Expect(tx.Commit()).ShouldNot(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.TransactionCommitTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})

		It("should increase successful to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			tx, err := db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			stmt, err := tx.Prepare(`INSERT INTO users (id, name)
				VALUES(2, 'LeBron James');`)
			Expect(err).ShouldNot(HaveOccurred())

			defer stmt.Close()

			rs, err := stmt.Exec()
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err := rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			Expect(tx.Commit()).ShouldNot(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.TransactionCommitSuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase successful to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			// First transaction
			tx, err := db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			rs, err := tx.Exec(`INSERT INTO users (id, name)
				VALUES(2, 'LeBron James');`)
			Expect(err).ShouldNot(HaveOccurred())

			rowsCounter, err := rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			Expect(tx.Commit()).ShouldNot(HaveOccurred())

			// Second transaction
			tx, err = db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			stmt, err := tx.Prepare(`INSERT INTO users (id, name)
				VALUES(3, 'Chistian Bale');`)
			Expect(err).ShouldNot(HaveOccurred())
			defer stmt.Close()

			rs, err = stmt.Exec()
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err = rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			Expect(tx.Commit()).ShouldNot(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.TransactionCommitSuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})

		It("should increase failed to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			// First transaction
			tx, err := db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			stmt, err := tx.Prepare(`INSERT INTO users (id, name)
			VALUES(1, 'WRONG REPEATED USER');`)
			Expect(err).ShouldNot(HaveOccurred())
			defer stmt.Close()

			_, err = stmt.Exec()
			Expect(err).Should(HaveOccurred())

			Expect(tx.Commit()).Should(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.TransactionCommitFailedCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase failed to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			// First transaction
			tx, err := db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			stmt, err := tx.Prepare(`INSERT INTO users (id, name)
				VALUES(1, 'WRONG REPEATED USER');`)
			Expect(err).ShouldNot(HaveOccurred())
			defer stmt.Close()

			_, err = stmt.Exec()
			Expect(err).Should(HaveOccurred())

			Expect(tx.Commit()).Should(HaveOccurred())

			// Second transaction
			tx, err = db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			stmt, err = tx.Prepare(`INSERT INTO users (id, name)
				VALUES(1, 'WRONG REPEATED USER');`)
			Expect(err).ShouldNot(HaveOccurred())
			defer stmt.Close()

			_, err = stmt.Exec()
			Expect(err).Should(HaveOccurred())

			Expect(tx.Commit()).Should(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.TransactionCommitFailedCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})
	})

	When("using transactions rollback", func() {
		It("should increase successful amount to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			tx, err := db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tx.Exec(`INSERT INTO users (id, name)
				VALUES(1, 'LeBron James');`)
			Expect(err).Should(HaveOccurred())

			Expect(tx.Rollback()).ShouldNot(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.TransactionCommitTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			Expect(driverCollector.TransactionRollbackTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			Expect(driverCollector.TransactionRollbackSuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase successful amount to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			// First transaction
			tx, err := db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tx.Exec(`INSERT INTO users (id, name)
				VALUES(1, 'LeBron James');`)
			Expect(err).Should(HaveOccurred())

			Expect(tx.Rollback()).ShouldNot(HaveOccurred())

			// Second transaction
			tx, err = db.Begin()
			Expect(err).ShouldNot(HaveOccurred())

			_, err = tx.Exec(`INSERT INTO users (id, name)
				VALUES(1, 'Chistian Bale');`)
			Expect(err).Should(HaveOccurred())

			Expect(tx.Rollback()).ShouldNot(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.TransactionCommitTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			Expect(driverCollector.TransactionRollbackTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
			Expect(driverCollector.TransactionRollbackSuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})
	})

	When("using executions", func() {
		It("should increase amount to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			rs, err := db.Exec(`INSERT INTO users (id, name)
				VALUES(2, 'LeBron James');`)
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err := rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			var metric dto.Metric
			Expect(driverCollector.ExecutionTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase amount to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			rs, err := db.Exec(`INSERT INTO users (id, name)
				VALUES(2, 'LeBron James');`)
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err := rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			stmt, err := db.Prepare(`INSERT INTO users (id, name)
			VALUES(3, 'Chistian Bale');`)
			Expect(err).ShouldNot(HaveOccurred())

			defer stmt.Close()

			rs, err = stmt.Exec()
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err = rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			var metric dto.Metric
			Expect(driverCollector.ExecutionTotalCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})

		It("should increase successful to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			rs, err := db.Exec(`INSERT INTO users (id, name)
				VALUES(2, 'LeBron James');`)
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err := rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			var metric dto.Metric
			Expect(driverCollector.ExecutionSuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase successful to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			rs, err := db.Exec(`INSERT INTO users (id, name)
				VALUES(2, 'LeBron James');`)
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err := rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			stmt, err := db.Prepare(`INSERT INTO users (id, name)
			VALUES(3, 'Chistian Bale');`)
			Expect(err).ShouldNot(HaveOccurred())

			rs, err = stmt.Exec()
			Expect(err).ShouldNot(HaveOccurred())
			rowsCounter, err = rs.RowsAffected()
			Expect(rowsCounter).To(BeEquivalentTo(1))

			var metric dto.Metric
			Expect(driverCollector.ExecutionSuccessfulCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})

		It("should increase failed to 1", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			_, err = db.Exec(`INSERT INTO users (id, name)
			VALUES(1, 'WRONG REPEATED USER');`)
			Expect(err).Should(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.ExecutionFailedCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
		})

		It("should increase failed to 2", func() {
			driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
				DriverName: "postgres",
			})
			Expect(err).ShouldNot(HaveOccurred())

			psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=pg-test sslmode=disable")
			db, err := sql.Open(driverCollector.DriverName, psqlInfo)
			Expect(db.Ping()).ShouldNot(HaveOccurred())

			_, err = db.Exec(`INSERT INTO users (id, name)
				VALUES(1, 'WRONG REPEATED USER');`)
			Expect(err).Should(HaveOccurred())

			_, err = db.Exec(`INSERT INTO users (id, name)
				VALUES(1, 'WRONG REPEATED USER');`)
			Expect(err).Should(HaveOccurred())

			var metric dto.Metric
			Expect(driverCollector.ExecutionFailedCounter.Write(&metric)).To(Succeed())
			Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
		})
	})
})
