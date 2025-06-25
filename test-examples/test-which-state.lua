-- Check which state module is loaded
local state = require("state")
print("State module loaded")
print("Has get?", state.get ~= nil)
print("Has create?", state.create ~= nil)

-- Check if it is the bridge state or our state
if state.create and not state.get then
    print("This is the bridge state module, not our custom state.lua")
end
