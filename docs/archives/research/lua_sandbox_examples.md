# Lua Sandbox Usage Examples

## Basic Sandbox Setup

### Creating a Minimal Sandbox
```go
// Create the most restrictive sandbox
sandbox := &LuaSandbox{
    config: SandboxConfig{
        AllowedLibraries: []string{"base", "table", "string", "math"},
        MaxInstructions:  1000000,  // 1M instructions
        MaxMemory:        10485760, // 10MB
        MaxExecutionTime: 5 * time.Second,
        DisableBytecode:  true,
        IsolateGlobals:   true,
    },
}

// Create sandboxed state
L, err := sandbox.CreateSafeState()
if err != nil {
    log.Fatal(err)
}
defer L.Close()

// Execute safe script
err = sandbox.ExecuteWithTimeout(L, `
    local function factorial(n)
        if n <= 1 then return 1 end
        return n * factorial(n - 1)
    end
    
    print("Factorial of 5:", factorial(5))
`, 1*time.Second)
```

### Sandbox with Bridge Access
```go
// Configure sandbox with bridge access
sandbox := &LuaSandbox{
    config: SandboxConfig{
        AllowedLibraries: []string{"base", "table", "string", "math", "coroutine"},
        MaxInstructions:  10000000,  // 10M for more complex operations
        MaxMemory:        52428800,  // 50MB for data processing
        MaxExecutionTime: 30 * time.Second,
        EnableAuditing:   true,
    },
}

L, _ := sandbox.CreateSafeState()

// Register safe bridges
llmBridge := &LLMBridge{/* ... */}
sandbox.RegisterSafeBridge(L, llmBridge)

// Script can now use LLM safely
sandbox.Execute(L, `
    local llm = llm  -- Global bridge access
    
    local response = llm.generate({
        prompt = "What is the capital of France?",
        max_tokens = 50
    })
    
    print("Response:", response)
`)
```

## Security Examples

### Preventing File System Access
```lua
-- These will all fail in sandbox:

-- Attempt 1: Direct file access
local file = io.open("/etc/passwd", "r")  -- Error: io is nil

-- Attempt 2: Load file
dofile("/etc/passwd")  -- Error: dofile is nil

-- Attempt 3: Require
require("os")  -- Error: require is nil

-- Attempt 4: Load code
load('os.execute("rm -rf /")')  -- Error: load is nil
```

### Preventing System Command Execution
```lua
-- All system access attempts fail:

-- Direct OS access
os.execute("whoami")  -- Error: os is nil

-- Try to get os through _G
_G["os"] = {execute = function() end}  -- Works, but it's just a table
_G.os.execute("whoami")  -- Does nothing (our dummy function)

-- Try to use debug to escape
debug.getinfo(1)  -- Error: debug is nil
```

### Safe String Operations
```lua
-- These string operations are allowed:
local s = "Hello, World!"
print(string.upper(s))  -- "HELLO, WORLD!"
print(string.sub(s, 1, 5))  -- "Hello"
print(string.format("Number: %d", 42))  -- "Number: 42"

-- But pattern matching is limited:
-- Complex patterns that could cause DoS are prevented
local huge = string.rep("a", 1000000)
-- This would timeout due to execution limits:
-- string.match(huge, "a*a*a*a*a*a*")
```

## Resource Limit Examples

### Instruction Count Limits
```lua
-- This will hit instruction limit:
local count = 0
while true do
    count = count + 1
    -- After ~1M instructions, raises error
end

-- Safe version with explicit limit:
local count = 0
for i = 1, 100000 do
    count = count + 1
end
print("Count:", count)  -- Works fine
```

### Memory Limits
```lua
-- This will hit memory limit:
local huge_table = {}
for i = 1, 10000000 do
    huge_table[i] = {data = string.rep("x", 1000)}
    -- Error: memory limit exceeded
end

-- Safe version:
local data = {}
for i = 1, 1000 do
    data[i] = i * 2  -- Small data
end
```

### Execution Timeout
```lua
-- This will timeout:
function slowFunction()
    local result = 0
    for i = 1, 1000000000 do
        result = result + math.sqrt(i)
    end
    return result
end

slowFunction()  -- Error: execution timeout

-- Safe version with yielding:
function safeSlowFunction()
    local result = 0
    for i = 1, 1000000 do
        result = result + math.sqrt(i)
        if i % 10000 == 0 then
            coroutine.yield()  -- Allow timeout check
        end
    end
    return result
end
```

## Bridge Integration Examples

### Safe LLM Usage
```lua
-- Sandboxed script using LLM bridge
local llm = llm  -- Pre-registered bridge

-- Safe text generation
local story = llm.generate({
    prompt = "Write a short story about a robot",
    max_tokens = 200,
    temperature = 0.8
})

-- Process the result safely
if story and type(story) == "string" then
    -- Limited string operations
    local words = {}
    for word in string.gmatch(story, "%S+") do
        table.insert(words, word)
    end
    print("Word count:", #words)
end
```

