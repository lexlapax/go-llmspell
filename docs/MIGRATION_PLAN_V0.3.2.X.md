# Go-LLMSpell Clean Slate Migration Plan: Multi-Engine Architecture for go-llms v0.3.2

## Introduction

### What is go-llmspell?

go-llmspell is a Go library that provides a **scriptable interface for LLM interactions** using embedded scripting languages. It acts as a wrapper around the go-llms library, providing scripting capabilities for AI agent orchestration and workflow automation. The key innovation is "spells" - scripts written in Lua, JavaScript, or Tengo that control LLMs without needing to compile Go code.

### Why This Migration?

We are migrating to go-llms v0.3.2 to leverage its significant architectural improvements:

1. **Advanced Multi-Agent Orchestration** - v0.3.2 provides enhanced capabilities for coordinating multiple AI agents
2. **Agent-First Architecture** - Everything is now an agent (tools, workflows, even scripts), enabling more flexible composition
3. **Better State Management** - Improved immutable state operations and inter-agent state sharing
4. **Enhanced Event System** - Real-time event streaming and better event-driven patterns
5. **Cloud-Native Ready** - Built for distributed execution, state persistence, and scalability

### Benefits of go-llmspell

1. **Scriptable Magic** 🪄 - Write "spells" to control LLMs without compiling Go code
2. **Rapid Prototyping** - Test AI workflows instantly with script hot-reloading
3. **Multi-Language Support** - Choose between Lua, JavaScript, or Tengo based on your preference
4. **Agent Orchestration** - Create and manage AI agents with tools and workflows through simple scripts
5. **Provider Flexibility** - Mix and match LLM providers (OpenAI, Anthropic, Gemini) at runtime
6. **Security** - Sandboxed script execution with resource limits protects your system
7. **Reusable Spells** - Build a library of spells for common AI tasks

### How go-llmspell is Used

Users write "spells" (scripts) that orchestrate AI agents:

```lua
-- Example: Research agent spell
local researcher = agent.create({
    name = "web_researcher",
    tools = {"web_fetch"},
    system_prompt = "You are a research assistant..."
})

local result = researcher.run("Research quantum computing")
fs.write("research.md", result)
```

This migration to v0.3.2 represents a fundamental shift to fully leverage the latest go-llms capabilities while maintaining the ease of scripting that makes go-llmspell valuable for AI developers.

## Executive Summary

This document outlines a complete redesign of go-llmspell to fully leverage go-llms v0.3.2's advanced multi-agent orchestration capabilities while supporting multiple scripting languages. The architecture is designed around a language-agnostic scripting engine interface, with Lua (GopherLua) as the first implementation, followed by JavaScript (Goja) and Tengo.

## Vision and Objectives

### Core Philosophy
- **Engine-Agnostic Design**: All features work across any supported scripting language
- **Agent-First Architecture**: Everything is an agent - tools, workflows, even scripts
- **Write Once, Run Anywhere**: Scripts portable across engines where possible
- **Progressive Enhancement**: Start with any engine, add others as needed
- **Observable by Default**: Built-in monitoring, debugging, and tracing
- **Cloud-Native Ready**: Distributed execution, state persistence, scalability

### Multi-Engine Strategy
1. **Common Interface**: Single scripting engine interface for all languages
2. **Shared Bridges**: Language-agnostic bridge implementations
3. **Unified Type System**: Common type conversion layer
4. **Consistent APIs**: Same functionality exposed to all languages
5. **Engine Selection**: Runtime engine selection and switching

## Architecture Design

### 1. Multi-Engine Architecture

#### Script Engine Interface
```go
// Core engine interface that all scripting languages must implement
type ScriptEngine interface {
    // Lifecycle
    Initialize(config EngineConfig) error
    Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error)
    ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (interface{}, error)
    Shutdown() error
    
    // Bridge Management
    RegisterBridge(name string, bridge Bridge) error
    UnregisterBridge(name string) error
    GetBridge(name string) (Bridge, error)
    
    // Type System
    ToNative(scriptValue interface{}) (interface{}, error)
    FromNative(goValue interface{}) (interface{}, error)
    
    // Metadata
    Name() string
    Version() string
    FileExtensions() []string
    
    // Resource Management
    SetMemoryLimit(bytes int64) error
    SetTimeout(duration time.Duration) error
    GetMetrics() EngineMetrics
}
```

