# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-llmspell is a Go library that provides a scriptable interface for LLM interactions using embedded scripting languages (starting with Lua, then JavaScript and Tengo). It acts as a wrapper around the go-llms library, providing scripting capabilities for AI agent orchestration and workflow automation.

## Current Status (Last Updated: May 31, 2025)

### Completed
- ✅ Initial project structure with comprehensive directory layout
- ✅ Architecture documentation (docs/architecture.md, implementation-guide.md, spell-development.md)
- ✅ go-llms v0.2.6 integration as git submodule
- ✅ Basic LLM bridge implementation (pkg/bridge/llm.go)
- ✅ Makefile with build, test, lint, and fmt targets
- ✅ .gitignore for Go projects
- ✅ Comprehensive TODO.md for tracking implementation
- ✅ **Phase 1: Core Infrastructure (COMPLETE)**
  - ✅ Engine interface system with comprehensive API
  - ✅ Thread-safe engine registry with factory pattern
  - ✅ Bridge infrastructure with lifecycle management
  - ✅ Security context with resource limits and monitoring
  - ✅ Complete test coverage using TDD approach
- ✅ **Phase 2: LLM Bridge Enhancement (COMPLETE)**
  - ✅ Multi-provider support (OpenAI, Anthropic, Gemini)
  - ✅ Dynamic provider switching at runtime
  - ✅ Model listing integration with go-llms inventory
  - ✅ Streaming support with proper error handling
  - ✅ Type conversion utilities for Go<->Script bridging
  - ✅ Comprehensive test coverage with race detection
  - ✅ Fixed concurrent access issues
- ✅ **Phase 3: Lua Engine Integration (COMPLETE)**
  - ✅ GopherLua integration with full Engine interface
  - ✅ Comprehensive Lua<->Go type conversions
  - ✅ LLM bridge adapter for Lua scripts
  - ✅ Complete standard library (JSON, HTTP, Storage, Log, Promise)
  - ✅ Security sandbox with disabled dangerous functions
  - ✅ Promise implementation for async patterns (using .next() instead of then)
  - ✅ Example spells: async-llm, provider-compare, chat-assistant
  - ✅ All tests passing with race detection
- ✅ **Phase 4: Tool System (COMPLETE)**
  - ✅ Tool interface and registry implementation
  - ✅ Thread-safe tool registration and execution
  - ✅ Parameter validation with JSON schemas
  - ✅ Lua bridge for tool system (tools module)
  - ✅ Script-based tool creation with tools.register()
  - ✅ Tool execution, listing, and management
  - ✅ Example tools: calculator, string tools, JSON processor
  - ✅ Comprehensive test coverage
- ✅ **Phase 5: Agent System (COMPLETE)**
  - ✅ Agent interface with comprehensive API (pkg/agents/interface.go)
  - ✅ Thread-safe agent registry with factory pattern
  - ✅ Default agent implementation wrapping go-llms agents
  - ✅ Tool integration with existing tool registry
  - ✅ Agent bridge for script access (pkg/bridge/agents.go)
  - ✅ Streaming support with callbacks
  - ✅ Comprehensive test coverage with mocks
  - ✅ Full Lua integration (agents_bridge.go, lua_agent.go, agents_wrapper.go)
  - ✅ Comprehensive lua-agent example with research, code analysis, and planning agents

### Recent Updates
- ✅ **Agent System Complete (May 31, 2025)**
  - Completed full Lua integration for agents
  - Created comprehensive lua-agent example with three agent patterns
  - Research agent demonstrates tool integration with web_fetch
  - Code analysis agent shows custom Lua tool usage
  - Planning agent illustrates multi-step orchestration
  - All tests passing with race detection

### In Progress
- 🔄 No active phase - ready for Phase 6 (Workflow System) or Phase 7 (Spell System)

### Next Steps
1. Begin Phase 6: Workflow System
   - Create workflow engine (`pkg/workflows/engine.go`)
   - Implement step execution logic
   - Add conditional branching and parallel execution
   - Create workflow bridge for scripts
