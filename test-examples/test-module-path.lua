-- Check module loading
print("Before require:")
print("  package.path:", package.path)

-- Try to find where state module comes from
local function find_module(name)
    for path in package.path:gmatch("[^;]+") do
        local file = path:gsub("?", name)
        print("  Checking:", file)
    end
end

print("
Looking for state module:")
find_module("state")

local state = require("state")
print("
State loaded from somewhere")
print("First few functions:")
local count = 0
for k, v in pairs(state) do
    print("  " .. k)
    count = count + 1
    if count > 5 then break end
end
