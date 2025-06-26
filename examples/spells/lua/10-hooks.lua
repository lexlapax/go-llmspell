-- ABOUTME: Example demonstrating LLM pipeline hooks for intercepting agent/LLM operations
-- ABOUTME: Shows hook registration, priorities, and lifecycle management using the hooks module

-- LLM Pipeline Hooks Example
-- This spell demonstrates the hooks module for LLM pipeline integration:
-- 1. Before/After generate hooks
-- 2. Before/After tool call hooks
-- 3. Hook priorities and ordering
-- 4. Hook management (enable/disable/list)
-- 5. Practical use cases

-- Required modules
local hooks = require("hooks")
local agent = require("agent")
local tools = require("tools")
local data = require("data")

-- Parameters
local model = params.model or "gpt-4"

print("=== LLM Pipeline Hooks Example ===")
print()

-- Example 1: Basic Generate Hooks
print("=== Example 1: Before/After Generate Hooks ===")

-- Register a before-generate hook to modify prompts
hooks.register_hook("enhance_prompt", {
    type = hooks.TYPES.BEFORE_GENERATE,
    priority = hooks.PRIORITY.HIGH,
    handler = function(context)
        print("  [BEFORE_GENERATE] Original prompt: " .. (context.prompt or "N/A"))
        
        -- Enhance the prompt with additional context
        if context.prompt then
            context.prompt = "Please provide a detailed response. " .. context.prompt
            print("  [BEFORE_GENERATE] Enhanced prompt: " .. context.prompt)
        end
        
        -- Track timing
        context.start_time = os.time()
        
        return context
    end
})

-- Register an after-generate hook to analyze responses
hooks.register_hook("analyze_response", {
    type = hooks.TYPES.AFTER_GENERATE,
    priority = hooks.PRIORITY.NORMAL,
    handler = function(context)
        local elapsed = os.time() - (context.start_time or 0)
        print("  [AFTER_GENERATE] Generation took " .. elapsed .. " seconds")
        
        if context.response then
            -- Count words in response
            local word_count = 0
            for word in string.gmatch(context.response, "%S+") do
                word_count = word_count + 1
            end
            print("  [AFTER_GENERATE] Response word count: " .. word_count)
            
            -- Add metadata
            context.metadata = context.metadata or {}
            context.metadata.word_count = word_count
            context.metadata.generation_time = elapsed
        end
        
        return context
    end
})

-- Create an agent that will trigger hooks
local assistant = agent.create("Assistant", {
    model = model,
    system = "You are a helpful assistant.",
    temperature = 0.7
})

-- Test the hooks with a simple query
print("\nTesting generate hooks:")
local response = assistant:run("What is the capital of France?")
print("Response: " .. response)
print()

-- Example 2: Tool Call Hooks
print("=== Example 2: Tool Call Hooks ===")

-- Define a simple calculator tool
local calc_tool = tools.define(
    "calculator",
    "Performs basic arithmetic operations",
    {
        parameters = {
            a = {type = "number", required = true},
            b = {type = "number", required = true},
            operation = {type = "string", required = true}
        }
    },
    function(args)
        if args.operation == "add" then
            return args.a + args.b
        elseif args.operation == "subtract" then
            return args.a - args.b
        elseif args.operation == "multiply" then
            return args.a * args.b
        elseif args.operation == "divide" then
            if args.b == 0 then
                error("Division by zero")
            end
            return args.a / args.b
        else
            error("Unknown operation: " .. args.operation)
        end
    end
)

-- Register before-tool-call hook for validation
hooks.register_hook("validate_tool_call", {
    type = hooks.TYPES.BEFORE_TOOL_CALL,
    priority = hooks.PRIORITY.HIGH,
    handler = function(context)
        print("  [BEFORE_TOOL_CALL] Tool: " .. (context.tool_name or "unknown"))
        print("  [BEFORE_TOOL_CALL] Arguments: " .. data.to_json(context.arguments or {}))
        
        -- Validate calculator inputs
        if context.tool_name == "calculator" then
            local args = context.arguments
            if args and args.operation == "divide" and args.b == 0 then
                print("  [BEFORE_TOOL_CALL] Warning: Division by zero detected!")
                -- Could modify args or throw error here
            end
        end
        
        context.tool_start_time = os.time()
        return context
    end
})

