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
	"github.com/lab259/go-rscsrv-prometheus/promexec"
	_ "github.com/lib/pq"
)

func TestPromsqlSQLExec(t *testing.T) {
	ginkgotest.Init("Promsql PromSQLExec Test Suite", t)
}

var _ = Describe("PromSQLExec", func() {

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

	When("using Exec", func() {
		When("executing INSERT", func() {

			var parentExecCollector *promexec.ExecCollector
			var usersNamedExec *promexec.NamedExecCollector
			BeforeEach(func() {
				parentExecCollector = promexec.NewExecCollector(&promexec.ExecCollectorOpts{})
				usersNamedExec = parentExecCollector.NewNamedExec("users")
			})

			It("should increase success to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.Exec("insert into users (id, name) values (2, 'Marçal')")
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(1))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			})

			It("should increase failures to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				// inserting without id
				_, err = promSQL.Exec("insert into users (name) values ('Fred')")
				Expect(err).Should(HaveOccurred())

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			})

			It("should increase rows affected to 2", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.Exec(`insert into users (id, name) values
					 (2, 'Marçal'),
					 (3, 'Beethoven')`)
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(2))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
			})
		})

		When("executing UPDATE", func() {

			var parentExecCollector *promexec.ExecCollector
			var usersNamedExec *promexec.NamedExecCollector
			BeforeEach(func() {
				parentExecCollector = promexec.NewExecCollector(&promexec.ExecCollectorOpts{})
				usersNamedExec = parentExecCollector.NewNamedExec("users")
			})

			It("should increase success to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.Exec("update users set name = 'Charles' where id = 1")
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(1))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			})

			It("should increase failures to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				// Updating with wrong field
				_, err = promSQL.Exec("update users set wrongfield = 'Ops' where id = 1")
				Expect(err).Should(HaveOccurred())

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			})

			It("should increase rows affected to 2", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				// Inserting data before update
				_, err = db.Exec(`insert into users (id, name) values
					 (2, 'Marçal'),
					 (3, 'Beethoven')`)
				Expect(err).ShouldNot(HaveOccurred())

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.Exec(`update users
					set name = 'New name'
					where id = 2 or id = 3`)
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(2))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
			})
		})

		When("executing DELETE", func() {

			var parentExecCollector *promexec.ExecCollector
			var usersNamedExec *promexec.NamedExecCollector
			BeforeEach(func() {
				parentExecCollector = promexec.NewExecCollector(&promexec.ExecCollectorOpts{})
				usersNamedExec = parentExecCollector.NewNamedExec("users")
			})

			It("should increase success to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.Exec("delete from users where id = 1")
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(1))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			})

			It("should increase failures to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				// Deleting with wrong field
				_, err = promSQL.Exec("delete from users where wrongfield = 1")
				Expect(err).Should(HaveOccurred())

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			})

			It("should increase rows affected to 2", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				// Inserting data before update
				_, err = db.Exec(`insert into users (id, name) values
					 (2, 'Marçal'),
					 (3, 'Beethoven')`)
				Expect(err).ShouldNot(HaveOccurred())

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.Exec(`delete from users
					where id = 2 or id = 3`)
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(2))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
			})
		})
	})

	When("using ExecContext", func() {
		When("executing INSERT", func() {

			var parentExecCollector *promexec.ExecCollector
			var usersNamedExec *promexec.NamedExecCollector
			BeforeEach(func() {
				parentExecCollector = promexec.NewExecCollector(&promexec.ExecCollectorOpts{})
				usersNamedExec = parentExecCollector.NewNamedExec("users")
			})

			It("should increase success to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.ExecContext(context.Background(), "insert into users (id, name) values (2, 'Marçal')")
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(1))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			})

			It("should increase failures to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				// inserting without id
				_, err = promSQL.ExecContext(context.Background(), "insert into users (name) values ('Fred')")
				Expect(err).Should(HaveOccurred())

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			})

			It("should increase rows affected to 2", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.ExecContext(context.Background(), `insert into users (id, name) values
					 (2, 'Marçal'),
					 (3, 'Beethoven')`)
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(2))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
			})
		})

		When("executing UPDATE", func() {

			var parentExecCollector *promexec.ExecCollector
			var usersNamedExec *promexec.NamedExecCollector
			BeforeEach(func() {
				parentExecCollector = promexec.NewExecCollector(&promexec.ExecCollectorOpts{})
				usersNamedExec = parentExecCollector.NewNamedExec("users")
			})

			It("should increase success to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.ExecContext(context.Background(), "update users set name = 'Charles' where id = 1")
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(1))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			})

			It("should increase failures to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				// Updating with wrong field
				_, err = promSQL.ExecContext(context.Background(), "update users set wrongfield = 'Ops' where id = 1")
				Expect(err).Should(HaveOccurred())

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			})

			It("should increase rows affected to 2", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				// Inserting data before update
				_, err = db.Exec(`insert into users (id, name) values
					 (2, 'Marçal'),
					 (3, 'Beethoven')`)
				Expect(err).ShouldNot(HaveOccurred())

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.ExecContext(context.Background(), `update users
					set name = 'New name'
					where id = 2 or id = 3`)
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(2))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
			})
		})

		When("executing DELETE", func() {

			var parentExecCollector *promexec.ExecCollector
			var usersNamedExec *promexec.NamedExecCollector
			BeforeEach(func() {
				parentExecCollector = promexec.NewExecCollector(&promexec.ExecCollectorOpts{})
				usersNamedExec = parentExecCollector.NewNamedExec("users")
			})

			It("should increase success to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.ExecContext(context.Background(), "delete from users where id = 1")
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(1))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
			})

			It("should increase failures to 1", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				promSQL := NewSQLExec(usersNamedExec, db)

				// Deleting with wrong field
				_, err = promSQL.ExecContext(context.Background(), "delete from users where wrongfield = 1")
				Expect(err).Should(HaveOccurred())

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
			})

			It("should increase rows affected to 2", func() {
				db, err := sql.Open("postgres", dbInfo)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(db.Ping()).ShouldNot(HaveOccurred())
				defer db.Close()

				// Inserting data before update
				_, err = db.Exec(`insert into users (id, name) values
					 (2, 'Marçal'),
					 (3, 'Beethoven')`)
				Expect(err).ShouldNot(HaveOccurred())

				promSQL := NewSQLExec(usersNamedExec, db)

				res, err := promSQL.ExecContext(context.Background(), `delete from users
					where id = 2 or id = 3`)
				Expect(err).ShouldNot(HaveOccurred())
				rows, err := res.RowsAffected()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(rows).To(BeEquivalentTo(2))

				var metric dto.Metric
				Expect(usersNamedExec.TotalCalls.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalSuccess.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(1))
				Expect(usersNamedExec.TotalFailures.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(0))
				Expect(usersNamedExec.TotalDuration.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeNumerically(">", 0))
				Expect(usersNamedExec.TotalRowsAffected.Write(&metric)).To(Succeed())
				Expect(metric.GetCounter().GetValue()).To(BeEquivalentTo(2))
			})
		})
	})
})
