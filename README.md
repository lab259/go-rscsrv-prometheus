# go-package-boilerplate [![CircleCI](https://circleci.com/gh/lab259/go-package-boilerplate.svg?style=shield&circle-token=224f68e222b4a6abeb01f2d0dda3b4cf264b806e)](https://circleci.com/gh/lab259/go-package-boilerplate)

> See here [how to create a repository from a template](https://help.github.com/en/articles/creating-a-repository-from-a-template).

## Getting Started

### Prerequisites

What things you need to setup the project:

- [go](https://golang.org/doc/install)
- [ginkgo](http://onsi.github.io/ginkgo/)

### Environment

Close the repository:

```bash
git clone git@github.com:lab259/go-package-boilerplate.git
```

Now, the dependencies must be installed.

```
cd go-package-boilerplate && go mod download
```

:wink: Finally, you are done to start developing.

### Running tests

In the `src/github.com/lab259/go-package-boilerplate` directory, execute:

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
