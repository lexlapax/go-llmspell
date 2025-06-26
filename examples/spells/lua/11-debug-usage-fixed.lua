-- ABOUTME: Fixed debug usage example using utils debug functions and logging for troubleshooting
-- ABOUTME: Demonstrates debugging patterns, logging levels, and performance monitoring

-- Debug Usage Example
-- This spell demonstrates debugging features available through utils and logging:
-- 1. Debug logging and levels
-- 2. Performance timing
-- 3. Error tracking and reporting
-- 4. Data inspection
-- 5. Debug configurations

-- Required modules
local utils = require("utils")
local logging = require("logging")
local llm = require("llm")
local agent = require("agent")
local data = require("data")
local errors = require("errors")

-- Parameters
local model = params and params.model or "gpt-4"

-- Get logger instance
local log = logging.default()

-- Example 1: Debug Logging Setup
print("=== Example 1: Debug Logging Setup ===")

-- Set debug level
local debug_level = "debug"
utils.debug_set_level(debug_level)
print("Debug level set to: " .. debug_level)

-- Get debug configuration
local debug_config = utils.debug_get_config()
print("Debug configuration: " .. data.to_json(debug_config))

-- Log at different levels
utils.debug_log("debug", "This is a debug message", {extra = "data"})
utils.debug_log("info", "This is an info message", {user = "test"})
utils.debug_log("warn", "This is a warning", {threshold = 0.8})
utils.debug_log("error", "This is an error", {code = "TEST_ERROR"})

-- Using logging module for structured logging
log:debug("Debug message from logging module")
log:info("Info message from logging module")
log:warn("Warning from logging module")
log:error("Error from logging module")
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

-- Create a function to profile
local function expensive_operation(n)
    local result = 0
    for i = 1, n do
        result = result + i
    end
    return result
end

-- Profile the function
print("\nProfiling expensive_operation:")
local sizes = {1000, 10000, 100000}
for _, size in ipairs(sizes) do
    local get_elapsed = timer()
    local result = expensive_operation(size)
    local elapsed = get_elapsed()
    
    utils.debug_log("info", string.format("Operation with n=%d took %.4f seconds", size, elapsed), {
        size = size,
        result = result,
        elapsed = elapsed
    })
end
print()

-- Example 3: Error Tracking
print("=== Example 3: Error Tracking ===")

-- Create custom error handler
local error_count = 0
local error_log = {}

local function track_error(err, context)
    error_count = error_count + 1
    local error_entry = {
        timestamp = os.time(),
        error = tostring(err),
        context = context,
        count = error_count
    }
    table.insert(error_log, error_entry)
    
    utils.debug_log("error", "Error tracked", error_entry)
    return error_entry
end

-- Test error tracking
local function risky_function(should_fail)
    if should_fail then
        error("Simulated error for debugging")
    end
    return "Success"
end

-- Try operations with error tracking
for i = 1, 3 do
    local success, result = pcall(risky_function, i == 2)
    if not success then
        track_error(result, {operation = "risky_function", iteration = i})
        print("  Error caught in iteration " .. i)
    else
        print("  Success in iteration " .. i)
    end
end

