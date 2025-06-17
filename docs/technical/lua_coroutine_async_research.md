# Lua Coroutine Support for Async Bridge Operations Research

This document investigates using Lua coroutines in GopherLua to handle asynchronous bridge operations, enabling non-blocking execution patterns.

## Executive Summary

Lua coroutines provide cooperative multitasking that can be effectively combined with Go's concurrency model to create async bridge operations. This enables non-blocking script execution while maintaining simple, synchronous-looking Lua code.

## Coroutine Fundamentals in GopherLua

### Basic Coroutine Operations

```go
// GopherLua supports standard Lua coroutine operations:
// - coroutine.create(f)    Create new coroutine
// - coroutine.resume(co)   Resume execution
// - coroutine.yield()      Suspend execution
// - coroutine.status(co)   Get coroutine status
// - coroutine.wrap(f)      Create wrapped coroutine

// Coroutine states in Lua:
// - "suspended"  Initial state or after yield
// - "running"    Currently executing
// - "normal"     Active but not running (resumed another)
// - "dead"       Finished execution
```

### GopherLua Coroutine Integration

```go
type CoroutineManager struct {
    L           *lua.LState
    coroutines  map[string]*lua.LState
    channels    map[string]chan lua.LValue
    mu          sync.RWMutex
}

func (cm *CoroutineManager) CreateCoroutine(L *lua.LState, fn lua.LValue) (*lua.LState, error) {
    // Create new coroutine
    co := L.NewThread()
    
    // Push function onto coroutine stack
    co.Push(fn)
    
    // Generate unique ID
    id := generateCoroutineID()
    
    cm.mu.Lock()
    cm.coroutines[id] = co
    cm.mu.Unlock()
    
    return co, nil
}
```

## Async Bridge Pattern

### 1. Promise-Based Async Pattern

```go
type LuaPromise struct {
    state    PromiseState
    value    lua.LValue
    err      error
    callbacks []PromiseCallback
    mu       sync.Mutex
}

type PromiseState int

const (
    PromisePending PromiseState = iota
    PromiseFulfilled
    PromiseRejected
)

type PromiseCallback struct {
    OnFulfilled lua.LValue
    OnRejected  lua.LValue
    Coroutine   *lua.LState
}

func CreatePromiseType(L *lua.LState) {
    mt := L.NewTypeMetatable("Promise")
    
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "then":    promiseThen,
        "catch":   promiseCatch,
        "finally": promiseFinally,
        "await":   promiseAwait,
    }))
}

func promiseAwait(L *lua.LState) int {
    promise := checkPromise(L, 1)
    
    promise.mu.Lock()
    state := promise.state
    promise.mu.Unlock()
    
    if state == PromisePending {
        // Yield the coroutine
        return L.Yield(1)
    }
    
    // Promise already resolved
    if state == PromiseFulfilled {
        L.Push(promise.value)
        return 1
    } else {
        L.Push(lua.LNil)
        L.Push(lua.LString(promise.err.Error()))
        return 2
    }
}
```

### 2. Async Bridge Method Implementation

```go
type AsyncBridge struct {
    bridge   bridge.Bridge
    executor *AsyncExecutor
}

type AsyncExecutor struct {
    workers    int
    taskQueue  chan AsyncTask
    resultMap  sync.Map
}

type AsyncTask struct {
    ID        string
    Method    string
    Args      []interface{}
    Promise   *LuaPromise
    Coroutine *lua.LState
}

func (ab *AsyncBridge) GenerateAsync(L *lua.LState) int {
    prompt := L.CheckString(1)
    options := L.OptTable(2, L.NewTable())
    
    // Create promise
    promise := &LuaPromise{
        state: PromisePending,
    }
    
    // Create task
    task := AsyncTask{
        ID:        generateTaskID(),
        Method:    "generate",
        Args:      []interface{}{prompt, tableToOptions(options)},
        Promise:   promise,
        Coroutine: L,
    }
    
    // Submit to executor
    ab.executor.Submit(task)
    
    // Return promise
    ud := L.NewUserData()
    ud.Value = promise
    L.SetMetatable(ud, L.GetTypeMetatable("Promise"))
    L.Push(ud)
    
    return 1
}

func (ae *AsyncExecutor) Submit(task AsyncTask) {
    ae.taskQueue <- task
}

func (ae *AsyncExecutor) Worker() {
    for task := range ae.taskQueue {
        // Execute bridge method
        result, err := ae.executeTask(task)
        
        // Update promise
        task.Promise.mu.Lock()
        if err != nil {
            task.Promise.state = PromiseRejected
            task.Promise.err = err
        } else {
            task.Promise.state = PromiseFulfilled
            task.Promise.value = result
        }
        callbacks := task.Promise.callbacks
        task.Promise.mu.Unlock()
        
        // Resume waiting coroutines
        for _, cb := range callbacks {
            if cb.Coroutine != nil {
                cb.Coroutine.Resume(cb.Coroutine, result)
            }
        }
    }
}
```

