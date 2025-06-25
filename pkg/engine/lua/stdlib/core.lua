-- ABOUTME: Core Utilities Library for go-llmspell Lua standard library
-- ABOUTME: Provides essential string, table, crypto, and time utilities

-- luacheck: globals string table os (allow extensions to built-in types)
-- luacheck: no unused args

local core = {}

-- Import other modules for reuse
local function safe_require(module)
    local ok, result = pcall(require, module)
    return ok and result or nil
end

-- Try to load other stdlib modules if available
local testing = safe_require("testing")
local _ = safe_require("data") -- Not used in current implementation

-- ============================================================================
-- String Utilities - Extend the string table
-- ============================================================================

-- Template string with variable substitution
function string.template(template, variables)
    if type(template) ~= "string" then
        error("template must be a string")
    end
    if type(variables) ~= "table" then
        error("variables must be a table")
    end

    -- Replace {{variable}} patterns
    return (
        string.gsub(template, "{{([^}]+)}}", function(key)
            key = string.match(key, "^%s*(.-)%s*$") -- trim whitespace
            local value = variables[key]
            if value == nil then
                return "{{" .. key .. "}}" -- Keep original if not found
            end
            return tostring(value)
        end)
    )
end

-- Convert to URL-safe slug
function string.slugify(text)
    if type(text) ~= "string" then
        error("text must be a string")
    end

    -- Convert to lowercase
    text = string.lower(text)

    -- Replace non-alphanumeric characters with hyphens
    text = string.gsub(text, "[^%w%-_]", "-")

    -- Replace multiple hyphens with single hyphen
    text = string.gsub(text, "%-+", "-")

    -- Remove leading/trailing hyphens
    text = string.gsub(text, "^%-+", "")
    text = string.gsub(text, "%-+$", "")

    return text
end

-- Truncate string with optional suffix
function string.truncate(text, length, suffix)
    if type(text) ~= "string" then
        error("text must be a string")
    end
    if type(length) ~= "number" or length < 0 then
        error("length must be a non-negative number")
    end

    suffix = suffix or "..."

    if #text <= length then
        return text
    end

    -- Account for suffix length
    local truncate_at = length - #suffix
    if truncate_at < 0 then
        truncate_at = 0
    end

    return string.sub(text, 1, truncate_at) .. suffix
end

-- Split string by delimiter
function string.split(str, delimiter)
    if type(str) ~= "string" then
        error("str must be a string")
    end

    delimiter = delimiter or " "
    local result = {}
    local pattern = "(.-)" .. delimiter
    local last_pos = 1
    local s, e, cap = string.find(str, pattern, 1)

    while s do
        if s ~= 1 or cap ~= "" then
            table.insert(result, cap)
        end
        last_pos = e + 1
        s, e, cap = string.find(str, pattern, last_pos)
    end

    if last_pos <= #str then
        cap = string.sub(str, last_pos)
        table.insert(result, cap)
    end

    return result
end

-- Remove leading/trailing whitespace
function string.trim(str)
    if type(str) ~= "string" then
        error("str must be a string")
    end
    return string.match(str, "^%s*(.-)%s*$")
end

-- Capitalize first letter
function string.capitalize(str)
    if type(str) ~= "string" then
        error("str must be a string")
    end
    return string.upper(string.sub(str, 1, 1)) .. string.sub(str, 2)
end

-- Convert to camelCase
function string.camelcase(str)
    if type(str) ~= "string" then
        error("str must be a string")
    end

    -- First, convert to lowercase and split by non-alphanumeric
    local words = {}
    for word in string.gmatch(str, "[%w]+") do
        table.insert(words, string.lower(word))
    end

    -- Capitalize all words except the first
    for i = 2, #words do
        words[i] = string.capitalize(words[i])
    end

    return table.concat(words, "")
end

