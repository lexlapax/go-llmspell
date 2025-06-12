// ABOUTME: Test suite for the model info bridge that provides LLM model discovery and metadata
// ABOUTME: Tests model inventory, provider-specific fetchers, caching, and service interfaces

package bridge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock model info provider
type mockModelInfoProvider struct {
	name      string
	models    []ModelInfo
	listError error
	getError  error
	callCount int
	mu        sync.Mutex
}

func (p *mockModelInfoProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.callCount++

	if p.listError != nil {
		return nil, p.listError
	}
	return p.models, nil
}

func (p *mockModelInfoProvider) GetModel(ctx context.Context, modelID string) (*ModelInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.getError != nil {
		return nil, p.getError
	}

	for _, model := range p.models {
		if model.ID == modelID {
			return &model, nil
		}
	}
	return nil, errors.New("model not found")
}

func (p *mockModelInfoProvider) GetCallCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.callCount
}

func TestNewModelInfoBridge(t *testing.T) {
	bridge := NewModelInfoBridge()
	require.NotNil(t, bridge)

	// Check metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "Model Info Bridge", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Equal(t, "Provides LLM model discovery and metadata", metadata.Description)
}

func TestModelInfoProviderManagement(t *testing.T) {
	t.Run("Register Provider", func(t *testing.T) {
		bridge := NewModelInfoBridge()
		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", Name: "GPT-4", Provider: "openai"},
				{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Provider: "openai"},
			},
		}

		err := bridge.RegisterProvider("openai", provider)
		assert.NoError(t, err)

		providers := bridge.ListProviders()
		assert.Contains(t, providers, "openai")
	})

	t.Run("Register Multiple Providers", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider1 := &mockModelInfoProvider{name: "openai"}
		provider2 := &mockModelInfoProvider{name: "anthropic"}

		_ = bridge.RegisterProvider("openai", provider1)
		_ = bridge.RegisterProvider("anthropic", provider2)

		providers := bridge.ListProviders()
		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "openai")
		assert.Contains(t, providers, "anthropic")
	})

	t.Run("Duplicate Provider Registration", func(t *testing.T) {
		bridge := NewModelInfoBridge()
		provider := &mockModelInfoProvider{name: "openai"}

		err := bridge.RegisterProvider("openai", provider)
		assert.NoError(t, err)

		err = bridge.RegisterProvider("openai", provider)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestModelInventory(t *testing.T) {
	t.Run("List All Models", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		// Register multiple providers
		openaiProvider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", Name: "GPT-4", Provider: "openai", MaxTokens: 8192},
				{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Provider: "openai", MaxTokens: 4096},
			},
		}

		anthropicProvider := &mockModelInfoProvider{
			name: "anthropic",
			models: []ModelInfo{
				{ID: "claude-3-opus", Name: "Claude 3 Opus", Provider: "anthropic", MaxTokens: 200000},
				{ID: "claude-3-sonnet", Name: "Claude 3 Sonnet", Provider: "anthropic", MaxTokens: 200000},
			},
		}

		_ = bridge.RegisterProvider("openai", openaiProvider)
		_ = bridge.RegisterProvider("anthropic", anthropicProvider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		models, err := bridge.ListAllModels(ctx)
		assert.NoError(t, err)
		assert.Len(t, models, 4)

		// Verify all models are present
		modelIDs := make([]string, len(models))
		for i, model := range models {
			modelIDs[i] = model.ID
		}
		assert.Contains(t, modelIDs, "gpt-4")
		assert.Contains(t, modelIDs, "gpt-3.5-turbo")
		assert.Contains(t, modelIDs, "claude-3-opus")
		assert.Contains(t, modelIDs, "claude-3-sonnet")
	})

	t.Run("List Models by Provider", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", Name: "GPT-4", Provider: "openai"},
				{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Provider: "openai"},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		models, err := bridge.ListModelsByProvider(ctx, "openai")
		assert.NoError(t, err)
		assert.Len(t, models, 2)

		// Non-existent provider
		models, err = bridge.ListModelsByProvider(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, models)
	})

	t.Run("Get Specific Model", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{
					ID:           "gpt-4",
					Name:         "GPT-4",
					Provider:     "openai",
					MaxTokens:    8192,
					Description:  "Most capable GPT-4 model",
					Capabilities: []string{"chat", "completion"},
				},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		model, err := bridge.GetModel(ctx, "openai", "gpt-4")
		assert.NoError(t, err)
		assert.NotNil(t, model)
		assert.Equal(t, "gpt-4", model.ID)
		assert.Equal(t, "GPT-4", model.Name)
		assert.Equal(t, 8192, model.MaxTokens)

		// Non-existent model
		model, err = bridge.GetModel(ctx, "openai", "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, model)
	})
}

