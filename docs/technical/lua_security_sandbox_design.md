# Lua Security Sandbox Design

## Overview
This document designs a comprehensive security sandbox for executing untrusted Lua scripts in go-llmspell, ensuring scripts cannot access system resources or escape the sandbox.

## Security Principles

### 1. Whitelist Approach
- Start with nothing, add only what's proven safe
- Explicitly allow functions rather than blocking dangerous ones
- Default deny for all system access

### 2. Defense in Depth
- Multiple layers of security
- Resource limits (CPU, memory, execution time)
- Input validation and sanitization
- Monitoring and logging

## Sandbox Architecture

```go
// LuaSandbox provides a secure execution environment
type LuaSandbox struct {
    config      SandboxConfig
    whitelists  Whitelists
    limits      ResourceLimits
    monitor     SecurityMonitor
}

// SandboxConfig defines security settings
type SandboxConfig struct {
    // Library control
    AllowedLibraries   []string          // Whitelist of safe libraries
    BlockedFunctions   map[string]bool   // Additional function blacklist
    
    // Resource limits
    MaxInstructions    int64             // Instruction count limit
    MaxMemory          int64             // Memory usage limit (bytes)
    MaxExecutionTime   time.Duration     // Total execution timeout
    MaxStackDepth      int               // Call stack depth limit
    MaxTableSize       int               // Limit table sizes
    MaxStringLength    int               // Limit string operations
    
    // Security features
    DisableBytecode    bool              // Prevent bytecode loading
    DisableDebugInfo   bool              // Strip debug information
    IsolateGlobals     bool              // Separate global environment
    
    // Monitoring
    EnableAuditing     bool              // Log all operations
    EnableProfiling    bool              // Track resource usage
}

// Whitelists define allowed operations
type Whitelists struct {
    Functions   map[string]bool       // Allowed global functions
    Modules     map[string]bool       // Allowed require modules
    Metatables  map[string]bool       // Allowed metatable access
    Properties  map[string]bool       // Allowed property access
}
```

## Library Security Matrix

| Library | Safety | Allowed Functions | Notes |
|---------|--------|------------------|-------|
| `base` | Partial | `assert`, `error`, `ipairs`, `next`, `pairs`, `pcall`, `print`, `select`, `tonumber`, `tostring`, `type`, `unpack`, `xpcall` | Remove: `dofile`, `load`, `loadfile`, `loadstring`, `require`, `rawget`, `rawset`, `getmetatable`, `setmetatable` |
| `coroutine` | Safe | All | Needed for async operations |
| `string` | Mostly Safe | All except pattern matching with `%b` | Potential DoS with complex patterns |
| `table` | Safe | All | No system access |
| `math` | Safe | All | Pure computation |
| `io` | **UNSAFE** | None | File system access |
| `os` | **UNSAFE** | `time`, `difftime`, `clock` only | System command execution |
| `debug` | **UNSAFE** | None | Can break sandbox |
| `package` | **UNSAFE** | None | Can load arbitrary code |

## Sandbox Implementation

### 1. State Initialization
```go
func (s *LuaSandbox) CreateSafeState() (*lua.LState, error) {
    // Create state without standard libraries
    L := lua.NewState(lua.Options{
        SkipOpenLibs:      true,
        CallStackSize:     s.config.MaxStackDepth,
        RegistrySize:      1024 * 20,  // Controlled size
        RegistryMaxSize:   1024 * 80,
        RegistryGrowStep:  32,
    })
    
    // Load only safe libraries
    for _, lib := range s.config.AllowedLibraries {
        switch lib {
        case "base":
            s.loadSafeBase(L)
        case "table":
            lua.OpenTable(L)
        case "string":
            s.loadSafeString(L)
        case "math":
            lua.OpenMath(L)
        case "coroutine":
            lua.OpenCoroutine(L)
        default:
            return nil, fmt.Errorf("unknown library: %s", lib)
        }
    }
    
    // Apply additional restrictions
    s.applyRestrictions(L)
    
    // Set up monitoring
    s.setupMonitoring(L)
    
    return L, nil
}
```

