# LState Lifecycle Management Examples

This document provides practical examples of implementing LState lifecycle management in GopherLua.

## Basic State Management

### Simple State Creation and Cleanup

```go
package main

import (
    "fmt"
    "log"
    lua "github.com/yuin/gopher-lua"
)

func main() {
    // Create state with options
    opts := lua.Options{
        CallStackSize:       128,
        RegistrySize:        1024 * 20,
        RegistryMaxSize:     1024 * 80,
        RegistryGrowStep:    32,
        SkipOpenLibs:        false,
        IncludeGoStackTrace: true,
    }
    
    L := lua.NewState(opts)
    defer L.Close()
    
    // Use the state
    err := L.DoString(`
        print("Hello from Lua!")
        local t = {}
        for i = 1, 1000 do
            t[i] = i * 2
        end
        return #t
    `)
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Get result
    result := L.Get(-1)
    fmt.Printf("Result: %v\n", result)
    
    // Clean up before closing
    L.SetTop(0)
    L.DoString(`collectgarbage("collect")`)
}
```

### State Factory Pattern

```go
package factory

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
)

type LStateFactory struct {
    options      lua.Options
    initScript   string
    libraries    []string
    modules      map[string]lua.LGFunction
}

func NewLStateFactory() *LStateFactory {
    return &LStateFactory{
        options: lua.Options{
            CallStackSize:   128,
            RegistrySize:    1024 * 20,
            RegistryMaxSize: 1024 * 80,
        },
        libraries: []string{"base", "string", "table", "math"},
        modules:   make(map[string]lua.LGFunction),
    }
}

func (f *LStateFactory) WithInitScript(script string) *LStateFactory {
    f.initScript = script
    return f
}

func (f *LStateFactory) WithModule(name string, loader lua.LGFunction) *LStateFactory {
    f.modules[name] = loader
    return f
}

func (f *LStateFactory) Create() (*lua.LState, error) {
    L := lua.NewState(f.options)
    
    // Load selected libraries
    for _, lib := range f.libraries {
        switch lib {
        case "base":
            L.OpenBase()
        case "string":
            L.OpenString()
        case "table":
            L.OpenTable()
        case "math":
            L.OpenMath()
        case "io":
            L.OpenIo()
        case "os":
            L.OpenOs()
        }
    }
    
    // Preload modules
    for name, loader := range f.modules {
        L.PreloadModule(name, loader)
    }
    
    // Run initialization script
    if f.initScript != "" {
        if err := L.DoString(f.initScript); err != nil {
            L.Close()
            return nil, fmt.Errorf("init script failed: %w", err)
        }
    }
    
    return L, nil
}

// Example usage
func ExampleFactory() {
    factory := NewLStateFactory().
        WithInitScript(`
            -- Global utilities
            function map(t, f)
                local result = {}
                for k, v in pairs(t) do
                    result[k] = f(v)
                end
                return result
            end
            
            function filter(t, f)
                local result = {}
                for k, v in pairs(t) do
                    if f(v) then
                        result[k] = v
                    end
                end
                return result
            end
        `).
        WithModule("utils", func(L *lua.LState) int {
            mod := L.NewTable()
            L.SetField(mod, "version", lua.LString("1.0"))
            L.Push(mod)
            return 1
        })
    
    L, err := factory.Create()
    if err != nil {
        log.Fatal(err)
    }
    defer L.Close()
    
    // Use the configured state
    L.DoString(`
        local utils = require("utils")
        print("Utils version:", utils.version)
        
        local numbers = {1, 2, 3, 4, 5}
        local doubled = map(numbers, function(x) return x * 2 end)
    `)
}
```

## State Pooling

### Basic State Pool

