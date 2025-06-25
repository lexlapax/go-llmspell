-- ABOUTME: State Management Library for go-llmspell - simplified wrapper around state_context bridge
-- ABOUTME: Provides state.get/set/update methods with dot-notation path support

local state = {}

-- Get the state_context bridge
local function get_state_context()
    if not bridges or not bridges.state_context then
        error("State context bridge not available. Ensure go-llmspell is properly initialized with state feature set.")
    end
    return bridges.state_context
end

-- Internal state storage using the bridge
local _context = nil
local _initialized = false

-- Initialize state context
local function ensure_initialized()
    if not _initialized then
        local bridge = get_state_context()
        -- Create a shared context for the script
        local success, ctx = pcall(function()
            return bridge.createSharedContext("script_state")
        end)
        if not success then
            error("Failed to create state context: " .. tostring(ctx))
        end
        _context = ctx
        _initialized = true
    end
    return _context
end

-- Parse dot-notation path into parts
local function parse_path(path)
    if not path or path == "" then
        return {}
    end
    
    local parts = {}
    for part in string.gmatch(path, "[^%.]+") do
        table.insert(parts, part)
    end
    return parts
end

-- Get value at path from a table
local function get_nested(tbl, parts)
    local current = tbl
    for _, part in ipairs(parts) do
        if type(current) ~= "table" then
            return nil
        end
        current = current[part]
    end
    return current
end

-- Set value at path in a table
local function set_nested(tbl, parts, value)
    if #parts == 0 then
        return value
    end
    
    local current = tbl
    for i = 1, #parts - 1 do
        local part = parts[i]
        if type(current[part]) ~= "table" then
            current[part] = {}
        end
        current = current[part]
    end
    
    current[parts[#parts]] = value
    return tbl
end

-- Get a value from state
function state.get(path)
    ensure_initialized()
    local bridge = get_state_context()
    
    if not path or path == "" then
        -- Return all state
        local all_state = {}
        local keys = bridge.keys(_context)
        if keys then
            for _, key in ipairs(keys) do
                all_state[key] = bridge.get(_context, key)
            end
        end
        return all_state
    end
    
    -- Check if it's a simple key or a path
    if not string.find(path, "%.") then
        -- Simple key
        local success, value = pcall(function()
            return bridge.get(_context, path)
        end)
        if success then
            return value
        else
            return nil
        end
    end
    
    -- Complex path - need to handle nested access
    local parts = parse_path(path)
    if #parts == 0 then
        return nil
    end
    
    -- Get the root object
    local root_key = parts[1]
    local root_value = bridge.get(_context, root_key)
    
    if #parts == 1 then
        return root_value
    end
    
    -- Navigate the nested structure
    if type(root_value) ~= "table" then
        return nil
    end
    
    -- Remove the first part and get nested value
    table.remove(parts, 1)
    return get_nested(root_value, parts)
end

-- Set a value in state
function state.set(path, value)
    ensure_initialized()
    local bridge = get_state_context()
    
    if not path or path == "" then
        error("Path is required for state.set")
    end
    
    -- Check if it's a simple key or a path
    if not string.find(path, "%.") then
        -- Simple key
        local success, err = pcall(function()
            return bridge.set(_context, path, value)
        end)
        if not success then
            error("Failed to set state: " .. tostring(err))
        end
        return
    end
    
    -- Complex path - need to handle nested setting
    local parts = parse_path(path)
    if #parts == 0 then
        return
    end
    
    -- Get the root object
    local root_key = parts[1]
    local root_value = bridge.get(_context, root_key)
    
    if #parts == 1 then
        -- Setting root level
        bridge.set(_context, root_key, value)
        return
    end
    
    -- Create nested structure if needed
    if type(root_value) ~= "table" then
        root_value = {}
    end
    
    -- Remove the first part and set nested value
    local root_key_backup = parts[1]
    table.remove(parts, 1)
    local updated_root = set_nested(root_value, parts, value)
    
    -- Set the updated root back
    bridge.set(_context, root_key_backup, updated_root)
end

-- Update a value in state with a function
function state.update(path, update_fn)
    if type(update_fn) ~= "function" then
        error("Update function is required for state.update")
    end
    
    local current_value = state.get(path)
    local new_value = update_fn(current_value)
    state.set(path, new_value)
    return new_value
end

-- Delete a value from state
function state.delete(path)
    ensure_initialized()
    local bridge = get_state_context()
    
    if not path or path == "" then
        error("Path is required for state.delete")
    end
    
    -- Simple key
    if not string.find(path, "%.") then
        bridge.delete(_context, path)
        return
    end
    
    -- Complex path - set to nil
    state.set(path, nil)
end

-- Check if a path exists in state
function state.has(path)
    ensure_initialized()
    local bridge = get_state_context()
    
    if not path or path == "" then
        return false
    end
    
    -- Simple key
    if not string.find(path, "%.") then
        return bridge.has(_context, path)
    end
    
    -- Complex path
    return state.get(path) ~= nil
end

-- Get all keys in state
function state.keys()
    ensure_initialized()
    local bridge = get_state_context()
    return bridge.keys(_context) or {}
end

-- Clear all state
function state.clear()
    ensure_initialized()
    local bridge = get_state_context()
    bridge.clearContext(_context)
end

-- Create a snapshot of current state
function state.snapshot()
    ensure_initialized()
    local bridge = get_state_context()
    return bridge.createSnapshot(_context)
end

-- Save state to persistent storage
function state.save(name)
    ensure_initialized()
    local bridge = get_state_context()
    name = name or "default"
    return bridge.saveState(_context, name)
end

-- Load state from persistent storage
function state.load(name)
    ensure_initialized()
    local bridge = get_state_context()
    name = name or "default"
    return bridge.loadState(_context, name)
end

-- Export state as a table
function state.export()
    ensure_initialized()
    local bridge = get_state_context()
    return bridge.exportState(_context)
end

-- Import state from a table
function state.import(data)
    ensure_initialized()
    local bridge = get_state_context()
    return bridge.importState(_context, data)
end

return state