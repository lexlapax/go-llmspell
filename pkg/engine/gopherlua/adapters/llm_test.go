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

		// Create adapter
		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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
			local models = llm.listModels()
			assert(#models == 2)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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

		adapter := NewLLMAdapter(llmBridge)
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
	id           string
	metadata     engine.BridgeMetadata
	initialized  bool
	callFunc     func(string, ...interface{}) (interface{}, error)
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

func (m *mockLLMBridge) ValidateMethod(name string, args []interface{}) error {
	// Basic validation
	for _, method := range m.Methods() {
		if method.Name == name {
			return nil
		}
	}
	return errors.New("unknown method")
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
