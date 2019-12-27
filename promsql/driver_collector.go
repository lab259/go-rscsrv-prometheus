package promsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Implementation based on https://github.com/opencensus-integrations/ocsql

type conn interface {
	driver.Pinger
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.Conn
	driver.ConnPrepareContext
	driver.ConnBeginTx
}

var (
	regMu sync.Mutex
)

func Register(options DriverCollectorOpts) (*DriverCollector, error) {
	// retrieve the driver implementation we need to wrap with instrumentation
	db, err := sql.Open(options.DriverName, "")
	if err != nil {
		return nil, err
	}
	dri := db.Driver()
	if err = db.Close(); err != nil {
		return nil, err
	}

	regMu.Lock()
	defer regMu.Unlock()

	driverName := options.DriverName + "-collector-"
	for i := int64(0); i < 1000; i++ {
		var (
			found   = false
			regName = driverName + strconv.FormatInt(i, 10)
		)
		for _, name := range sql.Drivers() {
			if name == regName {
				found = true
			}
		}
		if !found {
			driverCollector := Wrap(dri, options)
			sql.Register(regName, driverCollector)
			driverCollector.DriverName = regName
			return driverCollector, nil
		}
	}
	return nil, errors.New("unable to register driver, all slots have been taken")
}

func Wrap(d driver.Driver, options DriverCollectorOpts) *DriverCollector {
	return wrapDriver(d, options)
}

func wrapDriver(d driver.Driver, o DriverCollectorOpts) *DriverCollector {
	return NewDriverCollector(d, o)
}

type DriverCollector struct {
	DriverName string
	parent     driver.Driver
	connector  driver.Connector

	// prometheus counters
	QueryTotalCounter                  prometheus.Counter
	QuerySuccessfulCounter             prometheus.Counter
	QueryErroneousCounter              prometheus.Counter
	TransactionTotalCounter            prometheus.Counter
	TransactionSuccessfulAmountCounter prometheus.Counter
	TransactionErroneousAmountCounter  prometheus.Counter
	ExecutionTotalCounter              prometheus.Counter
	ExecutionSuccessfulAmountCounter   prometheus.Counter
	ExecutionErroneousAmountCounter    prometheus.Counter
}

type DriverCollectorOpts struct {
	DriverName string
	Prefix     string
}

func NewDriverCollector(driver driver.Driver, opts DriverCollectorOpts) *DriverCollector {
	prefix := opts.Prefix
	if prefix != "" && !strings.HasSuffix(opts.Prefix, "_") {
		prefix += "_"
	}

	return &DriverCollector{
		parent:                             driver,
		DriverName:                         opts.DriverName,
		QueryTotalCounter:                  prometheus.NewCounter(prometheus.CounterOpts{}),
		QuerySuccessfulCounter:             prometheus.NewCounter(prometheus.CounterOpts{}),
		QueryErroneousCounter:              prometheus.NewCounter(prometheus.CounterOpts{}),
		TransactionTotalCounter:            prometheus.NewCounter(prometheus.CounterOpts{}),
		TransactionSuccessfulAmountCounter: prometheus.NewCounter(prometheus.CounterOpts{}),
		TransactionErroneousAmountCounter:  prometheus.NewCounter(prometheus.CounterOpts{}),
		ExecutionTotalCounter:              prometheus.NewCounter(prometheus.CounterOpts{}),
		ExecutionSuccessfulAmountCounter:   prometheus.NewCounter(prometheus.CounterOpts{}),
		ExecutionErroneousAmountCounter:    prometheus.NewCounter(prometheus.CounterOpts{}),
	}
}

func (collector *DriverCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- collector.QueryTotalCounter.Desc()
	descs <- collector.QuerySuccessfulCounter.Desc()
	descs <- collector.QueryErroneousCounter.Desc()
	descs <- collector.TransactionTotalCounter.Desc()
	descs <- collector.TransactionSuccessfulAmountCounter.Desc()
	descs <- collector.TransactionErroneousAmountCounter.Desc()
	descs <- collector.ExecutionTotalCounter.Desc()
	descs <- collector.ExecutionSuccessfulAmountCounter.Desc()
	descs <- collector.ExecutionErroneousAmountCounter.Desc()
}

