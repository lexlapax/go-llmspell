-- Test LLM with provider
local bridge = bridges.llm_core

print("Testing generateWithProvider...")

-- Check if generateWithProvider exists
print("generateWithProvider exists:", bridge.generateWithProvider ~= nil)

-- Try to use it
if bridge.generateWithProvider then
    local success, result = pcall(function()
        return bridge.generateWithProvider("openai", "Say hello", {
            model = "gpt-3.5-turbo",
            max_tokens = 10
        })
    end)
    
    if success then
        print("Success! Result:", result)
    else
        print("Error:", result)
    end
else
    print("generateWithProvider not found")
end

return "done"