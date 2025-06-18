// ABOUTME: Tests for the LLM bridge that provides language model functionality
// ABOUTME: Tests provider management, text generation, streaming, and schema handling

package llm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEngine implements engine.ScriptEngine for testing
type MockEngine struct {
	bridges map[string]engine.Bridge
}

func NewMockEngine() *MockEngine {
	return &MockEngine{
		bridges: make(map[string]engine.Bridge),
	}
}

func (m *MockEngine) RegisterBridge(bridge engine.Bridge) error {
	m.bridges[bridge.GetID()] = bridge
	return nil
}

func (m *MockEngine) GetBridge(id string) (engine.Bridge, error) {
	bridge, ok := m.bridges[id]
	if !ok {
		return nil, fmt.Errorf("bridge not found: %s", id)
	}
	return bridge, nil
}

func (m *MockEngine) Execute(script string) (interface{}, error) {
	return nil, nil
}

func (m *MockEngine) ExecuteFile(filename string) (interface{}, error) {
	return nil, nil
}

func (m *MockEngine) Close() error {
	return nil
}

func TestLLMBridge_Initialization(t *testing.T) {
	bridge := NewLLMBridge()
	assert.NotNil(t, bridge)

	// Test initial state
	assert.False(t, bridge.IsInitialized())
	assert.Equal(t, "llm", bridge.GetID())

	// Test metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "llm", metadata.Name)
	assert.NotEmpty(t, metadata.Version)
	assert.NotEmpty(t, metadata.Description)

	// Test initialization
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test cleanup
	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestLLMBridge_Methods(t *testing.T) {
	bridge := NewLLMBridge()
	methods := bridge.Methods()

	// Check that we have methods defined
	assert.NotEmpty(t, methods)

	// Check for key methods
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	// Provider management methods
	assert.True(t, methodNames["setProvider"])
	assert.True(t, methodNames["getProvider"])
	assert.True(t, methodNames["listProviders"])

	// Generation methods
	assert.True(t, methodNames["generate"])
	assert.True(t, methodNames["generateMessage"])
	assert.True(t, methodNames["stream"])

	// Schema methods
	assert.True(t, methodNames["generateWithSchema"])
	assert.True(t, methodNames["addResponseSchema"])
	assert.True(t, methodNames["getResponseSchema"])
}

func TestLLMBridge_ProviderManagement(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test setProvider
	args := []engine.ScriptValue{
		engine.NewStringValue("mock-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"model":       "mock-model",
			"temperature": 0.7,
		})),
	}

	result, err := bridge.ExecuteMethod(ctx, "setProvider", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test getProvider
	args = []engine.ScriptValue{}
	result, err = bridge.ExecuteMethod(ctx, "getProvider", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	providerInfo := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "mock-provider", providerInfo["name"])

	// Test listProviders
	args = []engine.ScriptValue{}
	result, err = bridge.ExecuteMethod(ctx, "listProviders", args)
	assert.NoError(t, err)

	arr, ok := result.(engine.ArrayValue)
	require.True(t, ok)
	providers := arr.ToGo().([]interface{})
	assert.Len(t, providers, 1)
}

func TestLLMBridge_TextGeneration(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Set mock provider
	args := []engine.ScriptValue{
		engine.NewStringValue("mock-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"model": "mock-model",
		})),
	}
	_, err := bridge.ExecuteMethod(ctx, "setProvider", args)
	require.NoError(t, err)

	// Test generate
	args = []engine.ScriptValue{
		engine.NewStringValue("Hello, world!"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  100,
		})),
	}

	result, err := bridge.ExecuteMethod(ctx, "generate", args)
	assert.NoError(t, err)

	// Check result is a string
	str, ok := result.(engine.StringValue)
	require.True(t, ok)
	assert.NotEmpty(t, str.Value())
}

func TestLLMBridge_MessageGeneration(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Set mock provider
	args := []engine.ScriptValue{
		engine.NewStringValue("mock-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"model": "mock-model",
		})),
	}
	_, err := bridge.ExecuteMethod(ctx, "setProvider", args)
	require.NoError(t, err)

	// Test generateMessage
	messages := []interface{}{
		map[string]interface{}{
			"role":    "system",
			"content": "You are a helpful assistant.",
		},
		map[string]interface{}{
			"role":    "user",
			"content": "Hello!",
		},
	}

	args = []engine.ScriptValue{
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(messages)),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"temperature": 0.7,
		})),
	}

	result, err := bridge.ExecuteMethod(ctx, "generateMessage", args)
	assert.NoError(t, err)

	// Check result is an object with message
	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	response := obj.ToGo().(map[string]interface{})
	assert.NotNil(t, response["message"])
}

func TestLLMBridge_Streaming(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Set mock provider
	args := []engine.ScriptValue{
		engine.NewStringValue("mock-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"model": "mock-model",
		})),
	}
	_, err := bridge.ExecuteMethod(ctx, "setProvider", args)
	require.NoError(t, err)

	// Test stream
	args = []engine.ScriptValue{
		engine.NewStringValue("Tell me a story"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"temperature": 0.7,
		})),
	}

	result, err := bridge.ExecuteMethod(ctx, "stream", args)
	assert.NoError(t, err)

	// Check result has stream_id
	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	streamInfo := obj.ToGo().(map[string]interface{})
	assert.NotEmpty(t, streamInfo["stream_id"])
}

