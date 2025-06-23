-- ABOUTME: Debug usage example showing debugging tools, profiling, and troubleshooting
-- ABOUTME: Demonstrates debug module usage for development, testing, and performance analysis

-- Debug Usage Example
-- This spell demonstrates comprehensive debugging features including:
-- 1. Debug logging and tracing
-- 2. Performance profiling
-- 3. Memory usage tracking
-- 4. Call stack inspection
-- 5. Variable watching and breakpoints

-- Required modules
local debug = require("debug")
local log = require("log")
local llm = require("llm")
local agent = require("agent")
local core = require("core")
local data = require("data")
local errors = require("errors")

-- Enable debug mode
debug.enable()
log.set_level("debug")

-- Example 1: Debug Logging and Tracing
print("=== Example 1: Debug Logging and Tracing ===")

-- Create a traced function
local function calculate_fibonacci(n)
    debug.trace("calculate_fibonacci", {n = n})
    
    if n <= 1 then
        debug.trace("fibonacci_base_case", {n = n, result = n})
        return n
    end
    
    local result = calculate_fibonacci(n - 1) + calculate_fibonacci(n - 2)
    debug.trace("fibonacci_recursive", {n = n, result = result})
    
    return result
end

-- Enable function tracing
debug.start_trace()

-- Calculate with tracing
local fib_result = calculate_fibonacci(5)
print("Fibonacci(5) = " .. fib_result)

