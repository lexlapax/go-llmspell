# Hooks Adapter Analysis

## Summary
The HooksAdapter provides LLM pipeline hook functionality through the `bridges.agent_hooks` bridge. The hooks.lua module has been created from scratch to wrap all adapter functionality.

## Adapter Overview
- **File**: `/home/lexlapax/projects/lexlapax/go-llmspell/pkg/engine/lua/adapters/impl/hooks.go`
- **Bridge ID**: `bridges.agent_hooks`
- **Version**: 1.0.0

## Methods Implemented

### Hook Registration and Management (7 methods)
- `registerHook(hook_id, definition)` → `hooks.register_hook(hook_id, definition)`
- `unregisterHook(hook_id)` → `hooks.unregister_hook(hook_id)`
- `enableHook(hook_id)` → `hooks.enable_hook(hook_id)`
- `disableHook(hook_id)` → `hooks.disable_hook(hook_id)`
- `getHook(hook_id)` → `hooks.get_hook(hook_id)`
- `listHooks(filter?)` → `hooks.list_hooks(filter)`
- `clearHooks(hook_type?)` → `hooks.clear_hooks(hook_type)`

### Hook Execution (2 methods)
- `executeHooks(hook_type, context)` → `hooks.execute_hooks(hook_type, context)`
- `executeHook(hook_id, context)` → `hooks.execute_hook(hook_id, context)`

### Priority Management (2 methods)
- `setHookPriority(hook_id, priority)` → `hooks.set_hook_priority(hook_id, priority)`
- `getHookPriority(hook_id)` → `hooks.get_hook_priority(hook_id)`

### Hook State Queries (3 methods)
- `isHookEnabled(hook_id)` → `hooks.is_hook_enabled(hook_id)`
- `getEnabledHooks(hook_type?)` → `hooks.get_enabled_hooks(hook_type)`
- `getDisabledHooks(hook_type?)` → `hooks.get_disabled_hooks(hook_type)`

### Batch Operations (2 methods - from adapter convenience)
- `batchEnable(hook_ids)` → `hooks.batch_enable(hook_ids)`
- `batchDisable(hook_ids)` → `hooks.batch_disable(hook_ids)`

### Builder Pattern (1 method - from adapter convenience)
- `createHook(hook_id)` → `hooks.create_hook(hook_id)`

## Constants Implemented

### Hook Types
```lua
hooks.TYPES = {
    BEFORE_GENERATE = "beforeGenerate",
    AFTER_GENERATE = "afterGenerate",
    BEFORE_TOOL_CALL = "beforeToolCall",
    AFTER_TOOL_CALL = "afterToolCall"
}
```

### Priority Levels
```lua
hooks.PRIORITY = {
    HIGHEST = 1000,
    HIGH = 100,
    NORMAL = 0,
    LOW = -100,
    LOWEST = -1000
}
```

## Additional Features

### Lua Builder Pattern
In addition to the bridge's createHook method, the Lua module provides its own builder pattern implementation:

```lua
local hook = hooks.new_builder("hook_id")
    :with_priority(hooks.PRIORITY.HIGH)
    :before_generate(function(ctx) return ctx end)
    :after_generate(function(ctx) return ctx end)
    :before_tool_call(function(tool, args) return tool, args end)
    :after_tool_call(function(tool, result) return result end)
    :register()
```

### Simple Hook Creation Helper
```lua
local hook = hooks.create_simple_hook(
    "hook_id",
    hooks.TYPES.BEFORE_GENERATE,
    function(ctx) return ctx end,
    hooks.PRIORITY.HIGH  -- optional, defaults to NORMAL
)
```

### Bridge Access
The module exposes the raw bridge for advanced usage:
```lua
hooks.bridge -- Direct access to bridges.agent_hooks
```

## Hook Definition Structure
When registering hooks, the definition table can include:
- `priority`: Number value for execution order (default: 0)
- `beforeGenerate`: Function called before LLM generation
- `afterGenerate`: Function called after LLM generation
- `beforeToolCall`: Function called before tool execution
- `afterToolCall`: Function called after tool execution

## Implementation Status
✅ **COMPLETE** - All 18 adapter methods have been wrapped with snake_case naming convention. The module includes comprehensive constants, builder pattern support, batch operations, and full test coverage.

## Test Coverage
- Module loading and constant verification
- Hook registration and management (register, unregister, enable, disable, get)
- Hook lifecycle methods (list, clear, execute)
- Priority management (set, get)
- Batch operations (batch enable/disable)
- Builder pattern (both bridge and Lua implementations)
- Simple hook creation helper
- Hook state queries (enabled/disabled hooks)
- Hook execution (individual and by type)
- Graceful failure when bridge is missing
- Complete integration scenario test

Total: 11 test suites, all passing.

## Use Cases
This module is designed for LLM pipeline integration, allowing scripts to:
1. Intercept and modify prompts before generation
2. Process and validate responses after generation
3. Monitor and control tool calls
4. Implement rate limiting, logging, and security checks
5. Chain multiple hooks with priority-based execution order

Note: This is separate from the events.lua module's local hook system. The hooks.lua module specifically integrates with the LLM pipeline through the agent_hooks bridge.