# LState Lifecycle Management Research

This document investigates comprehensive lifecycle management for GopherLua LState instances, including creation patterns, pooling strategies, and cleanup procedures.

## Executive Summary

Effective LState lifecycle management is crucial for performance and resource efficiency. This involves careful state creation, intelligent pooling with health monitoring, and thorough cleanup procedures to prevent resource leaks.

## LState Lifecycle Phases

### 1. Creation Phase

```go
// LState creation involves:
// 1. Memory allocation
// 2. Registry initialization
// 3. Standard library loading
// 4. Custom configuration

type LStateFactory struct {
    options      lua.Options
    libraries    []LibraryLoader
    preloadMods  map[string]lua.LGFunction
    initScript   string
}

type LibraryLoader func(L *lua.LState)

func (lsf *LStateFactory) Create() (*lua.LState, error) {
    // Phase 1: Basic creation
    L := lua.NewState(lsf.options)
    
    // Phase 2: Library loading
    for _, loader := range lsf.libraries {
        loader(L)
    }
    
    // Phase 3: Module preloading
    for name, loader := range lsf.preloadMods {
        L.PreloadModule(name, loader)
    }
    
    // Phase 4: Initialization script
    if lsf.initScript != "" {
        if err := L.DoString(lsf.initScript); err != nil {
            L.Close()
            return nil, fmt.Errorf("init script failed: %w", err)
        }
    }
    
    return L, nil
}
```

### 2. Active Phase

During the active phase, LState is used for script execution:

```go
type LStateMetrics struct {
    CreatedAt    time.Time
    LastUsedAt   time.Time
    UseCount     int64
    ErrorCount   int64
    TotalRuntime time.Duration
    MemoryPeak   int
}

type ManagedLState struct {
    *lua.LState
    metrics  LStateMetrics
    healthy  bool
    inUse    bool
    mu       sync.Mutex
}

func (mls *ManagedLState) Execute(script string) error {
    mls.mu.Lock()
    mls.inUse = true
    mls.metrics.UseCount++
    startTime := time.Now()
    mls.mu.Unlock()
    
    defer func() {
        mls.mu.Lock()
        mls.inUse = false
        mls.metrics.LastUsedAt = time.Now()
        mls.metrics.TotalRuntime += time.Since(startTime)
        mls.mu.Unlock()
    }()
    
    err := mls.DoString(script)
    if err != nil {
        mls.mu.Lock()
        mls.metrics.ErrorCount++
        mls.mu.Unlock()
    }
    
    return err
}
```

### 3. Cleanup Phase

Proper cleanup is essential to prevent resource leaks:

```go
type LStateCleanup struct {
    resetGlobals    bool
    clearRegistry   bool
    collectGarbage  bool
    resetStack      bool
    clearModules    bool
}

func (lsc *LStateCleanup) Clean(L *lua.LState) error {
    // Step 1: Clear the stack
    if lsc.resetStack {
        L.SetTop(0)
    }
    
    // Step 2: Garbage collection
    if lsc.collectGarbage {
        L.DoString(`collectgarbage("collect")`)
        L.DoString(`collectgarbage("collect")`) // Double collect
    }
    
    // Step 3: Clear loaded modules
    if lsc.clearModules {
        L.DoString(`
            for k, v in pairs(package.loaded) do
                if type(k) == "string" and not k:match("^_") then
                    package.loaded[k] = nil
                end
            end
        `)
    }
    
    // Step 4: Reset globals (careful - preserves standard libs)
    if lsc.resetGlobals {
        if err := resetGlobalNamespace(L); err != nil {
            return err
        }
    }
    
    // Step 5: Clear registry (except critical entries)
    if lsc.clearRegistry {
        cleanRegistry(L)
    }
    
    return nil
}

func resetGlobalNamespace(L *lua.LState) error {
    return L.DoString(`
        -- Save standard libraries
        local saved = {}
        local standard = {
            "string", "table", "math", "io", "os", "debug",
            "package", "coroutine", "bit32", "_G", "_VERSION"
        }
        
        for _, name in ipairs(standard) do
            saved[name] = _G[name]
        end
        
        -- Clear globals
        for k, v in pairs(_G) do
            if not saved[k] then
                _G[k] = nil
            end
        end
    `)
}
```

