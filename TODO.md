# TODO: Go-LLMSpell Bridge-First Implementation

## Overview
Based on the bridge-first architecture in `docs/MIGRATION_PLAN_V0.3.3.md`, this TODO focuses on **bridging existing go-llms functionality** rather than reimplementing features. Our value is making go-llms scriptable through Lua, JavaScript, and Tengo.

## Key Principles
1. **Fundamental Rule**: If it's not in go-llms, we don't implement it in go-llmspell
2. **Bridge, Don't Build**: We ONLY bridge existing go-llms functionality. Bridging also means imports from go-llms and implementing the bridge function calls in the bridge.
3. **Clean Architecture**: Just `pkg/engine/` and `pkg/bridge/` - no business logic
4. **Script Infrastructure Only**: We only build what's needed for scripting (engines, type conversion, sandboxing)
5. **Type Safety**: Maintain type conversions at bridge boundaries

## Migration Status
- ✅ Updated go-llms to v0.3.5
- ✅ Phase 1.1: Script Engine Interface [COMPLETED]
- ✅ Phase 1.2: Core Bridge Foundation [COMPLETED]
- ✅ Phase 1.3: Core Bridge System [COMPLETED]
- ✅ Phase 1.4.1: Foundation Updates [COMPLETED - 2025-06-15]
- 🚧 Phase 1.4.2+: State Bridge Enhancements - NEXT
- 🚧 Phase 2-5: Engine Implementations - NOT STARTED

---

## Phase 1: Engine and Bridge Foundation

### ✅ 1.1 Script Engine Interface [COMPLETED]

### ✅ 1.2 Core Bridge Foundation [COMPLETED]

### ✅ 1.3 Core Bridge System [COMPLETED]
#### Items for revisit:
  - [ ] Support for async/promise-based tool execution (deferred to script engine implementation)
  - [ ] Test cross-engine compatibility (deferred to script engine implementation)

### 1.4 v0.3.5 Feature Integration

#### ✅ 1.4.1 Foundation Updates [COMPLETED - 2025-06-15]

All foundation updates for go-llms v0.3.5 integration completed. See TODO-DONE.md for detailed completion summary.

#### 1.4.2 State Bridge Enhancements

- [x] **Task 1.4.2.1: Add State Schema Validation** [COMPLETED - 2025-06-15]
  - [x] Ensure we leverage imports from go-llms pkg
  - [x] Add schemaRepo field to StateContextBridge
  - [x] Add stateSchema field for validation
  - [x] Implement ValidateState method
  - [x] Add schema versioning for states
  - [x] Support custom validation rules
  - [x] Add validation error details
  - [x] Check tests to use go-llms pkg/testutils

- [x] **Task 1.4.2.2: Add State Event Emission** [COMPLETED - 2025-06-15]
  - [x] Ensure we leverage imports from go-llms pkg
  - [x] Add eventEmitter to StateContextBridge
  - [x] Emit StateChangeEvent on set operations
  - [x] Emit StateDeleteEvent on delete operations
  - [x] Add state snapshot events
  - [x] Support event filtering by key patterns
  - [x] Add event replay for state reconstruction
  - [x] Check tests to use go-llms pkg/testutils


- [x] **Task 1.4.2.3: Add State Persistence with Schema Repository** [COMPLETED - 2025-06-15]
  - [x] Ensure we leverage imports from go-llms pkg
  - [x] Implement persistState method using schema repository
  - [x] Add loadState with schema validation
  - [x] Support versioned state snapshots
  - [x] Add state migration between versions
  - [x] Implement state diff generation
  - [x] Add compression for large states
  - [x] Check tests to use go-llms pkg/testutils

- [x] **Task 1.4.2.4: Add State Transformation Pipeline** [COMPLETED - 2025-06-15]
  - [x] Ensure we leverage imports from go-llms pkg
  - [x] Integrate with go-llms transformation pipeline
  - [x] Add pipeline configuration from scripts
  - [x] Support chained transformations
  - [x] Add transformation validation
  - [x] Implement transformation caching
  - [x] Add transformation metrics

#### 1.4.3 Utility Bridge Upgrades

- [x] **Task 1.4.3.1: Replace JSON Bridge with Structured Output Parser** [COMPLETED - 2025-06-15]
  - [x] Ensure we leverage imports from go-llms pkg
  - [x] Replace JSONBridge implementation with structured.JSONParser
  - [x] Add ParseWithRecovery for malformed JSON
  - [x] Add schema validation for parsed JSON
  - [x] Implement format conversion (JSON ↔ YAML ↔ XML)
  - [x] Add streaming JSON parsing support
  - [ ] Update tests for new parser capabilities
  - [ ] Check tests to use go-llms pkg/testutils

