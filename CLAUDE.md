# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

🚧 **ADAPTER-TODO.md Phase 3: Verify BridgeManager Integration** [IN PROGRESS - 2025-06-25]

**✅ Completed:**
- Phase 0: Radical Package Restructure [COMPLETED - 2025-06-25]
- Phase 1: Adapter Factory Infrastructure [COMPLETED - 2025-06-25]  
- Phase 2: Factory Integration with LuaEngine [COMPLETED - 2025-06-25]
- Phase 2.5: Remove RegisterAsModule [COMPLETED - 2025-06-25]
- Phase 2.6: Move Factory to Intended Location [COMPLETED - 2025-06-25]

**🔄 Current Task:**
- Phase 3.1: Verify BridgeManager adapter integration after restructure
- Issue: Test failures suggest fundamental BridgeManager problems that need verification before proceeding to Phase 4 adapter integration tests

**📋 Analysis:**
- Jumped to Phase 4.1.2 integration tests before completing Phase 3 verification
- Discovered Lua state access issues (Lua script sees `bridges` but Go can't access it)
- Need to start at Phase 3.1 to verify basic BridgeManager functionality post-restructure

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