# Lua Goroutine and Channel Integration Design

## Overview
This document designs the integration of Go's concurrency model (goroutines and channels) with Lua scripts in go-llmspell, leveraging GopherLua's LChannel support.

## Core Concepts

### 1. Concurrency Model
- **One LState per goroutine**: LState is NOT thread-safe
- **Channel-based communication**: Use LChannel for inter-script communication
- **No shared state**: Scripts communicate only through channels

### 2. Channel Support
GopherLua provides native channel support through:
- `LChannel` type: Wraps Go `chan lua.LValue`
- Channel module: Lua-side channel operations
- Bidirectional: Send/receive from both Go and Lua

## Architecture Design

### Channel Bridge
```go
// ChannelBridge provides channel operations for Lua scripts
type ChannelBridge struct {
    mu       sync.RWMutex
    channels map[string]chan lua.LValue
    config   ChannelConfig
}

type ChannelConfig struct {
    DefaultBufferSize int
    MaxChannels      int
    TimeoutDefault   time.Duration
}

// Channel operations exposed to Lua
func (cb *ChannelBridge) Methods() []engine.MethodInfo {
    return []engine.MethodInfo{
        {
            Name:        "make",
            Description: "Create a new channel",
            Parameters: []engine.ParameterInfo{
                {Name: "buffer_size", Type: "number", Required: false},
            },
            ReturnType: "channel",
        },
        {
            Name:        "select",
            Description: "Select from multiple channel operations",
            Parameters: []engine.ParameterInfo{
                {Name: "cases", Type: "array", Required: true},
                {Name: "timeout", Type: "number", Required: false},
            },
            ReturnType: "object",
        },
    }
}
```

### Goroutine Management
```go
// GoroutineManager manages Lua script execution in goroutines
type GoroutineManager struct {
    pool      *LStatePool
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
    
    // Tracking
    active    sync.Map // goroutine ID -> info
    errors    chan error
    
    // Limits
    maxActive int
}

// SpawnScript runs a Lua script in a new goroutine
func (gm *GoroutineManager) SpawnScript(script string, args map[string]interface{}) (string, error) {
    // Check limits
    if gm.getActiveCount() >= gm.maxActive {
        return "", ErrTooManyGoroutines
    }
    
    id := generateID()
    
    gm.wg.Add(1)
    go func() {
        defer gm.wg.Done()
        defer gm.removeActive(id)
        
        // Get LState from pool
        L, err := gm.pool.Get(gm.ctx)
        if err != nil {
            gm.errors <- fmt.Errorf("goroutine %s: %w", id, err)
            return
        }
        defer gm.pool.Put(L)
        
        // Setup goroutine context
        gm.setupGoroutineEnv(L, id, args)
        
        // Execute script
        if err := L.DoString(script); err != nil {
            gm.errors <- fmt.Errorf("goroutine %s: %w", id, err)
        }
    }()
    
    gm.addActive(id)
    return id, nil
}
```

## Channel Operations API

### Lua-side Channel Methods
```lua
-- Create channel
local ch = channel.make(10)  -- buffered channel with size 10
local ch2 = channel.make()    -- unbuffered channel

-- Send (blocking)
ch:send(value)

-- Send with timeout
local ok = ch:send(value, 1.0)  -- 1 second timeout
if not ok then
    print("send timeout")
end

-- Receive (blocking)
local value, ok = ch:receive()
if not ok then
    print("channel closed")
end

-- Receive with timeout
local value, ok = ch:receive(1.0)  -- 1 second timeout
if not ok then
    print("receive timeout or closed")
end

-- Try receive (non-blocking)
local value, ok = ch:try_receive()
if ok then
    print("got value:", value)
else
    print("would block")
end

-- Close channel
ch:close()

-- Check if closed
if ch:is_closed() then
    print("channel is closed")
end

-- Length and capacity
print("items in channel:", ch:len())
print("channel capacity:", ch:cap())
```

