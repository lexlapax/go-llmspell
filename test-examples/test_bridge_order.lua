-- ABOUTME: Test which order bridges are registered in production
-- ABOUTME: Confirm if util_core is registered before util_json

-- List bridges in the order they appear (Lua preserves insertion order in some cases)
local bridge_names = {}
local bridge_count = 0

if bridges then
    for name, _ in pairs(bridges) do
        bridge_count = bridge_count + 1
        table.insert(bridge_names, name)
    end
end

-- Try to use JSON
local json_test = "not_tested"
local json_error = nil
if bridges and bridges.util_core and bridges.util_core.jsonEncode then
    local result, err = bridges.util_core.jsonEncode({test = true})
    if err then
        json_error = err
        json_test = "failed"
    else
        json_test = "success"
    end
end

return {
    bridge_count = bridge_count,
    first_10_bridges = table.concat({unpack(bridge_names, 1, 10)}, ", "),
    has_util_core = bridges and bridges.util_core ~= nil,
    has_util_json = bridges and bridges.util_json ~= nil,
    json_test = json_test,
    json_error = json_error
}