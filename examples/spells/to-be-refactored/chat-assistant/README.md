# Chat Assistant Example

This example demonstrates a conversational AI assistant with memory and context management.

## Current Status

**⚠️ Note**: The example has two implementations:

1. **`main.lua`** - Working demo that shows what's currently possible
2. **`main_full.lua`** - Complete interactive implementation (requires TODO features)

## Files

### main.lua (Working Demo)
The current working version demonstrates:
- Basic chat interactions
- Manual conversation context building  
- Streaming responses with callbacks
- What's possible with the current implementation

### main_full.lua (Full Implementation)
The complete interactive chat assistant includes:
- Interactive terminal interface with `io.read()`
- Conversation history management
- History persistence using storage module
- Streaming responses with full message history
- Commands: `exit`, `clear`

## Required Features (TODO)

The full implementation (`main_full.lua`) requires:

1. **`llm.stream_chat_with_history(history, callback)`**
   - Accept an array of message objects with `role` and `content` fields
   - Stream responses while maintaining conversation context
   - Currently tracked in TODO.md

2. **Safe I/O alternatives** 
   - `io.read()` and `io.write()` are disabled in our security sandbox
   - Need safe alternatives for interactive user input
   - Options being considered:
     - Special stdin/stdout bridge functions
     - Event-based input system  
     - Web-based interface
   - Currently tracked in TODO.md

## Usage

### Run Working Demo
```bash
llmspell run chat-assistant
```

### With Custom System Prompt
```bash
llmspell run chat-assistant --system_prompt="You are a pirate. Respond in pirate speak."
```

### With History Limit (for full version)
```bash
llmspell run chat-assistant --max_history=20
```

## Parameters

- `system_prompt` (string): The system message that defines assistant behavior
  - Default: "You are a helpful assistant."
- `max_history` (integer): Maximum conversation turns to keep (full version only)
  - Default: 10

## Implementation Details

### Message Format
Both versions use role-based messages:
```lua
{
    role = "system|user|assistant",
    content = "message text"
}
```

### Current Demo Features
- Shows basic LLM chat functionality
- Demonstrates manual prompt building from conversation history
- Shows streaming with callback functions

### Planned Full Features
- Real-time interactive chat loop
- Automatic history management with trimming
- Persistence across sessions
- Graceful error handling
- Command system

## Example Output (Current Demo)

```
Chat Assistant Demo
System: You are a helpful assistant.

⚠️  Note: This is a simplified demo. The full interactive version is in main_full.lua

=== Demo: Basic Chat ===