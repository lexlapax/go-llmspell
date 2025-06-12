// ABOUTME: Test suite for the core LLM bridge that provides access to language model providers.
// ABOUTME: Tests provider interface bridging, message handling, provider switching, and streaming responses.

package bridge

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
)

// Mock LLM provider for testing
type mockLLMProvider struct {
	name         string
	model        string
	response     string
	error        error
	streamChunks []string
	callCount    int
	lastMessages []LLMMessage
	lastOptions  map[string]interface{}
	mu           sync.Mutex
}

func (m *mockLLMProvider) Complete(ctx context.Context, messages []LLMMessage, options map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	m.lastMessages = messages
	m.lastOptions = options

	if m.error != nil {
		return "", m.error
	}
	return m.response, nil
}

func (m *mockLLMProvider) CompleteStream(ctx context.Context, messages []LLMMessage, options map[string]interface{}) (<-chan string, <-chan error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	m.lastMessages = messages
	m.lastOptions = options

	chunkChan := make(chan string, len(m.streamChunks))
	errorChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errorChan)

		if m.error != nil {
			errorChan <- m.error
			return
		}

		for _, chunk := range m.streamChunks {
			select {
			case chunkChan <- chunk:
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			}
		}
	}()

	return chunkChan, errorChan
}

func (m *mockLLMProvider) GetName() string {
	return m.name
}

func (m *mockLLMProvider) GetModel() string {
	return m.model
}

func (m *mockLLMProvider) IsAvailable() bool {
	return m.error == nil
}

// Tests for LLMBridge
func TestNewLLMBridge(t *testing.T) {
	bridge := NewLLMBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "llm", bridge.GetID())
}

func TestLLMBridgeMetadata(t *testing.T) {
	bridge := NewLLMBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "llm", metadata.Name)
	assert.NotEmpty(t, metadata.Version)
	assert.NotEmpty(t, metadata.Description)
	assert.Contains(t, metadata.Description, "LLM")
}

func TestLLMProviderManagement(t *testing.T) {
	t.Run("Register Provider", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{name: "openai", model: "gpt-4"}

		err := bridge.RegisterProvider("openai", provider)
		assert.NoError(t, err)

		// Get registered provider
		retrieved, err := bridge.GetProvider("openai")
		assert.NoError(t, err)
		assert.Equal(t, provider, retrieved)
	})

	t.Run("Register Multiple Providers", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider1 := &mockLLMProvider{name: "openai", model: "gpt-4"}
		provider2 := &mockLLMProvider{name: "anthropic", model: "claude-3"}

		_ = bridge.RegisterProvider("openai", provider1)
		_ = bridge.RegisterProvider("anthropic", provider2)

		providers := bridge.ListProviders()
		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "openai")
		assert.Contains(t, providers, "anthropic")
	})

	t.Run("Duplicate Provider Registration", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider1 := &mockLLMProvider{name: "openai"}
		provider2 := &mockLLMProvider{name: "openai-2"}

		err := bridge.RegisterProvider("openai", provider1)
		assert.NoError(t, err)

		err = bridge.RegisterProvider("openai", provider2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("Set Active Provider", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider1 := &mockLLMProvider{name: "openai"}
		provider2 := &mockLLMProvider{name: "anthropic"}

		_ = bridge.RegisterProvider("openai", provider1)
		_ = bridge.RegisterProvider("anthropic", provider2)

		// Set active provider
		err := bridge.SetActiveProvider("anthropic")
		assert.NoError(t, err)

		active := bridge.GetActiveProvider()
		assert.Equal(t, "anthropic", active)
	})

	t.Run("Set Non-existent Active Provider", func(t *testing.T) {
		bridge := NewLLMBridge()

		err := bridge.SetActiveProvider("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestLLMCompletion(t *testing.T) {
	t.Run("Simple Completion", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:     "test",
			response: "Hello, world!",
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Say hello"},
		}

		response, err := bridge.Complete(ctx, messages, nil)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, world!", response)
		assert.Equal(t, 1, provider.callCount)
	})

	t.Run("Completion with Options", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:     "test",
			response: "Response",
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Test"},
		}
		options := map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  100,
		}

		_, err := bridge.Complete(ctx, messages, options)
		assert.NoError(t, err)
		assert.Equal(t, options, provider.lastOptions)
	})

	t.Run("Completion Error", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:  "test",
			error: errors.New("API error"),
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Test"},
		}

		_, err := bridge.Complete(ctx, messages, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("No Active Provider", func(t *testing.T) {
		bridge := NewLLMBridge()
		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Test"},
		}

		_, err := bridge.Complete(ctx, messages, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active provider")
	})
}

