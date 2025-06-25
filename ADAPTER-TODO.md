# ADAPTER-TODO.md: Fix Adapter Integration in Execution Path

## Goal
Ensure all Lua scripts follow the proper execution path: **Lua → Adapter → Bridge → go-llms**

## Current Problem
- BridgeManager.CreateLuaModule() creates modules directly from bridges
- Adapters are implemented but completely bypassed
- Current flow: Lua → Bridge → go-llms (adapter skipped)

## Architecture Decision
Make each engine responsible for creating/managing adapters when bridges are registered. This maintains clean separation while ensuring adapters are in the execution path.

**MAJOR UPDATE [2025-06-25]**: The adapter factory pattern was **abandoned due to unsolvable import cycles**. Replaced with a **callback pattern** to achieve the same goal without cycles:
- Engine has `SetAdapterCreator(func(bridgeID string, bridge engine.Bridge) (interface{}, error))`
- External code provides adapter creation logic via callback
- No factory class needed, no import cycles

**ARCHITECTURE BREAKTHROUGH [2025-06-25]**: After multiple failed attempts, **RADICAL PACKAGE RESTRUCTURE SOLVES ALL IMPORT CYCLES**:
- Complete reorganization of `pkg/engine/gopherlua/` into specialized subpackages
- **FACTORY PATTERN RESTORED** with zero import cycles through proper separation of concerns
- Static, type-safe adapter creation with clean dependency chain
- Separation of base classes, concrete implementations, converters, and factory logic

## Complex Architectural Considerations

### Multi-Bridge Adapter Dependencies
**Discovery**: Some adapters require multiple bridges with complex constructor signatures:
- `LLMAdapter(bridge, providersBridge, poolBridge)` - needs 3 bridges
- `ObservabilityAdapter(bridge, tracingBridge, guardrailsBridge)` - needs 3 bridges  
- `UtilsAdapter(bridge, authBridge, debugBridge, ...)` - needs 8+ bridges

**Challenges**:
1. **Bridge Registration Timing**: Bridges register independently, but adapters need multiple bridges
2. **Partial Dependencies**: Some adapters work with subset of bridges (graceful degradation)
3. **Error Handling**: Clear errors when required bridges are missing
4. **Bridge Grouping**: Which bridge IDs belong to the same adapter?

**Factory Pattern Solution**:
- AdapterFactory.CreateAdapter(bridgeID, bridgeMap) handles complex constructors
- Static method mapping with clear type signatures
- Bridge dependency resolution at factory level
- Graceful fallbacks for optional bridges (nil parameters)
- Clear error messages for missing required bridges

### Radical Package Restructure - Factory Pattern
**BREAKTHROUGH**: Complete reorganization eliminates all import cycles through separation of concerns

**New Package Structure:**
```
pkg/engine/lua/                         # Main engine (LuaEngine, LuaEngineFactory, BridgeManager)
pkg/engine/lua/converters/              # Type conversion utilities (Go ↔ Lua)
pkg/engine/lua/adapters/                # Base classes (BridgeAdapter, interfaces)
pkg/engine/lua/adapters/impl/           # Concrete adapter implementations
pkg/engine/lua/factory/                 # AdapterFactory (bridge-to-adapter mapping)
pkg/engine/lua/stdlib/                  # Stdlib modules (.lua files)
```

**Perfect Import Flow (NO CYCLES):**
```
lua/ → lua/factory/ + lua/stdlib/ ✅ (engine uses factory and stdlib)
lua/factory/ → lua/adapters/impl/ ✅ (factory creates concrete adapters)
lua/adapters/impl/ → lua/adapters/ ✅ (concrete adapters extend base classes)
lua/adapters/ → lua/converters/ + pkg/engine/ ✅ (base uses converters and Bridge interface)
lua/converters/ → pkg/engine/ ✅ (converters use core interfaces only)
```
**Result**: Perfect linear dependency chain - ZERO CYCLES!

## Tasks

