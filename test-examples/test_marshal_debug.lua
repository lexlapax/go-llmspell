-- ABOUTME: Debug JSON marshal function calls
-- ABOUTME: Test the actual marshal bridge method execution

-- Test direct bridge method call
local json_bridge = bridges.util_json
local test_obj = {name = "test", value = 42}

-- Test marshal method directly
local success, result = pcall(json_bridge.marshal, test_obj)

return {
    success = success,
    result_type = type(result),
    result_value = result,
    error_message = not success and result or nil
}