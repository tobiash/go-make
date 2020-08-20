# go-make : make-like build pipelines for go

## Overview [![PkgGoDev](https://pkg.go.dev/badge/tobiash/go-make)](https://pkg.go.dev/tobiash/go-make) [![Code Climate](https://codeclimate.com/github/tobiash/go-make/badges/gpa.svg)](https://codeclimate.com/github/tobiash/go-make) [![Go Report Card](https://goreportcard.com/badge/github.com/tobiash/go-make)](https://goreportcard.com/report/github.com/tobiash/go-make)

Experimental implementation of a (GNU-) `make`-like build pipeline for Go. The goal here is not to recreate or replace `make`, but to leverage its patterns and semantics as a library. Therefore, the provided CLI command is more of a proof-of-concept.

Key features:

- a simple DAG (directed acyclic graph) implementation with go-routine-based parallel walk feature
- an abstraction for `make`'s targets - this is not limited to filesystem based targets
- an abstraction of up-to-date checks for targets - while GNU make relies on file change timestamps only,
  go-make can be extended and by default uses a hash based system built on go-mod's `go.sum`

## Install

```
go get github.com/tobiash/go-make
```
