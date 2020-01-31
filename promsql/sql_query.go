package promsql

import (
	"context"
	"database/sql"
	"time"
)

// Query will serve as `sql.DB` and `sql.Tx` proxy. So, it can be
// intialized and used as replacement for db or tx calls.
type Query struct {
	namedQuery *NamedQuery
	db         DBQueryProxy
}

// NewQuery creates a new instance of *Query.
func NewQuery(namedQuery *NamedQuery, db DBQueryProxy) *Query {
	return &Query{
		namedQuery: namedQuery,
		db:         db,
	}
}

// DBQueryProxy is an abstraction of the operations needed from the `sql.DB`.
// Additionally, this approach enables using `sql.Tx` using the same strategy.
type DBQueryProxy interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Query is a proxy to `sql.DB.Query` that add some logic to count the number of
// times the method was called, how many times it succeeded or failed and how
// long it took.
func (srv *Query) Query(query string, args ...interface{}) (*sql.Rows, error) {
	srv.namedQuery.TotalCalls.Inc()

	start := time.Now()
	res, err := srv.db.Query(query, args...)
	srv.namedQuery.TotalDuration.Add(time.Since(start).Seconds())

	if err != nil {
		srv.namedQuery.TotalFailures.Inc()
	} else {
		srv.namedQuery.TotalSuccess.Inc()
	}

	return res, err
}

// QueryContext is a proxy to `sql.DB.QueryContext` that add some logic to count
// the number of times the method was called, how many times it succeeded or
// failed and how long it took.
func (srv *Query) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	srv.namedQuery.TotalCalls.Inc()

	start := time.Now()
	res, err := srv.db.QueryContext(ctx, query, args...)
	srv.namedQuery.TotalDuration.Add(time.Since(start).Seconds())

	if err != nil {
		srv.namedQuery.TotalFailures.Inc()
	} else {
		srv.namedQuery.TotalSuccess.Inc()
	}

	return res, err
}

// Exec is a proxy to `sql.DB.Exec` that add some logic to count the number of
// times the method was called, how many times it succeeded or failed and how
// long it took.
func (srv *Query) Exec(Exec string, args ...interface{}) (sql.Result, error) {
	srv.namedQuery.TotalCalls.Inc()

	start := time.Now()
	res, err := srv.db.Exec(Exec, args...)
	srv.namedQuery.TotalDuration.Add(time.Since(start).Seconds())

	if err != nil {
		srv.namedQuery.TotalFailures.Inc()
	} else {
		rowsAffected, rowErr := res.RowsAffected()

		if rowErr != nil {
			srv.namedQuery.TotalFailures.Inc()
		} else {
			srv.namedQuery.TotalSuccess.Inc()
			srv.namedQuery.TotalRowsAffected.Add(float64(rowsAffected))
		}
	}

	return res, err
}

// ExecContext is a proxy to `sql.DB.ExecContext` that add some logic to count
// the number of times the method was called, how many times it succeeded or
// failed and how long it took.
func (srv *Query) ExecContext(ctx context.Context, Exec string, args ...interface{}) (sql.Result, error) {
	srv.namedQuery.TotalCalls.Inc()

	start := time.Now()
	res, err := srv.db.ExecContext(ctx, Exec, args...)
	srv.namedQuery.TotalDuration.Add(time.Since(start).Seconds())

	if err != nil {
		srv.namedQuery.TotalFailures.Inc()
	} else {
		rowsAffected, rowErr := res.RowsAffected()

		if rowErr != nil {
			srv.namedQuery.TotalFailures.Inc()
		} else {
			srv.namedQuery.TotalSuccess.Inc()
			srv.namedQuery.TotalRowsAffected.Add(float64(rowsAffected))
		}
	}

	return res, err
}
