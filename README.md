# go-rscsrv-prometheus [![CircleCI](https://circleci.com/gh/lab259/go-rscsrv-prometheus.svg?style=svg&circle-token=870af825230a3bc9c94a153dad99b49cbebd696f)](https://circleci.com/gh/lab259/go-rscsrv-prometheus)

## Getting Started

### Prerequisites

What things you need to setup the project:

- [go](https://golang.org/doc/install)
- [ginkgo](http://onsi.github.io/ginkgo/)

### Environment

Close the repository:

```bash
git clone git@github.com:lab259/go-rscsrv-prometheus.git
```

Now, the dependencies must be installed.

```
cd go-rscsrv-prometheus && go mod download
```

:wink: Finally, you are done to start developing.

### Collectors

**database/sql.DB**

Given a `*sql.DB` instance, you can create a collector by using `promsql.NewDatabaseCollector(db, opts)`.

The information provided by the collector will generate these metrics:

- db_max_open_connections: Maximum number of open connections to the database.
- db_pool_open_connections: The number of established connections both in use and idle.
- db_pool_in_use: The number of connections currently in use.
- db_pool_idle: The number of idle connections.
- db_wait_count: The total number of connections waited for.
- db_wait_duration: The total time blocked waiting for a new connection.
- db_max_idle_closed: The total number of connections closed due to SetMaxIdleConns.
- db_max_lifetime_closed: The total number of connections closed due to SetConnMaxLifetime.

_All these metrics are provided by the `database/sql` package interface._

**opts: _promsql.DatabaseCollectorOpts_**
- Prefix `string`: That will add a prefix to the metrics names. So, for example, `db_max_open_connections` will become `db_PREFIX_max_open_connections`.


**database/sql/driver**

Given a driver name (e.g. `postgres`), you can create a collector by using `promsql.Register(opts)`.

The information provided by the collector will generate these metrics:

- db_query_total: The total number of queries processed.
- db_query_successful: The number of queries processed with success.
- db_query_failed: The number of queries processed with failure.
- db_transaction_total: The total number of transactions processed.
- db_transaction_successful: The number of transactions processed with success.
- db_transaction_failed: The number of transactions processed with failure.
- db_execution_total: The total number of executions processed.
- db_execution_successful: The number of executions processed with success.
- db_execution_failed: The number of executions processed with failure.

**opts: _promsql.DriverCollectorOpts**
- DriverName `string`: The base driver name that will be used by sql package (e.g. `postgres`, `mysql`)
- Prefix `string`: That will add a prefix to the metrics names. So, for example, `db_query_total` will become `db_PREFIX_query_total`.


### Usage

#### DriverCollector

First you have to start a PostgreSQL service

```
docker-compose up
```

Then you can run: 

```go
// Register our driver collector wrapper
driverCollector, err := promsql.Register(promsql.DriverCollectorOpts{
    DriverName: "postgres",
})
if err != nil {
    log.Fatalf("unable to register the driver collector: %v\n", err)
}

// Open sql connection with our driver name wrapper 
psqlInfo := fmt.Sprintf("user=postgres password=postgres dbname=db-example sslmode=disable")
db, err := sql.Open(driverCollector.DriverName, psqlInfo)

// Execute some query
rs, err := db.Query("select name from users where id = 1")
defer rs.Close()

// Retrieve the total number of queries processed
var metric dto.Metric
driverCollector.QueryTotalCounter.Write(&metric)
fmt.Println(metric.GetCounter().GetValue()) // 1
```

### Running tests

In order to run the tests, spin up the :

```bash
make dco-test-up
```

In the `src/github.com/lab259/go-rscsrv-prometheus` directory, execute:

```bash
make test
```

To enable coverage, execute:

```bash
make coverage
```

To generate the HTML coverage report, execute:

```bash
make coverage coverage-html
```