### Phase 0: Radical Package Restructure **PREREQUISITE FOR ALL OTHER PHASES**

**CRITICAL**: Must be completed FIRST - all subsequent phases depend on new structure

- [x] 0.1 **Create new package directories** [COMPLETED - 2025-06-25]
  - [x] **CREATE DIR**: `pkg/engine/lua/`
  - [x] **CREATE DIR**: `pkg/engine/lua/converters/`
  - [x] **CREATE DIR**: `pkg/engine/lua/adapters/`
  - [x] **CREATE DIR**: `pkg/engine/lua/adapters/impl/`
  - [x] **CREATE DIR**: `pkg/engine/lua/factory/`
  - [x] **CREATE DIR**: `pkg/engine/lua/stdlib/`

- [x] 0.2 **Move files to new structure** (maintain git history with `git mv`) [COMPLETED - 2025-06-25]
  - [x] **TO `lua/`**: `engine*.go`, `bridge_manager.go`, `modules*.go`, `async*.go`, `pool*.go`, `security*.go`, `health.go`, `profiling.go`, `debug.go`, `validator.go`, `test_helpers.go`
  - [x] **TO `lua/converters/`**: All `converter*.go` files, `chunkcache.go`, `compilation_optimization.go`
  - [x] **TO `lua/adapters/`**: `bridge_adapter.go` (base class)
  - [x] **TO `lua/adapters/impl/`**: All files from `pkg/engine/gopherlua/adapters/` directory
  - [x] **TO `lua/stdlib/`**: All files from `pkg/engine/gopherlua/stdlib/` directory

- [x] 0.3 **Update package declarations in moved files** [COMPLETED - 2025-06-25]
  - [x] Update `package gopherlua` → `package lua` in main engine files
  - [x] Update `package adapters` → `package impl` in concrete adapter files
  - [x] Update `package gopherlua` → `package adapters` in bridge_adapter.go
  - [x] Update `package gopherlua` → `package converters` in converter files
  - [x] Update all internal imports to reflect new package structure

- [x] 0.4 **Fix imports throughout codebase** [COMPLETED - 2025-06-25]
  - [x] Update imports in `pkg/runner/`, `pkg/bridge/`, `cmd/` and all other packages
  - [x] Change `github.com/lexlapax/go-llmspell/pkg/engine/gopherlua` → `github.com/lexlapax/go-llmspell/pkg/engine/lua`
  - [x] Update imports in `pkg/engine/lua/adapters/*` to import `pkg/engine/lua/converters`
  - [x] Update adapter imports to `pkg/engine/lua/adapters/impl`
  - [x] Update imports in `tests/integration/*`
  - [x] Update imports in `tests/regression/*`
  - [x] Update imports in `tests/stress/*`
  - [x] Update imports in `example/integration/*`
  - [x] Verify all tests still compile and pass

- [x] 0.5 **Remove old gopherlua directory** [COMPLETED - 2025-06-25]
  - [x] Verify all files have been moved successfully
  - [x] Remove empty `pkg/engine/gopherlua/` directory
  - [x] Update any remaining references in documentation or config files

### Phase 1: Create Adapter Factory Infrastructure **DEPENDS ON PHASE 0**

**Prerequisites**: Phase 0 (restructure) must be completed first

- [ ] 1.1 **Implement adapter factory** (`pkg/engine/lua/factory/adapter_factory.go`)
  - [ ] Create AdapterFactory struct
  - [ ] Implement CreateAdapter(bridgeID, bridgeMap) method
  - [ ] Handle all bridge IDs from ARCHITECTURE_ANALYSIS.md
  - [ ] Handle complex multi-bridge constructors (LLM, Observability, Utils)
  - [ ] Support graceful fallbacks for optional bridges
  - [ ] Clear error messages for unknown/missing bridges

