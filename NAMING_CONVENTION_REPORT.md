# Lua Stdlib Naming Convention Report

## Overview
This report summarizes the naming convention review of all Lua stdlib modules in go-llmspell to ensure they use snake_case naming for all exported functions.

## Summary
All 11 Lua stdlib modules have been reviewed. The modules are divided into three categories:

### 1. Modules with Perfect snake_case Naming (7 modules)
These modules already follow snake_case naming conventions throughout:
- **hooks.lua** - All functions use snake_case (e.g., `register_hook`, `unregister_hook`, `execute_hooks`)
- **modelinfo.lua** - All functions use snake_case (e.g., `discovery_scan`, `capabilities_check`, `selection_find`)
- **workflow.lua** - All functions use snake_case (e.g., `create_workflow`, `execute_workflow`, `pause_workflow`)
- **utils.lua** - All functions use snake_case (e.g., `auth_authenticate`, `debug_set_level`, `json_parse`)
- **state.lua** - All functions use snake_case (e.g., `get`, `set`, `update`, `delete`)
- **observability.lua** - All functions use snake_case (e.g., `counter`, `gauge`, `timer`, `start_span`)
- **events.lua** - All functions use snake_case (e.g., `create_emitter`, `wait_for`, `aggregate`)
- **structured.lua** - All functions use snake_case (e.g., `create_schema`, `validate_json`, `save_version`)

### 2. Modules with Mixed Naming (3 modules)
These modules have snake_case functions but also include camelCase adapter compatibility functions:

#### agent.lua
- **Snake_case functions**: `create`, `configure`, `clone`, `run`, `run_async`, `conversation`, `delegate`, `collaborate`, etc.
- **CamelCase adapter functions** (lines 717-1010): `createAgent`, `createLLMAgent`, `listAgents`, `getAgent`, `removeAgent`, `lifecycleCreate`, `lifecycleCreateLLM`, `registerTool`, `unregisterTool`, `stateGet`, `stateSet`, etc.

#### llm.lua  
- **Snake_case functions**: `quick_prompt`, `quick_prompt_async`, `chat_session`, `streaming_response`, `batch_process`, `use_provider`, `compare_providers`, etc.
- **CamelCase adapter functions** (lines 575-1137): `generate`, `generateMessage`, `countTokens`, `createAgent`, `agentComplete`, `agentStream`, `providersCreate`, `providersGet`, `poolCreate`, `poolGetHealth`, etc.

#### tools.lua
- **Snake_case functions**: `define`, `register_library`, `compose`, `execute_safe`, `pipeline`, `parallel_execute`, `validate_params`, `test_tool`, `benchmark_tool`, etc.
- **CamelCase adapter functions** (lines 772-1079): `listTools`, `searchTools`, `getToolInfo`, `getToolSchema`, `executeTool`, `executeAsync`, `registerCustomTool`, `validateToolInput`, `getToolMetrics`, `createBuilder`, etc.

### 3. Analysis of Mixed Naming Pattern

The camelCase functions in agent.lua, llm.lua, and tools.lua follow a consistent pattern:
1. They are grouped at the end of each module
2. They are clearly marked with comments like "MISSING LLMADAPTER METHODS - Adding for full compatibility"
3. They appear to be bridge/adapter methods that wrap the snake_case functions
4. They provide compatibility with Go-side adapters that expect camelCase naming

## Recommendations

1. **No Action Required for Pure snake_case Modules**: The 8 modules that already use snake_case exclusively are correctly implemented.

2. **Consider Adapter Pattern for Mixed Modules**: The current approach in agent.lua, llm.lua, and tools.lua appears intentional:
   - The primary API uses snake_case (Lua convention)
   - The adapter methods use camelCase (Go convention)
   - This provides compatibility while maintaining idiomatic Lua naming

3. **Documentation**: Consider adding a comment at the top of the mixed modules explaining the dual naming convention:
   ```lua
   -- This module uses snake_case for all primary functions (Lua convention)
   -- CamelCase functions at the end are adapter methods for Go-side compatibility
   ```

## Conclusion

The Lua stdlib modules follow a consistent pattern where:
- Primary Lua APIs use snake_case naming (idiomatic for Lua)
- Adapter/bridge compatibility functions use camelCase (matching Go-side expectations)
- This dual approach provides both idiomatic Lua interfaces and necessary Go interoperability

No changes are required as the current implementation appears to be intentional and well-structured.