# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

🚧 **Phase 2.3.4 ACTIVE**: Lua Standard Library - Starting development

**Completed Phases**:
- ✅ Phase 2.3.2.5: Test Utilities (40%+ code reduction, centralized testutils)
- ✅ Phase 2.3.3: Bridge Adapters (14/14 adapters with full test coverage)

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
/pkg/testutils/  # Centralized test utilities
```

## Implementation Workflow

1. **TDD mandatory** - Write tests first, use testutils for new tests
2. **Bridge-first** - Wrap go-llms, never reimplement  
3. **Research go-llms** - Check implementation in git submodule first
4. **Reuse code** - Use pkg/testutils, reduce duplication
5. **Run `make all`** - Complete dev cycle
6. **Update TODO.md** - Mark tasks with timestamps

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