-- Test agent creation directly
local agent = require("agent")

print("Testing direct agent creation...")

-- Try without pcall first
local name = "Test Agent"
local config = {
    model = "gpt-3.5-turbo",
    system = "You are a helpful assistant.",
    temperature = 0.7
}

print("Calling agent.create...")
local result = agent.create(name, config)

print("Result:", result)
if result then
    print("Type:", type(result))
    if type(result) == "table" then
        for k, v in pairs(result) do
            print("  " .. k .. ":", v)
        end
    end
end

return "done"