#### Bridge Interface
```go
// Language-agnostic bridge interface
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

// Concrete bridge implementations work with any engine
type AgentBridge struct {
    registry AgentRegistry
    // No language-specific code
}

type StateBridge struct {
    stateManager StateManager
    // No language-specific code
}
```

#### Type Conversion Layer
```go
// Universal type system for cross-engine compatibility
type TypeConverter interface {
    // Core conversions
    ToBoolean(v interface{}) (bool, error)
    ToNumber(v interface{}) (float64, error)
    ToString(v interface{}) (string, error)
    ToArray(v interface{}) ([]interface{}, error)
    ToMap(v interface{}) (map[string]interface{}, error)
    
    // Complex type handling
    ToStruct(v interface{}, target interface{}) error
    FromStruct(v interface{}) (map[string]interface{}, error)
    
    // Engine-specific adapters
    RegisterAdapter(engine string, adapter TypeAdapter) error
}
```

### 2. Engine Implementations

#### Lua Engine (First Implementation)
```
/pkg/engine/lua/
├── engine.go          # Implements ScriptEngine interface
├── converter.go       # Lua-specific type conversions
├── sandbox.go         # Lua security sandbox
├── stdlib/            # Lua standard library
│   ├── base.lua       # Core utilities
│   ├── async.lua      # Promise/async support
│   └── compat.lua     # Cross-engine compatibility
└── adapter.go         # Adapts GopherLua to our interface
```

#### JavaScript Engine (Second Implementation)
```
/pkg/engine/javascript/
├── engine.go          # Implements ScriptEngine interface
├── converter.go       # JS-specific type conversions
├── sandbox.go         # JS security sandbox
├── stdlib/            # JavaScript standard library
│   ├── base.js        # Core utilities
│   ├── async.js       # Native promise support
│   └── compat.js      # Cross-engine compatibility
├── adapter.go         # Adapts Goja to our interface
└── modules.go         # ES6 module support
```

#### Tengo Engine (Third Implementation)
```
/pkg/engine/tengo/
├── engine.go          # Implements ScriptEngine interface
├── converter.go       # Tengo-specific type conversions
├── sandbox.go         # Tengo security sandbox
├── stdlib/            # Tengo standard library
│   ├── base.tengo     # Core utilities
│   ├── async.tengo    # Async support
│   └── compat.tengo   # Cross-engine compatibility
└── adapter.go         # Adapts Tengo to our interface
```

### 3. Language-Agnostic Core

#### Core System Architecture
```
/pkg/core/
├── agent/              # Agent system (engine-agnostic)
│   ├── interface.go    # Agent interface
│   ├── base.go         # Base implementation
│   ├── registry.go     # Global registry
│   └── context.go      # Execution context
├── state/              # State management
│   ├── state.go        # Immutable state
│   ├── operations.go   # State transformations
│   ├── persistence.go  # Storage backends
│   └── sharing.go      # Inter-agent sharing
├── workflow/           # Workflow engine
│   ├── interface.go    # Workflow interface
│   ├── sequential.go   # Sequential execution
│   ├── parallel.go     # Parallel execution
│   ├── conditional.go  # Conditional branching
│   └── loop.go         # Iterative execution
├── event/              # Event system
│   ├── emitter.go      # Event emission
│   ├── handler.go      # Event handling
│   └── stream.go       # Event streaming
└── bridge/             # Bridge system
    ├── manager.go      # Bridge lifecycle
    ├── registry.go     # Bridge registry
    └── types.go        # Common types
```

#### Unified Bridge Layer
```
/pkg/bridge/
├── agent.go            # Agent bridge (works with all engines)
├── state.go            # State bridge
├── workflow.go         # Workflow bridge
├── event.go            # Event bridge
├── tool.go             # Tool bridge
└── converter.go        # Type conversion utilities
```

