.PHONY: build install test clean lint fmt run repl

# 版本号
VERSION := 0.1.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X github.com/opslang/opslang/internal/version.Version=$(VERSION) \
           -X github.com/opslang/opslang/internal/version.BuildTime=$(BUILD_TIME) \
           -X github.com/opslang/opslang/internal/version.GitCommit=$(GIT_COMMIT)

# Go 参数
GO := go
GOFLAGS := -trimpath
GOTEST := $(GO) test -v -race

## build: 编译 ops 命令行工具
build:
	@echo "🔨 Building OpsLang $(VERSION)..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/ops ./cmd/ops
	@echo "✅ Binary: bin/ops"

## install: 安装到 GOPATH/bin
install: build
	@echo "📦 Installing ops to GOPATH/bin..."
	cp bin/ops $(GOPATH)/bin/ops 2>/dev/null || cp bin/ops /usr/local/bin/ops
	@echo "✅ Installed"

## test: 运行测试
test:
	@echo "🧪 Running tests..."
	$(GOTEST) ./...

## test-cover: 运行测试并生成覆盖率报告
test-cover:
	@echo "🧪 Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

## lint: 代码检查
lint:
	@echo "🔍 Running linter..."
	golangci-lint run ./...

## fmt: 格式化代码
fmt:
	@echo "🎨 Formatting code..."
	$(GO) fmt ./...
	goimports -w .

## clean: 清理构建产物
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/ coverage.out coverage.html
	@echo "✅ Cleaned"

## run: 运行 ops 脚本
run:
	@echo "▶️  Running $(file)..."
	./bin/ops run $(file)

## repl: 启动交互式 REPL
repl: build
	@echo "🔄 Starting REPL..."
	./bin/ops repl

## examples: 运行示例
examples:
	@echo "▶️  Running examples..."
	@for f in examples/basic/*.ops; do \
		echo "--- $$f ---"; \
		./bin/ops run $$f; \
		echo; \
	done

## dev: 开发模式（编译+运行）
dev: build
	./bin/ops $(args)

## help: 显示帮助
help:
	@echo "OpsLang Build System"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
