# Lua Error Handling Examples

This document provides practical examples of implementing error handling and stack trace preservation in GopherLua.

## Basic Error Handling

### Simple Error Handling

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
    
    // Unprotected call - will panic on error
    err := L.DoString(`
        function riskyOperation()
            error("something went wrong")
        end
        
        riskyOperation()
    `)
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        
        // Check if it's an API error with stack trace
        if apiErr, ok := err.(*lua.ApiError); ok {
            fmt.Printf("Stack trace:\n%s\n", apiErr.StackTrace)
        }
    }
}
```

### Protected Function Calls

```go
package protected

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
)

func CallProtected(L *lua.LState, fnName string, args ...lua.LValue) ([]lua.LValue, error) {
    // Get function
    L.GetGlobal(fnName)
    fn := L.Get(-1)
    L.Pop(1)
    
    if fn == lua.LNil {
        return nil, fmt.Errorf("function %s not found", fnName)
    }
    
    // Protected call
    err := L.CallByParam(lua.P{
        Fn:      fn,
        NRet:    lua.MultRet,
        Protect: true,
    }, args...)
    
    if err != nil {
        return nil, err
    }
    
    // Collect results
    nret := L.GetTop()
    results := make([]lua.LValue, nret)
    for i := 0; i < nret; i++ {
        results[i] = L.Get(i + 1)
    }
    L.SetTop(0)
    
    return results, nil
}

// Example usage
func ExampleProtectedCall() {
    L := lua.NewState()
    defer L.Close()
    
    // Define a function that might error
    L.DoString(`
        function divide(a, b)
            if b == 0 then
                error("division by zero")
            end
            return a / b
        end
    `)
    
    // Safe call
    results, err := CallProtected(L, "divide", lua.LNumber(10), lua.LNumber(2))
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    
    fmt.Printf("Result: %v\n", results[0])
    
    // Call that will error
    _, err = CallProtected(L, "divide", lua.LNumber(10), lua.LNumber(0))
    if err != nil {
        fmt.Printf("Expected error: %v\n", err)
    }
}
```

## Enhanced Error Types

### Custom Error Objects

```go
package customerrors

import (
    "fmt"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type ScriptError struct {
    Type       string
    Message    string
    Code       int
    Details    map[string]interface{}
    StackTrace string
    Timestamp  time.Time
}

func RegisterErrorType(L *lua.LState) {
    mt := L.NewTypeMetatable("ScriptError")
    L.SetGlobal("ScriptError", mt)
    
    // Constructor
    L.SetField(mt, "new", L.NewFunction(newScriptError))
    
    // Methods
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "type":       errorType,
        "message":    errorMessage,
        "code":       errorCode,
        "details":    errorDetails,
        "stackTrace": errorStackTrace,
        "raise":      errorRaise,
    }))
    
    // String representation
    L.SetField(mt, "__tostring", L.NewFunction(errorToString))
}

func newScriptError(L *lua.LState) int {
    errType := L.CheckString(1)
    message := L.CheckString(2)
    code := L.OptInt(3, 0)
    details := L.OptTable(4, L.NewTable())
    
    err := &ScriptError{
        Type:      errType,
        Message:   message,
        Code:      code,
        Details:   tableToMap(details),
        Timestamp: time.Now(),
    }
    
    // Capture stack trace
    if dbg, ok := L.GetStack(1); ok {
        err.StackTrace = L.StackTrace()
    }
    
    ud := L.NewUserData()
    ud.Value = err
    L.SetMetatable(ud, L.GetTypeMetatable("ScriptError"))
    L.Push(ud)
    
    return 1
}

func errorRaise(L *lua.LState) int {
    err := checkScriptError(L, 1)
    L.RaiseError("%s: %s (code: %d)", err.Type, err.Message, err.Code)
    return 0
}

// Usage example in Lua
func ExampleCustomErrors() {
    L := lua.NewState()
    defer L.Close()
    
    RegisterErrorType(L)
    
    err := L.DoString(`
        -- Create custom error types
        function ValidationError(field, value)
            return ScriptError.new("ValidationError", 
                "Invalid value for field: " .. field,
                400,
                {field = field, value = value})
        end
        
        function NetworkError(url, status)
            return ScriptError.new("NetworkError",
                "Failed to connect to: " .. url,
                status,
                {url = url, timestamp = os.time()})
        end
        
        -- Usage
        function validateAge(age)
            if type(age) ~= "number" then
                ValidationError("age", tostring(age)):raise()
            end
            if age < 0 or age > 150 then
                ValidationError("age", age):raise()
            end
            return true
        end
        
        -- Test
        validateAge(-5)
    `)
    
    if err != nil {
        fmt.Printf("Caught error: %v\n", err)
    }
}
```

## Stack Trace Enhancement

### Rich Stack Traces

```go
package stacktrace

