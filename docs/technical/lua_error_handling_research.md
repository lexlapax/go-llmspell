# Lua Error Handling and Stack Trace Preservation Research

This document investigates error handling mechanisms and stack trace preservation strategies in GopherLua for robust script execution and debugging.

## Executive Summary

GopherLua provides comprehensive error handling through protected calls, error objects, and stack trace information. Effective error handling requires careful design of error propagation, stack trace capture, and error recovery mechanisms.

## Error Handling in GopherLua

### Error Types

GopherLua supports multiple error types:

```go
// Runtime errors
L.RaiseError("custom error: %s", message)

// API errors
err := L.DoString(script) // Returns Go error

// Protected calls
err := L.CallByParam(lua.P{
    Fn:      function,
    NRet:    1,
    Protect: true, // Protected mode
}, args...)

// Panic/recover mechanism
L.SetPCall(true) // Enable protected calls globally
```

### Error Object Structure

```go
// Lua error with stack trace
type ApiError struct {
    Type       ApiErrorType
    Message    string
    StackTrace string
    Cause      error
}

// Stack frame information
type DebugInfo struct {
    Source        string // Source file
    CurrentLine   int    // Current line number
    LineDefined   int    // Function definition line
    LastLineDefined int  // Function end line
    What          string // "Lua", "Go", "main", "tail"
    Name          string // Function name
}
```

## Stack Trace Preservation

### 1. Capturing Stack Traces

```go
type StackTraceCapture struct {
    maxDepth int
    skipGo   bool
}

func (stc *StackTraceCapture) CaptureStackTrace(L *lua.LState) []StackFrame {
    frames := make([]StackFrame, 0, stc.maxDepth)
    
    for i := 0; i < stc.maxDepth; i++ {
        dbg, ok := L.GetStack(i)
        if !ok {
            break
        }
        
        L.GetInfo("nSlu", dbg, lua.LNil)
        
        // Skip Go frames if requested
        if stc.skipGo && dbg.What == "Go" {
            continue
        }
        
        frame := StackFrame{
            Source:      dbg.Source,
            Line:        dbg.CurrentLine,
            Function:    dbg.Name,
            What:        dbg.What,
            LinesDefined: [2]int{dbg.LineDefined, dbg.LastLineDefined},
        }
        
        // Get local variables
        if i == 0 { // Only for current frame
            frame.Locals = stc.captureLocals(L, dbg)
        }
        
        frames = append(frames, frame)
    }
    
    return frames
}

func (stc *StackTraceCapture) captureLocals(L *lua.LState, dbg *lua.Debug) map[string]string {
    locals := make(map[string]string)
    
    for j := 1; ; j++ {
        name, value := L.GetLocal(dbg, j)
        if name == "" {
            break
        }
        
        locals[name] = stc.valueToString(value)
    }
    
    return locals
}
```

### 2. Error Context Enhancement

```go
type ErrorContext struct {
    Script     string
    Function   string
    Line       int
    Column     int
    StackTrace []StackFrame
    Variables  map[string]interface{}
    Timestamp  time.Time
}

func EnhanceError(L *lua.LState, originalErr error) *EnhancedError {
    ctx := ErrorContext{
        Timestamp: time.Now(),
    }
    
    // Get current position
    if dbg, ok := L.GetStack(0); ok {
        L.GetInfo("nSl", dbg, lua.LNil)
        ctx.Script = dbg.Source
        ctx.Function = dbg.Name
        ctx.Line = dbg.CurrentLine
    }
    
    // Capture stack trace
    capture := &StackTraceCapture{maxDepth: 20}
    ctx.StackTrace = capture.CaptureStackTrace(L)
    
    // Capture global state snapshot
    ctx.Variables = captureRelevantVariables(L)
    
    return &EnhancedError{
        Original: originalErr,
        Context:  ctx,
    }
}
```

### 3. Custom Error Types

