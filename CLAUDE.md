# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Project Overview

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 1 COMPLETE** [2025-06-17]: 38+ bridges, zero business logic duplication  
✅ **Phase 2.1 COMPLETE** [2025-06-17]: Lua engine research and architecture design  
🚧 **Phase 2.2 IN PROGRESS** [2025-06-18]: Lua Engine Core Implementation

### Current Status: Phase 2.2 - Core Engine Components
- ✅ 2.2.3: Security Sandbox [COMPLETED - 2025-06-17]
- ✅ 2.2.2: Type Converter System [COMPLETED - 2025-06-18]  
- 🔄 2.2.1: LState Pool Implementation [NEXT - Dependencies satisfied]
- ⏸️ 2.2.4: Core Engine Integration [WAITING - Depends on LState Pool]

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

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
make all   # Run complete dev cycle
make test  # Test with race detection
make lint  # Check code quality
```

## Implementation Rules

- **Bridge only** - Wrap go-llms, don't reimplement
- **Test everything** - Table-driven tests
- **Thread-safe** - Proper locking
- **No mocks** - Only in test files

## Important Files

- **TODO.md** - Current tasks (Phase 2.2+ detailed implementation tasks)
- **TODO-DONE.md** - Completed Phase 2+ tasks
- **TODO-DONE-ARCHIVE.md** - Phase 1 history
- **docs/technical/gopherlua_engine_architecture_design.md** - Lua implementation blueprint
- **docs/technical/gopherlua_engine_implementation_plan.md** - Detailed implementation roadmap
- **docs/archives/research/** - Archived research documents

## Implementation Guidance

### Phase 2.2 Current Status  
✅ **Security Sandbox** & **Type Converter** completed with comprehensive test coverage  
🔄 **Next**: LState Pool Implementation - dependencies satisfied, ready to start

### Phase 2.2 Next Steps
1. Begin with **Task 2.2.1.1**: Create State Factory (`/pkg/engine/gopherlua/factory.go`)
2. Follow TDD: Write tests first, then implement
3. Use architecture design as reference for all implementations
4. Each component should be independently testable

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.