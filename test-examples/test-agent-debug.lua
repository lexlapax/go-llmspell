-- Test agent creation with debugging
print("Loading agent module...")
local success, agent = pcall(require, "agent")
if not success then
    print("Failed to load agent module:", agent)
    return "failed"
end

print("Agent module loaded:", type(agent))

-- Check if create function exists
print("agent.create type:", type(agent.create))

if type(agent.create) ~= "function" then
    print("ERROR: agent.create is not a function!")
    return "failed"
end

-- Try to create agent
print("\nCreating agent...")
local name = "Test Agent"
local config = {
    model = "gpt-3.5-turbo",
    system = "You are a helpful assistant.",
    temperature = 0.7
}

-- Use pcall to catch error
local ok, result = pcall(agent.create, name, config)
print("Create result:", ok)
if not ok then
    print("Error:", result)
else
    print("Agent created:", result)
end

return "done"