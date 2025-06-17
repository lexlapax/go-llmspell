# GopherLua Engine Implementation Plan

## Executive Summary

This implementation plan details the step-by-step approach to building the Lua script engine for go-llmspell based on the comprehensive architecture design. The plan is organized into logical implementation phases with clear dependencies, testing requirements, and deliverables.

## Implementation Phases Overview

```
Phase 2.2: Core Engine Components (Foundation)
    ├── 2.2.1: LState Pool Implementation
    ├── 2.2.2: Type Converter System
    ├── 2.2.3: Security Sandbox
    └── 2.2.4: Core Engine Integration

Phase 2.3: Bridge Integration Layer
    ├── 2.3.1: Module System Architecture
    ├── 2.3.2: Bridge Adapters
    ├── 2.3.3: Lua Standard Library
    └── 2.3.4: Async/Coroutine Support

Phase 2.4: Advanced Features & Optimization
    ├── 2.4.1: Performance Optimization
    ├── 2.4.2: Development Tools
    ├── 2.4.3: Production Readiness
    └── 2.4.4: Documentation & Examples
```

## Phase 2.2: Core Engine Components

### 2.2.1: LState Pool Implementation

**Purpose**: Implement thread-safe state pooling with adaptive management

#### Tasks:

1. **Create State Factory** (`/pkg/engine/gopherlua/factory.go`)
   - [ ] Define `LStateFactory` struct with configuration options
   - [ ] Implement `Create()` method for new LState instances
   - [ ] Add library loading based on security level
   - [ ] Implement initialization script execution
   - [ ] Add warmup strategy for JIT optimization
   - [ ] Create factory configuration with sensible defaults

2. **Implement State Pool** (`/pkg/engine/gopherlua/pool.go`)
   - [ ] Define `LStatePool` struct with adaptive parameters
   - [ ] Implement `Get()` with health checking
   - [ ] Implement `Put()` with cleanup validation
   - [ ] Add pool metrics tracking (usage, health, performance)
   - [ ] Implement adaptive scaling logic
   - [ ] Add graceful shutdown with timeout
   - [ ] Create pool configuration options

3. **State Health Management** (`/pkg/engine/gopherlua/health.go`)
   - [ ] Define health metrics (memory, errors, execution time)
   - [ ] Implement health scoring algorithm
   - [ ] Add state recycling based on health
   - [ ] Create health monitoring goroutine
   - [ ] Implement state quarantine for unhealthy instances

4. **Testing** (`/pkg/engine/gopherlua/pool_test.go`)
   - [ ] Test concurrent state acquisition/release
   - [ ] Test pool scaling under load
   - [ ] Test health-based recycling
   - [ ] Test graceful shutdown
   - [ ] Benchmark pool performance
   - [ ] Test resource leak prevention

### 2.2.2: Type Converter System

**Purpose**: Bidirectional type conversion between Lua and Go with optimization

#### Tasks:

1. **Core Type Converter** (`/pkg/engine/gopherlua/converter.go`)
   - [ ] Define `LuaTypeConverter` interface matching engine.TypeConverter
   - [ ] Implement `ToLua()` for Go → Lua conversions
   - [ ] Implement `FromLua()` for Lua → Go conversions
   - [ ] Add circular reference detection
   - [ ] Implement conversion caching for performance
   - [ ] Add custom type registration system

2. **Primitive Type Handling** (`/pkg/engine/gopherlua/converter_primitives.go`)
   - [ ] Implement bool ↔ LBool conversion
   - [ ] Implement number ↔ LNumber conversion (int, float64)
   - [ ] Implement string ↔ LString conversion
   - [ ] Implement nil ↔ LNil handling
   - [ ] Add type validation and error reporting

3. **Complex Type Handling** (`/pkg/engine/gopherlua/converter_complex.go`)
   - [ ] Implement map ↔ LTable conversion
   - [ ] Implement slice/array ↔ LTable conversion
   - [ ] Implement struct ↔ LTable/LUserData conversion
   - [ ] Add struct tag support for field mapping
   - [ ] Implement interface{} handling

