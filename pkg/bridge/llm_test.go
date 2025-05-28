// ABOUTME: Tests for LLM bridge functionality including provider switching
// ABOUTME: Validates multi-provider support, model listing, and streaming

package bridge

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	modelinfodomain "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

// MockProvider implements a test provider
type MockProvider struct {
	name               string
	generateFunc       func(ctx context.Context, prompt string, options ...domain.Option) (string, error)
	generateMsgFunc    func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error)
	streamMsgFunc      func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error)
	generateSchemaFunc func(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...domain.Option) (interface{}, error)
	streamFunc         func(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error)
}

func (m *MockProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, prompt, options...)
	}
	return "mock response", nil
}

func (m *MockProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	if m.generateMsgFunc != nil {
		return m.generateMsgFunc(ctx, messages, options...)
	}
	return domain.Response{
		Content: "mock chat response",
	}, nil
}

func (m *MockProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	if m.streamMsgFunc != nil {
		return m.streamMsgFunc(ctx, messages, options...)
	}

	// Default streaming implementation
	ch := make(chan domain.Token)
	go func() {
		defer close(ch)
		chunks := []string{"Hello", " ", "from", " ", "mock", " ", "stream"}
		for i, chunk := range chunks {
			select {
			case <-ctx.Done():
				return
			case ch <- domain.Token{
				Text:     chunk,
				Finished: i == len(chunks)-1,
			}:
			}
		}
	}()
	return ch, nil
}

func (m *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...domain.Option) (interface{}, error) {
	if m.generateSchemaFunc != nil {
		return m.generateSchemaFunc(ctx, prompt, schema, options...)
	}
	return nil, errors.New("GenerateWithSchema not implemented")
}

func (m *MockProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, prompt, options...)
	}
	return nil, errors.New("Stream not implemented")
}

