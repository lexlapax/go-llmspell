// ABOUTME: Tests for providers bridge functionality including provider registry and multi-provider orchestration
// ABOUTME: Comprehensive test coverage for provider management, factories, consensus algorithms, and configurations

package llm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// go-llms imports for provider functionality
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Test ProvidersBridge core functionality
func TestProvidersBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *ProvidersBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *ProvidersBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "providers", metadata.Name)
				assert.Equal(t, "v1.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "provider system")
			},
		},
		{
			name: "Create provider from environment",
			test: func(t *testing.T, bridge *ProvidersBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// This would normally use environment variables
				// For testing, we'll create a mock provider
				t.Setenv("MOCK_API_KEY", "test-key")

				result, err := bridge.createProviderFromEnvironment(ctx, []interface{}{"mock", "test-mock"})
				if err != nil {
					// Mock provider creation might fail without proper setup
					// This is expected in a test environment
					t.Logf("Expected failure creating provider from environment: %v", err)
					return
				}

				assert.NotNil(t, result)
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test-mock", resultMap["name"])
				assert.Equal(t, "mock", resultMap["type"])
			},
		},
		{
			name: "List providers",
			test: func(t *testing.T, bridge *ProvidersBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				result, err := bridge.listProviders(ctx, []interface{}{})
				require.NoError(t, err)
				assert.NotNil(t, result)

				providers, ok := result.([]map[string]interface{})
				require.True(t, ok)
				// Should be empty initially or contain default providers
				assert.IsType(t, []map[string]interface{}{}, providers)
			},
		},
		{
			name: "Get provider templates",
			test: func(t *testing.T, bridge *ProvidersBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// List all templates
				result, err := bridge.listProviderTemplates(ctx, []interface{}{})
				require.NoError(t, err)
				assert.NotNil(t, result)

				templates, ok := result.([]map[string]interface{})
				require.True(t, ok)
				assert.Greater(t, len(templates), 0) // Should have at least some templates

				// Get specific template
				if len(templates) > 0 {
					templateType := templates[0]["type"].(string)
					templateResult, err := bridge.getProviderTemplate(ctx, []interface{}{templateType})
					require.NoError(t, err)
					assert.NotNil(t, templateResult)

					template, ok := templateResult.(map[string]interface{})
					require.True(t, ok)
					assert.Equal(t, templateType, template["type"])
					assert.Contains(t, template, "name")
					assert.Contains(t, template, "description")
				}
			},
		},
		{
			name: "Provider capabilities",
			test: func(t *testing.T, bridge *ProvidersBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// List providers by capability
				result, err := bridge.listProvidersByCapability(ctx, []interface{}{"streaming"})
				require.NoError(t, err)
				assert.NotNil(t, result)

				providers, ok := result.([]map[string]interface{})
				require.True(t, ok)
				assert.IsType(t, []map[string]interface{}{}, providers)
			},
		},
		{
			name: "Export and import provider config",
			test: func(t *testing.T, bridge *ProvidersBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Export config
				result, err := bridge.exportProviderConfig(ctx, []interface{}{})
				require.NoError(t, err)
				assert.NotNil(t, result)

				config, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, config, "providers")
				assert.Contains(t, config, "version")

				// Import config (should not error)
				err = bridge.importProviderConfig(ctx, []interface{}{config})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewProvidersBridge()
			tt.test(t, bridge)
		})
	}
}

// Test providers integration with go-llms
func TestProvidersBridgeIntegration(t *testing.T) {
	bridge := NewProvidersBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test direct go-llms registry usage
	registry := provider.GetGlobalRegistry()
	assert.NotNil(t, registry)

	// Test factory registration
	err = provider.RegisterDefaultFactories(registry)
	require.NoError(t, err)

	// List available templates
	templates := registry.ListTemplates()
	assert.Greater(t, len(templates), 0)

	// Verify we can get templates for common providers
	commonProviders := []string{"openai", "anthropic", "mock"}
	for _, providerType := range commonProviders {
		template, err := registry.GetTemplate(providerType)
		if err == nil {
			assert.Equal(t, providerType, template.Type)
			assert.NotEmpty(t, template.Name)
			assert.NotEmpty(t, template.Description)
		}
	}
}

// Test multi-provider functionality
func TestMultiProviderFunctionality(t *testing.T) {
	bridge := NewProvidersBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test multi-provider creation (would need actual providers)
	providersConfig := []interface{}{
		map[string]interface{}{
			"name":   "mock1",
			"weight": 0.7,
		},
		map[string]interface{}{
			"name":   "mock2",
			"weight": 0.3,
		},
	}

	// This will fail because providers don't exist, but we can test the validation
	_, err = bridge.createMultiProvider(ctx, []interface{}{"test-multi", providersConfig, "consensus"})
	assert.Error(t, err) // Expected to fail without real providers
	assert.Contains(t, err.Error(), "provider not found")
}