4. **Bridge Type Integration** (`/pkg/engine/gopherlua/converter_bridge.go`)
   - [ ] Implement Bridge → LUserData conversion
   - [ ] Add metatable generation for bridge methods
   - [ ] Implement method wrapping with error handling
   - [ ] Add type safety checks at boundaries
   - [ ] Create bridge type registry

5. **Function Wrapping** (`/pkg/engine/gopherlua/converter_function.go`)
   - [ ] Implement Go function → LFunction wrapper
   - [ ] Add argument conversion and validation
   - [ ] Implement return value handling
   - [ ] Add panic recovery and error propagation
   - [ ] Support variadic functions

6. **Testing** (`/pkg/engine/gopherlua/converter_test.go`)
   - [ ] Test all primitive type conversions
   - [ ] Test complex type conversions with nesting
   - [ ] Test circular reference handling
   - [ ] Test bridge object conversions
   - [ ] Test function wrapping and error handling
   - [ ] Benchmark conversion performance

### 2.2.3: Security Sandbox

**Purpose**: Multi-layer security with configurable restrictions

#### Tasks:

1. **Security Manager** (`/pkg/engine/gopherlua/security.go`)
   - [ ] Define `SecurityManager` with policy configuration
   - [ ] Implement security level presets (minimal, standard, strict)
   - [ ] Add library whitelist/blacklist system
   - [ ] Implement function filtering
   - [ ] Create security policy validation

2. **Library Restrictions** (`/pkg/engine/gopherlua/security_libraries.go`)
   - [ ] Implement safe library loader
   - [ ] Remove dangerous functions from os library
   - [ ] Remove io library in strict mode
   - [ ] Remove debug library completely
   - [ ] Add custom safe replacements for common functions

3. **Resource Limits** (`/pkg/engine/gopherlua/security_limits.go`)
   - [ ] Implement instruction count limiting via debug hooks
   - [ ] Add memory limit monitoring
   - [ ] Implement execution timeout with context
   - [ ] Add stack depth limits
   - [ ] Create resource limit profiles

4. **Sandbox Enforcement** (`/pkg/engine/gopherlua/security_sandbox.go`)
   - [ ] Implement `ApplySandbox()` for LState configuration
   - [ ] Add import/require restrictions
   - [ ] Implement global environment filtering
   - [ ] Add metatable protection
   - [ ] Create sandbox escape prevention

5. **Testing** (`/pkg/engine/gopherlua/security_test.go`)
   - [ ] Test library restrictions by security level
   - [ ] Test resource limit enforcement
   - [ ] Test sandbox escape attempts
   - [ ] Test malicious script execution
   - [ ] Benchmark security overhead

### 2.2.4: Core Engine Integration

**Purpose**: Implement ScriptEngine interface with all components

#### Tasks:

1. **Engine Implementation** (`/pkg/engine/gopherlua/engine.go`)
   - [ ] Define `LuaEngine` struct implementing engine.ScriptEngine
   - [ ] Implement `Initialize()` with component setup
   - [ ] Implement `Execute()` with full execution pipeline
   - [ ] Implement `ExecuteFile()` with file handling
   - [ ] Implement `Shutdown()` with cleanup
   - [ ] Add engine configuration system

2. **Bridge Registration** (`/pkg/engine/gopherlua/engine_bridge.go`)
   - [ ] Implement `RegisterBridge()` with module creation
   - [ ] Implement `UnregisterBridge()` with cleanup
   - [ ] Add bridge lifecycle management
   - [ ] Create bridge method wrapping
   - [ ] Implement bridge metadata handling

3. **Execution Pipeline** (`/pkg/engine/gopherlua/engine_execute.go`)
   - [ ] Implement state acquisition from pool
   - [ ] Add security sandbox application
   - [ ] Implement parameter injection
   - [ ] Add script compilation with caching
   - [ ] Implement result extraction
   - [ ] Add comprehensive error handling

4. **Chunk Caching** (`/pkg/engine/gopherlua/cache.go`)
   - [ ] Implement `ChunkCache` with LRU eviction
   - [ ] Add cache key generation
   - [ ] Implement size-based eviction
   - [ ] Add TTL support for entries
   - [ ] Create cache metrics tracking

