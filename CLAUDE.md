# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-llmspell is a Go library that provides a scriptable interface for LLM interactions using embedded scripting languages (starting with Lua, then JavaScript and Tengo). It acts as a wrapper around the go-llms library, providing scripting capabilities for AI agent orchestration and workflow automation.

## Current Status (Last Updated: May 27, 2025)

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

### In Progress
- 🔄 Phase 4: Tool System - Create tool interface

### Next Steps
1. Implement Tool System (Phase 4)
   - Create tool interface and registry
   - Integrate with go-llms built-in tools
   - Add tool bridge for script-defined tools
2. Implement Agent System (Phase 5)
   - Create agent interface and configuration
   - Integrate with go-llms agent capabilities
   - Add conversation memory management
3. Add missing LLM bridge features:
   - llm.stream_chat_with_history() for message-based streaming
   - Safe alternatives to io.read/write for interactive spells
4. Continue with Workflow system (Phase 6)

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
- `/pkg/bridge/` - Bridge implementations (LLM bridge complete, conversions added)
- `/pkg/security/` - Security context and resource management (implemented)
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
- `github.com/yuin/gopher-lua` - Lua scripting engine (to be added)
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