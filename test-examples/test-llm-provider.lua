-- Test LLM provider setup
local llm = require("llm")

print("Testing LLM provider setup...")

-- List available providers
print("\nAvailable providers:")
local providers = llm.list_providers()
for _, provider in ipairs(providers) do
    print("  -", provider)
end

-- Try to set provider
print("\nSetting provider to openai...")
local success, err = pcall(function()
    llm.use_provider("openai")
end)

if success then
    print("Provider set successfully")
    print("Current provider:", llm.get_current_provider())
else
    print("Error setting provider:", err)
end

-- Try a simple prompt
print("\nTesting quick_prompt...")
local result = llm.quick_prompt("Say hello", {
    model = "gpt-3.5-turbo",
    max_tokens = 10
})
print("Response:", result)

return "done"