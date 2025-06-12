# TODO: Go-LLMSpell Bridge-First Implementation

## Overview
Based on the bridge-first architecture in `docs/MIGRATION_PLAN_V0.3.3.md`, this TODO focuses on **bridging existing go-llms functionality** rather than reimplementing features. Our value is making go-llms scriptable through Lua, JavaScript, and Tengo.

## Key Principles
1. **Bridge, Don't Build**: If it exists in go-llms, we bridge it
2. **Clean Architecture**: Just `pkg/engine/` and `pkg/bridge/` - no agents, no core
3. **Script Engines Execute**: Engines handle script execution directly
4. **Type Safety**: Maintain type conversions at bridge boundaries

## Migration Status
- ✅ Updated go-llms submodule to v0.3.3
- ✅ Phase 1.1: Script Engine Interface [COMPLETED]
- ❌ Phase 1.2: ~~Script Execution Agent System~~ [DELETE - Architectural mismatch]
- 🚧 Migration to bridge-first architecture in progress

---


## Phase 1: Engine and Bridge Foundation

### ✅ 1.1 Script Engine Interface [COMPLETED]
- ✅ Core interfaces (ScriptEngine, Bridge, TypeConverter)
- ✅ Engine registry and discovery
- ✅ Type system foundation
- ✅ Bridge manager

### 1.2 State Bridge System
Bridge the comprehensive state management already in go-llms `/pkg/agent/core/state_manager.go`

- [ ] **Task 1.2.1: State Manager Bridge**
  - [ ] Create test file `/pkg/bridge/state_manager_test.go`
  - [ ] Test state lifecycle operations bridging
  - [ ] Test state transforms (filter, flatten, sanitize)
  - [ ] Test merge strategies (Last, MergeAll, Union)
  - [ ] Test snapshot and history operations
  - [ ] Create `/pkg/bridge/state_manager.go`
  - [ ] Bridge StateManager from go-llms
  - [ ] Expose state operations to scripts
  - [ ] Handle type conversions for state data
  - [ ] Support all merge strategies

- [ ] **Task 1.2.2: State Context Bridge**
  - [ ] Create test file `/pkg/bridge/state_context_test.go`
  - [ ] Test SharedStateContext bridging
  - [ ] Test parent-child state relationships
  - [ ] Test state isolation and sharing
  - [ ] Create `/pkg/bridge/state_context.go`
  - [ ] Bridge SharedStateContext from go-llms
  - [ ] Enable parent-child state sharing
  - [ ] Support state scoping for sub-agents

- [ ] **Task 1.2.3: State Persistence Bridge**
  - [ ] Create test file `/pkg/bridge/state_persistence_test.go`
  - [ ] Test persistence interface bridging
  - [ ] Test save/load operations
  - [ ] Test custom persistence implementations
  - [ ] Create `/pkg/bridge/state_persistence.go`
  - [ ] Bridge state persistence interface
  - [ ] Allow script-based persistence implementations
  - [ ] Support various storage backends from scripts

### 1.3 Core Bridge System

- [ ] **Task 1.3.1: LLM Agent Bridge** (CRITICAL - Replaces our agent duplication)
  - [ ] Create test file `/pkg/bridge/llm_agent_test.go`
  - [ ] Test go-llms agent creation and configuration
  - [ ] Test tool registration and execution
  - [ ] Test sub-agent orchestration
  - [ ] Test agent lifecycle hooks and events
  - [ ] Create `/pkg/bridge/llm_agent.go`
  - [ ] Bridge complete go-llms agent system from `/pkg/agent/`
  - [ ] Expose agent creation, configuration, and execution to scripts
  - [ ] Support tool registration and execution
  - [ ] Enable sub-agent orchestration
  - [ ] Bridge agent lifecycle hooks and events

- [ ] **Task 1.3.2: Workflow Engine Bridge**
  - [ ] Create test file `/pkg/bridge/workflow_engine_test.go`
  - [ ] Test workflow lifecycle bridging
  - [ ] Test all workflow types (sequential, parallel, conditional, loop)
  - [ ] Test workflow state and error handling
  - [ ] Create `/pkg/bridge/workflow_engine.go`
  - [ ] Bridge workflow system from `/pkg/agent/workflow/`
  - [ ] Expose workflow creation and execution
  - [ ] Support workflow composition from scripts

- [ ] **Task 1.3.3: Event System Bridge**
  - [ ] Create test file `/pkg/bridge/events_test.go`
  - [ ] Test event streaming to scripts
  - [ ] Test event filtering and subscription
  - [ ] Test all event types
  - [ ] Create `/pkg/bridge/events.go`
  - [ ] Bridge pkg/agent/domain event system
  - [ ] Support real-time event streaming to scripts
  - [ ] Enable event filtering and subscription by type
  - [ ] Handle lifecycle, execution, tool, and workflow events