- [ ] 1.2 **Write factory tests** (`pkg/engine/lua/factory/adapter_factory_test.go`)
  - [ ] Test all single-bridge adapters (StateAdapter, AgentAdapter, etc.)
  - [ ] Test multi-bridge adapters with full dependencies
  - [ ] Test multi-bridge adapters with partial dependencies (optional bridges nil)
  - [ ] Test error handling for unknown bridge IDs
  - [ ] Test error handling for missing required bridges
  - [ ] Test adapter interface compliance

- [ ] 1.3 **Verify no import cycles**
  - [ ] Run `go mod tidy` and check for import cycle errors
  - [ ] Test compilation of all packages
  - [ ] Verify clean dependency chain: lua/ → factory/ → adapters/impl/ → adapters/ → converters/

~~- [x] 1.4 **Fix existing adapter tests** - OBSOLETE (moved to Phase 0)~~

### Phase 2: Update Callback Infrastructure to Factory Pattern **MAJOR REVISION NEEDED**

**Status**: Existing callback infrastructure must be REPLACED with factory pattern integration

**UNDO PREVIOUS WORK**: Remove callback-based adapter management completely
- [UNDO] 2.1 Remove temporary callback-based tests (`pkg/engine/lua/engine_adapter_test.go` - formerly gopherlua)
  - [ ] Remove all SetAdapterCreator() related tests
  - [ ] Remove mockAdapter callback tests
  - [ ] Keep adapter storage/retrieval tests (GetAdapter, adapter map)
  - [ ] Remove callback error handling tests

- [UNDO] 2.2 Remove callback fields from LuaEngine **COMPLETE REMOVAL**
  - [ ] **MODIFY FILE**: `pkg/engine/lua/engine.go`
    - [ ] Remove `adapterCreator func(bridgeID string, bridge engine.Bridge) (interface{}, error)` field
    - [ ] Remove `SetAdapterCreator()` method completely
    - [ ] Keep `adapters map[string]interface{}` and `adapterMu sync.RWMutex` 
    - [ ] Keep `GetAdapter(bridgeID string) interface{}` method
    - [ ] Keep adapter management initialization in NewLuaEngine()

- [ ] 2.3 **NEW**: Integrate AdapterFactory with LuaEngine (depends on Phase 1)
  - [ ] **MODIFY FILE**: `pkg/engine/lua/engine.go`
    - [ ] Add import: `"github.com/lexlapax/go-llmspell/pkg/engine/lua/factory"`
    - [ ] Add `adapterFactory *factory.AdapterFactory` field to LuaEngine
    - [ ] Add `bridgeMap map[string]engine.Bridge` field for factory use
    - [ ] Initialize factory in NewLuaEngine(): `adapterFactory: factory.NewAdapterFactory()`
    - [ ] Modify RegisterBridge() to use factory.CreateAdapter(bridgeID, bridgeMap)
  - [ ] **WRITE TESTS**: `pkg/engine/lua/engine_adapter_integration_test.go` (NEW FILE)
    - [ ] Test engine creates real adapters using factory
    - [ ] Test bridge map maintenance during registration
    - [ ] Test adapter creation with complex multi-bridge dependencies

### Phase 3: Verify BridgeManager Integration **NEEDS VERIFICATION AFTER RESTRUCTURE**

**Status**: BridgeManager adapter integration completed but needs verification in new package structure

**Architecture Note**: BridgeManager uses ModuleCreator callback to get Lua modules from adapters. Factory pattern provides adapters, ModuleCreator interface remains unchanged.

- [VERIFY] 3.1 Verify adapter-based module creation tests (`pkg/engine/lua/engine_bridge_test.go` - moved from gopherlua)
  - [ ] Verify tests still pass after package restructure
  - [ ] Update import paths to new structure
  - [ ] Test module creation via adapter callback still works
  - [ ] Test error when no module creator configured still enforced

