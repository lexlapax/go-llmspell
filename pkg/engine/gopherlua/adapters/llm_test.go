// ABOUTME: Tests for LLM bridge adapter that exposes go-llms LLM functionality to Lua scripts
// ABOUTME: Validates agent creation, completion methods, streaming, model selection, and token counting

package adapters

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

func TestLLMAdapter_Creation(t *testing.T) {
	t.Run("create_llm_adapter", func(t *testing.T) {
		// Create LLM bridge mock
		llmBridge := &mockLLMBridge{
			id: "llm",
			metadata: engine.BridgeMetadata{
				Name:        "LLM Bridge",
				Version:     "1.0.0",
				Description: "Provides LLM agent functionality",
			},
		}

		// Create adapter with nil providers and pool bridges for basic test
		adapter := NewLLMAdapter(llmBridge, nil, nil)
		require.NotNil(t, adapter)

		// Should have specific LLM methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "createAgent")
		assert.Contains(t, methods, "complete")
		assert.Contains(t, methods, "stream")
		assert.Contains(t, methods, "listModels")
		assert.Contains(t, methods, "countTokens")
	})

	t.Run("llm_module_structure", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			metadata: engine.BridgeMetadata{
				Name: "LLM Bridge",
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Create module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module
		module := L.Get(-1).(*lua.LTable)
		L.Pop(1)

		// Check standard methods exist
		assert.NotEqual(t, lua.LNil, module.RawGetString("createAgent"))
		assert.NotEqual(t, lua.LNil, module.RawGetString("Agent")) // Constructor alias

		// Check flattened methods exist (no namespaces)
		assert.NotEqual(t, lua.LNil, module.RawGetString("modelsList"), "modelsList should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("modelsGetInfo"), "modelsGetInfo should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("poolCreate"), "poolCreate should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("providersCreate"), "providersCreate should exist")

		// Check namespaces don't exist
		assert.Equal(t, lua.LNil, module.RawGetString("models"), "models namespace should not exist")
		assert.Equal(t, lua.LNil, module.RawGetString("pool"), "pool namespace should not exist")
		assert.Equal(t, lua.LNil, module.RawGetString("providers"), "providers namespace should not exist")
	})
}

func TestLLMAdapter_AgentCreation(t *testing.T) {
	t.Run("create_simple_agent", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "createAgent" {
					// Return mock agent handle
					return map[string]interface{}{
						"id":    "agent-123",
						"model": "gpt-4",
						"type":  "llm",
					}, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Create agent from Lua
		err = L.DoString(`
			local llm = require("llm")
			local agent = llm.createAgent({
				model = "gpt-4",
				temperature = 0.7
			})
			assert(agent ~= nil)
			assert(agent.id == "agent-123")
			assert(agent.model == "gpt-4")
		`)
		assert.NoError(t, err)
	})

	t.Run("create_agent_with_tools", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "createAgent" && len(args) > 0 {
					config := args[0].(map[string]interface{})
					tools := config["tools"].([]interface{})
					return map[string]interface{}{
						"id":         "agent-456",
						"model":      config["model"],
						"toolsCount": len(tools),
					}, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Create agent with tools
		err = L.DoString(`
			local llm = require("llm")
			local agent = llm.createAgent({
				model = "gpt-4",
				tools = {"search", "calculator", "code_interpreter"}
			})
			assert(agent.toolsCount == 3)
		`)
		assert.NoError(t, err)
	})
}

func TestLLMAdapter_Completion(t *testing.T) {
	t.Run("simple_completion", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "complete" {
					prompt := args[0].(string)
					return map[string]interface{}{
						"text":   "Response to: " + prompt,
						"tokens": 10,
						"model":  "gpt-4",
					}, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test completion
		err = L.DoString(`
			local llm = require("llm")
			local response = llm.complete("Hello, world!")
			assert(response.text == "Response to: Hello, world!")
			assert(response.tokens == 10)
		`)
		assert.NoError(t, err)
	})

	t.Run("completion_with_options", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "complete" && len(args) >= 2 {
					options := args[1].(map[string]interface{})
					return map[string]interface{}{
						"text":        "Generated text",
						"temperature": options["temperature"],
						"maxTokens":   options["maxTokens"],
					}, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test completion with options
		err = L.DoString(`
			local llm = require("llm")
			local response = llm.complete("Generate text", {
				temperature = 0.8,
				maxTokens = 100
			})
			assert(response.temperature == 0.8)
			assert(response.maxTokens == 100)
		`)
		assert.NoError(t, err)
	})
}

func TestLLMAdapter_Streaming(t *testing.T) {
	t.Run("stream_completion", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "stream" {
					// Return a mock stream handle
					return map[string]interface{}{
						"streamId": "stream-789",
						"status":   "active",
					}, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test streaming
		err = L.DoString(`
			local llm = require("llm")
			local stream = llm.stream("Tell me a story", {
				onChunk = function(chunk)
					-- Mock callback
				end
			})
			assert(stream.streamId == "stream-789")
			assert(stream.status == "active")
		`)
		assert.NoError(t, err)
	})
}

func TestLLMAdapter_ModelManagement(t *testing.T) {
	t.Run("list_available_models", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "listModels" {
					// Return as a typed slice to avoid the []interface{} unpacking behavior
					models := []map[string]interface{}{
						{
							"id":          "gpt-4",
							"name":        "GPT-4",
							"maxTokens":   8192,
							"description": "Most capable model",
						},
						{
							"id":          "gpt-3.5-turbo",
							"name":        "GPT-3.5 Turbo",
							"maxTokens":   4096,
							"description": "Fast and efficient",
						},
					}
					return models, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test listing models
		err = L.DoString(`
			local llm = require("llm")
			-- Check that models namespace doesn't exist (flattened)
			assert(llm.models == nil, "models namespace should not exist (flattened)")
			-- Check that modelsList function exists instead
			assert(type(llm.modelsList) == "function", "modelsList should be a function")
			local models, err = llm.modelsList()
			assert(err == nil, "error calling modelsList: " .. tostring(err))
			assert(models ~= nil, "models should not be nil")
			assert(type(models) == "table", "models should be a table, got: " .. type(models))
			assert(#models == 2, "expected 2 models, got: " .. #models)
			assert(models[1].id == "gpt-4")
			assert(models[2].id == "gpt-3.5-turbo")
		`)
		assert.NoError(t, err)
	})

	t.Run("select_model", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "selectModel" {
					modelId := args[0].(string)
					return map[string]interface{}{
						"selected": modelId,
						"status":   "active",
					}, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test model selection
		err = L.DoString(`
			local llm = require("llm")
			local result = llm.selectModel("claude-3")
			assert(result.selected == "claude-3")
			assert(result.status == "active")
		`)
		assert.NoError(t, err)
	})
}

func TestLLMAdapter_TokenCounting(t *testing.T) {
	t.Run("count_tokens", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "countTokens" {
					text := args[0].(string)
					// Mock token counting (roughly 1 token per 4 chars)
					return map[string]interface{}{
						"tokens": len(text) / 4,
						"model":  "gpt-4",
					}, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test token counting
		err = L.DoString(`
			local llm = require("llm")
			local result = llm.countTokens("This is a test string for counting tokens")
			assert(result.tokens > 0)
			assert(result.model == "gpt-4")
		`)
		assert.NoError(t, err)
	})
}

func TestLLMAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_api_errors", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "complete" {
					return nil, errors.New("API rate limit exceeded")
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test error handling
		err = L.DoString(`
			local llm = require("llm")
			local response, err = llm.complete("Hello")
			assert(response == nil)
			assert(string.find(err, "rate limit"))
		`)
		assert.NoError(t, err)
	})
}

