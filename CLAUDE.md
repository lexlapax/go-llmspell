# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

🚧 **Phase 2.3.5 ACTIVE**: Lua Standard Library Implementation (5/18 tasks complete)
- ✅ Tasks 1-5: Promise & Async, LLM Operations, Agent Management, State Management, Event & Hooks Libraries complete
- Built with comprehensive testing, clean Lua linting (0 warnings), and mock bridge system
- Next: Task 2.3.5.6: Structured Data Library

**Next Phase**: Phase 2.4: Advanced Features & Optimization

**Completed**:
- ✅ Phase 1: Engine & Bridge Foundation (38+ bridges)
- ✅ Phase 2.1-2.3.4: Full Lua engine with async/coroutine support
- ✅ Phase 2.3.5.1-5: Full Lua stdlib foundation (Promise, LLM, Agent, State, Events) with complete test coverage

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