-- ABOUTME: Hooks example showing lifecycle hooks, execution hooks, and custom event handling
-- ABOUTME: Demonstrates hooks module usage for intercepting and modifying behavior at runtime

-- Hooks Example
-- This spell demonstrates comprehensive hook usage including:
-- 1. Pre/post execution hooks
-- 2. Error handling hooks
-- 3. Transform hooks for data modification
-- 4. Conditional hooks with predicates
-- 5. Hook priorities and ordering

-- Required modules
local hooks = require("hooks")
local llm = require("llm")
local agent = require("agent")
local log = require("log")
local errors = require("errors")
local data = require("data")
local core = require("core")
local utils = require("utils")

-- Example 1: Basic Pre/Post Execution Hooks
print("=== Example 1: Basic Pre/Post Execution Hooks ===")

-- Register a pre-execution hook for LLM calls
hooks.register("llm.complete.pre", function(args)
    log.debug("Pre-execution hook: LLM call starting")
    log.debug("Model: " .. (args.model or "default"))
    log.debug("Messages: " .. #(args.messages or {}))
    
    -- Modify arguments (e.g., inject system message)
    if not args.messages[1] or args.messages[1].role ~= "system" then
        table.insert(args.messages, 1, {
            role = "system",
            content = "You are a helpful assistant. Always be concise."
        })
        log.info("Injected system message via hook")
    end
    
    -- Track timing
    args.__start_time = os.time()
    
    return args  -- Return modified arguments
end)

-- Register a post-execution hook for LLM calls
hooks.register("llm.complete.post", function(result, args)
    local elapsed = os.time() - (args.__start_time or 0)
    log.debug("Post-execution hook: LLM call completed in " .. elapsed .. "s")
    
    -- Add metadata to result
    result.metadata = result.metadata or {}
    result.metadata.execution_time = elapsed
    result.metadata.hook_processed = true
    
    -- Log token usage
    if result.usage then
        log.info(string.format("Tokens used - Prompt: %d, Completion: %d, Total: %d",
            result.usage.prompt_tokens or 0,
            result.usage.completion_tokens or 0,
            result.usage.total_tokens or 0))
    end
    
    return result  -- Return modified result
end)

-- Make an LLM call that triggers hooks
local response = llm.complete({
    model = "gpt-3.5-turbo",
    messages = {
        {role = "user", content = "What is a hook in programming?"}
    }
})

print("Response: " .. string.sub(response.content, 1, 100) .. "...")
print("Hook metadata added: " .. tostring(response.metadata and response.metadata.hook_processed))
print()

-- Example 2: Error Handling Hooks
print("=== Example 2: Error Handling Hooks ===")

-- Global error counter
local error_count = 0

-- Register error hook
hooks.register("error", function(err, context)
    error_count = error_count + 1
    log.error("Error hook triggered (#" .. error_count .. ")")
    log.error("Error: " .. tostring(err))
    log.error("Context: " .. (context.operation or "unknown"))
    
    -- Attempt recovery for specific errors
    if string.find(tostring(err), "rate limit") then
        log.warn("Rate limit detected, adding delay...")
        utils.general_sleep(2000)
        return {retry = true, delay = 2}
    end
    
    -- Log stack trace for other errors
    if context.stack_trace then
        log.debug("Stack trace: " .. context.stack_trace)
    end
    
    return {retry = false}
end)

-- Simulate an error-prone operation
local function risky_operation(should_fail)
    hooks.trigger("operation.start", {name = "risky_operation"})
    
    if should_fail then
        local err = errors.new("SIMULATED_ERROR", "Simulated rate limit exceeded")
        local recovery = hooks.trigger("error", err, {
            operation = "risky_operation",
            attempt = 1
        })
        
        if recovery and recovery.retry then
            print("Error handled by hook, retrying...")
            return "Success after retry"
        else
            error(err)
        end
    end
    
    hooks.trigger("operation.end", {name = "risky_operation", status = "success"})
    return "Success"
end

-- Test error handling
local result = risky_operation(true)
print("Operation result: " .. result)
print("Total errors handled: " .. error_count)
print()

-- Example 3: Transform Hooks for Data Processing
print("=== Example 3: Transform Hooks for Data Processing ===")

-- Register transform hooks for agent responses
hooks.register("agent.response.transform", function(response)
    -- Add thinking process as metadata
    local thinking_pattern = "(?<thinking>.*?)</thinking>"
    local thinking = string.match(response.content, "<thinking>(.-)</thinking>")
    
    if thinking then
        response.metadata = response.metadata or {}
        response.metadata.thinking = thinking
        -- Remove thinking tags from visible response
        response.content = string.gsub(response.content, "<thinking>.-</thinking>", "")
        log.debug("Extracted thinking process via hook")
    end
    
    -- Add word count
    local word_count = 0
    for word in string.gmatch(response.content, "%S+") do
        word_count = word_count + 1
    end
    response.metadata = response.metadata or {}
    response.metadata.word_count = word_count
    
    return response
end)

-- Register output sanitization hook
hooks.register("output.sanitize", function(text)
    -- Remove any potential sensitive information
    text = string.gsub(text, "%d%d%d%d%d%d%d%d%d+", "[REDACTED_NUMBER]")
    text = string.gsub(text, "[%w%.%-]+@[%w%.%-]+", "[REDACTED_EMAIL]")
    
    -- Ensure proper formatting
    text = string.gsub(text, "%s+", " ")  -- Normalize whitespace
    text = string.gsub(text, "^%s+", "")  -- Trim start
    text = string.gsub(text, "%s+$", "")  -- Trim end
    
    return text
end)

-- Create an agent that will trigger transform hooks
local analytical_agent = agent.create({
    name = "Analytical Assistant",
    model = "gpt-4",
    instructions = [[
        When answering, first think through your response in <thinking> tags,
        then provide your answer. Include some numbers in your response.
    ]]
})

-- Get response that will be transformed
local agent_response = analytical_agent:run({
    prompt = "What are the top 3 benefits of using hooks? My phone is 1234567890.",
    max_tokens = 200
})

-- Apply transform hooks
agent_response = hooks.trigger("agent.response.transform", agent_response)
agent_response.content = hooks.trigger("output.sanitize", agent_response.content)

print("Transformed response: " .. agent_response.content)
print("Word count: " .. (agent_response.metadata and agent_response.metadata.word_count or "N/A"))
print()

-- Example 4: Conditional Hooks with Priorities
print("=== Example 4: Conditional Hooks with Priorities ===")

-- Register multiple hooks with different priorities
hooks.register("data.validate", function(data)
    log.debug("Validation hook 1 (priority 10)")
    if not data.name then
        error("Name is required")
    end
    return data
end, {priority = 10})

hooks.register("data.validate", function(data)
    log.debug("Validation hook 2 (priority 5)")
    if data.age and data.age < 0 then
        data.age = 0  -- Fix negative age
        log.warn("Fixed negative age")
    end
    return data
end, {priority = 5})

hooks.register("data.validate", function(data)
    log.debug("Validation hook 3 (priority 1)")
    -- Add timestamp
    data.validated_at = os.time()
    return data
end, {priority = 1})

-- Conditional hook that only runs for specific data types
hooks.register("data.process", function(data)
    if data.type ~= "user" then
        return data  -- Skip non-user data
    end
    
    log.debug("Processing user data")
    data.processed = true
    
    -- Normalize name
    if data.name then
        data.name = string.upper(string.sub(data.name, 1, 1)) .. 
                   string.lower(string.sub(data.name, 2))
    end
    
    return data
end, {
    condition = function(data)
        return data.type == "user"
    end
})

-- Test hook execution order and conditions
local test_data = {
    type = "user",
    name = "john doe",
    age = -5
}

print("Original data: " .. data.to_json(test_data))

-- Trigger validation hooks (executed by priority)
test_data = hooks.trigger("data.validate", test_data)
print("After validation: " .. data.to_json(test_data))

-- Trigger conditional processing
test_data = hooks.trigger("data.process", test_data)
print("After processing: " .. data.to_json(test_data))
print()

-- Example 5: Hook Chains and Middleware
print("=== Example 5: Hook Chains and Middleware ===")

-- Create a middleware system using hooks
local middleware_stack = {}

local function add_middleware(name, handler)
    table.insert(middleware_stack, {name = name, handler = handler})
    
    hooks.register("middleware.execute", function(context)
        log.debug("Middleware: " .. name)
        return handler(context)
    end, {
        name = name,
        condition = function(ctx)
            return ctx.middleware == name or ctx.middleware == "all"
        end
    })
end

-- Add authentication middleware
add_middleware("auth", function(context)
    if not context.user then
        error("Authentication required")
    end
    context.authenticated = true
    log.info("User authenticated: " .. context.user)
    return context
end)

-- Add logging middleware
add_middleware("logging", function(context)
    local start_time = os.time()
    
    -- Log request
    log.info(string.format("[%s] %s by %s",
        os.date("%Y-%m-%d %H:%M:%S"),
        context.action or "unknown",
        context.user or "anonymous"))
    
    -- Add cleanup hook
    hooks.register("middleware.cleanup", function()
        local duration = os.time() - start_time
        log.info("Request completed in " .. duration .. "s")
    end, {once = true})  -- Only run once
    
    return context
end)

-- Add rate limiting middleware
local request_counts = {}
add_middleware("rate_limit", function(context)
    local user = context.user or "anonymous"
    local current_time = os.time()
    
    -- Initialize or clean old entries
    if not request_counts[user] then
        request_counts[user] = {}
    end
    
    -- Count recent requests (last 60 seconds)
    local recent_count = 0
    local cutoff_time = current_time - 60
    local new_requests = {}
    
    for _, timestamp in ipairs(request_counts[user]) do
        if timestamp > cutoff_time then
            recent_count = recent_count + 1
            table.insert(new_requests, timestamp)
        end
    end
    
    request_counts[user] = new_requests
    
    -- Check rate limit
    if recent_count >= 10 then
        error("Rate limit exceeded: " .. recent_count .. " requests in last minute")
    end
    
    -- Record this request
    table.insert(request_counts[user], current_time)
    context.rate_limit_remaining = 10 - recent_count - 1
    
    return context
end)

-- Execute middleware chain
local function execute_request(context)
    -- Run all middleware
    for _, mw in ipairs(middleware_stack) do
        context.middleware = mw.name
        context = hooks.trigger("middleware.execute", context)
    end
    
    -- Execute the actual action
    print("Executing action: " .. (context.action or "none"))
    
    -- Cleanup
    hooks.trigger("middleware.cleanup")
    
    return context
end

-- Test middleware chain
local request_context = {
    user = "alice",
    action = "generate_report"
}

local result = execute_request(request_context)
print("Request completed. Rate limit remaining: " .. (result.rate_limit_remaining or "N/A"))
print()

-- Example 6: Dynamic Hook Management
print("=== Example 6: Dynamic Hook Management ===")

-- List all registered hooks
local registered_hooks = hooks.list()
print("Total registered hooks: " .. #registered_hooks)

-- Disable specific hooks temporarily
print("\nDisabling error hooks temporarily...")
hooks.disable("error")

-- This error won't trigger the error hook
local success, err = pcall(function()
    error("This error won't be caught by hooks")
end)
print("Error occurred: " .. tostring(not success))
print("Error caught by hook: false")

-- Re-enable error hooks
hooks.enable("error")

-- Remove specific hooks
print("\nCleaning up hooks...")
local removed = hooks.remove("middleware.cleanup")
print("Removed " .. removed .. " cleanup hooks")

-- Clear all hooks of a specific type
hooks.clear("data.validate")
print("Cleared all validation hooks")

-- Register a one-time hook
hooks.register("final.cleanup", function()
    print("One-time cleanup hook executed")
    return {
        hooks_remaining = #hooks.list(),
        errors_handled = error_count
    }
end, {once = true})

-- Trigger the one-time hook
local cleanup_result = hooks.trigger("final.cleanup")

-- Try triggering again (won't execute)
local second_result = hooks.trigger("final.cleanup")
print("Second trigger result: " .. tostring(second_result))

-- Return summary
return {
    success = true,
    total_errors_handled = error_count,
    hooks_executed = true,
    middleware_tested = true,
    final_hook_count = cleanup_result and cleanup_result.hooks_remaining or 0
}