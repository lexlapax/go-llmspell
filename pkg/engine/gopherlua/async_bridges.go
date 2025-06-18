// ABOUTME: Async bridge method wrappers for non-blocking operations in GopherLua engine
// ABOUTME: Provides promisification, streaming support, progress callbacks, and cancellation tokens

package gopherlua

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

// AsyncBridgeWrapper wraps a bridge to provide async method execution
type AsyncBridgeWrapper struct {
	bridge     engine.Bridge
	runtime    *AsyncRuntime
	channelMgr *ChannelManager
}

// Stream represents a stream of values from an async operation
type Stream struct {
	channelID string
	manager   *ChannelManager
	closed    bool
	mu        sync.Mutex
}

// CancellationToken provides cancellation support for async operations
type CancellationToken struct {
	id        string
	ctx       context.Context
	cancel    context.CancelFunc
	cancelled bool
	mu        sync.RWMutex
}

// ProgressCallback is called to report progress of async operations
type ProgressCallback func(progress float64)

// NewAsyncBridgeWrapper creates a new async wrapper for a bridge
func NewAsyncBridgeWrapper(b engine.Bridge, runtime *AsyncRuntime, channelMgr *ChannelManager) (*AsyncBridgeWrapper, error) {
	if b == nil {
		return nil, fmt.Errorf("bridge cannot be nil")
	}
	if runtime == nil {
		return nil, fmt.Errorf("async runtime cannot be nil")
	}
	if channelMgr == nil {
		return nil, fmt.Errorf("channel manager cannot be nil")
	}

	return &AsyncBridgeWrapper{
		bridge:     b,
		runtime:    runtime,
		channelMgr: channelMgr,
	}, nil
}

// Bridge interface delegation
func (w *AsyncBridgeWrapper) GetID() string {
	return w.bridge.GetID()
}

func (w *AsyncBridgeWrapper) GetMetadata() engine.BridgeMetadata {
	return w.bridge.GetMetadata()
}

func (w *AsyncBridgeWrapper) Initialize(ctx context.Context) error {
	return w.bridge.Initialize(ctx)
}

func (w *AsyncBridgeWrapper) Cleanup(ctx context.Context) error {
	return w.bridge.Cleanup(ctx)
}

func (w *AsyncBridgeWrapper) Methods() []engine.MethodInfo {
	return w.bridge.Methods()
}

func (w *AsyncBridgeWrapper) TypeMappings() map[string]engine.TypeMapping {
	return w.bridge.TypeMappings()
}

func (w *AsyncBridgeWrapper) IsInitialized() bool {
	return w.bridge.IsInitialized()
}

func (w *AsyncBridgeWrapper) RegisterWithEngine(eng engine.ScriptEngine) error {
	return w.bridge.RegisterWithEngine(eng)
}

func (w *AsyncBridgeWrapper) RequiredPermissions() []engine.Permission {
	return w.bridge.RequiredPermissions()
}

func (w *AsyncBridgeWrapper) ValidateMethod(method string, args []engine.ScriptValue) error {
	return w.bridge.ValidateMethod(method, args)
}

func (w *AsyncBridgeWrapper) ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	return w.bridge.ExecuteMethod(ctx, method, args)
}

// ExecuteMethodAsync executes a bridge method asynchronously and returns a promise
func (w *AsyncBridgeWrapper) ExecuteMethodAsync(ctx context.Context, L *lua.LState, method string, args []engine.ScriptValue) (*Promise, error) {
	// Create an empty promise that we'll resolve manually
	promise, err := w.runtime.CreateEmptyPromise(ctx)
	if err != nil {
		return nil, err
	}

	// Execute bridge method asynchronously
	go func() {
		// Execute the bridge method
		result, err := w.bridge.ExecuteMethod(ctx, method, args)

		// Convert result to Lua value
		var luaResult lua.LValue
		if err != nil {
			luaResult = lua.LNil
		} else {
			luaResult = ScriptValueToLValue(result)
		}

		// Set the result for the promise
		w.runtime.SetCoroutineResult(promise.GetCoroID(), luaResult, err)
	}()

	return promise, nil
}

// ExecuteMethodStream executes a method that returns a stream of values
func (w *AsyncBridgeWrapper) ExecuteMethodStream(ctx context.Context, L *lua.LState, method string, args []engine.ScriptValue) (*Stream, error) {
	// Execute the method
	result, err := w.bridge.ExecuteMethod(ctx, method, args)
	if err != nil {
		return nil, fmt.Errorf("stream method execution failed: %w", err)
	}

	// Check if result is a channel
	channelValue, ok := result.(engine.ChannelValue)
	if !ok {
		return nil, fmt.Errorf("method %s did not return a channel", method)
	}

	// Create a Lua channel for streaming
	channelID, err := w.channelMgr.CreateChannel(L, 10) // Buffered for performance
	if err != nil {
		return nil, fmt.Errorf("failed to create stream channel: %w", err)
	}

	// Start goroutine to forward values from engine channel to Lua channel
	go func() {
		defer func() { _ = w.channelMgr.CloseChannel(channelID) }()

		// Type assert the channel to a concrete channel type
		ch, ok := channelValue.Value().(chan engine.ScriptValue)
		if !ok {
			// Try chan interface{} as fallback
			if chInterface, ok := channelValue.Value().(chan interface{}); ok {
				for val := range chInterface {
					// Convert interface{} to ScriptValue
					var scriptVal engine.ScriptValue
					if sv, ok := val.(engine.ScriptValue); ok {
						scriptVal = sv
					} else {
						// Wrap in appropriate ScriptValue type
						switch v := val.(type) {
						case string:
							scriptVal = engine.NewStringValue(v)
						case float64:
							scriptVal = engine.NewNumberValue(v)
						case bool:
							scriptVal = engine.NewBoolValue(v)
						default:
							scriptVal = engine.NewNilValue()
						}
					}
					luaValue := ScriptValueToLValue(scriptVal)
					err := w.channelMgr.Send(ctx, channelID, luaValue)
					if err != nil {
						// Context cancelled or channel closed
						return
					}
				}
				return
			}
			return
		}

		for value := range ch {
			luaValue := ScriptValueToLValue(value)
			err := w.channelMgr.Send(ctx, channelID, luaValue)
			if err != nil {
				// Context cancelled or channel closed
				return
			}
		}
	}()

	return &Stream{
		channelID: channelID,
		manager:   w.channelMgr,
	}, nil
}