5. **Testing** (`/pkg/engine/gopherlua/engine_test.go`)
   - [ ] Test engine initialization and shutdown
   - [ ] Test script execution with various inputs
   - [ ] Test bridge registration and usage
   - [ ] Test error handling and recovery
   - [ ] Test concurrent execution
   - [ ] Benchmark execution performance

## Phase 2.3: Bridge Integration Layer

### 2.3.1: Module System Architecture

**Purpose**: Create flexible module system for bridge integration

#### Tasks:

1. **Module Registry** (`/pkg/engine/gopherlua/modules.go`)
   - [ ] Define `ModuleSystem` with registration
   - [ ] Implement module dependency resolution
   - [ ] Add lazy loading support
   - [ ] Create module priority system
   - [ ] Implement circular dependency detection

2. **Module Loader** (`/pkg/engine/gopherlua/modules_loader.go`)
   - [ ] Implement `PreloadModule()` for lazy loading
   - [ ] Add module initialization callbacks
   - [ ] Create profile-based loading
   - [ ] Implement module bundling
   - [ ] Add module version management

3. **Testing** (`/pkg/engine/gopherlua/modules_test.go`)
   - [ ] Test module registration and loading
   - [ ] Test dependency resolution
   - [ ] Test lazy loading behavior
   - [ ] Test circular dependency detection

### 2.3.2: Bridge Adapters

**Purpose**: Create Lua-specific adapters for go-llms bridges

#### Tasks:

1. **Bridge Adapter Base** (`/pkg/engine/gopherlua/bridge_adapter.go`)
   - [ ] Define `BridgeAdapter` interface
   - [ ] Implement base adapter with common functionality
   - [ ] Add method discovery and wrapping
   - [ ] Create error handling standards
   - [ ] Implement type hint system

2. **LLM Bridge Adapter** (`/pkg/engine/gopherlua/adapters/llm.go`)
   - [ ] Create LLM module with agent creation
   - [ ] Implement completion methods
   - [ ] Add streaming support
   - [ ] Implement model selection
   - [ ] Add token counting utilities

3. **State Bridge Adapter** (`/pkg/engine/gopherlua/adapters/state.go`)
   - [ ] Create state management module
   - [ ] Implement get/set operations
   - [ ] Add transform functions
   - [ ] Implement persistence methods
   - [ ] Add state merging capabilities

4. **Workflow Bridge Adapter** (`/pkg/engine/gopherlua/adapters/workflow.go`)
   - [ ] Create workflow module
   - [ ] Implement workflow builders
   - [ ] Add step definitions
   - [ ] Implement execution methods
   - [ ] Add state passing between steps

5. **Tools Bridge Adapter** (`/pkg/engine/gopherlua/adapters/tools.go`)
   - [ ] Create tools module
   - [ ] Implement tool registration
   - [ ] Add tool execution
   - [ ] Implement parameter validation
   - [ ] Add custom tool support

6. **Events Bridge Adapter** (`/pkg/engine/gopherlua/adapters/events.go`)
   - [ ] Create event module
   - [ ] Implement event subscription
   - [ ] Add event emission
   - [ ] Implement filtering
   - [ ] Add event correlation

7. **Testing** (`/pkg/engine/gopherlua/adapters/adapters_test.go`)
   - [ ] Test each adapter functionality
   - [ ] Test cross-adapter interaction
   - [ ] Test error propagation
   - [ ] Test type conversions

### 2.3.3: Lua Standard Library

**Purpose**: Create Lua-idiomatic APIs and utilities

#### Tasks:

1. **Core Utilities** (`/pkg/engine/gopherlua/stdlib/spell.lua`)
   - [ ] Create spell namespace with utilities
   - [ ] Add logging functions
   - [ ] Implement error handling helpers
   - [ ] Add type checking utilities
   - [ ] Create debugging helpers

2. **Promise Implementation** (`/pkg/engine/gopherlua/stdlib/promise.lua`)
   - [ ] Implement Promise class
   - [ ] Add then/catch/finally methods
   - [ ] Implement Promise.all()
   - [ ] Implement Promise.race()
   - [ ] Add async/await syntax sugar

