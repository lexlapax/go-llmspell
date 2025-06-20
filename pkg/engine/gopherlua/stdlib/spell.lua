-- ABOUTME: Spell Framework Library for go-llmspell Lua standard library
-- ABOUTME: Provides spell lifecycle management, composition, parameter handling, and execution context

-- luacheck: globals params (allow access to global params object)
-- luacheck: no unused args

local spell = {}

-- Internal state
local spell_config = {}
local spell_context = {}
local spell_params = {}
local spell_cache = {}
local spell_libraries = {}
local spell_hooks = {
    on_init = {},
    on_start = {},
    on_complete = {},
    on_cleanup = {},
    on_error = {},
}
local spell_resources = {}
local spell_environment = {}

-- Import other modules for functionality
local function safe_require(module)
    local ok, result = pcall(require, module)
    return ok and result or nil
end

local core = safe_require("core")

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to deep copy tables
local function deep_copy(obj)
    if core and core.table and core.table.deep_copy then
        return core.table.deep_copy(obj)
    end

    -- Fallback implementation
    if type(obj) ~= "table" then
        return obj
    end

    local copy = {}
    for k, v in pairs(obj) do
        copy[deep_copy(k)] = deep_copy(v)
    end
    return copy
end

-- Helper function to merge tables
local function merge_tables(...)
    if core and core.table and core.table.merge then
        return core.table.merge(...)
    end

    -- Fallback implementation
    local result = {}
    for _, tbl in ipairs({ ... }) do
        if type(tbl) == "table" then
            for k, v in pairs(tbl) do
                result[k] = v
            end
        end
    end
    return result
end

-- Helper function to generate UUID
local function generate_uuid()
    if core and core.crypto and core.crypto.uuid then
        return core.crypto.uuid()
    end

    -- Fallback simple ID
    return "spell_" .. os.time() .. "_" .. math.random(10000)
end

-- Helper function to validate parameter type
local function validate_param_type(value, expected_type, param_name)
    local actual_type = type(value)

    if expected_type == "array" then
        if actual_type ~= "table" then
            error(param_name .. " must be an array (table)")
        end
        -- Check if it's array-like
        local count = 0
        for _ in pairs(value) do
            count = count + 1
        end
        for i = 1, count do
            if value[i] == nil then
                error(param_name .. " must be an array (sequential table)")
            end
        end
    elseif expected_type == "object" then
        if actual_type ~= "table" then
            error(param_name .. " must be an object (table)")
        end
    elseif actual_type ~= expected_type then
        error(param_name .. " must be of type " .. expected_type .. ", got " .. actual_type)
    end
end

-- Helper function to validate enum values
local function validate_enum(value, enum_values, param_name)
    for _, valid_value in ipairs(enum_values) do
        if value == valid_value then
            return true
        end
    end

    local valid_str = table.concat(enum_values, ", ")
    error(param_name .. " must be one of: " .. valid_str .. ", got: " .. tostring(value))
end

-- Helper function to substitute template variables
local function substitute_variables(template, variables)
    if type(template) ~= "string" then
        return template
    end

    return string.gsub(template, "%$([%w_%.]+)", function(var_path)
        local value = variables
        for part in string.gmatch(var_path, "[^%.]+") do
            if type(value) == "table" and value[part] ~= nil then
                value = value[part]
            else
                return "$" .. var_path -- Keep original if not found
            end
        end
        return tostring(value)
    end)
end

-- ============================================================================
-- Spell Lifecycle Management
-- ============================================================================

-- Initialize spell with configuration and metadata
function spell.init(config)
    validate_required(config, "config")

    if type(config) ~= "table" then
        error("config must be a table")
    end

    -- Set default configuration
    spell_config = merge_tables({
        name = "unnamed-spell",
        version = "1.0.0",
        description = "",
        author = "",
        params = {},
        timeout = 300, -- 5 minutes default
        max_retries = 3,
        cache_ttl = 300, -- 5 minutes default
    }, config)

    -- Initialize execution context
    spell_context = {
        spell_name = spell_config.name,
        execution_id = generate_uuid(),
        start_time = os.time(),
        params = {},
        environment = spell_environment,
        metadata = {
            version = spell_config.version,
            description = spell_config.description,
            author = spell_config.author,
        },
    }

    -- Initialize parameters from config
    if spell_config.params then
        for param_name, param_config in pairs(spell_config.params) do
            spell_params[param_name] = param_config
        end
    end

    -- Execute init hooks
    for _, handler in ipairs(spell_hooks.on_init) do
        local success, err = pcall(handler, spell_config, spell_context)
        if not success then
            error("Init hook failed: " .. tostring(err))
        end
    end

    return spell_config
