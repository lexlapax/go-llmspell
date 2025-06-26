# Solution: LLM Agent Integration

## Overview

To enable real LLM agent creation, we need to implement the `createLLMAgent` case in the agent bridge and establish communication with the LLM bridge to get providers.

## Implementation Steps

### 1. Store Engine Reference in Agent Bridge

Modify `AgentBridge` to store the engine reference:

```go
// In pkg/bridge/agent/agent.go

type AgentBridge struct {
    mu            sync.RWMutex
    initialized   bool
    agents        map[string]types.BaseAgent
    registry      types.AgentRegistry
    eventStorage  events.EventStorage
    eventReplayer *events.EventReplayer
    profiler      *profiling.Profiler
    
    // Add engine reference
    engine        types.ScriptEngine
}

// Update RegisterWithEngine to store the engine
func (b *AgentBridge) RegisterWithEngine(engine types.ScriptEngine) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.engine = engine
    return nil
}
```

### 2. Implement createLLMAgent Case

Add the missing case in `ExecuteMethod`:

```go
case "createLLMAgent":
    b.mu.Lock()
    defer b.mu.Unlock()

    if len(args) < 2 {
        return types.NewErrorValue(fmt.Errorf("createLLMAgent requires name and model/config parameters")), nil
    }
    
    name := args[0].(types.StringValue).Value()
    
    // Handle two scenarios:
    // 1. String model name: "gpt-3.5-turbo"
    // 2. Object with model config: {model: "gpt-3.5-turbo", ...}
    
    var provider ldomain.Provider
    var config map[string]interface{}
    
    switch arg := args[1].(type) {
    case types.StringValue:
        // Just a model name - need to get provider from LLM bridge
        modelName := arg.Value()
        provider, err := b.getProviderForModel(modelName)
        if err != nil {
            return types.NewErrorValue(fmt.Errorf("failed to get provider for model %s: %w", modelName, err)), nil
        }
        
    case types.ObjectValue:
        // Configuration object
        config = arg.ToGo().(map[string]interface{})
        
        // Extract model name from config
        modelName, ok := config["model"].(string)
        if !ok {
            return types.NewErrorValue(fmt.Errorf("config must include 'model' field")), nil
        }
        
        // Get provider from LLM bridge
        provider, err := b.getProviderForModel(modelName)
        if err != nil {
            return types.NewErrorValue(fmt.Errorf("failed to get provider for model %s: %w", modelName, err)), nil
        }
        
    default:
        return types.NewErrorValue(fmt.Errorf("invalid argument type for createLLMAgent")), nil
    }
    
    // Extract additional options from config
    var options map[string]interface{}
    if len(args) > 2 {
        options = args[2].(types.ObjectValue).ToGo().(map[string]interface{})
    } else if config != nil {
        options = config
    }
    
    // Create the LLM agent using go-llms
    var agent types.BaseAgent
    
    // Handle system prompt if provided
    if systemPrompt, ok := options["system"].(string); ok {
        // Create with system prompt
        llmAgent := agentcore.NewAgent(name, provider)
        llmAgent.SetSystemPrompt(systemPrompt)
        
        // Set other options
        if temp, ok := options["temperature"].(float64); ok {
            llmAgent.SetTemperature(temp)
        }
        
        agent = llmAgent
    } else {
        // Create basic LLM agent
        agent = agentcore.NewAgent(name, provider)
    }
    
    // Store agent
    b.agents[agent.ID()] = agent
    
    // Return agent info
    result := map[string]types.ScriptValue{
        "id":   types.NewStringValue(agent.ID()),
        "type": types.NewStringValue("llm"),
        "name": types.NewStringValue(name),
    }
    return types.NewObjectValue(result), nil
```

### 3. Implement getProviderForModel Helper

```go
// getProviderForModel retrieves a provider from the LLM bridge
func (b *AgentBridge) getProviderForModel(modelName string) (ldomain.Provider, error) {
    if b.engine == nil {
        return nil, fmt.Errorf("engine not set - RegisterWithEngine not called")
    }
    
    // Get the LLM bridge
    llmBridge, err := b.engine.GetBridge("llm_core")
    if err != nil {
        return nil, fmt.Errorf("LLM bridge not found: %w", err)
    }
    
    // Call getProvider method on LLM bridge
    ctx := context.Background()
    result, err := llmBridge.ExecuteMethod(ctx, "getProviderForModel", []types.ScriptValue{
        types.NewStringValue(modelName),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get provider: %w", err)
    }
    
    // Extract provider from result
    // This requires the LLM bridge to return a provider object
    // For now, we'll need to create a mock provider
    
    // TODO: Implement proper provider retrieval
    // This would involve the LLM bridge managing real providers
    
    return nil, fmt.Errorf("provider retrieval not yet implemented")
}
```

### 4. Update LLM Bridge to Manage Real Providers

The LLM bridge needs to:
1. Create and manage real go-llms providers
2. Expose a method to get providers by model name
3. Return provider references that can be used by agent bridge

```go
// In pkg/bridge/llm/llm.go

// Add provider management
func (b *LLMBridge) getProviderForModel(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
    modelName := args[0].(types.StringValue).Value()
    
    // Map model names to provider types
    var provider ldomain.Provider
    
    switch {
    case strings.HasPrefix(modelName, "gpt"):
        // Create OpenAI provider
        provider = openai.NewProvider(openai.Config{
            APIKey: os.Getenv("OPENAI_API_KEY"),
        })
        
    case strings.HasPrefix(modelName, "claude"):
        // Create Anthropic provider
        provider = anthropic.NewProvider(anthropic.Config{
            APIKey: os.Getenv("ANTHROPIC_API_KEY"),
        })
        
    default:
        return types.NewErrorValue(fmt.Errorf("unknown model: %s", modelName)), nil
    }
    
    // Return provider reference
    // This is challenging because we need to pass a Go object between bridges
    // One solution is to use a provider registry with IDs
    
    providerID := b.registerProvider(provider)
    return types.NewStringValue(providerID), nil
}
```

## Alternative Solutions

### 1. Direct Provider Creation in Agent Bridge
Instead of getting providers from LLM bridge, create them directly in agent bridge based on model names.

### 2. Shared Provider Registry
Create a global provider registry that both bridges can access.

### 3. Configuration-Based Approach
Pass provider configuration through script and create providers on demand.

## Challenges

1. **Type Safety**: Passing Go objects (providers) between bridges through ScriptValue interface
2. **Provider Lifecycle**: Managing provider instances across bridges
3. **Configuration**: API keys and provider configuration need to be accessible
4. **Mock vs Real**: Current implementation uses mocks; transitioning to real providers requires proper configuration

## Next Steps

1. Decide on provider management strategy
2. Implement provider creation/retrieval mechanism
3. Update agent creation to use real providers
4. Add configuration for API keys
5. Test with real LLM providers