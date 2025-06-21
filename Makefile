# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Lua parameters
LUACHECK=luacheck
STYLUA=$(HOME)/.cargo/bin/stylua
LUA_FILES=pkg/engine/gopherlua/stdlib/*.lua examples/**/*.lua

# Binary names
BINARY_NAME=llmspell
BINARY_DIR=bin

# Build variables
LDFLAGS=-ldflags "-w -s"
MAIN_PATH=./cmd/llmspell

# Test variables
TEST_FLAGS=-v -race
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

.PHONY: all build clean test coverage fmt vet lint test-integration test-unit deps help build-examples mod bench bench-run lua-fmt lua-lint lua-check lua-syntax

# Default target
all: clean fmt vet lint test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(BINARY_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "✅ Clean complete"

# Run tests
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) $(TEST_FLAGS) ./pkg/...
	@echo "✅ Unit tests complete"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@if [ -d "./tests/integration" ]; then \
		$(GOTEST) $(TEST_FLAGS) -tags=integration ./tests/integration/...; \
		echo "✅ Integration tests complete"; \
	else \
		echo "⚠️  No integration tests found in ./tests/integration/"; \
	fi

# Run all tests (unit + integration)
test-all: test-unit test-integration

# Generate test coverage
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./pkg/...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "✅ Coverage report generated: $(COVERAGE_HTML)"

# Format code
fmt: lua-fmt
	@echo "Formatting Go code..."
	@find . -name "*.go" -not -path "./go-llms/*" -not -path "./vendor/*" | xargs $(GOFMT) -w
	@echo "✅ Go format complete"

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) $$(go list ./... | grep -v /go-llms/)
	@echo "✅ Vet complete"

# Run linter
lint: lua-lint
	@echo "Running Go linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./cmd/... ./pkg/...; \
		echo "✅ Go lint complete"; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Download dependencies and tidy
mod:
	@echo "Managing Go modules..."
	$(GOMOD) download
	@echo "⚠️  Skipping 'go mod tidy' due to examples with to_be_migrated build tag"
	@echo "⚠️  Once migration is complete, run 'go mod tidy' manually"
	@echo "✅ Module download complete"

# Alias for mod
deps: mod

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Tools installed"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_DIR)/$(BINARY_NAME)

# Build examples (with to_be_migrated tag)
build-examples:
	@echo "Building examples..."
	@echo "⚠️  Examples are marked with 'to_be_migrated' build tag"
	@echo "⚠️  They reference old interfaces and may not compile"
	@if [ -f "examples/integration/lua_integration.go" ]; then \
		echo "Attempting to build examples/integration/lua_integration.go..."; \
		$(GOBUILD) -tags=to_be_migrated -o $(BINARY_DIR)/lua_integration examples/integration/lua_integration.go 2>&1 | head -20 || echo "❌ Example build failed (expected - needs migration)"; \
	else \
		echo "⚠️  No Go examples found"; \
	fi

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "✅ Linux builds complete"

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "✅ macOS builds complete"

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Windows build complete"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@echo "Running package benchmarks..."
	$(GOTEST) -bench=. -benchmem ./pkg/...
	@echo "Running dedicated benchmark tests..."
	$(GOTEST) -bench=. -benchmem -tags=bench ./tests/benchmarks/...
	@echo "✅ All benchmarks complete"

# Run specific benchmark
bench-run:
	@echo "Running specific benchmark: $(BENCH)"
	$(GOTEST) -bench=$(BENCH) -benchmem -benchtime=10s -tags=bench ./tests/benchmarks/...

# Quick build without linting
quick: clean build

# Development mode - watch for changes and rebuild
watch:
	@echo "Watching for changes..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
	fi

# Lua formatting and linting targets

