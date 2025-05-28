-- ABOUTME: Simplified chat assistant that works with current implementation
-- ABOUTME: See main_full.lua for the complete interactive version (requires TODO items)

-- Get parameters with defaults
local system_prompt = params.system_prompt or "You are a helpful assistant."

print("Chat Assistant Demo")
print("System: " .. system_prompt)
print("\n⚠️  Note: This is a simplified demo. The full interactive version is in main_full.lua")
print("   It requires:")
print("   - llm.stream_chat_with_history() function (TODO)")
print("   - Safe io.read/write alternatives (TODO)")
print("")

-- Demonstrate basic chat functionality
print("=== Demo: Basic Chat ===\n")

local response, err = llm.chat("Hello! What can you help me with?")
if err then
    print("Error: " .. err)
else
    print("Assistant: " .. response)
end

-- Demonstrate building conversation context manually
print("\n=== Demo: Conversation with Context ===\n")

local conversation = {
    {role = "system", content = system_prompt},
    {role = "user", content = "What is the capital of France?"},
    {role = "assistant", content = "The capital of France is Paris."},
    {role = "user", content = "Tell me an interesting fact about it."}
}

-- Build the full prompt from conversation history
local full_prompt = ""
for _, msg in ipairs(conversation) do
    if msg.role == "system" then
        full_prompt = full_prompt .. "System: " .. msg.content .. "\n\n"
    elseif msg.role == "user" then
        full_prompt = full_prompt .. "User: " .. msg.content .. "\n"
    elseif msg.role == "assistant" then
        full_prompt = full_prompt .. "Assistant: " .. msg.content .. "\n"
    end
end
full_prompt = full_prompt .. "Assistant:"

print("Built prompt from history:")
print("---")
print(full_prompt)
print("---\n")

local response2, err2 = llm.chat(full_prompt)
if err2 then
    print("Error: " .. err2)
else
    print("Assistant: " .. response2)
end

-- Demonstrate streaming (basic version)
print("\n=== Demo: Streaming Response ===\n")

print("You: What are the three primary colors?")
io.write("Assistant: ")
io.flush()

local full_response = ""
local err3 = llm.stream_chat("What are the three primary colors?", function(chunk)
    io.write(chunk)
    io.flush()
    full_response = full_response .. chunk
    return nil
end)

if err3 then
    print("\nError: " .. err3)
else
    print("\n")
    print("(Streamed " .. #full_response .. " characters)")
end

print("\n✅ Demo complete!")
print("\nFor a full interactive chat experience, we need:")
print("1. llm.stream_chat_with_history() - to handle message arrays")
print("2. Safe alternatives to io.read() - for user input")
print("3. These are tracked in TODO.md")
print("\nSee main_full.lua for the planned implementation.")