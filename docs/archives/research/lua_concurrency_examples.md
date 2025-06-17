# Lua Concurrency Examples

## Basic Channel Operations

### Creating and Using Channels
```lua
-- Create channels
local ch = channel.make()      -- Unbuffered channel
local ch2 = channel.make(10)    -- Buffered channel with capacity 10

-- Send and receive
ch:send("hello")                -- Blocks until received
local msg = ch:receive()        -- Blocks until available
print(msg)                      -- "hello"

-- Non-blocking operations
local ok = ch:try_send("data")  -- Returns false if would block
local val, ok = ch:try_receive() -- Returns nil, false if empty

-- Close channel
ch:close()
local val, ok = ch:receive()    -- ok = false when closed
```

## Goroutine Spawning

### Basic Goroutine
```lua
-- Spawn a simple goroutine
go.spawn(function()
    print("Running in goroutine")
    channel.sleep(1)  -- Sleep 1 second
    print("Goroutine done")
end)

-- Spawn with arguments
go.spawn(function(name, count)
    for i = 1, count do
        print(name .. ": " .. i)
        channel.sleep(0.1)
    end
end, "worker", 5)
```

### Goroutine with Channels
```lua
-- Producer-consumer pattern
local data = channel.make(5)
local done = channel.make()

-- Producer
go.spawn(function()
    for i = 1, 10 do
        data:send(i)
        print("Produced:", i)
    end
    data:close()
end)

-- Consumer
go.spawn(function()
    for value in channel.range(data) do
        print("Consumed:", value)
        channel.sleep(0.2)  -- Simulate work
    end
    done:send(true)
end)

-- Wait for completion
done:receive()
print("All done!")
```

## Channel Select Operations

### Basic Select
```lua
local ch1 = channel.make()
local ch2 = channel.make()

-- Spawn senders
go.spawn(function()
    channel.sleep(1)
    ch1:send("from ch1")
end)

go.spawn(function()
    channel.sleep(2)
    ch2:send("from ch2")
end)

-- Select from multiple channels
for i = 1, 2 do
    local result = channel.select({
        {"|<-", ch1, function(val, ok)
            if ok then
                print("ch1:", val)
                return 1, val
            end
        end},
        {"|<-", ch2, function(val, ok)
            if ok then
                print("ch2:", val)
                return 2, val
            end
        end}
    })
    
    local case_num, value = result[1], result[2]
    print("Selected case", case_num, "with value", value)
end
```

### Select with Timeout
```lua
local ch = channel.make()

-- Try to receive with timeout
local result = channel.select({
    {"|<-", ch, function(val, ok)
        return "received", val
    end}
}, 1.0)  -- 1 second timeout

if result then
    print("Got:", result[2])
else
    print("Timeout!")
end
```

### Select with Default
```lua
local ch1 = channel.make(1)
local ch2 = channel.make(1)

-- Non-blocking select with default
local result = channel.select({
    {"<-|", ch1, "data1", function(ok)
        if ok then
            return "sent to ch1"
        end
    end},
    {"<-|", ch2, "data2", function(ok)
        if ok then
            return "sent to ch2"
        end
    end},
    {"default", function()
        return "nothing ready"
    end}
})

print("Result:", result[1])
```

## Advanced Patterns

### Worker Pool
```lua
local function create_worker_pool(num_workers, job_queue, result_queue)
    for i = 1, num_workers do
        go.spawn(function(id)
            print("Worker", id, "started")
            
            for job in channel.range(job_queue) do
                -- Process job
                local result = {
                    worker_id = id,
                    job_id = job.id,
                    result = job.data * 2  -- Simple processing
                }
                
                result_queue:send(result)
            end
            
            print("Worker", id, "finished")
        end, i)
    end
end

-- Usage
local jobs = channel.make(100)
local results = channel.make(100)

-- Create worker pool
create_worker_pool(5, jobs, results)

-- Submit jobs
go.spawn(function()
    for i = 1, 20 do
        jobs:send({id = i, data = i * 10})
    end
    jobs:close()
end)

-- Collect results
local count = 0
for result in channel.range(results) do
    print(string.format("Job %d processed by worker %d: %d", 
        result.job_id, result.worker_id, result.result))
    count = count + 1
    if count == 20 then
        break
    end
end
```

### Pipeline Processing
```lua
-- Create a data processing pipeline
local function pipeline()
    -- Stage channels
    local raw_data = channel.make(10)
    local parsed_data = channel.make(10)
    local validated_data = channel.make(10)
    local results = channel.make(10)
    
    -- Stage 1: Generate data
    go.spawn(function()
        for i = 1, 100 do
            raw_data:send({id = i, value = math.random(1, 100)})
        end
        raw_data:close()
    end)
    
    -- Stage 2: Parse/transform
    go.spawn(function()
        for data in channel.range(raw_data) do
            parsed_data:send({
                id = data.id,
                value = data.value,
                doubled = data.value * 2
            })
        end
        parsed_data:close()
    end)
    
    -- Stage 3: Validate
    go.spawn(function()
        for data in channel.range(parsed_data) do
            if data.doubled > 50 then  -- Simple validation
                validated_data:send(data)
            end
        end
        validated_data:close()
    end)
    
    -- Stage 4: Final processing
    go.spawn(function()
        for data in channel.range(validated_data) do
            results:send({
                id = data.id,
                final = data.doubled + 10
            })
        end
        results:close()
    end)
    
    return results
end

-- Run pipeline
local output = pipeline()
local count = 0
for result in channel.range(output) do
    print(string.format("Result %d: %d", result.id, result.final))
    count = count + 1
end
print("Processed", count, "items")
```

