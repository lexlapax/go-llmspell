# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 2.3.2.5 COMPLETE** [2025-06-19]: Test Utilities Extraction - ALL SUCCESS METRICS ACHIEVED  
🚧 **Phase 2.3.3 ACTIVE**: Bridge Adapters (2 of 14 completed)

- ScriptValue system complete: All 21 bridges converted with type safety
- Test infrastructure complete: Centralized testutils, 40%+ code reduction achieved
- Foundation established: Ready for bridge adapter development phase

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
/pkg/testutils/  # Centralized test utilities (NEW)
```

## Implementation Workflow

1. **TDD mandatory** - Write tests first, use testutils for new tests
2. **Bridge-first** - Wrap go-llms, never reimplement  
3. **Research go-llms** - Check implementation in git submodule first
4. **Reuse code** - Use pkg/testutils, reduce duplication
5. **Run `make all`** - Complete dev cycle
6. **Update TODO.md** - Mark tasks with timestamps

## Phase 2.3.2.5: Test Utilities Extraction - COMPLETE [2025-06-19]

**All 6 phases completed with all success metrics achieved**:
- ✅ **Phases 1-4**: Core infrastructure, helpers, package migrations 
- ✅ **Phase 5**: Advanced helpers (table/context/numeric) + GopherLua migration
- ✅ **Phase 6**: Final cleanup, comprehensive documentation, metrics verification

**Success Metrics Achieved**:
- Code reduction: 40%+ (956+ ScriptValue + 363+ duplicate lines removed)
- Test quality: 98.5% pass rate, zero race conditions
- Foundation: Complete testutils package with migration patterns established

## Commands

```bash
make all   # fmt, vet, lint, test, build
make test  # Test with race detection  
make lint  # Check code quality
```

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.