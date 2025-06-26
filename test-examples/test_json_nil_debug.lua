-- ABOUTME: Debug why jsonBridge is nil in UtilsAdapter
-- ABOUTME: Test the actual error message from jsonEncode

print("=== JSON Nil Debug Test ===")

-- Create a simple wrapper to see the actual error
local function test_json_encode(obj)
    if not bridges or not bridges.util_core then
        return nil, "No util_core bridge"
    end
    
    if not bridges.util_core.jsonEncode then
        return nil, "No jsonEncode method"
    end
    
    -- Call jsonEncode and capture all return values
    local result, err = bridges.util_core.jsonEncode(obj)
    
    -- Debug output
    return result, err
end

-- Test with a simple table
local test_obj = {name = "test", value = 42}
local result, err = test_json_encode(test_obj)

-- Also test data module
local data = require("data")
local data_ok, data_result = pcall(data.to_json, test_obj)

return {
    manual_result = result,
    manual_error = err,
    manual_result_type = type(result),
    data_ok = data_ok,
    data_result = data_result,
    data_result_type = type(data_result)
}