### 3. Coroutine-Based Async/Await

```go
func InstallAsyncAwait(L *lua.LState) {
    // Global async function wrapper
    L.SetGlobal("async", L.NewFunction(func(L *lua.LState) int {
        fn := L.CheckFunction(1)
        
        // Create async wrapper
        wrapper := L.NewFunction(func(L *lua.LState) int {
            // Create coroutine
            co := L.NewThread()
            
            // Copy arguments
            nargs := L.GetTop()
            for i := 1; i <= nargs; i++ {
                co.Push(L.Get(i))
            }
            
            // Push function
            co.Push(fn)
            
            // Start coroutine
            state, err, values := co.Resume(fn, L.Get(1))
            
            if state == lua.ResumeError {
                L.RaiseError("async error: %v", err)
            }
            
            // Return promise or values
            for _, v := range values {
                L.Push(v)
            }
            return len(values)
        })
        
        L.Push(wrapper)
        return 1
    }))
    
    // Global await function
    L.SetGlobal("await", L.NewFunction(func(L *lua.LState) int {
        // Check if we're in a coroutine
        if L.Status(L) == lua.ThreadNormal {
            L.RaiseError("await must be called from within an async function")
        }
        
        promise := L.Get(1)
        
        // Handle different async types
        switch v := promise.(type) {
        case *lua.LUserData:
            if p, ok := v.Value.(*LuaPromise); ok {
                return handlePromiseAwait(L, p)
            }
        }
        
        L.RaiseError("await expects a promise")
        return 0
    }))
}

func handlePromiseAwait(L *lua.LState, promise *LuaPromise) int {
    promise.mu.Lock()
    state := promise.state
    
    if state == PromisePending {
        // Register coroutine for resumption
        promise.callbacks = append(promise.callbacks, PromiseCallback{
            Coroutine: L,
        })
        promise.mu.Unlock()
        
        // Yield coroutine
        return L.Yield(0)
    }
    
    promise.mu.Unlock()
    
    // Promise already resolved
    if state == PromiseFulfilled {
        L.Push(promise.value)
        return 1
    } else {
        L.RaiseError(promise.err.Error())
        return 0
    }
}
```

## Channel-Based Coroutine Communication

### 1. Lua Channels for Coroutines

```go
type LuaChannel struct {
    ch       chan lua.LValue
    capacity int
    closed   bool
    mu       sync.RWMutex
}

func CreateChannelType(L *lua.LState) {
    mt := L.NewTypeMetatable("Channel")
    
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "send":     channelSend,
        "receive":  channelReceive,
        "select":   channelSelect,
        "close":    channelClose,
        "len":      channelLen,
        "cap":      channelCap,
    }))
}

func NewChannel(L *lua.LState) int {
    capacity := L.OptInt(1, 0)
    
    ch := &LuaChannel{
        ch:       make(chan lua.LValue, capacity),
        capacity: capacity,
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
    select {
    case ch.ch <- value:
        L.Push(lua.LTrue)
        return 1
    default:
        // Would block, yield coroutine
        if L.Status(L) != lua.ThreadNormal {
            // Store send operation for later
            registerPendingSend(L, ch, value)
            return L.Yield(0)
        }
        
        // Not in coroutine, block
        ch.ch <- value
        L.Push(lua.LTrue)
        return 1
    }
}

func channelReceive(L *lua.LState) int {
    ch := checkChannel(L, 1)
    
    // Non-blocking receive for coroutines
    select {
    case value, ok := <-ch.ch:
        if !ok {
            L.Push(lua.LNil)
            L.Push(lua.LFalse) // Channel closed
            return 2
        }
        L.Push(value)
        L.Push(lua.LTrue)
        return 2
    default:
        // Would block, yield coroutine
        if L.Status(L) != lua.ThreadNormal {
            registerPendingReceive(L, ch)
            return L.Yield(0)
        }
        
        // Not in coroutine, block
        value, ok := <-ch.ch
        if !ok {
            L.Push(lua.LNil)
            L.Push(lua.LFalse)
            return 2
        }
        L.Push(value)
        L.Push(lua.LTrue)
        return 2
    }
}
```

