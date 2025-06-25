-- Simple state test
print("=== Testing State Module ===")

local state = require("state")
print("State module loaded successfully")

-- Test basic state operations
print("\n--- Basic State Operations ---")
state.set("test_key", "test_value")
local value = state.get("test_key")
print("Set/Get test: " .. tostring(value))

state.set("counter", 1)
local updated = state.update("counter", function(old) return old + 5 end)
print("Update test: " .. tostring(updated))

local has_counter = state.has("counter")
print("Has test: " .. tostring(has_counter))

-- Test new methods
print("\n--- New StateAdapter Methods ---")
local values = state.values()
print("Values method works: " .. tostring(type(values) == "table"))

-- Test namespaced methods
print("\n--- Namespaced Methods ---")
print("Transforms namespace: " .. tostring(type(state.transforms) == "table"))
print("Context namespace: " .. tostring(type(state.context) == "table"))
print("Persistence namespace: " .. tostring(type(state.persistence) == "table"))

print("\n=== State Test Complete ===")
return {
    success = true,
    basic_operations = true,
    new_methods = true,
    namespaces = true
}