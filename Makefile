SHELL := /bin/bash
export VERSION := 0.1.0
export GOBIN = $(shell pwd)/bin

include scripts/Makefile.help
.DEFAULT_GOAL := help

include build/Makefile.deps

## Formats the Go source files
format:
	@ $(GOBIN)/golangci-lint run --fix
	@ $(GOBIN)/go-licenser -license ASL2

## Runs the linters
lint:
	@ $(GOBIN)/golangci-lint run
	@ $(GOBIN)/go-licenser -d -license ASL2

## Executes the unit tests
unit:
	@ go test -cover -race ./...

## Runs the integration tests
integration:
	@ go test -v -tags integration github.com/elastic/testcli/build/integration
