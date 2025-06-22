# Troubleshooting Guide

This guide helps you diagnose and fix common issues when writing and running Lua spells in go-llmspell.

## Table of Contents

- [Common Issues](#common-issues)
- [Debugging Techniques](#debugging-techniques)
- [Error Messages](#error-messages)
- [Performance Issues](#performance-issues)
- [LLM-Specific Issues](#llm-specific-issues)
- [Tool Problems](#tool-problems)
- [State and Memory Issues](#state-and-memory-issues)
- [Async and Promise Issues](#async-and-promise-issues)
- [Security and Permission Issues](#security-and-permission-issues)
- [Getting Help](#getting-help)

## Common Issues

### Spell Won't Run

**Problem**: Your spell fails to start or immediately exits.

**Common Causes & Solutions**:

1. **Syntax Errors**
   ```lua
   -- BAD: Missing end
   if condition then
       print("test")
   -- GOOD: Proper closure
   if condition then
       print("test")
   end
   ```

2. **Missing Required Parameters**
   ```lua
   -- Add parameter validation at the start
   assert(params.required_field, "required_field parameter is missing")
   ```

3. **Module Not Found**
   ```lua
   -- BAD: Wrong module name
   local llms = require("llms")  -- Module doesn't exist
   
   -- GOOD: Correct module name
   local llm = require("llm")
   ```

### Nil Value Errors

**Problem**: "attempt to index a nil value" or "attempt to call a nil value"

**Solutions**:

```lua
-- BAD: Not checking for nil
local result = api_call()
print(result.data)  -- Crashes if result is nil

-- GOOD: Defensive programming
local result = api_call()
if result and result.data then
    print(result.data)
else
    log.error("API call returned no data")
end

-- BETTER: Using safe navigation pattern
local function safe_get(obj, ...)
    local current = obj
    for _, key in ipairs({...}) do
        if type(current) ~= "table" then
            return nil
        end
        current = current[key]
    end
    return current
end

local data = safe_get(result, "response", "items", 1, "name")
```

### Type Mismatch Errors

**Problem**: "bad argument #1 to 'function' (string expected, got nil)"

**Solutions**:

```lua
-- BAD: Not converting types
local count = params.count  -- params are always strings
local doubled = count * 2   -- Error: can't multiply string

-- GOOD: Explicit type conversion
local count = tonumber(params.count) or 1
local doubled = count * 2

-- Type validation helper
local function ensure_type(value, expected_type, default)
    if type(value) == expected_type then
        return value
    end
    
    -- Try conversion for common cases
    if expected_type == "number" and type(value) == "string" then
        return tonumber(value) or default
    end
    
    return default
end

local temperature = ensure_type(params.temperature, "number", 0.7)
```

## Debugging Techniques

### 1. Debug Output

```lua
-- Create a debug function
local DEBUG = params.debug == "true"

local function debug_print(label, value)
    if DEBUG then
        print(string.format("[DEBUG] %s: %s", label, data.to_json(value)))
    end
end

-- Use throughout your spell
debug_print("LLM Request", request)
local response = llm.complete(request)
debug_print("LLM Response", response)
```

### 2. Interactive Debugging

```lua
-- Add breakpoints for interactive debugging
local function breakpoint(context)
    if params.interactive == "true" then
        print("\n=== BREAKPOINT ===")
        print("Context: " .. data.to_json(context, {pretty = true}))
        print("Press Enter to continue...")
        io.read()
    end
end

-- Use in your code
local result = complex_operation()
breakpoint({
    result = result,
    state = current_state,
    step = "after_complex_operation"
})
```

### 3. Execution Tracing

```lua
-- Trace function calls
local function trace_calls(module_name, module)
    local traced = {}
    
    for name, fn in pairs(module) do
        if type(fn) == "function" then
            traced[name] = function(...)
                log.debug(module_name .. "." .. name .. " called", {args = {...}})
                local results = {fn(...)}
                log.debug(module_name .. "." .. name .. " returned", {results = results})
                return table.unpack(results)
            end
        else
            traced[name] = fn
        end
    end
    
    return traced
end

-- Wrap modules for tracing
if params.trace == "true" then
    llm = trace_calls("llm", llm)
    tools = trace_calls("tools", tools)
end
```

### 4. State Inspection

```lua
-- Create state inspector
local function inspect_state(state_obj, name)
    print("\n=== State: " .. name .. " ===")
    local all_keys = state_obj:keys()
    for _, key in ipairs(all_keys) do
        local value = state_obj:get(key)
        print(string.format("  %s: %s (%s)", 
            key, 
            tostring(value), 
            type(value)
        ))
    end
    print("==================\n")
end

-- Use periodically
inspect_state(my_state, "before_processing")
-- ... do work ...
inspect_state(my_state, "after_processing")
```

## Error Messages

### Common Error Patterns

#### "API rate limit exceeded"
```lua
-- Solution: Implement retry with backoff
local function rate_limited_call(fn, ...)
    local max_retries = 3
    local base_delay = 60  -- 1 minute
    
    for attempt = 1, max_retries do
        local success, result = pcall(fn, ...)
        
        if success then
            return result
        end
        
        if result:find("rate limit") then
            local delay = base_delay * attempt
            log.warn("Rate limited, waiting " .. delay .. " seconds")
            core.sleep(delay)
        else
            error(result)  -- Re-throw non-rate-limit errors
        end
    end
    
    error("Max retries exceeded due to rate limiting")
end
```

#### "timeout exceeded"
```lua
-- Solution: Increase timeout or optimize operation
local response = llm.complete({
    model = "gpt-4",
    messages = messages,
    timeout = 120  -- Increase timeout to 2 minutes
})

-- Or break into smaller operations
local chunks = split_text(large_text, 1000)  -- 1000 chars each
local results = {}

for i, chunk in ipairs(chunks) do
    results[i] = llm.complete({
        model = "gpt-3.5-turbo",
        messages = {{role = "user", content = "Process: " .. chunk}}
    })
end
```

#### "insufficient_quota"
```lua
-- Solution: Check API limits and implement fallbacks
local function with_quota_check(primary_model, fallback_model)
    return function(request)
        request.model = primary_model
        
        local success, result = pcall(llm.complete, request)
        
        if not success and result:find("quota") then
            log.warn("Quota exceeded for " .. primary_model .. ", using fallback")
            request.model = fallback_model
            return llm.complete(request)
        end
        
        if not success then
            error(result)
        end
        
        return result
    end
end

local smart_complete = with_quota_check("gpt-4", "gpt-3.5-turbo")
```

## Performance Issues

### Slow Spell Execution

**Problem**: Spells take too long to complete.

**Solutions**:

1. **Parallel Processing**
   ```lua
   -- BAD: Sequential processing
   local results = {}
   for i, item in ipairs(items) do
       results[i] = process_item(item)  -- Each waits for previous
   end
   
   -- GOOD: Parallel processing
   local promises = {}
   for i, item in ipairs(items) do
       promises[i] = promise.new(function(resolve)
           core.async(function()
               resolve(process_item(item))
           end)
       end)
   end
   results = promise.all(promises):await()
   ```

2. **Caching Expensive Operations**
   ```lua
   local cache = {}
   
   local function cached_llm_call(prompt)
       local cache_key = prompt:sub(1, 50)  -- Use first 50 chars as key
       
       if cache[cache_key] then
           log.debug("Cache hit for prompt")
           return cache[cache_key]
       end
       
       local result = llm.complete({
           model = "gpt-3.5-turbo",
           messages = {{role = "user", content = prompt}}
       })
       
       cache[cache_key] = result
       return result
   end
   ```

3. **Batching Requests**
   ```lua
   -- Process in batches instead of one-by-one
   local function process_in_batches(items, batch_size)
       batch_size = batch_size or 10
       local results = {}
       
       for i = 1, #items, batch_size do
           local batch = {}
           for j = 0, batch_size - 1 do
               if items[i + j] then
                   table.insert(batch, items[i + j])
               end
           end
           
           -- Process batch in parallel
           local batch_results = process_batch(batch)
           for k, v in ipairs(batch_results) do
               results[i + k - 1] = v
           end
       end
       
       return results
   end
   ```

### Memory Usage Issues

**Problem**: Spell uses too much memory or crashes with out-of-memory errors.

**Solutions**:

```lua
-- 1. Stream large data instead of loading all at once
local function process_large_file(filepath)
    local file = io.open(filepath, "r")
    local line_count = 0
    
    while true do
        local line = file:read("*line")
        if not line then break end
        
        -- Process line
        process_line(line)
        
        line_count = line_count + 1
        if line_count % 1000 == 0 then
            -- Force garbage collection periodically
            collectgarbage("collect")
        end
    end
    
    file:close()
end

-- 2. Clear large variables when done
local huge_data = load_huge_dataset()
process_data(huge_data)
huge_data = nil  -- Allow garbage collection
collectgarbage("collect")

-- 3. Use weak references for caches
local cache = setmetatable({}, {__mode = "v"})  -- Values are weak
```

## LLM-Specific Issues

### Empty or Truncated Responses

**Problem**: LLM returns empty or cut-off responses.

**Solutions**:

```lua
-- 1. Check and adjust max_tokens
local response = llm.complete({
    model = "gpt-4",
    messages = messages,
    max_tokens = 2000  -- Increase from default
})

-- 2. Validate response
if not response.content or #response.content < 10 then
    log.warn("Received empty/short response, retrying...")
    -- Retry with different parameters
    response = llm.complete({
        model = "gpt-4",
        messages = messages,
        temperature = 0.8,  -- Increase temperature
        max_tokens = 2000
    })
end

-- 3. Handle continuation
local function get_complete_response(prompt, min_length)
    local full_response = ""
    local continuation_prompt = prompt
    
    while #full_response < min_length do
        local response = llm.complete({
            model = "gpt-4",
            messages = {{role = "user", content = continuation_prompt}}
        })
        
        full_response = full_response .. response.content
        
        if response.finish_reason == "length" then
            continuation_prompt = "Continue from: " .. 
                                response.content:sub(-100)
        else
            break
        end
    end
    
    return full_response
end
```

### Model-Specific Errors

```lua
-- Create model-specific handlers
local model_handlers = {
    ["gpt-4"] = {
        max_tokens = 8192,
        error_handler = function(error)
            if error:find("context_length_exceeded") then
                return "reduce_context"
            end
        end
    },
    ["claude-3-opus"] = {
        max_tokens = 100000,
        error_handler = function(error)
            if error:find("overloaded") then
                return "retry_later"
            end
        end
    }
}

local function smart_llm_call(request)
    local handler = model_handlers[request.model] or {}
    
    -- Apply model-specific limits
    if handler.max_tokens and not request.max_tokens then
        request.max_tokens = math.min(1000, handler.max_tokens)
    end
    
    local success, result = pcall(llm.complete, request)
    
    if not success and handler.error_handler then
        local action = handler.error_handler(tostring(result))
        
        if action == "reduce_context" then
            -- Trim messages
            request.messages = trim_conversation(request.messages)
            return llm.complete(request)
        elseif action == "retry_later" then
            core.sleep(30)
            return llm.complete(request)
        end
    end
    
    if not success then
        error(result)
    end
    
    return result
end
```

## Tool Problems

### Tool Not Found

```lua
-- Solution: Check available tools and provide fallbacks
local function safe_tool_execute(tool_name, args, fallback_fn)
    -- Check if tool exists
    local available = tools.list()
    local tool_exists = false
    
    for _, tool in ipairs(available) do
        if tool.name == tool_name then
            tool_exists = true
            break
        end
    end
    
    if tool_exists then
        return tools.execute(tool_name, args)
    elseif fallback_fn then
        log.warn("Tool " .. tool_name .. " not found, using fallback")
        return fallback_fn(args)
    else
        error("Tool not found: " .. tool_name)
    end
end

-- Usage with fallback
local result = safe_tool_execute("advanced_calculator", 
    {expression = "2 + 2"}, 
    function(args)
        -- Simple fallback
        return tostring(load("return " .. args.expression)())
    end
)
```

### Tool Execution Failures

```lua
-- Robust tool execution with error handling
local function execute_tool_safely(tool_name, args, options)
    options = options or {}
    local max_retries = options.max_retries or 3
    local retry_delay = options.retry_delay or 1
    
    for attempt = 1, max_retries do
        local success, result = pcall(tools.execute, tool_name, args)
        
        if success then
            return result
        end
        
        local error_msg = tostring(result)
        log.warn("Tool execution failed", {
            tool = tool_name,
            attempt = attempt,
            error = error_msg
        })
        
        -- Check if error is retryable
        local retryable = error_msg:find("timeout") or 
                         error_msg:find("temporary") or
                         error_msg:find("rate limit")
        
        if not retryable or attempt == max_retries then
            if options.fallback then
                return options.fallback(args)
            else
                error("Tool execution failed: " .. error_msg)
            end
        end
        
        core.sleep(retry_delay * attempt)
    end
end
```

## State and Memory Issues

### State Corruption

**Problem**: State becomes corrupted or inconsistent.

**Solutions**:

```lua
-- 1. Use transactions for state updates
local function transactional_update(state_obj, updates_fn)
    -- Save current state
    local backup = {}
    for _, key in ipairs(state_obj:keys()) do
        backup[key] = state_obj:get(key)
    end
    
    -- Try updates
    local success, error = pcall(updates_fn, state_obj)
    
    if not success then
        -- Rollback on error
        log.error("State update failed, rolling back", {error = error})
        for key, value in pairs(backup) do
            state_obj:set(key, value)
        end
        error("Transaction failed: " .. tostring(error))
    end
end

-- 2. Validate state consistency
local function validate_state(state_obj, schema)
    for field, validator in pairs(schema) do
        local value = state_obj:get(field)
        local valid, error = validator(value)
        
        if not valid then
            error("State validation failed for " .. field .. ": " .. error)
        end
    end
end

-- Usage
local state_schema = {
    user_id = function(v) 
        return type(v) == "string" and #v > 0, "user_id must be non-empty string" 
    end,
    credits = function(v) 
        return type(v) == "number" and v >= 0, "credits must be non-negative number" 
    end
}

transactional_update(app_state, function(state)
    state:set("credits", state:get("credits") - 10)
    validate_state(state, state_schema)
end)
```

### State Persistence Issues

```lua
-- Robust state persistence
local function save_state_safely(state_obj, filepath)
    -- Create backup first
    local backup_path = filepath .. ".backup"
    
    if tools.file_exists(filepath) then
        tools.file_copy(filepath, backup_path)
    end
    
    -- Try to save
    local success, error = pcall(function()
        local data = {}
        for _, key in ipairs(state_obj:keys()) do
            data[key] = state_obj:get(key)
        end
        
        local json = data.to_json(data, {pretty = true})
        tools.file_write(filepath .. ".tmp", json)
        
        -- Verify written data
        local verification = data.from_json(tools.file_read(filepath .. ".tmp"))
        assert(verification, "Failed to verify saved state")
        
        -- Atomic rename
        tools.file_move(filepath .. ".tmp", filepath)
    end)
    
    if not success then
        -- Restore from backup
        if tools.file_exists(backup_path) then
            tools.file_copy(backup_path, filepath)
        end
        error("State save failed: " .. tostring(error))
    end
    
    -- Clean up backup
    if tools.file_exists(backup_path) then
        tools.file_delete(backup_path)
    end
end
```

## Async and Promise Issues

### Promise Deadlocks

**Problem**: Promises never resolve, causing spell to hang.

**Solutions**:

```lua
-- 1. Always use timeouts
local function with_timeout(promise_fn, timeout_seconds)
    return promise.race({
        promise_fn(),
        promise.new(function(resolve, reject)
            core.async(function()
                core.sleep(timeout_seconds)
                reject("Operation timed out after " .. timeout_seconds .. " seconds")
            end)
        end)
    })
end

-- Usage
local result = with_timeout(function()
    return expensive_async_operation()
end, 30):await()

-- 2. Debug hanging promises
local active_promises = {}

local function tracked_promise(name, promise_fn)
    active_promises[name] = {
        started = os.time(),
        status = "running"
    }
    
    return promise.new(function(resolve, reject)
        promise_fn(
            function(value)
                active_promises[name].status = "resolved"
                resolve(value)
            end,
            function(error)
                active_promises[name].status = "rejected"
                reject(error)
            end
        )
    end)
end

-- Check for hanging promises
core.async(function()
    while true do
        core.sleep(10)
        for name, info in pairs(active_promises) do
            if info.status == "running" and 
               os.time() - info.started > 60 then
                log.warn("Promise '" .. name .. "' has been running for over 60 seconds")
            end
        end
    end
end)
```

### Coroutine Errors

```lua
-- Handle coroutine errors gracefully
local function safe_async(fn, error_handler)
    return core.async(function()
        local success, result = pcall(fn)
        if not success then
            if error_handler then
                error_handler(result)
            else
                log.error("Async error", {error = tostring(result)})
            end
        end
        return result
    end)
end

-- Usage with error recovery
safe_async(function()
    -- Async work that might fail
    local data = fetch_data()
    process_data(data)
end, function(error)
    -- Handle error
    log.error("Processing failed", {error = error})
    -- Try alternative approach
    use_cached_data()
end)
```

## Security and Permission Issues

### Permission Denied Errors

```lua
-- Handle permission errors gracefully
local function check_permissions(operation, resource)
    local permissions = {
        file_write = function(path)
            -- Check if path is in allowed directory
            local allowed_dirs = {"/tmp/", "./output/", "./data/"}
            for _, dir in ipairs(allowed_dirs) do
                if path:find("^" .. dir) then
                    return true
                end
            end
            return false, "Write not allowed to: " .. path
        end,
        
        web_fetch = function(url)
            -- Check if URL is allowed
            local blocked_domains = {"internal.company.com", "localhost"}
            for _, domain in ipairs(blocked_domains) do
                if url:find(domain) then
                    return false, "Access to " .. domain .. " is blocked"
                end
            end
            return true
        end
    }
    
    local checker = permissions[operation]
    if checker then
        return checker(resource)
    end
    
    return true  -- Allow by default
end

-- Wrap operations with permission checks
local function safe_file_write(path, content)
    local allowed, error = check_permissions("file_write", path)
    if not allowed then
        error("Permission denied: " .. error)
    end
    
    return tools.file_write(path, content)
end
```

### Sandbox Violations

```lua
-- Detect and handle sandbox violations
local function detect_sandbox_violation(code)
    local dangerous_patterns = {
        "os%.execute",
        "io%.popen",
        "load%(.*%)",
        "require%s*%(%s*[\"']os[\"']%s*%)",
        "debug%.",
        "_G%[",
        "rawset",
        "rawget"
    }
    
    for _, pattern in ipairs(dangerous_patterns) do
        if code:find(pattern) then
            return true, "Dangerous pattern detected: " .. pattern
        end
    end
    
    return false
end

-- Safe code execution
local function execute_user_code(code)
    local violated, reason = detect_sandbox_violation(code)
    if violated then
        error("Security violation: " .. reason)
    end
    
    -- Execute in restricted environment
    local env = {
        print = print,
        math = math,
        string = string,
        table = table,
        -- Limited set of safe functions
    }
    
    local fn, err = load(code, "user_code", "t", env)
    if not fn then
        error("Code compilation error: " .. err)
    end
    
    return fn()
end
```

## Getting Help

### 1. Enable Detailed Logging

```bash
# Run with debug logging
llmspell run my-spell.lua --log-level debug

# Or in the spell
log.set_level("debug")
```

### 2. Create Minimal Reproducible Example

```lua
-- minimal-repro.lua
-- Demonstrates the issue with minimal code

-- Setup
log.set_level("debug")
print("go-llmspell version: " .. spell.version())
print("Lua version: " .. _VERSION)

-- The problematic code
local function reproduce_issue()
    -- Minimal code that shows the problem
    local result = llm.complete({
        model = "gpt-3.5-turbo",
        messages = {{role = "user", content = "test"}}
    })
    -- Issue happens here
    print(result.nonexistent_field)  -- This will error
end

-- Run with error handling
local success, error = pcall(reproduce_issue)
if not success then
    print("Error reproduced: " .. tostring(error))
    print("Stack trace: " .. debug.traceback())
end
```

### 3. Collect Diagnostic Information

```lua
-- diagnostic-info.lua
-- Collects system information for bug reports

local function collect_diagnostics()
    local info = {
        spell_version = spell.version(),
        lua_version = _VERSION,
        os = os.getenv("OS") or "unknown",
        timestamp = os.date(),
        available_tools = tools.list(),
        available_models = llm.list_models(),
        memory_usage = collectgarbage("count") .. " KB",
        environment_vars = {
            OPENAI_API_KEY = os.getenv("OPENAI_API_KEY") and "set" or "not set",
            ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY") and "set" or "not set"
        }
    }
    
    return info
end

-- Save diagnostics
local diagnostics = collect_diagnostics()
tools.file_write("diagnostics.json", data.to_json(diagnostics, {pretty = true}))
print("Diagnostics saved to diagnostics.json")
```

### 4. Common Recovery Strategies

```lua
-- Master error recovery function
local function with_recovery(operation, recovery_strategies)
    local attempt = 0
    local max_attempts = #recovery_strategies + 1
    
    while attempt < max_attempts do
        attempt = attempt + 1
        
        local success, result = pcall(operation)
        
        if success then
            return result
        end
        
        local error_msg = tostring(result)
        log.warn("Operation failed (attempt " .. attempt .. ")", {error = error_msg})
        
        if attempt < max_attempts then
            local strategy = recovery_strategies[attempt]
            if strategy then
                log.info("Applying recovery strategy: " .. strategy.name)
                strategy.action(error_msg)
            end
        end
    end
    
    error("All recovery attempts failed")
end

-- Usage
local result = with_recovery(
    function()
        return llm.complete({
            model = "gpt-4",
            messages = {{role = "user", content = params.prompt}}
        })
    end,
    {
        {
            name = "retry_with_backoff",
            action = function() core.sleep(5) end
        },
        {
            name = "fallback_model",
            action = function()
                params.model = "gpt-3.5-turbo"
            end
        },
        {
            name = "reduce_complexity",
            action = function()
                params.prompt = params.prompt:sub(1, 500)
            end
        }
    }
)
```

## Quick Reference Card

### Debug Checklist

- [ ] Enable debug logging: `--log-level debug`
- [ ] Add parameter validation at spell start
- [ ] Check for nil values before accessing properties
- [ ] Implement proper error handling with pcall
- [ ] Add timeouts to async operations
- [ ] Use type checking for parameters
- [ ] Implement retry logic for network operations
- [ ] Add resource cleanup in finally blocks
- [ ] Monitor memory usage for large operations
- [ ] Validate external data before processing

### Performance Checklist

- [ ] Use parallel processing for independent operations
- [ ] Implement caching for expensive operations
- [ ] Batch API calls when possible
- [ ] Stream large files instead of loading entirely
- [ ] Clear large variables when done
- [ ] Use weak references for caches
- [ ] Profile slow sections with timing
- [ ] Optimize prompt length for LLM calls
- [ ] Use appropriate models for tasks
- [ ] Implement progress reporting for long operations

### Security Checklist

- [ ] Validate all user inputs
- [ ] Sanitize file paths
- [ ] Check permissions before operations
- [ ] Use sandboxed execution for untrusted code
- [ ] Never log sensitive information
- [ ] Implement rate limiting
- [ ] Validate URLs before fetching
- [ ] Use environment variables for secrets
- [ ] Implement proper error messages (no stack traces to users)
- [ ] Regular security audits of spell code

Remember: When in doubt, add more logging and validate your assumptions!