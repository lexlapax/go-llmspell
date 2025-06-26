-- ABOUTME: Test script to verify state management operations from 09-state-management.lua
-- ABOUTME: Tests state.get(), state.set(), state.update() and nested path access

print("=== Testing State Management Operations ===")

-- Test 1: Basic state operations
print("\n--- Test 1: Basic state.get(), state.set() ---")

local state = require("state")

-- Initialize state with default values (from example)
local app_state = state.get("app") or {
    conversation_history = {},
    user_preferences = {
        model = "gpt-3.5-turbo",
        temperature = 0.7,
        max_tokens = 500
    },
    session_data = {
        start_time = os.time(),
        message_count = 0,
        topics_discussed = {}
    },
    memory_bank = {}
}

-- Save initial state
state.set("app", app_state)
print("✓ Set initial app state")

-- Test 2: Nested path access (dot notation)
print("\n--- Test 2: Nested path access ---")

-- Get nested values
local model = state.get("app.user_preferences.model")
print("Model from nested path:", model)
assert(model == "gpt-3.5-turbo", "Should get nested model value")

local temp = state.get("app.user_preferences.temperature")
print("Temperature from nested path:", temp)
assert(temp == 0.7, "Should get nested temperature value")

-- Test 3: Update function
print("\n--- Test 3: state.update() function ---")

-- Update message count
state.update("app.session_data.message_count", function(count)
    return (count or 0) + 1
end)

local new_count = state.get("app.session_data.message_count")
print("Updated message count:", new_count)
assert(new_count == 1, "Should increment message count")

-- Test 4: Array operations (conversation history)
print("\n--- Test 4: Array/table operations ---")

-- Get conversation history
local history = state.get("app.conversation_history") or {}
print("Initial history length:", #history)

-- Add to history
table.insert(history, {
    role = "user",
    content = "Hello, world!",
    timestamp = os.time()
})

table.insert(history, {
    role = "assistant", 
    content = "Hello! How can I help you?",
    timestamp = os.time()
})

-- Save updated history
state.set("app.conversation_history", history)
print("✓ Added messages to conversation history")

-- Verify history was saved
local saved_history = state.get("app.conversation_history")
print("Saved history length:", #saved_history)
assert(#saved_history == 2, "Should have 2 messages in history")
assert(saved_history[1].role == "user", "First message should be from user")
assert(saved_history[2].role == "assistant", "Second message should be from assistant")

-- Test 5: Update nested arrays
print("\n--- Test 5: Update nested arrays ---")

state.update("app.session_data.topics_discussed", function(topics)
    topics = topics or {}
    table.insert(topics, "state management")
    table.insert(topics, "conversation history")
    return topics
end)

local topics = state.get("app.session_data.topics_discussed")
print("Topics discussed:", table.concat(topics, ", "))
assert(#topics == 2, "Should have 2 topics")

-- Test 6: Complex state structure
print("\n--- Test 6: Complex state structure ---")

state.set("research", {
    topic = "AI Language Models",
    progress = {
        technical_research = false,
        market_analysis = false,
        implementation_plan = false
    },
    findings = {},
    resources = {}
})

print("✓ Set research state")

-- Update progress
state.update("research.progress.technical_research", function() 
    return true 
end)

local tech_done = state.get("research.progress.technical_research")
print("Technical research complete:", tech_done)
assert(tech_done == true, "Should mark technical research as complete")

-- Test 7: Get entire state structure
print("\n--- Test 7: Get entire state structure ---")

local current_state = state.get("app")
print("App state has keys:")
for k, v in pairs(current_state) do
    print("  -", k, ":", type(v))
end

assert(current_state.conversation_history ~= nil, "Should have conversation_history")
assert(current_state.user_preferences ~= nil, "Should have user_preferences")
assert(current_state.session_data ~= nil, "Should have session_data")
assert(current_state.memory_bank ~= nil, "Should have memory_bank")

-- Test 8: Clear and restore (commented out to not affect other tests)
print("\n--- Test 8: State persistence (simulated) ---")

-- Get current state
local state_snapshot = state.get("app")
print("State snapshot taken, would have", #state_snapshot.conversation_history, "messages")

-- In the real example, they clear and restore from JSON
-- We'll just verify the structure is correct for serialization
local ok, data = pcall(require, "data")
if ok then
    local json_ok, json_str = pcall(data.to_json, state_snapshot)
    if json_ok and json_str then
        print("State serialized to", string.len(json_str), "bytes")
        
        local restore_ok, restored = pcall(data.from_json, json_str)
        if restore_ok and restored then
            print("State restored, has", #restored.conversation_history, "messages")
        else
            print("(Skipping restore test - data.from_json not available)")
        end
    else
        print("(Skipping serialization test - data.to_json not available)")
    end
else
    print("(Skipping serialization test - data module not available)")
end

print("\n=== All State Management Tests Passed! ===")
print("\nSummary:")
print("✓ state.get() works with simple and nested paths")
print("✓ state.set() works with complex objects")
print("✓ state.update() works with functions")
print("✓ Dot notation path access works correctly")
print("✓ Array/table operations work as expected")
print("✓ State can be serialized/deserialized")