// ABOUTME: Tests for the providers bridge that manages LLM provider creation and configuration
// ABOUTME: Tests provider templates, multi-provider setups, and environment-based configuration

package llm

import (
	"context"
	"os"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvidersBridge_Initialization(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	assert.NotNil(t, bridge)

	// Test initial state
	assert.False(t, bridge.IsInitialized())
	assert.Equal(t, "providers", bridge.GetID())

	// Test metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "providers", metadata.Name)
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

func TestProvidersBridge_Methods(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	methods := bridge.Methods()

	// Check that we have methods defined
	assert.NotEmpty(t, methods)

	// Check for key methods
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	// Provider creation methods
	assert.True(t, methodNames["createProvider"])
	assert.True(t, methodNames["createProviderFromEnvironment"])
	assert.True(t, methodNames["getProvider"])
	assert.True(t, methodNames["listProviders"])
	assert.True(t, methodNames["removeProvider"])

	// Template methods
	assert.True(t, methodNames["getProviderTemplate"])
	assert.True(t, methodNames["listProviderTemplates"])
	assert.True(t, methodNames["validateProviderConfig"])

	// Multi-provider methods
	assert.True(t, methodNames["createMultiProvider"])
	assert.True(t, methodNames["configureMultiProvider"])
	assert.True(t, methodNames["getMultiProvider"])
}

func TestProvidersBridge_CreateProvider(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test createProvider
	config := map[string]interface{}{
		"api_key":     "test-key",
		"model":       "gpt-3.5-turbo",
		"temperature": 0.7,
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("openai"),
		engine.NewStringValue("test-openai"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
	}

	result, err := bridge.ExecuteMethod(ctx, "createProvider", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	providerInfo := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "test-openai", providerInfo["name"])
	assert.Equal(t, "openai", providerInfo["type"])
	assert.NotEmpty(t, providerInfo["created"])
}

func TestProvidersBridge_CreateProviderFromEnvironment(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Set environment variables
	err = os.Setenv("OPENAI_API_KEY", "test-api-key")
	require.NoError(t, err)
	defer func() {
		_ = os.Unsetenv("OPENAI_API_KEY")
	}()

	// Test createProviderFromEnvironment
	args := []engine.ScriptValue{
		engine.NewStringValue("openai"),
		engine.NewStringValue("env-openai"),
	}

	result, err := bridge.ExecuteMethod(ctx, "createProviderFromEnvironment", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	providerInfo := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "env-openai", providerInfo["name"])
	assert.Equal(t, "openai", providerInfo["type"])
	assert.Equal(t, "environment", providerInfo["source"])
}

func TestProvidersBridge_ProviderTemplates(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test listProviderTemplates
	args := []engine.ScriptValue{}
	result, err := bridge.ExecuteMethod(ctx, "listProviderTemplates", args)
	assert.NoError(t, err)

	arr, ok := result.(engine.ArrayValue)
	require.True(t, ok)
	templates := arr.ToGo().([]interface{})
	assert.NotEmpty(t, templates)

	// Check for default templates
	templateTypes := make(map[string]bool)
	for _, template := range templates {
		tmpl := template.(map[string]interface{})
		templateTypes[tmpl["type"].(string)] = true
	}
	assert.True(t, templateTypes["openai"])
	assert.True(t, templateTypes["anthropic"])
	assert.True(t, templateTypes["mock"])

	// Test getProviderTemplate
	args = []engine.ScriptValue{
		engine.NewStringValue("openai"),
	}

	result, err = bridge.ExecuteMethod(ctx, "getProviderTemplate", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	template := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "openai", template["type"])
	assert.NotEmpty(t, template["description"])
	assert.NotEmpty(t, template["requiredEnvVars"])
}

func TestProvidersBridge_ValidateProviderConfig(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test valid config
	config := map[string]interface{}{
		"OPENAI_API_KEY": "test-key",
		"model":          "gpt-3.5-turbo",
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("openai"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
	}

	result, err := bridge.ExecuteMethod(ctx, "validateProviderConfig", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	validation := obj.ToGo().(map[string]interface{})
	assert.True(t, validation["valid"].(bool))
	errors := validation["errors"].([]interface{})
	assert.Empty(t, errors)

	// Test invalid config (missing required field)
	config = map[string]interface{}{
		"model": "gpt-3.5-turbo",
	}

	args = []engine.ScriptValue{
		engine.NewStringValue("openai"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
	}

	result, err = bridge.ExecuteMethod(ctx, "validateProviderConfig", args)
	assert.NoError(t, err)

	obj, ok = result.(engine.ObjectValue)
	require.True(t, ok)
	validation = obj.ToGo().(map[string]interface{})
	assert.False(t, validation["valid"].(bool))
	errors = validation["errors"].([]interface{})
	assert.NotEmpty(t, errors)
}

func TestProvidersBridge_MultiProvider(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create some providers first
	for i, name := range []string{"provider1", "provider2", "provider3"} {
		config := map[string]interface{}{
			"api_key": "test-key-" + string(rune('1'+i)),
			"model":   "mock-model",
		}

		args := []engine.ScriptValue{
			engine.NewStringValue("mock"),
			engine.NewStringValue(name),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
		}
		_, err := bridge.ExecuteMethod(ctx, "createProvider", args)
		require.NoError(t, err)
	}

	// Test createMultiProvider
	providers := []interface{}{
		map[string]interface{}{
			"name":    "provider1",
			"weight":  0.5,
			"primary": true,
		},
		map[string]interface{}{
			"name":   "provider2",
			"weight": 0.3,
		},
		map[string]interface{}{
			"name":   "provider3",
			"weight": 0.2,
		},
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("multi1"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("consensus"),
	}

	result, err := bridge.ExecuteMethod(ctx, "createMultiProvider", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	multiInfo := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "multi1", multiInfo["name"])
	assert.Equal(t, "consensus", multiInfo["strategy"])
	assert.Equal(t, float64(3), multiInfo["providers"])
}

func TestProvidersBridge_ConfigureMultiProvider(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create providers and multi-provider
	for _, name := range []string{"provider1", "provider2"} {
		config := map[string]interface{}{
			"api_key": "test-key",
			"model":   "mock-model",
		}

		args := []engine.ScriptValue{
			engine.NewStringValue("mock"),
			engine.NewStringValue(name),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
		}
		_, err := bridge.ExecuteMethod(ctx, "createProvider", args)
		require.NoError(t, err)
	}

	providers := []interface{}{
		map[string]interface{}{"name": "provider1", "weight": 0.6},
		map[string]interface{}{"name": "provider2", "weight": 0.4},
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("multi1"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("fastest"),
	}
	_, err := bridge.ExecuteMethod(ctx, "createMultiProvider", args)
	require.NoError(t, err)

	// Test configureMultiProvider
	config := map[string]interface{}{
		"consensusThreshold": 0.8,
		"timeout":            30.0,
		"retryOnFailure":     true,
	}

	args = []engine.ScriptValue{
		engine.NewStringValue("multi1"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
	}

	result, err := bridge.ExecuteMethod(ctx, "configureMultiProvider", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify configuration
	args = []engine.ScriptValue{
		engine.NewStringValue("multi1"),
	}

	result, err = bridge.ExecuteMethod(ctx, "getMultiProvider", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	multiInfo := obj.ToGo().(map[string]interface{})
	multiConfig := multiInfo["config"].(map[string]interface{})
	assert.Equal(t, 0.8, multiConfig["consensusThreshold"])
	assert.Equal(t, 30.0, multiConfig["timeout"])
	assert.True(t, multiConfig["retryOnFailure"].(bool))
}

func TestProvidersBridge_ProviderOperations(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create a provider
	config := map[string]interface{}{
		"api_key": "test-key",
		"model":   "mock-model",
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("mock"),
		engine.NewStringValue("test-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
	}
	_, err := bridge.ExecuteMethod(ctx, "createProvider", args)
	require.NoError(t, err)

	// Test generateWithProvider
	args = []engine.ScriptValue{
		engine.NewStringValue("test-provider"),
		engine.NewStringValue("Hello, world!"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  100,
		})),
	}

	result, err := bridge.ExecuteMethod(ctx, "generateWithProvider", args)
	assert.NoError(t, err)

	str, ok := result.(engine.StringValue)
	require.True(t, ok)
	assert.Contains(t, str.Value(), "Hello, world!")
}

func TestProvidersBridge_MockProvider(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test createMockProvider
	responses := []interface{}{
		"Response 1",
		"Response 2",
		"Response 3",
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("test-mock"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(responses)),
	}

	result, err := bridge.ExecuteMethod(ctx, "createMockProvider", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	mockInfo := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "test-mock", mockInfo["name"])
	assert.Equal(t, "mock", mockInfo["type"])
	assert.Equal(t, float64(3), mockInfo["responses"])

	// Test generating with mock provider
	// The mock uses len(prompt) % len(responses) to select response
	testCases := []struct {
		prompt   string
		expected string
	}{
		{"ABC", "Response 1"},   // len=3, 3%3=0
		{"ABCD", "Response 2"},  // len=4, 4%3=1
		{"ABCDE", "Response 3"}, // len=5, 5%3=2
	}

	for _, tc := range testCases {
		args = []engine.ScriptValue{
			engine.NewStringValue("test-mock"),
			engine.NewStringValue(tc.prompt),
		}
		result, err := bridge.ExecuteMethod(ctx, "generateWithProvider", args)
		assert.NoError(t, err)

		str, ok := result.(engine.StringValue)
		require.True(t, ok)
		assert.Equal(t, tc.expected, str.Value())
	}
}

func TestProvidersBridge_ExportImportConfig(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create some providers
	for i := 0; i < 3; i++ {
		config := map[string]interface{}{
			"api_key": "test-key-" + string(rune('1'+i)),
			"model":   "mock-model",
		}

		args := []engine.ScriptValue{
			engine.NewStringValue("mock"),
			engine.NewStringValue("provider" + string(rune('1'+i))),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
		}
		_, err := bridge.ExecuteMethod(ctx, "createProvider", args)
		require.NoError(t, err)
	}

	// Test exportProviderConfig
	args := []engine.ScriptValue{}
	result, err := bridge.ExecuteMethod(ctx, "exportProviderConfig", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	exportedConfig := obj.ToGo().(map[string]interface{})
	assert.NotNil(t, exportedConfig["providers"])
	assert.NotNil(t, exportedConfig["templates"])

	// Test importProviderConfig
	args = []engine.ScriptValue{
		engine.NewObjectValue(engine.ConvertMapToScriptValue(exportedConfig)),
	}

	result, err = bridge.ExecuteMethod(ctx, "importProviderConfig", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestProvidersBridge_ErrorHandling(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Test method call before initialization
	args := []engine.ScriptValue{
		engine.NewStringValue("test"),
	}

	result, err := bridge.ExecuteMethod(ctx, "getProvider", args)
	assert.NoError(t, err)

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

	// Test non-existent provider
	result, err = bridge.ExecuteMethod(ctx, "getProvider", args)
	assert.NoError(t, err)

	errVal, ok = result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "not found")

	// Test invalid provider type
	args = []engine.ScriptValue{
		engine.NewStringValue("invalid-type"),
		engine.NewStringValue("test-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(map[string]interface{}{})),
	}

	result, err = bridge.ExecuteMethod(ctx, "createProvider", args)
	assert.NoError(t, err)

	errVal, ok = result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "unknown provider type")
}

func TestProvidersBridge_TypeMappings(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	mappings := bridge.TypeMappings()

	assert.NotEmpty(t, mappings)

	// Check for key type mappings
	assert.Contains(t, mappings, "provider")
	assert.Contains(t, mappings, "provider_template")
	assert.Contains(t, mappings, "multi_provider")

	// Verify mapping structure
	providerMapping := mappings["provider"]
	assert.Equal(t, "bridge.Provider", providerMapping.GoType)
	assert.Equal(t, "object", providerMapping.ScriptType)
}

func TestProvidersBridge_RequiredPermissions(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	permissions := bridge.RequiredPermissions()

	assert.NotEmpty(t, permissions)

	// Check for permissions
	hasNetwork := false
	hasProcess := false

	for _, perm := range permissions {
		if perm.Type == engine.PermissionNetwork {
			hasNetwork = true
			assert.Equal(t, "llm.providers", perm.Resource)
		}
		if perm.Type == engine.PermissionProcess {
			hasProcess = true
			assert.Equal(t, "provider.registry", perm.Resource)
		}
	}

	assert.True(t, hasNetwork)
	assert.True(t, hasProcess)
}

func TestProvidersBridge_ProviderMetadata(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create a provider
	config := map[string]interface{}{
		"api_key": "test-key",
		"model":   "mock-model",
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("mock"),
		engine.NewStringValue("test-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
	}
	_, err := bridge.ExecuteMethod(ctx, "createProvider", args)
	require.NoError(t, err)

	// Test setProviderMetadata
	metadata := map[string]interface{}{
		"version":      "1.0.0",
		"capabilities": []string{"generate", "stream"},
		"maxTokens":    4096,
	}

	args = []engine.ScriptValue{
		engine.NewStringValue("test-provider"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(metadata)),
	}

	result, err := bridge.ExecuteMethod(ctx, "setProviderMetadata", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test getProviderMetadata
	args = []engine.ScriptValue{
		engine.NewStringValue("test-provider"),
	}

	result, err = bridge.ExecuteMethod(ctx, "getProviderMetadata", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	providerMeta := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "test-provider", providerMeta["name"])
	assert.NotNil(t, providerMeta["metadata"])
}

func TestProvidersBridge_ListProvidersByCapability(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewProvidersBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create providers
	for i := 0; i < 3; i++ {
		config := map[string]interface{}{
			"api_key": "test-key-" + string(rune('1'+i)),
			"model":   "mock-model",
		}

		args := []engine.ScriptValue{
			engine.NewStringValue("mock"),
			engine.NewStringValue("provider" + string(rune('1'+i))),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
		}
		_, err := bridge.ExecuteMethod(ctx, "createProvider", args)
		require.NoError(t, err)
	}

	// Test listProvidersByCapability
	args := []engine.ScriptValue{
		engine.NewStringValue("generate"),
	}

	result, err := bridge.ExecuteMethod(ctx, "listProvidersByCapability", args)
	assert.NoError(t, err)

	arr, ok := result.(engine.ArrayValue)
	require.True(t, ok)
	providers := arr.ToGo().([]interface{})
	assert.Len(t, providers, 3) // All mock providers have "generate" capability

	for _, provider := range providers {
		p := provider.(map[string]interface{})
		assert.Equal(t, "generate", p["capability"])
	}
}
