-- Simple agent test
local agent = require("agent")

-- Create agent
local my_agent = agent.create("Test Agent", {
    model = "gpt-4"
})

return "Agent created successfully"