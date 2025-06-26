-- ABOUTME: Test llm.generate_message directly
-- ABOUTME: Minimal test for the function

local llm = require("llm")

print("=== Testing llm.generate_message ===")
print("Type of llm:", type(llm))
print("Type of llm.generate_message:", type(llm.generate_message))
print("Type of llm.generateMessage:", type(llm.generateMessage))

-- Test with the exact pattern from 13-agent-handoff.lua
local classification = llm.generate_message({
    {role = "system", content = [[
        Classify the customer query into one of these categories:
        - general: General questions, greetings, account info
        - technical: Software bugs, hardware issues, technical problems
        - billing: Payments, subscriptions, invoices, refunds
        
        Respond with only the category name.
    ]]},
    {role = "user", content = "My computer is crashing"}
}, {
    model = "gpt-3.5-turbo",
    max_tokens = 10
})

print("\nResult type:", type(classification))
if type(classification) == "string" then
    print("String result:", classification)
elseif type(classification) == "table" then
    print("Table result with content:", classification.content)
else
    print("Unexpected type")
end

return {success = true}