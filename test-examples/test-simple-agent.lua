-- Simplest possible test
local agent = require("agent")
print("Agent module:", agent)
print("create function:", agent.create)

-- Direct call
agent.create("Test", {model = "gpt-3.5-turbo"})

return "done"