### 4. Cross-Engine Compatibility

#### Compatibility Layer
```
/pkg/compat/
├── api.go              # Common API definitions
├── types.go            # Shared type definitions
├── polyfill/           # Engine-specific polyfills
│   ├── lua.go          # Lua compatibility shims
│   ├── javascript.go   # JS compatibility shims
│   └── tengo.go        # Tengo compatibility shims
└── transpiler/         # Optional cross-compilation
    ├── ast.go          # Common AST
    └── converter.go    # Basic transpilation
```

#### Standard Library Design
```
/pkg/stdlib/
├── spec/               # Language-agnostic specifications
│   ├── agent.yaml      # Agent API spec
│   ├── state.yaml      # State API spec
│   ├── workflow.yaml   # Workflow API spec
│   └── tools.yaml      # Tools API spec
├── generator/          # Generates language-specific implementations
│   ├── lua.go          # Lua code generator
│   ├── javascript.go   # JS code generator
│   └── tengo.go        # Tengo code generator
└── tests/              # Cross-engine test suite
    └── conformance/    # Ensures all engines behave the same
```

## Implementation Plan

### Phase 1: Engine-Agnostic Foundation (Weeks 1-2)

#### 1.1 Script Engine Interface
**Task 1.1.1: Define Core Interfaces** ⭐ PRIORITY
- [ ] Create `/pkg/engine/interface.go`
- [ ] Define ScriptEngine interface
- [ ] Define Bridge interface
- [ ] Define TypeConverter interface
- [ ] Create EngineConfig structure

**Task 1.1.2: Engine Registry**
- [ ] Create `/pkg/engine/registry.go`
- [ ] Implement engine registration system
- [ ] Add engine discovery mechanism
- [ ] Support runtime engine switching
- [ ] Create engine factory pattern

**Task 1.1.3: Type System Foundation**
- [ ] Create `/pkg/engine/types.go`
- [ ] Define common type representations
- [ ] Create type mapping system
- [ ] Implement type validation
- [ ] Design error handling for type mismatches

**Task 1.1.4: Bridge Manager**
- [ ] Create `/pkg/bridge/manager.go`
- [ ] Implement bridge lifecycle management
- [ ] Create bridge registration system
- [ ] Add bridge dependency resolution
- [ ] Support hot-reloading of bridges

#### 1.2 Core Agent System (Engine-Agnostic)
**Task 1.2.1: Agent Interface** ⭐ PRIORITY
- [ ] Create `/pkg/core/agent/interface.go`
- [ ] Define lifecycle methods (init, run, cleanup)
- [ ] Add metadata and capability declaration
- [ ] Design extension points for custom agents
- [ ] Ensure engine independence

**Task 1.2.2: Base Agent Implementation**
- [ ] Create `/pkg/core/agent/base.go`
- [ ] Implement state management methods
- [ ] Add event emission capabilities
- [ ] Implement error handling and recovery
- [ ] Create agent metrics collection

**Task 1.2.3: Agent Registry**
- [ ] Create `/pkg/core/agent/registry.go`
- [ ] Implement thread-safe agent registration
- [ ] Add capability-based agent discovery
- [ ] Support dynamic agent lifecycle management
- [ ] Create agent templating system

**Task 1.2.4: Agent Context**
- [ ] Create `/pkg/core/agent/context.go`
- [ ] Design execution context with resource limits
- [ ] Add cancellation and timeout support
- [ ] Integrate distributed tracing
- [ ] Support multi-engine execution

#### 1.3 State Management System
**Task 1.3.1: State Object Design** ⭐ PRIORITY
- [ ] Create `/pkg/core/state/state.go`
- [ ] Implement immutable state operations
- [ ] Add metadata layer support
- [ ] Design artifact management system
- [ ] Implement state history tracking

**Task 1.3.2: State Operations**
- [ ] Create `/pkg/core/state/operations.go`
- [ ] Implement transformation methods
- [ ] Add state validation framework
- [ ] Design merge strategies
- [ ] Add serialization for all engines

