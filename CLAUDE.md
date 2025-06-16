# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-llmspell is a Go library providing **scriptable LLM interactions** using embedded scripting languages (Lua, JavaScript, Tengo). It bridges go-llms v0.3.5 functionality to scripts without reimplementing features.

## Current Status (June 2025)

🎯 **MILESTONE ACHIEVED**: Complete bridge ecosystem for go-llms v0.3.5

✅ **Phase 1 COMPLETE**: Bridge Foundation [2025-06-16]
- Script Engine Interface (1.1) + Core Bridge Foundation (1.2) + Core Bridge System (1.3)
- v0.3.5 Feature Integration (1.4.1-1.4.11) + Additional Original Bridges (1.5)
- 35+ bridges across 12 categories with comprehensive test coverage
- Pure bridge architecture: zero business logic duplication

🚧 **Phase 2 NEXT**: Lua Engine Implementation

### Bridge Ecosystem (35+ bridges)
**Core Infrastructure**: ScriptEngine interface, Engine Registry, Type System, Bridge Manager  
**State Management**: Manager, Context (schema validation, event emission)  
**Utilities**: Auth v2.0, JSON v2.0, LLM v2.0, Errors v2.0  
**Agent System**: Agent v2.0, Workflow v2.0, Events v2.0, Tools v2.1, Hooks  
**LLM Operations**: Schema validation, provider metadata, streaming  
**Schema System**: Versioning, migration, validation, import/export  
**Observability**: Tracing, Guardrails, Metrics  
**Provider Management**: Provider System, Provider Pool  
**Tool Discovery**: Tools Registry with MCP export  
**Engine Integration**: Event bus, type registry, profiling, API export

## Architecture Principle

**Fundamental Rule**: If it's not in go-llms, we don't implement it in go-llmspell.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
    /agent/      # Agent-related bridges
    /llm/        # LLM provider bridges  
    /state/      # State management bridges
    /util/       # Utility bridges
    ...          # Organized to mirror go-llms structure
```

## Development Workflow

1. **Read TODO.md** for current tasks
2. **Write tests first** (TDD mandatory)
3. **Implement feature** (bridge only, no business logic)
4. **Run `make all`** (fmt, vet, lint, test, build)
5. **Update TODO-DONE.md** when complete

## Key Commands

```bash
make all        # Run complete development cycle
make test       # Run tests with race detection
make fmt        # Format code
make vet        # Run go vet
make lint       # Check code quality
make build      # Build binary
```

## Implementation Guidelines

- **Bridge, don't build** - Only wrap go-llms functionality
- **No business logic** - All intelligence lives in go-llms
- **Type conversions only** - Convert between script and Go types
- **Test everything** - Comprehensive table-driven tests
- **Thread-safe** - Proper locking where needed
- **No mocks in production** - Only in test files
- **Leverage go-llms pkg/testutils** for tests as much as possible for reuse
- **Update TODO.md** as tasks complete

## Important Files

- **TODO.md** - Current implementation tasks
- **TODO-DONE.md** - Completed work tracking
- **docs/technical/architecture.md** - Bridge-first design

## Notes

- Bridge pattern strictly enforced
- Always run `make all` before committing  
- Never modify go-llms submodule
- Phase 1 bridge foundation complete - ready for Phase 2 engine implementations

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.