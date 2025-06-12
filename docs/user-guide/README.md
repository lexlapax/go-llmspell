# User Guide

This directory contains documentation for go-llmspell users - anyone who wants to write spells and use the scripting interface.

## 📚 Table of Contents

### Getting Started
- [**Quick Start Guide**](quick-start.md) *(Coming Soon)* - Get up and running with your first spell
- [**Installation Guide**](installation.md) *(Coming Soon)* - Detailed installation instructions
- [**Environment Setup**](environment-setup.md) *(Coming Soon)* - API keys, configuration, and troubleshooting

### Writing Spells
- [**Spell Basics**](spell-basics.md) *(Coming Soon)* - Fundamental concepts and spell structure
- [**Lua Spells Guide**](lua-spells.md) *(Coming Soon)* - Writing spells in Lua
- [**JavaScript Spells Guide**](javascript-spells.md) *(Coming Soon)* - Writing spells in JavaScript
- [**Tengo Spells Guide**](tengo-spells.md) *(Coming Soon)* - Writing spells in Tengo

### Core Concepts
- [**Agents Guide**](agents.md) *(Coming Soon)* - Working with LLM agents
- [**Tools Guide**](tools.md) *(Coming Soon)* - Using and creating tools
- [**Workflows Guide**](workflows.md) *(Coming Soon)* - Orchestrating complex workflows
- [**State Management Guide**](state-management.md) *(Coming Soon)* - Managing state across spells

### Advanced Topics
- [**Event System Guide**](events.md) *(Coming Soon)* - Working with events and hooks
- [**Security Guide**](security.md) *(Coming Soon)* - Understanding permissions and sandboxing
- [**Performance Tips**](performance.md) *(Coming Soon)* - Optimizing spell performance
- [**Debugging Guide**](debugging.md) *(Coming Soon)* - Troubleshooting and debugging spells

### Reference
- [**API Reference**](api-reference.md) *(Coming Soon)* - Complete scripting API documentation
- [**Built-in Functions**](builtin-functions.md) *(Coming Soon)* - All available built-in functions
- [**Error Reference**](error-reference.md) *(Coming Soon)* - Common errors and solutions
- [**Examples Collection**](examples/) *(Coming Soon)* - Comprehensive spell examples

## 🪄 What are Spells?

Spells are scripts written in Lua, JavaScript, or Tengo that control LLMs and AI agents. They're called "spells" because they bring AI capabilities to life through simple, expressive code.

### Basic Spell Structure

**Lua Example:**
```lua
-- research-spell.lua
local topic = params.topic or "artificial intelligence"

-- Create a research agent
local researcher = llm.agent({
    model = "claude-3-opus",
    tools = {"web_search", "file_write"},
    system = "You are a helpful research assistant"
})

-- Research the topic
local findings = researcher:run("Research recent developments in " .. topic)

-- Save results
tools.file_write("research_" .. topic .. ".md", findings)
log.info("Research completed for: " .. topic)
```

**JavaScript Example:**
```javascript
// research-spell.js
const topic = params.topic || "artificial intelligence";

// Create a research agent
const researcher = await llm.agent({
    model: "claude-3-opus",
    tools: ["web_search", "file_write"],
    system: "You are a helpful research assistant"
});

// Research the topic
const findings = await researcher.run(`Research recent developments in ${topic}`);

// Save results
await tools.fileWrite(`research_${topic}.md`, findings);
log.info(`Research completed for: ${topic}`);
```

## 🎯 Quick Examples

### Simple LLM Chat
```lua
local response = llm.complete({
    model = "gpt-4",
    messages = {
        {role = "user", content = "Explain quantum computing"}
    }
})
print(response.content)
```

### Agent with Tools
```javascript
const agent = await llm.agent({
    model = "claude-3-opus",
    tools = ["calculator", "web_search"],
    system = "You are a helpful assistant"
});

const result = await agent.run("What's 15% of 2847?");
```

