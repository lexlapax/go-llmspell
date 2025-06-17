# Lua Memory Limits Examples

This document provides practical examples of implementing memory limits and monitoring in GopherLua.

## Basic Memory Monitoring

### Simple Memory Check

```go
package main

import (
    "fmt"
    "log"
    lua "github.com/yuin/gopher-lua"
)

func main() {
    L := lua.NewState()
    defer L.Close()
    
    // Monitor memory usage
    initialMem := L.MemUsage()
    fmt.Printf("Initial memory: %d bytes\n", initialMem)
    
    // Execute script that allocates memory
    err := L.DoString(`
        local bigTable = {}
        for i = 1, 10000 do
            bigTable[i] = {
                id = i,
                data = string.rep("x", 100),
                nested = {a = 1, b = 2, c = 3}
            }
        end
    `)
    
    if err != nil {
        log.Fatal(err)
    }
    
    afterMem := L.MemUsage()
    fmt.Printf("After execution: %d bytes\n", afterMem)
    fmt.Printf("Memory used: %d bytes\n", afterMem-initialMem)
    
    // Force garbage collection
    L.DoString(`collectgarbage("collect")`)
    
    gcMem := L.MemUsage()
    fmt.Printf("After GC: %d bytes\n", gcMem)
    fmt.Printf("Memory freed: %d bytes\n", afterMem-gcMem)
}
```

## Hook-Based Memory Limiting

### Basic Memory Limiter

```go
package memlimit

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
)

type BasicMemoryLimiter struct {
    limit         int
    checkCounter  int
    checkInterval int
}

func NewBasicMemoryLimiter(limitBytes int) *BasicMemoryLimiter {
    return &BasicMemoryLimiter{
        limit:         limitBytes,
        checkInterval: 1000, // Check every 1000 instructions
    }
}

func (ml *BasicMemoryLimiter) Hook(L *lua.LState) {
    ml.checkCounter++
    if ml.checkCounter < ml.checkInterval {
        return
    }
    ml.checkCounter = 0
    
    usage := L.MemUsage()
    if usage > ml.limit {
        L.RaiseError("memory limit exceeded: %d bytes (limit: %d)", 
            usage, ml.limit)
    }
}

// Example usage
func ExecuteWithMemoryLimit(script string, limitMB int) error {
    L := lua.NewState()
    defer L.Close()
    
    limiter := NewBasicMemoryLimiter(limitMB * 1024 * 1024)
    L.SetHook(limiter.Hook, lua.Count, 100)
    
    return L.DoString(script)
}
```

### Advanced Memory Controller

```go
package memlimit

import (
    "fmt"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type AdvancedMemoryController struct {
    mu sync.RWMutex
    
    // Limits
    hardLimit int
    softLimit int
    
    // GC settings
    gcThreshold   int
    lastGC        time.Time
    minGCInterval time.Duration
    
    // Metrics
    peakUsage      int
    totalGCs       int
    totalFreed     int64
    violations     int
    
    // Callbacks
    onSoftLimit func(usage int)
    onGCTrigger func(before, after int)
}

func NewAdvancedMemoryController(config MemoryConfig) *AdvancedMemoryController {
    return &AdvancedMemoryController{
        hardLimit:     config.HardLimit,
        softLimit:     config.SoftLimit,
        gcThreshold:   config.GCThreshold,
        minGCInterval: config.MinGCInterval,
    }
}

func (mc *AdvancedMemoryController) CheckMemory(L *lua.LState) error {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    
    usage := L.MemUsage()
    
    // Update peak usage
    if usage > mc.peakUsage {
        mc.peakUsage = usage
    }
    
    // Check soft limit
    if usage > mc.softLimit && mc.onSoftLimit != nil {
        mc.onSoftLimit(usage)
    }
    
    // Try GC if above threshold
    if usage > mc.gcThreshold && time.Since(mc.lastGC) > mc.minGCInterval {
        before := usage
        L.DoString(`collectgarbage("collect")`)
        after := L.MemUsage()
        
        mc.lastGC = time.Now()
        mc.totalGCs++
        mc.totalFreed += int64(before - after)
        
        if mc.onGCTrigger != nil {
            mc.onGCTrigger(before, after)
        }
        
        usage = after // Update usage after GC
    }
    
    // Check hard limit
    if usage > mc.hardLimit {
        mc.violations++
        return fmt.Errorf("memory limit exceeded: %d/%d bytes", 
            usage, mc.hardLimit)
    }
    
    return nil
}

func (mc *AdvancedMemoryController) GetMetrics() MemoryMetrics {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    
    return MemoryMetrics{
        PeakUsage:    mc.peakUsage,
        TotalGCs:     mc.totalGCs,
        TotalFreed:   mc.totalFreed,
        Violations:   mc.violations,
        AverageFreed: mc.totalFreed / int64(max(mc.totalGCs, 1)),
    }
}
```