import (
    "fmt"
    "strings"
    lua "github.com/yuin/gopher-lua"
)

type EnhancedStackTrace struct {
    Frames []StackFrame
}

type StackFrame struct {
    Source      string
    Line        int
    Function    string
    IsGo        bool
    Locals      map[string]string
    Upvalues    map[string]string
}

func CaptureEnhancedStackTrace(L *lua.LState, maxDepth int) *EnhancedStackTrace {
    trace := &EnhancedStackTrace{
        Frames: make([]StackFrame, 0, maxDepth),
    }
    
    for level := 0; level < maxDepth; level++ {
        dbg, ok := L.GetStack(level)
        if !ok {
            break
        }
        
        L.GetInfo("nSlu", dbg, lua.LNil)
        
        frame := StackFrame{
            Source:   dbg.Source,
            Line:     dbg.CurrentLine,
            Function: dbg.Name,
            IsGo:     dbg.What == "Go",
            Locals:   make(map[string]string),
            Upvalues: make(map[string]string),
        }
        
        // Capture locals (only for Lua frames)
        if !frame.IsGo {
            for i := 1; ; i++ {
                name, value := L.GetLocal(dbg, i)
                if name == "" {
                    break
                }
                frame.Locals[name] = valueToString(value)
            }
            
            // Capture upvalues
            if level == 0 { // Only for current function
                fn := L.Get(-(level + 1))
                if fn.Type() == lua.LTFunction {
                    for i := 1; ; i++ {
                        name, value := L.GetUpvalue(fn, i)
                        if name == "" {
                            break
                        }
                        frame.Upvalues[name] = valueToString(value)
                    }
                }
            }
        }
        
        trace.Frames = append(trace.Frames, frame)
    }
    
    return trace
}

func (est *EnhancedStackTrace) Format(verbose bool) string {
    var buf strings.Builder
    
    for i, frame := range est.Frames {
        // Frame header
        if frame.Function != "" {
            buf.WriteString(fmt.Sprintf("#%d %s at %s:%d\n", 
                i, frame.Function, frame.Source, frame.Line))
        } else {
            buf.WriteString(fmt.Sprintf("#%d <anonymous> at %s:%d\n", 
                i, frame.Source, frame.Line))
        }
        
        if verbose && !frame.IsGo {
            // Local variables
            if len(frame.Locals) > 0 {
                buf.WriteString("    Locals:\n")
                for name, value := range frame.Locals {
                    buf.WriteString(fmt.Sprintf("      %s = %s\n", name, value))
                }
            }
            
            // Upvalues
            if len(frame.Upvalues) > 0 {
                buf.WriteString("    Upvalues:\n")
                for name, value := range frame.Upvalues {
                    buf.WriteString(fmt.Sprintf("      %s = %s\n", name, value))
                }
            }
        }
    }
    
    return buf.String()
}

// Example with error handler
func ExecuteWithEnhancedErrors(L *lua.LState, script string) error {
    // Set error handler
    L.Push(L.NewFunction(func(L *lua.LState) int {
        // Get error message
        errMsg := L.CheckString(1)
        
        // Capture enhanced stack trace
        trace := CaptureEnhancedStackTrace(L, 20)
        
        // Create enhanced error
        enhancedErr := fmt.Sprintf(
            "Error: %s\n\nStack trace:\n%s",
            errMsg,
            trace.Format(true),
        )
        
        L.Push(lua.LString(enhancedErr))
        return 1
    }))
    
    // Compile script
    fn, err := L.LoadString(script)
    if err != nil {
        return err
    }
    
    // Call with error handler
    return L.PCall(0, lua.MultRet, -2)
}
```

## Error Recovery Patterns

### Retry with Backoff

```go
package recovery

import (
    "fmt"
    "math"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type RetryExecutor struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Factor      float64
}

func NewRetryExecutor() *RetryExecutor {
    return &RetryExecutor{
        MaxAttempts: 3,
        BaseDelay:   100 * time.Millisecond,
        MaxDelay:    5 * time.Second,
        Factor:      2.0,
    }
}

