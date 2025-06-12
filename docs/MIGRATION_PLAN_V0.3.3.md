# Go-LLMSpell Clean Slate Migration Plan: Multi-Engine Architecture for go-llms v0.3.3

## Introduction

### What is go-llmspell?

go-llmspell is a Go library that provides a **scriptable interface for LLM interactions** using embedded scripting languages. It acts as a wrapper around the go-llms library, providing scripting capabilities for AI agent orchestration and workflow automation. The key innovation is "spells" - scripts written in Lua, JavaScript, or Tengo that control LLMs without needing to compile Go code.

### Why This Migration?

We are migrating to go-llms v0.3.3 to leverage its significant architectural improvements:

1. **Advanced Multi-Agent Orchestration** - v0.3.3 provides enhanced capabilities for coordinating multiple AI agents
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

This migration to v0.3.3 represents a fundamental shift to fully leverage the latest go-llms capabilities while maintaining the ease of scripting that makes go-llmspell valuable for AI developers.

## Executive Summary

This document outlines a complete redesign of go-llmspell to fully leverage go-llms v0.3.3's advanced multi-agent orchestration capabilities while supporting multiple scripting languages. The architecture is designed around a language-agnostic scripting engine interface, with Lua (GopherLua) as the first implementation, followed by JavaScript (Goja) and Tengo.

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
├── llm.go              # LLM provider bridge (pkg/llm)
├── builtins.go         # Built-in tools bridge (pkg/agent/builtins)
├── schema.go           # Schema validation bridge (pkg/schema)
├── structured.go       # Structured output bridge (pkg/structured)
├── util.go             # Utilities bridge (pkg/util)
├── modelinfo.go        # Model info bridge (pkg/util/llmutil/modelinfo)
├── logging.go          # Logging bridge (pkg/internal/debug + slog)
├── hooks.go            # Hook bridge (pkg/agent/domain Hook interface)
├── events.go           # Event bridge (pkg/agent/domain event system)
├── manager.go          # Bridge lifecycle management
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
**Task 1.1.1: Define Core Interfaces**
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
**Task 1.2.1: Agent Interface**
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
**Task 1.3.1: State Object Design**
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
**Task 1.4.1: Agent Bridge**
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

**Task 1.4.5: LLM Bridge**
- [ ] Create `/pkg/bridge/llm.go`
- [ ] Bridge pkg/llm provider interfaces
- [ ] Expose message handling and options
- [ ] Support provider switching and pooling
- [ ] Add streaming and non-streaming responses

**Task 1.4.6: Built-in Tools Bridge**
- [ ] Create `/pkg/bridge/builtins.go`
- [ ] Expose pkg/agent/builtins/tools
- [ ] Bridge file, web, data, datetime tools
- [ ] Add math, system, and feed tools
- [ ] Support tool discovery and registration

**Task 1.4.7: Schema Bridge**
- [ ] Create `/pkg/bridge/schema.go`
- [ ] Bridge pkg/schema validation system
- [ ] Expose reflection-based generation
- [ ] Add coercion and validation utilities
- [ ] Support custom validator registration

**Task 1.4.8: Structured Output Bridge**
- [ ] Create `/pkg/bridge/structured.go`
- [ ] Bridge pkg/structured processing
- [ ] Expose JSON extraction utilities
- [ ] Add prompt enhancement features
- [ ] Support schema caching

**Task 1.4.9: Utilities Bridge**
- [ ] Create `/pkg/bridge/util.go`
- [ ] Bridge core pkg/util functions
- [ ] Expose JSON utilities and helpers
- [ ] Add environment variable access
- [ ] Include auth and metrics utilities

**Task 1.4.10: Model Info Bridge**
- [ ] Create `/pkg/bridge/modelinfo.go`
- [ ] Bridge pkg/util/llmutil/modelinfo
- [ ] Expose model inventory and discovery
- [ ] Add provider-specific model fetchers
- [ ] Include caching and service interfaces

