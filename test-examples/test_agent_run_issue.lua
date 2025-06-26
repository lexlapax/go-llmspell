-- ABOUTME: Test to isolate agent:run() issue
-- ABOUTME: Check what agent.create returns and how :run() works

local agent = require("agent")

print("Testing agent creation and run...")

-- Create a simple agent
local test_agent = agent.create("Test Agent", {
    model = "gpt-3.5-turbo",
    system = "You are a test agent"
})

print("Agent created:", test_agent ~= nil)
print("Agent type:", type(test_agent))

-- Check what's in the agent object
if test_agent then
    print("Agent ID:", test_agent.id)
    print("Agent has run method:", type(test_agent.run))
    
    -- Try to run with a simple string
    print("\nTrying to run agent with string input...")
    local ok, result = pcall(function()
        return test_agent:run("Hello, world!")
    end)
    
    if ok then
        print("Run succeeded!")
        print("Result type:", type(result))
        if type(result) == "table" then
            print("Result content:", result.content)
        else
            print("Result:", result)
        end
    else
        print("Run failed:", result)
    end
end

return {
    agent_created = test_agent ~= nil,
    error = not ok and result or nil
}