// ABOUTME: Tests for the agent registry that manages thread-safe agent registration and discovery
// ABOUTME: Ensures capability-based discovery, lifecycle management, and templating work correctly

package agent

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistry_Creation tests registry creation and initialization
func TestRegistry_Creation(t *testing.T) {
	t.Run("create new registry", func(t *testing.T) {
		registry := NewRegistry()
		require.NotNil(t, registry)

		// Check initial state
		agents := registry.List()
		assert.Empty(t, agents)
	})

	t.Run("global registry", func(t *testing.T) {
		// Get global registry
		registry := GlobalRegistry()
		require.NotNil(t, registry)

		// Should be the same instance
		registry2 := GlobalRegistry()
		assert.Same(t, registry, registry2)
	})
}

// TestRegistry_Registration tests agent registration functionality
func TestRegistry_Registration(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	t.Run("register agent", func(t *testing.T) {
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Version: "1.0.0",
		})
		require.NoError(t, err)

		// Register agent
		err = registry.Register(ctx, agent)
		assert.NoError(t, err)

		// Check agent is registered
		retrieved, err := registry.Get("test-agent")
		assert.NoError(t, err)
		assert.Equal(t, agent, retrieved)
	})

	t.Run("register duplicate agent", func(t *testing.T) {
		agent1, err := NewBaseAgent(BaseAgentConfig{
			ID:      "duplicate",
			Name:    "Agent 1",
			Version: "1.0.0",
		})
		require.NoError(t, err)

		agent2, err := NewBaseAgent(BaseAgentConfig{
			ID:      "duplicate",
			Name:    "Agent 2",
			Version: "1.0.0",
		})
		require.NoError(t, err)

		// Register first agent
		err = registry.Register(ctx, agent1)
		assert.NoError(t, err)

		// Try to register duplicate
		err = registry.Register(ctx, agent2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("register with force", func(t *testing.T) {
		agent1, err := NewBaseAgent(BaseAgentConfig{
			ID:      "force-test",
			Name:    "Agent 1",
			Version: "1.0.0",
		})
		require.NoError(t, err)

		agent2, err := NewBaseAgent(BaseAgentConfig{
			ID:      "force-test",
			Name:    "Agent 2",
			Version: "2.0.0",
		})
		require.NoError(t, err)

		// Register first agent
		err = registry.Register(ctx, agent1)
		assert.NoError(t, err)

		// Force register second agent
		err = registry.RegisterWithOptions(ctx, agent2, RegisterOptions{Force: true})
		assert.NoError(t, err)

		// Check second agent replaced first
		retrieved, err := registry.Get("force-test")
		assert.NoError(t, err)
		assert.Equal(t, "2.0.0", retrieved.Version())
	})

	t.Run("unregister agent", func(t *testing.T) {
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:      "unregister-test",
			Name:    "Test Agent",
			Version: "1.0.0",
		})
		require.NoError(t, err)

		// Register and unregister
		err = registry.Register(ctx, agent)
		assert.NoError(t, err)

		err = registry.Unregister(ctx, "unregister-test")
		assert.NoError(t, err)

		// Check agent is gone
		_, err = registry.Get("unregister-test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestRegistry_Discovery tests capability-based agent discovery
func TestRegistry_Discovery(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	// Register agents with different capabilities
	agents := []struct {
		id           string
		capabilities map[string]interface{}
	}{
		{
			id: "agent1",
			capabilities: map[string]interface{}{
				"streaming": true,
				"tools":     []string{"file_read", "file_write"},
				"language":  "lua",
			},
		},
		{
			id: "agent2",
			capabilities: map[string]interface{}{
				"streaming": false,
				"tools":     []string{"web_fetch"},
				"language":  "javascript",
			},
		},
		{
			id: "agent3",
			capabilities: map[string]interface{}{
				"streaming": true,
				"tools":     []string{"file_read", "web_fetch"},
				"language":  "lua",
			},
		},
	}

	for _, agentData := range agents {
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:           agentData.id,
			Name:         "Test Agent",
			Version:      "1.0.0",
			Capabilities: agentData.capabilities,
		})
		require.NoError(t, err)
		require.NoError(t, registry.Register(ctx, agent))
	}

	t.Run("find by single capability", func(t *testing.T) {
		// Find agents with streaming
		found := registry.FindByCapability("streaming", true)
		assert.Len(t, found, 2)
		ids := []string{found[0].ID(), found[1].ID()}
		assert.Contains(t, ids, "agent1")
		assert.Contains(t, ids, "agent3")
	})

	t.Run("find by multiple capabilities", func(t *testing.T) {
		// Find Lua agents with streaming
		filter := CapabilityFilter{
			Required: map[string]interface{}{
				"streaming": true,
				"language":  "lua",
			},
		}
		found := registry.FindByCapabilities(filter)
		assert.Len(t, found, 2)
	})

	t.Run("find with any match", func(t *testing.T) {
		// Find agents with either file_read or web_fetch tools
		filter := CapabilityFilter{
			Any: map[string]interface{}{
				"tools": []string{"file_read"},
			},
		}
		found := registry.FindByCapabilities(filter)
		assert.Len(t, found, 2) // agent1 and agent3
	})

	t.Run("find with excluded capabilities", func(t *testing.T) {
		// Find non-streaming agents
		filter := CapabilityFilter{
			Excluded: map[string]interface{}{
				"streaming": true,
			},
		}
		found := registry.FindByCapabilities(filter)
		assert.Len(t, found, 1)
		assert.Equal(t, "agent2", found[0].ID())
	})

	t.Run("complex filter", func(t *testing.T) {
		// Find Lua agents with file_read but not web_fetch
		filter := CapabilityFilter{
			Required: map[string]interface{}{
				"language": "lua",
			},
			Any: map[string]interface{}{
				"tools": []string{"file_read"},
			},
			Excluded: map[string]interface{}{
				"tools": []string{"web_fetch"},
			},
		}
		found := registry.FindByCapabilities(filter)
		assert.Len(t, found, 1)
		assert.Equal(t, "agent1", found[0].ID())
	})
}

// TestRegistry_Lifecycle tests agent lifecycle management
func TestRegistry_Lifecycle(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	t.Run("auto initialize on register", func(t *testing.T) {
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:      "auto-init",
			Name:    "Test Agent",
			Version: "1.0.0",
		})
		require.NoError(t, err)

		// Register with auto-init
		err = registry.RegisterWithOptions(ctx, agent, RegisterOptions{
			AutoInit: true,
		})
		assert.NoError(t, err)

		// Check agent was initialized
		assert.Equal(t, StatusReady, agent.Status())
	})

	t.Run("cleanup on unregister", func(t *testing.T) {
		agent, err := NewBaseAgent(BaseAgentConfig{
			ID:      "auto-cleanup",
			Name:    "Test Agent",
			Version: "1.0.0",
		})
		require.NoError(t, err)

		// Initialize and register
		err = agent.Init(ctx)
		require.NoError(t, err)
		err = registry.Register(ctx, agent)
		require.NoError(t, err)

		// Unregister with cleanup
		err = registry.UnregisterWithCleanup(ctx, "auto-cleanup")
		assert.NoError(t, err)

		// Check agent was cleaned up
		assert.Equal(t, StatusStopped, agent.Status())
	})

	t.Run("batch lifecycle operations", func(t *testing.T) {
		// Register multiple agents
		var agents []Agent
		for i := 0; i < 3; i++ {
			agent, err := NewBaseAgent(BaseAgentConfig{
				ID:      string(rune('a' + i)),
				Name:    "Test Agent",
				Version: "1.0.0",
			})
			require.NoError(t, err)
			err = registry.Register(ctx, agent)
			require.NoError(t, err)
			agents = append(agents, agent)
		}

		// Initialize all
		err := registry.InitAll(ctx)
		assert.NoError(t, err)

		// Check all initialized
		for _, agent := range agents {
			if ea, ok := agent.(ExtendedAgent); ok {
				assert.Equal(t, StatusReady, ea.Status())
			}
		}

		// Cleanup all
		err = registry.CleanupAll(ctx)
		assert.NoError(t, err)

		// Check all cleaned up
		for _, agent := range agents {
			if ea, ok := agent.(ExtendedAgent); ok {
				assert.Equal(t, StatusStopped, ea.Status())
			}
		}
	})
}

