-- ABOUTME: Simplified debug example using available logging and monitoring capabilities
-- ABOUTME: Demonstrates debugging patterns without the debug module

-- Debug Usage Example (Simplified)
-- This spell demonstrates debugging patterns:
-- 1. Structured logging with the log module
-- 2. Performance timing and profiling
-- 3. Error tracking and reporting
-- 4. Custom debug utilities

-- Required modules
local log = require("log")
local llm = require("llm")
local agent = require("agent")
local data = require("data")
local errors = require("errors")
local utils = require("utils")

-- Parameters
local model = params.model or "gpt-4"
local output_dir = params.output_dir or "./debug-output"

print("=== Debug Usage Example (Simplified) ===")
print()

-- Ensure output directory exists
if not utils.file_exists(output_dir) then
    utils.mkdir(output_dir)
end

-- Set log level
log.set_level("debug")

-- Custom debug utilities
local Debug = {}

function Debug.new()
    return {
        traces = {},
        timers = {},
        counters = {},
        errors = {}
    }
end

function Debug.trace(debug, name, data)
    local trace = {
        name = name,
        data = data,
        timestamp = os.time(),
        time_ms = os.clock() * 1000
    }
    debug.traces[#debug.traces + 1] = trace
    
    log.debug(string.format("TRACE: %s | %s", name, data and data.to_json and data:to_json() or tostring(data)))
end

function Debug.start_timer(debug, name)
    debug.timers[name] = {
        start = os.clock(),
        count = (debug.timers[name] and debug.timers[name].count or 0) + 1
    }
    log.debug("TIMER START: " .. name)
end

function Debug.stop_timer(debug, name)
    if debug.timers[name] and debug.timers[name].start then
        local elapsed = os.clock() - debug.timers[name].start
        debug.timers[name].last_duration = elapsed
        debug.timers[name].total_duration = (debug.timers[name].total_duration or 0) + elapsed
        debug.timers[name].start = nil
        
        log.debug(string.format("TIMER STOP: %s | Duration: %.3fs", name, elapsed))
        return elapsed
    end
    return 0
end

function Debug.increment(debug, counter, value)
    value = value or 1
    debug.counters[counter] = (debug.counters[counter] or 0) + value
    log.debug(string.format("COUNTER: %s = %d", counter, debug.counters[counter]))
end

function Debug.record_error(debug, error_type, message, context)
    local error_record = {
        type = error_type,
        message = message,
        context = context,
        timestamp = os.time()
    }
    debug.errors[#debug.errors + 1] = error_record
    
    log.error(string.format("ERROR: [%s] %s", error_type, message))
    if context then
        log.debug("ERROR CONTEXT: " .. (type(context) == "table" and data.to_json(context) or tostring(context)))
    end
end

-- Create debug instance
local debugger = Debug.new()

-- Example 1: Function Tracing
print("=== Example 1: Function Tracing ===")

-- Traced fibonacci function
local function calculate_fibonacci(n, depth)
    depth = depth or 0
    local indent = string.rep("  ", depth)
    
    Debug.trace(debugger, "fibonacci_call", {n = n, depth = depth})
    Debug.increment(debugger, "fibonacci_calls")
    
    log.debug(indent .. "-> fibonacci(" .. n .. ")")
    
    if n <= 1 then
        log.debug(indent .. "<- fibonacci(" .. n .. ") = " .. n .. " (base case)")
        return n
    end
    
    local result = calculate_fibonacci(n - 1, depth + 1) + calculate_fibonacci(n - 2, depth + 1)
    log.debug(indent .. "<- fibonacci(" .. n .. ") = " .. result)
    
    return result
end

-- Calculate with tracing
print("\nCalculating Fibonacci(5) with tracing:")
Debug.start_timer(debugger, "fibonacci_calculation")
local fib_result = calculate_fibonacci(5)
local fib_time = Debug.stop_timer(debugger, "fibonacci_calculation")

print("Result: " .. fib_result)
print("Time: " .. string.format("%.3f", fib_time) .. " seconds")
print("Total calls: " .. debugger.counters.fibonacci_calls)
print()

-- Example 2: Performance Profiling
print("=== Example 2: Performance Profiling ===")

-- Profile different operations
local operations = {
    {
        name = "string_concatenation",
        fn = function()
            local s = ""
            for i = 1, 1000 do
                s = s .. "x"
            end
            return #s
        end
    },
    {
        name = "table_append",
        fn = function()
            local t = {}
            for i = 1, 1000 do
                t[#t + 1] = "x"
            end
            return #t
        end
    },
    {
        name = "json_operations",
        fn = function()
            local obj = {data = {}}
            for i = 1, 100 do
                obj.data[i] = {id = i, value = i * 2}
            end
            local json = data.to_json(obj)
            local parsed = data.from_json(json)
            return #parsed.data
        end
    }
}

print("\nProfiling operations:")
for _, op in ipairs(operations) do
    Debug.start_timer(debugger, op.name)
    
    local success, result = pcall(op.fn)
    
    local duration = Debug.stop_timer(debugger, op.name)
    
    if success then
        print(string.format("  %s: %.3fs (result: %s)", op.name, duration, tostring(result)))
    else
        Debug.record_error(debugger, "profile_error", tostring(result), op.name)
        print(string.format("  %s: FAILED", op.name))
    end
end

print()

-- Example 3: Agent Call Debugging
print("=== Example 3: Agent Call Debugging ===")

-- Create debug-wrapped agent
local debug_agent = agent.create("Debug Assistant", {
    model = model,
    system = "You are a helpful assistant. Be very concise.",
    temperature = 0.7
})

-- Wrap agent run method with debugging
local original_run = debug_agent.run
debug_agent.run = function(self, prompt)
    Debug.trace(debugger, "agent_call_start", {prompt = prompt})
    Debug.start_timer(debugger, "agent_call")
    Debug.increment(debugger, "agent_calls")
    
    log.info("Agent prompt: " .. string.sub(prompt, 1, 50) .. "...")
    
    local success, result = pcall(original_run, self, prompt)
    
    local duration = Debug.stop_timer(debugger, "agent_call")
    
    if success then
        Debug.trace(debugger, "agent_call_complete", {
            duration = duration,
            response_length = #result
        })
        log.info("Agent response: " .. string.sub(result, 1, 50) .. "...")
        return result
    else
        Debug.record_error(debugger, "agent_error", tostring(result), {prompt = prompt})
        error(result)
    end
end

-- Test agent with debugging
print("\nTesting agent with debug wrapper:")
local test_prompts = {
    "What is 2+2?",
    "Name a color.",
    "Say hello."
}

for _, prompt in ipairs(test_prompts) do
    print("\nPrompt: " .. prompt)
    local response = debug_agent:run(prompt)
    print("Response: " .. response)
end

print()

-- Example 4: Error Tracking
print("=== Example 4: Error Tracking ===")

-- Function that might fail
local function risky_operation(value)
    Debug.trace(debugger, "risky_operation", {input = value})
    
    if type(value) ~= "number" then
        error("Expected number, got " .. type(value))
    end
    
    if value < 0 then
        error("Value must be non-negative")
    end
    
    if value == 13 then
        error("Unlucky number!")
    end
    
    return math.sqrt(value)
end

-- Test with various inputs
print("\nTesting error tracking:")
local test_values = {4, "hello", -5, 13, 16}

for _, value in ipairs(test_values) do
    print("  Testing with: " .. tostring(value))
    
    local success, result = pcall(risky_operation, value)
    
    if success then
        print("    Success: " .. result)
    else
        Debug.record_error(debugger, "risky_operation_failed", result, {input = value})
        print("    Failed: " .. result)
    end
end

print()

-- Generate debug report
print("=== Debug Report ===")

-- Summary
print("\nOperation Summary:")
print("  Total traces: " .. #debugger.traces)
print("  Total errors: " .. #debugger.errors)

print("\nCounters:")
for name, count in pairs(debugger.counters) do
    print("  " .. name .. ": " .. count)
end

print("\nTimers:")
for name, timer in pairs(debugger.timers) do
    if timer.total_duration then
        print(string.format("  %s: %.3fs total (%d calls, %.3fs avg)",
            name,
            timer.total_duration,
            timer.count,
            timer.total_duration / timer.count
        ))
    end
end

-- Save detailed report
local report = {
    summary = {
        total_traces = #debugger.traces,
        total_errors = #debugger.errors,
        counters = debugger.counters,
        timers = {}
    },
    traces = debugger.traces,
    errors = debugger.errors
}

-- Process timer data for report
for name, timer in pairs(debugger.timers) do
    if timer.total_duration then
        report.summary.timers[name] = {
            total_duration = timer.total_duration,
            count = timer.count,
            average_duration = timer.total_duration / timer.count
        }
    end
end

-- Save report
utils.file_write(output_dir .. "/debug_report.json", data.to_json(report))
print("\nDebug report saved to: " .. output_dir .. "/debug_report.json")

-- Save trace log
local trace_log = "Debug Trace Log\n"
trace_log = trace_log .. "===============\n\n"

for _, trace in ipairs(debugger.traces) do
    trace_log = trace_log .. string.format(
        "[%s] %s: %s\n",
        os.date("%H:%M:%S", trace.timestamp),
        trace.name,
        type(trace.data) == "table" and data.to_json(trace.data) or tostring(trace.data)
    )
end

utils.file_write(output_dir .. "/trace.log", trace_log)
print("Trace log saved to: " .. output_dir .. "/trace.log")

print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Function tracing with structured logging")
print("2. Performance profiling with timers")
print("3. Agent call debugging with wrappers")
print("4. Error tracking and reporting")
print("5. Debug report generation")
print()
print("Key insights:")
print("- Structured logging helps track execution flow")
print("- Timers reveal performance bottlenecks")
print("- Wrapping functions enables monitoring")
print("- Error tracking improves reliability")
print("- Reports provide debugging insights")
print()

-- Return summary
return {
    total_traces = #debugger.traces,
    total_errors = #debugger.errors,
    performance_tests = #operations,
    files_created = 2
}