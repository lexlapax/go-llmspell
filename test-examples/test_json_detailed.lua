-- ABOUTME: Detailed test of JSON functionality with error checking
-- ABOUTME: Tests each step of JSON encoding/decoding process

print("=== Detailed JSON Test ===")

-- Load data module
local ok, data = pcall(require, "data")
if not ok then
    print("ERROR: Failed to load data module:", data)
    return {success = false, error = "Failed to load data module"}
end

-- Test 1: Basic encoding
print("\n--- Test 1: Basic Encoding ---")
local test_obj = {name = "test", value = 42}
print("Input object:", type(test_obj))
for k, v in pairs(test_obj) do
    print("  " .. k .. ":", v)
end

local encode_ok, json_result = pcall(data.to_json, test_obj)
print("Encode call success:", encode_ok)
if encode_ok then
    print("Result type:", type(json_result))
    print("Result:", json_result)
    print("Length:", json_result and #json_result or "nil")
else
    print("Encode error:", json_result)
end

-- Test 2: Direct util_core access
print("\n--- Test 2: Direct util_core Access ---")
if bridges and bridges.util_core and bridges.util_core.jsonEncode then
    print("Calling jsonEncode directly...")
    local direct_ok, direct_result = pcall(bridges.util_core.jsonEncode, test_obj)
    print("Direct call success:", direct_ok)
    print("Direct result type:", type(direct_result))
    print("Direct result:", direct_result)
end

-- Test 3: Check for nil/empty returns
print("\n--- Test 3: Return Value Analysis ---")
if encode_ok and json_result then
    if type(json_result) == "string" then
        if #json_result == 0 then
            print("WARNING: JSON encode returned empty string")
        else
            print("JSON encode returned non-empty string of length", #json_result)
        end
    else
        print("WARNING: JSON encode returned non-string type:", type(json_result))
    end
end

return {
    success = encode_ok,
    result_type = encode_ok and type(json_result) or "error",
    result_length = encode_ok and json_result and type(json_result) == "string" and #json_result or 0
}