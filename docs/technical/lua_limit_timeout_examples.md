# Lua Instruction Limits and Timeout Examples

This document provides practical examples of implementing instruction count limits and timeout mechanisms in GopherLua.

## Basic Instruction Counting

### Simple Counter Hook

```go
package main

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
)

func main() {
    L := lua.NewState()
    defer L.Close()
    
    instructionCount := 0
    
    // Set hook to count every 10 instructions
    L.SetHook(func(L *lua.LState) {
        instructionCount += 10
        if instructionCount > 1000 {
            L.RaiseError("instruction limit exceeded: %d", instructionCount)
        }
    }, lua.Count, 10)
    
    // This will trigger the limit
    err := L.DoString(`
        local sum = 0
        for i = 1, 1000 do
            sum = sum + i
        end
        return sum
    `)
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Context-Based Timeouts

### Basic Timeout Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    lua "github.com/yuin/gopher-lua"
)

func executeWithTimeout(script string, timeout time.Duration) error {
    L := lua.NewState()
    defer L.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    L.SetContext(ctx)
    
    // Execute script
    err := L.DoString(script)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("script timeout after %v", timeout)
        }
        return err
    }
    
    return nil
}

func main() {
    // This will timeout
    err := executeWithTimeout(`
        while true do
            -- Infinite loop
        end
    `, 2*time.Second)
    
    fmt.Printf("Result: %v\n", err)
}
```

## Combined Resource Limiter

### Full Implementation Example

```go
package resourcelimit

import (
    "context"
    "fmt"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type ResourceLimiter struct {
    mu               sync.RWMutex
    instructionLimit int
    instructionCount int
    timeout          time.Duration
    startTime        time.Time
    memoryLimit      int
    checkInterval    int
    
    // Metrics
    peakInstructions int
    peakMemory       int
}

func NewResourceLimiter(opts LimiterOptions) *ResourceLimiter {
    return &ResourceLimiter{
        instructionLimit: opts.InstructionLimit,
        timeout:          opts.Timeout,
        memoryLimit:      opts.MemoryLimit,
        checkInterval:    opts.CheckInterval,
        startTime:        time.Now(),
    }
}

func (rl *ResourceLimiter) Hook(L *lua.LState) {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    // Update counts
    rl.instructionCount += rl.checkInterval
    currentMem := L.MemUsage()
    
    // Track peaks
    if rl.instructionCount > rl.peakInstructions {
        rl.peakInstructions = rl.instructionCount
    }
    if currentMem > rl.peakMemory {
        rl.peakMemory = currentMem
    }
    
    // Check limits
    if rl.instructionLimit > 0 && rl.instructionCount >= rl.instructionLimit {
        L.RaiseError("instruction limit exceeded: %d/%d", 
            rl.instructionCount, rl.instructionLimit)
    }
    
    if rl.timeout > 0 && time.Since(rl.startTime) > rl.timeout {
        L.RaiseError("execution timeout: %v", time.Since(rl.startTime))
    }
    
    if rl.memoryLimit > 0 && currentMem > rl.memoryLimit {
        L.RaiseError("memory limit exceeded: %d/%d bytes", 
            currentMem, rl.memoryLimit)
    }
}

func (rl *ResourceLimiter) GetMetrics() ResourceMetrics {
    rl.mu.RLock()
    defer rl.mu.RUnlock()
    
    return ResourceMetrics{
        InstructionCount: rl.instructionCount,
        PeakInstructions: rl.peakInstructions,
        PeakMemory:       rl.peakMemory,
        Duration:         time.Since(rl.startTime),
    }
}

// Usage example
func ExecuteScriptWithLimits(script string) error {
    L := lua.NewState()
    defer L.Close()
    
    limiter := NewResourceLimiter(LimiterOptions{
        InstructionLimit: 1_000_000,
        Timeout:          5 * time.Second,
        MemoryLimit:      10 * 1024 * 1024, // 10MB
        CheckInterval:    1000,
    })
    
    L.SetHook(limiter.Hook, lua.Count, limiter.checkInterval)
    
    err := L.DoString(script)
    
    metrics := limiter.GetMetrics()
    fmt.Printf("Execution metrics: %+v\n", metrics)
    
    return err
}
```

## Adaptive Limiting

### Dynamic Check Intervals