// ExecuteMethodAsyncWithProgress executes a method with progress reporting
func (w *AsyncBridgeWrapper) ExecuteMethodAsyncWithProgress(ctx context.Context, L *lua.LState, method string, args []engine.ScriptValue, progressCb ProgressCallback) (*Promise, error) {
	// Create the base promise
	promise, err := w.ExecuteMethodAsync(ctx, L, method, args)
	if err != nil {
		return nil, err
	}

	// Start progress reporting in a separate goroutine
	go func() {
		ticker := time.NewTicker(25 * time.Millisecond)
		defer ticker.Stop()

		start := time.Now()
		estimatedDuration := 100 * time.Millisecond // Default estimate

		for {
			select {
			case <-ticker.C:
				// Check if promise is resolved
				if promise.IsResolved() {
					progressCb(1.0) // 100% complete
					return
				}

				elapsed := time.Since(start)
				progress := float64(elapsed) / float64(estimatedDuration)
				if progress > 0.99 {
					progress = 0.99 // Cap at 99% until done
				}
				progressCb(progress)
			case <-ctx.Done():
				return
			}
		}
	}()

	return promise, nil
}

// CreateCancellationToken creates a new cancellation token
func (w *AsyncBridgeWrapper) CreateCancellationToken() *CancellationToken {
	ctx, cancel := context.WithCancel(context.Background())
	return &CancellationToken{
		id:     uuid.New().String(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// ExecuteMethodAsyncWithToken executes a method with cancellation token support
func (w *AsyncBridgeWrapper) ExecuteMethodAsyncWithToken(ctx context.Context, L *lua.LState, method string, args []engine.ScriptValue, token *CancellationToken) (*Promise, error) {
	// Combine contexts - either the provided context or token can cancel
	combinedCtx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-token.ctx.Done():
			cancel()
		}
	}()

	return w.ExecuteMethodAsync(combinedCtx, L, method, args)
}

// AwaitAll waits for all promises to resolve
func (w *AsyncBridgeWrapper) AwaitAll(ctx context.Context, promises ...*Promise) ([]lua.LValue, error) {
	results := make([]lua.LValue, len(promises))
	errors := make([]error, len(promises))
	var wg sync.WaitGroup

	for i, promise := range promises {
		wg.Add(1)
		go func(index int, p *Promise) {
			defer wg.Done()
			result, err := p.Await(ctx)
			results[index] = result
			errors[index] = err
		}(i, promise)
	}

	wg.Wait()

	// Check for any errors
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("promise %d failed: %w", i, err)
		}
	}

	return results, nil
}

// AwaitRace waits for the first promise to resolve
func (w *AsyncBridgeWrapper) AwaitRace(ctx context.Context, promises ...*Promise) (lua.LValue, int, error) {
	if len(promises) == 0 {
		return lua.LNil, -1, fmt.Errorf("no promises provided")
	}

	type result struct {
		value lua.LValue
		index int
		err   error
	}

	resultCh := make(chan result, len(promises))

	for i, promise := range promises {
		go func(index int, p *Promise) {
			value, err := p.Await(ctx)
			select {
			case resultCh <- result{value: value, index: index, err: err}:
			case <-ctx.Done():
			}
		}(i, promise)
	}

	select {
	case res := <-resultCh:
		return res.value, res.index, res.err
	case <-ctx.Done():
		return lua.LNil, -1, ctx.Err()
	}
}

// Stream methods

// Next returns the next value from the stream
func (s *Stream) Next(ctx context.Context) (lua.LValue, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return lua.LNil, fmt.Errorf("stream is closed")
	}
	s.mu.Unlock()

	return s.manager.Receive(ctx, s.channelID)
}

// Close closes the stream
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	return s.manager.CloseChannel(s.channelID)
}

// CancellationToken methods

// Cancel cancels the token
func (t *CancellationToken) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.cancelled {
		t.cancelled = true
		t.cancel()
	}
}

// IsCancelled checks if the token is cancelled
func (t *CancellationToken) IsCancelled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cancelled
}

// Context returns the context associated with the token
func (t *CancellationToken) Context() context.Context {
	return t.ctx
}

// GetID returns the token ID
func (t *CancellationToken) GetID() string {
	return t.id
}

// Helper function to convert ScriptValue to LValue
func ScriptValueToLValue(sv engine.ScriptValue) lua.LValue {
	if sv == nil {
		return lua.LNil
	}

	switch v := sv.(type) {
	case engine.StringValue:
		return lua.LString(v.Value())
	case engine.NumberValue:
		return lua.LNumber(v.Value())
	case engine.BoolValue:
		return lua.LBool(v.Value())
	case engine.ArrayValue:
		table := &lua.LTable{}
		for i, elem := range v.Elements() {
			table.RawSetInt(i+1, ScriptValueToLValue(elem))
		}
		return table
	case engine.ObjectValue:
		table := &lua.LTable{}
		for k, val := range v.Fields() {
			table.RawSetString(k, ScriptValueToLValue(val))
		}
		return table
	default:
		return lua.LNil
	}
}
