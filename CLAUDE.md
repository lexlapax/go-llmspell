# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-llmspell is a Go library providing **scriptable LLM interactions** using embedded scripting languages (Lua, JavaScript, Tengo). It bridges go-llms v0.3.5 functionality to scripts without reimplementing features.

## Current Status (June 2025)

✅ **Phase 1.1 Complete**: Script Engine Interface foundation  
✅ **Phase 1.2 Complete**: Core Bridge Foundation (state, utilities)  
✅ **Phase 1.3 Complete**: Core Bridge System (agents, workflows, events, tools, hooks)  
✅ **Phase 1.4.1 Complete**: Foundation Updates [2025-06-15]
✅ **Phase 1.4.2 Complete**: State Bridge Enhancements [2025-06-15]
✅ **Phase 1.4.3 Complete**: Utility Bridge Upgrades [2025-06-16]
✅ **Phase 1.4.4 Complete**: LLM Bridge Advanced Features [2025-06-16]
✅ **Phase 1.4.5 Complete**: Schema Bridge Full Implementation [2025-06-16]
⏸️ **Phase 1.4.6 Deferred**: Model Info Bridge Intelligence (features not in go-llms)
✅ **Phase 1.4.7 Complete**: Agent Bridge Advanced Features [2025-06-16]
✅ **Phase 1.4.8 Complete**: Event Bridge Replacement [2025-06-16]
✅ **Phase 1.4.9 Complete**: Tools Bridge Enhancement [2025-06-16]
🚧 **Phase 1.4.10 Next**: Workflow Bridge Serialization
🎯 **Target**: Pure bridge architecture exposing go-llms to scripts

### Completed Components
- Core interfaces (ScriptEngine, Bridge, TypeConverter)
- Engine Registry with thread-safe operations
- Type System with cross-engine conversions and v0.3.5 integration
- Bridge Manager with lifecycle management, events, documentation, and state serialization
- State bridges: Manager, Context with schema validation and event emission
- Utility bridges: Auth (v2.0), JSON (v2.0), LLM (v2.0), Errors (v2.0)
- Agent bridges: Agent (v2.0), Workflow, Events (v2.0), Tools (v2.1), Hooks
- LLM bridge: Schema validation, provider metadata, streaming with events
- Schema bridge: Versioning, migration, tag generation, import/export, custom validators
- Event bridge v2.0: Complete event system with bus, storage, filtering, serialization, aggregation, and replay
- Agent bridge v2.0: State serialization, event replay, performance profiling
- Tools bridge v2.1: Schema validation, documentation generation, execution analytics
- v0.3.5 integration: schemas, structured outputs, events, docs, errors
- Event-driven bridge lifecycle with metrics and monitoring
- Multi-format documentation generation (OpenAPI, Markdown, JSON)
- Bridge state serialization with versioning and validation

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