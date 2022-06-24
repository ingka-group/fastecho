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
	    '

.PHONY: check
check: # @HELP runs pre-commit checks
check:
	@pre-commit run --all

.PHONY: clean
clean: # @HELP removes unnecessary files
clean:
	@rm -rf bin/
	@rm -f coverage.out
