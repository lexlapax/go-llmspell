# LState Pool Implementation Design

## Design Goals
1. Thread-safe pool management
2. Efficient state reuse with proper isolation
3. Configurable sizing and limits
4. Graceful degradation under load
5. Comprehensive metrics and monitoring

## Pool Architecture

```go
// LStatePool manages a pool of Lua VM instances
type LStatePool struct {
    // Pool storage
    states   chan *pooledState
    mu       sync.RWMutex
    
    // Configuration
    config   PoolConfig
    
    // Lifecycle
    closed   atomic.Bool
    wg       sync.WaitGroup
    
    // Metrics
    metrics  PoolMetrics
    
    // State management
    factory  StateFactory
    resetter StateResetter
}

// pooledState wraps LState with metadata
type pooledState struct {
    L          *lua.LState
    id         string
    created    time.Time
    lastUsed   time.Time
    useCount   int64
    globalHash uint64 // For detecting modifications
}

// PoolConfig defines pool behavior
type PoolConfig struct {
    MinSize          int           // Minimum states to maintain
    MaxSize          int           // Maximum states allowed
    MaxIdleTime      time.Duration // Evict states idle longer than this
    MaxLifetime      time.Duration // Force recreate states older than this
    MaxUseCount      int64         // Recreate after N uses
    CreateTimeout    time.Duration // Timeout for state creation
    ResetTimeout     time.Duration // Timeout for state reset
    HealthCheckInterval time.Duration // How often to check pool health
}

// PoolMetrics tracks pool performance
type PoolMetrics struct {
    Created      atomic.Int64
    Destroyed    atomic.Int64
    Active       atomic.Int64
    Idle         atomic.Int64
    Resets       atomic.Int64
    ResetErrors  atomic.Int64
    Timeouts     atomic.Int64
    WaitTime     atomic.Int64 // Cumulative wait time in ns
}
```

## State Factory Pattern

```go
// StateFactory creates new LState instances
type StateFactory interface {
    Create() (*lua.LState, error)
    Configure(*lua.LState) error
}

// DefaultStateFactory with security settings
type DefaultStateFactory struct {
    options      lua.Options
    safeLibs     []string
    bridgeLoader BridgeLoader
}

func (f *DefaultStateFactory) Create() (*lua.LState, error) {
    L := lua.NewState(f.options)
    
    // Load only safe libraries
    for _, lib := range f.safeLibs {
        switch lib {
        case "base":
            L.OpenBase()
        case "table":
            L.OpenTable()
        case "string":
            L.OpenString()
        case "math":
            L.OpenMath()
        // Explicitly exclude: io, os, debug, package
        }
    }
    
    // Configure bridges
    if err := f.bridgeLoader.Load(L); err != nil {
        L.Close()
        return nil, err
    }
    
    return L, nil
}
```

## State Reset Strategy

```go
// StateResetter resets LState for reuse
type StateResetter interface {
    Reset(*lua.LState) error
    Validate(*lua.LState) error
}

// SafeStateResetter with isolation guarantees
type SafeStateResetter struct {
    globalSnapshot map[string]lua.LValue
    maxStackSize   int
}

func (r *SafeStateResetter) Reset(L *lua.LState) error {
    // 1. Clear stack completely
    L.SetTop(0)
    
    // 2. Reset global environment
    if err := r.resetGlobals(L); err != nil {
        return fmt.Errorf("failed to reset globals: %w", err)
    }
    
    // 3. Clear registry user values
    // Note: Keep system values intact
    
    // 4. Terminate any running coroutines
    // This is implicit when clearing stack
    
    // 5. Reset memory allocator stats if available
    
    return nil
}

func (r *SafeStateResetter) resetGlobals(L *lua.LState) error {
    global := L.Get(lua.GlobalsIndex).(*lua.LTable)
    
    // Option 1: Restore from snapshot
    toRemove := []string{}
    global.ForEach(func(k, v lua.LValue) {
        key := lua.LVAsString(k)
        if _, exists := r.globalSnapshot[key]; !exists {
            toRemove = append(toRemove, key)
        }
    })
    
    for _, key := range toRemove {
        L.SetGlobal(key, lua.LNil)
    }
    
    // Option 2: Whitelist approach (alternative)
    // Only keep known safe globals
    
    return nil
}
```

## Pool Operations

