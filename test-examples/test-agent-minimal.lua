-- Test agent module
local agent = require("agent")

print("Agent module loaded:", type(agent))
print("agent.create:", type(agent.create))

-- Test without API call (which would fail without key)
print("Agent module test successful")
return "Success"