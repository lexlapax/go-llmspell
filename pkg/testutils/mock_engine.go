// ABOUTME: MockScriptEngine provides a unified mock implementation of the ScriptEngine interface for testing
// ABOUTME: Consolidates mock engine patterns from across the codebase with builder pattern support

package testutils

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ExecuteCall records details of each Execute method call for verification
type ExecuteCall struct {
	Script string
	Params map[string]interface{}
	Result engine.ScriptValue
	Error  error
}

// MockScriptEngine provides a configurable mock implementation of engine.ScriptEngine
type MockScriptEngine struct {
	mu             sync.RWMutex
	name           string
	version        string
	initialized    bool
	bridges        map[string]engine.Bridge
	memoryLimit    int64
	timeout        time.Duration
	resourceLimits engine.ResourceLimits
	metrics        engine.EngineMetrics
	contexts       map[string]engine.ScriptContext

	// Test configuration
	executeCalls    []ExecuteCall
	executeFunc     func(ctx context.Context, script string, params map[string]interface{}) (engine.ScriptValue, error)
	executeFileFunc func(ctx context.Context, path string, params map[string]interface{}) (engine.ScriptValue, error)
	initializeFunc  func(config engine.EngineConfig) error
	shutdownFunc    func() error

	// Error injection
	initError     error
	executeError  error
	shutdownError error

	// State tracking
	shutdownCalled  bool
	registerCalls   []string
	unregisterCalls []string
}

// NewMockScriptEngine creates a new mock engine with default behavior
func NewMockScriptEngine() *MockScriptEngine {
	return &MockScriptEngine{
		name:            "mock-engine",
		version:         "1.0.0",
		bridges:         make(map[string]engine.Bridge),
		contexts:        make(map[string]engine.ScriptContext),
		executeCalls:    make([]ExecuteCall, 0),
		registerCalls:   make([]string, 0),
		unregisterCalls: make([]string, 0),
	}
}

// WithName sets the engine name
func (m *MockScriptEngine) WithName(name string) *MockScriptEngine {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.name = name
	return m
}

// WithVersion sets the engine version
func (m *MockScriptEngine) WithVersion(version string) *MockScriptEngine {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.version = version
	return m
}

// WithExecuteFunc sets a custom execute function
func (m *MockScriptEngine) WithExecuteFunc(f func(ctx context.Context, script string, params map[string]interface{}) (engine.ScriptValue, error)) *MockScriptEngine {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executeFunc = f
	return m
}

// WithExecuteFileFunc sets a custom execute file function
func (m *MockScriptEngine) WithExecuteFileFunc(f func(ctx context.Context, path string, params map[string]interface{}) (engine.ScriptValue, error)) *MockScriptEngine {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executeFileFunc = f
	return m
}

// WithInitializeFunc sets a custom initialize function
func (m *MockScriptEngine) WithInitializeFunc(f func(config engine.EngineConfig) error) *MockScriptEngine {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initializeFunc = f
	return m
}

// WithInitError sets an error to be returned by Initialize
func (m *MockScriptEngine) WithInitError(err error) *MockScriptEngine {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initError = err
	return m
}

// WithExecuteError sets an error to be returned by Execute
func (m *MockScriptEngine) WithExecuteError(err error) *MockScriptEngine {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executeError = err
	return m
}

// GetExecuteCalls returns all recorded execute calls
func (m *MockScriptEngine) GetExecuteCalls() []ExecuteCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	calls := make([]ExecuteCall, len(m.executeCalls))
	copy(calls, m.executeCalls)
	return calls
}

// GetRegisterCalls returns all bridge registration calls
func (m *MockScriptEngine) GetRegisterCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	calls := make([]string, len(m.registerCalls))
	copy(calls, m.registerCalls)
	return calls
}

// IsInitialized returns whether the engine has been initialized
func (m *MockScriptEngine) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// IsShutdown returns whether shutdown has been called
func (m *MockScriptEngine) IsShutdown() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shutdownCalled
}

// ScriptEngine interface implementation

func (m *MockScriptEngine) Initialize(config engine.EngineConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return errors.New("already initialized")
	}

	if m.initError != nil {
		return m.initError
	}

	if m.initializeFunc != nil {
		return m.initializeFunc(config)
	}

	m.initialized = true
	m.memoryLimit = config.MemoryLimit
	m.timeout = config.TimeoutLimit
	return nil
}

