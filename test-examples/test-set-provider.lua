-- Test setting provider in bridge
local bridge = bridges.llm_core

print("Testing provider setup...")

-- Try setProvider
print("\n1. Setting provider with setProvider...")
local success1, result1 = pcall(function()
    return bridge.setProvider("openai", {
        api_key = os.getenv("OPENAI_API_KEY")
    })
end)
print("setProvider result:", success1, result1)

-- Try getProvider
print("\n2. Getting current provider...")
local success2, result2 = pcall(function()
    return bridge.getProvider()
end)
print("getProvider result:", success2, result2)

-- Now try generate
print("\n3. Trying generate after setting provider...")
local success3, result3 = pcall(function()
    return bridge.generate("Say hello", {
        model = "gpt-3.5-turbo",
        max_tokens = 10
    })
end)

if success3 then
    print("Generate success! Result:", result3)
else
    print("Generate error:", result3)
end

return "done"