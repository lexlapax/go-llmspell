# Provider Compare Spell

Compares responses from multiple LLM providers to the same prompt.

## Features

- Tests multiple providers with the same prompt
- Handles provider switching and errors gracefully
- Provides detailed comparison results
- Restores original provider after testing

Note: Response time measurement is not available due to sandbox restrictions (os.clock is disabled)

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

Testing openai...
  Success!

Testing anthropic...
  Success!

Testing gemini...
  Success!

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
```

## Notes

- Requires API keys for the providers you want to test
- Response times may vary based on network conditions and provider load
- Some providers may have different capabilities or limits