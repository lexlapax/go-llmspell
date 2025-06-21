# Phase 3.2 Fix Plan - Core CLI Implementation

## Current State Summary

### What's Working:
- ✅ Configuration package (`/pkg/config/`) - Complete with tests
- ✅ Error handling (`/pkg/errors/`) - Complete with tests
- ✅ Security profiles (`/pkg/security/`) - Complete with tests
- ✅ Validator interface (`/pkg/validator/`) - Complete with tests
- ✅ Basic CLI structure with Kong - Commands parse correctly

### What's Broken:
- ❌ Runner package has skipped tests (3 test files marked .skip)
- ❌ Using SimpleExecutor instead of real ScriptExecutor
- ❌ No actual engine integration in CLI
- ❌ Commands are stubs that don't execute scripts
- ❌ REPL package doesn't exist
- ❌ Template package doesn't exist

## Architecture Overview

```
User → CLI (Kong) → Commands → Runner → Engine Registry → Script Engine
                                  ↓
                            Security/Validator
```

### Component Responsibilities:

1. **CLI (`/cmd/llmspell/`)**: Parse commands, setup context, handle errors
2. **Commands (`/cmd/llmspell/commands/`)**: Execute specific actions
3. **Runner (`/pkg/runner/`)**: Orchestrate script execution
4. **Engine Registry**: Manage script engines (Lua, JS, Tengo)
5. **Security/Validator**: Enforce security profiles and validate scripts

## Fix Implementation Plan

### Phase 1: Fix Runner Package Tests (Task 3.2.5.1)

1. **Enable skipped test files**:
   ```bash
   mv pkg/runner/engine_registry_test.go.skip pkg/runner/engine_registry_test.go
   mv pkg/runner/engine_selector_test.go.skip pkg/runner/engine_selector_test.go
   mv pkg/runner/executor_test.go.skip pkg/runner/executor_test.go
   ```

2. **Fix test compilation issues**:
   - Update test files to work with current interfaces
   - Mock engine.Registry properly
   - Fix any API mismatches

3. **Clean up backup files**:
   ```bash
   rm pkg/runner/executor_simple.go.bak
   rm cmd/llmspell/commands/simple.go.bak
   ```

4. **Ensure all tests pass**:
   ```bash
   go test -v ./pkg/runner/...
   ```

### Phase 2: Wire Up Full Executor (Task 3.2.5.2)

**CRITICAL FINDING**: The LuaEngine doesn't have an EngineFactory implementation!

1. **Create LuaEngineFactory** (`/pkg/engine/gopherlua/engine_factory.go`):
   ```go
   type LuaEngineFactory struct {
       stateFactory *LStateFactory
   }
   
   func NewLuaEngineFactory() *LuaEngineFactory {
       config := DefaultFactoryConfig()
       return &LuaEngineFactory{
           stateFactory: NewLStateFactory(config),
       }
   }
   
   // Implement EngineFactory interface
   func (f *LuaEngineFactory) Create(config engine.EngineConfig) (engine.ScriptEngine, error) {
       return NewLuaEngine(), nil
   }
   
   func (f *LuaEngineFactory) Name() string { return "lua" }
   func (f *LuaEngineFactory) Version() string { return "5.1" }
   func (f *LuaEngineFactory) Description() string { return "Lua 5.1 script engine" }
   func (f *LuaEngineFactory) FileExtensions() []string { return []string{".lua"} }
   // ... etc
   ```

2. **Fix main.go engine registration**:
   ```go
   // Create registry with proper config
   registryConfig := engine.DefaultRegistryConfig()
   registry := engine.NewRegistry(registryConfig)
   
   // Register Lua engine factory
   luaFactory := gopherlua.NewLuaEngineFactory()
   if err := registry.Register(luaFactory); err != nil {
       // handle error
   }
   
   // Initialize registry
   if err := registry.Initialize(); err != nil {
       // handle error
   }
   ```

