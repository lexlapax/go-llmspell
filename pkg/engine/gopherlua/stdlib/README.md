# go-llmspell Lua Standard Library

A comprehensive, bridge-first Lua standard library for go-llmspell that provides scriptable LLM interactions through clean, idiomatic APIs.

## Philosophy

The go-llmspell Lua standard library follows a **bridge-first architecture** where we wrap existing functionality from the go-llms ecosystem rather than reimplementing features. This ensures consistency, reliability, and compatibility while providing script-friendly APIs.

### Core Principles

1. **Bridge-First Design**: All functionality wraps go-llms capabilities, never reimplements them
2. **Script-Friendly APIs**: Clean, Lua-idiomatic interfaces that feel natural to script authors
3. **Comprehensive Testing**: Each module has extensive Go-based tests ensuring reliability
4. **Async/Promise Support**: Full support for asynchronous operations and Promise patterns
5. **Type Safety**: Parameter validation and type checking throughout
6. **Resource Management**: Automatic cleanup and resource lifecycle management

## Library Structure

The standard library is organized into feature-oriented modules:

### Core Infrastructure
- **`promise.lua`** - Promise/async support with coroutine integration
- **`testing.lua`** - Test utilities and assertion framework
- **`errors.lua`** - Error handling and recovery utilities
- **`logging.lua`** - Structured logging with multiple levels and formatters

### LLM & AI Operations
- **`llm.lua`** - LLM provider abstraction and chat operations
- **`agent.lua`** - AI agent management and lifecycle
- **`tools.lua`** - Tool registration and execution framework

### Data & State Management
- **`data.lua`** - Data transformation and validation utilities
- **`state.lua`** - State persistence and management
- **`auth.lua`** - Authentication and authorization utilities

### Framework & Composition
- **`spell.lua`** - Spell framework for composition and reuse
- **`events.lua`** - Event system and lifecycle hooks
- **`observability.lua`** - Metrics, monitoring, and debugging
- **`core.lua`** - Core utilities (string, table, crypto, time)

## Key Features

### Asynchronous Programming
```lua
local promise = require("promise")

-- Promise-based async operations
local result = promise.new(function(resolve, reject)
    llm.chat("Hello, world!")
        :andThen(resolve)
        :onError(reject)
end)

-- Async/await syntax
local async_function = promise.async(function()
    local response = promise.await(llm.chat("What is 2+2?"))
    return "Answer: " .. response
end)
```

### Spell Framework
```lua
local spell = require("spell")

-- Initialize spell with metadata
spell.init({
    name = "web-summarizer",
    version = "1.0.0",
    params = {
        url = { type = "string", required = true },
        style = { type = "string", default = "brief", enum = {"brief", "detailed"} }
    }
})

-- Compose multiple spells into workflows
local results = spell.compose({
    { name = "fetch", spell = "web-fetcher", params = { url = "$url" } },
    { name = "summarize", spell = "text-summarizer", params = { text = "$fetch.content" } }
})
```

### Type-Safe Parameters
```lua
-- Parameter validation with types and constraints
spell.params("count", { type = "number", min = 1, max = 100, default = 10 })
spell.params("format", { type = "string", enum = {"json", "text", "markdown"} })
spell.params("config", { type = "table", required = true })

local count = spell.params("count")  -- Validated and type-safe
```

### Event-Driven Architecture
```lua
local events = require("events")

-- Subscribe to events
events.on("llm:response", function(data)
    logging.info("LLM response received", data)
end)

-- Emit events with data
events.emit("spell:complete", { 
    spell = "web-summarizer", 
    duration = 2.5,
    tokens = 150 
})

-- Promise-based event waiting
local response = events.wait_for("llm:response", 5000)  -- 5 second timeout
```

### Agent Management
```lua
local agent = require("agent")

-- Create and manage AI agents
local researcher = agent.create({
    name = "research-agent",
    model = "gpt-4",
    system_prompt = "You are a helpful research assistant",
    tools = {"web-search", "document-reader"}
})

-- Agent conversations with context
local result = researcher:chat("Research the history of Lua programming language")
```

## Testing Framework

Each module includes comprehensive testing using our built-in testing framework:

```lua
local testing = require("testing")

testing.describe("Promise operations", function()
    testing.it("should resolve with correct value", function()
        local p = promise.resolve(42)
        local result = promise.await(p)
        testing.assert.equal(result, 42)
    end)
    
    testing.it("should handle async operations", function()
        local p = promise.new(function(resolve)
            promise.sleep(100):andThen(function()
                resolve("completed")
            end)
        end)
        local result = promise.await(p, 1000)
        testing.assert.equal(result, "completed")
    end)
end)
```

## Architecture Integration

### Bridge Pattern
All modules follow the bridge pattern, wrapping go-llms functionality:

```lua
-- llm.lua bridges go-llms LLM providers
function llm.chat(message, options)
    -- Validates parameters and forwards to Go bridge
    return bridge.llm.chat(message, options or {})
end

-- agent.lua bridges go-llms agent management
function agent.create(config)
    -- Type validation then bridge call
    validate_agent_config(config)
    return bridge.agent.create(config)
end
```

### Resource Management
Automatic cleanup and resource lifecycle management:

```lua
-- Resources are automatically tracked and cleaned up
spell.resource("temp-file", {
    data = "/tmp/analysis.json",
    cleanup = function(resource)
        os.remove(resource.data)
    end
})

-- Automatic cleanup on spell completion
spell.cleanup_resources()
```

### Context Isolation
Each spell execution runs in isolated context:

```lua
-- Parent spell context
spell.init({ name = "main-workflow" })

-- Composed spells get isolated contexts
spell.compose({
    { name = "step1", spell = "data-processor" },  -- Isolated context
    { name = "step2", spell = "formatter" }        -- Isolated context
})
```

## Performance Considerations

- **Lazy Loading**: Modules are loaded on-demand to minimize startup time
- **Caching**: Built-in TTL-based caching for expensive operations
- **Memory Management**: Automatic cleanup of resources and intermediate results
- **Concurrent Execution**: Promise-based concurrency for I/O operations

## Error Handling

Comprehensive error handling with recovery mechanisms:

```lua
-- Global error handlers
spell.on_error(function(err)
    logging.error("Spell execution failed", { error = err })
    return { recoverable = true, retry_after = 1000 }
end)

-- Promise error handling
promise.new(function(resolve, reject)
    -- Risky operation
end):onError(function(err)
    logging.warn("Operation failed, using fallback", { error = err })
    return fallback_value
end)
```

## Getting Started

1. **Basic Spell**: Start with a simple spell using the framework
2. **Add Parameters**: Define typed parameters for your spell
3. **Use Libraries**: Incorporate LLM operations, tools, or agents
4. **Compose Workflows**: Combine multiple spells into workflows
5. **Add Testing**: Write tests using the testing framework
6. **Monitor & Debug**: Use observability tools for production monitoring

## Best Practices

1. **Always validate parameters** using the spell framework
2. **Use promises for async operations** to avoid blocking
3. **Implement proper error handling** with recovery strategies
4. **Test thoroughly** using the built-in testing framework
5. **Follow bridge-first principles** - wrap, don't reimplement
6. **Document your spells** with clear descriptions and examples
7. **Use resource management** for cleanup of temporary resources
8. **Leverage composition** to build complex workflows from simple spells

For detailed API documentation, see [API_REFERENCE.md](API_REFERENCE.md).
For practical examples, see [EXAMPLES.md](EXAMPLES.md).