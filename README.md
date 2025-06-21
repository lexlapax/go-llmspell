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

- ✅ **Phase 1** - Engine and Bridge Foundation [COMPLETED - 2025-06-17]
  - 38+ bridges across 13 categories
  - Pure bridge architecture with zero business logic duplication
  - Complete ScriptValue type system for cross-engine compatibility

- ✅ **Phase 2** - Lua Engine Implementation [COMPLETED - 2025-06-20]
  - ✅ Complete Lua engine with GopherLua integration
  - ✅ Async/coroutine support with promises and channels
  - ✅ Comprehensive Lua standard library (18 modules)
  - ✅ Bridge integration layer with namespace flattening
  - ✅ Development tools: Debugger & Script Validator (100% coverage)
  - ✅ Performance optimization and profiling infrastructure

- ✅ **Phase 3** - Spell Runner CLI [COMPLETED - 2025-06-21]
  - ✅ Complete command-line interface with 11 commands
  - ✅ Interactive REPL with syntax highlighting and history
  - ✅ Template system for spell generation (5 template types)
  - ✅ Three-tier security profiles (sandbox, development, production)
  - ✅ Comprehensive documentation suite and shell completion
  - ✅ Integration tests and cross-platform compatibility

- 🔄 **Phase 4** - JavaScript Engine Implementation [READY TO START]
  - Research goja integration and ES6+ support design
  - Implement complete JavaScript engine with async/await
  - Create JavaScript standard library bridging go-llms

- 🔲 **Phase 5** - Tengo Engine Implementation [PLANNED]
- 🔲 **Phase 6** - Integration and Examples [PLANNED]

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

## 🎯 CLI Usage

The `llmspell` CLI provides comprehensive spell execution and management capabilities:

### Common Commands
```bash
# Execute spells
llmspell run script.lua --param input=data.txt
llmspell run agent.js --engine javascript --timeout 5m

# Interactive development
llmspell repl                    # Start Lua REPL
llmspell repl --engine lua       # Explicit engine selection

# Create new spells
llmspell new my-agent --type agent --author "Gold Space"
llmspell new workflow-spell --type workflow
llmspell new --list              # Show available templates

# Validation and debugging
llmspell validate script.lua --security --profile sandbox
llmspell debug complex-spell.lua # Interactive debugger

# Configuration management
llmspell config view             # Show current config
llmspell config set engine.default javascript
llmspell config init             # Create default config
```

### Advanced Features
```bash
# Security profiles
llmspell run spell.lua --profile sandbox    # Restricted execution
llmspell security list                      # View all profiles
llmspell security compare sandbox production

# Engine management
llmspell engines list            # Show available engines
llmspell engines info lua        # Engine capabilities
llmspell engines detect script.unknown

# Development tools
llmspell run dev.lua --watch --verbose      # Auto-reload on changes
llmspell run script.lua --dry-run          # Preview execution
```

### Shell Integration
```bash
# Enable tab completion
source <(llmspell completion bash)
llmspell completion zsh > ~/.zsh/completions/_llmspell

# Generate documentation
llmspell man > llmspell.1        # Generate man page
llmspell man --all --install     # Install all man pages
```

For complete CLI documentation, see: [CLI Usage Guide](docs/cli-usage.md)

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