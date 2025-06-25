-- Test module loading
print("Testing module loading order...")

-- Check what state module we get
local state = require("state")
print("\nState module loaded")

-- Check if it has our methods
print("Has get?", state.get ~= nil)
print("Has set?", state.set ~= nil) 
print("Has update?", state.update ~= nil)

-- Check if it has adapter methods
print("Has create?", state.create ~= nil)
print("Has merge?", state.merge ~= nil)

-- Check if bridges.state exists
if bridges and bridges.state then
    print("\nbridges.state exists")
end

-- Check if bridges.state_manager exists
if bridges and bridges.state_manager then
    print("bridges.state_manager exists")
end

-- Check if bridges.state_context exists
if bridges and bridges.state_context then
    print("bridges.state_context exists")
end
