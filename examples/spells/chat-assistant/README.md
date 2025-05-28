# Chat Assistant Spell

An interactive chat assistant that maintains conversation history across sessions.

**Note**: The full interactive version (`main.lua`) requires features not yet implemented:
- `storage.exists()`, `storage.read()`, `storage.write()` for persistence
- `json.encode()` and `json.decode()` for JSON handling
- `io.read()` and `io.write()` for interactive input (disabled for security)
- `llm.stream_chat_with_history()` for streaming with message history

For now, use `main_simple.lua` which demonstrates basic chat functionality.

## Features

- Interactive chat interface
- Maintains conversation context
- Persists history between sessions
- Configurable system prompt
- History management with size limits
- Streaming responses

## Usage

Basic usage:
```bash
llmspell run chat-assistant
```

With custom system prompt:
```bash
llmspell run chat-assistant --param system_prompt="You are a technical expert in Go programming."
```

With custom history limit:
```bash
llmspell run chat-assistant --param max_history=20
```

## Commands

- `exit` - Quit the chat and save history
- `clear` - Clear conversation history
- Empty line - Skip turn

## Configuration

The spell accepts the following parameters:

- `system_prompt` (string): The system message that defines the assistant's behavior
- `max_history` (integer): Maximum number of conversation turns to keep (default: 10)

## Example Session

```
Chat Assistant Started
System: You are a helpful assistant.
Type 'exit' to quit, 'clear' to reset history

You: What is Go programming language?