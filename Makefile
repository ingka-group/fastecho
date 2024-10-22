SHELL := /bin/bash

EXECUTABLE := fastecho
GOBIN := $(or $(GOBIN),$(shell pwd)/bin)
BUILD_PATH := $(GOBIN)/$(EXECUTABLE)

# Code coverage
COVERAGE_REPORT := coverage.out

# Dependency versions
GO_IMPORT_LINT_VERSION ?= latest

.PHONY: help
help: ## @HELP prints this message
help:
	@echo
	@echo 'Variables:'
	@echo '  EXECUTABLE      = $(EXECUTABLE)'
	@echo '  BUILD_PATH      = $(BUILD_PATH)'
	@echo '  COVERAGE_REPORT = $(COVERAGE_REPORT)'
	@echo
	@echo 'Versions:'
	@echo '  GO_IMPORT_LINT_VERSION = $(GO_IMPORT_LINT_VERSION)'
	@echo
	@echo 'Usage:'
	@echo '  make <target>'
	@echo
	@echo 'Targets:'
	@grep -E '^.*: *## *@HELP' $(MAKEFILE_LIST)   \
	    | awk '                                   \
	        BEGIN {FS = ": *## *@HELP"};          \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    ' | sort

.PHONY: all
all: ## @HELP builds binaries, runs tests and generates documentation
all: clean compile test

.PHONY: clean
clean: ## @HELP clean-up
clean:
	@rm -rf bin/
	@rm -f $(COVERAGE_REPORT)
	@GOBIN=$(GOBIN) go clean -r --modcache --testcache ./...

.PHONY: pre-commit
pre-commit: ## @HELP runs pre-commit checks on the entire repo
pre-commit:
	@pre-commit run --all-files

.PHONY: compile
compile: ## @HELP runs the actual `go build` which updates the service binary
compile:
	@echo -n "Building $(EXECUTABLE) into $(GOBIN) ... "
	@GOBIN=$(GOBIN) CGO_ENABLED=0 go build -o $(BUILD_PATH) cmd/main.go
	@echo "done."

.PHONY: run
run: ## @HELP runs the service
run: compile
	@$(BUILD_PATH)

.PHONY: test
test: ## @HELP runs all tests and generates a RAW coverage report to be picked up by analysis tools
test:
	@echo -n "Running full tests (unit and integration) ... "
	@GOBIN=$(GOBIN) go test -coverpkg $(COVERAGE_PKG) -coverprofile=$(COVERAGE_REPORT) ./...
	@GOBIN=$(GOBIN) go tool cover -func=$(COVERAGE_REPORT) | tail -n 1
	@echo "done."

.PHONY: test-short
test-short: ## @HELP runs unit tests only (no integration tests)
test-short:
	@echo -n "Running short tests (no integration tests) ... "
	@GOBIN=$(GOBIN) go test -v -short ./...
	@echo "done."

.PHONY: coverage-html
coverage-html: ## @HELP generates an HTML coverage report
coverage-html: test
	@GOBIN=$(GOBIN) go tool cover -html=$(COVERAGE_REPORT)

.PHONY: go-import-lint
go-import-lint: ## @HELP verifies the imports order
go-import-lint: bin/go-import-lint check-go-import-lint
	@echo "Verifying import order"
	@$(GOBIN)/go-import-lint

.PHONY: check-go-import-lint
check-go-import-lint:
	@which $(GOBIN)/go-import-lint >/dev/null || ( echo "Install go-import-lint from https://github.com/hedhyw/go-import-lint and retry." && exit 1 )

.PHONY: check-gh
check-gh:
	@gh --version >/dev/null || ( echo "Install gh CLI (e.g. brew install gh), run `gh auth login` retry." && exit 1 )

.PHONY: bin/go-import-lint
bin/go-import-lint:
	@echo "Downloading go-import-lint..."
	@GOBIN=$(GOBIN) go install -v github.com/hedhyw/go-import-lint/cmd/go-import-lint@$(GO_IMPORT_LINT_VERSION)
