-- ABOUTME: Basic LLM interaction example showing simple completions
-- ABOUTME: Demonstrates core LLM module usage with different models and parameters

-- Required modules
local llm = require("llm")

-- Basic LLM Interaction Example
-- This spell demonstrates fundamental LLM operations including:
-- 1. Simple text generation
-- 2. Conversation with context
-- 3. Different model usage

-- Validate parameters
local model = params.model or "gpt-3.5-turbo"
local prompt = params.prompt or "Tell me an interesting fact about Lua programming language."

-- Example 1: Simple text generation
print("=== Example 1: Simple Text Generation ===")
local response = llm.generate(prompt, {
    model = model
})

print("Response: " .. tostring(response))
print()

-- Example 2: Message-based generation with system prompt
print("=== Example 2: Message-based Generation ===")
local response2 = llm.generate_message({
    {role = "system", content = "You are a helpful assistant who speaks concisely."},
    {role = "user", content = "What is Lua?"}
}, {
    model = model,
    temperature = 0.7
})

print("Response: " .. tostring(response2))
print()

-- Example 3: Multi-turn conversation
print("=== Example 3: Multi-turn Conversation ===")
local conversation = {
    {role = "system", content = "You are a helpful coding assistant."},
    {role = "user", content = "What's a good use case for Lua?"}
}

local response3 = llm.generate_message(conversation, {
    model = model
})

print("Assistant: " .. tostring(response3))

-- Add assistant response to conversation
table.insert(conversation, {role = "assistant", content = tostring(response3)})
table.insert(conversation, {role = "user", content = "Can you show me a simple example?"})

-- Continue conversation
local response4 = llm.generate_message(conversation, {
    model = model
})

print("\nUser: Can you show me a simple example?")
print("Assistant: " .. tostring(response4))
print()

-- Example 4: Using quick_prompt for simple queries
print("=== Example 4: Quick Prompt ===")
local quick_response = llm.quick_prompt("Write a haiku about programming", {
    model = model,
    temperature = 0.9
})

print("Haiku: " .. tostring(quick_response))
print()

-- Example 5: Error handling
print("=== Example 5: Error Handling ===")
local success, result = pcall(function()
    return llm.generate(nil)  -- This should fail
end)

if success then
    print("Unexpected success")
else
    print("Error caught: " .. tostring(result))
end

-- Return summary
return {
    success = true,
    examples_demonstrated = {
        simple_generation = true,
        message_generation = true,
        conversation = true,
        quick_prompt = true,
        error_handling = true
    }
}