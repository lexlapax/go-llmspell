// ABOUTME: Tests for the Agent interface and related types
// ABOUTME: Ensures the Agent system follows established patterns

package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentInterface(t *testing.T) {
	t.Run("Config validation", func(t *testing.T) {
		tests := []struct {
			name    string
			config  Config
			wantErr bool
		}{
			{
				name: "valid config",
				config: Config{
					Name:         "test-agent",
					SystemPrompt: "You are a helpful assistant",
					Provider:     "openai",
					Model:        "gpt-4",
				},
				wantErr: false,
			},
			{
				name: "missing name",
				config: Config{
					SystemPrompt: "You are a helpful assistant",
					Provider:     "openai",
					Model:        "gpt-4",
				},
				wantErr: true,
			},
			{
				name: "missing provider",
				config: Config{
					Name:         "test-agent",
					SystemPrompt: "You are a helpful assistant",
					Model:        "gpt-4",
				},
				wantErr: true,
			},
			{
				name: "missing model",
				config: Config{
					Name:         "test-agent",
					SystemPrompt: "You are a helpful assistant",
					Provider:     "openai",
				},
				wantErr: true,
			},
			{
				name: "with tools",
				config: Config{
					Name:         "test-agent",
					SystemPrompt: "You are a helpful assistant",
					Provider:     "openai",
					Model:        "gpt-4",
					Tools:        []string{"web_fetch", "calculator"},
				},
				wantErr: false,
			},
			{
				name: "with options",
				config: Config{
					Name:         "test-agent",
					SystemPrompt: "You are a helpful assistant",
					Provider:     "openai",
					Model:        "gpt-4",
					MaxTokens:    1000,
					Temperature:  0.7,
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.config.Validate()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("Message types", func(t *testing.T) {
		// Test message creation
		userMsg := NewUserMessage("Hello")
		assert.Equal(t, UserRole, userMsg.Role)
		assert.Equal(t, "Hello", userMsg.Content)

		assistantMsg := NewAssistantMessage("Hi there!")
		assert.Equal(t, AssistantRole, assistantMsg.Role)
		assert.Equal(t, "Hi there!", assistantMsg.Content)

		systemMsg := NewSystemMessage("You are helpful")
		assert.Equal(t, SystemRole, systemMsg.Role)
		assert.Equal(t, "You are helpful", systemMsg.Content)
	})

	t.Run("ExecutionOptions", func(t *testing.T) {
		opts := &ExecutionOptions{
			Stream:      true,
			MaxTokens:   500,
			Temperature: 0.8,
			Timeout:     30 * time.Second,
		}

		assert.True(t, opts.Stream)
		assert.Equal(t, 500, opts.MaxTokens)
		assert.Equal(t, 0.8, opts.Temperature)
		assert.Equal(t, 30*time.Second, opts.Timeout)

		// Test nil options
		var nilOpts *ExecutionOptions
		assert.Nil(t, nilOpts)
	})

	t.Run("ExecutionResult", func(t *testing.T) {
		result := &ExecutionResult{
			Response: "Test response",
			Messages: []Message{
				NewUserMessage("Question"),
				NewAssistantMessage("Answer"),
			},
			TokensUsed: 50,
			Duration:   100 * time.Millisecond,
		}

		assert.Equal(t, "Test response", result.Response)
		assert.Len(t, result.Messages, 2)
		assert.Equal(t, 50, result.TokensUsed)
		assert.Equal(t, 100*time.Millisecond, result.Duration)
	})
}

// MockAgent implementation is now in mock_agent.go

func TestMockAgent(t *testing.T) {
	ctx := context.Background()

	t.Run("basic operations", func(t *testing.T) {
		agent := NewMockAgent("test-agent")

		// Test name
		assert.Equal(t, "test-agent", agent.Name())

		// Test initialization
		err := agent.Initialize(ctx)
		require.NoError(t, err)
		assert.True(t, agent.initialized)

		// Test system prompt
		agent.SetSystemPrompt("Test prompt")
		assert.Equal(t, "Test prompt", agent.GetSystemPrompt())

		// Test tools
		err = agent.AddTool("calculator")
		require.NoError(t, err)
		err = agent.AddTool("web_fetch")
		require.NoError(t, err)
		assert.Equal(t, []string{"calculator", "web_fetch"}, agent.GetTools())

		// Test cleanup
		err = agent.Cleanup()
		require.NoError(t, err)
		assert.False(t, agent.initialized)
	})

	t.Run("execution", func(t *testing.T) {
		agent := NewMockAgent("test-agent")
		agent.SetResponse("Custom response")

		result, err := agent.Execute(ctx, "Test input", nil)
		require.NoError(t, err)
		assert.Equal(t, "Custom response", result.Response)
		assert.Len(t, result.Messages, 2)
		assert.Equal(t, UserRole, result.Messages[0].Role)
		assert.Equal(t, "Test input", result.Messages[0].Content)
		assert.Equal(t, AssistantRole, result.Messages[1].Role)
		assert.Equal(t, "Custom response", result.Messages[1].Content)
	})

	t.Run("streaming", func(t *testing.T) {
		agent := NewMockAgent("test-agent")
		agent.SetResponse("Hello world")

		var collected string
		err := agent.Stream(ctx, "Test", nil, func(chunk string) error {
			collected += chunk
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, "Hello world", collected)
	})
}