```go
package pooling

import (
    "fmt"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type BasicStatePool struct {
    factory   *LStateFactory
    pool      chan *lua.LState
    maxSize   int
    created   int
    mu        sync.Mutex
}

func NewBasicStatePool(factory *LStateFactory, size int) *BasicStatePool {
    return &BasicStatePool{
        factory: factory,
        pool:    make(chan *lua.LState, size),
        maxSize: size,
    }
}

func (p *BasicStatePool) Get() (*lua.LState, error) {
    select {
    case L := <-p.pool:
        return L, nil
    default:
        // Pool is empty, try to create new state
        p.mu.Lock()
        if p.created >= p.maxSize {
            p.mu.Unlock()
            // Wait for available state
            L := <-p.pool
            return L, nil
        }
        p.created++
        p.mu.Unlock()
        
        return p.factory.Create()
    }
}

func (p *BasicStatePool) Put(L *lua.LState) {
    // Clean the state before returning to pool
    L.SetTop(0)
    L.DoString(`
        -- Clear globals except standard libraries
        for k, v in pairs(_G) do
            if type(k) == "string" and not string.match(k, "^[A-Z_]") then
                _G[k] = nil
            end
        end
        collectgarbage("collect")
    `)
    
    select {
    case p.pool <- L:
        // Successfully returned to pool
    default:
        // Pool is full, close the state
        L.Close()
        p.mu.Lock()
        p.created--
        p.mu.Unlock()
    }
}

func (p *BasicStatePool) Close() {
    close(p.pool)
    for L := range p.pool {
        L.Close()
    }
}

// Example with concurrent usage
func ExampleBasicPool() {
    factory := NewLStateFactory()
    pool := NewBasicStatePool(factory, 5)
    defer pool.Close()
    
    var wg sync.WaitGroup
    
    // Simulate concurrent script execution
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            L, err := pool.Get()
            if err != nil {
                log.Printf("Worker %d: Failed to get state: %v", id, err)
                return
            }
            defer pool.Put(L)
            
            // Execute script
            err = L.DoString(fmt.Sprintf(`
                print("Worker %d executing")
                local sum = 0
                for i = 1, 1000 do
                    sum = sum + i
                end
                return sum
            `, id))
            
            if err != nil {
                log.Printf("Worker %d: Script error: %v", id, err)
            }
        }(i)
    }
    
    wg.Wait()
}
```

### Advanced Pool with Health Monitoring

