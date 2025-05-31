# Lua Agent Example

This example demonstrates how to create custom agents that properly orchestrate LLM calls with tools.

## Features Demonstrated

1. **Research Agent**: Combines web_fetch tool with LLM to research topics
   - Uses LLM to extract search terms
   - Fetches web content using tools
   - Summarizes findings with LLM

2. **Code Analysis Agent**: Integrates custom Lua tool with LLM for code review
   - Custom tool analyzes code metrics
   - LLM provides detailed review based on metrics
   - Demonstrates tool creation and usage in agents

3. **Planning Agent**: Multi-step LLM orchestration
   - Breaks down complex tasks into steps
   - Analyzes tool requirements for each step
   - Shows pure LLM orchestration patterns

## Running the Example

```bash
llmspell run lua-agent
```

## Key Concepts

### Registering a Function as an Agent

```lua
function my_agent(input, options)
    return "Response: " .. input
end

agents.register("my-agent", my_agent)
```

### Registering a Table as an Agent

```lua
local my_agent = {
    name = "custom-name",  -- Optional, overrides registration name
    system_prompt = "...", -- Optional system prompt
    
    -- Required: execute method
    execute = function(self, input, options)
        return "Response"
    end,
    
    -- Optional: streaming support
    stream = function(self, input, options, callback)
        callback("chunk1 ")
        callback("chunk2 ")
        return nil  -- or error string
    end,
    
    -- Optional: getters/setters
    get_system_prompt = function(self)
        return self.system_prompt
    end,
    
    set_system_prompt = function(self, prompt)
        self.system_prompt = prompt
    end
}

agents.register("my-agent", my_agent)
```

### Using Registered Agents

```lua
-- Execute
local result, err = agents.execute("my-agent", "input text", {maxTokens = 100})

-- Stream
agents.stream("my-agent", "input", function(chunk)
    print(chunk)
    return nil  -- or return error to stop
end)

-- Get info
local info = agents.get("my-agent")

-- Remove
agents.remove("my-agent")
```

## Architecture

Lua agents are registered in the global agent registry with a unique provider name (`lua-<agent-name>`). This allows them to be used just like any other agent in the system while maintaining the flexibility of Lua scripting.

The `LuaAgent` wrapper handles:
- Type conversions between Lua and Go
- Thread-safe execution
- Error handling
- Optional method detection (streaming, tool management, etc.)