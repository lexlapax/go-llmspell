# Lua Memory Limits via Registry Configuration Research

This document investigates memory limiting mechanisms in GopherLua, focusing on registry configuration, memory tracking, and practical implementation strategies.

## Executive Summary

GopherLua provides memory usage tracking through `LState.MemUsage()` but does not enforce hard memory limits natively. Memory control must be implemented through a combination of registry size configuration, hook-based monitoring, and careful management of Lua object lifecycles.

## Memory Management in GopherLua

### Memory Tracking Capabilities

GopherLua tracks memory usage but doesn't provide native enforcement:

```go
// Get current memory usage
memUsage := L.MemUsage() // Returns approximate bytes used

// No built-in memory limit enforcement
// Must implement custom solution
```

### Registry Configuration

The Lua registry is the global storage for all Lua values:

```go
type LState struct {
    // Registry configuration options
    Options Options
    
    // Registry size affects memory usage
    reg *registry
}

type Options struct {
    // Registry size configuration
    RegistrySize         int // Initial registry array size
    RegistryMaxSize      int // Maximum registry growth
    RegistryGrowStep     int // Registry growth increment
    
    // Other memory-related options
    CallStackSize        int // Call stack size
    MinimizeStackMemory  bool // Minimize stack allocations
}
```

## Memory Limiting Strategies

### 1. Registry Size Limits

```go
// Configure registry with size limits
opts := lua.Options{
    RegistrySize:     1024 * 4,    // Initial: 4K entries
    RegistryMaxSize:  1024 * 1024, // Max: 1M entries
    RegistryGrowStep: 32,          // Grow by 32 entries
    CallStackSize:    128,         // Limit call stack depth
}

L := lua.NewState(opts)
```

### 2. Hook-Based Memory Monitoring

```go
type MemoryMonitor struct {
    limit        int
    checkCounter int
    checkInterval int
}

func (mm *MemoryMonitor) Hook(L *lua.LState) {
    mm.checkCounter++
    if mm.checkCounter < mm.checkInterval {
        return
    }
    mm.checkCounter = 0
    
    usage := L.MemUsage()
    if usage > mm.limit {
        L.RaiseError("memory limit exceeded: %d bytes (limit: %d)", usage, mm.limit)
    }
}

// Usage
monitor := &MemoryMonitor{
    limit:         10 * 1024 * 1024, // 10MB
    checkInterval: 1000,              // Check every 1000 instructions
}
L.SetHook(monitor.Hook, lua.Count, 100)
```

### 3. Object Allocation Tracking

```go
type AllocationTracker struct {
    tableCount    int
    tableLimit    int
    stringBytes   int
    stringLimit   int
    functionCount int
    functionLimit int
}

// Custom allocator wrapper
func (at *AllocationTracker) TrackAllocation(L *lua.LState, typ lua.LValueType) error {
    switch typ {
    case lua.LTTable:
        at.tableCount++
        if at.tableCount > at.tableLimit {
            return fmt.Errorf("table allocation limit exceeded: %d", at.tableLimit)
        }
    case lua.LTString:
        // Track string memory (approximate)
        at.stringBytes += 64 // Average string overhead
        if at.stringBytes > at.stringLimit {
            return fmt.Errorf("string memory limit exceeded: %d bytes", at.stringLimit)
        }
    case lua.LTFunction:
        at.functionCount++
        if at.functionCount > at.functionLimit {
            return fmt.Errorf("function allocation limit exceeded: %d", at.functionLimit)
        }
    }
    return nil
}
```

## Memory Usage Patterns

### Memory Consumption by Type

| Lua Type | Approximate Memory Usage | Notes |
|----------|-------------------------|-------|
| nil | 0 bytes | Singleton value |
| boolean | 0 bytes | Singleton values (true/false) |
| number | 8 bytes | 64-bit float |
| string | 24 + len bytes | Header + content |
| table | 64 + entries * 32 bytes | Base + entry overhead |
| function | 100+ bytes | Varies with closure size |
| userdata | 32 + data bytes | Header + user data |
| thread | 1KB+ | Separate stack |

