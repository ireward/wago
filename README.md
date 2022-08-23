<h2 align="center"> A utility library for Go projects within iReward </h2>

---

## About

Wago is a shared library that is used in different Go projects within iReward.

## Prerequisites

In order to use this library, you need to have at least version `1.18` of Go installed on your device, since it makes use of Generics, which were introducted in version `1.18`. To check if you have go and the corresponding version installed, the the following command in your command line:

```bash
go version
```

If Go is installed, you should see the version thats installed.

## Installation

To install `wago` and its dependencies, you have to options:

- If using Go modules, just import the packages that you want to use and afterwards run `go mod tidy`.

```bash
import (
    "github.com/ireward/wago/log"
)
```

- To download the library from GitHub, use the `go get` command

```bash
go get github.com/ireward/wago
```
