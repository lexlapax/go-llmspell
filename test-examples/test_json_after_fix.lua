-- ABOUTME: Test JSON functionality after engine fix
-- ABOUTME: Verify the multi-bridge adapter recreation works

-- Check bridges
if not bridges or not bridges.util_core then
    return {error = "No util_core bridge"}
end

-- Test JSON encoding
local test_obj = {name = "test", value = 42, nested = {key = "value"}}
local result, err = bridges.util_core.jsonEncode(test_obj)

-- Also test through data module
local data = require("data")
local data_result = nil
local data_error = nil
local ok, res = pcall(function()
    return data.to_json(test_obj)
end)
if ok then
    data_result = res
else
    data_error = res
end

return {
    direct_result = result,
    direct_error = err,
    direct_type = type(result),
    data_result = data_result,
    data_error = data_error and tostring(data_error) or nil,
    data_type = type(data_result),
    success = result ~= nil and type(result) == "string"
}