### 2. Select Operation for Multiple Channels

```go
func channelSelect(L *lua.LState) int {
    cases := L.CheckTable(1)
    
    // Build select cases
    var selectCases []reflect.SelectCase
    var caseInfo []SelectCaseInfo
    
    cases.ForEach(func(idx, caseVal lua.LValue) {
        caseTable := caseVal.(*lua.LTable)
        
        ch := checkChannel(L, caseTable.RawGetInt(1))
        op := lua.LVAsString(caseTable.RawGetInt(2))
        
        switch op {
        case "send":
            value := caseTable.RawGetInt(3)
            selectCases = append(selectCases, reflect.SelectCase{
                Dir:  reflect.SelectSend,
                Chan: reflect.ValueOf(ch.ch),
                Send: reflect.ValueOf(value),
            })
        case "receive":
            selectCases = append(selectCases, reflect.SelectCase{
                Dir:  reflect.SelectRecv,
                Chan: reflect.ValueOf(ch.ch),
            })
        case "default":
            selectCases = append(selectCases, reflect.SelectCase{
                Dir: reflect.SelectDefault,
            })
        }
        
        caseInfo = append(caseInfo, SelectCaseInfo{
            Channel: ch,
            Op:      op,
            Index:   int(idx.(lua.LNumber)),
        })
    })
    
    // Perform select
    chosen, recv, recvOK := reflect.Select(selectCases)
    
    // Return results
    L.Push(lua.LNumber(caseInfo[chosen].Index))
    
    if caseInfo[chosen].Op == "receive" {
        if recvOK {
            L.Push(luaValueOf(L, recv))
            L.Push(lua.LTrue)
            return 3
        } else {
            L.Push(lua.LNil)
            L.Push(lua.LFalse)
            return 3
        }
    }
    
    return 1
}

type SelectCaseInfo struct {
    Channel *LuaChannel
    Op      string
    Index   int
}
```

## Stream Processing with Coroutines

### 1. Streaming Bridge Operations

```go
type StreamingBridge struct {
    bridge bridge.Bridge
}

func (sb *StreamingBridge) StreamGenerate(L *lua.LState) int {
    prompt := L.CheckString(1)
    options := L.OptTable(2, L.NewTable())
    
    // Create stream channel
    streamCh := make(chan bridge.StreamChunk, 10)
    
    // Start streaming in goroutine
    go func() {
        defer close(streamCh)
        
        err := sb.bridge.GenerateStream(prompt, streamCh, 
            convertOptions(options))
        if err != nil {
            // Send error as special chunk
            streamCh <- bridge.StreamChunk{
                Error: err,
            }
        }
    }()
    
    // Create Lua iterator
    iter := L.NewFunction(func(L *lua.LState) int {
        select {
        case chunk, ok := <-streamCh:
            if !ok {
                // Stream ended
                L.Push(lua.LNil)
                return 1
            }
            
            if chunk.Error != nil {
                L.Push(lua.LNil)
                L.Push(lua.LString(chunk.Error.Error()))
                return 2
            }
            
            // Return chunk data
            chunkTable := L.NewTable()
            L.SetField(chunkTable, "content", lua.LString(chunk.Content))
            L.SetField(chunkTable, "done", lua.LBool(chunk.Done))
            L.Push(chunkTable)
            return 1
            
        default:
            // Would block, yield if in coroutine
            if L.Status(L) != lua.ThreadNormal {
                registerStreamChannel(L, streamCh)
                return L.Yield(0)
            }
            
            // Not in coroutine, block
            chunk, ok := <-streamCh
            if !ok {
                L.Push(lua.LNil)
                return 1
            }
            
            chunkTable := L.NewTable()
            L.SetField(chunkTable, "content", lua.LString(chunk.Content))
            L.Push(chunkTable)
            return 1
        }
    })
    
    L.Push(iter)
    return 1
}
```

### 2. Async Pipeline Processing

