-- Test bridge method call directly
print("Testing bridge method call...")

if bridges and bridges.agent_core then
    local bridge = bridges.agent_core
    
    print("Bridge type:", type(bridge))
    print("createAgent type:", type(bridge.createAgent))
    
    if type(bridge.createAgent) == "function" then
        print("\nCalling createAgent...")
        local agent_id = "test_agent_123"
        local config = {
            name = "Test Agent",
            model = "gpt-3.5-turbo",
            type = "basic"
        }
        
        print("Parameters:")
        print("  agent_id:", agent_id)
        print("  config:", config)
        
        local success, result = pcall(bridge.createAgent, agent_id, config)
        print("\nCall result:")
        print("  success:", success)
        print("  result:", result)
    end
end

return "done"