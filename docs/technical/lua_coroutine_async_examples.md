# Lua Coroutine Async Examples

This document provides practical examples of implementing asynchronous bridge operations using Lua coroutines in GopherLua.

## Basic Coroutine Usage

### Simple Coroutine Example

```go
package main

import (
    "fmt"
    "time"
    lua "github.com/yuin/gopher-lua"
)

func main() {
    L := lua.NewState()
    defer L.Close()
    
    // Create a coroutine
    err := L.DoString(`
        -- Create coroutine
        local co = coroutine.create(function(name)
            print("Coroutine started with:", name)
            
            local value = coroutine.yield("first yield")
            print("Resumed with:", value)
            
            coroutine.yield("second yield")
            
            return "done", name
        end)
        
        -- Resume coroutine
        local ok, result = coroutine.resume(co, "test")
        print("First resume:", ok, result)
        
        ok, result = coroutine.resume(co, "hello")
        print("Second resume:", ok, result)
        
        ok, result1, result2 = coroutine.resume(co)
        print("Final resume:", ok, result1, result2)
        
        print("Status:", coroutine.status(co))
    `)
    
    if err != nil {
        panic(err)
    }
}
```

### Coroutine with Go Integration

```go
package gointegration

import (
    "fmt"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type CoroutineManager struct {
    L          *lua.LState
    coroutines map[string]*lua.LState
}

func NewCoroutineManager(L *lua.LState) *CoroutineManager {
    cm := &CoroutineManager{
        L:          L,
        coroutines: make(map[string]*lua.LState),
    }
    
    // Register coroutine functions
    L.SetGlobal("spawn", L.NewFunction(cm.spawn))
    L.SetGlobal("sleep", L.NewFunction(cm.sleep))
    
    return cm
}

func (cm *CoroutineManager) spawn(L *lua.LState) int {
    fn := L.CheckFunction(1)
    
    // Create new coroutine
    co := L.NewThread()
    
    // Copy arguments
    nargs := L.GetTop() - 1
    for i := 2; i <= nargs+1; i++ {
        co.Push(L.Get(i))
    }
    
    // Start coroutine in goroutine
    go func() {
        co.Push(fn)
        for i := 0; i < nargs; i++ {
            co.Push(co.Get(i + 1))
        }
        
        state, err, values := co.Resume(fn, co.Get(1))
        if state == lua.ResumeError {
            fmt.Printf("Coroutine error: %v\n", err)
        }
    }()
    
    // Return coroutine reference
    L.Push(co)
    return 1
}

func (cm *CoroutineManager) sleep(L *lua.LState) int {
    duration := L.CheckNumber(1)
    
    // Only yield if in coroutine
    if L.Status(L) != lua.ThreadNormal {
        // Schedule resumption
        go func() {
            time.Sleep(time.Duration(duration) * time.Second)
            L.Resume(L)
        }()
        
        return L.Yield(0)
    }
    
    // Not in coroutine, block
    time.Sleep(time.Duration(duration) * time.Second)
    return 0
}

// Example usage
func ExampleCoroutineManager() {
    L := lua.NewState()
    defer L.Close()
    
    cm := NewCoroutineManager(L)
    
    err := L.DoString(`
        -- Spawn async tasks
        local task1 = spawn(function(name)
            print(name .. " started")
            sleep(1)
            print(name .. " middle")
            sleep(1)
            print(name .. " done")
        end, "Task1")
        
        local task2 = spawn(function(name)
            for i = 1, 3 do
                print(name .. " count:", i)
                sleep(0.5)
            end
        end, "Task2")
        
        -- Main continues
        print("Main continues...")
        sleep(3)
        print("Main done")
    `)
    
    if err != nil {
        panic(err)
    }
}
```

## Promise Implementation

### Basic Promise Pattern