func (collector *DriverCollector) Collect(metrics chan<- prometheus.Metric) {
	collector.QueryTotalCounter.Collect(metrics)
	collector.QuerySuccessfulCounter.Collect(metrics)
	collector.QueryErroneousCounter.Collect(metrics)
	collector.TransactionTotalCounter.Collect(metrics)
	collector.TransactionSuccessfulAmountCounter.Collect(metrics)
	collector.TransactionErroneousAmountCounter.Collect(metrics)
	collector.ExecutionTotalCounter.Collect(metrics)
	collector.ExecutionSuccessfulAmountCounter.Collect(metrics)
	collector.ExecutionErroneousAmountCounter.Collect(metrics)
}

func (d *DriverCollector) Open(name string) (driver.Conn, error) {
	c, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}
	return wrapConn(c, d), nil
}

func wrapConn(parent driver.Conn, collector *DriverCollector) driver.Conn {
	var (
		n, hasNameValueChecker = parent.(driver.NamedValueChecker)
		s, hasSessionResetter  = parent.(driver.SessionResetter)
	)
	c := &ocConn{parent: parent, collector: collector}
	switch {
	case !hasNameValueChecker && !hasSessionResetter:
		return c
	case hasNameValueChecker && !hasSessionResetter:
		return struct {
			conn
			driver.NamedValueChecker
		}{c, n}
	case !hasNameValueChecker && hasSessionResetter:
		return struct {
			conn
			driver.SessionResetter
		}{c, s}
	case hasNameValueChecker && hasSessionResetter:
		return struct {
			conn
			driver.NamedValueChecker
			driver.SessionResetter
		}{c, n, s}
	}
	panic("unreachable")
}

// ocConn implements driver.Conn
type ocConn struct {
	parent    driver.Conn
	collector *DriverCollector
}

func (c *ocConn) Ping(ctx context.Context) (err error) {
	if pinger, ok := c.parent.(driver.Pinger); ok {
		err = pinger.Ping(ctx)
	}
	return
}

func (c *ocConn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	if exec, ok := c.parent.(driver.Execer); ok {
		c.collector.ExecutionTotalCounter.Inc()

		if res, err = exec.Exec(query, args); err != nil {
			c.collector.ExecutionErroneousAmountCounter.Inc()
			return nil, err
		}

		c.collector.ExecutionSuccessfulAmountCounter.Inc()
		return ocResult{parent: res}, nil
	}

	return nil, driver.ErrSkip
}

func (c *ocConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (res driver.Result, err error) {
	if execCtx, ok := c.parent.(driver.ExecerContext); ok {
		c.collector.ExecutionTotalCounter.Inc()

		if res, err = execCtx.ExecContext(ctx, query, args); err != nil {
			c.collector.ExecutionErroneousAmountCounter.Inc()
			return nil, err
		}

		c.collector.ExecutionSuccessfulAmountCounter.Inc()
		return res, nil
	}

	return nil, driver.ErrSkip
}

func (c *ocConn) Query(query string, args []driver.Value) (rows driver.Rows, err error) {
	if queryer, ok := c.parent.(driver.Queryer); ok {
		c.collector.QueryTotalCounter.Inc()

		rows, err = queryer.Query(query, args)
		if err != nil {

			c.collector.QueryErroneousCounter.Inc()
			return nil, err
		}

		c.collector.QuerySuccessfulCounter.Inc()
		return rows, nil
	}

	return nil, driver.ErrSkip
}

func (c *ocConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	if queryerCtx, ok := c.parent.(driver.QueryerContext); ok {

		c.collector.QueryTotalCounter.Inc()
		rows, err = queryerCtx.QueryContext(ctx, query, args)
		if err != nil {

			c.collector.QueryErroneousCounter.Inc()
			return nil, err
		}

		c.collector.QuerySuccessfulCounter.Inc()
		return rows, nil
	}

	return nil, driver.ErrSkip
}

func (c *ocConn) Prepare(query string) (stmt driver.Stmt, err error) {
	stmt, err = c.parent.Prepare(query)
	if err != nil {
		return nil, err
	}

	stmt = wrapStmt(stmt, query, c.collector)
	return
}

func (c *ocConn) Close() error {
	return c.parent.Close()
}

func (c *ocConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.TODO(), driver.TxOptions{})
}