- [ ] **Task 1.4.3.2: Enhance Auth Bridge with OAuth2 Discovery**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add OAuth2 .well-known endpoint discovery
  - [ ] Implement token validation with schema system
  - [ ] Add auth event logging for security audit
  - [ ] Implement credential serialization
  - [ ] Add token refresh automation
  - [ ] Support multiple auth schemes per endpoint
  - [ ] Check tests to use go-llms pkg/testutils

- [ ] **Task 1.4.3.3: Enhance LLM Utility Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add provider capability metadata exposure
  - [ ] Integrate model discovery API
  - [ ] Add response parsing with recovery
  - [ ] Implement streaming event emission
  - [ ] Add cost tracking per request
  - [ ] Support provider-specific options
  - [ ] Check tests to use go-llms pkg/testutils

- [ ] **Task 1.4.3.4: Add Error Serialization Utilities**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement SerializableError wrapper
  - [ ] Add error recovery strategy support
  - [ ] Create error event emission
  - [ ] Add error categorization
  - [ ] Implement error aggregation
  - [ ] Support custom error handlers
  - [ ] Check tests to use go-llms pkg/testutils

#### 1.4.4 LLM Bridge Advanced Features

- [ ] **Task 1.4.4.1: Add Schema-Validated Generation**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add responseSchemas map to LLMBridge
  - [ ] Implement generateWithSchema method
  - [ ] Add schema validation for responses
  - [ ] Support multiple schema versions
  - [ ] Add schema inference from examples
  - [ ] Implement schema caching

- [ ] **Task 1.4.4.2: Add Provider Metadata Discovery**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement getProviderCapabilities method
  - [ ] Expose model-specific features
  - [ ] Add capability-based routing
  - [ ] Support dynamic provider selection
  - [ ] Add provider health monitoring
  - [ ] Implement fallback strategies

- [ ] **Task 1.4.4.3: Add Streaming with Event Emission**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement streaming response handling
  - [ ] Emit events for each stream chunk
  - [ ] Add stream aggregation support
  - [ ] Implement stream error recovery
  - [ ] Add stream performance metrics
  - [ ] Support stream transformation

#### 1.4.5 Schema Bridge Full Implementation

- [ ] **Task 1.4.5.1: Add Schema Versioning and Migration**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add fileRepo for file-based persistence
  - [ ] Implement saveSchemaVersion method
  - [ ] Add schema migration support
  - [ ] Create migration registry
  - [ ] Implement automatic migration
  - [ ] Add migration validation

- [ ] **Task 1.4.5.2: Add Tag-Based Schema Generation**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add tagGenerator field
  - [ ] Implement generateFromTags method
  - [ ] Support struct tag parsing
  - [ ] Add custom tag handlers
  - [ ] Generate documentation from tags
  - [ ] Support nested struct generation

- [ ] **Task 1.4.5.3: Add Schema Import/Export**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement schema export to JSON Schema
  - [ ] Add OpenAPI schema export
  - [ ] Support schema import from files
  - [ ] Add schema format conversion
  - [ ] Implement schema merging
  - [ ] Add schema diff generation

- [ ] **Task 1.4.5.4: Add Custom Validators**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Extend validator registration
  - [ ] Support script-based validators
  - [ ] Add async validation support
  - [ ] Implement validation caching
  - [ ] Add validation performance metrics
  - [ ] Support conditional validation

#### 1.4.6 Model Info Bridge Intelligence

- [ ] **Task 1.4.6.1: Add Model Performance Analytics**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add metricsStore field
  - [ ] Implement getModelPerformance method
  - [ ] Add performance report generation
  - [ ] Track latency, token usage, costs
  - [ ] Generate performance trends
  - [ ] Add anomaly detection

- [ ] **Task 1.4.6.2: Add Model Recommendation Engine**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement findModelsWithCapabilities
  - [ ] Add task-based model selection
  - [ ] Consider cost/performance tradeoffs
  - [ ] Support multi-criteria optimization
  - [ ] Add recommendation explanations
  - [ ] Implement A/B testing support

- [ ] **Task 1.4.6.3: Add Model Catalog Export**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement OpenAPI export for models
  - [ ] Add interactive documentation
  - [ ] Include pricing information
  - [ ] Add capability matrices
  - [ ] Generate comparison charts
  - [ ] Support custom export formats

#### 1.4.7 Agent Bridge Advanced Features

