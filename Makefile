.PHONY: run build test test-coverage integration-test integration-tests lint vuln sast guardrails guardrails-local install-hooks tools tidy generate stop

BINARY  := ./alfredo
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
TOOLS_BIN := $(CURDIR)/.bin
GOLANGCI_LINT_VERSION := v2.11.4
GOVULNCHECK_VERSION := v1.2.0
GOSEC_VERSION := v2.24.7
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
GOVULNCHECK := $(TOOLS_BIN)/govulncheck
GOSEC := $(TOOLS_BIN)/gosec
GUARDRAIL_ARTIFACTS := tmp/guardrails

export GOCACHE := $(CURDIR)/tmp/go-build-cache
export GOLANGCI_LINT_CACHE := $(CURDIR)/tmp/golangci-lint-cache

run: build
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	$(BINARY)

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/server

tools: $(GOLANGCI_LINT) $(GOVULNCHECK) $(GOSEC)

$(GOLANGCI_LINT):
	mkdir -p $(TOOLS_BIN)
	GOBIN=$(TOOLS_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(GOVULNCHECK):
	mkdir -p $(TOOLS_BIN)
	GOBIN=$(TOOLS_BIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

$(GOSEC):
	mkdir -p $(TOOLS_BIN)
	GOBIN=$(TOOLS_BIN) go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)

lint: tools
	$(GOLANGCI_LINT) run ./...

test:
	go test ./internal/...

test-coverage:
	mkdir -p $(GUARDRAIL_ARTIFACTS)
	go test ./internal/... -covermode=atomic -coverprofile=$(GUARDRAIL_ARTIFACTS)/unit-cover.out

integration-test:
	go test -count=1 ./tests/integration/...

integration-tests: integration-test

vuln: tools
	$(GOVULNCHECK) ./...

sast: tools
	$(GOSEC) -severity high ./...

guardrails: tools
	go run ./tools/guardrails \
		-mode ci \
		-golangci-lint $(GOLANGCI_LINT) \
		-govulncheck $(GOVULNCHECK) \
		-gosec $(GOSEC)

guardrails-local: tools
	go run ./tools/guardrails \
		-mode local \
		-html \
		-golangci-lint $(GOLANGCI_LINT) \
		-govulncheck $(GOVULNCHECK) \
		-gosec $(GOSEC)

install-hooks:
	git config core.hooksPath .githooks

tidy:
	go mod tidy

generate:
	go generate ./...
