# Provider Compare Spell

Compares responses from multiple LLM providers to the same prompt. Tests each provider sequentially due to provider switching requirements.

## Features

- Tests multiple providers with the same prompt
- Shows progress as each provider is tested
- Handles provider switching and errors gracefully
- Provides detailed comparison results
- Restores original provider after testing

## Usage

Basic comparison:
```bash
llmspell run provider-compare --param prompt="Explain quantum computing in one sentence."
```

Test specific providers (string format):
```bash
llmspell run provider-compare \
  --param prompt="What is 2+2?" \
  --param providers="openai,anthropic"
```

Test specific providers (array format):
```bash
llmspell run provider-compare \
  --param prompt="What is 2+2?" \
  --param providers='["openai", "anthropic"]'
```

## Parameters

- `prompt` (string, required): The prompt to send to all providers
- `providers` (string or array, optional): List of providers to compare
  - String format: `"openai,anthropic,gemini"` (comma-separated)
  - Array format: `["openai", "anthropic", "gemini"]`
  - Default: `"openai,anthropic,gemini"`

## Example Output

```
Provider Comparison
==================
Prompt: Explain quantum computing in one sentence.

Testing providers: openai, anthropic, gemini

Sending requests to all providers...

  [1/3] Testing openai...
       Result: ✓ Success
  [2/3] Testing anthropic...
       Result: ✓ Success
  [3/3] Testing gemini...
       Result: ✓ Success

================================================================================
RESULTS
================================================================================

OPENAI
------
Status: Success
Response:
Quantum computing uses quantum mechanical phenomena like superposition and 
entanglement to process information in ways that classical computers cannot.

ANTHROPIC
---------
Status: Success
Response:
Quantum computing harnesses quantum mechanical effects to perform certain 
calculations exponentially faster than traditional computers.

GEMINI
------
Status: Success
Response:
Quantum computing leverages the principles of quantum mechanics to solve 
complex problems that are intractable for classical computers.

================================================================================
SUMMARY
================================================================================
Providers tested: 3
Successful: 3
Failed: 0

✅ Comparison complete!
```

## Notes

- Requires API keys for the providers you want to test
- Only providers with valid API keys at startup will be available
- Provider switching must happen in the main thread, preventing true parallel execution
- For true parallel execution with a single provider, see the async-callbacks example

## Technical Limitations

Due to the way provider state is managed, true parallel execution across different providers is not currently possible within a single spell. Each provider switch must complete before the next request can be made. 

For parallel execution options:
1. Use multiple spell instances (one per provider)
2. Use async callbacks with a single provider (see async-callbacks example)