end

-- Define or get parameters with validation
function spell.params(name, config)
    validate_required(name, "parameter name")

    if config then
        -- Define parameter
        if type(config) ~= "table" then
            error("parameter config must be a table")
        end

        spell_params[name] = config
        return config
    else
        -- Get parameter value with validation
        local param_config = spell_params[name]
        local value = _G.params and _G.params[name]

        -- Apply default if value is nil
        if value == nil and param_config and param_config.default ~= nil then
            value = param_config.default
        end

        -- Check required
        if param_config and param_config.required and value == nil then
            error("Required parameter '" .. name .. "' is missing")
        end

        -- Validate type
        if value ~= nil and param_config and param_config.type then
            validate_param_type(value, param_config.type, name)
        end

        -- Validate enum
        if value ~= nil and param_config and param_config.enum then
            validate_enum(value, param_config.enum, name)
        end

        -- Store in context
        spell_context.params[name] = value

        return value
    end
end

-- Output results in structured format
function spell.output(data, format, metadata)
    validate_required(data, "data")

    format = format or "auto"
    metadata = metadata or {}

    local output = {
        spell = spell_config.name,
        execution_id = spell_context.execution_id,
        timestamp = os.time(),
        format = format,
        metadata = metadata,
        data = data,
    }

    -- Execute complete hooks
    for _, handler in ipairs(spell_hooks.on_complete) do
        local success, err = pcall(handler, output)
        if not success then
            error("Complete hook failed: " .. tostring(err))
        end
    end

    -- Format output based on type
    if format == "json" then
        -- Simple JSON-like formatting (requires proper JSON library for full implementation)
        print("=== SPELL OUTPUT (JSON) ===")
        print("{")
        print('  "spell": "' .. output.spell .. '",')
        print('  "execution_id": "' .. output.execution_id .. '",')
        print('  "timestamp": ' .. output.timestamp .. ",")
        print('  "data": ' .. tostring(data))
        print("}")
    elseif format == "text" or format == "auto" then
        print("=== SPELL OUTPUT ===")
        print("Spell: " .. output.spell)
        print("Execution ID: " .. output.execution_id)
        print("Timestamp: " .. os.date("%Y-%m-%d %H:%M:%S", output.timestamp))
        if next(metadata) then
            print("Metadata: " .. tostring(metadata))
        end
        print("")
        print(tostring(data))
    else
        -- Custom format - just print data
        print(tostring(data))
    end

    return output
end

-- ============================================================================
-- Spell Composition and Reuse
-- ============================================================================

-- Include other spells or libraries
function spell.include(path_or_name)
    validate_required(path_or_name, "path or library name")

    -- Check if it's a registered library first
    if spell_libraries[path_or_name] then
        return spell_libraries[path_or_name]
    end

    -- Try to load as file
    local success, result = pcall(dofile, path_or_name)
    if success then
        return result
    end

    -- Try to load as module
    success, result = pcall(require, path_or_name)
    if success then
        return result
    end

    error("Failed to include '" .. path_or_name .. "': not found as library, file, or module")
end

