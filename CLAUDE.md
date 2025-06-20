# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

🚧 **Phase 2.4.1 ACTIVE**: Performance Optimization (4/4 tasks complete)
- ✅ Tasks 1-4: Profiling, Type Conversion, State Pool, Script Compilation complete
- All optimization features implemented with comprehensive testing
- Next: Task 2.4.2: Spell Runner command line

**Next Phase**: Phase 2.4: Advanced Features & Optimization

**Completed**:
- ✅ Phase 1: Engine & Bridge Foundation (38+ bridges)
- ✅ Phase 2.1-2.3.4: Full Lua engine with async/coroutine support
- ✅ Phase 2.3.5: Comprehensive Lua stdlib (Promise, LLM, Agent, State, Events, Auth, Data, Observability, Tools, Errors, Logging) with complete test coverage
- ✅ Phase 2.4.1.1-4: Performance optimizations (Profiling, Type Conversion, State Pool, Script Compilation)

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