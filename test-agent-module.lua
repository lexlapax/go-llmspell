-- Test agent module loading
local agent = require("agent")

print("Agent module loaded: " .. type(agent))
if type(agent) == "table" then
    print("agent.create: " .. type(agent.create))
    
    -- Try to create a simple agent
    local test_agent = agent.create("Test", {
        model = "gpt-4"
    })
    
    print("Agent created successfully")
end