### Fan-Out/Fan-In
```lua
-- Fan-out: Distribute work to multiple goroutines
local function fan_out(input, num_workers)
    local workers = {}
    
    for i = 1, num_workers do
        local output = channel.make(10)
        workers[i] = output
        
        go.spawn(function(id, in_ch, out_ch)
            for item in channel.range(in_ch) do
                -- Process item
                local result = {
                    worker = id,
                    original = item,
                    processed = item * item  -- Square it
                }
                out_ch:send(result)
            end
            out_ch:close()
        end, i, input, output)
    end
    
    return workers
end

-- Fan-in: Merge multiple channels
local function fan_in(channels)
    local output = channel.make(10)
    local done_count = 0
    
    for i, ch in ipairs(channels) do
        go.spawn(function(input)
            for item in channel.range(input) do
                output:send(item)
            end
            done_count = done_count + 1
            if done_count == #channels then
                output:close()
            end
        end, ch)
    end
    
    return output
end

-- Usage
local source = channel.make(10)

-- Generate data
go.spawn(function()
    for i = 1, 20 do
        source:send(i)
    end
    source:close()
end)

-- Fan-out to 3 workers
local worker_outputs = fan_out(source, 3)

-- Fan-in results
local merged = fan_in(worker_outputs)

-- Collect all results
for result in channel.range(merged) do
    print(string.format("Worker %d processed %d -> %d", 
        result.worker, result.original, result.processed))
end
```

## Integration with Bridges

### Async Bridge Calls
```lua
-- Asynchronous LLM calls
local llm = require("llm")
local responses = channel.make(10)

-- Spawn multiple LLM requests
local prompts = {
    "What is the capital of France?",
    "Explain quantum computing",
    "Write a haiku about coding"
}

for i, prompt in ipairs(prompts) do
    go.spawn(function(id, p)
        local result = llm.generate({
            prompt = p,
            max_tokens = 100
        })
        
        responses:send({
            id = id,
            prompt = p,
            response = result
        })
    end, i, prompt)
end

-- Collect responses
for i = 1, #prompts do
    local result = responses:receive()
    print(string.format("Response %d:\nPrompt: %s\nAnswer: %s\n", 
        result.id, result.prompt, result.response))
end
```

### Parallel Tool Execution
```lua
-- Execute multiple tools in parallel
local tools = require("tools")
local results = channel.make()

local tool_calls = {
    {name = "web_search", params = {query = "Lua programming"}},
    {name = "calculator", params = {expression = "2 + 2"}},
    {name = "weather", params = {location = "San Francisco"}}
}

-- Execute tools concurrently
for _, call in ipairs(tool_calls) do
    go.spawn(function(tool_info)
        local ok, result = pcall(function()
            return tools.execute(tool_info.name, tool_info.params)
        end)
        
        results:send({
            tool = tool_info.name,
            success = ok,
            result = ok and result or tostring(result)
        })
    end, call)
end

-- Collect results
for i = 1, #tool_calls do
    local result = results:receive()
    print(string.format("Tool: %s\nSuccess: %s\nResult: %s\n",
        result.tool, tostring(result.success), result.result))
end
```

## Error Handling in Concurrent Code

### Goroutine Error Collection
```lua
local errors = channel.make(10)
local done = channel.make()

-- Worker that might fail
local function risky_worker(id, error_ch)
    return function()
        local ok, err = pcall(function()
            if math.random() > 0.7 then
                error("Worker " .. id .. " failed!")
            end
            -- Do work
            channel.sleep(math.random())
            print("Worker", id, "completed successfully")
        end)
        
        if not ok then
            error_ch:send({worker = id, error = err})
        end
    end
end

-- Spawn workers
for i = 1, 5 do
    go.spawn(risky_worker(i, errors))
end

-- Error collector
go.spawn(function()
    channel.sleep(2)  -- Wait for workers
    
    -- Collect any errors
    local error_count = 0
    while true do
        local err, ok = errors:try_receive()
        if not ok then
            break
        end
        print("Error from worker", err.worker, ":", err.error)
        error_count = error_count + 1
    end
    
    done:send(error_count)
end)

local total_errors = done:receive()
print("Total errors:", total_errors)
```

## Best Practices

### 1. Channel Cleanup
```lua
-- Always close channels when done
local ch = channel.make()

go.spawn(function()
    defer(function() ch:close() end)  -- Ensure cleanup
    
    for i = 1, 10 do
        ch:send(i)
    end
end)
```

### 2. Timeout Protection
```lua
-- Protect against deadlocks with timeouts
local function safe_receive(ch, timeout)
    local result = channel.select({
        {"|<-", ch, function(val, ok)
            return val, ok
        end}
    }, timeout or 5.0)
    
    if result then
        return result[1], result[2]
    else
        return nil, false, "timeout"
    end
end
```

### 3. Graceful Shutdown
```lua
-- Implement graceful shutdown pattern
local shutdown = channel.make()
local workers_done = channel.make()

-- Worker with shutdown support
go.spawn(function()
    while true do
        local result = channel.select({
            {"|<-", work_queue, function(job, ok)
                if ok then
                    process_job(job)
                    return true
                end
            end},
            {"|<-", shutdown, function()
                print("Worker shutting down")
                return false
            end}
        })
        
        if not result[1] then
            break
        end
    end
    
    workers_done:send(true)
end)

-- Trigger shutdown
shutdown:close()
workers_done:receive()
print("Shutdown complete")
```