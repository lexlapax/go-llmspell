-- ABOUTME: Debug usage example that works with or without debug bridge
-- ABOUTME: Demonstrates debugging patterns using available features

-- Debug Usage Example
-- This spell demonstrates debugging features:
-- 1. Safe debug logging
-- 2. Performance monitoring
-- 3. Error tracking
-- 4. Agent debugging

-- Required modules
local utils = require("utils")
local logging = require("logging")
local llm = require("llm")
local agent = require("agent")
local data = require("data")

-- Parameters
local model = params and params.model or "gpt-4"

-- Get logger instance
local log = logging.default()

-- Safe wrapper for debug functions
local function safe_debug_log(level, message, data)
    -- Try to use utils.debug_log if available
    local success, err = pcall(function()
        utils.debug_log(level, message, data)
    end)
    
    if not success then
        -- Fallback to logging module
        if level == "error" then
            log:error(message)
        elseif level == "warn" then
            log:warn(message)
        elseif level == "info" then
            log:info(message)
        else
            log:debug(message)
        end
    end
end

-- Example 1: Debug Configuration
print("=== Example 1: Debug Configuration ===")

-- Try to set debug level
local success, err = pcall(function()
    utils.debug_set_level("debug")
end)

if success then
    print("Debug level set to: debug")
else
    print("Note: Debug bridge not available, using logging module instead")
end

-- Try to get debug config
local debug_config = nil
success, debug_config = pcall(function()
    return utils.debug_get_config()
end)

if success and debug_config then
    print("Debug configuration: " .. data.to_json(debug_config))
else
    print("Using default logging configuration")
end

print()

-- Example 2: Performance Monitoring
print("=== Example 2: Performance Monitoring ===")

-- Simple timer utility
local function timer()
    local start_time = os.clock()
    return function()
        return os.clock() - start_time
    end
end

-- Profile a function
local function expensive_operation(n)
    local result = 0
    for i = 1, n do
        result = result + i
    end
    return result
end

print("\nProfiling operations:")
local sizes = {1000, 10000, 100000}
for _, size in ipairs(sizes) do
    local get_elapsed = timer()
    local result = expensive_operation(size)
    local elapsed = get_elapsed()
    
    print(string.format("  n=%d: %.4f seconds (result=%d)", size, elapsed, result))
    
    safe_debug_log("info", "Performance test completed", {
        size = size,
        elapsed = elapsed,
        result = result
    })
end

print()

-- Example 3: Error Tracking
print("=== Example 3: Error Tracking ===")

local error_count = 0
local error_log = {}

local function track_error(err, context)
    error_count = error_count + 1
    local error_entry = {
        timestamp = os.time(),
        error = tostring(err),
        context = context
    }
    table.insert(error_log, error_entry)
    
    safe_debug_log("error", "Error tracked: " .. tostring(err), context)
    
    return error_entry
end

-- Test error handling
for i = 1, 3 do
    local function test_func()
        if i == 2 then
            error("Test error #" .. i)
        end
        return "Success #" .. i
    end
    
    local ok, result = pcall(test_func)
    if not ok then
        track_error(result, {iteration = i})
        print("  Iteration " .. i .. ": Failed")
    else
        print("  Iteration " .. i .. ": " .. result)
    end
end

print("\nError summary:")
print("  Total errors: " .. error_count)
print("  Errors logged: " .. #error_log)

print()

-- Example 4: Agent Debugging
print("=== Example 4: Agent Debugging ===")

-- Create agent with error handling
local agent_ok, debug_agent = pcall(function()
    return agent.create("Debug Assistant", {
        model = model,
        system = "You are a helpful assistant. Be very concise - one sentence max.",
        temperature = 0.7
    })
end)

if not agent_ok then
    print("Failed to create agent: " .. tostring(debug_agent))
    print("Skipping agent tests")
else
    print("Agent created successfully")
    
    -- Test prompts
    local prompts = {
        "What is 2+2?",
        "Name a color."
    }
    
    for i, prompt in ipairs(prompts) do
        print("\nPrompt " .. i .. ": " .. prompt)
        
        local get_elapsed = timer()
        safe_debug_log("debug", "Starting agent call", {prompt = prompt})
        
        local ok, response = pcall(function()
            return debug_agent:run(prompt)
        end)
        
        local elapsed = get_elapsed()
        
        if ok then
            print("Response: " .. response)
            print("Time: " .. string.format("%.2f seconds", elapsed))
            
            safe_debug_log("info", "Agent call completed", {
                prompt = prompt,
                response_length = #response,
                elapsed = elapsed
            })
        else
            print("Error: " .. tostring(response))
            track_error(response, {prompt = prompt})
        end
    end
end

print()

-- Example 5: Debug Utilities
print("=== Example 5: Debug Utilities ===")

-- Custom debug helper
local function debug_value(name, value)
    local value_type = type(value)
    local value_str = tostring(value)
    
    if value_type == "table" then
        value_str = data.to_json(value)
    elseif value_type == "string" then
        value_str = string.format('"%s"', value)
    end
    
    print(string.format("  %s (%s): %s", name, value_type, value_str))
    
    safe_debug_log("debug", "Debug value inspection", {
        name = name,
        type = value_type,
        value = value
    })
end

print("\nInspecting values:")
debug_value("simple_string", "Hello, World!")
debug_value("number", 42)
debug_value("boolean", true)
debug_value("table", {a = 1, b = 2, c = {nested = true}})
debug_value("nil_value", nil)

print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Safe debug logging with fallbacks")
print("2. Performance monitoring with timers")
print("3. Error tracking and recovery")
print("4. Agent debugging with timing")
print("5. Value inspection utilities")
print()

return {
    examples_completed = 5,
    errors_tracked = error_count,
    debug_bridge_available = success
}