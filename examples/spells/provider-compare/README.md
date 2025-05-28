# Provider Compare Spell

Compares responses from multiple LLM providers to the same prompt.

## Features

- Tests multiple providers with the same prompt
- Measures response time for each provider
- Handles provider switching and errors gracefully
- Provides detailed comparison results
- Restores original provider after testing

## Usage

Basic comparison:
```bash
llmspell run provider-compare --param prompt="Explain quantum computing in one sentence."
```

Test specific providers:
```bash
llmspell run provider-compare \
  --param prompt="What is 2+2?" \
  --param providers='["openai", "anthropic"]'
```

## Parameters

- `prompt` (string, required): The prompt to send to all providers
- `providers` (array, optional): List of providers to compare (default: ["openai", "anthropic", "gemini"])

## Example Output

```
Provider Comparison
==================
Prompt: Explain quantum computing in one sentence.

Testing providers: openai, anthropic, gemini

Testing openai...
  Success! Response time: 1.23s

Testing anthropic...
  Success! Response time: 0.98s

Testing gemini...
  Success! Response time: 1.45s

================================================================================
RESULTS
================================================================================

OPENAI
------
Status: Success
Time: 1.23 seconds
Response:
Quantum computing uses quantum mechanical phenomena like superposition and 
entanglement to process information in ways that classical computers cannot.

ANTHROPIC
---------
Status: Success
Time: 0.98 seconds
Response:
Quantum computing harnesses quantum mechanical effects to perform certain 
calculations exponentially faster than traditional computers.

GEMINI
------
Status: Success
Time: 1.45 seconds
Response:
Quantum computing leverages the principles of quantum mechanics to solve 
complex problems that are intractable for classical computers.

================================================================================
SUMMARY
================================================================================
Providers tested: 3
Successful: 3
Failed: 0
Average response time: 1.22 seconds
```

## Notes

- Requires API keys for the providers you want to test
- Response times may vary based on network conditions and provider load
- Some providers may have different capabilities or limits