## Registry Configuration

### Custom Registry Settings

```go
package main

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
)

func createLimitedLuaState(memoryClass string) *lua.LState {
    var opts lua.Options
    
    switch memoryClass {
    case "minimal":
        opts = lua.Options{
            RegistrySize:        128,    // Very small registry
            RegistryMaxSize:     1024,   // Max 1K entries
            RegistryGrowStep:    16,     // Grow slowly
            CallStackSize:       64,     // Limited call depth
            MinimizeStackMemory: true,   // Minimize allocations
        }
    
    case "standard":
        opts = lua.Options{
            RegistrySize:        1024,   // Standard size
            RegistryMaxSize:     65536,  // Max 64K entries
            RegistryGrowStep:    128,    // Normal growth
            CallStackSize:       128,    // Normal call depth
            MinimizeStackMemory: false,
        }
    
    case "large":
        opts = lua.Options{
            RegistrySize:        4096,    // Large initial size
            RegistryMaxSize:     1048576, // Max 1M entries
            RegistryGrowStep:    1024,    // Fast growth
            CallStackSize:       256,     // Deep call stacks
            MinimizeStackMemory: false,
        }
    
    default:
        // Use defaults
        return lua.NewState()
    }
    
    return lua.NewState(opts)
}

// Example usage
func demonstrateRegistryLimits() {
    classes := []string{"minimal", "standard", "large"}
    
    for _, class := range classes {
        L := createLimitedLuaState(class)
        defer L.Close()
        
        fmt.Printf("\n=== %s configuration ===\n", class)
        
        // Test allocation limits
        err := L.DoString(`
            local count = 0
            local tables = {}
            
            -- Try to allocate many tables
            while count < 100000 do
                tables[count] = {id = count}
                count = count + 1
                
                -- Check every 1000
                if count % 1000 == 0 then
                    collectgarbage("collect")
                end
            end
            
            return count
        `)
        
        if err != nil {
            fmt.Printf("Failed at approximately %d allocations: %v\n", 
                L.MemUsage()/64, err) // Rough estimate
        } else {
            fmt.Printf("Successfully allocated 100000 tables\n")
            fmt.Printf("Memory usage: %d bytes\n", L.MemUsage())
        }
    }
}
```

## Memory Profiling

### Memory Usage Profiler

```go
package profiling

import (
    "fmt"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type MemoryProfiler struct {
    samples    []MemorySample
    maxSamples int
    interval   time.Duration
    lastSample time.Time
}

type MemorySample struct {
    Timestamp    time.Time
    MemoryUsage  int
    AllocRate    float64 // bytes per second
    StackDepth   int
    Description  string
}

func NewMemoryProfiler(maxSamples int, interval time.Duration) *MemoryProfiler {
    return &MemoryProfiler{
        maxSamples: maxSamples,
        interval:   interval,
        samples:    make([]MemorySample, 0, maxSamples),
    }
}

func (mp *MemoryProfiler) Sample(L *lua.LState, description string) {
    now := time.Now()
    if now.Sub(mp.lastSample) < mp.interval {
        return // Skip this sample
    }
    
    usage := L.MemUsage()
    
    var allocRate float64
    if len(mp.samples) > 0 {
        lastSample := mp.samples[len(mp.samples)-1]
        timeDelta := now.Sub(lastSample.Timestamp).Seconds()
        memDelta := float64(usage - lastSample.MemoryUsage)
        allocRate = memDelta / timeDelta
    }
    
    sample := MemorySample{
        Timestamp:    now,
        MemoryUsage:  usage,
        AllocRate:    allocRate,
        StackDepth:   L.GetTop(),
        Description:  description,
    }
    
    mp.samples = append(mp.samples, sample)
    if len(mp.samples) > mp.maxSamples {
        mp.samples = mp.samples[1:]
    }
    
    mp.lastSample = now
}

func (mp *MemoryProfiler) Report() string {
    if len(mp.samples) == 0 {
        return "No samples collected"
    }
    
    var totalMem int
    var peakMem int
    var peakAllocRate float64
    
    for _, s := range mp.samples {
        totalMem += s.MemoryUsage
        if s.MemoryUsage > peakMem {
            peakMem = s.MemoryUsage
        }
        if s.AllocRate > peakAllocRate {
            peakAllocRate = s.AllocRate
        }
    }
    
    avgMem := totalMem / len(mp.samples)
    duration := mp.samples[len(mp.samples)-1].Timestamp.Sub(mp.samples[0].Timestamp)
    
    return fmt.Sprintf(`Memory Profile Report:
