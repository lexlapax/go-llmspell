# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-llmspell is a Go library that provides a scriptable interface for LLM interactions using embedded scripting languages (starting with Lua, then JavaScript and Tengo). It acts as a wrapper around the go-llms library, providing scripting capabilities for AI agent orchestration and workflow automation.

## Current Status (Last Updated: May 30, 2025)

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
- ✅ **Phase 5: Agent System (CORE COMPLETE)**
  - ✅ Agent interface with comprehensive API (pkg/agents/interface.go)
  - ✅ Thread-safe agent registry with factory pattern
  - ✅ Default agent implementation wrapping go-llms agents
  - ✅ Tool integration with existing tool registry
  - ✅ Agent bridge for script access (pkg/bridge/agents.go)
  - ✅ Streaming support with callbacks
  - ✅ Comprehensive test coverage with mocks
  - 🔄 Lua integration pending (agents_bridge.go)

### Recent Updates
- ✅ **Promise Async Tests Fix (COMPLETE)**
  - Fixed all failing promise async tests in stdlib
  - Resolved variable scope issues in Lua
  - Added promise state exposure through metamethods
  - Fixed timing and table length issues
  - All tests now passing
- ✅ **Agent System Core Implementation (COMPLETE)**
  - Created comprehensive agent interface following go-llms patterns
  - Implemented thread-safe registry with global instance
  - Built default agent wrapping go-llms DefaultAgent
  - Added tool adapter for seamless integration
  - Created agent bridge for script access
  - Full test coverage using TDD approach

### In Progress
- 🔄 Phase 5: Agent System - Lua integration (agents_bridge.go)

### Next Steps
1. Complete Agent System (Phase 5)
   - Create Lua bridge for agents (agents_bridge.go)
   - Add agent examples in examples/spells/
   - Update documentation with agent usage patterns
2. Add missing LLM bridge features:
   - llm.stream_chat_with_history() for message-based streaming
   - Safe alternatives to io.read/write for interactive spells
3. Continue with Workflow system (Phase 6)
4. Investigate and integrate more built-in tools from go-llms

## Development Commands

### Go Module Management
```bash
# Initialize module dependencies
go mod tidy

# Download dependencies
go mod download

# Run tests (TDD approach)
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
make fmt

# Run linter
make lint

# Run vet
make vet

# Build project
make build

# Clean build artifacts
make clean
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
- `/pkg/engine/` - Script engine interface (implemented)
- `/pkg/bridge/` - Bridge implementations (LLM, tools, and agents bridges complete)
- `/pkg/security/` - Security context and resource management (implemented)
- `/pkg/tools/` - Tool system with registry and validation (implemented)
- `/pkg/agents/` - Agent system with registry and go-llms integration (implemented)
- `/pkg/spells/` - Spell management (basic structure created)
- `/docs/` - Comprehensive documentation
- `/go-llms/` - Submodule for go-llms reference

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
- Always use TDD

### Post-Feature Workflow
- Run make build, make test, make lint, make fmt, make vet and fix errors after feature completion

### Dependency Management
- Never change underlying dependency libraries even if you have access to source via git sub-modules etc.