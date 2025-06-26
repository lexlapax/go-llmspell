-- ABOUTME: Test JSON bridge return value handling
-- ABOUTME: Debug why JSON methods return nil

print("=== JSON Bridge Return Value Test ===")

-- Test direct bridge call with manual error handling
print("\n--- Testing util_core.jsonEncode ---")
if bridges and bridges.util_core and bridges.util_core.jsonEncode then
    local test_obj = {name = "test", value = 42}
    
    -- Call jsonEncode - it returns (result, error)
    local result, error = bridges.util_core.jsonEncode(test_obj)
    
    print("Return values:")
    print("  Result:", result)
    print("  Result type:", type(result))
    print("  Error:", error)
    print("  Error type:", type(error))
    
    -- If result is nil, check error
    if result == nil then
        print("Result is nil - checking error message")
        if error then
            print("Error message:", tostring(error))
        end
    elseif type(result) == "string" then
        print("Success! JSON string:", result)
        print("Length:", #result)
    else
        print("Unexpected result type:", type(result))
    end
end

-- Test data module handling
print("\n--- Testing data.to_json handling ---")
local data = require("data")
local test_obj2 = {foo = "bar", num = 123}

-- Manually wrap to see what happens
local ok, result_or_error = pcall(function()
    return data.to_json(test_obj2)
end)

print("pcall results:")
print("  Success:", ok)
print("  Result/Error:", result_or_error)
print("  Type:", type(result_or_error))

return {
    direct_call_works = bridges and bridges.util_core and bridges.util_core.jsonEncode ~= nil,
    data_module_works = ok and type(result_or_error) == "string" and #result_or_error > 0
}