**Task 1.4.11: Logging Bridge**
- [ ] Create `/pkg/bridge/logging.go`
- [ ] Bridge pkg/internal/debug logging system with structured log/slog integration
- [ ] Support script-agnostic logging (info, warn, error, debug) via slog interface
- [ ] Integrate with component-based debug filtering using DebugComponent constants
- [ ] Enable structured logging with metadata using slog.With() and slog.Group()
- [ ] Support dynamic log level changes per component (e.g., slog.Debug("Component.Agent", ...))
- [ ] Provide script access to debug.IsEnabled() for conditional debug output
- [ ] Ensure thread-safety across engines with proper context propagation
- [ ] Bridge LoggingHook integration for automatic agent conversation logging

**Task 1.4.12: Hook Bridge**
- [ ] Create `/pkg/bridge/hooks.go`
- [ ] Bridge pkg/agent/domain Hook interface with full method support
- [ ] Support BeforeGenerate(ctx, request) and AfterGenerate(ctx, request, response) hooks
- [ ] Support BeforeToolCall(ctx, tool, input) and AfterToolCall(ctx, tool, input, output) hooks
- [ ] Enable agent lifecycle hooks via OnStart(ctx, agent) and OnStop(ctx, agent)
- [ ] Allow multiple script-based hooks with priority ordering and chain execution
- [ ] Integrate with built-in LoggingHook and MetricsHook for automatic instrumentation
- [ ] Support context propagation and metadata passing between hooks
- [ ] Enable conditional hook execution based on agent state and request properties
- [ ] Provide hook registry for dynamic registration/deregistration from scripts
- [ ] Support async hooks with proper error handling and timeout management

**Task 1.4.13: Event Bridge**
- [ ] Create `/pkg/bridge/events.go`
- [ ] Bridge pkg/agent/domain event system
- [ ] Support real-time event streaming to scripts
- [ ] Enable event filtering and subscription by type
- [ ] Handle lifecycle, execution, tool, and workflow events
- [ ] Support event metadata and correlation

**Task 1.4.14: Tracing Bridge**
- [ ] Create `/pkg/bridge/tracing.go`
- [ ] Bridge core/tracing.go distributed tracing system
- [ ] Support OpenTelemetry span creation and management
- [ ] Enable trace correlation across agents and tools
- [ ] Provide span annotation and attribute setting
- [ ] Support trace sampling and export configuration
- [ ] Integrate with agent execution context

**Task 1.4.15: Event Utilities Bridge**
- [ ] Create `/pkg/bridge/event_utils.go`
- [ ] Bridge event utility functions and helpers
- [ ] Support event transformation and filtering
- [ ] Enable event batching and aggregation
- [ ] Provide event pattern matching utilities
- [ ] Support event correlation and causality tracking
- [ ] Include event replay and debugging tools

**Task 1.4.16: State Utilities Bridge**
- [ ] Create `/pkg/bridge/state_utils.go`
- [ ] Bridge state utility functions and helpers
- [ ] Support state validation and transformation
- [ ] Enable state diff and merge operations
- [ ] Provide state serialization utilities
- [ ] Support state migration and versioning
- [ ] Include state debugging and inspection tools

**Task 1.4.17: Artifact Bridge**
- [ ] Create `/pkg/bridge/artifact.go`
- [ ] Bridge artifact.go agent artifact management
- [ ] Support file and data artifact creation
- [ ] Enable artifact sharing between agents
- [ ] Provide artifact versioning and metadata
- [ ] Support artifact storage backends (local, cloud)
- [ ] Include artifact lifecycle management

**Task 1.4.18: Tool Context Bridge**
- [ ] Create `/pkg/bridge/tool_context.go`
- [ ] Bridge tool execution context system
- [ ] Support context propagation to tools
- [ ] Enable tool metadata and configuration access
- [ ] Provide tool resource limits and monitoring
- [ ] Support tool cancellation and timeout
- [ ] Include tool error handling and recovery

**Task 1.4.19: Agent Handoff Bridge**
- [ ] Create `/pkg/bridge/handoff.go`
- [ ] Bridge handoff.go agent handoff system
- [ ] Support agent-to-agent state transfer
- [ ] Enable handoff condition evaluation
- [ ] Provide handoff metadata and context
- [ ] Support handoff validation and rollback
- [ ] Include handoff monitoring and debugging

