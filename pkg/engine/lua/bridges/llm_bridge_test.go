// ABOUTME: Tests for the Lua LLM bridge implementation
// ABOUTME: Verifies LLM operations exposed to Lua scripts work correctly

package bridges

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

// mockLLMBridge is a test double for bridge.LLMBridge
type mockLLMBridge struct {
	bridge.LLMBridge // Embed to get all methods

	// Track calls and responses
	chatCalled        bool
	chatResponse      string
	chatError         error
	completeCalled    bool
	completeResponse  string
	completeError     error
	streamCalled      bool
	streamChunks      []string
	streamError       error
	listModelsCalled  bool
	models            []map[string]interface{}
	listModelsError   error
	providers         []string
	currentProvider   string
	setProviderError  error
	setProviderCalled bool
}

func newMockLLMBridge() *mockLLMBridge {
	return &mockLLMBridge{
		providers:       []string{"openai", "anthropic", "gemini"},
		currentProvider: "openai",
		models: []map[string]interface{}{
			{"provider": "openai", "name": "gpt-4"},
			{"provider": "openai", "name": "gpt-3.5-turbo"},
			{"provider": "anthropic", "name": "claude-3-sonnet"},
		},
		streamChunks: []string{"Chunk 1: ", "Processing ", "data"},
	}
}

func (m *mockLLMBridge) Chat(ctx context.Context, prompt string) (string, error) {
	m.chatCalled = true
	if m.chatError != nil {
		return "", m.chatError
	}
	if m.chatResponse != "" {
		return m.chatResponse, nil
	}
	return fmt.Sprintf("Response to: %s", prompt), nil
}

func (m *mockLLMBridge) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	m.completeCalled = true
	if m.completeError != nil {
		return "", m.completeError
	}
	if m.completeResponse != "" {
		return m.completeResponse, nil
	}
	return fmt.Sprintf("Completion for: %s (max tokens: %d)", prompt, maxTokens), nil
}

