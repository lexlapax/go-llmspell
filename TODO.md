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
- ✅ Phase 1: Engine and Bridge Foundation [COMPLETED - 2025-06-17]
  - 38+ bridges across 13 categories with comprehensive test coverage
  - Pure bridge architecture: zero business logic duplication
- 🚧 Phase 2: Lua Engine Implementation - RESEARCH COMPLETE, CORE COMPONENTS COMPLETE
  - Security Sandbox: ✅ COMPLETED [2025-06-17]  
  - Type Converter: ✅ COMPLETED [2025-06-18]
  - LState Pool: ✅ COMPLETED [2025-06-18]
  - Core Engine: 🔄 NEXT (All dependencies satisfied)
- 🚧 Phase 3: JavaScript Engine Implementation - NOT STARTED
- 🚧 Phase 4: Tengo Engine Implementation - NOT STARTED
- 🚧 Phase 5: Integration and Examples - NOT STARTED

---
## DEFERRED TASKS from different Phases - For Revisit 
- See `TODO-DONE-ARCHIVE.md` for completed tasks history

### Section 1.3.
  - [ ] **Task 1.3.20: Support for async/promise-based tool execution** (**[DEFERRED]** to script engine implementation)
  - [ ] **Task 1.3.21: Test cross-engine compatibility** (**[DEFERRED]** to script engine implementation)

#### ⏸️ 1.4.6 Model Info Bridge Intelligence **[DEFERRED]** - Features not in go-llms
**Status**: Tasks deferred - missing features documented in `go-llms-upstream-request.md`

- [ ] **Task 1.4.6.1: Add Model Performance Analytics** ⏸️ **[DEFERRED]**
  - Missing from go-llms: Model performance tracking, analytics, metrics
  - Documented in upstream request #1

- [ ] **Task 1.4.6.2: Add Model Recommendation Engine** ⏸️ **[DEFERRED]**  
  - Missing from go-llms: Recommendation algorithms, model selection
  - Documented in upstream request #2

- [ ] **Task 1.4.6.3: Add Model Catalog Export** ⏸️ **[DEFERRED]**
  - Missing from go-llms: Catalog export, OpenAPI generation for models
  - Documented in upstream request #3
- [ ] **Task 1.5.8: Memory Bridge** ⏸️ **[DEFERRED]** - Not in go-llms yet
  - [ ] Will implement when available in go-llms

### Section 1.5
- [ ] **Task 1.5.9: Conversation Bridge** ⏸️ **[DEFERRED]** - Not in go-llms yet
  - [ ] Will implement when available in go-llms


---

## Phase 2: Lua Engine Implementation
### 2.1 Lua Engine Research and Planning
✅ **COMPLETED [2025-06-17]** - All 14 research tasks completed. See TODO-DONE.md for details.

### Phase 2.2: Core Engine Components

**Implementation Order Based on Dependencies:**
1. **FIRST**: Security Sandbox (2.2.3) & Type Converter (2.2.2) - No dependencies, can be done in parallel ✅ **SECURITY COMPLETED** ✅ **CORE TYPE CONVERTER COMPLETED**
2. **SECOND**: LState Pool (2.2.1) - Depends on Security Sandbox for library loading configuration
3. **THIRD**: Core Engine (2.2.4) - Depends on all above components

**Current Implementation Status:**
- ✅ **Phase 2.2.1: LState Pool Implementation** - COMPLETED [2025-06-18] (All 4 tasks) ➜ Moved to TODO-DONE.md
- ✅ **Phase 2.2.2: Type Converter System** - COMPLETED [2025-06-18] (All 6 tasks) ➜ Moved to TODO-DONE.md
- ✅ **Phase 2.2.3: Security Sandbox System** - COMPLETED [2025-06-17] (All 5 tasks) ➜ Moved to TODO-DONE.md
- 🔄 **Phase 2.2.4: Core Engine Integration** - READY TO START (Dependencies satisfied)