```go
package advanced

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type StateHealth struct {
    UseCount    int64
    ErrorCount  int64
    LastUsed    time.Time
    CreatedAt   time.Time
    MemoryUsage int
}

type PooledState struct {
    L      *lua.LState
    health *StateHealth
    id     int64
}

type AdvancedPool struct {
    factory       *LStateFactory
    available     chan *PooledState
    all           map[int64]*PooledState
    maxSize       int
    maxAge        time.Duration
    maxUses       int64
    nextID        int64
    mu            sync.RWMutex
    
    // Metrics
    totalCreated  int64
    totalDestroyed int64
    currentSize   int32
}

func NewAdvancedPool(config PoolConfig) *AdvancedPool {
    pool := &AdvancedPool{
        factory:     config.Factory,
        available:   make(chan *PooledState, config.MaxSize),
        all:         make(map[int64]*PooledState),
        maxSize:     config.MaxSize,
        maxAge:      config.MaxAge,
        maxUses:     config.MaxUses,
    }
    
    // Pre-create minimum states
    for i := 0; i < config.MinSize; i++ {
        if ps, err := pool.createState(); err == nil {
            pool.available <- ps
        }
    }
    
    // Start maintenance routine
    go pool.maintenanceLoop()
    
    return pool
}

func (p *AdvancedPool) createState() (*PooledState, error) {
    L, err := p.factory.Create()
    if err != nil {
        return nil, err
    }
    
    id := atomic.AddInt64(&p.nextID, 1)
    ps := &PooledState{
        L:  L,
        id: id,
        health: &StateHealth{
            CreatedAt: time.Now(),
            LastUsed:  time.Now(),
        },
    }
    
    p.mu.Lock()
    p.all[id] = ps
    p.mu.Unlock()
    
    atomic.AddInt64(&p.totalCreated, 1)
    atomic.AddInt32(&p.currentSize, 1)
    
    return ps, nil
}

func (p *AdvancedPool) Get() (*lua.LState, int64, error) {
    for {
        select {
        case ps := <-p.available:
            if p.isHealthy(ps) {
                ps.health.LastUsed = time.Now()
                atomic.AddInt64(&ps.health.UseCount, 1)
                return ps.L, ps.id, nil
            }
            
            // Unhealthy state, destroy it
            p.destroyState(ps)
            
        case <-time.After(100 * time.Millisecond):
            // Try to create new state
            if atomic.LoadInt32(&p.currentSize) < int32(p.maxSize) {
                ps, err := p.createState()
                if err != nil {
                    return nil, 0, err
                }
                ps.health.LastUsed = time.Now()
                atomic.AddInt64(&ps.health.UseCount, 1)
                return ps.L, ps.id, nil
            }
        }
    }
}

func (p *AdvancedPool) Put(L *lua.LState, id int64, hadError bool) {
    p.mu.RLock()
    ps, exists := p.all[id]
    p.mu.RUnlock()
    
    if !exists {
        return
    }
    
    if hadError {
        atomic.AddInt64(&ps.health.ErrorCount, 1)
    }
    
    // Update memory usage
    ps.health.MemoryUsage = L.MemUsage()
    
    // Clean state
    p.cleanState(L)
    
    // Return to pool if healthy
    if p.isHealthy(ps) {
        select {
        case p.available <- ps:
        default:
            // Pool full, destroy state
            p.destroyState(ps)
        }
    } else {
        p.destroyState(ps)
    }
}

func (p *AdvancedPool) isHealthy(ps *PooledState) bool {
    health := ps.health
    
    // Check age
    if p.maxAge > 0 && time.Since(health.CreatedAt) > p.maxAge {
        return false
    }
    
    // Check use count
    if p.maxUses > 0 && health.UseCount >= p.maxUses {
        return false
    }
    
    // Check error rate
    if health.UseCount > 0 {
        errorRate := float64(health.ErrorCount) / float64(health.UseCount)
        if errorRate > 0.1 { // 10% error rate threshold
            return false
        }
    }
    
    // Check memory
    if health.MemoryUsage > 50*1024*1024 { // 50MB threshold
        return false
    }
    
    return true
}

func (p *AdvancedPool) cleanState(L *lua.LState) {
    L.SetTop(0)
    L.DoString(`
        -- Clear non-standard globals
        for k, v in pairs(_G) do
            if type(k) == "string" and k:sub(1,1) ~= "_" then
                local isStandard = false
                for _, std in ipairs({"string", "table", "math", "io", "os", 
                                     "coroutine", "package", "debug", "bit32"}) do
                    if k == std then
                        isStandard = true
                        break
                    end
                end
                if not isStandard then
                    _G[k] = nil
                end
            end
        end
        
        -- Clear package.loaded (except standard modules)
        for k, v in pairs(package.loaded) do
            if type(k) == "string" and not k:match("^_") then
                package.loaded[k] = nil
            end
        end
        
        collectgarbage("collect")
        collectgarbage("collect")
    `)
}

func (p *AdvancedPool) destroyState(ps *PooledState) {
    p.mu.Lock()
    delete(p.all, ps.id)
    p.mu.Unlock()
    
    ps.L.Close()
    
    atomic.AddInt64(&p.totalDestroyed, 1)
    atomic.AddInt32(&p.currentSize, -1)
}

func (p *AdvancedPool) maintenanceLoop() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        p.performMaintenance()
    }
}

func (p *AdvancedPool) performMaintenance() {
    p.mu.RLock()
    states := make([]*PooledState, 0, len(p.all))
    for _, ps := range p.all {
        states = append(states, ps)
    }
    p.mu.RUnlock()
    
    for _, ps := range states {
        // Check if state is idle and unhealthy
        if time.Since(ps.health.LastUsed) > 1*time.Minute && !p.isHealthy(ps) {
            p.destroyState(ps)
        }
    }
}

// Example usage
func ExampleAdvancedPool() {
    config := PoolConfig{
        Factory: NewLStateFactory(),
        MinSize: 2,
        MaxSize: 10,
        MaxAge:  5 * time.Minute,
        MaxUses: 1000,
    }
    
    pool := NewAdvancedPool(config)
    
    // Execute script with error handling
    L, id, err := pool.Get()
    if err != nil {
        log.Fatal(err)
    }
    
    hadError := false
    err = L.DoString(`
        function processData(data)
            local result = {}
            for i, v in ipairs(data) do
                result[i] = v * 2
            end
            return result
        end
    `)
    
    if err != nil {
        hadError = true
        log.Printf("Script error: %v", err)
    }
    
    pool.Put(L, id, hadError)
    
    // Check pool metrics
    fmt.Printf("Pool Stats:\n")
    fmt.Printf("  Current Size: %d\n", atomic.LoadInt32(&pool.currentSize))
    fmt.Printf("  Total Created: %d\n", atomic.LoadInt64(&pool.totalCreated))
    fmt.Printf("  Total Destroyed: %d\n", atomic.LoadInt64(&pool.totalDestroyed))
}
```

