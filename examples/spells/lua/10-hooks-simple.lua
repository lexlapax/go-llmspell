-- ABOUTME: Simplified hooks example demonstrating hook-like patterns without hooks module
-- ABOUTME: Shows interceptors, decorators, and event-based patterns for behavior modification

-- Hooks Example (Simplified)
-- This spell demonstrates hook-like patterns:
-- 1. Function wrapping for pre/post execution
-- 2. Error handling decorators
-- 3. Transform functions for data modification
-- 4. Event-based interception patterns

-- Required modules
local llm = require("llm")
local agent = require("agent")
local logging = require("logging")
local data = require("data")
local utils = require("utils")

-- Create logger
local log = logging.default()

-- Parameters
local model = params.model or "gpt-4"
local output_dir = params.output_dir or "./hooks-output"

print("=== Hooks Pattern Example (Simplified) ===")
print()

-- Note: File operations removed - focus on hook patterns
-- In production, results would be saved to files

-- Simple hook system implementation
local SimpleHooks = {}

function SimpleHooks.new()
    return {
        hooks = {},
        metrics = {
            calls = {},
            errors = {},
            durations = {}
        }
    }
end

function SimpleHooks.wrap(hooks, name, original_fn)
    return function(...)
        local args = {...}
        local start_time = os.time()
        
        -- Pre-execution hooks
        local pre_hooks = hooks.hooks[name .. ".pre"] or {}
        for _, hook in ipairs(pre_hooks) do
            local modified_args = hook(args)
            if modified_args then
                args = modified_args
            end
        end
        
        -- Execute original function
        local success, result = pcall(original_fn, unpack(args))
        
        -- Track metrics
        hooks.metrics.calls[name] = (hooks.metrics.calls[name] or 0) + 1
        hooks.metrics.durations[name] = (hooks.metrics.durations[name] or 0) + (os.time() - start_time)
        
        if not success then
            hooks.metrics.errors[name] = (hooks.metrics.errors[name] or 0) + 1
            
            -- Error hooks
            local error_hooks = hooks.hooks[name .. ".error"] or {}
            for _, hook in ipairs(error_hooks) do
                hook(result, args)
            end
            
            error(result)
        end
        
        -- Post-execution hooks
        local post_hooks = hooks.hooks[name .. ".post"] or {}
        for _, hook in ipairs(post_hooks) do
            local modified_result = hook(result, args)
            if modified_result ~= nil then
                result = modified_result
            end
        end
        
        return result
    end
end

