// ABOUTME: Test helpers for the engine package to avoid import cycles with testutils
// ABOUTME: Provides mock implementations that can be used within engine tests

// NOTE: This file contains test helpers specific to the engine package.
// The centralized testutils package cannot be used here due to import cycles
// (testutils imports engine, so engine tests cannot import testutils).
// These mocks are simplified versions for use within engine package tests only.
// For tests outside the engine package, use the full testutils.MockScriptEngine.

// Common helper functions to reduce duplication

package engine

import (
	"context"
	"errors"
	"sync"
	"time"
)

// testMockScriptEngine provides a simple mock implementation for engine tests
type testMockScriptEngine struct {
	mu             sync.RWMutex
	name           string
	version        string
	initialized    bool
	bridges        map[string]Bridge
	memoryLimit    int64
	timeout        time.Duration
	resourceLimits ResourceLimits
	metrics        EngineMetrics
	contexts       map[string]ScriptContext

	// Test configuration
	executeFunc     func(ctx context.Context, script string, params map[string]interface{}) (ScriptValue, error)
	executeFileFunc func(ctx context.Context, path string, params map[string]interface{}) (ScriptValue, error)
	shutdownError   error
}

// newTestMockScriptEngine creates a new mock engine for testing
func newTestMockScriptEngine(name string) *testMockScriptEngine {
	return &testMockScriptEngine{
		name:     name,
		version:  "1.0.0",
		bridges:  make(map[string]Bridge),
		contexts: make(map[string]ScriptContext),
	}
}

// withExecuteFunc sets a custom execute function
func (m *testMockScriptEngine) withExecuteFunc(f func(ctx context.Context, script string, params map[string]interface{}) (ScriptValue, error)) *testMockScriptEngine {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executeFunc = f
	return m
}

func (m *testMockScriptEngine) Initialize(config EngineConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return errors.New("already initialized")
	}

	m.initialized = true
	m.memoryLimit = config.MemoryLimit
	m.timeout = config.TimeoutLimit
	return nil
}

func (m *testMockScriptEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (ScriptValue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	if m.executeFunc != nil {
		return m.executeFunc(ctx, script, params)
	}

	return NewStringValue("executed: " + script), nil
}

func (m *testMockScriptEngine) ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (ScriptValue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	if m.executeFileFunc != nil {
		return m.executeFileFunc(ctx, path, params)
	}

	return NewStringValue("executed file: " + path), nil
}

func (m *testMockScriptEngine) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	if m.shutdownError != nil {
		return m.shutdownError
	}

	m.initialized = false
	return nil
}

func (m *testMockScriptEngine) RegisterBridge(bridge Bridge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := bridge.GetID()
	if _, exists := m.bridges[id]; exists {
		return errors.New("bridge already registered")
	}

	m.bridges[id] = bridge
	return nil
}

func (m *testMockScriptEngine) UnregisterBridge(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.bridges[name]; !exists {
		return errors.New("bridge not found")
	}

	delete(m.bridges, name)
	return nil
}

func (m *testMockScriptEngine) GetBridge(name string) (Bridge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bridge, exists := m.bridges[name]
	if !exists {
		return nil, errors.New("bridge not found")
	}
	return bridge, nil
}

func (m *testMockScriptEngine) ListBridges() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.bridges))
	for name := range m.bridges {
		names = append(names, name)
	}
	return names
}

func (m *testMockScriptEngine) ToNative(scriptValue ScriptValue) (interface{}, error) {
	if scriptValue == nil {
		return nil, nil
	}
	return scriptValue.ToGo(), nil
}