// TestRegistry_Templates tests agent templating system
func TestRegistry_Templates(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	t.Run("register template", func(t *testing.T) {
		// Create template
		template := &AgentTemplate{
			ID:          "chat-template",
			Name:        "Chat Agent Template",
			Description: "Template for chat agents",
			Config: BaseAgentConfig{
				Name:        "Chat Agent",
				Description: "A chat agent",
				Version:     "1.0.0",
				Capabilities: map[string]interface{}{
					"streaming": true,
					"chat":      true,
				},
			},
			Factory: func(id string, params map[string]interface{}) (Agent, error) {
				config := BaseAgentConfig{
					ID:           id,
					Name:         "Chat Agent",
					Description:  "A chat agent",
					Version:      "1.0.0",
					Capabilities: map[string]interface{}{"streaming": true, "chat": true},
				}
				// Apply params
				if name, ok := params["name"].(string); ok {
					config.Name = name
				}
				return NewBaseAgent(config)
			},
		}

		// Register template
		err := registry.RegisterTemplate(template)
		assert.NoError(t, err)

		// Get template
		retrieved := registry.GetTemplate("chat-template")
		assert.NotNil(t, retrieved)
		assert.Equal(t, template.ID, retrieved.ID)
	})

	t.Run("create from template", func(t *testing.T) {
		// Register template
		template := &AgentTemplate{
			ID:   "worker-template",
			Name: "Worker Agent Template",
			Config: BaseAgentConfig{
				Name:    "Worker",
				Version: "1.0.0",
			},
			Factory: func(id string, params map[string]interface{}) (Agent, error) {
				config := BaseAgentConfig{
					ID:      id,
					Name:    "Worker",
					Version: "1.0.0",
				}
				if name, ok := params["name"].(string); ok {
					config.Name = name
				}
				return NewBaseAgent(config)
			},
		}
		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		// Create agent from template
		agent, err := registry.CreateFromTemplate(ctx, "worker-template", "worker1", map[string]interface{}{
			"name": "Custom Worker",
		})
		assert.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "worker1", agent.ID())
		assert.Equal(t, "Custom Worker", agent.Name())

		// Agent should be registered
		retrieved, err := registry.Get("worker1")
		assert.NoError(t, err)
		assert.Equal(t, agent, retrieved)
	})

	t.Run("list templates", func(t *testing.T) {
		// Register multiple templates
		for i := 0; i < 3; i++ {
			template := &AgentTemplate{
				ID:   "template" + string(rune('0'+i)),
				Name: "Template " + string(rune('0'+i)),
				Factory: func(id string, params map[string]interface{}) (Agent, error) {
					return NewBaseAgent(BaseAgentConfig{ID: id, Name: "Agent", Version: "1.0.0"})
				},
			}
			err := registry.RegisterTemplate(template)
			require.NoError(t, err)
		}

		// List templates
		templates := registry.ListTemplates()
		assert.GreaterOrEqual(t, len(templates), 3)
	})
}