print("\nError summary:")
print("  Total errors: " .. error_count)
print("  Error log entries: " .. #error_log)
print()

-- Example 4: Agent and LLM Debugging
print("=== Example 4: Agent and LLM Debugging ===")

-- Create a debug-enabled agent
local success, debug_agent = pcall(agent.create, "Debug Assistant", {
    model = model,
    system = "You are a helpful assistant. Be concise.",
    temperature = 0.5
})

if not success then
    print("Failed to create agent: " .. tostring(debug_agent))
    -- Skip agent testing
else
    -- Function to debug agent calls
    local function debug_agent_call(prompt)
        utils.debug_log("debug", "Agent call starting", {prompt = prompt})
        
        local get_elapsed = timer()
        local response = debug_agent:run(prompt)
        local elapsed = get_elapsed()
        
        -- Log details
        local debug_info = {
            prompt = prompt,
            response_preview = string.sub(response, 1, 100),
            response_length = #response,
            elapsed_time = elapsed
        }
        
        utils.debug_log("info", "Agent call completed", debug_info)
        
        return response, debug_info
    end

    -- Test debug agent calls
    local test_prompts = {
        "What is 2+2?",
        "Write a haiku about debugging",
        "Explain recursion in one sentence"
    }

    print("\nTesting agent with debug logging:")
    for i, prompt in ipairs(test_prompts) do
        print("\nPrompt " .. i .. ": " .. prompt)
        local response, debug_info = debug_agent_call(prompt)
        print("Response: " .. response)
        print("Debug info: " .. data.to_json(debug_info))
    end
end
print()

-- Example 5: Data Inspection and Validation
print("=== Example 5: Data Inspection and Validation ===")

-- Debug data structures
local complex_data = {
    users = {
        {id = 1, name = "Alice", active = true},
        {id = 2, name = "Bob", active = false},
        {id = 3, name = "Charlie", active = true}
    },
    settings = {
        debug = true,
        log_level = "debug",
        features = {"auth", "api", "webhooks"}
    },
    metrics = {
        requests = 1523,
        errors = 23,
        success_rate = 0.985
    }
}

-- Debug inspection function
local function inspect_data(data_obj, path)
    path = path or "root"
    utils.debug_log("debug", "Inspecting data at: " .. path, {
        type = type(data_obj),
        size = type(data_obj) == "table" and #data_obj or nil
    })
    
    if type(data_obj) == "table" then
        for key, value in pairs(data_obj) do
            local new_path = path .. "." .. tostring(key)
            if type(value) == "table" then
                inspect_data(value, new_path)
            else
                utils.debug_log("debug", "  " .. new_path .. " = " .. tostring(value), {
                    type = type(value)
                })
            end
        end
    end
end

print("\nInspecting complex data structure:")
inspect_data(complex_data)

-- Data validation with debug output
local function validate_user(user)
    local validation_errors = {}
    
    if not user.id then
        table.insert(validation_errors, "Missing user ID")
    end
    
    if not user.name or #user.name < 2 then
        table.insert(validation_errors, "Invalid user name")
    end
    
    if type(user.active) ~= "boolean" then
        table.insert(validation_errors, "Active status must be boolean")
    end
    
    if #validation_errors > 0 then
        utils.debug_log("warn", "User validation failed", {
            user = user,
            errors = validation_errors
        })
        return false, validation_errors
    end
    
    utils.debug_log("info", "User validation passed", {user = user})
    return true
end

print("\nValidating users:")
for _, user in ipairs(complex_data.users) do
    local valid, errors = validate_user(user)
    print("  User " .. user.name .. ": " .. (valid and "Valid" or "Invalid"))
end
print()

-- Example 6: Debug Configuration Management
print("=== Example 6: Debug Configuration Management ===")

-- Check current debug settings
local current_config = utils.debug_get_config()
print("\nCurrent debug configuration:")
print(data.to_json(current_config, {pretty = true}))

-- Create debug session manager
local DebugSession = {}

function DebugSession:new()
    local obj = {
        start_time = os.time(),
        events = {},
        metrics = {
            agent_calls = 0,
            errors = 0,
            warnings = 0
        }
    }
    setmetatable(obj, {__index = self})
    return obj
end

function DebugSession:log_event(event_type, details)
    local event = {
        timestamp = os.time(),
        type = event_type,
        details = details
    }
    table.insert(self.events, event)
    
    -- Update metrics
    if event_type == "agent_call" then
        self.metrics.agent_calls = self.metrics.agent_calls + 1
    elseif event_type == "error" then
        self.metrics.errors = self.metrics.errors + 1
    elseif event_type == "warning" then
        self.metrics.warnings = self.metrics.warnings + 1
    end
    
    utils.debug_log("debug", "Session event", event)
end

function DebugSession:get_summary()
    local elapsed = os.time() - self.start_time
    return {
        duration = elapsed,
        total_events = #self.events,
        metrics = self.metrics
    }
end

-- Use debug session
local session = DebugSession:new()

-- Simulate some events
session:log_event("start", {user = "developer"})
session:log_event("agent_call", {prompt = "test prompt"})
session:log_event("warning", {message = "Rate limit approaching"})
session:log_event("agent_call", {prompt = "another test"})
session:log_event("end", {status = "success"})

print("\nDebug session summary:")
print(data.to_json(session:get_summary(), {pretty = true}))

-- Summary
print("\n=== Summary ===")
print("This example demonstrated:")
print("1. Debug logging setup and configuration")
print("2. Performance monitoring with timers")
print("3. Error tracking and reporting")
print("4. Agent/LLM call debugging")
print("5. Data inspection and validation")
print("6. Debug session management")
print()
print("Debug features used:")
print("- utils.debug_set_level()")
print("- utils.debug_log()")
print("- utils.debug_get_config()")
print("- logging module for structured logs")
print("- Custom error tracking")
print("- Performance profiling")

return {
    debug_examples = 6,
    error_count = error_count,
    features_demonstrated = {
        "logging_levels",
        "performance_monitoring",
        "error_tracking",
        "agent_debugging",
        "data_inspection",
        "session_management"
    }
}