```go
type AsyncPipeline struct {
    stages []PipelineStage
}

type PipelineStage struct {
    Name      string
    Processor lua.LValue
    Parallel  int
}

func CreateAsyncPipeline(L *lua.LState) int {
    stages := L.CheckTable(1)
    
    pipeline := &AsyncPipeline{
        stages: make([]PipelineStage, 0),
    }
    
    stages.ForEach(func(_, stageVal lua.LValue) {
        stage := stageVal.(*lua.LTable)
        
        pipeline.stages = append(pipeline.stages, PipelineStage{
            Name:      getStringField(L, stage, "name", ""),
            Processor: stage.RawGetString("processor"),
            Parallel:  int(getNumberField(L, stage, "parallel", 1)),
        })
    })
    
    // Create pipeline processor
    processor := L.NewFunction(func(L *lua.LState) int {
        input := L.Get(1)
        
        // Create coroutine for pipeline
        co := L.NewThread()
        
        // Process through stages
        go func() {
            current := input
            
            for _, stage := range pipeline.stages {
                // Create stage channels
                inCh := make(chan lua.LValue, stage.Parallel)
                outCh := make(chan lua.LValue, stage.Parallel)
                
                // Start stage workers
                var wg sync.WaitGroup
                for i := 0; i < stage.Parallel; i++ {
                    wg.Add(1)
                    go func() {
                        defer wg.Done()
                        
                        for item := range inCh {
                            // Call processor
                            co.Push(stage.Processor)
                            co.Push(item)
                            co.Call(1, 1)
                            result := co.Get(-1)
                            co.Pop(1)
                            
                            outCh <- result
                        }
                    }()
                }
                
                // Feed input
                go func() {
                    inCh <- current
                    close(inCh)
                }()
                
                // Collect output
                go func() {
                    wg.Wait()
                    close(outCh)
                }()
                
                // Get result
                current = <-outCh
            }
            
            // Resume main coroutine with result
            L.Resume(co, current)
        }()
        
        // Yield main coroutine
        return L.Yield(0)
    })
    
    L.Push(processor)
    return 1
}
```

## Error Handling in Async Operations

### 1. Async Error Propagation

```go
type AsyncError struct {
    Error     error
    Stage     string
    Timestamp time.Time
    Context   map[string]interface{}
}

func HandleAsyncError(L *lua.LState, err error, context string) {
    asyncErr := &AsyncError{
        Error:     err,
        Stage:     context,
        Timestamp: time.Now(),
        Context:   captureErrorContext(L),
    }
    
    // Check if we're in a coroutine
    if L.Status(L) != lua.ThreadNormal {
        // Store error for coroutine
        setCoroutineError(L, asyncErr)
        L.Yield(0)
    } else {
        // Regular error handling
        L.RaiseError("async error in %s: %v", context, err)
    }
}

func setCoroutineError(L *lua.LState, err *AsyncError) {
    errTable := L.NewTable()
    L.SetField(errTable, "message", lua.LString(err.Error.Error()))
    L.SetField(errTable, "stage", lua.LString(err.Stage))
    L.SetField(errTable, "timestamp", lua.LNumber(err.Timestamp.Unix()))
    
    // Store in coroutine registry
    L.SetGlobal("__coroutine_error", errTable)
}
```

### 2. Try-Catch for Coroutines

```go
func InstallCoroutineTryCatch(L *lua.LState) {
    L.SetGlobal("try", L.NewFunction(func(L *lua.LState) int {
        tryFn := L.CheckFunction(1)
        catchFn := L.CheckFunction(2)
        finallyFn := L.OptFunction(3, nil)
        
        // Create protected coroutine
        co := L.NewThread()
        
        // Execute try block
        co.Push(tryFn)
        state, err, values := co.Resume(tryFn)
        
        if state == lua.ResumeError {
            // Execute catch block
            L.Push(catchFn)
            L.Push(lua.LString(err.Error()))
            L.Call(1, lua.MultRet)
        } else {
            // Return try block results
            for _, v := range values {
                L.Push(v)
            }
        }
        
        // Execute finally block
        if finallyFn != nil {
            L.Push(finallyFn)
            L.Call(0, 0)
        }
        
        return L.GetTop()
    }))
}
```

## Performance Considerations

### 1. Coroutine Pool

