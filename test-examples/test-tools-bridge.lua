-- Test tools bridge methods
local tools = require("tools")

print("Testing tools bridge...")

-- Get the bridge directly
local bridge = bridges and bridges.agent_tools

if not bridge then
    print("No tools bridge found")
    return "Failed"
end

print("Bridge found:", type(bridge))
print("\nBridge methods:")

-- List all methods
for k, v in pairs(bridge) do
    print("  " .. k .. ":", type(v))
end

-- Try the correct method name
print("\nTrying to get tool info for 'system_info':")
if bridge.getToolInfo then
    local success, result = pcall(function()
        return bridge.getToolInfo(bridge, "system_info")
    end)
    print("getToolInfo result:", success, result)
end

if bridge.executeTool then
    print("\nTrying executeTool:")
    local success, result = pcall(function()
        return bridge.executeTool(bridge, "system_info", {})
    end)
    print("executeTool result:", success, result)
end

return "Test completed"