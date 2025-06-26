-- ABOUTME: Test how jsonEncode returns values
-- ABOUTME: Check if it returns (result, error) or just result

print("=== JSON Return Values Test ===")

-- Test without pcall to see raw returns
if bridges and bridges.util_core and bridges.util_core.jsonEncode then
    print("\n--- Raw jsonEncode call ---")
    local test_obj = {name = "test", value = 42}
    
    -- Call without pcall to see all return values
    local r1, r2, r3 = bridges.util_core.jsonEncode(test_obj)
    
    print("Return value 1:", r1, "Type:", type(r1))
    print("Return value 2:", r2, "Type:", type(r2))
    print("Return value 3:", r3, "Type:", type(r3))
    
    -- Test with pcall
    print("\n--- With pcall ---")
    local success, result = pcall(bridges.util_core.jsonEncode, test_obj)
    print("pcall success:", success)
    print("pcall result:", result, "Type:", type(result))
    
    -- If pcall succeeded but result is nil, try without pcall
    if success and result == nil then
        print("\n--- Trying different approach ---")
        -- Maybe the adapter expects the table in a different way?
        local json_result = bridges.util_core.jsonEncode(test_obj)
        print("Direct result:", json_result, "Type:", type(json_result))
    end
end

-- Test the data module's approach
print("\n--- Data module approach ---")
local data = require("data")

-- Manually replicate what data.to_json does
local function test_to_json(object)
    local util = bridges.util_core
    if not util then
        return nil, "No util_core bridge"
    end
    
    print("Calling util.jsonEncode...")
    local success, result = pcall(util.jsonEncode, object)
    print("  pcall success:", success)
    print("  pcall result:", result)
    print("  result type:", type(result))
    
    if not success then
        return nil, "JSON encoding failed: " .. tostring(result)
    end
    
    return result
end

local json, err = test_to_json({foo = "bar"})
print("\nFinal result:", json)
print("Final error:", err)

return {
    test_complete = true
}