func (re *RetryExecutor) ExecuteWithRetry(L *lua.LState, fn lua.LValue, args ...lua.LValue) ([]lua.LValue, error) {
    var lastErr error
    
    for attempt := 0; attempt < re.MaxAttempts; attempt++ {
        if attempt > 0 {
            // Calculate backoff delay
            delay := time.Duration(float64(re.BaseDelay) * math.Pow(re.Factor, float64(attempt-1)))
            if delay > re.MaxDelay {
                delay = re.MaxDelay
            }
            
            fmt.Printf("Retry attempt %d after %v\n", attempt+1, delay)
            time.Sleep(delay)
        }
        
        // Try execution
        err := L.CallByParam(lua.P{
            Fn:      fn,
            NRet:    lua.MultRet,
            Protect: true,
        }, args...)
        
        if err == nil {
            // Success - collect results
            nret := L.GetTop()
            results := make([]lua.LValue, nret)
            for i := 0; i < nret; i++ {
                results[i] = L.Get(i + 1)
            }
            L.SetTop(0)
            return results, nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryable(err) {
            break
        }
    }
    
    return nil, fmt.Errorf("failed after %d attempts: %w", re.MaxAttempts, lastErr)
}

func isRetryable(err error) bool {
    errStr := err.Error()
    
    // Define retryable error patterns
    retryablePatterns := []string{
        "timeout",
        "connection refused",
        "temporary failure",
        "resource temporarily unavailable",
    }
    
    for _, pattern := range retryablePatterns {
        if strings.Contains(errStr, pattern) {
            return true
        }
    }
    
    return false
}

// Example usage
func ExampleRetryExecution() {
    L := lua.NewState()
    defer L.Close()
    
    // Simulate a flaky operation
    L.DoString(`
        attempts = 0
        function flakyOperation()
            attempts = attempts + 1
            if attempts < 3 then
                error("temporary failure")
            end
            return "success", attempts
        end
    `)
    
    executor := NewRetryExecutor()
    
    L.GetGlobal("flakyOperation")
    fn := L.Get(-1)
    L.Pop(1)
    
    results, err := executor.ExecuteWithRetry(L, fn)
    if err != nil {
        fmt.Printf("Final error: %v\n", err)
    } else {
        fmt.Printf("Success after %v attempts\n", results[1])
    }
}
```

## Error Context Preservation

### Contextual Error Information

```go
package context

import (
    "fmt"
    "strings"
    "sync"
    lua "github.com/yuin/gopher-lua"
)

type ErrorContext struct {
    mu        sync.RWMutex
    contexts  []ContextEntry
    maxDepth  int
}

type ContextEntry struct {
    Level       string
    Description string
    Variables   map[string]interface{}
    Timestamp   time.Time
}

func NewErrorContext(maxDepth int) *ErrorContext {
    return &ErrorContext{
        maxDepth: maxDepth,
        contexts: make([]ContextEntry, 0, maxDepth),
    }
}

func (ec *ErrorContext) Push(level, description string, vars map[string]interface{}) {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    
    entry := ContextEntry{
        Level:       level,
        Description: description,
        Variables:   vars,
        Timestamp:   time.Now(),
    }
    
    ec.contexts = append(ec.contexts, entry)
    
    // Maintain max depth
    if len(ec.contexts) > ec.maxDepth {
        ec.contexts = ec.contexts[1:]
    }
}

func (ec *ErrorContext) Pop() {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    
    if len(ec.contexts) > 0 {
        ec.contexts = ec.contexts[:len(ec.contexts)-1]
    }
}

func (ec *ErrorContext) WrapError(err error) error {
    ec.mu.RLock()
    defer ec.mu.RUnlock()
    
    if len(ec.contexts) == 0 {
        return err
    }
    
    var contextStr strings.Builder
    contextStr.WriteString("Error context:\n")
    
    for i := len(ec.contexts) - 1; i >= 0; i-- {
        ctx := ec.contexts[i]
        contextStr.WriteString(fmt.Sprintf("  [%s] %s\n", ctx.Level, ctx.Description))
        
        if len(ctx.Variables) > 0 {
            for k, v := range ctx.Variables {
                contextStr.WriteString(fmt.Sprintf("    %s: %v\n", k, v))
            }
        }
    }
    
    return fmt.Errorf("%s\nOriginal error: %w", contextStr.String(), err)
}

// Integration with Lua
func InstallErrorContext(L *lua.LState, ec *ErrorContext) {
    // Push context function
    L.SetGlobal("pushContext", L.NewFunction(func(L *lua.LState) int {
        level := L.CheckString(1)
        desc := L.CheckString(2)
        vars := L.OptTable(3, L.NewTable())
        
        varsMap := make(map[string]interface{})
        vars.ForEach(func(k, v lua.LValue) {
            varsMap[k.String()] = luaValueToGo(v)
        })
        
        ec.Push(level, desc, varsMap)
        return 0
    }))
    
    // Pop context function
    L.SetGlobal("popContext", L.NewFunction(func(L *lua.LState) int {
        ec.Pop()
        return 0
    }))
    
    // With context function
    L.SetGlobal("withContext", L.NewFunction(func(L *lua.LState) int {
        level := L.CheckString(1)
        desc := L.CheckString(2)
        fn := L.CheckFunction(3)
        
        // Push context
        ec.Push(level, desc, nil)
        defer ec.Pop()
        
        // Call function
        err := L.CallByParam(lua.P{
            Fn:      fn,
            NRet:    lua.MultRet,
            Protect: true,
        })
        
        if err != nil {
            // Wrap error with context
            L.RaiseError(ec.WrapError(err).Error())
        }
        
        return L.GetTop()
    }))
}

// Example usage
func ExampleErrorContext() {
    L := lua.NewState()
    defer L.Close()
    
    ec := NewErrorContext(10)
    InstallErrorContext(L, ec)
    
    err := L.DoString(`
        function processUser(userId)
            return withContext("business", "processing user", function()
                pushContext("data", "loading user data", {userId = userId})
                
                -- Simulate error
                error("user not found")
                
                popContext()
            end)
        end
        
        processUser(12345)
    `)
    
    if err != nil {
        fmt.Printf("Error with context:\n%v\n", err)
    }
}
```

## Error Logging and Monitoring

### Structured Error Logger

```go
package logging

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    lua "github.com/yuin/gopher-lua"
)