-- Register after-tool-call hook for logging
hooks.register_hook("log_tool_result", {
    type = hooks.TYPES.AFTER_TOOL_CALL,
    priority = hooks.PRIORITY.NORMAL,
    handler = function(context)
        local elapsed = os.time() - (context.tool_start_time or 0)
        print("  [AFTER_TOOL_CALL] Tool execution took " .. elapsed .. " seconds")
        
        if context.result then
            print("  [AFTER_TOOL_CALL] Result: " .. tostring(context.result))
        end
        
        if context.error then
            print("  [AFTER_TOOL_CALL] Error: " .. tostring(context.error))
        end
        
        return context
    end
})

-- Create an agent with tool access
local math_assistant = agent.create("Math Assistant", {
    model = model,
    system = "You are a math assistant. Use the calculator tool for computations.",
    tools = {"calculator"},
    temperature = 0.3
})

-- Test tool call hooks
print("\nTesting tool call hooks:")
local result = math_assistant:run("Calculate 15 divided by 3")
print("Result: " .. result)
print()

-- Example 3: Hook Priorities and Multiple Handlers
print("=== Example 3: Hook Priorities ===")

-- Register multiple before-generate hooks with different priorities
hooks.register_hook("priority_highest", {
    type = hooks.TYPES.BEFORE_GENERATE,
    priority = hooks.PRIORITY.HIGHEST,
    handler = function(context)
        print("  [Priority 1000] Highest priority hook")
        context.hook_chain = (context.hook_chain or "") .. "1-"
        return context
    end
})

hooks.register_hook("priority_normal", {
    type = hooks.TYPES.BEFORE_GENERATE,
    priority = hooks.PRIORITY.NORMAL,
    handler = function(context)
        print("  [Priority 0] Normal priority hook")
        context.hook_chain = (context.hook_chain or "") .. "2-"
        return context
    end
})

hooks.register_hook("priority_low", {
    type = hooks.TYPES.BEFORE_GENERATE,
    priority = hooks.PRIORITY.LOW,
    handler = function(context)
        print("  [Priority -100] Low priority hook")
        context.hook_chain = (context.hook_chain or "") .. "3"
        return context
    end
})

-- Test priority ordering
print("\nTesting hook priorities (should execute in order: highest, normal, low):")
local test_response = assistant:run("Say hello")
print("Hook execution chain: " .. (assistant.last_context and assistant.last_context.hook_chain or "N/A"))
print()

-- Example 4: Hook Management
print("=== Example 4: Hook Management ===")

