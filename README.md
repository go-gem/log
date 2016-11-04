# Gem Log
[![Go Report Card](https://goreportcard.com/badge/github.com/go-gem/log)](https://goreportcard.com/report/github.com/go-gem/log)
[![GoDoc](https://godoc.org/github.com/go-gem/log?status.svg)](https://godoc.org/github.com/go-gem/log)
[![Build Status](https://travis-ci.org/go-gem/log.svg?branch=master)](https://travis-ci.org/go-gem/log)
[![Coverage Status](https://coveralls.io/repos/github/go-gem/log/badge.svg?branch=master)](https://coveralls.io/github/go-gem/log?branch=master)

a simple and leveled logging package written in Go(golang), it is an extended edition of the standard logging package.

## Install
```
go get github.com/go-gem/log
```
Requires Go 1.5 or above.

## Example
```
package main

import (
	"github.com/go-gem/log"
	"os"
)

func main() {
	var logger log.Logger

	logger = log.New(os.Stderr, log.Lshortfile, log.LevelWarning | log.LevelError)

	logger.Debug("debug log.") // ignored.
	logger.Info("info log.") // ignored.
	logger.Warning("warning log.")
	logger.Error("error log.")
}
```