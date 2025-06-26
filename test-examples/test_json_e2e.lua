-- ABOUTME: End-to-end test of JSON functionality through data module
-- ABOUTME: Tests that data.to_json() works with fixed UtilsAdapter

print("=== End-to-End JSON Test ===")

-- Test data module
local data = require("data")

-- Test 1: Simple object to JSON
print("\n--- Test 1: Object to JSON ---")
local obj = {
    name = "test",
    value = 42,
    nested = {
        key = "value"
    }
}

local json_str = data.to_json(obj)
print("Object converted to JSON")
print("Type:", type(json_str))
print("Length:", json_str and #json_str or "nil")
print("Content:", json_str)

-- Test 2: JSON back to object
print("\n--- Test 2: JSON to Object ---")
if json_str then
    local parsed = data.parse_json(json_str)
    print("JSON parsed back to object")
    print("Type:", type(parsed))
    if parsed then
        print("Name:", parsed.name)
        print("Value:", parsed.value)
        print("Nested key:", parsed.nested and parsed.nested.key)
    end
end

-- Test 3: Pretty print
print("\n--- Test 3: Pretty JSON ---")
local pretty_json = data.to_json(obj, "pretty")
print("Pretty JSON:")
print(pretty_json)

-- Test 4: State module integration
print("\n--- Test 4: State + JSON Integration ---")
local state = require("state")

state.set("test_data", obj)
local state_obj = state.get("test_data")
local state_json = data.to_json(state_obj)
print("State object as JSON:", state_json)

print("\n=== All JSON Tests Complete ===")

return {
    success = true,
    json_length = json_str and #json_str or 0,
    pretty_length = pretty_json and #pretty_json or 0,
    state_json_length = state_json and #state_json or 0
}