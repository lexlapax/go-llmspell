# CLAUDE.md

go-llmspell: **Scriptable LLM interactions** via Lua, JavaScript, and Tengo. Bridges go-llms v0.3.5 to scripts without reimplementing features.

## Current Status

✅ **Phase 3: Spell Runner CLI COMPLETED [2025-06-21]**

**🎉 Phase 3 Successfully Delivered:**
- ✅ Complete CLI with 11 commands (run, repl, new, validate, config, security, engines, debug, version, completion, man)
- ✅ Interactive REPL with syntax highlighting and history
- ✅ Comprehensive template system for spell generation
- ✅ Three-tier security profiles (sandbox, development, production)
- ✅ Full documentation suite including man pages and shell completion
- ✅ Integration tests covering all functionality
- ✅ Production-ready CLI tool

**Next Phase:**
- 🔄 **Phase 4: JavaScript Engine Implementation** [READY TO START]
- Research goja integration and design ES6+ support
- Implement complete JavaScript engine with async/await
- Create JavaScript standard library bridging go-llms

**Completed Milestones**:
- ✅ Phase 1: Engine & Bridge Foundation (38+ bridges)
- ✅ Phase 2: Complete Lua Engine Implementation
  - Full Lua engine with async/coroutine support
  - Comprehensive Lua stdlib (18 modules) with complete test coverage
  - Development Tools: Debugger & Script Validator with 100% coverage
- ✅ Phase 3: Spell Runner CLI Implementation
  - Complete command-line interface with 11 commands
  - Interactive REPL, debugger, and template system
  - Comprehensive documentation and testing infrastructure

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

## Phase 4 Focus: JavaScript Engine Implementation

Next major milestone: Implement complete JavaScript engine support with ES6+ features and comprehensive standard library bridging go-llms functionality.

## Key Reminders

- **Complete tasks fully** - No lazy implementations or deferrals
- Do what's asked; nothing more, nothing less
- Prefer editing existing files over creating new ones
- Never create docs unless explicitly requested
- If it's in go-llms, bridge it - don't reimplement
- **Phase 3 Ready**: Lua engine is production-ready for CLI integration