func TestLLMBridge(t *testing.T) {
	t.Run("create bridge with no API keys", func(t *testing.T) {
		// Save current env vars
		oldOpenAI := os.Getenv("OPENAI_API_KEY")
		oldAnthropic := os.Getenv("ANTHROPIC_API_KEY")
		oldGemini := os.Getenv("GEMINI_API_KEY")

		// Clear all API keys
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("GEMINI_API_KEY")

		bridge, err := NewLLMBridge()
		if err == nil {
			t.Error("expected error when no API keys are set")
		}
		if bridge != nil {
			t.Error("expected nil bridge when creation fails")
		}

		// Restore env vars
		if oldOpenAI != "" {
			os.Setenv("OPENAI_API_KEY", oldOpenAI)
		}
		if oldAnthropic != "" {
			os.Setenv("ANTHROPIC_API_KEY", oldAnthropic)
		}
		if oldGemini != "" {
			os.Setenv("GEMINI_API_KEY", oldGemini)
		}
	})

	t.Run("create bridge with mock providers", func(t *testing.T) {
		bridge := &LLMBridge{
			providers: make(map[string]domain.Provider),
		}

		// Add mock providers
		bridge.providers["openai"] = &MockProvider{name: "openai"}
		bridge.providers["anthropic"] = &MockProvider{name: "anthropic"}
		bridge.current = "openai"

		// Test listing providers
		providers := bridge.ListProviders()
		if len(providers) != 2 {
			t.Errorf("expected 2 providers, got %d", len(providers))
		}

		// Test current provider
		current := bridge.GetCurrentProvider()
		if current != "openai" {
			t.Errorf("expected current provider to be openai, got %s", current)
		}

		// Test switching provider
		err := bridge.SetProvider("anthropic")
		if err != nil {
			t.Errorf("failed to switch provider: %v", err)
		}

		if bridge.GetCurrentProvider() != "anthropic" {
			t.Errorf("expected current provider to be anthropic after switch")
		}

		// Test switching to non-existent provider
		err = bridge.SetProvider("nonexistent")
		if err == nil {
			t.Error("expected error when switching to non-existent provider")
		}
	})

	t.Run("chat functionality", func(t *testing.T) {
		bridge := &LLMBridge{
			providers: make(map[string]domain.Provider),
			current:   "test",
		}

		called := false
		bridge.providers["test"] = &MockProvider{
			generateMsgFunc: func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
				called = true
				if len(messages) != 1 {
					t.Errorf("expected 1 message, got %d", len(messages))
				}
				if messages[0].Role != domain.RoleUser {
					t.Errorf("expected user role, got %s", messages[0].Role)
				}
				return domain.Response{Content: "test response"}, nil
			},
		}

		ctx := context.Background()
		response, err := bridge.Chat(ctx, "test prompt")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if response != "test response" {
			t.Errorf("expected 'test response', got %s", response)
		}
		if !called {
			t.Error("provider GenerateMessage was not called")
		}
	})

	t.Run("complete functionality", func(t *testing.T) {
		bridge := &LLMBridge{
			providers: make(map[string]domain.Provider),
			current:   "test",
		}

		called := false
		bridge.providers["test"] = &MockProvider{
			generateFunc: func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
				called = true
				if prompt != "test prompt" {
					t.Errorf("expected 'test prompt', got %s", prompt)
				}
				// We can't easily verify options as they're provider-specific
				return "completion result", nil
			},
		}

		ctx := context.Background()
		response, err := bridge.Complete(ctx, "test prompt", 100)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if response != "completion result" {
			t.Errorf("expected 'completion result', got %s", response)
		}
		if !called {
			t.Error("provider Generate was not called")
		}
	})

	t.Run("streaming functionality", func(t *testing.T) {
		bridge := &LLMBridge{
			providers: make(map[string]domain.Provider),
			current:   "test",
		}

		bridge.providers["test"] = &MockProvider{} // Uses default streaming implementation

		ctx := context.Background()
		var chunks []string
		err := bridge.StreamChat(ctx, "test prompt", func(chunk string) error {
			chunks = append(chunks, chunk)
			return nil
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := "Hello from mock stream"
		result := strings.Join(chunks, "")
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("streaming with callback error", func(t *testing.T) {
		bridge := &LLMBridge{
			providers: make(map[string]domain.Provider),
			current:   "test",
		}

		bridge.providers["test"] = &MockProvider{} // Uses default streaming implementation

		ctx := context.Background()
		callbackErr := errors.New("callback error")
		err := bridge.StreamChat(ctx, "test prompt", func(chunk string) error {
			return callbackErr
		})

		if err == nil {
			t.Error("expected error from callback")
		}
		if !strings.Contains(err.Error(), "callback error") {
			t.Errorf("expected callback error, got: %v", err)
		}
	})

	t.Run("bridge interface implementation", func(t *testing.T) {
		bridge := &LLMBridge{
			providers: make(map[string]domain.Provider),
			current:   "test",
		}

		// Test Name
		if bridge.Name() != "llm" {
			t.Errorf("expected bridge name 'llm', got %s", bridge.Name())
		}

		// Test Methods
		methods := bridge.Methods()
		if len(methods) != 8 {
			t.Errorf("expected 8 methods, got %d", len(methods))
		}

		// Verify key methods exist
		methodNames := make(map[string]bool)
		for _, m := range methods {
			methodNames[m.Name] = true
		}

		expectedMethods := []string{
			"chat", "complete", "streamChat", "setProvider",
			"getCurrentProvider", "listProviders", "listModels", "listModelsForProvider",
		}

		for _, expected := range expectedMethods {
			if !methodNames[expected] {
				t.Errorf("missing expected method: %s", expected)
			}
		}

		// Test Initialize
		ctx := context.Background()
		if err := bridge.Initialize(ctx); err != nil {
			t.Errorf("unexpected error from Initialize: %v", err)
		}

		// Test Cleanup
		bridge.providers["test"] = &MockProvider{}
		if err := bridge.Cleanup(ctx); err != nil {
			t.Errorf("unexpected error from Cleanup: %v", err)
		}

		if len(bridge.providers) != 0 {
			t.Error("expected providers to be cleared after cleanup")
		}
		if bridge.current != "" {
			t.Error("expected current provider to be cleared after cleanup")
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		bridge := &LLMBridge{
			providers: make(map[string]domain.Provider),
			current:   "test",
		}

		bridge.providers["test"] = &MockProvider{}
		bridge.providers["test2"] = &MockProvider{}

		var wg sync.WaitGroup
		errors := make([]error, 0)
		var errorsMu sync.Mutex

		// Concurrent operations
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				ctx := context.Background()

				// Alternate between providers
				provider := "test"
				if i%2 == 0 {
					provider = "test2"
				}

				if err := bridge.SetProvider(provider); err != nil {
					errorsMu.Lock()
					errors = append(errors, err)
					errorsMu.Unlock()
					return
				}

				// Try to use the provider
				_, err := bridge.Chat(ctx, "concurrent test")
				if err != nil {
					errorsMu.Lock()
					errors = append(errors, err)
					errorsMu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		if len(errors) > 0 {
			t.Errorf("concurrent access errors: %v", errors)
		}
	})
}

// TestModelListing tests model listing functionality
func TestModelListing(t *testing.T) {
	t.Run("list models with empty inventory", func(t *testing.T) {
		bridge := &LLMBridge{
			providers: make(map[string]domain.Provider),
		}

		ctx := context.Background()
		_ = ctx    // Mark as used
		_ = bridge // Mark as used

		// This test would require mocking llmutil.GetAvailableModels
		// For now, we'll skip the actual implementation test
		// In a real scenario, we'd use dependency injection or a test double
		t.Skip("Requires mocking of llmutil.GetAvailableModels")
	})

	t.Run("filter models by provider", func(t *testing.T) {
		// This test would require mocking llmutil.GetAvailableModels
		// For a proper test, we'd need to refactor to allow dependency injection
		t.Skip("Requires proper mocking setup for model inventory")
	})

	t.Run("convert model info", func(t *testing.T) {
		bridge := &LLMBridge{}

		// Test model conversion
		model := modelinfodomain.Model{
			Provider:       "openai",
			Name:           "gpt-4",
			DisplayName:    "GPT-4",
			Description:    "Advanced language model",
			ContextWindow:  8192,
			ModelFamily:    "gpt-4",
			TrainingCutoff: "2023-09",
			Capabilities: modelinfodomain.Capabilities{
				FunctionCalling: true,
				Streaming:       true,
				JSONMode:        true,
			},
		}

		info := bridge.convertToModelInfo(model)

		if info.ID != "gpt-4" {
			t.Errorf("expected ID 'gpt-4', got %s", info.ID)
		}
		if info.Name != "GPT-4" {
			t.Errorf("expected Name 'GPT-4', got %s", info.Name)
		}
		if info.Provider != "openai" {
			t.Errorf("expected Provider 'openai', got %s", info.Provider)
		}
		if info.ContextSize != 8192 {
			t.Errorf("expected ContextSize 8192, got %d", info.ContextSize)
		}

		// Check metadata
		if info.Metadata["description"] != "Advanced language model" {
			t.Errorf("expected description in metadata")
		}
		if info.Metadata["training_cutoff"] != "2023-09" {
			t.Errorf("expected training_cutoff in metadata")
		}
		if info.Metadata["model_family"] != "gpt-4" {
			t.Errorf("expected model_family in metadata")
		}

		// Check capabilities
		caps := info.Metadata["capabilities"]
		if !strings.Contains(caps, "function_calling") {
			t.Errorf("expected function_calling in capabilities")
		}
		if !strings.Contains(caps, "streaming") {
			t.Errorf("expected streaming in capabilities")
		}
		if !strings.Contains(caps, "json_mode") {
			t.Errorf("expected json_mode in capabilities")
		}
	})
}