- [ ] **Task 1.3.4: Tool System Bridge**
  - [ ] Create test file `/pkg/bridge/tools_test.go`
  - [ ] Test tool interface bridging
  - [ ] Test built-in tools exposure
  - [ ] Test tool registration and execution
  - [ ] Create `/pkg/bridge/tools.go`
  - [ ] Bridge pkg/agent/tools interfaces
  - [ ] Expose built-in tools from pkg/agent/builtins/tools
  - [ ] Support tool registration and execution
  - [ ] Enable tool composition and chaining

- [ ] **Task 1.3.5: Hook System Bridge**
  - [ ] Create test file `/pkg/bridge/hooks_test.go`
  - [ ] Test Hook interface bridging
  - [ ] Test all hook types (BeforeGenerate, AfterGenerate, etc.)
  - [ ] Test hook priority and chaining
  - [ ] Create `/pkg/bridge/hooks.go`
  - [ ] Bridge pkg/agent/domain Hook interface
  - [ ] Support all hook types with priority ordering
  - [ ] Enable script-based hook implementations
  - [ ] Integrate with built-in hooks

### 1.4 Additional Bridges

- [ ] **Task 1.4.1: Schema Bridge**
  - [ ] Create `/pkg/bridge/schema.go`
  - [ ] Bridge pkg/schema validation system
  - [ ] Expose reflection-based generation
  - [ ] Support custom validator registration

- [ ] **Task 1.4.2: Structured Output Bridge**
  - [ ] Create `/pkg/bridge/structured.go`
  - [ ] Bridge pkg/structured processing
  - [ ] Expose JSON extraction utilities
  - [ ] Support schema caching

- [ ] **Task 1.4.3: Tracing Bridge**
  - [ ] Create `/pkg/bridge/tracing.go`
  - [ ] Bridge core/tracing.go distributed tracing
  - [ ] Support OpenTelemetry integration
  - [ ] Enable trace correlation

- [ ] **Task 1.4.4: Guardrails Bridge**
  - [ ] Create `/pkg/bridge/guardrails.go`
  - [ ] Bridge guardrails.go safety system
  - [ ] Support content filtering
  - [ ] Enable behavioral constraints

- [ ] **Task 1.4.5: Metrics Bridge**
  - [ ] Create `/pkg/bridge/metrics.go`
  - [ ] Bridge performance metrics system
  - [ ] Support custom metric collection
  - [ ] Enable metric aggregation

- [ ] **Task 1.4.6: Memory Bridge** ⏸️ **[DEFERRED - Not in go-llms yet]**
  - [ ] Will implement when available in go-llms

- [ ] **Task 1.4.7: Conversation Bridge** ⏸️ **[DEFERRED - Not in go-llms yet]**
  - [ ] Will implement when available in go-llms

---

## Phase 2: Lua Engine Implementation

### 2.1 Lua Engine Core
- [ ] **Task 2.1.1: Engine Implementation**
  - [ ] Create test file `/pkg/engine/lua/engine_test.go`
  - [ ] Test ScriptEngine interface implementation
  - [ ] Test GopherLua integration
  - [ ] Test resource limits enforcement
  - [ ] Create `/pkg/engine/lua/engine.go`
  - [ ] Implement ScriptEngine interface for Lua
  - [ ] Integrate GopherLua
  - [ ] Implement resource limits

- [ ] **Task 2.1.2: Type Converter**
  - [ ] Create test file `/pkg/engine/lua/converter_test.go`
  - [ ] Test Lua ↔ Go type conversions
  - [ ] Test Lua tables → Go maps/arrays
  - [ ] Create `/pkg/engine/lua/converter.go`
  - [ ] Implement type conversions
  - [ ] Optimize for performance

- [ ] **Task 2.1.3: Security Sandbox**
  - [ ] Create test file `/pkg/engine/lua/sandbox_test.go`
  - [ ] Test dangerous functions disabled
  - [ ] Test file system restrictions
  - [ ] Create `/pkg/engine/lua/sandbox.go`
  - [ ] Disable dangerous Lua functions
  - [ ] Implement restrictions

