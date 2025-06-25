-- ABOUTME: LLM pipeline hooks library for Lua that wraps the agent_hooks bridge
-- ABOUTME: Provides hook registration, priority ordering, lifecycle execution, and management operations

local hooks = {}

-- Version information
hooks._VERSION = "1.0.0"
hooks._DESCRIPTION = "LLM pipeline hooks library for Lua"

-- Get the hooks bridge
local function get_hooks_bridge()
    if bridges and bridges.agent_hooks then
        return bridges.agent_hooks
    end
    return nil
end

-- Check if bridge is available
local function with_hooks_bridge(method_name, ...)
    local hooks_bridge = get_hooks_bridge()
    if not hooks_bridge then
        error("Hooks bridge not available for " .. method_name)
    end
    return hooks_bridge[method_name](...)
end

-- Constants
hooks.TYPES = {
    BEFORE_GENERATE = "beforeGenerate",
    AFTER_GENERATE = "afterGenerate",
    BEFORE_TOOL_CALL = "beforeToolCall",
    AFTER_TOOL_CALL = "afterToolCall"
}

hooks.PRIORITY = {
    HIGHEST = 1000,
    HIGH = 100,
    NORMAL = 0,
    LOW = -100,
    LOWEST = -1000
}

-- Hook registration and management methods (snake_case)
function hooks.register_hook(hook_id, definition)
    return with_hooks_bridge("registerHook", hook_id, definition)
end

function hooks.unregister_hook(hook_id)
    return with_hooks_bridge("unregisterHook", hook_id)
end

function hooks.enable_hook(hook_id)
    return with_hooks_bridge("enableHook", hook_id)
end

function hooks.disable_hook(hook_id)
    return with_hooks_bridge("disableHook", hook_id)
end

function hooks.get_hook(hook_id)
    return with_hooks_bridge("getHook", hook_id)
end

function hooks.list_hooks(filter)
    if filter then
        return with_hooks_bridge("listHooks", filter)
    else
        return with_hooks_bridge("listHooks")
    end
end

function hooks.clear_hooks(hook_type)
    if hook_type then
        return with_hooks_bridge("clearHooks", hook_type)
    else
        return with_hooks_bridge("clearHooks")
    end
end

-- Hook execution methods (snake_case)
function hooks.execute_hooks(hook_type, context)
    return with_hooks_bridge("executeHooks", hook_type, context)
end

function hooks.execute_hook(hook_id, context)
    return with_hooks_bridge("executeHook", hook_id, context)
end

-- Priority management (snake_case)
function hooks.set_hook_priority(hook_id, priority)
    return with_hooks_bridge("setHookPriority", hook_id, priority)
end

function hooks.get_hook_priority(hook_id)
    return with_hooks_bridge("getHookPriority", hook_id)
end

-- Hook state queries (snake_case)
function hooks.is_hook_enabled(hook_id)
    return with_hooks_bridge("isHookEnabled", hook_id)
end

function hooks.get_enabled_hooks(hook_type)
    if hook_type then
        return with_hooks_bridge("getEnabledHooks", hook_type)
    else
        return with_hooks_bridge("getEnabledHooks")
    end
end

function hooks.get_disabled_hooks(hook_type)
    if hook_type then
        return with_hooks_bridge("getDisabledHooks", hook_type)
    else
        return with_hooks_bridge("getDisabledHooks")
    end
end

-- Batch operations (snake_case) - using the adapter's convenience methods
function hooks.batch_enable(hook_ids)
    return with_hooks_bridge("batchEnable", hook_ids)
end

function hooks.batch_disable(hook_ids)
    return with_hooks_bridge("batchDisable", hook_ids)
end

-- Hook builder pattern (snake_case)
function hooks.create_hook(hook_id)
    -- Use the adapter's createHook method
    return with_hooks_bridge("createHook", hook_id)
end

-- Alternative builder implementation
hooks.builder = {}
hooks.builder.__index = hooks.builder

function hooks.builder:new(hook_id)
    local builder = {
        _hook_id = hook_id,
        _definition = {
            priority = hooks.PRIORITY.NORMAL
        }
    }
    setmetatable(builder, self)
    return builder
end

function hooks.builder:with_priority(priority)
    self._definition.priority = priority
    return self
end

function hooks.builder:before_generate(fn)
    self._definition.beforeGenerate = fn
    return self
end

function hooks.builder:after_generate(fn)
    self._definition.afterGenerate = fn
    return self
end

function hooks.builder:before_tool_call(fn)
    self._definition.beforeToolCall = fn
    return self
end

function hooks.builder:after_tool_call(fn)
    self._definition.afterToolCall = fn
    return self
end

function hooks.builder:register()
    return hooks.register_hook(self._hook_id, self._definition)
end

-- Alternative constructor for builder
function hooks.new_builder(hook_id)
    return hooks.builder:new(hook_id)
end

-- Helper function to create a simple hook
function hooks.create_simple_hook(hook_id, hook_type, fn, priority)
    priority = priority or hooks.PRIORITY.NORMAL
    local definition = {
        priority = priority
    }
    
    if hook_type == hooks.TYPES.BEFORE_GENERATE then
        definition.beforeGenerate = fn
    elseif hook_type == hooks.TYPES.AFTER_GENERATE then
        definition.afterGenerate = fn
    elseif hook_type == hooks.TYPES.BEFORE_TOOL_CALL then
        definition.beforeToolCall = fn
    elseif hook_type == hooks.TYPES.AFTER_TOOL_CALL then
        definition.afterToolCall = fn
    else
        error("Invalid hook type: " .. tostring(hook_type))
    end
    
    return hooks.register_hook(hook_id, definition)
end

-- Bridge access to raw methods if available
if bridges and bridges.agent_hooks then
    -- Allow access to raw bridge methods
    hooks.bridge = bridges.agent_hooks
end

return hooks