**Task 1.4.20: Guardrails Bridge**
- [ ] Create `/pkg/bridge/guardrails.go`
- [ ] Bridge guardrails.go agent safety system
- [ ] Support content filtering and validation
- [ ] Enable behavioral constraint enforcement
- [ ] Provide safety policy configuration
- [ ] Support custom guardrail implementation
- [ ] Include guardrail violation reporting

**Task 1.4.21: Tool Event Emitter Bridge**
- [ ] Create `/pkg/bridge/tool_events.go`
- [ ] Bridge tool event emission system
- [ ] Support tool execution event streaming
- [ ] Enable tool performance monitoring
- [ ] Provide tool error event handling
- [ ] Support tool lifecycle events
- [ ] Include tool usage analytics

**Task 1.4.22: Memory Management Bridge** ⏸️ **[DEFERRED - Awaiting go-llms implementation]**
- [ ] Create `/pkg/bridge/memory.go`
- [ ] Bridge agent memory management system
- [ ] Support short-term and long-term memory
- [ ] Enable memory persistence and retrieval
- [ ] Provide memory search and indexing
- [ ] Support memory compression and optimization
- [ ] Include memory debugging and inspection
- [ ] **NOTE: Memory subsystem not yet implemented in go-llms v0.3.3**

**Task 1.4.23: Conversation Bridge**
- [ ] Create `/pkg/bridge/conversation.go`
- [ ] Bridge conversation management system
- [ ] Support multi-turn conversation handling
- [ ] Enable conversation state persistence
- [ ] Provide conversation branching and merging
- [ ] Support conversation templates and patterns
- [ ] Include conversation analytics and insights

**Task 1.4.24: Model Management Bridge**
- [ ] Create `/pkg/bridge/model_mgmt.go`
- [ ] Bridge dynamic model management system
- [ ] Support runtime model switching
- [ ] Enable model performance monitoring
- [ ] Provide model capability discovery
- [ ] Support model pooling and load balancing
- [ ] Include model cost optimization

**Task 1.4.25: Provider Pooling Bridge**
- [ ] Create `/pkg/bridge/provider_pool.go`
- [ ] Bridge provider connection pooling system
- [ ] Support connection lifecycle management
- [ ] Enable load balancing across providers
- [ ] Provide connection health monitoring
- [ ] Support failover and redundancy
- [ ] Include connection performance metrics

**Task 1.4.26: Resilience Bridge**
- [ ] Create `/pkg/bridge/resilience.go`
- [ ] Bridge retry and circuit breaker patterns
- [ ] Support configurable retry policies
- [ ] Enable circuit breaker state management
- [ ] Provide timeout and deadline handling
- [ ] Support rate limiting and throttling
- [ ] Include resilience pattern monitoring

**Task 1.4.27: Collaboration Bridge**
- [ ] Create `/pkg/bridge/collaboration.go`
- [ ] Bridge multi-agent collaboration system
- [ ] Support agent coordination patterns
- [ ] Enable agent communication protocols
- [ ] Provide collaboration state management
- [ ] Support collaborative workflow execution
- [ ] Include collaboration monitoring and debugging

**Task 1.4.28: Security Bridge**
- [ ] Create `/pkg/bridge/security.go`
- [ ] Bridge authentication and authorization system
- [ ] Support user and agent identity management
- [ ] Enable permission and role-based access
- [ ] Provide security policy enforcement
- [ ] Support audit logging and compliance
- [ ] Include security threat detection

**Task 1.4.29: Metrics Bridge**
- [ ] Create `/pkg/bridge/metrics.go`
- [ ] Bridge performance and usage metrics system
- [ ] Support custom metric collection
- [ ] Enable metric aggregation and reporting
- [ ] Provide metric alerting and notification
- [ ] Support metric visualization and dashboards
- [ ] Include metric-based optimization

### Phase 2: Lua Engine Implementation (Weeks 3-4)

#### 2.1 Lua Engine Core
**Task 2.1.1: Engine Implementation**
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

