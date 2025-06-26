-- ABOUTME: Test bridge availability and basic functionality
-- ABOUTME: Debug bridge connections and method calls

print("=== Bridge Debug Test ===")

-- Check if bridges table exists
print("Bridges available:", type(bridges))
if bridges then
    print("Bridges keys:")
    for k, v in pairs(bridges) do
        print("  " .. k .. ":", type(v))
    end
end

-- Check util_core specifically
print("util_core available:", type(bridges and bridges.util_core))
if bridges and bridges.util_core then
    local util = bridges.util_core
    print("util_core methods:")
    for k, v in pairs(util) do
        print("  " .. k .. ":", type(v))
    end
    
    -- Test basic JSON encoding
    print("\nTesting JSON encoding...")
    local test_obj = {name = "test", value = 42}
    
    local ok, result = pcall(util.jsonEncode, test_obj)
    print("jsonEncode result:", ok, type(result), result)
end

return {
    bridges_available = bridges ~= nil,
    util_core_available = bridges and bridges.util_core ~= nil
}