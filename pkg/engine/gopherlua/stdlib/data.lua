-- ABOUTME: Structured Data Library for go-llmspell Lua standard library
-- ABOUTME: Provides JSON processing, schema validation, data transformation, and format conversion

local data = {}

-- Internal tracking (reserved for future use)
-- local cached_schemas = {}
-- local format_converters = {}
-- local transformation_cache = {}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to get utils bridge for JSON operations
local function get_utils()
    if not _G.util then
        error("Utils bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.util
end

-- Helper function to get structured bridge for schema operations (reserved for future use)
-- local function get_structured()
--     if not _G.structured then
--         error("Structured bridge not available. Ensure go-llmspell is properly initialized.")
--     end
--     return _G.structured
-- end

-- JSON and Data Processing Utilities

-- Parse JSON with optional schema validation
function data.parse_json(text, schema)
    validate_required(text, "text")

    if type(text) ~= "string" then
        error("text must be a string")
    end

    local util = get_utils()

    if schema then
        -- Parse with schema validation
        if type(schema) ~= "table" then
            error("schema must be a table")
        end

        local success, result = pcall(util.jsonParse, text, schema)
        if not success then
            error("JSON parsing failed: " .. tostring(result))
        end

        return result
    else
        -- Parse without validation
        local success, result = pcall(util.jsonDecode, text)
        if not success then
            error("JSON parsing failed: " .. tostring(result))
        end

        return result
    end
end

-- Convert object to JSON with formatting options
function data.to_json(object, format)
    validate_required(object, "object")

    local util = get_utils()
    format = format or {}

    if type(format) == "string" then
        -- Simple format string
        if format == "pretty" then
            local success, result = pcall(util.jsonPrettify, util.jsonEncode(object))
            if not success then
                error("JSON formatting failed: " .. tostring(result))
            end
            return result
        else
            local success, result = pcall(util.jsonEncode, object)
            if not success then
                error("JSON encoding failed: " .. tostring(result))
            end
            return result
        end
    elseif type(format) == "table" then
        -- Detailed format options
        local success, result = pcall(util.jsonToJSON, object, format)
        if not success then
            error("JSON formatting failed: " .. tostring(result))
        end
        return result
    else
        -- Default encoding
        local success, result = pcall(util.jsonEncode, object)
        if not success then
            error("JSON encoding failed: " .. tostring(result))
        end
        return result
    end
end

-- Extract structured data from text (especially LLM output) with schema
function data.extract_structured(text, schema)
    validate_required(text, "text")
    validate_required(schema, "schema")

    if type(text) ~= "string" then
        error("text must be a string")
    end

    if type(schema) ~= "table" then
        error("schema must be a table")
    end

    local util = get_utils()

    local success, result = pcall(util.jsonExtractStructuredData, text, schema)
    if not success then
        error("Structured data extraction failed: " .. tostring(result))
    end

    return result
end

-- Convert data between formats (JSON, YAML, XML, etc.)
function data.convert_format(data_obj, from_format, to_format)
    validate_required(data_obj, "data")
    validate_required(from_format, "from_format")
    validate_required(to_format, "to_format")

    -- For now, implement basic JSON conversion
    -- In a full implementation, this would use the bridge's format conversion
    if from_format == "json" and to_format == "json" then
        -- JSON to JSON (essentially a format/pretty operation)
        return data.to_json(data_obj, "pretty")
    elseif from_format == "string" and to_format == "json" then
        -- Parse string as JSON
        return data.parse_json(data_obj)
    elseif to_format == "json" then
        -- Convert to JSON
        return data.to_json(data_obj)
    else
        -- For complex conversions, we'd use the bridge
        error(
            "Format conversion from "
                .. from_format
                .. " to "
                .. to_format
                .. " not yet implemented"
        )
    end
end

-- Schema Validation Helpers

-- Validate data against a schema
function data.validate(data_obj, schema)
    validate_required(data_obj, "data")
    validate_required(schema, "schema")

    if type(schema) ~= "table" then
        error("schema must be a table")
    end

    local util = get_utils()

    -- Convert data to JSON for validation
    local json_data = data.to_json(data_obj)

    local success, result = pcall(util.jsonValidateJSONSchema, json_data, schema)
    if not success then
        return {
            valid = false,
            errors = { tostring(result) },
        }
    end

    if result.valid == false or result.valid == nil then
        return {
            valid = false,
            errors = result.errors or { "Validation failed" },
        }
    end

    return {
        valid = true,
        errors = {},
    }
end

-- Infer schema from data
function data.infer_schema(data_obj)
    validate_required(data_obj, "data")

    -- Basic schema inference implementation
    local function infer_type(value)
        local value_type = type(value)

        if value_type == "string" then
            return { type = "string" }
        elseif value_type == "number" then
            if value % 1 == 0 then
                return { type = "integer" }
            else
                return { type = "number" }
            end
        elseif value_type == "boolean" then
            return { type = "boolean" }
        elseif value_type == "table" then
            -- Check if it's an array or object
            local is_array = true
            local count = 0
            for k, _ in pairs(value) do
                count = count + 1
                if type(k) ~= "number" or k ~= count then
                    is_array = false
                    break
                end
            end

            if is_array and count > 0 then
                -- Array type - infer item schema from first element
                local item_schema = infer_type(value[1])
                return {
                    type = "array",
                    items = item_schema,
                }
            else
                -- Object type - infer properties
                local properties = {}
                local required = {}

                for k, v in pairs(value) do
                    properties[k] = infer_type(v)
                    table.insert(required, k)
                end

                return {
                    type = "object",
                    properties = properties,
                    required = required,
                }
            end
        else
            return { type = "null" }
        end
    end

    return infer_type(data_obj)
end

-- Migrate data from old schema to new schema
function data.migrate_schema(data_obj, old_schema, new_schema)
    validate_required(data_obj, "data")
    validate_required(old_schema, "old_schema")
    validate_required(new_schema, "new_schema")

    if type(old_schema) ~= "table" or type(new_schema) ~= "table" then
        error("schemas must be tables")
    end

    -- Basic migration: validate against old, then transform to new
    local old_validation = data.validate(data_obj, old_schema)
    if not old_validation.valid then
        error("Data doesn't match old schema: " .. table.concat(old_validation.errors, ", "))
    end

    -- For now, return the data as-is
    -- In a full implementation, this would perform intelligent migration
    local new_validation = data.validate(data_obj, new_schema)
    if new_validation.valid then
        return data_obj
    else
        -- Try basic property mapping
        if type(data_obj) == "table" and new_schema.type == "object" then
            local migrated = {}
            local new_props = new_schema.properties or {}

            for prop, prop_schema in pairs(new_props) do
                if data_obj[prop] ~= nil then
                    migrated[prop] = data_obj[prop]
                elseif prop_schema.default ~= nil then
                    migrated[prop] = prop_schema.default
                end
            end

            return migrated
        end

        error("Schema migration failed: " .. table.concat(new_validation.errors, ", "))
    end
end

-- Data Transformation Utilities

-- Map function for collections
function data.map(collection, mapper)
    validate_required(collection, "collection")
    validate_required(mapper, "mapper")

    if type(collection) ~= "table" then
        error("collection must be a table")
    end

    if type(mapper) ~= "function" then
        error("mapper must be a function")
    end

    local result = {}

    -- Check if it's an array-like table
    local is_array = true
    local count = 0
    for k, _ in pairs(collection) do
        count = count + 1
        if type(k) ~= "number" or k ~= count then
            is_array = false
            break
        end
    end

    if is_array then
        -- Array mapping
        for i, value in ipairs(collection) do
            local mapped_value = mapper(value, i, collection)
            table.insert(result, mapped_value)
        end
    else
        -- Object mapping
        for key, value in pairs(collection) do
            local mapped_value = mapper(value, key, collection)
            result[key] = mapped_value
        end
    end

    return result
end

-- Filter function for collections
function data.filter(collection, predicate)
    validate_required(collection, "collection")
    validate_required(predicate, "predicate")

    if type(collection) ~= "table" then
        error("collection must be a table")
    end

    if type(predicate) ~= "function" then
        error("predicate must be a function")
    end

    local result = {}

    -- Check if it's an array-like table
    local is_array = true
    local count = 0
    for k, _ in pairs(collection) do
        count = count + 1
        if type(k) ~= "number" or k ~= count then
            is_array = false
            break
        end
    end

    if is_array then
        -- Array filtering
        for i, value in ipairs(collection) do
            if predicate(value, i, collection) then
                table.insert(result, value)
            end
        end
    else
        -- Object filtering
        for key, value in pairs(collection) do
            if predicate(value, key, collection) then
                result[key] = value
            end
        end
    end

    return result
end

-- Reduce function for aggregation
function data.reduce(collection, reducer, initial)
    validate_required(collection, "collection")
    validate_required(reducer, "reducer")

    if type(collection) ~= "table" then
        error("collection must be a table")
    end

    if type(reducer) ~= "function" then
        error("reducer must be a function")
    end

    local accumulator = initial
    local has_initial = initial ~= nil

    -- Check if it's an array-like table
    local is_array = true
    local count = 0
    for k, _ in pairs(collection) do
        count = count + 1
        if type(k) ~= "number" or k ~= count then
            is_array = false
            break
        end
    end

    if is_array then
        -- Array reduction
        local start_index = 1

        if not has_initial and #collection > 0 then
            accumulator = collection[1]
            start_index = 2
        end

        for i = start_index, #collection do
            accumulator = reducer(accumulator, collection[i], i, collection)
        end
    else
        -- Object reduction
        local first = true

        for key, value in pairs(collection) do
            if not has_initial and first then
                accumulator = value
                first = false
            else
                accumulator = reducer(accumulator, value, key, collection)
            end
        end
    end

    return accumulator
end

-- Advanced Utilities

-- Deep clone a data structure
function data.clone(obj)
    validate_required(obj, "obj")

    local function deep_copy(orig)
        local orig_type = type(orig)
        local copy

        if orig_type == "table" then
            copy = {}
            for orig_key, orig_value in next, orig, nil do
                copy[deep_copy(orig_key)] = deep_copy(orig_value)
            end
            setmetatable(copy, deep_copy(getmetatable(orig)))
        else
            copy = orig
        end

        return copy
    end

    return deep_copy(obj)
end

-- Deep merge two objects
function data.merge(obj1, obj2)
    validate_required(obj1, "obj1")
    validate_required(obj2, "obj2")

    if type(obj1) ~= "table" or type(obj2) ~= "table" then
        error("both objects must be tables")
    end

    local function deep_merge(target, source)
        for key, value in pairs(source) do
            if type(value) == "table" and type(target[key]) == "table" then
                target[key] = deep_merge(target[key], value)
            else
                target[key] = value
            end
        end
        return target
    end

    return deep_merge(data.clone(obj1), obj2)
end

-- Get nested value with dot notation
function data.get_path(obj, path)
    validate_required(obj, "obj")
    validate_required(path, "path")

    if type(obj) ~= "table" then
        return nil
    end

    if type(path) ~= "string" then
        error("path must be a string")
    end

    local keys = {}
    for key in path:gmatch("[^%.]+") do
        table.insert(keys, key)
    end

    local current = obj
    for _, key in ipairs(keys) do
        if type(current) ~= "table" then
            return nil
        end

        -- Try as string key first, then as number
        current = current[key]
        if current == nil then
            local num_key = tonumber(key)
            if num_key then
                current = current[num_key]
            end
        end

        if current == nil then
            return nil
        end
    end

    return current
end

-- Set nested value with dot notation
function data.set_path(obj, path, value)
    validate_required(obj, "obj")
    validate_required(path, "path")

    if type(obj) ~= "table" then
        error("obj must be a table")
    end

    if type(path) ~= "string" then
        error("path must be a string")
    end

    local keys = {}
    for key in path:gmatch("[^%.]+") do
        table.insert(keys, key)
    end

    local current = obj
    for i = 1, #keys - 1 do
        local key = keys[i]

        if type(current[key]) ~= "table" then
            current[key] = {}
        end

        current = current[key]
    end

    current[keys[#keys]] = value
    return obj
end

-- Export the module
return data
