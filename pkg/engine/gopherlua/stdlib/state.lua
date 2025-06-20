-- ABOUTME: State Management Library for go-llmspell Lua standard library
-- ABOUTME: Provides state creation, persistence, transforms, validation, and context management

local state = {}

-- Import promise library for async operations
local promise = _G.promise or require("promise")

-- Internal state tracking
local active_states = {}
-- local state_schemas = {} -- Reserved for future schema registration
local state_transforms = {}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to get state manager bridge
local function get_state_manager()
    if not _G.state_manager then
        error("State manager bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.state_manager
end

-- Helper function to get state context bridge
local function get_state_context()
    if not _G.state_context then
        error("State context bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.state_context
end

-- Helper function to generate unique state ID
local function generate_state_id()
    return "state_" .. tostring(os.time()) .. "_" .. tostring(math.random(1000, 9999))
end

-- Helper function to merge tables (currently unused but may be needed for advanced merging)
-- local function merge_tables(...)
--     local result = {}
--     for _, t in ipairs({ ... }) do
--         if type(t) == "table" then
--             for k, v in pairs(t) do
--                 result[k] = v
--             end
--         end
--     end
--     return result
-- end

-- Context and State Utilities

-- Create a new state with optional initial data
function state.create(initial_data)
    local manager = get_state_manager()

    -- Create state through bridge
    local state_obj = manager:createState()

    if state_obj then
        -- Set initial data if provided
        if initial_data and type(initial_data) == "table" then
            for key, value in pairs(initial_data) do
                manager:set(state_obj, key, value)
            end
        end

        -- Track active state
        local state_id = state_obj.id or generate_state_id()
        active_states[state_id] = {
            created = os.time(),
            modified = os.time(),
            data = initial_data or {},
        }

        return state_obj
    end

    return nil
end

-- Merge multiple states into a new state
function state.merge(state1, state2, ...)
    validate_required(state1, "state1")
    validate_required(state2, "state2")

    local manager = get_state_manager()
    local states = { state1, state2, ... }

    -- Use merge strategy "merge_all" by default
    return manager:mergeStates(states, "merge_all")
end

-- Create a snapshot of current state
function state.snapshot(state_obj)
    validate_required(state_obj, "state")

    -- Create a deep copy of the state
    local snapshot = state.create()
    local manager = get_state_manager()

    -- Copy all keys from original state
    local keys = manager:keys(state_obj)
    for _, key in ipairs(keys) do
        local result = manager:get(state_obj, key)
        if result and result.value ~= nil then
            manager:set(snapshot, key, result.value)
        end
    end

    -- Copy metadata
    local metadata = manager:getAllMetadata(state_obj)
    if metadata then
        for key, value in pairs(metadata) do
            manager:setMetadata(snapshot, key, value)
        end
    end

    return snapshot
end

-- Restore state from a snapshot
function state.restore(snapshot)
    validate_required(snapshot, "snapshot")

    -- Create new state from snapshot data
    return state.create(snapshot.data)
end

-- State Persistence Helpers

-- Save state with optional key for identification
function state.save(state_obj, key)
    validate_required(state_obj, "state")

    local save_key = key or generate_state_id()
    local manager = get_state_manager()

    -- Set save key as metadata
    manager:setMetadata(state_obj, "_save_key", save_key)

    -- Save state through bridge
    local success = manager:saveState(state_obj)

    if success then
        -- Update tracking
        if active_states[state_obj.id] then
            active_states[state_obj.id].saved = os.time()
            active_states[state_obj.id].save_key = save_key
        end
    end

    return success, save_key
end

-- Load state by key with optional default
function state.load(key, default)
    validate_required(key, "key")

    local manager = get_state_manager()

    -- Try to load state
    local loaded = manager:loadState(key)

    if loaded then
        -- Track loaded state
        if loaded.id then
            active_states[loaded.id] = {
                loaded = os.time(),
                load_key = key,
            }
        end
        return loaded
    end

    -- Return default if provided
    if default ~= nil then
        return state.create(default)
    end

    return nil
end

-- Set expiration for state (TTL support)
function state.expire(key, duration)
    validate_required(key, "key")
    validate_required(duration, "duration")

    -- Schedule state deletion after duration
    -- This is a simplified implementation - in production, use a proper scheduler
    local manager = get_state_manager()

    promise.Promise.new(function(resolve, reject)
        -- Wait for duration (in seconds)
        promise
            .sleep(duration * 1000)
            :andThen(function()
                -- Delete the state
                manager:deleteState(key)
                resolve(true)
            end)
            :onError(function(err)
                reject(err)
            end)
    end)

    return true
end

-- State Transformation Utilities

-- Transform state using a transformer function
function state.transform(state_obj, transformer)
    validate_required(state_obj, "state")
    validate_required(transformer, "transformer")

    if type(transformer) ~= "function" then
        error("transformer must be a function")
    end

    -- Create a snapshot for transformation
    local transformed = state.snapshot(state_obj)

    -- Apply transformer
    local success, result = pcall(transformer, transformed)

    if success then
        return result or transformed
    else
        error("Transform failed: " .. tostring(result))
    end
end

-- Filter state based on predicate
function state.filter(state_obj, predicate)
    validate_required(state_obj, "state")
    validate_required(predicate, "predicate")

    if type(predicate) ~= "function" then
        error("predicate must be a function")
    end

    local manager = get_state_manager()
    local filtered = state.create()

    -- Get all keys and filter
    local keys = manager:keys(state_obj)
    for _, key in ipairs(keys) do
        local result = manager:get(state_obj, key)
        if result and result.value ~= nil then
            -- Apply predicate
            if predicate(key, result.value) then
                manager:set(filtered, key, result.value)
            end
        end
    end

    return filtered
end

-- Validate state against schema
function state.validate(state_obj, schema)
    validate_required(state_obj, "state")
    validate_required(schema, "schema")

    -- Simple validation implementation
    local errors = {}
    local manager = get_state_manager()

    -- Check required fields
    if schema.required and type(schema.required) == "table" then
        for _, field in ipairs(schema.required) do
            local result = manager:get(state_obj, field)
            if not result or result.value == nil then
                table.insert(errors, "Missing required field: " .. field)
            end
        end
    end

    -- Check field types
    if schema.properties and type(schema.properties) == "table" then
        local keys = manager:keys(state_obj)
        for _, key in ipairs(keys) do
            if schema.properties[key] then
                local result = manager:get(state_obj, key)
                if result and result.value ~= nil then
                    local expected_type = schema.properties[key].type
                    local actual_type = type(result.value)

                    -- Basic type checking
                    if expected_type and actual_type ~= expected_type then
                        -- Handle number/integer special case
                        if not (expected_type == "integer" and actual_type == "number") then
                            table.insert(
                                errors,
                                string.format(
                                    "Field '%s' has wrong type. Expected: %s, Got: %s",
                                    key,
                                    expected_type,
                                    actual_type
                                )
                            )
                        end
                    end
                end
            end
        end
    end

    return {
        valid = #errors == 0,
        errors = errors,
    }
end

-- Advanced State Operations

-- Create a state context with parent-child relationships
function state.create_context(parent_state)
    local context = get_state_context()

    -- Create shared context with optional parent
    return context:createSharedContext(parent_state)
end

-- Configure inheritance for shared context
function state.configure_inheritance(context, inherit_messages, inherit_artifacts, inherit_metadata)
    validate_required(context, "context")

    local ctx = get_state_context()

    -- Set inheritance configuration
    return ctx:withInheritanceConfig(
        context,
        inherit_messages ~= false, -- Default true
        inherit_artifacts ~= false, -- Default true
        inherit_metadata ~= false -- Default true
    )
end

-- Get value from shared context (with inheritance)
function state.get_from_context(context, key)
    validate_required(context, "context")
    validate_required(key, "key")

    local ctx = get_state_context()
    return ctx:get(context, key)
end

-- Set value in shared context (local only)
function state.set_in_context(context, key, value)
    validate_required(context, "context")
    validate_required(key, "key")

    local ctx = get_state_context()
    return ctx:set(context, key, value)
end

-- Register a named transform
function state.register_transform(name, transform_func)
    validate_required(name, "name")
    validate_required(transform_func, "transform_func")

    if type(transform_func) ~= "function" then
        error("transform_func must be a function")
    end

    -- Store transform locally
    state_transforms[name] = transform_func

    -- Could also register with bridge if supported
    -- local manager = get_state_manager()
    -- manager:registerTransform(name, transform_func)

    return true
end

-- Apply a named transform
function state.apply_transform(state_obj, transform_name)
    validate_required(state_obj, "state")
    validate_required(transform_name, "transform_name")

    local transform_func = state_transforms[transform_name]
    if not transform_func then
        error("Transform not found: " .. transform_name)
    end

    return state.transform(state_obj, transform_func)
end

-- Batch operations for efficiency
function state.batch_set(state_obj, data)
    validate_required(state_obj, "state")
    validate_required(data, "data")

    if type(data) ~= "table" then
        error("data must be a table")
    end

    local manager = get_state_manager()

    -- Set all key-value pairs
    for key, value in pairs(data) do
        manager:set(state_obj, key, value)
    end

    return state_obj
end

-- Get all state data as a table
function state.to_table(state_obj)
    validate_required(state_obj, "state")

    local manager = get_state_manager()
    local result = {}

    -- Get all keys and values
    local keys = manager:keys(state_obj)
    for _, key in ipairs(keys) do
        local value_result = manager:get(state_obj, key)
        if value_result and value_result.value ~= nil then
            result[key] = value_result.value
        end
    end

    return result
end

-- Create state from table
function state.from_table(data)
    validate_required(data, "data")

    if type(data) ~= "table" then
        error("data must be a table")
    end

    return state.create(data)
end

-- State comparison
function state.equals(state1, state2)
    validate_required(state1, "state1")
    validate_required(state2, "state2")

    local manager = get_state_manager()

    -- Get keys from both states
    local keys1 = manager:keys(state1)
    local keys2 = manager:keys(state2)

    -- Quick check: different number of keys
    if #keys1 ~= #keys2 then
        return false
    end

    -- Check all keys and values
    for _, key in ipairs(keys1) do
        local result1 = manager:get(state1, key)
        local result2 = manager:get(state2, key)

        if not result2 or result1.value ~= result2.value then
            return false
        end
    end

    return true
end

-- State diffing
function state.diff(state1, state2)
    validate_required(state1, "state1")
    validate_required(state2, "state2")

    local manager = get_state_manager()
    local diff = {
        added = {},
        removed = {},
        modified = {},
    }

    -- Get all keys
    local keys1 = manager:keys(state1)
    local keys2 = manager:keys(state2)

    -- Create key sets for comparison
    local key_set1 = {}
    local key_set2 = {}

    for _, key in ipairs(keys1) do
        key_set1[key] = true
    end

    for _, key in ipairs(keys2) do
        key_set2[key] = true
    end

    -- Find added and modified keys
    for _, key in ipairs(keys2) do
        if not key_set1[key] then
            -- Key added in state2
            local result = manager:get(state2, key)
            if result and result.value ~= nil then
                diff.added[key] = result.value
            end
        else
            -- Key exists in both, check if modified
            local result1 = manager:get(state1, key)
            local result2 = manager:get(state2, key)

            if result1 and result2 and result1.value ~= result2.value then
                diff.modified[key] = {
                    old = result1.value,
                    new = result2.value,
                }
            end
        end
    end

    -- Find removed keys
    for _, key in ipairs(keys1) do
        if not key_set2[key] then
            local result = manager:get(state1, key)
            if result and result.value ~= nil then
                diff.removed[key] = result.value
            end
        end
    end

    return diff
end

-- Utility function to clean up old states
function state.cleanup(age_seconds)
    local current_time = os.time()
    local cleaned = 0

    for state_id, info in pairs(active_states) do
        if info.created and (current_time - info.created) > age_seconds then
            active_states[state_id] = nil
            cleaned = cleaned + 1
        end
    end

    return cleaned
end

-- Export state module
return state
