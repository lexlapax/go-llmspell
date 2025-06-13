# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-llmspell is a Go library providing **scriptable LLM interactions** using embedded scripting languages (Lua, JavaScript, Tengo). It bridges go-llms v0.3.3 functionality to scripts without reimplementing features.

## Current Status (June 2025)

✅ **Phase 1.1 Complete**: Script Engine Interface foundation  
✅ **Phase 1.2 Complete**: Core Bridge Foundation (state, utilities)  
🚧 **Phase 1.3 Active**: Core Bridge System (agents, workflows, events, tools)  
🎯 **Target**: Pure bridge architecture exposing go-llms to scripts

### Completed Components
- Core interfaces (ScriptEngine, Bridge, TypeConverter)
- Engine Registry with thread-safe operations
- Type System with cross-engine conversions
- Bridge Manager with lifecycle management
- State bridges: Manager, Context, Persistence
- Utility bridges: Auth, JSON, LLM, General
- Bridge type system with go-llms aliases

## Architecture Principle

**Fundamental Rule**: If it's not in go-llms, we don't implement it in go-llmspell.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
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

## Important Files

- **TODO.md** - Current implementation tasks
- **TODO-DONE.md** - Completed work tracking
- **docs/technical/architecture.md** - Bridge-first design

## Notes

- Bridge pattern strictly enforced
- Always run `make all` before committing
- Never modify go-llms submodule