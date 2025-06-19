# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

🚧 **Phase 2.3.3 ACTIVE**: Bridge Adapters Enhancement & Namespace Flattening (14/24 tasks complete)
- Tasks 15-17: Add missing bridge functionality (tool_registry, llm_pool, llm_providers)
- Tasks 18-24: Flatten 51 namespaces across 10 adapters for consistency
- Dependency order: 15 → 16-17 → 18-24

**Next Phase**: Phase 2.3.4: Async/Coroutine Support → 2.3.5: Lua Standard Library

**Completed**:
- ✅ Phase 1: Engine & Bridge Foundation (38+ bridges)
- ✅ Phase 2.1-2.3.2: Lua Core, Module System, ScriptValue, Test Utilities

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
/pkg/testutils/  # Centralized test utilities
```

## Implementation Workflow

1. **Be thorough** - No shortcuts or deferrals. Ask questions when needed
2. **TDD mandatory** - Write tests first, use testutils
3. **Bridge-first** - Wrap go-llms, never reimplement  
4. **Research go-llms** - Check git submodule first
5. **Run `make all`** - Complete dev cycle
6. **Update TODO.md** - Mark tasks with timestamps

## Commands

```bash
make all   # fmt, vet, lint, test, build
make test  # Test with race detection  
make lint  # Check code quality
```

## Key Reminders

- **Complete tasks fully** - No lazy implementations or deferrals
- Do what's asked; nothing more, nothing less
- Prefer editing existing files over creating new ones
- Never create docs unless explicitly requested
- If it's in go-llms, bridge it - don't reimplement