**Task 2.2.5: Logging Module**
- [ ] Create `/pkg/engine/lua/stdlib/log.lua`
- [ ] Wrap logging bridge for Lua with idiomatic API
- [ ] Expose log.info(), log.warn(), log.error(), log.debug() with table-based structured args
- [ ] Support component-based debug logging with log.debug_for(component, message, attrs)
- [ ] Enable structured logging with metadata via log.with(attrs) and log.group(name, fn)
- [ ] Provide log.set_level(level) and log.is_enabled(component) for dynamic control
- [ ] Include thread-safe logging utilities with automatic context propagation
- [ ] Support Lua table serialization for complex log data structures
- [ ] Add log.capture() and log.replay() for testing and debugging
- [ ] Integrate with LoggingHook for automatic agent conversation logging

**Task 2.2.6: Hooks Module**
- [ ] Create `/pkg/engine/lua/stdlib/hooks.lua`
- [ ] Wrap hook bridge for Lua with function-based registration
- [ ] Expose hooks.before_generate(fn) and hooks.after_generate(fn) registration
- [ ] Add hooks.before_tool(fn) and hooks.after_tool(fn) hook registration
- [ ] Support agent lifecycle hooks via hooks.on_start(fn) and hooks.on_stop(fn)
- [ ] Enable multiple hook registration with priority ordering
- [ ] Support conditional hook execution with predicate functions
- [ ] Provide hooks.remove(id) and hooks.clear() for dynamic management
- [ ] Handle async hooks with coroutine support and error handling
- [ ] Integrate with built-in LoggingHook and MetricsHook automation

**Task 2.2.7: Tracing Module**
- [ ] Create `/pkg/engine/lua/stdlib/tracing.lua`
- [ ] Wrap tracing bridge for Lua with span management
- [ ] Expose trace.start_span(name, attrs) and trace.finish_span(span)
- [ ] Support trace correlation with trace.current_span() and trace.set_span(span)
- [ ] Enable span annotation with span:set_attribute(key, value)
- [ ] Provide trace sampling control with trace.set_sampling_rate(rate)
- [ ] Include trace export configuration and debugging utilities
- [ ] Support trace context propagation across agent boundaries

**Task 2.2.8: Event Utilities Module**
- [ ] Create `/pkg/engine/lua/stdlib/event_utils.lua`
- [ ] Wrap event utilities bridge for Lua
- [ ] Expose event transformation functions (filter, map, reduce)
- [ ] Support event batching with event_utils.batch(events, size)
- [ ] Enable event pattern matching with event_utils.match(pattern, event)
- [ ] Provide event correlation utilities
- [ ] Include event replay and debugging tools

**Task 2.2.9: State Utilities Module**
- [ ] Create `/pkg/engine/lua/stdlib/state_utils.lua`
- [ ] Wrap state utilities bridge for Lua
- [ ] Expose state validation functions
- [ ] Support state diff and merge operations
- [ ] Enable state serialization with state_utils.serialize(state, format)
- [ ] Provide state migration utilities
- [ ] Include state debugging and inspection tools

**Task 2.2.10: Artifacts Module**
- [ ] Create `/pkg/engine/lua/stdlib/artifacts.lua`
- [ ] Wrap artifact bridge for Lua with file/data management
- [ ] Expose artifacts.create(type, data, metadata) and artifacts.get(id)
- [ ] Support artifact sharing with artifacts.share(id, agent_id)
- [ ] Enable artifact versioning and metadata management
- [ ] Provide artifact storage backend configuration
- [ ] Include artifact lifecycle management utilities

**Task 2.2.11: Tool Context Module**
- [ ] Create `/pkg/engine/lua/stdlib/tool_context.lua`
- [ ] Wrap tool context bridge for Lua
- [ ] Expose context.get_metadata() and context.set_config(config)
- [ ] Support context propagation utilities
- [ ] Enable tool resource monitoring
- [ ] Provide cancellation and timeout handling
- [ ] Include error handling and recovery patterns

