-- Test bridge method call with self
print("Testing bridge method call with self...")

if bridges and bridges.agent_core then
    local bridge = bridges.agent_core
    
    local agent_id = "test_agent_456"
    local config = {
        name = "Test Agent 2",
        model = "gpt-3.5-turbo",
        type = "basic"
    }
    
    -- Method 1: Direct call (works)
    print("Method 1 - Direct call:")
    local result1 = bridge.createAgent(agent_id, config)
    print("  Result:", result1)
    
    -- Method 2: pcall without self (fails)
    print("\nMethod 2 - pcall without self:")
    local success2, result2 = pcall(bridge.createAgent, agent_id .. "_2", config)
    print("  Success:", success2)
    print("  Result:", result2)
    
    -- Method 3: pcall with self (should work)
    print("\nMethod 3 - pcall with self:")
    local success3, result3 = pcall(bridge.createAgent, bridge, agent_id .. "_3", config)
    print("  Success:", success3)
    print("  Result:", result3)
end

return "done"