```go
package promise

import (
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type Promise struct {
    state     PromiseState
    value     lua.LValue
    err       error
    callbacks []Callback
    mu        sync.Mutex
}

type PromiseState int

const (
    Pending PromiseState = iota
    Fulfilled
    Rejected
)

type Callback struct {
    OnFulfilled lua.LValue
    OnRejected  lua.LValue
    Promise     *Promise // For chaining
}

func RegisterPromiseType(L *lua.LState) {
    mt := L.NewTypeMetatable("Promise")
    L.SetGlobal("Promise", mt)
    
    // Static methods
    L.SetField(mt, "new", L.NewFunction(promiseNew))
    L.SetField(mt, "resolve", L.NewFunction(promiseResolve))
    L.SetField(mt, "reject", L.NewFunction(promiseReject))
    L.SetField(mt, "all", L.NewFunction(promiseAll))
    L.SetField(mt, "race", L.NewFunction(promiseRace))
    
    // Instance methods
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "then":    promiseThen,
        "catch":   promiseCatch,
        "finally": promiseFinally,
        "await":   promiseAwait,
    }))
}

func promiseNew(L *lua.LState) int {
    executor := L.CheckFunction(1)
    
    promise := &Promise{
        state: Pending,
    }
    
    ud := L.NewUserData()
    ud.Value = promise
    L.SetMetatable(ud, L.GetTypeMetatable("Promise"))
    
    // Create resolve/reject functions
    resolve := L.NewClosure(func(L *lua.LState) int {
        value := L.Get(1)
        promise.Resolve(value)
        return 0
    }, ud)
    
    reject := L.NewClosure(func(L *lua.LState) int {
        err := L.Get(1)
        promise.Reject(err)
        return 0
    }, ud)
    
    // Call executor
    L.Push(executor)
    L.Push(resolve)
    L.Push(reject)
    L.PCall(2, 0, nil)
    
    L.Push(ud)
    return 1
}

func (p *Promise) Resolve(value lua.LValue) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if p.state != Pending {
        return
    }
    
    p.state = Fulfilled
    p.value = value
    
    // Execute callbacks
    for _, cb := range p.callbacks {
        if cb.OnFulfilled != lua.LNil {
            go p.executeCallback(cb.OnFulfilled, value, cb.Promise)
        }
    }
}

func (p *Promise) Reject(err lua.LValue) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if p.state != Pending {
        return
    }
    
    p.state = Rejected
    p.err = fmt.Errorf("%v", err)
    
    // Execute callbacks
    for _, cb := range p.callbacks {
        if cb.OnRejected != lua.LNil {
            go p.executeCallback(cb.OnRejected, err, cb.Promise)
        }
    }
}

func promiseThen(L *lua.LState) int {
    promise := checkPromise(L, 1)
    onFulfilled := L.OptFunction(2, nil)
    onRejected := L.OptFunction(3, nil)
    
    // Create chained promise
    newPromise := &Promise{
        state: Pending,
    }
    
    callback := Callback{
        OnFulfilled: onFulfilled,
        OnRejected:  onRejected,
        Promise:     newPromise,
    }
    
    promise.mu.Lock()
    if promise.state == Pending {
        promise.callbacks = append(promise.callbacks, callback)
    } else if promise.state == Fulfilled {
        go promise.executeCallback(onFulfilled, promise.value, newPromise)
    } else {
        go promise.executeCallback(onRejected, lua.LString(promise.err.Error()), newPromise)
    }
    promise.mu.Unlock()
    
    // Return chained promise
    ud := L.NewUserData()
    ud.Value = newPromise
    L.SetMetatable(ud, L.GetTypeMetatable("Promise"))
    L.Push(ud)
    
    return 1
}

// Example usage
func ExamplePromises() {
    L := lua.NewState()
    defer L.Close()
    
    RegisterPromiseType(L)
    
    // Async function that returns promise
    L.SetGlobal("fetchData", L.NewFunction(func(L *lua.LState) int {
        url := L.CheckString(1)
        
        promise := &Promise{state: Pending}
        
        go func() {
            // Simulate async operation
            time.Sleep(1 * time.Second)
            
            if url == "error" {
                promise.Reject(lua.LString("Network error"))
            } else {
                promise.Resolve(lua.LString("Data from " + url))
            }
        }()
        
        ud := L.NewUserData()
        ud.Value = promise
        L.SetMetatable(ud, L.GetTypeMetatable("Promise"))
        L.Push(ud)
        return 1
    }))
    
    err := L.DoString(`
        -- Promise chaining
        fetchData("https://api.example.com")
            :then(function(data)
                print("Received:", data)
                return fetchData("https://api.example.com/2")
            end)
            :then(function(data2)
                print("Second data:", data2)
            end)
            :catch(function(err)
                print("Error:", err)
            end)
        
        -- Promise.all
        Promise.all({
            fetchData("url1"),
            fetchData("url2"),
            fetchData("url3")
        }):then(function(results)
            print("All results:", #results)
        end)
    `)
    
    if err != nil {
        panic(err)
    }
    
    time.Sleep(3 * time.Second) // Wait for async operations
}
```

## Async/Await Implementation

### Full Async/Await Pattern

```go
package asyncawait

import (
    "fmt"
    "sync"
    lua "github.com/yuin/gopher-lua"
)

type AsyncRuntime struct {
    L              *lua.LState
    pendingAwait   map[*lua.LState]chan lua.LValue
    mu             sync.RWMutex
}

func NewAsyncRuntime(L *lua.LState) *AsyncRuntime {
    ar := &AsyncRuntime{
        L:            L,
        pendingAwait: make(map[*lua.LState]chan lua.LValue),
    }
    
    // Install async/await
    ar.install()
    
    return ar
}

func (ar *AsyncRuntime) install() {
    // async function wrapper
    ar.L.SetGlobal("async", ar.L.NewFunction(func(L *lua.LState) int {
        fn := L.CheckFunction(1)
        
        // Create async wrapper
        wrapper := L.NewFunction(func(L *lua.LState) int {
            // Create coroutine for async function
            co := L.NewThread()
            
            // Copy arguments
            nargs := L.GetTop()
            for i := 1; i <= nargs; i++ {
                co.Push(L.Get(i))
            }
            
            // Create promise for result
            promise := &Promise{state: Pending}
            
            // Start coroutine
            go func() {
                co.Push(fn)
                for i := 1; i <= nargs; i++ {
                    co.Push(co.Get(i))
                }
                
                state, err, values := co.Resume(fn)
                
                if state == lua.ResumeError {
                    promise.Reject(lua.LString(err.Error()))
                } else {
                    if len(values) > 0 {
                        promise.Resolve(values[0])
                    } else {
                        promise.Resolve(lua.LNil)
                    }
                }
            }()
            
            // Return promise
            ud := L.NewUserData()
            ud.Value = promise
            L.SetMetatable(ud, L.GetTypeMetatable("Promise"))
            L.Push(ud)
            return 1
        })
        
        L.Push(wrapper)
        return 1
    }))
    
    // await function
    ar.L.SetGlobal("await", ar.L.NewFunction(func(L *lua.LState) int {
        // Must be called from coroutine
        if L.Status(L) == lua.ThreadNormal {
            L.RaiseError("await must be called from async function")
            return 0
        }
        
        promise := checkPromise(L, 1)
        
        promise.mu.Lock()
        state := promise.state
        value := promise.value
        err := promise.err
        promise.mu.Unlock()
        
        if state == Fulfilled {
            L.Push(value)
            return 1
        } else if state == Rejected {
            L.RaiseError(err.Error())
            return 0
        }
        
        // Promise pending, set up resumption
        resultCh := make(chan lua.LValue, 1)
        ar.mu.Lock()
        ar.pendingAwait[L] = resultCh
        ar.mu.Unlock()
        
        // Register callback
        promise.mu.Lock()
        promise.callbacks = append(promise.callbacks, Callback{
            OnFulfilled: L.NewFunction(func(L *lua.LState) int {
                value := L.Get(1)
                resultCh <- value
                return 0
            }),
            OnRejected: L.NewFunction(func(L *lua.LState) int {
                err := L.Get(1)
                resultCh <- err
                return 0
            }),
        })
        promise.mu.Unlock()
        
        // Yield coroutine
        return L.Yield(0)
    }))
}

// Example async/await usage
func ExampleAsyncAwait() {
    L := lua.NewState()
    defer L.Close()
    
    RegisterPromiseType(L)
    ar := NewAsyncRuntime(L)
    
    // Create async bridge function
    L.SetGlobal("queryDatabase", L.NewFunction(func(L *lua.LState) int {
        query := L.CheckString(1)
        
        promise := &Promise{state: Pending}
        
        go func() {
            // Simulate database query
            time.Sleep(500 * time.Millisecond)
            promise.Resolve(lua.LString("Result for: " + query))
        }()
        
        ud := L.NewUserData()
        ud.Value = promise
        L.SetMetatable(ud, L.GetTypeMetatable("Promise"))
        L.Push(ud)
        return 1
    }))
    
    err := L.DoString(`
        -- Define async function
        local processData = async(function(id)
            print("Processing ID:", id)
            
            -- Await database query
            local userData = await(queryDatabase("SELECT * FROM users WHERE id = " .. id))
            print("User data:", userData)
            
            -- Multiple awaits
            local orders = await(queryDatabase("SELECT * FROM orders WHERE user_id = " .. id))
            print("Orders:", orders)
            
            -- Process results
            return {
                user = userData,
                orders = orders
            }
        end)
        
        -- Call async function (returns promise)
        processData(123):then(function(result)
            print("Processing complete")
        end):catch(function(err)
            print("Error:", err)
        end)
        
        -- Async function with error handling
        local safeProcess = async(function(id)
            local ok, result = pcall(function()
                return await(queryDatabase("INVALID QUERY"))
            end)
            
            if not ok then
                print("Query failed:", result)
                return nil
            end
            
            return result
        end)
    `)
    
    if err != nil {
        panic(err)
    }
    
    time.Sleep(2 * time.Second) // Wait for async operations
}
```

## Channel-Based Communication

### Lua Channels Implementation

```go
package channels

import (
    "sync"
    lua "github.com/yuin/gopher-lua"
)

type Channel struct {
    ch       chan lua.LValue
    closed   bool
    mu       sync.RWMutex
    waiters  []chan struct{}
}

func RegisterChannelType(L *lua.LState) {
    mt := L.NewTypeMetatable("Channel")
    L.SetGlobal("Channel", mt)
    
    // Constructor
    L.SetField(mt, "new", L.NewFunction(channelNew))
    
    // Methods
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "send":    channelSend,
        "receive": channelReceive,
        "close":   channelClose,
        "select":  channelSelect,
    }))
}

func channelNew(L *lua.LState) int {
    capacity := L.OptInt(1, 0)
    
    ch := &Channel{
        ch:      make(chan lua.LValue, capacity),
        waiters: make([]chan struct{}, 0),
    }
    
    ud := L.NewUserData()
    ud.Value = ch
    L.SetMetatable(ud, L.GetTypeMetatable("Channel"))
    L.Push(ud)
    
    return 1
}

func channelSend(L *lua.LState) int {
    ch := checkChannel(L, 1)
    value := L.Get(2)
    
    // Non-blocking send for coroutines
    if L.Status(L) != lua.ThreadNormal {
        select {
        case ch.ch <- value:
            L.Push(lua.LTrue)
            return 1
        default:
            // Would block, register and yield
            waiter := make(chan struct{})
            ch.mu.Lock()
            ch.waiters = append(ch.waiters, waiter)
            ch.mu.Unlock()
            
            go func() {
                <-waiter
                ch.ch <- value
                L.Resume(L, lua.LTrue)
            }()
            
            return L.Yield(0)
        }
    }
    
    // Regular blocking send
    ch.ch <- value
    L.Push(lua.LTrue)
    return 1
}

// Example producer/consumer with channels
func ExampleChannels() {
    L := lua.NewState()
    defer L.Close()
    
    RegisterChannelType(L)
    
    // Install coroutine helpers
    L.SetGlobal("go", L.NewFunction(func(L *lua.LState) int {
        fn := L.CheckFunction(1)
        
        co := L.NewThread()
        
        // Copy arguments
        nargs := L.GetTop() - 1
        for i := 2; i <= nargs+1; i++ {
            co.Push(L.Get(i))
        }
        
        go func() {
            co.Push(fn)
            for i := 0; i < nargs; i++ {
                co.Push(co.Get(i + 1))
            }
            co.PCall(nargs, 0, nil)
        }()
        
        return 0
    }))
    
    err := L.DoString(`
        local ch = Channel.new(5)
        
        -- Producer coroutine
        go(function()
            for i = 1, 10 do
                print("Producing:", i)
                ch:send(i)
                -- Simulate work
            end
            ch:close()
        end)
        
        -- Consumer coroutines
        for i = 1, 3 do
            go(function(id)
                while true do
                    local value, ok = ch:receive()
                    if not ok then
                        print("Consumer", id, "done")
                        break
                    end
                    print("Consumer", id, "received:", value)
                end
            end, i)
        end
        
        -- Select example
        local ch1 = Channel.new()
        local ch2 = Channel.new()
        
        go(function()
            ch1:send("from ch1")
        end)
        
        go(function()
            ch2:send("from ch2")
        end)
        
        local cases = {
            {ch1, "receive"},
            {ch2, "receive"},
            {nil, "default"}
        }
        
        local idx, value = Channel.select(cases)
        print("Selected case", idx, "with value", value)
    `)
    
    if err != nil {
        panic(err)
    }
    
    time.Sleep(1 * time.Second)
}
```

## Stream Processing

### Async Stream Implementation

```go
package streams

import (
    "bufio"
    "io"
    lua "github.com/yuin/gopher-lua"
)

type AsyncStream struct {
    source   io.Reader
    chunks   chan StreamChunk
    done     chan struct{}
    bufSize  int
}

type StreamChunk struct {
    Data  []byte
    Error error
    EOF   bool
}

func RegisterStreamType(L *lua.LState) {
    mt := L.NewTypeMetatable("Stream")
    L.SetGlobal("Stream", mt)
    
    // Methods
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "read":      streamRead,
        "readLine":  streamReadLine,
        "forEach":   streamForEach,
        "map":       streamMap,
        "filter":    streamFilter,
        "reduce":    streamReduce,
        "collect":   streamCollect,
    }))
}

func streamForEach(L *lua.LState) int {
    stream := checkStream(L, 1)
    fn := L.CheckFunction(2)
    
    if L.Status(L) != lua.ThreadNormal {
        // Async iteration in coroutine
        go func() {
            co := L
            for chunk := range stream.chunks {
                if chunk.Error != nil {
                    co.Push(lua.LNil)
                    co.Push(lua.LString(chunk.Error.Error()))
                    co.Resume(co)
                    break
                }
                
                co.Push(fn)
                co.Push(lua.LString(chunk.Data))
                co.PCall(1, 0, nil)
                
                if chunk.EOF {
                    break
                }
            }
            
            co.Push(lua.LTrue)
            co.Resume(co)
        }()
        
        return L.Yield(0)
    }
    
    // Synchronous iteration
    for chunk := range stream.chunks {
        if chunk.Error != nil {
            L.Push(lua.LNil)
            L.Push(lua.LString(chunk.Error.Error()))
            return 2
        }
        
        L.Push(fn)
        L.Push(lua.LString(chunk.Data))
        L.Call(1, 0)
        
        if chunk.EOF {
            break
        }
    }
    
    L.Push(lua.LTrue)
    return 1
}

// Example async stream processing
func ExampleAsyncStreams() {
    L := lua.NewState()
    defer L.Close()
    
    RegisterStreamType(L)
    
    // Create async file reader
    L.SetGlobal("readFileStream", L.NewFunction(func(L *lua.LState) int {
        filename := L.CheckString(1)
        
        file, err := os.Open(filename)
        if err != nil {
            L.Push(lua.LNil)
            L.Push(lua.LString(err.Error()))
            return 2
        }
        
        stream := &AsyncStream{
            source:  file,
            chunks:  make(chan StreamChunk, 10),
            done:    make(chan struct{}),
            bufSize: 1024,
        }
        
        // Start reading in background
        go func() {
            defer close(stream.chunks)
            defer file.Close()
            
            reader := bufio.NewReaderSize(file, stream.bufSize)
            for {
                data := make([]byte, stream.bufSize)
                n, err := reader.Read(data)
                
                if n > 0 {
                    stream.chunks <- StreamChunk{
                        Data: data[:n],
                        EOF:  err == io.EOF,
                    }
                }
                
                if err != nil {
                    if err != io.EOF {
                        stream.chunks <- StreamChunk{
                            Error: err,
                        }
                    }
                    break
                }
            }
        }()
        
        ud := L.NewUserData()
        ud.Value = stream
        L.SetMetatable(ud, L.GetTypeMetatable("Stream"))
        L.Push(ud)
        return 1
    }))
    
    err := L.DoString(`
        -- Async stream processing
        local processFile = async(function(filename)
            local stream = readFileStream(filename)
            
            local lineCount = 0
            local wordCount = 0
            
            -- Process stream asynchronously
            await(stream:forEach(function(chunk)
                -- Count lines
                for line in chunk:gmatch("[^\n]+") do
                    lineCount = lineCount + 1
                    
                    -- Count words
                    for word in line:gmatch("%S+") do
                        wordCount = wordCount + 1
                    end
                end
            end))
            
            return {
                lines = lineCount,
                words = wordCount
            }
        end)
        
        -- Use the async function
        processFile("test.txt"):then(function(stats)
            print("File stats:", stats.lines, "lines,", stats.words, "words")
        end):catch(function(err)
            print("Error processing file:", err)
        end)
        
        -- Stream pipeline
        local pipeline = async(function(filename)
            local stream = readFileStream(filename)
            
            local result = await(
                stream
                    :map(function(chunk)
                        return chunk:upper()
                    end)
                    :filter(function(chunk)
                        return #chunk > 0
                    end)
                    :collect()
            )
            
            return result
        end)
    `)
    
    if err != nil {
        panic(err)
    }
}
```

## Error Handling in Async Code

### Async Error Propagation

```go
package asyncerrors

type AsyncError struct {
    Message   string
    Stack     []StackFrame
    Timestamp time.Time
    Context   map[string]interface{}
}

type StackFrame struct {
    Function string
    File     string
    Line     int
    Async    bool
}

func InstallAsyncErrorHandling(L *lua.LState) {
    // try/catch for async functions
    L.SetGlobal("asyncTry", L.NewFunction(func(L *lua.LState) int {
        tryFn := L.CheckFunction(1)
        catchFn := L.OptFunction(2, nil)
        finallyFn := L.OptFunction(3, nil)
        
        // Create wrapper promise
        promise := &Promise{state: Pending}
        
        // Execute in coroutine
        co := L.NewThread()
        
        go func() {
            // Set up error handler
            co.Push(L.NewFunction(func(L *lua.LState) int {
                err := L.Get(1)
                
                asyncErr := &AsyncError{
                    Message:   lua.LVAsString(err),
                    Timestamp: time.Now(),
                    Stack:     captureAsyncStack(L),
                    Context:   captureContext(L),
                }
                
                if catchFn != nil {
                    L.Push(catchFn)
                    L.Push(convertAsyncError(L, asyncErr))
                    L.PCall(1, 1, nil)
                    result := L.Get(-1)
                    L.Pop(1)
                    
                    promise.Resolve(result)
                } else {
                    promise.Reject(lua.LString(asyncErr.Message))
                }
                
                return 0
            }))
            
            errorHandler := co.Get(-1)
            
            // Execute try block
            co.Push(tryFn)
            state, err, values := co.Resume(tryFn)
            
            if state == lua.ResumeError {
                // Handle error
                co.Push(errorHandler)
                co.Push(lua.LString(err.Error()))
                co.PCall(1, 0, nil)
            } else {
                // Success
                if len(values) > 0 {
                    promise.Resolve(values[0])
                } else {
                    promise.Resolve(lua.LNil)
                }
            }
            
            // Execute finally block
            if finallyFn != nil {
                L.Push(finallyFn)
                L.PCall(0, 0, nil)
            }
        }()
        
        // Return promise
        ud := L.NewUserData()
        ud.Value = promise
        L.SetMetatable(ud, L.GetTypeMetatable("Promise"))
        L.Push(ud)
        return 1
    }))
}

// Example with error handling
func ExampleAsyncErrorHandling() {
    L := lua.NewState()
    defer L.Close()
    
    RegisterPromiseType(L)
    InstallAsyncErrorHandling(L)
    
    err := L.DoString(`
        -- Async function with error handling
        local fetchUserData = async(function(userId)
            if userId < 0 then
                error("Invalid user ID")
            end
            
            -- Simulate async operation
            local userData = await(queryDatabase("users", userId))
            
            if not userData then
                error("User not found")
            end
            
            return userData
        end)
        
        -- Using asyncTry for error handling
        asyncTry(
            function()
                local user = await(fetchUserData(-1))
                print("User:", user)
            end,
            function(err)
                print("Caught error:", err.message)
                print("Stack trace:")
                for i, frame in ipairs(err.stack) do
                    print(string.format("  %s:%d in %s%s", 
                        frame.file, frame.line, frame.func,
                        frame.async and " (async)" or ""))
                end
            end,
            function()
                print("Cleanup complete")
            end
        )
        
        -- Promise chain with error handling
        fetchUserData(123)
            :then(function(user)
                return fetchUserData(user.managerId)
            end)
            :then(function(manager)
                print("Manager:", manager)
            end)
            :catch(function(err)
                print("Chain error:", err)
            end)
    `)
    
    if err != nil {
        panic(err)
    }
}
```

## Complete Async Bridge Example

### Full Implementation

```go
package complete

import (
    "context"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
    "github.com/your/go-llms/bridge"
)

type AsyncLLMBridge struct {
    bridge   bridge.LLMBridge
    executor *AsyncExecutor
}

type AsyncExecutor struct {
    workers   int
    taskQueue chan Task
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
}

type Task struct {
    ID       string
    Method   string
    Args     []interface{}
    ResultCh chan TaskResult
}

type TaskResult struct {
    Value lua.LValue
    Error error
}

func NewAsyncLLMBridge(bridge bridge.LLMBridge, workers int) *AsyncLLMBridge {
    ctx, cancel := context.WithCancel(context.Background())
    
    executor := &AsyncExecutor{
        workers:   workers,
        taskQueue: make(chan Task, workers*2),
        ctx:       ctx,
        cancel:    cancel,
    }
    
    // Start workers
    for i := 0; i < workers; i++ {
        executor.wg.Add(1)
        go executor.worker()
    }
    
    return &AsyncLLMBridge{
        bridge:   bridge,
        executor: executor,
    }
}

func (ae *AsyncExecutor) worker() {
    defer ae.wg.Done()
    
    for {
        select {
        case task := <-ae.taskQueue:
            result := ae.executeTask(task)
            task.ResultCh <- result
            
        case <-ae.ctx.Done():
            return
        }
    }
}

func RegisterAsyncLLMBridge(L *lua.LState, bridge *AsyncLLMBridge) {
    // Create bridge object
    bridgeTable := L.NewTable()
    
    // Async methods
    L.SetField(bridgeTable, "generateAsync", L.NewFunction(func(L *lua.LState) int {
        prompt := L.CheckString(1)
        options := L.OptTable(2, L.NewTable())
        
        // Create promise
        promise := &Promise{state: Pending}
        
        // Create task
        task := Task{
            ID:       generateID(),
            Method:   "generate",
            Args:     []interface{}{prompt, tableToOptions(options)},
            ResultCh: make(chan TaskResult, 1),
        }
        
        // Submit task
        go func() {
            bridge.executor.taskQueue <- task
            result := <-task.ResultCh
            
            if result.Error != nil {
                promise.Reject(lua.LString(result.Error.Error()))
            } else {
                promise.Resolve(result.Value)
            }
        }()
        
        // Return promise
        ud := L.NewUserData()
        ud.Value = promise
        L.SetMetatable(ud, L.GetTypeMetatable("Promise"))
        L.Push(ud)
        return 1
    }))
    
    // Stream method
    L.SetField(bridgeTable, "streamGenerate", L.NewFunction(func(L *lua.LState) int {
        prompt := L.CheckString(1)
        onChunk := L.CheckFunction(2)
        options := L.OptTable(3, L.NewTable())
        
        // Create stream
        stream := &AsyncStream{
            chunks: make(chan StreamChunk, 10),
        }
        
        // Start streaming
        go func() {
            defer close(stream.chunks)
            
            err := bridge.bridge.GenerateStream(prompt, 
                func(chunk string) error {
                    stream.chunks <- StreamChunk{
                        Data: []byte(chunk),
                    }
                    return nil
                },
                convertOptions(options))
                
            if err != nil {
                stream.chunks <- StreamChunk{Error: err}
            }
        }()
        
        // Process stream in coroutine
        if L.Status(L) != lua.ThreadNormal {
            go func() {
                for chunk := range stream.chunks {
                    if chunk.Error != nil {
                        L.Push(onChunk)
                        L.Push(lua.LNil)
                        L.Push(lua.LString(chunk.Error.Error()))
                        L.PCall(2, 0, nil)
                        break
                    }
                    
                    L.Push(onChunk)
                    L.Push(lua.LString(chunk.Data))
                    L.Push(lua.LNil)
                    L.PCall(2, 1, nil)
                    
                    cont := L.Get(-1)
                    L.Pop(1)
                    
                    if !lua.LVAsBool(cont) {
                        break
                    }
                }
                
                L.Resume(L)
            }()
            
            return L.Yield(0)
        }
        
        // Synchronous processing
        for chunk := range stream.chunks {
            if chunk.Error != nil {
                L.Push(onChunk)
                L.Push(lua.LNil)
                L.Push(lua.LString(chunk.Error.Error()))
                L.Call(2, 0)
                break
            }
            
            L.Push(onChunk)
            L.Push(lua.LString(chunk.Data))
            L.Push(lua.LNil)
            L.Call(2, 1)
            
            cont := L.Get(-1)
            L.Pop(1)
            
            if !lua.LVAsBool(cont) {
                break
            }
        }
        
        return 0
    }))
    
    L.SetGlobal("llmBridge", bridgeTable)
}

// Example usage
func ExampleCompleteAsync() {
    L := lua.NewState()
    defer L.Close()
    
    // Set up async runtime
    RegisterPromiseType(L)
    InstallAsyncAwait(L)
    
    // Create and register bridge
    llmBridge, _ := bridge.NewLLMBridge(bridge.Config{})
    asyncBridge := NewAsyncLLMBridge(llmBridge, 4)
    RegisterAsyncLLMBridge(L, asyncBridge)
    
    err := L.DoString(`
        -- Async chat conversation
        local chat = async(function()
            local conversation = {}
            
            -- First message
            local response1 = await(llmBridge.generateAsync(
                "Hello, I need help with async programming in Lua",
                {temperature = 0.7}
            ))
            table.insert(conversation, {role = "assistant", content = response1})
            
            -- Follow-up
            local response2 = await(llmBridge.generateAsync(
                "Can you show me an example with coroutines?",
                {temperature = 0.7}
            ))
            table.insert(conversation, {role = "assistant", content = response2})
            
            return conversation
        end)
        
        -- Execute async chat
        chat():then(function(conversation)
            print("Conversation complete:", #conversation, "messages")
        end):catch(function(err)
            print("Chat error:", err)
        end)
        
        -- Streaming example
        local streamChat = async(function(prompt)
            local fullResponse = ""
            
            await(llmBridge.streamGenerate(
                prompt,
                function(chunk, err)
                    if err then
                        error(err)
                    end
                    
                    io.write(chunk)
                    io.flush()
                    fullResponse = fullResponse .. chunk
                    
                    return true -- Continue streaming
                end,
                {temperature = 0.7}
            ))
            
            return fullResponse
        end)
        
        -- Parallel requests
        local parallelQueries = async(function(queries)
            local promises = {}
            
            for i, query in ipairs(queries) do
                promises[i] = llmBridge.generateAsync(query)
            end
            
            return await(Promise.all(promises))
        end)
        
        parallelQueries({
            "What is async programming?",
            "Explain coroutines",
            "What are promises?"
        }):then(function(responses)
            print("Got", #responses, "responses")
        end)
    `)
    
    if err != nil {
        panic(err)
    }
    
    time.Sleep(5 * time.Second) // Wait for async operations
    
    // Cleanup
    asyncBridge.executor.cancel()
    asyncBridge.executor.wg.Wait()
}
```

## Summary

These examples demonstrate comprehensive async patterns using Lua coroutines:
1. Basic coroutine usage and Go integration
2. Promise-based async operations
3. Full async/await implementation
4. Channel-based communication
5. Async stream processing
6. Error handling in async code
7. Complete async bridge implementation

Key features include non-blocking execution, natural async syntax, efficient resource usage, and seamless integration with Go's concurrency model.