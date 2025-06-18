# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phases 1-2.3.2 COMPLETE** [2025-06-19]: All 21 bridges converted to ScriptValue system  
🚧 **Phase 2.3.2.5 ACTIVE**: Update GopherLua Engine to use ScriptValue internally

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

## Current Phase: GopherLua Engine ScriptValue Integration

**Tasks**: LValue↔ScriptValue converters, update engine pipeline, bridge adapter integration, comprehensive testing

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