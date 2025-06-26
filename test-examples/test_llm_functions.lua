-- ABOUTME: Test llm function availability
-- ABOUTME: Check which functions exist

local llm = require("llm")

print("=== Testing LLM Functions ===")

-- Check generate_message
print("\nChecking generate_message:")
print("  llm.generate_message type:", type(llm.generate_message))
print("  llm.generateMessage type:", type(llm.generateMessage))
print("  llm.complete type:", type(llm.complete))

-- Check if they're the same
if llm.generate_message then
    print("\nAre they the same?")
    print("  generate_message == generateMessage:", llm.generate_message == llm.generateMessage)
    print("  generate_message == complete:", llm.generate_message == llm.complete)
end

-- Try calling it
if llm.generate_message then
    print("\nTrying to call generate_message...")
    local messages = {
        {role = "system", content = "Reply with just 'test'"},
        {role = "user", content = "test"}
    }
    
    local ok, result = pcall(llm.generate_message, messages, {
        model = "gpt-3.5-turbo",
        max_tokens = 10
    })
    
    if ok then
        print("  Success! Type:", type(result))
        if type(result) == "table" then
            print("  Has content field:", result.content ~= nil)
        end
    else
        print("  Error:", result)
    end
end

-- Check bridge access
print("\nChecking bridge availability:")
print("  bridges type:", type(bridges))
if bridges then
    print("  bridges.llm_core type:", type(bridges.llm_core))
    if bridges.llm_core then
        print("  bridges.llm_core.generateMessage type:", type(bridges.llm_core.generateMessage))
    end
end

return {done = true}