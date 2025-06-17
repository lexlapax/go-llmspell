# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Project Overview

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 1 COMPLETE** [2025-06-17]: 38+ bridges, zero business logic duplication  
✅ **Phase 2.1-2.2 COMPLETE** [2025-06-18]: Core Engine Components - LuaEngine fully implemented  
✅ **Phase 2.3.1 COMPLETE** [2025-06-19]: Module System Architecture  
🚧 **Phase 2.3.2 NEXT**: Async/Coroutine Support - Foundation for bridge operations

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
```

## Implementation Workflow

1. **Read TODO.md** - Current Phase 2.3.2 tasks (Async/Coroutine Support)
2. **TDD mandatory** - Write tests first, then implement
3. **Bridge-first** - Wrap go-llms, never reimplement business logic
4. **Run `make all`** - Complete dev cycle (fmt, vet, lint, test, build)
5. **Update TODO-DONE.md** - Mark completed tasks with timestamps

## Phase 2.3.2: Async/Coroutine Support

**Current Focus**: Building async foundation for bridge operations

- Task 2.3.2.1: Async Runtime (`/pkg/engine/gopherlua/async.go`)
- Task 2.3.2.2: Channel Integration (`/pkg/engine/gopherlua/channels.go`)  
- Task 2.3.2.3: Async Bridge Methods (`/pkg/engine/gopherlua/async_bridges.go`)
- Task 2.3.2.4: Async Testing (`/pkg/engine/gopherlua/async_test.go`)

**Why First**: Bridge adapters need async foundation for streaming, timeouts, and concurrency.

## Commands

```bash
make all   # Complete dev cycle
make test  # Test with race detection  
make lint  # Code quality checks
```

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.