func (m *testMockScriptEngine) FromNative(goValue interface{}) (ScriptValue, error) {
	switch v := goValue.(type) {
	case nil:
		return NewNilValue(), nil
	case bool:
		return NewBoolValue(v), nil
	case string:
		return NewStringValue(v), nil
	case int:
		return NewNumberValue(float64(v)), nil
	case int8:
		return NewNumberValue(float64(v)), nil
	case int16:
		return NewNumberValue(float64(v)), nil
	case int32:
		return NewNumberValue(float64(v)), nil
	case int64:
		return NewNumberValue(float64(v)), nil
	case uint:
		return NewNumberValue(float64(v)), nil
	case uint8:
		return NewNumberValue(float64(v)), nil
	case uint16:
		return NewNumberValue(float64(v)), nil
	case uint32:
		return NewNumberValue(float64(v)), nil
	case uint64:
		return NewNumberValue(float64(v)), nil
	case float32:
		return NewNumberValue(float64(v)), nil
	case float64:
		return NewNumberValue(v), nil
	default:
		return NewCustomValue("unknown", v), nil
	}
}

func (m *testMockScriptEngine) Name() string {
	return m.name
}

func (m *testMockScriptEngine) Version() string {
	return m.version
}

func (m *testMockScriptEngine) FileExtensions() []string {
	return []string{"mock", "test"}
}

func (m *testMockScriptEngine) Features() []EngineFeature {
	return []EngineFeature{FeatureAsync, FeatureDebugging}
}

func (m *testMockScriptEngine) SetMemoryLimit(bytes int64) error {
	if bytes < 0 {
		return errors.New("invalid memory limit")
	}
	m.memoryLimit = bytes
	return nil
}

func (m *testMockScriptEngine) SetTimeout(duration time.Duration) error {
	if duration < 0 {
		return errors.New("invalid timeout")
	}
	m.timeout = duration
	return nil
}

func (m *testMockScriptEngine) SetResourceLimits(limits ResourceLimits) error {
	m.resourceLimits = limits
	return nil
}

func (m *testMockScriptEngine) GetMetrics() EngineMetrics {
	return m.metrics
}

func (m *testMockScriptEngine) CreateContext(options ContextOptions) (ScriptContext, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := options.ID
	if id == "" {
		id = "ctx-" + time.Now().Format("20060102150405")
	}
	ctx := &testMockScriptContext{
		id:        id,
		variables: make(map[string]interface{}),
	}
	// Initialize with provided variables
	for k, v := range options.Variables {
		ctx.variables[k] = v
	}
	m.contexts[ctx.ID()] = ctx
	return ctx, nil
}

func (m *testMockScriptEngine) DestroyContext(ctx ScriptContext) error {
	delete(m.contexts, ctx.ID())
	return nil
}

func (m *testMockScriptEngine) ExecuteScript(ctx context.Context, script string, options ExecutionOptions) (*ExecutionResult, error) {
	start := time.Now()
	val, err := m.Execute(ctx, script, options.Variables)
	result := &ExecutionResult{
		Value:    val,
		Duration: time.Since(start),
		Metadata: make(map[string]interface{}),
	}
	if err != nil {
		result.Error = err
	}
	return result, nil
}

func (m *testMockScriptEngine) GetEventBus() EventBus {
	return NewDefaultEventBus()
}

func (m *testMockScriptEngine) RegisterTypeConverter(fromType, toType string, converter TypeConverterFunc) error {
	return nil
}

func (m *testMockScriptEngine) GetTypeRegistry() TypeRegistry {
	return NewDefaultTypeRegistry()
}

func (m *testMockScriptEngine) EnableProfiling(config ProfilingConfig) error {
	return nil
}

func (m *testMockScriptEngine) DisableProfiling() error {
	return nil
}

func (m *testMockScriptEngine) GetProfilingReport() (*ProfilingReport, error) {
	return &ProfilingReport{}, nil
}

func (m *testMockScriptEngine) ExportAPI(format ExportFormat) ([]byte, error) {
	return []byte("{}"), nil
}

func (m *testMockScriptEngine) GenerateClientLibrary(language string, options ClientLibraryOptions) ([]byte, error) {
	return []byte("{}"), nil
}

// testMockScriptContext provides a simple context implementation
type testMockScriptContext struct {
	id        string
	variables map[string]interface{}
}

func (m *testMockScriptContext) ID() string {
	return m.id
}

func (m *testMockScriptContext) SetVariable(name string, value interface{}) error {
	m.variables[name] = value
	return nil
}

func (m *testMockScriptContext) GetVariable(name string) (interface{}, error) {
	val, exists := m.variables[name]
	if !exists {
		return nil, errors.New("variable not found")
	}
	return val, nil
}

