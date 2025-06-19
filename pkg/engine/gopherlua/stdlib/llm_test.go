// ABOUTME: Comprehensive test suite for LLM Operations Library in Lua standard library
// ABOUTME: Tests LLM operations, provider management, model discovery, and integration with promises

package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// MockLLMBridge represents a mock LLM bridge for testing
type MockLLMBridge struct {
	responses    map[string]interface{}
	streamChunks map[string][]string
	providers    []string
	models       map[string]interface{}
	callLog      []string
}

// NewMockLLMBridge creates a new mock LLM bridge
func NewMockLLMBridge() *MockLLMBridge {
	return &MockLLMBridge{
		responses:    make(map[string]interface{}),
		streamChunks: make(map[string][]string),
		providers:    []string{"openai", "anthropic", "mock"},
		models:       make(map[string]interface{}),
		callLog:      []string{},
	}
}

// SetupMockBridges sets up mock bridges in the Lua state
func setupMockBridges(L *lua.LState, mockBridge *MockLLMBridge) {
	// Create mock LLM bridge table
	llmBridge := L.NewTable()

	// Mock generate method
	llmBridge.RawSetString("generate", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		prompt := L.CheckString(2)
		_ = L.OptTable(3, L.NewTable()) // options parameter
		mockBridge.callLog = append(mockBridge.callLog, "generate:"+prompt)

		response := L.NewTable()
		response.RawSetString("content", lua.LString("Mock response to: "+prompt))
		response.RawSetString("usage", L.NewTable())

		L.Push(response)
		return 1
	}))

	// Mock generateWithProvider method
	llmBridge.RawSetString("generateWithProvider", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		provider := L.CheckString(2)
		prompt := L.CheckString(3)
		_ = L.OptTable(4, L.NewTable()) // options parameter
		mockBridge.callLog = append(mockBridge.callLog, "generateWithProvider:"+provider+":"+prompt)

		response := L.NewTable()
		response.RawSetString("content", lua.LString("Mock response from "+provider+" to: "+prompt))
		response.RawSetString("provider", lua.LString(provider))

		L.Push(response)
		return 1
	}))

	// Mock generateMessage method
	llmBridge.RawSetString("generateMessage", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)             // self (bridge table)
		_ = L.CheckTable(2)             // messages table - not used in mock
		_ = L.OptTable(3, L.NewTable()) // options parameter
		mockBridge.callLog = append(mockBridge.callLog, "generateMessage")

		response := L.NewTable()
		response.RawSetString("content", lua.LString("Mock conversation response"))

		L.Push(response)
		return 1
	}))

	// Mock listProviders method
	llmBridge.RawSetString("listProviders", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		providers := L.NewTable()
		for i, provider := range mockBridge.providers {
			providers.RawSetInt(i+1, lua.LString(provider))
		}
		L.Push(providers)
		return 1
	}))

	// Mock setProvider method
	llmBridge.RawSetString("setProvider", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		provider := L.CheckString(2)
		_ = L.OptTable(3, L.NewTable()) // config parameter (optional)
		mockBridge.callLog = append(mockBridge.callLog, "setProvider:"+provider)
		L.Push(lua.LTrue)
		return 1
	}))

	// Mock testProviderConnection method
	llmBridge.RawSetString("testProviderConnection", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		provider := L.CheckString(2)
		mockBridge.callLog = append(mockBridge.callLog, "testProviderConnection:"+provider)
		L.Push(lua.LTrue)
		return 1
	}))

	// Mock getModelInfo method
	llmBridge.RawSetString("getModelInfo", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		modelId := L.CheckString(2)
		mockBridge.callLog = append(mockBridge.callLog, "getModelInfo:"+modelId)

		info := L.NewTable()
		info.RawSetString("id", lua.LString(modelId))
		info.RawSetString("context_length", lua.LNumber(4096))
		info.RawSetString("supports_streaming", lua.LTrue)
		info.RawSetString("cost_per_token", lua.LNumber(0.001))
		info.RawSetString("currency", lua.LString("USD"))

		L.Push(info)
		return 1
	}))

	// Mock stream method
	llmBridge.RawSetString("stream", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		prompt := L.CheckString(2)
		_ = L.OptTable(3, L.NewTable()) // options parameter
		mockBridge.callLog = append(mockBridge.callLog, "stream:"+prompt)
		L.Push(lua.LString("stream_123"))
		return 1
	}))

	// Mock readStream method
	llmBridge.RawSetString("readStream", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		streamId := L.CheckString(2)
		mockBridge.callLog = append(mockBridge.callLog, "readStream:"+streamId)

		// Return chunks for testing
		chunks := mockBridge.streamChunks[streamId]
		if len(chunks) == 0 {
			L.Push(lua.LNil) // Stream ended
		} else {
			chunk := chunks[0]
			mockBridge.streamChunks[streamId] = chunks[1:] // Remove first chunk
			L.Push(lua.LString(chunk))
		}
		return 1
	}))

	// Mock closeStream method
	llmBridge.RawSetString("closeStream", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		streamId := L.CheckString(2)
		mockBridge.callLog = append(mockBridge.callLog, "closeStream:"+streamId)
		delete(mockBridge.streamChunks, streamId)
		L.Push(lua.LTrue)
		return 1
	}))

	// Add missing methods for model discovery
	llmBridge.RawSetString("listModels", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)  // self (bridge table)
		_ = L.CheckString(2) // provider parameter
		mockBridge.callLog = append(mockBridge.callLog, "listModels")

		models := L.NewTable()
		model := L.NewTable()
		model.RawSetString("id", lua.LString("gpt-4"))
		models.RawSetInt(1, model)

		L.Push(models)
		return 1
	}))

	// Add missing streamWithProvider method
	llmBridge.RawSetString("streamWithProvider", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)             // self (bridge table)
		_ = L.CheckString(2)            // provider parameter
		_ = L.CheckString(3)            // prompt parameter
		_ = L.OptTable(4, L.NewTable()) // options parameter
		mockBridge.callLog = append(mockBridge.callLog, "streamWithProvider")
		L.Push(lua.LString("stream_123"))
		return 1
	}))

	// Set up mock bridges as globals
	L.SetGlobal("llm_bridge", llmBridge)
	L.SetGlobal("provider_bridge", llmBridge) // Same for simplicity
	L.SetGlobal("pool_bridge", llmBridge)     // Same for simplicity
	L.SetGlobal("llm_util_bridge", llmBridge) // Same for simplicity
}