- [VERIFY] 3.2 Verify BridgeManager adapter integration **NEEDS PATH UPDATES**
  - [ ] **VERIFY FILE**: `pkg/engine/lua/engine_bridge.go` (moved from gopherlua)
    - [ ] ModuleCreator func type: `type ModuleCreator func(bridgeID string) (lua.LGFunction, error)` *(should be unchanged)*
    - [ ] moduleCreator field in BridgeManager *(should be unchanged)*
    - [ ] CreateLuaModule() uses moduleCreator callback *(should be unchanged)*
    - [ ] No fallback to direct bridge methods *(should be unchanged)*
  - [ ] Update import paths to new package structure

- [VERIFY] 3.3 Verify BridgeManager initialization **UPDATE PATHS ONLY**
  - [ ] **VERIFY FILE**: `pkg/engine/lua/engine_bridge.go`
    - [ ] `NewBridgeManagerWithModuleCreator()` constructor exists *(should be unchanged)*
    - [ ] Existing `NewBridgeManager()` for backward compatibility *(should be unchanged)*
  - [ ] **VERIFY FILE**: `pkg/engine/lua/engine.go`  
    - [ ] Engine passes closure that accesses GetAdapter() method *(should be unchanged)*

- [ ] 3.4 **UPDATE**: Run bridge manager tests after restructure
  - [ ] Verify all bridge manager tests pass with new package structure
  - [ ] Update test imports to new paths
  - [ ] Confirm `TestLuaEngine_BridgeIntegration` still fails correctly (proving adapter enforcement works)

### Phase 4: Complete Factory Integration **DEPENDS ON PHASES 0-2**

**Status**: Final integration of AdapterFactory with LuaEngine (replaces old Phase 3.5-3.7)

- [ ] 4.1 **Wire factory to engine initialization**
  - [ ] **MODIFY FILE**: `pkg/engine/lua/engine_factory.go`
    - [ ] Engine factory automatically includes AdapterFactory (no external setup needed)
    - [ ] Verify initialization flow: NewLuaEngine() → factory initialized automatically
    - [ ] Remove any obsolete callback-related code if still present
  - [ ] **WRITE TESTS**: `pkg/engine/lua/engine_factory_integration_test.go` (NEW FILE)
    - [ ] Test factory creates engine with working adapter factory
    - [ ] Test bridge registration creates real adapters (not mocks)
    - [ ] Test complete flow: EngineFactory.Create() → NewLuaEngine() → RegisterBridge() → CreateRealAdapter()

- [ ] 4.2 **Update bridge registration logic for complex multi-bridge scenarios**
  - [ ] **MODIFY FILE**: `pkg/engine/lua/engine.go`
    - [ ] Enhance RegisterBridge() to handle deferred adapter creation
    - [ ] Support multi-bridge adapters (wait until all required bridges available)
    - [ ] Handle factory errors (unknown bridge IDs, missing dependencies)
    - [ ] Test graceful fallbacks for optional bridge dependencies
  - [ ] **WRITE TESTS**: Add to `pkg/engine/lua/engine_adapter_integration_test.go`
    - [ ] Test single-bridge adapter creation (immediate)
    - [ ] Test multi-bridge adapter creation (deferred until all bridges available)
    - [ ] Test error handling for unknown bridge IDs
    - [ ] Test graceful fallbacks for optional bridge dependencies

- [ ] 4.3 **Verify complete adapter integration**
  - [ ] **RUN TESTS**: Verify all adapter integration tests pass
  - [ ] **VERIFY**: `TestLuaEngine_BridgeIntegration` now passes (was failing before)
  - [ ] **VERIFY**: All bridge modules use real adapters, not direct bridge methods
  - [ ] **VERIFY**: Complete execution path: Lua → Stdlib → Adapter → Bridge → go-llms

### Phase 5: End-to-End Integration and Testing **DEPENDS ON PHASES 0-4**

**Prerequisites**: Package restructure, factory implementation, and engine integration must be completed first

