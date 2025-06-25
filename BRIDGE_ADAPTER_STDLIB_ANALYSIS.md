# Bridge-Adapter-Stdlib Deep Analysis Report

## Executive Summary

After analyzing the three layers (bridges, adapters, and stdlib), I found:

1. **Architecture is sound** - Clean separation of concerns with proper layering
2. **Notation inconsistency exists** - Colon notation (`:`) used in events.lua while dot notation (`.`) used everywhere else
3. **Bridge access pattern varies** - Most files use local bridge variables, but events.lua uses global `bridges` table directly

## Key Findings

### 1. Method Call Notation Inconsistency

**INCONSISTENCY FOUND**: In `events.lua` lines 714 and 724:
```lua
-- Using colon notation (INCONSISTENT)
bridges.agent_events:publishEvent(event)    -- line 714
bridges.agent_events:subscribe(pattern, handler)  -- line 724
```

**Everywhere else uses dot notation**:
```lua
-- Standard pattern in other files
bridge.createAgent(agent_id, agent_config)  -- agent.lua
bridge.generate(prompt, opts)               -- llm.lua
bridge.set(key, value)                      -- state.lua
```

### 2. Bridge Access Patterns

Two distinct patterns observed:

**Pattern A: Local bridge variable** (used in most stdlib files)
```lua
-- Get bridge once at module level
local function get_agent_bridge()
    if not bridges or not bridges.agent_core then
        error("Agent bridge not available")
    end
    return bridges.agent_core
end

-- Use local bridge variable
local bridge = get_agent_bridge()
bridge.createAgent(...)  -- dot notation
```

**Pattern B: Direct global access** (used in events.lua)
```lua
-- Direct access to global bridges table
if bridges and bridges.agent_events then
    bridges.agent_events:publishEvent(event)  -- colon notation
end
```

### 3. Bridge ID Mappings (from adapters)

Complete mapping of bridge IDs to their adapter files:

| Adapter File | Bridge IDs |
|--------------|------------|
| agent.go | `bridges.agent_core` |
| events.go | `bridges.agent_events` |
| hooks.go | `bridges.agent_hooks` |
| tools.go | `bridges.agent_tools`, `bridges.agent_tools_registry` |
| workflow.go | `bridges.agent_workflow` |
| llm.go | `bridges.llm_core`, `bridges.llm_providers`, `bridges.llm_pool` |
| modelinfo.go | `bridges.llm_modelinfo` |
| observability.go | `bridges.observability_metrics`, `bridges.observability_tracing`, `bridges.observability_guardrails` |
| state.go | `bridges.state_manager`, `bridges.state_context` |
| structured.go | `bridges.structured_schema` |
| utils.go | `bridges.util_core`, `bridges.util_debug`, `bridges.util_slog`, `bridges.util_script_logger`, `bridges.util_auth`, `bridges.util_errors`, `bridges.util_json`, `bridges.util_llm` |

### 4. Architectural Relationships

**Data Flow**:
1. Lua script calls stdlib function
2. Stdlib validates and prepares parameters
3. Stdlib calls bridge method via `bridges.*` global
4. Adapter (if present) handles type conversion
5. Bridge delegates to go-llms
6. Response flows back through same path

**Type Conversion**:
- Handled by `LuaTypeConverter` in adapter layer
- Lua tables ↔ Go maps/structs
- Lua functions ↔ Go callbacks
- Lua userdata ↔ Go objects

## Recommendations

### 1. Fix Notation Inconsistency
The colon notation in events.lua should be changed to dot notation for consistency:
```lua
-- Change from:
bridges.agent_events:publishEvent(event)
bridges.agent_events:subscribe(pattern, handler)

-- To:
bridges.agent_events.publishEvent(event)
bridges.agent_events.subscribe(pattern, handler)
```

### 2. Standardize Bridge Access Pattern
Consider adopting Pattern A (local bridge variable) consistently across all stdlib files for:
- Better error handling
- Cleaner code
- Consistent patterns

### 3. Verify Method Signatures
The colon notation suggests these methods might expect `self` as first parameter, but based on the architecture analysis, all bridge methods should use dot notation as they're not object methods but module functions.

## Conclusion

The architecture follows the "no reimplementation" principle well, with clean separation between:
- **Bridges**: Wrap go-llms functionality
- **Adapters**: Handle Lua-specific type conversions
- **Stdlib**: Provide high-level Lua APIs

The only issue is the notation inconsistency in events.lua which should be fixed for consistency.