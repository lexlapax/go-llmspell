-- ABOUTME: Basic LLM interaction example showing simple completions and streaming
-- ABOUTME: Demonstrates core LLM module usage with different models and parameters

-- Basic LLM Interaction Example
-- This spell demonstrates fundamental LLM operations including:
-- 1. Simple text completion
-- 2. Conversation with context
-- 3. Streaming responses
-- 4. Different model usage
-- 5. Error handling

-- Validate parameters
local model = params.model or "gpt-3.5-turbo"
local prompt = params.prompt or "Tell me an interesting fact about Lua programming language."

-- Example 1: Simple completion
print("=== Example 1: Simple Completion ===")
local response = llm.complete({
    model = model,
    messages = {
        {role = "user", content = prompt}
    }
})

print("Response: " .. response.content)
print("Tokens used: " .. tostring(response.usage.total_tokens))
print()

-- Example 2: Completion with system message and temperature
print("=== Example 2: Completion with System Message ===")
local creative_response = llm.complete({
    model = model,
    messages = {
        {role = "system", content = "You are a creative storyteller who speaks in a whimsical, poetic style."},
        {role = "user", content = "Describe a rainy day in 3 sentences."}
    },
    temperature = 0.9,  -- Higher temperature for more creativity
    max_tokens = 150
})

print("Creative response: " .. creative_response.content)
print()

-- Example 3: Multi-turn conversation
print("=== Example 3: Multi-turn Conversation ===")
local conversation = {
    {role = "system", content = "You are a helpful coding assistant specializing in Lua."},
    {role = "user", content = "What's the difference between pairs and ipairs in Lua?"}
}

local response1 = llm.complete({
    model = model,
    messages = conversation
})

print("Assistant: " .. response1.content)

-- Add the response to conversation history
table.insert(conversation, {role = "assistant", content = response1.content})
table.insert(conversation, {role = "user", content = "Can you show me a code example?"})

-- Continue the conversation
local response2 = llm.complete({
    model = model,
    messages = conversation
})

print("\nUser: Can you show me a code example?")
print("Assistant: " .. response2.content)
print()

-- Example 4: Streaming response
print("=== Example 4: Streaming Response ===")
print("Streaming response: ")

local full_content = ""
llm.stream({
    model = model,
    messages = {
        {role = "user", content = "Write a haiku about programming"}
    },
    on_content = function(content)
        -- Print each chunk as it arrives
        io.write(content)
        io.flush()
        full_content = full_content .. content
    end,
    on_complete = function(response)
        -- Called when streaming is complete
        print("\n\nStreaming complete!")
        print("Total tokens: " .. tostring(response.usage.total_tokens))
    end,
    on_error = function(error)
        print("\nStreaming error: " .. error)
    end
})

print()

-- Example 5: Using different models
print("=== Example 5: Model Comparison ===")
local test_prompt = "Explain recursion in one sentence."

-- Try different models (if available)
local models_to_try = {
    "gpt-3.5-turbo",
    "gpt-4",
    -- Add more models as needed
}

for _, test_model in ipairs(models_to_try) do
    -- Use pcall for error handling in case model is not available
    local success, result = pcall(function()
        return llm.complete({
            model = test_model,
            messages = {{role = "user", content = test_prompt}},
            temperature = 0  -- Use 0 for consistent comparison
        })
    end)
    
    if success then
        print(test_model .. ": " .. result.content)
    else
        print(test_model .. ": Not available or error - " .. tostring(result))
    end
end

print()

-- Example 6: Error handling
print("=== Example 6: Error Handling ===")

-- Function to safely call LLM with retry logic
local function safe_llm_call(request, max_retries)
    max_retries = max_retries or 3
    local last_error = nil
    
    for attempt = 1, max_retries do
        local success, result = pcall(function()
            return llm.complete(request)
        end)
        
        if success then
            return result
        else
            last_error = result
            log.warn("LLM call failed (attempt " .. attempt .. "): " .. tostring(result))
            
            -- Wait before retry with exponential backoff
            if attempt < max_retries then
                core.sleep(math.pow(2, attempt - 1))
            end
        end
    end
    
    error("LLM call failed after " .. max_retries .. " attempts: " .. tostring(last_error))
end

-- Test the safe call function
local safe_response = safe_llm_call({
    model = model,
    messages = {{role = "user", content = "Say 'Hello, safe world!'"}},
    temperature = 0.5
})

print("Safe call response: " .. safe_response.content)

-- Example 7: JSON mode (if supported by model)
print("\n=== Example 7: Structured Output ===")
local json_response = llm.complete({
    model = model,
    messages = {
        {role = "system", content = "You must respond with valid JSON only."},
        {role = "user", content = "Create a JSON object with name, age, and city fields for a fictional person."}
    },
    temperature = 0.3
})

print("JSON response: " .. json_response.content)

-- Try to parse the JSON
local success, parsed = pcall(function()
    return data.from_json(json_response.content)
end)

if success then
    print("Successfully parsed JSON:")
    print("  Name: " .. (parsed.name or "N/A"))
    print("  Age: " .. tostring(parsed.age or "N/A"))
    print("  City: " .. (parsed.city or "N/A"))
else
    print("Failed to parse JSON: " .. tostring(parsed))
end

-- Example 8: Token counting and limits
print("\n=== Example 8: Token Management ===")

local long_prompt = string.rep("This is a test sentence. ", 100)  -- Create a long prompt

-- First, try with limited tokens
local limited_response = llm.complete({
    model = model,
    messages = {{role = "user", content = "Summarize this text: " .. long_prompt}},
    max_tokens = 50  -- Limit response length
})

print("Limited response (50 tokens max):")
print(limited_response.content)
print("Finish reason: " .. (limited_response.finish_reason or "unknown"))

-- Summary of the examples
print("\n=== Summary ===")
print("This spell demonstrated:")
print("1. Simple LLM completions")
print("2. System messages for behavior control")
print("3. Multi-turn conversations")
print("4. Streaming responses")
print("5. Multiple model usage")
print("6. Error handling and retries")
print("7. Structured output generation")
print("8. Token management")

-- Return some useful data
return {
    examples_run = 8,
    model_used = model,
    total_tokens_estimate = response.usage.total_tokens + creative_response.usage.total_tokens,
    streaming_content = full_content
}