#### 2.2.4: Core Engine Integration [IN PROGRESS - 1/5 tasks completed]
- [x] **Task 2.2.4.1: Engine Implementation** (`/pkg/engine/gopherlua/engine.go`) ✅ COMPLETED
  - [x] Define `LuaEngine` struct implementing engine.ScriptEngine
  - [x] Implement `Initialize()` with component setup
  - [x] Implement `Execute()` with full execution pipeline
  - [x] Implement `ExecuteFile()` with file handling
  - [x] Implement `Shutdown()` with cleanup
  - [x] Add engine configuration system
      Key Features Implemented:
      1. LuaEngine - Complete ScriptEngine interface implementation with pool, converter, security integration
      2. Initialize() - Configurable initialization with SecurityManager, pool setup, factory creation
      3. Execute() - Full execution pipeline with parameter injection, chunk caching, timeout handling
      4. ExecuteFile() - File-based script execution with extension validation
      5. Shutdown() - Graceful shutdown with pool cleanup and bridge management
      6. Type conversion - Go ↔ Lua conversion through integrated TypeConverter
      7. Bridge management - RegisterBridge, UnregisterBridge, GetBridge, ListBridges
      8. Resource limits - Memory, timeout, and comprehensive resource management
      9. Error handling - Proper EngineError wrapping with type categorization
      10. Thread-safe metrics - Atomic operations for all metrics tracking to prevent data races
      11. Comprehensive testing - 8 test suites with 100% pass rate covering all functionality, race-condition free

- [ ] **Task 2.2.4.2: Bridge Registration** (`/pkg/engine/gopherlua/engine_bridge.go`)
  - [ ] Implement `RegisterBridge()` with module creation
  - [ ] Implement `UnregisterBridge()` with cleanup
  - [ ] Add bridge lifecycle management
  - [ ] Create bridge method wrapping
  - [ ] Implement bridge metadata handling

- [ ] **Task 2.2.4.3: Execution Pipeline** (`/pkg/engine/gopherlua/engine_execute.go`)
  - [ ] Implement state acquisition from pool
  - [ ] Add security sandbox application
  - [ ] Implement parameter injection
  - [ ] Add script compilation with caching
  - [ ] Implement result extraction
  - [ ] Add comprehensive error handling

- [ ] **Task 2.2.4.4: Chunk Caching** (`/pkg/engine/gopherlua/cache.go`)
  - [ ] Implement `ChunkCache` with LRU eviction
  - [ ] Add cache key generation
  - [ ] Implement size-based eviction
  - [ ] Add TTL support for entries
  - [ ] Create cache metrics tracking

- [ ] **Task 2.2.4.5: Engine Testing** (`/pkg/engine/gopherlua/engine_test.go`)
  - [ ] Test engine initialization and shutdown
  - [ ] Test script execution with various inputs
  - [ ] Test bridge registration and usage
  - [ ] Test error handling and recovery
  - [ ] Test concurrent execution
  - [ ] Benchmark execution performance

### Phase 2.3: Bridge Integration Layer

#### 2.3.1: Module System Architecture
- [ ] **Task 2.3.1.1: Module Registry** (`/pkg/engine/gopherlua/modules.go`)
  - [ ] Define `ModuleSystem` with registration
  - [ ] Implement module dependency resolution
  - [ ] Add lazy loading support
  - [ ] Create module priority system
  - [ ] Implement circular dependency detection

- [ ] **Task 2.3.1.2: Module Loader** (`/pkg/engine/gopherlua/modules_loader.go`)
  - [ ] Implement `PreloadModule()` for lazy loading
  - [ ] Add module initialization callbacks
  - [ ] Create profile-based loading
  - [ ] Implement module bundling
  - [ ] Add module version management

- [ ] **Task 2.3.1.3: Module Testing** (`/pkg/engine/gopherlua/modules_test.go`)
  - [ ] Test module registration and loading
  - [ ] Test dependency resolution
  - [ ] Test lazy loading behavior
  - [ ] Test circular dependency detection

