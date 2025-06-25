-- Test state.update
local state = require("state")

print("Testing state.update...")

-- Set initial value
state.set("counter", 0)
print("Initial counter:", state.get("counter"))

-- Try update
print("\nUpdating counter...")
local success, result = pcall(function()
    return state.update("counter", function(current)
        return (current or 0) + 1
    end)
end)

print("Update result:", success, result)
print("Counter after update:", state.get("counter"))

return "done"