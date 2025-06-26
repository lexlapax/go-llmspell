-- ABOUTME: Test state and data module interaction
-- ABOUTME: Specifically test state.get() and data.to_json() operations

print("=== State JSON Test ===")

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
print("Retrieved state:", type(test_state))

-- Try to convert to JSON
local ok, result = pcall(data.to_json, test_state)
if ok then
    print("JSON conversion successful")
    print("Result type:", type(result))
    if type(result) == "string" then
        print("JSON length:", #result)
        print("JSON content:", result)
    else
        print("JSON result:", result)
    end
else
    print("JSON conversion failed:", result)
end

return {
    success = ok,
    result_type = type(result),
    error = not ok and result or nil
}