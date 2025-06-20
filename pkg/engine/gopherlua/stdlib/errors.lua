-- ABOUTME: Error Handling & Recovery Library for go-llmspell Lua standard library
-- ABOUTME: Provides unified access to error handling, retry logic, circuit breakers, and recovery strategies

local errors = {}

-- Import promise library for async error handling
-- local promise = _G.promise or require("promise")

-- Compatibility for unpack function (Lua 5.1 vs 5.3+)
local unpack = unpack or table.unpack -- luacheck: ignore 113

-- Internal state
local error_handlers = {}
local recovery_strategies = {}
local circuit_breakers = {}
local error_categories = {}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to get error utilities bridge
local function get_error_bridge()
    if not _G.util_errors then
        error("Error utilities bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.util_errors
end

-- Enhanced Error Handling

-- Try-catch-finally pattern for Lua
function errors.try(func, catch_func, finally_func)
    validate_required(func, "function")

    if type(func) ~= "function" then
        error("first argument must be a function")
    end

    local success, result, error_obj

    -- Execute the main function
    success, result = pcall(func)

    if not success then
        -- Handle error
        error_obj = result
        if catch_func and type(catch_func) == "function" then
            local catch_success, catch_result = pcall(catch_func, error_obj)
            if catch_success then
                result = catch_result
                success = true
            else
                -- If catch function also fails, use the catch error
                error_obj = catch_result
            end
        end
    end

    -- Always execute finally block
    if finally_func and type(finally_func) == "function" then
        pcall(finally_func)
    end

    if success then
        return result
    else
        error(error_obj)
    end
end

-- Create custom error with context
function errors.create(message, code, context)
    validate_required(message, "error message")

    if type(message) ~= "string" then
        error("error message must be a string")
    end

    local bridge = get_error_bridge()
    local success, error_obj =
        pcall(bridge.createError, message, code or "CUSTOM_ERROR", context or {})

    if not success then
        error("Failed to create error: " .. tostring(error_obj))
    end

    return error_obj
end

-- Wrap error with additional context
function errors.wrap(original_error, context)
    validate_required(original_error, "original error")
    validate_required(context, "error context")

    if type(context) ~= "table" then
        error("context must be a table")
    end

    local bridge = get_error_bridge()
    local success, wrapped_error = pcall(bridge.wrapError, original_error, context)

    if not success then
        error("Failed to wrap error: " .. tostring(wrapped_error))
    end

    return wrapped_error
end

-- Chain multiple errors together
function errors.chain(errors_array)
    validate_required(errors_array, "errors array")

    if type(errors_array) ~= "table" then
        error("errors must be an array")
    end

    local bridge = get_error_bridge()
    local success, chained_error = pcall(bridge.chainErrors, errors_array)

    if not success then
        error("Failed to chain errors: " .. tostring(chained_error))
    end

    return chained_error
end

-- Retry and Recovery Mechanisms

-- Retry function with configurable options
function errors.retry(func, options)
    validate_required(func, "function")

    if type(func) ~= "function" then
        error("first argument must be a function")
    end

    options = options or {}
    local max_attempts = options.max_attempts or 3
    local backoff_strategy = options.strategy or "exponential"
    local initial_delay = options.initial_delay or 1000 -- milliseconds
    local max_delay = options.max_delay or 30000 -- 30 seconds
    local jitter = options.jitter ~= false -- default true

    local bridge = get_error_bridge()

    -- Create backoff strategy
    local strategy
    if backoff_strategy == "exponential" then
        local success_exp, strategy_exp =
            pcall(bridge.createExponentialBackoffStrategy, initial_delay, max_delay, jitter)
        if not success_exp then
            error("Failed to create exponential backoff strategy: " .. tostring(strategy_exp))
        end
        strategy = strategy_exp
    elseif backoff_strategy == "linear" then
        local success_lin, strategy_lin =
            pcall(bridge.createLinearBackoffStrategy, initial_delay, max_delay)
        if not success_lin then
            error("Failed to create linear backoff strategy: " .. tostring(strategy_lin))
        end
        strategy = strategy_lin
    else
        error("Unknown backoff strategy: " .. tostring(backoff_strategy))
    end

    -- Perform retries
    local last_error
    for attempt = 1, max_attempts do
        local success, result = pcall(func)

        if success then
            return result
        end

        last_error = result

        -- Check if error is retryable
        local retryable_success, is_retryable = pcall(bridge.isRetryableError, last_error)
        if retryable_success and not is_retryable then
            break -- Don't retry non-retryable errors
        end

        -- Don't delay on the last attempt
        if attempt < max_attempts then
            local delay_success, delay = pcall(bridge.calculateBackoffDelay, strategy, attempt)
            if delay_success and delay > 0 then
                -- Sleep for the calculated delay (simplified - in real implementation would use proper sleep)
                os.execute("sleep " .. math.floor(delay / 1000))
            end
        end
    end

    -- All attempts failed
    error("Retry attempts exhausted. Last error: " .. tostring(last_error))
end

-- Circuit breaker pattern implementation
function errors.circuit_breaker(func, config)
    validate_required(func, "function")

    if type(func) ~= "function" then
        error("first argument must be a function")
    end

    config = config or {}
    local threshold = config.failure_threshold or 5
    local timeout = config.timeout or 60000 -- 60 seconds
    local reset_timeout = config.reset_timeout or 30000 -- 30 seconds

    local bridge = get_error_bridge()
    local success, circuit_breaker =
        pcall(bridge.createCircuitBreakerStrategy, threshold, timeout, reset_timeout)

    if not success then
        error("Failed to create circuit breaker: " .. tostring(circuit_breaker))
    end

    -- Store circuit breaker for reuse
    local func_id = tostring(func)
    circuit_breakers[func_id] = circuit_breaker

    return function(...)
        local args = { ... }
        local execute_success, result = pcall(
            bridge.executeWithCircuitBreaker,
            circuit_breaker,
            function()
                return func(unpack(args))
            end
        )

        if not execute_success then
            error("Circuit breaker execution failed: " .. tostring(result))
        end

        return result
    end
end

-- Fallback strategy implementation
function errors.fallback(primary_func, fallback_func)
    validate_required(primary_func, "primary function")
    validate_required(fallback_func, "fallback function")

    if type(primary_func) ~= "function" then
        error("primary function must be a function")
    end

    if type(fallback_func) ~= "function" then
        error("fallback function must be a function")
    end

    return function(...)
        local args = { ... }
        local success, result = pcall(primary_func, unpack(args))

        if success then
            return result
        else
            -- Log primary function failure
            errors.log_error("fallback_triggered", {
                primary_error = result,
                timestamp = os.time(),
            })

            -- Execute fallback
            local fallback_success, fallback_result = pcall(fallback_func, unpack(args))
            if fallback_success then
                return fallback_result
            else
                error(
                    "Both primary and fallback functions failed. Primary: "
                        .. tostring(result)
                        .. ", Fallback: "
                        .. tostring(fallback_result)
                )
            end
        end
    end
end

-- Error Categorization and Reporting

-- Categorize error by type and characteristics
function errors.categorize(error_obj)
    validate_required(error_obj, "error object")

    local bridge = get_error_bridge()
    local success, category = pcall(bridge.categorizeError, error_obj)

    if not success then
        error("Failed to categorize error: " .. tostring(category))
    end

    return category
end

-- Check if error is retryable
function errors.is_retryable(error_obj)
    validate_required(error_obj, "error object")

    local bridge = get_error_bridge()
    local success, retryable = pcall(bridge.isRetryableError, error_obj)

    if not success then
        return false -- Conservative default
    end

    return retryable
end

-- Check if error is fatal (non-recoverable)
function errors.is_fatal(error_obj)
    validate_required(error_obj, "error object")

    local bridge = get_error_bridge()
    local success, fatal = pcall(bridge.isFatalError, error_obj)

    if not success then
        return false -- Conservative default
    end

    return fatal
end

-- Aggregate multiple errors
function errors.aggregate(errors_array)
    validate_required(errors_array, "errors array")

    if type(errors_array) ~= "table" then
        error("errors must be an array")
    end

    local bridge = get_error_bridge()
    local success, aggregator = pcall(bridge.createErrorAggregator)

    if not success then
        error("Failed to create error aggregator: " .. tostring(aggregator))
    end

    -- Add all errors to aggregator
    for _, err in ipairs(errors_array) do
        local add_success, _ = pcall(bridge.addErrorToAggregator, aggregator, err)
        -- Continue with other errors even if one fails to add (graceful degradation)
        _ = add_success
    end

    local finalize_success, aggregated_error = pcall(bridge.finalizeErrorAggregator, aggregator)
    if not finalize_success then
        error("Failed to finalize error aggregation: " .. tostring(aggregated_error))
    end

    return aggregated_error
end

-- Serialization and Deserialization

-- Convert error to JSON string
function errors.to_json(error_obj)
    validate_required(error_obj, "error object")

    local bridge = get_error_bridge()
    local success, json_string = pcall(bridge.errorToJSON, error_obj)

    if not success then
        error("Failed to serialize error to JSON: " .. tostring(json_string))
    end

    return json_string
end

-- Create error from JSON string
function errors.from_json(json_string)
    validate_required(json_string, "JSON string")

    if type(json_string) ~= "string" then
        error("JSON input must be a string")
    end

    local bridge = get_error_bridge()
    local success, error_obj = pcall(bridge.errorFromJSON, json_string)

    if not success then
        error("Failed to deserialize error from JSON: " .. tostring(error_obj))
    end

    return error_obj
end

-- Error Context Management

-- Get error context information
function errors.get_context(error_obj)
    validate_required(error_obj, "error object")

    local bridge = get_error_bridge()
    local success, context = pcall(bridge.getErrorContext, error_obj)

    if not success then
        return {} -- Return empty context if unable to get
    end

    return context or {}
end

-- Add context to existing error
function errors.add_context(error_obj, key, value)
    validate_required(error_obj, "error object")
    validate_required(key, "context key")

    if type(key) ~= "string" then
        error("context key must be a string")
    end

    local bridge = get_error_bridge()
    local success, updated_error = pcall(bridge.addErrorContext, error_obj, key, value)

    if not success then
        error("Failed to add context to error: " .. tostring(updated_error))
    end

    return updated_error
end

-- Error Event Handling

-- Log error event for monitoring
function errors.log_error(error_type, metadata)
    validate_required(error_type, "error type")
    validate_required(metadata, "error metadata")

    if type(error_type) ~= "string" then
        error("error type must be a string")
    end

    if type(metadata) ~= "table" then
        error("metadata must be a table")
    end

    local bridge = get_error_bridge()
    local success, _ = pcall(bridge.emitErrorEvent, error_type, metadata)
    -- Don't throw error for logging failures, just continue (non-blocking)
    _ = success

    return true
end

-- Subscribe to error events
function errors.subscribe_to_errors(error_types, handler)
    validate_required(error_types, "error types")
    validate_required(handler, "error handler")

    if type(error_types) ~= "table" then
        error("error types must be an array")
    end

    if type(handler) ~= "function" then
        error("handler must be a function")
    end

    local bridge = get_error_bridge()
    local success, subscription_id = pcall(bridge.subscribeToErrorEvents, error_types, handler)

    if not success then
        error("Failed to subscribe to error events: " .. tostring(subscription_id))
    end

    return subscription_id
end

-- Recovery Strategy Management

-- Create custom recovery strategy
function errors.create_recovery_strategy(strategy_type, config)
    validate_required(strategy_type, "strategy type")

    if type(strategy_type) ~= "string" then
        error("strategy type must be a string")
    end

    config = config or {}
    local bridge = get_error_bridge()
    local success, strategy

    if strategy_type == "exponential_backoff" then
        success, strategy = pcall(
            bridge.createExponentialBackoffStrategy,
            config.initial_delay or 1000,
            config.max_delay or 30000,
            config.jitter ~= false
        )
    elseif strategy_type == "linear_backoff" then
        success, strategy = pcall(
            bridge.createLinearBackoffStrategy,
            config.initial_delay or 1000,
            config.max_delay or 30000
        )
    elseif strategy_type == "circuit_breaker" then
        success, strategy = pcall(
            bridge.createCircuitBreakerStrategy,
            config.failure_threshold or 5,
            config.timeout or 60000,
            config.reset_timeout or 30000
        )
    elseif strategy_type == "fallback" then
        success, strategy = pcall(bridge.createFallbackStrategy, config.fallback_value)
    else
        error("Unknown recovery strategy type: " .. strategy_type)
    end

    if not success then
        error("Failed to create recovery strategy: " .. tostring(strategy))
    end

    return strategy
end

-- Register custom error category
function errors.register_category(category_name, matcher_func)
    validate_required(category_name, "category name")
    validate_required(matcher_func, "matcher function")

    if type(category_name) ~= "string" then
        error("category name must be a string")
    end

    if type(matcher_func) ~= "function" then
        error("matcher function must be a function")
    end

    local bridge = get_error_bridge()
    local success, _ = pcall(bridge.registerErrorCategory, category_name, matcher_func)

    if not success then
        error("Failed to register error category: " .. tostring(_))
    end

    error_categories[category_name] = matcher_func
    return true
end

-- Utility Functions

-- Create safe wrapper for any function
function errors.safe(func, default_value)
    validate_required(func, "function")

    if type(func) ~= "function" then
        error("first argument must be a function")
    end

    return function(...)
        local args = { ... }
        local success, result = pcall(func, unpack(args))

        if success then
            return result
        else
            return default_value
        end
    end
end

-- Create timeout wrapper for functions
function errors.timeout(func, timeout_ms)
    validate_required(func, "function")
    validate_required(timeout_ms, "timeout in milliseconds")

    if type(func) ~= "function" then
        error("first argument must be a function")
    end

    if type(timeout_ms) ~= "number" or timeout_ms <= 0 then
        error("timeout must be a positive number")
    end

    return function(...)
        local args = { ... }
        local start_time = os.clock() * 1000 -- Convert to milliseconds

        local success, result = pcall(func, unpack(args))

        local elapsed_time = (os.clock() * 1000) - start_time

        if elapsed_time > timeout_ms then
            error("Function execution timed out after " .. timeout_ms .. "ms")
        end

        if success then
            return result
        else
            error(result)
        end
    end
end

-- Get system error information
function errors.get_system_info()
    return {
        lua_version = _VERSION,
        bridges_available = {
            errors = _G.util_errors ~= nil,
        },
        error_categories = error_categories,
        active_circuit_breakers = #circuit_breakers,
        recovery_strategies = #recovery_strategies,
    }
end

-- Clean up resources
function errors.cleanup()
    -- Clear circuit breakers
    for id, _ in pairs(circuit_breakers) do
        circuit_breakers[id] = nil
    end

    -- Clear recovery strategies
    for id, _ in pairs(recovery_strategies) do
        recovery_strategies[id] = nil
    end

    -- Clear error handlers
    for id, _ in pairs(error_handlers) do
        error_handlers[id] = nil
    end

    errors.log_error("errors_cleanup_completed", {
        timestamp = os.time(),
    })
end

-- Export the module
return errors
