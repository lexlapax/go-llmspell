-- ABOUTME: Compares responses from multiple LLM providers
-- ABOUTME: Useful for evaluating different models on the same prompt

-- Get parameters
local prompt = params.prompt
local providers_to_test = params.providers or {"openai", "anthropic", "gemini"}

if not prompt then
    error("Prompt parameter is required")
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
        local start_time = os.clock()
        local response, chat_err = llm.chat(prompt)
        local elapsed = os.clock() - start_time
        
        if chat_err then
            print("  Error: " .. chat_err)
            results[provider] = {
                success = false,
                error = chat_err
            }
        else
            print("  Success! Response time: " .. string.format("%.2f", elapsed) .. "s")
            results[provider] = {
                success = true,
                response = response,
                time = elapsed
            }
        end
    end
    
    print("")
end

-- Restore original provider
llm.set_provider(original_provider)

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
        print("Time: " .. string.format("%.2f", result.time) .. " seconds")
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
local total_time = 0

for _, result in pairs(results) do
    if result.success then
        successful = successful + 1
        total_time = total_time + result.time
    end
end

print("Providers tested: " .. #providers_to_test)
print("Successful: " .. successful)
print("Failed: " .. (#providers_to_test - successful))

if successful > 0 then
    print("Average response time: " .. string.format("%.2f", total_time / successful) .. " seconds")
end