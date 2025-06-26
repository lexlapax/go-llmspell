-- ABOUTME: State management example showing persistent state, context sharing, and data flow
-- ABOUTME: Demonstrates state module usage for maintaining conversation context and data persistence

-- State Management Example
-- This spell demonstrates comprehensive state management including:
-- 1. Persistent state across script execution
-- 2. Shared context between components
-- 3. State merging and updates
-- 4. State serialization and restoration
-- 5. Transaction-like state updates

-- Required modules
local state = require("state")
local llm = require("llm")
local agent = require("agent")
local data = require("data")
local core = require("core")
local log = require("log")
local utils = require("utils")

-- Initialize state with default values
local function initialize_state()
    -- Get or create application state
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
    return app_state
end

-- Example 1: Conversation with Memory
print("=== Example 1: Conversation with Memory ===")

local app_state = initialize_state()

-- Function to add message to conversation history
local function add_to_history(role, content)
    local history = state.get("app.conversation_history") or {}
    table.insert(history, {
        role = role,
        content = content,
        timestamp = os.time()
    })
    
    -- Keep only last 10 messages for context window
    if #history > 10 then
        table.remove(history, 1)
    end
    
    state.set("app.conversation_history", history)
    state.update("app.session_data.message_count", function(count)
        return (count or 0) + 1
    end)
end

-- Function to get conversation context
local function get_conversation_context()
    local history = state.get("app.conversation_history") or {}
    local messages = {}
    
    -- Add system message with context
    table.insert(messages, {
        role = "system",
        content = "You are a helpful assistant with memory of our conversation."
    })
    
    -- Add conversation history
    for _, msg in ipairs(history) do
        table.insert(messages, {
            role = msg.role,
            content = msg.content
        })
    end
    
    return messages
end

-- Simulate conversation
local topics = {"Lua programming", "state management", "AI assistants"}

-- Create an agent instead of using llm directly
print("Creating assistant agent...")
print("Model:", state.get("app.user_preferences.model"))
print("Temperature:", state.get("app.user_preferences.temperature"))

local assistant = agent.create("Assistant", {
    model = state.get("app.user_preferences.model"),
    system = "You are a helpful assistant with memory of our conversation.",
    temperature = state.get("app.user_preferences.temperature")
})
print("Assistant created successfully")

for i, topic in ipairs(topics) do
    print("\nTurn " .. i .. ": Discussing " .. topic)
    
    -- User message
    local user_message = "Tell me about " .. topic
    add_to_history("user", user_message)
    
    -- Get response with context
    local messages = get_conversation_context()
    
    -- Build context-aware prompt
    local context_prompt = ""
    for _, msg in ipairs(messages) do
        if msg.role ~= "system" then
            context_prompt = context_prompt .. msg.role .. ": " .. msg.content .. "\n"
        end
    end
    context_prompt = context_prompt .. "user: " .. user_message .. "\nassistant: "
    
    -- Get response using agent
    local response = assistant:run(context_prompt)
    
    -- Save response to history
    add_to_history("assistant", response.content or response)
    
    -- Track topics
    state.update("app.session_data.topics_discussed", function(topics)
        topics = topics or {}
        table.insert(topics, topic)
        return topics
    end)
    
    print("Assistant: " .. string.sub(response.content or tostring(response), 1, 100) .. "...")
end

-- Show session summary
local session = state.get("app.session_data")
print("\n=== Session Summary ===")
print("Messages exchanged: " .. session.message_count)
print("Topics discussed: " .. table.concat(session.topics_discussed, ", "))
print()

-- Example 2: State Persistence and Restoration
print("=== Example 2: State Persistence and Restoration ===")

