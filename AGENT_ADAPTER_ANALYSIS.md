# Agent Adapter vs agent.lua Method Analysis

## AgentAdapter Methods (from GetMethods())

### Standard Convenience Methods
- ❌ `createAgent` - Missing from agent.lua (has `agent.create` instead, different signature)
- ❌ `createLLMAgent` - Missing from agent.lua
- ❌ `listAgents` - Missing from agent.lua (has `agent.list` instead)
- ❌ `getAgent` - Missing from agent.lua (has `agent.get` instead)
- ❌ `removeAgent` - Missing from agent.lua (has `agent.remove` instead)

### Lifecycle Methods (flattened)
- ❌ `lifecycleCreate` - Missing from agent.lua
- ❌ `lifecycleCreateLLM` - Missing from agent.lua
- ❌ `lifecycleList` - Missing from agent.lua
- ❌ `lifecycleGet` - Missing from agent.lua
- ❌ `lifecycleRemove` - Missing from agent.lua
- ❌ `lifecycleGetMetrics` - Missing from agent.lua

### Communication Methods
- ✅ `run` - Implemented as `agent.run(agent_id, input, options)`
- ✅ `runAsync` - Implemented as `agent.run_async(agent_id, input, options)`
- ❌ `registerTool` - Missing from agent.lua (has `agent.add_tools` and `agent.create_tool`)
- ❌ `unregisterTool` - Missing from agent.lua
- ❌ `listTools` - Missing from agent.lua (has `agent.get_tools`)

### State Methods (flattened)
- ❌ `stateGet` - Missing from agent.lua
- ❌ `stateSet` - Missing from agent.lua
- ❌ `stateExport` - Missing from agent.lua
- ❌ `stateImport` - Missing from agent.lua
- ❌ `stateSaveSnapshot` - Missing from agent.lua
- ❌ `stateLoadSnapshot` - Missing from agent.lua
- ❌ `stateListSnapshots` - Missing from agent.lua

### Events Methods (flattened)
- ❌ `eventsEmit` - Missing from agent.lua
- ❌ `eventsSubscribe` - Missing from agent.lua
- ❌ `eventsUnsubscribe` - Missing from agent.lua
- ❌ `eventsStartRecording` - Missing from agent.lua
- ❌ `eventsStopRecording` - Missing from agent.lua
- ❌ `eventsReplay` - Missing from agent.lua

### Profiling Methods (flattened)
- ❌ `profilingStart` - Missing from agent.lua
- ❌ `profilingStop` - Missing from agent.lua
- ❌ `profilingGetMetrics` - Missing from agent.lua
- ❌ `profilingGetReport` - Missing from agent.lua

### Workflow Methods (flattened)
- ❌ `workflowCreate` - Missing from agent.lua (has `agent.workflow_create`)
- ❌ `workflowExecute` - Missing from agent.lua (has `agent.workflow_run`)
- ❌ `workflowAddStep` - Missing from agent.lua

### Hooks Methods (flattened)
- ❌ All hooks methods missing from agent.lua

## Methods in agent.lua NOT in AgentAdapter
- ✅ `agent.create(name, config)` - High-level wrapper
- ✅ `agent.configure(agent_id, settings)` - Configuration helper
- ✅ `agent.clone(source_agent_id, new_name, modifications)` - Cloning functionality
- ✅ `agent.conversation(agent_id, system_prompt)` - Conversation helper
- ✅ `agent.delegate(from_agent_id, to_agent_id, task, options)` - Delegation
- ✅ `agent.collaborate(agent_ids, task, options)` - Multi-agent collaboration
- ✅ `agent.add_tools(agent_id, tools)` - Tool management
- ✅ `agent.create_tool(name, func, schema)` - Tool creation
- ✅ `agent.get_tools(agent_id)` - Tool listing
- ✅ `agent.tool_chain(tools, initial_data, options)` - Tool chaining
- ✅ `agent.workflow_create(name, steps, options)` - Workflow creation
- ✅ `agent.workflow_run(workflow_id, input, options)` - Workflow execution
- ✅ `agent.workflow_parallel(steps, input, options)` - Parallel workflows
- ✅ `agent.workflow_conditional(condition, then_step, else_step, input, options)` - Conditional workflows
- ✅ `agent.get_status(agent_id)` - Status checking
- ✅ `agent.get_workflow_status(workflow_id)` - Workflow status
- ✅ `agent.list_active()` - Active agents listing
- ✅ `agent.list_active_workflows()` - Active workflows listing

## Critical Issues Found

1. **Different API Philosophy**: agent.lua provides high-level workflow helpers while AgentAdapter exposes low-level bridge methods
2. **Missing Bridge Methods**: Many core AgentAdapter methods not exposed in agent.lua
3. **Naming Inconsistencies**: agent.lua uses different method names than AgentAdapter expects
4. **Missing Namespaces**: No state, events, profiling, or hooks functionality exposed

## Key Problem with 09-state-management.lua

The error `assistant:run(context_prompt)` fails because:
1. `assistant` is created using `agent.create()` which returns an agent ID, not an object
2. agent.lua doesn't provide object-oriented syntax `agent:run()`
3. Should be `agent.run(assistant, context_prompt)` instead

## Recommended Actions

1. ✅ **COMPLETED**: Add missing AgentAdapter methods to agent.lua
2. ✅ **COMPLETED**: Add object-oriented wrapper for agent objects
3. ✅ **COMPLETED**: Add namespaced methods for state, events, profiling
4. ✅ **COMPLETED**: Fix the object-oriented syntax issue that's breaking examples

## Implementation Complete ✅

**Phase 6.1.3 COMPLETED [2025-06-25]**: All AgentAdapter methods now wrapped in agent.lua

### Methods Added:
- **Standard convenience methods**: `createAgent`, `createLLMAgent`, `listAgents`, `getAgent`, `removeAgent`
- **Lifecycle methods**: `lifecycleCreate`, `lifecycleCreateLLM`, `lifecycleList`, `lifecycleGet`, `lifecycleRemove`, `lifecycleGetMetrics`
- **Tool management methods**: `registerTool`, `unregisterTool`, `listTools`
- **State methods**: `stateGet`, `stateSet`, `stateExport`, `stateImport`, `stateSaveSnapshot`, `stateLoadSnapshot`, `stateListSnapshots`
- **Events methods**: `eventsEmit`, `eventsSubscribe`, `eventsUnsubscribe`, `eventsStartRecording`, `eventsStopRecording`, `eventsReplay`
- **Profiling methods**: `profilingStart`, `profilingStop`, `profilingGetMetrics`, `profilingGetReport`
- **Workflow methods**: `workflowCreate`, `workflowExecute`, `workflowAddStep`
- **Hooks methods**: `hooksRegister`, `hooksUnregister`, `hooksExecute`, `hooksList`

### Namespace organization:
- `agent.state.*` - Agent state management
- `agent.events.*` - Agent event handling
- `agent.profiling.*` - Agent performance profiling
- `agent.hooks.*` - Agent lifecycle hooks

### API Coverage:
- **38/38 AgentAdapter methods** now wrapped in agent.lua
- **Object-oriented syntax working** - `assistant:run()` now works
- **Input format compatibility** - Accepts both string and table inputs
- **Backward compatible** - All existing methods still work
- **Full test coverage** - All tests passing

### Critical Fixes:
- **Object-oriented support**: Added metatable to agent objects for `assistant:run()` syntax
- **Input format handling**: `agent.run()` converts string inputs to expected message format
- **Bridge compatibility**: Works with multiple bridge IDs for different agent types