func (c *ocConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	if prepCtx, ok := c.parent.(driver.ConnPrepareContext); ok {
		stmt, err = prepCtx.PrepareContext(ctx, query)
	} else {
		stmt, err = c.parent.Prepare(query)
	}
	if err != nil {
		return nil, err
	}

	stmt = wrapStmt(stmt, query, c.collector)
	return
}

func (c *ocConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	if ctx == nil || ctx == context.TODO() {
		ctx = context.Background()
	}
	if connBeginTx, ok := c.parent.(driver.ConnBeginTx); ok {
		tx, err = connBeginTx.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return ocTx{parent: tx, ctx: ctx, collector: c.collector}, nil
	}

	tx, err = c.parent.Begin()
	if err != nil {
		return nil, err
	}
	return ocTx{parent: tx, ctx: ctx, collector: c.collector}, nil
}

// ocResult implements driver.Result
type ocResult struct {
	parent driver.Result
	ctx    context.Context
}

func (r ocResult) LastInsertId() (int64, error) {
	return r.parent.LastInsertId()
}

func (r ocResult) RowsAffected() (int64, error) {
	return r.parent.RowsAffected()
}

// ocStmt implements driver.Stmt
type ocStmt struct {
	parent    driver.Stmt
	query     string
	collector *DriverCollector
}

func (s ocStmt) Exec(args []driver.Value) (res driver.Result, err error) {
	s.collector.ExecutionTotalCounter.Inc()

	res, err = s.parent.Exec(args)
	if err != nil {
		s.collector.ExecutionErroneousAmountCounter.Inc()
		return nil, err
	}

	s.collector.ExecutionSuccessfulAmountCounter.Inc()

	res, err = ocResult{parent: res}, nil
	return
}

func (s ocStmt) Close() error {
	return s.parent.Close()
}

func (s ocStmt) NumInput() int {
	return s.parent.NumInput()
}

func (s ocStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	s.collector.QueryTotalCounter.Inc()

	rows, err = s.parent.Query(args)
	if err != nil {

		s.collector.QueryErroneousCounter.Inc()
		return nil, err
	}

	s.collector.QuerySuccessfulCounter.Inc()
	return rows, nil
}

func (s ocStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	s.collector.QueryTotalCounter.Inc()

	queryContext := s.parent.(driver.StmtQueryContext)
	rows, err = queryContext.QueryContext(ctx, args)
	if err != nil {

		s.collector.QueryErroneousCounter.Inc()
		return nil, err

	}

	s.collector.QuerySuccessfulCounter.Inc()
	rows, err = wrapRows(ctx, rows), nil
	return
}

func (s ocStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	s.collector.ExecutionTotalCounter.Inc()

	execContext := s.parent.(driver.StmtExecContext)
	res, err = execContext.ExecContext(ctx, args)
	if err != nil {

		s.collector.ExecutionErroneousAmountCounter.Inc()
		return nil, err
	}

	s.collector.ExecutionSuccessfulAmountCounter.Inc()
	res, err = ocResult{parent: res, ctx: ctx}, nil
	return
}

type withRowsColumnTypeScanType interface {
	ColumnTypeScanType(index int) reflect.Type
}

type ocRows struct {
	parent driver.Rows
	ctx    context.Context
}

func (r ocRows) HasNextResultSet() bool {
	if v, ok := r.parent.(driver.RowsNextResultSet); ok {
		return v.HasNextResultSet()
	}

	return false
}

func (r ocRows) NextResultSet() error {
	if v, ok := r.parent.(driver.RowsNextResultSet); ok {
		return v.NextResultSet()
	}

	return io.EOF
}

