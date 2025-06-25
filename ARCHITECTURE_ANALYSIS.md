# Architecture Analysis: Bridge, Adapter, and Stdlib Layers

## Overview

This document analyzes the relationships between three key layers in the go-llmspell architecture:
1. **Bridge Layer** (`/pkg/bridge/`) - Go code that wraps go-llms functionality
2. **Adapter Layer** (`/pkg/engine/gopherlua/adapters/`) - Go code that adapts bridges for Lua
3. **Stdlib Layer** (`/pkg/engine/gopherlua/stdlib/`) - Lua code that provides high-level APIs

## Layer Structure

### 1. Bridge Layer (`/pkg/bridge/`)
The bridge layer contains Go packages that wrap go-llms functionality:

```
bridge/
├── agent/          # Agent functionality (core, events, hooks, tools, workflow)
├── llm/            # LLM operations (core, pool, providers)
├── manager         # Bridge lifecycle management
├── modelinfo       # Model information access
├── observability/  # Metrics, tracing, guardrails
├── registry/       # Bridge registration and feature sets
├── state/          # State management (context, manager)
├── structured/     # Schema validation
├── types/          # Shared types
└── util/           # Utilities (auth, debug, errors, json, logging)
```

### 2. Adapter Layer (`/pkg/engine/gopherlua/adapters/`)
The adapter layer provides Lua-specific wrappers around bridges:

```
adapters/
├── agent.go          # bridge_id=bridges.agent_core
├── events.go         # bridge_id=bridges.agent_events
├── hooks.go          # bridge_id=bridges.agent_hooks
├── llm.go            # bridge_id=bridges.llm_core, llm_providers, llm_pool
├── modelinfo.go      # bridge_id=bridges.llm_modelinfo
├── observability.go  # bridge_id=bridges.observability_metrics/tracing/guardrails
├── state.go          # bridge_id=bridges.state_manager, state_context
├── structured.go     # bridge_id=bridges.structured_schema
├── tools.go          # bridge_id=bridges.agent_tools, agent_tools_registry
├── utils.go          # bridge_id=bridges.util_core/debug/slog/script_logger/auth
└── workflow.go       # bridge_id=bridges.agent_workflow
```

### 3. Stdlib Layer (`/pkg/engine/gopherlua/stdlib/`)
The stdlib layer provides Lua modules that use bridges via the global `bridges` table:

```
stdlib/
├── agent.lua         # Uses: bridges.agent_core, bridges.agent_workflow
├── auth.lua          # Uses: bridges.util_auth, bridges.security
├── data.lua          # Uses: bridges.util_core
├── errors.lua        # Uses: bridges.util_errors
├── events.lua        # Uses: bridges.agent_events
├── llm.lua           # Uses: bridges.llm_core, bridges.util_llm
├── logging.lua       # Uses: bridges.util_debug/slog/script_logger
├── observability.lua # Uses: bridges.observability_metrics/tracing/guardrails
├── state.lua         # Uses: bridges.state_manager, bridges.state_context
├── structured.lua    # Uses: bridges.structured_schema
└── tools.lua         # Uses: bridges.agent_tools, bridges.agent_tools_registry
```

## Data Flow Architecture

### 1. Registration Flow
```
Bridge Implementation → Registry → BridgeManager → Global 'bridges' table in Lua
```

### 2. Execution Flow
```
Lua Script → Stdlib Module → bridges.* call → Adapter → Bridge → go-llms
```

### 3. Type Conversion Flow
```
Lua Types ←→ LuaTypeConverter (in Adapter) ←→ Go Types
```

## Key Relationships

### Bridge to Adapter Mapping

| Bridge Package | Bridge IDs | Adapter File |
|----------------|------------|--------------|
| agent/* | agent_core, agent_events, agent_hooks, agent_tools, agent_tools_registry, agent_workflow | agent.go, events.go, hooks.go, tools.go, workflow.go |
| llm/* | llm_core, llm_providers, llm_pool, llm_modelinfo | llm.go, modelinfo.go |
| observability/* | observability_metrics, observability_tracing, observability_guardrails | observability.go |
| state/* | state_manager, state_context | state.go |
| structured/* | structured_schema | structured.go |
| util/* | util_core, util_auth, util_debug, util_errors, util_json, util_llm, util_slog, util_script_logger | utils.go |

### Adapter to Stdlib Mapping

| Adapter | Bridge IDs | Stdlib Modules |
|---------|------------|----------------|
| agent.go | bridges.agent_core | agent.lua |
| events.go | bridges.agent_events | events.lua, agent.lua |
| hooks.go | bridges.agent_hooks | agent.lua |
| llm.go | bridges.llm_core, etc. | llm.lua |
| modelinfo.go | bridges.llm_modelinfo | llm.lua |
| observability.go | bridges.observability_* | observability.lua |
| state.go | bridges.state_* | state.lua |
| structured.go | bridges.structured_schema | structured.lua |
| tools.go | bridges.agent_tools* | tools.lua, agent.lua |
| utils.go | bridges.util_* | auth.lua, data.lua, errors.lua, logging.lua |
| workflow.go | bridges.agent_workflow | agent.lua |

## Architectural Patterns

### 1. Bridge Pattern
- Each bridge wraps specific go-llms functionality
- Bridges implement the `engine.Bridge` interface
- Bridges are registered with unique IDs

### 2. Adapter Pattern
- Adapters inherit from `BridgeAdapter`
- They provide Lua-specific type conversions
- Minimal logic - mostly delegation to bridges

### 3. Module Pattern
- Lua stdlib modules provide high-level APIs
- They access bridges via the global `bridges` table
- Each module focuses on a specific domain

### 4. Registry Pattern
- Bridges are organized into feature sets
- Feature sets map to security profiles
- Dynamic registration based on configuration

## Key Design Principles

1. **Separation of Concerns**: Each layer has a specific responsibility
2. **No Reimplementation**: Bridges wrap go-llms, never reimplement
3. **Type Safety**: Adapters handle all type conversions
4. **Modularity**: Feature sets allow selective bridge loading
5. **Security**: Bridge availability controlled by security profiles

## Example Flow: Creating an Agent

1. **Lua Script**: `agent.create("my-agent", {model = "gpt-4"})`
2. **Stdlib (agent.lua)**: 
   - Validates parameters
   - Calls `bridges.agent_core:createAgent(...)`
3. **Adapter (agent.go)**: 
   - Converts Lua table to Go map
   - Delegates to bridge
4. **Bridge (agent/agent.go)**:
   - Calls go-llms agent creation
   - Returns agent instance
5. **Type Conversion**: 
   - Go agent → Adapter → Lua userdata
6. **Return to Script**: Agent handle available in Lua

This architecture ensures clean separation between scripting concerns and core LLM functionality while maintaining type safety and security boundaries.