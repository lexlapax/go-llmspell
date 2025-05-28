# go-llmspell Examples

This directory contains example spells and integration code for go-llmspell.

## Directory Structure

### /spells/
Contains ready-to-run spell examples that demonstrate various capabilities:

- **hello-llm/** - Basic LLM interaction example
- **chat-assistant/** - Interactive chat with conversation history
- **provider-compare/** - Compare responses from multiple LLM providers

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
llmspell run hello-llm
llmspell run chat-assistant
llmspell run provider-compare --param prompt="What is the meaning of life?"
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