# Lua Logging & Debug Library Design

## Overview
A comprehensive logging and debug library that bridges the existing go-llmspell logging infrastructure:
- util/debug.go - Debug logging with component control
- util/slog.go - Structured logging with slog
- util/script_logger.go - Unified script-friendly logger

## Core API Design

### 1. Logger Creation and Configuration
```lua
-- Create a logger instance
local logger = logging.create(name, config)

-- Get default logger
local log = logging.default()

-- Configure logging
logging.configure({
    level = "info",          -- debug, info, warn, error
    format = "json",         -- json, text, pretty
    output = "stdout",       -- stdout, stderr, file
    file = "/path/to/log",   -- if output="file"
    emoji = true,            -- enable emoji in logs
    components = {"auth", "llm"}  -- debug components
})
```

### 2. Basic Logging Methods
```lua
-- Standard log levels
logger:debug(message, attributes)
logger:info(message, attributes)
logger:warn(message, attributes)
logger:error(message, attributes)

-- Conditional debug logging
logger:debug_if(condition, message, attributes)

-- Component-based debug
logger:debug_component(component, message, ...)
```

### 3. Structured Logging
```lua
-- Log with structured attributes
logger:log("info", "User action", {
    user_id = "123",
    action = "login",
    ip = "192.168.1.1",
    timestamp = os.time()
})

-- Add persistent context
logger:with_context({
    request_id = "req_123",
    service = "auth"
})

-- Child logger with additional context
local child = logger:child({
    module = "database"
})
```

### 4. Debug Control
```lua
-- Enable/disable debug logging
logging.debug.enable()
logging.debug.disable()

-- Component-based debug control
logging.debug.enable_component("auth")
logging.debug.disable_component("auth")
logging.debug.is_enabled("auth")

-- Get active components
local components = logging.debug.get_components()
```

### 5. Log Formatting and Output
```lua
-- Custom formatters
logging.formatters.register("custom", function(entry)
    return string.format("[%s] %s: %s",
        entry.level,
        entry.timestamp,
        entry.message
    )
end)

-- Set formatter
logger:set_formatter("custom")

-- Log rotation (if file output)
logger:rotate({
    max_size = "100MB",
    max_age = "7d",
    max_backups = 5
})
```

### 6. Performance and Profiling
```lua
-- Measure operation performance
local timer = logging.timer("operation_name")
-- ... do work ...
timer:stop()  -- Logs duration

-- Profile code block
logging.profile("heavy_computation", function()
    -- ... intensive work ...
end)

-- Manual timing
local start = logging.time()
-- ... work ...
logging.duration("task_name", start)
```

### 7. Hook Integration
```lua
-- Hook into LLM operations
logging.hooks.before_generate(function(params)
    logger:info("LLM generation starting", {
        model = params.model,
        prompt_length = #params.prompt
    })
end)

logging.hooks.after_generate(function(params, result)
    logger:info("LLM generation complete", {
        model = params.model,
        tokens = result.usage.total_tokens,
        duration = result.duration
    })
end)

-- Tool call hooks
logging.hooks.before_tool_call(function(tool, args)
    logger:debug("Tool call", {tool = tool, args = args})
end)

logging.hooks.after_tool_call(function(tool, args, result)
    logger:debug("Tool result", {
        tool = tool,
        success = result.success
    })
end)
```

### 8. Error and Exception Logging
```lua
-- Log errors with stack traces
logger:error_with_stack(error_obj, message)

-- Catch and log exceptions
logging.catch(function()
    -- risky operation
end, "operation_name")

-- Assert with logging
logging.assert(condition, message, attributes)
```

### 9. Metrics and Monitoring
```lua
-- Count events
logging.metrics.count("user_login", 1, {
    status = "success"
})

-- Gauge values
logging.metrics.gauge("queue_size", 42)

-- Histogram for distributions
logging.metrics.histogram("response_time", 123.45)
```

### 10. Audit Logging
```lua
-- Security-relevant logging
logging.audit.log("access_granted", {
    user = "john@example.com",
    resource = "/api/users",
    action = "READ",
    ip = "192.168.1.1"
})

-- Compliance logging
logging.audit.compliance("data_export", {
    user = "admin",
    records = 1000,
    destination = "s3://backup"
})
```

### 11. Log Querying and Analysis
```lua
-- Search recent logs
local entries = logging.search({
    level = {"error", "warn"},
    since = os.time() - 3600,  -- last hour
    component = "auth",
    limit = 100
})

-- Get log statistics
local stats = logging.stats({
    period = "1h",
    group_by = "level"
})
```

### 12. Integration Helpers
```lua
-- Create logger from bridge
local bridge_logger = logging.from_bridge(_G.util_script_logger)

-- Export logs
logging.export({
    format = "json",
    file = "/tmp/logs.json",
    filter = {level = "error"}
})

-- Real-time log streaming
logging.stream(function(entry)
    -- Process each log entry
    print(entry.message)
end)
```

## Implementation Notes

1. **Bridge Integration**: Wrap existing bridges (debug, slog, script_logger)
2. **Performance**: Lazy evaluation of expensive operations
3. **Thread Safety**: Use appropriate synchronization from bridges
4. **Memory Management**: Implement log buffer limits
5. **Error Handling**: Graceful degradation if bridges unavailable

## Example Usage

```lua
local logging = require("logging")

-- Configure logging
logging.configure({
    level = "info",
    format = "json",
    emoji = true
})

-- Get logger
local log = logging.create("myapp")

-- Add context
log:with_context({
    version = "1.0.0",
    environment = "production"
})

-- Log with structure
log:info("Application started", {
    pid = process.pid,
    memory = process.memory_usage()
})

-- Debug logging
logging.debug.enable_component("database")
log:debug_component("database", "Query executed", {
    query = "SELECT * FROM users",
    duration = 45.2
})

-- Error handling
local success, err = pcall(risky_operation)
if not success then
    log:error_with_stack(err, "Operation failed")
end

-- Performance monitoring
logging.profile("data_processing", function()
    process_large_dataset()
end)

-- Audit logging
logging.audit.log("permission_changed", {
    admin = current_user,
    target_user = user_id,
    old_role = "viewer",
    new_role = "editor"
})
```