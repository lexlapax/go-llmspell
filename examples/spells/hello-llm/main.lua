-- ABOUTME: Example spell demonstrating basic LLM interactions
-- ABOUTME: Shows chat, completion, and streaming capabilities

-- Check if LLM module is available
if not llm then
    error("LLM module not available")
end

-- Get current provider
local provider = llm.get_provider()
print("Current LLM provider: " .. provider)

-- List all available providers
local providers = llm.list_providers()
print("\nAvailable providers:")
for i, p in ipairs(providers) do
    print("  " .. i .. ". " .. p)
end

-- Simple chat example
print("\nSending chat message...")
local response, err = llm.chat("Hello! Please respond with a simple greeting.")
if err then
    print("Error: " .. err)
else
    print("Response: " .. response)
end

-- Completion with max tokens
print("\nGenerating completion...")
local completion, err = llm.complete("The capital of France is", 10)
if err then
    print("Error: " .. err)
else
    print("Completion: " .. completion)
end

-- Example with streaming (if supported)
print("\nStreaming example:")
local function stream_callback(chunk)
    -- Print each chunk as it arrives (io.write is disabled for security)
    print(chunk)
    return nil -- Return nil to continue, or error string to stop
end

-- Note: streaming might not work without proper io library
-- This is just to show the API
local err = llm.stream_chat("Tell me a very short joke.", stream_callback)
if err then
    print("\nStreaming error: " .. err)
else
    print("\nStreaming complete.")
end