func TestLLMAdapter_ChainedOperations(t *testing.T) {
	t.Run("agent_with_chained_calls", func(t *testing.T) {
		callCount := 0
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				callCount++
				switch method {
				case "createAgent":
					return map[string]interface{}{
						"id":    "agent-chain",
						"model": "gpt-4",
					}, nil
				case "agentComplete":
					agentId := args[0].(string)
					prompt := args[1].(string)
					return map[string]interface{}{
						"text":    fmt.Sprintf("Agent %s response to: %s", agentId, prompt),
						"agentId": agentId,
					}, nil
				}
				return nil, errors.New("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test chained operations
		err = L.DoString(`
			local llm = require("llm")
			
			-- Create agent
			local agent = llm.createAgent({model = "gpt-4"})
			
			-- Use agent for completion
			local response = llm.agentComplete(agent.id, "Hello from agent")
			assert(response.agentId == agent.id)
			assert(string.find(response.text, "Agent agent%-chain response"))
		`)
		assert.NoError(t, err)
		assert.Equal(t, 2, callCount) // createAgent + agentComplete
	})
}

// Mock LLM bridge for testing
type mockLLMBridge struct {
	id          string
	metadata    engine.BridgeMetadata
	initialized bool
	callFunc    func(string, ...interface{}) (interface{}, error)
}

func (m *mockLLMBridge) GetID() string {
	return m.id
}

func (m *mockLLMBridge) GetMetadata() engine.BridgeMetadata {
	return m.metadata
}

func (m *mockLLMBridge) Initialize(ctx context.Context) error {
	m.initialized = true
	return nil
}

func (m *mockLLMBridge) Cleanup(ctx context.Context) error {
	m.initialized = false
	return nil
}

func (m *mockLLMBridge) IsInitialized() bool {
	return m.initialized
}

func (m *mockLLMBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return nil
}

func (m *mockLLMBridge) Methods() []engine.MethodInfo {
	// Return LLM-specific methods
	return []engine.MethodInfo{
		{Name: "createAgent", Description: "Create an LLM agent"},
		{Name: "complete", Description: "Generate completion"},
		{Name: "stream", Description: "Stream completion"},
		{Name: "listModels", Description: "List available models"},
		{Name: "selectModel", Description: "Select a model"},
		{Name: "countTokens", Description: "Count tokens in text"},
		{Name: "agentComplete", Description: "Complete using specific agent"},
	}
}

func (m *mockLLMBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	// Basic validation
	for _, method := range m.Methods() {
		if method.Name == name {
			return nil
		}
	}
	return errors.New("unknown method")
}

func (m *mockLLMBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// If callFunc is set, convert args and call it
	if m.callFunc != nil {
		// Convert ScriptValue args to interface{} for callFunc
		convertedArgs := make([]interface{}, len(args))
		for i, arg := range args {
			convertedArgs[i] = arg.ToGo()
		}

		result, err := m.callFunc(name, convertedArgs...)
		if err != nil {
			return nil, err
		}

		// Convert result to ScriptValue
		return convertResultToScriptValue(result), nil
	}

	// Default mock implementation
	switch name {
	case "createAgent":
		return engine.NewStringValue("mock-agent-id"), nil
	case "complete":
		return engine.NewStringValue("Mock completion response"), nil
	case "listModels":
		return engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("gpt-4"),
			engine.NewStringValue("gpt-3.5-turbo"),
		}), nil
	case "countTokens":
		return engine.NewNumberValue(42), nil
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

func (m *mockLLMBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"Agent": {
			GoType:     "*core.LLMAgent",
			ScriptType: "table",
		},
		"CompletionResponse": {
			GoType:     "llmdomain.Response",
			ScriptType: "table",
		},
	}
}

