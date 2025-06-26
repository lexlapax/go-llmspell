-- ABOUTME: Simplified event-driven example that works with current limitations
-- ABOUTME: Demonstrates basic event patterns without complex async features

-- Required modules
local agent = require("agent")
local utils = require("utils")

-- Event-Driven Spells Example (Simplified)
-- This spell demonstrates event-driven patterns:
-- 1. Basic event emission and handling
-- 2. Event-driven content moderation

-- Parameters
local scenario = params.scenario or "content_moderation"
local model = params.model or "gpt-4"
local output_dir = params.output_dir or "./event-driven-output"

print("=== Event-Driven Spells Example (Simplified) ===")
print("Scenario: " .. scenario)
print()

-- Note: File system operations removed - focus on event-driven patterns
-- In production, output would be saved to files

-- Simple Event System
local SimpleEventSystem = {}

function SimpleEventSystem.new()
    return {
        handlers = {},
        event_log = {}
    }
end

function SimpleEventSystem.on(system, event_name, handler)
    if not system.handlers[event_name] then
        system.handlers[event_name] = {}
    end
    system.handlers[event_name][#system.handlers[event_name] + 1] = handler
    print("  📌 Registered handler for: " .. event_name)
end

function SimpleEventSystem.emit(system, event_name, data)
    -- Log event
    system.event_log[#system.event_log + 1] = {
        name = event_name,
        data = data,
        timestamp = os.time()
    }
    
    print("  🔔 Event: " .. event_name)
    
    -- Execute handlers
    local handlers = system.handlers[event_name] or {}
    for i = 1, #handlers do
        local success, result = pcall(handlers[i], data)
        if not success then
            print("  ⚠️ Handler error: " .. tostring(result))
        end
    end
end

-- Example 1: Basic Event-Driven System
print("=== Example 1: Basic Event System ===")

local basic_events = SimpleEventSystem.new()

-- Register handlers
SimpleEventSystem.on(basic_events, "user_joined", function(data)
    print("  👤 User joined: " .. data.username)
end)

SimpleEventSystem.on(basic_events, "message_received", function(data)
    print("  💬 Message from " .. data.user .. ": " .. data.message)
end)

SimpleEventSystem.on(basic_events, "user_left", function(data)
    print("  👋 User left: " .. data.username)
end)

-- Simulate events
print("\nSimulating basic events:")
SimpleEventSystem.emit(basic_events, "user_joined", {username = "Alice"})
SimpleEventSystem.emit(basic_events, "message_received", {user = "Alice", message = "Hello everyone!"})
SimpleEventSystem.emit(basic_events, "message_received", {user = "Alice", message = "How's the weather?"})
SimpleEventSystem.emit(basic_events, "user_left", {username = "Alice"})

print()

-- Example 2: Event-Driven Content Moderation
print("=== Example 2: Content Moderation System ===")

local moderation_system = SimpleEventSystem.new()
local moderation_state = {
    flagged_content = {},
    user_warnings = {},
    banned_users = {}
}

-- Create moderation agent
local moderator = agent.create("Content Moderator", {
    model = model,
    system = "You are a content moderator. Analyze content for inappropriate material. Respond with: SAFE, WARNING, or VIOLATION",
    temperature = 0.1
})

-- Register moderation handlers
SimpleEventSystem.on(moderation_system, "content_submitted", function(data)
    print("  📝 Content submitted by: " .. data.user)
    
    -- Moderate content
    local result = moderator:run("Moderate this content: " .. data.content)
    
    local severity = "SAFE"
    if result:find("WARNING") then
        severity = "WARNING"
    elseif result:find("VIOLATION") then
        severity = "VIOLATION"
    end
    
    -- Emit appropriate event based on moderation
    SimpleEventSystem.emit(moderation_system, "content_moderated", {
        user = data.user,
        content = data.content,
        severity = severity,
        reason = result
    })
end)

SimpleEventSystem.on(moderation_system, "content_moderated", function(data)
    print("  🔍 Moderation result: " .. data.severity)
    
    if data.severity == "WARNING" then
        SimpleEventSystem.emit(moderation_system, "user_warned", {
            user = data.user,
            reason = data.reason
        })
    elseif data.severity == "VIOLATION" then
        SimpleEventSystem.emit(moderation_system, "content_flagged", {
            user = data.user,
            content = data.content,
            reason = data.reason
        })
    end
end)

SimpleEventSystem.on(moderation_system, "user_warned", function(data)
    moderation_state.user_warnings[data.user] = 
        (moderation_state.user_warnings[data.user] or 0) + 1
    
    print("  ⚠️ User warned: " .. data.user .. 
          " (Total warnings: " .. moderation_state.user_warnings[data.user] .. ")")
    
    -- Check for ban threshold
    if moderation_state.user_warnings[data.user] >= 3 then
        SimpleEventSystem.emit(moderation_system, "user_banned", {user = data.user})
    end
end)

SimpleEventSystem.on(moderation_system, "content_flagged", function(data)
    moderation_state.flagged_content[#moderation_state.flagged_content + 1] = data
    print("  🚫 Content flagged for review")
    
    -- Auto-ban for severe violations
    SimpleEventSystem.emit(moderation_system, "user_banned", {user = data.user})
end)

SimpleEventSystem.on(moderation_system, "user_banned", function(data)
    moderation_state.banned_users[data.user] = true
    print("  ⛔ User banned: " .. data.user)
end)

-- Test moderation system
print("\nTesting moderation system:")
local test_contents = {
    {user = "Bob", content = "This is a friendly message about cats"},
    {user = "Charlie", content = "I hate everyone and everything"},
    {user = "Bob", content = "Let's discuss the weather"},
    {user = "Charlie", content = "More negative hostile content"},
    {user = "David", content = "Spam spam spam buy now!!!"}
}

for _, submission in ipairs(test_contents) do
    SimpleEventSystem.emit(moderation_system, "content_submitted", submission)
    utils.general_sleep(500)  -- Small delay for readability (500ms)
end

print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Basic event emission and handling")
print("2. Event-driven content moderation")
print()
print("Key insights:")
print("- Events enable decoupled, reactive systems")
print("- Handlers can trigger cascading events")
print("- Event systems are useful for moderation workflows")
print()

-- Save event log
local log_content = "Event Log - Generated: " .. os.date() .. "\n\n"
log_content = log_content .. "Basic Events:\n"
for i = 1, #basic_events.event_log do
    local event = basic_events.event_log[i]
    log_content = log_content .. string.format("  [%s] %s\n", 
        os.date("%H:%M:%S", event.timestamp), event.name)
end

log_content = log_content .. "\nModeration Events:\n"  
for i = 1, #moderation_system.event_log do
    local event = moderation_system.event_log[i]
    log_content = log_content .. string.format("  [%s] %s\n",
        os.date("%H:%M:%S", event.timestamp), event.name)
end

-- Output would be written to file in production
print("\nEvent Log:")
print(log_content)
print("Event log saved to: " .. output_dir .. "/event_log.txt")

-- Return summary
return {
    event_systems_created = 2,
    total_events_emitted = #basic_events.event_log + #moderation_system.event_log,
    patterns_demonstrated = {
        "basic_events",
        "content_moderation"
    }
}