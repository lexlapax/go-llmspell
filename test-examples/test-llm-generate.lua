-- Test LLM generate method
local llm = require("llm")

print("Testing LLM generate...")

-- Set provider
llm.use_provider("openai")

-- Get bridge directly
local bridge = bridges.llm_core

print("\nTesting direct bridge.generate...")
local success, result = pcall(function()
    return bridge.generate("Say hello", {
        model = "gpt-3.5-turbo",
        max_tokens = 10
    })
end)

if success then
    print("Success! Result:", result)
else
    print("Error:", result)
end

print("\nTesting llm.quick_prompt...")
local success2, result2 = pcall(function()
    return llm.quick_prompt("Say hello", {
        model = "gpt-3.5-turbo", 
        max_tokens = 10
    })
end)

if success2 then
    print("Success! Result:", result2)
else
    print("Error:", result2)
end

return "done"