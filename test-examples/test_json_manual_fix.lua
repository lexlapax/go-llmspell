-- ABOUTME: Test a manual fix for JSON encoding
-- ABOUTME: See if we can work around the nil return issue

print("=== JSON Manual Fix Test ===")

-- Test 1: Check what util_core.jsonEncode actually returns
if bridges and bridges.util_core and bridges.util_core.jsonEncode then
    print("\n--- Direct jsonEncode test ---")
    local test_obj = {name = "test", value = 42}
    
    -- Since jsonEncode returns (result, error), capture both
    local result, err = bridges.util_core.jsonEncode(test_obj)
    
    print("Raw return values:")
    print("  1st value:", result)
    print("  1st type:", type(result))
    print("  2nd value:", err)
    print("  2nd type:", type(err))
    
    -- If result is nil but error is also nil, something's wrong with the bridge
    if result == nil and err == nil then
        print("ERROR: Both result and error are nil - bridge not working correctly")
    end
end

-- Test 2: Try to patch data.to_json temporarily
local data = require("data")
local original_to_json = data.to_json

-- Override to_json with a working version
data.to_json = function(obj)
    -- For now, just return a hardcoded JSON string to test
    return '{"mocked":true,"original_failed":true}'
end

-- Test the override
print("\n--- Testing patched to_json ---")
local json = data.to_json({test = "value"})
print("Patched result:", json)
print("Type:", type(json))
print("Length:", json and #json or "nil")

-- Restore original
data.to_json = original_to_json

-- Test 3: Check if we can access the JSON bridge directly
print("\n--- Checking JSON bridge directly ---")
if bridges and bridges.util_json then
    print("util_json bridge exists")
    -- Try to iterate its methods (won't work but let's try)
    local count = 0
    for k, v in pairs(bridges.util_json) do
        print("  Found key:", k)
        count = count + 1
    end
    print("Total keys found:", count)
end

return {
    jsonEncode_exists = bridges and bridges.util_core and bridges.util_core.jsonEncode ~= nil,
    json_bridge_exists = bridges and bridges.util_json ~= nil
}