3. **Error Handling** (`/pkg/engine/gopherlua/stdlib/errors.lua`)
   - [ ] Create error class hierarchy
   - [ ] Add stack trace capture
   - [ ] Implement error chaining
   - [ ] Add retry mechanisms
   - [ ] Create error recovery patterns

4. **Testing Utilities** (`/pkg/engine/gopherlua/stdlib/test.lua`)
   - [ ] Create assertion library
   - [ ] Add test runner
   - [ ] Implement mocking utilities
   - [ ] Add performance helpers
   - [ ] Create test reporting

5. **Documentation** (`/pkg/engine/gopherlua/stdlib/README.md`)
   - [ ] Document all stdlib modules
   - [ ] Add usage examples
   - [ ] Create API reference
   - [ ] Add best practices guide

### 2.3.4: Async/Coroutine Support

**Purpose**: Enable async operations with coroutines

#### Tasks:

1. **Async Runtime** (`/pkg/engine/gopherlua/async.go`)
   - [ ] Implement `AsyncRuntime` for coroutine management
   - [ ] Add promise-coroutine integration
   - [ ] Create async execution context
   - [ ] Implement cancellation support
   - [ ] Add timeout handling

2. **Channel Integration** (`/pkg/engine/gopherlua/channels.go`)
   - [ ] Implement Go channel ↔ LChannel bridge
   - [ ] Add select operation support
   - [ ] Create buffered channel support
   - [ ] Implement channel closing
   - [ ] Add deadlock detection

3. **Async Bridge Methods** (`/pkg/engine/gopherlua/async_bridges.go`)
   - [ ] Wrap bridge methods for async execution
   - [ ] Add automatic promisification
   - [ ] Implement streaming support
   - [ ] Add progress callbacks
   - [ ] Create cancellation tokens

4. **Testing** (`/pkg/engine/gopherlua/async_test.go`)
   - [ ] Test coroutine lifecycle
   - [ ] Test promise integration
   - [ ] Test channel operations
   - [ ] Test cancellation and timeouts
   - [ ] Test concurrent async operations

## Phase 2.4: Advanced Features & Optimization

### 2.4.1: Performance Optimization

**Purpose**: Optimize critical paths and reduce overhead

#### Tasks:

1. **Profiling Infrastructure** (`/pkg/engine/gopherlua/profiling.go`)
   - [ ] Add execution time tracking
   - [ ] Implement memory profiling
   - [ ] Create allocation tracking
   - [ ] Add hot path identification
   - [ ] Implement profiling API

2. **Type Conversion Optimization**
   - [ ] Implement conversion caching
   - [ ] Add fast paths for common types
   - [ ] Optimize table traversal
   - [ ] Reduce allocation in hot paths
   - [ ] Add benchmarks for all conversions

3. **State Pool Optimization**
   - [ ] Implement predictive scaling
   - [ ] Optimize state reset process
   - [ ] Add state pre-warming
   - [ ] Implement memory pooling
   - [ ] Create performance metrics

4. **Script Compilation Optimization**
   - [ ] Enhance chunk caching
   - [ ] Add AST optimization
   - [ ] Implement dead code elimination
   - [ ] Add constant folding
   - [ ] Create compilation benchmarks

### 2.4.2: Development Tools

**Purpose**: Enhance developer experience

#### Tasks:

1. **REPL Implementation** (`/cmd/llmspell-lua/main.go`)
   - [ ] Create interactive Lua REPL
   - [ ] Add command history
   - [ ] Implement auto-completion
   - [ ] Add syntax highlighting
   - [ ] Create help system

2. **Debugger Support** (`/pkg/engine/gopherlua/debug.go`)
   - [ ] Implement breakpoint support
   - [ ] Add step debugging
   - [ ] Create variable inspection
   - [ ] Implement stack trace visualization
   - [ ] Add watch expressions

3. **Script Validator** (`/pkg/engine/gopherlua/validator.go`)
   - [ ] Implement syntax validation
   - [ ] Add type checking where possible
   - [ ] Create linting rules
   - [ ] Implement security validation
   - [ ] Add performance warnings

