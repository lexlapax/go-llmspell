-- Test state API
local state = require("state")

print("Testing state API...")

-- Try creating a state
local s = state.create()
print("Created state:", s)
print("State type:", type(s))

-- Check if it has get/set methods
if type(s) == "table" then
    print("State methods:")
    for k, v in pairs(s) do
        print("  ", k, type(v))
    end
end

-- Try the module-level functions
print("\nTrying state.to_table:")
local t = state.to_table()
print("Result:", t, type(t))

return "done"