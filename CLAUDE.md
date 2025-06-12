# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-llmspell is a Go library that provides a **scriptable interface for LLM interactions** using embedded scripting languages (Lua, JavaScript, Tengo). It acts as a wrapper around the go-llms library, providing scripting capabilities for AI agent orchestration and workflow automation.

## Current Status (Last Updated: January 2025)

### Migration to go-llms v0.3.3
- 🚧 **In Progress**: Complete redesign for multi-engine architecture
- 🚧 **Clean Slate**: Moving from legacy single-engine to polyglot design
- 🚧 **Target**: Full go-llms v0.3.3 multi-agent orchestration capabilities

### Legacy Implementation (Archived)
- ✅ Phases 1-5 completed with Lua-only implementation
- ✅ Basic tools, agents, and spells working with go-llms v0.3.0
- ⚠️ Legacy code will be preserved but superseded by new architecture

### New Architecture Highlights
- **Multi-Engine Support**: Lua, JavaScript, Tengo with common interface
- **Engine-Agnostic Core**: All features work across any scripting language
- **Advanced Features**: Distributed tracing, agent handoffs, guardrails, artifacts
- **Production Ready**: Provider pooling, resilience patterns, security, metrics
- **Cloud Native**: Built for distributed execution and scalability

## Development Commands

### Essential Commands
```bash
make all        # Format, vet, lint, test, build
make test       # Run tests with race detection
make coverage   # Generate coverage report
make example SPELL=hello-llm  # Run example spell
```

### Build Commands
```bash
make build      # Build binary
make run        # Build and run
make clean      # Clean artifacts
make build-all  # Cross-platform builds
```

## Architecture Overview

### Layered Architecture
1. **Spell Layer**: User scripts in Lua/JS/Tengo
2. **Engine Layer**: Language-specific interpreters
3. **Bridge Layer**: Go-script interop
4. **go-llms Layer**: Core LLM integration

### Key Design Principles
- **Engine-Agnostic**: Core functionality independent of scripting language
- **Bridge Pattern**: Clean separation between Go and script environments
- **Registry Pattern**: Dynamic registration of engines, tools, agents
- **Security First**: Sandboxed execution with resource limits
- **TDD Approach**: Test-driven development for all features

## Package Structure (v0.3.3)
```
/pkg/engine/         # Script engine interface and registry
    interface.go     # Common ScriptEngine interface
    registry.go      # Engine registration and discovery
    types.go         # Common type system
    /lua/           # Lua engine implementation
    /javascript/    # JavaScript engine (planned)
    /tengo/         # Tengo engine (planned)

/pkg/bridge/        # Language-agnostic bridges
    llm.go          # LLM provider access
    agents.go       # Agent management
    tools.go        # Tool registration
    hooks.go        # Lifecycle hooks
    events.go       # Event system
    tracing.go      # OpenTelemetry integration
    [17 more bridges for v0.3.3 features]

/pkg/core/          # Engine-agnostic core systems
    /agent/         # Agent interface and registry
    /state/         # State management
    /workflow/      # Workflow engine
```

## Key Dependencies
- `github.com/lexlapax/go-llms` v0.3.3 - Core LLM library
- `github.com/yuin/gopher-lua` v1.1.1 - Lua engine
- `github.com/dop251/goja` - JavaScript engine (planned)
- `github.com/d5/tengo` - Tengo engine (planned)

## Development Guidelines

### Workflow
1. Read TODO.md and docs/MIGRATION_PLAN_V0.3.3.md
2. Write tests first (TDD)
3. Implement feature
4. Run `make all` for quality checks
5. Update documentation

### Testing
- Unit tests with >80% coverage
- Table-driven tests
- Race detection enabled
- Mock external dependencies

### Security
- Sandbox all script execution
- Validate inputs
- Enforce resource limits
- Use context for cancellation

## Environment Setup
```bash
cp .env.example .env
# Add API keys:
# OPENAI_API_KEY=sk-...
# ANTHROPIC_API_KEY=sk-ant-...
# GEMINI_API_KEY=AI...
```

## Script Development

### Lua Example
```lua
local agent = agents.create({
    name = "researcher",
    model = "gpt-4",
    tools = {"web_fetch"}
})

local result = agent:run("Research quantum computing")
```

### Common Patterns
- Use promises for async operations (`.next()` in Lua)
- Register custom tools with `tools.register()`
- Create agents with `agents.create()` or `agents.register()`
- Access all go-llms features through bridges

## Important Notes
- 📋 TODO.md tracks all implementation tasks
- 📄 docs/MIGRATION_PLAN_V0.3.3.md has detailed design
- 🔄 Migration to v0.3.3 is complete redesign, not incremental
- ⏸️ Memory subsystem deferred (not in go-llms yet)
- 🧪 Always use TDD approach
- 🔒 Security and sandboxing are mandatory
- 🧵 Ensure thread-safety in all components