// setupLLMLibrary loads the LLM library and sets it as global
func setupLLMLibrary(t *testing.T, L *lua.LState) *MockLLMBridge {
	t.Helper()

	// Setup mock bridges first
	mockBridge := NewMockLLMBridge()
	setupMockBridges(L, mockBridge)

	// Load promise library (required dependency)
	promisePath := filepath.Join(".", "promise.lua")
	err := L.DoFile(promisePath)
	if err != nil {
		t.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)

	// Load LLM library
	llmPath := filepath.Join(".", "llm.lua")
	err = L.DoFile(llmPath)
	if err != nil {
		t.Fatalf("Failed to load LLM library: %v", err)
	}
	llmLib := L.Get(-1)
	L.SetGlobal("llm", llmLib)

	return mockBridge
}

// TestLLMLibraryLoading tests that the LLM library can be loaded
func TestLLMLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := setupLLMLibrary(t, L)

	// Check that LLM table exists and has expected structure
	script := `
		if type(llm) ~= "table" then
			error("LLM module should be a table")
		end
		
		if type(llm.quick_prompt) ~= "function" then
			error("llm.quick_prompt function should be available")
		end
		
		if type(llm.chat_session) ~= "function" then
			error("llm.chat_session function should be available")
		end
		
		if type(llm.use_provider) ~= "function" then
			error("llm.use_provider function should be available")
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("LLM library structure test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected library structure test to pass")
	}

	_ = mockBridge // Use mockBridge to avoid unused variable warning
}

// TestQuickPrompt tests the quick_prompt functionality
func TestQuickPrompt(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "basic_quick_prompt",
			script: `
				local response = llm.quick_prompt("Hello, world!")
				return response and response.content
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "Mock response to: Hello, world!" {
					t.Errorf("Expected mock response, got %v", result.String())
				}
			},
		},
		{
			name: "quick_prompt_with_options",
			script: `
				local response = llm.quick_prompt("Test prompt", {temperature = 0.5})
				return response and response.content
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "Mock response to: Test prompt" {
					t.Errorf("Expected mock response, got %v", result.String())
				}
			},
		},
		{
			name: "quick_prompt_validation",
			script: `
				local success, err = pcall(function()
					llm.quick_prompt(nil)
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for nil prompt")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupLLMLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestChatSession tests the chat session functionality
func TestChatSession(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_chat_session",
			script: `
				local session = llm.chat_session("You are a helpful assistant")
				return type(session) == "table" and type(session.send) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected chat session to be created properly")
				}
			},
		},
		{
			name: "chat_session_send",
			script: `
				local session = llm.chat_session("You are helpful")
				local response = session:send("Hello!")
				return response and response.content
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "Mock conversation response" {
					t.Errorf("Expected mock conversation response, got %v", result.String())
				}
			},
		},
		{
			name: "chat_session_history",
			script: `
				local session = llm.chat_session("System prompt")
				session:send("User message")
				local history = session:get_history()
				return #history >= 2 -- System + User + Assistant
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected chat history to contain messages")
				}
			},
		},
		{
			name: "chat_session_clear",
			script: `
				local session = llm.chat_session("System prompt")
				session:send("Message")
				session:clear()
				local history = session:get_history()
				return #history == 1 -- Only system message
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected chat history to be cleared to only system message")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupLLMLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestProviderManagement tests provider management functionality
func TestProviderManagement(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue, mockBridge *MockLLMBridge)
	}{
		{
			name: "use_provider",
			script: `
				local success = llm.use_provider("openai")
				local current = llm.get_current_provider()
				return success and current == "openai"
			`,
			check: func(t *testing.T, result lua.LValue, mockBridge *MockLLMBridge) {
				if result != lua.LTrue {
					t.Errorf("Expected provider to be set successfully")
				}
			},
		},
		{
			name: "list_providers",
			script: `
				local providers = llm.list_providers()
				return type(providers) == "table" and #providers > 0
			`,
			check: func(t *testing.T, result lua.LValue, mockBridge *MockLLMBridge) {
				if result != lua.LTrue {
					t.Errorf("Expected providers list to be non-empty table")
				}
			},
		},
		{
			name: "compare_providers",
			script: `
				local results = llm.compare_providers("Test prompt", {"openai", "anthropic"})
				return type(results) == "table" and #results == 2
			`,
			check: func(t *testing.T, result lua.LValue, mockBridge *MockLLMBridge) {
				if result != lua.LTrue {
					t.Errorf("Expected comparison results for 2 providers")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			mockBridge := setupLLMLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result, mockBridge)
		})
	}
}