-- List all registered hooks
local all_hooks = hooks.list_hooks()
print("\nRegistered hooks: " .. #all_hooks)
for i, hook in ipairs(all_hooks) do
    print(string.format("  %d. %s (type: %s, priority: %s)", 
        i, hook.id, hook.type, hook.priority))
end

-- Disable specific hooks
print("\nDisabling 'priority_normal' hook...")
hooks.disable_hook("priority_normal")

-- Enable it again
print("Re-enabling 'priority_normal' hook...")
hooks.enable_hook("priority_normal")

-- Remove hooks we don't need anymore
print("\nCleaning up priority test hooks...")
hooks.unregister_hook("priority_highest")
hooks.unregister_hook("priority_normal")
hooks.unregister_hook("priority_low")

-- Example 5: Practical Use Cases
print("\n=== Example 5: Practical Use Cases ===")

-- Use case 1: Token counting and limits
hooks.register_hook("token_limiter", {
    type = hooks.TYPES.BEFORE_GENERATE,
    priority = hooks.PRIORITY.HIGH,
    handler = function(context)
        -- In production, would use actual tokenizer
        local estimated_tokens = context.prompt and #context.prompt / 4 or 0
        print("  [Token Limiter] Estimated tokens: " .. math.floor(estimated_tokens))
        
        if estimated_tokens > 1000 then
            print("  [Token Limiter] Warning: Prompt may be too long!")
            -- Could truncate or modify prompt here
        end
        
        return context
    end
})

-- Use case 2: Response caching
local response_cache = {}
hooks.register_hook("response_cache", {
    type = hooks.TYPES.BEFORE_GENERATE,
    priority = hooks.PRIORITY.HIGHEST,
    handler = function(context)
        local cache_key = context.prompt or ""
        if response_cache[cache_key] then
            print("  [Cache] Hit! Returning cached response")
            context.response = response_cache[cache_key]
            context.skip_generation = true  -- If supported by the bridge
        end
        return context
    end
})

hooks.register_hook("cache_store", {
    type = hooks.TYPES.AFTER_GENERATE,
    priority = hooks.PRIORITY.LOWEST,
    handler = function(context)
        if context.response and context.prompt then
            response_cache[context.prompt] = context.response
            print("  [Cache] Stored response in cache")
        end
        return context
    end
})

-- Test practical hooks
print("\nTesting practical hooks:")
local query = "What is 2+2?"

-- First call - should cache
print("First call:")
assistant:run(query)

-- Second call - should hit cache
print("\nSecond call (should hit cache):")
assistant:run(query)

-- Example 6: Advanced Hook Patterns
print("\n=== Example 6: Advanced Patterns ===")

-- Pattern 1: Conditional hooks using context
hooks.register_hook("dev_mode_logger", {
    type = hooks.TYPES.AFTER_GENERATE,
    priority = hooks.PRIORITY.LOW,
    handler = function(context)
        -- Only log in development mode
        if context.dev_mode then
            print("  [DEV] Full context: " .. data.to_json(context))
        end
        return context
    end
})

-- Pattern 2: Error handling in hooks
hooks.register_hook("safe_processor", {
    type = hooks.TYPES.AFTER_GENERATE,
    priority = hooks.PRIORITY.NORMAL,
    handler = function(context)
        local success, result = pcall(function()
            -- Some processing that might fail
            if context.response and string.find(context.response, "error") then
                error("Response contains error word")
            end
            return context
        end)
        
        if not success then
            print("  [Safe Processor] Error caught: " .. tostring(result))
            context.processing_error = tostring(result)
        end
        
        return context
    end
})

-- Pattern 3: Chain of responsibility
hooks.register_hook("response_filter", {
    type = hooks.TYPES.AFTER_GENERATE,
    priority = hooks.PRIORITY.HIGH,
    handler = function(context)
        if context.response then
            -- Remove any potential sensitive data
            context.response = string.gsub(context.response, "%d%d%d%-?%d%d%-?%d%d%d%d", "[REDACTED]")
            print("  [Filter] Applied sensitive data filter")
        end
        return context
    end
})

-- Test advanced patterns
print("\nTesting advanced patterns:")
local sensitive_response = assistant:run("My SSN is 123-45-6789")
print("Filtered response: " .. sensitive_response)

-- Cleanup
print("\n=== Cleanup ===")
print("Unregistering all example hooks...")

-- Get current hooks and unregister them
local final_hooks = hooks.list_hooks()
for _, hook in ipairs(final_hooks) do
    hooks.unregister_hook(hook.id)
end

print("All hooks cleaned up. Total remaining: " .. #hooks.list_hooks())

-- Summary
print("\n=== Summary ===")
print("This example demonstrated:")
print("1. Before/After generate hooks for modifying prompts and responses")
print("2. Before/After tool call hooks for monitoring tool usage")
print("3. Hook priorities and execution order")
print("4. Hook management (list, enable, disable, unregister)")
print("5. Practical use cases (token limiting, caching, filtering)")
print("6. Advanced patterns (conditional hooks, error handling)")

return {
    hooks_demonstrated = 6,
    patterns_shown = {
        "basic_lifecycle",
        "tool_monitoring",
        "priority_ordering",
        "management_operations",
        "practical_applications",
        "advanced_patterns"
    }
}