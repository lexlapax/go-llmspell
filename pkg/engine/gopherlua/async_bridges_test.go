// ABOUTME: Tests for async bridge methods in GopherLua engine
// ABOUTME: Tests async wrapping, promisification, streaming, and cancellation for bridge operations

package gopherlua

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

// Mock bridge for testing async operations
type mockAsyncBridge struct {
	id          string
	initialized bool
	methods     map[string]bridge.MethodInfo
}

func newMockAsyncBridge() *mockAsyncBridge {
	return &mockAsyncBridge{
		id: "mockAsync",
		methods: map[string]bridge.MethodInfo{
			"fastMethod": {
				Description: "A fast method that returns immediately",
				Parameters:  []bridge.ParameterInfo{{Name: "input", Type: "string"}},
				Returns:     []bridge.ReturnInfo{{Type: "string"}},
			},
			"slowMethod": {
				Description: "A slow method that takes time",
				Parameters:  []bridge.ParameterInfo{{Name: "delay", Type: "number"}},
				Returns:     []bridge.ReturnInfo{{Type: "string"}},
			},
			"streamMethod": {
				Description: "A method that streams data",
				Parameters:  []bridge.ParameterInfo{{Name: "count", Type: "number"}},
				Returns:     []bridge.ReturnInfo{{Type: "channel"}},
			},
			"errorMethod": {
				Description: "A method that returns an error",
				Parameters:  []bridge.ParameterInfo{},
				Returns:     []bridge.ReturnInfo{{Type: "error"}},
			},
		},
	}
}

func (m *mockAsyncBridge) GetID() string                                                 { return m.id }
func (m *mockAsyncBridge) GetMetadata() bridge.BridgeMetadata                            { return bridge.BridgeMetadata{} }
func (m *mockAsyncBridge) Initialize(ctx context.Context) error                          { m.initialized = true; return nil }
func (m *mockAsyncBridge) Cleanup(ctx context.Context) error                             { m.initialized = false; return nil }
func (m *mockAsyncBridge) GetMethods() map[string]bridge.MethodInfo                      { return m.methods }
func (m *mockAsyncBridge) GetTypeMappings() map[string]bridge.TypeMapping                { return nil }
func (m *mockAsyncBridge) IsInitialized() bool                                           { return m.initialized }
func (m *mockAsyncBridge) GetRequiredPermissions() []string                              { return nil }
func (m *mockAsyncBridge) ValidateMethod(method string, args []engine.ScriptValue) error { return nil }