- [ ] 5.1 **Fix ExecutionPipeline and REPL integration**
  - [ ] **VERIFY FILES**: Update import paths in external packages
    - [ ] Update `pkg/runner/executor.go` imports from gopherlua → lua
    - [ ] Ensure ExecutionPipeline.loadBridgeModules() works with real adapters
    - [ ] Update REPL imports and verify LoadBridgeModulesIntoState() works
  - [ ] **FIX TESTS**: Update import paths in all test files
    - [ ] Fix `TestLuaEngine_BridgeIntegration` test (should pass once real adapters connected)  
    - [ ] Verify all engine integration tests pass with new structure

- [ ] 5.2 **Adapter interface standardization**
  - [ ] **AUDIT FILES**: Verify all adapters in `pkg/engine/lua/adapters/impl/`
    - [ ] Verify all adapters have CreateLuaModule() returning lua.LGFunction
    - [ ] Check adapter constructors match expected signatures
    - [ ] Ensure adapters handle bridge dependencies correctly
  - [ ] **STANDARDIZE**: Define common adapter interface if needed
    - [ ] Create adapter interface in `pkg/engine/lua/adapters/interface.go`
    - [ ] Ensure consistent error handling across adapters
    - [ ] Document adapter creation patterns

- [ ] 5.3 **Comprehensive end-to-end testing**
  - [ ] **WRITE TESTS**: `pkg/engine/lua/integration_test.go` (NEW FILE)
    - [ ] Test complete flow: Lua Script → Stdlib → Adapter → Bridge → go-llms
    - [ ] Test all adapters from ARCHITECTURE_ANALYSIS.md
    - [ ] Test error propagation through complete stack
    - [ ] Test adapter creation failures and missing bridge handling
  - [ ] **FUNCTIONAL TESTS**: Test real adapter functionality
    - [ ] Test state management via StateAdapter
    - [ ] Test agent creation via AgentAdapter  
    - [ ] Test LLM operations via LLMAdapter
    - [ ] Test events via EventsAdapter

### Phase 6: Fix Example Scripts and Stdlib **DEPENDS ON PHASES 0-5**

**Prerequisites**: Real adapters must be working in engine before example scripts can be tested

