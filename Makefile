BIN = catalog-cd

DOTBIN      = $(CURDIR)/.bin
GOLANGCI_VERSION = v1.58.0

TIMEOUT_UNIT = 20m

GOFLAGS ?= -v
GOFLAGS_TEST ?= -v -cover

SH_FILES := $(shell find ./ -not -regex '^./vendor/.*' -type f -regex ".*\.sh" -print)
YAML_FILES := $(shell find . -not -regex '^./vendor/.*' -type f -regex ".*y[a]ml" -print)
MD_FILES := $(shell find . -type f -regex ".*md"  -not -regex '^./vendor/.*'  -not -regex '^./.vale/.*'  -not -regex "^./docs/themes/.*" -not -regex "^./.git/.*" -not -regex ".*/testdata/.*" -print)

ARGS ?=

.EXPORT_ALL_VARIABLES:

$(DOTBIN):
	@mkdir -p $@

GOLANGCILINT = $(DOTBIN)/golangci-lint
$(DOTBIN)/golangci-lint: $(DOTBIN) ; $(info $(M) getting golangci-lint $(GOLANGCI_VERSION))
	GOBIN=$(DOTBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION)

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

##@ Linting
.PHONY: lint
lint: lint-go lint-yaml lint-md lint-shell ## run all linters

.PHONY: lint-go
lint-go: $(GOLANGCILINT) ## runs go linter on all go files
	@echo "Linting go files..."
	@$(GOLANGCILINT) run ./... --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--timeout $(TIMEOUT_UNIT)

.PHONY: lint-shell
lint-shell: ${SH_FILES} ## runs shellcheck on all python files
	@echo "Linting shell script files..."
	@shellcheck $(SH_FILES)


.PHONY: lint-yaml
lint-yaml: ${YAML_FILES} ## runs yamllint on all yaml files
	@echo "Linting yaml files..."
	@yamllint -c .yamllint $(YAML_FILES)


.PHONY: lint-md
lint-md: ## runs markdownlint and vale on all markdown files
	@echo "Linting markdown files..."
	@markdownlint $(MD_FILES)
	@echo "Grammar check with vale of documentation..."
	@vale docs actions *.md --minAlertLevel=error --output=line
	@echo "CodeSpell on docs content"
	@codespell docs actions

.PHONY: pre-commit
pre-commit: ## Run pre-commit hooks script manually
	@pre-commit run --all-files

.PHONY: help
help:
	@grep -hE '^[ a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'