### Channel Select
```lua
-- Select with multiple cases
local result = channel.select({
    -- Receive case
    {"|<-", ch1, function(value, ok)
        if ok then
            print("received from ch1:", value)
            return "ch1", value
        end
    end},
    
    -- Send case
    {"<-|", ch2, "hello", function(ok)
        if ok then
            print("sent to ch2")
            return "ch2_sent"
        end
    end},
    
    -- Default case (non-blocking)
    {"default", function()
        print("no channel ready")
        return "default"
    end}
}, 2.0)  -- 2 second timeout

if result then
    local case, value = result[1], result[2]
    print("selected case:", case, "value:", value)
else
    print("select timeout")
end
```

## Type Restrictions

### What CAN be sent through channels:
- Primitives: nil, bool, number, string
- Simple tables (without metatables)
- Errors (as userdata)
- Other channels

### What CANNOT be sent through channels:
- Functions (not thread-safe)
- Tables with metatables
- Userdata (except controlled types)
- Thread/coroutine objects
- Bridge objects (unless specially handled)

### Safe Channel Value Converter
```go
// SafeChannelValue ensures value is safe to send through channel
func (c *LuaTypeConverter) SafeChannelValue(L *lua.LState, lv lua.LValue) (lua.LValue, error) {
    switch lv.Type() {
    case lua.LTNil, lua.LTBool, lua.LTNumber, lua.LTString, lua.LTChannel:
        return lv, nil // Safe primitives
        
    case lua.LTTable:
        table := lv.(*lua.LTable)
        
        // Check for metatable
        if L.GetMetatable(table) != lua.LNil {
            return nil, ErrUnsafeChannelValue{Type: "table with metatable"}
        }
        
        // Deep copy table to ensure safety
        safeCopy := L.NewTable()
        err := c.deepCopyTable(L, table, safeCopy, make(map[*lua.LTable]bool))
        if err != nil {
            return nil, err
        }
        
        return safeCopy, nil
        
    case lua.LTUserData:
        // Only allow specific userdata types
        ud := lv.(*lua.LUserData)
        switch ud.Value.(type) {
        case error:
            return lv, nil // Errors are safe
        default:
            return nil, ErrUnsafeChannelValue{Type: "userdata"}
        }
        
    default:
        return nil, ErrUnsafeChannelValue{Type: lv.Type().String()}
    }
}
```

## Coroutine Integration

### Lua Coroutines vs Go Goroutines
- **Lua coroutines**: Cooperative, single-threaded
- **Go goroutines**: Preemptive, multi-threaded
- **Integration**: Use coroutines within single LState, goroutines for parallelism

### Coroutine Support
```lua
-- Create coroutine
local co = coroutine.create(function(initial)
    print("coroutine started with:", initial)
    local value = coroutine.yield("first yield")
    print("resumed with:", value)
    return "done"
end)

-- Resume coroutine
local ok, result = coroutine.resume(co, "initial value")
print("yielded:", result)

-- Resume again
ok, result = coroutine.resume(co, "second value")
print("returned:", result)

-- Check status
print("status:", coroutine.status(co))  -- "dead"
```

### Async Bridge Operations
```go
// AsyncBridgeCall wraps bridge calls with coroutine support
func (e *LuaEngine) AsyncBridgeCall(L *lua.LState) int {
    // Get bridge and method
    bridgeName := L.CheckString(1)
    methodName := L.CheckString(2)
    args := L.CheckTable(3)
    
    // Convert arguments
    goArgs, err := e.converter.TableToSlice(args)
    if err != nil {
        L.RaiseError("invalid arguments: %v", err)
    }
    
    // Create coroutine for async operation
    co, _ := L.NewThread()
    
    // Run bridge call in goroutine
    go func() {
        // Get bridge
        bridge := e.bridgeManager.GetBridge(bridgeName)
        if bridge == nil {
            L.Resume(co, lua.LString("bridge not found"))
            return
        }
        
        // Call method
        ctx := context.Background()
        results, err := bridge.Call(ctx, methodName, goArgs)
        
        // Resume coroutine with results
        if err != nil {
            L.Resume(co, lua.LNil, lua.LString(err.Error()))
        } else {
            // Convert results
            luaResults := e.convertResults(L, results)
            L.Resume(co, luaResults...)
        }
    }()
    
    // Return coroutine
    L.Push(co)
    return 1
}
```