func (m *mockLLMBridge) StreamChat(ctx context.Context, prompt string, callback func(string) error) error {
	m.streamCalled = true
	if m.streamError != nil {
		return m.streamError
	}

	// Simulate streaming with predefined chunks
	for _, chunk := range m.streamChunks {
		if err := callback(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockLLMBridge) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	m.listModelsCalled = true
	if m.listModelsError != nil {
		return nil, m.listModelsError
	}
	return m.models, nil
}

func (m *mockLLMBridge) ListProviders() []string {
	return m.providers
}

func (m *mockLLMBridge) GetCurrentProvider() string {
	return m.currentProvider
}

func (m *mockLLMBridge) SetProvider(name string) error {
	m.setProviderCalled = true
	if m.setProviderError != nil {
		return m.setProviderError
	}
	// Check if provider exists
	for _, p := range m.providers {
		if p == name {
			m.currentProvider = name
			return nil
		}
	}
	return fmt.Errorf("provider not found: %s", name)
}

func TestLLMBridgeRegister(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockLLMBridge()
	llmBridge := NewLLMBridge(mockBridge)

	err := llmBridge.Register(L)
	require.NoError(t, err)

	// Check if llm module is registered
	llm := L.GetGlobal("llm")
	require.NotEqual(t, lua.LNil, llm)
	require.Equal(t, lua.LTTable, llm.Type())

	// Check if all functions are registered
	llmTable := llm.(*lua.LTable)
	functions := []string{
		"chat", "complete", "stream_chat", "list_models",
		"list_providers", "get_provider", "set_provider",
		"chat_async", "complete_async",
	}

	for _, fn := range functions {
		f := llmTable.RawGetString(fn)
		assert.NotEqual(t, lua.LNil, f, "Function %s should be registered", fn)
		assert.Equal(t, lua.LTFunction, f.Type(), "llm.%s should be a function", fn)
	}
}

func TestLLMBridgeChat(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockLLMBridge()
	llmBridge := NewLLMBridge(mockBridge)
	require.NoError(t, llmBridge.Register(L))

	// Test successful chat
	err := L.DoString(`
		local response, err = llm.chat("Hello, world!")
		assert(response == "Response to: Hello, world!", "Response should match")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.chatCalled)

	// Test chat with error
	mockBridge.chatError = errors.New("chat failed")
	err = L.DoString(`
		local response, err = llm.chat("Test prompt")
		assert(response == nil, "Response should be nil on error")
		assert(err == "chat failed", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestLLMBridgeComplete(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockLLMBridge()
	llmBridge := NewLLMBridge(mockBridge)
	require.NoError(t, llmBridge.Register(L))

	// Test completion without maxTokens
	err := L.DoString(`
		local response, err = llm.complete("Complete this:")
		assert(response == "Completion for: Complete this: (max tokens: 0)", "Response should match")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.completeCalled)

	// Test completion with maxTokens
	mockBridge.completeCalled = false
	err = L.DoString(`
		local response, err = llm.complete("Complete this:", 100)
		assert(response == "Completion for: Complete this: (max tokens: 100)", "Response should match")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.completeCalled)

	// Test completion with error
	mockBridge.completeError = errors.New("completion failed")
	err = L.DoString(`
		local response, err = llm.complete("Test")
		assert(response == nil, "Response should be nil on error")
		assert(err == "completion failed", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestLLMBridgeStreamChat(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockLLMBridge()
	llmBridge := NewLLMBridge(mockBridge)
	require.NoError(t, llmBridge.Register(L))

	// Test successful streaming
	err := L.DoString(`
		local chunks = {}
		local err = llm.stream_chat("Hello", function(chunk)
			table.insert(chunks, chunk)
		end)
		
		assert(err == nil, "Error should be nil")
		assert(#chunks == 3, "Should receive 3 chunks")
		assert(chunks[1] == "Chunk 1: ", "First chunk should match")
		assert(chunks[2] == "Processing ", "Second chunk should match")
		assert(chunks[3] == "data", "Third chunk should match")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.streamCalled)

	// Test streaming with callback error
	err = L.DoString(`
		local err = llm.stream_chat("Test", function(chunk)
			return "callback error"
		end)
		
		assert(err == "callback error", "Error should match callback error")
	`)
	require.NoError(t, err)

	// Test streaming with bridge error
	mockBridge.streamError = errors.New("stream failed")
	err = L.DoString(`
		local err = llm.stream_chat("Test", function(chunk)
			-- This won't be called
		end)
		
		assert(err == "stream failed", "Error should match")
	`)
	require.NoError(t, err)
}

func TestLLMBridgeListModels(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockLLMBridge()
	llmBridge := NewLLMBridge(mockBridge)
	require.NoError(t, llmBridge.Register(L))

	// Test successful list models
	err := L.DoString(`
		local models, err = llm.list_models()
		
		assert(models ~= nil, "Models should not be nil")
		assert(err == nil, "Error should be nil")
		assert(#models == 3, "Should have 3 models")
		
		-- Check first model
		assert(models[1].provider == "openai", "First model provider should be openai")
		assert(models[1].name == "gpt-4", "First model name should be gpt-4")
		
		-- Check third model
		assert(models[3].provider == "anthropic", "Third model provider should be anthropic")
		assert(models[3].name == "claude-3-sonnet", "Third model name should be claude-3-sonnet")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.listModelsCalled)

	// Test list models with error
	mockBridge.listModelsError = errors.New("failed to list models")
	err = L.DoString(`
		local models, err = llm.list_models()
		
		assert(models == nil, "Models should be nil on error")
		assert(err == "failed to list models", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestLLMBridgeProviders(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockLLMBridge()
	llmBridge := NewLLMBridge(mockBridge)
	require.NoError(t, llmBridge.Register(L))

	// Test list providers
	err := L.DoString(`
		local providers = llm.list_providers()
		
		assert(#providers == 3, "Should have 3 providers")
		assert(providers[1] == "openai", "First provider should be openai")
		assert(providers[2] == "anthropic", "Second provider should be anthropic")
		assert(providers[3] == "gemini", "Third provider should be gemini")
	`)
	require.NoError(t, err)

	// Test get current provider
	err = L.DoString(`
		local provider = llm.get_provider()
		assert(provider == "openai", "Current provider should be openai")
	`)
	require.NoError(t, err)

	// Test set provider
	err = L.DoString(`
		local err = llm.set_provider("anthropic")
		assert(err == nil, "Error should be nil")
		
		local provider = llm.get_provider()
		assert(provider == "anthropic", "Current provider should be anthropic")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.setProviderCalled)

	// Test set invalid provider
	err = L.DoString(`
		local err = llm.set_provider("invalid-provider")
		assert(err == "provider not found: invalid-provider", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestLLMBridgeAsyncFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockLLMBridge()
	llmBridge := NewLLMBridge(mockBridge)
	require.NoError(t, llmBridge.Register(L))

	// Just verify async functions are registered
	// Actual async functionality is tested in llm_bridge_async_test.go
	err := L.DoString(`
		assert(type(llm.chat_async) == "function", "chat_async should be a function")
		assert(type(llm.complete_async) == "function", "complete_async should be a function")
	`)
	require.NoError(t, err)
}

func TestLLMBridgeIntegration(t *testing.T) {
	// Skip if not integration test
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	L := lua.NewState()
	defer L.Close()

	// Create a real LLM bridge
	realBridge, err := bridge.NewLLMBridge()
	require.NoError(t, err)

	// Wrap it in the adapter
	adapter := NewLLMBridgeAdapter(realBridge)
	llmBridge := NewLLMBridge(adapter)
	require.NoError(t, llmBridge.Register(L))

	// Test basic functionality with real bridge
	err = L.DoString(`
		-- List providers
		local providers = llm.list_providers()
		assert(type(providers) == "table", "Providers should be a table")
		
		-- Get current provider (may be empty if no providers configured)
		local provider = llm.get_provider()
		assert(type(provider) == "string", "Provider should be a string")
	`)
	require.NoError(t, err)
}