-- Get trace results
local trace_data = debug.stop_trace()
print("Function calls traced: " .. #trace_data)
print("First few traces:")
for i = 1, math.min(3, #trace_data) do
    local trace = trace_data[i]
    print(string.format("  %s: %s", trace.name, data.to_json(trace.args)))
end
print()

-- Example 2: Performance Profiling
print("=== Example 2: Performance Profiling ===")

-- Start profiling
debug.start_profile()

-- Expensive operation 1: String manipulation
local function expensive_string_operation()
    local result = ""
    for i = 1, 1000 do
        result = result .. "x"  -- Inefficient concatenation
    end
    return result
end

-- Expensive operation 2: Table operations
local function expensive_table_operation()
    local t = {}
    for i = 1, 1000 do
        table.insert(t, i)
        table.sort(t)  -- Sorting after each insert (inefficient)
    end
    return t
end

-- Expensive operation 3: LLM call
local function expensive_llm_operation()
    return llm.complete({
        model = "gpt-3.5-turbo",
        messages = {
            {role = "user", content = "Count from 1 to 10"}
        },
        max_tokens = 50
    })
end

-- Run operations
print("Running expensive operations...")
expensive_string_operation()
expensive_table_operation()
expensive_llm_operation()

-- Get profiling results
local profile = debug.stop_profile()
print("\nProfile Results:")
print(string.format("Total time: %.3fs", profile.total_time))
print(string.format("CPU time: %.3fs", profile.cpu_time))
print(string.format("Memory allocated: %.2f MB", profile.memory_allocated / 1024 / 1024))

print("\nTop functions by time:")
for i, func in ipairs(profile.functions) do
    if i > 5 then break end
    print(string.format("  %s: %.3fs (%.1f%%)", 
        func.name, 
        func.time, 
        func.time / profile.total_time * 100))
end
print()

-- Example 3: Memory Usage Tracking
print("=== Example 3: Memory Usage Tracking ===")

-- Get initial memory snapshot
local mem_start = debug.memory_snapshot()

-- Create some objects
local large_table = {}
for i = 1, 10000 do
    large_table[i] = {
        id = i,
        data = string.rep("x", 100),
        nested = {a = 1, b = 2, c = 3}
    }
end

-- Verify table was created
print(string.format("Created large table with %d entries", #large_table))

-- Create an agent (holds state)
local memory_test_agent = agent.create({
    name = "Memory Test Agent",
    model = "gpt-3.5-turbo",
    instructions = "You are testing memory usage."
})

-- Get memory snapshot after allocations
local mem_after = debug.memory_snapshot()

-- Calculate differences
local mem_diff = debug.memory_diff(mem_start, mem_after)
print(string.format("Memory allocated: %.2f MB", mem_diff.allocated / 1024 / 1024))
print(string.format("Tables created: %d", mem_diff.tables))
print(string.format("Strings created: %d", mem_diff.strings))

-- Force garbage collection
collectgarbage("collect")
local mem_gc = debug.memory_snapshot()
local gc_diff = debug.memory_diff(mem_after, mem_gc)
print(string.format("Memory freed by GC: %.2f MB", -gc_diff.allocated / 1024 / 1024))
print()

-- Example 4: Call Stack Inspection
print("=== Example 4: Call Stack Inspection ===")

local function level3()
    -- Inspect call stack
    local stack = debug.get_stack(10)  -- Get up to 10 frames
    
    print("Current call stack:")
    for i, frame in ipairs(stack) do
        print(string.format("  %d: %s at %s:%d", 
            i, 
            frame.function_name or "anonymous",
            frame.source or "unknown",
            frame.line or 0))
        
        -- Show local variables for top frame
        if i == 1 and frame.locals then
            print("    Locals:")
            for name, value in pairs(frame.locals) do
                print(string.format("      %s = %s", name, tostring(value)))
            end
        end
    end
    
    return "level3_result"
end

local function level2(param)
    local local_var = "level2_local"
    return level3()
end

local function level1()
    return level2("test_param")
end

-- Call through multiple levels
level1()
print()

-- Example 5: Variable Watching and Breakpoints
print("=== Example 5: Variable Watching and Breakpoints ===")

-- Set up variable watchers
local watch_values = {}
debug.watch("counter", function(name, old_value, new_value)
    print(string.format("Watch: %s changed from %s to %s", 
        name, tostring(old_value), tostring(new_value)))
    table.insert(watch_values, {name = name, old = old_value, new = new_value})
end)

-- Function with conditional breakpoint
local function process_items(items)
    local counter = 0
    local results = {}
    
    for i, item in ipairs(items) do
        counter = counter + 1
        debug.set_watched("counter", counter)
        
        -- Conditional breakpoint
        if debug.breakpoint_enabled and item.value > 50 then
            debug.breakpoint({
                message = "High value item found",
                item = item,
                counter = counter
            })
        end
        
        -- Process item
        local result = item.value * 2
        table.insert(results, result)
        
        -- Simulate some work
        core.sleep(0.01)
    end
    
    return results
end

-- Enable breakpoints
debug.breakpoint_enabled = true
debug.on_breakpoint = function(info)
    print("BREAKPOINT HIT: " .. info.message)
    print("  Item: " .. data.to_json(info.item))
    print("  Counter: " .. info.counter)
    -- In a real debugger, execution would pause here
end

-- Process items with debugging
local test_items = {
    {id = 1, value = 10},
    {id = 2, value = 25},
    {id = 3, value = 75},  -- Will trigger breakpoint
    {id = 4, value = 30}
}

local results = process_items(test_items)
print("Processed " .. #results .. " items")
print("Watch triggers: " .. #watch_values)
print()

-- Example 6: Error Debugging and Recovery
print("=== Example 6: Error Debugging and Recovery ===")

-- Enable detailed error tracking
debug.track_errors = true

-- Function that may fail
local function risky_llm_call(prompt, should_fail)
    debug.trace("risky_llm_call_start", {prompt = prompt, should_fail = should_fail})
    
    if should_fail then
        -- Simulate an error with full context
        local err = errors.new("LLM_ERROR", "Simulated API failure")
        err:with_context("prompt", prompt)
        err:with_context("timestamp", os.time())
        err:with_context("stack_trace", debug.get_stack(5))
        
        debug.trace("risky_llm_call_error", {error = err})
        error(err)
    end
    
    local result = llm.complete({
        model = "gpt-3.5-turbo",
        messages = {{role = "user", content = prompt}},
        max_tokens = 50
    })
    
    debug.trace("risky_llm_call_success", {result_length = #result.content})
    return result
end

-- Try with error capture
local success, result = pcall(function()
    return risky_llm_call("This will fail", true)
end)

if not success then
    print("Error caught: " .. tostring(result))
    
    -- Get detailed error info
    local error_info = debug.get_last_error()
    if error_info then
        print("Error details:")
        print("  Type: " .. (error_info.type or "unknown"))
        print("  Message: " .. (error_info.message or "none"))
        print("  Location: " .. (error_info.location or "unknown"))
        
        if error_info.context then
            print("  Context:")
            for k, v in pairs(error_info.context) do
                if k ~= "stack_trace" then  -- Skip large stack trace
                    print(string.format("    %s: %s", k, tostring(v)))
                end
            end
        end
    end
end
print()

-- Example 7: Debug Report Generation
print("=== Example 7: Debug Report Generation ===")

-- Generate comprehensive debug report
local report = debug.generate_report({
    include_profile = true,
    include_memory = true,
    include_traces = true,
    include_errors = true,
    include_config = true
})

print("Debug Report Summary:")
print("  Script: " .. (report.script_name or "unknown"))
print("  Runtime: " .. string.format("%.2fs", report.runtime or 0))
print("  Memory peak: " .. string.format("%.2f MB", (report.memory_peak or 0) / 1024 / 1024))
print("  Errors encountered: " .. (report.error_count or 0))
print("  Functions traced: " .. (report.trace_count or 0))

-- Save detailed report
local report_json = data.to_json(report, {pretty = true})
print("\nDetailed report size: " .. #report_json .. " bytes")

-- Disable debug mode
debug.disable()
print("\nDebug mode disabled")

-- Return debug summary
return {
    success = true,
    debug_enabled = debug.is_enabled(),
    traces_collected = #(trace_data or {}),
    errors_tracked = report.error_count or 0,
    memory_peak_mb = (report.memory_peak or 0) / 1024 / 1024,
    profiling_completed = profile ~= nil
}