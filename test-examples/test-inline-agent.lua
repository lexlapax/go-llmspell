-- Create agent functionality inline to bypass module loading issues
print("Creating inline agent implementation...")

-- Get the bridge directly
if not bridges or not bridges.agent_core then
    error("Agent bridge not available")
end

local bridge = bridges.agent_core

-- Create agent directly
local agent_id = "test_agent_" .. tostring(os.time())
local config = {
    name = "Test Agent",
    model = "gpt-3.5-turbo",
    type = "basic"
}

print("Creating agent with bridge.createAgent...")
print("  agent_id:", agent_id)
print("  config:", config)

local agent_obj = bridge.createAgent(agent_id, config)
print("Agent created:", agent_obj)

if agent_obj then
    print("Agent details:")
    print("  id:", agent_obj.id)
    print("  name:", agent_obj.name)
    print("  type:", agent_obj.type)
end

return "done"