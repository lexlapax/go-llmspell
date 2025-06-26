-- ABOUTME: Test the exact coordinator:run() call from line 207
-- ABOUTME: Isolate the issue with passing table input to agent:run()

local agent = require("agent")
local state = require("state")

print("=== Testing coordinator:run() pattern ===")

-- Set up state like in the example
state.set("research", {
    topic = "Test Research Topic"
})

-- Create coordinator agent like in example
print("\n1. Creating coordinator agent...")
local coordinator = agent.create("Research Coordinator", {
    model = "gpt-4",
    system = [[
        You coordinate research tasks. 
        Track progress in shared state under 'research.progress'.
        Assign tasks to researcher agents.
    ]]
})

if not coordinator then
    print("ERROR: Failed to create coordinator")
    return {error = "coordinator creation failed"}
end

print("Coordinator created with ID:", coordinator.id)

-- Try the exact same call pattern
print("\n2. Testing coordinator:run() with table input...")
local ok, result = pcall(function()
    return coordinator:run({
        prompt = "Create a research plan for: " .. state.get("research.topic"),
        max_tokens = 200
    })
end)

if ok then
    print("SUCCESS: coordinator:run() worked!")
    print("Result type:", type(result))
    if type(result) == "table" and result.content then
        print("Result content preview:", string.sub(result.content, 1, 50) .. "...")
    end
else
    print("ERROR: coordinator:run() failed")
    print("Error message:", result)
    
    -- Try alternate approach
    print("\n3. Trying with string input instead...")
    ok, result = pcall(function()
        return coordinator:run("Create a research plan for: " .. state.get("research.topic"))
    end)
    
    if ok then
        print("String input worked!")
    else
        print("String input also failed:", result)
    end
end

return {
    coordinator_created = coordinator ~= nil,
    table_input_ok = ok,
    error = not ok and result or nil
}