func (m *MockScriptEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (engine.ScriptValue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	// Record the call
	call := ExecuteCall{
		Script: script,
		Params: params,
	}

	// Use custom function if provided
	if m.executeFunc != nil {
		result, err := m.executeFunc(ctx, script, params)
		call.Result = result
		call.Error = err
		m.executeCalls = append(m.executeCalls, call)
		return result, err
	}

	// Return configured error
	if m.executeError != nil {
		call.Error = m.executeError
		m.executeCalls = append(m.executeCalls, call)
		return nil, m.executeError
	}

	// Default behavior
	result := engine.NewStringValue("executed: " + script)
	call.Result = result
	m.executeCalls = append(m.executeCalls, call)
	return result, nil
}

func (m *MockScriptEngine) ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (engine.ScriptValue, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	if m.executeFileFunc != nil {
		return m.executeFileFunc(ctx, path, params)
	}

	return engine.NewStringValue("executed file: " + path), nil
}

func (m *MockScriptEngine) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shutdownCalled {
		return errors.New("already shutdown")
	}

	if m.shutdownError != nil {
		return m.shutdownError
	}

	if m.shutdownFunc != nil {
		return m.shutdownFunc()
	}

	m.shutdownCalled = true
	m.initialized = false
	return nil
}

func (m *MockScriptEngine) RegisterBridge(bridge engine.Bridge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	id := bridge.GetID()
	if _, exists := m.bridges[id]; exists {
		return errors.New("bridge already registered: " + id)
	}

	m.bridges[id] = bridge
	m.registerCalls = append(m.registerCalls, id)
	return nil
}

func (m *MockScriptEngine) UnregisterBridge(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	if _, exists := m.bridges[name]; !exists {
		return errors.New("bridge not found: " + name)
	}

	delete(m.bridges, name)
	m.unregisterCalls = append(m.unregisterCalls, name)
	return nil
}

func (m *MockScriptEngine) GetBridge(name string) (engine.Bridge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bridge, exists := m.bridges[name]
	if !exists {
		return nil, errors.New("bridge not found: " + name)
	}
	return bridge, nil
}

func (m *MockScriptEngine) ListBridges() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.bridges))
	for name := range m.bridges {
		names = append(names, name)
	}
	return names
}

func (m *MockScriptEngine) ToNative(scriptValue engine.ScriptValue) (interface{}, error) {
	if scriptValue == nil {
		return nil, nil
	}

	switch v := scriptValue.(type) {
	case engine.StringValue:
		return v.Value(), nil
	case engine.NumberValue:
		return v.Value(), nil
	case engine.BoolValue:
		return v.Value(), nil
	case engine.NilValue:
		return nil, nil
	case engine.ObjectValue:
		result := make(map[string]interface{})
		for k, sv := range v.Fields() {
			native, err := m.ToNative(sv)
			if err != nil {
				return nil, err
			}
			result[k] = native
		}
		return result, nil
	case engine.ArrayValue:
		elements := v.Elements()
		result := make([]interface{}, len(elements))
		for i, elem := range elements {
			native, err := m.ToNative(elem)
			if err != nil {
				return nil, err
			}
			result[i] = native
		}
		return result, nil
	default:
		return nil, errors.New("unsupported script value type")
	}
}

func (m *MockScriptEngine) ToScriptValue(value interface{}) (engine.ScriptValue, error) {
	if value == nil {
		return engine.NewNilValue(), nil
	}

	switch v := value.(type) {
	case string:
		return engine.NewStringValue(v), nil
	case float64:
		return engine.NewNumberValue(v), nil
	case int:
		return engine.NewNumberValue(float64(v)), nil
	case bool:
		return engine.NewBoolValue(v), nil
	case map[string]interface{}:
		fields := make(map[string]engine.ScriptValue)
		for k, val := range v {
			sv, err := m.ToScriptValue(val)
			if err != nil {
				return nil, err
			}
			fields[k] = sv
		}
		return engine.NewObjectValue(fields), nil
	case []interface{}:
		elements := make([]engine.ScriptValue, len(v))
		for i, val := range v {
			sv, err := m.ToScriptValue(val)
			if err != nil {
				return nil, err
			}
			elements[i] = sv
		}
		return engine.NewArrayValue(elements), nil
	default:
		return nil, errors.New("unsupported native value type")
	}
}

func (m *MockScriptEngine) CreateContext(options engine.ContextOptions) (engine.ScriptContext, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	if _, exists := m.contexts[options.ID]; exists {
		return nil, errors.New("context already exists: " + options.ID)
	}

	ctx := &mockScriptContext{
		id:     options.ID,
		values: make(map[string]interface{}),
	}

	// Initialize with provided variables
	if options.Variables != nil {
		for k, v := range options.Variables {
			ctx.values[k] = v
		}
	}

	m.contexts[options.ID] = ctx
	return ctx, nil
}

func (m *MockScriptEngine) GetContext(id string) (engine.ScriptContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, exists := m.contexts[id]
	if !exists {
		return nil, errors.New("context not found: " + id)
	}
	return ctx, nil
}

func (m *MockScriptEngine) DestroyContext(ctx engine.ScriptContext) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	if ctx == nil {
		return errors.New("context is nil")
	}

	id := ctx.ID()
	if _, exists := m.contexts[id]; !exists {
		return errors.New("context not found: " + id)
	}

	// Destroy the context
	err := ctx.Destroy()
	if err != nil {
		return err
	}

	delete(m.contexts, id)
	return nil
}

