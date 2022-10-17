GOBIN := $(or $(GOBIN),$(shell pwd)/bin)

COVERAGE_REPORT := coverage.out

.PHONY: help
help: # @HELP shows this help
help:
	@echo
	@echo 'Variables:'
	@echo '  COVERAGE_REPORT = $(COVERAGE_REPORT)'
	@echo
	@echo 'Usage:'
	@echo '  make <target>'
	@echo 'Targets:'
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST)    \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    '

.PHONY: check
check: # @HELP runs pre-commit checks
check:
	@pre-commit run --all

.PHONY: clean
clean: # @HELP removes unnecessary files
clean:
	@rm -rf bin/
	@rm -f ${COVERAGE_REPORT}

.PHONY: compile
compile: # @HELP runs the actual `go build` command which re-compiles the library
compile:
	@go build ./...

.PHONY: test
test: # @HELP runs unit tests
test:
	@echo -n "Running unit tests..."
	GOBIN=$(GOBIN) go test -v ./...

.PHONY: coverage
coverage: # @HELP generates an HTML coverage report
coverage: coverage-report
	GOBIN=$(GOBIN) go tool cover -func=${COVERAGE_REPORT} | tail -n 1
	GOBIN=$(GOBIN) go tool cover -html=${COVERAGE_REPORT}

.PHONY: coverage-report
coverage-report: # @HELP runs all tests and generates a RAW coverage report to be picked up by analysis tools
coverage-report:
	GOBIN=$(GOBIN) go test -coverpkg ./date/...,./echozap/...,./gcstorage/...,./rest/...,./util/... -coverprofile=${COVERAGE_REPORT} ./...