// TestRegistry_Concurrency tests thread-safe operations
func TestRegistry_Concurrency(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	t.Run("concurrent registration", func(t *testing.T) {
		var wg sync.WaitGroup
		numAgents := 100

		for i := 0; i < numAgents; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				agent, err := NewBaseAgent(BaseAgentConfig{
					ID:      "concurrent-" + string(rune('0'+n)),
					Name:    "Concurrent Agent",
					Version: "1.0.0",
				})
				require.NoError(t, err)
				err = registry.Register(ctx, agent)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()

		// Check all agents registered
		agents := registry.List()
		assert.GreaterOrEqual(t, len(agents), numAgents)
	})

	t.Run("concurrent discovery", func(t *testing.T) {
		// Register agents with capabilities
		for i := 0; i < 10; i++ {
			agent, err := NewBaseAgent(BaseAgentConfig{
				ID:      "discovery-" + string(rune('0'+i)),
				Name:    "Discovery Agent",
				Version: "1.0.0",
				Capabilities: map[string]interface{}{
					"concurrent": true,
					"index":      i,
				},
			})
			require.NoError(t, err)
			err = registry.Register(ctx, agent)
			require.NoError(t, err)
		}

		var wg sync.WaitGroup
		numReaders := 50

		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Perform discovery
				found := registry.FindByCapability("concurrent", true)
				assert.GreaterOrEqual(t, len(found), 10)
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent lifecycle", func(t *testing.T) {
		// Register agents
		var agents []Agent
		for i := 0; i < 5; i++ {
			agent, err := NewBaseAgent(BaseAgentConfig{
				ID:      "lifecycle-" + string(rune('0'+i)),
				Name:    "Lifecycle Agent",
				Version: "1.0.0",
			})
			require.NoError(t, err)
			err = registry.Register(ctx, agent)
			require.NoError(t, err)
			agents = append(agents, agent)
		}

		var wg sync.WaitGroup

		// Concurrent init
		for _, agent := range agents {
			wg.Add(1)
			go func(a Agent) {
				defer wg.Done()
				err := a.Init(ctx)
				assert.NoError(t, err)
			}(agent)
		}

		wg.Wait()

		// Check all initialized
		for _, agent := range agents {
			if ea, ok := agent.(ExtendedAgent); ok {
				assert.Equal(t, StatusReady, ea.Status())
			}
		}
	})
}