## Advanced Pooling Strategies

### 1. Adaptive Pool Management

```go
type AdaptivePool struct {
    factory      *LStateFactory
    cleanup      *LStateCleanup
    
    // Pool configuration
    minSize      int
    maxSize      int
    targetSize   int
    
    // Adaptive parameters
    scaleUpThreshold   float64 // Usage ratio to trigger scale up
    scaleDownThreshold float64 // Usage ratio to trigger scale down
    scaleFactor        float64 // Multiplication factor for scaling
    
    // State tracking
    states       []*PooledState
    available    chan *PooledState
    metrics      PoolMetrics
    mu           sync.RWMutex
    
    // Background management
    quit         chan struct{}
    wg           sync.WaitGroup
}

type PooledState struct {
    state       *ManagedLState
    pooledAt    time.Time
    generation  int
    healthScore float64
}

func (ap *AdaptivePool) Start() {
    // Pre-warm pool
    for i := 0; i < ap.minSize; i++ {
        if state, err := ap.createPooledState(); err == nil {
            ap.available <- state
        }
    }
    
    // Start background workers
    ap.wg.Add(2)
    go ap.monitorLoop()
    go ap.maintenanceLoop()
}

func (ap *AdaptivePool) monitorLoop() {
    defer ap.wg.Done()
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ap.adjustPoolSize()
        case <-ap.quit:
            return
        }
    }
}

func (ap *AdaptivePool) adjustPoolSize() {
    ap.mu.Lock()
    defer ap.mu.Unlock()
    
    usage := ap.metrics.getUsageRatio()
    currentSize := len(ap.states)
    
    if usage > ap.scaleUpThreshold && currentSize < ap.maxSize {
        // Scale up
        newSize := int(float64(currentSize) * ap.scaleFactor)
        if newSize > ap.maxSize {
            newSize = ap.maxSize
        }
        
        for i := currentSize; i < newSize; i++ {
            ap.wg.Add(1)
            go func() {
                defer ap.wg.Done()
                if state, err := ap.createPooledState(); err == nil {
                    ap.available <- state
                }
            }()
        }
        
        ap.targetSize = newSize
    } else if usage < ap.scaleDownThreshold && currentSize > ap.minSize {
        // Scale down
        newSize := int(float64(currentSize) / ap.scaleFactor)
        if newSize < ap.minSize {
            newSize = ap.minSize
        }
        
        ap.targetSize = newSize
    }
}
```

### 2. Health-Based State Management

```go
type StateHealthChecker struct {
    memoryThreshold   int
    errorRateLimit    float64
    maxAge            time.Duration
    maxUseCount       int64
    performanceDecay  float64
}

func (shc *StateHealthChecker) CheckHealth(ps *PooledState) HealthStatus {
    state := ps.state
    metrics := state.metrics
    
    health := HealthStatus{
        Healthy: true,
        Score:   100.0,
        Reasons: []string{},
    }
    
    // Check memory usage
    memUsage := state.MemUsage()
    if memUsage > shc.memoryThreshold {
        health.Score -= 20
        health.Reasons = append(health.Reasons, 
            fmt.Sprintf("high memory: %d bytes", memUsage))
    }
    
    // Check error rate
    if metrics.UseCount > 0 {
        errorRate := float64(metrics.ErrorCount) / float64(metrics.UseCount)
        if errorRate > shc.errorRateLimit {
            health.Score -= 30
            health.Reasons = append(health.Reasons,
                fmt.Sprintf("high error rate: %.2f%%", errorRate*100))
        }
    }
    
    // Check age
    age := time.Since(metrics.CreatedAt)
    if age > shc.maxAge {
        health.Score -= 25
        health.Reasons = append(health.Reasons,
            fmt.Sprintf("old state: %v", age))
    }
    
    // Check use count
    if metrics.UseCount > shc.maxUseCount {
        health.Score -= 15
        health.Reasons = append(health.Reasons,
            fmt.Sprintf("high use count: %d", metrics.UseCount))
    }
    
    // Apply performance decay
    avgRuntime := metrics.TotalRuntime / time.Duration(max(metrics.UseCount, 1))
    if avgRuntime > 100*time.Millisecond {
        decay := shc.performanceDecay * float64(avgRuntime.Milliseconds())
        health.Score -= decay
        health.Reasons = append(health.Reasons,
            fmt.Sprintf("slow performance: %v avg", avgRuntime))
    }
    
    health.Healthy = health.Score > 50
    ps.healthScore = health.Score
    
    return health
}

type HealthStatus struct {
    Healthy bool
    Score   float64
    Reasons []string
}
```