-- Compose multiple spells into a workflow
function spell.compose(spells_config)
    validate_required(spells_config, "spells configuration")

    if type(spells_config) ~= "table" then
        error("spells configuration must be a table")
    end

    local results = {}
    local variables = deep_copy(spell_context.params)

    for i, spell_step in ipairs(spells_config) do
        if type(spell_step) ~= "table" then
            error("spell step " .. i .. " must be a table")
        end

        local step_name = spell_step.name or ("step_" .. i)
        local spell_name = spell_step.spell
        local step_params = spell_step.params or {}

        if not spell_name then
            error("spell step " .. i .. " missing 'spell' field")
        end

        -- Substitute variables in parameters
        local resolved_params = {}
        for key, value in pairs(step_params) do
            resolved_params[key] = substitute_variables(value, variables)
        end

        -- Create isolated context for composed spell
        local old_context = spell_context
        local step_context = deep_copy(spell_context)
        step_context.execution_id = generate_uuid()
        step_context.parent_execution_id = old_context.execution_id
        step_context.step_name = step_name
        step_context.params = resolved_params

        spell_context = step_context

        -- Execute the spell step (simplified - in real implementation this would
        -- invoke the actual spell execution system)
        local step_result = {
            step = step_name,
            spell = spell_name,
            params = resolved_params,
            execution_id = step_context.execution_id,
            status = "completed",
            timestamp = os.time(),
        }

        -- For demo purposes, simulate some results based on spell name
        if spell_name == "web-fetcher" then
            step_result.content = "Fetched content from " .. (resolved_params.url or "unknown URL")
        elseif spell_name == "text-summarizer" then
            step_result.summary = "Summary of: " .. tostring(resolved_params.text or "no text")
        elseif spell_name == "web-search" then
            step_result.results = { "result1", "result2", "result3" }
        else
            step_result.output = "Output from " .. spell_name
        end

        -- Store result and update variables
        results[step_name] = step_result
        variables[step_name] = step_result

        -- Restore original context
        spell_context = old_context
    end

    return results
end

-- Create reusable libraries
function spell.library(name, functions)
    validate_required(name, "library name")
    validate_required(functions, "library functions")

    if type(functions) ~= "table" then
        error("library functions must be a table")
    end

    -- Validate that all values are functions
    for func_name, func in pairs(functions) do
        if type(func) ~= "function" then
            error("library function '" .. func_name .. "' must be a function")
        end
    end

    -- Register the library
    spell_libraries[name] = functions

    return functions
end

-- ============================================================================
-- Execution Context Management
-- ============================================================================

-- Access current execution context
function spell.context()
    return deep_copy(spell_context)
end

-- Access configuration values
function spell.config(key, default)
    if key then
        local value = spell_config[key]
        return value ~= nil and value or default
    else
        return deep_copy(spell_config)
    end
end

-- Caching with TTL support
function spell.cache(key, value, ttl)
    validate_required(key, "cache key")

    if value ~= nil then
        -- Set cache value
        ttl = ttl or spell_config.cache_ttl or 300
        spell_cache[key] = {
            value = value,
            expires = os.time() + ttl,
            created = os.time(),
        }
        return value
    else
        -- Get cache value
        local entry = spell_cache[key]
        if entry and os.time() < entry.expires then
            return entry.value
        elseif entry then
            -- Expired, remove it
            spell_cache[key] = nil
        end
        return nil
    end
end

-- ============================================================================
-- Advanced Features
-- ============================================================================

-- Error handling and recovery
function spell.on_error(handler)
    validate_required(handler, "error handler")

    if type(handler) ~= "function" then
        error("error handler must be a function")
    end

    table.insert(spell_hooks.on_error, handler)
end

-- Lifecycle hooks
function spell.on_init(handler)
    validate_required(handler, "init handler")

    if type(handler) ~= "function" then
        error("init handler must be a function")
    end

    table.insert(spell_hooks.on_init, handler)
end

function spell.on_start(handler)
    validate_required(handler, "start handler")

    if type(handler) ~= "function" then
        error("start handler must be a function")
    end

    table.insert(spell_hooks.on_start, handler)
end

function spell.on_complete(handler)
    validate_required(handler, "complete handler")

    if type(handler) ~= "function" then
        error("complete handler must be a function")
    end

    table.insert(spell_hooks.on_complete, handler)
end

function spell.on_cleanup(handler)
    validate_required(handler, "cleanup handler")

    if type(handler) ~= "function" then
        error("cleanup handler must be a function")
    end

    table.insert(spell_hooks.on_cleanup, handler)
end

-- Environment management
function spell.env(key, value)
    validate_required(key, "environment key")

    if value ~= nil then
        -- Set environment variable
        spell_environment[key] = value
        return value
    else
        -- Get environment variable
        return spell_environment[key]
    end
end

-- Create sandboxed execution environment
function spell.sandbox(config)
    config = config or {}

    local sandbox_env = {
        -- Safe globals
        type = type,
        pairs = pairs,
        ipairs = ipairs,
        next = next,
        tostring = tostring,
        tonumber = tonumber,
        string = string,
        table = table,
        math = math,
        os = {
            time = os.time,
            date = os.date,
            clock = os.clock,
        },

        -- Spell framework
        spell = spell,

        -- Custom additions from config
    }

    if config.globals then
        for k, v in pairs(config.globals) do
            sandbox_env[k] = v
        end
    end

    return sandbox_env