func TestModelInfoCaching(t *testing.T) {
	t.Run("Cache Model List", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", Name: "GPT-4", Provider: "openai"},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// First call - should hit provider
		models1, err := bridge.ListModelsByProvider(ctx, "openai")
		assert.NoError(t, err)
		assert.Len(t, models1, 1)
		assert.Equal(t, 1, provider.GetCallCount())

		// Second call - should use cache
		models2, err := bridge.ListModelsByProvider(ctx, "openai")
		assert.NoError(t, err)
		assert.Len(t, models2, 1)
		assert.Equal(t, 1, provider.GetCallCount()) // No additional calls

		// Models should be identical
		assert.Equal(t, models1, models2)
	})

	t.Run("Cache Expiration", func(t *testing.T) {
		bridge := NewModelInfoBridge()
		bridge.SetCacheTTL(100 * time.Millisecond) // Short TTL for testing

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", Name: "GPT-4", Provider: "openai"},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// First call
		_, err := bridge.ListModelsByProvider(ctx, "openai")
		assert.NoError(t, err)
		assert.Equal(t, 1, provider.GetCallCount())

		// Wait for cache to expire
		time.Sleep(150 * time.Millisecond)

		// Second call - should hit provider again
		_, err = bridge.ListModelsByProvider(ctx, "openai")
		assert.NoError(t, err)
		assert.Equal(t, 2, provider.GetCallCount())
	})

	t.Run("Clear Cache", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", Name: "GPT-4", Provider: "openai"},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// First call
		_, err := bridge.ListModelsByProvider(ctx, "openai")
		assert.NoError(t, err)
		assert.Equal(t, 1, provider.GetCallCount())

		// Clear cache
		bridge.ClearCache()

		// Second call - should hit provider again
		_, err = bridge.ListModelsByProvider(ctx, "openai")
		assert.NoError(t, err)
		assert.Equal(t, 2, provider.GetCallCount())
	})
}

func TestModelFiltering(t *testing.T) {
	t.Run("Filter by Capabilities", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", Capabilities: []string{"chat", "completion", "vision"}},
				{ID: "gpt-3.5-turbo", Capabilities: []string{"chat", "completion"}},
				{ID: "text-embedding-ada", Capabilities: []string{"embedding"}},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// Filter for chat capability
		models, err := bridge.FilterModels(ctx, ModelFilter{
			Capabilities: []string{"chat"},
		})
		assert.NoError(t, err)
		assert.Len(t, models, 2)

		// Filter for vision capability
		models, err = bridge.FilterModels(ctx, ModelFilter{
			Capabilities: []string{"vision"},
		})
		assert.NoError(t, err)
		assert.Len(t, models, 1)
		assert.Equal(t, "gpt-4", models[0].ID)
	})

	t.Run("Filter by Token Limit", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", MaxTokens: 8192},
				{ID: "gpt-3.5-turbo", MaxTokens: 4096},
				{ID: "gpt-4-32k", MaxTokens: 32768},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// Filter for models with at least 8k tokens
		models, err := bridge.FilterModels(ctx, ModelFilter{
			MinTokens: 8000,
		})
		assert.NoError(t, err)
		assert.Len(t, models, 2)

		// Filter for models with at least 32k tokens
		models, err = bridge.FilterModels(ctx, ModelFilter{
			MinTokens: 32000,
		})
		assert.NoError(t, err)
		assert.Len(t, models, 1)
		assert.Equal(t, "gpt-4-32k", models[0].ID)
	})
}

func TestModelInfoBridgeEngineIntegration(t *testing.T) {
	t.Run("Register With Engine", func(t *testing.T) {
		bridge := NewModelInfoBridge()
		mockEngine := &mockScriptEngine{}

		err := bridge.RegisterWithEngine(mockEngine)
		assert.NoError(t, err)
		assert.Len(t, mockEngine.bridges, 1)
		assert.Equal(t, "modelinfo", mockEngine.bridges[0].GetID())
	})

	t.Run("Bridge Methods", func(t *testing.T) {
		bridge := NewModelInfoBridge()
		methods := bridge.Methods()

		expectedMethods := []string{
			"list_all_models",
			"list_models_by_provider",
			"get_model",
			"filter_models",
			"list_providers",
			"clear_cache",
			"set_cache_ttl",
		}

		assert.Len(t, methods, len(expectedMethods))

		methodNames := make(map[string]bool)
		for _, method := range methods {
			methodNames[method.Name] = true
		}

		for _, expected := range expectedMethods {
			assert.True(t, methodNames[expected], "Expected method %s not found", expected)
		}
	})

	t.Run("Type Mappings", func(t *testing.T) {
		bridge := NewModelInfoBridge()
		mappings := bridge.TypeMappings()

		// Should have mappings for model info structures
		assert.Contains(t, mappings, "ModelInfo")
		assert.Contains(t, mappings, "ModelFilter")
		assert.Contains(t, mappings, "[]ModelInfo")
	})
}

