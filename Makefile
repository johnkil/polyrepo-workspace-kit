GO ?= go
GO_BIN ?= $(shell $(GO) env GOPATH)/bin
GOIMPORTS ?= $(GO_BIN)/goimports
GOLANGCI_LINT ?= $(GO_BIN)/golangci-lint
GOVULNCHECK ?= $(GO_BIN)/govulncheck
GORELEASER ?= $(GO_BIN)/goreleaser

GOIMPORTS_VERSION ?= v0.29.0
GOLANGCI_LINT_VERSION ?= $(shell cat .golangci-lint-version)
GOVULNCHECK_VERSION ?= v1.2.0
GORELEASER_VERSION ?= v2.15.3
GOIMPORTS_LOCAL ?= github.com/johnkil/polyrepo-workspace-kit

BIN_DIR ?= bin
WKIT_BIN ?= $(BIN_DIR)/wkit
COVERAGE_FILE ?= coverage.out
FUZZTIME ?= 5s

GO_FILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './vendor/*')

.PHONY: tools release-tools fmt fmt-check tidy-check vet lint test test-race coverage fuzz vuln build demo failure-demo install-script-check check release-check release-snapshot clean

tools:
	$(GO) install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

release-tools:
	$(GO) install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

fmt:
	$(GOIMPORTS) -local $(GOIMPORTS_LOCAL) -w $(GO_FILES)

fmt-check:
	@unformatted="$$( $(GOIMPORTS) -local $(GOIMPORTS_LOCAL) -l $(GO_FILES) )"; \
	if [ -n "$$unformatted" ]; then \
		echo "goimports needed for:"; \
		echo "$$unformatted"; \
		echo "Run: make fmt"; \
		exit 1; \
	fi

tidy-check:
	$(GO) mod tidy -diff

vet:
	$(GO) vet ./...

lint:
	$(GOLANGCI_LINT) run

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

coverage:
	$(GO) test -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

fuzz:
	@set -eu; \
	for package in $$($(GO) list ./...); do \
		if $(GO) test -list '^Fuzz' $$package | grep -q '^Fuzz'; then \
			echo "fuzz $$package"; \
			$(GO) test -run='^$$' -fuzz='Fuzz' -fuzztime=$(FUZZTIME) $$package; \
		fi; \
	done

vuln:
	$(GOVULNCHECK) ./...

build:
	$(GO) build -o $(WKIT_BIN) ./cmd/wkit

demo: build
	WKIT_BIN="$(PWD)/$(WKIT_BIN)" sh examples/minimal-workspace/run-demo.sh

failure-demo: build
	WKIT_BIN="$(PWD)/$(WKIT_BIN)" sh examples/failure-workspace/run-demo.sh

install-script-check:
	sh -n install.sh
	sh install.sh --help >/dev/null

check: tidy-check fmt-check vet lint test test-race vuln demo failure-demo install-script-check

release-check:
	$(GORELEASER) check

release-snapshot:
	$(GORELEASER) release --snapshot --clean --release-notes docs/release-notes.md

clean:
	rm -f $(WKIT_BIN) $(COVERAGE_FILE)
	rm -rf dist
