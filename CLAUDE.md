# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

🚧 **Active**: Phase 3.2.3 - Core Runner Package  
✅ **Completed**: Phase 1 (Engine & Bridge), Phase 2 (Lua Engine), Tasks 3.2.1-3.2.2

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
6. **Use and Update TODO.md** - Mark done tasks with timestamps, use to track project progression

## Commands

```bash
make all   # fmt, vet, lint, test, build
make test  # Test with race detection  
make lint  # Check code quality
```

## Phase 3 Focus: Spell Runner CLI

See TODO.md for detailed task list (3.2.1 through 3.2.9)

## Key Reminders

- **Complete tasks fully** - No lazy implementations or deferrals
- Do what's asked; nothing more, nothing less
- Prefer editing existing files over creating new ones
- Never create docs unless explicitly requested
- If it's in go-llms, bridge it - don't reimplement
- **Phase 3 Ready**: Lua engine is production-ready for CLI integration