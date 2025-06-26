-- ABOUTME: Test what methods are available on util_core
-- ABOUTME: Check if JSON methods are accessible through util_core

print("=== Utils Methods Test ===")

-- Check util_core methods
print("\n--- util_core methods ---")
if bridges and bridges.util_core then
    local methods = {}
    for k, v in pairs(bridges.util_core) do
        if type(v) == "function" then
            table.insert(methods, k)
        end
    end
    table.sort(methods)
    
    print("Found", #methods, "methods:")
    for _, method in ipairs(methods) do
        print("  " .. method)
    end
    
    -- Test JSON methods
    print("\n--- Testing JSON methods on util_core ---")
    if bridges.util_core.jsonEncode then
        print("jsonEncode exists - testing...")
        local test_obj = {name = "test", value = 42}
        local ok, result = pcall(bridges.util_core.jsonEncode, test_obj)
        print("  Success:", ok)
        print("  Result:", result)
    end
    
    if bridges.util_core.jsonDecode then
        print("\njsonDecode exists - testing...")
        local test_json = '{"name":"test","value":42}'
        local ok, result = pcall(bridges.util_core.jsonDecode, test_json)
        print("  Success:", ok)
        if ok and type(result) == "table" then
            print("  Decoded name:", result.name)
            print("  Decoded value:", result.value)
        end
    end
else
    print("util_core not found!")
end

return {
    util_core_exists = bridges and bridges.util_core ~= nil,
    jsonEncode_exists = bridges and bridges.util_core and bridges.util_core.jsonEncode ~= nil,
    jsonDecode_exists = bridges and bridges.util_core and bridges.util_core.jsonDecode ~= nil
}