### Registry Growth Patterns

```go
// Registry growth visualization
type RegistryMonitor struct {
    measurements []MemoryMeasurement
    mu           sync.Mutex
}

type MemoryMeasurement struct {
    Time         time.Time
    MemUsage     int
    ObjectCount  int
    RegistrySize int
}

func (rm *RegistryMonitor) Measure(L *lua.LState) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    measurement := MemoryMeasurement{
        Time:     time.Now(),
        MemUsage: L.MemUsage(),
        // Approximate object count
        ObjectCount: L.GetTop(),
    }
    
    rm.measurements = append(rm.measurements, measurement)
}
```

## Advanced Memory Control

### 1. Garbage Collection Control

```go
type GCController struct {
    memoryThreshold int
    lastGC          time.Time
    minGCInterval   time.Duration
}

func (gc *GCController) MaybeGC(L *lua.LState) {
    if L.MemUsage() > gc.memoryThreshold && 
       time.Since(gc.lastGC) > gc.minGCInterval {
        
        before := L.MemUsage()
        
        // Full garbage collection
        L.DoString(`collectgarbage("collect")`)
        
        after := L.MemUsage()
        gc.lastGC = time.Now()
        
        log.Printf("GC: freed %d bytes (%d -> %d)", 
            before-after, before, after)
    }
}
```

### 2. Memory Pool Implementation

```go
type LuaMemoryPool struct {
    states      []*lua.LState
    maxStates   int
    stateMemory int // Memory per state
    mu          sync.Mutex
}

func NewLuaMemoryPool(maxStates, memoryPerState int) *LuaMemoryPool {
    return &LuaMemoryPool{
        maxStates:   maxStates,
        stateMemory: memoryPerState,
        states:      make([]*lua.LState, 0, maxStates),
    }
}

func (p *LuaMemoryPool) Get() (*lua.LState, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if len(p.states) > 0 {
        L := p.states[len(p.states)-1]
        p.states = p.states[:len(p.states)-1]
        return L, nil
    }
    
    // Check if we can create new state
    totalMemory := len(p.states) * p.stateMemory
    if totalMemory+p.stateMemory > p.maxStates*p.stateMemory {
        return nil, fmt.Errorf("memory pool exhausted")
    }
    
    // Create new state with memory limits
    opts := lua.Options{
        RegistrySize:    1024,
        RegistryMaxSize: p.stateMemory / 32, // Approximate
    }
    
    return lua.NewState(opts), nil
}
```

### 3. Memory Quota System

```go
type MemoryQuota struct {
    total     int64
    used      int64
    reserved  map[string]int64 // Per-script reservations
    mu        sync.RWMutex
}

func (mq *MemoryQuota) Reserve(scriptID string, bytes int64) error {
    mq.mu.Lock()
    defer mq.mu.Unlock()
    
    if mq.used+bytes > mq.total {
        return fmt.Errorf("insufficient memory quota: need %d, available %d",
            bytes, mq.total-mq.used)
    }
    
    mq.used += bytes
    mq.reserved[scriptID] = bytes
    return nil
}

func (mq *MemoryQuota) Release(scriptID string) {
    mq.mu.Lock()
    defer mq.mu.Unlock()
    
    if bytes, ok := mq.reserved[scriptID]; ok {
        mq.used -= bytes
        delete(mq.reserved, scriptID)
    }
}
```

## Implementation Patterns

### 1. Comprehensive Memory Limiter

