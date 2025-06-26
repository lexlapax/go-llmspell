-- ABOUTME: Test data.from_json alias
-- ABOUTME: Verify the alias works correctly

local data = require("data")

-- Check if functions exist
local has_parse_json = type(data.parse_json) == "function"
local has_from_json = type(data.from_json) == "function"
local same_function = data.parse_json == data.from_json

-- Test from_json
local json_str = '{"test":"value","number":42}'
local result = nil
local error = nil

if has_from_json then
    local ok, res = pcall(data.from_json, json_str)
    if ok then
        result = res
    else
        error = res
    end
end

return {
    has_parse_json = has_parse_json,
    has_from_json = has_from_json,
    same_function = same_function,
    result = result,
    error = error and tostring(error) or nil,
    result_type = type(result),
    result_test = result and result.test or nil,
    result_number = result and result.number or nil
}