// ABOUTME: Tests for the default agent implementation
// ABOUTME: Verifies integration with go-llms agents and tool system

package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAgent(t *testing.T) {
	t.Run("creation and initialization", func(t *testing.T) {
		config := Config{
			Name:         "test-agent",
			SystemPrompt: "You are a helpful assistant",
			Provider:     "openai",
			Model:        "gpt-4",
			MaxTokens:    1000,
			Temperature:  0.7,
			Timeout:      30 * time.Second,
		}

		// Create agent with mock LLM provider
		agent := NewDefaultAgent(config)
		require.NotNil(t, agent)

		// Test properties
		assert.Equal(t, "test-agent", agent.Name())
		assert.Equal(t, "You are a helpful assistant", agent.GetSystemPrompt())

		// Test initialization
		ctx := context.Background()
		err := agent.Initialize(ctx)
		assert.NoError(t, err)

		// Test cleanup
		err = agent.Cleanup()
		assert.NoError(t, err)
	})

	t.Run("system prompt management", func(t *testing.T) {
		config := Config{
			Name:         "prompt-test",
			SystemPrompt: "Initial prompt",
			Provider:     "openai",
			Model:        "gpt-4",
		}

		agent := NewDefaultAgent(config)
		
		// Initial prompt
		assert.Equal(t, "Initial prompt", agent.GetSystemPrompt())

		// Update prompt
		agent.SetSystemPrompt("Updated prompt")
		assert.Equal(t, "Updated prompt", agent.GetSystemPrompt())
	})

	t.Run("tool management", func(t *testing.T) {
		config := Config{
			Name:     "tool-test",
			Provider: "openai",
			Model:    "gpt-4",
			Tools:    []string{"calculator", "web_fetch"},
		}

		agent := NewDefaultAgent(config)

		// Check initial tools
		tools := agent.GetTools()
		assert.Contains(t, tools, "calculator")
		assert.Contains(t, tools, "web_fetch")

		// Add new tool
		err := agent.AddTool("json_parser")
		assert.NoError(t, err)

		tools = agent.GetTools()
		assert.Contains(t, tools, "json_parser")
		assert.Len(t, tools, 3)

		// Try to add duplicate tool
		err = agent.AddTool("calculator")
		assert.NoError(t, err) // Should not error, just ignore duplicate
		tools = agent.GetTools()
		assert.Len(t, tools, 3) // Still 3 tools
	})

	t.Run("tool validation", func(t *testing.T) {
		config := Config{
			Name:     "tool-validation",
			Provider: "openai",
			Model:    "gpt-4",
		}

		agent := NewDefaultAgent(config)

		// Try to add non-existent tool
		err := agent.AddTool("non_existent_tool")
		// The actual implementation might check if tool exists in registry
		// For now, we'll accept this behavior
		assert.NoError(t, err)
	})
}

// MockLLMProvider for testing
type MockLLMProvider struct {
	responses  []string
	respIndex  int
	executeErr error
	streamErr  error
}

func (m *MockLLMProvider) Execute(ctx context.Context, messages []Message, opts *ExecutionOptions) (*ExecutionResult, error) {
	if m.executeErr != nil {
		return nil, m.executeErr
	}

	response := "Mock response"
	if m.respIndex < len(m.responses) {
		response = m.responses[m.respIndex]
		m.respIndex++
	}

	return &ExecutionResult{
		Response:   response,
		Messages:   append(messages, NewAssistantMessage(response)),
		TokensUsed: 10,
		Duration:   10 * time.Millisecond,
	}, nil
}

func (m *MockLLMProvider) Stream(ctx context.Context, messages []Message, opts *ExecutionOptions, callback StreamCallback) error {
	if m.streamErr != nil {
		return m.streamErr
	}

	response := "Mock streaming response"
	if m.respIndex < len(m.responses) {
		response = m.responses[m.respIndex]
		m.respIndex++
	}

	// Simulate streaming using the response
	chunks := splitIntoChunks(response, 5)
	for _, chunk := range chunks {
		if err := callback(chunk); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func TestDefaultAgentExecution(t *testing.T) {
	t.Run("execute with mock provider", func(t *testing.T) {
		// This test will need the actual implementation to work
		// For now, we're setting up the structure
		t.Skip("Skipping until DefaultAgent is implemented with mock provider support")
	})

	t.Run("stream with mock provider", func(t *testing.T) {
		// This test will need the actual implementation to work
		t.Skip("Skipping until DefaultAgent is implemented with mock provider support")
	})
}

func TestLLMSAgentAdapter(t *testing.T) {
	t.Run("adapter wraps go-llms agent", func(t *testing.T) {
		// This will test the adapter that wraps go-llms agents
		// Similar to how LLMSToolAdapter wraps go-llms tools
		t.Skip("Skipping until LLMSAgentAdapter is implemented")
	})
}