// TestStreamingResponse tests streaming functionality
func TestStreamingResponse(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := setupLLMLibrary(t, L)

	// Set up stream chunks for testing
	mockBridge.streamChunks["stream_123"] = []string{"chunk1", "chunk2", "chunk3"}

	script := `
		local chunks = {}
		local promise = llm.streaming_response("Stream test", function(chunk)
			table.insert(chunks, chunk)
			return true -- continue
		end)
		
		-- For testing, we'll just check that the function returns a promise-like object
		return type(promise) == "table" and type(promise.andThen) == "function"
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Streaming test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected streaming to return a promise")
	}
}

// TestBatchProcessing tests batch processing functionality
func TestBatchProcessing(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "batch_process_sync",
			script: `
				local prompts = {"Hello", "How are you?", "Goodbye"}
				local results = llm.batch_process(prompts)
				return type(results) == "table" and #results == 3
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected batch processing to return 3 results")
				}
			},
		},
		{
			name: "batch_process_async",
			script: `
				local prompts = {"Hello", "Hi"}
				local promise = llm.batch_process_async(prompts)
				return type(promise) == "table" and type(promise.andThen) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected async batch processing to return a promise")
				}
			},
		},
		{
			name: "batch_process_validation",
			script: `
				local success, err = pcall(function()
					llm.batch_process("not a table")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid prompts parameter")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupLLMLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestModelDiscovery tests model discovery functionality
func TestModelDiscovery(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "model_info",
			script: `
				local info = llm.model_info("gpt-4")
				return type(info) == "table" and info.id == "gpt-4"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected model info to be returned")
				}
			},
		},
		{
			name: "cost_estimate",
			script: `
				local estimate = llm.cost_estimate("Hello world", "gpt-4", "openai")
				return type(estimate) == "table" and estimate.estimated_cost ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected cost estimate to be returned")
				}
			},
		},
		{
			name: "find_model",
			script: `
				local requirements = {
					min_context_length = 2000,
					supports_streaming = true
				}
				local models = llm.find_model(requirements)
				return type(models) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected model search to return a table")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupLLMLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestUtilities tests utility functions
func TestUtilities(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "set_and_get_defaults",
			script: `
				llm.set_defaults({temperature = 0.8, max_tokens = 2000})
				local defaults = llm.get_defaults()
				return defaults.temperature == 0.8 and defaults.max_tokens == 2000
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected defaults to be set and retrieved correctly")
				}
			},
		},
		{
			name: "reset_defaults",
			script: `
				llm.set_defaults({temperature = 0.9})
				llm.reset_defaults()
				local defaults = llm.get_defaults()
				return defaults.temperature == 0.7 -- original default
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected defaults to be reset to original values")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupLLMLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestLLMErrorHandling tests error handling in various scenarios
func TestLLMErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "missing_prompt_error",
			script: `
				local success, err = pcall(function()
					llm.quick_prompt()
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing prompt")
				}
			},
		},
		{
			name: "invalid_streaming_callback",
			script: `
				local success, err = pcall(function()
					llm.streaming_response("test", "not a function")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid callback")
				}
			},
		},
		{
			name: "invalid_defaults_type",
			script: `
				local success, err = pcall(function()
					llm.set_defaults("not a table")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid defaults type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupLLMLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// BenchmarkLLMOperations benchmarks basic LLM operations
func BenchmarkLLMOperations(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	// Setup LLM library for benchmarking
	mockBridge := NewMockLLMBridge()
	setupMockBridges(L, mockBridge)

	// Load promise library (required dependency)
	promisePath := filepath.Join(".", "promise.lua")
	err := L.DoFile(promisePath)
	if err != nil {
		b.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)

	// Load LLM library
	llmPath := filepath.Join(".", "llm.lua")
	err = L.DoFile(llmPath)
	if err != nil {
		b.Fatalf("Failed to load LLM library: %v", err)
	}
	llmLib := L.Get(-1)
	L.SetGlobal("llm", llmLib)

	script := `
		local response = llm.quick_prompt("Benchmark test")
		return response
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := L.DoString(script)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
		L.Pop(1) // Clean stack
	}
}

// TestLLMPackageRequire tests that the LLM module can be required as a package
func TestLLMPackageRequire(t *testing.T) {
	// Change to the stdlib directory for testing
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	err = os.Chdir(filepath.Dir(filepath.Join(wd, "llm.lua")))
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	L := lua.NewState()
	defer L.Close()

	// Setup mock bridges and promise dependency
	mockBridge := NewMockLLMBridge()
	setupMockBridges(L, mockBridge)

	// Load promise library first
	promisePath := filepath.Join(".", "promise.lua")
	err = L.DoFile(promisePath)
	if err != nil {
		t.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)

	script := `
		local llm = require('llm')
		
		-- Test that the module loads correctly
		if type(llm) ~= "table" then
			error("LLM module should return a table")
		end
		
		if type(llm.quick_prompt) ~= "function" then
			error("quick_prompt function should be available")
		end
		
		if type(llm.chat_session) ~= "function" then
			error("chat_session function should be available")
		end
		
		return true
	`

	err = L.DoString(script)
	if err != nil {
		t.Fatalf("Package require test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected require test to pass, got %v", result)
	}
}
