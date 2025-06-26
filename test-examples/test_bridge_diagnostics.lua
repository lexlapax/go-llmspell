-- ABOUTME: Diagnostic test to check bridge and adapter availability
-- ABOUTME: Verifies util_json bridge methods are callable

print("=== Bridge Diagnostics ===")

-- Check bridges table
print("\n--- Checking bridges table ---")
if bridges then
    print("bridges table exists")
    for k, v in pairs(bridges) do
        print("  " .. k .. ":", type(v))
    end
else
    print("bridges table is nil!")
end

-- Check util_json specifically
print("\n--- Checking util_json bridge ---")
if bridges and bridges.util_json then
    print("util_json bridge exists")
    print("Type:", type(bridges.util_json))
    
    -- Try to list methods
    local count = 0
    for k, v in pairs(bridges.util_json) do
        print("  Method:", k, "Type:", type(v))
        count = count + 1
    end
    print("Total methods found:", count)
    
    -- Try to call marshal directly
    print("\n--- Testing marshal method ---")
    if bridges.util_json.marshal then
        print("marshal method exists")
        local test_obj = {name = "test", value = 42}
        local ok, result = pcall(bridges.util_json.marshal, test_obj)
        print("Marshal call success:", ok)
        print("Result type:", type(result))
        print("Result:", result)
    else
        print("marshal method not found!")
    end
else
    print("util_json bridge not found!")
end

-- Check util_core for comparison
print("\n--- Checking util_core bridge ---")
if bridges and bridges.util_core then
    print("util_core bridge exists")
    local count = 0
    for k, v in pairs(bridges.util_core) do
        count = count + 1
    end
    print("Total methods found:", count)
    
    -- Check if jsonEncode exists
    if bridges.util_core.jsonEncode then
        print("jsonEncode method exists in util_core")
    end
else
    print("util_core bridge not found!")
end

return {
    bridges_exist = bridges ~= nil,
    util_json_exists = bridges and bridges.util_json ~= nil,
    util_core_exists = bridges and bridges.util_core ~= nil,
    marshal_exists = bridges and bridges.util_json and bridges.util_json.marshal ~= nil
}