-- Convert to snake_case
function string.snakecase(str)
    if type(str) ~= "string" then
        error("str must be a string")
    end

    -- Handle camelCase by inserting underscores before capitals
    str = string.gsub(str, "([a-z])([A-Z])", "%1_%2")

    -- Replace non-alphanumeric with underscores
    str = string.gsub(str, "[^%w_]", "_")

    -- Replace multiple underscores with single
    str = string.gsub(str, "_+", "_")

    -- Remove leading/trailing underscores and lowercase
    str = string.lower(str)
    str = string.gsub(str, "^_+", "")
    str = string.gsub(str, "_+$", "")

    return str
end

-- ============================================================================
-- Table Utilities - Extend the table table
-- ============================================================================

-- Extract keys from table
function table.keys(tbl)
    if type(tbl) ~= "table" then
        error("tbl must be a table")
    end

    local keys = {}
    for k, _ in pairs(tbl) do
        table.insert(keys, k)
    end
    return keys
end

-- Extract values from table
function table.values(tbl)
    if type(tbl) ~= "table" then
        error("tbl must be a table")
    end

    local values = {}
    for _, v in pairs(tbl) do
        table.insert(values, v)
    end
    return values
end

-- Shallow merge tables
function table.merge(...)
    local result = {}

    for _, tbl in ipairs({ ... }) do
        if type(tbl) ~= "table" then
            error("all arguments must be tables")
        end
        for k, v in pairs(tbl) do
            result[k] = v
        end
    end

    return result
end

-- Deep copy table (use data.clone if available, otherwise implement)
function table.deep_copy(tbl)
    -- Always use our implementation which handles circular references
    -- The data.clone function may not handle all edge cases properly
    local function deep_copy_impl(obj, seen)
        if type(obj) ~= "table" then
            return obj
        end

        if seen[obj] then
            return seen[obj]
        end

        local copy = {}
        seen[obj] = copy

        for k, v in pairs(obj) do
            copy[deep_copy_impl(k, seen)] = deep_copy_impl(v, seen)
        end

        return setmetatable(copy, getmetatable(obj))
    end

    return deep_copy_impl(tbl, {})
end

-- Check if table is empty
function table.is_empty(tbl)
    if type(tbl) ~= "table" then
        error("tbl must be a table")
    end
    return next(tbl) == nil
end

-- Array slice
function table.slice(tbl, start_idx, end_idx)
    if type(tbl) ~= "table" then
        error("tbl must be a table")
    end

    start_idx = start_idx or 1
    end_idx = end_idx or #tbl

    local result = {}
    for i = start_idx, end_idx do
        table.insert(result, tbl[i])
    end
    return result
end

-- Reverse array in-place
function table.reverse(tbl)
    if type(tbl) ~= "table" then
        error("tbl must be a table")
    end

    local len = #tbl
    for i = 1, math.floor(len / 2) do
        tbl[i], tbl[len - i + 1] = tbl[len - i + 1], tbl[i]
    end
    return tbl
end

-- Shuffle array in-place
function table.shuffle(tbl)
    if type(tbl) ~= "table" then
        error("tbl must be a table")
    end

    for i = #tbl, 2, -1 do
        local j = math.random(1, i)
        tbl[i], tbl[j] = tbl[j], tbl[i]
    end
    return tbl
end

-- Check if table contains value
function table.contains(tbl, value)
    if type(tbl) ~= "table" then
        error("tbl must be a table")
    end

    for _, v in pairs(tbl) do
        if v == value then
            return true
        end
    end
    return false
end

-- ============================================================================
-- Crypto Utilities
-- ============================================================================

local crypto = {}

-- Generate UUID (use testing.data.uuid if available)
function crypto.uuid()
    if testing and testing.data and testing.data.uuid then
        return testing.data.uuid()
    end

    -- Fallback UUID v4 implementation
    local template = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
    return string.gsub(template, "[xy]", function(c)
        local v = (c == "x") and math.random(0, 0xf) or math.random(8, 0xb)
        return string.format("%x", v)
    end)
