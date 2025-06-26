-- ABOUTME: Test direct access to JSON bridge
-- ABOUTME: Verify bridge availability and marshal function

print("=== Direct JSON Bridge Test ===")

-- Check bridges table
print("Bridges available:", type(bridges))
if bridges then
    for k, v in pairs(bridges) do
        print("Bridge:", k, type(v))
    end
end

-- Check util_json specifically
print("util_json available:", type(bridges and bridges.util_json))

if bridges and bridges.util_json then
    local json_bridge = bridges.util_json
    print("JSON bridge methods:")
    for k, v in pairs(json_bridge) do
        print("  " .. k .. ":", type(v))
    end
    
    -- Test marshal function directly
    print("\nTesting marshal...")
    local test_obj = {name = "test", value = 42}
    
    local ok, result = pcall(json_bridge.marshal, test_obj)
    print("Marshal result:", ok, type(result), result)
    
    if ok and result then
        print("JSON string length:", #result)
        print("JSON content:", result)
    end
else
    print("util_json bridge not available")
end

return {
    bridges_available = bridges ~= nil,
    util_json_available = bridges and bridges.util_json ~= nil,
    marshal_test = bridges and bridges.util_json and true or false
}