-- Test state module functions
local state = require("state")
print("State module functions:")
for k, v in pairs(state) do
    print("  " .. k .. ":", type(v))
end