**Task 2.2.12: Handoff Module**
- [ ] Create `/pkg/engine/lua/stdlib/handoff.lua`
- [ ] Wrap agent handoff bridge for Lua
- [ ] Expose handoff.to(agent_id, state, conditions)
- [ ] Support handoff condition evaluation
- [ ] Enable handoff metadata and context management
- [ ] Provide handoff validation and rollback
- [ ] Include handoff monitoring and debugging

**Task 2.2.13: Guardrails Module**
- [ ] Create `/pkg/engine/lua/stdlib/guardrails.lua`
- [ ] Wrap guardrails bridge for Lua
- [ ] Expose guardrails.validate(content, policy) and guardrails.filter(content)
- [ ] Support behavioral constraint enforcement
- [ ] Enable safety policy configuration
- [ ] Provide custom guardrail implementation
- [ ] Include guardrail violation reporting

**Task 2.2.14: Memory Module** ⏸️ **[DEFERRED - Awaiting go-llms implementation]**
- [ ] Create `/pkg/engine/lua/stdlib/memory.lua`
- [ ] Wrap memory management bridge for Lua
- [ ] Expose memory.store(key, value) and memory.recall(key, query)
- [ ] Support short-term and long-term memory operations
- [ ] Enable memory search and indexing
- [ ] Provide memory compression and optimization
- [ ] Include memory debugging and inspection
- [ ] **NOTE: Memory subsystem not yet implemented in go-llms v0.3.3**

**Task 2.2.15: Conversation Module**
- [ ] Create `/pkg/engine/lua/stdlib/conversation.lua`
- [ ] Wrap conversation bridge for Lua
- [ ] Expose conversation.start(template) and conversation.continue(message)
- [ ] Support multi-turn conversation handling
- [ ] Enable conversation state persistence
- [ ] Provide conversation branching and merging
- [ ] Include conversation analytics and insights

**Task 2.2.16: Advanced Modules**
- [ ] Create `/pkg/engine/lua/stdlib/model_mgmt.lua` - Dynamic model management
- [ ] Create `/pkg/engine/lua/stdlib/provider_pool.lua` - Provider connection pooling
- [ ] Create `/pkg/engine/lua/stdlib/resilience.lua` - Retry and circuit breaker patterns
- [ ] Create `/pkg/engine/lua/stdlib/collaboration.lua` - Multi-agent collaboration
- [ ] Create `/pkg/engine/lua/stdlib/security.lua` - Authentication and authorization
- [ ] Create `/pkg/engine/lua/stdlib/metrics.lua` - Performance and usage metrics

### Phase 3: Workflow System (Weeks 5-6)

#### 3.1 Engine-Agnostic Workflow Engine
**Task 3.1.1: Workflow Interface**
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
**Task 4.1.1: Engine Implementation**
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

**Task 4.2.5: Logging Module**
- [ ] Create `/pkg/engine/javascript/stdlib/log.js`
- [ ] Wrap logging bridge for JavaScript with modern API
- [ ] Expose log.info(), log.warn(), log.error(), log.debug() with object-based structured args
- [ ] Support component-based debug logging with log.debugFor(component, message, metadata)
- [ ] Enable structured logging with metadata via log.with(attrs) and log.group(name, fn)
- [ ] Provide log.setLevel(level) and log.isEnabled(component) for dynamic control
- [ ] Include thread-safe logging utilities with automatic context propagation
- [ ] Support JSON serialization for complex log data structures
- [ ] Add log.capture() and log.replay() for testing and debugging
- [ ] Integrate with LoggingHook for automatic agent conversation logging

**Task 4.2.6: Hooks Module**
- [ ] Create `/pkg/engine/javascript/stdlib/hooks.js`
- [ ] Wrap hook bridge for JavaScript with event-emitter pattern
- [ ] Expose hooks.beforeGenerate(fn) and hooks.afterGenerate(fn) registration
- [ ] Add hooks.beforeTool(fn) and hooks.afterTool(fn) hook registration
- [ ] Support agent lifecycle hooks via hooks.onStart(fn) and hooks.onStop(fn)
- [ ] Enable multiple hook registration with priority ordering
- [ ] Support conditional hook execution with predicate functions
- [ ] Provide hooks.remove(id) and hooks.clear() for dynamic management
- [ ] Handle async hooks with native Promise support and error handling
- [ ] Integrate with built-in LoggingHook and MetricsHook automation