**Task 1.3.3: State Persistence**
- [ ] Create `/pkg/core/state/persistence.go`
- [ ] Define persistence interface
- [ ] Implement memory store
- [ ] Add file-based store
- [ ] Design cloud storage adapters

**Task 1.3.4: State Sharing**
- [ ] Create `/pkg/core/state/sharing.go`
- [ ] Implement inter-agent state sharing
- [ ] Add state isolation mechanisms
- [ ] Design access control system
- [ ] Handle concurrent state access

#### 1.4 Universal Bridge System
**Task 1.4.1: Agent Bridge** ⭐ PRIORITY
- [ ] Create `/pkg/bridge/agent.go`
- [ ] Implement engine-agnostic agent bridge
- [ ] Add type conversion layer
- [ ] Support all agent operations
- [ ] Create comprehensive tests

**Task 1.4.2: State Bridge**
- [ ] Create `/pkg/bridge/state.go`
- [ ] Implement engine-agnostic state bridge
- [ ] Handle complex type conversions
- [ ] Support metadata operations
- [ ] Add performance optimizations

**Task 1.4.3: Event Bridge**
- [ ] Create `/pkg/bridge/event.go`
- [ ] Implement engine-agnostic event bridge
- [ ] Support async event handling
- [ ] Add event filtering
- [ ] Create event batching

**Task 1.4.4: Workflow Bridge**
- [ ] Create `/pkg/bridge/workflow.go`
- [ ] Implement workflow operations
- [ ] Support all workflow types
- [ ] Add workflow composition
- [ ] Handle workflow state

### Phase 2: Lua Engine Implementation (Weeks 3-4)

#### 2.1 Lua Engine Core
**Task 2.1.1: Engine Implementation** ⭐ PRIORITY
- [ ] Create `/pkg/engine/lua/engine.go`
- [ ] Implement ScriptEngine interface for Lua
- [ ] Integrate GopherLua
- [ ] Add Lua-specific optimizations
- [ ] Implement resource limits

**Task 2.1.2: Type Converter**
- [ ] Create `/pkg/engine/lua/converter.go`
- [ ] Implement Lua ↔ Go type conversions
- [ ] Handle Lua tables → Go maps/arrays
- [ ] Support userdata conversions
- [ ] Optimize for performance

**Task 2.1.3: Security Sandbox**
- [ ] Create `/pkg/engine/lua/sandbox.go`
- [ ] Disable dangerous Lua functions
- [ ] Implement file system restrictions
- [ ] Add network access control
- [ ] Create resource quotas

**Task 2.1.4: Lua Adapter**
- [ ] Create `/pkg/engine/lua/adapter.go`
- [ ] Adapt GopherLua to ScriptEngine interface
- [ ] Handle Lua-specific features
- [ ] Implement error mapping
- [ ] Add performance monitoring

#### 2.2 Lua Standard Library
**Task 2.2.1: Core Module**
- [ ] Create `/pkg/engine/lua/stdlib/core.lua`
- [ ] Implement basic utilities
- [ ] Add type checking functions
- [ ] Create debugging helpers
- [ ] Include performance utilities

**Task 2.2.2: Async Module**
- [ ] Create `/pkg/engine/lua/stdlib/async.lua`
- [ ] Implement Promise-like API
- [ ] Add async/await patterns
- [ ] Support coroutines
- [ ] Create timer functions

**Task 2.2.3: Agent Module**
- [ ] Create `/pkg/engine/lua/stdlib/agent.lua`
- [ ] Wrap agent bridge for Lua
- [ ] Add Lua-idiomatic API
- [ ] Support method chaining
- [ ] Include helper functions

**Task 2.2.4: Workflow Module**
- [ ] Create `/pkg/engine/lua/stdlib/workflow.lua`
- [ ] Implement workflow builders
- [ ] Add DSL for workflows
- [ ] Support composition
- [ ] Create debugging tools

### Phase 3: Workflow System (Weeks 5-6)

#### 3.1 Engine-Agnostic Workflow Engine
**Task 3.1.1: Workflow Interface** ⭐ PRIORITY
- [ ] Create `/pkg/core/workflow/interface.go`
- [ ] Define workflow step interface
- [ ] Support multiple execution strategies
- [ ] Add workflow metadata
- [ ] Design extension points