### 3. Generation-Based Recycling

```go
type GenerationalPool struct {
    *AdaptivePool
    
    maxGeneration    int
    generationSize   int
    currentGen       int
    recycleThreshold float64
}

func (gp *GenerationalPool) Get() (*ManagedLState, error) {
    select {
    case ps := <-gp.available:
        // Check if state needs recycling
        if gp.shouldRecycle(ps) {
            go gp.recycleState(ps)
            return gp.Get() // Recursive call to get another
        }
        
        return ps.state, nil
        
    case <-time.After(100 * time.Millisecond):
        // Create new state if pool is empty
        return gp.createNewState()
    }
}

func (gp *GenerationalPool) shouldRecycle(ps *PooledState) bool {
    // Check generation
    if ps.generation < gp.currentGen-1 {
        return true
    }
    
    // Check health score
    if ps.healthScore < gp.recycleThreshold {
        return true
    }
    
    // Check staleness
    if time.Since(ps.pooledAt) > 5*time.Minute {
        return true
    }
    
    return false
}

func (gp *GenerationalPool) recycleState(ps *PooledState) {
    // Close old state
    ps.state.Close()
    
    // Create new state for current generation
    if newPs, err := gp.createPooledState(); err == nil {
        newPs.generation = gp.currentGen
        
        select {
        case gp.available <- newPs:
        default:
            // Pool is full, close the state
            newPs.state.Close()
        }
    }
}
```

## State Isolation and Security

### 1. Sandboxed State Creation

```go
type SandboxConfig struct {
    AllowedLibraries []string
    MaxMemory        int
    MaxCPU           time.Duration
    AllowFileAccess  bool
    AllowNetAccess   bool
}

func CreateSandboxedState(config SandboxConfig) (*lua.LState, error) {
    opts := lua.Options{
        RegistryMaxSize:     config.MaxMemory / 32,
        MinimizeStackMemory: true,
    }
    
    L := lua.NewState(opts)
    
    // Load only allowed libraries
    for _, lib := range config.AllowedLibraries {
        switch lib {
        case "base":
            L.OpenBase()
        case "string":
            L.OpenString()
        case "table":
            L.OpenTable()
        case "math":
            L.OpenMath()
        case "coroutine":
            L.OpenCoroutine()
        // Dangerous libraries excluded by default
        }
    }
    
    // Remove dangerous functions
    L.DoString(`
        io = nil
        os.execute = nil
        os.exit = nil
        os.setenv = nil
        loadfile = nil
        dofile = nil
    `)
    
    // Install resource limits
    InstallResourceLimits(L, config)
    
    return L, nil
}
```

### 2. State Checkpoint and Restore

