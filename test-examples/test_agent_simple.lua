-- Simple agent test
print("=== Testing Agent Module ===")

local success, agent = pcall(require, "agent")
if not success then
    print("Failed to load agent module: " .. tostring(agent))
    return {success = false, error = agent}
end

print("Agent module loaded successfully")

-- Test agent creation
print("\n--- Testing Agent Creation ---")
local create_success, assistant = pcall(function()
    return agent.create("TestAgent", {
        model = "gpt-3.5-turbo",
        system = "You are a test assistant."
    })
end)

if not create_success then
    print("Agent creation failed: " .. tostring(assistant))
    return {success = false, error = assistant}
end

print("Agent created successfully")
print("Agent type: " .. tostring(type(assistant)))

if assistant and assistant.id then
    print("Agent ID: " .. tostring(assistant.id))
else
    print("WARNING: Agent object missing ID")
end

-- Test object-oriented syntax
print("\n--- Testing Object-Oriented Syntax ---")
if assistant and type(assistant.run) == "function" then
    print("assistant:run method available")
    
    local run_success, response = pcall(function()
        return assistant:run("Hello, can you respond?")
    end)
    
    if run_success then
        print("Agent run succeeded")
        print("Response type: " .. tostring(type(response)))
        if response and response.content then
            print("Response: " .. tostring(response.content))
        end
    else
        print("Agent run failed: " .. tostring(response))
    end
else
    print("ERROR: assistant:run method not available")
end

print("\n=== Agent Test Complete ===")
return {
    success = true,
    agent_created = assistant ~= nil,
    has_id = assistant and assistant.id ~= nil,
    has_run_method = assistant and type(assistant.run) == "function"
}