### Workflow Example
```lua
local workflow = workflow.sequential({
    steps = {
        {name = "research", agent = researcher},
        {name = "summarize", agent = summarizer},
        {name = "save", tool = "file_write"}
    }
})

local result = workflow:run({topic = "climate change"})
```

## 🔧 Available Script Engines

| Engine | Language | Status | Best For |
|--------|----------|--------|----------|
| **Lua** | Lua 5.1 | ✅ Ready | Simple scripts, fast execution |
| **JavaScript** | ES6+ | 🚧 Coming Soon | Complex logic, async patterns |
| **Tengo** | Tengo | 🚧 Coming Soon | Performance-critical scripts |

## 🛠️ Built-in Capabilities

### LLM Operations
- **llm.complete()** - Basic LLM completions
- **llm.stream()** - Streaming responses
- **llm.agent()** - Create AI agents
- **llm.models()** - List available models

### Tool System
- **tools.file_read()** - Read files
- **tools.file_write()** - Write files
- **tools.web_fetch()** - HTTP requests
- **tools.calculator()** - Math operations
- **tools.datetime()** - Date/time utilities

### State Management
- **state.create()** - Create state containers
- **state.get()** / **state.set()** - Access state
- **state.persist()** - Save state to disk
- **state.transform()** - Apply transformations

### Workflow Orchestration
- **workflow.sequential()** - Step-by-step execution
- **workflow.parallel()** - Concurrent execution
- **workflow.conditional()** - Branching logic
- **workflow.loop()** - Iterative patterns

## 🔒 Security & Permissions

All spells run in sandboxed environments with:

- **Resource Limits** - Memory, CPU, and time constraints
- **File System Controls** - Restricted file access
- **Network Controls** - Controlled external access
- **Function Restrictions** - Dangerous operations blocked

See the [Security Guide](security.md) for details.

## 📖 Learning Path

**New to Spells?**
1. Start with [Quick Start Guide](quick-start.md)
2. Read [Spell Basics](spell-basics.md)
3. Try the examples in your preferred language
4. Explore [Agents Guide](agents.md) and [Tools Guide](tools.md)

**Experienced Scripter?**
1. Check [API Reference](api-reference.md)
2. Browse [Examples Collection](examples/)
3. Dive into [Advanced Topics](#advanced-topics)

**Building Complex Workflows?**
1. Read [Workflows Guide](workflows.md)
2. Learn [State Management](state-management.md)
3. Explore [Event System](events.md)

## 🤝 Community & Support

- **Examples** - See [examples/](../../examples/) in the main repository
- **Issues** - Report bugs or request features on GitHub
- **Discussions** - Ask questions in GitHub Discussions
- **Contributing** - Help improve documentation and examples

## 🔗 Related Documentation

- [**Technical Documentation**](../technical/) - For developers and contributors
- [**Main README**](../../README.md) - Project overview and installation
- [**Architecture**](../technical/architecture.md) - How go-llmspell works internally

## 📝 Documentation Format

User guide documentation follows these conventions:

- **Practical Examples** - Every concept includes working code
- **Multi-Language** - Examples in Lua, JavaScript, and Tengo when applicable
- **Progressive Complexity** - Start simple, build to advanced topics
- **Real-World Focus** - Examples solve actual problems
- **Cross-References** - Links to related concepts and APIs

## 🆘 Getting Help

If you're stuck:

1. **Check Examples** - Look in [examples/](../../examples/) for similar use cases
2. **Search Documentation** - Use your browser's search (Ctrl/Cmd+F)
3. **API Reference** - Consult [API Reference](api-reference.md) for function details
4. **GitHub Issues** - Search existing issues or create a new one
5. **Discussions** - Ask the community in GitHub Discussions

---

**Back to:** [Project Root](../../README.md) | [All Documentation](../README.md) | [Technical Docs](../technical/README.md)