func (r ocRows) ColumnTypeDatabaseTypeName(index int) string {
	if v, ok := r.parent.(driver.RowsColumnTypeDatabaseTypeName); ok {
		return v.ColumnTypeDatabaseTypeName(index)
	}

	return ""
}

func (r ocRows) ColumnTypeLength(index int) (length int64, ok bool) {
	if v, ok := r.parent.(driver.RowsColumnTypeLength); ok {
		return v.ColumnTypeLength(index)
	}

	return 0, false
}

func (r ocRows) ColumnTypeNullable(index int) (nullable, ok bool) {
	if v, ok := r.parent.(driver.RowsColumnTypeNullable); ok {
		return v.ColumnTypeNullable(index)
	}

	return false, false
}

func (r ocRows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	if v, ok := r.parent.(driver.RowsColumnTypePrecisionScale); ok {
		return v.ColumnTypePrecisionScale(index)
	}

	return 0, 0, false
}

func (r ocRows) Columns() []string {
	return r.parent.Columns()
}

func (r ocRows) Close() error {
	return r.parent.Close()
}

func (r ocRows) Next(dest []driver.Value) error {
	return r.parent.Next(dest)
}

func wrapRows(ctx context.Context, parent driver.Rows) driver.Rows {
	var (
		ts, hasColumnTypeScan = parent.(driver.RowsColumnTypeScanType)
	)

	r := ocRows{
		parent: parent,
		ctx:    ctx,
	}

	if hasColumnTypeScan {
		return struct {
			ocRows
			withRowsColumnTypeScanType
		}{r, ts}
	}

	return r
}

type ocTx struct {
	parent    driver.Tx
	ctx       context.Context
	collector *DriverCollector
}

func (t ocTx) Commit() (err error) {
	t.collector.TransactionTotalCounter.Inc()

	err = t.parent.Commit()
	if err != nil {
		t.collector.TransactionErroneousAmountCounter.Inc()
	} else {
		t.collector.TransactionSuccessfulAmountCounter.Inc()
	}
	return
}

func (t ocTx) Rollback() (err error) {
	return t.parent.Rollback()
}

func wrapStmt(stmt driver.Stmt, query string, collector *DriverCollector) driver.Stmt {
	var (
		_, hasExeCtx    = stmt.(driver.StmtExecContext)
		_, hasQryCtx    = stmt.(driver.StmtQueryContext)
		c, hasColConv   = stmt.(driver.ColumnConverter)
		n, hasNamValChk = stmt.(driver.NamedValueChecker)
	)

	s := ocStmt{parent: stmt, query: query, collector: collector}
	switch {
	case !hasExeCtx && !hasQryCtx && !hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
		}{s}
	case !hasExeCtx && hasQryCtx && !hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
		}{s, s}
	case hasExeCtx && !hasQryCtx && !hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
		}{s, s}
	case hasExeCtx && hasQryCtx && !hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
		}{s, s, s}
	case !hasExeCtx && !hasQryCtx && hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.ColumnConverter
		}{s, c}
	case !hasExeCtx && hasQryCtx && hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.ColumnConverter
		}{s, s, c}
	case hasExeCtx && !hasQryCtx && hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.ColumnConverter
		}{s, s, c}
	case hasExeCtx && hasQryCtx && hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.ColumnConverter
		}{s, s, s, c}

	case !hasExeCtx && !hasQryCtx && !hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.NamedValueChecker
		}{s, n}
	case !hasExeCtx && hasQryCtx && !hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.NamedValueChecker
		}{s, s, n}
	case hasExeCtx && !hasQryCtx && !hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.NamedValueChecker
		}{s, s, n}
	case hasExeCtx && hasQryCtx && !hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.NamedValueChecker
		}{s, s, s, n}
	case !hasExeCtx && !hasQryCtx && hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, c, n}
	case !hasExeCtx && hasQryCtx && hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, c, n}
	case hasExeCtx && !hasQryCtx && hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, c, n}
	case hasExeCtx && hasQryCtx && hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, s, c, n}
	}
	panic("unreachable")
}

var errConnDone = sql.ErrConnDone