### Safe Tool Usage
```lua
-- Using tools bridge in sandbox
local tools = tools

-- Calculator is safe (no system access)
local result = tools.calculator({
    expression = "2 + 2 * 3"
})
print("Result:", result)  -- 8

-- Web search is controlled by bridge
local results = tools.web_search({
    query = "Lua programming",
    limit = 5
})

-- Process results safely
if results and type(results) == "table" then
    for i, result in ipairs(results) do
        print(i, result.title)
        -- URLs are just strings, can't fetch in sandbox
    end
end
```

### Safe State Management
```lua
-- State bridge provides controlled persistence
local state = state

-- Save data (controlled by bridge)
state.set("user_score", 100)
state.set("user_name", "Alice")

-- Retrieve data
local score = state.get("user_score")
local name = state.get("user_name")

-- Safe data processing
if score and type(score) == "number" then
    score = score + 10
    state.set("user_score", score)
end

-- List keys (controlled enumeration)
local keys = state.list()
for _, key in ipairs(keys) do
    print("Key:", key)
end
```

## Sandbox Configuration Examples

### Development Sandbox (More Permissive)
```go
devSandbox := &LuaSandbox{
    config: SandboxConfig{
        AllowedLibraries: []string{
            "base", "table", "string", "math", 
            "coroutine", // For async
        },
        MaxInstructions:  100000000,    // 100M
        MaxMemory:        104857600,    // 100MB
        MaxExecutionTime: 60 * time.Second,
        MaxStringLength:  1048576,       // 1MB strings
        EnableProfiling:  true,          // Performance monitoring
        EnableAuditing:   true,          // Full logging
    },
}
```

### Production Sandbox (Restrictive)
```go
prodSandbox := &LuaSandbox{
    config: SandboxConfig{
        AllowedLibraries: []string{"base", "table", "math"},
        BlockedFunctions: map[string]bool{
            "collectgarbage": true,  // Prevent GC manipulation
        },
        MaxInstructions:  10000000,     // 10M
        MaxMemory:        20971520,     // 20MB
        MaxExecutionTime: 10 * time.Second,
        MaxStringLength:  65536,        // 64KB strings
        DisableBytecode:  true,
        IsolateGlobals:   true,
        EnableAuditing:   true,
    },
}
```

### Minimal Computation Sandbox
```go
// For simple calculations only
calcSandbox := &LuaSandbox{
    config: SandboxConfig{
        AllowedLibraries: []string{"math"}, // Math only
        WhitelistFunctions: map[string]bool{
            "type":     true,
            "tonumber": true,
            "tostring": true,
            "assert":   true,
            "error":    true,
        },
        MaxInstructions:  100000,      // 100K
        MaxMemory:        1048576,     // 1MB
        MaxExecutionTime: 1 * time.Second,
    },
}
```

## Error Handling in Sandbox

### Catching Security Violations
```go
// Execute with violation tracking
violations, err := sandbox.ExecuteAndTrack(L, `
    -- Attempt various violations
    local attempts = {
        function() return os.execute("ls") end,
        function() return io.open("file.txt") end,
        function() return require("socket") end,
        function() return load("return 1")() end,
    }
    
    for i, attempt in ipairs(attempts) do
        local ok, err = pcall(attempt)
        if not ok then
            print("Attempt", i, "blocked:", err)
        end
    end
`)

// Check violations
for _, v := range violations {
    log.Printf("Violation: %s - %s", v.Type, v.Message)
}
```

### Graceful Error Recovery
```lua
-- Safe error handling pattern
local function safeOperation()
    local ok, result = pcall(function()
        -- Potentially dangerous operation
        return someFunction()
    end)
    
    if ok then
        return result
    else
        -- Log error safely
        print("Operation failed:", tostring(result))
        return nil, "operation failed"
    end
end

-- Multiple attempts with timeout
local function retryWithTimeout(func, maxAttempts)
    for i = 1, maxAttempts do
        local ok, result = pcall(func)
        if ok then
            return result
        end
        -- No sleep available in sandbox!
        -- Just continue to next attempt
    end
    return nil, "all attempts failed"
end
```

## Performance Monitoring

### Tracking Resource Usage
```go
// Monitor sandbox performance
metrics := sandbox.Execute(L, script)

fmt.Printf("Execution metrics:\n")
fmt.Printf("  Instructions: %d\n", metrics.InstructionCount)
fmt.Printf("  Memory used: %d bytes\n", metrics.MemoryUsed)
fmt.Printf("  Execution time: %v\n", metrics.ExecutionTime)
fmt.Printf("  Table allocations: %d\n", metrics.TableAllocations)
fmt.Printf("  String allocations: %d\n", metrics.StringAllocations)
```

### Profiling Safe Scripts
```lua
-- Built-in safe profiler
local profiler = sandbox.profiler

profiler.start()

-- Code to profile
for i = 1, 10000 do
    local t = {}
    for j = 1, 100 do
        t[j] = j * 2
    end
end

local report = profiler.stop()
print("Profile report:")
print("  Total time:", report.total_time)
print("  Allocations:", report.allocations)
print("  Peak memory:", report.peak_memory)
```