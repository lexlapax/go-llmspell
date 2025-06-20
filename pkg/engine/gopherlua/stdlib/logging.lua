-- ABOUTME: Logging & Debug Library for go-llmspell Lua standard library
-- ABOUTME: Provides comprehensive logging, debugging, profiling, and monitoring capabilities

local logging = {}

-- Note: unpack is defined but not used in this file currently

-- Internal state
local loggers = {}
local default_logger = nil
local global_config = {
    level = "info",
    format = "text",
    output = "stdout",
    emoji = false,
    components = {},
}
local formatters = {}
local hooks = {
    before_generate = {},
    after_generate = {},
    before_tool_call = {},
    after_tool_call = {},
}
local audit_handlers = {}

-- Log levels
local LOG_LEVELS = {
    debug = 0,
    info = 1,
    warn = 2,
    error = 3,
}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to get appropriate bridge
local function get_bridge(bridge_type)
    local bridges = {
        debug = _G.util_debug,
        slog = _G.util_slog,
        script_logger = _G.util_script_logger,
    }

    local bridge = bridges[bridge_type]
    if not bridge then
        error(bridge_type .. " bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return bridge
end

-- Logger class
local Logger = {}
Logger.__index = Logger

function Logger.new(_, name, config)
    local self = setmetatable({}, Logger)
    self.name = name or "default"
    self.config = config or {}
    self.context = {}
    self.level = LOG_LEVELS[self.config.level or global_config.level] or LOG_LEVELS.info
    self.formatter = self.config.format or global_config.format

    -- Get script logger bridge for unified logging
    self.bridge = get_bridge("script_logger")

    return self
end

-- Add persistent context to logger
function Logger:with_context(context)
    validate_required(context, "context")

    if type(context) ~= "table" then
        error("context must be a table")
    end

    for k, v in pairs(context) do
        self.context[k] = v
    end

    return self
end

-- Create child logger with additional context
function Logger:child(context)
    validate_required(context, "context")

    if type(context) ~= "table" then
        error("context must be a table")
    end

    local child = Logger.new(nil, self.name, self.config)

    -- Copy parent context
    for k, v in pairs(self.context) do
        child.context[k] = v
    end

    -- Add child context
    for k, v in pairs(context) do
        child.context[k] = v
    end

    return child
end

-- Core logging method
function Logger:log(level, message, attributes)
    validate_required(level, "log level")
    validate_required(message, "message")

    if type(level) == "string" then
        level = LOG_LEVELS[level]
        if not level then
            error("invalid log level")
        end
    end

    -- Check if we should log this level
    if level < self.level then
        return
    end

    -- Merge attributes with context
    local final_attrs = {}
    for k, v in pairs(self.context) do
        final_attrs[k] = v
    end
    if attributes then
        for k, v in pairs(attributes) do
            final_attrs[k] = v
        end
    end

    -- Add metadata
    final_attrs.logger = self.name
    final_attrs.timestamp = os.time()

    -- Use script logger bridge
    local success, _ = pcall(self.bridge.log, message, level, final_attrs)
    if not success then
        -- Fallback to print if bridge fails
        print(string.format("[%s] %s: %s", level, self.name, message))
    end
end

-- Convenience logging methods
function Logger:debug(message, attributes)
    self:log("debug", message, attributes)
end

function Logger:info(message, attributes)
    self:log("info", message, attributes)
end

function Logger:warn(message, attributes)
    self:log("warn", message, attributes)
end

function Logger:error(message, attributes)
    self:log("error", message, attributes)
end

-- Conditional debug logging
function Logger:debug_if(condition, message, attributes)
    if condition then
        self:debug(message, attributes)
    end
end

-- Component-based debug logging
function Logger.debug_component(_, component, message, ...)
    validate_required(component, "component")
    validate_required(message, "message")

    local debug_bridge = get_bridge("debug")
    pcall(debug_bridge.debugPrintf, component, message, ...)
end

-- Error logging with stack trace
function Logger:error_with_stack(error_obj, message)
    local stack = debug.traceback(error_obj, 2)
    self:error(message or "Error occurred", {
        error = tostring(error_obj),
        stack_trace = stack,
    })
end

-- Set formatter for this logger
function Logger:set_formatter(formatter_name)
    validate_required(formatter_name, "formatter name")

    if not formatters[formatter_name] then
        error("unknown formatter: " .. formatter_name)
    end

    self.formatter = formatter_name
    return self
end

-- Main logging module functions

-- Create a new logger
function logging.create(name, config)
    name = name or "default"

    if loggers[name] then
        return loggers[name]
    end

    local logger = Logger.new(nil, name, config)
    loggers[name] = logger

    return logger
end

-- Get default logger
function logging.default()
    if not default_logger then
        default_logger = logging.create("default")
    end
    return default_logger
end

-- Configure global logging
function logging.configure(config)
    validate_required(config, "configuration")

    if type(config) ~= "table" then
        error("configuration must be a table")
    end

    -- Update global config
    for k, v in pairs(config) do
        global_config[k] = v
    end

    -- Configure script logger bridge if available
    local bridge = get_bridge("script_logger")
    local success, _ = pcall(bridge.configure, {
        level = config.level,
        format = config.format,
        emoji = config.emoji,
    })
    _ = success

    -- Configure debug components if specified
    if config.components then
        local debug_bridge = get_bridge("debug")
        for _, component in ipairs(config.components) do
            local enable_success, _ = pcall(debug_bridge.enableDebugComponent, component)
            _ = enable_success
        end
    end
end

-- Debug control namespace
logging.debug = {}

-- Enable debug logging globally
function logging.debug.enable()
    local debug_bridge = get_bridge("debug")
    -- Enable by setting empty component (enables all)
    local success, _ = pcall(debug_bridge.enableDebugComponent, "")
    _ = success
end

-- Disable debug logging globally
function logging.debug.disable()
    -- No direct API in bridge to disable all, so we track components
    local debug_bridge = get_bridge("debug")
    for component, _ in
        pairs(debug_bridge.getDebugComponents and debug_bridge.getDebugComponents() or {})
    do
        local success, _ = pcall(debug_bridge.disableDebugComponent, component)
        _ = success
    end
end

-- Enable debug for specific component
function logging.debug.enable_component(component)
    validate_required(component, "component name")

    local debug_bridge = get_bridge("debug")
    local success, _ = pcall(debug_bridge.enableDebugComponent, component)
    if not success then
        error("failed to enable debug component: " .. component)
    end
end

-- Disable debug for specific component
function logging.debug.disable_component(component)
    validate_required(component, "component name")

    local debug_bridge = get_bridge("debug")
    local success, _ = pcall(debug_bridge.disableDebugComponent, component)
    if not success then
        error("failed to disable debug component: " .. component)
    end
end

-- Check if debug is enabled for component
function logging.debug.is_enabled(component)
    validate_required(component, "component name")

    local debug_bridge = get_bridge("debug")
    local success, enabled = pcall(debug_bridge.isDebugEnabled, component)
    if not success then
        return false
    end
    return enabled
end

-- Get list of active debug components
function logging.debug.get_components()
    -- This would need to be exposed by the bridge
    -- For now, return empty table
    return {}
end

-- Formatter registration
logging.formatters = {}

function logging.formatters.register(name, formatter_func)
    validate_required(name, "formatter name")
    validate_required(formatter_func, "formatter function")

    if type(formatter_func) ~= "function" then
        error("formatter must be a function")
    end

    formatters[name] = formatter_func
end

-- Built-in formatters
logging.formatters.register("json", function(entry)
    -- Simple JSON formatter
    local json_parts = {
        string.format('"level":"%s"', entry.level),
        string.format('"timestamp":%d', entry.timestamp),
        string.format('"message":"%s"', entry.message:gsub('"', '\\"')),
    }

    if entry.attributes then
        for k, v in pairs(entry.attributes) do
            if type(v) == "string" then
                table.insert(json_parts, string.format('"%s":"%s"', k, v:gsub('"', '\\"')))
            elseif type(v) == "number" then
                table.insert(json_parts, string.format('"%s":%g', k, v))
            elseif type(v) == "boolean" then
                table.insert(json_parts, string.format('"%s":%s', k, tostring(v)))
            end
        end
    end

    return "{" .. table.concat(json_parts, ",") .. "}"
end)

logging.formatters.register("text", function(entry)
    local parts = {
        string.format("[%s]", entry.level:upper()),
        os.date("%Y-%m-%d %H:%M:%S", entry.timestamp),
        entry.message,
    }

    if entry.attributes and next(entry.attributes) then
        local attrs = {}
        for k, v in pairs(entry.attributes) do
            table.insert(attrs, string.format("%s=%s", k, tostring(v)))
        end
        table.insert(parts, "(" .. table.concat(attrs, " ") .. ")")
    end

    return table.concat(parts, " ")
end)

-- Performance and profiling functions

-- Create a timer
function logging.timer(name)
    validate_required(name, "timer name")

    local timer = {
        name = name,
        start_time = os.clock() * 1000, -- Convert to milliseconds
        logger = logging.default(),
    }

    function timer:stop()
        local duration = (os.clock() * 1000) - self.start_time
        self.logger:debug("Timer completed", {
            timer = self.name,
            duration_ms = duration,
        })
        return duration
    end

    return timer
end

-- Profile a function execution
function logging.profile(name, func)
    validate_required(name, "profile name")
    validate_required(func, "function to profile")

    if type(func) ~= "function" then
        error("second argument must be a function")
    end

    local timer = logging.timer(name)
    local success, result = pcall(func)
    local duration = timer:stop()

    local logger = logging.default()
    logger:info("Profile completed", {
        profile = name,
        duration_ms = duration,
        success = success,
    })

    if success then
        return result
    else
        error(result)
    end
end

-- Get current time for manual timing
function logging.time()
    return os.clock() * 1000
end

-- Log duration from manual timing
function logging.duration(name, start_time)
    validate_required(name, "operation name")
    validate_required(start_time, "start time")

    local duration = (os.clock() * 1000) - start_time
    local logger = logging.default()
    logger:debug("Duration measured", {
        operation = name,
        duration_ms = duration,
    })
    return duration
end

-- Hook management
logging.hooks = {}

-- Register before_generate hook
function logging.hooks.before_generate(func)
    validate_required(func, "hook function")

    if type(func) ~= "function" then
        error("hook must be a function")
    end

    table.insert(hooks.before_generate, func)

    -- Register with slog bridge if available
    local slog_bridge = get_bridge("slog")
    local success, _ = pcall(slog_bridge.registerBeforeGenerateHook, func)
    _ = success
end

-- Register after_generate hook
function logging.hooks.after_generate(func)
    validate_required(func, "hook function")

    if type(func) ~= "function" then
        error("hook must be a function")
    end

    table.insert(hooks.after_generate, func)

    -- Register with slog bridge if available
    local slog_bridge = get_bridge("slog")
    local success, _ = pcall(slog_bridge.registerAfterGenerateHook, func)
    _ = success
end

-- Register before_tool_call hook
function logging.hooks.before_tool_call(func)
    validate_required(func, "hook function")

    if type(func) ~= "function" then
        error("hook must be a function")
    end

    table.insert(hooks.before_tool_call, func)

    -- Register with slog bridge if available
    local slog_bridge = get_bridge("slog")
    local success, _ = pcall(slog_bridge.registerBeforeToolCallHook, func)
    _ = success
end

-- Register after_tool_call hook
function logging.hooks.after_tool_call(func)
    validate_required(func, "hook function")

    if type(func) ~= "function" then
        error("hook must be a function")
    end

    table.insert(hooks.after_tool_call, func)

    -- Register with slog bridge if available
    local slog_bridge = get_bridge("slog")
    local success, _ = pcall(slog_bridge.registerAfterToolCallHook, func)
    _ = success
end

-- Error and exception handling

-- Catch and log exceptions
function logging.catch(func, operation_name)
    validate_required(func, "function")
    validate_required(operation_name, "operation name")

    if type(func) ~= "function" then
        error("first argument must be a function")
    end

    local logger = logging.default()
    local success, result = pcall(func)

    if not success then
        logger:error_with_stack(result, "Operation failed: " .. operation_name)
        return nil, result
    end

    return result, nil
end

-- Assert with logging
function logging.assert(condition, message, attributes)
    if not condition then
        local logger = logging.default()
        logger:error("Assertion failed", {
            message = message or "assertion failed",
            attributes = attributes,
        })
        error(message or "assertion failed")
    end
end

-- Metrics namespace
logging.metrics = {}

-- Count events
function logging.metrics.count(metric_name, value, tags)
    validate_required(metric_name, "metric name")
    value = value or 1

    local logger = logging.default()
    logger:debug("Metric count", {
        metric = metric_name,
        value = value,
        type = "counter",
        tags = tags,
    })
end

-- Record gauge value
function logging.metrics.gauge(metric_name, value, tags)
    validate_required(metric_name, "metric name")
    validate_required(value, "gauge value")

    if type(value) ~= "number" then
        error("gauge value must be a number")
    end

    local logger = logging.default()
    logger:debug("Metric gauge", {
        metric = metric_name,
        value = value,
        type = "gauge",
        tags = tags,
    })
end

-- Record histogram value
function logging.metrics.histogram(metric_name, value, tags)
    validate_required(metric_name, "metric name")
    validate_required(value, "histogram value")

    if type(value) ~= "number" then
        error("histogram value must be a number")
    end

    local logger = logging.default()
    logger:debug("Metric histogram", {
        metric = metric_name,
        value = value,
        type = "histogram",
        tags = tags,
    })
end

-- Audit logging namespace
logging.audit = {}

-- Log audit events
function logging.audit.log(event_type, details)
    validate_required(event_type, "event type")
    validate_required(details, "event details")

    if type(details) ~= "table" then
        error("event details must be a table")
    end

    local logger = logging.default()
    logger:info("AUDIT: " .. event_type, {
        audit_type = event_type,
        audit_details = details,
        audit_timestamp = os.time(),
        audit_id = os.time() .. "_" .. math.random(10000),
    })

    -- Call registered audit handlers
    for _, handler in ipairs(audit_handlers) do
        local success, _ = pcall(handler, event_type, details)
        _ = success
    end
end

-- Compliance-specific logging
function logging.audit.compliance(event_type, details)
    validate_required(event_type, "compliance event type")
    validate_required(details, "compliance details")

    details.compliance = true
    details.compliance_type = event_type

    logging.audit.log("compliance_" .. event_type, details)
end

-- Register audit handler
function logging.audit.register_handler(handler)
    validate_required(handler, "audit handler")

    if type(handler) ~= "function" then
        error("audit handler must be a function")
    end

    table.insert(audit_handlers, handler)
end

-- Log search and analysis
function logging.search(criteria)
    validate_required(criteria, "search criteria")

    if type(criteria) ~= "table" then
        error("search criteria must be a table")
    end

    -- This would require a log storage backend
    -- For now, return empty results
    return {}
end

-- Get log statistics
function logging.stats(options)
    local _ = options -- Store in local variable

    -- This would require a log storage backend
    -- For now, return basic stats
    return {
        total_logs = 0,
        by_level = {
            debug = 0,
            info = 0,
            warn = 0,
            error = 0,
        },
    }
end

-- Integration helpers

-- Create logger from bridge
function logging.from_bridge(bridge)
    validate_required(bridge, "bridge")

    local logger = logging.create("bridge_logger")
    logger.bridge = bridge
    return logger
end

-- Export logs
function logging.export(options)
    validate_required(options, "export options")

    if type(options) ~= "table" then
        error("export options must be a table")
    end

    -- This would require log storage
    -- For now, just log the export request
    local logger = logging.default()
    logger:info("Log export requested", options)
end

-- Real-time log streaming
function logging.stream(handler)
    validate_required(handler, "stream handler")

    if type(handler) ~= "function" then
        error("stream handler must be a function")
    end

    -- This would require a streaming backend
    -- For now, just register the handler
    local logger = logging.default()
    logger:debug("Log streaming registered")
end

-- System information
function logging.get_system_info()
    return {
        lua_version = _VERSION,
        bridges_available = {
            debug = _G.util_debug ~= nil,
            slog = _G.util_slog ~= nil,
            script_logger = _G.util_script_logger ~= nil,
        },
        loggers = loggers,
        global_config = global_config,
        hooks_registered = {
            before_generate = #hooks.before_generate,
            after_generate = #hooks.after_generate,
            before_tool_call = #hooks.before_tool_call,
            after_tool_call = #hooks.after_tool_call,
        },
    }
end

-- Cleanup resources
function logging.cleanup()
    -- Clear loggers
    for name, _ in pairs(loggers) do
        loggers[name] = nil
    end
    default_logger = nil

    -- Clear hooks
    hooks.before_generate = {}
    hooks.after_generate = {}
    hooks.before_tool_call = {}
    hooks.after_tool_call = {}

    -- Clear audit handlers
    audit_handlers = {}

    -- Log cleanup
    local logger = logging.default()
    logger:debug("Logging cleanup completed")
end

-- Export the module
return logging
