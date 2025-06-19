// ABOUTME: MockBridge provides a unified mock implementation of the Bridge interface for testing
// ABOUTME: Consolidates mock bridge patterns from across the codebase with flexible method handlers

package testutils

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// MethodHandler defines a function that handles method execution for mock bridges
type MethodHandler func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error)

// MockBridge provides a configurable mock implementation of engine.Bridge
type MockBridge struct {
	mu          sync.RWMutex
	id          string
	metadata    engine.BridgeMetadata
	initialized bool
	cleanedUp   bool

	// Method handling
	methods        map[string]engine.MethodInfo
	methodHandlers map[string]MethodHandler
	validateFunc   func(method string, args []engine.ScriptValue) error

	// Dependencies and permissions
	dependencies []string
	permissions  []engine.Permission
	typeMappings map[string]engine.TypeMapping

	// Error injection
	initError     error
	cleanupError  error
	registerError error

	// Call tracking
	initCalls     int
	cleanupCalls  int
	methodCalls   map[string][]methodCall
	validateCalls []validateCall

	// Custom behavior functions
	initFunc    func(ctx context.Context) error
	cleanupFunc func(ctx context.Context) error
}

type methodCall struct {
	Method string
	Args   []engine.ScriptValue
	Result engine.ScriptValue
	Error  error
}

type validateCall struct {
	Method string
	Args   []engine.ScriptValue
	Error  error
}

// NewMockBridge creates a new mock bridge with the given ID
func NewMockBridge(id string) *MockBridge {
	return &MockBridge{
		id: id,
		metadata: engine.BridgeMetadata{
			Name:        id,
			Version:     "1.0.0",
			Description: "Mock bridge for testing",
			Author:      "testutils",
			License:     "MIT",
		},
		methods:        make(map[string]engine.MethodInfo),
		methodHandlers: make(map[string]MethodHandler),
		methodCalls:    make(map[string][]methodCall),
		typeMappings:   make(map[string]engine.TypeMapping),
		validateCalls:  make([]validateCall, 0),
		permissions:    make([]engine.Permission, 0),
		dependencies:   make([]string, 0),
	}
}

// WithMetadata sets custom metadata
func (b *MockBridge) WithMetadata(metadata engine.BridgeMetadata) *MockBridge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.metadata = metadata
	return b
}

// WithMethod adds a method to the bridge with a handler
func (b *MockBridge) WithMethod(name string, info engine.MethodInfo, handler MethodHandler) *MockBridge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.methods[name] = info
	if handler != nil {
		b.methodHandlers[name] = handler
	}
	return b
}

// WithMethodHandler adds or updates a method handler
func (b *MockBridge) WithMethodHandler(name string, handler MethodHandler) *MockBridge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.methodHandlers[name] = handler
	return b
}

// WithValidateFunc sets a custom validation function
func (b *MockBridge) WithValidateFunc(f func(method string, args []engine.ScriptValue) error) *MockBridge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.validateFunc = f
	return b
}

// WithDependencies sets bridge dependencies
func (b *MockBridge) WithDependencies(deps ...string) *MockBridge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dependencies = deps
	b.metadata.Dependencies = deps
	return b
}

// WithPermissions sets required permissions
func (b *MockBridge) WithPermissions(perms ...engine.Permission) *MockBridge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.permissions = perms
	return b
}

// WithInitError sets an error to be returned by Initialize
func (b *MockBridge) WithInitError(err error) *MockBridge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initError = err
	return b
}

// WithInitFunc sets a custom initialization function
func (b *MockBridge) WithInitFunc(f func(ctx context.Context) error) *MockBridge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initFunc = f
	return b
}

// GetMethodCalls returns all calls for a specific method
func (b *MockBridge) GetMethodCalls(method string) []methodCall {
	b.mu.RLock()
	defer b.mu.RUnlock()
	calls := b.methodCalls[method]
	result := make([]methodCall, len(calls))
	copy(result, calls)
	return result
}

// GetValidateCalls returns all validation calls
func (b *MockBridge) GetValidateCalls() []validateCall {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]validateCall, len(b.validateCalls))
	copy(result, b.validateCalls)
	return result
}

// GetInitCallCount returns the number of times Initialize was called
func (b *MockBridge) GetInitCallCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initCalls
}

// Bridge interface implementation

func (b *MockBridge) GetID() string {
	return b.id
}

func (b *MockBridge) GetMetadata() engine.BridgeMetadata {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.metadata
}

