# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 1 COMPLETE** [2025-06-17]: 38+ bridges across 13 categories  
✅ **Phase 2.1-2.2 COMPLETE** [2025-06-18]: Lua core engine implemented  
✅ **Phase 2.3.2 COMPLETE** [2025-06-19]: Async support & ScriptValue integration  
✅ **Phase 2.3.2.5 COMPLETE** [2025-06-19]: GopherLua Engine uses ScriptValue internally
🚧 **Phase 2.3.3 ACTIVE**: Bridge adapters (2/14 done) - State adapter next

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
```

## Implementation Workflow

1. **Read TODO.md** - TDD mandatory - Write tests first
2. **Bridge-first** - Wrap go-llms, never reimplement business logic  
3. **Run `make all`** - Complete dev cycle (fmt, vet, lint, test, build)
4. **Update TODO-DONE.md** - Mark completed tasks with timestamps

### ScriptValue System
- ExecuteMethod(ctx, name, args []engine.ScriptValue) (engine.ScriptValue, error)
- ValidateMethod(name, args []engine.ScriptValue) error
- Use engine.NewXXXValue() constructors, type check with args[i].Type()

## Current Phase: Lua Bridge Adapters

**Completed**: 2.3.3.2 - LLM and Provider Bridge Adapter ✅

**Next Task**: 2.3.3.3 - State Bridge Adapter (`/pkg/engine/gopherlua/adapters/state.go`)
- Create state and context management module
- Implement get/set operations
- Add transform functions
- Implement persistence methods
- Add state merging capabilities

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