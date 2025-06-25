-- Find file tools
local tools = require("tools")

print("Finding file tools...")

local all_tools = tools.list()

-- Find file-related tools
local file_tools = {}
for _, tool in ipairs(all_tools) do
    local name = tool.name or tostring(tool)
    if string.find(name, "file") then
        table.insert(file_tools, name)
    end
end

print("\nFile-related tools:")
for _, name in ipairs(file_tools) do
    print("  - " .. name)
end

return "Test completed"