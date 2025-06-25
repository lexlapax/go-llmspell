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

## Tasks

### Phase 1: Create Adapter Factory Infrastructure ~~**ABANDONED - REPLACED BY CALLBACK PATTERN**~~

- [~~] 1.1 ~~Write tests for adapter factory~~ **REMOVED** - Import cycle made this impossible
- [~~] 1.2 ~~Implement adapter factory~~ **REMOVED** - Import cycle: engine→adapters→gopherlua→engine  
- [~~] 1.3 ~~Run adapter factory tests~~ **REMOVED**
- [x] 1.4 **Fix existing adapter tests** (`pkg/engine/gopherlua/adapters/adapters_test.go`)
  - [x] Remove broken `TestAdapterFactoryIntegration` test that depended on removed factory
  - [x] Remove unused imports (bridge/agent, bridge/llm, bridge/state)
  - [x] Verify remaining adapter tests still pass
  - [x] Existing adapter unit tests (TestAllAdaptersIntegration) work unchanged

**REPLACEMENT IMPLEMENTED**: Callback pattern in engine:
- `engine.SetAdapterCreator(func(bridgeID string, bridge engine.Bridge) (interface{}, error))`
- `engine.GetAdapter(bridgeID string) interface{}`
- External code provides adapter creation logic without import cycles

### Phase 2: Integrate Adapter Management into LuaEngine **COMPLETED [2025-06-25]**

- [x] 2.1 Write tests for adapter management in engine (`pkg/engine/gopherlua/engine_adapter_test.go`)
  - [x] Test adapter creation during bridge registration using callback pattern
  - [x] Test adapter retrieval by bridge ID  
  - [x] Test adapter lifecycle (cleanup)
  - [x] Test concurrent access to adapter map
  - [x] Test error handling when adapter creation fails
  - [x] Test behavior when no adapter creator is set

- [x] 2.2 Modify LuaEngine to manage adapters **USING CALLBACK PATTERN**
  - [x] Add `adapterCreator func(bridgeID string, bridge engine.Bridge) (interface{}, error)` field
  - [x] Add `adapters map[string]interface{}` and `adapterMu sync.RWMutex`
  - [x] Initialize adapter management in NewLuaEngine()
  - [x] Modify RegisterBridge() to create and store adapters via callback
  - [x] Add `GetAdapter(bridgeID string) interface{}` method
  - [x] Add `SetAdapterCreator()` method for dependency injection
  - [x] Fixed MockBridge RegisterWithEngine() infinite recursion bug

- [x] 2.3 Run engine adapter tests
  - [x] All tests pass
  - [x] Fixed test failures related to missing adapter creators

### Phase 3: Update BridgeManager to Use Adapters **COMPLETED [2025-06-25]**

**Architecture Note**: Using callback pattern to avoid import cycles. Lua BridgeManager will use a callback function to get Lua modules from adapters instead of creating them directly from bridges.

- [x] 3.1 Write tests for adapter-based module creation (`pkg/engine/gopherlua/engine_bridge_test.go`)
  - [x] Test module creation via adapter callback
  - [x] Test error handling when no adapter exists  
  - [x] Test adapter module integration
  - [x] Test error when no module creator configured (every bridge requires an adapter)

- [x] 3.2 Modify BridgeManager to use module creation callback **COMPLETED [2025-06-25]**
  - [x] Add ModuleCreator func type: `type ModuleCreator func(bridgeID string) (lua.LGFunction, error)`
  - [x] Add moduleCreator field to BridgeManager
  - [x] Modify CreateLuaModule() to use moduleCreator callback (required for all bridges)
  - [x] Remove fallback to direct bridge methods (every bridge requires an adapter)
  - [x] Ensure adapters are in execution path for ALL bridges

- [x] 3.3 Update BridgeManager initialization **COMPLETED [2025-06-25]**
  - [x] Add `NewBridgeManagerWithModuleCreator()` constructor
  - [x] Keep existing `NewBridgeManager()` for backward compatibility (but will fail at runtime without adapters)
  - [x] Engine will pass a closure that accesses its GetAdapter() method

- [x] 3.4 Run bridge manager tests **COMPLETED [2025-06-25]**
  - [x] New adapter tests pass
  - [x] Expected test failure: `TestLuaEngine_BridgeIntegration` fails because it doesn't use adapters
  - [x] This confirms our architecture enforcement is working correctly

### Phase 3.5: Create Real Adapter Factory **CRITICAL MISSING PHASE - NEXT TASK**