type ErrorLogger struct {
    logger     *log.Logger
    structured bool
    logLevel   LogLevel
}

type LogLevel int

const (
    LogError LogLevel = iota
    LogWarning
    LogInfo
    LogDebug
)

type ErrorLogEntry struct {
    Timestamp  time.Time              `json:"timestamp"`
    Level      string                 `json:"level"`
    Error      string                 `json:"error"`
    Script     string                 `json:"script,omitempty"`
    Function   string                 `json:"function,omitempty"`
    Line       int                    `json:"line,omitempty"`
    StackTrace []StackFrameLog        `json:"stack_trace,omitempty"`
    Context    map[string]interface{} `json:"context,omitempty"`
}

type StackFrameLog struct {
    File     string `json:"file"`
    Line     int    `json:"line"`
    Function string `json:"function"`
}

func NewErrorLogger(structured bool) *ErrorLogger {
    return &ErrorLogger{
        logger:     log.New(os.Stderr, "[LUA] ", log.LstdFlags),
        structured: structured,
        logLevel:   LogError,
    }
}

func (el *ErrorLogger) LogError(L *lua.LState, err error) {
    entry := el.createLogEntry(L, err, "error")
    
    if el.structured {
        jsonData, _ := json.Marshal(entry)
        el.logger.Println(string(jsonData))
    } else {
        el.logger.Printf("ERROR: %s at %s:%d - %s",
            entry.Error, entry.Script, entry.Line, entry.Function)
    }
}

func (el *ErrorLogger) createLogEntry(L *lua.LState, err error, level string) ErrorLogEntry {
    entry := ErrorLogEntry{
        Timestamp: time.Now(),
        Level:     level,
        Error:     err.Error(),
        Context:   make(map[string]interface{}),
    }
    
    // Get current position
    if dbg, ok := L.GetStack(0); ok {
        L.GetInfo("nSl", dbg, lua.LNil)
        entry.Script = dbg.Source
        entry.Function = dbg.Name
        entry.Line = dbg.CurrentLine
    }
    
    // Get stack trace
    for i := 0; i < 10; i++ {
        dbg, ok := L.GetStack(i)
        if !ok {
            break
        }
        
        L.GetInfo("nSl", dbg, lua.LNil)
        entry.StackTrace = append(entry.StackTrace, StackFrameLog{
            File:     dbg.Source,
            Line:     dbg.CurrentLine,
            Function: dbg.Name,
        })
    }
    
    return entry
}

