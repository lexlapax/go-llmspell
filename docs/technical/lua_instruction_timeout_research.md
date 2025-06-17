# Lua Instruction Count Limits and Timeout Mechanisms Research

This document investigates GopherLua's instruction count limits and timeout mechanisms for implementing resource control and preventing infinite loops in script execution.

## Executive Summary

GopherLua provides built-in support for instruction counting through Lua's debug hooks, enabling precise control over script execution limits. Combined with Go's context and timer mechanisms, we can implement robust timeout and resource limiting capabilities.

## Instruction Count Limits

### GopherLua Hook System

GopherLua implements Lua 5.1's debug hook system, allowing us to set hooks that trigger after a specific number of VM instructions:

```go
// Set a hook that triggers every N instructions
L.SetContext(ctx)
L.CallByParam(lua.P{
    Fn:      function,
    NRet:    0,
    Protect: true,
}, args...)
```

### Hook-Based Instruction Counting

```go
type InstructionLimiter struct {
    limit    int
    count    int
    exceeded bool
}

func (il *InstructionLimiter) Hook(L *lua.LState) {
    il.count++
    if il.count >= il.limit {
        il.exceeded = true
        L.RaiseError("instruction limit exceeded")
    }
}

// Usage
limiter := &InstructionLimiter{limit: 1000000}
L.SetHook(limiter.Hook, lua.Count, 100) // Check every 100 instructions
```

### Instruction Counting Characteristics

1. **Granularity**: Hooks can be set to trigger every N instructions
2. **Overhead**: Frequent hooks (e.g., every instruction) add significant overhead
3. **Accuracy**: Instruction count is approximate due to hook intervals
4. **Types**: Count hooks, line hooks, call/return hooks available

## Timeout Mechanisms

### Context-Based Timeouts

GopherLua supports Go contexts for cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

L.SetContext(ctx)
err := L.DoString(script)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("script timeout after 5 seconds")
    }
    return err
}
```

### Hook-Based Timeouts

Alternative approach using time-based hooks:

```go
type TimeoutEnforcer struct {
    start   time.Time
    timeout time.Duration
}

func (te *TimeoutEnforcer) Hook(L *lua.LState) {
    if time.Since(te.start) > te.timeout {
        L.RaiseError("execution timeout")
    }
}

// Usage
enforcer := &TimeoutEnforcer{
    start:   time.Now(),
    timeout: 5 * time.Second,
}
L.SetHook(enforcer.Hook, lua.Count, 1000) // Check every 1000 instructions
```

## Combined Limits Implementation

### Comprehensive Resource Limiter

```go
type ResourceLimiter struct {
    instructionLimit int
    instructionCount int
    timeout          time.Duration
    startTime        time.Time
    memoryLimit      int
    checkInterval    int
}

func NewResourceLimiter(instrLimit int, timeout time.Duration, memLimit int) *ResourceLimiter {
    return &ResourceLimiter{
        instructionLimit: instrLimit,
        timeout:          timeout,
        memoryLimit:      memLimit,
        checkInterval:    1000, // Check every 1000 instructions
        startTime:        time.Now(),
    }
}

func (rl *ResourceLimiter) Hook(L *lua.LState) {
    // Check instruction count
    rl.instructionCount += rl.checkInterval
    if rl.instructionCount >= rl.instructionLimit {
        L.RaiseError("instruction limit %d exceeded", rl.instructionLimit)
    }
    
    // Check timeout
    if time.Since(rl.startTime) > rl.timeout {
        L.RaiseError("execution timeout after %v", rl.timeout)
    }
    
    // Check memory (approximate)
    if L.MemUsage() > rl.memoryLimit {
        L.RaiseError("memory limit %d exceeded", rl.memoryLimit)
    }
}
```

### Context Integration

```go
type LuaExecutor struct {
    pool           *LStatePool
    defaultTimeout time.Duration
    defaultLimit   int
}

func (le *LuaExecutor) Execute(script string, opts ExecutionOptions) error {
    ctx := opts.Context
    if ctx == nil {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(context.Background(), le.defaultTimeout)
        defer cancel()
    }
    
    L := le.pool.Get()
    defer le.pool.Put(L)
    
    // Set context for cancellation
    L.SetContext(ctx)
    
    // Set resource limiter hook
    limiter := NewResourceLimiter(
        opts.InstructionLimit,
        opts.Timeout,
        opts.MemoryLimit,
    )
    L.SetHook(limiter.Hook, lua.Count, limiter.checkInterval)
    
    // Execute with protection
    return L.DoString(script)
}
```

## Performance Considerations

### Hook Overhead Analysis

| Check Interval | Overhead | Use Case |
|----------------|----------|----------|
| 1              | Very High (~50-100%) | Debugging only |
| 100            | High (~10-20%) | Strict security |
| 1000           | Moderate (~2-5%) | Normal security |
| 10000          | Low (~0.5-1%) | Performance-critical |

### Optimization Strategies

1. **Dynamic Intervals**: Adjust check interval based on script trust level
2. **Adaptive Checking**: Increase frequency near limits
3. **Cached Decisions**: Store security decisions to avoid repeated checks
4. **Batch Operations**: Check multiple conditions in single hook

```go
func (rl *ResourceLimiter) AdaptiveHook(L *lua.LState) {
    rl.instructionCount += rl.checkInterval
    
    // Increase check frequency as we approach limits
    utilizationRatio := float64(rl.instructionCount) / float64(rl.instructionLimit)
    if utilizationRatio > 0.8 {
        rl.checkInterval = 100 // Check more frequently near limit
    }
    
    // Batch all checks
    if rl.shouldTerminate() {
        L.RaiseError(rl.terminationReason())
    }
}
```

## Error Handling and Recovery

### Graceful Termination

```go
type GracefulLimiter struct {
    *ResourceLimiter
    warningThreshold float64
    warningIssued    bool
}

