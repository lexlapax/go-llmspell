-- ABOUTME: Observability & Monitoring Library for go-llmspell Lua standard library
-- ABOUTME: Provides unified access to metrics, tracing, logging, and performance monitoring

local observability = {}

-- Import promise library for async operations (reserved for future use)
-- local promise = _G.promise or require("promise")

-- Internal state
local active_spans = {}
-- local performance_timers = {}  -- Reserved for future use
local custom_loggers = {}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to get metrics bridge
local function get_metrics_bridge()
    if not _G.metrics then
        error("Metrics bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.metrics
end

-- Helper function to get tracing bridge
local function get_tracing_bridge()
    if not _G.tracing then
        error("Tracing bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.tracing
end

-- Helper function to get slog bridge
local function get_slog_bridge()
    if not _G.slog then
        error("Slog bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.slog
end

-- Helper function to get events bridge
local function get_events_bridge()
    if not _G.events then
        error("Events bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.events
end

-- Helper function to get guardrails bridge
local function get_guardrails_bridge()
    -- Guardrails bridge is optional, return nil if not available
    return _G.guardrails
end

-- Helper function to generate unique IDs (reserved for future use)
-- local function generate_id(prefix)
--     return (prefix or "id")
--         .. "_"
--         .. tostring(os.time())
--         .. "_"
--         .. tostring(math.random(1000, 9999))
-- end

-- Metrics Management

-- Create a counter metric
function observability.counter(name, description, tags)
    validate_required(name, "metric name")

    if type(name) ~= "string" then
        error("metric name must be a string")
    end

    local bridge = get_metrics_bridge()
    local success, result = pcall(bridge.createCounter, name, description or "", tags or {})

    if not success then
        error("Failed to create counter: " .. tostring(result))
    end

    return {
        name = name,
        increment = function(value, tags_override)
            local inc_tags = tags_override or tags or {}
            local inc_value = value or 1

            local inc_success, err = pcall(bridge.incrementCounter, name, inc_value, inc_tags)
            if not inc_success then
                error("Failed to increment counter: " .. tostring(err))
            end
        end,
        get = function()
            local get_success, value = pcall(bridge.getCounter, name)
            if not get_success then
                error("Failed to get counter value: " .. tostring(value))
            end
            return value
        end,
        reset = function()
            local reset_success, err = pcall(bridge.resetCounter, name)
            if not reset_success then
                error("Failed to reset counter: " .. tostring(err))
            end
        end,
    }
end

-- Create a gauge metric
function observability.gauge(name, description, tags)
    validate_required(name, "metric name")

    if type(name) ~= "string" then
        error("metric name must be a string")
    end

    local bridge = get_metrics_bridge()
    local success, result = pcall(bridge.createGauge, name, description or "", tags or {})

    if not success then
        error("Failed to create gauge: " .. tostring(result))
    end

    return {
        name = name,
        set = function(value, tags_override)
            validate_required(value, "gauge value")
            local gauge_tags = tags_override or tags or {}

            local set_success, err = pcall(bridge.setGauge, name, value, gauge_tags)
            if not set_success then
                error("Failed to set gauge: " .. tostring(err))
            end
        end,
        increase = function(value, tags_override)
            local inc_value = value or 1
            local gauge_tags = tags_override or tags or {}

            local inc_success, err = pcall(bridge.increaseGauge, name, inc_value, gauge_tags)
            if not inc_success then
                error("Failed to increase gauge: " .. tostring(err))
            end
        end,
        decrease = function(value, tags_override)
            local dec_value = value or 1
            local gauge_tags = tags_override or tags or {}

            local dec_success, err = pcall(bridge.decreaseGauge, name, dec_value, gauge_tags)
            if not dec_success then
                error("Failed to decrease gauge: " .. tostring(err))
            end
        end,
        get = function()
            local get_success, value = pcall(bridge.getGauge, name)
            if not get_success then
                error("Failed to get gauge value: " .. tostring(value))
            end
            return value
        end,
    }
end

-- Create a timer metric
function observability.timer(name, description, tags)
    validate_required(name, "metric name")

    if type(name) ~= "string" then
        error("metric name must be a string")
    end

    local bridge = get_metrics_bridge()
    local success, result = pcall(bridge.createTimer, name, description or "", tags or {})

    if not success then
        error("Failed to create timer: " .. tostring(result))
    end

    return {
        name = name,
        record = function(duration, tags_override)
            validate_required(duration, "duration")
            local timer_tags = tags_override or tags or {}

            local record_success, err = pcall(bridge.recordTimer, name, duration, timer_tags)
            if not record_success then
                error("Failed to record timer: " .. tostring(err))
            end
        end,
        start = function()
            local start_time = os.clock()
            return {
                stop = function(tags_override)
                    local duration = (os.clock() - start_time) * 1000 -- Convert to milliseconds
                    local timer_tags = tags_override or tags or {}

                    local stop_success, err = pcall(bridge.recordTimer, name, duration, timer_tags)
                    if not stop_success then
                        error("Failed to record timer: " .. tostring(err))
                    end
                    return duration
                end,
            }
        end,
        get_stats = function()
            local stats_success, stats = pcall(bridge.getTimerStats, name)
            if not stats_success then
                error("Failed to get timer stats: " .. tostring(stats))
            end
            return stats
        end,
    }
end

-- Create a ratio counter metric
function observability.ratio_counter(name, description, tags)
    validate_required(name, "metric name")

    if type(name) ~= "string" then
        error("metric name must be a string")
    end

    local bridge = get_metrics_bridge()
    local success, result = pcall(bridge.createRatioCounter, name, description or "", tags or {})

    if not success then
        error("Failed to create ratio counter: " .. tostring(result))
    end

    return {
        name = name,
        increment_numerator = function(value, tags_override)
            local inc_value = value or 1
            local ratio_tags = tags_override or tags or {}

            local num_success, err =
                pcall(bridge.incrementRatioNumerator, name, inc_value, ratio_tags)
            if not num_success then
                error("Failed to increment ratio numerator: " .. tostring(err))
            end
        end,
        increment_denominator = function(value, tags_override)
            local inc_value = value or 1
            local ratio_tags = tags_override or tags or {}

            local den_success, err =
                pcall(bridge.incrementRatioDenominator, name, inc_value, ratio_tags)
            if not den_success then
                error("Failed to increment ratio denominator: " .. tostring(err))
            end
        end,
        get_ratio = function()
            local ratio_success, ratio = pcall(bridge.getRatio, name)
            if not ratio_success then
                error("Failed to get ratio: " .. tostring(ratio))
            end
            return ratio
        end,
        get_counts = function()
            local counts_success, counts = pcall(bridge.getRatioCounts, name)
            if not counts_success then
                error("Failed to get ratio counts: " .. tostring(counts))
            end
            return counts
        end,
    }
end

-- Performance Monitoring

-- Track function execution time
function observability.track(func, name, options)
    validate_required(func, "function")

    if type(func) ~= "function" then
        error("func must be a function")
    end

    options = options or {}
    local metric_name = name or options.name or "tracked_function"
    local include_args = options.include_args or false
    local include_result = options.include_result or false
    local auto_metric = options.auto_metric ~= false -- Default to true

    -- Create timer metric if auto_metric is enabled
    local timer_metric = nil
    if auto_metric then
        timer_metric =
            observability.timer(metric_name .. "_duration", "Execution time for " .. metric_name)
    end

    return function(...)
        local args = { ... }
        local start_time = os.clock()

        -- Start tracing span if available
        local span_id = nil
        if get_tracing_bridge() then
            local success, result = pcall(get_tracing_bridge().startSpan, metric_name, {
                operation = "function_call",
                function_name = metric_name,
            })
            if success then
                span_id = result
                active_spans[span_id] = true
            end
        end

        local success, result
        if #args == 0 then
            success, result = pcall(func)
        elseif #args == 1 then
            success, result = pcall(func, args[1])
        elseif #args == 2 then
            success, result = pcall(func, args[1], args[2])
        elseif #args == 3 then
            success, result = pcall(func, args[1], args[2], args[3])
        elseif #args == 4 then
            success, result = pcall(func, args[1], args[2], args[3], args[4])
        elseif #args == 5 then
            success, result = pcall(func, args[1], args[2], args[3], args[4], args[5])
        else
            success, result = pcall(func, args[1], args[2], args[3], args[4], args[5])
        end
        local duration = (os.clock() - start_time) * 1000 -- Convert to milliseconds

        -- End tracing span
        if span_id and active_spans[span_id] then
            local tracing = get_tracing_bridge()
            if success then
                pcall(tracing.setSpanStatus, span_id, "ok")
            else
                pcall(tracing.setSpanStatus, span_id, "error")
                pcall(tracing.recordError, span_id, tostring(result))
            end
            pcall(tracing.endSpan, span_id)
            active_spans[span_id] = nil
        end

        -- Record metrics
        if timer_metric then
            timer_metric.record(duration, {
                success = tostring(success),
                function_name = metric_name,
            })
        end

        -- Log execution details if requested
        if include_args or include_result then
            local slog = get_slog_bridge()
            local log_data = {
                function_name = metric_name,
                duration_ms = duration,
                success = success,
            }

            if include_args then
                log_data.args = args
            end

            if include_result and success then
                log_data.result = result
            elseif not success then
                log_data.error = tostring(result)
            end

            if success then
                pcall(slog.info, "Function execution completed", log_data)
            else
                pcall(slog.error, "Function execution failed", log_data)
            end
        end

        if not success then
            error(result)
        end

        return result
    end
end

-- Distributed Tracing

-- Start a new tracing span
function observability.start_span(name, options)
    validate_required(name, "span name")

    if type(name) ~= "string" then
        error("span name must be a string")
    end

    options = options or {}
    local bridge = get_tracing_bridge()

    local success, span_id = pcall(bridge.startSpan, name, options)
    if not success then
        error("Failed to start span: " .. tostring(span_id))
    end

    active_spans[span_id] = true

    return {
        id = span_id,
        name = name,
        add_attribute = function(key, value)
            validate_required(key, "attribute key")
            local attr_success, err = pcall(bridge.setSpanAttribute, span_id, key, value)
            if not attr_success then
                error("Failed to add span attribute: " .. tostring(err))
            end
        end,
        add_event = function(event_name, attributes)
            validate_required(event_name, "event name")
            local event_success, err =
                pcall(bridge.addSpanEvent, span_id, event_name, attributes or {})
            if not event_success then
                error("Failed to add span event: " .. tostring(err))
            end
        end,
        set_status = function(status, description)
            validate_required(status, "status")
            local status_success, err = pcall(bridge.setSpanStatus, span_id, status, description)
            if not status_success then
                error("Failed to set span status: " .. tostring(err))
            end
        end,
        record_error = function(error_msg)
            validate_required(error_msg, "error message")
            local error_success, err = pcall(bridge.recordError, span_id, error_msg)
            if not error_success then
                error("Failed to record span error: " .. tostring(err))
            end
        end,
        finish = function()
            if active_spans[span_id] then
                local end_success, err = pcall(bridge.endSpan, span_id)
                if not end_success then
                    error("Failed to end span: " .. tostring(err))
                end
                active_spans[span_id] = nil
            end
        end,
    }
end

-- Create a traced version of a function
function observability.trace(func, span_name, options)
    validate_required(func, "function")

    if type(func) ~= "function" then
        error("func must be a function")
    end

    local name = span_name or "traced_function"
    options = options or {}

    return function(...)
        local span = observability.start_span(name, options)

        local success, result = pcall(func, ...)

        if success then
            span.set_status("ok")
        else
            span.set_status("error")
            span.record_error(tostring(result))
        end

        span.finish()

        if not success then
            error(result)
        end

        return result
    end
end

-- Structured Logging

-- Create a custom logger with context
function observability.logger(name, options)
    validate_required(name, "logger name")

    if type(name) ~= "string" then
        error("logger name must be a string")
    end

    options = options or {}
    local level = options.level or "info"
    local context = options.context or {}
    local bridge = get_slog_bridge()

    -- Store logger configuration
    custom_loggers[name] = {
        level = level,
        context = context,
    }

    return {
        name = name,
        debug = function(message, data)
            validate_required(message, "message")
            local log_data = {}
            for k, v in pairs(context) do
                log_data[k] = v
            end
            if data then
                for k, v in pairs(data) do
                    log_data[k] = v
                end
            end
            log_data.logger = name

            local debug_success, err = pcall(bridge.debug, message, log_data)
            if not debug_success then
                error("Failed to log debug message: " .. tostring(err))
            end
        end,
        info = function(message, data)
            validate_required(message, "message")
            local log_data = {}
            for k, v in pairs(context) do
                log_data[k] = v
            end
            if data then
                for k, v in pairs(data) do
                    log_data[k] = v
                end
            end
            log_data.logger = name

            local info_success, err = pcall(bridge.info, message, log_data)
            if not info_success then
                error("Failed to log info message: " .. tostring(err))
            end
        end,
        warn = function(message, data)
            validate_required(message, "message")
            local log_data = {}
            for k, v in pairs(context) do
                log_data[k] = v
            end
            if data then
                for k, v in pairs(data) do
                    log_data[k] = v
                end
            end
            log_data.logger = name

            local warn_success, err = pcall(bridge.warn, message, log_data)
            if not warn_success then
                error("Failed to log warn message: " .. tostring(err))
            end
        end,
        error = function(message, data)
            validate_required(message, "message")
            local log_data = {}
            for k, v in pairs(context) do
                log_data[k] = v
            end
            if data then
                for k, v in pairs(data) do
                    log_data[k] = v
                end
            end
            log_data.logger = name

            local error_success, err = pcall(bridge.error, message, log_data)
            if not error_success then
                error("Failed to log error message: " .. tostring(err))
            end
        end,
        with_context = function(new_context)
            validate_required(new_context, "context")
            local merged_context = {}
            for k, v in pairs(context) do
                merged_context[k] = v
            end
            for k, v in pairs(new_context) do
                merged_context[k] = v
            end

            return observability.logger(name, {
                level = level,
                context = merged_context,
            })
        end,
    }
end

-- Simple logging functions (using default logger)
function observability.log(level, message, data)
    validate_required(level, "log level")
    validate_required(message, "message")

    local bridge = get_slog_bridge()
    local method = bridge[level]

    if type(method) ~= "function" then
        error("Invalid log level: " .. tostring(level))
    end

    local log_success, err = pcall(method, message, data or {})
    if not log_success then
        error("Failed to log message: " .. tostring(err))
    end
end

-- Convenience logging functions
function observability.debug(message, data)
    observability.log("debug", message, data)
end

function observability.info(message, data)
    observability.log("info", message, data)
end

function observability.warn(message, data)
    observability.log("warn", message, data)
end

function observability.error(message, data)
    observability.log("error", message, data)
end

-- Health and Monitoring

-- Create a health check
function observability.health_check(name, check_func, options)
    validate_required(name, "health check name")
    validate_required(check_func, "check function")

    if type(check_func) ~= "function" then
        error("check_func must be a function")
    end

    options = options or {}
    -- local timeout = options.timeout or 5000 -- 5 seconds default (reserved for future use)
    local tags = options.tags or {}

    -- Create metrics for health check
    local success_counter =
        observability.counter(name .. "_success", "Health check successes", tags)
    local failure_counter = observability.counter(name .. "_failure", "Health check failures", tags)
    local timer_metric =
        observability.timer(name .. "_duration", "Health check execution time", tags)

    return {
        name = name,
        check = function()
            local span = observability.start_span("health_check_" .. name, {
                operation = "health_check",
                check_name = name,
            })

            local start_time = os.clock()
            local success, result = pcall(function()
                -- Simple timeout simulation (would need actual timeout implementation)
                return check_func()
            end)
            local duration = (os.clock() - start_time) * 1000

            timer_metric.record(duration)

            if success then
                success_counter.increment(1)
                span.set_status("ok")
                span.add_attribute("result", "healthy")
                span.finish()

                return {
                    healthy = true,
                    status = "ok",
                    duration_ms = duration,
                    result = result,
                }
            else
                failure_counter.increment(1)
                span.set_status("error")
                span.record_error(tostring(result))
                span.add_attribute("result", "unhealthy")
                span.finish()

                return {
                    healthy = false,
                    status = "error",
                    duration_ms = duration,
                    error = tostring(result),
                }
            end
        end,
        get_metrics = function()
            return {
                successes = success_counter.get(),
                failures = failure_counter.get(),
                stats = timer_metric.get_stats(),
            }
        end,
    }
end

-- Event Monitoring

-- Monitor specific events
function observability.monitor_events(pattern, handler, options)
    validate_required(pattern, "event pattern")
    validate_required(handler, "event handler")

    if type(handler) ~= "function" then
        error("handler must be a function")
    end

    options = options or {}
    local bridge = get_events_bridge()

    -- Create metrics for event monitoring
    local event_counter =
        observability.counter("events_" .. pattern, "Monitored events: " .. pattern)

    local wrapped_handler = function(event_data)
        event_counter.increment(1)

        -- Add logging
        observability.debug("Event monitored", {
            pattern = pattern,
            event = event_data,
        })

        -- Call original handler
        local success, result = pcall(handler, event_data)
        if not success then
            observability.error("Event handler failed", {
                pattern = pattern,
                error = tostring(result),
                event = event_data,
            })
        end

        return result
    end

    local sub_success, subscription = pcall(bridge.subscribe, pattern, wrapped_handler, options)
    if not sub_success then
        error("Failed to monitor events: " .. tostring(subscription))
    end

    return {
        pattern = pattern,
        unsubscribe = function()
            local unsub_success, err = pcall(bridge.unsubscribe, subscription)
            if not unsub_success then
                error("Failed to unsubscribe from events: " .. tostring(err))
            end
        end,
        get_count = function()
            return event_counter.get()
        end,
    }
end

-- Safety and Compliance Monitoring

-- Create a guardrail for monitoring safety violations
function observability.guardrail(name, validation_func, options)
    validate_required(name, "guardrail name")
    validate_required(validation_func, "validation function")

    if type(validation_func) ~= "function" then
        error("validation_func must be a function")
    end

    options = options or {}
    local bridge = get_guardrails_bridge()

    if not bridge then
        observability.warn("Guardrails bridge not available, creating local-only guardrail")

        -- Create local-only guardrail with metrics
        local violation_counter =
            observability.counter(name .. "_violations", "Guardrail violations: " .. name)
        local validation_counter =
            observability.counter(name .. "_validations", "Guardrail validations: " .. name)

        return {
            name = name,
            validate = function(data)
                validation_counter.increment(1)

                local success, result = pcall(validation_func, data)
                if not success then
                    violation_counter.increment(1)
                    observability.error("Guardrail validation error", {
                        guardrail = name,
                        error = tostring(result),
                        data = data,
                    })
                    return false, result
                end

                if not result then
                    violation_counter.increment(1)
                    observability.warn("Guardrail violation", {
                        guardrail = name,
                        data = data,
                    })
                end

                return result, nil
            end,
            get_metrics = function()
                return {
                    violations = violation_counter.get(),
                    validations = validation_counter.get(),
                }
            end,
        }
    end

    -- Use bridge-based guardrail
    local create_success, guardrail_id =
        pcall(bridge.createGuardrail, name, validation_func, options)
    if not create_success then
        error("Failed to create guardrail: " .. tostring(guardrail_id))
    end

    return {
        name = name,
        id = guardrail_id,
        validate = function(data)
            local validate_success, result = pcall(bridge.validateGuardrail, guardrail_id, data)
            if not validate_success then
                error("Failed to validate guardrail: " .. tostring(result))
            end
            return result
        end,
        get_metrics = function()
            local metrics_success, metrics = pcall(bridge.getGuardrailMetrics, guardrail_id)
            if not metrics_success then
                error("Failed to get guardrail metrics: " .. tostring(metrics))
            end
            return metrics
        end,
    }
end

-- Utility Functions

-- Get all metrics summary
function observability.get_metrics_summary()
    local bridge = get_metrics_bridge()
    local summary_success, summary = pcall(bridge.getAllMetrics)

    if not summary_success then
        error("Failed to get metrics summary: " .. tostring(summary))
    end

    return summary
end

-- Get system information
function observability.get_system_info()
    return {
        lua_version = _VERSION,
        os_time = os.time(),
        bridges_available = {
            metrics = _G.metrics ~= nil,
            tracing = _G.tracing ~= nil,
            slog = _G.slog ~= nil,
            events = _G.events ~= nil,
            guardrails = _G.guardrails ~= nil,
        },
        active_spans = #active_spans,
        custom_loggers = #custom_loggers,
    }
end

-- Clean up resources
function observability.cleanup()
    -- End all active spans
    for span_id, _ in pairs(active_spans) do
        local tracing = get_tracing_bridge()
        pcall(tracing.endSpan, span_id)
        active_spans[span_id] = nil
    end

    -- Clear performance timers (reserved for future use)
    -- performance_timers = {}

    -- Clear custom loggers
    custom_loggers = {}

    observability.info("Observability cleanup completed")
end

-- Export the module
return observability