func (b *MockBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initCalls++

	if b.initError != nil {
		return b.initError
	}

	if b.initFunc != nil {
		return b.initFunc(ctx)
	}

	if b.initialized {
		return errors.New("already initialized")
	}

	b.initialized = true
	return nil
}

func (b *MockBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cleanupCalls++

	if b.cleanupError != nil {
		return b.cleanupError
	}

	if b.cleanupFunc != nil {
		return b.cleanupFunc(ctx)
	}

	b.cleanedUp = true
	b.initialized = false
	return nil
}

func (b *MockBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

func (b *MockBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.registerError != nil {
		return b.registerError
	}

	return engine.RegisterBridge(b)
}

func (b *MockBridge) Methods() []engine.MethodInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	methods := make([]engine.MethodInfo, 0, len(b.methods))
	for _, method := range b.methods {
		methods = append(methods, method)
	}
	return methods
}

func (b *MockBridge) ValidateMethod(method string, args []engine.ScriptValue) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Record the call
	call := validateCall{
		Method: method,
		Args:   args,
	}

	// Use custom validation function if provided
	if b.validateFunc != nil {
		err := b.validateFunc(method, args)
		call.Error = err
		b.validateCalls = append(b.validateCalls, call)
		return err
	}

	// Check if method exists
	info, exists := b.methods[method]
	if !exists {
		err := fmt.Errorf("unknown method: %s", method)
		call.Error = err
		b.validateCalls = append(b.validateCalls, call)
		return err
	}

	// Basic parameter count validation
	if len(info.Parameters) > 0 && len(args) < len(info.Parameters) {
		// Count required parameters
		requiredCount := 0
		for _, param := range info.Parameters {
			if param.Required {
				requiredCount++
			}
		}

		if len(args) < requiredCount {
			err := fmt.Errorf("method %s requires at least %d arguments, got %d", method, requiredCount, len(args))
			call.Error = err
			b.validateCalls = append(b.validateCalls, call)
			return err
		}
	}

	b.validateCalls = append(b.validateCalls, call)
	return nil
}

func (b *MockBridge) ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.initialized {
		return nil, errors.New("bridge not initialized")
	}

	// Record the call
	call := methodCall{
		Method: method,
		Args:   args,
	}

	// Use custom handler if provided
	if handler, exists := b.methodHandlers[method]; exists {
		result, err := handler(ctx, args)
		call.Result = result
		call.Error = err

		if b.methodCalls[method] == nil {
			b.methodCalls[method] = make([]methodCall, 0)
		}
		b.methodCalls[method] = append(b.methodCalls[method], call)

		return result, err
	}

	// Check if method exists
	if _, exists := b.methods[method]; !exists {
		err := fmt.Errorf("unknown method: %s", method)
		call.Error = err

		if b.methodCalls[method] == nil {
			b.methodCalls[method] = make([]methodCall, 0)
		}
		b.methodCalls[method] = append(b.methodCalls[method], call)

		return nil, err
	}

	// Default behavior - return success
	result := engine.NewStringValue(fmt.Sprintf("%s result", method))
	call.Result = result

	if b.methodCalls[method] == nil {
		b.methodCalls[method] = make([]methodCall, 0)
	}
	b.methodCalls[method] = append(b.methodCalls[method], call)

	return result, nil
}

func (b *MockBridge) RequiredPermissions() []engine.Permission {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.permissions
}

func (b *MockBridge) TypeMappings() map[string]engine.TypeMapping {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.typeMappings
}

// MockAsyncBridge extends MockBridge with async-specific functionality
type MockAsyncBridge struct {
	*MockBridge
	asyncMethods map[string]bool
}

// NewMockAsyncBridge creates a new async-capable mock bridge
func NewMockAsyncBridge(id string) *MockAsyncBridge {
	return &MockAsyncBridge{
		MockBridge:   NewMockBridge(id),
		asyncMethods: make(map[string]bool),
	}
}

// WithAsyncMethod adds an async method to the bridge
func (b *MockAsyncBridge) WithAsyncMethod(name string, info engine.MethodInfo, handler MethodHandler) *MockAsyncBridge {
	b.WithMethod(name, info, handler)
	b.asyncMethods[name] = true
	return b
}

// IsAsyncMethod checks if a method is async
func (b *MockAsyncBridge) IsAsyncMethod(method string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.asyncMethods[method]
}