**Problem**: Engine creates mockAdapters, no mapping from bridge IDs to real adapters

- [ ] 3.5.1 Create adapter factory mapping (`pkg/engine/gopherlua/adapter_factory.go`)
  - [ ] Map bridge IDs to adapter constructors (e.g., "state_manager" → NewStateAdapter)
  - [ ] Handle multi-bridge adapters (LLM, Observability, Utils)
  - [ ] Include all adapters from ARCHITECTURE_ANALYSIS.md
  - [ ] Return proper error for unknown bridge IDs

- [ ] 3.5.2 Write tests for adapter factory (`pkg/engine/gopherlua/adapter_factory_test.go`)
  - [ ] Test all bridge ID mappings
  - [ ] Test error handling for unknown bridges
  - [ ] Test adapter creation with dependencies
  - [ ] Test adapter interface compliance

- [ ] 3.5.3 Replace mockAdapter with real adapter creation
  - [ ] Update engine_adapter_test_mock.go or remove it
  - [ ] Update LuaEngine.SetAdapterCreator to use real factory
  - [ ] Test with real adapters instead of mocks

### Phase 3.6: Connect LuaEngine to Real Adapters **CRITICAL MISSING PHASE**

**Problem**: Engine has adapter management but BridgeManager not connected to it

- [ ] 3.6.1 Wire LuaEngine adapter system to BridgeManager
  - [ ] Update NewLuaEngine() to create BridgeManager with ModuleCreator callback
  - [ ] ModuleCreator should call engine.GetAdapter(bridgeID).CreateLuaModule()
  - [ ] Ensure adapter creation happens during bridge registration

- [ ] 3.6.2 Update engine initialization flow
  - [ ] Set up real adapter creator during engine initialization
  - [ ] Ensure bridges create real adapters when registered
  - [ ] Test complete flow: RegisterBridge → CreateAdapter → ModuleCreator works

- [ ] 3.6.3 Fix LoadBridgeModules to use adapters
  - [ ] Ensure ExecutionPipeline.loadBridgeModules() works with adapters
  - [ ] Fix TestLuaEngine_BridgeIntegration test failure
  - [ ] Test REPL bridge loading (LoadBridgeModulesIntoState)

### Phase 3.7: Adapter Interface Verification **MISSING PHASE**

**Problem**: Need to ensure all adapters implement correct interface consistently

- [ ] 3.7.1 Audit all adapter implementations
  - [ ] Verify all adapters have CreateLuaModule() returning lua.LGFunction
  - [ ] Check adapter constructors match expected signatures
  - [ ] Ensure adapters handle bridge dependencies correctly

- [ ] 3.7.2 Standardize adapter interface
  - [ ] Define common Adapter interface if needed
  - [ ] Ensure consistent error handling across adapters  
  - [ ] Document adapter creation patterns

- [ ] 3.7.3 Test adapter compatibility
  - [ ] Test each adapter's CreateLuaModule() method
  - [ ] Verify Lua modules work correctly
  - [ ] Test error cases and edge conditions

### Phase 4: Complete End-to-End Integration Testing **NEXT PHASE AFTER 3.5-3.7**

**Prerequisites**: Phases 3.5-3.7 must be completed first

- [ ] 4.1 Write comprehensive integration tests (`pkg/engine/gopherlua/adapter_integration_test.go`)
  - [ ] Test complete flow: Lua Script → Stdlib → Adapter → Bridge → go-llms
  - [ ] Test engine initialization with real adapters
  - [ ] Test LoadBridgeModules with real adapters
  - [ ] Test ExecutionPipeline.loadBridgeModules() flow
  - [ ] Test REPL LoadBridgeModulesIntoState() flow

- [ ] 4.2 Test real adapter functionality 
  - [ ] Test state management via StateAdapter
  - [ ] Test agent creation via AgentAdapter
  - [ ] Test LLM operations via LLMAdapter
  - [ ] Test events via EventsAdapter
  - [ ] Test all adapters from ARCHITECTURE_ANALYSIS.md

- [ ] 4.3 Test error propagation through complete stack
  - [ ] Lua error → Adapter error → Bridge error
  - [ ] Bridge error → Adapter error → Lua error
  - [ ] Adapter creation failures
  - [ ] Missing bridge handling

### Phase 4.5: Fix Broken Engine Integration **CRITICAL**

**Problem**: Current engine tests fail because they don't use adapter pattern

