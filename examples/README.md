# go-llmspell Examples

This directory contains example spells and integration code for go-llmspell.

## Directory Structure

### /spells/
Contains ready-to-run spell examples that demonstrate various capabilities:

#### Basic Examples
- **hello-llm/** - Basic LLM interaction example with chat, completion, and streaming
- **web-summarizer/** - Fetches and summarizes web content using LLM

#### Advanced LLM Features
- **async-llm/** - Demonstrates promise-based async patterns with LLMs
- **async-callbacks/** - Shows how async callbacks enable parallel LLM operations
- **provider-compare/** - Compare responses from multiple LLM providers
- **chat-assistant/** - Interactive chat with conversation history

#### Tool System
- **builtin-tools/** - Demonstrates using built-in tools with LLM tool calling
- **tool-example/** - Shows how to create and use custom tools from Lua scripts

### /integration/
Contains Go code examples showing how to integrate the spell engine:

- **lua_integration.go** - Demonstrates Lua engine integration and security features

## Running Spells

To run a spell:
```bash
llmspell run <spell-name>
```

For example:
```bash
# Basic examples
llmspell run hello-llm
llmspell run web-summarizer --param url="https://example.com"

# Async/parallel examples
llmspell run async-llm
llmspell run async-callbacks --param mode=parallel
llmspell run provider-compare --param prompt="What is the meaning of life?"

# Interactive examples
llmspell run chat-assistant

# Tool examples
llmspell run builtin-tools --param query="What is 25 * 4 + 10?"
llmspell run tool-example
```

## Creating Your Own Spells

Each spell requires:
1. A directory under `spells/`
2. A `spell.yaml` metadata file
3. A main script file (e.g., `main.lua`)
4. Optional supporting files and documentation

See the [Spell Development Guide](../docs/spell-development.md) for detailed instructions.

## Integration Examples

The integration examples show how to:
- Create and configure script engines
- Register Go functions for script access
- Exchange variables between Go and scripts
- Implement security sandboxing
- Handle errors and timeouts

To run integration examples:
```bash
cd integration
go run lua_integration.go
```