func TestLLMBridge_SchemaValidation(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test addResponseSchema
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"age": map[string]interface{}{
				"type": "number",
			},
		},
		"required": []string{"name"},
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("person"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(schema)),
	}

	result, err := bridge.ExecuteMethod(ctx, "addResponseSchema", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test getResponseSchema
	args = []engine.ScriptValue{
		engine.NewStringValue("person"),
	}

	result, err = bridge.ExecuteMethod(ctx, "getResponseSchema", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	retrievedSchema := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "person", retrievedSchema["name"])
}

func TestLLMBridge_StructuredGeneration(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Set mock provider
	args := []engine.ScriptValue{
		engine.NewStringValue("mock-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"model": "mock-model",
		})),
	}
	_, err := bridge.ExecuteMethod(ctx, "setProvider", args)
	require.NoError(t, err)

	// Add schema
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"items": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	args = []engine.ScriptValue{
		engine.NewStringValue("list"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(schema)),
	}
	_, err = bridge.ExecuteMethod(ctx, "addResponseSchema", args)
	require.NoError(t, err)

	// Test generateWithSchema
	args = []engine.ScriptValue{
		engine.NewStringValue("Generate a list of colors"),
		engine.NewStringValue("list"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"temperature": 0.7,
		})),
	}

	result, err := bridge.ExecuteMethod(ctx, "generateWithSchema", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	response := obj.ToGo().(map[string]interface{})
	assert.NotNil(t, response["parsed"])
}

func TestLLMBridge_ErrorHandling(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Test method call before initialization
	args := []engine.ScriptValue{
		engine.NewStringValue("test"),
	}

	result, err := bridge.ExecuteMethod(ctx, "generate", args)
	assert.NoError(t, err) // Error is returned as ErrorValue

	errVal, ok := result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "not initialized")

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test invalid method
	result, err = bridge.ExecuteMethod(ctx, "invalidMethod", args)
	assert.NoError(t, err)

	errVal, ok = result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "unknown method")

	// Test wrong argument types
	args = []engine.ScriptValue{
		engine.NewNumberValue(123), // Should be string
	}

	result, err = bridge.ExecuteMethod(ctx, "generate", args)
	assert.NoError(t, err)

	errVal, ok = result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "expected string")
}

func TestLLMBridge_ProviderMetrics(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Set provider
	args := []engine.ScriptValue{
		engine.NewStringValue("test-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"model": "test-model",
		})),
	}
	_, err := bridge.ExecuteMethod(ctx, "setProvider", args)
	require.NoError(t, err)

	// Generate some text to create metrics
	args = []engine.ScriptValue{
		engine.NewStringValue("Hello"),
	}
	_, err = bridge.ExecuteMethod(ctx, "generate", args)
	require.NoError(t, err)

	// Test getProviderMetrics
	args = []engine.ScriptValue{
		engine.NewStringValue("test-provider"),
	}

	result, err := bridge.ExecuteMethod(ctx, "getProviderMetrics", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	metrics := obj.ToGo().(map[string]interface{})
	assert.NotNil(t, metrics["total_requests"])
}

func TestLLMBridge_FallbackChain(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test setFallbackChain
	providers := []interface{}{
		"primary-provider",
		"secondary-provider",
		"tertiary-provider",
	}

	args := []engine.ScriptValue{
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
	}

	result, err := bridge.ExecuteMethod(ctx, "setFallbackChain", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test getFallbackChain
	args = []engine.ScriptValue{}
	result, err = bridge.ExecuteMethod(ctx, "getFallbackChain", args)
	assert.NoError(t, err)

	arr, ok := result.(engine.ArrayValue)
	require.True(t, ok)
	chain := arr.ToGo().([]interface{})
	assert.Len(t, chain, 3)
}

func TestLLMBridge_ValidateMethod(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test valid method with correct args
	args := []engine.ScriptValue{
		engine.NewStringValue("test"),
	}
	err := bridge.ValidateMethod("generate", args)
	assert.NoError(t, err)

	// Test invalid method
	err = bridge.ValidateMethod("invalidMethod", args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown method")

	// Test insufficient args
	args = []engine.ScriptValue{}
	err = bridge.ValidateMethod("generate", args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least")
}

func TestLLMBridge_TypeMappings(t *testing.T) {
	bridge := NewLLMBridge()
	mappings := bridge.TypeMappings()

	assert.NotEmpty(t, mappings)

	// Check for key type mappings
	assert.Contains(t, mappings, "provider")
	assert.Contains(t, mappings, "response")
	assert.Contains(t, mappings, "schema")

	// Verify mapping structure
	providerMapping := mappings["provider"]
	assert.Equal(t, "bridge.Provider", providerMapping.GoType)
	assert.Equal(t, "object", providerMapping.ScriptType)
}

func TestLLMBridge_RequiredPermissions(t *testing.T) {
	bridge := NewLLMBridge()
	permissions := bridge.RequiredPermissions()

	assert.NotEmpty(t, permissions)

	// Check for network permission (API calls)
	hasNetwork := false
	for _, perm := range permissions {
		if perm.Type == engine.PermissionNetwork {
			hasNetwork = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
	}
	assert.True(t, hasNetwork)
}

func TestLLMBridge_Concurrency(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Set provider
	args := []engine.ScriptValue{
		engine.NewStringValue("test-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"model": "test-model",
		})),
	}
	_, err := bridge.ExecuteMethod(ctx, "setProvider", args)
	require.NoError(t, err)

	// Test concurrent generate calls
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(n int) {
			defer func() { done <- true }()

			args := []engine.ScriptValue{
				engine.NewStringValue("Hello " + string(rune(n))),
			}

			result, err := bridge.ExecuteMethod(ctx, "generate", args)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}(i)
	}

	// Wait for all goroutines
	timeout := time.After(5 * time.Second)
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}