- Duration: %v
- Samples: %d
- Average Memory: %s
- Peak Memory: %s
- Peak Alloc Rate: %s/sec
`, 
        duration,
        len(mp.samples),
        formatBytes(avgMem),
        formatBytes(peakMem),
        formatBytes(int(peakAllocRate)),
    )
}

// Example usage with profiling
func ProfileScriptExecution(script string) {
    L := lua.NewState()
    defer L.Close()
    
    profiler := NewMemoryProfiler(100, 10*time.Millisecond)
    
    // Set up profiling hook
    L.SetHook(func(L *lua.LState) {
        profiler.Sample(L, "execution")
    }, lua.Count, 1000)
    
    // Profile initial state
    profiler.Sample(L, "initial")
    
    // Execute script
    err := L.DoString(script)
    if err != nil {
        fmt.Printf("Script error: %v\n", err)
    }
    
    // Profile after execution
    profiler.Sample(L, "after_execution")
    
    // Force GC and profile
    L.DoString(`collectgarbage("collect")`)
    profiler.Sample(L, "after_gc")
    
    // Print report
    fmt.Println(profiler.Report())
}
```

## Memory Quota Management

### Per-Script Memory Quotas

```go
package quota

import (
    "fmt"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type MemoryQuotaManager struct {
    mu          sync.RWMutex
    totalQuota  int64
    available   int64
    allocations map[string]*ScriptAllocation
}

type ScriptAllocation struct {
    ScriptID    string
    Quota       int64
    Used        int64
    State       *lua.LState
    StartTime   time.Time
    LastChecked time.Time
}

func NewMemoryQuotaManager(totalQuotaMB int) *MemoryQuotaManager {
    totalBytes := int64(totalQuotaMB * 1024 * 1024)
    return &MemoryQuotaManager{
        totalQuota:  totalBytes,
        available:   totalBytes,
        allocations: make(map[string]*ScriptAllocation),
    }
}

func (mqm *MemoryQuotaManager) AllocateQuota(scriptID string, quotaMB int) (*lua.LState, error) {
    mqm.mu.Lock()
    defer mqm.mu.Unlock()
    
    quotaBytes := int64(quotaMB * 1024 * 1024)
    
    if quotaBytes > mqm.available {
        return nil, fmt.Errorf("insufficient quota: requested %d MB, available %d MB",
            quotaMB, mqm.available/1024/1024)
    }
    
    // Create limited Lua state
    opts := lua.Options{
        RegistryMaxSize: int(quotaBytes / 32), // Rough estimate
        CallStackSize:   128,
    }
    L := lua.NewState(opts)
    
    allocation := &ScriptAllocation{
        ScriptID:    scriptID,
        Quota:       quotaBytes,
        Used:        0,
        State:       L,
        StartTime:   time.Now(),
        LastChecked: time.Now(),
    }
    
    // Set up monitoring hook
    L.SetHook(func(L *lua.LState) {
        mqm.checkQuota(scriptID)
    }, lua.Count, 1000)
    
    mqm.allocations[scriptID] = allocation
    mqm.available -= quotaBytes
    
    return L, nil
}

func (mqm *MemoryQuotaManager) checkQuota(scriptID string) {
    mqm.mu.Lock()
    defer mqm.mu.Unlock()
    
    alloc, ok := mqm.allocations[scriptID]
    if !ok {
        return
    }
    
    currentUsage := int64(alloc.State.MemUsage())
    alloc.Used = currentUsage
    alloc.LastChecked = time.Now()
    
    if currentUsage > alloc.Quota {
        alloc.State.RaiseError("memory quota exceeded: %d/%d bytes",
            currentUsage, alloc.Quota)
    }
}

func (mqm *MemoryQuotaManager) ReleaseQuota(scriptID string) {
    mqm.mu.Lock()
    defer mqm.mu.Unlock()
    
    if alloc, ok := mqm.allocations[scriptID]; ok {
        mqm.available += alloc.Quota
        alloc.State.Close()
        delete(mqm.allocations, scriptID)
    }
}

func (mqm *MemoryQuotaManager) GetStatus() QuotaStatus {
    mqm.mu.RLock()
    defer mqm.mu.RUnlock()
    
    var totalUsed int64
    activeScripts := make([]ScriptStatus, 0, len(mqm.allocations))
    
    for _, alloc := range mqm.allocations {
        totalUsed += alloc.Used
        activeScripts = append(activeScripts, ScriptStatus{
            ScriptID:   alloc.ScriptID,
            QuotaMB:    alloc.Quota / 1024 / 1024,
            UsedMB:     alloc.Used / 1024 / 1024,
            Uptime:     time.Since(alloc.StartTime),
            Efficiency: float64(alloc.Used) / float64(alloc.Quota) * 100,
        })
    }
    
    return QuotaStatus{
        TotalQuotaMB:    mqm.totalQuota / 1024 / 1024,
        AvailableMB:     mqm.available / 1024 / 1024,
        AllocatedMB:     (mqm.totalQuota - mqm.available) / 1024 / 1024,
        ActualUsedMB:    totalUsed / 1024 / 1024,
        ActiveScripts:   activeScripts,
        OverallEfficiency: float64(totalUsed) / float64(mqm.totalQuota-mqm.available) * 100,
    }
}
```

