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

.PHONY: all build clean test coverage fmt vet lint test-integration test-unit deps help

# Default target
all: clean fmt vet lint test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_DIR)/$(BINARY_NAME)"

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
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_DIR)/$(BINARY_NAME)

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

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
	@echo "  make build-all      - Build for all platforms"
	@echo "  make bench          - Run benchmarks"
	@echo "  make watch          - Watch for changes and rebuild"
	@echo "  make help           - Show this help message"