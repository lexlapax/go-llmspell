-- Test if built-in tools are accessible
local tools = require("tools")

print("Testing built-in tools access...")

-- Try to list available tools
print("\nTrying tools.list():")
local success, result = pcall(function()
    return tools.list()
end)

if success then
    print("Success! Found " .. #result .. " tools")
    -- Show first few tools
    for i = 1, math.min(5, #result) do
        print("  - " .. tostring(result[i].name or result[i]))
    end
else
    print("Error:", result)
end

-- Try to execute a simple tool directly
print("\nTrying to execute system_info tool:")
success, result = pcall(function()
    return tools.execute_safe("system_info", {})
end)

if success then
    print("Success!")
    if result then
        print("  OS:", result.os and result.os.name or "unknown")
    end
else
    print("Error:", result)
end

return "Test completed"