```go
// Get retrieves a state from the pool
func (p *LStatePool) Get(ctx context.Context) (*lua.LState, error) {
    if p.closed.Load() {
        return nil, ErrPoolClosed
    }
    
    start := time.Now()
    defer func() {
        p.metrics.WaitTime.Add(time.Since(start).Nanoseconds())
    }()
    
    select {
    case state := <-p.states:
        // Validate state health
        if p.shouldRecycle(state) {
            state.L.Close()
            p.metrics.Destroyed.Add(1)
            return p.createNew(ctx)
        }
        
        // Reset for reuse
        if err := p.resetter.Reset(state.L); err != nil {
            state.L.Close()
            p.metrics.ResetErrors.Add(1)
            return p.createNew(ctx)
        }
        
        state.lastUsed = time.Now()
        state.useCount++
        p.metrics.Active.Add(1)
        p.metrics.Idle.Add(-1)
        
        return state.L, nil
        
    case <-ctx.Done():
        p.metrics.Timeouts.Add(1)
        return nil, ctx.Err()
        
    default:
        // Pool empty, create new
        return p.createNew(ctx)
    }
}

// Put returns a state to the pool
func (p *LStatePool) Put(L *lua.LState) {
    if p.closed.Load() {
        L.Close()
        return
    }
    
    // Find the wrapped state
    state := p.findPooledState(L)
    if state == nil {
        L.Close() // Unknown state, close it
        return
    }
    
    p.metrics.Active.Add(-1)
    
    select {
    case p.states <- state:
        p.metrics.Idle.Add(1)
    default:
        // Pool full, close the state
        L.Close()
        p.metrics.Destroyed.Add(1)
    }
}
```

## Lifecycle Management

```go
// Start initializes the pool
func (p *LStatePool) Start(ctx context.Context) error {
    // Pre-warm pool to minimum size
    for i := 0; i < p.config.MinSize; i++ {
        state, err := p.createNew(ctx)
        if err != nil {
            return fmt.Errorf("failed to pre-warm pool: %w", err)
        }
        p.Put(state)
    }
    
    // Start health check routine
    p.wg.Add(1)
    go p.healthCheckLoop()
    
    return nil
}

// healthCheckLoop maintains pool health
func (p *LStatePool) healthCheckLoop() {
    defer p.wg.Done()
    
    ticker := time.NewTicker(p.config.HealthCheckInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        if p.closed.Load() {
            return
        }
        
        p.evictStale()
        p.maintainMinSize()
    }
}

// Shutdown gracefully closes the pool
func (p *LStatePool) Shutdown(timeout time.Duration) error {
    p.closed.Store(true)
    
    // Stop creating new states
    close(p.states)
    
    // Wait for health check to stop
    p.wg.Wait()
    
    // Close all remaining states
    count := 0
    for state := range p.states {
        state.L.Close()
        count++
    }
    
    p.metrics.Destroyed.Add(int64(count))
    
    return nil
}
```

## Integration Points

### With ScriptEngine
```go
type LuaEngine struct {
    pool    *LStatePool
    config  engine.EngineConfig
}

func (e *LuaEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
    L, err := e.pool.Get(ctx)
    if err != nil {
        return nil, err
    }
    defer e.pool.Put(L)
    
    // Execute script with L
    // ...
}
```

### Configuration Example
```go
poolConfig := PoolConfig{
    MinSize:       2,
    MaxSize:       10,
    MaxIdleTime:   5 * time.Minute,
    MaxLifetime:   30 * time.Minute,
    MaxUseCount:   1000,
    CreateTimeout: 5 * time.Second,
    ResetTimeout:  100 * time.Millisecond,
    HealthCheckInterval: 30 * time.Second,
}
```

## Testing Strategy

1. **Isolation Tests**: Verify no state leakage between executions
2. **Concurrency Tests**: Stress test with multiple goroutines
3. **Performance Tests**: Benchmark pool operations vs direct creation
4. **Resilience Tests**: Test behavior under various failure modes
5. **Memory Tests**: Verify no memory leaks with extended usage

## Monitoring and Metrics

```go
// Expose metrics for monitoring
func (p *LStatePool) Stats() PoolStats {
    return PoolStats{
        Created:     p.metrics.Created.Load(),
        Destroyed:   p.metrics.Destroyed.Load(),
        Active:      p.metrics.Active.Load(),
        Idle:        p.metrics.Idle.Load(),
        PoolSize:    len(p.states),
        Resets:      p.metrics.Resets.Load(),
        ResetErrors: p.metrics.ResetErrors.Load(),
        AvgWaitTime: time.Duration(p.metrics.WaitTime.Load() / p.metrics.Created.Load()),
    }
}
```

## Implementation Priority

1. Basic pool with Get/Put
2. State reset mechanism
3. Lifecycle management
4. Health checks and eviction
5. Metrics and monitoring
6. Advanced features (timeouts, recycling)