# Built-in Tools Example

This example demonstrates how to use the built-in tool system in go-llmspell with an LLM that supports tool calling.

## Overview

The spell showcases:
- Registration and listing of built-in tools
- Tool parameter validation using JSON schemas
- Direct tool execution from Lua scripts
- Integration with LLM tool calling capabilities

## Built-in Tools

The example uses these built-in tools:
- **calculator**: Performs basic arithmetic operations (add, subtract, multiply, divide)
- **string_length**: Returns the length of a string
- **string_reverse**: Reverses a string
- **json_extract**: Extracts values from JSON using dot notation paths

## Usage

```bash
# Run the spell from the project root
llmspell run builtin-tools

# Or with the full path
llmspell run examples/spells/builtin-tools
```

## Code Structure

The spell demonstrates:
1. Listing available tools
2. Direct tool execution with validation
3. Using tools with an LLM for automated problem solving
4. Error handling for invalid parameters

## Requirements

- An LLM provider that supports tool calling (OpenAI, Anthropic, etc.)
- Valid API credentials set in environment variables