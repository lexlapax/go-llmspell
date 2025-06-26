-- ABOUTME: Test llm.complete return value
-- ABOUTME: Debug what the function returns

local llm = require("llm")

print("=== Testing llm.complete ===")

-- Test simple call
local result = llm.complete({
    model = "gpt-3.5-turbo",
    messages = {
        {role = "user", content = "Say hello"}
    }
})

print("Result type:", type(result))
print("Result value:", result)

if type(result) == "table" then
    print("\nTable contents:")
    for k, v in pairs(result) do
        print("  " .. tostring(k) .. " = " .. tostring(v))
    end
end

return {done = true}