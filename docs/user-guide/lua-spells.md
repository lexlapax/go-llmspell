# Getting Started with Lua Spells

This guide will help you write your first Lua spells and understand the core concepts of scripting LLM interactions with go-llmspell.

## Table of Contents

- [What is a Lua Spell?](#what-is-a-lua-spell)
- [Your First Spell](#your-first-spell)
- [Running Spells](#running-spells)
- [Core Concepts](#core-concepts)
- [Available Modules](#available-modules)
- [Working with LLMs](#working-with-llms)
- [Using Tools](#using-tools)
- [Managing State](#managing-state)
- [Async Operations](#async-operations)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Next Steps](#next-steps)

## What is a Lua Spell?

A Lua spell is a script that orchestrates AI agents, LLMs, and tools to accomplish tasks. Spells can:

- Generate text using various LLM providers
- Create and manage AI agents with specific capabilities
- Use tools to interact with files, web services, and external systems
- Coordinate complex workflows with multiple steps
- Handle state persistence and event-driven logic

## Your First Spell

Let's start with a simple "Hello, LLM" spell:

```lua
-- hello-llm.lua
-- A simple spell that asks an LLM to greet the user

-- Get the user's name from parameters (or use default)
local name = params.name or "World"

-- Call the LLM to generate a greeting
local response = llm.complete({
    model = "gpt-3.5-turbo",
    messages = {
        {role = "system", content = "You are a friendly assistant who creates personalized greetings."},
        {role = "user", content = "Create a warm greeting for " .. name}
    }
})

-- Print the greeting
print(response.content)

-- Return the result (optional, but useful for testing)
return response.content
```

Save this as `hello-llm.lua` and run it:

```bash
llmspell run hello-llm.lua -p name="Alice"
```

## Running Spells

There are several ways to run Lua spells:

### Command Line

```bash
# Run a spell file
llmspell run my-spell.lua

# Pass parameters
llmspell run my-spell.lua -p key1=value1 -p key2=value2

# With specific security profile
llmspell run my-spell.lua --profile development

# Debug mode
llmspell run my-spell.lua --debug
```

### Interactive REPL

```bash
# Start the REPL
llmspell repl

# In the REPL:
> local response = llm.complete({model = "gpt-4", messages = {{role = "user", content = "Hello!"}}})
> print(response.content)
```

### From Another Spell

```lua
-- Load and run another spell
local result = spell.run("helper-spell.lua", {param1 = "value1"})
```

## Core Concepts

### 1. Global Objects

Lua spells have access to several global objects:

- `llm` - LLM operations (completions, streaming, agents)
- `tools` - Built-in and custom tools
- `state` - State management
- `workflow` - Workflow orchestration
- `events` - Event system
- `log` - Logging utilities
- `params` - Input parameters
- `spell` - Spell utilities

### 2. Parameters

Spells can receive parameters:

```lua
-- Access parameters passed to the spell
local model = params.model or "gpt-3.5-turbo"
local temperature = tonumber(params.temperature) or 0.7

-- Parameters are always strings from CLI, convert as needed
local max_tokens = tonumber(params.max_tokens) or 1000
```

### 3. Return Values

Spells can return values for testing or chaining:

```lua
-- Return a single value
return result

-- Return multiple values
return result, metadata

-- Return a table
return {
    success = true,
    data = result,
    metadata = metadata
}
```

## Available Modules

### Core Modules

```lua
-- LLM operations
local llm = require("llm")

-- Logging
local log = require("logging")

-- State management
local state = require("state")

-- Events
local events = require("events")

-- Tools
local tools = require("tools")

-- Workflow orchestration
local workflow = require("workflow") -- Not yet implemented

-- Agent management
local agent = require("agent")

-- Data utilities
local data = require("data")

-- Error handling
local errors = require("errors")

-- Authentication
local auth = require("auth")

-- Testing utilities (for spell testing)
local testing = require("testing")

-- Observability
local observability = require("observability")
```

### Async Support

```lua
-- Promise-based async operations
local promise = require("promise")

-- Core async utilities
local core = require("core")
```

## Working with LLMs

### Basic Completion

```lua
-- Simple completion
local response = llm.complete({
    model = "gpt-4",
    messages = {
        {role = "user", content = "Explain quantum computing in simple terms"}
    }
})

print(response.content)
```

### With Options

```lua
-- Completion with options
local response = llm.complete({
    model = "claude-3-opus-20240229",
    messages = {
        {role = "system", content = "You are a helpful coding assistant"},
        {role = "user", content = "Write a Python function to calculate factorial"}
    },
    temperature = 0.2,
    max_tokens = 500,
    top_p = 0.9
})
```

### Streaming Responses

```lua
-- Stream responses for real-time output
llm.stream({
    model = "gpt-4",
    messages = {
        {role = "user", content = "Write a short story about a robot"}
    },
    on_content = function(content)
        -- Called for each chunk of content
        io.write(content)
        io.flush()
    end,
    on_complete = function(full_response)
        -- Called when streaming is complete
        print("\n\nTotal tokens: " .. full_response.usage.total_tokens)
    end,
    on_error = function(error)
        print("Error: " .. error)
    end
})
```

### Creating Agents

```lua
-- Create an agent with tools
local researcher = agent.create({
    name = "Research Assistant",
    model = "gpt-4",
    system = "You are a helpful research assistant with access to web search and file operations.",
    tools = {"web_search", "file_read", "file_write"},
    temperature = 0.3
})

-- Use the agent
local result = researcher:run("Research the latest developments in quantum computing and save a summary")
```

### Multiple Providers

```lua
-- Use different providers
local models = {
    "gpt-4",                       -- OpenAI
    "claude-3-opus-20240229",      -- Anthropic
    "gemini-pro",                  -- Google
    "llama2:70b"                   -- Ollama (local)
}

-- Compare responses
for _, model in ipairs(models) do
    local response = llm.complete({
        model = model,
        messages = {{role = "user", content = "What is 2+2?"}}
    })
    print(model .. ": " .. response.content)
end
```

## Using Tools

### Built-in Tools

```lua
-- File operations
local content = tools.file_read("input.txt")
tools.file_write("output.txt", "Processed: " .. content)

-- Web operations
local webpage = tools.web_fetch("https://example.com")
local weather = tools.web_search("current weather in Tokyo")

-- Datetime
local now = tools.datetime_now()
local formatted = tools.datetime_format(now, "YYYY-MM-DD HH:mm:ss")

-- Calculator
local result = tools.calculator("sqrt(16) + 3^2")
```

### Tool Discovery

```lua
-- List available tools
local available_tools = tools.list()
for _, tool in ipairs(available_tools) do
    print(tool.name .. ": " .. tool.description)
end

-- Get tool metadata
local tool_info = tools.get_info("web_search")
print("Parameters: " .. data.to_json(tool_info.parameters))
```

### Custom Tools

```lua
-- Register a custom tool
tools.register({
    name = "word_counter",
    description = "Count words in text",
    parameters = {
        text = {type = "string", required = true}
    },
    execute = function(args)
        local count = 0
        for word in string.gmatch(args.text, "%S+") do
            count = count + 1
        end
        return {word_count = count}
    end
})

-- Use the custom tool
local result = tools.execute("word_counter", {text = "Hello world from Lua"})
print("Word count: " .. result.word_count)
```

## Managing State

### Local State

```lua
-- Create a state container
local session_state = state.create("session")

-- Set values
session_state:set("user_name", "Alice")
session_state:set("conversation_history", {})

-- Get values
local user_name = session_state:get("user_name")
local history = session_state:get("conversation_history")

-- Update nested values
session_state:update("stats", function(current)
    current = current or {messages = 0}
    current.messages = current.messages + 1
    return current
end)
```

### Persistent State

```lua
-- Create persistent state
local app_state = state.create("app_data", {
    persistent = true,
    file = "app_state.json"
})

-- Load existing state
app_state:load()

-- Make changes
app_state:set("last_run", os.date())
app_state:increment("run_count")

-- Save to disk
app_state:save()
```

### Shared State

```lua
-- Create state that can be shared between agents
local shared_state = state.create("shared", {
    thread_safe = true
})

-- Multiple agents can access it
local agent1 = agent.create({name = "Agent 1", state = shared_state})
local agent2 = agent.create({name = "Agent 2", state = shared_state})
```

## Async Operations

### Using Promises

```lua
-- Create a promise for async operation
local p = promise.new(function(resolve, reject)
    -- Simulate async work
    core.async(function()
        core.sleep(1)  -- Sleep for 1 second
        resolve("Operation completed!")
    end)
end)

-- Wait for completion
local result = p:await()
print(result)
```

### Parallel Operations

```lua
-- Run multiple LLM calls in parallel
local promises = {}

for i = 1, 3 do
    promises[i] = promise.new(function(resolve)
        core.async(function()
            local response = llm.complete({
                model = "gpt-3.5-turbo",
                messages = {{role = "user", content = "Generate a random number"}}
            })
            resolve(response.content)
        end)
    end)
end

-- Wait for all to complete
local results = promise.all(promises):await()
for i, result in ipairs(results) do
    print("Result " .. i .. ": " .. result)
end
```

### Async Error Handling

```lua
-- Handle async errors
local p = promise.new(function(resolve, reject)
    core.async(function()
        local success, result = pcall(function()
            -- Some operation that might fail
            error("Something went wrong!")
        end)
        
        if success then
            resolve(result)
        else
            reject(result)
        end
    end)
end)

-- Handle success and failure
p:next(function(result)
    print("Success: " .. result)
end):catch(function(error)
    print("Error: " .. error)
end)
```

## Error Handling

### Basic Error Handling

```lua
-- Use pcall for safe execution
local success, result = pcall(function()
    return llm.complete({
        model = "gpt-4",
        messages = {{role = "user", content = "Hello"}}
    })
end)

if success then
    print(result.content)
else
    log.error("LLM call failed: " .. tostring(result))
end
```

### Custom Error Types

```lua
-- Create custom errors
local errors = require("errors")

-- Create a specific error
local err = errors.new("RATE_LIMIT", "API rate limit exceeded")
err:add_detail("provider", "openai")
err:add_detail("retry_after", 60)

-- Check error type
if errors.is_type(err, "RATE_LIMIT") then
    local retry_after = err:get_detail("retry_after")
    log.info("Waiting " .. retry_after .. " seconds before retry")
    core.sleep(retry_after)
end
```

### Error Recovery

```lua
-- Implement retry logic
local function with_retry(fn, max_attempts)
    max_attempts = max_attempts or 3
    
    for attempt = 1, max_attempts do
        local success, result = pcall(fn)
        
        if success then
            return result
        end
        
        log.warn("Attempt " .. attempt .. " failed: " .. tostring(result))
        
        if attempt < max_attempts then
            -- Exponential backoff
            local delay = math.pow(2, attempt - 1)
            core.sleep(delay)
        end
    end
    
    error("All attempts failed")
end

-- Use with retry
local response = with_retry(function()
    return llm.complete({
        model = "gpt-4",
        messages = {{role = "user", content = "Hello"}}
    })
end)
```

## Best Practices

### 1. Parameter Validation

```lua
-- Always validate parameters
local function validate_params()
    assert(params.api_key, "API key is required")
    assert(params.model, "Model parameter is required")
    
    local temperature = tonumber(params.temperature)
    assert(temperature and temperature >= 0 and temperature <= 2, 
           "Temperature must be between 0 and 2")
end

validate_params()
```

### 2. Resource Cleanup

```lua
-- Use finally blocks for cleanup
local file_handle = nil

local success, result = pcall(function()
    file_handle = io.open("data.txt", "r")
    -- Process file
    return file_handle:read("*all")
end)

-- Always close resources
if file_handle then
    file_handle:close()
end

if not success then
    error(result)
end
```

### 3. Logging

```lua
-- Use structured logging
log.info("Starting spell execution", {
    spell_name = "data_processor",
    model = params.model,
    input_size = #params.input
})

-- Different log levels
log.debug("Detailed information for debugging")
log.info("General information")
log.warn("Warning messages")
log.error("Error messages", {error = err})
```

### 4. State Management

```lua
-- Initialize state with defaults
local state = state.create("app", {
    defaults = {
        settings = {
            model = "gpt-3.5-turbo",
            temperature = 0.7,
            max_retries = 3
        },
        stats = {
            total_requests = 0,
            total_tokens = 0
        }
    }
})
```

### 5. Error Context

```lua
-- Add context to errors
local function process_data(data)
    if not data then
        local err = errors.new("INVALID_INPUT", "Data is required")
        err:add_detail("function", "process_data")
        err:add_detail("timestamp", os.date())
        error(err)
    end
    -- Process data...
end
```

## Common Patterns

### Chat Loop

```lua
-- Interactive chat loop
local conversation = {}

while true do
    io.write("You: ")
    local user_input = io.read()
    
    if user_input == "exit" then
        break
    end
    
    table.insert(conversation, {role = "user", content = user_input})
    
    local response = llm.complete({
        model = "gpt-4",
        messages = conversation
    })
    
    table.insert(conversation, {role = "assistant", content = response.content})
    print("Assistant: " .. response.content)
end
```

### Data Processing Pipeline

```lua
-- Process data through multiple steps
local function pipeline(input_file)
    -- Step 1: Read data
    local raw_data = tools.file_read(input_file)
    
    -- Step 2: Clean data
    local cleaned = llm.complete({
        model = "gpt-3.5-turbo",
        messages = {
            {role = "system", content = "Clean and format the following data"},
            {role = "user", content = raw_data}
        }
    }).content
    
    -- Step 3: Analyze
    local analysis = llm.complete({
        model = "gpt-4",
        messages = {
            {role = "system", content = "Analyze this data and provide insights"},
            {role = "user", content = cleaned}
        }
    }).content
    
    -- Step 4: Save results
    tools.file_write("analysis_" .. os.date("%Y%m%d_%H%M%S") .. ".md", analysis)
    
    return analysis
end
```

### Multi-Agent Collaboration

```lua
-- Create specialized agents
local researcher = agent.create({
    name = "Researcher",
    model = "gpt-4",
    system = "You are a research specialist who gathers information",
    tools = {"web_search", "file_read"}
})

local writer = agent.create({
    name = "Writer", 
    model = "claude-3-opus-20240229",
    system = "You are a professional writer who creates engaging content"
})

local editor = agent.create({
    name = "Editor",
    model = "gpt-4",
    system = "You are an editor who reviews and improves content"
})

-- Collaborate on a task
local topic = params.topic or "artificial intelligence"

-- Research phase
local research = researcher:run("Research recent developments in " .. topic)

-- Writing phase
local draft = writer:run("Write an article about: " .. research)

-- Editing phase
local final = editor:run("Edit and improve this article: " .. draft)

-- Save result
tools.file_write("article_" .. topic:gsub(" ", "_") .. ".md", final)
```

## Debugging Tips

### 1. Use Debug Mode

```bash
# Run with debug output
llmspell run my-spell.lua --debug
```

### 2. Interactive Debugging

```lua
-- Add breakpoints in your code
local function debug_here(context)
    print("=== DEBUG BREAK ===")
    print("Context: " .. tostring(context))
    print("Press Enter to continue...")
    io.read()
end

-- Use in your spell
local result = some_operation()
debug_here({result = result, state = current_state})
```

### 3. Inspect Objects

```lua
-- Pretty print objects
local data = require("data")

local complex_object = {
    name = "test",
    nested = {
        values = {1, 2, 3},
        map = {key = "value"}
    }
}

print(data.to_json(complex_object, {pretty = true}))
```

### 4. Trace Execution

```lua
-- Add execution tracing
local trace_enabled = params.trace == "true"

local function trace(message)
    if trace_enabled then
        log.debug("[TRACE] " .. message)
    end
end

trace("Starting main processing loop")
```

## Performance Tips

### 1. Reuse Agents

```lua
-- Create agent once, use multiple times
local assistant = agent.create({
    model = "gpt-3.5-turbo",
    system = "You are a helpful assistant"
})

-- Reuse for multiple queries
for _, query in ipairs(queries) do
    local response = assistant:run(query)
    -- Process response
end
```

### 2. Batch Operations

```lua
-- Process multiple items efficiently
local items = {"item1", "item2", "item3"}
local results = {}

-- Create promise for each item
local promises = {}
for i, item in ipairs(items) do
    promises[i] = promise.new(function(resolve)
        core.async(function()
            local result = process_item(item)
            resolve(result)
        end)
    end)
end

-- Wait for all to complete
results = promise.all(promises):await()
```

### 3. Cache Results

```lua
-- Simple caching pattern
local cache = {}

local function get_or_compute(key, compute_fn)
    if cache[key] then
        log.debug("Cache hit for: " .. key)
        return cache[key]
    end
    
    log.debug("Cache miss for: " .. key)
    cache[key] = compute_fn()
    return cache[key]
end

-- Use the cache
local result = get_or_compute("expensive_operation", function()
    return llm.complete({
        model = "gpt-4",
        messages = {{role = "user", content = "Complex query"}}
    })
end)
```

## Security Considerations

### 1. API Key Management

```lua
-- Never hardcode API keys
-- Use environment variables or config files
local api_key = os.getenv("OPENAI_API_KEY")
if not api_key then
    error("OPENAI_API_KEY environment variable not set")
end
```

### 2. Input Sanitization

```lua
-- Sanitize user input
local function sanitize_input(input)
    -- Remove potential command injection
    input = input:gsub("[;&|`$]", "")
    
    -- Limit length
    if #input > 1000 then
        input = input:sub(1, 1000)
    end
    
    return input
end

local safe_input = sanitize_input(params.user_input or "")
```

### 3. File Access Control

```lua
-- Validate file paths
local function safe_file_path(path)
    -- Ensure path doesn't escape allowed directory
    local allowed_dir = "./data/"
    local full_path = allowed_dir .. path
    
    -- Normalize path
    full_path = full_path:gsub("%.%.", "")
    
    -- Check if still within allowed directory
    if not full_path:find("^" .. allowed_dir) then
        error("Invalid file path")
    end
    
    return full_path
end
```

## Next Steps

Now that you understand the basics of Lua spells:

1. **Explore the Examples**: Check out the [examples directory](../../examples/spells/) for more complex spells
2. **Learn About Agents**: Read the [Agents Guide](agents.md) to understand agent capabilities
3. **Master Tools**: Explore the [Tools Guide](tools.md) for all available tools
4. **Build Workflows**: Learn about [Workflows](workflows.md) for complex orchestration
5. **API Reference**: Consult the [API Reference](api-reference.md) for detailed documentation

## Troubleshooting

### Common Issues

**LLM calls failing:**
```lua
-- Check API key is set
assert(os.getenv("OPENAI_API_KEY"), "OpenAI API key not set")

-- Verify model name
local valid_models = {"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
assert(table.contains(valid_models, model), "Invalid model: " .. model)
```

**Memory issues:**
```lua
-- Clear large variables when done
large_data = nil
collectgarbage("collect")
```

**Async operations hanging:**
```lua
-- Always use timeouts
local result = promise.race({
    actual_operation(),
    promise.timeout(30)  -- 30 second timeout
}):await()
```

## Examples Repository

Find complete working examples at:
- [Basic Examples](../../examples/spells/hello-llm/)
- [Agent Examples](../../examples/spells/lua-agent/)
- [Tool Examples](../../examples/spells/tool-example/)
- [Async Examples](../../examples/spells/async-llm/)
- [Advanced Patterns](../../examples/spells/)

---

**Next Guide**: [JavaScript Spells Guide](javascript-spells.md) | **Back to**: [User Guide](README.md)