# Phase 2.2 Refactored Implementation Order

Based on dependency analysis, here's the correct implementation order for Phase 2.2:

## Implementation Order (By Dependencies)

### 1. FIRST: Components with No Dependencies (Can be done in parallel)

#### 2.2.3: Security Sandbox (PREREQUISITE for State Factory)
- [ ] **Task 2.2.3.1: Security Manager** (`/pkg/engine/gopherlua/security.go`)
  - [ ] Define `SecurityManager` with policy configuration
  - [ ] Implement security level presets (minimal, standard, strict)
  - [ ] Add library whitelist/blacklist system
  - [ ] Implement function filtering
  - [ ] Create security policy validation

- [ ] **Task 2.2.3.2: Library Restrictions** (`/pkg/engine/gopherlua/security_libraries.go`)
  - [ ] Implement safe library loader
  - [ ] Remove dangerous functions from os library
  - [ ] Remove io library in strict mode
  - [ ] Remove debug library completely
  - [ ] Add custom safe replacements for common functions

- [ ] **Task 2.2.3.3: Resource Limits** (`/pkg/engine/gopherlua/security_limits.go`)
  - [ ] Implement instruction count limiting via debug hooks
  - [ ] Add memory limit monitoring
  - [ ] Implement execution timeout with context
  - [ ] Add stack depth limits
  - [ ] Create resource limit profiles

- [ ] **Task 2.2.3.4: Sandbox Enforcement** (`/pkg/engine/gopherlua/security_sandbox.go`)
  - [ ] Implement `ApplySandbox()` for LState configuration
  - [ ] Add import/require restrictions
  - [ ] Implement global environment filtering
  - [ ] Add metatable protection
  - [ ] Create sandbox escape prevention

- [ ] **Task 2.2.3.5: Security Testing** (`/pkg/engine/gopherlua/security_test.go`)
  - [ ] Test library restrictions by security level
  - [ ] Test resource limit enforcement
  - [ ] Test sandbox escape attempts
  - [ ] Test malicious script execution
  - [ ] Benchmark security overhead

#### 2.2.2: Type Converter System (Independent)
- [ ] **Task 2.2.2.1: Core Type Converter** (`/pkg/engine/gopherlua/converter.go`)
  - [ ] Define `LuaTypeConverter` interface matching engine.TypeConverter
  - [ ] Implement `ToLua()` for Go → Lua conversions
  - [ ] Implement `FromLua()` for Lua → Go conversions
  - [ ] Add circular reference detection
  - [ ] Implement conversion caching for performance
  - [ ] Add custom type registration system

- [ ] **Task 2.2.2.2: Primitive Type Handling** (`/pkg/engine/gopherlua/converter_primitives.go`)
  - [ ] Implement bool ↔ LBool conversion
  - [ ] Implement number ↔ LNumber conversion (int, float64)
  - [ ] Implement string ↔ LString conversion
  - [ ] Implement nil ↔ LNil handling
  - [ ] Add type validation and error reporting

- [ ] **Task 2.2.2.3: Complex Type Handling** (`/pkg/engine/gopherlua/converter_complex.go`)
  - [ ] Implement map ↔ LTable conversion
  - [ ] Implement slice/array ↔ LTable conversion
  - [ ] Implement struct ↔ LTable/LUserData conversion
  - [ ] Add struct tag support for field mapping
  - [ ] Implement interface{} handling

- [ ] **Task 2.2.2.4: Bridge Type Integration** (`/pkg/engine/gopherlua/converter_bridge.go`)
  - [ ] Implement Bridge → LUserData conversion
  - [ ] Add metatable generation for bridge methods
  - [ ] Implement method wrapping with error handling
  - [ ] Add type safety checks at boundaries
  - [ ] Create bridge type registry

- [ ] **Task 2.2.2.5: Function Wrapping** (`/pkg/engine/gopherlua/converter_function.go`)
  - [ ] Implement Go function → LFunction wrapper
  - [ ] Add argument conversion and validation
  - [ ] Implement return value handling
  - [ ] Add panic recovery and error propagation
  - [ ] Support variadic functions

- [ ] **Task 2.2.2.6: Converter Testing** (`/pkg/engine/gopherlua/converter_test.go`)
  - [ ] Test all primitive type conversions
  - [ ] Test complex type conversions with nesting
  - [ ] Test circular reference handling
  - [ ] Test bridge object conversions
  - [ ] Test function wrapping and error handling
  - [ ] Benchmark conversion performance

### 2. SECOND: Components that depend on Security

#### 2.2.1: LState Pool Implementation (DEPENDS ON: Security Sandbox)
- [ ] **Task 2.2.1.1: Create State Factory** (`/pkg/engine/gopherlua/factory.go`)
  - [ ] Define `LStateFactory` struct with configuration options
  - [ ] Accept SecurityManager/SecurityConfig in factory config
  - [ ] Use SecurityManager to determine library loading
  - [ ] Implement `Create()` method with security sandbox application
  - [ ] Add initialization script execution
  - [ ] Add warmup strategy for JIT optimization
  - [ ] Create factory configuration with sensible defaults

- [ ] **Task 2.2.1.2: Implement State Pool** (`/pkg/engine/gopherlua/pool.go`)
  - [ ] Define `LStatePool` struct with adaptive parameters
  - [ ] Implement `Get()` with health checking
  - [ ] Implement `Put()` with cleanup validation
  - [ ] Add pool metrics tracking (usage, health, performance)
  - [ ] Implement adaptive scaling logic
  - [ ] Add graceful shutdown with timeout
  - [ ] Create pool configuration options

- [ ] **Task 2.2.1.3: State Health Management** (`/pkg/engine/gopherlua/health.go`)
  - [ ] Define health metrics (memory, errors, execution time)
  - [ ] Implement health scoring algorithm
  - [ ] Add state recycling based on health
  - [ ] Create health monitoring goroutine
  - [ ] Implement state quarantine for unhealthy instances

- [ ] **Task 2.2.1.4: Pool Testing** (`/pkg/engine/gopherlua/pool_test.go`)
  - [ ] Test concurrent state acquisition/release
  - [ ] Test pool scaling under load
  - [ ] Test health-based recycling
  - [ ] Test graceful shutdown
  - [ ] Benchmark pool performance
  - [ ] Test resource leak prevention

### 3. THIRD: Core Engine that depends on all components

#### 2.2.4: Core Engine Integration (DEPENDS ON: Security, Type Converter, State Pool)
- [ ] **Task 2.2.4.1: Engine Implementation** (`/pkg/engine/gopherlua/engine.go`)
  - [ ] Define `LuaEngine` struct implementing engine.ScriptEngine
  - [ ] Integrate SecurityManager for sandbox configuration
  - [ ] Use LStatePool for state management
  - [ ] Use TypeConverter for parameter/result conversion
  - [ ] Implement `Initialize()` with component setup
  - [ ] Implement `Execute()` with full execution pipeline
  - [ ] Implement `ExecuteFile()` with file handling
  - [ ] Implement `Shutdown()` with cleanup
  - [ ] Add engine configuration system

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

## Summary of Changes:

1. **Security Sandbox (2.2.3)** moved to FIRST - it's a prerequisite
2. **Type Converter (2.2.2)** moved to FIRST - it's independent
3. **State Factory (2.2.1)** moved to SECOND - depends on Security
4. **Core Engine (2.2.4)** remains LAST - depends on all components

This ordering ensures we build components in dependency order, avoiding the situation where we're hardcoding security into the factory instead of using a proper security system.