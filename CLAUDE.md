# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

🚧 **Phase 3.2: Core CLI Implementation ACTIVE**
- ✅ Task 3.2.1: Configuration Foundation [COMPLETED - 2025-06-20]
  - Koanf v2 integration with layered config (defaults → file → env → flags)
  - Thread-safe implementation with mutex protection
  - Fixed all race conditions in tests
  - Comprehensive test coverage, lint-clean
- Next: Task 3.2.2: Error Handling Infrastructure

**Completed Milestones**:
- ✅ Phase 1: Engine & Bridge Foundation (38+ bridges)
- ✅ Phase 2: Complete Lua Engine Implementation
  - Full Lua engine with async/coroutine support
  - Comprehensive Lua stdlib (18 modules) with complete test coverage
  - Development Tools: Debugger & Script Validator with 100% coverage
- ✅ Phase 3.1: Spell Runner Research & Planning

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
/pkg/testutils/  # Centralized test utilities
/cmd/llmspell/   # CLI implementation (Phase 3)
/pkg/config/     # Configuration management (Phase 3)
/pkg/runner/     # Script execution runtime (Phase 3)
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

## Phase 3 Focus: Spell Runner CLI

**Priority Order** (from TODO.md):
1. **3.2.1**: Configuration Foundation (`/pkg/config/`) - Koanf v2, YAML, env vars
2. **3.2.2**: Error Handling Infrastructure (`/pkg/errors/`) - CLI exit codes, formatting
3. **3.2.3**: Core Runner Package (`/pkg/runner/`) - Engine registry, spell loading
4. **3.2.4**: Security & Validation Integration (`/pkg/security/`) - Profiles, validation
5. **3.2.5**: CLI Structure with Kong (`/cmd/llmspell/`) - Commands, flags, help
6. **3.2.6**: Debug Command Implementation - Integrate existing debugger
7. **3.2.7**: REPL Implementation (`/pkg/repl/`) - Interactive shell
8. **3.2.8**: Template & Utilities (`/pkg/template/`) - Spell scaffolding
9. **3.2.9**: Testing & Integration (`/test/integration/`) - End-to-end tests

## Key Reminders

- **Complete tasks fully** - No lazy implementations or deferrals
- Do what's asked; nothing more, nothing less
- Prefer editing existing files over creating new ones
- Never create docs unless explicitly requested
- If it's in go-llms, bridge it - don't reimplement
- **Phase 3 Ready**: Lua engine is production-ready for CLI integration