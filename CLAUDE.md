# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

🚧 **Phase 2.4.4: Production Readiness - Comprehensive Testing** [IN PROGRESS - 2025-06-23]

**✅ Completed:**
- Task 2.4.5.1: CODE documentation [COMPLETED - 2025-06-22]
- Task 2.4.5.2: User Guide & Example Spells [COMPLETED - 2025-06-22]
- Task 2.4.4.1: 90%+ test coverage - comprehensive_test.go created [COMPLETED - 2025-06-22]
- Task 2.4.4.1: Integration test suite - integration_test.go created [COMPLETED - 2025-06-22]

**🔄 Current Task:**
- Task 2.4.4.1: Create stress tests (tests/stress/)
- Task 2.4.4.1: Implement chaos testing 
- Task 2.4.4.1: Add regression test suite

## Architecture

**Fundamental Rule**: If it's not in go-llms, we don't implement it.

```
/pkg/engine/     # Script engine interfaces (our code)
/pkg/bridge/     # Thin wrappers around go-llms (no business logic)
/cmd/llmspell/   # CLI implementation
```

## Implementation Workflow

1. **Be thorough** - Complete tasks fully
2. **TDD mandatory** - Write tests first
3. **Bridge-first** - Wrap go-llms, never reimplement  
4. **Use TODO.md** - Track progress with timestamps

## Key Reminders

- Do what's asked; nothing more, nothing less
- Prefer editing existing files over creating new ones
- If it's in go-llms, bridge it - don't reimplement