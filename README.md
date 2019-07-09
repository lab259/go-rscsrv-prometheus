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

### Running tests

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
