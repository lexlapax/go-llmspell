-- Test agent bridge directly
print("Testing agent bridge...")

-- Check if bridges are available
print("bridges:", bridges)
print("type(bridges):", type(bridges))

if bridges then
    print("bridges.agent_core:", bridges.agent_core)
    print("type(bridges.agent_core):", type(bridges.agent_core))
    
    if bridges.agent_core then
        print("\nAvailable methods:")
        for k, v in pairs(bridges.agent_core) do
            print("  " .. k .. ":", type(v))
        end
        
        -- Try to create an agent
        print("\nTrying lifecycleCreate...")
        local success, result = pcall(function()
            return bridges.agent_core.lifecycleCreate("Test", {model = "gpt-3.5-turbo"})
        end)
        print("Result:", success, result)
    end
end

return "done"