-- ABOUTME: Test llm response structure in detail
-- ABOUTME: Print all fields to understand format

local llm = require("llm")

print("=== Testing LLM Response Structure ===")

-- Test generateMessage which is what the example uses
local messages = {
    {role = "system", content = "Reply with just 'technical'"},
    {role = "user", content = "classify this: my computer crashed"}
}

print("\nCalling llm.generateMessage...")
local response = llm.generateMessage(messages, {
    model = "gpt-3.5-turbo",
    max_tokens = 10
})

print("\nResponse details:")
print("Type:", type(response))

if type(response) == "table" then
    print("\nAll fields:")
    for k, v in pairs(response) do
        print("  " .. k .. ":", type(v), "=", tostring(v))
    end
    
    -- Check common field names
    print("\nChecking common fields:")
    print("  response.content:", response.content)
    print("  response.message:", response.message)
    print("  response.text:", response.text)
    print("  response.choices:", response.choices)
    print("  response.result:", response.result)
    print("  response[1]:", response[1])
elseif type(response) == "string" then
    print("Response is a string:", response)
else
    print("Unexpected response type")
end

-- Also test llm.generate for comparison
print("\n\nTesting llm.generate...")
local response2 = llm.generate("Reply with just 'hello'", {
    model = "gpt-3.5-turbo", 
    max_tokens = 10
})

print("Type:", type(response2))
if type(response2) == "string" then
    print("String response:", response2)
elseif type(response2) == "table" then
    print("Table response with keys:")
    for k, v in pairs(response2) do
        print("  " .. k .. ":", type(v))
    end
end

return {done = true}