### 2. Safe Base Library
```go
func (s *LuaSandbox) loadSafeBase(L *lua.LState) {
    // Load base library
    lua.OpenBase(L)
    
    // Remove dangerous functions
    dangerousFuncs := []string{
        "dofile", "load", "loadfile", "loadstring",
        "require", "rawget", "rawset", 
        "getmetatable", "setmetatable",
        "getfenv", "setfenv", // Lua 5.1
        "collectgarbage", // Can affect performance
    }
    
    for _, fn := range dangerousFuncs {
        L.SetGlobal(fn, lua.LNil)
    }
    
    // Override print for controlled output
    L.SetGlobal("print", L.NewFunction(s.safePrint))
}

func (s *LuaSandbox) safePrint(L *lua.LState) int {
    n := L.GetTop()
    parts := make([]string, n)
    
    for i := 1; i <= n; i++ {
        parts[i-1] = L.ToStringMeta(L.Get(i)).String()
        
        // Check string length
        if len(parts[i-1]) > s.config.MaxStringLength {
            parts[i-1] = parts[i-1][:s.config.MaxStringLength] + "..."
        }
    }
    
    // Log to sandbox output
    s.monitor.LogOutput(strings.Join(parts, "\t"))
    
    return 0
}
```

### 3. Resource Limits

#### Instruction Count Limiting
```go
func (s *LuaSandbox) setupInstructionLimit(L *lua.LState, limit int64) {
    count := int64(0)
    
    L.SetContext(context.WithValue(L.Context(), "instruction_count", &count))
    
    // Set debug hook for instruction counting
    L.SetDebugHook(func(L *lua.LState, event lua.DebugHookEvent) {
        if event == lua.HookCount {
            atomic.AddInt64(&count, 1)
            
            if atomic.LoadInt64(&count) > limit {
                L.RaiseError("instruction limit exceeded")
            }
        }
    }, lua.HookCount, 100) // Check every 100 instructions
}
```

#### Memory Limiting
```go
func (s *LuaSandbox) setupMemoryLimit(L *lua.LState, limit int64) {
    // Monitor registry size
    originalNewTable := L.NewTable
    L.NewTable = func() *lua.LTable {
        // Check current memory usage
        if s.getMemoryUsage(L) > limit {
            L.RaiseError("memory limit exceeded")
        }
        return originalNewTable()
    }
    
    // Monitor string creation
    // (Similar wrapping for string operations)
}
```

#### Execution Timeout
```go
func (s *LuaSandbox) ExecuteWithTimeout(L *lua.LState, script string, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    L.SetContext(ctx)
    
    done := make(chan error, 1)
    go func() {
        done <- L.DoString(script)
    }()
    
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        // Force stop execution
        L.RaiseError("execution timeout")
        return fmt.Errorf("script execution timeout after %v", timeout)
    }
}
```

### 4. Input Validation

