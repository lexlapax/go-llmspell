// ABOUTME: Tests for the agent bridge that exposes agents to scripts
// ABOUTME: Verifies agent creation, execution, and management from scripts

package bridge

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentBridge(t *testing.T) {
	t.Run("creation and initialization", func(t *testing.T) {
		ctx := context.Background()
		bridge, err := NewAgentBridge(ctx)
		require.NoError(t, err)
		require.NotNil(t, bridge)

		// Check interface implementation
		var _ AgentBridge = bridge
	})

	t.Run("agent creation", func(t *testing.T) {
		ctx := context.Background()
		bridge, err := NewAgentBridge(ctx)
		require.NoError(t, err)

		// Register a mock agent factory
		factory := func(config agents.Config) (agents.Agent, error) {
			if config.Name == "error-agent" {
				return nil, errors.New("failed to create agent")
			}
			agent := agents.NewMockAgent(config.Name)
			agent.SetSystemPrompt(config.SystemPrompt)
			for _, tool := range config.Tools {
				err := agent.AddTool(tool)
				if err != nil {
					return nil, err
				}
			}
			return agent, nil
		}

		err = agents.RegisterAgentFactory("mock", factory)
		require.NoError(t, err)

		// Test successful creation
		config := map[string]interface{}{
			"name":         "test-agent",
			"provider":     "mock",
			"model":        "mock-model",
			"systemPrompt": "You are a test assistant",
		}

		agentName, err := bridge.Create(config)
		require.NoError(t, err)
		assert.Equal(t, "test-agent", agentName)

		// Test creation with missing fields
		badConfig := map[string]interface{}{
			"provider": "mock",
			"model":    "mock-model",
		}

		_, err = bridge.Create(badConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")

		// Test creation with factory error
		errorConfig := map[string]interface{}{
			"name":     "error-agent",
			"provider": "mock",
			"model":    "mock-model",
		}

		_, err = bridge.Create(errorConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create agent")
	})

	t.Run("agent execution", func(t *testing.T) {
		ctx := context.Background()
		bridge, err := NewAgentBridge(ctx)
		require.NoError(t, err)

		// Create a test agent
		config := map[string]interface{}{
			"name":     "exec-agent",
			"provider": "mock",
			"model":    "mock-model",
		}

		agentName, err := bridge.Create(config)
		require.NoError(t, err)

		// Execute the agent
		result, err := bridge.Execute(agentName, "Test input", nil)
		require.NoError(t, err)
		assert.NotEmpty(t, result)

		// Test execution with non-existent agent
		_, err = bridge.Execute("non-existent", "Test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test execution with options
		opts := map[string]interface{}{
			"maxTokens":   100,
			"temperature": 0.7,
		}

		result, err = bridge.Execute(agentName, "Test with options", opts)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("agent listing", func(t *testing.T) {
		ctx := context.Background()
		bridge, err := NewAgentBridge(ctx)
		require.NoError(t, err)

		// Create multiple agents
		createdAgents := []string{}
		for i := 0; i < 3; i++ {
			config := map[string]interface{}{
				"name":     fmt.Sprintf("list-agent-%d", i),
				"provider": "mock",
				"model":    "mock-model",
			}
			name, err := bridge.Create(config)
			require.NoError(t, err)
			createdAgents = append(createdAgents, name)
		}

		// Verify we can get info for each created agent
		for _, agentName := range createdAgents {
			info, err := bridge.GetInfo(agentName)
			require.NoError(t, err)
			assert.Equal(t, agentName, info["name"])
		}

		// TODO: When List() is implemented to return created agents,
		// update this test to verify the list contains our agents
		agentList := bridge.List()
		assert.NotNil(t, agentList)
	})

	t.Run("agent info", func(t *testing.T) {
		ctx := context.Background()
		bridge, err := NewAgentBridge(ctx)
		require.NoError(t, err)

		// Create an agent with tools
		config := map[string]interface{}{
			"name":         "info-agent",
			"provider":     "mock",
			"model":        "mock-model",
			"systemPrompt": "Test prompt",
			"tools":        []string{"calculator", "web_fetch"},
		}

		agentName, err := bridge.Create(config)
		require.NoError(t, err)

		// Get agent info
		info, err := bridge.GetInfo(agentName)
		require.NoError(t, err)
		assert.Equal(t, "info-agent", info["name"])
		assert.Equal(t, "Test prompt", info["systemPrompt"])

		tools, ok := info["tools"].([]string)
		assert.True(t, ok)
		assert.Contains(t, tools, "calculator")
		assert.Contains(t, tools, "web_fetch")

		// Test get info for non-existent agent
		_, err = bridge.GetInfo("non-existent")
		assert.Error(t, err)
	})

	t.Run("streaming execution", func(t *testing.T) {
		ctx := context.Background()
		bridge, err := NewAgentBridge(ctx)
		require.NoError(t, err)

		// Create an agent
		config := map[string]interface{}{
			"name":     "stream-agent",
			"provider": "mock",
			"model":    "mock-model",
		}

		agentName, err := bridge.Create(config)
		require.NoError(t, err)

		// Stream execution
		chunks := []string{}
		err = bridge.Stream(agentName, "Test streaming", nil, func(chunk string) error {
			chunks = append(chunks, chunk)
			return nil
		})
		require.NoError(t, err)
		assert.NotEmpty(t, chunks)

		// Verify we got multiple chunks
		assert.Greater(t, len(chunks), 1)

		// Test streaming with error callback
		err = bridge.Stream(agentName, "Test", nil, func(chunk string) error {
			return errors.New("callback error")
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "callback error")
	})

	t.Run("agent removal", func(t *testing.T) {
		ctx := context.Background()
		bridge, err := NewAgentBridge(ctx)
		require.NoError(t, err)

		// Create an agent
		config := map[string]interface{}{
			"name":     "remove-agent",
			"provider": "mock",
			"model":    "mock-model",
		}

		agentName, err := bridge.Create(config)
		require.NoError(t, err)

		// Verify it exists
		_, err = bridge.GetInfo(agentName)
		require.NoError(t, err)

		// Remove the agent
		err = bridge.Remove(agentName)
		require.NoError(t, err)

		// Verify it's gone
		_, err = bridge.GetInfo(agentName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test removing non-existent agent
		err = bridge.Remove("non-existent")
		assert.Error(t, err)
	})
}

// Using MockAgent from agents package
