print("Testing state module loading...")
local state = require("state")
print("state type:", type(state))
print("state.get type:", type(state.get))

-- Test basic get
print("\nTesting state.get...")
local result = state.get("test")
print("Got result:", result)