```go
type StateCheckpoint struct {
    Globals   map[string]interface{}
    Registry  map[string]interface{}
    Modules   map[string]bool
    Timestamp time.Time
}

func CreateCheckpoint(L *lua.LState) (*StateCheckpoint, error) {
    checkpoint := &StateCheckpoint{
        Globals:   make(map[string]interface{}),
        Registry:  make(map[string]interface{}),
        Modules:   make(map[string]bool),
        Timestamp: time.Now(),
    }
    
    // Capture globals
    err := L.DoString(`
        _CHECKPOINT_GLOBALS = {}
        for k, v in pairs(_G) do
            if type(v) ~= "function" and type(v) ~= "userdata" then
                _CHECKPOINT_GLOBALS[k] = v
            end
        end
    `)
    if err != nil {
        return nil, err
    }
    
    // Convert to Go map
    L.GetGlobal("_CHECKPOINT_GLOBALS")
    if globals, ok := L.Get(-1).(*lua.LTable); ok {
        globals.ForEach(func(k, v lua.LValue) {
            checkpoint.Globals[k.String()] = luaValueToGo(v)
        })
    }
    L.Pop(1)
    
    // Capture loaded modules
    L.DoString(`
        for k, v in pairs(package.loaded) do
            _CHECKPOINT_MODULES = _CHECKPOINT_MODULES or {}
            _CHECKPOINT_MODULES[k] = true
        end
    `)
    
    return checkpoint, nil
}

func RestoreCheckpoint(L *lua.LState, checkpoint *StateCheckpoint) error {
    // Clear current state
    cleanup := &LStateCleanup{
        resetGlobals:   true,
        clearRegistry:  true,
        collectGarbage: true,
        resetStack:     true,
        clearModules:   true,
    }
    
    if err := cleanup.Clean(L); err != nil {
        return err
    }
    
    // Restore globals
    for k, v := range checkpoint.Globals {
        L.SetGlobal(k, goValueToLua(L, v))
    }
    
    return nil
}
```

## Lifecycle Event Handling

### 1. Event-Driven Lifecycle

```go
type LifecycleEventType int

const (
    EventCreated LifecycleEventType = iota
    EventActivated
    EventDeactivated
    EventRecycled
    EventDestroyed
    EventError
)

type LifecycleEvent struct {
    Type      LifecycleEventType
    State     *ManagedLState
    Timestamp time.Time
    Metadata  map[string]interface{}
}

type LifecycleManager struct {
    handlers map[LifecycleEventType][]LifecycleHandler
    events   chan LifecycleEvent
    mu       sync.RWMutex
}

type LifecycleHandler func(event LifecycleEvent)

func (lm *LifecycleManager) RegisterHandler(eventType LifecycleEventType, handler LifecycleHandler) {
    lm.mu.Lock()
    defer lm.mu.Unlock()
    
    lm.handlers[eventType] = append(lm.handlers[eventType], handler)
}

func (lm *LifecycleManager) EmitEvent(event LifecycleEvent) {
    select {
    case lm.events <- event:
    default:
        // Event queue full, log and continue
    }
}

func (lm *LifecycleManager) processEvents() {
    for event := range lm.events {
        lm.mu.RLock()
        handlers := lm.handlers[event.Type]
        lm.mu.RUnlock()
        
        for _, handler := range handlers {
            handler(event)
        }
    }
}
```

### 2. Lifecycle Hooks in Lua

```go
func InstallLifecycleHooks(L *lua.LState, lm *LifecycleManager) {
    hooks := L.NewTable()
    
    // onCreate hook
    L.SetField(hooks, "onCreate", L.NewFunction(func(L *lua.LState) int {
        fn := L.CheckFunction(1)
        
        lm.RegisterHandler(EventCreated, func(event LifecycleEvent) {
            if event.State.LState == L {
                L.Push(fn)
                L.Call(0, 0)
            }
        })
        
        return 0
    }))
    
    // onActivate hook
    L.SetField(hooks, "onActivate", L.NewFunction(func(L *lua.LState) int {
        fn := L.CheckFunction(1)
        
        lm.RegisterHandler(EventActivated, func(event LifecycleEvent) {
            if event.State.LState == L {
                L.Push(fn)
                L.Call(0, 0)
            }
        })
        
        return 0
    }))
    
    L.SetGlobal("lifecycle", hooks)
}
```

## Performance Optimization

### 1. Warm-up Strategies