**Task 3.1.2: Sequential Workflow**
- [ ] Create `/pkg/core/workflow/sequential.go`
- [ ] Implement ordered execution
- [ ] Add state passing
- [ ] Support error handling
- [ ] Include retry logic

**Task 3.1.3: Parallel Workflow**
- [ ] Create `/pkg/core/workflow/parallel.go`
- [ ] Implement concurrent execution
- [ ] Add synchronization
- [ ] Support merge strategies
- [ ] Handle partial failures

**Task 3.1.4: Conditional Workflow**
- [ ] Create `/pkg/core/workflow/conditional.go`
- [ ] Implement branch evaluation
- [ ] Support multiple conditions
- [ ] Add default branches
- [ ] Enable dynamic conditions

**Task 3.1.5: Loop Workflow**
- [ ] Create `/pkg/core/workflow/loop.go`
- [ ] Implement iteration logic
- [ ] Add break conditions
- [ ] Support state accumulation
- [ ] Handle infinite loop prevention

#### 3.2 Workflow Runtime
**Task 3.2.1: Execution Engine**
- [ ] Create `/pkg/runtime/executor/engine.go`
- [ ] Implement workflow scheduler
- [ ] Add resource management
- [ ] Support cancellation
- [ ] Enable monitoring

**Task 3.2.2: State Management**
- [ ] Implement checkpointing
- [ ] Add state versioning
- [ ] Support rollback
- [ ] Handle state conflicts
- [ ] Enable state inspection

**Task 3.2.3: Error Handling**
- [ ] Design error strategies
- [ ] Implement compensation
- [ ] Add saga patterns
- [ ] Support circuit breakers
- [ ] Create error analytics

### Phase 4: JavaScript Engine Implementation (Weeks 7-8)

#### 4.1 JavaScript Engine Core
**Task 4.1.1: Engine Implementation** ⭐ PRIORITY
- [ ] Create `/pkg/engine/javascript/engine.go`
- [ ] Implement ScriptEngine interface for JS
- [ ] Integrate Goja
- [ ] Add ES6+ support
- [ ] Implement module system

**Task 4.1.2: Type Converter**
- [ ] Create `/pkg/engine/javascript/converter.go`
- [ ] Implement JS ↔ Go type conversions
- [ ] Handle JS objects → Go structs
- [ ] Support Promise conversions
- [ ] Optimize for Goja

**Task 4.1.3: Security Sandbox**
- [ ] Create `/pkg/engine/javascript/sandbox.go`
- [ ] Restrict global access
- [ ] Implement CSP-like policies
- [ ] Add resource limits
- [ ] Control prototype access

**Task 4.1.4: Module System**
- [ ] Create `/pkg/engine/javascript/modules.go`
- [ ] Implement CommonJS support
- [ ] Add ES6 module support
- [ ] Create module loader
- [ ] Support npm-like packages

#### 4.2 JavaScript Standard Library
**Task 4.2.1: Core Module**
- [ ] Create `/pkg/engine/javascript/stdlib/core.js`
- [ ] Implement utilities
- [ ] Add polyfills
- [ ] Create type helpers
- [ ] Include debugging tools

**Task 4.2.2: Async Module**
- [ ] Create `/pkg/engine/javascript/stdlib/async.js`
- [ ] Leverage native Promises
- [ ] Add async/await support
- [ ] Implement observables
- [ ] Create reactive patterns

**Task 4.2.3: Agent Module**
- [ ] Create `/pkg/engine/javascript/stdlib/agent.js`
- [ ] Implement JS-idiomatic API
- [ ] Support class-based agents
- [ ] Add decorators
- [ ] Include TypeScript definitions

**Task 4.2.4: Workflow Module**
- [ ] Create `/pkg/engine/javascript/stdlib/workflow.js`
- [ ] Implement fluent API
- [ ] Add JSX-like syntax
- [ ] Support functional composition
- [ ] Create visual debugger

