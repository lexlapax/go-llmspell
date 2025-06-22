# Common Patterns and Idioms

This guide covers common patterns and idiomatic ways to structure Lua spells in go-llmspell. These patterns help you write more maintainable, efficient, and robust spells.

## Table of Contents

- [Spell Structure Patterns](#spell-structure-patterns)
- [LLM Interaction Patterns](#llm-interaction-patterns)
- [Agent Orchestration Patterns](#agent-orchestration-patterns)
- [Tool Usage Patterns](#tool-usage-patterns)
- [State Management Patterns](#state-management-patterns)
- [Error Handling Patterns](#error-handling-patterns)
- [Async Patterns](#async-patterns)
- [Testing Patterns](#testing-patterns)
- [Performance Patterns](#performance-patterns)
- [Security Patterns](#security-patterns)

## Spell Structure Patterns

### The Setup-Execute-Cleanup Pattern

```lua
-- spell-template.lua
-- Standard structure for robust spells

-- 1. Setup Phase
local function setup()
    -- Validate parameters
    assert(params.required_param, "required_param is missing")
    
    -- Initialize state
    local state = state.create("spell_state")
    state:set("start_time", os.time())
    
    -- Configure logging
    log.set_level(params.log_level or "info")
    
    return state
end

-- 2. Main Execution
local function execute(state)
    -- Main spell logic here
    local result = llm.complete({
        model = params.model or "gpt-3.5-turbo",
        messages = {{role = "user", content = params.prompt}}
    })
    
    -- Update state
    state:increment("requests_made")
    state:set("last_result", result.content)
    
    return result
end

-- 3. Cleanup Phase
local function cleanup(state, success, result)
    -- Save state if persistent
    if params.save_state then
        state:save()
    end
    
    -- Log completion
    log.info("Spell completed", {
        success = success,
        duration = os.time() - state:get("start_time"),
        requests = state:get("requests_made") or 0
    })
    
    -- Release resources
    state:clear()
end

-- Main spell execution with error handling
local function main()
    local state = setup()
    local success, result = pcall(execute, state)
    cleanup(state, success, result)
    
    if not success then
        error(result)
    end
    
    return result
end

-- Execute spell
return main()
```

### The Module Pattern

```lua
-- spell-modules.lua
-- Organizing complex spells into modules

-- Module: Data Processing
local DataProcessor = {}

function DataProcessor.clean(text)
    -- Remove extra whitespace
    text = text:gsub("%s+", " ")
    -- Trim
    text = text:gsub("^%s*(.-)%s*$", "%1")
    return text
end

function DataProcessor.validate(data)
    assert(type(data) == "string", "Data must be string")
    assert(#data > 0, "Data cannot be empty")
    assert(#data < 10000, "Data too large")
    return true
end

-- Module: LLM Operations
local LLMOps = {}

function LLMOps.summarize(text, options)
    options = options or {}
    return llm.complete({
        model = options.model or "gpt-3.5-turbo",
        messages = {
            {role = "system", content = "You are a helpful summarizer. Be concise."},
            {role = "user", content = "Summarize this text: " .. text}
        },
        temperature = options.temperature or 0.3,
        max_tokens = options.max_tokens or 200
    })
end

function LLMOps.analyze_sentiment(text)
    local response = llm.complete({
        model = "gpt-3.5-turbo",
        messages = {
            {role = "system", content = "Analyze sentiment. Reply with only: positive, negative, or neutral"},
            {role = "user", content = text}
        },
        temperature = 0
    })
    return response.content:lower():trim()
end

-- Use modules
local text = params.text
DataProcessor.validate(text)
text = DataProcessor.clean(text)

local summary = LLMOps.summarize(text)
local sentiment = LLMOps.analyze_sentiment(text)

return {
    summary = summary.content,
    sentiment = sentiment,
    word_count = #text:split(" ")
}
```

## LLM Interaction Patterns

### The Retry with Backoff Pattern

```lua
-- retry-pattern.lua
-- Robust LLM calls with exponential backoff

local function call_llm_with_retry(request, options)
    options = options or {}
    local max_retries = options.max_retries or 3
    local base_delay = options.base_delay or 1
    
    for attempt = 1, max_retries do
        local success, result = pcall(function()
            return llm.complete(request)
        end)
        
        if success then
            return result
        end
        
        -- Check if error is retryable
        local error_msg = tostring(result)
        local is_rate_limit = error_msg:find("rate limit")
        local is_timeout = error_msg:find("timeout")
        
        if not (is_rate_limit or is_timeout) then
            -- Non-retryable error
            error(result)
        end
        
        -- Log retry attempt
        log.warn("LLM call failed, retrying", {
            attempt = attempt,
            error = error_msg,
            next_delay = base_delay * (2 ^ (attempt - 1))
        })
        
        -- Exponential backoff
        if attempt < max_retries then
            core.sleep(base_delay * (2 ^ (attempt - 1)))
        end
    end
    
    error("Max retries exceeded")
end

-- Usage
local response = call_llm_with_retry({
    model = "gpt-4",
    messages = {{role = "user", content = "Hello"}}
})
```

### The Conversation Context Pattern

```lua
-- conversation-pattern.lua
-- Managing conversation history efficiently

local ConversationManager = {}

function ConversationManager:new(options)
    local obj = {
        messages = {},
        max_messages = options.max_messages or 20,
        max_tokens = options.max_tokens or 2000,
        system_prompt = options.system_prompt
    }
    setmetatable(obj, {__index = self})
    
    if obj.system_prompt then
        table.insert(obj.messages, {role = "system", content = obj.system_prompt})
    end
    
    return obj
end

function ConversationManager:add_message(role, content)
    table.insert(self.messages, {role = role, content = content})
    self:trim_history()
end

function ConversationManager:trim_history()
    -- Keep system message + last N messages
    if #self.messages > self.max_messages then
        local system_msg = nil
        if self.messages[1].role == "system" then
            system_msg = table.remove(self.messages, 1)
        end
        
        -- Remove oldest messages
        while #self.messages > self.max_messages - 1 do
            table.remove(self.messages, 1)
        end
        
        -- Restore system message
        if system_msg then
            table.insert(self.messages, 1, system_msg)
        end
    end
end

function ConversationManager:estimate_tokens(text)
    -- Rough estimate: 1 token per 4 characters
    return math.ceil(#text / 4)
end

function ConversationManager:get_messages_for_llm()
    -- Ensure we don't exceed token limit
    local messages = {}
    local total_tokens = 0
    
    -- Always include system message
    if self.messages[1] and self.messages[1].role == "system" then
        messages[1] = self.messages[1]
        total_tokens = self:estimate_tokens(self.messages[1].content)
    end
    
    -- Add messages from newest to oldest until token limit
    for i = #self.messages, 1, -1 do
        local msg = self.messages[i]
        if msg.role ~= "system" then
            local msg_tokens = self:estimate_tokens(msg.content)
            if total_tokens + msg_tokens <= self.max_tokens then
                table.insert(messages, 2, msg)  -- Insert after system message
                total_tokens = total_tokens + msg_tokens
            else
                break
            end
        end
    end
    
    return messages
end

-- Usage
local conversation = ConversationManager:new({
    system_prompt = "You are a helpful assistant",
    max_messages = 10,
    max_tokens = 1500
})

-- Chat loop
while true do
    io.write("You: ")
    local user_input = io.read()
    
    if user_input == "exit" then break end
    
    conversation:add_message("user", user_input)
    
    local response = llm.complete({
        model = "gpt-4",
        messages = conversation:get_messages_for_llm()
    })
    
    conversation:add_message("assistant", response.content)
    print("Assistant: " .. response.content)
end
```

### The Multi-Model Pattern

```lua
-- multi-model-pattern.lua
-- Using multiple models for different tasks

local ModelSelector = {}

-- Model capabilities and costs
ModelSelector.models = {
    ["gpt-3.5-turbo"] = {
        capabilities = {"general", "fast", "cheap"},
        cost_per_1k = 0.002,
        max_tokens = 4096
    },
    ["gpt-4"] = {
        capabilities = {"complex", "reasoning", "accurate"},
        cost_per_1k = 0.03,
        max_tokens = 8192
    },
    ["claude-3-haiku"] = {
        capabilities = {"fast", "cheap", "general"},
        cost_per_1k = 0.0025,
        max_tokens = 100000
    },
    ["claude-3-opus"] = {
        capabilities = {"complex", "creative", "accurate"},
        cost_per_1k = 0.06,
        max_tokens = 100000
    }
}

function ModelSelector.select_model(task_type, constraints)
    constraints = constraints or {}
    
    -- Task-based selection
    local preferred_models = {
        summarization = {"gpt-3.5-turbo", "claude-3-haiku"},
        analysis = {"gpt-4", "claude-3-opus"},
        creative = {"claude-3-opus", "gpt-4"},
        simple = {"gpt-3.5-turbo", "claude-3-haiku"}
    }
    
    local candidates = preferred_models[task_type] or preferred_models.simple
    
    -- Apply constraints
    if constraints.max_cost then
        candidates = table.filter(candidates, function(model)
            return ModelSelector.models[model].cost_per_1k <= constraints.max_cost
        end)
    end
    
    if constraints.min_context then
        candidates = table.filter(candidates, function(model)
            return ModelSelector.models[model].max_tokens >= constraints.min_context
        end)
    end
    
    -- Return first available or default
    return candidates[1] or "gpt-3.5-turbo"
end

-- Usage
local model = ModelSelector.select_model("analysis", {
    max_cost = 0.05,  -- Max $0.05 per 1k tokens
    min_context = 8000 -- Need at least 8k context
})

local response = llm.complete({
    model = model,
    messages = {{role = "user", content = "Analyze this data..."}}
})
```

## Agent Orchestration Patterns

### The Agent Pipeline Pattern

```lua
-- agent-pipeline.lua
-- Chain agents for complex workflows

local AgentPipeline = {}

function AgentPipeline:new()
    local obj = {
        agents = {},
        results = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function AgentPipeline:add_agent(name, config)
    config.name = name
    self.agents[#self.agents + 1] = {
        name = name,
        agent = agent.create(config)
    }
    return self
end

function AgentPipeline:run(initial_input)
    local current_input = initial_input
    
    for _, agent_info in ipairs(self.agents) do
        log.info("Running agent: " .. agent_info.name)
        
        local result = agent_info.agent:run(current_input)
        
        -- Store intermediate result
        self.results[agent_info.name] = result
        
        -- Pass result to next agent
        current_input = result
    end
    
    return {
        final_result = current_input,
        intermediate_results = self.results
    }
end

-- Usage: Research -> Analyze -> Summarize pipeline
local pipeline = AgentPipeline:new()

pipeline:add_agent("researcher", {
    model = "gpt-4",
    system = "You are a research assistant. Find relevant information.",
    tools = {"web_search"},
    temperature = 0.3
})

pipeline:add_agent("analyzer", {
    model = "claude-3-opus",
    system = "You are a data analyst. Analyze the research findings.",
    temperature = 0.2
})

pipeline:add_agent("summarizer", {
    model = "gpt-3.5-turbo",
    system = "You are a summarizer. Create a concise summary.",
    temperature = 0.1
})

local result = pipeline:run("Research recent developments in quantum computing")
print("Final Summary: " .. result.final_result)
```

### The Agent Team Pattern

```lua
-- agent-team.lua
-- Multiple agents working in parallel

local AgentTeam = {}

function AgentTeam:new(config)
    local obj = {
        agents = {},
        coordinator = config.coordinator,
        voting_threshold = config.voting_threshold or 0.5
    }
    setmetatable(obj, {__index = self})
    return obj
end

function AgentTeam:add_member(name, config)
    config.name = name
    self.agents[name] = agent.create(config)
    return self
end

function AgentTeam:discuss(topic, rounds)
    rounds = rounds or 3
    local discussion_history = {}
    
    for round = 1, rounds do
        local round_responses = {}
        
        -- Each agent responds to the topic and previous discussions
        local promises = {}
        for name, agent in pairs(self.agents) do
            promises[name] = promise.new(function(resolve)
                core.async(function()
                    local context = topic
                    if #discussion_history > 0 then
                        context = context .. "\n\nPrevious discussion:\n" .. 
                                 table.concat(discussion_history, "\n")
                    end
                    
                    local response = agent:run(context)
                    resolve({name = name, response = response})
                end)
            end)
        end
        
        -- Wait for all responses
        for name, p in pairs(promises) do
            local result = p:await()
            round_responses[result.name] = result.response
            log.debug("Round " .. round .. " - " .. result.name .. ": " .. result.response)
        end
        
        -- Add to history
        for name, response in pairs(round_responses) do
            table.insert(discussion_history, name .. ": " .. response)
        end
    end
    
    -- Coordinator summarizes
    if self.coordinator then
        local final_context = "Summarize this team discussion:\n" .. 
                            table.concat(discussion_history, "\n")
        return self.coordinator:run(final_context)
    else
        return discussion_history
    end
end

function AgentTeam:vote(proposal)
    local votes = {}
    local yes_count = 0
    
    -- Collect votes in parallel
    local promises = {}
    for name, agent in pairs(self.agents) do
        promises[name] = promise.new(function(resolve)
            core.async(function()
                local prompt = "Vote YES or NO on this proposal: " .. proposal .. 
                             "\nRespond with only YES or NO."
                local response = agent:run(prompt)
                local vote = response:upper():find("YES") and "YES" or "NO"
                resolve({name = name, vote = vote})
            end)
        end)
    end
    
    -- Count votes
    for name, p in pairs(promises) do
        local result = p:await()
        votes[result.name] = result.vote
        if result.vote == "YES" then
            yes_count = yes_count + 1
        end
    end
    
    local total = 0
    for _ in pairs(votes) do total = total + 1 end
    
    return {
        approved = (yes_count / total) >= self.voting_threshold,
        votes = votes,
        yes_percentage = (yes_count / total) * 100
    }
end

-- Usage: Create a team of analysts
local team = AgentTeam:new({
    voting_threshold = 0.6,
    coordinator = agent.create({
        model = "gpt-4",
        system = "You are a team coordinator. Synthesize team discussions."
    })
})

team:add_member("optimist", {
    model = "gpt-3.5-turbo",
    system = "You are an optimistic analyst. Focus on opportunities."
})

team:add_member("pessimist", {
    model = "gpt-3.5-turbo",
    system = "You are a cautious analyst. Focus on risks."
})

team:add_member("realist", {
    model = "gpt-4",
    system = "You are a balanced analyst. Consider all factors."
})

-- Have team discuss
local conclusion = team:discuss("Should we invest in this new technology?", 2)

-- Vote on a proposal
local vote_result = team:vote("Proceed with a pilot program")
print("Vote passed: " .. tostring(vote_result.approved))
```

## Tool Usage Patterns

### The Tool Chain Pattern

```lua
-- tool-chain.lua
-- Chain multiple tools for complex operations

local ToolChain = {}

function ToolChain:new()
    local obj = {
        steps = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function ToolChain:add_step(tool_name, args_fn, error_handler)
    table.insert(self.steps, {
        tool = tool_name,
        args_fn = args_fn or function(prev) return prev end,
        error_handler = error_handler
    })
    return self
end

function ToolChain:execute(initial_input)
    local current_result = initial_input
    local results = {}
    
    for i, step in ipairs(self.steps) do
        local success, result = pcall(function()
            local args = step.args_fn(current_result, results)
            return tools.execute(step.tool, args)
        end)
        
        if not success then
            if step.error_handler then
                result = step.error_handler(result, current_result)
            else
                error("Step " .. i .. " failed: " .. tostring(result))
            end
        end
        
        results[step.tool] = result
        current_result = result
    end
    
    return current_result, results
end

-- Usage: Download -> Process -> Save chain
local chain = ToolChain:new()

chain:add_step("web_fetch", 
    function() return {url = params.url} end
)

chain:add_step("extract_text",
    function(html) return {html = html, selector = "article"} end
)

chain:add_step("summarize",
    function(text) return {text = text, max_length = 500} end
)

chain:add_step("file_write",
    function(summary) 
        return {
            path = "summaries/" .. os.date("%Y%m%d_%H%M%S") .. ".txt",
            content = summary
        }
    end
)

local final_result, all_results = chain:execute()
log.info("Processing complete", {saved_to = final_result})
```

### The Tool Registry Pattern

```lua
-- tool-registry.lua
-- Dynamic tool management and discovery

local ToolRegistry = {}
ToolRegistry.custom_tools = {}

function ToolRegistry.register_tool(definition)
    -- Validate tool definition
    assert(definition.name, "Tool must have a name")
    assert(definition.description, "Tool must have a description")
    assert(definition.execute, "Tool must have an execute function")
    
    -- Create parameter schema if not provided
    if not definition.parameters then
        definition.parameters = {}
    end
    
    -- Register tool
    ToolRegistry.custom_tools[definition.name] = definition
    
    -- Make available to tools module
    tools.register(definition)
    
    log.info("Registered custom tool: " .. definition.name)
end

function ToolRegistry.create_llm_tool(name, prompt_template)
    return {
        name = name,
        description = "LLM-based tool: " .. name,
        parameters = {
            input = {type = "string", required = true}
        },
        execute = function(args)
            local prompt = prompt_template:gsub("{{input}}", args.input)
            local response = llm.complete({
                model = "gpt-3.5-turbo",
                messages = {{role = "user", content = prompt}},
                temperature = 0.3
            })
            return response.content
        end
    }
end

-- Register custom tools
ToolRegistry.register_tool({
    name = "json_validator",
    description = "Validate JSON and provide detailed error messages",
    parameters = {
        json_string = {type = "string", required = true}
    },
    execute = function(args)
        local success, result = pcall(data.from_json, args.json_string)
        if success then
            return {valid = true, data = result}
        else
            return {valid = false, error = tostring(result)}
        end
    end
})

ToolRegistry.register_tool(
    ToolRegistry.create_llm_tool(
        "code_reviewer",
        "Review this code and provide feedback:\n\n{{input}}\n\nProvide constructive feedback."
    )
)

-- Usage
local validation = tools.execute("json_validator", {
    json_string = '{"name": "test", "value": 123}'
})

if validation.valid then
    print("JSON is valid")
else
    print("JSON error: " .. validation.error)
end
```

## State Management Patterns

### The State Machine Pattern

```lua
-- state-machine.lua
-- Implement state machines for complex workflows

local StateMachine = {}

function StateMachine:new(initial_state)
    local obj = {
        current_state = initial_state,
        states = {},
        transitions = {},
        history = {},
        context = state.create("state_machine_context")
    }
    setmetatable(obj, {__index = self})
    return obj
end

function StateMachine:add_state(name, handler)
    self.states[name] = handler
    return self
end

function StateMachine:add_transition(from, to, condition)
    if not self.transitions[from] then
        self.transitions[from] = {}
    end
    table.insert(self.transitions[from], {
        to = to,
        condition = condition or function() return true end
    })
    return self
end

function StateMachine:transition_to(state)
    table.insert(self.history, {
        from = self.current_state,
        to = state,
        timestamp = os.time()
    })
    
    log.info("State transition", {
        from = self.current_state,
        to = state
    })
    
    self.current_state = state
    
    -- Execute state handler
    if self.states[state] then
        self.states[state](self.context)
    end
end

function StateMachine:run()
    while self.current_state ~= "end" do
        -- Check available transitions
        local transitions = self.transitions[self.current_state] or {}
        local transitioned = false
        
        for _, transition in ipairs(transitions) do
            if transition.condition(self.context) then
                self:transition_to(transition.to)
                transitioned = true
                break
            end
        end
        
        if not transitioned then
            error("No valid transition from state: " .. self.current_state)
        end
    end
    
    return self.context
end

-- Usage: Document processing workflow
local workflow = StateMachine:new("start")

workflow:add_state("start", function(ctx)
    ctx:set("document", tools.file_read(params.file))
    ctx:set("status", "loaded")
end)

workflow:add_state("validate", function(ctx)
    local doc = ctx:get("document")
    local is_valid = #doc > 100 and doc:find("BEGIN") and doc:find("END")
    ctx:set("valid", is_valid)
end)

workflow:add_state("process", function(ctx)
    local doc = ctx:get("document")
    local processed = llm.complete({
        model = "gpt-3.5-turbo",
        messages = {{role = "user", content = "Process this document: " .. doc}}
    })
    ctx:set("processed", processed.content)
end)

workflow:add_state("save", function(ctx)
    tools.file_write("output.txt", ctx:get("processed"))
    ctx:set("status", "completed")
end)

workflow:add_state("error", function(ctx)
    log.error("Document processing failed", {
        reason = ctx:get("error_reason")
    })
end)

-- Define transitions
workflow:add_transition("start", "validate")
workflow:add_transition("validate", "process", function(ctx)
    return ctx:get("valid") == true
end)
workflow:add_transition("validate", "error", function(ctx)
    return ctx:get("valid") == false
end)
workflow:add_transition("process", "save")
workflow:add_transition("save", "end")
workflow:add_transition("error", "end")

-- Run workflow
local result = workflow:run()
```

### The Event-Driven State Pattern

```lua
-- event-state.lua
-- State management with event system

local EventState = {}

function EventState:new(name)
    local obj = {
        name = name,
        state = state.create(name),
        listeners = {},
        event_history = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function EventState:on(event, handler)
    if not self.listeners[event] then
        self.listeners[event] = {}
    end
    table.insert(self.listeners[event], handler)
    return self
end

function EventState:emit(event, data)
    -- Record event
    table.insert(self.event_history, {
        event = event,
        data = data,
        timestamp = os.time()
    })
    
    -- Call listeners
    local handlers = self.listeners[event] or {}
    for _, handler in ipairs(handlers) do
        local success, err = pcall(handler, data, self.state)
        if not success then
            log.error("Event handler error", {
                event = event,
                error = err
            })
        end
    end
    
    -- Emit to global event system
    events.emit(self.name .. ":" .. event, data)
end

function EventState:set(key, value)
    local old_value = self.state:get(key)
    self.state:set(key, value)
    
    self:emit("changed:" .. key, {
        key = key,
        old_value = old_value,
        new_value = value
    })
end

function EventState:get(key)
    return self.state:get(key)
end

-- Usage: Shopping cart with events
local cart = EventState:new("shopping_cart")

-- Set up event handlers
cart:on("item_added", function(data, state)
    local items = state:get("items") or {}
    table.insert(items, data.item)
    state:set("items", items)
    
    -- Recalculate total
    local total = 0
    for _, item in ipairs(items) do
        total = total + item.price
    end
    state:set("total", total)
end)

cart:on("changed:total", function(data)
    log.info("Cart total changed", {
        from = data.old_value,
        to = data.new_value
    })
    
    -- Check for discounts
    if data.new_value > 100 then
        cart:emit("discount_eligible", {
            total = data.new_value,
            discount = 0.1
        })
    end
end)

cart:on("discount_eligible", function(data, state)
    state:set("discount", data.discount)
    state:set("final_total", data.total * (1 - data.discount))
end)

-- Use the cart
cart:emit("item_added", {
    item = {name = "Book", price = 29.99}
})

cart:emit("item_added", {
    item = {name = "Laptop", price = 999.99}
})

print("Final total: $" .. cart:get("final_total"))
```

## Error Handling Patterns

### The Circuit Breaker Pattern

```lua
-- circuit-breaker.lua
-- Prevent cascading failures

local CircuitBreaker = {}

function CircuitBreaker:new(config)
    local obj = {
        failure_threshold = config.failure_threshold or 5,
        timeout = config.timeout or 60,
        half_open_requests = config.half_open_requests or 1,
        state = "closed",
        failure_count = 0,
        last_failure_time = 0,
        success_count = 0
    }
    setmetatable(obj, {__index = self})
    return obj
end

function CircuitBreaker:call(fn, ...)
    if self.state == "open" then
        if os.time() - self.last_failure_time >= self.timeout then
            self.state = "half-open"
            self.success_count = 0
        else
            error("Circuit breaker is open")
        end
    end
    
    local success, result = pcall(fn, ...)
    
    if success then
        self:on_success()
        return result
    else
        self:on_failure()
        error(result)
    end
end

function CircuitBreaker:on_success()
    self.failure_count = 0
    
    if self.state == "half-open" then
        self.success_count = self.success_count + 1
        if self.success_count >= self.half_open_requests then
            self.state = "closed"
            log.info("Circuit breaker closed")
        end
    end
end

function CircuitBreaker:on_failure()
    self.failure_count = self.failure_count + 1
    self.last_failure_time = os.time()
    
    if self.failure_count >= self.failure_threshold then
        self.state = "open"
        log.error("Circuit breaker opened", {
            failures = self.failure_count
        })
    end
end

-- Usage: Protect external API calls
local api_breaker = CircuitBreaker:new({
    failure_threshold = 3,
    timeout = 30
})

local function call_external_api(endpoint)
    return api_breaker:call(function()
        return tools.web_fetch(endpoint)
    end)
end

-- Safely call API
for i = 1, 10 do
    local success, result = pcall(call_external_api, "https://api.example.com/data")
    if success then
        print("API call successful")
    else
        print("API call failed: " .. tostring(result))
        core.sleep(5)  -- Wait before retry
    end
end
```

### The Error Aggregation Pattern

```lua
-- error-aggregation.lua
-- Collect and analyze errors

local ErrorCollector = {}

function ErrorCollector:new()
    local obj = {
        errors = {},
        error_counts = {},
        handlers = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function ErrorCollector:add_error(error_type, error_msg, context)
    local error_entry = {
        type = error_type,
        message = error_msg,
        context = context or {},
        timestamp = os.time(),
        stack_trace = debug.traceback()
    }
    
    table.insert(self.errors, error_entry)
    
    -- Update counts
    self.error_counts[error_type] = (self.error_counts[error_type] or 0) + 1
    
    -- Call type-specific handler if exists
    if self.handlers[error_type] then
        self.handlers[error_type](error_entry)
    end
    
    return error_entry
end

function ErrorCollector:register_handler(error_type, handler)
    self.handlers[error_type] = handler
end

function ErrorCollector:get_summary()
    local total_errors = #self.errors
    local error_rate = {}
    
    for error_type, count in pairs(self.error_counts) do
        error_rate[error_type] = {
            count = count,
            percentage = (count / total_errors) * 100
        }
    end
    
    return {
        total = total_errors,
        by_type = error_rate,
        recent = table.slice(self.errors, -10)  -- Last 10 errors
    }
end

function ErrorCollector:wrap(fn, error_type)
    return function(...)
        local success, result = pcall(fn, ...)
        if not success then
            self:add_error(error_type or "general", tostring(result), {
                args = {...}
            })
        end
        return success, result
    end
end

-- Global error collector
local collector = ErrorCollector:new()

-- Register handlers for specific error types
collector:register_handler("rate_limit", function(error)
    log.warn("Rate limit hit, implementing backoff")
    core.sleep(60)  -- Wait 1 minute
end)

collector:register_handler("api_error", function(error)
    -- Send alert for API errors
    if collector.error_counts.api_error > 10 then
        log.error("High API error rate detected!")
    end
end)

-- Wrap functions with error collection
local safe_llm_call = collector:wrap(function(messages)
    return llm.complete({
        model = "gpt-4",
        messages = messages
    })
end, "llm_error")

-- Use wrapped function
local success, response = safe_llm_call({
    {role = "user", content = "Hello"}
})

-- Get error summary
local summary = collector:get_summary()
log.info("Error summary", summary)
```

## Async Patterns

### The Promise Pool Pattern

```lua
-- promise-pool.lua
-- Manage concurrent promises with limits

local PromisePool = {}

function PromisePool:new(config)
    local obj = {
        max_concurrent = config.max_concurrent or 5,
        running = 0,
        queue = {},
        results = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function PromisePool:add(fn, id)
    table.insert(self.queue, {fn = fn, id = id or #self.queue + 1})
    return self
end

function PromisePool:run()
    local all_promises = {}
    
    while #self.queue > 0 or self.running > 0 do
        -- Start new promises up to limit
        while self.running < self.max_concurrent and #self.queue > 0 do
            local task = table.remove(self.queue, 1)
            self.running = self.running + 1
            
            local p = promise.new(function(resolve, reject)
                core.async(function()
                    local success, result = pcall(task.fn)
                    self.running = self.running - 1
                    
                    if success then
                        self.results[task.id] = result
                        resolve(result)
                    else
                        reject(result)
                    end
                end)
            end)
            
            table.insert(all_promises, {promise = p, id = task.id})
        end
        
        -- Wait a bit for promises to complete
        core.yield()
    end
    
    -- Wait for all to complete
    for _, p_info in ipairs(all_promises) do
        p_info.promise:await()
    end
    
    return self.results
end

-- Usage: Process multiple items with concurrency limit
local pool = PromisePool:new({max_concurrent = 3})

-- Add 10 tasks
for i = 1, 10 do
    pool:add(function()
        log.info("Processing item " .. i)
        local response = llm.complete({
            model = "gpt-3.5-turbo",
            messages = {{role = "user", content = "Generate a random fact #" .. i}}
        })
        return response.content
    end, "task_" .. i)
end

-- Run all tasks
local results = pool:run()

-- Display results
for id, result in pairs(results) do
    print(id .. ": " .. result)
end
```

### The Async Iterator Pattern

```lua
-- async-iterator.lua
-- Process collections asynchronously

local AsyncIterator = {}

function AsyncIterator:new(items)
    local obj = {
        items = items,
        index = 0
    }
    setmetatable(obj, {__index = self})
    return obj
end

function AsyncIterator:map(fn)
    local results = {}
    local promises = {}
    
    for i, item in ipairs(self.items) do
        promises[i] = promise.new(function(resolve)
            core.async(function()
                local result = fn(item, i)
                resolve(result)
            end)
        end)
    end
    
    -- Wait for all
    results = promise.all(promises):await()
    
    return AsyncIterator:new(results)
end

function AsyncIterator:filter(fn)
    local filtered = {}
    
    for i, item in ipairs(self.items) do
        if fn(item, i) then
            table.insert(filtered, item)
        end
    end
    
    return AsyncIterator:new(filtered)
end

function AsyncIterator:reduce(fn, initial)
    local result = initial
    
    for i, item in ipairs(self.items) do
        result = fn(result, item, i)
    end
    
    return result
end

function AsyncIterator:for_each(fn)
    local promises = {}
    
    for i, item in ipairs(self.items) do
        promises[i] = promise.new(function(resolve)
            core.async(function()
                fn(item, i)
                resolve()
            end)
        end)
    end
    
    promise.all(promises):await()
end

function AsyncIterator:chunk(size)
    local chunks = {}
    
    for i = 1, #self.items, size do
        local chunk = {}
        for j = 0, size - 1 do
            if self.items[i + j] then
                table.insert(chunk, self.items[i + j])
            end
        end
        table.insert(chunks, chunk)
    end
    
    return AsyncIterator:new(chunks)
end

function AsyncIterator:to_array()
    return self.items
end

-- Usage: Process URLs in parallel
local urls = {
    "https://example.com/1",
    "https://example.com/2",
    "https://example.com/3",
    "https://example.com/4"
}

local results = AsyncIterator:new(urls)
    :map(function(url)
        return tools.web_fetch(url)
    end)
    :filter(function(content)
        return #content > 1000  -- Only keep substantial pages
    end)
    :map(function(content)
        return llm.complete({
            model = "gpt-3.5-turbo",
            messages = {{role = "user", content = "Summarize: " .. content}}
        }).content
    end)
    :to_array()

print("Processed " .. #results .. " pages")
```

## Testing Patterns

### The Spell Testing Pattern

```lua
-- spell-testing.lua
-- Comprehensive spell testing framework

local SpellTester = {}

function SpellTester:new(spell_path)
    local obj = {
        spell_path = spell_path,
        mocks = {},
        assertions = 0,
        failures = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function SpellTester:mock(module_name, method_name, mock_fn)
    if not self.mocks[module_name] then
        self.mocks[module_name] = {}
    end
    self.mocks[module_name][method_name] = mock_fn
    return self
end

function SpellTester:run_test(test_name, params, expected)
    log.info("Running test: " .. test_name)
    
    -- Apply mocks
    local original_modules = {}
    for module_name, mocks in pairs(self.mocks) do
        original_modules[module_name] = {}
        for method_name, mock_fn in pairs(mocks) do
            original_modules[module_name][method_name] = _G[module_name][method_name]
            _G[module_name][method_name] = mock_fn
        end
    end
    
    -- Run spell
    local success, result = pcall(function()
        return spell.run(self.spell_path, params)
    end)
    
    -- Restore original modules
    for module_name, methods in pairs(original_modules) do
        for method_name, original_fn in pairs(methods) do
            _G[module_name][method_name] = original_fn
        end
    end
    
    -- Check result
    self:assert_equals(expected.success, success, test_name .. " success")
    
    if expected.result then
        self:assert_equals(expected.result, result, test_name .. " result")
    end
    
    if expected.error_pattern and not success then
        self:assert_matches(expected.error_pattern, tostring(result), test_name .. " error")
    end
    
    return success, result
end

function SpellTester:assert_equals(expected, actual, message)
    self.assertions = self.assertions + 1
    if expected ~= actual then
        local failure = {
            message = message,
            expected = expected,
            actual = actual
        }
        table.insert(self.failures, failure)
        log.error("Assertion failed", failure)
    end
end

function SpellTester:assert_matches(pattern, text, message)
    self.assertions = self.assertions + 1
    if not string.find(text, pattern) then
        local failure = {
            message = message,
            pattern = pattern,
            text = text
        }
        table.insert(self.failures, failure)
        log.error("Pattern match failed", failure)
    end
end

function SpellTester:get_report()
    return {
        total_assertions = self.assertions,
        failures = #self.failures,
        passed = self.assertions - #self.failures,
        failure_details = self.failures
    }
end

-- Usage: Test a summarizer spell
local tester = SpellTester:new("summarizer.lua")

-- Mock LLM calls
tester:mock("llm", "complete", function(request)
    return {
        content = "Mocked summary of: " .. request.messages[1].content,
        usage = {total_tokens = 100}
    }
end)

-- Run tests
tester:run_test("basic_summarization", {
    text = "Long text to summarize..."
}, {
    success = true,
    result = "Mocked summary of: Long text to summarize..."
})

tester:run_test("empty_text", {
    text = ""
}, {
    success = false,
    error_pattern = "Text cannot be empty"
})

-- Get test report
local report = tester:get_report()
print(string.format("Tests: %d passed, %d failed", 
    report.passed, report.failures))
```

## Performance Patterns

### The Caching Pattern

```lua
-- caching-pattern.lua
-- Multi-level caching system

local CacheManager = {}

function CacheManager:new(config)
    local obj = {
        memory_cache = {},
        disk_cache_dir = config.disk_cache_dir or "./cache/",
        ttl = config.ttl or 3600,  -- 1 hour default
        max_memory_items = config.max_memory_items or 100
    }
    setmetatable(obj, {__index = self})
    return obj
end

function CacheManager:generate_key(...)
    local parts = {...}
    local key_data = table.concat(parts, ":")
    -- Simple hash function
    local hash = 0
    for i = 1, #key_data do
        hash = ((hash * 31) + string.byte(key_data, i)) % 2147483647
    end
    return tostring(hash)
end

function CacheManager:get(key)
    -- Check memory cache first
    local mem_entry = self.memory_cache[key]
    if mem_entry and os.time() - mem_entry.timestamp < self.ttl then
        log.debug("Cache hit (memory): " .. key)
        return mem_entry.value
    end
    
    -- Check disk cache
    local disk_path = self.disk_cache_dir .. key .. ".json"
    if tools.file_exists(disk_path) then
        local disk_data = data.from_json(tools.file_read(disk_path))
        if os.time() - disk_data.timestamp < self.ttl then
            log.debug("Cache hit (disk): " .. key)
            -- Promote to memory cache
            self:set_memory(key, disk_data.value)
            return disk_data.value
        end
    end
    
    log.debug("Cache miss: " .. key)
    return nil
end

function CacheManager:set(key, value)
    -- Save to memory
    self:set_memory(key, value)
    
    -- Save to disk
    local disk_data = {
        value = value,
        timestamp = os.time()
    }
    tools.file_write(
        self.disk_cache_dir .. key .. ".json",
        data.to_json(disk_data)
    )
end

function CacheManager:set_memory(key, value)
    self.memory_cache[key] = {
        value = value,
        timestamp = os.time()
    }
    
    -- Evict oldest if over limit
    if self:count_memory() > self.max_memory_items then
        self:evict_oldest()
    end
end

function CacheManager:count_memory()
    local count = 0
    for _ in pairs(self.memory_cache) do
        count = count + 1
    end
    return count
end

function CacheManager:evict_oldest()
    local oldest_key = nil
    local oldest_time = os.time()
    
    for key, entry in pairs(self.memory_cache) do
        if entry.timestamp < oldest_time then
            oldest_time = entry.timestamp
            oldest_key = key
        end
    end
    
    if oldest_key then
        self.memory_cache[oldest_key] = nil
        log.debug("Evicted from memory cache: " .. oldest_key)
    end
end

function CacheManager:with_cache(key_parts, fn)
    local key = self:generate_key(table.unpack(key_parts))
    
    -- Try to get from cache
    local cached = self:get(key)
    if cached then
        return cached
    end
    
    -- Compute and cache
    local result = fn()
    self:set(key, result)
    
    return result
end

-- Usage
local cache = CacheManager:new({
    ttl = 1800,  -- 30 minutes
    max_memory_items = 50
})

-- Cache LLM responses
local function get_llm_response(prompt, model)
    return cache:with_cache({model, prompt}, function()
        return llm.complete({
            model = model,
            messages = {{role = "user", content = prompt}}
        }).content
    end)
end

-- First call - computes
local response1 = get_llm_response("What is the capital of France?", "gpt-3.5-turbo")

-- Second call - from cache
local response2 = get_llm_response("What is the capital of France?", "gpt-3.5-turbo")
```

### The Batch Processing Pattern

```lua
-- batch-processing.lua
-- Efficient batch operations

local BatchProcessor = {}

function BatchProcessor:new(config)
    local obj = {
        batch_size = config.batch_size or 10,
        process_fn = config.process_fn,
        flush_interval = config.flush_interval or 5,  -- seconds
        queue = {},
        last_flush = os.time(),
        auto_flush = config.auto_flush ~= false
    }
    setmetatable(obj, {__index = self})
    
    if obj.auto_flush then
        obj:start_auto_flush()
    end
    
    return obj
end

function BatchProcessor:add(item)
    table.insert(self.queue, item)
    
    if #self.queue >= self.batch_size then
        self:flush()
    end
end

function BatchProcessor:flush()
    if #self.queue == 0 then
        return
    end
    
    local batch = {}
    -- Take up to batch_size items
    for i = 1, math.min(self.batch_size, #self.queue) do
        table.insert(batch, table.remove(self.queue, 1))
    end
    
    -- Process batch
    local success, result = pcall(self.process_fn, batch)
    if not success then
        log.error("Batch processing failed", {
            error = result,
            batch_size = #batch
        })
        -- Re-queue failed items
        for _, item in ipairs(batch) do
            table.insert(self.queue, 1, item)
        end
    end
    
    self.last_flush = os.time()
    
    return result
end

function BatchProcessor:start_auto_flush()
    core.async(function()
        while true do
            core.sleep(1)
            if os.time() - self.last_flush >= self.flush_interval then
                self:flush()
            end
        end
    end)
end

function BatchProcessor:wait_completion()
    -- Flush remaining items
    while #self.queue > 0 do
        self:flush()
    end
end

-- Usage: Batch LLM requests
local llm_batcher = BatchProcessor:new({
    batch_size = 5,
    flush_interval = 3,
    process_fn = function(batch)
        log.info("Processing batch of " .. #batch .. " requests")
        
        -- Create promises for parallel processing
        local promises = {}
        for i, item in ipairs(batch) do
            promises[i] = promise.new(function(resolve)
                core.async(function()
                    local response = llm.complete(item.request)
                    resolve({
                        id = item.id,
                        response = response
                    })
                end)
            end)
        end
        
        -- Wait for all
        local results = promise.all(promises):await()
        
        -- Call callbacks
        for _, result in ipairs(results) do
            if batch[result.id].callback then
                batch[result.id].callback(result.response)
            end
        end
    end
})

-- Queue multiple requests
for i = 1, 20 do
    llm_batcher:add({
        id = i,
        request = {
            model = "gpt-3.5-turbo",
            messages = {{role = "user", content = "Tell me fact #" .. i}}
        },
        callback = function(response)
            print("Response " .. i .. ": " .. response.content)
        end
    })
end

-- Wait for all to complete
llm_batcher:wait_completion()
```

## Security Patterns

### The Input Validation Pattern

```lua
-- input-validation.lua
-- Comprehensive input validation

local Validator = {}

Validator.rules = {
    required = function(value)
        return value ~= nil and value ~= "", "Value is required"
    end,
    
    min_length = function(min)
        return function(value)
            return #tostring(value) >= min, 
                   "Value must be at least " .. min .. " characters"
        end
    end,
    
    max_length = function(max)
        return function(value)
            return #tostring(value) <= max, 
                   "Value must be at most " .. max .. " characters"
        end
    end,
    
    pattern = function(pattern, message)
        return function(value)
            return string.match(tostring(value), pattern) ~= nil,
                   message or "Value does not match required pattern"
        end
    end,
    
    enum = function(valid_values)
        return function(value)
            for _, valid in ipairs(valid_values) do
                if value == valid then
                    return true
                end
            end
            return false, "Value must be one of: " .. table.concat(valid_values, ", ")
        end
    end,
    
    custom = function(fn, message)
        return function(value)
            local valid = fn(value)
            return valid, message or "Custom validation failed"
        end
    end
}

function Validator:new()
    local obj = {
        schemas = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function Validator:define_schema(name, schema)
    self.schemas[name] = schema
end

function Validator:validate(schema_name, data)
    local schema = self.schemas[schema_name]
    if not schema then
        error("Unknown schema: " .. schema_name)
    end
    
    local errors = {}
    
    for field, rules in pairs(schema) do
        local value = data[field]
        
        for _, rule in ipairs(rules) do
            local valid, message = rule(value)
            if not valid then
                if not errors[field] then
                    errors[field] = {}
                end
                table.insert(errors[field], message)
            end
        end
    end
    
    if next(errors) then
        return false, errors
    end
    
    return true, nil
end

function Validator:sanitize(data)
    local sanitized = {}
    
    for key, value in pairs(data) do
        if type(value) == "string" then
            -- Remove potential XSS
            value = value:gsub("<[^>]+>", "")
            -- Remove potential SQL injection
            value = value:gsub("[';\"--]", "")
            -- Trim whitespace
            value = value:gsub("^%s*(.-)%s*$", "%1")
        end
        
        sanitized[key] = value
    end
    
    return sanitized
end

-- Usage
local validator = Validator:new()

-- Define validation schema
validator:define_schema("llm_request", {
    prompt = {
        Validator.rules.required,
        Validator.rules.min_length(10),
        Validator.rules.max_length(1000),
        Validator.rules.pattern("^[^<>]+$", "Prompt cannot contain < or >")
    },
    model = {
        Validator.rules.required,
        Validator.rules.enum({"gpt-3.5-turbo", "gpt-4", "claude-3-opus"})
    },
    temperature = {
        Validator.rules.custom(function(v)
            local num = tonumber(v)
            return num and num >= 0 and num <= 2
        end, "Temperature must be between 0 and 2")
    }
})

-- Validate input
local input = validator:sanitize(params)
local valid, errors = validator:validate("llm_request", input)

if not valid then
    error("Validation failed: " .. data.to_json(errors))
end

-- Safe to use validated input
local response = llm.complete({
    model = input.model,
    messages = {{role = "user", content = input.prompt}},
    temperature = tonumber(input.temperature) or 0.7
})
```

### The Sandbox Execution Pattern

```lua
-- sandbox-execution.lua
-- Safe execution environment

local Sandbox = {}

function Sandbox:new(config)
    local obj = {
        allowed_globals = config.allowed_globals or {
            "math", "string", "table", "tonumber", "tostring",
            "pairs", "ipairs", "type", "assert", "error"
        },
        max_execution_time = config.max_execution_time or 10,
        max_memory = config.max_memory or 50 * 1024 * 1024,  -- 50MB
        disabled_functions = config.disabled_functions or {
            "load", "loadfile", "loadstring", "dofile",
            "io", "os", "debug", "require"
        }
    }
    setmetatable(obj, {__index = self})
    return obj
end

function Sandbox:create_environment()
    local env = {}
    
    -- Copy allowed globals
    for _, name in ipairs(self.allowed_globals) do
        env[name] = _G[name]
    end
    
    -- Add safe print function
    env.print = function(...)
        local args = {...}
        for i, v in ipairs(args) do
            args[i] = tostring(v)
        end
        log.info("Sandbox output: " .. table.concat(args, " "))
    end
    
    -- Add limited table access
    env.table = {
        insert = table.insert,
        remove = table.remove,
        concat = table.concat,
        sort = table.sort
    }
    
    -- Add safe string functions
    env.string = {
        find = string.find,
        match = string.match,
        gsub = string.gsub,
        sub = string.sub,
        len = string.len,
        upper = string.upper,
        lower = string.lower
    }
    
    return env
end

function Sandbox:execute(code, args)
    local env = self:create_environment()
    
    -- Add arguments to environment
    if args then
        for k, v in pairs(args) do
            env[k] = v
        end
    end
    
    -- Compile code in sandbox
    local fn, err = load(code, "sandboxed", "t", env)
    if not fn then
        error("Compilation error: " .. err)
    end
    
    -- Execute with timeout
    local start_time = os.time()
    local result = nil
    local completed = false
    
    local success, exec_result = pcall(function()
        -- Check execution time in a separate coroutine
        core.async(function()
            while not completed do
                if os.time() - start_time > self.max_execution_time then
                    error("Execution timeout exceeded")
                end
                core.sleep(0.1)
            end
        end)
        
        result = fn()
        completed = true
    end)
    
    if not success then
        error("Execution error: " .. tostring(exec_result))
    end
    
    return result
end

-- Usage: Execute untrusted code safely
local sandbox = Sandbox:new({
    max_execution_time = 5,
    allowed_globals = {"math", "string", "table", "print"}
})

-- Safe execution
local code = [[
    local function factorial(n)
        if n <= 1 then return 1 end
        return n * factorial(n - 1)
    end
    
    print("Factorial of 5 is: " .. factorial(5))
    
    return {
        result = factorial(input_number),
        message = "Calculation complete"
    }
]]

local result = sandbox:execute(code, {input_number = 10})
print("Sandbox returned: " .. data.to_json(result))
```

## Summary

These patterns provide a foundation for building robust, maintainable, and efficient Lua spells. Key takeaways:

1. **Structure spells** with clear setup, execution, and cleanup phases
2. **Handle errors gracefully** with retries, circuit breakers, and validation
3. **Manage state effectively** using state machines and event-driven patterns
4. **Optimize performance** with caching, batching, and async operations
5. **Ensure security** through input validation and sandboxed execution

Remember to:
- Start simple and add complexity only when needed
- Test thoroughly, especially error conditions
- Monitor and log appropriately
- Document your patterns for team consistency
- Consider performance implications of your choices

For more examples and advanced patterns, see the [examples directory](../../examples/spells/).