## Testing Memory Behavior

### Memory Stress Tests

```go
package memtest

import (
    "fmt"
    "testing"
    lua "github.com/yuin/gopher-lua"
)

func TestMemoryAllocationPatterns(t *testing.T) {
    patterns := []struct {
        name   string
        script string
        desc   string
    }{
        {
            name: "linear_growth",
            script: `
                local data = {}
                for i = 1, 10000 do
                    data[i] = string.rep("x", 100)
                end
            `,
            desc: "Linear memory growth",
        },
        {
            name: "exponential_growth",
            script: `
                local data = ""
                for i = 1, 20 do
                    data = data .. data .. "x"
                end
            `,
            desc: "Exponential memory growth",
        },
        {
            name: "table_nesting",
            script: `
                local function deepNest(depth)
                    if depth == 0 then return {} end
                    return {child = deepNest(depth - 1)}
                end
                local nested = deepNest(1000)
            `,
            desc: "Deep table nesting",
        },
        {
            name: "circular_references",
            script: `
                for i = 1, 1000 do
                    local a, b = {}, {}
                    a.ref = b
                    b.ref = a
                    a.data = string.rep("x", 1000)
                end
                collectgarbage()
            `,
            desc: "Circular references with GC",
        },
    }
    
    for _, p := range patterns {
        t.Run(p.name, func(t *testing.T) {
            L := lua.NewState()
            defer L.Close()
            
            initial := L.MemUsage()
            
            // Track memory during execution
            var samples []int
            L.SetHook(func(L *lua.LState) {
                samples = append(samples, L.MemUsage())
            }, lua.Count, 10000)
            
            err := L.DoString(p.script)
            if err != nil {
                t.Logf("Script failed: %v", err)
            }
            
            peak := L.MemUsage()
            L.DoString(`collectgarbage("collect")`)
            afterGC := L.MemUsage()
            
            t.Logf("%s:", p.desc)
            t.Logf("  Initial: %s", formatBytes(initial))
            t.Logf("  Peak: %s", formatBytes(peak))
            t.Logf("  After GC: %s", formatBytes(afterGC))
            t.Logf("  Freed: %s", formatBytes(peak-afterGC))
            t.Logf("  Samples: %d", len(samples))
        })
    }
}

func formatBytes(bytes int) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
    )
    
    switch {
    case bytes >= GB:
        return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
    case bytes >= MB:
        return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
    case bytes >= KB:
        return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
    default:
        return fmt.Sprintf("%d B", bytes)
    }
}
```

## Complete Integration Example

### Memory-Limited Script Engine

