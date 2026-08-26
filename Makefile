# OpsLang Makefile
# Usage: make <target>

MODULE   := github.com/j4ckzh0u/opslang
LDFLAGS  := -s -w

GO       := go
GOBUILD  := $(GO) build
GOTEST   := $(GO) test
GOVET    := $(GO) vet
GOFMT    := gofmt
GOCLEAN  := $(GO) clean

BIN_DIR  := bin
DIST_DIR := dist

# Platforms for cross-compilation
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Binaries
CMDS := opsctl ops-runner

.PHONY: all build test vet fmt lint clean install help
.PHONY: build-all dist coverage bench scale-test tidy check ci
.PHONY: run repl examples

# ─── Default ────────────────────────────────────────────────────────

all: check build test ## Run check, build, and test

# ─── Build ──────────────────────────────────────────────────────────

build: $(addprefix $(BIN_DIR)/,$(CMDS)) ## Build all binaries for current platform

$(BIN_DIR)/opsctl: $(shell find cmd/opsctl internal pkg -name '*.go' 2>/dev/null)
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $@ ./cmd/opsctl

$(BIN_DIR)/ops-runner: $(shell find cmd/ops-runner internal pkg -name '*.go' 2>/dev/null)
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $@ ./cmd/ops-runner

build-all: ## Build for all platforms
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		for cmd in $(CMDS); do \
			echo "Building $$cmd-$$os-$$arch..."; \
			mkdir -p $(DIST_DIR); \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GOBUILD) -ldflags="$(LDFLAGS)" \
				-o $(DIST_DIR)/$$cmd-$$os-$$arch ./cmd/$$cmd; \
		done; \
	done
	@echo "All binaries built in $(DIST_DIR)/"

dist: build-all ## Alias for build-all

install: ## Install opsctl to $GOPATH/bin
	CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(GOPATH)/bin/opsctl ./cmd/opsctl
	@echo "Installed opsctl to $(GOPATH)/bin/opsctl"

# ─── Test ───────────────────────────────────────────────────────────

test: ## Run all tests
	$(GOTEST) -race ./...

coverage: ## Run tests with coverage report
	$(GOTEST) -race -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1
	@echo "Full report: $(GO) tool cover -html=coverage.out"

bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

# OPS_SCALE_N hosts (default 10000), OPS_SCALE_FILE_KB / OPS_SCALE_FAIL_RATE
# / OPS_SCALE_LATENCY_MS tune payload size, fault injection and latency.
scale-test: ## Run 10k-host distribute/collect simulation (full tier)
	OPS_SCALE_N=10000 $(GOTEST) -run 'TestScale' -v -timeout 15m ./pkg/ops-core-sdk/file/

# ─── Quality ────────────────────────────────────────────────────────

vet: ## Run go vet
	$(GOVET) ./...

fmt: ## Check formatting
	@test -z "$$($(GOFMT) -l .)" || { echo "Unformatted files:"; $(GOFMT) -l .; exit 1; }

fmt-fix: ## Fix formatting
	$(GOFMT) -w .

lint: vet fmt ## Run all linters

tidy: ## Tidy go modules
	$(GO) mod tidy

check: lint vet ## Run all checks (lint + vet)

# ─── CI ─────────────────────────────────────────────────────────────

ci: check test build ## Run full CI pipeline locally

# ─── Clean ──────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	$(GOCLEAN)
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out

# ─── Development ────────────────────────────────────────────────────

run: build ## Run opsctl (usage: make run ARGS="run examples/helloworld.ops")
	$(BIN_DIR)/opsctl $(ARGS)

repl: build ## Start OpsLang REPL
	$(BIN_DIR)/opsctl repl

examples: build ## Run all examples
	@for f in examples/*.ops; do \
		echo "=== $$f ==="; \
		timeout 10 $(BIN_DIR)/opsctl run "$$f" 2>&1 | head -5; \
		echo ""; \
	done

# ─── Help ───────────────────────────────────────────────────────────

help: ## Show this help
	@echo "OpsLang Build System"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
