-- ABOUTME: Test adapter creation order and bridge availability
-- ABOUTME: Debug why jsonBridge is nil in UtilsAdapter

-- Check what bridges are available
print("=== Bridge and Adapter Creation Order Test ===")

-- List all bridges
local bridge_list = {}
if bridges then
    for name, _ in pairs(bridges) do
        table.insert(bridge_list, name)
    end
    table.sort(bridge_list)
end

-- Check if util_core has JSON methods (from UtilsAdapter)
local has_json_methods = false
if bridges and bridges.util_core then
    has_json_methods = bridges.util_core.jsonEncode ~= nil
end

-- Try to understand the problem
local json_error = nil
if has_json_methods then
    local test_obj = {test = "value"}
    local result, err = bridges.util_core.jsonEncode(test_obj)
    if err then
        json_error = err
    end
end

return {
    bridge_count = #bridge_list,
    bridges = table.concat(bridge_list, ", "),
    util_core_exists = bridges and bridges.util_core ~= nil,
    util_json_exists = bridges and bridges.util_json ~= nil,
    has_json_methods = has_json_methods,
    json_error = json_error
}