## State Isolation and Security

### Sandboxed State Creation

```go
package sandbox

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
)

type SandboxLevel int

const (
    SandboxMinimal SandboxLevel = iota
    SandboxStandard
    SandboxStrict
)

func CreateSandboxedState(level SandboxLevel) *lua.LState {
    L := lua.NewState()
    
    switch level {
    case SandboxMinimal:
        // Load all standard libraries
        L.OpenLibs()
        // Remove only the most dangerous functions
        L.DoString(`
            os.execute = nil
            os.exit = nil
            io.popen = nil
        `)
        
    case SandboxStandard:
        // Load safe libraries only
        L.OpenBase()
        L.OpenTable()
        L.OpenString()
        L.OpenMath()
        L.OpenCoroutine()
        
        // Remove file operations
        L.DoString(`
            loadfile = nil
            dofile = nil
            load = function(code)
                if type(code) ~= "string" then
                    error("only string code is allowed")
                end
                return loadstring(code)
            end
        `)
        
    case SandboxStrict:
        // Minimal libraries
        L.OpenBase()
        L.OpenTable()
        L.OpenString()
        L.OpenMath()
        
        // Remove all potentially dangerous functions
        L.DoString(`
            -- Remove all file/system access
            io = nil
            os = nil
            loadfile = nil
            dofile = nil
            require = nil
            package = nil
            debug = nil
            
            -- Restrict load function
            local original_load = load
            load = nil
            loadstring = nil
            
            -- Remove other dangerous functions
            rawget = nil
            rawset = nil
            setfenv = nil
            getfenv = nil
            setmetatable = nil
            getmetatable = nil
            
            -- Provide safe alternatives
            _G.safe_load = function(code)
                if type(code) ~= "string" then
                    error("only string code is allowed")
                end
                if #code > 10000 then
                    error("code too large")
                end
                return original_load(code, "sandbox", "t", {})
            end
        `)
    }
    
    return L
}

// Example with different sandbox levels
func ExampleSandbox() {
    levels := []struct {
        name  string
        level SandboxLevel
        code  string
    }{
        {
            name:  "Minimal Sandbox",
            level: SandboxMinimal,
            code: `
                -- This works in minimal sandbox
                local f = io.open("test.txt", "r")
                if f then
                    f:close()
                    print("File access allowed")
                end
            `,
        },
        {
            name:  "Standard Sandbox",
            level: SandboxStandard,
            code: `
                -- This fails in standard sandbox
                -- io.open would error
                print("Math still works:", math.sin(1))
            `,
        },
        {
            name:  "Strict Sandbox",
            level: SandboxStrict,
            code: `
                -- Very limited environment
                local t = {1, 2, 3}
                local sum = 0
                for _, v in ipairs(t) do
                    sum = sum + v
                end
                print("Sum:", sum)
            `,
        },
    }
    
    for _, test := range levels {
        fmt.Printf("\n=== %s ===\n", test.name)
        L := CreateSandboxedState(test.level)
        defer L.Close()
        
        err := L.DoString(test.code)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    }
}
```