#### 2.3.2: Bridge Adapters
- [ ] **Task 2.3.2.1: Bridge Adapter Base** (`/pkg/engine/gopherlua/bridge_adapter.go`)
  - [ ] Define `BridgeAdapter` interface
  - [ ] Implement base adapter with common functionality
  - [ ] Add method discovery and wrapping
  - [ ] Create error handling standards
  - [ ] Implement type hint system

- [ ] **Task 2.3.2.2: LLM Bridge Adapter** (`/pkg/engine/gopherlua/adapters/llm.go`)
  - [ ] Create LLM module with agent creation
  - [ ] Implement completion methods
  - [ ] Add streaming support
  - [ ] Implement model selection
  - [ ] Add token counting utilities

- [ ] **Task 2.3.2.3: State Bridge Adapter** (`/pkg/engine/gopherlua/adapters/state.go`)
  - [ ] Create state management module
  - [ ] Implement get/set operations
  - [ ] Add transform functions
  - [ ] Implement persistence methods
  - [ ] Add state merging capabilities

- [ ] **Task 2.3.2.4: Workflow Bridge Adapter** (`/pkg/engine/gopherlua/adapters/workflow.go`)
  - [ ] Create workflow module
  - [ ] Implement workflow builders
  - [ ] Add step definitions
  - [ ] Implement execution methods
  - [ ] Add state passing between steps

- [ ] **Task 2.3.2.5: Tools Bridge Adapter** (`/pkg/engine/gopherlua/adapters/tools.go`)
  - [ ] Create tools module
  - [ ] Implement tool registration
  - [ ] Add tool execution
  - [ ] Implement parameter validation
  - [ ] Add custom tool support

- [ ] **Task 2.3.2.6: Events Bridge Adapter** (`/pkg/engine/gopherlua/adapters/events.go`)
  - [ ] Create event module
  - [ ] Implement event subscription
  - [ ] Add event emission
  - [ ] Implement filtering
  - [ ] Add event correlation

- [ ] **Task 2.3.2.7: Adapter Testing** (`/pkg/engine/gopherlua/adapters/adapters_test.go`)
  - [ ] Test each adapter functionality
  - [ ] Test cross-adapter interaction
  - [ ] Test error propagation
  - [ ] Test type conversions

#### 2.3.3: Lua Standard Library
- [ ] **Task 2.3.3.1: Core Utilities** (`/pkg/engine/gopherlua/stdlib/spell.lua`)
  - [ ] Create spell namespace with utilities
  - [ ] Add logging functions
  - [ ] Implement error handling helpers
  - [ ] Add type checking utilities
  - [ ] Create debugging helpers

- [ ] **Task 2.3.3.2: Promise Implementation** (`/pkg/engine/gopherlua/stdlib/promise.lua`)
  - [ ] Implement Promise class
  - [ ] Add then/catch/finally methods
  - [ ] Implement Promise.all()
  - [ ] Implement Promise.race()
  - [ ] Add async/await syntax sugar

- [ ] **Task 2.3.3.3: Error Handling** (`/pkg/engine/gopherlua/stdlib/errors.lua`)
  - [ ] Create error class hierarchy
  - [ ] Add stack trace capture
  - [ ] Implement error chaining
  - [ ] Add retry mechanisms
  - [ ] Create error recovery patterns

- [ ] **Task 2.3.3.4: Testing Utilities** (`/pkg/engine/gopherlua/stdlib/test.lua`)
  - [ ] Create assertion library
  - [ ] Add test runner
  - [ ] Implement mocking utilities
  - [ ] Add performance helpers
  - [ ] Create test reporting

- [ ] **Task 2.3.3.5: Documentation** (`/pkg/engine/gopherlua/stdlib/README.md`)
  - [ ] Document all stdlib modules
  - [ ] Add usage examples
  - [ ] Create API reference
  - [ ] Add best practices guide

#### 2.3.4: Async/Coroutine Support
- [ ] **Task 2.3.4.1: Async Runtime** (`/pkg/engine/gopherlua/async.go`)
  - [ ] Implement `AsyncRuntime` for coroutine management
  - [ ] Add promise-coroutine integration
  - [ ] Create async execution context
  - [ ] Implement cancellation support
  - [ ] Add timeout handling

