# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 2.3.2 COMPLETE** [2025-12-19]: ScriptValue system + test fixes  
✅ **Phase 2.3.2.5 COMPLETE** [2025-12-19]: Test utilities extraction - Major progress

- All 21 bridges converted to ScriptValue with type safety
- Fixed workflow bridge, deadlocks, JSON type assertions
- Migrated 8 test files (305 ScriptValue replacements) in agent & llm packages
- Achieved significant code reduction through helper functions

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

**Completed Tasks**:
- ✅ Created `/pkg/testutils` with MockScriptEngine & MockBridge
- ✅ Migrated `/pkg/engine` tests (enhanced local test_helpers.go)
- ✅ Migrated `/pkg/bridge/agent` tests (5 files, 171 replacements)
- ✅ Migrated `/pkg/bridge/llm` tests (3 files, 134 replacements)
- ✅ Created sv(), svMap(), svArray() helpers in each package
- ✅ Removed duplicate MockEngine from llm package

**Remaining Tasks**:
- [ ] Migrate `/pkg/bridge/util` tests (8 files)
- [ ] Migrate `/pkg/bridge/observability` tests (3 files)
- [ ] Migrate `/pkg/bridge/structured` tests (1 file)
- [ ] Extract table test patterns
- [ ] Create advanced test utilities

**Key Findings**:
- Import cycle prevents direct testutils usage in engine package
- Helper functions pattern (sv, svMap, svArray) very effective
- Achieved ~30% code reduction in migrated files

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