- [ ] 6.1 **Audit and fix stdlib files for new structure**
  - [ ] **VERIFY FILES**: Review all stdlib/*.lua files in `pkg/engine/lua/stdlib/`
    - [ ] Verify import paths updated to new package structure
    - [ ] Ensure all use bridges.bridge_id.method() pattern (dot notation)
    - [ ] Verify compatibility with factory-created adapter modules
  - [ ] **UPDATE TESTS**: Fix stdlib test import paths to new structure

- [ ] 6.2 **Test original failing examples**
  - [ ] **FIX**: `examples/spells/lua/09-state-management.lua`
    - [ ] Verify state.get(), state.set(), state.update() work via StateAdapter
    - [ ] Document resolution of original state management issues
  - [ ] **FIX**: `examples/spells/lua/13-agent-handoff.lua`
    - [ ] Verify agent creation and state sharing work via AgentAdapter
    - [ ] Document resolution of original agent handoff issues

- [ ] 6.3 **Test all example scripts with factory-created adapters**
  - [ ] **RUN**: All examples/spells/lua/*.lua scripts
  - [ ] Verify they work with adapter-based bridge modules
  - [ ] Fix any adapter-related issues discovered
  - [ ] Update examples to use new package structure if needed
  - [ ] Document any breaking changes

### Phase 7: Documentation and Final Verification

- [ ] 7.1 **Update architectural documentation**
  - [ ] Update ARCHITECTURE_ANALYSIS.md with new package structure
  - [ ] Document factory pattern and bridge ID mappings (NOT callbacks)
  - [ ] Update execution flow diagrams: Lua → Stdlib → Adapter → Bridge → go-llms
  - [ ] Document the package restructure solution and why it was necessary

- [ ] 7.2 **Update inline code documentation**
  - [ ] Document AdapterFactory functions and bridge ID mappings
  - [ ] Document LuaEngine adapter management (factory-based, NOT callbacks)
  - [ ] Document BridgeManager ModuleCreator callback pattern
  - [ ] Document the complete integration flow with new package structure

- [ ] 7.3 **Clean up obsolete code and files**
  - [ ] Remove obsolete callback infrastructure completely
  - [ ] Remove any debug prints or temporary code from restructure
  - [ ] Remove obsolete import references to old gopherlua package
  - [ ] Ensure consistent code style across all new packages

- [ ] 7.4 **Final verification and testing**
  - [ ] **RUN**: Complete test suite (unit, integration, adapter, engine, examples)
  - [ ] **VERIFY**: No import cycles in final implementation
  - [ ] **VERIFY**: Adapters are in execution path for ALL bridges
  - [ ] **VERIFY**: All original failing examples now work
  - [ ] **PERFORMANCE**: Test adapter overhead is minimal
  - [ ] Document any performance implications of new structure

## CRITICAL DISCOVERY: Major Architecture Gaps Identified

### What We Thought Was Done vs Reality:

**ASSUMED COMPLETE**: Phases 1-3 seemed finished
**REALITY**: Critical integration layers are completely missing

### Key Problems Discovered:

1. **Engine Creates BridgeManager Without Adapters**
   - `NewLuaEngine()` calls `NewBridgeManager(converter)` (no ModuleCreator)
   - This causes `LoadBridgeModules()` to fail at runtime
   - Test failure in `TestLuaEngine_BridgeIntegration` proves this

2. **Mock Adapters vs Real Adapters Disconnect** 
   - LuaEngine creates `mockAdapter` instances (not real adapters)
   - Real adapters exist in `/adapters/` but aren't used
   - No mapping from bridge IDs to real adapter constructors

3. **Execution Pipeline Broken**
   - `ExecutionPipeline.loadBridgeModules()` calls `LoadBridgeModules()` 
   - This fails because BridgeManager has no ModuleCreator
   - Scripts can't access bridge functionality

### Critical Missing Phases (COMPLETELY REVISED):
- **Phase 0**: Radical package restructure to eliminate import cycles (PREREQUISITE FOR ALL OTHER WORK)
- **Phase 1**: Create AdapterFactory in new `pkg/engine/lua/factory/` structure  
- **Phase 2**: Remove callback infrastructure, integrate factory with LuaEngine
- **Phase 4**: Complete factory integration and multi-bridge adapter support
- **Phase 5**: End-to-end testing and adapter interface standardization

### Next Steps (COMPLETELY REVISED):
1. **CRITICAL**: Complete Phase 0 (Package Restructure) - PREREQUISITE for everything else
2. **MUST complete Phases 0-2** before any advanced integration
3. **Phases 4-5** depend entirely on 0-2 being completed correctly  
4. **Phase 6** (example scripts) cannot work until factory integration is complete

### **IMMEDIATE NEXT TASK**: Phase 0.1 - Create new package directories

**BREAKTHROUGH**: Package restructure solves ALL import cycle issues while maintaining factory pattern!

## Success Criteria (UPDATED FOR NEW STRUCTURE)
- All Lua scripts follow: Lua → Stdlib → Adapter → Bridge → go-llms
- Every bridge has a corresponding adapter created via **AdapterFactory pattern** in `pkg/engine/lua/factory/`
- LuaEngine properly wires adapters to BridgeManager via ModuleCreator callback
- No direct bridge method calls from Lua (enforced by BridgeManager)  
- All tests pass (including currently failing TestLuaEngine_BridgeIntegration)
- All example scripts work correctly (09-state-management.lua, 13-agent-handoff.lua)
- ExecutionPipeline.loadBridgeModules() works with factory-created adapters
- REPL LoadBridgeModulesIntoState() works with factory-created adapters
- **ZERO import cycles** in final implementation (achieved through package restructure)
- **Perfect separation of concerns**: engine / converters / adapters / impl / factory / stdlib
- **Factory pattern** handles complex multi-bridge dependencies elegantly
- **Type-safe** adapter creation with clean linear dependency chain
- **Package restructure** maintains git history and minimizes breaking changes
- **All callback infrastructure removed** - pure factory pattern throughout