- [ ] **Task 1.4.7.1: Add Agent State Serialization**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement agent state export
  - [ ] Add state compression
  - [ ] Support incremental snapshots
  - [ ] Add state encryption option
  - [ ] Implement state versioning
  - [ ] Test state portability

- [ ] **Task 1.4.7.2: Add Agent Replay from Events**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement event-based replay
  - [ ] Add replay speed control
  - [ ] Support partial replay
  - [ ] Add replay debugging
  - [ ] Implement deterministic replay
  - [ ] Add replay visualization

- [ ] **Task 1.4.7.3: Add Agent Performance Profiling**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add profiling hooks
  - [ ] Track execution times
  - [ ] Monitor resource usage
  - [ ] Generate flame graphs
  - [ ] Add bottleneck detection
  - [ ] Support continuous profiling

#### 1.4.8 Event Bridge Replacement

- [ ] **Task 1.4.8.1: Replace with v0.3.5 Event System**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Remove current event bridge implementation
  - [ ] Integrate v0.3.5 EventEmitter
  - [ ] Add EventStore support
  - [ ] Implement event filtering
  - [ ] Add event serialization
  - [ ] Update all event tests

- [ ] **Task 1.4.8.2: Add Event Aggregation**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement event aggregation rules
  - [ ] Add time-window aggregation
  - [ ] Support custom aggregators
  - [ ] Add aggregation caching
  - [ ] Implement real-time dashboards
  - [ ] Export aggregated metrics

- [ ] **Task 1.4.8.3: Add Event Replay System**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement EventReplayer
  - [ ] Add replay filtering
  - [ ] Support speed control
  - [ ] Add checkpoint support
  - [ ] Implement replay hooks
  - [ ] Test replay accuracy

#### 1.4.9 Tools Bridge Enhancement

- [ ] **Task 1.4.9.1: Add Tool Schema Validation**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add schemaValidator field
  - [ ] Implement executeToolValidated
  - [ ] Validate input parameters
  - [ ] Validate output format
  - [ ] Add validation caching
  - [ ] Generate validation reports

- [ ] **Task 1.4.9.2: Add Tool Documentation Generation**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add docGenerator field
  - [ ] Generate tool documentation
  - [ ] Include examples and schemas
  - [ ] Add interactive playground
  - [ ] Generate SDK snippets
  - [ ] Support multiple languages

- [ ] **Task 1.4.9.3: Add Tool Execution Analytics**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Track tool execution metrics
  - [ ] Monitor success/failure rates
  - [ ] Add performance profiling
  - [ ] Generate usage reports
  - [ ] Implement cost tracking
  - [ ] Add anomaly alerts

#### 1.4.10 Workflow Bridge Serialization

- [ ] **Task 1.4.10.1: Add Workflow Import/Export**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Implement exportWorkflow method
  - [ ] Implement importWorkflow method
  - [ ] Add format validation
  - [ ] Support version compatibility
  - [ ] Add migration support
  - [ ] Test round-trip accuracy

- [ ] **Task 1.4.10.2: Add Script Step Handlers**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Support script-based workflow steps
  - [ ] Add step validation
  - [ ] Implement step debugging
  - [ ] Add step composition
  - [ ] Support async steps
  - [ ] Add step visualization

- [ ] **Task 1.4.10.3: Add Workflow Templates**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create template registry
  - [ ] Add template validation
  - [ ] Support parameterized templates
  - [ ] Add template composition
  - [ ] Generate template documentation
  - [ ] Implement template versioning

#### 1.4.11 Engine Integration

- [ ] **Task 1.4.11.1: Add Engine Event Bus**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add GetEventBus to ScriptEngine interface
  - [ ] Implement event bus per engine
  - [ ] Support cross-engine events
  - [ ] Add event routing
  - [ ] Implement event priorities
  - [ ] Test event isolation

- [ ] **Task 1.4.11.2: Add Type Conversion Registry**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add RegisterTypeConverter method
  - [ ] Implement conversion registry
  - [ ] Support bidirectional conversions
  - [ ] Add conversion caching
  - [ ] Generate conversion docs
  - [ ] Test conversion accuracy

- [ ] **Task 1.4.11.3: Add Engine Profiling**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add EnableProfiling method
  - [ ] Implement profiler interface
  - [ ] Track script execution
  - [ ] Monitor memory usage
  - [ ] Generate performance reports
  - [ ] Add optimization hints

