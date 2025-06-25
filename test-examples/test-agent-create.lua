-- Test agent creation
local agent = require("agent")

print("Testing agent creation...")

-- Try creating agent
local success, result = pcall(function()
    return agent.create("Test Agent", {
        model = "gpt-3.5-turbo",
        system = "You are a helpful assistant.",
        temperature = 0.7
    })
end)

print("Create result:", success)
if not success then
    print("Error:", result)
else
    print("Agent created:", result)
    print("Agent type:", type(result))
    
    -- Check if it has run method
    if result and result.run then
        print("Agent has run method")
    end
end

return "done"