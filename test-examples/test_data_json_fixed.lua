-- ABOUTME: Test data.to_json with proper error handling  
-- ABOUTME: Tests the specific issue found in 09-state-management.lua at line 153

print("=== Data JSON Fixed Test ===")

local state = require("state")
local data = require("data")

-- Set up simple state
state.set("test", {
    name = "test", 
    value = 42,
    nested = {
        key = "value"
    }
})

-- Get the state
local test_state = state.get("test")
print("Retrieved state type:", type(test_state))

-- Try to convert to JSON with error handling
local ok, result = pcall(data.to_json, test_state)
if ok then
    print("JSON conversion successful")
    print("Result type:", type(result))
    if type(result) == "string" then
        print("JSON length:", #result)
        print("JSON preview:", string.sub(result, 1, 50) .. "...")
    else
        print("Unexpected result type, got:", result)
    end
else
    print("JSON conversion failed:", result)
end

-- Test the specific line that was failing (line 153 of 09-state-management.lua)
if ok and type(result) == "string" then
    print("Testing line 153 equivalent...")
    local size_message = "State size: " .. #result .. " bytes"
    print(size_message)
    print("✓ Line 153 equivalent works!")
else
    print("✗ Line 153 equivalent would fail")
end

return {
    success = ok,
    result_type = type(result),
    length_test = ok and type(result) == "string"
}