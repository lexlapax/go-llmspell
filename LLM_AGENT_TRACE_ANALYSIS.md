# LLM Agent Creation Trace Analysis

## The Problem

When creating an LLM agent in Lua scripts, real LLM providers aren't being used. Instead, mock agents are created that return canned responses.

## Complete Execution Path

### 1. Script Level: `03-agent-plain.lua`
```lua
local analyst = agent.create("Data Analyst", {
    model = model,  -- e.g., "gpt-3.5-turbo"
    system = "You are a data analyst...",
    temperature = 0.3
})
```

### 2. Stdlib Wrapper: `agent.lua`
```lua
function agent.create(name, config)
    local bridge = get_agent_bridge()  -- Gets bridges.agent_core
    local agent_obj = bridge.lifecycleCreate(name, agent_config)
    -- ...
end
```

### 3. Adapter: `agent.go` (AgentAdapter)
```go
// In addFlattenedMethods
L.SetField(module, "lifecycleCreate", L.NewFunction(func(L *lua.LState) int {
    agentID := L.CheckString(1)
    config := L.CheckTable(2)
    
    result, err := aa.GetBridge().ExecuteMethod(ctx, "createAgent", args)
    // ...
}))
```

### 4. Bridge: `agent.go` (AgentBridge)
```go
case "createAgent":
    // Extract name, description, and type from config
    agentTypeStr, _ := config["type"].(string)
    if agentTypeStr == "" {
        agentTypeStr = "basic"
    }
    
    // Create agent based on type
    baseAgent := agentcore.NewBaseAgent(name, description, domain.AgentType(agentTypeStr))
    agent = &MockAgent{
        BaseAgent:   baseAgent,
        AgentConfig: config,
    }
```

## Key Issues Identified

### 1. **No createLLMAgent Implementation**
- The bridge defines `createLLMAgent` in Methods() but doesn't handle it in ExecuteMethod()
- When the method is called, it returns "method not found"

### 2. **Missing Provider Integration**
- `createLLMAgent` expects a Provider parameter (from go-llms)
- Scripts only pass a model name string (e.g., "gpt-3.5-turbo")
- No mechanism to convert model name to Provider instance

### 3. **Bridge Coordination Issue**
- LLM bridge manages providers separately
- Agent bridge has no way to access LLM bridge's providers
- No inter-bridge communication mechanism

### 4. **Type Mismatch**
- Scripts pass: `{model: "gpt-3.5-turbo", system: "...", temperature: 0.3}`
- go-llms expects: `provider ldomain.Provider` instance
- No conversion logic exists

## How go-llms Creates LLM Agents

From `go-llms/pkg/agent/core/llm_agent.go`:

```go
// Simple creation
func NewAgent(name string, provider ldomain.Provider) *LLMAgent {
    deps := LLMDeps{
        Provider: provider,
        Logger:   slog.Default(),
        Tracer:   nil,
    }
    return NewLLMAgent(name, "LLM Agent", deps)
}

// Full creation
func NewLLMAgent(name, description string, deps LLMDeps) *LLMAgent {
    baseAgent := NewBaseAgent(name, description, domain.AgentTypeLLM)
    agent := &LLMAgent{
        BaseAgentImpl:    baseAgent,
        deps:             deps,
        tools:            make(map[string]domain.Tool),
        // ...
    }
    return agent
}
```

## What Should Happen

1. Script calls `agent.create()` with model configuration
2. Agent bridge should:
   - Detect model configuration in config
   - Get provider from LLM bridge based on model name
   - Call go-llms `NewAgent()` with the provider
   - Return a real LLM agent, not a mock

## Solutions Needed

### Option 1: Inter-Bridge Communication
- Add mechanism for agent bridge to access LLM bridge
- Query LLM bridge for provider by model name
- Use provider to create real LLM agent

### Option 2: Provider Registration in Agent Bridge
- Allow providers to be registered directly with agent bridge
- Map model names to providers
- Use registered providers when creating LLM agents

### Option 3: Bridge Context/Registry
- Create a shared bridge registry accessible to all bridges
- Register providers in the registry
- Agent bridge queries registry for providers

### Option 4: Implement createLLMAgent Case
```go
case "createLLMAgent":
    name := args[0].(types.StringValue).Value()
    modelOrProvider := args[1] // Could be string (model) or provider object
    
    // Get provider - either from LLM bridge or create based on model
    var provider ldomain.Provider
    if modelStr, ok := modelOrProvider.(types.StringValue); ok {
        // Get provider from LLM bridge or create based on model name
        provider = b.getOrCreateProvider(modelStr.Value())
    } else {
        // Direct provider object passed
        provider = modelOrProvider.ToGo().(ldomain.Provider)
    }
    
    // Create real LLM agent
    agent := agentcore.NewAgent(name, provider)
    b.agents[agent.ID()] = agent
    
    // Return agent info
    result := map[string]types.ScriptValue{
        "id":   types.NewStringValue(agent.ID()),
        "type": types.NewStringValue("llm"),
        "name": types.NewStringValue(name),
    }
    return types.NewObjectValue(result), nil
```

## Current Workaround

The current implementation creates MockAgent instances that:
- Accept any configuration
- Return canned responses like "[Mock Agent Response]"
- Don't actually call LLM providers
- Are sufficient for testing but not real usage

## Impact

- All agent examples work but don't use real LLMs
- Scripts can't actually interact with OpenAI, Anthropic, etc.
- The bridge architecture needs enhancement for real provider integration