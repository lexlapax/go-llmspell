# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phases 1-2.3.2 COMPLETE** [2025-06-19]: All 21 bridges converted to ScriptValue system  
🚧 **Phase 2.3.3 ACTIVE**: Bridge adapters (2/14 done) - State adapter next

- ScriptValue refactoring complete with full type safety
- Performance benchmarks show minimal overhead  
- Migration guide available for bridge authors

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
```

## Implementation Workflow

1. **Read TODO.md** - TDD mandatory - Write tests first
2. **Bridge-first** - Wrap go-llms, never reimplement business logic  
3. **Research and Plan codebase** - Always look at `go-llms/` git sub-module and the implementation there to inform what files and what to implement locally in go-llmspell
4. **Run `make all`** - Complete dev cycle (fmt, vet, lint, test, build)
5. **Update TODO.md** - Mark completed tasks with timestamps

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