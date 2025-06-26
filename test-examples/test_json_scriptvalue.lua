-- ABOUTME: Test JSON bridge ScriptValue handling
-- ABOUTME: Verify the adapter correctly converts ScriptValue results

print("=== JSON ScriptValue Test ===")

-- Create a test to see what happens at each step
local data = require("data")

-- Override the util_core methods to debug
local original_jsonEncode = bridges.util_core.jsonEncode

-- Wrap jsonEncode to see what's happening
bridges.util_core.jsonEncode = function(obj)
    print("\n--- jsonEncode wrapper called ---")
    print("Input type:", type(obj))
    print("Input value:", obj)
    
    -- Call original
    local result, err = original_jsonEncode(obj)
    
    print("Result:", result)
    print("Result type:", type(result))
    print("Error:", err)
    print("Error type:", type(err))
    
    -- Check if result is a userdata (ScriptValue wrapper)
    if type(result) == "userdata" then
        print("Result is userdata - might be ScriptValue wrapper")
        local mt = getmetatable(result)
        if mt then
            print("Has metatable")
            for k, v in pairs(mt) do
                print("  Meta key:", k, "Type:", type(v))
            end
        end
    end
    
    return result, err
end

-- Now test data.to_json
print("\n=== Testing data.to_json ===")
local test_obj = {name = "test", value = 42}
local json_result = data.to_json(test_obj)

print("\nFinal result:")
print("  Type:", type(json_result))
print("  Value:", json_result)
print("  Length:", json_result and type(json_result) == "string" and #json_result or "N/A")

-- Restore original
bridges.util_core.jsonEncode = original_jsonEncode

return {
    success = json_result ~= nil and type(json_result) == "string" and #json_result > 0
}