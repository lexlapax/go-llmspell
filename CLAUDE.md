# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 2.3.2.5 COMPLETE** [2025-06-19]: Test utilities extraction - ALL PHASES COMPLETED  
🚧 **Phase 2.3.3 ACTIVE**: Bridge Adapters (2 of 14 completed)

- All 21 bridges converted to ScriptValue with type safety + ALL TEST FAILURES FIXED
- Complete state manager bridge implementation (missing ExecuteMethod cases added)
- ALL bridge packages migrated to testutils (956+ ScriptValue call reductions)
- Achieved 30-40% code reduction target with 100% test pass rate

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

## Completed: Test Utilities Extraction [2025-06-19]

**All phases completed successfully**:
- ✅ Created centralized `/pkg/testutils` with comprehensive mock implementations
- ✅ Migrated ALL engine package tests (enhanced local test_helpers.go)
- ✅ Migrated ALL bridge package tests (13 packages, 56+ test files)
- ✅ FIXED ALL bridge test failures including complete state manager implementation
- ✅ Total: 956+ ScriptValue call reductions across entire codebase
- ✅ Achieved 30-40% code reduction target with 100% test pass rate

**Key Technical Achievements**:
- State object preservation through __state custom fields
- Comprehensive ExecuteMethod implementations for all bridge operations
- Flexible type conversion handling (int/float64 JSON conversions)
- Import cycle resolution patterns for engine package constraints

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