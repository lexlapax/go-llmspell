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

.PHONY: all build clean test coverage fmt vet lint test-integration test-unit deps help examples example example-mock

# Default target
all: clean fmt vet lint test build
	@echo ""
	@echo "⚠️  NOTE: The build step above may have failed due to ongoing v0.3.3 migration"
	@echo "⚠️  Only pkg/engine/* is currently implemented. See TODO.md for progress"

# Build the binary
build:
	@echo "⚠️  WARNING: The cmd/llmspell is not yet migrated to v0.3.3 architecture"
	@echo "⚠️  This build may fail or produce non-functional binaries"
	@echo "⚠️  See TODO.md for migration progress"
	@echo ""
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	-$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH) 2>/dev/null || echo "❌ Build failed - cmd not yet implemented for v0.3.3"
	@echo ""

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Clean complete"

# Run tests
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) $(TEST_FLAGS) ./pkg/...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) $(TEST_FLAGS) -tags=integration ./test/integration/...

# Run all tests (unit + integration)
test-all: test-unit test-integration

# Generate test coverage
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./pkg/...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .
	@echo "Format complete"

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet complete"

# Run linter
lint:
	@echo "Running linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

# Run the application
run: build
	@echo "⚠️  WARNING: The application may not run correctly until migration is complete"
	@echo "Running $(BINARY_NAME)..."
	-./$(BINARY_DIR)/$(BINARY_NAME) 2>/dev/null || echo "❌ Run failed - cmd not yet implemented for v0.3.3"

# Example targets
examples:
	@echo "⚠️  WARNING: Examples are not yet migrated to v0.3.3 architecture"
	@echo "⚠️  They may fail or produce unexpected results"
	@echo ""
	@echo "Available example spells (NOT YET MIGRATED):"
	@echo "  hello-llm         - Basic LLM interaction example"
	@echo "  chat-assistant    - Interactive chat assistant"
	@echo "  provider-compare  - Compare responses across providers"
	@echo "  web-summarizer    - Summarize web content (requires http module)"
	@echo ""
	@echo "Run an example with: make example SPELL=<spell-name>"
	@echo "Example: make example SPELL=hello-llm"

# Run a specific example spell
example:
	@echo "⚠️  WARNING: Examples are not yet migrated to v0.3.3 architecture"
	@if [ -z "$(SPELL)" ]; then \
		echo "Error: SPELL not specified. Usage: make example SPELL=<spell-name>"; \
		echo "Run 'make examples' to see available spells"; \
		exit 1; \
	fi
	@if [ ! -d "examples/spells/$(SPELL)" ]; then \
		echo "Error: Spell '$(SPELL)' not found in examples/spells/"; \
		echo "Run 'make examples' to see available spells"; \
		exit 1; \
	fi
	@echo "Running example spell: $(SPELL)"
	@echo "❌ Examples not yet implemented for v0.3.3 - see TODO.md for progress"

# Run example with mock LLM (no API key required)
example-mock:
	@echo "⚠️  WARNING: Examples are not yet migrated to v0.3.3 architecture"
	@if [ -z "$(SPELL)" ]; then \
		echo "Error: SPELL not specified. Usage: make example-mock SPELL=<spell-name>"; \
		echo "Run 'make examples' to see available spells"; \
		exit 1; \
	fi
	@if [ ! -d "examples/spells/$(SPELL)" ]; then \
		echo "Error: Spell '$(SPELL)' not found in examples/spells/"; \
		echo "Run 'make examples' to see available spells"; \
		exit 1; \
	fi
	@echo "Running example spell with mock LLM: $(SPELL)"
	@echo "❌ Examples not yet implemented for v0.3.3 - see TODO.md for progress"

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "⚠️  WARNING: Cross-platform builds not yet supported for v0.3.3"
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)
	-GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH) 2>/dev/null || echo "❌ Linux build failed"
	-GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH) 2>/dev/null || echo "❌ Linux ARM build failed"

build-darwin:
	@echo "⚠️  WARNING: Cross-platform builds not yet supported for v0.3.3"
	@echo "Building for macOS..."
	@mkdir -p $(BINARY_DIR)
	-GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH) 2>/dev/null || echo "❌ macOS build failed"
	-GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH) 2>/dev/null || echo "❌ macOS ARM build failed"

build-windows:
	@echo "⚠️  WARNING: Cross-platform builds not yet supported for v0.3.3"
	@echo "Building for Windows..."
	@mkdir -p $(BINARY_DIR)
	-GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH) 2>/dev/null || echo "❌ Windows build failed"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Check for security vulnerabilities
security:
	@echo "Checking for vulnerabilities..."
	$(GOCMD) list -json -m all | nancy sleuth

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

# Quick build without linting
quick: clean build

# Development mode - watch for changes and rebuild
watch:
	@echo "Watching for changes..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
	fi

# Show migration status
migration-status:
	@echo "🔄 Go-LLMSpell v0.3.3 Migration Status"
	@echo "======================================"
	@echo "✅ Implemented:"
	@echo "  - pkg/engine/interface.go (ScriptEngine, Bridge, TypeConverter)"
	@echo "  - pkg/engine/registry.go (Engine Registry)"
	@echo "  - pkg/engine/types.go (Type System)"
	@echo ""
	@echo "⏳ In Progress:"
	@echo "  - See TODO.md for current tasks"
	@echo ""
	@echo "❌ Not Started:"
	@echo "  - cmd/llmspell (CLI application)"
	@echo "  - examples/* (Example spells)"
	@echo "  - pkg/bridge/* (All bridges)"
	@echo "  - pkg/core/* (Agent and state systems)"
	@echo "  - pkg/engine/lua/* (Lua engine)"
	@echo "  - pkg/engine/javascript/* (JavaScript engine)"
	@echo "  - pkg/engine/tengo/* (Tengo engine)"
	@echo ""
	@echo "📋 For detailed progress, see TODO.md and TODO-DONE.md"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make test           - Run unit tests"
	@echo "  make test-all       - Run all tests (unit + integration)"
	@echo "  make coverage       - Generate test coverage report"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run linter"
	@echo "  make deps           - Download dependencies"
	@echo "  make install-tools  - Install development tools"
	@echo "  make run            - Build and run the application"
	@echo "  make examples       - List available example spells"
	@echo "  make example SPELL=<name> - Run a specific example spell"
	@echo "  make example-mock SPELL=<name> - Run example with mock LLM"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make bench          - Run benchmarks"
	@echo "  make watch          - Watch for changes and rebuild"
	@echo "  make migration-status - Show v0.3.3 migration progress"
	@echo "  make help           - Show this help message"