// TestRegistry_Stats tests registry statistics and monitoring
func TestRegistry_Stats(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	t.Run("registry stats", func(t *testing.T) {
		// Register agents with different states
		for i := 0; i < 5; i++ {
			agent, err := NewBaseAgent(BaseAgentConfig{
				ID:      "stats-" + string(rune('0'+i)),
				Name:    "Stats Agent",
				Version: "1.0.0",
			})
			require.NoError(t, err)

			if i < 3 {
				err = agent.Init(ctx)
				require.NoError(t, err)
			}

			err = registry.Register(ctx, agent)
			require.NoError(t, err)
		}

		// Get stats
		stats := registry.Stats()
		assert.Equal(t, 5, stats.TotalAgents)
		assert.Equal(t, 3, stats.ReadyAgents)
		assert.Equal(t, 2, stats.CreatedAgents)
		assert.Equal(t, 0, stats.RunningAgents)
		assert.Equal(t, 0, stats.ErrorAgents)
		assert.Equal(t, 0, stats.StoppedAgents)
	})

	t.Run("health check", func(t *testing.T) {
		// Create a fresh registry for this test
		registry := NewRegistry()
		ctx := context.Background()

		// Create healthy and unhealthy agents
		healthy, err := NewBaseAgent(BaseAgentConfig{
			ID:      "healthy",
			Name:    "Healthy Agent",
			Version: "1.0.0",
		})
		require.NoError(t, err)
		err = healthy.Init(ctx)
		require.NoError(t, err)
		err = registry.Register(ctx, healthy)
		require.NoError(t, err)

		unhealthy, err := NewBaseAgent(BaseAgentConfig{
			ID:      "unhealthy",
			Name:    "Unhealthy Agent",
			Version: "1.0.0",
		})
		require.NoError(t, err)
		unhealthy.HandleError(ctx, assert.AnError)
		err = registry.Register(ctx, unhealthy)
		require.NoError(t, err)

		// Perform health check
		report := registry.HealthCheck(ctx)
		assert.False(t, report.Healthy)
		assert.Len(t, report.HealthyAgents, 1)
		assert.Len(t, report.UnhealthyAgents, 1)
		assert.Contains(t, report.Issues, "1 agent(s) in error state")
	})
}

// Helper types are defined in registry.go