### 2.2 Lua Standard Library
- [ ] **Task 2.2.1: Core Modules**
  - [ ] Create `/pkg/engine/lua/stdlib/core.lua`
  - [ ] Create `/pkg/engine/lua/stdlib/llm.lua` - LLM bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/tools.lua` - Tools bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/workflow.lua` - Workflow bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/state.lua` - State bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/events.lua` - Events bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/hooks.lua` - Hooks bridge wrapper

---

## Phase 3: JavaScript Engine Implementation

### 3.1 JavaScript Engine Core
- [ ] **Task 3.1.1: Engine Implementation**
  - [ ] Create test file `/pkg/engine/javascript/engine_test.go`
  - [ ] Test ScriptEngine interface implementation
  - [ ] Test Goja integration
  - [ ] Test ES6+ support
  - [ ] Create `/pkg/engine/javascript/engine.go`
  - [ ] Implement ScriptEngine interface for JS
  - [ ] Integrate Goja
  - [ ] Add ES6+ support

- [ ] **Task 3.1.2: Type Converter**
  - [ ] Create test file `/pkg/engine/javascript/converter_test.go`
  - [ ] Test JS ↔ Go type conversions
  - [ ] Test Promise handling
  - [ ] Create `/pkg/engine/javascript/converter.go`
  - [ ] Implement type conversions
  - [ ] Handle async patterns

- [ ] **Task 3.1.3: Security Sandbox**
  - [ ] Create test file `/pkg/engine/javascript/sandbox_test.go`
  - [ ] Test global access restrictions
  - [ ] Test resource limits
  - [ ] Create `/pkg/engine/javascript/sandbox.go`
  - [ ] Restrict global access
  - [ ] Implement CSP-like policies

### 3.2 JavaScript Standard Library
- [ ] **Task 3.2.1: Core Modules**
  - [ ] Create `/pkg/engine/javascript/stdlib/core.js`
  - [ ] Create `/pkg/engine/javascript/stdlib/llm.js` - LLM bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/tools.js` - Tools bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/workflow.js` - Workflow bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/state.js` - State bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/events.js` - Events bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/hooks.js` - Hooks bridge wrapper

---

## Phase 4: Tengo Engine Implementation

### 4.1 Tengo Engine Core
- [ ] **Task 4.1.1: Engine Implementation**
  - [ ] Create `/pkg/engine/tengo/engine.go`
  - [ ] Implement ScriptEngine interface for Tengo
  - [ ] Integrate Tengo VM
  - [ ] Optimize for performance

- [ ] **Task 4.1.2: Type Converter**
  - [ ] Create `/pkg/engine/tengo/converter.go`
  - [ ] Implement Tengo ↔ Go conversions
  - [ ] Handle Tengo objects

- [ ] **Task 4.1.3: Security Sandbox**
  - [ ] Create `/pkg/engine/tengo/sandbox.go`
  - [ ] Implement Tengo restrictions
  - [ ] Add import controls

---

## Phase 5: Integration and Examples

### 5.1 Example Spells
- [ ] **Task 5.1.1: Basic Examples**
  - [ ] Hello World spell (all engines)
  - [ ] LLM chat spell
  - [ ] Tool usage spell
  - [ ] State management spell

- [ ] **Task 5.1.2: Advanced Examples**
  - [ ] Multi-agent orchestration spell
  - [ ] Complex workflow spell
  - [ ] Event-driven spell
  - [ ] Hook-based customization spell

### 5.2 Testing
- [ ] **Task 5.2.1: Cross-Engine Tests**
  - [ ] Create conformance test suite
  - [ ] Verify API compatibility
  - [ ] Test performance characteristics

- [ ] **Task 5.2.2: Integration Tests**
  - [ ] Test bridge functionality
  - [ ] Test type conversions
  - [ ] Test error handling

---

## Documentation

### API Documentation
- [ ] Bridge API reference
- [ ] Engine-specific features
- [ ] Type conversion guide

### User Guides
- [ ] Getting started guide
- [ ] Migration from direct go-llms usage
- [ ] Best practices

### Tutorials
- [ ] First spell tutorial
- [ ] Using go-llms agents from scripts
- [ ] Building workflows in scripts

---

## Success Metrics

### Development
- [ ] Zero duplicate implementations of go-llms features
- [ ] Clean two-package architecture maintained
- [ ] All bridges properly tested

### Performance
- [ ] < 5% overhead from bridging
- [ ] Type conversions optimized
- [ ] Memory usage minimal

### Adoption
- [ ] Clear examples for all major features
- [ ] Comprehensive documentation
- [ ] Easy migration path

---

## Notes

### Development Order
1. Delete `/pkg/core/agent/` immediately
2. Implement core bridges (state, llm_agent, workflow)
3. Complete Lua engine and stdlib
4. Add JavaScript engine
5. Add Tengo engine
6. Create comprehensive examples

### Testing Strategy
- TDD for all new code
- Test bridges thoroughly
- Cross-engine conformance tests
- Performance benchmarks

### What We DON'T Build
- ❌ Any agent abstractions (scripts execute directly)
- ❌ State management (bridge to go-llms)
- ❌ Workflow engine (bridge to go-llms)
- ❌ Event system (bridge to go-llms)
- ❌ Complex abstractions (keep it simple)

### What We DO Build
- ✅ Script engines (Lua, JS, Tengo)
- ✅ Type converters
- ✅ Bridges to go-llms
- ✅ Language-specific stdlib
- ✅ Examples and documentation

---

**Remember**: If it exists in go-llms, we bridge it. We only build what's unique to our scripting layer.