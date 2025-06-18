# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 2.3.2 COMPLETE** [2025-12-19]: ScriptValue system + test fixes  
🚧 **Phase 2.3.2.5 ACTIVE**: Test utilities extraction - Consolidating 56 test files

- All 21 bridges converted to ScriptValue with type safety
- Fixed workflow bridge, deadlocks, JSON type assertions
- Starting 6-week test utilities extraction (30-40% code reduction target)

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

## Current Task: Test Utilities Extraction

**Week 1 Goals**:
- Create `/pkg/testutils` structure
- Consolidate 4+ mock engine implementations → `mock_engine.go`
- Extract 3+ mock bridge patterns → `mock_bridges.go`
- Move existing `scriptvalue_helpers.go`
- Add comprehensive tests for mocks

**Key Patterns to Extract**:
- Mock engines (12+ files)
- Bridge setup/teardown (20+ files)
- ScriptValue creation (30+ files)
- Type assertions (25+ files)
- Table test structures (40+ files)

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