func (m *testMockScriptContext) Execute(script string) (interface{}, error) {
	return "context executed: " + script, nil
}

func (m *testMockScriptContext) Destroy() error {
	m.variables = nil
	return nil
}

// testMockBridge provides a simple mock bridge for testing the engine package
type testMockBridge struct {
	mu           sync.RWMutex
	id           string
	metadata     BridgeMetadata
	initialized  bool
	methods      []MethodInfo
	typeMappings map[string]TypeMapping
	permissions  []Permission
	executeFunc  func(ctx context.Context, method string, args []ScriptValue) (ScriptValue, error)
}

func newTestMockBridge(id string) *testMockBridge {
	return &testMockBridge{
		id: id,
		metadata: BridgeMetadata{
			Name:        id,
			Version:     "1.0.0",
			Description: "Mock bridge for testing",
		},
		methods: []MethodInfo{
			{
				Name:        "testMethod",
				Description: "A test method",
				Parameters: []ParameterInfo{
					{
						Name:        "input",
						Type:        "string",
						Required:    true,
						Description: "Input parameter",
					},
				},
				ReturnType: "string",
			},
		},
		typeMappings: map[string]TypeMapping{
			"string": {
				GoType:     "string",
				ScriptType: "string",
				Converter:  "direct",
			},
		},
		permissions: []Permission{
			{
				Type:     PermissionNetwork,
				Resource: "http://api.example.com",
				Actions:  []string{"GET", "POST"},
			},
		},
	}
}

func (b *testMockBridge) GetID() string {
	return b.id
}

func (b *testMockBridge) GetMetadata() BridgeMetadata {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.metadata
}

func (b *testMockBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.initialized {
		return errors.New("already initialized")
	}
	b.initialized = true
	return nil
}

func (b *testMockBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.initialized {
		return errors.New("not initialized")
	}
	b.initialized = false
	return nil
}

func (b *testMockBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

func (b *testMockBridge) RegisterWithEngine(engine ScriptEngine) error {
	return engine.RegisterBridge(b)
}

func (b *testMockBridge) Methods() []MethodInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.methods
}

func (b *testMockBridge) TypeMappings() map[string]TypeMapping {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.typeMappings
}

func (b *testMockBridge) ValidateMethod(name string, args []ScriptValue) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, method := range b.methods {
		if method.Name == name {
			return nil
		}
	}
	return errors.New("method not found")
}

func (b *testMockBridge) ExecuteMethod(ctx context.Context, name string, args []ScriptValue) (ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, errors.New("bridge not initialized")
	}

	if b.executeFunc != nil {
		return b.executeFunc(ctx, name, args)
	}

	// Default implementation
	for _, method := range b.methods {
		if method.Name == name {
			if len(args) > 0 {
				return NewStringValue("result: " + args[0].String()), nil
			}
			return NewStringValue("result: no args"), nil
		}
	}
	return nil, errors.New("method not found")
}

func (b *testMockBridge) RequiredPermissions() []Permission {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.permissions
}

// Common helper functions for engine tests

// createTestArgs creates a common set of test arguments for validation testing
func createTestArgs() []ScriptValue {
	return []ScriptValue{
		NewStringValue("hello"),
		NewNumberValue(42),
		NewBoolValue(true),
		NewNilValue(),
	}
}

// createTestObject creates a test object with standard fields
func createTestObject() ScriptValue {
	return NewObjectValue(map[string]ScriptValue{
		"name":   NewStringValue("test"),
		"age":    NewNumberValue(25),
		"active": NewBoolValue(true),
	})
}

// createTestArray creates a test array with mixed types
func createTestArray() ScriptValue {
	return NewArrayValue([]ScriptValue{
		NewStringValue("hello"),
		NewNumberValue(42),
		NewBoolValue(true),
	})
}

// createMixedTypeArray creates an array with different types for testing
func createMixedTypeArray() ScriptValue {
	return NewArrayValue([]ScriptValue{
		NewNumberValue(1),
		NewStringValue("two"),
		NewBoolValue(true),
	})
}