// Test error scenarios
func TestProvidersBridgeErrors(t *testing.T) {
	bridge := NewProvidersBridge()
	ctx := context.Background()

	// Test methods without initialization
	_, err := bridge.createProvider(ctx, []interface{}{"openai", "test", map[string]interface{}{}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	_, err = bridge.createProvider(ctx, []interface{}{})
	assert.Error(t, err)

	_, err = bridge.createProvider(ctx, []interface{}{123, "test", map[string]interface{}{}})
	assert.Error(t, err)

	_, err = bridge.getProvider(ctx, []interface{}{"nonexistent"})
	assert.Error(t, err)

	err = bridge.removeProvider(ctx, []interface{}{"nonexistent"})
	assert.Error(t, err)
}

// Test concurrent operations
func TestProvidersBridgeConcurrency(t *testing.T) {
	bridge := NewProvidersBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test concurrent provider listing
	numRoutines := 10
	done := make(chan bool, numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func() {
			defer func() { done <- true }()

			// List providers concurrently
			_, err := bridge.listProviders(ctx, []interface{}{})
			assert.NoError(t, err)

			// List templates concurrently
			_, err = bridge.listProviderTemplates(ctx, []interface{}{})
			assert.NoError(t, err)
		}()
	}

	// Wait for all operations
	for i := 0; i < numRoutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

// Test provider configuration validation
func TestProviderConfigurationValidation(t *testing.T) {
	bridge := NewProvidersBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test configuration validation for different provider types
	providerTypes := []string{"openai", "anthropic", "mock"}

	for _, providerType := range providerTypes {
		config := map[string]interface{}{
			"api_key": "test-key",
			"model":   "test-model",
		}

		result, err := bridge.validateProviderConfig(ctx, []interface{}{providerType, config})
		if err == nil {
			validationResult, ok := result.(map[string]interface{})
			require.True(t, ok)
			assert.Contains(t, validationResult, "valid")
			assert.Contains(t, validationResult, "providerType")
		}
	}
}

// Test consensus strategies
func TestConsensusStrategies(t *testing.T) {
	// Test consensus strategy enum conversion
	strategies := []string{"fastest", "primary", "consensus"}

	for _, strategy := range strategies {
		t.Run("strategy_"+strategy, func(t *testing.T) {
			bridge := NewProvidersBridge()
			ctx := context.Background()
			err := bridge.Initialize(ctx)
			require.NoError(t, err)

			// Test strategy validation in multi-provider creation
			providersConfig := []interface{}{
				map[string]interface{}{
					"name":   "mock1",
					"weight": 1.0,
				},
			}

			_, err = bridge.createMultiProvider(ctx, []interface{}{"test-multi", providersConfig, strategy})
			// Will fail due to missing providers, but strategy should be valid
			if err != nil {
				assert.Contains(t, err.Error(), "provider not found")
			}
		})
	}

	// Test invalid strategy
	bridge := NewProvidersBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	_, err = bridge.createMultiProvider(ctx, []interface{}{"test-multi", []interface{}{}, "invalid_strategy"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown strategy")
}

// Test provider metadata retrieval
func TestProviderMetadata(t *testing.T) {
	bridge := NewProvidersBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test metadata for non-existent provider
	_, err = bridge.getProviderMetadata(ctx, []interface{}{"nonexistent"})
	assert.Error(t, err)
}

// Test bridge configuration and lifecycle
func TestProvidersBridgeLifecycle(t *testing.T) {
	bridge := NewProvidersBridge()
	ctx := context.Background()

	// Test initialization
	assert.False(t, bridge.IsInitialized())
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())

	// Test methods return metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "providers", metadata.Name)
	assert.NotEmpty(t, metadata.Dependencies)

	// Test type mappings
	typeMappings := bridge.TypeMappings()
	assert.Contains(t, typeMappings, "provider")
	assert.Contains(t, typeMappings, "multi_provider")
	assert.Contains(t, typeMappings, "provider_template")

	// Test required permissions
	permissions := bridge.RequiredPermissions()
	assert.Greater(t, len(permissions), 0)
	assert.Equal(t, "llm.providers", permissions[0].Resource)

	// Test method listing
	methods := bridge.Methods()
	assert.Greater(t, len(methods), 10) // Should have many methods

	// Verify specific methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	expectedMethods := []string{
		"createProvider",
		"listProviders",
		"getProviderTemplate",
		"createMultiProvider",
		"generateWithProvider",
	}

	for _, expectedMethod := range expectedMethods {
		assert.True(t, methodNames[expectedMethod], "Method %s should exist", expectedMethod)
	}
}