**Task 4.2.7: Tracing Module**
- [ ] Create `/pkg/engine/javascript/stdlib/tracing.js`
- [ ] Wrap tracing bridge for JavaScript with modern API
- [ ] Expose trace.startSpan(name, attrs) and trace.finishSpan(span)
- [ ] Support trace correlation with trace.currentSpan() and trace.setSpan(span)
- [ ] Enable span annotation with span.setAttribute(key, value)
- [ ] Provide trace sampling control with trace.setSamplingRate(rate)
- [ ] Include trace export configuration and debugging utilities
- [ ] Support trace context propagation with async/await patterns

**Task 4.2.8: Event Utilities Module**
- [ ] Create `/pkg/engine/javascript/stdlib/eventUtils.js`
- [ ] Wrap event utilities bridge for JavaScript
- [ ] Expose event transformation functions (filter, map, reduce)
- [ ] Support event batching with EventUtils.batch(events, size)
- [ ] Enable event pattern matching with EventUtils.match(pattern, event)
- [ ] Provide event correlation utilities with Promise chains
- [ ] Include event replay and debugging tools

**Task 4.2.9: State Utilities Module**
- [ ] Create `/pkg/engine/javascript/stdlib/stateUtils.js`
- [ ] Wrap state utilities bridge for JavaScript
- [ ] Expose state validation functions with Promise-based API
- [ ] Support state diff and merge operations
- [ ] Enable state serialization with StateUtils.serialize(state, format)
- [ ] Provide state migration utilities
- [ ] Include state debugging and inspection tools

**Task 4.2.10: Artifacts Module**
- [ ] Create `/pkg/engine/javascript/stdlib/artifacts.js`
- [ ] Wrap artifact bridge for JavaScript with async/await API
- [ ] Expose artifacts.create(type, data, metadata) and artifacts.get(id)
- [ ] Support artifact sharing with artifacts.share(id, agentId)
- [ ] Enable artifact versioning and metadata management
- [ ] Provide artifact storage backend configuration
- [ ] Include artifact lifecycle management utilities

**Task 4.2.11: Tool Context Module**
- [ ] Create `/pkg/engine/javascript/stdlib/toolContext.js`
- [ ] Wrap tool context bridge for JavaScript
- [ ] Expose context.getMetadata() and context.setConfig(config)
- [ ] Support context propagation utilities
- [ ] Enable tool resource monitoring
- [ ] Provide cancellation and timeout handling with AbortController
- [ ] Include error handling and recovery patterns

**Task 4.2.12: Handoff Module**
- [ ] Create `/pkg/engine/javascript/stdlib/handoff.js`
- [ ] Wrap agent handoff bridge for JavaScript
- [ ] Expose handoff.to(agentId, state, conditions) with Promise API
- [ ] Support handoff condition evaluation
- [ ] Enable handoff metadata and context management
- [ ] Provide handoff validation and rollback
- [ ] Include handoff monitoring and debugging

**Task 4.2.13: Guardrails Module**
- [ ] Create `/pkg/engine/javascript/stdlib/guardrails.js`
- [ ] Wrap guardrails bridge for JavaScript
- [ ] Expose guardrails.validate(content, policy) and guardrails.filter(content)
- [ ] Support behavioral constraint enforcement
- [ ] Enable safety policy configuration
- [ ] Provide custom guardrail implementation
- [ ] Include guardrail violation reporting

**Task 4.2.14: Memory Module** ⏸️ **[DEFERRED - Awaiting go-llms implementation]**
- [ ] Create `/pkg/engine/javascript/stdlib/memory.js`
- [ ] Wrap memory management bridge for JavaScript
- [ ] Expose memory.store(key, value) and memory.recall(key, query)
- [ ] Support short-term and long-term memory operations
- [ ] Enable memory search and indexing with Promise-based API
- [ ] Provide memory compression and optimization
- [ ] Include memory debugging and inspection
- [ ] **NOTE: Memory subsystem not yet implemented in go-llms v0.3.3**

