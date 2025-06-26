-- ABOUTME: Example demonstrating event-driven programming patterns in spells
-- ABOUTME: Shows event handling, reactive systems, and asynchronous event flows

-- Required modules
local agent = require("agent")
local utils = require("utils")  -- For utilities
local events = require("events")  -- For event system

-- Event-Driven Spells Example
-- This spell demonstrates event-driven patterns:
-- 1. Basic event emission and handling
-- 2. Event-driven state changes
-- 3. Asynchronous event processing
-- 4. Event chains and cascades
-- 5. Event filtering and transformation
-- 6. Complex event-driven systems

-- Parameters
local scenario = params.scenario or "content_moderation"
local model = params.model or "gpt-4"
local output_dir = params.output_dir or "./event-driven-output"

print("=== Event-Driven Spells Example ===")
print("Scenario: " .. scenario)
print()

-- File operations not available in current implementation
-- In production, would ensure output directory exists

-- Helper function for counting table entries
local function table_count(t)
    local count = 0
    for _ in pairs(t) do count = count + 1 end
    return count
end

-- Enhanced Event System
local EventSystem = {}

function EventSystem:new()
    local obj = {
        handlers = {},
        event_log = {},
        filters = {},
        transformers = {},
        async_queue = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function EventSystem:on(event_name, handler, options)
    options = options or {}
    if not self.handlers[event_name] then
        self.handlers[event_name] = {}
    end
    
    -- Debug: Check types
    if type(self.handlers[event_name]) ~= "table" then
        print("ERROR: self.handlers[event_name] is not a table!", type(self.handlers[event_name]))
        return
    end
    
    table.insert(self.handlers[event_name], {
        handler = handler,
        priority = options.priority or 0,
        once = options.once or false,
        async = options.async or false,
        id = options.id or ("handler_" .. #self.handlers[event_name] + 1)
    })
    
    -- Sort by priority (higher first)
    table.sort(self.handlers[event_name], function(a, b)
        return a.priority > b.priority
    end)
    
    print("  📌 Registered handler for: " .. event_name)
end

function EventSystem:off(event_name, handler_id)
    if self.handlers[event_name] then
        for i, h in ipairs(self.handlers[event_name]) do
            if h.id == handler_id then
                table.remove(self.handlers[event_name], i)
                print("  📌 Removed handler: " .. handler_id)
                break
            end
        end
    end
end

function EventSystem:emit(event_name, data)
    -- Log event
    table.insert(self.event_log, {
        name = event_name,
        data = data,
        timestamp = os.time()
    })
    
    print("  🔔 Event: " .. event_name)
    
    -- Apply filters
    if self.filters[event_name] then
        for _, filter in ipairs(self.filters[event_name]) do
            if not filter(data) then
                print("  🚫 Event filtered out")
                return
            end
        end
    end
    
    -- Apply transformers
    if self.transformers[event_name] then
        for _, transformer in ipairs(self.transformers[event_name]) do
            data = transformer(data)
        end
    end
    
    -- Get handlers
    local handlers = self.handlers[event_name] or {}
    local handlers_to_remove = {}
    
    -- Execute handlers
    for i, handler_info in ipairs(handlers) do
        if handler_info.async then
            -- Queue async handlers
            table.insert(self.async_queue, {
                handler = handler_info.handler,
                data = data,
                event_name = event_name
            })
        else
            -- Execute sync handlers
            local success, result = pcall(handler_info.handler, data)
            if not success then
                print("  ⚠️ Handler error: " .. tostring(result))
            end
        end
        
        -- Mark for removal if "once"
        if handler_info.once then
            table.insert(handlers_to_remove, i)
        end
    end
    
    -- Remove "once" handlers
    for i = #handlers_to_remove, 1, -1 do
        table.remove(self.handlers[event_name], handlers_to_remove[i])
    end
end

function EventSystem:add_filter(event_name, filter_fn)
    if not self.filters[event_name] then
        self.filters[event_name] = {}
    end
    table.insert(self.filters[event_name], filter_fn)
end

function EventSystem:add_transformer(event_name, transformer_fn)
    if not self.transformers[event_name] then
        self.transformers[event_name] = {}
    end
    table.insert(self.transformers[event_name], transformer_fn)
end

function EventSystem:process_async_queue()
    while #self.async_queue > 0 do
        local task = table.remove(self.async_queue, 1)
        -- Process async task (simplified without core.async)
        print("  ⚡ Processing async event: " .. task.event_name)
        local success, result = pcall(task.handler, task.data)
        if not success then
            print("  ⚠️ Async handler error: " .. tostring(result))
        end
    end
end

function EventSystem:get_event_stats()
    local stats = {}
    for _, event in ipairs(self.event_log) do
        stats[event.name] = (stats[event.name] or 0) + 1
    end
    return stats
end

-- Example 1: Basic Event-Driven System
print("=== Example 1: Basic Event System ===")

local basic_events = EventSystem:new()

-- Register handlers
basic_events:on("user_joined", function(data)
    print("  👤 User joined: " .. data.username)
end)

basic_events:on("message_received", function(data)
    print("  💬 Message from " .. data.user .. ": " .. data.message)
end)

basic_events:on("user_left", function(data)
    print("  👋 User left: " .. data.username)
end)

-- Simulate events
print("\nSimulating basic events:")
basic_events:emit("user_joined", {username = "Alice"})
basic_events:emit("message_received", {user = "Alice", message = "Hello everyone!"})
basic_events:emit("message_received", {user = "Alice", message = "How's the weather?"})
basic_events:emit("user_left", {username = "Alice"})

print()

-- Example 2: Event-Driven Content Moderation
print("=== Example 2: Content Moderation System ===")

local moderation_system = EventSystem:new()
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
moderation_system:on("content_submitted", function(data)
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
    moderation_system:emit("content_moderated", {
        user = data.user,
        content = data.content,
        severity = severity,
        reason = result
    })
end)

moderation_system:on("content_moderated", function(data)
    print("  🔍 Moderation result: " .. data.severity)
    
    if data.severity == "WARNING" then
        moderation_system:emit("user_warned", {
            user = data.user,
            reason = data.reason
        })
    elseif data.severity == "VIOLATION" then
        moderation_system:emit("content_flagged", {
            user = data.user,
            content = data.content,
            reason = data.reason
        })
    end
end)

moderation_system:on("user_warned", function(data)
    moderation_state.user_warnings[data.user] = 
        (moderation_state.user_warnings[data.user] or 0) + 1
    
    print("  ⚠️ User warned: " .. data.user .. 
          " (Total warnings: " .. moderation_state.user_warnings[data.user] .. ")")
    
    -- Check for ban threshold
    if moderation_state.user_warnings[data.user] >= 3 then
        moderation_system:emit("user_banned", {user = data.user})
    end
end)

moderation_system:on("content_flagged", function(data)
    table.insert(moderation_state.flagged_content, data)
    print("  🚫 Content flagged for review")
    
    -- Auto-ban for severe violations
    moderation_system:emit("user_banned", {user = data.user})
end)

moderation_system:on("user_banned", function(data)
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
    moderation_system:emit("content_submitted", submission)
    utils.general_sleep(0.5)  -- Small delay for readability
end

print()

-- Example 3: Event-Driven Workflow with State Machine
print("=== Example 3: Event-Driven State Machine ===")

local StateMachine = {}

function StateMachine:new(initial_state)
    local obj = {
        current_state = initial_state,
        transitions = {},
        state_handlers = {},
        events = EventSystem:new()
    }
    setmetatable(obj, {__index = self})
    
    -- Auto-emit state change events
    obj.events:on("transition", function(data)
        obj.current_state = data.to
        obj.events:emit("state:" .. data.to, data)
    end)
    
    return obj
end

function StateMachine:add_transition(from, to, event)
    if not self.transitions[from] then
        self.transitions[from] = {}
    end
    self.transitions[from][event] = to
end

function StateMachine:on_state(state, handler)
    self.events:on("state:" .. state, handler)
end

function StateMachine:trigger(event, data)
    local transitions = self.transitions[self.current_state]
    if transitions and transitions[event] then
        local new_state = transitions[event]
        print("  🔄 State transition: " .. self.current_state .. " → " .. new_state)
        
        self.events:emit("transition", {
            from = self.current_state,
            to = new_state,
            event = event,
            data = data
        })
    else
        print("  ❌ No transition for event '" .. event .. "' in state '" .. self.current_state .. "'")
    end
end

-- Create document processing state machine
local doc_processor = StateMachine:new("idle")

-- Define transitions
doc_processor:add_transition("idle", "uploading", "upload_start")
doc_processor:add_transition("uploading", "processing", "upload_complete")
doc_processor:add_transition("processing", "reviewing", "process_complete")
doc_processor:add_transition("reviewing", "approved", "approve")
doc_processor:add_transition("reviewing", "rejected", "reject")
doc_processor:add_transition("approved", "idle", "reset")
doc_processor:add_transition("rejected", "idle", "reset")

-- Define state handlers
doc_processor:on_state("uploading", function(data)
    print("  📤 Uploading document...")
    utils.general_sleep(1)
    doc_processor:trigger("upload_complete", {filename = "document.pdf"})
end)

doc_processor:on_state("processing", function(data)
    print("  ⚙️ Processing document: " .. (data.data.filename or "unknown"))
    
    -- Simulate processing with agent
    local processor = agent.create("Document Processor", {
        model = model,
        system = "You process documents. Extract key information.",
        temperature = 0.3
    })
    
    local result = processor:run("Process this document: [Document content here]")
    utils.general_sleep(1)
    
    doc_processor:trigger("process_complete", {result = result})
end)

doc_processor:on_state("reviewing", function(data)
    print("  👁️ Reviewing processed document...")
    
    -- Simulate review decision
    local decision = math.random() > 0.3 and "approve" or "reject"
    utils.general_sleep(1)
    
    doc_processor:trigger(decision, {reason = "Review complete"})
end)

doc_processor:on_state("approved", function(data)
    print("  ✅ Document approved!")
    -- Would save: output_dir .. "/approved_doc.txt"
    print("      Would save approval record")
    doc_processor:trigger("reset")
end)

doc_processor:on_state("rejected", function(data)
    print("  ❌ Document rejected!")
    doc_processor:trigger("reset")
end)

-- Test state machine
print("\nTesting document processing state machine:")
doc_processor:trigger("upload_start")
print()

-- Example 4: Event Cascades and Chains
print("=== Example 4: Event Cascades ===")

local cascade_system = EventSystem:new()
local cascade_state = {
    data_points = {},
    analysis_results = {},
    reports_generated = 0
}

-- Create event cascade for data pipeline
cascade_system:on("data_received", function(data)
    print("  📊 Data received: " .. data.type)
    table.insert(cascade_state.data_points, data)
    
    -- Trigger validation
    cascade_system:emit("validate_data", data)
end)

cascade_system:on("validate_data", function(data)
    print("  ✔️ Validating data...")
    
    -- Simple validation
    local is_valid = data.value and type(data.value) == "number"
    
    if is_valid then
        cascade_system:emit("data_validated", data)
    else
        cascade_system:emit("data_invalid", data)
    end
end)

cascade_system:on("data_validated", function(data)
    print("  ✅ Data valid, processing...")
    
    -- Process data
    local processed = {
        original = data.value,
        squared = data.value ^ 2,
        sqrt = math.sqrt(math.abs(data.value))
    }
    
    cascade_system:emit("data_processed", processed)
end)

cascade_system:on("data_invalid", function(data)
    print("  ⚠️ Invalid data, attempting correction...")
    
    -- Try to correct
    if data.value then
        data.value = tonumber(tostring(data.value)) or 0
        cascade_system:emit("validate_data", data)  -- Re-validate
    end
end)

cascade_system:on("data_processed", function(data)
    print("  📈 Data processed, checking thresholds...")
    
    table.insert(cascade_state.analysis_results, data)
    
    -- Check if we have enough data for report
    if #cascade_state.analysis_results >= 3 then
        cascade_system:emit("generate_report", cascade_state.analysis_results)
    end
end)

cascade_system:on("generate_report", function(results)
    print("  📄 Generating report...")
    
    cascade_state.reports_generated = cascade_state.reports_generated + 1
    
    local report = "Data Analysis Report #" .. cascade_state.reports_generated .. "\n"
    report = report .. "Generated: " .. os.date() .. "\n\n"
    
    for i, result in ipairs(results) do
        report = report .. string.format(
            "Data Point %d: Original=%.2f, Squared=%.2f, Sqrt=%.2f\n",
            i, result.original, result.squared, result.sqrt
        )
    end
    
    -- Would save report to: output_dir .. "/cascade_report_" .. cascade_state.reports_generated .. ".txt"
    print("    Cascade report generated: #" .. cascade_state.reports_generated)
    
    -- Clear processed results
    cascade_state.analysis_results = {}
    
    cascade_system:emit("report_completed", {
        report_number = cascade_state.reports_generated
    })
end)

cascade_system:on("report_completed", function(data)
    print("  ✅ Report #" .. data.report_number .. " completed!")
end)

-- Test cascade system
print("\nTesting event cascade system:")
local test_data = {
    {type = "temperature", value = 23.5},
    {type = "pressure", value = "101.3"},  -- String, will need correction
    {type = "humidity", value = 65},
    {type = "wind", value = nil},  -- Invalid
    {type = "rainfall", value = 12.7}
}

for _, data_point in ipairs(test_data) do
    cascade_system:emit("data_received", data_point)
    utils.general_sleep(0.3)
end

print()

-- Example 5: Async Event Processing
print("=== Example 5: Async Event Processing ===")

local async_system = EventSystem:new()

-- Register async handlers
async_system:on("analyze_text", function(data)
    local analyzer = agent.create("Text Analyzer", {
        model = model,
        system = "Analyze text sentiment and key themes.",
        temperature = 0.4
    })
    
    local analysis = analyzer:run("Analyze: " .. data.text)
    
    async_system:emit("analysis_complete", {
        id = data.id,
        text = data.text,
        analysis = analysis
    })
end, {async = true})

async_system:on("analysis_complete", function(data)
    print("  ✅ Analysis complete for ID: " .. data.id)
    -- Would save analysis to: output_dir .. "/analysis_" .. data.id .. ".txt"
    -- Content: text and analysis results
end)

-- Submit multiple texts for async processing
print("\nSubmitting texts for async analysis:")
local texts = {
    {id = "001", text = "The future of AI looks bright and promising."},
    {id = "002", text = "Climate change requires urgent action."},
    {id = "003", text = "Technology transforms how we communicate."}
}

for _, text_data in ipairs(texts) do
    print("  📤 Submitting: " .. text_data.id)
    async_system:emit("analyze_text", text_data)
end

-- Process async queue
print("  ⏳ Processing async queue...")
async_system:process_async_queue()
utils.general_sleep(2)  -- Wait for async processing

print()

-- Example 6: Event Filtering and Transformation
print("=== Example 6: Event Filtering & Transformation ===")

local filtered_system = EventSystem:new()

-- Add filter: only process high-priority events
filtered_system:add_filter("task_submitted", function(data)
    return data.priority and data.priority >= 7
end)

-- Add transformer: enrich event data
filtered_system:add_transformer("task_submitted", function(data)
    data.timestamp = os.time()
    data.enhanced = true
    data.estimated_duration = data.priority * 10  -- Higher priority = longer tasks
    return data
end)

-- Register handler
filtered_system:on("task_submitted", function(data)
    print(string.format(
        "  🎯 High-priority task: %s (Priority: %d, Est. Duration: %d min)",
        data.name, data.priority, data.estimated_duration
    ))
end)

-- Test filtering
print("\nTesting event filtering (only priority >= 7 will be processed):")
local tasks = {
    {name = "Critical bug fix", priority = 10},
    {name = "Feature request", priority = 5},
    {name = "Security patch", priority = 9},
    {name = "Documentation", priority = 3},
    {name = "Performance optimization", priority = 7}
}

for _, task in ipairs(tasks) do
    print("  Submitting: " .. task.name .. " (Priority: " .. task.priority .. ")")
    filtered_system:emit("task_submitted", task)
end

print()

-- Summary and Statistics
print("=== Event System Statistics ===")

print("\nBasic System Stats:")
local basic_stats = basic_events:get_event_stats()
for event, count in pairs(basic_stats) do
    print("  " .. event .. ": " .. count .. " times")
end

print("\nModeration System Stats:")
print("  Warnings issued: " .. table_count(moderation_state.user_warnings))
print("  Users banned: " .. table_count(moderation_state.banned_users))
print("  Content flagged: " .. #moderation_state.flagged_content)

print("\nCascade System Stats:")
print("  Reports generated: " .. cascade_state.reports_generated)
print("  Data points processed: " .. #cascade_state.data_points)

print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Basic event emission and handling")
print("2. Event-driven content moderation")
print("3. State machines with events")
print("4. Event cascades and chains")
print("5. Asynchronous event processing")
print("6. Event filtering and transformation")
print()
print("Key insights:")
print("- Events enable decoupled, reactive systems")
print("- State machines provide structure to event flows")
print("- Cascades allow complex multi-step processes")
print("- Async processing improves performance")
print("- Filtering and transformation add flexibility")
print()

-- In production, files would be created in output directory
print("Event processing complete - events handled asynchronously")

-- Return summary
return {
    event_systems_created = 6,
    total_events_emitted = #basic_events.event_log + #moderation_system.event_log + 
                          #cascade_system.event_log + #async_system.event_log + 
                          #filtered_system.event_log,
    patterns_demonstrated = {
        "basic_events",
        "content_moderation",
        "state_machine",
        "event_cascades",
        "async_processing",
        "filtering_transformation"
    },
    files_generated = files and #files or 0
}