```go
type AdaptiveLimiter struct {
    *ResourceLimiter
    baseInterval     int
    minInterval      int
    warningThreshold float64
}

func (al *AdaptiveLimiter) AdaptiveHook(L *lua.LState) {
    al.mu.Lock()
    
    // Calculate utilization
    instrUtil := float64(al.instructionCount) / float64(al.instructionLimit)
    timeUtil := float64(time.Since(al.startTime)) / float64(al.timeout)
    maxUtil := math.Max(instrUtil, timeUtil)
    
    // Adjust check interval based on utilization
    if maxUtil > al.warningThreshold {
        // Near limit - check more frequently
        al.checkInterval = al.minInterval
    } else if maxUtil < 0.5 {
        // Far from limit - check less frequently
        al.checkInterval = al.baseInterval * 2
    }
    
    al.mu.Unlock()
    
    // Call base hook
    al.ResourceLimiter.Hook(L)
}

// Usage
func ExecuteWithAdaptiveLimits(script string) error {
    L := lua.NewState()
    defer L.Close()
    
    limiter := &AdaptiveLimiter{
        ResourceLimiter: NewResourceLimiter(LimiterOptions{
            InstructionLimit: 10_000_000,
            Timeout:          30 * time.Second,
            MemoryLimit:      50 * 1024 * 1024,
            CheckInterval:    10000,
        }),
        baseInterval:     10000,
        minInterval:      100,
        warningThreshold: 0.8,
    }
    
    L.SetHook(limiter.AdaptiveHook, lua.Count, limiter.checkInterval)
    
    return L.DoString(script)
}
```

## Graceful Warning System

### Soft Limits with Callbacks

```go
type GracefulLimiter struct {
    *ResourceLimiter
    softInstructionLimit int
    softTimeout          time.Duration
    onWarning            func(warning string)
    warningsIssued       map[string]bool
}

func (gl *GracefulLimiter) GracefulHook(L *lua.LState) {
    gl.mu.Lock()
    defer gl.mu.Unlock()
    
    // Check soft limits first
    if gl.softInstructionLimit > 0 && 
       gl.instructionCount >= gl.softInstructionLimit &&
       !gl.warningsIssued["instruction"] {
        
        warning := fmt.Sprintf("Warning: approaching instruction limit (%d/%d)",
            gl.instructionCount, gl.instructionLimit)
        
        gl.warningsIssued["instruction"] = true
        
        // Call Lua warning function if exists
        L.GetGlobal("onResourceWarning")
        if L.IsFunction(-1) {
            L.Push(lua.LString(warning))
            L.Push(lua.LString("instruction"))
            L.Push(lua.LNumber(float64(gl.instructionCount) / float64(gl.instructionLimit)))
            L.Call(3, 0)
        } else {
            L.Pop(1)
        }
        
        // Call Go warning callback
        if gl.onWarning != nil {
            gl.onWarning(warning)
        }
    }
    
    // Check timeout soft limit
    elapsed := time.Since(gl.startTime)
    if gl.softTimeout > 0 && 
       elapsed >= gl.softTimeout &&
       !gl.warningsIssued["timeout"] {
        
        warning := fmt.Sprintf("Warning: approaching timeout (%v/%v)",
            elapsed, gl.timeout)
        
        gl.warningsIssued["timeout"] = true
        
        if gl.onWarning != nil {
            gl.onWarning(warning)
        }
    }
    
    // Call base hook for hard limits
    gl.ResourceLimiter.Hook(L)
}

// Usage with warning handling
func ExecuteWithWarnings(script string) error {
    L := lua.NewState()
    defer L.Close()
    
    // Define warning handler in Lua
    L.DoString(`
        function onResourceWarning(message, type, utilization)
            print(string.format("[LUA WARNING] %s (%.1f%% used)", 
                message, utilization * 100))
            
            -- Script can take action
            if type == "instruction" and utilization > 0.9 then
                -- Try to wrap up work
                if cleanupFunction then
                    cleanupFunction()
                end
            end
        end
    `)
    
    limiter := &GracefulLimiter{
        ResourceLimiter: NewResourceLimiter(LimiterOptions{
            InstructionLimit: 1_000_000,
            Timeout:          10 * time.Second,
            CheckInterval:    1000,
        }),
        softInstructionLimit: 800_000,  // Warn at 80%
        softTimeout:          8 * time.Second,
        warningsIssued:       make(map[string]bool),
        onWarning: func(warning string) {
            fmt.Printf("[GO WARNING] %s\n", warning)
        },
    }
    
    L.SetHook(limiter.GracefulHook, lua.Count, limiter.checkInterval)
    
    return L.DoString(script)
}
```

## Testing Resource Limits

### Test Helper Functions

