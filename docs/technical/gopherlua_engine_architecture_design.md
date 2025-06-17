# GopherLua Engine Architecture Design

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Core Components](#core-components)
4. [Implementation Design](#implementation-design)
5. [Bridge Integration](#bridge-integration)
6. [Performance & Optimization](#performance--optimization)
7. [Security Model](#security-model)
8. [Timeout Handling](#timeout-handling--state-abandonment)
9. [Development Roadmap](#development-roadmap)
10. [Testing Strategy](#testing-strategy)
11. [API Reference](#api-reference)

## Executive Summary

This document synthesizes all research conducted on implementing the Lua script engine for go-llmspell using GopherLua. The design provides a robust, secure, and performant scripting environment that bridges Lua scripts to go-llms functionality while maintaining the project's core philosophy of "bridge, don't build."

**Note**: The detailed research documents that informed this architecture have been archived to `/docs/archives/research/` for reference.

### Key Design Decisions

1. **GopherLua as Engine**: Lua 5.1 VM implementation in pure Go
2. **UserData for Bridges**: Type-safe bridge object representation
3. **Adaptive State Pooling**: Intelligent LState lifecycle management
4. **Coroutine-Based Async**: Non-blocking bridge operations
5. **Multi-Layer Security**: Sandboxing, resource limits, and monitoring
6. **Lazy Module Loading**: Optimized startup and memory usage

### Architecture Principles

- **Thread Safety**: Each goroutine gets its own LState instance
- **Resource Efficiency**: State pooling with health monitoring
- **Type Safety**: Strong typing at bridge boundaries via UserData
- **Security First**: Multiple layers of sandboxing and limits
- **Performance Optimized**: Compiled chunk caching, lazy loading
- **Error Resilience**: Comprehensive error handling with stack preservation

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Lua Scripts                          │
├─────────────────────────────────────────────────────────────┤
│                    Script Engine API                        │
├─────────────────┬─────────────────┬────────────────────────┤
│  Module System  │  Type Converter │  Security Sandbox      │
├─────────────────┼─────────────────┼────────────────────────┤
│  State Manager  │  Bridge Layer   │  Resource Monitor      │
├─────────────────┴─────────────────┴────────────────────────┤
│                      GopherLua Core                         │
├─────────────────────────────────────────────────────────────┤
│                    go-llms Bridges                          │
└─────────────────────────────────────────────────────────────┘
```

### Component Interaction Flow

```
Script Execution Request
    ↓
Engine.Execute()
    ↓
State Pool ← Get/Create LState
    ↓
Security Sandbox ← Apply Limits
    ↓
Module Loader ← Preload Bridges
    ↓
Type Converter ← Script ↔ Go Types
    ↓
Bridge Methods ← Execute Operations
    ↓
Result Conversion
    ↓
State Cleanup → Return to Pool
```

## Core Components

### 1. LState Management

**Design**: Thread-safe state pooling with lifecycle management

```go
type LStatePool struct {
    factory      *LStateFactory
    cleanup      *LStateCleanup
    
    // Pool configuration
    minSize      int
    maxSize      int
    targetSize   int
    
    // Adaptive parameters
    scaleUpThreshold   float64
    scaleDownThreshold float64
    
    // Health monitoring
    healthChecker *StateHealthChecker
    
    // State tracking
    states       []*PooledState
    available    chan *PooledState
    metrics      PoolMetrics
}

type PooledState struct {
    state       *lua.LState
    id          string
    createdAt   time.Time
    lastUsedAt  time.Time
    useCount    int64
    errorCount  int64
    healthScore float64
    
    // Execution tracking for timeout safety
    executing   bool           // true when state is in use
    done        chan struct{}  // closed when execution completes
    mu          sync.Mutex     // protects executing flag
}
```

**Key Features**:
- Adaptive scaling based on usage patterns
- Health-based state recycling
- Generation-based lifecycle management
- Comprehensive metrics tracking
- Safe timeout handling with state abandonment
- Execution tracking for race condition prevention

### 2. Type Conversion System

**Design**: Bidirectional type conversion with optimization

```go
type LuaTypeConverter struct {
    // Caches for performance
    structCache  map[reflect.Type]*structInfo
    methodCache  map[reflect.Type]map[string]lua.LGFunction
    
    // Conversion strategies
    strategies   map[reflect.Type]ConversionStrategy
}

// Conversion flow
ScriptValue ↔ LValue ↔ Go Types ↔ go-llms Types
```

**Type Mappings**:
- **Primitives**: Direct mapping (bool, number, string, nil)
- **Collections**: Tables ↔ maps/slices with circular reference handling
- **Functions**: Wrapped with proper error propagation
- **Bridge Objects**: UserData with metatable methods
- **Channels**: LChannel for coroutine communication

### 3. Module System

**Design**: Lazy loading with dependency resolution

```go
type ModuleSystem struct {
    // Module registry
    modules      map[string]ModuleDefinition
    loaded       map[string]bool
    
    // Loading strategies
    preloadList  []string
    lazyModules  map[string]bool
    
    // Dependency graph
    dependencies map[string][]string
}

type ModuleDefinition struct {
    Name         string
    Dependencies []string
    InitFunc     func() error
    LoadFunc     lua.LGFunction
    Priority     int
}
```

**Loading Strategy**:
1. Core modules preloaded (spell, error handling)
2. Bridge modules loaded on demand
3. Profile-based module sets (minimal, standard, full)
4. Circular dependency detection

### 4. Security Sandbox

**Design**: Multi-layer security with configurable policies

```go
type SecuritySandbox struct {
    // Library restrictions
    allowedLibraries []string
    deniedFunctions  map[string]bool
    
    // Resource limits
    instructionLimit int
    memoryLimit      int
    timeout          time.Duration
    
    // Monitoring hooks
    instructionHook  lua.LHook
    memoryHook       lua.LHook
    
    // Security levels
    level            SecurityLevel
}
```

**Security Levels**:
- **Minimal**: Basic restrictions, most libraries available
- **Standard**: No file/network access, safe libraries only
- **Strict**: Minimal libraries, aggressive limits

### 5. Bridge Integration Layer

**Design**: UserData-based bridge objects with type safety

```go
type BridgeUserData struct {
    bridge      bridge.Bridge
    bridgeType  string
    methods     map[string]lua.LGFunction
}

// Bridge registration
func RegisterBridge(L *lua.LState, name string, b bridge.Bridge) {
    mt := L.NewTypeMetatable(name)
    
    // Method table
    methods := L.NewTable()
    for _, method := range b.Methods() {
        methods.RawSetString(method.Name, 
            L.NewFunction(createBridgeMethod(b, method)))
    }
    
    L.SetField(mt, "__index", methods)
    L.SetField(mt, "__tostring", L.NewFunction(bridgeToString))
    
    // Prevent modification
    L.SetField(mt, "__newindex", L.NewFunction(denyModification))
}
```

### 6. Async/Coroutine Support

**Design**: Promise-based async with coroutine integration

```go
type AsyncRuntime struct {
    promises     map[string]*Promise
    coroutines   map[*lua.LState]CoroutineInfo
    executor     *AsyncExecutor
}

type Promise struct {
    state    PromiseState
    value    lua.LValue
    error    error
    handlers []PromiseHandler
}

// Async/await pattern
async(function()
    local result = await(bridge:asyncMethod())
    return result
end)
```

### 7. Error Handling

**Design**: Enhanced error tracking with stack preservation

```go
type ErrorHandler struct {
    // Error enhancement
    stackCapture  *StackTraceCapture
    contextStore  *ErrorContext
    
    // Recovery strategies
    retryPolicy   *RetryPolicy
    fallbacks     map[string]RecoveryStrategy
}

type EnhancedError struct {
    Original    error
    StackTrace  []StackFrame
    Context     map[string]interface{}
    Timestamp   time.Time
    ScriptInfo  ScriptLocation
}
```

### 8. Resource Monitoring

**Design**: Comprehensive resource tracking and limits

```go
type ResourceMonitor struct {
    // Limits
    limits       ResourceLimits
    
    // Current usage
    instructions int64
    memory       int64
    startTime    time.Time
    
    // Enforcement
    enforcer     *ResourceEnforcer
    
    // Metrics
    collector    *MetricsCollector
}

type ResourceLimits struct {
    MaxInstructions int64
    MaxMemory       int64
    MaxDuration     time.Duration
    CheckInterval   int
}
```

### 9. Timeout Handling & State Abandonment

**Design**: Safe handling of script timeouts without race conditions

```go
// State abandonment for timeout scenarios
func (p *LStatePool) AbandonState(state *lua.LState) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    pooledState, exists := p.inUse[state]
    if !exists {
        return // State not from this pool
    }
    
    // Mark as not executing
    pooledState.mu.Lock()
    pooledState.executing = false
    if pooledState.done != nil {
        close(pooledState.done)
    }
    pooledState.mu.Unlock()
    
    // Remove from inUse tracking
    delete(p.inUse, state)
    
    // Update metrics
    atomic.AddInt64(&p.metrics.inUse, -1)
    atomic.AddInt64(&p.metrics.recycled, 1)
    
    // Note: We do NOT close the state here
    // It will be garbage collected when PCall completes
}

// Execution pipeline timeout handling
func (ep *ExecutionPipeline) executeScript(execCtx *ExecutionContext) error {
    resultChan := make(chan error, 1)
    
    go func() {
        err := execCtx.State.PCall(0, lua.MultRet, nil)
        select {
        case resultChan <- err:
        default: // Timed out, discard result
        }
    }()
    
    select {
    case err := <-resultChan:
        return err
    case <-execCtx.Context.Done():
        // Abandon state instead of closing
        abandonedState := execCtx.State
        execCtx.State = nil // Prevent normal cleanup
        ep.pool.AbandonState(abandonedState)
        return ErrTimeout
    }
}
```

**Key Design Decision**: 
- Never close an LState while PCall() is executing (not thread-safe)
- Abandoned states complete naturally and are garbage collected
- Pool tracks executing states and waits during shutdown
- Prevents data races while maintaining clean timeout behavior

```go
// Pool shutdown waits for executing states
func (p *LStatePool) Shutdown(ctx context.Context) error {
    // Wait for executing states with timeout
    waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    for {
        p.mu.RLock()
        hasExecuting := false
        for _, pooledState := range p.inUse {
            pooledState.mu.Lock()
            if pooledState.executing {
                hasExecuting = true
            }
            pooledState.mu.Unlock()
            if hasExecuting {
                break
            }
        }
        p.mu.RUnlock()
        
        if !hasExecuting {
            break
        }
        
        select {
        case <-waitCtx.Done():
            return fmt.Errorf("timeout waiting for executing states")
        case <-time.After(10 * time.Millisecond):
            // Check again
        }
    }
    
    // Safe to close all states now
    for _, state := range p.states {
        state.Close()
    }
    return nil
}
```

## Implementation Design

### Engine Implementation Structure

```go
package gopherlua

type LuaEngine struct {
    // Core components
    pool         *LStatePool
    factory      *LStateFactory
    modules      *ModuleSystem
    security     *SecurityManager
    
    // Type system
    converter    *LuaTypeConverter
    
    // Bridge registry
    bridges      map[string]bridge.Bridge
    
    // Configuration
    config       EngineConfig
    
    // Metrics
    metrics      *EngineMetrics
}

// ScriptEngine interface implementation
func (e *LuaEngine) Initialize(config engine.EngineConfig) error {
    e.config = config
    
    // Initialize components
    e.factory = NewLStateFactory(config)
    e.pool = NewLStatePool(e.factory, config.PoolSize)
    e.modules = NewModuleSystem()
    e.security = NewSecurityManager(config.Security)
    e.converter = NewLuaTypeConverter()
    
    // Start background workers
    e.pool.Start()
    e.metrics.Start()
    
    return nil
}

func (e *LuaEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
    // Get state from pool
    state, err := e.pool.Get()
    if err != nil {
        return nil, err
    }
    defer e.pool.Put(state)
    
    // Apply security
    if err := e.security.ApplySandbox(state.L); err != nil {
        return nil, err
    }
    
    // Set execution context
    state.L.SetContext(ctx)
    
    // Install resource monitors
    monitor := e.security.CreateMonitor(state.L)
    defer monitor.Stop()
    
    // Convert parameters
    for k, v := range params {
        state.L.SetGlobal(k, e.converter.ToLua(state.L, v))
    }
    
    // Execute script with timeout handling via pipeline
    pipeline := NewExecutionPipeline(e)
    result, err := pipeline.Execute(ctx, script, params)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

### State Factory Configuration

```go
type LStateFactory struct {
    // Options
    options      lua.Options
    
    // Libraries to load
    libraries    []LibraryLoader
    
    // Preload modules
    preloadMods  map[string]lua.LGFunction
    
    // Initialization
    initScript   string
    warmupFunc   WarmupStrategy
}

func (f *LStateFactory) Create() (*lua.LState, error) {
    // Create with options
    L := lua.NewState(f.options)
    
    // Load libraries based on security level
    for _, loader := range f.libraries {
        if err := loader.Load(L); err != nil {
            L.Close()
            return nil, err
        }
    }
    
    // Preload modules
    for name, module := range f.preloadMods {
        L.PreloadModule(name, module)
    }
    
    // Run initialization
    if f.initScript != "" {
        if err := L.DoString(f.initScript); err != nil {
            L.Close()
            return nil, err
        }
    }
    
    // Warmup for performance
    if f.warmupFunc != nil {
        f.warmupFunc(L)
    }
    
    return L, nil
}
```

### Module Implementation Pattern

```go
// Standard module structure
func CreateLLMModule(engine *LuaEngine) lua.LGFunction {
    return func(L *lua.LState) int {
        module := L.NewTable()
        
        // Module metadata
        L.SetField(module, "_VERSION", lua.LString("1.0.0"))
        L.SetField(module, "_DESCRIPTION", lua.LString("LLM bridge module"))
        
        // Constructors
        L.SetField(module, "agent", L.NewFunction(createAgent))
        L.SetField(module, "complete", L.NewFunction(complete))
        
        // Async variants
        L.SetField(module, "completeAsync", L.NewFunction(completeAsync))
        
        // Push module
        L.Push(module)
        return 1
    }
}

// Bridge method implementation
func createAgent(L *lua.LState) int {
    config := L.CheckTable(1)
    
    // Convert config
    agentConfig := convertAgentConfig(L, config)
    
    // Get bridge
    llmBridge := getBridge(L, "llm").(*bridge.LLMBridge)
    
    // Create agent via bridge
    agent, err := llmBridge.CreateAgent(agentConfig)
    if err != nil {
        L.RaiseError("failed to create agent: %v", err)
        return 0
    }
    
    // Wrap in UserData
    ud := L.NewUserData()
    ud.Value = &AgentUserData{
        agent: agent,
        id:    generateID(),
    }
    L.SetMetatable(ud, L.GetTypeMetatable("Agent"))
    
    L.Push(ud)
    return 1
}
```

## Bridge Integration

### Bridge Registration Flow

```go
func (e *LuaEngine) RegisterBridge(name string, b bridge.Bridge) error {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    // Store bridge
    e.bridges[name] = b
    
    // Create module loader
    loader := createBridgeModule(e, name, b)
    
    // Register with module system
    e.modules.Register(ModuleDefinition{
        Name:     name,
        LoadFunc: loader,
        Priority: 10, // Bridge modules load after core
    })
    
    return nil
}

func createBridgeModule(engine *LuaEngine, name string, b bridge.Bridge) lua.LGFunction {
    return func(L *lua.LState) int {
        // Create module table
        module := L.NewTable()
        
        // Add bridge info
        info := b.GetMetadata()
        L.SetField(module, "_bridge", lua.LString(info.Name))
        L.SetField(module, "_version", lua.LString(info.Version))
        
        // Register methods
        for _, method := range b.Methods() {
            fn := createBridgeMethod(engine, b, method)
            L.SetField(module, method.Name, L.NewFunction(fn))
        }
        
        L.Push(module)
        return 1
    }
}
```

### Type Conversion for Bridges

```go
func createBridgeMethod(engine *LuaEngine, b bridge.Bridge, method MethodInfo) lua.LGFunction {
    return func(L *lua.LState) int {
        // Collect arguments
        args := make([]interface{}, L.GetTop())
        for i := 1; i <= L.GetTop(); i++ {
            args[i-1] = engine.converter.FromLua(L.Get(i))
        }
        
        // Call bridge method
        result, err := b.Call(method.Name, args...)
        if err != nil {
            L.Push(lua.LNil)
            L.Push(lua.LString(err.Error()))
            return 2
        }
        
        // Convert result
        L.Push(engine.converter.ToLua(L, result))
        return 1
    }
}
```

## Performance & Optimization

### 1. State Pool Optimization

```go
// Adaptive pool sizing
func (p *LStatePool) adjustPoolSize() {
    usage := p.metrics.getUsageRatio()
    current := len(p.states)
    
    if usage > p.scaleUpThreshold && current < p.maxSize {
        newSize := int(float64(current) * 1.2)
        p.expand(newSize - current)
    } else if usage < p.scaleDownThreshold && current > p.minSize {
        newSize := int(float64(current) * 0.8)
        p.shrink(current - newSize)
    }
}
```

### 2. Chunk Caching

```go
type ChunkCache struct {
    cache    map[string]*CompiledChunk
    maxSize  int
    lru      *list.List
}

type CompiledChunk struct {
    Key      string
    Proto    *lua.FunctionProto
    Size     int
    LastUsed time.Time
}

func (cc *ChunkCache) GetOrCompile(L *lua.LState, source, name string) (*lua.FunctionProto, error) {
    key := cc.generateKey(source, name)
    
    if chunk, ok := cc.cache[key]; ok {
        cc.touch(chunk)
        return chunk.Proto, nil
    }
    
    // Compile and cache
    proto, err := parse.Parse(strings.NewReader(source), name)
    if err != nil {
        return nil, err
    }
    
    cc.add(key, proto)
    return proto, nil
}
```

### 3. Memory Management

```go
// Memory-aware execution
func (e *LuaEngine) executeWithMemoryLimit(L *lua.LState, script string, limit int64) error {
    monitor := &MemoryMonitor{
        limit:    limit,
        interval: 1000,
    }
    
    L.SetHook(monitor.Hook, lua.Count, monitor.interval)
    defer L.RemoveHook()
    
    return L.DoString(script)
}
```

### 4. Optimization Strategies

1. **Lazy Module Loading**: Load only required modules
2. **State Warmup**: Pre-JIT common operations
3. **Type Conversion Caching**: Cache struct reflection
4. **Batch Operations**: Group related operations
5. **Coroutine Pooling**: Reuse coroutines

## Security Model

### Security Layers

1. **Library Restrictions**
   ```go
   type LibraryRestrictions struct {
       Allowed   []string // Whitelist
       Denied    []string // Blacklist
       Custom    map[string]lua.LGFunction // Replacements
   }
   ```

2. **Resource Limits**
   ```go
   type ResourceLimits struct {
       CPU      CPULimits
       Memory   MemoryLimits
       IO       IOLimits
       Network  NetworkLimits
   }
   ```

3. **Sandbox Enforcement**
   ```go
   func (s *SecurityManager) ApplySandbox(L *lua.LState) error {
       // Remove dangerous functions
       s.removeDangerousFunctions(L)
       
       // Apply resource hooks
       s.installResourceHooks(L)
       
       // Set execution limits
       s.configureExecution(L)
       
       return nil
   }
   ```

### Security Profiles

```go
var SecurityProfiles = map[string]SecurityConfig{
    "minimal": {
        Libraries: []string{"base", "string", "table", "math", "coroutine", "os", "io"},
        Limits: ResourceLimits{
            Instructions: 100_000_000,
            Memory:       100 * MB,
            Timeout:      5 * time.Minute,
        },
    },
    "standard": {
        Libraries: []string{"base", "string", "table", "math", "coroutine"},
        Limits: ResourceLimits{
            Instructions: 10_000_000,
            Memory:       50 * MB,
            Timeout:      30 * time.Second,
        },
    },
    "strict": {
        Libraries: []string{"base", "string", "table", "math"},
        Limits: ResourceLimits{
            Instructions: 1_000_000,
            Memory:       10 * MB,
            Timeout:      5 * time.Second,
        },
    },
}
```

## Development Roadmap

### Phase 1: Core Engine (Current)
- [x] LState management and pooling
- [x] Type conversion system
- [x] Module preloading
- [x] Error handling
- [x] Resource limits
- [x] Security sandbox
- [ ] Basic bridge integration
- [ ] Unit tests

### Phase 2: Bridge Integration
- [ ] LLM bridge implementation
- [ ] Tools bridge implementation
- [ ] State management bridge
- [ ] Workflow bridge
- [ ] Event system bridge

### Phase 3: Advanced Features
- [ ] Async/await implementation
- [ ] Stream processing
- [ ] Coroutine pools
- [ ] Advanced error recovery
- [ ] Performance profiling

### Phase 4: Production Readiness
- [ ] Comprehensive testing
- [ ] Performance optimization
- [ ] Security hardening
- [ ] Documentation
- [ ] Examples and tutorials

## Testing Strategy

### Test Categories

1. **Unit Tests**
   - Type conversion accuracy
   - State pool management
   - Module loading
   - Security restrictions

2. **Integration Tests**
   - Bridge functionality
   - Cross-module interaction
   - Error propagation
   - Resource limits

3. **Performance Tests**
   - Execution benchmarks
   - Memory usage
   - Concurrent execution
   - State pool efficiency

4. **Security Tests**
   - Sandbox escape attempts
   - Resource exhaustion
   - Malicious scripts
   - Permission violations

### Test Implementation

```go
func TestLuaEngine(t *testing.T) {
    engine := NewLuaEngine(DefaultConfig())
    
    t.Run("BasicExecution", func(t *testing.T) {
        result, err := engine.Execute(context.Background(), 
            `return 1 + 1`, nil)
        assert.NoError(t, err)
        assert.Equal(t, 2.0, result)
    })
    
    t.Run("BridgeIntegration", func(t *testing.T) {
        mockBridge := &MockLLMBridge{}
        engine.RegisterBridge("llm", mockBridge)
        
        _, err := engine.Execute(context.Background(), `
            local llm = require("llm")
            return llm.complete({prompt = "test"})
        `, nil)
        assert.NoError(t, err)
    })
}
```

## API Reference

### Engine API

```go
// Core engine interface
type LuaEngine interface {
    engine.ScriptEngine
    
    // Lua-specific methods
    SetGlobal(name string, value interface{}) error
    GetGlobal(name string) (interface{}, error)
    
    // Module management
    PreloadModule(name string, loader lua.LGFunction) error
    RequireModule(name string) error
    
    // Coroutine support
    CreateCoroutine(fn lua.LValue) (*lua.LState, error)
    ResumeCoroutine(co *lua.LState, args ...lua.LValue) ([]lua.LValue, error)
}
```

### Script API

```lua
-- Core modules
local spell = require("spell")     -- Core utilities
local async = require("async")     -- Async/await support
local errors = require("errors")   -- Error handling

-- Bridge modules  
local llm = require("llm")         -- LLM operations
local tools = require("tools")     -- Tool system
local state = require("state")     -- State management
local workflow = require("workflow") -- Workflows
local events = require("events")   -- Event system

-- Example usage
local agent = llm.agent({
    model = "gpt-4",
    temperature = 0.7
})

local response = await(agent:generateAsync("Hello!"))
```

## Conclusion

This architecture provides a robust, secure, and performant Lua scripting environment for go-llmspell. By leveraging GopherLua's mature implementation and combining it with careful design of state management, type conversion, security, and bridge integration, we create a powerful platform for scripting LLM interactions while maintaining the project's core philosophy of being a pure bridge to go-llms functionality.

The modular design allows for incremental implementation, starting with core engine functionality and progressively adding bridge integrations and advanced features. The comprehensive testing strategy ensures reliability, while the performance optimizations enable efficient execution even under heavy load.

This design serves as the blueprint for implementing the GopherLua engine in go-llmspell, providing clear guidance for development while remaining flexible enough to accommodate future enhancements.