- [ ] **Task 2.3.4.2: Channel Integration** (`/pkg/engine/gopherlua/channels.go`)
  - [ ] Implement Go channel ↔ LChannel bridge
  - [ ] Add select operation support
  - [ ] Create buffered channel support
  - [ ] Implement channel closing
  - [ ] Add deadlock detection

- [ ] **Task 2.3.4.3: Async Bridge Methods** (`/pkg/engine/gopherlua/async_bridges.go`)
  - [ ] Wrap bridge methods for async execution
  - [ ] Add automatic promisification
  - [ ] Implement streaming support
  - [ ] Add progress callbacks
  - [ ] Create cancellation tokens

- [ ] **Task 2.3.4.4: Async Testing** (`/pkg/engine/gopherlua/async_test.go`)
  - [ ] Test coroutine lifecycle
  - [ ] Test promise integration
  - [ ] Test channel operations
  - [ ] Test cancellation and timeouts
  - [ ] Test concurrent async operations

### Phase 2.4: Advanced Features & Optimization

#### 2.4.1: Performance Optimization
- [ ] **Task 2.4.1.1: Profiling Infrastructure** (`/pkg/engine/gopherlua/profiling.go`)
  - [ ] Add execution time tracking
  - [ ] Implement memory profiling
  - [ ] Create allocation tracking
  - [ ] Add hot path identification
  - [ ] Implement profiling API

- [ ] **Task 2.4.1.2: Type Conversion Optimization**
  - [ ] Implement conversion caching
  - [ ] Add fast paths for common types
  - [ ] Optimize table traversal
  - [ ] Reduce allocation in hot paths
  - [ ] Add benchmarks for all conversions

- [ ] **Task 2.4.1.3: State Pool Optimization**
  - [ ] Implement predictive scaling
  - [ ] Optimize state reset process
  - [ ] Add state pre-warming
  - [ ] Implement memory pooling
  - [ ] Create performance metrics

- [ ] **Task 2.4.1.4: Script Compilation Optimization**
  - [ ] Enhance chunk caching
  - [ ] Add AST optimization
  - [ ] Implement dead code elimination
  - [ ] Add constant folding
  - [ ] Create compilation benchmarks

#### 2.4.2: Development Tools
- [ ] **Task 2.4.2.1: REPL Implementation** (`/cmd/llmspell-lua/main.go`)
  - [ ] Create interactive Lua REPL
  - [ ] Add command history
  - [ ] Implement auto-completion
  - [ ] Add syntax highlighting
  - [ ] Create help system

- [ ] **Task 2.4.2.2: Debugger Support** (`/pkg/engine/gopherlua/debug.go`)
  - [ ] Implement breakpoint support
  - [ ] Add step debugging
  - [ ] Create variable inspection
  - [ ] Implement stack trace visualization
  - [ ] Add watch expressions

- [ ] **Task 2.4.2.3: Script Validator** (`/pkg/engine/gopherlua/validator.go`)
  - [ ] Implement syntax validation
  - [ ] Add type checking where possible
  - [ ] Create linting rules
  - [ ] Implement security validation
  - [ ] Add performance warnings

- [ ] **Task 2.4.2.4: Documentation Generator** (`/pkg/engine/gopherlua/docs.go`)
  - [ ] Extract API from bridges
  - [ ] Generate Lua documentation
  - [ ] Create example extraction
  - [ ] Add type annotations
  - [ ] Generate completion data

#### 2.4.3: Production Readiness
- [ ] **Task 2.4.3.1: Comprehensive Testing**
  - [ ] Achieve 90%+ test coverage
  - [ ] Add integration test suite
  - [ ] Create stress tests
  - [ ] Implement chaos testing
  - [ ] Add regression test suite

- [ ] **Task 2.4.3.2: Error Handling Enhancement**
  - [ ] Standardize error types
  - [ ] Add error categorization
  - [ ] Implement error recovery
  - [ ] Create error reporting
  - [ ] Add error metrics