### Phase 5: Built-in Components (Weeks 9-10)

#### 5.1 Engine-Agnostic Components
**Task 5.1.1: LLM Agents** ⭐ PRIORITY
- [ ] Create `/pkg/components/agents/llm/`
- [ ] Implement provider abstraction
- [ ] Add model selection
- [ ] Support streaming
- [ ] Create prompt templates

**Task 5.1.2: Tool Library**
- [ ] Create `/pkg/components/tools/`
- [ ] Implement core tools
- [ ] Add tool discovery
- [ ] Support tool composition
- [ ] Create tool testing framework

**Task 5.1.3: Workflow Templates**
- [ ] Create `/pkg/components/workflows/`
- [ ] Build common patterns
- [ ] Add parameterization
- [ ] Support inheritance
- [ ] Create template gallery

**Task 5.1.4: Cross-Engine Examples**
- [ ] Create examples for each engine
- [ ] Show language-specific features
- [ ] Demonstrate portability
- [ ] Include benchmarks
- [ ] Add migration guides

### Phase 6: Tengo Engine Implementation (Weeks 11-12)

#### 6.1 Tengo Engine Core
**Task 6.1.1: Engine Implementation**
- [ ] Create `/pkg/engine/tengo/engine.go`
- [ ] Implement ScriptEngine interface for Tengo
- [ ] Integrate Tengo VM
- [ ] Add Tengo-specific features
- [ ] Optimize for performance

**Task 6.1.2: Type Converter**
- [ ] Create `/pkg/engine/tengo/converter.go`
- [ ] Implement Tengo ↔ Go conversions
- [ ] Handle Tengo objects
- [ ] Support compiled scripts
- [ ] Add type validation

**Task 6.1.3: Security Sandbox**
- [ ] Create `/pkg/engine/tengo/sandbox.go`
- [ ] Implement Tengo restrictions
- [ ] Add import controls
- [ ] Limit built-ins
- [ ] Control execution time

**Task 6.1.4: Tengo Adapter**
- [ ] Create `/pkg/engine/tengo/adapter.go`
- [ ] Adapt Tengo to interface
- [ ] Handle compilation
- [ ] Support hot reload
- [ ] Add debugging support

#### 6.2 Cross-Engine Testing
**Task 6.2.1: Conformance Suite**
- [ ] Create `/pkg/test/conformance/`
- [ ] Test all engines equally
- [ ] Verify API compatibility
- [ ] Check performance
- [ ] Validate behavior

**Task 6.2.2: Integration Tests**
- [ ] Test cross-engine workflows
- [ ] Verify type conversions
- [ ] Check error handling
- [ ] Test resource limits
- [ ] Validate security

### Phase 7: Production Features (Week 13-14)

#### 7.1 Multi-Engine Runtime
**Task 7.1.1: Engine Manager**
- [ ] Create `/pkg/runtime/manager.go`
- [ ] Implement engine pooling
- [ ] Add load balancing
- [ ] Support hot swapping
- [ ] Enable A/B testing

**Task 7.1.2: Script Router**
- [ ] Create `/pkg/runtime/router.go`
- [ ] Route by file extension
- [ ] Support shebang detection
- [ ] Add performance routing
- [ ] Enable feature routing

**Task 7.1.3: Cross-Engine State**
- [ ] Implement state sharing
- [ ] Add type preservation
- [ ] Support state migration
- [ ] Handle engine failures
- [ ] Enable debugging

#### 7.2 Observability
**Task 7.2.1: Unified Metrics**
- [ ] Create engine metrics
- [ ] Add performance tracking
- [ ] Support custom metrics
- [ ] Enable comparison
- [ ] Build dashboards

**Task 7.2.2: Distributed Tracing**
- [ ] Trace across engines
- [ ] Add engine metadata
- [ ] Support correlation
- [ ] Enable sampling
- [ ] Create analysis tools

### Phase 8: Advanced Features (Week 15-16)

#### 8.1 Advanced Patterns
**Task 8.1.1: Polyglot Workflows**
- [ ] Mix engines in workflows
- [ ] Optimize engine selection
- [ ] Support transitions
- [ ] Handle failures
- [ ] Enable debugging