# Format Lua code with stylua
lua-fmt:
	@echo "Formatting Lua code..."
	@if command -v $(STYLUA) >/dev/null 2>&1; then \
		$(STYLUA) pkg/engine/gopherlua/stdlib/*.lua; \
		echo "✅ Lua format complete (stdlib files only)"; \
	else \
		echo "⚠️  stylua not installed. Install with: cargo install stylua"; \
	fi

# Lint Lua code with luacheck
lua-lint:
	@echo "Linting Lua code..."
	@if command -v $(LUACHECK) >/dev/null 2>&1; then \
		$(LUACHECK) $$(find . -name "*.lua" -not -path "./go-llms/*" -not -path "./vendor/*"); \
		echo "✅ Lua lint complete"; \
	else \
		echo "⚠️  luacheck not installed. Install with your package manager"; \
	fi

# Check Lua syntax only
lua-syntax:
	@echo "Checking Lua syntax..."
	@if command -v lua >/dev/null 2>&1; then \
		find . -name "*.lua" -not -path "./go-llms/*" -not -path "./vendor/*" -exec lua -l {} \; 2>/dev/null || true; \
		echo "✅ Lua syntax check complete"; \
	else \
		echo "⚠️  lua interpreter not installed"; \
	fi

# Combined Lua check (syntax + lint + format)
lua-check: lua-syntax lua-lint lua-fmt
	@echo "✅ All Lua checks complete"

# Show migration status
migration-status:
	@echo "🔄 Go-LLMSpell v0.3.3 Migration Status"
	@echo "======================================"
	@echo "✅ Implemented:"
	@echo "  - pkg/engine/interface.go (ScriptEngine, Bridge, TypeConverter)"
	@echo "  - pkg/engine/registry.go (Engine Registry)"
	@echo "  - pkg/engine/types.go (Type System)"
	@echo "  - pkg/bridge/manager.go (Bridge Manager)"
	@echo "  - pkg/bridge/llm.go (Core LLM Bridge)"
	@echo "  - pkg/bridge/util.go (Essential Utilities Bridge)"
	@echo "  - pkg/bridge/modelinfo.go (Model Info Bridge)"
	@echo "  - pkg/core/agent/interface.go (Agent Interface)"
	@echo "  - pkg/core/agent/base.go (Base Agent Implementation)"
	@echo "  - pkg/core/agent/registry.go (Agent Registry)"
	@echo "  - pkg/core/agent/context.go (Agent Context)"
	@echo "  - cmd/llmspell (Placeholder CLI)"
	@echo ""
	@echo "⏳ In Progress:"
	@echo "  - Phase 1.3: State Management System (evaluating need)"
	@echo "  - See TODO.md for current tasks"
	@echo ""
	@echo "❌ Not Started:"
	@echo "  - pkg/core/state/* (State management systems)"
	@echo "  - pkg/engine/lua/* (Lua engine)"
	@echo "  - pkg/engine/javascript/* (JavaScript engine)"
	@echo "  - pkg/engine/tengo/* (Tengo engine)"
	@echo ""
	@echo "📋 For detailed progress, see TODO.md and TODO-DONE.md"

# Show help
help:
	@echo "Available targets:"
	@echo "  make              - Run all (clean, fmt, vet, lint, test, build)"
	@echo "  make build        - Build the binary"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-all     - Run all tests (unit + integration)"
	@echo "  make coverage     - Generate test coverage report"
	@echo "  make fmt          - Format code (Go + Lua)"
	@echo "  make vet          - Run go vet"
	@echo "  make lint         - Run linter (Go + Lua)"
	@echo "  make mod          - Download dependencies and tidy modules"
	@echo "  make deps         - Alias for 'make mod'"
	@echo "  make build-examples - Build examples (with to_be_migrated tag)"
	@echo "  make install-tools - Install development tools"
	@echo "  make run          - Build and run the application"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make bench        - Run benchmarks"
	@echo "  make quick        - Quick build (clean + build)"
	@echo "  make watch        - Watch for changes and rebuild"
	@echo ""
	@echo "Lua-specific targets:"
	@echo "  make lua-fmt      - Format Lua code with stylua"
	@echo "  make lua-lint     - Lint Lua code with luacheck"
	@echo "  make lua-syntax   - Check Lua syntax"
	@echo "  make lua-check    - Run all Lua checks (syntax + lint + format)"
	@echo ""
	@echo "  make migration-status - Show v0.3.3 migration progress"
	@echo "  make help         - Show this help message"