2. Alternative: Begin Phase 7: Spell System
   - Create spell loader and runner
   - Implement spell metadata parsing
   - Add spell packaging format
3. Add missing features (lower priority):
   - llm.stream_chat_with_history() for message-based streaming
   - Safe alternatives to io.read/write for interactive spells
   - Additional agent examples (multi-turn conversation)
4. Investigate and integrate more built-in tools from go-llms

## Development Commands

### Essential Daily Commands
```bash
# Development workflow (format, vet, lint, test, build)
make all

# Quick build without linting
make quick

# Run tests with race detection (recommended for development)
make test

# Run all tests (unit + integration)
make test-all

# Generate test coverage report
make coverage

# Format, vet, and lint code
make fmt && make vet && make lint
```

### Build and Run Commands
```bash
# Build the binary
make build

# Build and run the application
make run

# Clean build artifacts
make clean

# Build for all platforms (Linux, macOS, Windows)
make build-all
```

### Example Spell Commands
```bash
# List available example spells
make examples

# Run a specific example spell
make example SPELL=hello-llm
make example SPELL=async-llm
make example SPELL=provider-compare

# Run example with mock LLM (no API key required)
make example-mock SPELL=hello-llm
```

### Testing Commands
```bash
# Run unit tests only
make test-unit

# Run integration tests (requires API keys)
make test-integration

# Run benchmarks
make bench

# Watch for changes and rebuild (requires air)
make watch
```

### Development Tools
```bash
# Install development tools (golangci-lint)
make install-tools

# Check for security vulnerabilities (requires nancy)
make security

# Generate documentation server
make docs

# Download and update dependencies
make deps
```

## Architecture Overview

The project follows a layered architecture:

1. **Spell Layer**: User-created scripts in Lua/JS/Tengo
2. **Script Engine Layer**: Language-specific interpreters with common interface
3. **Bridge Layer**: Go code exposing functionality to scripts
4. **go-llms Layer**: Core LLM library integration

### Key Components
- **Engine Interface**: Common abstraction for all script engines
- **Registry System**: Manages engines, tools, and agents
- **Bridge Pattern**: Clean separation between Go and script environments
- **Security**: Sandboxing with resource limits

### Current Package Structure
- `/cmd/llmspell/` - CLI entry point
- `/pkg/engine/` - Script engine interface and implementations
  - `/lua/` - GopherLua engine implementation with full bridge system
  - `/registry.go` - Thread-safe engine registry with factory pattern
- `/pkg/bridge/` - Bridge implementations for script-Go interop
  - `llm.go` - LLM provider bridge (OpenAI, Anthropic, Gemini)
  - `tools.go` - Tool system bridge for script access
  - `agents.go` - Agent management bridge
  - `conversions.go` - Type conversion utilities
- `/pkg/security/` - Security context and resource management
- `/pkg/tools/` - Tool system with registry and JSON schema validation
- `/pkg/agents/` - Agent system with go-llms integration
  - Thread-safe registry, factory pattern, tool integration
- `/examples/spells/` - Example spells demonstrating features
  - `async-llm/`, `provider-compare/`, `chat-assistant/`, etc.
- `/docs/` - Architecture and development documentation
- `/go-llms/` - Git submodule for go-llms library reference

## Development Guidelines

### Testing Strategy (TDD)
1. Write tests before implementation
2. Use table-driven tests for comprehensive coverage
3. Mock external dependencies
4. Aim for >80% test coverage

### Code Style
- Follow standard Go conventions
- Use meaningful variable names
- Add godoc comments for all exported types
- Keep functions focused and small

### Security First
- Implement sandboxing from the start
- Validate all inputs
- Enforce resource limits
- Use context for cancellation

### Commit Workflow
1. Create feature branch
2. Write tests first (TDD)
3. Implement feature
4. Run `make fmt`, `make vet`, `make lint`, `make test`
5. Fix any issues before committing
6. Update documentation as needed

