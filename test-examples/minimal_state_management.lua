-- ABOUTME: Minimal version of state management example
-- ABOUTME: Isolate the exact cause of the error

local state = require("state")
local agent = require("agent")
local data = require("data")
local utils = require("utils")

print("=== Minimal State Management Test ===")

-- Test 1: Basic state operations
print("\n1. Testing basic state operations...")
state.set("test", {value = 42})
print("Set test.value = 42")
print("Get test.value =", state.get("test.value"))

-- Test 2: JSON operations
print("\n2. Testing JSON operations...")
local test_data = {name = "test", items = {1, 2, 3}}
local json_str = data.to_json(test_data)
print("to_json result:", json_str ~= nil)

local parsed = data.from_json(json_str)
print("from_json result:", parsed ~= nil)

-- Test 3: Agent creation and run
print("\n3. Testing agent creation...")
local assistant = agent.create("Assistant", {
    model = "gpt-3.5-turbo",
    system = "You are a test assistant."
})
print("Agent created:", assistant ~= nil)

print("\n4. Testing agent:run()...")
local response = assistant:run("Hello")
print("Response received:", response ~= nil)

-- Test 4: State watcher
print("\n5. Testing state watcher...")
local function create_watcher(path, callback)
    local last_value = state.get(path)
    
    return function()
        local current_value = state.get(path)
        if current_value ~= last_value then
            callback(current_value, last_value)
            last_value = current_value
        end
    end
end

state.set("counter", 0)
local watcher = create_watcher("counter", function(new_val, old_val)
    print("Counter changed from", old_val, "to", new_val)
end)

state.set("counter", 1)
watcher()

-- Test 5: Sleep function
print("\n6. Testing sleep...")
utils.general_sleep(100)
print("Sleep completed")

print("\n=== All tests completed ===")

return {
    success = true
}