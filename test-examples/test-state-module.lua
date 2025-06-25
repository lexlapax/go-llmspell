-- Test state module
print("Loading state module...")
local state = require("state")
print("State module loaded:", state)
print("state.get type:", type(state.get))

-- Try to use state.get
print("Testing state.get...")
local result = state.get("test")
print("Result:", result)

return "done"
