-- ABOUTME: Structured Data and Schema Library for go-llmspell Lua standard library
-- ABOUTME: Provides schema validation, generation, and conversion utilities for structured data handling

local structured = {}

-- Import promise library for async operations
local promise = _G.promise or require("promise")

-- Helper function to get structured bridge
local function get_structured_bridge()
    if not bridges or not bridges.structured_schema then
        error("Structured bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return bridges.structured_schema
end

-- Helper function to validate required parameters
local function validate_required(param, name)
    if not param then
        error(name .. " is required")
    end
end

-- Schema Creation and Management

-- Create a new schema from a definition
function structured.create_schema(schema_data)
    validate_required(schema_data, "schema_data")
    local bridge = get_structured_bridge()
    return bridge:createSchema(schema_data)
end

-- Create a property definition with constraints
function structured.create_property(property_type, constraints)
    validate_required(property_type, "property_type")
    local bridge = get_structured_bridge()
    return bridge:createProperty(property_type, constraints)
end

-- Validation Functions

-- Validate JSON data against a schema
function structured.validate_json(schema, data)
    validate_required(schema, "schema")
    validate_required(data, "data")
    local bridge = get_structured_bridge()
    return bridge:validateJSON(schema, data)
end

-- Validate a struct against a schema
function structured.validate_struct(schema, data)
    validate_required(schema, "schema")
    validate_required(data, "data")
    local bridge = get_structured_bridge()
    return bridge:validateStruct(schema, data)
end

-- Asynchronous validation with promise support
function structured.validate_async(schema, data, callback)
    validate_required(schema, "schema")
    validate_required(data, "data")

    local Promise = promise.Promise or promise
    return Promise.new(function(resolve, reject)
        local bridge = get_structured_bridge()
        local result, err = bridge:validateAsync(schema, data, callback)
        if err then
            reject(err)
        else
            resolve(result)
        end
    end)
end

-- Schema Generation

-- Generate schema from type information
function structured.generate_from_type(type_info)
    validate_required(type_info, "type_info")
    local bridge = get_structured_bridge()
    return bridge:generateSchemaFromType(type_info)
end

-- Generate schema from struct tags
function structured.generate_from_tags(struct_data)
    validate_required(struct_data, "struct_data")
    local bridge = get_structured_bridge()
    return bridge:generateFromTags(struct_data)
end

-- Generate schema with documentation
function structured.generate_with_docs(struct_data, include_examples)
    validate_required(struct_data, "struct_data")
    local bridge = get_structured_bridge()
    return bridge:generateWithDocumentation(struct_data, include_examples)
end

-- Schema Conversion

-- Convert JSON Schema string to schema object
function structured.from_json_schema(json_schema)
    validate_required(json_schema, "json_schema")
    local bridge = get_structured_bridge()
    return bridge:convertJSONSchema(json_schema)
end

-- Export schema to JSON Schema format
function structured.to_json_schema(schema)
    validate_required(schema, "schema")
    local bridge = get_structured_bridge()
    return bridge:exportToJSONSchema(schema)
end

-- Export schema to OpenAPI format
function structured.to_openapi(schema)
    validate_required(schema, "schema")
    local bridge = get_structured_bridge()
    return bridge:exportToOpenAPI(schema)
end

-- Convert between different schema formats
function structured.convert_format(schema, from_format, to_format)
    validate_required(schema, "schema")
    validate_required(from_format, "from_format")
    validate_required(to_format, "to_format")
    local bridge = get_structured_bridge()
    return bridge:convertFormat(schema, from_format, to_format)
end

-- Repository Operations

-- Initialize file-based repository
function structured.init_repository(directory)
    validate_required(directory, "directory")
    local bridge = get_structured_bridge()
    return bridge:initializeFileRepository(directory)
end

-- Save schema to repository
function structured.save_schema(name, schema)
    validate_required(name, "name")
    validate_required(schema, "schema")
    local bridge = get_structured_bridge()
    return bridge:saveSchema(name, schema)
end

-- Get schema from repository
function structured.get_schema(name)
    validate_required(name, "name")
    local bridge = get_structured_bridge()
    return bridge:getSchema(name)
end

-- Delete schema from repository
function structured.delete_schema(name)
    validate_required(name, "name")
    local bridge = get_structured_bridge()
    return bridge:deleteSchema(name)
end

-- Versioning Support

-- Save a specific version of a schema
function structured.save_version(name, schema, version)
    validate_required(name, "name")
    validate_required(schema, "schema")
    validate_required(version, "version")
    local bridge = get_structured_bridge()
    return bridge:saveSchemaVersion(name, schema, version)
end

-- Get a specific version of a schema
function structured.get_version(name, version)
    validate_required(name, "name")
    validate_required(version, "version")
    local bridge = get_structured_bridge()
    return bridge:getSchemaVersion(name, version)
end

-- List all versions of a schema
function structured.list_versions(name)
    validate_required(name, "name")
    local bridge = get_structured_bridge()
    return bridge:listSchemaVersions(name)
end

-- Schema Operations

-- Merge multiple schemas
function structured.merge_schemas(schemas, strategy)
    validate_required(schemas, "schemas")
    if type(schemas) ~= "table" or #schemas < 2 then
        error("schemas must be an array with at least 2 schemas")
    end
    local bridge = get_structured_bridge()
    return bridge:mergeSchemas(schemas, strategy)
end

-- Generate diff between schemas
function structured.diff_schemas(old_schema, new_schema)
    validate_required(old_schema, "old_schema")
    validate_required(new_schema, "new_schema")
    local bridge = get_structured_bridge()
    return bridge:generateDiff(old_schema, new_schema)
end

-- Import/Export Operations

-- Import schema from file
function structured.import_from_file(file_path, format)
    validate_required(file_path, "file_path")
    local bridge = get_structured_bridge()
    return bridge:importFromFile(file_path, format)
end

-- Import schema from string
function structured.import_from_string(content, format)
    validate_required(content, "content")
    local bridge = get_structured_bridge()
    return bridge:importFromString(content, format)
end

-- Export repository to JSON
function structured.export_repository()
    local bridge = get_structured_bridge()
    return bridge:exportRepository()
end

-- Import repository from JSON
function structured.import_repository(data)
    validate_required(data, "data")
    local bridge = get_structured_bridge()
    return bridge:importRepository(data)
end

-- Custom Validators

-- Register custom validator
function structured.register_validator(name, validator)
    validate_required(name, "name")
    validate_required(validator, "validator")
    local bridge = get_structured_bridge()
    return bridge:registerCustomValidator(name, validator)
end

-- Register conditional validator
function structured.register_conditional_validator(name, validator)
    validate_required(name, "name")
    validate_required(validator, "validator")
    local bridge = get_structured_bridge()
    return bridge:registerConditionalValidator(name, validator)
end

-- Validate with custom validator
function structured.validate_custom(data, validator_name)
    validate_required(data, "data")
    validate_required(validator_name, "validator_name")
    local bridge = get_structured_bridge()
    return bridge:validateWithCustom(data, validator_name)
end

-- List all custom validators
function structured.list_validators()
    local bridge = get_structured_bridge()
    return bridge:listCustomValidators()
end

-- Utility Functions

-- Get validation metrics
function structured.get_metrics()
    local bridge = get_structured_bridge()
    return bridge:getValidationMetrics()
end

-- Clear validation cache
function structured.clear_cache()
    local bridge = get_structured_bridge()
    return bridge:clearValidationCache()
end

-- Convenience Functions

-- Quick validate function with automatic schema detection
function structured.validate(schema_or_name, data)
    validate_required(schema_or_name, "schema_or_name")
    validate_required(data, "data")

    local schema = schema_or_name
    if type(schema_or_name) == "string" then
        -- If string, assume it's a schema name and fetch from repository
        schema = structured.get_schema(schema_or_name)
    end

    return structured.validate_json(schema, data)
end

-- Create and save schema in one operation
function structured.create_and_save(name, schema_data)
    validate_required(name, "name")
    validate_required(schema_data, "schema_data")

    local schema = structured.create_schema(schema_data)
    structured.save_schema(name, schema)
    return schema
end

-- Batch validation
function structured.validate_batch(schema, data_array)
    validate_required(schema, "schema")
    validate_required(data_array, "data_array")

    if type(data_array) ~= "table" then
        error("data_array must be an array")
    end

    local results = {}
    for i, data in ipairs(data_array) do
        local result, err = structured.validate_json(schema, data)
        local is_valid = result and result.valid or false
        local error_msg = nil
        if not is_valid and result and result.errors and #result.errors > 0 then
            error_msg = table.concat(result.errors, "; ")
        end
        results[i] = {
            index = i,
            valid = is_valid,
            result = result,
            error = error_msg or err
        }
    end

    return results
end

-- Example schema builder helper
function structured.builder()
    local builder = {
        schema = {
            type = "object",
            properties = {},
            required = {}
        }
    }

    function builder:add_property(name, property_type, required, constraints)
        self.schema.properties[name] = structured.create_property(property_type, constraints)
        if required then
            table.insert(self.schema.required, name)
        end
        return self
    end

    function builder:add_string(name, required, min_length, max_length, pattern)
        local constraints = {}
        if min_length then constraints.minLength = min_length end
        if max_length then constraints.maxLength = max_length end
        if pattern then constraints.pattern = pattern end
        return self:add_property(name, "string", required, constraints)
    end

    function builder:add_number(name, required, minimum, maximum)
        local constraints = {}
        if minimum then constraints.minimum = minimum end
        if maximum then constraints.maximum = maximum end
        return self:add_property(name, "number", required, constraints)
    end

    function builder:add_boolean(name, required)
        return self:add_property(name, "boolean", required, {})
    end

    function builder:add_array(name, required, items, min_items, max_items)
        local constraints = {items = items}
        if min_items then constraints.minItems = min_items end
        if max_items then constraints.maxItems = max_items end
        return self:add_property(name, "array", required, constraints)
    end

    function builder:add_object(name, required, properties)
        local constraints = {properties = properties}
        return self:add_property(name, "object", required, constraints)
    end

    function builder:build()
        return structured.create_schema(self.schema)
    end

    return builder
end

return structured