func (gl *GracefulLimiter) Hook(L *lua.LState) {
    utilization := float64(gl.instructionCount) / float64(gl.instructionLimit)
    
    // Issue warning at threshold
    if utilization > gl.warningThreshold && !gl.warningIssued {
        // Push warning to Lua script
        L.GetGlobal("onResourceWarning")
        if L.IsFunction(-1) {
            L.Push(lua.LNumber(utilization))
            L.Call(1, 0)
        } else {
            L.Pop(1)
        }
        gl.warningIssued = true
    }
    
    // Hard limit
    if utilization >= 1.0 {
        L.RaiseError("instruction limit exceeded")
    }
}
```

### State Cleanup on Termination

```go
func (le *LuaExecutor) ExecuteWithCleanup(script string) error {
    L := le.pool.Get()
    defer func() {
        // Always return state to pool, even on panic
        if r := recover(); r != nil {
            L.SetTop(0) // Clear stack
            le.pool.Put(L)
            panic(r) // Re-panic
        }
        le.pool.Put(L)
    }()
    
    // Set hooks and execute
    // ...
}
```

## Integration with Script Engine

### Configuration Structure

```go
type ScriptLimits struct {
    InstructionLimit   int           `json:"instruction_limit"`
    Timeout            time.Duration `json:"timeout"`
    MemoryLimit        int           `json:"memory_limit"`
    CheckInterval      int           `json:"check_interval"`
    WarningThreshold   float64       `json:"warning_threshold"`
    EnableProfiling    bool          `json:"enable_profiling"`
}

type LuaEngineConfig struct {
    PoolSize        int          `json:"pool_size"`
    DefaultLimits   ScriptLimits `json:"default_limits"`
    SecurityLevel   string       `json:"security_level"`
    EnableDebugHook bool         `json:"enable_debug_hook"`
}
```

### Profile-Based Limits

```go
var SecurityProfiles = map[string]ScriptLimits{
    "strict": {
        InstructionLimit: 1_000_000,
        Timeout:          5 * time.Second,
        MemoryLimit:      10 * 1024 * 1024, // 10MB
        CheckInterval:    100,
        WarningThreshold: 0.8,
    },
    "normal": {
        InstructionLimit: 10_000_000,
        Timeout:          30 * time.Second,
        MemoryLimit:      50 * 1024 * 1024, // 50MB
        CheckInterval:    1000,
        WarningThreshold: 0.9,
    },
    "relaxed": {
        InstructionLimit: 100_000_000,
        Timeout:          5 * time.Minute,
        MemoryLimit:      200 * 1024 * 1024, // 200MB
        CheckInterval:    10000,
        WarningThreshold: 0.95,
    },
}
```

## Testing Strategies

### Unit Tests for Limits

```go
func TestInstructionLimit(t *testing.T) {
    tests := []struct {
        name     string
        script   string
        limit    int
        shouldFail bool
    }{
        {
            name:   "within limit",
            script: `for i=1,100 do end`,
            limit:  10000,
            shouldFail: false,
        },
        {
            name:   "exceed limit",
            script: `while true do end`,
            limit:  1000,
            shouldFail: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Benchmark Tests

```go
func BenchmarkHookOverhead(b *testing.B) {
    intervals := []int{1, 10, 100, 1000, 10000}
    
    for _, interval := range intervals {
        b.Run(fmt.Sprintf("interval_%d", interval), func(b *testing.B) {
            // Benchmark with different check intervals
        })
    }
}
```

## Best Practices

1. **Default Limits**: Always set reasonable defaults
2. **User Control**: Allow script-specific limit overrides
3. **Monitoring**: Log limit violations for analysis
4. **Graceful Degradation**: Warn before hard limits
5. **Testing**: Test limits with worst-case scripts

## Implementation Checklist

- [ ] Basic instruction counting hook
- [ ] Context-based timeout integration
- [ ] Combined resource limiter
- [ ] Adaptive check intervals
- [ ] Warning system for soft limits
- [ ] Profile-based configuration
- [ ] Metrics and monitoring
- [ ] Comprehensive test suite
- [ ] Performance benchmarks
- [ ] Documentation and examples

## References

- [GopherLua Debug Hooks](https://github.com/yuin/gopher-lua#lua-module-debug)
- [Lua 5.1 Debug Library](https://www.lua.org/manual/5.1/manual.html#5.9)
- [Go Context Best Practices](https://go.dev/blog/context)
- [Resource Limiting Patterns](https://github.com/yuin/gopher-lua/issues/124)