func (m *mockLLMBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:     engine.PermissionNetwork,
			Resource: "llm-api",
			Actions:  []string{"read", "write"},
		},
	}
}

func (m *mockLLMBridge) Call(method string, args ...interface{}) (interface{}, error) {
	if m.callFunc != nil {
		return m.callFunc(method, args...)
	}
	return nil, errors.New("method not implemented")
}

// Helper function to convert result to ScriptValue
func convertResultToScriptValue(result interface{}) engine.ScriptValue {
	if result == nil {
		return engine.NewNilValue()
	}

	switch v := result.(type) {
	case string:
		return engine.NewStringValue(v)
	case int:
		return engine.NewNumberValue(float64(v))
	case float64:
		return engine.NewNumberValue(v)
	case bool:
		return engine.NewBoolValue(v)
	case []interface{}:
		elements := make([]engine.ScriptValue, len(v))
		for i, elem := range v {
			elements[i] = convertResultToScriptValue(elem)
		}
		return engine.NewArrayValue(elements)
	case []map[string]interface{}:
		elements := make([]engine.ScriptValue, len(v))
		for i, elem := range v {
			elements[i] = convertResultToScriptValue(elem)
		}
		return engine.NewArrayValue(elements)
	case map[string]interface{}:
		fields := make(map[string]engine.ScriptValue)
		for k, val := range v {
			fields[k] = convertResultToScriptValue(val)
		}
		return engine.NewObjectValue(fields)
	case error:
		return engine.NewErrorValue(v)
	default:
		// For other types, convert to string
		return engine.NewStringValue(fmt.Sprintf("%v", v))
	}
}

