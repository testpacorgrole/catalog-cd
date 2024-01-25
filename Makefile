BIN = catalog-cd

GOFLAGS ?= -v
GOFLAGS_TEST ?= -v -cover

ARGS ?=

.EXPORT_ALL_VARIABLES:

all: help

.PHONY: $(BIN)
$(BIN):
	go build -o $(BIN) . $(ARGS)

.PHONY: build
build: $(BIN)

.PHONY: run
run:
	go run . $(ARGS)

install:
	go install $(CMD)

test: test-unit

.PHONY: test-unit
test-unit:
	go test $(GOFLAGS_TEST) ./...

.PHONY: watch
catalog-cd-watch: ## Watch go files and rebuild catalog-cd on changes (needs entr).
	find . -name '*.go' | entr -r go build -v .

.PHONY: help
help:
	@grep -hE '^[ a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
