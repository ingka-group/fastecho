SHELL := /bin/bash

.PHONY: help
help: # @HELP shows this help
help:
	@echo
	@echo 'Usage:'
	@echo '  make <target>'
	@echo 'Targets:'
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST)    \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    ' | sort

.PHONY: check
check: # @HELP runs pre-commit checks
check:
	@pre-commit run --all

.PHONY: clean
clean: # @HELP removes unnecessary files
clean:
	@rm -rf bin/
	@rm -f coverage.out

.PHONY: compile
compile: # @HELP runs the actual `go build` command which re-compiles the library
compile:
	@go build ./...

.PHONY: test
test: # @HELP runs unit tests and performs coverage evaluation
test: compile  # you don't even want to start the tests if compilation failed
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out