```go
type ScriptError struct {
    Type       string
    Message    string
    StackTrace string
    Source     string
    Line       int
    UserData   lua.LValue
}

func CreateScriptError(L *lua.LState, errType, message string) *lua.LUserData {
    err := &ScriptError{
        Type:    errType,
        Message: message,
    }
    
    // Capture stack trace
    if dbg, ok := L.GetStack(1); ok {
        L.GetInfo("Sl", dbg, lua.LNil)
        err.Source = dbg.Source
        err.Line = dbg.CurrentLine
    }
    
    err.StackTrace = L.StackTrace()
    
    // Create userdata
    ud := L.NewUserData()
    ud.Value = err
    L.SetMetatable(ud, L.GetTypeMetatable("ScriptError"))
    
    return ud
}

// Register error type metatable
func RegisterScriptErrorType(L *lua.LState) {
    mt := L.NewTypeMetatable("ScriptError")
    
    L.SetField(mt, "__tostring", L.NewFunction(func(L *lua.LState) int {
        err := checkScriptError(L, 1)
        L.Push(lua.LString(fmt.Sprintf("%s: %s\n%s", 
            err.Type, err.Message, err.StackTrace)))
        return 1
    }))
    
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "type":       scriptErrorType,
        "message":    scriptErrorMessage,
        "stackTrace": scriptErrorStackTrace,
        "source":     scriptErrorSource,
        "line":       scriptErrorLine,
    }))
}
```

## Error Propagation Strategies

### 1. Error Chain Management

```go
type ErrorChain struct {
    errors []ChainedError
    mu     sync.RWMutex
}

type ChainedError struct {
    Error      error
    Context    ErrorContext
    Timestamp  time.Time
    ScriptName string
}

func (ec *ErrorChain) AddError(L *lua.LState, err error, scriptName string) {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    
    chained := ChainedError{
        Error:      err,
        Timestamp:  time.Now(),
        ScriptName: scriptName,
        Context:    captureContext(L),
    }
    
    ec.errors = append(ec.errors, chained)
}

func (ec *ErrorChain) Format() string {
    ec.mu.RLock()
    defer ec.mu.RUnlock()
    
    var buf strings.Builder
    
    for i, e := range ec.errors {
        buf.WriteString(fmt.Sprintf("Error #%d in %s at %s:\n", 
            i+1, e.ScriptName, e.Timestamp.Format(time.RFC3339)))
        buf.WriteString(fmt.Sprintf("  %v\n", e.Error))
        
        if len(e.Context.StackTrace) > 0 {
            buf.WriteString("  Stack trace:\n")
            for _, frame := range e.Context.StackTrace {
                buf.WriteString(fmt.Sprintf("    %s:%d in %s\n", 
                    frame.Source, frame.Line, frame.Function))
            }
        }
        buf.WriteString("\n")
    }
    
    return buf.String()
}
```

### 2. Error Recovery Mechanisms

```go
type ErrorRecovery struct {
    strategies map[string]RecoveryStrategy
    fallback   RecoveryStrategy
}

type RecoveryStrategy func(L *lua.LState, err error) (recovered bool, result lua.LValue)

func NewErrorRecovery() *ErrorRecovery {
    er := &ErrorRecovery{
        strategies: make(map[string]RecoveryStrategy),
    }
    
    // Default strategies
    er.RegisterStrategy("timeout", func(L *lua.LState, err error) (bool, lua.LValue) {
        if strings.Contains(err.Error(), "timeout") {
            // Return partial results if available
            L.GetGlobal("_partial_results")
            if !L.IsNil(-1) {
                return true, L.Get(-1)
            }
            L.Pop(1)
        }
        return false, lua.LNil
    })
    
    er.RegisterStrategy("memory", func(L *lua.LState, err error) (bool, lua.LValue) {
        if strings.Contains(err.Error(), "memory") {
            // Trigger GC and retry once
            L.DoString(`collectgarbage("collect")`)
            return true, lua.LNil // Signal retry
        }
        return false, lua.LNil
    })
    
    return er
}

func (er *ErrorRecovery) Recover(L *lua.LState, err error) (bool, lua.LValue) {
    // Try specific strategies
    for pattern, strategy := range er.strategies {
        if strings.Contains(err.Error(), pattern) {
            return strategy(L, err)
        }
    }
    
    // Fallback strategy
    if er.fallback != nil {
        return er.fallback(L, err)
    }
    
    return false, lua.LNil
}
```

### 3. Error Handling Wrapper