```go
// ValidateScript checks script for dangerous patterns
func (s *LuaSandbox) ValidateScript(script string) error {
    // Check for bytecode header
    if strings.HasPrefix(script, "\x1bLua") {
        return ErrBytecodeNotAllowed
    }
    
    // Check for suspicious patterns
    dangerousPatterns := []string{
        `load\s*\(`,        // Dynamic code loading
        `debug\.`,          // Debug library access
        `io\.`,             // IO library access
        `os\.`,             // OS library access
        `package\.`,        // Package library access
        `_G\[`,             // Global table manipulation
        `rawset`,           // Raw table access
        `rawget`,           
        `getmetatable`,     // Metatable manipulation
        `setmetatable`,
    }
    
    for _, pattern := range dangerousPatterns {
        if matched, _ := regexp.MatchString(pattern, script); matched {
            return fmt.Errorf("dangerous pattern detected: %s", pattern)
        }
    }
    
    return nil
}
```

### 5. Global Environment Isolation

```go
func (s *LuaSandbox) createIsolatedEnvironment(L *lua.LState) *lua.LTable {
    // Create new environment table
    env := L.NewTable()
    
    // Copy only safe globals
    globals := L.Get(lua.GlobalsIndex).(*lua.LTable)
    globals.ForEach(func(k, v lua.LValue) {
        key := lua.LVAsString(k)
        
        if s.whitelists.Functions[key] {
            env.RawSetString(key, v)
        }
    })
    
    // Add sandbox-specific globals
    env.RawSetString("_VERSION", lua.LString("Sandbox 1.0"))
    env.RawSetString("_SANDBOX", lua.LTrue)
    
    return env
}
```

### 6. Metatable Protection

```go
func (s *LuaSandbox) protectMetatables(L *lua.LState) {
    // Protect string metatable
    L.DoString(`
        local string_mt = getmetatable("")
        if string_mt then
            string_mt.__metatable = false
        end
    `)
    
    // Remove metatable access functions
    L.SetGlobal("getmetatable", lua.LNil)
    L.SetGlobal("setmetatable", lua.LNil)
    L.SetGlobal("rawget", lua.LNil)
    L.SetGlobal("rawset", lua.LNil)
}
```

## Security Monitoring

```go
// SecurityMonitor tracks sandbox activity
type SecurityMonitor struct {
    logs         []SecurityEvent
    violations   []Violation
    metrics      SecurityMetrics
    alertHandler AlertHandler
}

type SecurityEvent struct {
    Timestamp   time.Time
    EventType   string
    Details     map[string]interface{}
    ScriptID    string
    Severity    SecuritySeverity
}

type Violation struct {
    Timestamp   time.Time
    Type        ViolationType
    Message     string
    ScriptID    string
    StackTrace  string
}

func (m *SecurityMonitor) RecordViolation(v Violation) {
    m.violations = append(m.violations, v)
    
    if v.Type == ViolationTypeCritical {
        m.alertHandler.Alert(v)
    }
}
```

## Integration with Bridges

### Safe Bridge Access
```go
func (s *LuaSandbox) registerSafeBridge(L *lua.LState, bridge engine.Bridge) {
    bridgeTable := L.NewTable()
    
    for _, method := range bridge.Methods() {
        // Wrap method with security checks
        wrappedMethod := s.wrapBridgeMethod(bridge, method)
        bridgeTable.RawSetString(method.Name, wrappedMethod)
    }
    
    // Register with read-only metatable
    mt := L.NewTable()
    L.SetField(mt, "__index", bridgeTable)
    L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
        L.RaiseError("cannot modify bridge")
        return 0
    }))
    L.SetField(mt, "__metatable", lua.LFalse)
    
    proxy := L.NewTable()
    L.SetMetatable(proxy, mt)
    
    L.SetGlobal(bridge.GetID(), proxy)
}
```

## Testing Strategy

### 1. Escape Attempt Tests
```go
// Test various sandbox escape attempts
escapeAttempts := []string{
    `load("os.execute('rm -rf /')")()`,
    `require("os").execute("whoami")`,
    `debug.getinfo(print).func("malicious")`,
    `getmetatable("").__index = function() os.execute("bad") end`,
    `_G["os"] = require("os")`,
    `rawset(_G, "os", require("os"))`,
}
```

### 2. Resource Exhaustion Tests
```go
// Test resource limits
exhaustionTests := []string{
    `while true do end`,                    // Infinite loop
    `local t = {} while true do t[#t+1] = t end`, // Memory bomb
    `string.rep("x", 2^30)`,               // Large string
    `function f() return f() end f()`,     // Stack overflow
}
```

### 3. Performance Tests
- Measure sandbox overhead
- Test concurrent sandbox execution
- Benchmark with/without monitoring

## Best Practices

1. **Principle of Least Privilege**
   - Only enable what's absolutely necessary
   - Start restrictive, loosen carefully

2. **Regular Security Audits**
   - Review whitelist regularly
   - Monitor violation logs
   - Update based on new threats

3. **Layer Security**
   - Combine multiple techniques
   - Don't rely on single defense
   - Monitor at multiple levels

4. **Clear Documentation**
   - Document what's allowed/blocked
   - Provide safe coding guidelines
   - Warn about limitations

5. **Graceful Degradation**
   - Handle security violations gracefully
   - Provide meaningful error messages
   - Don't expose internal details