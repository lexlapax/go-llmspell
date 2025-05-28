-- ABOUTME: Simple chat assistant example that works with current implementation
-- ABOUTME: Shows basic conversation without persistence or streaming

-- Get parameters with defaults
local system_prompt = params.system_prompt or "You are a helpful assistant."

print("Simple Chat Assistant")
print("System: " .. system_prompt)
print("(This is a simplified version - no history or streaming yet)")
print("")

-- Simulate a few exchanges
local exchanges = {
    "What is the capital of France?",
    "Tell me a fun fact about it.",
    "What's the weather like there?"
}

for i, question in ipairs(exchanges) do
    print("You: " .. question)
    
    -- For now, just use simple chat without history
    -- In a full implementation, we'd build the full conversation context
    local full_prompt = system_prompt .. "\n\nUser: " .. question .. "\n\nAssistant:"
    
    local response, err = llm.chat(full_prompt)
    if err then
        print("Error: " .. err)
    else
        print("Assistant: " .. response)
    end
    print("")
end

print("(End of demo - full interactive chat requires io module)")