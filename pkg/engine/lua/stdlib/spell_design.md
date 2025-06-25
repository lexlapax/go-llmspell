# Spell Framework Library Design

## Overview
The Spell Framework Library provides a unified API for spell lifecycle management, composition, parameter handling, and execution context management. It standardizes how spells are structured, initialized, and composed while maintaining compatibility with existing spell patterns.

## Design Principles
1. **Backward Compatibility**: Work with existing spell patterns and global objects
2. **Framework-agnostic**: No assumptions about the hosting environment
3. **Composition-focused**: Enable spell reuse and library creation
4. **Type Safety**: Parameter validation and type checking
5. **Context Isolation**: Clean execution environments for composed spells

## Core API Design

### 1. Spell Lifecycle Management
```lua
-- Initialize spell with configuration and metadata
spell.init(config)
-- Example: spell.init({
--   name = "web-summarizer",
--   version = "1.0.0",
--   description = "Summarizes web pages using LLM",
--   author = "User",
--   params = {
--     url = { type = "string", required = true, description = "URL to summarize" },
--     style = { type = "string", default = "brief", enum = {"brief", "detailed", "bullet-points"} }
--   }
-- })

-- Define parameters with validation
spell.params(name, config)
-- Example: spell.params("url", { type = "string", required = true })
-- Example: spell.params("style", { type = "string", default = "brief" })

-- Output results in structured format
spell.output(data, format, metadata)
-- Example: spell.output(summary, "text", { format = "markdown" })
-- Example: spell.output({ result = "success", data = results }, "json")
```

### 2. Spell Composition and Reuse
```lua
-- Include other spells or libraries
spell.include(path_or_name)
-- Example: spell.include("./lib/http-utils.lua")
-- Example: spell.include("stdlib.string-utils")

-- Compose multiple spells into a workflow
spell.compose(spells_config)
-- Example: spell.compose({
--   { name = "fetch", spell = "web-fetcher", params = { url = "$url" } },
--   { name = "summarize", spell = "text-summarizer", params = { text = "$fetch.content" } }
-- })

-- Create reusable libraries
spell.library(name, functions)
-- Example: spell.library("http-utils", {
--   validate_url = function(url) ... end,
--   safe_fetch = function(url, options) ... end
-- })
```

### 3. Execution Context Management
```lua
-- Access current execution context
spell.context()
-- Returns: { spell_name, execution_id, start_time, params, environment }

-- Access configuration values
spell.config(key, default)
-- Example: spell.config("max_retries", 3)
-- Example: spell.config("timeout", 30)

-- Caching with TTL support
spell.cache(key, value, ttl)
-- Example: spell.cache("api_response_" .. url, response, 300) -- 5 min TTL
-- Example: local cached = spell.cache("expensive_calculation") -- Read from cache
```

### 4. Advanced Features
```lua
-- Error handling and recovery
spell.on_error(handler)
-- Example: spell.on_error(function(err) 
--   logging.error("Spell failed", { error = err, context = spell.context() })
--   return { error = err, recoverable = true }
-- end)

-- Lifecycle hooks
spell.on_init(handler)     -- Called after spell.init()
spell.on_start(handler)    -- Called before main execution
spell.on_complete(handler) -- Called after successful execution
spell.on_cleanup(handler)  -- Called during cleanup (always)

-- Environment management
spell.env(key, value)      -- Set environment variable
spell.env(key)             -- Get environment variable
spell.sandbox(config)      -- Create sandboxed execution environment

-- Resource management
spell.resource(name, config) -- Register managed resource
spell.cleanup_resources()    -- Cleanup all managed resources
```

## Implementation Strategy

### 1. Global Integration
- Work with existing global objects (`params`, `llm`, `tools`, etc.)
- Provide spell-scoped versions of globals for composition
- Maintain context isolation between composed spells

### 2. Parameter System
- Extend the global `params` object functionality
- Add validation, type checking, and default values
- Support parameter transformation and dependency resolution

### 3. Composition Engine
- Enable spell-to-spell communication through parameter passing
- Support conditional execution and branching
- Provide template variable substitution (`$variable` syntax)

### 4. Cache Integration
- Use existing cache infrastructure if available
- Provide in-memory fallback for development
- Support distributed caching for production environments

### 5. Context Management
- Track execution metadata (timing, parameters, results)
- Provide context inheritance for composed spells
- Enable context-aware debugging and logging

## Usage Patterns

### 1. Simple Spell Framework
```lua
-- Traditional spell with framework
spell.init({
  name = "hello-world",
  params = {
    name = { type = "string", default = "World" }
  }
})

local name = spell.params("name")
local greeting = "Hello, " .. name .. "!"

spell.output(greeting, "text")
```

### 2. Composed Spell Workflow
```lua
spell.init({
  name = "research-workflow",
  params = {
    topic = { type = "string", required = true }
  }
})

local results = spell.compose({
  {
    name = "search",
    spell = "web-search",
    params = { query = "$topic", limit = 5 }
  },
  {
    name = "analyze",
    spell = "content-analyzer", 
    params = { urls = "$search.results" }
  },
  {
    name = "summarize",
    spell = "multi-source-summarizer",
    params = { 
      sources = "$analyze.content",
      style = "academic"
    }
  }
})

spell.output(results.summarize, "markdown")
```

### 3. Library Creation
```lua
spell.library("ml-utils", {
  classify_sentiment = function(text)
    return llm.chat("Classify sentiment (positive/negative/neutral): " .. text)
  end,
  
  extract_entities = function(text)
    local prompt = "Extract named entities from: " .. text
    return llm.chat(prompt)
  end,
  
  batch_process = function(texts, processor)
    local results = {}
    for i, text in ipairs(texts) do
      results[i] = processor(text)
    end
    return results
  end
})
```

## Testing Requirements

1. **Initialization Testing**: Verify spell.init() correctly sets up metadata and parameters
2. **Parameter Validation**: Test type checking, required fields, and default values  
3. **Composition Testing**: Verify spell composition and parameter passing
4. **Context Isolation**: Ensure composed spells don't interfere with each other
5. **Cache Testing**: Verify TTL behavior and cache invalidation
6. **Output Formatting**: Test various output formats and metadata handling
7. **Library Loading**: Test spell.include() and library registration
8. **Error Handling**: Test error propagation and recovery mechanisms
9. **Resource Management**: Test cleanup and resource lifecycle
10. **Integration Testing**: Test with existing spell patterns and modules

## Integration with Existing Systems

- **Bridge Architecture**: Leverage existing bridge system for advanced features
- **Module System**: Integrate with current Lua module loading
- **Global Objects**: Extend rather than replace existing globals
- **Backward Compatibility**: Ensure existing spells continue to work unchanged
- **Testing Framework**: Use existing testing infrastructure for validation