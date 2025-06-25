-- Test what methods are available on the agent bridge
print("Testing agent bridge methods...")

-- Check if bridges are available
if bridges and bridges.agent_core then
    local bridge = bridges.agent_core
    print("\nChecking for expected methods:")
    
    local methods = {
        "lifecycleCreate",
        "createAgent", 
        "run",
        "runAgent",
        "stateGet",
        "getAgentState"
    }
    
    for _, method in ipairs(methods) do
        local m = bridge[method]
        print(string.format("  %-20s: %s", method, type(m)))
    end
    
    -- Try direct table access
    print("\nAll methods in bridge:")
    for k, v in pairs(bridge) do
        if type(v) == "function" then
            print("  " .. k .. " (function)")
        end
    end
else
    print("Agent bridge not available")
end

return "done"