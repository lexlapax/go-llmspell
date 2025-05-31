# go-llmspell

**Cast scripting spells to animate LLM golems** 🧙‍♂️✨

go-llmspell transforms complex LLM interactions into simple, magical scripts. Write spells in Lua, JavaScript, or Tengo that bring AI agents to life, automate conversations, and orchestrate intelligent workflows—all with the reliability and performance of Go.

```lua
-- example spell: creative-writer.lua
local topic = params.topic or "the future of AI"

-- Create an agent with creative writing abilities
local writer = agent.create({
    name = "creative_writer",
    system_prompt = "You are a creative writer with a vivid imagination."
})

-- Generate a story
local story = writer.run("Write a short story about " .. topic)

-- Save the result
fs.write("story.md", story)
log.info("Story created!", {topic = topic})
```

## 🚀 Key Features

- **🪄 Scriptable Magic**: Write spells in Lua, JavaScript, or Tengo to control LLMs
- **🤖 Agent Orchestration**: Create and manage AI agents with tools and workflows
- **🔧 Tool Integration**: Build custom tools that agents can use
- **⚡ Go Performance**: Native Go speed with embedded scripting flexibility
- **🔒 Secure Execution**: Sandboxed script execution with resource limits
- **📚 Spellbook Library**: Pre-written spells for common AI tasks

## 📖 Documentation

- [**Architecture Overview**](docs/architecture.md) - System design and components
- [**Implementation Guide**](docs/implementation-guide.md) - Development roadmap
- [**Spell Development Guide**](docs/spell-development.md) - How to write spells
- [**Getting Started**](docs/getting-started.md) - Quick start guide
- [**Documentation Index**](docs/README.md) - All documentation

## 🏗️ Project Status

This project is under active development. See our [TODO](TODO.md) for current tasks and [TODO-DONE](TODO-DONE.md) for completed work.

### Current Status
- ✅ Architecture designed and documented
- ✅ go-llms integration complete
- ✅ Basic project structure
- ✅ Core infrastructure implementation (Phase 1 complete)
  - ✅ Engine interface and registry system
  - ✅ Bridge infrastructure with lifecycle management
  - ✅ Security context with resource limits
- ✅ LLM Bridge enhancement (Phase 2 complete)
  - ✅ Multi-provider support (OpenAI, Anthropic, Gemini)
  - ✅ Provider switching and model discovery
  - ✅ Type conversion utilities
  - ✅ Comprehensive test coverage
- ✅ Lua engine implementation (Phase 3 complete)
  - ✅ Full Lua VM integration with security sandbox
  - ✅ Complete standard library (JSON, HTTP, Storage, Log, Promise)
  - ✅ LLM bridge for Lua scripts
  - ✅ Example spells demonstrating capabilities
- ✅ Tool System implementation (Phase 4 complete)
  - ✅ Tool interface and registry for managing tools
  - ✅ Script-based tool creation with parameter validation
  - ✅ Lua bridge for tool system (tools module)
  - ✅ Example tools demonstrating the system
- ✅ Agent System core implementation (Phase 5 mostly complete)
  - ✅ Agent interface and registry system
  - ✅ Default agent implementation with go-llms integration
  - ✅ Tool integration for agent capabilities
  - ✅ Agent bridge for script access
  - 🔄 Lua integration pending

## 🛠️ Installation

```bash
# Clone the repository
git clone https://github.com/lexlapax/go-llmspell.git
cd go-llmspell

# Initialize submodules (for go-llms reference)
git submodule update --init --recursive

# Build the project
make build

# Run tests
make test
```

## 🎯 Quick Start

### Setting Up API Keys

The easiest way to configure API keys is using a `.env` file:

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env and add your API keys:
# OPENAI_API_KEY=sk-...
# ANTHROPIC_API_KEY=sk-ant-...
# GEMINI_API_KEY=AI...
```

The CLI will automatically load the `.env` file. See [Environment Setup](docs/env-setup.md) for more details.

### Running Example Spells

```bash
# Run the async LLM example (demonstrates promises)
./bin/llmspell run examples/spells/async-llm

# Compare multiple LLM providers
./bin/llmspell run examples/spells/provider-compare --param prompt="What is AI?"

# Simple chat assistant demo
./bin/llmspell run examples/spells/chat-assistant
```

### Available Example Spells

- **async-llm**: Demonstrates promise-based async patterns with LLMs
- **provider-compare**: Compares responses from multiple providers
- **chat-assistant**: Interactive chat with conversation history (demo version)
- **hello-world**: Basic spell structure example

### Future CLI (Coming Soon)

```bash
# Run a spell
llmspell run my-spell.lua

# List available spells
llmspell list

# Create a new spell
llmspell create my-spell
```

## 🏛️ Architecture

go-llmspell uses a layered architecture:

```
┌─────────────────────────────────────┐
│        Spell Scripts (Lua/JS)       │
├─────────────────────────────────────┤
│         Script Engines              │
├─────────────────────────────────────┤
│          Bridge Layer               │
├─────────────────────────────────────┤
│           go-llms                   │
└─────────────────────────────────────┘
```

### Core Components
- **Script Engines**: Lua (gopher-lua), JavaScript (goja), Tengo
- **Bridges**: LLM, Tools, Agents, Workflows, StdLib
- **Security**: Sandboxing, resource limits, filesystem jail
- **Spells**: Reusable scripts for common tasks

## 🔮 Example Spells (Coming Soon)

### Web Researcher
```lua
-- Research a topic using web search and LLM analysis
local researcher = spell.load("web-researcher")
local report = researcher.run({topic = "quantum computing"})
```

### Code Reviewer
```javascript
// Automated code review with AI
const reviewer = await spell.load("code-reviewer");
const review = await reviewer.run({
    file: "main.go",
    style: "golang"
});
```

### Blog Writer
```lua
-- Generate blog posts with research and editing
local writer = spell.load("blog-writer")
local post = writer.run({
    topic = "The Future of AI",
    tone = "professional",
    length = 1000
})
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) (coming soon).

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Write tests first (TDD approach)
4. Implement your feature
5. Run quality checks: `make fmt vet lint test`
6. Submit a pull request

## 📦 Dependencies

- [go-llms](https://github.com/lexlapax/go-llms) v0.2.6 - LLM provider abstraction
- [gopher-lua](https://github.com/yuin/gopher-lua) v1.1.1 - Lua 5.1 VM (integrated)
- [goja](https://github.com/dop251/goja) - JavaScript engine (planned)
- [tengo](https://github.com/d5/tengo) - Embeddable script language (planned)

## 📄 License

[License information to be added]

## 🎉 Acknowledgments

Built on top of the excellent [go-llms](https://github.com/lexlapax/go-llms) library.

---

**Note**: This project is under active development. APIs and features may change. Check the [documentation](docs/) for the latest information.