// Mock pool bridge for testing
type mockPoolBridge struct {
	id          string
	initialized bool
	methods     map[string]func(args []engine.ScriptValue) (engine.ScriptValue, error)
}

func (m *mockPoolBridge) GetID() string                                               { return m.id }
func (m *mockPoolBridge) GetMetadata() engine.BridgeMetadata                          { return engine.BridgeMetadata{} }
func (m *mockPoolBridge) Initialize(ctx context.Context) error                        { m.initialized = true; return nil }
func (m *mockPoolBridge) Cleanup(ctx context.Context) error                           { m.initialized = false; return nil }
func (m *mockPoolBridge) IsInitialized() bool                                         { return m.initialized }
func (m *mockPoolBridge) RegisterWithEngine(e engine.ScriptEngine) error              { return nil }
func (m *mockPoolBridge) Methods() []engine.MethodInfo                                { return nil }
func (m *mockPoolBridge) ValidateMethod(name string, args []engine.ScriptValue) error { return nil }
func (m *mockPoolBridge) TypeMappings() map[string]engine.TypeMapping                 { return nil }
func (m *mockPoolBridge) RequiredPermissions() []engine.Permission                    { return nil }

func (m *mockPoolBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if method, ok := m.methods[name]; ok {
		return method(args)
	}
	return engine.NewErrorValue(fmt.Errorf("method not found: %s", name)), nil
}