```go
type SafeExecutor struct {
    recovery      *ErrorRecovery
    errorHandler  func(error)
    maxRetries    int
    retryDelay    time.Duration
}

func (se *SafeExecutor) ExecuteProtected(L *lua.LState, fn lua.LValue, args ...lua.LValue) ([]lua.LValue, error) {
    var lastErr error
    
    for attempt := 0; attempt <= se.maxRetries; attempt++ {
        if attempt > 0 {
            time.Sleep(se.retryDelay)
        }
        
        // Protected call
        err := L.CallByParam(lua.P{
            Fn:      fn,
            NRet:    lua.MultRet,
            Protect: true,
            Handler: L.NewFunction(se.errorHandler),
        }, args...)
        
        if err == nil {
            // Success - collect results
            results := make([]lua.LValue, L.GetTop())
            for i := range results {
                results[i] = L.Get(i + 1)
            }
            L.SetTop(0)
            return results, nil
        }
        
        lastErr = err
        
        // Try recovery
        if recovered, result := se.recovery.Recover(L, err); recovered {
            if result != lua.LNil {
                return []lua.LValue{result}, nil
            }
            // Continue to retry
            continue
        }
        
        // No recovery possible
        break
    }
    
    return nil, lastErr
}

func (se *SafeExecutor) errorHandler(L *lua.LState) int {
    err := L.Get(1)
    
    // Enhance error with context
    enhanced := EnhanceError(L, fmt.Errorf("%v", err))
    
    // Log error
    if se.errorHandler != nil {
        se.errorHandler(enhanced)
    }
    
    // Re-raise with enhanced information
    L.Push(lua.LString(enhanced.String()))
    return 1
}
```

## Debug Information Management

### 1. Debug Context Preservation

```go
type DebugContext struct {
    breakpoints  map[string][]int // file -> line numbers
    watchedVars  []string
    stepMode     StepMode
    callDepth    int
    debugHandler DebugHandler
}

type DebugHandler func(L *lua.LState, event DebugEvent)

type DebugEvent struct {
    Type      string // "call", "return", "line"
    File      string
    Line      int
    Function  string
    Depth     int
    Variables map[string]lua.LValue
}

func (dc *DebugContext) InstallHooks(L *lua.LState) {
    L.SetHook(func(L *lua.LState) {
        dbg, _ := L.GetStack(0)
        L.GetInfo("nSl", dbg, lua.LNil)
        
        event := DebugEvent{
            File:     dbg.Source,
            Line:     dbg.CurrentLine,
            Function: dbg.Name,
            Depth:    dc.callDepth,
        }
        
        // Check breakpoints
        if dc.isBreakpoint(dbg.Source, dbg.CurrentLine) {
            event.Type = "breakpoint"
            event.Variables = dc.captureWatchedVariables(L, dbg)
            dc.debugHandler(L, event)
        }
        
        // Step mode handling
        switch dc.stepMode {
        case StepInto:
            event.Type = "step"
            dc.debugHandler(L, event)
        case StepOver:
            if dc.callDepth <= dc.stepOverDepth {
                event.Type = "step"
                dc.debugHandler(L, event)
            }
        }
    }, lua.MaskLine|lua.MaskCall|lua.MaskRet, 0)
}
```

### 2. Error Reporting Formatter

```go
type ErrorFormatter struct {
    style         FormatStyle
    maxFrames     int
    contextLines  int
    colorEnabled  bool
}

type FormatStyle int

const (
    FormatSimple FormatStyle = iota
    FormatDetailed
    FormatJSON
)

func (ef *ErrorFormatter) Format(err *EnhancedError) string {
    switch ef.style {
    case FormatSimple:
        return ef.formatSimple(err)
    case FormatDetailed:
        return ef.formatDetailed(err)
    case FormatJSON:
        return ef.formatJSON(err)
    default:
        return err.Error()
    }
}

func (ef *ErrorFormatter) formatDetailed(err *EnhancedError) string {
    var buf strings.Builder
    
    // Error header
    if ef.colorEnabled {
        buf.WriteString(color.RedString("Error: %s\n", err.Message))
    } else {
        buf.WriteString(fmt.Sprintf("Error: %s\n", err.Message))
    }
    
    // Location
    buf.WriteString(fmt.Sprintf("  at %s:%d", err.Context.Script, err.Context.Line))
    if err.Context.Function != "" {
        buf.WriteString(fmt.Sprintf(" in function '%s'", err.Context.Function))
    }
    buf.WriteString("\n\n")
    
    // Source context
    if err.Context.Script != "@string" {
        if sourceContext := ef.getSourceContext(err.Context.Script, err.Context.Line); sourceContext != "" {
            buf.WriteString("Source:\n")
            buf.WriteString(sourceContext)
            buf.WriteString("\n")
        }
    }
    
    // Stack trace
    buf.WriteString("Stack trace:\n")
    for i, frame := range err.Context.StackTrace {
        if i >= ef.maxFrames {
            buf.WriteString(fmt.Sprintf("  ... %d more frames\n", 
                len(err.Context.StackTrace)-ef.maxFrames))
            break
        }
        
        buf.WriteString(ef.formatFrame(frame, i))
    }
    
    // Variables (if in debug mode)
    if len(err.Context.Variables) > 0 {
        buf.WriteString("\nLocal variables:\n")
        for name, value := range err.Context.Variables {
            buf.WriteString(fmt.Sprintf("  %s = %v\n", name, value))
        }
    }
    
    return buf.String()
}

func (ef *ErrorFormatter) getSourceContext(file string, line int) string {
    // Read source file and extract context
    content, err := ioutil.ReadFile(file)
    if err != nil {
        return ""
    }
    
    lines := strings.Split(string(content), "\n")
    start := max(0, line-ef.contextLines-1)
    end := min(len(lines), line+ef.contextLines)
    
    var buf strings.Builder
    for i := start; i < end; i++ {
        lineNum := i + 1
        prefix := "  "
        if lineNum == line {
            prefix = "> "
            if ef.colorEnabled {
                buf.WriteString(color.RedString("%s%4d | %s\n", prefix, lineNum, lines[i]))
            } else {
                buf.WriteString(fmt.Sprintf("%s%4d | %s\n", prefix, lineNum, lines[i]))
            }
        } else {
            buf.WriteString(fmt.Sprintf("%s%4d | %s\n", prefix, lineNum, lines[i]))
        }
    }
    
    return buf.String()
}
```