func TestLLMStreamingCompletion(t *testing.T) {
	t.Run("Stream Completion", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:         "test",
			streamChunks: []string{"Hello", ", ", "world", "!"},
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Say hello"},
		}

		chunkChan, errorChan := bridge.CompleteStream(ctx, messages, nil)

		var chunks []string
		done := false

		for !done {
			select {
			case chunk, ok := <-chunkChan:
				if !ok {
					done = true
					break
				}
				chunks = append(chunks, chunk)
			case err := <-errorChan:
				assert.NoError(t, err)
			}
		}

		assert.Equal(t, []string{"Hello", ", ", "world", "!"}, chunks)
		assert.Equal(t, 1, provider.callCount)
	})

	t.Run("Stream Completion Error", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:  "test",
			error: errors.New("stream error"),
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Test"},
		}

		chunkChan, errorChan := bridge.CompleteStream(ctx, messages, nil)

		select {
		case <-chunkChan:
			t.Fatal("Expected error, got chunk")
		case err := <-errorChan:
			if err == nil {
				t.Fatal("Expected error but got nil")
			}
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "stream error")
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for error")
		}
	})

	t.Run("Stream Context Cancellation", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:         "test",
			streamChunks: []string{"chunk1", "chunk2", "chunk3"},
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx, cancel := context.WithCancel(context.Background())
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Test"},
		}

		chunkChan, errorChan := bridge.CompleteStream(ctx, messages, nil)

		// Cancel context immediately
		cancel()

		// Should get context cancellation error
		select {
		case <-chunkChan:
			// May get some chunks before cancellation
		case err := <-errorChan:
			assert.Error(t, err)
			assert.True(t, errors.Is(err, context.Canceled))
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for cancellation")
		}
	})
}

func TestLLMBridgeEngineIntegration(t *testing.T) {
	t.Run("Register With Engine", func(t *testing.T) {
		bridge := NewLLMBridge()
		engine := &mockScriptEngine{}

		err := bridge.RegisterWithEngine(engine)
		assert.NoError(t, err)
		assert.Len(t, engine.bridges, 1)
		assert.Equal(t, bridge, engine.bridges[0])
	})

	t.Run("Bridge Methods", func(t *testing.T) {
		bridge := NewLLMBridge()
		methods := bridge.Methods()

		// Should expose key LLM methods
		methodNames := make([]string, len(methods))
		for i, m := range methods {
			methodNames[i] = m.Name
		}

		assert.Contains(t, methodNames, "complete")
		assert.Contains(t, methodNames, "completeStream")
		assert.Contains(t, methodNames, "setProvider")
		assert.Contains(t, methodNames, "listProviders")
		assert.Contains(t, methodNames, "createMessage")
	})

	t.Run("Type Mappings", func(t *testing.T) {
		bridge := NewLLMBridge()
		mappings := bridge.TypeMappings()

		// Should provide mappings for LLM types
		assert.Contains(t, mappings, "LLMMessage")
		assert.Contains(t, mappings, "CompletionOptions")
	})

	t.Run("Required Permissions", func(t *testing.T) {
		bridge := NewLLMBridge()
		permissions := bridge.RequiredPermissions()

		// LLM bridge should require network permission
		hasNetwork := false
		for _, p := range permissions {
			if p.Type == engine.PermissionNetwork {
				hasNetwork = true
				break
			}
		}
		assert.True(t, hasNetwork, "LLM bridge should require network permission")
	})
}

func TestMessageHandling(t *testing.T) {
	t.Run("Create Messages", func(t *testing.T) {
		bridge := NewLLMBridge()

		// Create system message
		sysMsg := bridge.CreateMessage("system", "You are a helpful assistant")
		assert.Equal(t, "system", sysMsg.Role)
		assert.Equal(t, "You are a helpful assistant", sysMsg.Content)

		// Create user message
		userMsg := bridge.CreateMessage("user", "Hello")
		assert.Equal(t, "user", userMsg.Role)
		assert.Equal(t, "Hello", userMsg.Content)

		// Create assistant message
		assistMsg := bridge.CreateMessage("assistant", "Hi there!")
		assert.Equal(t, "assistant", assistMsg.Role)
		assert.Equal(t, "Hi there!", assistMsg.Content)
	})

	t.Run("Message Validation", func(t *testing.T) {
		bridge := NewLLMBridge()

		// Valid message
		msg := LLMMessage{Role: "user", Content: "Test"}
		err := bridge.ValidateMessage(msg)
		assert.NoError(t, err)

		// Invalid role
		msg = LLMMessage{Role: "invalid", Content: "Test"}
		err = bridge.ValidateMessage(msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")

		// Empty content
		msg = LLMMessage{Role: "user", Content: ""}
		err = bridge.ValidateMessage(msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty content")
	})

	t.Run("Message History", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:     "test",
			response: "Response",
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// Create conversation history
		messages := []LLMMessage{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "What's 2+2?"},
			{Role: "assistant", Content: "2+2 equals 4"},
			{Role: "user", Content: "What's 3+3?"},
		}

		_, err := bridge.Complete(ctx, messages, nil)
		assert.NoError(t, err)
		assert.Equal(t, messages, provider.lastMessages)
	})
}

