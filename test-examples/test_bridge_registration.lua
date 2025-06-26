-- ABOUTME: Test what bridges are actually registered
-- ABOUTME: Check if util_json bridge is available

print("=== Bridge Registration Test ===")

-- List all bridges
print("\n--- All registered bridges ---")
if bridges then
    local bridge_list = {}
    for name, bridge in pairs(bridges) do
        table.insert(bridge_list, name)
    end
    table.sort(bridge_list)
    
    for _, name in ipairs(bridge_list) do
        print("  " .. name)
    end
    print("Total bridges:", #bridge_list)
else
    print("No bridges table!")
end

-- Check specific utility bridges
print("\n--- Utility bridges ---")
local util_bridges = {
    "util_core",
    "util_json", 
    "util_debug",
    "util_errors",
    "util_script_logger",
    "util_slog",
    "auth"
}

for _, name in ipairs(util_bridges) do
    local exists = bridges and bridges[name] ~= nil
    print("  " .. name .. ":", exists and "registered" or "NOT FOUND")
end

-- Check if util_core has JSON methods
print("\n--- util_core JSON methods ---")
if bridges and bridges.util_core then
    local json_methods = {
        "jsonEncode",
        "jsonDecode",
        "jsonExtractStructuredData",
        "jsonValidateJSONSchema",
        "jsonToJSON",
        "jsonPrettify"
    }
    
    for _, method in ipairs(json_methods) do
        local exists = bridges.util_core[method] ~= nil
        print("  " .. method .. ":", exists and "exists" or "missing")
    end
end

return {
    bridges_exist = bridges ~= nil,
    util_core_exists = bridges and bridges.util_core ~= nil,
    util_json_exists = bridges and bridges.util_json ~= nil
}