## Lifecycle Monitoring

### State Lifecycle Tracker

```go
package lifecycle

import (
    "fmt"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type StateEvent struct {
    Type      string
    StateID   string
    Timestamp time.Time
    Details   map[string]interface{}
}

type LifecycleTracker struct {
    events chan StateEvent
    states map[string]*TrackedState
    mu     sync.RWMutex
}

type TrackedState struct {
    L           *lua.LState
    ID          string
    CreatedAt   time.Time
    Events      []StateEvent
    Metrics     StateMetrics
}

type StateMetrics struct {
    ExecutionCount int64
    TotalRuntime   time.Duration
    LastError      error
    MemoryPeak     int
}

func NewLifecycleTracker() *LifecycleTracker {
    lt := &LifecycleTracker{
        events: make(chan StateEvent, 1000),
        states: make(map[string]*TrackedState),
    }
    
    go lt.processEvents()
    
    return lt
}

func (lt *LifecycleTracker) CreateTrackedState(id string) (*TrackedState, error) {
    L := lua.NewState()
    
    ts := &TrackedState{
        L:         L,
        ID:        id,
        CreatedAt: time.Now(),
        Events:    make([]StateEvent, 0),
    }
    
    lt.mu.Lock()
    lt.states[id] = ts
    lt.mu.Unlock()
    
    lt.emitEvent(StateEvent{
        Type:      "created",
        StateID:   id,
        Timestamp: time.Now(),
    })
    
    // Install hooks for lifecycle tracking
    lt.installHooks(ts)
    
    return ts, nil
}

func (lt *LifecycleTracker) installHooks(ts *TrackedState) {
    L := ts.L
    
    // Track function calls
    L.SetGlobal("__track_call", L.NewFunction(func(L *lua.LState) int {
        funcName := L.CheckString(1)
        lt.emitEvent(StateEvent{
            Type:      "function_call",
            StateID:   ts.ID,
            Timestamp: time.Now(),
            Details: map[string]interface{}{
                "function": funcName,
            },
        })
        return 0
    }))
    
    // Override error function
    L.DoString(`
        local original_error = error
        error = function(message, level)
            __track_call("error")
            original_error(message, level or 1)
        end
    `)
}

func (lt *LifecycleTracker) ExecuteTracked(ts *TrackedState, code string) error {
    start := time.Now()
    
    lt.emitEvent(StateEvent{
        Type:      "execution_start",
        StateID:   ts.ID,
        Timestamp: start,
    })
    
    err := ts.L.DoString(code)
    
    duration := time.Since(start)
    ts.Metrics.ExecutionCount++
    ts.Metrics.TotalRuntime += duration
    
    if err != nil {
        ts.Metrics.LastError = err
    }
    
    // Check memory
    memUsage := ts.L.MemUsage()
    if memUsage > ts.Metrics.MemoryPeak {
        ts.Metrics.MemoryPeak = memUsage
    }
    
    lt.emitEvent(StateEvent{
        Type:      "execution_end",
        StateID:   ts.ID,
        Timestamp: time.Now(),
        Details: map[string]interface{}{
            "duration": duration,
            "error":    err != nil,
            "memory":   memUsage,
        },
    })
    
    return err
}

func (lt *LifecycleTracker) DestroyTracked(ts *TrackedState) {
    lt.emitEvent(StateEvent{
        Type:      "destroyed",
        StateID:   ts.ID,
        Timestamp: time.Now(),
        Details: map[string]interface{}{
            "lifetime":       time.Since(ts.CreatedAt),
            "execution_count": ts.Metrics.ExecutionCount,
            "total_runtime":  ts.Metrics.TotalRuntime,
        },
    })
    
    ts.L.Close()
    
    lt.mu.Lock()
    delete(lt.states, ts.ID)
    lt.mu.Unlock()
}

func (lt *LifecycleTracker) emitEvent(event StateEvent) {
    select {
    case lt.events <- event:
    default:
        // Event queue full, drop event
    }
}

func (lt *LifecycleTracker) processEvents() {
    for event := range lt.events {
        lt.mu.RLock()
        if ts, exists := lt.states[event.StateID]; exists {
            ts.Events = append(ts.Events, event)
        }
        lt.mu.RUnlock()
        
        // Log or process event
        fmt.Printf("[%s] %s: %s\n", 
            event.Timestamp.Format("15:04:05.000"),
            event.StateID,
            event.Type)
    }
}

// Example usage
func ExampleLifecycleTracking() {
    tracker := NewLifecycleTracker()
    
    // Create tracked state
    ts, err := tracker.CreateTrackedState("worker-1")
    if err != nil {
        log.Fatal(err)
    }
    defer tracker.DestroyTracked(ts)
    
    // Execute some code
    for i := 0; i < 3; i++ {
        err := tracker.ExecuteTracked(ts, fmt.Sprintf(`
            print("Execution %d")
            local sum = 0
            for j = 1, 1000 do
                sum = sum + j
            end
            if sum %% 2 == 0 then
                error("even sum!")
            end
        `, i+1))
        
        if err != nil {
            fmt.Printf("Execution %d failed: %v\n", i+1, err)
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    
    // Print metrics
    fmt.Printf("\nState Metrics:\n")
    fmt.Printf("  Execution Count: %d\n", ts.Metrics.ExecutionCount)
    fmt.Printf("  Total Runtime: %v\n", ts.Metrics.TotalRuntime)
    fmt.Printf("  Memory Peak: %d bytes\n", ts.Metrics.MemoryPeak)
    fmt.Printf("  Event Count: %d\n", len(ts.Events))
}
```

