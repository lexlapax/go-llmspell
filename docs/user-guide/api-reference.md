# Complete API Reference

This is the comprehensive API reference for all modules and functions available in go-llmspell Lua scripts.

## Table of Contents

- [Global Objects](#global-objects)
- [Core Module](#core-module)
- [LLM Module](#llm-module)
- [Agent Module](#agent-module)
- [Tools Module](#tools-module)
- [State Module](#state-module)
- [Events Module](#events-module)
- [Logging Module](#logging-module)
- [Promise Module](#promise-module)
- [Data Module](#data-module)
- [Errors Module](#errors-module)
- [Auth Module](#auth-module)
- [Observability Module](#observability-module)
- [Spell Module](#spell-module)
- [Testing Module](#testing-module)

---

## Global Objects

These objects are available in all Lua spells without requiring imports.

### `params`
**Type**: `table`  
**Description**: Input parameters passed to the spell from command line or caller.

```lua
-- Access parameters
local name = params.name or "default"
local count = tonumber(params.count) or 1
```

### `llm`
**Type**: `table`  
**Description**: Global LLM operations object (alias for require("llm")).

### `log`
**Type**: `table`  
**Description**: Global logging object (alias for require("logging")).

### `spell`
**Type**: `table`  
**Description**: Global spell utilities object (alias for require("spell")).

---

## Core Module

The core module provides fundamental async operations and utilities.

### Functions

#### `core.async(fn)`
Runs a function asynchronously in a new coroutine.

**Parameters:**
- `fn` (function): The function to run asynchronously

**Returns:**
- `string`: Coroutine ID

**Example:**
```lua
local core = require("core")

core.async(function()
    -- Async operations here
    core.sleep(1)
    print("Async operation completed")
end)
```

#### `core.sleep(seconds)`
Pauses execution for the specified duration.

**Parameters:**
- `seconds` (number): Duration to sleep in seconds

**Example:**
```lua
core.sleep(0.5)  -- Sleep for 500ms
```

#### `core.yield()`
Yields control to allow other coroutines to run.

**Example:**
```lua
for i = 1, 1000 do
    -- Process item
    if i % 100 == 0 then
        core.yield()  -- Give other coroutines a chance
    end
end
```

#### `core.channel(capacity)`
Creates a new channel for coroutine communication.

**Parameters:**
- `capacity` (number, optional): Buffer capacity (default: 0 for unbuffered)

**Returns:**
- `table`: Channel object with send() and receive() methods

**Example:**
```lua
local ch = core.channel(10)

-- Send value
ch:send("hello")

-- Receive value
local value = ch:receive()
```

#### `core.select(cases)`
Performs a select operation on multiple channels.

**Parameters:**
- `cases` (table): Array of case objects

**Returns:**
- `number`: Index of the selected case
- `any`: Value received (for receive cases)

**Example:**
```lua
local ch1 = core.channel()
local ch2 = core.channel()

local index, value = core.select({
    {channel = ch1, op = "receive"},
    {channel = ch2, op = "receive"},
    {channel = ch1, op = "send", value = "hello"},
    {op = "default"}  -- Non-blocking
})
```

---

## LLM Module

The LLM module provides language model operations.

### Functions

#### `llm.complete(options)`
Performs a completion request to an LLM.

**Parameters:**
- `options` (table): Request options
  - `model` (string, required): Model identifier
  - `messages` (table, required): Array of message objects
  - `temperature` (number, optional): Sampling temperature (0-2)
  - `max_tokens` (number, optional): Maximum tokens to generate
  - `top_p` (number, optional): Top-p sampling parameter
  - `stop` (string[], optional): Stop sequences
  - `stream` (boolean, optional): Enable streaming (use llm.stream instead)

**Returns:**
- `table`: Response object
  - `content` (string): Generated text
  - `usage` (table): Token usage information
  - `model` (string): Model used
  - `finish_reason` (string): Why generation stopped

**Example:**
```lua
local response = llm.complete({
    model = "gpt-4",
    messages = {
        {role = "system", content = "You are a helpful assistant"},
        {role = "user", content = "Explain quantum computing"}
    },
    temperature = 0.7,
    max_tokens = 500
})

print(response.content)
print("Tokens used: " .. response.usage.total_tokens)
```

#### `llm.stream(options)`
Performs a streaming completion request.

**Parameters:**
- `options` (table): Request options (same as complete)
  - `on_content` (function, required): Called with each content chunk
  - `on_complete` (function, optional): Called when streaming completes
  - `on_error` (function, optional): Called on error

**Example:**
```lua
llm.stream({
    model = "gpt-4",
    messages = {{role = "user", content = "Write a story"}},
    on_content = function(chunk)
        io.write(chunk)
        io.flush()
    end,
    on_complete = function(response)
        print("\nTotal tokens: " .. response.usage.total_tokens)
    end,
    on_error = function(err)
        print("Error: " .. err)
    end
})
```

#### `llm.models(provider)`
Lists available models for a provider or all providers.

**Parameters:**
- `provider` (string, optional): Provider name to filter by

**Returns:**
- `table`: Array of model information objects

**Example:**
```lua
-- List all models
local all_models = llm.models()

-- List OpenAI models
local openai_models = llm.models("openai")

for _, model in ipairs(openai_models) do
    print(model.id .. " - " .. model.description)
end
```

#### `llm.structured(options)`
Performs structured output generation with schema validation.

**Parameters:**
- `options` (table): Request options
  - All parameters from `complete()`
  - `schema` (table, required): JSON schema for output structure
  - `strict` (boolean, optional): Enforce strict schema validation

**Returns:**
- `table`: Parsed and validated response object

**Example:**
```lua
local result = llm.structured({
    model = "gpt-4",
    messages = {{role = "user", content = "List 3 programming languages"}},
    schema = {
        type = "object",
        properties = {
            languages = {
                type = "array",
                items = {
                    type = "object",
                    properties = {
                        name = {type = "string"},
                        year = {type = "number"},
                        paradigm = {type = "string"}
                    },
                    required = {"name", "year", "paradigm"}
                }
            }
        },
        required = {"languages"}
    }
})

for _, lang in ipairs(result.languages) do
    print(lang.name .. " (" .. lang.year .. ") - " .. lang.paradigm)
end
```

---

## Agent Module

The agent module provides AI agent creation and management.

### Functions

#### `agent.create(config)`
Creates a new AI agent with specified configuration.

**Parameters:**
- `config` (table): Agent configuration
  - `name` (string, optional): Agent name
  - `model` (string, required): LLM model to use
  - `system` (string, optional): System prompt
  - `tools` (string[], optional): Tool names to enable
  - `temperature` (number, optional): Sampling temperature
  - `max_iterations` (number, optional): Max tool use iterations
  - `state` (table, optional): Shared state object

**Returns:**
- `table`: Agent object

**Example:**
```lua
local agent = require("agent")

local assistant = agent.create({
    name = "Research Assistant",
    model = "gpt-4",
    system = "You are a helpful research assistant",
    tools = {"web_search", "file_read", "file_write"},
    temperature = 0.3,
    max_iterations = 5
})
```

### Agent Methods

#### `agent:run(prompt, context)`
Runs the agent with a prompt.

**Parameters:**
- `prompt` (string): The task or question for the agent
- `context` (table, optional): Additional context

**Returns:**
- `string`: Agent's response

**Example:**
```lua
local result = assistant:run("Research quantum computing breakthroughs in 2024")
```

#### `agent:chat(messages)`
Continues a conversation with the agent.

**Parameters:**
- `messages` (table): Conversation history

**Returns:**
- `table`: Response message

**Example:**
```lua
local response = assistant:chat({
    {role = "user", content = "What did you find?"},
    {role = "assistant", content = previous_response},
    {role = "user", content = "Can you elaborate on the first point?"}
})
```

#### `agent:reset()`
Resets the agent's conversation history and state.

**Example:**
```lua
assistant:reset()
```

#### `agent:get_tools()`
Returns the list of tools available to the agent.

**Returns:**
- `table`: Array of tool names

#### `agent:add_tool(tool_name)`
Adds a tool to the agent's available tools.

**Parameters:**
- `tool_name` (string): Name of the tool to add

#### `agent:remove_tool(tool_name)`
Removes a tool from the agent's available tools.

**Parameters:**
- `tool_name` (string): Name of the tool to remove

---

## Tools Module

The tools module provides access to built-in and custom tools.

### Functions

#### `tools.list()`
Lists all available tools.

**Returns:**
- `table`: Array of tool information objects

**Example:**
```lua
local available = tools.list()
for _, tool in ipairs(available) do
    print(tool.name .. ": " .. tool.description)
end
```

#### `tools.get_info(name)`
Gets detailed information about a tool.

**Parameters:**
- `name` (string): Tool name

**Returns:**
- `table`: Tool information including parameters

**Example:**
```lua
local info = tools.get_info("web_search")
print("Parameters: " .. data.to_json(info.parameters))
```

#### `tools.execute(name, args)`
Executes a tool with given arguments.

**Parameters:**
- `name` (string): Tool name
- `args` (table): Tool arguments

**Returns:**
- `any`: Tool execution result

**Example:**
```lua
local result = tools.execute("calculator", {expression = "sqrt(16) + 3^2"})
```

#### `tools.register(definition)`
Registers a custom tool.

**Parameters:**
- `definition` (table): Tool definition
  - `name` (string, required): Tool name
  - `description` (string, required): Tool description
  - `parameters` (table, required): Parameter schema
  - `execute` (function, required): Execution function

**Example:**
```lua
tools.register({
    name = "word_count",
    description = "Count words in text",
    parameters = {
        text = {type = "string", required = true, description = "Text to count words in"}
    },
    execute = function(args)
        local count = 0
        for word in string.gmatch(args.text, "%S+") do
            count = count + 1
        end
        return {count = count}
    end
})
```

### Built-in Tools

#### File Operations

##### `tools.file_read(path)`
Reads a file's contents.

**Parameters:**
- `path` (string): File path

**Returns:**
- `string`: File contents

##### `tools.file_write(path, content)`
Writes content to a file.

**Parameters:**
- `path` (string): File path
- `content` (string): Content to write

##### `tools.file_append(path, content)`
Appends content to a file.

**Parameters:**
- `path` (string): File path
- `content` (string): Content to append

##### `tools.file_delete(path)`
Deletes a file.

**Parameters:**
- `path` (string): File path

##### `tools.file_exists(path)`
Checks if a file exists.

**Parameters:**
- `path` (string): File path

**Returns:**
- `boolean`: True if file exists

##### `tools.file_list(directory)`
Lists files in a directory.

**Parameters:**
- `directory` (string): Directory path

**Returns:**
- `table`: Array of file names

#### Web Operations

##### `tools.web_fetch(url, options)`
Fetches content from a URL.

**Parameters:**
- `url` (string): URL to fetch
- `options` (table, optional): Request options
  - `method` (string): HTTP method
  - `headers` (table): HTTP headers
  - `body` (string): Request body

**Returns:**
- `table`: Response object

##### `tools.web_search(query, options)`
Performs a web search.

**Parameters:**
- `query` (string): Search query
- `options` (table, optional): Search options
  - `count` (number): Number of results

**Returns:**
- `table`: Search results

#### Date/Time Operations

##### `tools.datetime_now()`
Gets the current date and time.

**Returns:**
- `string`: ISO 8601 formatted timestamp

##### `tools.datetime_format(timestamp, format)`
Formats a timestamp.

**Parameters:**
- `timestamp` (string): ISO 8601 timestamp
- `format` (string): Format string

**Returns:**
- `string`: Formatted date/time

##### `tools.datetime_parse(date_string)`
Parses a date string.

**Parameters:**
- `date_string` (string): Date string to parse

**Returns:**
- `string`: ISO 8601 timestamp

#### Utility Tools

##### `tools.calculator(expression)`
Evaluates a mathematical expression.

**Parameters:**
- `expression` (string): Math expression

**Returns:**
- `number`: Calculation result

##### `tools.json_parse(json_string)`
Parses a JSON string.

**Parameters:**
- `json_string` (string): JSON string

**Returns:**
- `any`: Parsed value

##### `tools.json_stringify(value, pretty)`
Converts a value to JSON.

**Parameters:**
- `value` (any): Value to convert
- `pretty` (boolean, optional): Pretty print

**Returns:**
- `string`: JSON string

---

## State Module

The state module provides state management capabilities.

### Functions

#### `state.create(name, options)`
Creates a new state container.

**Parameters:**
- `name` (string): State container name
- `options` (table, optional): Configuration options
  - `persistent` (boolean): Enable persistence
  - `file` (string): Persistence file path
  - `thread_safe` (boolean): Enable thread safety
  - `defaults` (table): Default values

**Returns:**
- `table`: State container object

**Example:**
```lua
local state = require("state")

local session = state.create("session", {
    persistent = true,
    file = "session_state.json",
    defaults = {
        user = nil,
        messages = {},
        created_at = os.date()
    }
})
```

### State Container Methods

#### `state:get(key, default)`
Gets a value from state.

**Parameters:**
- `key` (string): Key to retrieve
- `default` (any, optional): Default value if key not found

**Returns:**
- `any`: Value or default

**Example:**
```lua
local username = session:get("user.name", "Anonymous")
```

#### `state:set(key, value)`
Sets a value in state.

**Parameters:**
- `key` (string): Key to set (supports dot notation)
- `value` (any): Value to set

**Example:**
```lua
session:set("user.name", "Alice")
session:set("user.preferences.theme", "dark")
```

#### `state:update(key, updater)`
Updates a value using an updater function.

**Parameters:**
- `key` (string): Key to update
- `updater` (function): Function that receives current value and returns new value

**Example:**
```lua
session:update("stats.message_count", function(current)
    return (current or 0) + 1
end)
```

#### `state:delete(key)`
Deletes a key from state.

**Parameters:**
- `key` (string): Key to delete

#### `state:clear()`
Clears all state data.

#### `state:increment(key, amount)`
Increments a numeric value.

**Parameters:**
- `key` (string): Key to increment
- `amount` (number, optional): Amount to increment by (default: 1)

#### `state:append(key, value)`
Appends to an array value.

**Parameters:**
- `key` (string): Key of array
- `value` (any): Value to append

#### `state:save()`
Saves state to persistent storage (if enabled).

#### `state:load()`
Loads state from persistent storage (if enabled).

#### `state:transaction(fn)`
Executes a function within a transaction.

**Parameters:**
- `fn` (function): Function to execute

**Example:**
```lua
session:transaction(function()
    local count = session:get("count", 0)
    session:set("count", count + 1)
    session:set("last_updated", os.date())
end)
```

---

## Events Module

The events module provides event system functionality.

### Functions

#### `events.create_emitter()`
Creates a new event emitter.

**Returns:**
- `table`: Event emitter object

**Example:**
```lua
local events = require("events")
local emitter = events.create_emitter()
```

### Event Emitter Methods

#### `emitter:on(event, handler)`
Registers an event handler.

**Parameters:**
- `event` (string): Event name
- `handler` (function): Handler function

**Returns:**
- `string`: Handler ID

**Example:**
```lua
local handler_id = emitter:on("message", function(data)
    print("Message received: " .. data.text)
end)
```

#### `emitter:once(event, handler)`
Registers a one-time event handler.

**Parameters:**
- `event` (string): Event name
- `handler` (function): Handler function

**Returns:**
- `string`: Handler ID

#### `emitter:emit(event, data)`
Emits an event.

**Parameters:**
- `event` (string): Event name
- `data` (any, optional): Event data

**Example:**
```lua
emitter:emit("message", {text = "Hello", user = "Alice"})
```

#### `emitter:off(event, handler_id)`
Removes an event handler.

**Parameters:**
- `event` (string): Event name
- `handler_id` (string): Handler ID to remove

#### `emitter:remove_all_listeners(event)`
Removes all handlers for an event.

**Parameters:**
- `event` (string, optional): Event name (all events if not specified)

### Global Events

#### `events.global`
Global event emitter instance.

**Common Events:**
- `agent.started`: Agent execution started
- `agent.completed`: Agent execution completed
- `tool.executed`: Tool was executed
- `state.changed`: State was modified
- `error.occurred`: An error occurred

**Example:**
```lua
events.global:on("tool.executed", function(data)
    log.debug("Tool executed", {
        tool = data.tool,
        duration = data.duration,
        success = data.success
    })
end)
```

---

## Logging Module

The logging module provides structured logging capabilities.

### Functions

#### `log.debug(message, data)`
Logs a debug message.

**Parameters:**
- `message` (string): Log message
- `data` (table, optional): Structured data

#### `log.info(message, data)`
Logs an info message.

**Parameters:**
- `message` (string): Log message
- `data` (table, optional): Structured data

#### `log.warn(message, data)`
Logs a warning message.

**Parameters:**
- `message` (string): Log message
- `data` (table, optional): Structured data

#### `log.error(message, data)`
Logs an error message.

**Parameters:**
- `message` (string): Log message
- `data` (table, optional): Structured data

#### `log.with_fields(fields)`
Creates a logger with persistent fields.

**Parameters:**
- `fields` (table): Fields to include in all log messages

**Returns:**
- `table`: Logger instance with the same methods

**Example:**
```lua
local logger = log.with_fields({
    component = "data_processor",
    version = "1.0"
})

logger.info("Processing started", {items = 100})
-- Logs with component and version fields automatically included
```

#### `log.set_level(level)`
Sets the minimum log level.

**Parameters:**
- `level` (string): Log level ("debug", "info", "warn", "error")

---

## Promise Module

The promise module provides async/await-style programming.

### Functions

#### `promise.new(executor)`
Creates a new promise.

**Parameters:**
- `executor` (function): Executor function that receives resolve and reject callbacks

**Returns:**
- `table`: Promise object

**Example:**
```lua
local promise = require("promise")

local p = promise.new(function(resolve, reject)
    core.async(function()
        local success, result = pcall(some_async_operation)
        if success then
            resolve(result)
        else
            reject(result)
        end
    end)
end)
```

### Promise Methods

#### `promise:next(on_fulfilled, on_rejected)`
Chains promise handlers (then in JavaScript).

**Parameters:**
- `on_fulfilled` (function, optional): Success handler
- `on_rejected` (function, optional): Error handler

**Returns:**
- `table`: New promise

**Example:**
```lua
p:next(function(value)
    print("Success: " .. value)
    return value * 2
end):next(function(value)
    print("Doubled: " .. value)
end)
```

#### `promise:catch(on_rejected)`
Handles promise rejection.

**Parameters:**
- `on_rejected` (function): Error handler

**Returns:**
- `table`: New promise

**Example:**
```lua
p:catch(function(error)
    log.error("Operation failed", {error = error})
end)
```

#### `promise:finally(on_finally)`
Executes cleanup regardless of outcome.

**Parameters:**
- `on_finally` (function): Cleanup function

**Returns:**
- `table`: New promise

#### `promise:await()`
Waits for promise resolution synchronously.

**Returns:**
- `any`: Resolved value

**Throws:**
- Error if promise is rejected

**Example:**
```lua
local result = p:await()
```

### Static Methods

#### `promise.resolve(value)`
Creates a resolved promise.

**Parameters:**
- `value` (any): Resolution value

**Returns:**
- `table`: Resolved promise

#### `promise.reject(reason)`
Creates a rejected promise.

**Parameters:**
- `reason` (any): Rejection reason

**Returns:**
- `table`: Rejected promise

#### `promise.all(promises)`
Waits for all promises to resolve.

**Parameters:**
- `promises` (table): Array of promises

**Returns:**
- `table`: Promise that resolves to array of results

**Example:**
```lua
local results = promise.all({p1, p2, p3}):await()
```

#### `promise.race(promises)`
Resolves/rejects with the first settled promise.

**Parameters:**
- `promises` (table): Array of promises

**Returns:**
- `table`: Promise that settles with first result

#### `promise.all_settled(promises)`
Waits for all promises to settle (resolve or reject).

**Parameters:**
- `promises` (table): Array of promises

**Returns:**
- `table`: Promise that resolves to array of settlement objects

#### `promise.timeout(seconds)`
Creates a promise that rejects after timeout.

**Parameters:**
- `seconds` (number): Timeout duration

**Returns:**
- `table`: Promise that rejects on timeout

---

## Data Module

The data module provides data manipulation utilities.

### Functions

#### `data.to_json(value, options)`
Converts a Lua value to JSON string.

**Parameters:**
- `value` (any): Value to convert
- `options` (table, optional): Conversion options
  - `pretty` (boolean): Pretty print
  - `indent` (string): Indentation string

**Returns:**
- `string`: JSON string

**Example:**
```lua
local data = require("data")

local json = data.to_json({
    name = "test",
    values = {1, 2, 3}
}, {pretty = true})
```

#### `data.from_json(json_string)`
Parses a JSON string to Lua value.

**Parameters:**
- `json_string` (string): JSON string

**Returns:**
- `any`: Parsed value

#### `data.to_yaml(value)`
Converts a Lua value to YAML string.

**Parameters:**
- `value` (any): Value to convert

**Returns:**
- `string`: YAML string

#### `data.from_yaml(yaml_string)`
Parses a YAML string to Lua value.

**Parameters:**
- `yaml_string` (string): YAML string

**Returns:**
- `any`: Parsed value

#### `data.to_xml(value, root_name)`
Converts a Lua value to XML string.

**Parameters:**
- `value` (any): Value to convert
- `root_name` (string, optional): Root element name

**Returns:**
- `string`: XML string

#### `data.from_xml(xml_string)`
Parses an XML string to Lua value.

**Parameters:**
- `xml_string` (string): XML string

**Returns:**
- `any`: Parsed value

#### `data.merge(target, source, deep)`
Merges two tables.

**Parameters:**
- `target` (table): Target table
- `source` (table): Source table
- `deep` (boolean, optional): Deep merge

**Returns:**
- `table`: Merged table (target is modified)

#### `data.clone(value, deep)`
Clones a value.

**Parameters:**
- `value` (any): Value to clone
- `deep` (boolean, optional): Deep clone

**Returns:**
- `any`: Cloned value

#### `data.filter(table, predicate)`
Filters a table based on predicate.

**Parameters:**
- `table` (table): Table to filter
- `predicate` (function): Filter function

**Returns:**
- `table`: Filtered table

#### `data.map(table, transform)`
Maps a table using transform function.

**Parameters:**
- `table` (table): Table to map
- `transform` (function): Transform function

**Returns:**
- `table`: Mapped table

#### `data.reduce(table, reducer, initial)`
Reduces a table to a single value.

**Parameters:**
- `table` (table): Table to reduce
- `reducer` (function): Reducer function
- `initial` (any, optional): Initial value

**Returns:**
- `any`: Reduced value

---

## Errors Module

The errors module provides enhanced error handling.

### Functions

#### `errors.new(code, message)`
Creates a new error object.

**Parameters:**
- `code` (string): Error code
- `message` (string): Error message

**Returns:**
- `table`: Error object

**Example:**
```lua
local errors = require("errors")

local err = errors.new("VALIDATION_ERROR", "Invalid input format")
err:add_detail("field", "email")
err:add_detail("value", user_input)
```

### Error Methods

#### `error:add_detail(key, value)`
Adds detail to the error.

**Parameters:**
- `key` (string): Detail key
- `value` (any): Detail value

#### `error:get_detail(key)`
Gets error detail.

**Parameters:**
- `key` (string): Detail key

**Returns:**
- `any`: Detail value

#### `error:to_string()`
Converts error to string representation.

**Returns:**
- `string`: Error string

#### `error:wrap(inner_error)`
Wraps another error.

**Parameters:**
- `inner_error` (any): Error to wrap

### Utility Functions

#### `errors.is_type(error, type)`
Checks if an error is of a specific type.

**Parameters:**
- `error` (any): Error to check
- `type` (string): Error type/code

**Returns:**
- `boolean`: True if error matches type

#### `errors.catch(fn, handler)`
Executes function with error handling.

**Parameters:**
- `fn` (function): Function to execute
- `handler` (function): Error handler

**Returns:**
- `any`: Function result or handler result

**Example:**
```lua
local result = errors.catch(
    function()
        return risky_operation()
    end,
    function(err)
        log.error("Operation failed", {error = err})
        return default_value
    end
)
```

---

## Auth Module

The auth module provides authentication utilities.

### Functions

#### `auth.oauth2(config)`
Creates an OAuth2 authentication flow.

**Parameters:**
- `config` (table): OAuth2 configuration
  - `client_id` (string): Client ID
  - `client_secret` (string): Client secret
  - `auth_url` (string): Authorization URL
  - `token_url` (string): Token URL
  - `redirect_uri` (string): Redirect URI
  - `scopes` (string[]): Required scopes

**Returns:**
- `table`: OAuth2 handler

#### `auth.api_key(key)`
Creates API key authentication.

**Parameters:**
- `key` (string): API key

**Returns:**
- `table`: API key handler

#### `auth.bearer(token)`
Creates bearer token authentication.

**Parameters:**
- `token` (string): Bearer token

**Returns:**
- `table`: Bearer token handler

#### `auth.basic(username, password)`
Creates basic authentication.

**Parameters:**
- `username` (string): Username
- `password` (string): Password

**Returns:**
- `table`: Basic auth handler

---

## Observability Module

The observability module provides monitoring and tracing.

### Functions

#### `observability.metrics`
Access to metrics functionality.

##### `observability.metrics.counter(name, tags)`
Creates a counter metric.

**Parameters:**
- `name` (string): Metric name
- `tags` (table, optional): Metric tags

**Returns:**
- `table`: Counter object with increment() method

##### `observability.metrics.gauge(name, tags)`
Creates a gauge metric.

**Parameters:**
- `name` (string): Metric name
- `tags` (table, optional): Metric tags

**Returns:**
- `table`: Gauge object with set() method

##### `observability.metrics.histogram(name, tags)`
Creates a histogram metric.

**Parameters:**
- `name` (string): Metric name
- `tags` (table, optional): Metric tags

**Returns:**
- `table`: Histogram object with observe() method

#### `observability.tracing`
Access to tracing functionality.

##### `observability.tracing.start_span(name, options)`
Starts a new trace span.

**Parameters:**
- `name` (string): Span name
- `options` (table, optional): Span options

**Returns:**
- `table`: Span object

**Example:**
```lua
local span = observability.tracing.start_span("process_data", {
    attributes = {
        data_size = #data,
        source = "api"
    }
})

-- Do work...

span:add_event("processing_complete", {
    items_processed = 100
})

span:finish()
```

#### `observability.guardrails`
Access to guardrail functionality.

##### `observability.guardrails.content_filter(config)`
Creates a content filter guardrail.

**Parameters:**
- `config` (table): Filter configuration

**Returns:**
- `table`: Guardrail object

##### `observability.guardrails.rate_limiter(config)`
Creates a rate limiting guardrail.

**Parameters:**
- `config` (table): Rate limit configuration

**Returns:**
- `table`: Guardrail object

---

## Spell Module

The spell module provides spell management utilities.

### Functions

#### `spell.run(path, params)`
Runs another spell file.

**Parameters:**
- `path` (string): Path to spell file
- `params` (table, optional): Parameters to pass

**Returns:**
- `any`: Spell return value

**Example:**
```lua
local result = spell.run("helpers/data_processor.lua", {
    input_file = "data.csv",
    output_format = "json"
})
```

#### `spell.info()`
Gets information about the current spell.

**Returns:**
- `table`: Spell information
  - `path` (string): Spell file path
  - `name` (string): Spell name
  - `params` (table): Input parameters
  - `start_time` (number): Start timestamp

#### `spell.require(module_path)`
Requires a Lua module relative to spell directory.

**Parameters:**
- `module_path` (string): Module path

**Returns:**
- `any`: Module exports

#### `spell.exit(code, message)`
Exits the spell with code and message.

**Parameters:**
- `code` (number): Exit code (0 for success)
- `message` (string, optional): Exit message

---

## Testing Module

The testing module provides utilities for testing spells.

### Functions

#### `testing.describe(name, fn)`
Creates a test suite.

**Parameters:**
- `name` (string): Suite name
- `fn` (function): Suite function

**Example:**
```lua
local testing = require("testing")

testing.describe("Data Processor", function()
    testing.it("should parse JSON", function()
        local result = parse_json('{"key": "value"}')
        testing.expect(result.key):to_equal("value")
    end)
    
    testing.it("should handle errors", function()
        testing.expect(function()
            parse_json("invalid json")
        end):to_throw("Invalid JSON")
    end)
end)
```

#### `testing.it(name, fn)`
Creates a test case.

**Parameters:**
- `name` (string): Test name
- `fn` (function): Test function

#### `testing.before_each(fn)`
Registers a function to run before each test.

**Parameters:**
- `fn` (function): Setup function

#### `testing.after_each(fn)`
Registers a function to run after each test.

**Parameters:**
- `fn` (function): Teardown function

#### `testing.expect(value)`
Creates an expectation.

**Parameters:**
- `value` (any): Value to test

**Returns:**
- `table`: Expectation object with matcher methods

### Matchers

- `to_equal(expected)`: Exact equality
- `to_deep_equal(expected)`: Deep equality for tables
- `to_be_truthy()`: Truthy value
- `to_be_falsy()`: Falsy value
- `to_be_nil()`: Nil value
- `to_be_type(type)`: Type check
- `to_contain(value)`: Contains value (string or table)
- `to_match(pattern)`: String pattern match
- `to_throw(message)`: Function throws error
- `to_be_greater_than(value)`: Greater than
- `to_be_less_than(value)`: Less than
- `to_be_close_to(value, precision)`: Numeric closeness

#### `testing.mock(object, method)`
Creates a mock function.

**Parameters:**
- `object` (table, optional): Object to mock method on
- `method` (string, optional): Method name

**Returns:**
- `table`: Mock object with tracking

**Example:**
```lua
local mock = testing.mock()
mock.returns("mocked value")

-- Use mock
local result = mock("arg1", "arg2")

-- Verify
testing.expect(mock.called):to_be_truthy()
testing.expect(mock.call_count):to_equal(1)
testing.expect(mock.calls[1]):to_deep_equal({"arg1", "arg2"})
```

#### `testing.run()`
Runs all registered tests.

**Returns:**
- `table`: Test results

---

## Type Conversions

### Lua to Go Type Mappings

| Lua Type | Go Type | Notes |
|----------|---------|-------|
| nil | nil | |
| boolean | bool | |
| number | float64 | Integers preserved when possible |
| string | string | |
| table | map[string]interface{} or []interface{} | Arrays vs objects detected |
| function | Not supported | Cannot pass Lua functions to Go |
| userdata | Wrapped Go object | Bridge objects |

### Go to Lua Type Mappings

| Go Type | Lua Type | Notes |
|---------|----------|-------|
| nil | nil | |
| bool | boolean | |
| int*, uint*, float* | number | |
| string | string | |
| []T | table | Sequential array |
| map[string]T | table | Object/dictionary |
| struct | table | Field names as keys |
| func | function | Wrapped as Lua function |
| interface{} | Appropriate Lua type | Based on concrete type |

---

## Error Codes

Common error codes used throughout the system:

- `INVALID_ARGUMENT`: Invalid argument provided
- `NOT_FOUND`: Resource not found
- `PERMISSION_DENIED`: Permission denied
- `RATE_LIMIT`: Rate limit exceeded
- `TIMEOUT`: Operation timed out
- `CANCELLED`: Operation cancelled
- `INTERNAL_ERROR`: Internal system error
- `NETWORK_ERROR`: Network-related error
- `VALIDATION_ERROR`: Validation failed
- `TYPE_ERROR`: Type mismatch
- `SCRIPT_ERROR`: Script execution error

---

## Security Profiles

Available security profiles for spell execution:

### sandbox (default)
- No file system access outside working directory
- No network access
- Limited memory (100MB)
- Limited execution time (30s)
- No system command execution

### development
- Full file system access
- Network access allowed
- Higher memory limit (1GB)
- Longer execution time (5m)
- System commands restricted

### production
- Configured file system access
- Configured network access
- Custom resource limits
- Custom timeout
- Audit logging enabled

---

## Environment Variables

Environment variables available in spells:

- `LLMSPELL_DEBUG`: Debug mode enabled
- `LLMSPELL_PROFILE`: Active security profile
- `LLMSPELL_CONFIG`: Configuration file path
- `OPENAI_API_KEY`: OpenAI API key
- `ANTHROPIC_API_KEY`: Anthropic API key
- `GEMINI_API_KEY`: Google Gemini API key

---

## Best Practices Summary

1. **Always validate inputs** - Check parameters and external data
2. **Handle errors gracefully** - Use pcall and error objects
3. **Clean up resources** - Close files, cancel operations
4. **Use structured logging** - Include context in log messages
5. **Respect rate limits** - Implement backoff and retry logic
6. **Monitor performance** - Use metrics and tracing
7. **Test your spells** - Use the testing module
8. **Document your code** - Comments and clear naming
9. **Version your spells** - Track changes over time
10. **Security first** - Never expose secrets, validate paths

---

**Navigation**: [Back to User Guide](README.md) | [Lua Spells Guide](lua-spells.md) | [Examples](../../examples/)