func (m *MockScriptEngine) ExecuteScript(ctx context.Context, script string, options engine.ExecutionOptions) (*engine.ExecutionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	startTime := time.Now()

	// Use the executeFunc or default behavior
	var result engine.ScriptValue
	var err error

	params := options.Variables

	// Temporarily unlock for Execute call
	m.mu.Unlock()
	result, err = m.Execute(ctx, script, params)
	m.mu.Lock()

	duration := time.Since(startTime)

	execResult := &engine.ExecutionResult{
		Value:    result,
		Duration: duration,
		Metadata: make(map[string]interface{}),
	}

	if err != nil {
		execResult.Error = err
	}

	if options.CaptureOutput {
		execResult.Output = "mock output"
	}

	return execResult, nil
}

func (m *MockScriptEngine) SetResourceLimits(limits engine.ResourceLimits) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	m.resourceLimits = limits
	return nil
}

func (m *MockScriptEngine) GetMetrics() engine.EngineMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics
}

func (m *MockScriptEngine) GetEngineInfo() engine.EngineInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return engine.EngineInfo{
		Name:           m.name,
		Version:        m.version,
		Description:    "Mock script engine for testing",
		FileExtensions: []string{".mock"},
		Features:       []engine.EngineFeature{engine.FeatureAsync, engine.FeatureModules},
		Status:         engine.EngineStatusActive,
	}
}

func (m *MockScriptEngine) EnableProfiling(config engine.ProfilingConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	// Mock implementation - just return success
	return nil
}

func (m *MockScriptEngine) DisableProfiling() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	// Mock implementation - just return success
	return nil
}

func (m *MockScriptEngine) GetProfilingReport() (*engine.ProfilingReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	// Return a mock profiling report
	return &engine.ProfilingReport{
		StartTime:       time.Now().Add(-1 * time.Hour),
		EndTime:         time.Now(),
		Duration:        time.Hour,
		Executions:      100,
		AverageExecTime: 10 * time.Millisecond,
		MemoryStats: engine.MemoryStats{
			Allocated:      1024 * 1024,
			TotalAllocated: 2048 * 1024,
		},
		Metrics: map[string]interface{}{
			"calls":  m.metrics.BridgeCallsCount,
			"errors": m.metrics.ErrorCount,
		},
	}, nil
}

func (m *MockScriptEngine) ExportAPI(format engine.ExportFormat) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	// Return mock API export
	apiData := map[string]interface{}{
		"engine":  m.name,
		"version": m.version,
		"bridges": m.ListBridges(),
		"format":  string(format),
	}

	switch format {
	case engine.ExportFormatJSON:
		return json.Marshal(apiData)
	case engine.ExportFormatMarkdown:
		return []byte("# Mock Engine API\n\nThis is a mock API export."), nil
	default:
		return nil, errors.New("unsupported export format: " + string(format))
	}
}

func (m *MockScriptEngine) GenerateClientLibrary(language string, options engine.ClientLibraryOptions) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}

	// Return mock client library
	clientData := map[string]interface{}{
		"language":    language,
		"packageName": options.PackageName,
		"version":     options.Version,
		"engine":      m.name,
	}

	return json.Marshal(clientData)
}

func (m *MockScriptEngine) Features() []engine.EngineFeature {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return []engine.EngineFeature{engine.FeatureAsync, engine.FeatureModules}
}

func (m *MockScriptEngine) FileExtensions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return []string{".mock", ".test"}
}

func (m *MockScriptEngine) FromNative(value interface{}) (engine.ScriptValue, error) {
	// FromNative is the same as ToScriptValue
	return m.ToScriptValue(value)
}

func (m *MockScriptEngine) GetEventBus() engine.EventBus {
	// Return nil for mock implementation
	return nil
}

func (m *MockScriptEngine) GetTypeRegistry() engine.TypeRegistry {
	// Return nil for mock implementation
	return nil
}

func (m *MockScriptEngine) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

func (m *MockScriptEngine) RegisterTypeConverter(from string, to string, converter engine.TypeConverterFunc) error {
	// Mock implementation - just return success
	return nil
}

func (m *MockScriptEngine) SetMemoryLimit(limit int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	m.memoryLimit = limit
	return nil
}

func (m *MockScriptEngine) SetTimeout(timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return errors.New("engine not initialized")
	}

	m.timeout = timeout
	return nil
}

func (m *MockScriptEngine) Version() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.version
}

// mockScriptContext provides a simple context implementation
type mockScriptContext struct {
	id     string
	values map[string]interface{}
	mu     sync.RWMutex
}

func (c *mockScriptContext) ID() string {
	return c.id
}

func (c *mockScriptContext) SetVariable(name string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[name] = value
	return nil
}

func (c *mockScriptContext) GetVariable(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.values[name]
	if !ok {
		return nil, errors.New("variable not found: " + name)
	}
	return val, nil
}

func (c *mockScriptContext) Execute(script string) (interface{}, error) {
	// Mock implementation - just return the script
	return "executed: " + script, nil
}

func (c *mockScriptContext) Destroy() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values = nil
	return nil
}
