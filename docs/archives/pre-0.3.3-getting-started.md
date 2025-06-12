# Getting Started with go-llmspell

## Installation

```bash
go get github.com/lexlapax/go-llmspell
```

## Basic Usage

### Using the CLI

```bash
# Run a Lua spell
llmspell -script examples/hello.lua

# Run with verbose output
llmspell -script examples/chat.lua -v

# Use a different engine
llmspell -script examples/hello.js -engine js
```

### Writing Your First Spell

Create a file `hello.lua`:

```lua
-- A simple spell that greets the world
print("🧙‍♂️ Casting hello spell...")

-- Access LLM functionality
local response = llm.chat("Hello, magical world!")
print("✨ LLM says: " .. response)
```

### Using as a Library

```go
package main

import (
    "github.com/lexlapax/go-llmspell/pkg/engine"
    "github.com/lexlapax/go-llmspell/pkg/spells"
)

func main() {
    // Create a spell library
    library := spells.NewLibrary()
    
    // Load and execute a spell
    // (Implementation coming soon)
}
```

## Next Steps

- Explore [example spells](../examples/)
- Read the [architecture overview](architecture.md)
- Check the [API reference](api-reference.md)