end

-- Resource management
function spell.resource(name, config)
    validate_required(name, "resource name")
    validate_required(config, "resource config")

    if type(config) ~= "table" then
        error("resource config must be a table")
    end

    local resource = {
        name = name,
        type = config.type or "generic",
        created = os.time(),
        cleanup = config.cleanup,
        data = config.data,
    }

    spell_resources[name] = resource

    return resource
end

function spell.cleanup_resources()
    for name, resource in pairs(spell_resources) do
        if resource.cleanup and type(resource.cleanup) == "function" then
            local success, err = pcall(resource.cleanup, resource)
            if not success then
                print("Warning: Failed to cleanup resource '" .. name .. "': " .. tostring(err))
            end
        end
    end

    -- Execute cleanup hooks
    for _, handler in ipairs(spell_hooks.on_cleanup) do
        local success, err = pcall(handler)
        if not success then
            print("Warning: Cleanup hook failed: " .. tostring(err))
        end
    end

    -- Clear resources
    spell_resources = {}
end

-- ============================================================================
-- Utility Functions
-- ============================================================================

-- Get registered libraries
function spell.get_libraries()
    return deep_copy(spell_libraries)
end

-- Get cache statistics
function spell.get_cache_stats()
    local stats = {
        total_entries = 0,
        expired_entries = 0,
        total_size = 0,
    }

    local current_time = os.time()
    for key, entry in pairs(spell_cache) do
        stats.total_entries = stats.total_entries + 1
        if current_time >= entry.expires then
            stats.expired_entries = stats.expired_entries + 1
        end
        stats.total_size = stats.total_size + 1 -- Simplified size calculation
    end

    return stats
end

-- Clear expired cache entries
function spell.clear_expired_cache()
    local current_time = os.time()
    local cleared = 0

    for key, entry in pairs(spell_cache) do
        if current_time >= entry.expires then
            spell_cache[key] = nil
            cleared = cleared + 1
        end
    end

    return cleared
end

-- Validate spell configuration
function spell.validate_config(config)
    if type(config) ~= "table" then
        return false, "config must be a table"
    end

    if config.name and type(config.name) ~= "string" then
        return false, "config.name must be a string"
    end

    if config.version and type(config.version) ~= "string" then
        return false, "config.version must be a string"
    end

    if config.params and type(config.params) ~= "table" then
        return false, "config.params must be a table"
    end

    -- Validate parameter definitions
    if config.params then
        for param_name, param_config in pairs(config.params) do
            if type(param_config) ~= "table" then
                return false, "parameter '" .. param_name .. "' config must be a table"
            end

            if param_config.type and type(param_config.type) ~= "string" then
                return false, "parameter '" .. param_name .. "' type must be a string"
            end

            if param_config.enum and type(param_config.enum) ~= "table" then
                return false, "parameter '" .. param_name .. "' enum must be a table"
            end
        end
    end

    return true
end

-- Get system information
function spell.get_system_info()
    return {
        lua_version = _VERSION,
        spell_framework_version = "1.0.0",
        current_spell = spell_config.name or "none",
        execution_id = spell_context.execution_id,
        libraries_loaded = (function()
            local count = 0
            for _ in pairs(spell_libraries) do
                count = count + 1
            end
            return count
        end)(),
        cache_entries = (function()
            local count = 0
            for _ in pairs(spell_cache) do
                count = count + 1
            end
            return count
        end)(),
        resources_managed = (function()
            local count = 0
            for _ in pairs(spell_resources) do
                count = count + 1
            end
            return count
        end)(),
        hooks_registered = {
            on_init = #spell_hooks.on_init,
            on_start = #spell_hooks.on_start,
            on_complete = #spell_hooks.on_complete,
            on_cleanup = #spell_hooks.on_cleanup,
            on_error = #spell_hooks.on_error,
        },
    }
end

-- Reset spell state (useful for testing)
function spell.reset()
    spell_config = {}
    spell_context = {}
    spell_params = {}
    spell_cache = {}
    spell_libraries = {}
    spell_hooks = {
        on_init = {},
        on_start = {},
        on_complete = {},
        on_cleanup = {},
        on_error = {},
    }
    spell_resources = {}
    spell_environment = {}
end

-- Export the module
return spell