- [ ] **Task 1.4.11.4: Add Engine API Export**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Add ExportAPI method
  - [ ] Generate API specifications
  - [ ] Include type information
  - [ ] Add method signatures
  - [ ] Generate client libraries
  - [ ] Support API versioning

### 1.5 Additional Original Bridges

- [ ] **Task 1.5.1: Tracing Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/observability/tracing.go`
  - [ ] Bridge core/tracing.go distributed tracing
  - [ ] Support OpenTelemetry integration
  - [ ] Enable trace correlation
  - [ ] Add trace sampling configuration
  - [ ] Support custom trace attributes

- [ ] **Task 1.5.2: Guardrails Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/guardrails.go`
  - [ ] Bridge guardrails.go safety system
  - [ ] Support content filtering
  - [ ] Enable behavioral constraints
  - [ ] Add custom guardrail rules
  - [ ] Implement guardrail analytics

- [ ] **Task 1.5.3: Metrics Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/observability/metrics.go`
  - [ ] Bridge performance metrics system
  - [ ] Support custom metric collection
  - [ ] Enable metric aggregation
  - [ ] Add metric export formats
  - [ ] Implement alerting rules

- [ ] **Task 1.5.4: Provider System Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/llm/providers.go`
  - [ ] Bridge all provider implementations (Anthropic, OpenAI, etc.)
  - [ ] Bridge consensus provider for multi-LLM voting
  - [ ] Bridge multi-provider with strategies (primary/fallback, sequential)
  - [ ] Expose provider configuration and options
  - [ ] Add provider-specific optimizations

- [ ] **Task 1.5.5: Provider Pool Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/llm/pool.go`
  - [ ] Bridge connection pooling from go-llms
  - [ ] Expose pool metrics and management
  - [ ] Support connection limits and timeouts
  - [ ] Add pool health monitoring
  - [ ] Implement adaptive pooling

- [ ] **Task 1.5.6: Built-in Tools Registry Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/agent/tools/registry.go`
  - [ ] Bridge the tool registry system
  - [ ] Expose tool discovery and metadata
  - [ ] Support dynamic tool loading
  - [ ] Add tool versioning support
  - [ ] Implement tool deprecation handling

- [ ] **Task 1.5.7: Profiling Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/observability/profiling.go`
  - [ ] Bridge performance profiling utilities
  - [ ] Support integration test profiling
  - [ ] Enable performance monitoring from scripts
  - [ ] Add CPU and memory profiling
  - [ ] Generate profiling reports

- [ ] **Task 1.5.8: Memory Bridge** ⏸️ **[DEFERRED - Not in go-llms yet]**
  - [ ] Will implement when available in go-llms

- [ ] **Task 1.5.9: Conversation Bridge** ⏸️ **[DEFERRED - Not in go-llms yet]**
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
1. Complete core bridges (llm_agent, workflow, events, tools)
2. Implement provider and pool bridges
3. Complete Lua engine and stdlib
4. Add JavaScript engine
5. Add Tengo engine
6. Create comprehensive examples

### Testing Strategy
- TDD for all new code
- Test bridges thoroughly
- Cross-engine conformance tests
- Performance benchmarks

### What We DON'T Build (CRITICAL)
- ❌ **NO LLM Logic**: No provider implementations, no API calls, no response parsing
- ❌ **NO Agent Logic**: No agent orchestration, no tool execution logic
- ❌ **NO State Management**: No state storage, transforms, or merging logic
- ❌ **NO Workflow Engine**: No workflow execution or state passing
- ❌ **NO Event System**: No event dispatching or subscription logic
- ❌ **NO Tools Implementation**: No tool logic, only bridging to go-llms tools
- ❌ **NO Business Features**: If it should be in go-llms, contribute it there first
- ❌ **NO Custom Abstractions**: No "improved" versions of go-llms features

### What We DO Build (Our ONLY Value-Add)
- ✅ **Script Engines**: Lua, JavaScript, Tengo execution environments
- ✅ **Type Converters**: Script ↔ Go type conversion infrastructure
- ✅ **Bridge Interfaces**: Thin wrappers that expose go-llms to scripts
- ✅ **Security Sandboxes**: Script execution isolation and resource limits
- ✅ **Language Bindings**: Idiomatic script APIs for each language
- ✅ **Examples/Documentation**: How to use go-llms from scripts

### If You're Tempted to Implement Something...
1. **STOP**: Does it exist in go-llms? → Bridge it
2. **STOP**: Should it exist in go-llms? → Contribute upstream first
3. **STOP**: Is it script-specific? → Only then implement it here

---

**Remember**: If it exists in go-llms, we bridge it. We only build what's unique to our scripting layer.