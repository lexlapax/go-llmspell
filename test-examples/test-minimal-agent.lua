-- Test minimal agent
package.path = "/tmp/?.lua;" .. package.path

local agent = require("minimal_agent")
print("Loaded minimal agent")

local my_agent = agent.create("Test", {model = "test"})
print("Created agent:", my_agent)
print("Agent name:", my_agent.name)

local result = my_agent:run("Hello")
print("Run result:", result)

return "done"