**Task 8.1.2: Engine Plugins**
- [ ] Create plugin system
- [ ] Support custom engines
- [ ] Add engine marketplace
- [ ] Enable community engines
- [ ] Create certification

#### 8.2 Developer Experience
**Task 8.2.1: Unified CLI**
- [ ] Support all engines
- [ ] Add engine management
- [ ] Include migration tools
- [ ] Support debugging
- [ ] Enable profiling

**Task 8.2.2: IDE Support**
- [ ] Multi-language highlighting
- [ ] Cross-engine navigation
- [ ] Unified debugging
- [ ] Performance profiling
- [ ] Integrated testing

## Engine-Specific API Examples

### Common API (All Engines)
```lua
-- Lua
local agent = Agent.new({name = "analyzer", model = "gpt-4"})
local state = State.new({data = "..."})
local result = agent:run(state)
```

```javascript
// JavaScript
const agent = new Agent({name: "analyzer", model: "gpt-4"});
const state = new State({data: "..."});
const result = await agent.run(state);
```

```go
// Tengo
agent := agent.new({name: "analyzer", model: "gpt-4"})
state := state.new({data: "..."})
result := agent.run(state)
```

### Engine-Specific Features
```lua
-- Lua: Coroutine-based async
local function async_workflow()
    local results = {}
    for i = 1, 3 do
        coroutine.yield(agent:run_async(state))
    end
    return results
end
```

```javascript
// JavaScript: Native async/await
async function asyncWorkflow() {
    const results = await Promise.all([
        agent1.run(state),
        agent2.run(state),
        agent3.run(state)
    ]);
    return results;
}
```

```go
// Tengo: Channel-based concurrency
results := []
ch := make_chan(3)
for i := 0; i < 3; i++ {
    go func() { 
        ch <- agent.run(state) 
    }()
}
for i := 0; i < 3; i++ {
    results = append(results, <-ch)
}
```

## Migration Strategy

### 1. Phased Rollout
1. **Phase 1**: Lua engine with full features
2. **Phase 2**: JavaScript engine for web developers
3. **Phase 3**: Tengo for Go developers
4. **Phase 4**: Community engines

### 2. Compatibility Guarantees
- Core API identical across engines
- Engine-specific features clearly marked
- Automatic polyfills where possible
- Clear migration paths between engines

### 3. Performance Considerations
- Lua: Fastest startup, lowest memory
- JavaScript: Best async support, familiar syntax
- Tengo: Best Go integration, compiled scripts

## Testing Strategy

### 1. Conformance Testing
- Single test suite runs on all engines
- Validates identical behavior
- Checks performance characteristics
- Ensures security compliance

### 2. Engine-Specific Testing
- Tests for unique features
- Performance benchmarks
- Security penetration tests
- Resource limit validation

### 3. Cross-Engine Testing
- State sharing between engines
- Workflow handoff
- Type preservation
- Error propagation

## Success Metrics

### 1. Developer Experience
- Same script runs on 3+ engines: 95% compatibility
- Engine switch time: < 5 minutes
- Learning curve: 1 engine → all engines in 1 hour
- Documentation: 100% coverage all engines

### 2. Performance
- Lua: < 10ms startup, < 10MB memory
- JavaScript: < 50ms startup, < 50MB memory
- Tengo: < 5ms startup, < 20MB memory
- Cross-engine overhead: < 5%

### 3. Adoption
- 3 production engines in 6 months
- 5+ community engines in 1 year
- 80% scripts portable across engines
- Active multi-engine deployments: 50+

## Conclusion

This multi-engine architecture transforms go-llmspell into a truly polyglot AI orchestration platform. By designing around a language-agnostic core with pluggable script engines, we enable developers to choose the best language for their needs while maintaining access to the full power of go-llms v0.3.2's multi-agent capabilities.

The phased approach ensures each engine is production-ready before moving to the next, while the common API guarantees that investments in one engine transfer to others. This positions go-llmspell as the most flexible and powerful AI orchestration platform available.