**Task 4.2.15: Conversation Module**
- [ ] Create `/pkg/engine/javascript/stdlib/conversation.js`
- [ ] Wrap conversation bridge for JavaScript
- [ ] Expose conversation.start(template) and conversation.continue(message)
- [ ] Support multi-turn conversation handling with async/await
- [ ] Enable conversation state persistence
- [ ] Provide conversation branching and merging
- [ ] Include conversation analytics and insights

**Task 4.2.16: Advanced Modules**
- [ ] Create `/pkg/engine/javascript/stdlib/modelMgmt.js` - Dynamic model management
- [ ] Create `/pkg/engine/javascript/stdlib/providerPool.js` - Provider connection pooling
- [ ] Create `/pkg/engine/javascript/stdlib/resilience.js` - Retry and circuit breaker patterns
- [ ] Create `/pkg/engine/javascript/stdlib/collaboration.js` - Multi-agent collaboration
- [ ] Create `/pkg/engine/javascript/stdlib/security.js` - Authentication and authorization
- [ ] Create `/pkg/engine/javascript/stdlib/metrics.js` - Performance and usage metrics

### Phase 5: Agent Built-ins Integration (Weeks 9-10)

#### 5.1 Built-in Tool Categories
**Task 5.1.1: File System Tools**
- [ ] Expose pkg/agent/builtins/tools/file via bridges
- [ ] Bridge file_read, file_write, file_delete
- [ ] Add file_list, file_search, file_move
- [ ] Include permission and sandboxing
- [ ] Create comprehensive tests

**Task 5.1.2: Web and API Tools**
- [ ] Expose pkg/agent/builtins/tools/web via bridges
- [ ] Bridge web_fetch, web_scrape, web_search
- [ ] Add api_client, graphql, openapi tools
- [ ] Include authentication support
- [ ] Support rate limiting and caching

**Task 5.1.3: Data Processing Tools**
- [ ] Expose pkg/agent/builtins/tools/data via bridges
- [ ] Bridge csv_process, json_process, xml_process
- [ ] Add data_transform utilities
- [ ] Include format conversion
- [ ] Support large data handling

**Task 5.1.4: DateTime Tools**
- [ ] Expose pkg/agent/builtins/tools/datetime via bridges
- [ ] Bridge datetime_now, datetime_parse, datetime_format
- [ ] Add datetime_calculate, datetime_compare
- [ ] Include timezone support
- [ ] Support various date formats

**Task 5.1.5: Math and Calculation Tools**
- [ ] Expose pkg/agent/builtins/tools/math via bridges
- [ ] Bridge calculator tool with full functions
- [ ] Add mathematical constants
- [ ] Include statistical functions
- [ ] Support complex calculations

**Task 5.1.6: System Tools**
- [ ] Expose pkg/agent/builtins/tools/system via bridges
- [ ] Bridge env_var, system_info tools
- [ ] Add process_list, execute tools
- [ ] Include security restrictions
- [ ] Support cross-platform operations

**Task 5.1.7: Feed and Content Tools**
- [ ] Expose pkg/agent/builtins/tools/feed via bridges
- [ ] Bridge feed_fetch, feed_parse, feed_filter
- [ ] Add feed_discover, feed_aggregate
- [ ] Include content extraction
- [ ] Support multiple feed formats

#### 5.2 Built-in Agent Templates
**Task 5.2.1: Agent Registry Integration**
- [ ] Expose pkg/agent/builtins/agents registry via bridges
- [ ] Bridge pre-built agent templates
- [ ] Add provider-specific optimizations
- [ ] Support streaming patterns
- [ ] Include model selection helpers

#### 5.3 Built-in Workflow Patterns
**Task 5.3.1: Workflow Registry Integration**
- [ ] Expose pkg/agent/builtins/workflows registry via bridges
- [ ] Bridge common workflow patterns
- [ ] Add parameterization support
- [ ] Include error handling templates
- [ ] Support workflow inheritance

#### 5.4 Cross-Engine Integration Examples
**Task 5.4.1: Multi-Engine Examples**
- [ ] Create examples for each engine
- [ ] Show language-specific features
- [ ] Demonstrate built-ins integration
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