2. **Update RunCmd to use real executor**:
   ```go
   // Get registry from context
   engineRegistry := GetEngineRegistry(ctx).(*runner.EngineRegistryManager)
   
   // Create executor with real components
   selector := runner.NewEngineSelector(engineRegistry)
   executor := runner.NewScriptExecutor(runnerConfig, engineRegistry, selector)
   
   // Execute script
   result, err := executor.ExecuteFile(ctx, c.Script, params)
   ```

3. **Update ValidateCmd to use real validation**:
   - Use engine's actual validation capabilities
   - Check script syntax through engine
   - Validate against security profiles

4. **Create integration test**:
   ```go
   // test/integration/basic_execution_test.go
   func TestBasicLuaExecution(t *testing.T) {
       // Create test script
       // Run through CLI
       // Verify output
   }
   ```

### Phase 3: Complete Command Implementations

1. **Run Command**:
   - Execute scripts with real engines
   - Pass parameters correctly
   - Handle timeouts
   - Show progress if verbose

2. **Validate Command**:
   - Validate spell.yaml files
   - Validate script syntax
   - Check security constraints

3. **Engines Command**:
   - List registered engines from registry
   - Show engine capabilities
   - Display version info

4. **Config Command**:
   - Actually read/write config files
   - Use Koanf for layered configuration
   - Support get/set operations

5. **Security Command**:
   - Use real security profiles
   - Show actual permissions
   - Validate profile configurations

### Phase 4: Implement REPL (Task 3.2.6)

1. **Create `/pkg/repl/` package**:
   - `repl.go` - REPL interface
   - `base_repl.go` - Common functionality
   - `lua_repl.go` - Lua-specific REPL
   - Tests for each

2. **Integrate readline**:
   ```bash
   go get github.com/chzyer/readline
   ```

3. **Implement REPL features**:
   - History persistence
   - Auto-completion
   - Multi-line input
   - Engine switching

### Phase 5: Implement Debug Command (Task 3.2.7)

1. **Integrate with Lua debugger**:
   - Use existing `/pkg/engine/gopherlua/debug.go`
   - Support breakpoints
   - Variable inspection
   - Step execution

### Phase 6: Template Package (Task 3.2.8)

1. **Create `/pkg/template/` package**:
   - Spell scaffolding
   - Example templates
   - Project initialization

### Phase 7: Integration Tests (Task 3.2.9)

1. **Create `/test/integration/`**:
   - End-to-end CLI tests
   - Multi-engine tests
   - Error scenarios
   - Performance tests

## Testing Strategy

### For Each Component:
1. **Unit Tests**: Test individual functions
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test full CLI execution

### Test Coverage Requirements:
- Minimum 80% coverage per package
- All error paths tested
- All command variations tested

## Success Criteria

1. **Working CLI that actually executes scripts**:
   ```bash
   llmspell run examples/hello.lua
   # Should print: Hello from Lua!
   ```

2. **All tests passing**:
   ```bash
   go test -v ./...
   # No skipped tests, all green
   ```

3. **Commands work as documented**:
   - run: Executes scripts
   - validate: Validates scripts
   - engines: Lists engines
   - config: Manages configuration
   - security: Shows profiles
   - repl: Interactive mode
   - debug: Debug scripts

## Implementation Order

1. Fix runner tests (Priority 0)
2. Wire up full executor (Priority 1)
3. Complete run command (Priority 1)
4. Complete validate command (Priority 2)
5. Complete other commands (Priority 2)
6. Implement REPL (Priority 3)
7. Implement debug (Priority 3)
8. Create templates (Priority 4)
9. Add integration tests (Priority 4)

## Time Estimate

- Phase 1 (Fix tests): 2-3 hours
- Phase 2 (Wire executor): 3-4 hours
- Phase 3 (Commands): 2-3 hours
- Phase 4 (REPL): 3-4 hours
- Phase 5 (Debug): 2-3 hours
- Phase 6 (Templates): 2 hours
- Phase 7 (Integration): 3-4 hours

Total: ~20-25 hours of focused work