```go
type WarmupStrategy struct {
    preloadScripts []string
    warmupScript   string
    jitWarmup      bool
}

func (ws *WarmupStrategy) Warmup(L *lua.LState) error {
    // Preload common scripts
    for _, script := range ws.preloadScripts {
        if _, err := L.LoadString(script); err != nil {
            return fmt.Errorf("warmup preload failed: %w", err)
        }
    }
    
    // Run warmup script
    if ws.warmupScript != "" {
        if err := L.DoString(ws.warmupScript); err != nil {
            return fmt.Errorf("warmup script failed: %w", err)
        }
    }
    
    // JIT warmup (if applicable)
    if ws.jitWarmup {
        L.DoString(`
            -- Run hot functions to trigger JIT compilation
            for i = 1, 1000 do
                -- Common operations
                local t = {}
                for j = 1, 10 do
                    t[j] = j * 2
                end
                table.sort(t)
            end
        `)
    }
    
    // Clear warmup artifacts
    L.DoString(`collectgarbage("collect")`)
    L.SetTop(0)
    
    return nil
}
```

### 2. Memory-Efficient Pooling

```go
type MemoryEfficientPool struct {
    basePool     *AdaptivePool
    memoryBudget int64
    currentUsage int64
    mu           sync.Mutex
}

func (mep *MemoryEfficientPool) Get() (*ManagedLState, error) {
    mep.mu.Lock()
    defer mep.mu.Unlock()
    
    // Check memory budget
    state, err := mep.basePool.Get()
    if err != nil {
        return nil, err
    }
    
    stateMemory := int64(state.MemUsage())
    if mep.currentUsage+stateMemory > mep.memoryBudget {
        // Over budget, return state and error
        mep.basePool.Put(state)
        return nil, fmt.Errorf("memory budget exceeded")
    }
    
    mep.currentUsage += stateMemory
    return state, nil
}

func (mep *MemoryEfficientPool) Put(state *ManagedLState) {
    mep.mu.Lock()
    defer mep.mu.Unlock()
    
    stateMemory := int64(state.MemUsage())
    mep.currentUsage -= stateMemory
    
    // Run GC before returning to pool
    state.DoString(`collectgarbage("collect")`)
    
    mep.basePool.Put(state)
}
```

## Testing and Monitoring

### Lifecycle Testing

```go
func TestLifecycleManagement(t *testing.T) {
    pool := NewAdaptivePool(PoolConfig{
        MinSize: 2,
        MaxSize: 10,
        Factory: &LStateFactory{
            options: lua.Options{},
        },
    })
    
    pool.Start()
    defer pool.Stop()
    
    // Test state creation and pooling
    states := make([]*ManagedLState, 5)
    for i := 0; i < 5; i++ {
        state, err := pool.Get()
        if err != nil {
            t.Fatalf("Failed to get state: %v", err)
        }
        states[i] = state
    }
    
    // Return states to pool
    for _, state := range states {
        pool.Put(state)
    }
    
    // Verify pool metrics
    metrics := pool.GetMetrics()
    if metrics.TotalCreated < 5 {
        t.Errorf("Expected at least 5 states created, got %d", 
            metrics.TotalCreated)
    }
}
```

## Best Practices

1. **Pre-warm Pools**: Create states ahead of time to avoid latency
2. **Monitor Health**: Regular health checks prevent degraded performance
3. **Clean Between Uses**: Reset state to prevent cross-contamination
4. **Limit Lifetime**: Recycle states periodically to prevent memory leaks
5. **Profile Usage**: Track metrics to optimize pool configuration

## Implementation Checklist

- [ ] Basic state factory with configuration
- [ ] Managed state wrapper with metrics
- [ ] Cleanup procedures and state reset
- [ ] Adaptive pool with auto-scaling
- [ ] Health monitoring and scoring
- [ ] Generation-based recycling
- [ ] State isolation and sandboxing
- [ ] Checkpoint and restore functionality
- [ ] Lifecycle event system
- [ ] Performance optimization strategies

## Summary

Effective LState lifecycle management requires:
1. Structured creation with proper initialization
2. Intelligent pooling with health monitoring
3. Thorough cleanup between uses
4. Adaptive scaling based on demand
5. Security through isolation and sandboxing

This ensures optimal performance, resource efficiency, and reliability in production environments.