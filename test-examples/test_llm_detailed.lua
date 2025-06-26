-- ABOUTME: Detailed test of llm module functions
-- ABOUTME: Compare generate vs generateMessage

local llm = require("llm")
local data = require("data")

print("=== Testing LLM functions ===")

-- Test 1: llm.generate with string prompt
print("\n1. Testing llm.generate with string prompt...")
local ok1, result1 = pcall(function()
    return llm.generate("Say hello", {
        model = "gpt-3.5-turbo",
        max_tokens = 50
    })
end)

if ok1 then
    print("generate() succeeded")
    print("Result type:", type(result1))
    if type(result1) == "string" then
        print("Result:", result1)
    else
        print("Result:", data.to_json(result1))
    end
else
    print("generate() failed:", result1)
end

-- Test 2: llm.generateMessage with messages
print("\n2. Testing llm.generateMessage with messages...")
local ok2, result2 = pcall(function()
    return llm.generateMessage({
        {role = "system", content = "You are a helpful assistant"},
        {role = "user", content = "Say hello"}
    }, {
        model = "gpt-3.5-turbo",
        max_tokens = 50
    })
end)

if ok2 then
    print("generateMessage() succeeded")
    print("Result type:", type(result2))
    if type(result2) == "string" then
        print("Result:", result2)
    else
        print("Result:", data.to_json(result2))
    end
else
    print("generateMessage() failed:", result2)
end

-- Test 3: llm.quick_prompt
print("\n3. Testing llm.quick_prompt...")
local ok3, result3 = pcall(function()
    return llm.quick_prompt("Say hello", {
        model = "gpt-3.5-turbo",
        max_tokens = 50
    })
end)

if ok3 then
    print("quick_prompt() succeeded")
    print("Result type:", type(result3))
    if type(result3) == "string" then
        print("Result:", result3)
    else
        print("Result:", data.to_json(result3))
    end
else
    print("quick_prompt() failed:", result3)
end

return {
    generate_ok = ok1,
    generateMessage_ok = ok2,
    quick_prompt_ok = ok3
}