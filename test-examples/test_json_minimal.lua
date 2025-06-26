-- ABOUTME: Minimal test to check JSON functionality
-- ABOUTME: Isolate the exact problem with JSON encoding

-- Check bridges exist
if not bridges then
    return {error = "No bridges table"}
end

if not bridges.util_core then
    return {error = "No util_core bridge"}
end

if not bridges.util_json then
    return {error = "No util_json bridge"}
end

-- Try the simplest possible JSON encoding
local test_table = {key = "value"}

-- Method 1: Direct util_core.jsonEncode
local result1, err1 = nil, nil
if bridges.util_core.jsonEncode then
    result1, err1 = bridges.util_core.jsonEncode(test_table)
end

-- Method 2: Using data module
local data = require("data")
local result2 = nil
local ok, res = pcall(data.to_json, test_table)
if ok then
    result2 = res
end

return {
    util_core_exists = bridges.util_core ~= nil,
    util_json_exists = bridges.util_json ~= nil,
    jsonEncode_exists = bridges.util_core.jsonEncode ~= nil,
    direct_result = result1,
    direct_error = err1,
    data_result = result2,
    result1_type = type(result1),
    result2_type = type(result2)
}