func TestConcurrentModelInfoAccess(t *testing.T) {
	t.Run("Concurrent List Operations", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", Name: "GPT-4"},
				{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo"},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// Run concurrent list operations
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				models, err := bridge.ListAllModels(ctx)
				if err != nil {
					errors <- err
					return
				}
				if len(models) != 2 {
					errors <- fmt.Errorf("unexpected model count")
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
		}
	})

	t.Run("Concurrent Provider Registration", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		var wg sync.WaitGroup
		errors := make(chan error, 5)

		// Try to register different providers concurrently
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				provider := &mockModelInfoProvider{
					name: fmt.Sprintf("provider%d", id),
				}
				err := bridge.RegisterProvider(fmt.Sprintf("provider%d", id), provider)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// All registrations should succeed
		for err := range errors {
			t.Errorf("Registration error: %v", err)
		}

		// Verify all providers registered
		providers := bridge.ListProviders()
		assert.Len(t, providers, 5)
	})
}

func TestModelInfoErrorHandling(t *testing.T) {
	t.Run("Provider List Error", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name:      "openai",
			listError: errors.New("API error"),
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		models, err := bridge.ListModelsByProvider(ctx, "openai")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
		assert.Nil(t, models)
	})

	t.Run("Provider Get Error", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name:     "openai",
			getError: errors.New("model fetch error"),
			models: []ModelInfo{
				{ID: "gpt-4", Name: "GPT-4"},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		model, err := bridge.GetModel(ctx, "openai", "gpt-4")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model fetch error")
		assert.Nil(t, model)
	})
}

func TestModelInfoEdgeCases(t *testing.T) {
	t.Run("Empty Provider List", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		models, err := bridge.ListAllModels(ctx)
		assert.NoError(t, err)
		assert.Empty(t, models)
	})

	t.Run("Invalid Filter", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		provider := &mockModelInfoProvider{
			name: "openai",
			models: []ModelInfo{
				{ID: "gpt-4", MaxTokens: 8192},
			},
		}

		_ = bridge.RegisterProvider("openai", provider)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// Filter with impossible criteria
		models, err := bridge.FilterModels(ctx, ModelFilter{
			MinTokens: 1000000, // No model has this many tokens
		})
		assert.NoError(t, err)
		assert.Empty(t, models)
	})

	t.Run("Nil Context", func(t *testing.T) {
		bridge := NewModelInfoBridge()

		// Should handle nil context gracefully
		models, err := bridge.ListAllModels(nil) //nolint:staticcheck // Testing nil context handling
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context")
		assert.Nil(t, models)
	})
}

// Benchmark tests
func BenchmarkModelInfoListAll(b *testing.B) {
	bridge := NewModelInfoBridge()

	// Create providers with many models
	for i := 0; i < 5; i++ {
		models := make([]ModelInfo, 100)
		for j := 0; j < 100; j++ {
			models[j] = ModelInfo{
				ID:       fmt.Sprintf("model-%d-%d", i, j),
				Name:     fmt.Sprintf("Model %d-%d", i, j),
				Provider: fmt.Sprintf("provider%d", i),
			}
		}

		provider := &mockModelInfoProvider{
			name:   fmt.Sprintf("provider%d", i),
			models: models,
		}
		_ = bridge.RegisterProvider(fmt.Sprintf("provider%d", i), provider)
	}

	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bridge.ListAllModels(ctx)
	}
}

func BenchmarkModelInfoFilter(b *testing.B) {
	bridge := NewModelInfoBridge()

	// Create provider with models
	models := make([]ModelInfo, 1000)
	for i := 0; i < 1000; i++ {
		models[i] = ModelInfo{
			ID:           fmt.Sprintf("model-%d", i),
			Name:         fmt.Sprintf("Model %d", i),
			Provider:     "test",
			MaxTokens:    (i + 1) * 1000,
			Capabilities: []string{"chat", "completion"},
		}
	}

	provider := &mockModelInfoProvider{
		name:   "test",
		models: models,
	}
	_ = bridge.RegisterProvider("test", provider)

	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	filter := ModelFilter{
		MinTokens:    5000,
		Capabilities: []string{"chat"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bridge.FilterModels(ctx, filter)
	}
}
