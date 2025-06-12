# Go-LLMSpell Architecture

## Table of Contents

1. [Introduction](#introduction)
2. [Why Go-LLMSpell Exists](#why-go-llmspell-exists)
3. [Core Philosophy](#core-philosophy)
4. [Architecture Overview](#architecture-overview)
5. [Bridge-First Design](#bridge-first-design)
6. [Component Architecture](#component-architecture)
7. [Script Engine System](#script-engine-system)
8. [Bridge Layer](#bridge-layer)
9. [Type System](#type-system)
10. [Security Model](#security-model)
11. [Examples](#examples)
12. [Future Considerations](#future-considerations)

---

**Quick Links:**
- [TODO.md](../../TODO.md) - Current implementation tasks
- [TODO-DONE.md](../../TODO-DONE.md) - Completed work tracking
- [Architecture Diagrams](../images/) - SVG diagrams referenced in this document

## Introduction

Go-LLMSpell is a **scriptable interface for LLM interactions** that bridges multiple scripting languages (Lua, JavaScript, Tengo) to the powerful go-llms library. It enables developers to write "spells" - scripts that orchestrate AI agents, workflows, and tools without needing to compile Go code.

### What is a Spell?

A spell is a script written in your language of choice that controls LLMs and their associated tools. Think of it as magical incantations that bring AI capabilities to life through simple, expressive code.

```lua
-- Example spell: Research Assistant
local researcher = llm.agent({
    model = "claude-3-opus",
    tools = {"web_search", "file_write"},
    system = "You are a helpful research assistant"
})

local topic = "quantum computing breakthroughs 2025"
local research = researcher:run("Research " .. topic .. " and summarize findings")
tools.file_write("research_summary.md", research)
```

## Why Go-LLMSpell Exists

### The Problem

1. **Compilation Barrier**: Working with LLMs in Go requires compilation for every change
2. **Rapid Prototyping**: AI workflows need constant iteration and experimentation
3. **Language Preference**: Different teams prefer different scripting languages
4. **Hot Reloading**: Production systems need to update AI behaviors without downtime
5. **Complexity**: Direct go-llms usage requires deep Go knowledge

### The Solution

Go-LLMSpell provides a **bridge-first architecture** that:
- Exposes go-llms functionality through simple scripting APIs
- Supports multiple scripting languages with the same interface
- Enables hot-reloading of AI behaviors
- Maintains type safety at bridge boundaries
- Provides security through sandboxed execution

### Key Benefits

🚀 **Rapid Development**: Test AI workflows instantly without compilation  
🔄 **Hot Reloading**: Update spells in production without restarts  
🌐 **Multi-Language**: Choose Lua, JavaScript, or Tengo based on your needs  
🔒 **Secure**: Sandboxed script execution with resource limits  
🎯 **Type Safe**: Automatic conversions between scripts and Go  
📚 **Reusable**: Build libraries of spells for common tasks  

## Core Philosophy

### 1. Bridge, Don't Build

We don't reimplement features that exist in go-llms. Instead, we create thin bridges that expose go-llms functionality to scripts. This ensures:
- Automatic compatibility with go-llms updates
- Minimal maintenance burden
- Consistent behavior with direct go-llms usage

### 2. Engine-Agnostic Design

All features work identically across Lua, JavaScript, and Tengo. Scripts are portable between engines with minimal changes.

### 3. Script-Friendly APIs

We hide go-llms complexity behind intuitive scripting interfaces. Complex Go patterns become simple script calls.

### 4. Security First

All script execution happens in sandboxed environments with configurable resource limits and permission controls.

### 5. Upstream-First Development

When new features or improvements are needed for core LLM, agent, tools, or workflow functionality, we contribute them upstream to go-llms rather than implementing them locally. This ensures:

- **Community Benefit**: Improvements benefit the entire go-llms ecosystem
- **Maintenance Reduction**: No custom forks or divergent implementations to maintain
- **Quality Assurance**: Changes go through go-llms review and testing processes
- **Long-term Compatibility**: Ensures go-llmspell stays aligned with go-llms evolution

**Contribution Process**:
1. **Identify Need**: Feature required for spell functionality
2. **Evaluate Scope**: Determine if it's core go-llms functionality vs. scripting-specific
3. **Upstream First**: Submit change request (CR) to go-llms if core functionality
4. **Bridge Second**: Create bridge to expose new go-llms feature to scripts
5. **Document**: Update bridges and examples once upstream change is merged

This approach reinforces our bridge-first philosophy and strengthens the broader Go LLM ecosystem.

## Architecture Overview

![Architecture Overview](../images/architecture-overview.svg)

Go-LLMSpell consists of two main packages:

```
go-llmspell/
├── pkg/
│   ├── engine/          # Script engine interfaces and implementations
│   │   ├── interface.go # Core interfaces (ScriptEngine, Bridge, TypeConverter)
│   │   ├── registry.go  # Engine registry and discovery
│   │   ├── types.go     # Type system and conversions
│   │   ├── lua/        # Lua engine implementation
│   │   ├── javascript/ # JavaScript engine implementation
│   │   └── tengo/      # Tengo engine implementation
│   │
│   └── bridge/         # Bridges to go-llms functionality
│       ├── manager.go       # Bridge lifecycle management
│       ├── llm_agent.go     # LLM agent orchestration bridge
│       ├── state.go         # State management bridge
│       ├── workflow.go      # Workflow engine bridge
│       ├── tools.go         # Tool system bridge
│       ├── events.go        # Event system bridge
│       └── ... (other bridges)
```

### Data Flow

```
User Script (Lua/JS/Tengo)
    ↓ [Script API Calls]
Script Engine (pkg/engine/)
    ↓ [Type Conversion]
Bridge Layer (pkg/bridge/)
    ↓ [Go Function Calls]
go-llms Library
    ↓ [Provider APIs]
LLM Providers (OpenAI, Anthropic, etc.)
```

## Bridge-First Design

### What We Build vs What We Bridge

#### We Build (Our Value-Add)
1. **Script Engine System** - Multi-language execution environments
2. **Type Conversion Layer** - Seamless script ↔ Go type conversions
3. **Bridge Interfaces** - Clean APIs to go-llms functionality
4. **Language Bindings** - Idiomatic APIs for each scripting language
5. **Developer Experience** - Hot reload, debugging, examples

#### We Bridge (From go-llms)
1. **LLM Agents** - Full agent system with tools and orchestration
2. **State Management** - StateManager, transforms, persistence
3. **Workflows** - Sequential, parallel, conditional, loop workflows
4. **Event System** - Real-time event streaming and subscriptions
5. **Tools** - File, web, data, datetime, math, system tools
6. **Infrastructure** - Tracing, metrics, hooks, guardrails

#### We Defer (Not Yet in go-llms)
1. **Memory System** - Await go-llms implementation
2. **Conversation Management** - Await go-llms implementation

### Bridge Implementation Strategy

Each bridge follows these principles:

1. **Thin Wrappers**: Minimal code between scripts and go-llms
2. **Type Safety**: Handle all conversions at bridge boundaries
3. **Error Mapping**: Convert Go errors to script-friendly formats
4. **Performance**: Cache conversions, optimize hot paths

Example bridge structure:

```go
// Bridge to go-llms StateManager
type StateBridge struct {
    manager *golms.StateManager
}

func (b *StateBridge) Register(engine ScriptEngine) error {
    // Expose methods to scripts
    engine.RegisterFunction("state.create", b.wrapCreate)
    engine.RegisterFunction("state.get", b.wrapGet)
    engine.RegisterFunction("state.set", b.wrapSet)
    return nil
}

func (b *StateBridge) wrapGet(key string) (interface{}, error) {
    // Get from go-llms
    value, err := b.manager.Get(key)
    if err != nil {
        return nil, b.mapError(err)
    }
    // Convert to script-friendly type
    return b.toScriptType(value), nil
}
```

## Component Architecture

### Script Engine System

![Script Engine Architecture](../images/engine-architecture.svg)

The Script Engine provides the execution environment for spells:

```go
type ScriptEngine interface {
    // Lifecycle
    Initialize(config EngineConfig) error
    Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error)
    ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (interface{}, error)
    Shutdown() error
    
    // Bridge Management
    RegisterBridge(name string, bridge Bridge) error
    UnregisterBridge(name string) error
    
    // Type System
    ToNative(scriptValue interface{}) (interface{}, error)
    FromNative(goValue interface{}) (interface{}, error)
    
    // Resource Management
    SetMemoryLimit(bytes int64) error
    SetTimeout(duration time.Duration) error
}
```

### Bridge Layer

![Bridge Layer Architecture](../images/bridge-architecture.svg)

Bridges connect script engines to go-llms functionality:

```go
type Bridge interface {
    // Registration
    Register(engine ScriptEngine) error
    Unregister() error
    
    // Metadata
    Name() string
    Methods() []MethodInfo
    
    // Type conversion hints
    TypeMappings() map[string]TypeMapping
}
```

#### Core Bridges

1. **LLM Agent Bridge** (`llm_agent.go`)
   - Agent creation and configuration
   - Tool registration and execution
   - Sub-agent orchestration
   - Lifecycle hooks and events

2. **State Management Bridge** (`state.go`)
   - State lifecycle operations
   - Transforms (filter, flatten, sanitize)
   - Merge strategies
   - Persistence interface

3. **Workflow Bridge** (`workflow.go`)
   - All workflow types (sequential, parallel, conditional, loop)
   - Workflow composition
   - State passing between steps
   - Error handling and recovery

4. **Event System Bridge** (`events.go`)
   - Real-time event streaming
   - Event filtering and subscriptions
   - Event metadata and correlation
   - All event types (lifecycle, tool, workflow)

5. **Tool System Bridge** (`tools.go`)
   - Built-in tool access
   - Custom tool registration
   - Tool composition and chaining
   - Parameter validation

## Type System

### Type Conversion Flow

```
Script Type → TypeConverter → Go Type → go-llms
     ↑                                      ↓
     └──────── TypeConverter ←─────────────┘
```

### Supported Type Mappings

| Script Type | Go Type | Notes |
|-------------|---------|-------|
| boolean | bool | Direct mapping |
| number | float64 | All numbers are float64 |
| string | string | UTF-8 encoded |
| table/object | map[string]interface{} | Recursive conversion |
| array | []interface{} | Preserves order |
| function | func(...) ... | With adapter |
| nil/null | nil | Direct mapping |

### Type Converter Interface

```go
type TypeConverter interface {
    // Basic type conversions
    ToBoolean(value interface{}) (bool, error)
    ToNumber(value interface{}) (float64, error)
    ToString(value interface{}) (string, error)
    ToArray(value interface{}) ([]interface{}, error)
    ToMap(value interface{}) (map[string]interface{}, error)
    
    // Advanced conversions
    ToStruct(source interface{}, target interface{}) error
    ToFunction(fn interface{}) (interface{}, error)
    FromFunction(fn interface{}) (interface{}, error)
    
    // Type information
    GetType(value interface{}) TypeInfo
    SupportsType(t reflect.Type) bool
}
```

## Security Model

### Script Sandboxing

Each script runs in an isolated environment with:

1. **Resource Limits**
   - Memory usage caps
   - CPU time limits
   - Execution timeouts
   - Goroutine limits

2. **Permission Controls**
   - File system access restrictions
   - Network access controls
   - Environment variable filtering
   - System call blocking

3. **Built-in Function Restrictions**
   - Dangerous functions disabled
   - Import/require limitations
   - Reflection controls

### Security Configuration

```go
type SecurityConfig struct {
    // Resource limits
    MaxMemory      int64         // bytes
    MaxCPUTime     time.Duration // execution time
    Timeout        time.Duration // total timeout
    MaxGoroutines  int           // concurrent routines
    
    // Permissions
    AllowFileRead  bool
    AllowFileWrite bool
    AllowNetwork   bool
    AllowEnvVars   []string // whitelist
    
    // Script restrictions
    DisabledBuiltins []string
    AllowedImports   []string
}
```

## Examples

### Basic LLM Interaction

**Lua:**
```lua
local response = llm.complete({
    model = "gpt-4",
    messages = {
        {role = "user", content = "Explain quantum computing"}
    }
})
print(response.content)
```

**JavaScript:**
```javascript
const response = await llm.complete({
    model: "gpt-4",
    messages: [
        {role: "user", content: "Explain quantum computing"}
    ]
});
console.log(response.content);
```

**Tengo:**
```tengo
response := llm.complete({
    model: "gpt-4",
    messages: [
        {role: "user", content: "Explain quantum computing"}
    ]
})
fmt.println(response.content)
```

### Agent with Tools

```lua
-- Create an agent with tools
local agent = llm.agent({
    model = "claude-3-opus",
    tools = {"web_search", "calculator", "file_read"},
    system = "You are a helpful assistant"
})

-- Register custom tool
tools.register({
    name = "get_weather",
    description = "Get current weather for a location",
    parameters = {
        location = {type = "string", required = true}
    },
    handler = function(params)
        -- Implementation
        return {temperature = 72, condition = "sunny"}
    end
})

-- Run agent with tool access
local result = agent:run("What's the weather in NYC?")
```

### State Management

```javascript
// Create state with persistence
const state = await state.create({
    persistence: "file",
    path: "./agent_state.json"
});

// Set values
await state.set("conversation_id", "abc123");
await state.set("context", {
    user: "John",
    topic: "quantum computing",
    messages: []
});

// Transform state
const filtered = await state.transform({
    type: "filter",
    keys: ["conversation_id", "context.topic"]
});

// Merge states
const merged = await state.merge(otherState, {
    strategy: "deep",
    conflictResolution: "theirs"
});
```

### Workflow Orchestration

```lua
-- Define a research workflow
local workflow = workflow.create({
    name = "research_workflow",
    type = "sequential"
})

-- Add steps
workflow:addStep({
    name = "search",
    agent = researcher,
    prompt = "Search for information about {{topic}}"
})

workflow:addStep({
    name = "summarize",
    agent = summarizer,
    prompt = "Summarize the findings: {{search.output}}"
})

workflow:addStep({
    name = "save",
    tool = "file_write",
    params = {
        path = "research_{{topic}}.md",
        content = "{{summarize.output}}"
    }
})

-- Execute workflow
local result = workflow:run({
    topic = "quantum computing"
})
```

### Event-Driven Patterns

```javascript
// Subscribe to events
events.on("agent.started", (event) => {
    console.log(`Agent ${event.agentId} started`);
});

events.on("tool.executed", (event) => {
    metrics.record("tool_usage", {
        tool: event.toolName,
        duration: event.duration,
        success: event.success
    });
});

// Emit custom events
events.emit("custom.milestone", {
    workflow: "research",
    stage: "completed",
    duration: 45.2
});
```

### Hook System

```lua
-- Register lifecycle hooks
hooks.register({
    beforeGenerate = function(ctx, request)
        -- Modify request before LLM call
        request.temperature = 0.7
        return request
    end,
    
    afterGenerate = function(ctx, request, response)
        -- Process response after LLM call
        logger.info("Tokens used: " .. response.usage.total_tokens)
        return response
    end,
    
    beforeToolCall = function(ctx, tool, input)
        -- Validate tool inputs
        if tool == "file_write" and input.path:match("^/etc/") then
            error("Cannot write to system directories")
        end
        return input
    end
})
```

## Future Considerations

### Planned Enhancements

1. **Memory System Bridge** - Once available in go-llms
   - Short-term and long-term memory
   - Memory search and retrieval
   - Memory compression

2. **Conversation Management Bridge** - Once available in go-llms
   - Multi-turn conversation handling
   - Context window management
   - Conversation branching

3. **Additional Script Engines**
   - Python (via embedded interpreter)
   - WebAssembly support
   - User-provided engines

4. **Enhanced Developer Experience**
   - Visual workflow builder
   - Spell marketplace
   - IDE plugins

### Extension Points

The architecture supports extensions through:

1. **Custom Bridges** - Add new go-llms functionality
2. **Custom Type Adapters** - Support new type conversions
3. **Custom Security Policies** - Fine-grained permissions
4. **Custom Script Engines** - Bring your own language

### Upstream Collaboration

As go-llmspell grows, we anticipate contributing to go-llms in areas such as:

**Likely Upstream Contributions**:
- **Agent Enhancements** - Improved sub-agent coordination patterns
- **Workflow Improvements** - New workflow types or execution optimizations
- **Tool Ecosystem** - Additional built-in tools for common scripting needs
- **State Management** - Enhanced persistence backends or transformation functions
- **Event System** - Extended event types or filtering capabilities
- **Provider Support** - New LLM provider integrations discovered through script usage

**Collaboration Benefits**:
- **Script-Driven Insights**: Real-world spell usage reveals go-llms improvement opportunities
- **Rapid Prototyping**: Script-based testing can validate go-llms features before implementation
- **User Feedback Loop**: Scripters provide valuable feedback on go-llms API usability
- **Ecosystem Growth**: Strengthens both go-llms and go-llmspell communities

This collaborative model ensures go-llmspell remains a first-class citizen in the go-llms ecosystem while driving innovation upstream.

## Conclusion

Go-LLMSpell's bridge-first architecture provides a powerful, flexible, and secure way to script LLM interactions. By bridging to go-llms rather than reimplementing features, we ensure compatibility, reduce maintenance burden, and focus on our core value: making LLM orchestration accessible through simple scripts.

The clean separation between script engines and bridges, combined with comprehensive type conversion and security features, creates a robust platform for building AI-powered applications in any scripting language.