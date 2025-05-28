# Hello LLM Spell

A simple example spell that demonstrates basic LLM interaction capabilities.

## Features

- Lists available LLM providers
- Sends a simple chat message
- Generates text completion with token limit
- Demonstrates streaming API (if supported)

## Usage

```bash
llmspell run hello-llm
```

## Requirements

At least one of the following environment variables must be set:
- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `GEMINI_API_KEY`

## Example Output

```
Current LLM provider: openai

Available providers:
  1. openai
  2. anthropic
  3. gemini

Sending chat message...
Response: Hello! How can I help you today?

Generating completion...
Completion: The capital of France is Paris.

Streaming example:
Why don't scientists trust atoms?
Because they make up everything!
Streaming complete.
```