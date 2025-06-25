print("Testing agent bridge...")

-- Check if bridges exist
print("bridges type:", type(bridges))
if bridges then
    print("bridges.agent_core type:", type(bridges.agent_core))
    
    if bridges.agent_core then
        -- List all methods on agent_core
        print("\nMethods on bridges.agent_core:")
        for k, v in pairs(bridges.agent_core) do
            print("  " .. k .. ":", type(v))
        end
    end
end

-- Try loading agent module
local agent = require("agent")
print("\nagent module loaded:", type(agent))

-- Try creating a simple agent
print("\nTrying to create agent...")
local test_agent = agent.create("TestAgent", {
    model = "gpt-3.5-turbo",
    system = "You are a test agent"
})
print("Agent created:", test_agent and test_agent.id or "FAILED")