## Key Dependencies
- `github.com/lexlapax/go-llms` v0.2.6 - Core LLM wrapper library (integrated)
- `github.com/yuin/gopher-lua` v1.1.1 - Lua scripting engine (integrated)
- `github.com/joho/godotenv` v1.5.1 - Environment file loading (integrated)
- `github.com/dop251/goja` - JavaScript scripting engine (to be added)
- `github.com/d5/tengo` - Tengo scripting engine (to be added)

## Important Notes
- Always use TDD approach for new features
- Run quality checks (fmt, vet, lint, test) after feature completion
- Refer to TODO.md for current tasks and priorities
- Check TODO-DONE.md for completed work reference

### Development Principles
- **TDD Approach**: Write tests before implementation for all new features
- **Security First**: Implement sandboxing and validation from the start
- **Thread Safety**: Use mutex protection for all shared resources
- **Registry Pattern**: Follow established patterns for extensibility

### Post-Feature Workflow
```bash
# Run complete quality check pipeline
make fmt vet lint test build

# Or use the all target
make all
```

### Dependency Management
- Never change underlying dependency libraries even if you have access to source via git sub-modules
- Always use `go mod tidy` after adding dependencies
- Keep go-llms submodule as reference only, use Go module for actual dependency

### Environment Setup
- Copy `.env.example` to `.env` and add your API keys:
  ```bash
  OPENAI_API_KEY=sk-...
  ANTHROPIC_API_KEY=sk-ant-...
  GEMINI_API_KEY=AI...
  ```
- The CLI automatically loads `.env` file if present
- Use `make example-mock SPELL=<name>` to run examples without API keys

### Debugging and Development Tips
- Use `make test` for development (includes race detection)
- Use `make coverage` to see test coverage reports
- Enable debug logging in spells with `log.set_level("debug")`
- Use `make watch` for automatic rebuilds during development (requires air)
- Check `/examples/spells/` for implementation patterns

### Script Engine Development (Lua)
- Security sandbox disables dangerous functions (os.execute, io operations)
- Use stdlib modules: json, http, storage, log, promise
- Agents can be created in Lua using `agents.register()`
- Tools can be registered from Lua using `tools.register()`
- Promises use `.next()` instead of `then()` (reserved keyword in Lua)

## Additional Context and Reminders

### LLM Integration Patterns
- **Provider Switching**: Use `llm.set_provider()` to switch between OpenAI, Anthropic, Gemini
- **Model Selection**: Use `llm.set_model()` after setting provider
- **Streaming**: Use `llm.stream_chat()` with callback functions for real-time responses
- **Async Patterns**: Use Promise system in Lua for parallel operations

### Built-in Tools Available
- **web_fetch**: Fetch content from URLs (built into go-llms)
- **calculator**: Basic mathematical operations
- **string_tools**: Text manipulation (reverse, uppercase, etc.)
- **json_processor**: JSON validation and transformation
- Scripts can register additional tools using `tools.register()`

### Agent System Patterns
- **Default Agents**: Use `agents.create()` to create go-llms wrapped agents
- **Lua Agents**: Register custom agents using `agents.register()` with function or table
- **Tool Integration**: Agents can use any registered tools
- **Streaming**: Use `agents.stream()` for streaming agent responses

### Testing Approach
- Unit tests for all components using TDD
- Integration tests for full spell execution
- Race detection enabled by default in test runs
- Mock implementations available for testing without API keys
- Benchmarks for performance-critical components

### Security Implementation
- Filesystem access completely sandboxed
- Network requests restricted to HTTPS
- Resource limits enforced (memory, CPU, goroutines)
- Dangerous Lua functions disabled (os.execute, io.*, debug.*)
- Storage module provides safe file operations

### Common Development Tasks
- **Single Test**: `go test ./pkg/specific/package -v`
- **Race Detection**: `go test -race ./...`
- **With Coverage**: `go test -cover ./pkg/...`
- **Specific Function**: `go test -run TestFunctionName ./pkg/package`
- **Integration Only**: `make test-integration`
- **Benchmarks**: `go test -bench=. ./pkg/...`