```go
type MemoryLimiter struct {
    hardLimit      int
    softLimit      int
    checkInterval  int
    gcThreshold    int
    
    // Metrics
    peakUsage      int
    gcCount        int
    violationCount int
    
    // Callbacks
    onSoftLimit    func(usage int)
    onApproaching  func(usage int, limit int)
}

func (ml *MemoryLimiter) EnforceLimit(L *lua.LState) error {
    usage := L.MemUsage()
    
    // Update peak
    if usage > ml.peakUsage {
        ml.peakUsage = usage
    }
    
    // Soft limit warning
    if usage > ml.softLimit && ml.onSoftLimit != nil {
        ml.onSoftLimit(usage)
    }
    
    // Approaching limit warning (90%)
    if usage > int(float64(ml.hardLimit)*0.9) && ml.onApproaching != nil {
        ml.onApproaching(usage, ml.hardLimit)
    }
    
    // Try GC before failing
    if usage > ml.gcThreshold {
        before := usage
        L.DoString(`collectgarbage("collect")`)
        after := L.MemUsage()
        ml.gcCount++
        
        // Recalculate
        usage = after
        
        log.Printf("Auto-GC: freed %d bytes", before-after)
    }
    
    // Hard limit enforcement
    if usage > ml.hardLimit {
        ml.violationCount++
        return fmt.Errorf("memory limit exceeded: %d/%d bytes", 
            usage, ml.hardLimit)
    }
    
    return nil
}
```

### 2. Script-Specific Memory Tracking

```go
type ScriptMemoryTracker struct {
    scripts map[string]*ScriptMemoryInfo
    mu      sync.RWMutex
}

type ScriptMemoryInfo struct {
    ID            string
    StartMemory   int
    CurrentMemory int
    PeakMemory    int
    Allocations   int
    GCCount       int
    LastUpdated   time.Time
}

func (smt *ScriptMemoryTracker) StartTracking(scriptID string, L *lua.LState) {
    smt.mu.Lock()
    defer smt.mu.Unlock()
    
    info := &ScriptMemoryInfo{
        ID:            scriptID,
        StartMemory:   L.MemUsage(),
        CurrentMemory: L.MemUsage(),
        PeakMemory:    L.MemUsage(),
        LastUpdated:   time.Now(),
    }
    
    smt.scripts[scriptID] = info
}

func (smt *ScriptMemoryTracker) Update(scriptID string, L *lua.LState) {
    smt.mu.Lock()
    defer smt.mu.Unlock()
    
    if info, ok := smt.scripts[scriptID]; ok {
        current := L.MemUsage()
        info.CurrentMemory = current
        if current > info.PeakMemory {
            info.PeakMemory = current
        }
        info.Allocations++
        info.LastUpdated = time.Now()
    }
}
```

### 3. Memory Profiling

```go
type MemoryProfiler struct {
    samples   []MemorySample
    interval  time.Duration
    maxSamples int
}

type MemorySample struct {
    Timestamp   time.Time
    MemoryUsage int
    GCInfo      GCInfo
    StackDepth  int
}

type GCInfo struct {
    Count      int
    TotalPause time.Duration
}

func (mp *MemoryProfiler) Profile(L *lua.LState) {
    sample := MemorySample{
        Timestamp:   time.Now(),
        MemoryUsage: L.MemUsage(),
        StackDepth:  L.GetTop(),
    }
    
    mp.samples = append(mp.samples, sample)
    
    // Maintain sample limit
    if len(mp.samples) > mp.maxSamples {
        mp.samples = mp.samples[1:]
    }
}

func (mp *MemoryProfiler) GenerateReport() MemoryReport {
    if len(mp.samples) == 0 {
        return MemoryReport{}
    }
    
    var total, peak int
    for _, s := range mp.samples {
        total += s.MemoryUsage
        if s.MemoryUsage > peak {
            peak = s.MemoryUsage
        }
    }
    
    return MemoryReport{
        AverageUsage: total / len(mp.samples),
        PeakUsage:    peak,
        Duration:     mp.samples[len(mp.samples)-1].Timestamp.Sub(mp.samples[0].Timestamp),
        SampleCount:  len(mp.samples),
    }
}
```

## Testing Memory Limits

### Memory Stress Tests