```go
type CoroutinePool struct {
    available chan *lua.LState
    factory   func() *lua.LState
    maxSize   int
}

func NewCoroutinePool(L *lua.LState, size int) *CoroutinePool {
    pool := &CoroutinePool{
        available: make(chan *lua.LState, size),
        factory: func() *lua.LState {
            return L.NewThread()
        },
        maxSize: size,
    }
    
    // Pre-populate pool
    for i := 0; i < size/2; i++ {
        pool.available <- pool.factory()
    }
    
    return pool
}

func (cp *CoroutinePool) Get() *lua.LState {
    select {
    case co := <-cp.available:
        // Reset coroutine state
        co.SetTop(0)
        return co
    default:
        // Create new coroutine
        return cp.factory()
    }
}

func (cp *CoroutinePool) Put(co *lua.LState) {
    select {
    case cp.available <- co:
        // Returned to pool
    default:
        // Pool full, let GC handle it
    }
}
```

### 2. Batch Async Operations

```go
func BatchAsyncExecute(L *lua.LState) int {
    operations := L.CheckTable(1)
    batchSize := L.OptInt(2, 10)
    
    var results []lua.LValue
    var errors []error
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    // Process in batches
    batch := make([]AsyncOperation, 0, batchSize)
    
    operations.ForEach(func(idx, op lua.LValue) {
        batch = append(batch, AsyncOperation{
            Index: int(idx.(lua.LNumber)),
            Op:    op,
        })
        
        if len(batch) >= batchSize {
            // Process batch
            wg.Add(1)
            processBatch := batch
            batch = make([]AsyncOperation, 0, batchSize)
            
            go func(ops []AsyncOperation) {
                defer wg.Done()
                
                for _, op := range ops {
                    // Execute operation
                    result, err := executeAsyncOp(L, op.Op)
                    
                    mu.Lock()
                    if err != nil {
                        errors = append(errors, err)
                    } else {
                        results = append(results, result)
                    }
                    mu.Unlock()
                }
            }(processBatch)
        }
    })
    
    // Process remaining
    if len(batch) > 0 {
        wg.Add(1)
        go func(ops []AsyncOperation) {
            defer wg.Done()
            // Process remaining operations
        }(batch)
    }
    
    // Wait for completion
    wg.Wait()
    
    // Return results
    resultsTable := L.NewTable()
    for i, result := range results {
        resultsTable.RawSetInt(i+1, result)
    }
    
    errorsTable := L.NewTable()
    for i, err := range errors {
        errorsTable.RawSetInt(i+1, lua.LString(err.Error()))
    }
    
    L.Push(resultsTable)
    L.Push(errorsTable)
    return 2
}
```

## Testing Async Patterns

### Testing Coroutine Integration

```go
func TestCoroutineAsyncBridge(t *testing.T) {
    L := lua.NewState()
    defer L.Close()
    
    // Install async support
    InstallAsyncAwait(L)
    CreatePromiseType(L)
    
    // Create async bridge
    bridge := &AsyncBridge{
        bridge:   mockBridge,
        executor: NewAsyncExecutor(4),
    }
    
    RegisterAsyncBridge(L, bridge)
    
    // Test async/await pattern
    err := L.DoString(`
        local async_test = async(function()
            local result = await(bridge:generateAsync("test prompt"))
            assert(result == "test response")
            
            -- Multiple awaits
            local results = {}
            for i = 1, 3 do
                results[i] = await(bridge:generateAsync("prompt " .. i))
            end
            
            return results
        end)
        
        local results = async_test()
        assert(#results == 3)
    `)
    
    if err != nil {
        t.Fatalf("Async test failed: %v", err)
    }
}
```

## Best Practices

1. **Use Coroutines for I/O**: Ideal for network calls, file operations
2. **Avoid Blocking**: Never block in coroutine without yielding
3. **Error Propagation**: Ensure errors bubble up through coroutine chain
4. **Resource Cleanup**: Always clean up resources in finally blocks
5. **Pool Coroutines**: Reuse coroutines to reduce allocation overhead

## Implementation Checklist

- [ ] Basic coroutine integration with bridges
- [ ] Promise-based async pattern
- [ ] Async/await syntax support
- [ ] Channel-based communication
- [ ] Stream processing with coroutines
- [ ] Error handling in async operations
- [ ] Coroutine pooling
- [ ] Batch async operations
- [ ] Testing framework for async patterns
- [ ] Performance optimization

## Summary

Coroutine support for async bridge operations enables:
1. Non-blocking script execution
2. Natural async/await patterns in Lua
3. Efficient stream processing
4. Channel-based communication
5. Seamless integration with Go's concurrency

This provides a powerful model for handling async operations while maintaining simple, readable Lua code.