```go
package testing

import (
    "testing"
    "time"
    lua "github.com/yuin/gopher-lua"
)

func TestInstructionLimits(t *testing.T) {
    tests := []struct {
        name          string
        script        string
        limit         int
        expectError   bool
        errorContains string
    }{
        {
            name: "simple loop within limit",
            script: `
                local sum = 0
                for i = 1, 100 do
                    sum = sum + i
                end
                return sum
            `,
            limit:       10000,
            expectError: false,
        },
        {
            name: "infinite loop exceeds limit",
            script: `
                while true do
                    -- This will exceed any limit
                end
            `,
            limit:         1000,
            expectError:   true,
            errorContains: "instruction limit exceeded",
        },
        {
            name: "recursive function",
            script: `
                function fib(n)
                    if n <= 1 then return n end
                    return fib(n-1) + fib(n-2)
                end
                return fib(30)
            `,
            limit:         100000,
            expectError:   true,
            errorContains: "instruction limit exceeded",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            L := lua.NewState()
            defer L.Close()
            
            count := 0
            L.SetHook(func(L *lua.LState) {
                count += 100
                if count >= tt.limit {
                    L.RaiseError("instruction limit exceeded: %d", count)
                }
            }, lua.Count, 100)
            
            err := L.DoString(tt.script)
            
            if tt.expectError {
                if err == nil {
                    t.Errorf("expected error but got none")
                } else if tt.errorContains != "" && 
                         !strings.Contains(err.Error(), tt.errorContains) {
                    t.Errorf("expected error containing %q, got %v", 
                        tt.errorContains, err)
                }
            } else if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

## Performance Benchmarks

### Hook Overhead Measurement

```go
func BenchmarkHookOverhead(b *testing.B) {
    script := `
        local sum = 0
        for i = 1, 10000 do
            sum = sum + i
        end
        return sum
    `
    
    intervals := []int{1, 10, 100, 1000, 10000}
    
    for _, interval := range intervals {
        b.Run(fmt.Sprintf("interval_%d", interval), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                L := lua.NewState()
                
                if interval > 0 {
                    count := 0
                    L.SetHook(func(L *lua.LState) {
                        count++
                    }, lua.Count, interval)
                }
                
                L.DoString(script)
                L.Close()
            }
        })
    }
}
```

## Integration with Script Engine

### Complete Example with Engine

```go
type LuaEngine struct {
    pool    *LStatePool
    limits  ScriptLimits
    metrics *MetricsCollector
}

func (e *LuaEngine) ExecuteScript(ctx context.Context, script string, opts ExecutionOptions) (*ScriptResult, error) {
    // Get state from pool
    L := e.pool.Get()
    defer e.pool.Put(L)
    
    // Apply context
    if ctx == nil {
        ctx = context.Background()
    }
    ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
    defer cancel()
    
    L.SetContext(ctx)
    
    // Create resource limiter
    limiter := NewResourceLimiter(LimiterOptions{
        InstructionLimit: opts.InstructionLimit,
        Timeout:          opts.Timeout,
        MemoryLimit:      opts.MemoryLimit,
        CheckInterval:    opts.CheckInterval,
    })
    
    // Set hook
    L.SetHook(limiter.Hook, lua.Count, limiter.checkInterval)
    
    // Execute script
    start := time.Now()
    err := L.DoString(script)
    duration := time.Since(start)
    
    // Collect metrics
    metrics := limiter.GetMetrics()
    e.metrics.RecordExecution(metrics, duration, err)
    
    // Build result
    result := &ScriptResult{
        Success:  err == nil,
        Error:    err,
        Duration: duration,
        Metrics:  metrics,
    }
    
    // Extract return values if successful
    if err == nil {
        result.Values = extractReturnValues(L)
    }
    
    return result, nil
}
```

## Common Patterns

### Pattern: Progressive Limits

```go
// Start with strict limits, relax if needed
limits := []ScriptLimits{
    {InstructionLimit: 100_000, Timeout: 1 * time.Second},
    {InstructionLimit: 1_000_000, Timeout: 5 * time.Second},
    {InstructionLimit: 10_000_000, Timeout: 30 * time.Second},
}

for i, limit := range limits {
    result, err := engine.Execute(script, limit)
    if err == nil {
        return result, nil
    }
    
    if isResourceError(err) && i < len(limits)-1 {
        log.Printf("Retrying with relaxed limits: %+v", limits[i+1])
        continue
    }
    
    return nil, err
}
```

### Pattern: Limit Profiles

```go
profiles := map[string]ScriptLimits{
    "interactive": {
        InstructionLimit: 100_000,
        Timeout:          1 * time.Second,
        CheckInterval:    100,
    },
    "background": {
        InstructionLimit: 10_000_000,
        Timeout:          5 * time.Minute,
        CheckInterval:    10000,
    },
    "batch": {
        InstructionLimit: 0, // No limit
        Timeout:          1 * time.Hour,
        CheckInterval:    100000,
    },
}

// Use based on context
limits := profiles[executionContext]
```

## Summary

These examples demonstrate practical implementations of instruction counting and timeout mechanisms in GopherLua, from basic usage to advanced patterns with adaptive limiting, graceful warnings, and comprehensive testing strategies.