## Integration Patterns

### 1. Worker Pool Pattern
```lua
-- Worker function
local function worker(id, jobs, results)
    while true do
        local job = jobs:receive()
        if not job then
            break  -- Channel closed
        end
        
        -- Process job
        local result = process_job(job)
        results:send({id = id, job = job, result = result})
    end
end

-- Create channels
local jobs = channel.make(100)
local results = channel.make(100)

-- Spawn workers
for i = 1, 5 do
    go.spawn(worker, i, jobs, results)
end

-- Send jobs
for i = 1, 50 do
    jobs:send({id = i, data = "job" .. i})
end
jobs:close()

-- Collect results
for i = 1, 50 do
    local result = results:receive()
    print("Result:", result.id, result.result)
end
```

### 2. Pipeline Pattern
```lua
-- Pipeline stages
local function stage1(input, output)
    for value in channel.range(input) do
        output:send(value * 2)
    end
    output:close()
end

local function stage2(input, output)
    for value in channel.range(input) do
        output:send(value + 10)
    end
    output:close()
end

-- Create pipeline
local ch1 = channel.make()
local ch2 = channel.make()
local ch3 = channel.make()

go.spawn(stage1, ch1, ch2)
go.spawn(stage2, ch2, ch3)

-- Send data
go.spawn(function()
    for i = 1, 10 do
        ch1:send(i)
    end
    ch1:close()
end)

-- Collect results
for result in channel.range(ch3) do
    print("Result:", result)
end
```

### 3. Fan-out/Fan-in Pattern
```lua
-- Fan-out: distribute work
local function fanOut(input, workers)
    local outputs = {}
    for i = 1, workers do
        local out = channel.make()
        outputs[i] = out
        
        go.spawn(function()
            for value in channel.range(input) do
                -- Process value
                local result = process(value)
                out:send(result)
            end
            out:close()
        end)
    end
    return outputs
end

-- Fan-in: merge results
local function fanIn(inputs)
    local output = channel.make()
    
    for _, input in ipairs(inputs) do
        go.spawn(function()
            for value in channel.range(input) do
                output:send(value)
            end
        end)
    end
    
    -- Close output when all inputs are done
    go.spawn(function()
        for _ in ipairs(inputs) do
            -- Wait for all to complete
        end
        output:close()
    end)
    
    return output
end
```

## Error Handling

### Channel Errors
```go
type ChannelError struct {
    Op      string // "send", "receive", "close"
    Channel string // Channel identifier
    Reason  string // Error reason
}

// Lua-side error handling
local ok, err = pcall(function()
    ch:send(unsafe_value)
end)
if not ok then
    print("Channel error:", err)
end
```

### Goroutine Errors
```go
// Error collection from goroutines
func (gm *GoroutineManager) CollectErrors() []error {
    var errors []error
    
    // Non-blocking collection
    for {
        select {
        case err := <-gm.errors:
            errors = append(errors, err)
        default:
            return errors
        }
    }
}
```

## Performance Considerations

### 1. Channel Buffer Sizing
- Small buffers (1-10) for tight synchronization
- Large buffers (100-1000) for producer/consumer patterns
- Unbuffered for strict synchronization

### 2. Goroutine Limits
- Limit concurrent Lua goroutines (default: 100)
- Monitor goroutine lifecycle
- Proper cleanup on script completion

### 3. Value Copying
- Deep copy tables for channel safety
- Consider serialization for large data
- Use references where safe

## Testing Strategy

### 1. Concurrency Tests
- Race condition detection
- Deadlock detection
- Channel operation timeouts

### 2. Integration Tests
- Multi-goroutine scripts
- Channel communication patterns
- Error propagation

### 3. Performance Tests
- Channel throughput
- Goroutine spawn/cleanup overhead
- Memory usage under load

## Security Considerations

### 1. Resource Limits
- Maximum goroutines per script
- Channel buffer size limits
- Timeout enforcement

### 2. Value Sanitization
- Prevent unsafe values in channels
- Validate bridge object handling
- Control cross-script communication

### 3. Isolation
- No shared state between scripts
- Channel-only communication
- Proper cleanup on termination