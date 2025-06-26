-- ABOUTME: Debug JSON type conversion issues
-- ABOUTME: Check what types are being passed and returned

print("=== JSON Type Debug Test ===")

-- Check if util_core is available
if not bridges or not bridges.util_core then
    print("ERROR: util_core bridge not available")
    return {success = false}
end

-- Test 1: Simple string encoding
print("\n--- Test 1: Encode Simple String ---")
local str_result, str_err = bridges.util_core.jsonEncode("hello")
print("Input type: string")
print("Result:", str_result)
print("Result type:", type(str_result))
print("Error:", str_err)

-- Test 2: Number encoding
print("\n--- Test 2: Encode Number ---")
local num_result, num_err = bridges.util_core.jsonEncode(42)
print("Input type: number")
print("Result:", num_result)
print("Result type:", type(num_result))
print("Error:", num_err)

-- Test 3: Table encoding
print("\n--- Test 3: Encode Table ---")
local tbl = {name = "test", value = 42}
local tbl_result, tbl_err = bridges.util_core.jsonEncode(tbl)
print("Input type: table")
print("Result:", tbl_result)
print("Result type:", type(tbl_result))
print("Error:", tbl_err)

-- Test 4: Check data module internals
print("\n--- Test 4: Data Module Flow ---")
local data = require("data")

-- Capture internal steps
local function debug_to_json(obj)
    print("Calling data.to_json with:", type(obj))
    
    -- Try calling get_utils
    local util_ok, util = pcall(function()
        if not bridges or not bridges.util_core then
            error("Utils bridge not available")
        end
        return bridges.util_core
    end)
    
    if not util_ok then
        print("Failed to get utils:", util)
        return nil, util
    end
    
    print("Got utils bridge")
    
    -- Try calling jsonEncode
    local encode_ok, result = pcall(util.jsonEncode, obj)
    print("jsonEncode pcall success:", encode_ok)
    print("jsonEncode result:", result)
    print("jsonEncode result type:", type(result))
    
    return result
end

local debug_result = debug_to_json({test = "value"})
print("Debug result:", debug_result)

return {
    string_works = str_result ~= nil and type(str_result) == "string",
    number_works = num_result ~= nil and type(num_result) == "string",
    table_works = tbl_result ~= nil and type(tbl_result) == "string"
}