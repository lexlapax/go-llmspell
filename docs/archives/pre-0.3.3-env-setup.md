# Environment Setup

The llmspell CLI now automatically loads environment variables from a `.env` file in the current directory.

## Quick Start

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit the `.env` file and add your API keys:
   ```bash
   # LLM Provider API Keys
   OPENAI_API_KEY=sk-...
   ANTHROPIC_API_KEY=sk-ant-...
   GEMINI_API_KEY=AI...
   ```

3. Run your spells - the API keys will be automatically loaded:
   ```bash
   llmspell run examples/spells/provider-compare
   ```

## Manual Environment Setup

If you prefer not to use a `.env` file, you can still set environment variables manually:

```bash
# Set individual variables
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GEMINI_API_KEY="AI..."

# Or source from a file
source ~/.llmspell-env
```

## Provider Requirements

- **OpenAI**: Requires `OPENAI_API_KEY`
- **Anthropic**: Requires `ANTHROPIC_API_KEY`
- **Gemini**: Requires `GEMINI_API_KEY`

Only providers with valid API keys will be available at runtime.

## Mock Mode

For testing without API keys, set `MOCK_LLM=true`:

```bash
MOCK_LLM=true llmspell run examples/spells/hello-llm
```

Or in your `.env` file:
```
MOCK_LLM=true
```

## Security Notes

- Never commit your `.env` file to version control
- The `.gitignore` file already excludes `.env`
- Keep your API keys secure and rotate them regularly