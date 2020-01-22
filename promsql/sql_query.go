package promsql

import (
	"context"
	"database/sql"
	"time"

	"github.com/lab259/go-rscsrv-prometheus/promquery"
)

type PromSQLQuery struct {
	namedQuery *promquery.NamedQueryCollector
	db         DBQueryProxy
}

func NewSQLQuery(namedQuery *promquery.NamedQueryCollector, db DBQueryProxy) *PromSQLQuery {
	return &PromSQLQuery{
		namedQuery: namedQuery,
		db:         db,
	}
}

type DBQueryProxy interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (srv *PromSQLQuery) Query(query string, args ...interface{}) (*sql.Rows, error) {
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

func (srv *PromSQLQuery) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
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

func (srv *PromSQLQuery) Exec(Exec string, args ...interface{}) (sql.Result, error) {
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

func (srv *PromSQLQuery) ExecContext(ctx context.Context, Exec string, args ...interface{}) (sql.Result, error) {
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
