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
	GOBIN=$(GOBIN) go test -cover -coverprofile=${COVERAGE_REPORT} ./...

.PHONY: go-import-lint
go-import-lint: # @HELP verifies the imports order
go-import-lint: bin/go-import-lint
	@echo -n "Verifying import order"
	PATH="$(PATH):$(GOBIN)" go-import-lint

bin/go-import-lint:
	@echo "Downloading go-import-lint..."
	GOBIN=$(GOBIN) go install -v github.com/hedhyw/go-import-lint/cmd/go-import-lint@latest
	@echo "Verifying go-import-lint installation..."
	${GOBIN}/go-import-lint --help

.PHONY: clean-tags
clean-tags: # @HELP cleans local git tags
clean-tags:
	git tag -l | xargs git tag -d
	git fetch --tags
