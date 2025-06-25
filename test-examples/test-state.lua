-- Test state module
local state = require("state")
print("State module loaded:", state ~= nil)
if state then
    print("State functions:")
    for k, v in pairs(state) do
        print("  ", k, type(v))
    end
end
return "done"