function SimpleHooks.register(hooks, event, fn)
    if not hooks.hooks[event] then
        hooks.hooks[event] = {}
    end
    hooks.hooks[event][#hooks.hooks[event] + 1] = fn
    print("  📌 Registered hook for: " .. event)
end

-- Create hook system
local hook_system = SimpleHooks.new()

-- Example 1: Basic Pre/Post Execution Hooks
print("=== Example 1: Basic Pre/Post Execution Hooks ===")

-- Register pre-execution hook
SimpleHooks.register(hook_system, "agent.run.pre", function(args)
    print("  🔵 Pre-hook: Agent starting")
    print("    Input: " .. string.sub(tostring(args[1]), 1, 50) .. "...")
    
    -- Log to file
    local log_entry = string.format("[%s] PRE: %s\n", 
        os.date("%H:%M:%S"), 
        tostring(args[1])
    )
    -- In production would append to log file
    print(log_entry)
    
    return args
end)

-- Register post-execution hook
SimpleHooks.register(hook_system, "agent.run.post", function(result, args)
    print("  🟢 Post-hook: Agent completed")
    print("    Output length: " .. #tostring(result))
    
    -- Add metadata
    if type(result) == "string" then
        result = result .. "\n[Processed by hook system]"
    end
    
    -- Log to file
    local log_entry = string.format("[%s] POST: %d chars\n",
        os.date("%H:%M:%S"),
        #tostring(result)
    )
    -- In production would append to log file
    print(log_entry)
    
    return result
end)

-- Create wrapped agent
local original_agent = agent.create("Assistant", {
    model = model,
    system = "You are a helpful assistant.",
    temperature = 0.7
})

-- Wrap the run method
local wrapped_run = SimpleHooks.wrap(hook_system, "agent.run", function(prompt)
    return original_agent:run(prompt)
end)

-- Test wrapped function
print("\nTesting hooked agent:")
local result = wrapped_run("Explain hooks in one sentence.")
print("Result: " .. string.sub(result, 1, 100) .. "...")
print()

-- Example 2: Error Handling Hooks
print("=== Example 2: Error Handling Hooks ===")

-- Register error hook
SimpleHooks.register(hook_system, "calculation.error", function(error_msg, args)
    print("  ❌ Error hook triggered: " .. error_msg)
    print("    Arguments: " .. data.to_json(args))
    
    -- Log error (in production would write to file)
    print(string.format("[%s] ERROR: %s",
        os.date("%H:%M:%S"),
        error_msg
    ))
end)

-- Create error-prone function
local function risky_calculation(a, b)
    if b == 0 then
        error("Division by zero")
    end
    return a / b
end

-- Wrap with error handling
local safe_calculation = SimpleHooks.wrap(hook_system, "calculation", risky_calculation)

-- Test error handling
print("\nTesting error hooks:")
local test_cases = {{10, 2}, {10, 0}, {20, 4}}

for _, case in ipairs(test_cases) do
    local a, b = case[1], case[2]
    print(string.format("  Calculating %d / %d", a, b))
    
    local success, result = pcall(safe_calculation, a, b)
    if success then
        print("    Result: " .. result)
    else
        print("    Failed (error was handled)")
    end
end

print()

-- Example 3: Transform Hooks
print("=== Example 3: Transform Hooks ===")

-- Create data transformer with hooks
local DataTransformer = {}

function DataTransformer.new()
    return {
        transforms = {}
    }
end

function DataTransformer.add_transform(transformer, name, fn)
    transformer.transforms[name] = fn
    print("  📐 Added transform: " .. name)
end

function DataTransformer.apply(transformer, data)
    local result = data
    for name, fn in pairs(transformer.transforms) do
        print("  Applying transform: " .. name)
        result = fn(result)
    end
    return result
end

-- Create transformer
local transformer = DataTransformer.new()

-- Add transforms
DataTransformer.add_transform(transformer, "lowercase", function(text)
    return text:lower()
end)

DataTransformer.add_transform(transformer, "trim", function(text)
    return text:gsub("^%s+", ""):gsub("%s+$", "")
end)

DataTransformer.add_transform(transformer, "sanitize", function(text)
    return text:gsub("[^%w%s]", "")
end)

-- Test transformer
print("\nTesting data transforms:")
local test_data = "  HELLO, World! 123  "
print("Original: '" .. test_data .. "'")

local transformed = DataTransformer.apply(transformer, test_data)
print("Transformed: '" .. transformed .. "'")
print()

-- Example 4: Conditional Hooks
print("=== Example 4: Conditional Hooks ===")

-- Add conditional hook system
function SimpleHooks.register_conditional(hooks, event, condition, fn)
    SimpleHooks.register(hooks, event, function(...)
        if condition(...) then
            return fn(...)
        end
    end)
    print("  📎 Registered conditional hook for: " .. event)
end

-- Register conditional hooks
SimpleHooks.register_conditional(
    hook_system,
    "agent.run.pre",
    function(args) return #args[1] > 100 end,  -- Only for long prompts
    function(args)
        print("  ⚡ Long prompt detected, adding optimization hint")
        args[1] = args[1] .. "\n\nPlease be concise in your response."
        return args
    end
)

-- Test conditional hook
print("\nTesting conditional hooks:")
print("Short prompt:")
wrapped_run("Hi")

print("\nLong prompt:")
local long_prompt = string.rep("Please explain in detail ", 10)
wrapped_run(long_prompt)

print()

-- Show metrics
print("=== Hook System Metrics ===")
print("Call counts:")
for name, count in pairs(hook_system.metrics.calls) do
    print("  " .. name .. ": " .. count .. " calls")
end

print("\nError counts:")
for name, count in pairs(hook_system.metrics.errors) do
    print("  " .. name .. ": " .. count .. " errors")
end

print("\nTotal durations:")
for name, duration in pairs(hook_system.metrics.durations) do
    print("  " .. name .. ": " .. duration .. " seconds")
end

-- Save metrics
-- In production, would save metrics to file
print("\nHook metrics (would be saved to metrics.json)")
if hook_system and hook_system.metrics then
    print("  Total calls:", #(hook_system.metrics.calls or {}))
    print("  Total errors:", #(hook_system.metrics.errors or {}))
end

print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Function wrapping for pre/post execution")
print("2. Error handling with hook patterns")
print("3. Data transformation pipelines")
print("4. Conditional hook execution")
print("5. Metrics collection via hooks")
print()
print("Key insights:")
print("- Hooks enable behavior modification without changing original code")
print("- Pre/post hooks are useful for logging and monitoring")
print("- Error hooks provide centralized error handling")
print("- Transform hooks create data pipelines")
print("- Conditional hooks add dynamic behavior")
print()

-- Return summary
return {
    total_hooks = #hook_system.hooks,
    total_calls = hook_system.metrics.calls["agent.run"] or 0,
    transforms_applied = 3,
    files_created = 3
}