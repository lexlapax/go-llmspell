# LState Management and Pooling Analysis

## Overview
LState is the core VM instance in GopherLua. Proper management is critical for performance and safety in a concurrent environment.

## Thread Safety Analysis

### Core Principle
**LState is NOT thread-safe**. Each goroutine must have its own LState instance.

### Implications
1. Cannot share LState between goroutines
2. Must use synchronization for pool access
3. Communication between scripts via Go channels (LChannel)

## Pooling Strategy

### Why Pool?
- LState creation is expensive (~1-2ms per instance)
- Memory allocation overhead
- Avoids GC pressure from frequent create/destroy cycles

### Pool Implementation Pattern

```go
type LStatePool struct {
    mu    sync.Mutex
    saved []*lua.LState
    
    // Configuration
    maxSize      int
    createFunc   func() *lua.LState
    resetFunc    func(*lua.LState)
}

func (p *LStatePool) Get() *lua.LState {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if len(p.saved) > 0 {
        L := p.saved[len(p.saved)-1]
        p.saved = p.saved[:len(p.saved)-1]
        p.resetFunc(L)
        return L
    }
    
    return p.createFunc()
}

func (p *LStatePool) Put(L *lua.LState) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if len(p.saved) < p.maxSize {
        p.saved = append(p.saved, L)
    } else {
        L.Close() // Discard if pool is full
    }
}
```

## State Reset Requirements

### Between Script Executions
1. **Stack Cleanup**: `L.SetTop(0)` or `L.Pop(L.GetTop())`
2. **Global Environment**: Monitor and potentially reset
3. **Registry Cleanup**: Remove user-added values
4. **Coroutines**: Ensure all coroutines are terminated
5. **Channels**: Close any open channels

### Reset Implementation
```go
func resetLState(L *lua.LState) {
    // Clear stack
    L.SetTop(0)
    
    // Reset global environment if needed
    // Option 1: Keep pristine copy and restore
    // Option 2: Whitelist allowed globals
    // Option 3: Create new environment table
    
    // Clear any pending errors
    L.Push(lua.LNil)
    L.SetGlobal("_LAST_ERROR")
}
```

## Memory Management

### Registry Size Configuration
```go
func createLState() *lua.LState {
    L := lua.NewState(lua.Options{
        CallStackSize: 120,        // Default: 128
        RegistrySize:  1024*20,    // Default: 1024*20
        RegistryMaxSize: 1024*80,  // Default: 1024*80
        RegistryGrowStep: 32,      // Default: 32
        SkipOpenLibs: true,        // For security
    })
    return L
}
```

### Memory Considerations
1. **Registry Size**: Balance between memory usage and script requirements
2. **Call Stack**: Fixed size is fastest but less flexible
3. **Pool Size**: Consider memory vs creation overhead tradeoff

## Lifecycle Management

### Creation Phase
1. Create LState with appropriate options
2. Load safe libraries only
3. Preload bridge modules
4. Take snapshot of clean environment

### Usage Phase
1. Get from pool
2. Load/execute script
3. Handle results/errors
4. Reset state

### Return Phase
1. Clear stack
2. Reset globals if modified
3. Clear custom registry entries
4. Return to pool

### Shutdown Phase
1. Stop accepting new requests
2. Wait for active states to return
3. Close all pooled states
4. Clear pool

## Performance Optimization

### Benchmarks
- LState creation: ~1-2ms
- Reset operations: ~10-50μs
- Pool get/put: ~1-5μs

### Optimization Strategies
1. **Pre-warm Pool**: Create initial states at startup
2. **Lazy Expansion**: Grow pool on demand
3. **Size Limits**: Prevent unbounded growth
4. **Reset Optimization**: Minimal necessary cleanup

## Error Handling

### Pool Exhaustion
```go
func (p *LStatePool) GetWithTimeout(timeout time.Duration) (*lua.LState, error) {
    // Implement with channels and select
}
```

### Panic Recovery
```go
func safeLuaCall(L *lua.LState, fn lua.LGFunction) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("lua panic: %v", r)
        }
    }()
    // Execute function
}
```

## Integration with go-llmspell

### Pool per Engine Instance
- Each LuaEngine has its own pool
- Pool size based on expected concurrency
- Configurable via EngineConfig

### Bridge Considerations
- Bridges are stateless (good for pooling)
- Type converters must handle reset
- Module preloading in pool creation

### Security Implications
- Ensure complete isolation between executions
- No state leakage between scripts
- Careful with global environment modifications

## Best Practices

1. **Always Reset**: Never trust previous state
2. **Monitor Pool**: Track metrics (size, wait time, etc.)
3. **Graceful Degradation**: Handle pool exhaustion
4. **Clean Shutdown**: Properly close all states
5. **Test Isolation**: Verify no state leakage

## Implementation Checklist

- [ ] Pool struct with mutex protection
- [ ] Configurable pool size limits
- [ ] State creation with security options
- [ ] Comprehensive reset function
- [ ] Timeout-based Get operation
- [ ] Metrics collection
- [ ] Graceful shutdown mechanism
- [ ] Integration tests for isolation
- [ ] Benchmarks for pool operations
- [ ] Documentation for configuration