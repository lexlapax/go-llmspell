-- ABOUTME: Compares responses from multiple LLM providers in parallel
-- ABOUTME: Uses async callbacks to run all provider queries concurrently

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

print("Provider Comparison (Parallel)")
print("==============================")
print("Prompt: " .. prompt)
print("\nTesting providers: " .. table.concat(providers_to_test, ", "))

-- Results table
local results = {}
local completed = 0
local total = #providers_to_test

-- Check which providers are actually available
local available_providers = llm.list_providers()
print("\nAvailable providers: " .. table.concat(available_providers, ", "))

-- Helper function to test a single provider
function test_provider(provider_name, callback)
    -- Check if provider is available
    local is_available = false
    for _, p in ipairs(available_providers) do
        if p == provider_name then
            is_available = true
            break
        end
    end
    
    if not is_available then
        callback(nil, "provider '" .. provider_name .. "' not initialized (check API key)")
        return
    end
    
    -- Switch to provider
    local switch_err = llm.set_provider(provider_name)
    
    if switch_err then
        callback(nil, switch_err)
        return
    end
    
    -- Make synchronous call since we can't guarantee provider state in async
    local response, chat_err = llm.chat(prompt)
    
    if chat_err then
        callback(nil, chat_err)
    else
        callback(response, nil)
    end
end

-- Start all tests using async pattern for consistency
print("\nStarting parallel-style requests...")
print("(Note: Due to provider switching requirements, requests run sequentially)")

for i, provider in ipairs(providers_to_test) do
    print("\n  [" .. i .. "/" .. total .. "] Testing " .. provider .. "...")
    
    test_provider(provider, function(response, err)
        if err then
            results[provider] = {
                success = false,
                error = err
            }
            print("       Result: ✗ Failed - " .. err)
        else
            results[provider] = {
                success = true,
                response = response
            }
            print("       Result: ✓ Success")
        end
        completed = completed + 1
    end)
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
    
    if result and result.success then
        print("Status: Success")
        print("Response:")
        print(result.response)
    else
        print("Status: Failed")
        print("Error: " .. (result and result.error or "No response"))
    end
end

-- Summary
print("\n" .. string.rep("=", 80))
print("SUMMARY")
print(string.rep("=", 80))

local successful = 0

for provider, result in pairs(results) do
    if result and result.success then
        successful = successful + 1
    end
end

print("Providers tested: " .. total)
print("Successful: " .. successful)
print("Failed: " .. (total - successful))

print("\n====================")
print("TRUE PARALLEL OPTION")
print("====================")
print([[
For true parallel provider comparison, you could:

1. Run multiple spell instances:
   llmspell run provider-compare --param providers="openai" &
   llmspell run provider-compare --param providers="anthropic" &
   llmspell run provider-compare --param providers="gemini" &

2. Use async callbacks with a single provider:
   See async-callbacks example for true parallel execution

The limitation is that provider switching must happen in the main 
thread, preventing true parallel execution across providers.
]])

print("\n✅ Comparison complete!")