## Error Testing Infrastructure

### 1. Error Injection

```go
type ErrorInjector struct {
    injections map[string]ErrorInjection
    enabled    bool
}

type ErrorInjection struct {
    Condition func(L *lua.LState) bool
    Error     func() error
    Frequency float64 // 0.0 to 1.0
}

func (ei *ErrorInjector) InjectErrors(L *lua.LState) {
    if !ei.enabled {
        return
    }
    
    L.SetHook(func(L *lua.LState) {
        for name, injection := range ei.injections {
            if injection.Condition(L) && rand.Float64() < injection.Frequency {
                L.RaiseError("injected error: %s", injection.Error())
            }
        }
    }, lua.MaskCall, 0)
}

// Example usage
injector := &ErrorInjector{
    enabled: true,
    injections: map[string]ErrorInjection{
        "memory_pressure": {
            Condition: func(L *lua.LState) bool {
                return L.MemUsage() > 5*1024*1024 // 5MB
            },
            Error: func() error {
                return errors.New("simulated memory error")
            },
            Frequency: 0.1, // 10% chance
        },
    },
}
```

### 2. Error Verification

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name          string
        script        string
        expectedError string
        checkStack    bool
    }{
        {
            name: "runtime error",
            script: `
                function divide(a, b)
                    return a / b
                end
                divide(10, 0)
            `,
            expectedError: "divide by zero",
            checkStack:    true,
        },
        {
            name: "custom error",
            script: `
                error("custom error message", 2)
            `,
            expectedError: "custom error message",
            checkStack:    true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            L := lua.NewState()
            defer L.Close()
            
            err := L.DoString(tt.script)
            
            if err == nil {
                t.Fatal("expected error but got none")
            }
            
            if !strings.Contains(err.Error(), tt.expectedError) {
                t.Errorf("expected error containing %q, got %v", 
                    tt.expectedError, err)
            }
            
            if tt.checkStack {
                // Verify stack trace is preserved
                if apiErr, ok := err.(*lua.ApiError); ok {
                    if apiErr.StackTrace == "" {
                        t.Error("expected stack trace but got none")
                    }
                }
            }
        })
    }
}
```

## Best Practices

1. **Always Use Protected Calls**: Use `CallByParam` with `Protect: true`
2. **Capture Stack Traces Early**: Stack information is transient
3. **Provide Context**: Include relevant variables and state in errors
4. **Design for Recovery**: Implement appropriate recovery strategies
5. **Test Error Paths**: Use error injection to test error handling

## Implementation Checklist

- [ ] Basic error capture and enhancement
- [ ] Stack trace preservation system
- [ ] Custom error types with metadata
- [ ] Error chain management
- [ ] Recovery strategy framework
- [ ] Debug context preservation
- [ ] Error formatting and reporting
- [ ] Source code context display
- [ ] Error injection for testing
- [ ] Comprehensive error handling tests

## Summary

Effective error handling in GopherLua requires:
1. Comprehensive error capture with stack traces
2. Context preservation for debugging
3. Flexible recovery mechanisms
4. Clear error reporting
5. Testing infrastructure for error paths

This enables robust script execution with excellent debugging capabilities.