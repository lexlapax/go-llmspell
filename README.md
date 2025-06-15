# go-llmspell

**Cast scripting spells to animate LLM golems** 🧙‍♂️✨

go-llmspell makes LLM interactions scriptable. Write spells in Lua, JavaScript, or Tengo that control AI agents, tools, and workflows—all with Go's reliability and performance.

```lua
-- example spell: creative-writer.lua
local writer = llm.agent({
    model = "claude-3-opus",
    system = "You are a creative writer with vivid imagination."
})

local story = writer:run("Write a short story about " .. params.topic)
fs.write("story.md", story)
log.info("Story created!")
```

## 🚀 Key Features

- **🪄 Multi-Language Spells**: Lua, JavaScript, or Tengo scripting
- **🤖 Agent Orchestration**: AI agents with tools and workflows  
- **⚡ Go Performance**: Native speed with scripting flexibility
- **🔒 Secure Execution**: Sandboxed scripts with resource limits
- **🌉 Bridge Architecture**: Leverages go-llms without duplication

## 📖 Documentation

### 📚 **Essential Reading**
- [**📋 Documentation Hub**](docs/README.md) - Complete documentation index
- [**🏛️ Architecture Overview**](docs/technical/architecture.md) - System design and philosophy
- [**👥 User Guide**](docs/user-guide/README.md) - Writing spells and using APIs

### 🔧 **Development**
- [**📋 Current Tasks**](TODO.md) - Active development priorities
- [**✅ Completed Work**](TODO-DONE.md) - Development history
- [**🤝 Contributing Guide**](CONTRIBUTING.md) - How to contribute

## 🏗️ Project Status (June 2025)

- ✅ **Architecture** - Bridge-first design documented and implemented
- ✅ **Phase 1.1** - Script Engine Interface complete
- ✅ **Phase 1.2** - Core Bridge Foundation complete
  - State management bridges (manager, context)
  - Bridge type system with go-llms aliases
  - Utility bridges (auth, json, llm, general)
- ✅ **Phase 1.3** - Core Bridge System complete
  - Agent, workflow, events, tools, and hooks bridges
  - Enhanced custom tool support with go-llms v0.3.5
  - Comprehensive test coverage with go-llms testutils
- 🚧 **Phase 1.4** - v0.3.5 Feature Integration in progress
- 🔮 **Coming Soon** - Lua engine, then JavaScript and Tengo

## 🛠️ Quick Start

### Installation
```bash
git clone https://github.com/lexlapax/go-llmspell.git
cd go-llmspell
git submodule update --init --recursive
make build && make test
```

### Setup API Keys
```bash
cp .env.example .env
# Edit .env with your API keys:
# OPENAI_API_KEY=sk-...
# ANTHROPIC_API_KEY=sk-ant-...
```

### Run Examples
```bash
# Basic LLM interaction
./bin/llmspell run examples/hello-llm

# Agent with tools
./bin/llmspell run examples/research-agent

# Provider comparison
./bin/llmspell run examples/provider-compare --param prompt="Explain AI"
```

## 🏛️ Architecture

```
┌─────────────────────────┐
│   Spell Scripts         │  ← Your spells (Lua/JS/Tengo)
├─────────────────────────┤
│   Script Engines        │  ← Multi-language execution
├─────────────────────────┤  
│   Bridge Layer          │  ← Type-safe go-llms access
├─────────────────────────┤
│   go-llms Library       │  ← LLM providers & tools
└─────────────────────────┘
```

**Key Principle**: We bridge to go-llms functionality rather than reimplementing it.

## 🔮 Example Spells

### Simple LLM Chat
```lua
local response = llm.complete({
    model = "gpt-4",
    messages = {{role = "user", content = "Explain quantum computing"}}
})
print(response.content)
```

### Agent with Tools
```javascript
const agent = await llm.agent({
    model: "claude-3-opus",
    tools: ["calculator", "web_search"]
});
const result = await agent.run("What's 15% of 2847?");
```

### Workflow Orchestration
```lua
local workflow = workflow.sequential({
    {name = "research", agent = researcher},
    {name = "summarize", agent = summarizer},
    {name = "save", tool = "file_write"}
})
local result = workflow:run({topic = "climate change"})
```

## 🤝 Contributing

We welcome contributions! See our [Contributing Guidelines](CONTRIBUTING.md) for:

- **Development setup** and TDD workflow
- **Code standards** and quality requirements  
- **Architecture principles** and bridge-first design
- **Community guidelines** and communication channels

**Quick development workflow:**
```bash
git checkout -b feature/my-feature
# Write tests first, then implement
make all  # Must pass before submitting
```

## 📦 Core Dependencies

- [**go-llms**](https://github.com/lexlapax/go-llms) v0.3.5 - LLM providers and agent framework
- [**gopher-lua**](https://github.com/yuin/gopher-lua) - Lua 5.1 VM
- [**goja**](https://github.com/dop251/goja) - JavaScript engine *(planned)*
- [**tengo**](https://github.com/d5/tengo) - Compiled script language *(planned)*

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🎉 Acknowledgments

Built on the excellent [go-llms](https://github.com/lexlapax/go-llms) library.

---

**⚡ Ready to cast your first spell?** Start with the [User Guide](docs/user-guide/README.md)!