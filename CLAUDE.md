# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Project Overview

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 1 COMPLETE** [2025-06-17]: 38+ bridges, zero business logic duplication  
✅ **Phase 2.1-2.2 COMPLETE** [2025-06-18]: Core Engine Components - LuaEngine fully implemented  
✅ **Phase 2.3.1 COMPLETE** [2025-06-19]: Module System Architecture  
🚧 **Phase 2.3.2 IN PROGRESS**: ScriptValue Type System Refactoring - Converting bridges from []interface{} to []engine.ScriptValue

### Latest Progress [2025-06-19]
✅ **util/* bridges COMPLETE** - All 8 util bridges converted (auth, debug, errors, json, script_logger, slog, llm, util)
✅ **state/* bridges COMPLETE** - Both state bridges converted (manager, context) with comprehensive tests
🚧 **agent/* bridges STARTED** - Beginning systematic conversion of 6 agent bridges

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
```

## Implementation Workflow

1. **Read TODO.md** - Current Phase 2.3.2 tasks (ScriptValue Refactoring)
2. **Backup Pattern** - mv old.go old.go.backup → create new from scratch → compare methods
3. **TDD mandatory** - Write tests first, then implement  
4. **Bridge-first** - Wrap go-llms, never reimplement business logic
5. **Run `make all`** - Complete dev cycle (fmt, vet, lint, test, build)
6. **Update TODO-DONE.md** - Mark completed tasks with timestamps

### ScriptValue Conversion Rules
- ExecuteMethod(ctx, name, args []engine.ScriptValue) (engine.ScriptValue, error)
- ValidateMethod(name, args []engine.ScriptValue) error  
- Use engine.NewXXXValue() constructors for return values
- Type check with args[i].Type() == engine.TypeString before casting
- Helper: convertToScriptValue() for complex conversions

## Phase 2.3.2: ScriptValue Type System Refactoring

**Current Focus**: Converting all bridges from []interface{} to []engine.ScriptValue for type safety and consistency

### Progress Summary [2025-06-19]
- ✅ **util/auth.go** - Rewritten from scratch with ScriptValue (backup pattern)
- ✅ **util/debug.go** - Converted in-place with ExecuteMethod dispatcher  
- ✅ **util/errors.go** - Rewritten from scratch with ScriptValue
- ✅ **util/json.go** - Rewritten from scratch with ScriptValue
- ✅ **util/script_logger.go** - Rewritten from scratch, unified logger
- ✅ **util/slog.go** - Rewritten from scratch with ScriptValue
- ✅ **util/llm.go** - Converted in-place with minimal changes
- ✅ **util/util.go** - Converted in-place with minimal changes
- ✅ **state/manager.go** - Converted in-place to ScriptValue
- ✅ **state/context.go** - Rewritten from scratch (45KB vs 109KB backup - focused on core functionality)
- 🚧 **agent/agent.go** - STARTED: Backed up, ready for rewrite

### Next: agent/* bridges (6 remaining)
- agent.go, events.go, hooks.go, tool_registry.go, tools.go, workflow.go

**Pattern**: backup-old-file → create-new-from-scratch → compare-methods → backup-test-file → create-new-test

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