end

-- Hash data using various algorithms
function crypto.hash(data_str, algorithm)
    if type(data_str) ~= "string" then
        error("data must be a string")
    end
    if type(algorithm) ~= "string" then
        error("algorithm must be a string")
    end

    algorithm = string.lower(algorithm)

    -- This would typically call into Go crypto functions
    -- For now, return a placeholder that indicates the function needs bridge support
    error("crypto.hash requires bridge implementation for " .. algorithm)
end

-- Generate random string
function crypto.random_string(length, charset)
    if testing and testing.data and testing.data.random_string then
        return testing.data.random_string(length)
    end

    -- Fallback implementation
    length = length or 16
    charset = charset or "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

    local result = {}
    for i = 1, length do
        local idx = math.random(1, #charset)
        result[i] = string.sub(charset, idx, idx)
    end
    return table.concat(result)
end

-- Base64 encoding
function crypto.base64_encode(data_str)
    if type(data_str) ~= "string" then
        error("data must be a string")
    end

    -- This requires bridge implementation
    error("crypto.base64_encode requires bridge implementation")
end

-- Base64 decoding
function crypto.base64_decode(encoded)
    if type(encoded) ~= "string" then
        error("encoded must be a string")
    end

    -- This requires bridge implementation
    error("crypto.base64_decode requires bridge implementation")
end

-- URL-safe base64 encoding
function crypto.base64url_encode(data_str)
    if type(data_str) ~= "string" then
        error("data must be a string")
    end

    -- This requires bridge implementation
    error("crypto.base64url_encode requires bridge implementation")
end

-- URL-safe base64 decoding
function crypto.base64url_decode(encoded)
    if type(encoded) ~= "string" then
        error("encoded must be a string")
    end

    -- This requires bridge implementation
    error("crypto.base64url_decode requires bridge implementation")
end

-- ============================================================================
-- Time Utilities - Extend the os table
-- ============================================================================

-- Get current timestamp in seconds with fractional part
function os.now()
    -- Use os.clock for higher precision if available
    return os.time() + (os.clock() % 1)
end

-- Format timestamp using strftime patterns
function os.format(timestamp, format)
    if type(timestamp) ~= "number" then
        error("timestamp must be a number")
    end
    format = format or "%Y-%m-%d %H:%M:%S"

    return os.date(format, math.floor(timestamp))
end

-- Calculate duration between timestamps
function os.duration(start_time, end_time)
    if type(start_time) ~= "number" then
        error("start_time must be a number")
    end
    if type(end_time) ~= "number" then
        error("end_time must be a number")
    end

    local seconds = end_time - start_time

    return {
        seconds = seconds,
        minutes = seconds / 60,
        hours = seconds / 3600,
        days = seconds / 86400,
    }
end

-- Parse time string to timestamp
function os.parse_time(time_str, format)
    if type(time_str) ~= "string" then
        error("time_str must be a string")
    end

    -- This is complex and would benefit from bridge implementation
    -- For now, provide basic ISO date parsing
    if not format or format == "%Y-%m-%d" then
        local year, month, day = string.match(time_str, "(%d+)-(%d+)-(%d+)")
        if year and month and day then
            return os.time({ year = tonumber(year), month = tonumber(month), day = tonumber(day) })
        end
    end

    error("os.parse_time: complex format parsing requires bridge implementation")
end

-- Add time to timestamp
function os.add_time(timestamp, duration)
    if type(timestamp) ~= "number" then
        error("timestamp must be a number")
    end
    if type(duration) ~= "table" then
        error("duration must be a table")
    end

    local seconds = timestamp

    if duration.seconds then
        seconds = seconds + duration.seconds
    end
    if duration.minutes then
        seconds = seconds + (duration.minutes * 60)
    end
    if duration.hours then
        seconds = seconds + (duration.hours * 3600)
    end
    if duration.days then
        seconds = seconds + (duration.days * 86400)
    end

    return seconds
end

-- Human-readable duration
function os.humanize_duration(seconds)
    if type(seconds) ~= "number" then
        error("seconds must be a number")
    end

    local parts = {}

    -- Days
    if seconds >= 86400 then
        local days = math.floor(seconds / 86400)
        table.insert(parts, days .. " day" .. (days ~= 1 and "s" or ""))
        seconds = seconds % 86400
    end

    -- Hours
    if seconds >= 3600 then
        local hours = math.floor(seconds / 3600)
        table.insert(parts, hours .. " hour" .. (hours ~= 1 and "s" or ""))
        seconds = seconds % 3600
    end

    -- Minutes
    if seconds >= 60 then
        local minutes = math.floor(seconds / 60)
        table.insert(parts, minutes .. " minute" .. (minutes ~= 1 and "s" or ""))
        seconds = seconds % 60
    end

    -- Seconds
    if seconds > 0 or #parts == 0 then
        local secs = math.floor(seconds)
        table.insert(parts, secs .. " second" .. (secs ~= 1 and "s" or ""))
    end

    return table.concat(parts, " ")
end

-- ============================================================================
-- Miscellaneous Utilities
-- ============================================================================

-- Type checking utilities
function core.is_callable(value)
    local t = type(value)
    if t == "function" then
        return true
    elseif t == "table" then
        local mt = getmetatable(value)
        return mt ~= nil and type(mt.__call) == "function"
    end
    return false
end

function core.is_array(value)
    if type(value) ~= "table" then
        return false
    end

    local count = 0
    for _ in pairs(value) do
        count = count + 1
    end

    -- Check if all keys are sequential integers starting from 1
    for i = 1, count do
        if value[i] == nil then
            return false
        end
    end

    return true
end

function core.is_object(value)
    return type(value) == "table" and not core.is_array(value)
end

-- Function utilities

-- Create debounced function
function core.debounce(func, delay)
    if type(func) ~= "function" then
        error("func must be a function")
    end
    if type(delay) ~= "number" or delay < 0 then
        error("delay must be a non-negative number")
    end

    local timer = nil
    local pending_args = nil

    return function(...)
        pending_args = { ... }

        if timer then
            -- Cancel existing timer
            timer = nil
        end

        -- This is simplified - real implementation would need proper timer support
        -- For now, just call immediately after storing args
        local args = pending_args
        pending_args = nil
        return func(unpack(args)) -- luacheck: ignore 113
    end
end

-- Create throttled function
function core.throttle(func, delay)
    if type(func) ~= "function" then
        error("func must be a function")
    end
    if type(delay) ~= "number" or delay < 0 then
        error("delay must be a non-negative number")
    end

    local last_call = 0

    return function(...)
        local now = os.clock()
        if now - last_call >= delay then
            last_call = now
            return func(...)
        end
    end
end

-- Create memoized function
function core.memoize(func)
    if type(func) ~= "function" then
        error("func must be a function")
    end

    local cache = {}

    return function(...)
        local key = table.concat({ ... }, "\0")
        if cache[key] == nil then
            cache[key] = { func(...) }
        end
        return unpack(cache[key]) -- luacheck: ignore 113
    end
end

-- Error handling

-- Try-catch pattern
function core.try(func, catch_func)
    if type(func) ~= "function" then
        error("func must be a function")
    end

    local success, result = pcall(func)

    if success then
        return result
    elseif catch_func then
        return catch_func(result)
    else
        error(result)
    end
end

-- Safe function call
function core.safe_call(func, ...)
    if type(func) ~= "function" then
        error("func must be a function")
    end

    return pcall(func, ...)
end

-- ============================================================================
-- Module Setup
-- ============================================================================

-- Add crypto namespace to core
core.crypto = crypto

-- Export the module
return core