func TestLLMAdapter_PoolEnhancement(t *testing.T) {
	t.Run("pool_methods_with_bridge", func(t *testing.T) {
		// Create mock LLM bridge
		llmBridge := &mockLLMBridge{
			id: "llm",
			metadata: engine.BridgeMetadata{
				Name: "LLM Bridge",
			},
		}

		// Create mock pool bridge
		poolBridge := &mockPoolBridge{
			id:          "pool",
			initialized: true,
			methods: map[string]func(args []engine.ScriptValue) (engine.ScriptValue, error){
				"createPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":     args[0],
						"strategy": args[2],
						"created":  engine.NewBoolValue(true),
					}), nil
				},
				"getPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name": args[0],
						"providers": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewStringValue("openai"),
							engine.NewStringValue("anthropic"),
						}),
						"strategy": engine.NewStringValue("round_robin"),
					}), nil
				},
				"listPools": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("main-pool"),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("backup-pool"),
						}),
					}), nil
				},
				"getPoolMetrics": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"total_requests": engine.NewNumberValue(100),
						"success_rate":   engine.NewNumberValue(0.95),
					}), nil
				},
				"getProviderHealth": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"provider": engine.NewStringValue("openai"),
							"health":   engine.NewStringValue("healthy"),
						}),
					}), nil
				},
			},
		}

		// Create adapter with pool bridge
		adapter := NewLLMAdapter(llmBridge, nil, poolBridge)

		// Should have pool methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "poolCreate")
		assert.Contains(t, methods, "poolGet")
		assert.Contains(t, methods, "poolList")
		assert.Contains(t, methods, "poolGetProviderHealth")
		assert.Contains(t, methods, "poolGetResponse")
		assert.Contains(t, methods, "poolReturnResponse")
	})

	t.Run("pool_methods_in_lua", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				switch method {
				case "createPool":
					return map[string]interface{}{
						"name":    "test-pool",
						"created": true,
					}, nil
				case "getPoolMetrics":
					return map[string]interface{}{
						"requests": 50,
						"errors":   2,
					}, nil
				}
				return nil, fmt.Errorf("unknown method: %s", method)
			},
		}

		poolBridge := &mockPoolBridge{
			id:          "pool",
			initialized: true,
			methods: map[string]func(args []engine.ScriptValue) (engine.ScriptValue, error){
				"getPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":     args[0],
						"strategy": engine.NewStringValue("failover"),
					}), nil
				},
				"listPools": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("pool1"),
						}),
					}), nil
				},
				"removePool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"resetPoolMetrics": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"generateMessageWithPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"role":    engine.NewStringValue("assistant"),
						"content": engine.NewStringValue("Hello from pool!"),
					}), nil
				},
				"streamWithPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"stream_id": engine.NewStringValue("stream-123"),
						"started":   engine.NewBoolValue(true),
					}), nil
				},
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, poolBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test flattened pool methods
		err = L.DoString(`
			local llm = require("llm")
			
			-- Test poolCreate (basic method)
			local pool, err = llm.poolCreate("test-pool", {"openai", "anthropic"}, "round_robin")
			assert(err == nil, "should not error")
			assert(pool.name == "test-pool", "should create pool")
			assert(pool.created == true, "should be created")
			
			-- Test poolGet (enhanced method)
			local pool2, err = llm.poolGet("my-pool")
			assert(err == nil, "should not error")
			assert(pool2.name == "my-pool", "should get pool")
			assert(pool2.strategy == "failover", "should have strategy")
			
			-- Test poolList
			local pools, err = llm.poolList()
			assert(err == nil, "should not error")
			assert(#pools == 1, "should have one pool")
			assert(pools[1].name == "pool1", "should have pool1")
			
			-- Test poolRemove
			local success, err = llm.poolRemove("old-pool")
			assert(err == nil, "should not error")
			assert(success == true, "should remove successfully")
			
			-- Test poolGetMetrics (basic method)
			local metrics, err = llm.poolGetMetrics("test-pool")
			assert(err == nil, "should not error")
			assert(metrics.requests == 50, "should have requests")
			assert(metrics.errors == 2, "should have errors")
			
			-- Test poolResetMetrics
			local success, err = llm.poolResetMetrics("test-pool")
			assert(err == nil, "should not error")
			assert(success == true, "should reset successfully")
			
			-- Test poolGenerateMessage
			local msg, err = llm.poolGenerateMessage("test-pool", {{role="user", content="Hi"}})
			assert(err == nil, "should not error")
			assert(msg.content == "Hello from pool!", "should generate message")
			
			-- Test poolStream
			local stream, err = llm.poolStream("test-pool", "Hello", {})
			assert(err == nil, "should not error")
			assert(stream.stream_id == "stream-123", "should have stream id")
			assert(stream.started == true, "should be started")
		`)
		assert.NoError(t, err)
	})

	t.Run("pool_object_pooling_methods", func(t *testing.T) {
		llmBridge := &mockLLMBridge{id: "llm"}

		poolBridge := &mockPoolBridge{
			id:          "pool",
			initialized: true,
			methods: map[string]func(args []engine.ScriptValue) (engine.ScriptValue, error){
				"getResponseFromPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":      engine.NewStringValue("resp-123"),
						"content": engine.NewStringValue(""),
					}), nil
				},
				"returnResponseToPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"getTokenFromPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"value": engine.NewStringValue("token-456"),
						"used":  engine.NewBoolValue(false),
					}), nil
				},
				"returnTokenToPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"getChannelFromPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":   engine.NewStringValue("chan-789"),
						"open": engine.NewBoolValue(true),
					}), nil
				},
				"returnChannelToPool": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, poolBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test object pooling methods
		err = L.DoString(`
			local llm = require("llm")
			
			-- Test response pooling
			local resp, err = llm.poolGetResponse()
			assert(err == nil, "should not error")
			assert(resp.id == "resp-123", "should get response")
			
			local success, err = llm.poolReturnResponse(resp)
			assert(err == nil, "should not error")
			assert(success == true, "should return response")
			
			-- Test token pooling
			local token, err = llm.poolGetToken()
			assert(err == nil, "should not error")
			assert(token.value == "token-456", "should get token")
			assert(token.used == false, "should not be used")
			
			local success, err = llm.poolReturnToken(token)
			assert(err == nil, "should not error")
			assert(success == true, "should return token")
			
			-- Test channel pooling
			local chan, err = llm.poolGetChannel()
			assert(err == nil, "should not error")
			assert(chan.id == "chan-789", "should get channel")
			assert(chan.open == true, "should be open")
			
			local success, err = llm.poolReturnChannel(chan)
			assert(err == nil, "should not error")
			assert(success == true, "should return channel")
		`)
		assert.NoError(t, err)
	})

	t.Run("pool_methods_without_bridge", func(t *testing.T) {
		// Create adapter without pool bridge
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "createPool" {
					return map[string]interface{}{"created": true}, nil
				}
				return nil, fmt.Errorf("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Enhanced pool methods should not exist
		err = L.DoString(`
			local llm = require("llm")
			
			-- Basic pool methods should still work
			local pool, err = llm.poolCreate("test", {"openai"}, "round_robin")
			assert(err == nil, "poolCreate should work")
			assert(pool.created == true)
			
			-- Enhanced pool methods should not exist
			assert(llm.poolGet == nil, "poolGet should not exist")
			assert(llm.poolList == nil, "poolList should not exist")
			assert(llm.poolRemove == nil, "poolRemove should not exist")
			assert(llm.poolGetProviderHealth == nil, "poolGetProviderHealth should not exist")
			assert(llm.poolGetResponse == nil, "poolGetResponse should not exist")
		`)
		assert.NoError(t, err)
	})

	t.Run("namespace_flattening", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				switch method {
				case "createProvider":
					return map[string]interface{}{"name": "test-provider"}, nil
				case "listProviders":
					return []map[string]interface{}{{"name": "provider1"}}, nil
				case "listModels":
					return []map[string]interface{}{{"name": "gpt-4"}}, nil
				case "getModelInfo":
					return map[string]interface{}{"name": "gpt-4", "context": 8192}, nil
				}
				return nil, fmt.Errorf("unknown method: %s", method)
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test that namespaces have been flattened
		err = L.DoString(`
			local llm = require("llm")
			
			-- Provider methods should be flattened
			local provider, err = llm.providersCreate("openai", "my-provider", {})
			assert(err == nil, "providersCreate should work")
			assert(provider.name == "test-provider")
			
			local providers, err = llm.providersList()
			assert(err == nil, "providersList should work")
			assert(#providers == 1)
			assert(providers[1].name == "provider1")
			
			-- Model methods should be flattened
			local models, err = llm.modelsList()
			assert(err == nil, "modelsList should work")
			assert(#models == 1)
			assert(models[1].name == "gpt-4")
			
			local info, err = llm.modelsGetInfo("gpt-4")
			assert(err == nil, "modelsGetInfo should work")
			assert(info.name == "gpt-4")
			assert(info.context == 8192)
			
			-- Old namespaces should not exist
			assert(llm.providers == nil, "providers namespace should not exist")
			assert(llm.models == nil, "models namespace should not exist")
			assert(llm.pool == nil, "pool namespace should not exist")
		`)
		assert.NoError(t, err)
	})
}
