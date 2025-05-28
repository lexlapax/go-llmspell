-- ABOUTME: Compares responses from multiple LLM providers
-- ABOUTME: Useful for evaluating different models on the same prompt

-- Get parameters
local prompt = params.prompt
local providers_param = params.providers or "openai,anthropic,gemini"

if not prompt then
    error("Prompt parameter is required")
end

-- Parse providers list (handle both string and array format)
local providers_to_test = {}
if type(providers_param) == "string" then
    for provider in string.gmatch(providers_param, "[^,]+") do
        table.insert(providers_to_test, provider:match("^%s*(.-)%s*$")) -- trim whitespace
    end
elseif type(providers_param) == "table" then
    providers_to_test = providers_param
else
    error("Providers parameter must be a string or array")
end

-- Get current provider to restore later
local original_provider = llm.get_provider()

-- Results table
local results = {}

print("Provider Comparison")
print("==================")
print("Prompt: " .. prompt)
print("\nTesting providers: " .. table.concat(providers_to_test, ", "))
print("\n")

-- Test each provider
for _, provider in ipairs(providers_to_test) do
    print("Testing " .. provider .. "...")
    
    -- Try to switch to provider
    local switch_err = llm.set_provider(provider)
    
    if switch_err then
        print("  Error: Cannot switch to " .. provider .. " - " .. switch_err)
        results[provider] = {
            success = false,
            error = switch_err
        }
    else
        -- Send the prompt
        local response, chat_err = llm.chat(prompt)
        
        if chat_err then
            print("  Error: " .. chat_err)
            results[provider] = {
                success = false,
                error = chat_err
            }
        else
            print("  Success!")
            results[provider] = {
                success = true,
                response = response
            }
        end
    end
    
    print("")
end

-- Restore original provider
if original_provider then
    llm.set_provider(original_provider)
end

-- Display results
print("\n" .. string.rep("=", 80))
print("RESULTS")
print(string.rep("=", 80))

for _, provider in ipairs(providers_to_test) do
    local result = results[provider]
    
    print("\n" .. string.upper(provider))
    print(string.rep("-", #provider))
    
    if result.success then
        print("Status: Success")
        print("Response:")
        print(result.response)
    else
        print("Status: Failed")
        print("Error: " .. result.error)
    end
end

-- Summary
print("\n" .. string.rep("=", 80))
print("SUMMARY")
print(string.rep("=", 80))

local successful = 0

for provider, result in pairs(results) do
    if result.success then
        successful = successful + 1
    end
end

print("Providers tested: " .. #providers_to_test)
print("Successful: " .. successful)
print("Failed: " .. (#providers_to_test - successful))

print("\n✅ Comparison complete!")