```go
func TestMemoryLimits(t *testing.T) {
    tests := []struct {
        name        string
        script      string
        memoryLimit int
        shouldFail  bool
    }{
        {
            name: "simple allocation",
            script: `
                local t = {}
                for i = 1, 1000 do
                    t[i] = "test" .. i
                end
            `,
            memoryLimit: 1024 * 1024, // 1MB
            shouldFail:  false,
        },
        {
            name: "excessive allocation",
            script: `
                local tables = {}
                for i = 1, 100000 do
                    tables[i] = {data = string.rep("x", 1000)}
                end
            `,
            memoryLimit: 10 * 1024 * 1024, // 10MB
            shouldFail:  true,
        },
        {
            name: "circular references",
            script: `
                for i = 1, 10000 do
                    local a, b = {}, {}
                    a.ref = b
                    b.ref = a
                end
            `,
            memoryLimit: 5 * 1024 * 1024, // 5MB
            shouldFail:  false, // Should handle with GC
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            L := lua.NewState()
            defer L.Close()
            
            limiter := &MemoryLimiter{
                hardLimit:   tt.memoryLimit,
                softLimit:   int(float64(tt.memoryLimit) * 0.8),
                gcThreshold: int(float64(tt.memoryLimit) * 0.7),
            }
            
            var err error
            L.SetHook(func(L *lua.LState) {
                if err == nil {
                    err = limiter.EnforceLimit(L)
                }
                if err != nil {
                    L.RaiseError(err.Error())
                }
            }, lua.Count, 1000)
            
            scriptErr := L.DoString(tt.script)
            
            if tt.shouldFail && scriptErr == nil {
                t.Errorf("expected memory limit error, got none")
            } else if !tt.shouldFail && scriptErr != nil {
                t.Errorf("unexpected error: %v", scriptErr)
            }
        })
    }
}
```

## Best Practices

1. **Conservative Limits**: Set memory limits conservatively to account for GC overhead
2. **Regular Monitoring**: Check memory usage at reasonable intervals (not every instruction)
3. **GC Integration**: Trigger GC before failing on memory limits
4. **Soft Limits**: Implement warning thresholds before hard limits
5. **Script Isolation**: Track memory per script for multi-tenant scenarios

## Configuration Examples

### Development Configuration
```go
devConfig := MemoryConfig{
    HardLimit:     100 * 1024 * 1024, // 100MB
    SoftLimit:     80 * 1024 * 1024,  // 80MB
    CheckInterval: 10000,              // Low overhead
    EnableGC:      true,
    GCThreshold:   50 * 1024 * 1024,  // 50MB
}
```

### Production Configuration
```go
prodConfig := MemoryConfig{
    HardLimit:     10 * 1024 * 1024, // 10MB
    SoftLimit:     8 * 1024 * 1024,  // 8MB
    CheckInterval: 1000,              // More frequent
    EnableGC:      true,
    GCThreshold:   5 * 1024 * 1024,  // 5MB
}
```

### Embedded Configuration
```go
embeddedConfig := MemoryConfig{
    HardLimit:     1 * 1024 * 1024, // 1MB
    SoftLimit:     768 * 1024,      // 768KB
    CheckInterval: 100,              // Very frequent
    EnableGC:      true,
    GCThreshold:   512 * 1024,      // 512KB
}
```

## Implementation Checklist

- [ ] Memory usage monitoring via hooks
- [ ] Registry size configuration
- [ ] Garbage collection integration
- [ ] Soft and hard limit enforcement
- [ ] Per-script memory tracking
- [ ] Memory profiling tools
- [ ] Quota management system
- [ ] Testing framework for memory limits
- [ ] Documentation of memory patterns
- [ ] Performance impact analysis

## Summary

While GopherLua doesn't provide native memory limit enforcement, effective memory control can be achieved through:
1. Registry configuration to limit growth
2. Hook-based monitoring with `MemUsage()`
3. Integrated garbage collection strategies
4. Careful tracking of allocations
5. Multi-level limit enforcement (soft/hard)

The key is balancing security and performance while providing useful feedback to scripts approaching their limits.