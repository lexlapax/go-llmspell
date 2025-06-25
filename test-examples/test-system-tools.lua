-- Find system tools
local tools = require("tools")

print("Finding system tools...")

local all_tools = tools.list()
print("\nAll available tools (" .. #all_tools .. " total):")

-- Find system-related tools
local system_tools = {}
for _, tool in ipairs(all_tools) do
    local name = tool.name or tostring(tool)
    if string.find(name, "system") or string.find(name, "exec") or string.find(name, "env") then
        table.insert(system_tools, name)
    end
end

print("\nSystem-related tools:")
for _, name in ipairs(system_tools) do
    print("  - " .. name)
end

-- Try execute_command instead
print("\nTrying execute_command tool:")
local success, result = pcall(function()
    return tools.execute_safe("execute_command", {
        command = "echo 'Hello from system tool'"
    })
end)

if success and result then
    print("Success!")
    print("  stdout:", result.stdout)
    print("  success:", result.success)
else
    print("Error:", result or "unknown")
end

return "Test completed"