-- Save current state to JSON
local current_state = state.get("app")
local state_json = data.to_json(current_state)
print("State size: " .. #state_json .. " bytes")

-- Simulate clearing state (new session)
state.clear("app")
print("State cleared")

-- Restore from saved state
local restored_state = data.from_json(state_json)
state.set("app", restored_state)
print("State restored")

-- Verify restoration
local history = state.get("app.conversation_history")
print("Restored " .. #history .. " messages from history")
print()

-- Example 3: Shared State Between Agents
print("=== Example 3: Shared State Between Agents ===")

-- Create a research coordinator agent
local coordinator = agent.create("Research Coordinator", {
    model = "gpt-4",
    system = [[
        You coordinate research tasks. 
        Track progress in shared state under 'research.progress'.
        Assign tasks to researcher agents.
    ]]
})

-- Create researcher agents
local researcher1 = agent.create("Researcher 1", {
    model = "gpt-3.5-turbo",
    system = "You research technical topics and update shared state with findings."
})

local researcher2 = agent.create("Researcher 2", {
    model = "gpt-3.5-turbo",
    system = "You research practical applications and update shared state with findings."
})

-- Initialize research state
state.set("research", {
    topic = "State Management in Distributed Systems",
    progress = {
        technical_research = "pending",
        practical_research = "pending",
        synthesis = "pending"
    },
    findings = {}
})

-- Coordinator assigns tasks
print("Coordinator planning research...")
local plan = coordinator:run({
    prompt = "Create a research plan for: " .. state.get("research.topic"),
    max_tokens = 200
})

print("Research plan created")

-- Researchers work in parallel (simulated)
print("\nResearchers working...")

-- Researcher 1 updates state
state.update("research.progress.technical_research", function() 
    return "in_progress" 
end)

local tech_findings = researcher1:run({
    prompt = "Research technical aspects of: " .. state.get("research.topic"),
    max_tokens = 300
})

state.update("research.findings", function(findings)
    findings = findings or {}
    findings.technical = tech_findings.content
    return findings
end)

state.update("research.progress.technical_research", function() 
    return "completed" 
end)

-- Researcher 2 updates state
state.update("research.progress.practical_research", function() 
    return "in_progress" 
end)

local practical_findings = researcher2:run({
    prompt = "Research practical applications of: " .. state.get("research.topic"),
    max_tokens = 300
})

state.update("research.findings", function(findings)
    findings = findings or {}
    findings.practical = practical_findings.content
    return findings
end)

state.update("research.progress.practical_research", function() 
    return "completed" 
end)

-- Coordinator synthesizes findings
print("\nCoordinator synthesizing findings...")
local findings = state.get("research.findings")
local synthesis = coordinator:run({
    prompt = string.format(
        "Synthesize these research findings on \"%s\":\n\nTechnical: %s\n\nPractical: %s\n\nCreate a comprehensive summary with key insights and recommendations.",
        state.get("research.topic"), findings.technical or "none", findings.practical or "none"
    ),
    max_tokens = 400
})

state.update("research.findings", function(findings)
    findings.synthesis = synthesis.content
    return findings
end)

state.update("research.progress.synthesis", function() 
    return "completed" 
end)

-- Show research results
local progress = state.get("research.progress")
print("\n=== Research Progress ===")
for task, status in pairs(progress) do
    print(task .. ": " .. status)
end
print()

-- Example 4: State Transactions and Rollback
print("=== Example 4: State Transactions and Rollback ===")

-- Create a transaction-like update
local function transactional_update(updates, validator)
    -- Save current state
    local backup = state.get("app")
    
    -- Apply updates
    local success = true
    local error_msg = nil
    
    for key, value in pairs(updates) do
        state.set("app." .. key, value)
    end
    
    -- Validate the new state
    if validator then
        success, error_msg = validator(state.get("app"))
    end
    
    -- Rollback if validation fails
    if not success then
        state.set("app", backup)
        return false, error_msg
    end
    
    return true
end

-- Try valid update
print("Attempting valid preference update...")
local success, err = transactional_update({
    ["user_preferences.temperature"] = 0.8,
    ["user_preferences.max_tokens"] = 1000
}, function(new_state)
    -- Validate temperature is in range
    local temp = new_state.user_preferences.temperature
    if temp < 0 or temp > 2 then
        return false, "Temperature out of range"
    end
    return true
end)

print("Update result: " .. (success and "Success" or ("Failed: " .. (err or "unknown"))))

-- Try invalid update
print("\nAttempting invalid preference update...")
success, err = transactional_update({
    ["user_preferences.temperature"] = 3.0,  -- Invalid
    ["user_preferences.max_tokens"] = 2000
}, function(new_state)
    local temp = new_state.user_preferences.temperature
    if temp < 0 or temp > 2 then
        return false, "Temperature out of range"
    end
    return true
end)

print("Update result: " .. (success and "Success" or ("Failed: " .. (err or "unknown"))))
print("Current temperature: " .. state.get("app.user_preferences.temperature"))
print()

-- Example 5: State Queries and Aggregation
print("=== Example 5: State Queries and Aggregation ===")

-- Add some memory bank entries
local memories = {
    {type = "fact", content = "Lua was created in Brazil", importance = 0.9},
    {type = "fact", content = "Lua means moon in Portuguese", importance = 0.7},
    {type = "skill", content = "State management patterns", importance = 0.95},
    {type = "context", content = "User prefers concise explanations", importance = 0.8}
}

state.set("app.memory_bank", memories)

-- Query memories by type
local function get_memories_by_type(mem_type)
    local all_memories = state.get("app.memory_bank") or {}
    local filtered = {}
    
    for _, memory in ipairs(all_memories) do
        if memory.type == mem_type then
            table.insert(filtered, memory)
        end
    end
    
    return filtered
end

-- Aggregate importance scores
local function get_average_importance()
    local all_memories = state.get("app.memory_bank") or {}
    local total = 0
    local count = 0
    
    for _, memory in ipairs(all_memories) do
        total = total + (memory.importance or 0)
        count = count + 1
    end
    
    return count > 0 and (total / count) or 0
end

-- Display memory statistics
print("Memory Bank Statistics:")
print("Total memories: " .. #(state.get("app.memory_bank") or {}))
print("Facts: " .. #get_memories_by_type("fact"))
print("Skills: " .. #get_memories_by_type("skill"))
print("Context: " .. #get_memories_by_type("context"))
print("Average importance: " .. string.format("%.2f", get_average_importance()))
print()

-- Example 6: State Watchers and Reactive Updates
print("=== Example 6: State Watchers and Reactive Updates ===")

-- Create a state watcher function
local function create_watcher(path, callback)
    local last_value = state.get(path)
    
    return function()
        local current_value = state.get(path)
        if current_value ~= last_value then
            callback(current_value, last_value)
            last_value = current_value
        end
    end
end

-- Watch message count changes
local message_watcher = create_watcher("app.session_data.message_count", function(new_val, old_val)
    print("Message count changed from " .. (old_val or 0) .. " to " .. (new_val or 0))
    
    -- Trigger milestone actions
    if new_val % 5 == 0 then
        print("  Milestone reached! Saving checkpoint...")
        state.set("app.last_checkpoint", os.time())
    end
end)

-- Simulate activity
for i = 1, 3 do
    add_to_history("user", "Test message " .. i)
    message_watcher()  -- Check for changes
    utils.general_sleep(100)  -- Sleep for 100ms
end

-- Final state summary
print("\n=== Final State Summary ===")
local final_state = state.get("app")
print("Conversation messages: " .. #(final_state.conversation_history or {}))
print("Total messages sent: " .. (final_state.session_data.message_count or 0))
print("Memory bank size: " .. #(final_state.memory_bank or {}))
print("Last checkpoint: " .. os.date("%Y-%m-%d %H:%M:%S", final_state.last_checkpoint or 0))

-- Return state statistics
return {
    success = true,
    messages_processed = final_state.session_data.message_count,
    memories_stored = #(final_state.memory_bank or {}),
    topics_discussed = final_state.session_data.topics_discussed
}