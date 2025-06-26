-- ABOUTME: Test if CheckTable is causing JSON issues
-- ABOUTME: Verify that jsonEncode only accepts tables

print("=== JSON Table Check Test ===")

-- Test 1: Table input (should work)
local table_result = nil
local table_err = nil
if bridges and bridges.util_core and bridges.util_core.jsonEncode then
    local ok, r1, r2 = pcall(function()
        return bridges.util_core.jsonEncode({test = "table"})
    end)
    if ok then
        table_result = r1
        table_err = r2
    else
        table_err = r1  -- Error from pcall
    end
end

-- Test 2: String input (might fail due to CheckTable)
local string_result = nil
local string_err = nil
if bridges and bridges.util_core and bridges.util_core.jsonEncode then
    local ok, r1, r2 = pcall(function()
        return bridges.util_core.jsonEncode("just a string")
    end)
    if ok then
        string_result = r1
        string_err = r2
    else
        string_err = r1  -- Error from pcall
    end
end

-- Test 3: Number input
local number_result = nil
local number_err = nil
if bridges and bridges.util_core and bridges.util_core.jsonEncode then
    local ok, r1, r2 = pcall(function()
        return bridges.util_core.jsonEncode(42)
    end)
    if ok then
        number_result = r1
        number_err = r2
    else
        number_err = r1  -- Error from pcall
    end
end

return {
    table_works = table_result ~= nil,
    table_error = tostring(table_err or "none"),
    string_works = string_result ~= nil,
    string_error = tostring(string_err or "none"),
    number_works = number_result ~= nil,
    number_error = tostring(number_err or "none")
}