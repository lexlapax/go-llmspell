-- ABOUTME: Test if agent bridge is available
-- ABOUTME: Check bridge registration and agent creation

print("Checking bridges...")

-- Check if bridges exist
if bridges then
    print("bridges table exists")
    
    -- List all bridges
    print("\nAvailable bridges:")
    for name, bridge in pairs(bridges) do
        print("  - " .. name)
    end
    
    -- Check agent_core specifically
    if bridges.agent_core then
        print("\nagent_core bridge is available")
        print("Bridge type:", type(bridges.agent_core))
        
        -- Check bridge methods
        if bridges.agent_core.lifecycleCreate then
            print("lifecycleCreate method exists")
        end
    else
        print("\nERROR: agent_core bridge NOT available")
    end
else
    print("ERROR: bridges table does not exist")
end

-- Now try agent module
local agent = require("agent")
print("\nagent module loaded")

-- Try creating an agent
print("\nAttempting to create agent...")
local ok, result = pcall(function()
    return agent.create("Test", {model = "gpt-3.5-turbo"})
end)

if ok then
    print("Agent created successfully")
    print("Agent ID:", result and result.id or "nil")
else
    print("Agent creation failed:", result)
end

return {
    bridges_exist = bridges ~= nil,
    agent_core_exists = bridges and bridges.agent_core ~= nil,
    agent_created = ok
}