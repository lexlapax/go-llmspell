-- ABOUTME: Simplified debug example using logging module
-- ABOUTME: Demonstrates basic debugging patterns without file operations

-- Debug Usage Example (Simplified)
-- This spell demonstrates basic debugging:
-- 1. Structured logging
-- 2. Performance timing
-- 3. Error tracking

-- Required modules
local logging = require("logging")
local agent = require("agent")
local data = require("data")

-- Parameters
local model = params and params.model or "gpt-4"

-- Get logger
local log = logging.default()

print("=== Debug Usage Example (Simplified) ===")
print()

-- Example 1: Structured Logging
print("=== Example 1: Structured Logging ===")

-- Log at different levels
log:debug("This is a debug message")
log:info("This is an info message")
log:warn("This is a warning message")
log:error("This is an error message")

-- Log with attributes
log:info("User action", {
    action = "login",
    user_id = 123,
    timestamp = os.time()
})

print("Logged messages at different levels")
print()

-- Example 2: Performance Timing
print("=== Example 2: Performance Timing ===")

-- Simple timer
local function timer()
    local start = os.clock()
    return function()
        return os.clock() - start
    end
end

-- Test function
local function calculate_sum(n)
    local sum = 0
    for i = 1, n do
        sum = sum + i
    end
    return sum
end

-- Time the function
print("\nTiming calculations:")
local sizes = {1000, 10000, 100000}
for _, size in ipairs(sizes) do
    local elapsed_timer = timer()
    local result = calculate_sum(size)
    local elapsed = elapsed_timer()
    
    print(string.format("  n=%d: %.4f seconds (sum=%d)", size, elapsed, result))
    
    log:info("Performance test", {
        size = size,
        elapsed = elapsed,
        result = result
    })
end

print()

-- Example 3: Error Tracking
print("=== Example 3: Error Tracking ===")

local errors_found = {}

-- Function that might error
local function risky_operation(value)
    if type(value) ~= "number" then
        error("Expected number, got " .. type(value))
    end
    if value < 0 then
        error("Value must be non-negative")
    end
    return math.sqrt(value)
end

-- Test with different inputs
local test_values = {4, "hello", -1, 16}
print("\nTesting error handling:")

for _, value in ipairs(test_values) do
    local ok, result = pcall(risky_operation, value)
    
    if ok then
        print("  " .. tostring(value) .. " -> Success: " .. result)
        log:info("Operation succeeded", {
            input = value,
            output = result
        })
    else
        print("  " .. tostring(value) .. " -> Error: " .. result)
        table.insert(errors_found, {
            input = value,
            error = result
        })
        log:error("Operation failed", {
            input = value,
            error = result
        })
    end
end

print("\nTotal errors: " .. #errors_found)
print()

-- Example 4: Agent Debugging
print("=== Example 4: Agent Debugging ===")

-- Create agent with error handling
local ok, agent_result = pcall(function()
    return agent.create("Debug Assistant", {
        model = model,
        system = "You are a helpful assistant. Answer in one sentence.",
        temperature = 0.7
    })
end)

if not ok then
    print("Failed to create agent: " .. tostring(agent_result))
    log:error("Agent creation failed", {error = tostring(agent_result)})
else
    local my_agent = agent_result
    print("Agent created successfully")
    
    -- Test the agent
    local prompt = "What is the capital of France?"
    print("\nPrompt: " .. prompt)
    
    local start_time = timer()
    local ok2, response = pcall(function()
        return my_agent:run(prompt)
    end)
    local elapsed = start_time()
    
    if ok2 then
        print("Response: " .. response)
        print("Time: " .. string.format("%.2f seconds", elapsed))
        
        log:info("Agent call completed", {
            prompt = prompt,
            response = response,
            elapsed = elapsed
        })
    else
        print("Agent call failed: " .. tostring(response))
        log:error("Agent call failed", {
            prompt = prompt,
            error = tostring(response)
        })
    end
end

print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Structured logging with the logging module")
print("2. Performance timing with simple timers")
print("3. Error tracking and recovery")
print("4. Agent debugging with timing")
print()

return {
    examples = 4,
    errors_found = #errors_found,
    success = true
}