func TestProviderSwitching(t *testing.T) {
	t.Run("Dynamic Provider Switching", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider1 := &mockLLMProvider{
			name:     "provider1",
			response: "Response from provider 1",
		}
		provider2 := &mockLLMProvider{
			name:     "provider2",
			response: "Response from provider 2",
		}

		_ = bridge.RegisterProvider("p1", provider1)
		_ = bridge.RegisterProvider("p2", provider2)

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Test"},
		}

		// Use provider 1
		_ = bridge.SetActiveProvider("p1")
		resp1, err := bridge.Complete(ctx, messages, nil)
		assert.NoError(t, err)
		assert.Equal(t, "Response from provider 1", resp1)

		// Switch to provider 2
		_ = bridge.SetActiveProvider("p2")
		resp2, err := bridge.Complete(ctx, messages, nil)
		assert.NoError(t, err)
		assert.Equal(t, "Response from provider 2", resp2)

		assert.Equal(t, 1, provider1.callCount)
		assert.Equal(t, 1, provider2.callCount)
	})

	t.Run("Provider Availability Check", func(t *testing.T) {
		bridge := NewLLMBridge()
		availableProvider := &mockLLMProvider{
			name:     "available",
			response: "OK",
		}
		unavailableProvider := &mockLLMProvider{
			name:  "unavailable",
			error: errors.New("service unavailable"),
		}

		_ = bridge.RegisterProvider("available", availableProvider)
		_ = bridge.RegisterProvider("unavailable", unavailableProvider)

		// Check availability
		assert.True(t, bridge.IsProviderAvailable("available"))
		assert.False(t, bridge.IsProviderAvailable("unavailable"))
		assert.False(t, bridge.IsProviderAvailable("non-existent"))
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("Concurrent Completions", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:     "test",
			response: "Concurrent response",
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Run 10 concurrent completions
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				messages := []LLMMessage{
					{Role: "user", Content: fmt.Sprintf("Request %d", id)},
				}
				_, err := bridge.Complete(ctx, messages, nil)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			assert.NoError(t, err)
		}

		assert.Equal(t, 10, provider.callCount)
	})

	t.Run("Concurrent Provider Management", func(t *testing.T) {
		bridge := NewLLMBridge()
		var wg sync.WaitGroup

		// Concurrent provider registration
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				provider := &mockLLMProvider{
					name: fmt.Sprintf("provider%d", id),
				}
				_ = bridge.RegisterProvider(fmt.Sprintf("p%d", id), provider)
			}(i)
		}

		// Concurrent provider listing
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				bridge.ListProviders()
			}()
		}

		wg.Wait()

		// All providers should be registered
		providers := bridge.ListProviders()
		assert.Len(t, providers, 5)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Empty Message List", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:     "test",
			response: "Response",
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		_, err := bridge.Complete(ctx, []LLMMessage{}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no messages")
	})

	t.Run("Nil Options", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:     "test",
			response: "Response",
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		messages := []LLMMessage{
			{Role: "user", Content: "Test"},
		}

		// Should handle nil options gracefully
		_, err := bridge.Complete(ctx, messages, nil)
		assert.NoError(t, err)
	})

	t.Run("Very Long Message", func(t *testing.T) {
		bridge := NewLLMBridge()
		provider := &mockLLMProvider{
			name:     "test",
			response: "OK",
		}

		_ = bridge.RegisterProvider("test", provider)
		_ = bridge.SetActiveProvider("test")

		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// Create a very long message
		longContent := strings.Repeat("This is a test. ", 1000)
		messages := []LLMMessage{
			{Role: "user", Content: longContent},
		}

		_, err := bridge.Complete(ctx, messages, nil)
		assert.NoError(t, err)
		assert.Equal(t, longContent, provider.lastMessages[0].Content)
	})
}

// Benchmark tests
func BenchmarkLLMComplete(b *testing.B) {
	bridge := NewLLMBridge()
	provider := &mockLLMProvider{
		name:     "test",
		response: "Benchmark response",
	}

	_ = bridge.RegisterProvider("test", provider)
	_ = bridge.SetActiveProvider("test")

	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	messages := []LLMMessage{
		{Role: "system", Content: "You are a benchmark assistant"},
		{Role: "user", Content: "This is a benchmark test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bridge.Complete(ctx, messages, nil)
	}
}

func BenchmarkMessageCreation(b *testing.B) {
	bridge := NewLLMBridge()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.CreateMessage("user", "Benchmark message content")
	}
}

func BenchmarkProviderSwitch(b *testing.B) {
	bridge := NewLLMBridge()

	// Register multiple providers
	for i := 0; i < 5; i++ {
		provider := &mockLLMProvider{
			name: fmt.Sprintf("provider%d", i),
		}
		_ = bridge.RegisterProvider(fmt.Sprintf("p%d", i), provider)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bridge.SetActiveProvider(fmt.Sprintf("p%d", i%5))
	}
}
