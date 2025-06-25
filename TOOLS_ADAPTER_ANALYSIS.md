# Tools Adapter Analysis: tools.lua vs ToolsAdapter

## Current Status
**FILE**: `pkg/engine/lua/stdlib/tools.lua`
**ADAPTER**: `pkg/engine/lua/adapters/impl/tools.go`

## Method Coverage Analysis

### ✅ EXISTING in tools.lua (Custom high-level wrappers)
- `tools.define()` - Custom tool registration
- `tools.register_library()` - Bulk registration
- `tools.compose()` - Tool composition
- `tools.execute_safe()` - Safe execution with metrics
- `tools.pipeline()` - Pipeline execution
- `tools.parallel_execute()` - Parallel execution
- `tools.validate_params()` - Parameter validation
- `tools.test_tool()` - Tool testing
- `tools.benchmark_tool()` - Performance benchmarking
- `tools.list()` - Tool listing
- `tools.search()` - Tool search
- `tools.get_info()` - Tool information
- `tools.get_metrics()` - Local metrics
- `tools.get_history()` - Execution history
- `tools.clear_history()` - History management
- `tools.reset_metrics()` - Metrics reset
- `tools.export_tool()` - Tool export

### ❌ MISSING from tools.lua (ToolsAdapter methods)

#### Core Tool Discovery Methods
- `tools.listTools()` - Bridge method for listing tools
- `tools.searchTools(query)` - Bridge tool search
- `tools.getToolInfo(toolName)` - Bridge tool info
- `tools.getToolSchema(toolName)` - Tool schema access
- `tools.getCategories()` - Available categories
- `tools.listByCategory(category)` - Filter by category
- `tools.listByTags(tags)` - Filter by tags

#### Tool Execution Methods
- `tools.executeTool(toolName, params)` - Bridge execution
- `tools.executeAsync(toolName, params)` - Async execution

#### Custom Tool Registration
- `tools.registerCustomTool(toolDef)` - Bridge registration

#### Validation Methods
- `tools.validateToolInput(toolName, params)` - Bridge validation

#### Metrics Methods
- `tools.getToolMetrics(toolName)` - Bridge metrics

#### Builder Pattern Support
- `tools.createBuilder(toolName)` - Fluent builder

#### Registry Bridge Methods (Optional)
- `tools.getTool(toolName)` - Complete tool from registry
- `tools.listToolsByPermission(permission)` - Filter by permission
- `tools.listToolsByResourceUsage(criteria)` - Filter by resource usage
- `tools.getToolDocumentation(toolName)` - Comprehensive docs
- `tools.exportToolToMCP(toolName)` - MCP export single
- `tools.exportAllToolsToMCP()` - MCP export all
- `tools.clearRegistry()` - Clear registry
- `tools.getRegistryStats()` - Registry statistics

#### Constants from ToolsAdapter
- `tools.CATEGORIES` - Category constants
- `tools.PERMISSIONS` - Permission types
- `tools.RESOURCE_USAGE` - Resource usage levels

## Architecture Issues

### 1. Bridge Access Pattern
**CURRENT tools.lua**: Uses `bridges.agent_tools` only
**ToolsAdapter**: Multi-bridge pattern:
- Main bridge (`agent_tools`)
- Registry bridge (`agent_tools_registry`) - optional

### 2. API Philosophy Mismatch
**CURRENT tools.lua**: High-level orchestration and custom tools
**ToolsAdapter**: Direct bridge method exposure + convenience

### 3. Missing Integration
- No registry bridge integration
- Missing direct bridge method wrappers
- No MCP export functionality
- No builder pattern support

## Implementation Strategy

1. **KEEP** existing high-level convenience methods in tools.lua
2. **ADD** all missing ToolsAdapter methods to tools.lua
3. **IMPLEMENT** registry bridge support (optional)
4. **ADD** builder pattern support
5. **ADD** constants and categories
6. **MAINTAIN** backward compatibility

## Bridge Access Pattern

### Current
```lua
local function get_tools_bridge()
    if not bridges or not bridges.agent_tools then
        error("Tools bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return bridges.agent_tools
end
```

### Required
```lua
local function get_tools_bridge()
    if not bridges or not bridges.agent_tools then
        error("Tools bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return bridges.agent_tools
end

local function get_registry_bridge()
    if bridges and bridges.agent_tools_registry then
        return bridges.agent_tools_registry
    end
    return nil
end
```

## Method Implementation Plan

1. Add direct bridge method wrappers
2. Add registry bridge methods with availability checks
3. Add builder pattern implementation
4. Add constants tables
5. Organize methods in namespaces for clarity
6. Keep existing high-level methods unchanged