-- ABOUTME: Tools & Registry Library for go-llmspell Lua standard library
-- ABOUTME: Provides tool registration, execution, validation, composition, and registry management

local tools = {}

-- Import promise library for async operations (reserved for future use)
-- local promise = _G.promise or require("promise")

-- Internal tracking
local registered_tools = {}
local tool_metrics = {}
local execution_history = {}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to get tools bridge
local function get_tools_bridge()
    if not _G.tools then
        error("Tools bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.tools
end

-- Helper function to get registry bridge if available (reserved for future use)
-- local function get_registry_bridge()
--     -- Registry bridge is optional, return nil if not available
--     return _G.registry
-- end

-- Helper function to generate unique execution ID
local function generate_execution_id()
    return "exec_" .. tostring(os.time()) .. "_" .. tostring(math.random(1000, 9999))
end

-- Helper function to record execution metrics
local function record_execution(tool_name, duration, success, error_msg)
    if not tool_metrics[tool_name] then
        tool_metrics[tool_name] = {
            total_calls = 0,
            successful_calls = 0,
            failed_calls = 0,
            total_duration = 0,
            avg_duration = 0,
            last_error = nil,
            last_execution = nil,
        }
    end

    local metrics = tool_metrics[tool_name]
    metrics.total_calls = metrics.total_calls + 1
    metrics.total_duration = metrics.total_duration + duration
    metrics.avg_duration = metrics.total_duration / metrics.total_calls
    metrics.last_execution = os.time()

    if success then
        metrics.successful_calls = metrics.successful_calls + 1
    else
        metrics.failed_calls = metrics.failed_calls + 1
        metrics.last_error = error_msg
    end
end

-- Tool Registration and Management

-- Define a new tool with comprehensive metadata
function tools.define(name, description, schema, func)
    validate_required(name, "tool name")
    validate_required(description, "tool description")
    validate_required(func, "tool function")

    if type(name) ~= "string" then
        error("tool name must be a string")
    end

    if type(description) ~= "string" then
        error("tool description must be a string")
    end

    if type(func) ~= "function" then
        error("tool function must be a function")
    end

    schema = schema or {}

    -- Create comprehensive tool definition
    local tool_def = {
        name = name,
        description = description,
        version = schema.version or "1.0.0",
        category = schema.category or "custom",
        tags = schema.tags or {},
        parameters = schema.parameters or {},
        output_schema = schema.output_schema or {},
        examples = schema.examples or {},
        constraints = schema.constraints or {},
        permissions = schema.permissions or {},
        resource_usage = schema.resource_usage or {},
        is_deterministic = schema.is_deterministic ~= false,
        is_destructive = schema.is_destructive or false,
        requires_confirmation = schema.requires_confirmation or false,
        execute = func,
        created_at = os.time(),
        usage_count = 0,
    }

    -- Store locally
    registered_tools[name] = tool_def

    -- Register with bridge if available
    local bridge = get_tools_bridge()
    if bridge and bridge.registerCustomTool then
        local success = bridge.registerCustomTool(tool_def)
        if not success then
            error("Failed to register tool with bridge: " .. name)
        end
    end

    return tool_def
end

-- Register a library of tools
function tools.register_library(library)
    validate_required(library, "library")

    if type(library) ~= "table" then
        error("library must be a table")
    end

    local registered_count = 0
    local failed_tools = {}

    for tool_name, tool_def in pairs(library) do
        local success, err = pcall(function()
            if type(tool_def) == "function" then
                -- Simple function - create basic definition
                tools.define(tool_name, "Library tool: " .. tool_name, {}, tool_def)
            elseif type(tool_def) == "table" then
                -- Full tool definition
                tools.define(
                    tool_name,
                    tool_def.description or ("Library tool: " .. tool_name),
                    tool_def.schema or tool_def,
                    tool_def.execute or tool_def.func
                )
            else
                error("Invalid tool definition for " .. tool_name)
            end
        end)

        if success then
            registered_count = registered_count + 1
        else
            table.insert(failed_tools, { name = tool_name, error = err })
        end
    end

    return {
        registered = registered_count,
        failed = #failed_tools,
        failures = failed_tools,
        total = registered_count + #failed_tools,
    }
end

-- Compose multiple tools into a new composite tool
function tools.compose(tool_list, options)
    validate_required(tool_list, "tools")

    if type(tool_list) ~= "table" or #tool_list == 0 then
        error("tools must be a non-empty table")
    end

    options = options or {}
    local composition_name = options.name or "composed_tool_" .. tostring(os.time())
    local mode = options.mode or "pipeline" -- pipeline, parallel, conditional

    local composite_func
    if mode == "pipeline" then
        composite_func = function(input)
            local result = input
            for i, tool in ipairs(tool_list) do
                if type(tool) == "string" then
                    result = tools.execute_safe(tool, result)
                elseif type(tool) == "table" and tool.execute then
                    result = tool.execute(result)
                elseif type(tool) == "function" then
                    result = tool(result)
                else
                    error("Invalid tool at position " .. i)
                end
            end
            return result
        end
    elseif mode == "parallel" then
        composite_func = function(input)
            return tools.parallel_execute(tool_list, input)
        end
    elseif mode == "conditional" then
        composite_func = function(input)
            for _, tool in ipairs(tool_list) do
                local condition = tool.condition
                if not condition or condition(input) then
                    if type(tool.execute) == "function" then
                        return tool.execute(input)
                    elseif type(tool.func) == "function" then
                        return tool.func(input)
                    end
                end
            end
            error("No tool condition matched")
        end
    else
        error("Invalid composition mode: " .. mode)
    end

    return tools.define(
        composition_name,
        "Composite tool: " .. mode .. " of " .. #tool_list .. " tools",
        {
            category = "composite",
            tags = { "composed", mode },
            composition = {
                mode = mode,
                tools = tool_list,
                options = options,
            },
        },
        composite_func
    )
end

-- Tool Execution Utilities

-- Execute a tool safely with error handling and metrics
function tools.execute_safe(tool, params, options)
    validate_required(tool, "tool")

    options = options or {}
    local start_time = os.clock()
    local execution_id = generate_execution_id()

    -- Resolve tool
    local tool_def
    local tool_name
    if type(tool) == "string" then
        tool_name = tool
        tool_def = registered_tools[tool_name]
        if not tool_def then
            -- Try to get from bridge
            local bridge = get_tools_bridge()
            if bridge and bridge.getToolInfo then
                local info = bridge:getToolInfo(tool_name)
                if info then
                    tool_def = info
                end
            end
        end
        if not tool_def then
            error("Tool not found: " .. tool_name)
        end
    elseif type(tool) == "table" then
        tool_def = tool
        tool_name = tool.name or "anonymous"
    else
        error("tool must be a string (name) or table (definition)")
    end

    -- Validate parameters if schema is available
    if tool_def.parameters and next(tool_def.parameters) then
        local validation_result = tools.validate_params(tool_def, params)
        if not validation_result.valid then
            error("Parameter validation failed: " .. table.concat(validation_result.errors, ", "))
        end
    end

    -- Execute with error handling
    local success, result, error_msg
    if options.timeout then
        -- Execute with timeout (simplified - would need promise integration)
        success, result = pcall(tool_def.execute, params)
        error_msg = not success and result or nil
    else
        success, result = pcall(tool_def.execute, params)
        error_msg = not success and result or nil
    end

    local duration = os.clock() - start_time

    -- Record metrics
    record_execution(tool_name, duration, success, error_msg)

    -- Record execution history
    table.insert(execution_history, {
        id = execution_id,
        tool = tool_name,
        timestamp = os.time(),
        duration = duration,
        success = success,
        error = error_msg,
        params = params,
        result = success and result or nil,
    })

    -- Limit history size
    if #execution_history > 1000 then
        table.remove(execution_history, 1)
    end

    if not success then
        if options.silent then
            return nil, error_msg
        else
            error("Tool execution failed: " .. tostring(error_msg))
        end
    end

    return result
end

-- Execute tools in a pipeline
function tools.pipeline(tool_list, initial_data, options)
    validate_required(tool_list, "tools")

    if type(tool_list) ~= "table" or #tool_list == 0 then
        error("tools must be a non-empty table")
    end

    options = options or {}
    local results = {}
    local current_data = initial_data

    for i, tool in ipairs(tool_list) do
        local step_options = {
            silent = options.continue_on_error,
            timeout = options.step_timeout,
        }

        local success, result, error_msg =
            pcall(tools.execute_safe, tool, current_data, step_options)

        if success then
            table.insert(results, result)
            if options.pass_output ~= false then
                current_data = result
            end
        else
            table.insert(results, { error = error_msg })
            if not options.continue_on_error then
                error("Pipeline failed at step " .. i .. ": " .. tostring(error_msg))
            end
        end
    end

    return {
        results = results,
        final_output = current_data,
        success = #results == #tool_list,
        completed_steps = #results,
        total_steps = #tool_list,
    }
end

-- Execute tools in parallel
function tools.parallel_execute(tool_list, params, options)
    validate_required(tool_list, "tools")

    if type(tool_list) ~= "table" or #tool_list == 0 then
        error("tools must be a non-empty table")
    end

    options = options or {}
    local max_concurrent = options.max_concurrent or #tool_list

    -- For simplicity, execute sequentially but collect all results
    -- In a full implementation, this would use actual concurrency
    local results = {}
    local errors = {}

    for i, tool in ipairs(tool_list) do
        if i <= max_concurrent then
            local success, result = pcall(tools.execute_safe, tool, params, { silent = true })
            if success then
                results[i] = result
            else
                errors[i] = result
            end
        end
    end

    return {
        results = results,
        errors = errors,
        success_count = #results,
        error_count = #errors,
        total = #tool_list,
    }
end

-- Tool Validation and Testing

-- Validate tool parameters against schema
function tools.validate_params(tool, params)
    validate_required(tool, "tool")

    local tool_def
    if type(tool) == "string" then
        tool_def = registered_tools[tool]
        if not tool_def then
            error("Tool not found: " .. tool)
        end
    elseif type(tool) == "table" then
        tool_def = tool
    else
        error("tool must be a string or table")
    end

    local schema = tool_def.parameters
    if not schema or not next(schema) then
        return { valid = true, errors = {} }
    end

    local errors = {}

    -- Check required parameters
    if schema.required and type(schema.required) == "table" then
        for _, field in ipairs(schema.required) do
            if not params or params[field] == nil then
                table.insert(errors, "Missing required parameter: " .. field)
            end
        end
    end

    -- Check parameter types
    if schema.properties and type(schema.properties) == "table" and params then
        for param_name, param_value in pairs(params) do
            local param_schema = schema.properties[param_name]
            if param_schema and param_schema.type then
                local expected_type = param_schema.type
                local actual_type = type(param_value)

                -- Basic type checking
                if expected_type == "integer" and actual_type == "number" then
                    if param_value % 1 ~= 0 then
                        table.insert(errors, "Parameter '" .. param_name .. "' must be an integer")
                    end
                elseif expected_type ~= actual_type then
                    table.insert(
                        errors,
                        string.format(
                            "Parameter '%s' has wrong type. Expected: %s, Got: %s",
                            param_name,
                            expected_type,
                            actual_type
                        )
                    )
                end
            end
        end
    end

    return {
        valid = #errors == 0,
        errors = errors,
    }
end

-- Test a tool with provided test cases
function tools.test_tool(
    tool,
    test_cases --[[, options]]
)
    validate_required(tool, "tool")
    validate_required(test_cases, "test_cases")

    if type(test_cases) ~= "table" then
        error("test_cases must be a table")
    end

    -- options = options or {}  -- Unused in current implementation
    local results = {}
    local passed = 0
    local failed = 0

    for i, test_case in ipairs(test_cases) do
        local test_result = {
            name = test_case.name or ("Test " .. i),
            input = test_case.input,
            expected = test_case.expected,
            passed = false,
            error = nil,
            actual = nil,
            duration = 0,
        }

        local start_time = os.clock()
        local success, result = pcall(tools.execute_safe, tool, test_case.input, { silent = true })
        test_result.duration = os.clock() - start_time

        if success then
            test_result.actual = result
            if test_case.expected then
                -- Simple equality check
                test_result.passed = test_result.actual == test_case.expected
            elseif test_case.validator and type(test_case.validator) == "function" then
                test_result.passed = test_case.validator(result)
            else
                test_result.passed = true -- No validation specified
            end
        else
            test_result.error = result
            test_result.passed = false
        end

        if test_result.passed then
            passed = passed + 1
        else
            failed = failed + 1
        end

        table.insert(results, test_result)
    end

    return {
        results = results,
        summary = {
            total = #test_cases,
            passed = passed,
            failed = failed,
            pass_rate = passed / #test_cases,
        },
    }
end

-- Benchmark a tool's performance
function tools.benchmark_tool(tool, params, options)
    validate_required(tool, "tool")

    options = options or {}
    local iterations = options.iterations or 10
    local warmup = options.warmup or 2

    local durations = {}
    local errors = 0

    -- Warmup runs
    for _ = 1, warmup do
        pcall(tools.execute_safe, tool, params, { silent = true })
    end

    -- Benchmark runs
    for _ = 1, iterations do
        local start_time = os.clock()
        local success = pcall(tools.execute_safe, tool, params, { silent = true })
        local duration = os.clock() - start_time

        if success then
            table.insert(durations, duration)
        else
            errors = errors + 1
        end
    end

    if #durations == 0 then
        error("No successful executions during benchmark")
    end

    -- Calculate statistics
    local total_time = 0
    local min_time = durations[1]
    local max_time = durations[1]

    for _, duration in ipairs(durations) do
        total_time = total_time + duration
        if duration < min_time then
            min_time = duration
        end
        if duration > max_time then
            max_time = duration
        end
    end

    local avg_time = total_time / #durations

    -- Calculate median
    table.sort(durations)
    local median_time
    if #durations % 2 == 0 then
        median_time = (durations[#durations / 2] + durations[#durations / 2 + 1]) / 2
    else
        median_time = durations[math.ceil(#durations / 2)]
    end

    return {
        iterations = iterations,
        successful = #durations,
        errors = errors,
        total_time = total_time,
        avg_time = avg_time,
        min_time = min_time,
        max_time = max_time,
        median_time = median_time,
        throughput = #durations / total_time, -- executions per second
    }
end

-- Discovery and Information

-- List all available tools
function tools.list()
    local bridge = get_tools_bridge()
    local all_tools = {}

    -- Add locally registered tools
    for name, tool_def in pairs(registered_tools) do
        table.insert(all_tools, {
            name = name,
            description = tool_def.description,
            category = tool_def.category,
            version = tool_def.version,
            source = "local",
        })
    end

    -- Add bridge tools if available
    if bridge and bridge.listTools then
        local bridge_tools = bridge.listTools()
        if bridge_tools then
            for _, tool in ipairs(bridge_tools) do
                table.insert(all_tools, {
                    name = tool.name,
                    description = tool.description,
                    category = tool.category,
                    version = tool.version,
                    source = "bridge",
                })
            end
        end
    end

    return all_tools
end

-- Search tools by query
function tools.search(query)
    validate_required(query, "query")

    local bridge = get_tools_bridge()
    if bridge and bridge.searchTools then
        return bridge.searchTools(query)
    else
        -- Fallback: simple local search
        local results = {}
        for name, tool_def in pairs(registered_tools) do
            if
                string.find(name:lower(), query:lower())
                or string.find(tool_def.description:lower(), query:lower())
            then
                table.insert(results, tool_def)
            end
        end
        return results
    end
end

-- Get detailed tool information
function tools.get_info(name)
    validate_required(name, "tool name")

    -- Check local registry first
    if registered_tools[name] then
        return registered_tools[name]
    end

    -- Try bridge
    local bridge = get_tools_bridge()
    if bridge and bridge.getToolInfo then
        return bridge.getToolInfo(name)
    end

    return nil
end

-- Get tool execution metrics
function tools.get_metrics(name)
    if name then
        return tool_metrics[name]
    else
        return tool_metrics
    end
end

-- Get execution history
function tools.get_history(limit)
    limit = limit or 50
    local history = {}
    local start_index = math.max(1, #execution_history - limit + 1)

    for i = start_index, #execution_history do
        table.insert(history, execution_history[i])
    end

    return history
end

-- Utility Functions

-- Clear execution history
function tools.clear_history()
    execution_history = {}
end

-- Reset tool metrics
function tools.reset_metrics(tool_name)
    if tool_name then
        tool_metrics[tool_name] = nil
    else
        tool_metrics = {}
    end
end

-- Export tool definition
function tools.export_tool(name, format)
    local tool_def = tools.get_info(name)
    if not tool_def then
        error("Tool not found: " .. name)
    end

    format = format or "lua"

    if format == "lua" then
        return tool_def
    elseif format == "json" then
        -- Would use data.to_json if available
        return tostring(tool_def)
    else
        error("Unsupported export format: " .. format)
    end
end

-- If tools global already exists (from bridge), extend it instead of replacing
if _G.tools and type(_G.tools) == "table" then
    -- Merge our functions into the existing bridge
    for k, v in pairs(tools) do
        _G.tools[k] = v
    end
    return _G.tools
else
    -- No existing bridge, return our module
    return tools
end