func (m *mockAsyncBridge) ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	switch method {
	case "fastMethod":
		if len(args) > 0 {
			return engine.NewStringValue("result: " + args[0].String()), nil
		}
		return engine.NewStringValue("result: no input"), nil

	case "slowMethod":
		delay := 100 * time.Millisecond
		if len(args) > 0 {
			if num, ok := args[0].(engine.NumberValue); ok {
				delay = time.Duration(num.Value()) * time.Millisecond
			}
		}
		select {
		case <-time.After(delay):
			return engine.NewStringValue("slow result"), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	case "streamMethod":
		count := 5
		if len(args) > 0 {
			if num, ok := args[0].(engine.NumberValue); ok {
				count = int(num.Value())
			}
		}
		// Return a channel value for streaming
		ch := make(chan engine.ScriptValue, count)
		go func() {
			defer close(ch)
			for i := 0; i < count; i++ {
				select {
				case ch <- engine.NewNumberValue(float64(i)):
				case <-ctx.Done():
					return
				}
			}
		}()
		return engine.NewChannelValue("stream-channel", ch), nil

	case "errorMethod":
		return nil, fmt.Errorf("intentional error")

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func TestAsyncBridgeWrapper_Creation(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	mockBridge := newMockAsyncBridge()

	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	if wrapper == nil {
		t.Fatalf("NewAsyncBridgeWrapper returned nil")
	}

	// Verify wrapper properties
	if wrapper.GetID() != mockBridge.GetID() {
		t.Errorf("Expected ID %s, got %s", mockBridge.GetID(), wrapper.GetID())
	}

	// Check that methods are wrapped
	methods := wrapper.GetMethods()
	if len(methods) != len(mockBridge.methods) {
		t.Errorf("Expected %d methods, got %d", len(mockBridge.methods), len(methods))
	}
}

func TestAsyncBridgeWrapper_Promisification(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	ctx := context.Background()

	// Test fast method promisification
	args := []engine.ScriptValue{engine.NewStringValue("test")}
	promise, err := wrapper.ExecuteMethodAsync(ctx, L, "fastMethod", args)
	if err != nil {
		t.Fatalf("ExecuteMethodAsync failed: %v", err)
	}

	if promise == nil {
		t.Fatalf("ExecuteMethodAsync returned nil promise")
	}

	// Await the promise
	result, err := promise.Await(ctx)
	if err != nil {
		t.Errorf("Promise.Await failed: %v", err)
	}

	if result == nil || result.Type() != lua.LTString {
		t.Errorf("Expected string result, got %v", result)
	}
}

func TestAsyncBridgeWrapper_SlowMethodWithTimeout(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	// Test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Execute slow method that takes 200ms
	args := []engine.ScriptValue{engine.NewNumberValue(200)}
	promise, err := wrapper.ExecuteMethodAsync(ctx, L, "slowMethod", args)
	if err != nil {
		t.Fatalf("ExecuteMethodAsync failed: %v", err)
	}

	// Should timeout
	_, err = promise.Await(ctx)
	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}

func TestAsyncBridgeWrapper_StreamingSupport(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	ctx := context.Background()

	// Test streaming method
	args := []engine.ScriptValue{engine.NewNumberValue(3)}
	stream, err := wrapper.ExecuteMethodStream(ctx, L, "streamMethod", args)
	if err != nil {
		t.Fatalf("ExecuteMethodStream failed: %v", err)
	}

	if stream == nil {
		t.Fatalf("ExecuteMethodStream returned nil stream")
	}

	// Collect stream values
	var values []float64
	for {
		value, err := stream.Next(ctx)
		if err != nil {
			break
		}
		if num, ok := value.(lua.LNumber); ok {
			values = append(values, float64(num))
		}
	}

	// Should have received 3 values
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	// Values should be 0, 1, 2
	for i, v := range values {
		if v != float64(i) {
			t.Errorf("Expected value %d, got %f", i, v)
		}
	}
}

func TestAsyncBridgeWrapper_ProgressCallbacks(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	ctx := context.Background()

	// Track progress callbacks with thread-safe access
	var mu sync.Mutex
	var progressUpdates []float64
	progressCallback := func(progress float64) {
		mu.Lock()
		progressUpdates = append(progressUpdates, progress)
		mu.Unlock()
	}

	// Execute with progress tracking
	args := []engine.ScriptValue{engine.NewNumberValue(100)}
	promise, err := wrapper.ExecuteMethodAsyncWithProgress(ctx, L, "slowMethod", args, progressCallback)
	if err != nil {
		t.Fatalf("ExecuteMethodAsyncWithProgress failed: %v", err)
	}

	// Wait for completion
	_, err = promise.Await(ctx)
	if err != nil {
		t.Errorf("Promise.Await failed: %v", err)
	}

	// Should have received progress updates
	mu.Lock()
	numUpdates := len(progressUpdates)
	mu.Unlock()

	if numUpdates == 0 {
		t.Errorf("Expected progress updates, got none")
	}
}

func TestAsyncBridgeWrapper_CancellationTokens(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	// Create cancellation token
	token := wrapper.CreateCancellationToken()
	if token == nil {
		t.Fatalf("CreateCancellationToken returned nil")
	}

	// Execute slow method with token
	args := []engine.ScriptValue{engine.NewNumberValue(500)}
	promise, err := wrapper.ExecuteMethodAsyncWithToken(token.Context(), L, "slowMethod", args, token)
	if err != nil {
		t.Fatalf("ExecuteMethodAsyncWithToken failed: %v", err)
	}

	// Cancel after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		token.Cancel()
	}()

	// Should be cancelled
	_, err = promise.Await(context.Background())
	if err == nil {
		t.Errorf("Expected cancellation error, got nil")
	}

	if !token.IsCancelled() {
		t.Errorf("Token should be cancelled")
	}
}

func TestAsyncBridgeWrapper_ErrorHandling(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	ctx := context.Background()

	// Test error method
	promise, err := wrapper.ExecuteMethodAsync(ctx, L, "errorMethod", nil)
	if err != nil {
		t.Fatalf("ExecuteMethodAsync failed: %v", err)
	}

	// Should get error from promise
	_, err = promise.Await(ctx)
	if err == nil {
		t.Errorf("Expected error from promise, got nil")
	}

	if err.Error() != "intentional error" {
		t.Errorf("Expected 'intentional error', got %v", err)
	}
}

func TestAsyncBridgeWrapper_MultiplePromises(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	ctx := context.Background()

	// Create multiple promises
	var promises []*Promise
	for i := 0; i < 5; i++ {
		args := []engine.ScriptValue{engine.NewStringValue(fmt.Sprintf("test%d", i))}
		promise, err := wrapper.ExecuteMethodAsync(ctx, L, "fastMethod", args)
		if err != nil {
			t.Fatalf("ExecuteMethodAsync %d failed: %v", i, err)
		}
		promises = append(promises, promise)
	}

	// Wait for all promises
	results, err := wrapper.AwaitAll(ctx, promises...)
	if err != nil {
		t.Errorf("AwaitAll failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	// Check results
	for i, result := range results {
		expected := fmt.Sprintf("result: test%d", i)
		if str, ok := result.(lua.LString); ok {
			if string(str) != expected {
				t.Errorf("Result %d: expected %s, got %s", i, expected, str)
			}
		} else {
			t.Errorf("Result %d is not a string", i)
		}
	}
}

func TestAsyncBridgeWrapper_PromiseRace(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	ctx := context.Background()

	// Create promises with different delays
	promise1, _ := wrapper.ExecuteMethodAsync(ctx, L, "slowMethod", []engine.ScriptValue{engine.NewNumberValue(100)})
	promise2, _ := wrapper.ExecuteMethodAsync(ctx, L, "slowMethod", []engine.ScriptValue{engine.NewNumberValue(50)})
	promise3, _ := wrapper.ExecuteMethodAsync(ctx, L, "slowMethod", []engine.ScriptValue{engine.NewNumberValue(200)})

	// Race promises - promise2 should win
	result, index, err := wrapper.AwaitRace(ctx, promise1, promise2, promise3)
	if err != nil {
		t.Errorf("AwaitRace failed: %v", err)
	}

	if index != 1 { // promise2 is at index 1
		t.Errorf("Expected promise at index 1 to win, got index %d", index)
	}

	if result == nil {
		t.Errorf("AwaitRace returned nil result")
	}
}