4. **Documentation Generator** (`/pkg/engine/gopherlua/docs.go`)
   - [ ] Extract API from bridges
   - [ ] Generate Lua documentation
   - [ ] Create example extraction
   - [ ] Add type annotations
   - [ ] Generate completion data

### 2.4.3: Production Readiness

**Purpose**: Ensure production-quality implementation

#### Tasks:

1. **Comprehensive Testing**
   - [ ] Achieve 90%+ test coverage
   - [ ] Add integration test suite
   - [ ] Create stress tests
   - [ ] Implement chaos testing
   - [ ] Add regression test suite

2. **Error Handling Enhancement**
   - [ ] Standardize error types
   - [ ] Add error categorization
   - [ ] Implement error recovery
   - [ ] Create error reporting
   - [ ] Add error metrics

3. **Monitoring & Metrics**
   - [ ] Add Prometheus metrics
   - [ ] Implement health checks
   - [ ] Create performance dashboards
   - [ ] Add distributed tracing
   - [ ] Implement alerting rules

4. **Security Hardening**
   - [ ] Conduct security audit
   - [ ] Add input validation
   - [ ] Implement rate limiting
   - [ ] Create security benchmarks
   - [ ] Add CVE scanning

### 2.4.4: Documentation & Examples

**Purpose**: Complete user-facing documentation

#### Tasks:

1. **User Guide** (`/docs/user-guide/lua/`)
   - [ ] Getting started with Lua spells
   - [ ] Complete API reference
   - [ ] Common patterns and idioms
   - [ ] Troubleshooting guide
   - [ ] Migration from pure Lua

2. **Example Spells** (`/examples/lua/`)
   - [ ] Basic LLM interaction
   - [ ] Agent with tools
   - [ ] Complex workflows
   - [ ] Event-driven spells
   - [ ] Performance patterns

3. **Developer Documentation**
   - [ ] Architecture deep dive
   - [ ] Extension guide
   - [ ] Performance tuning
   - [ ] Security best practices
   - [ ] Contribution guide

## Implementation Timeline

### Week 1-2: Foundation (Phase 2.2.1-2.2.2)
- LState Pool implementation
- Type Converter system
- Basic testing

### Week 3-4: Security & Integration (Phase 2.2.3-2.2.4)
- Security Sandbox
- Core Engine integration
- Comprehensive testing

### Week 5-6: Bridge Layer (Phase 2.3.1-2.3.2)
- Module System
- Bridge Adapters
- Initial stdlib

### Week 7-8: Advanced Features (Phase 2.3.3-2.3.4)
- Complete stdlib
- Async support
- Integration testing

### Week 9-10: Optimization & Tools (Phase 2.4.1-2.4.2)
- Performance optimization
- Development tools
- Profiling

### Week 11-12: Production Ready (Phase 2.4.3-2.4.4)
- Production hardening
- Documentation
- Examples

## Success Criteria

1. **Functional Requirements**
   - [ ] All ScriptEngine interface methods implemented
   - [ ] All bridges accessible from Lua
   - [ ] Async operations fully supported
   - [ ] Security sandbox enforced

2. **Performance Requirements**
   - [ ] < 5% overhead vs direct go-llms usage
   - [ ] < 100ms script initialization time
   - [ ] Support 1000+ concurrent scripts
   - [ ] Memory usage < 10MB per state

3. **Quality Requirements**
   - [ ] 90%+ test coverage
   - [ ] Zero critical security issues
   - [ ] Comprehensive documentation
   - [ ] Production-ready stability

## Risk Mitigation

1. **Performance Risks**
   - Continuous benchmarking
   - Early optimization of hot paths
   - Fallback to simpler implementations

2. **Security Risks**
   - Regular security audits
   - Strict sandboxing by default
   - Continuous vulnerability scanning

3. **Compatibility Risks**
   - Extensive integration testing
   - Version pinning for dependencies
   - Backward compatibility tests

## Conclusion

This implementation plan provides a comprehensive roadmap for building the GopherLua engine for go-llmspell. By following this structured approach, we ensure a robust, secure, and performant implementation that maintains the project's bridge-first philosophy while providing an excellent developer experience for Lua spell writers.