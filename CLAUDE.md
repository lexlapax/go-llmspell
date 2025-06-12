# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-llmspell is a Go library providing **scriptable LLM interactions** using embedded scripting languages (Lua, JavaScript, Tengo). It wraps go-llms v0.3.3 with multi-engine scripting for AI agent orchestration.

## Current Status (June 2025)

✅ **Phase 1.1 Complete**: Script Engine Interface foundation
✅ **Phase 1.2 Complete**: Core Agent System implementation
🚧 **Phase 1.3 Next**: State Management System (evaluating need)
🎯 **Target**: Full multi-engine support with go-llms v0.3.3 features

### Completed Components
- Core interfaces (ScriptEngine, Bridge, TypeConverter)
- Engine Registry with thread-safe operations
- Type System with cross-engine conversions
- Bridge Manager with lifecycle management
- Core bridges: LLM (streaming), Utilities, Model Info
- Agent System (interface, base, registry, context)

## Development Workflow

1. **Read TODO.md** for current tasks
2. **Write tests first** (TDD mandatory)
3. **Implement feature**
4. **Run `make all`** (fmt, vet, lint, test, build)
5. **Update TODO-DONE.md** when complete

## Key Commands

```bash
make all        # Run complete development cycle
make test       # Run tests with race detection
make lint       # Check code quality
make build      # Build binary
```

## Architecture

```
/pkg/engine/     # Core interfaces & registry
/pkg/bridge/     # Go-script bridges (LLM, utils, etc.)
/pkg/core/       # Engine-agnostic systems
  /agent/        # Agent interface, base, registry, context
```

### Design Principles
- **Engine-agnostic** core functionality
- **Test-driven** development (tests before code)
- **Thread-safe** implementations
- **Security-first** with sandboxed execution

## Implementation Guidelines

- Follow existing patterns in completed components
- Ensure thread-safety with proper locking
- Write comprehensive table-driven tests
- Document public APIs with godoc comments
- Handle errors explicitly, never panic
- Use contexts for cancellation/timeouts

## Important Files

- **TODO.md** - Current implementation tasks
- **TODO-DONE.md** - Completed work tracking
- **docs/MIGRATION_PLAN_V0.3.3.md** - Architecture design

## Notes

- Legacy Lua-only code exists but being replaced
- Memory subsystem deferred (not in go-llms yet)
- Always run `make all` before committing

## Memories
- Do not ever modify the git sub-modules like go-llms