# Lua Spell Examples

This directory contains example Lua spells demonstrating various features and patterns in go-llmspell.

## Examples Overview

### Basic Examples

2. **[01-tools-usage.lua](01-tools-usage.lua)** - Using built-in tools
   - File operations
   - Web fetching
   - Date/time utilities
   - Calculator
   - Tool discovery

1. **[02-basic-llm.lua](02-basic-llm.lua)** - Basic LLM interaction
   - Simple completions
   - Streaming responses
   - Multi-turn conversations
   - Error handling
   - Multiple models

3. **[03-agent-plain.lua](03-agent-plain.lua)** - Agent without tools
   - Creating agents
   - Agent conversations
   - Agent personalities
   - Chain of thought

4. **[04-agent-with-tools.lua](04-agent-with-tools.lua)** - Agent with tools
   - Tool-enabled agents
   - Autonomous task completion
   - Tool selection
   - Error recovery

5. **[05-agent-as-tool.lua](05-agent-as-tool.lua)** - Agent wrapped as tool
   - Nested agents
   - Agent delegation
   - Complex hierarchies
   - Specialized agents

### Advanced Examples

6. **[06-complex-workflows.lua](06-complex-workflows.lua)** - Complex workflows
   - Multi-step processes
   - Conditional logic
   - Parallel execution
   - State management

7. **[07-event-driven.lua](07-event-driven.lua)** - Event-driven spells
   - Event handling
   - Reactive programming
   - State changes
   - Async events

8. **[08-performance-patterns.lua](08-performance-patterns.lua)** - Performance optimization
   - Caching strategies
   - Batch processing
   - Parallel operations
   - Resource management

## Running Examples

### Basic Execution

```bash
# Run with default parameters
llmspell run examples/spells/lua/01-basic-llm.lua

# Run with custom parameters
llmspell run examples/spells/lua/01-basic-llm.lua -p model=gpt-4 -p prompt="Hello, world!"

# Run with debug output
llmspell run examples/spells/lua/02-tools-usage.lua --debug
```

### Environment Setup

Make sure you have API keys configured:

```bash
export OPENAI_API_KEY="your-key-here"
export ANTHROPIC_API_KEY="your-key-here"  # For Claude models
```

### Parameters

Most examples accept parameters to customize behavior:

- `model` - LLM model to use (default: gpt-3.5-turbo)
- `prompt` - Input prompt for examples
- `debug` - Enable debug output ("true"/"false")
- `output_dir` - Directory for output files

Example:
```bash
llmspell run examples/spells/lua/04-agent-with-tools.lua \
  -p model=gpt-4 \
  -p task="Research climate change" \
  -p output_dir=./results
```

## Learning Path

1. **Start with basics**: Begin with `01-basic-llm.lua` to understand LLM interactions
2. **Explore tools**: Try `02-tools-usage.lua` to see available tools
3. **Understand agents**: Progress through `03-agent-plain.lua` and `04-agent-with-tools.lua`
4. **Advanced patterns**: Study workflows, events, and performance optimizations

## Common Patterns

### Error Handling
```lua
local success, result = pcall(function()
    return llm.complete({...})
end)
if not success then
    log.error("LLM call failed", {error = result})
end
```

### Retry Logic
```lua
local function with_retry(fn, max_attempts)
    for attempt = 1, max_attempts do
        local success, result = pcall(fn)
        if success then return result end
        if attempt < max_attempts then
            core.sleep(2 ^ (attempt - 1))
        end
    end
    error("Max retries exceeded")
end
```

### State Management
```lua
local state = state.create("example_state")
state:set("counter", 0)
state:increment("counter")
```

## Troubleshooting

### API Key Issues
- Ensure API keys are set in environment variables
- Check key permissions and quotas
- Verify model availability for your API key

### Performance Issues
- Use appropriate models for tasks (GPT-3.5 for simple tasks)
- Implement caching for repeated operations
- Batch operations when possible

### Debugging
- Use `--debug` flag for detailed output
- Add `log.debug()` statements in spells
- Check `llmspell.log` for errors

## Contributing

To add new examples:
1. Follow the naming pattern: `XX-description.lua`
2. Include comprehensive comments
3. Add error handling
4. Update this README
5. Test with multiple models

## Resources

- [User Guide](../../../docs/user-guide/lua-spells.md)
- [API Reference](../../../docs/user-guide/api-reference.md)
- [Common Patterns](../../../docs/user-guide/common-patterns.md)
- [Troubleshooting](../../../docs/user-guide/troubleshooting.md)