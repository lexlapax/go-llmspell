# Async Callbacks Example

This example demonstrates how async callbacks could enable true parallel execution in go-llmspell without the complexity of multiple Lua states.

## Overview

The spell shows three modes:
1. **Simple**: Basic async callback pattern
2. **Parallel**: Simulates parallel execution with callbacks
3. **Promise**: Integration with promise API

## Usage

```bash
# Simple async example
llmspell run async-callbacks --param mode=simple

# Parallel execution simulation
llmspell run async-callbacks --param mode=parallel

# Promise integration
llmspell run async-callbacks --param mode=promise
```

## How Async Callbacks Enable Parallelism

### Current Limitation
```lua
-- Current promises execute synchronously
local p1 = promise.new(function(resolve)
    local result = llm.chat("Question 1") -- Blocks here
    resolve(result)
end)

local p2 = promise.new(function(resolve)
    -- This doesn't start until p1's executor completes
    local result = llm.chat("Question 2")
    resolve(result)
end)
```

### With Async Callbacks
```lua
-- Async callbacks enable true parallelism
llm.chat_async("Question 1", function(result)
    print("Got result 1:", result)
end)

llm.chat_async("Question 2", function(result)
    print("Got result 2:", result)
end)

-- Both operations run concurrently in Go
-- Callbacks execute when results are ready
while llm.has_pending() do
    llm.process_callbacks()
end
```

## Implementation Architecture

### 1. Go Side (Bridge Layer)
```go
func (lb *LLMBridge) chatAsync(L *lua.LState) int {
    prompt := L.CheckString(1)
    callback := L.CheckFunction(2)
    
    // Start goroutine for async operation
    go func() {
        result, err := lb.actualLLMCall(prompt)
        lb.queueCallback(callback, result, err)
    }()
    
    return 0 // Non-blocking return
}
```

### 2. Callback Queue
- Goroutines queue results when complete
- Main thread processes queue safely
- No direct Lua calls from goroutines

### 3. Lua Event Loop
```lua
-- Main event loop pattern
while running do
    -- Process any ready callbacks
    llm.process_callbacks()
    
    -- Do other work
    update_ui()
    handle_input()
    
    -- Small delay to prevent busy waiting
    sleep(0.01)
end
```

## Benefits Over Multiple Lua States

1. **Simpler Implementation**
   - No need to manage Lua state pool
   - No complex state synchronization
   - No function/data serialization

2. **Better Resource Usage**
   - Single Lua state (less memory)
   - Shared module state
   - Easier debugging

3. **Familiar Pattern**
   - Similar to Node.js async model
   - Well-understood callback patterns
   - Easy to integrate with promises

## Comparison with Current Promises

| Feature | Current Promises | Async Callbacks |
|---------|-----------------|-----------------|
| Execution | Sequential | Parallel |
| Blocking | Yes (in executor) | No |
| Complexity | Low | Medium |
| True Parallelism | No | Yes |
| Thread Safety | N/A (single thread) | Managed by queue |

## Integration with Promises

Async callbacks can be wrapped to provide promise interface:

```lua
function llm.chat_promise(prompt)
    return promise.new(function(resolve, reject)
        local resolved = false
        
        llm.chat_async(prompt, 
            function(result)
                resolved = true
                resolve(result)
            end,
            function(err)
                resolved = true
                reject(err)
            end
        )
        
        -- Wait for callback
        while not resolved do
            llm.process_callbacks()
            coroutine.yield()
        end
    end)
end

-- Now can use promise.all for true parallel execution
local results = promise.all({
    llm.chat_promise("Question 1"),
    llm.chat_promise("Question 2"),
    llm.chat_promise("Question 3")
}):await()
```

## Implementation Steps

1. **Modify LLM Bridge**
   - Add async variants of methods
   - Implement callback queue
   - Add process_callbacks function

2. **Update Lua Runtime**
   - Add event loop support
   - Integrate callback processing
   - Handle coroutine yields

3. **Enhance Promise Implementation**
   - Support async operations
   - Integrate with callback system
   - Maintain backward compatibility

## Conclusion

Async callbacks provide a practical path to parallel execution in go-llmspell. They avoid the complexity of multiple Lua states while enabling true concurrent operations. This pattern is well-proven in other environments and would significantly enhance go-llmspell's capabilities for I/O-bound operations like LLM calls.