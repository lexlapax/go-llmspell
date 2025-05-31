// ABOUTME: Tests for the Agent registry implementation
// ABOUTME: Ensures thread-safe agent management following established patterns

package agents

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRegistry(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		registry := NewRegistry()
		require.NotNil(t, registry)

		// Test factory registration
		factory := func(config Config) (Agent, error) {
			return NewMockAgent(config.Name), nil
		}

		err := registry.Register("mock", factory)
		require.NoError(t, err)

		// Test duplicate registration
		err = registry.Register("mock", factory)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")

		// Test listing
		factories := registry.List()
		assert.Contains(t, factories, "mock")
	})

	t.Run("agent creation", func(t *testing.T) {
		registry := NewRegistry()

		// Register factory
		factory := func(config Config) (Agent, error) {
			if config.Name == "error-agent" {
				return nil, errors.New("failed to create agent")
			}
			agent := NewMockAgent(config.Name)
			agent.SetSystemPrompt(config.SystemPrompt)
			return agent, nil
		}

		err := registry.Register("mock", factory)
		require.NoError(t, err)

		// Test successful creation
		config := Config{
			Name:         "test-agent",
			SystemPrompt: "Test prompt",
			Provider:     "mock",
			Model:        "mock-model",
		}

		agent, err := registry.Create(config)
		require.NoError(t, err)
		require.NotNil(t, agent)
		assert.Equal(t, "test-agent", agent.Name())
		assert.Equal(t, "Test prompt", agent.GetSystemPrompt())

		// Test creation with unregistered factory
		config.Provider = "unknown"
		agent, err = registry.Create(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, agent)

		// Test creation with factory error
		config.Name = "error-agent"
		config.Provider = "mock"
		agent, err = registry.Create(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create agent")
		assert.Nil(t, agent)
	})

	t.Run("agent management", func(t *testing.T) {
		registry := NewRegistry()

		// Register factory
		factory := func(config Config) (Agent, error) {
			return NewMockAgent(config.Name), nil
		}
		err := registry.Register("mock", factory)
		require.NoError(t, err)

		// Create and store agent
		config := Config{
			Name:     "managed-agent",
			Provider: "mock",
			Model:    "mock-model",
		}

		agent, err := registry.Create(config)
		require.NoError(t, err)

		// Test Get
		retrieved, err := registry.Get("managed-agent")
		require.NoError(t, err)
		assert.Equal(t, agent, retrieved)

		// Test Get non-existent
		_, err = registry.Get("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test Remove
		err = registry.Remove("managed-agent")
		require.NoError(t, err)

		// Verify removed
		_, err = registry.Get("managed-agent")
		assert.Error(t, err)

		// Test Remove non-existent
		err = registry.Remove("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("concurrent access", func(t *testing.T) {
		registry := NewRegistry()

		// Register factory
		factory := func(config Config) (Agent, error) {
			return NewMockAgent(config.Name), nil
		}
		err := registry.Register("mock", factory)
		require.NoError(t, err)

		// Test concurrent creates
		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				config := Config{
					Name:     "concurrent-agent",
					Provider: "mock",
					Model:    "mock-model",
				}

				agent, err := registry.Create(config)
				if err != nil {
					errChan <- err
					return
				}

				// Try to get the agent
				_, err = registry.Get(agent.Name())
				if err != nil {
					errChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		// Check for any errors
		for err := range errChan {
			t.Errorf("Concurrent access error: %v", err)
		}
	})

	t.Run("global registry", func(t *testing.T) {
		// Test global functions
		factory := func(config Config) (Agent, error) {
			return NewMockAgent(config.Name), nil
		}

		// Clear any existing registrations
		globalRegistry = NewRegistry()

		err := RegisterAgentFactory("global-mock", factory)
		require.NoError(t, err)

		config := Config{
			Name:     "global-agent",
			Provider: "global-mock",
			Model:    "mock-model",
		}

		agent, err := CreateAgent(config)
		require.NoError(t, err)
		assert.Equal(t, "global-agent", agent.Name())

		retrieved, err := GetAgent("global-agent")
		require.NoError(t, err)
		assert.Equal(t, agent, retrieved)

		factories := ListAgentFactories()
		assert.Contains(t, factories, "global-mock")

		err = RemoveAgent("global-agent")
		require.NoError(t, err)

		_, err = GetAgent("global-agent")
		assert.Error(t, err)
	})
}

func TestDefaultRegistry(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		// Test that default registry is properly initialized
		registry := DefaultRegistry()
		require.NotNil(t, registry)

		// Should be the same instance on multiple calls
		registry2 := DefaultRegistry()
		assert.Equal(t, registry, registry2)
	})
}