## Complete Lifecycle Manager

### Integrated State Manager

```go
package manager

import (
    "context"
    "fmt"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type StateManager struct {
    factory     *LStateFactory
    pool        *AdvancedPool
    tracker     *LifecycleTracker
    config      ManagerConfig
    
    // Shutdown coordination
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

type ManagerConfig struct {
    PoolSize        int
    MaxStateAge     time.Duration
    MaxStateUses    int64
    CleanupInterval time.Duration
    EnableTracking  bool
    SandboxLevel    SandboxLevel
}

func NewStateManager(config ManagerConfig) *StateManager {
    ctx, cancel := context.WithCancel(context.Background())
    
    sm := &StateManager{
        factory: NewLStateFactory(),
        config:  config,
        ctx:     ctx,
        cancel:  cancel,
    }
    
    // Configure factory based on sandbox level
    sm.configureSandbox()
    
    // Create pool
    poolConfig := PoolConfig{
        Factory: sm.factory,
        MinSize: 2,
        MaxSize: config.PoolSize,
        MaxAge:  config.MaxStateAge,
        MaxUses: config.MaxStateUses,
    }
    sm.pool = NewAdvancedPool(poolConfig)
    
    // Create tracker if enabled
    if config.EnableTracking {
        sm.tracker = NewLifecycleTracker()
    }
    
    // Start background workers
    sm.wg.Add(1)
    go sm.cleanupWorker()
    
    return sm
}

func (sm *StateManager) configureSandbox() {
    switch sm.config.SandboxLevel {
    case SandboxStrict:
        sm.factory.libraries = []string{"base", "table", "string", "math"}
        sm.factory.initScript = `
            -- Remove dangerous functions
            io = nil
            os = nil
            debug = nil
            require = nil
            loadfile = nil
            dofile = nil
        `
    case SandboxStandard:
        sm.factory.libraries = []string{"base", "table", "string", "math", "coroutine"}
    default:
        sm.factory.libraries = []string{"base", "table", "string", "math", "coroutine", "os", "io"}
    }
}

func (sm *StateManager) Execute(ctx context.Context, script string) (interface{}, error) {
    // Get state from pool
    L, id, err := sm.pool.Get()
    if err != nil {
        return nil, fmt.Errorf("failed to get state: %w", err)
    }
    
    hadError := false
    defer func() {
        sm.pool.Put(L, id, hadError)
    }()
    
    // Set execution context
    L.SetContext(ctx)
    
    // Track execution if enabled
    if sm.tracker != nil {
        // Wrap execution with tracking
        // Implementation depends on tracker integration
    }
    
    // Execute script
    err = L.DoString(script)
    if err != nil {
        hadError = true
        return nil, err
    }
    
    // Get return values
    results := make([]interface{}, L.GetTop())
    for i := 0; i < L.GetTop(); i++ {
        results[i] = luaValueToGo(L.Get(i + 1))
    }
    
    return results, nil
}

func (sm *StateManager) cleanupWorker() {
    defer sm.wg.Done()
    
    ticker := time.NewTicker(sm.config.CleanupInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            sm.performCleanup()
        case <-sm.ctx.Done():
            return
        }
    }
}

func (sm *StateManager) performCleanup() {
    // Pool maintenance is handled by the pool itself
    // This is for additional cleanup tasks
    
    // Log current stats
    fmt.Printf("State Manager Stats:\n")
    fmt.Printf("  Pool Size: %d\n", atomic.LoadInt32(&sm.pool.currentSize))
    fmt.Printf("  States Created: %d\n", atomic.LoadInt64(&sm.pool.totalCreated))
    fmt.Printf("  States Destroyed: %d\n", atomic.LoadInt64(&sm.pool.totalDestroyed))
}

func (sm *StateManager) Shutdown() error {
    sm.cancel()
    sm.wg.Wait()
    
    // Close pool
    sm.pool.Close()
    
    return nil
}

// Utility function to convert Lua values to Go
func luaValueToGo(lv lua.LValue) interface{} {
    switch v := lv.(type) {
    case lua.LBool:
        return bool(v)
    case lua.LNumber:
        return float64(v)
    case lua.LString:
        return string(v)
    case *lua.LTable:
        // Simple conversion, doesn't handle circular references
        result := make(map[string]interface{})
        v.ForEach(func(k, v lua.LValue) {
            if key, ok := k.(lua.LString); ok {
                result[string(key)] = luaValueToGo(v)
            }
        })
        return result
    case *lua.LNilType:
        return nil
    default:
        return fmt.Sprintf("%v", v)
    }
}

// Example usage
func ExampleStateManager() {
    config := ManagerConfig{
        PoolSize:        5,
        MaxStateAge:     10 * time.Minute,
        MaxStateUses:    100,
        CleanupInterval: 1 * time.Minute,
        EnableTracking:  true,
        SandboxLevel:    SandboxStandard,
    }
    
    manager := NewStateManager(config)
    defer manager.Shutdown()
    
    // Execute scripts concurrently
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            script := fmt.Sprintf(`
                local worker_id = %d
                local result = {}
                
                for i = 1, 10 do
                    table.insert(result, worker_id * 10 + i)
                end
                
                return result
            `, id)
            
            result, err := manager.Execute(ctx, script)
            if err != nil {
                fmt.Printf("Worker %d error: %v\n", id, err)
                return
            }
            
            fmt.Printf("Worker %d result: %v\n", id, result)
        }(i)
    }
    
    wg.Wait()
}
```

## Summary

These examples demonstrate comprehensive LState lifecycle management:
1. Basic state creation and cleanup patterns
2. Factory pattern for consistent state configuration
3. Simple and advanced pooling strategies
4. Health monitoring and automatic recycling
5. Sandbox creation for security isolation
6. Lifecycle tracking and metrics collection
7. Complete integrated state management system

Key features include automatic scaling, health-based recycling, comprehensive monitoring, and secure execution environments.