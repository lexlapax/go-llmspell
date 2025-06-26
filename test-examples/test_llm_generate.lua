-- ABOUTME: Test llm.generateMessage to see what it returns
-- ABOUTME: Check the response structure

local llm = require("llm")

print("Testing llm.generateMessage...")

local ok, result = pcall(function()
    return llm.generateMessage({
        {role = "system", content = "You are a helpful assistant"},
        {role = "user", content = "Say hello"}
    }, {
        model = "gpt-3.5-turbo",
        max_tokens = 50
    })
end)

if ok then
    print("Success!")
    print("Result type:", type(result))
    if type(result) == "table" then
        print("Result keys:")
        for k, v in pairs(result) do
            print("  -", k, ":", type(v))
            if k == "content" and type(v) == "string" then
                print("    Content preview:", string.sub(v, 1, 50))
            end
        end
    else
        print("Result:", result)
    end
else
    print("Error:", result)
end

return {
    success = ok,
    has_content = ok and result and result.content ~= nil
}