// Install in Lua
func InstallErrorLogger(L *lua.LState, logger *ErrorLogger) {
    L.SetGlobal("logError", L.NewFunction(func(L *lua.LState) int {
        msg := L.CheckString(1)
        logger.LogError(L, fmt.Errorf(msg))
        return 0
    }))
}
```

## Complete Error Handling System

### Integrated Error Manager

```go
package integrated

import (
    "fmt"
    "sync"
    lua "github.com/yuin/gopher-lua"
)

type ErrorManager struct {
    logger      *ErrorLogger
    context     *ErrorContext
    recovery    *RetryExecutor
    handlers    map[string]ErrorHandler
    mu          sync.RWMutex
}

type ErrorHandler func(L *lua.LState, err error) (handled bool, result lua.LValue)

func NewErrorManager() *ErrorManager {
    return &ErrorManager{
        logger:   NewErrorLogger(true),
        context:  NewErrorContext(10),
        recovery: NewRetryExecutor(),
        handlers: make(map[string]ErrorHandler),
    }
}

func (em *ErrorManager) RegisterHandler(pattern string, handler ErrorHandler) {
    em.mu.Lock()
    defer em.mu.Unlock()
    em.handlers[pattern] = handler
}

func (em *ErrorManager) ExecuteScript(L *lua.LState, script string) error {
    // Install error management functions
    em.installFunctions(L)
    
    // Set global error handler
    L.Push(L.NewFunction(em.globalErrorHandler))
    
    // Load and execute script
    fn, err := L.LoadString(script)
    if err != nil {
        return fmt.Errorf("compilation error: %w", err)
    }
    
    // Execute with error handling
    return L.PCall(0, 0, -2)
}

func (em *ErrorManager) globalErrorHandler(L *lua.LState) int {
    err := fmt.Errorf("%v", L.Get(1))
    
    // Log error
    em.logger.LogError(L, err)
    
    // Try handlers
    em.mu.RLock()
    for pattern, handler := range em.handlers {
        if strings.Contains(err.Error(), pattern) {
            em.mu.RUnlock()
            
            if handled, result := handler(L, err); handled {
                if result != lua.LNil {
                    L.Push(result)
                    return 1
                }
                return 0
            }
            
            em.mu.RLock()
        }
    }
    em.mu.RUnlock()
    
    // Wrap with context
    contextErr := em.context.WrapError(err)
    L.Push(lua.LString(contextErr.Error()))
    return 1
}

func (em *ErrorManager) installFunctions(L *lua.LState) {
    // Context functions
    InstallErrorContext(L, em.context)
    
    // Logger functions
    InstallErrorLogger(L, em.logger)
    
    // Safe execution function
    L.SetGlobal("safeCall", L.NewFunction(func(L *lua.LState) int {
        fn := L.CheckFunction(1)
        
        // Remove function from stack
        args := make([]lua.LValue, L.GetTop()-1)
        for i := 0; i < len(args); i++ {
            args[i] = L.Get(i + 2)
        }
        L.SetTop(0)
        
        // Execute with retry
        results, err := em.recovery.ExecuteWithRetry(L, fn, args...)
        
        if err != nil {
            L.Push(lua.LBool(false))
            L.Push(lua.LString(err.Error()))
            return 2
        }
        
        L.Push(lua.LBool(true))
        for _, result := range results {
            L.Push(result)
        }
        return len(results) + 1
    }))
}

// Example usage
func ExampleIntegratedErrorHandling() {
    L := lua.NewState()
    defer L.Close()
    
    em := NewErrorManager()
    
    // Register custom handlers
    em.RegisterHandler("network", func(L *lua.LState, err error) (bool, lua.LValue) {
        fmt.Println("Network error detected, returning cached data")
        return true, lua.LString("cached_response")
    })
    
    err := em.ExecuteScript(L, `
        function fetchData(url)
            return withContext("api", "fetching data from " .. url, function()
                -- Simulate network error
                error("network timeout")
            end)
        end
        
        -- Safe execution with retry
        local ok, result = safeCall(fetchData, "https://api.example.com")
        if ok then
            print("Success:", result)
        else
            print("Failed:", result)
        end
    `)
    
    if err != nil {
        fmt.Printf("Script execution failed: %v\n", err)
    }
}
```

## Summary

These examples demonstrate comprehensive error handling in GopherLua:
1. Basic error capture and protected calls
2. Custom error types with rich metadata
3. Enhanced stack traces with local variables
4. Retry mechanisms with exponential backoff
5. Context preservation for better debugging
6. Structured logging and monitoring
7. Integrated error management system

Key patterns include protected execution, contextual information, recovery strategies, and comprehensive error tracking.