- [ ] 4.5.1 Fix TestLuaEngine_BridgeIntegration
  - [ ] Update test to use real adapter pattern
  - [ ] Ensure LoadBridgeModules works in tests
  - [ ] Test with actual bridge registration and execution

- [ ] 4.5.2 Fix ExecutionPipeline integration
  - [ ] Ensure ep.bridges.LoadBridgeModules() works with adapters
  - [ ] Test script execution with bridge modules loaded
  - [ ] Verify bridge modules accessible in Lua scripts

- [ ] 4.5.3 Fix REPL integration
  - [ ] Ensure LoadBridgeModulesIntoState() works with adapters
  - [ ] Test REPL with bridge modules available
  - [ ] Verify persistent state includes bridge functionality

### Phase 5: Fix Stdlib Files and Test Example Scripts **DEPENDS ON PHASES 3.5-4.5**

**Prerequisites**: Real adapters must be working in engine before stdlib can be tested

- [ ] 5.1 Audit and fix stdlib files for adapter compatibility
  - [ ] Review all stdlib/*.lua files in pkg/engine/gopherlua/stdlib/
  - [ ] Ensure all use bridges.bridge_id.method() pattern (dot notation)
  - [ ] Update any files still using direct bridge access
  - [ ] Verify compatibility with adapter-created modules

- [ ] 5.2 Test 09-state-management.lua and 13-agent-handoff.lua
  - [ ] These were the original failing examples that started this effort
  - [ ] Verify state.get(), state.set(), state.update() work via StateAdapter
  - [ ] Verify agent creation and state sharing work via AgentAdapter
  - [ ] Document resolution of original issues

- [ ] 5.3 Test all example scripts with adapter integration
  - [ ] Run all examples/spells/lua/*.lua scripts
  - [ ] Verify they work with adapter-based bridge modules
  - [ ] Fix any adapter-related issues
  - [ ] Document any breaking changes

### Phase 6: Documentation and Final Verification

- [ ] 6.1 Update architectural documentation
  - [ ] Update ARCHITECTURE_ANALYSIS.md with callback pattern
  - [ ] Document adapter factory pattern and bridge ID mappings
  - [ ] Update execution flow diagrams: Lua → Stdlib → Adapter → Bridge → go-llms
  - [ ] Document the abandoned factory pattern and why callback pattern was chosen

- [ ] 6.2 Update inline code documentation
  - [ ] Document adapter factory functions and bridge ID mappings
  - [ ] Document LuaEngine adapter management (SetAdapterCreator, GetAdapter)
  - [ ] Document BridgeManager ModuleCreator callback pattern
  - [ ] Document the complete integration flow

- [ ] 6.3 Clean up temporary and obsolete code
  - [ ] Remove or update engine_adapter_test_mock.go
  - [ ] Remove any debug prints or temporary code
  - [ ] Remove commented old code from adapter integration work
  - [ ] Ensure consistent code style across all new files

- [ ] 6.4 Final verification and testing
  - [ ] Run complete test suite (unit, integration, adapter, engine, examples)
  - [ ] Verify no import cycles in final implementation
  - [ ] Verify adapters are in execution path for ALL bridges
  - [ ] Performance test: ensure adapter overhead is minimal
  - [ ] Document any performance implications

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

### Critical Missing Phases:
- **Phase 3.5**: Create real adapter factory with bridge ID mapping
- **Phase 3.6**: Connect LuaEngine adapter system to BridgeManager  
- **Phase 3.7**: Verify all adapters implement consistent interface
- **Phase 4.5**: Fix broken engine integration tests

### Next Steps:
1. **MUST complete Phases 3.5-3.7** before any integration testing
2. **Phase 4** depends entirely on 3.5-3.7 being done correctly
3. **Phase 5** (example scripts) cannot work until engine integration is fixed

## Success Criteria
- All Lua scripts follow: Lua → Stdlib → Adapter → Bridge → go-llms
- Every bridge has a corresponding adapter created via real adapter factory
- LuaEngine properly wires adapters to BridgeManager via ModuleCreator callback
- No direct bridge method calls from Lua (enforced by BridgeManager)
- All tests pass (including currently failing TestLuaEngine_BridgeIntegration)
- All example scripts work correctly (09-state-management.lua, 13-agent-handoff.lua)
- ExecutionPipeline.loadBridgeModules() works with real adapters
- REPL LoadBridgeModulesIntoState() works with real adapters
- No import cycles in final implementation
- Clean separation of concerns maintained