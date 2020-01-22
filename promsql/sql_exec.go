package promsql

import (
	"context"
	"database/sql"
	"time"

	"github.com/lab259/go-rscsrv-prometheus/promexec"
)

type PromSQLExec struct {
	namedExec *promexec.NamedExecCollector
	db        DBExecProxy
}

func NewSQLExec(namedExec *promexec.NamedExecCollector, db DBExecProxy) *PromSQLExec {
	return &PromSQLExec{
		namedExec: namedExec,
		db:        db,
	}
}

type DBExecProxy interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (srv *PromSQLExec) Exec(Exec string, args ...interface{}) (sql.Result, error) {
	srv.namedExec.TotalCalls.Inc()

	start := time.Now()
	res, err := srv.db.Exec(Exec, args...)
	srv.namedExec.TotalDuration.Add(time.Since(start).Seconds())

	if err != nil {
		srv.namedExec.TotalFailures.Inc()
	} else {
		rowsAffected, err := res.RowsAffected()

		if err != nil {
			srv.namedExec.TotalFailures.Inc()
		} else {
			srv.namedExec.TotalSuccess.Inc()
			srv.namedExec.TotalRowsAffected.Add(float64(rowsAffected))
		}
	}

	return res, err
}

func (srv *PromSQLExec) ExecContext(ctx context.Context, Exec string, args ...interface{}) (sql.Result, error) {
	srv.namedExec.TotalCalls.Inc()

	start := time.Now()
	res, err := srv.db.ExecContext(ctx, Exec, args...)
	srv.namedExec.TotalDuration.Add(time.Since(start).Seconds())

	if err != nil {
		srv.namedExec.TotalFailures.Inc()
	} else {
		rowsAffected, err := res.RowsAffected()

		if err != nil {
			srv.namedExec.TotalFailures.Inc()
		} else {
			srv.namedExec.TotalSuccess.Inc()
			srv.namedExec.TotalRowsAffected.Add(float64(rowsAffected))
		}
	}

	return res, err
}