```go
package engine

import (
    "context"
    "fmt"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type MemoryLimitedEngine struct {
    pool         *LStatePool
    quotaManager *MemoryQuotaManager
    profiler     *MemoryProfiler
    config       EngineConfig
}

type EngineConfig struct {
    TotalMemoryMB    int
    ScriptMemoryMB   int
    CheckIntervalMS  int
    EnableProfiling  bool
    EnableQuotas     bool
}

func NewMemoryLimitedEngine(config EngineConfig) *MemoryLimitedEngine {
    engine := &MemoryLimitedEngine{
        config: config,
    }
    
    if config.EnableQuotas {
        engine.quotaManager = NewMemoryQuotaManager(config.TotalMemoryMB)
    }
    
    if config.EnableProfiling {
        engine.profiler = NewMemoryProfiler(1000, time.Millisecond*time.Duration(config.CheckIntervalMS))
    }
    
    // Initialize pool with memory-aware states
    engine.pool = NewLStatePool(PoolConfig{
        MaxStates:      config.TotalMemoryMB / config.ScriptMemoryMB,
        StateMemoryMB:  config.ScriptMemoryMB,
    })
    
    return engine
}

func (e *MemoryLimitedEngine) Execute(ctx context.Context, scriptID, script string) (*ExecutionResult, error) {
    // Allocate quota if enabled
    var L *lua.LState
    var err error
    
    if e.quotaManager != nil {
        L, err = e.quotaManager.AllocateQuota(scriptID, e.config.ScriptMemoryMB)
        if err != nil {
            return nil, fmt.Errorf("quota allocation failed: %w", err)
        }
        defer e.quotaManager.ReleaseQuota(scriptID)
    } else {
        L = e.pool.Get()
        defer e.pool.Put(L)
    }
    
    // Set up memory monitoring
    controller := NewAdvancedMemoryController(MemoryConfig{
        HardLimit:     e.config.ScriptMemoryMB * 1024 * 1024,
        SoftLimit:     int(float64(e.config.ScriptMemoryMB) * 0.8 * 1024 * 1024),
        GCThreshold:   int(float64(e.config.ScriptMemoryMB) * 0.7 * 1024 * 1024),
        MinGCInterval: 100 * time.Millisecond,
    })
    
    controller.onSoftLimit = func(usage int) {
        // Notify script about approaching limit
        L.GetGlobal("onMemoryWarning")
        if L.IsFunction(-1) {
            L.Push(lua.LNumber(usage))
            L.Push(lua.LNumber(controller.hardLimit))
            L.Call(2, 0)
        } else {
            L.Pop(1)
        }
    }
    
    // Set hooks
    checkCounter := 0
    L.SetHook(func(L *lua.LState) {
        checkCounter++
        
        // Memory check
        if checkCounter%e.config.CheckIntervalMS == 0 {
            if err := controller.CheckMemory(L); err != nil {
                L.RaiseError(err.Error())
            }
        }
        
        // Profiling
        if e.profiler != nil && checkCounter%1000 == 0 {
            e.profiler.Sample(L, scriptID)
        }
        
        // Context check
        if ctx.Err() != nil {
            L.RaiseError("context cancelled")
        }
    }, lua.Count, 100)
    
    // Execute script
    start := time.Now()
    execErr := L.DoString(script)
    duration := time.Since(start)
    
    // Collect results
    result := &ExecutionResult{
        ScriptID:      scriptID,
        Success:       execErr == nil,
        Error:         execErr,
        Duration:      duration,
        MemoryMetrics: controller.GetMetrics(),
    }
    
    if e.profiler != nil {
        result.ProfileReport = e.profiler.Report()
    }
    
    return result, nil
}

// Example usage
func ExampleUsage() {
    engine := NewMemoryLimitedEngine(EngineConfig{
        TotalMemoryMB:   100,  // 100MB total
        ScriptMemoryMB:  10,   // 10MB per script
        CheckIntervalMS: 1000, // Check every 1000 instructions
        EnableProfiling: true,
        EnableQuotas:    true,
    })
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    result, err := engine.Execute(ctx, "test-script", `
        -- Memory-intensive operation
        local data = {}
        for i = 1, 100000 do
            data[i] = {
                id = i,
                value = string.rep("x", 100)
            }
            
            -- Cooperate with memory management
            if i % 10000 == 0 then
                collectgarbage("step")
            end
        end
        
        function onMemoryWarning(current, limit)
            print(string.format("Memory warning: %d/%d bytes", current, limit))
            -- Clean up some data
            for i = 1, #data/2 do
                data[i] = nil
            end
            collectgarbage("collect")
        end
        
        return #data
    `)
    
    if err != nil {
        fmt.Printf("Execution failed: %v\n", err)
    } else {
        fmt.Printf("Result: %+v\n", result)
    }
}
```

## Summary

These examples demonstrate comprehensive memory management in GopherLua:
1. Basic memory monitoring and limits
2. Advanced controllers with GC integration
3. Registry configuration for different memory profiles
4. Memory profiling and analysis
5. Quota-based multi-tenant systems
6. Complete integration with script engines

Key patterns include soft/hard limits, automatic GC triggering, memory profiling, and quota management for multi-script environments.