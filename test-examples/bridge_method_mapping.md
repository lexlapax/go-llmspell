# Bridge Method Mapping

## Agent.lua expects (adapter methods) → Bridge actually has:

- `lifecycleCreate` → `createAgent`
- `lifecycleGet` → `getAgent`
- `lifecycleList` → `listAgents`
- `lifecycleRemove` → `removeAgent`
- `stateSet` → `setAgentState`
- `stateGet` → `getAgentState`
- `run` → `runAgent`
- `runAsync` → `runAgentAsync`
- `registerTool` → `registerAgentTool`
- `listTools` → `listAgentTools`

## Workflow methods:
- `create` → `createWorkflow`
- `addStep` → `addWorkflowStep`
- `execute` → `executeAgentWorkflow`
- `get` → No direct equivalent (need to check)

## The Problem:
The adapter layer that provides these flattened method names is not being used.
The `bridges.agent_core` is the raw bridge, not the adapter-wrapped version.