- [ ] **Task 2.4.3.3: Monitoring & Metrics**
  - [ ] Add Prometheus metrics
  - [ ] Implement health checks
  - [ ] Create performance dashboards
  - [ ] Add distributed tracing
  - [ ] Implement alerting rules

- [ ] **Task 2.4.3.4: Security Hardening**
  - [ ] Conduct security audit
  - [ ] Add input validation
  - [ ] Implement rate limiting
  - [ ] Create security benchmarks
  - [ ] Add CVE scanning

#### 2.4.4: Documentation & Examples
- [ ] **Task 2.4.4.1: User Guide** (`/docs/user-guide/lua/`)
  - [ ] Getting started with Lua spells
  - [ ] Complete API reference
  - [ ] Common patterns and idioms
  - [ ] Troubleshooting guide
  - [ ] Migration from pure Lua

- [ ] **Task 2.4.4.2: Example Spells** (`/examples/lua/`)
  - [ ] Basic LLM interaction
  - [ ] Agent with tools
  - [ ] Complex workflows
  - [ ] Event-driven spells
  - [ ] Performance patterns

- [ ] **Task 2.4.4.3: Developer Documentation**
  - [ ] Architecture deep dive
  - [ ] Extension guide
  - [ ] Performance tuning
  - [ ] Security best practices
  - [ ] Contribution guide
---

## Phase 3: JavaScript Engine Implementation

### 3.1 JavaScript Engine Research and Planning
- [ ] 3.1.1. Research goja (https://github.com/dop251/goja) go. Find the best javascript engine to work with in go-llmspell (There are others). 
- [ ] 3.1.2. Research how to integrate the chosen javascript engine into this go-llmspell library. add additional TODO.md entries as needed 
- [ ] 3.1.3. Analyze state management and memory integration
- [ ] 3.1.4. Design ScriptValue ↔ javascript type conversion system 
- [ ] 3.1.5. Plan goroutine integration for async operations
- [ ] 3.1.6. Design security sandboxing approach
- [ ] 3.1.7. Create detailed implementation roadmap
- [ ] 3.1.8. Research  bytecode validation and security implications - may not apply to gopher-lua
- [ ] 3.1.9. Investigate warning system integration 
- [ ] 3.1.10. Study generational GC vs incremental GC trade-offs if it applies
- [ ] 3.1.11. Research goja debug introspection capabilities for development tools
- [ ] 3.1.12. Combine all research documents and re-synthesize into one javascript_engine_architecture.md based on `docs/technical/architecture.md` and a detailed implementation roadmap

### 3.2 JavaScript Engine Core
- [ ] **Task 3.2.1: Engine Implementation**
  - [ ] Create test file `/pkg/engine/javascript/engine_test.go`
  - [ ] Test ScriptEngine interface implementation
  - [ ] Test Goja integration
  - [ ] Test ES6+ or ES5.1+ whichever is the lstest support
  - [ ] Create `/pkg/engine/javascript/engine.go`
  - [ ] Implement ScriptEngine interface for JS
  - [ ] Integrate Goja
  - [ ] Add ES6+ support

- [ ] **Task 3.2.2: Type Converter**
  - [ ] Create test file `/pkg/engine/javascript/converter_test.go`
  - [ ] Test JS ↔ Go type conversions
  - [ ] Test Promise handling
  - [ ] Create `/pkg/engine/javascript/converter.go`
  - [ ] Implement type conversions
  - [ ] Handle async patterns

- [ ] **Task 3.2.3: Security Sandbox**
  - [ ] Create test file `/pkg/engine/javascript/sandbox_test.go`
  - [ ] Test global access restrictions
  - [ ] Test resource limits
  - [ ] Create `/pkg/engine/javascript/sandbox.go`
  - [ ] Restrict global access
  - [ ] Implement CSP-like policies

### 3.3 JavaScript Standard Library
- [ ] **Task 3.3.1: Core Modules**
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