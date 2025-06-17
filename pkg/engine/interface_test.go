// ABOUTME: Test suite for the core interfaces of the multi-engine scripting architecture.
// ABOUTME: Validates the ScriptEngine, Bridge, and TypeConverter interfaces and related types.

package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toFloat64 converts numeric types to float64
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

// Mock implementation of ScriptEngine for testing
type mockScriptEngine struct {
	name           string
	version        string
	initialized    bool
	bridges        map[string]Bridge
	memoryLimit    int64
	timeout        time.Duration
	resourceLimits ResourceLimits
	metrics        EngineMetrics
	contexts       map[string]ScriptContext
}

func newMockScriptEngine(name string) *mockScriptEngine {
	return &mockScriptEngine{
		name:     name,
		version:  "1.0.0",
		bridges:  make(map[string]Bridge),
		contexts: make(map[string]ScriptContext),
	}
}

func (m *mockScriptEngine) Initialize(config EngineConfig) error {
	if m.initialized {
		return errors.New("already initialized")
	}
	m.initialized = true
	m.memoryLimit = config.MemoryLimit
	m.timeout = config.TimeoutLimit
	return nil
}

func (m *mockScriptEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (ScriptValue, error) {
	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}
	if script == "error" {
		return nil, &EngineError{
			Type:    ErrorTypeRuntime,
			Message: "runtime error",
		}
	}
	return NewStringValue("executed: " + script), nil
}

func (m *mockScriptEngine) ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (ScriptValue, error) {
	if !m.initialized {
		return nil, errors.New("engine not initialized")
	}
	return NewStringValue("executed file: " + path), nil
}

func (m *mockScriptEngine) Shutdown() error {
	if !m.initialized {
		return errors.New("engine not initialized")
	}
	m.initialized = false
	return nil
}

func (m *mockScriptEngine) RegisterBridge(bridge Bridge) error {
	id := bridge.GetID()
	if _, exists := m.bridges[id]; exists {
		return errors.New("bridge already registered")
	}
	m.bridges[id] = bridge
	return nil
}

func (m *mockScriptEngine) UnregisterBridge(name string) error {
	if _, exists := m.bridges[name]; !exists {
		return errors.New("bridge not found")
	}
	delete(m.bridges, name)
	return nil
}

func (m *mockScriptEngine) GetBridge(name string) (Bridge, error) {
	bridge, exists := m.bridges[name]
	if !exists {
		return nil, errors.New("bridge not found")
	}
	return bridge, nil
}

func (m *mockScriptEngine) ListBridges() []string {
	names := make([]string, 0, len(m.bridges))
	for name := range m.bridges {
		names = append(names, name)
	}
	return names
}

func (m *mockScriptEngine) ToNative(scriptValue ScriptValue) (interface{}, error) {
	if scriptValue == nil {
		return nil, nil
	}
	return scriptValue.ToGo(), nil
}

func (m *mockScriptEngine) FromNative(goValue interface{}) (ScriptValue, error) {
	switch v := goValue.(type) {
	case nil:
		return NewNilValue(), nil
	case bool:
		return NewBoolValue(v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		floatVal, _ := toFloat64(v)
		return NewNumberValue(floatVal), nil
	case string:
		return NewStringValue(v), nil
	case []interface{}:
		elems := make([]ScriptValue, len(v))
		for i, elem := range v {
			sv, err := m.FromNative(elem)
			if err != nil {
				return nil, err
			}
			elems[i] = sv
		}
		return NewArrayValue(elems), nil
	case map[string]interface{}:
		fields := make(map[string]ScriptValue)
		for k, val := range v {
			sv, err := m.FromNative(val)
			if err != nil {
				return nil, err
			}
			fields[k] = sv
		}
		return NewObjectValue(fields), nil
	default:
		return NewCustomValue("unknown", v), nil
	}
}

func (m *mockScriptEngine) Name() string {
	return m.name
}

func (m *mockScriptEngine) Version() string {
	return m.version
}

func (m *mockScriptEngine) FileExtensions() []string {
	return []string{"mock", "test"}
}

func (m *mockScriptEngine) Features() []EngineFeature {
	return []EngineFeature{FeatureAsync, FeatureDebugging}
}

func (m *mockScriptEngine) SetMemoryLimit(bytes int64) error {
	if bytes < 0 {
		return errors.New("invalid memory limit")
	}
	m.memoryLimit = bytes
	return nil
}

func (m *mockScriptEngine) SetTimeout(duration time.Duration) error {
	if duration < 0 {
		return errors.New("invalid timeout")
	}
	m.timeout = duration
	return nil
}

func (m *mockScriptEngine) SetResourceLimits(limits ResourceLimits) error {
	m.resourceLimits = limits
	return nil
}

func (m *mockScriptEngine) GetMetrics() EngineMetrics {
	return m.metrics
}

func (m *mockScriptEngine) CreateContext(options ContextOptions) (ScriptContext, error) {
	id := options.ID
	if id == "" {
		id = "ctx-" + time.Now().Format("20060102150405")
	}
	ctx := &mockScriptContext{
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

func (m *mockScriptEngine) DestroyContext(ctx ScriptContext) error {
	delete(m.contexts, ctx.ID())
	return nil
}

func (m *mockScriptEngine) ExecuteScript(ctx context.Context, script string, options ExecutionOptions) (*ExecutionResult, error) {
	start := time.Now()
	result := &ExecutionResult{
		Value:    NewStringValue("executed: " + script),
		Duration: time.Since(start),
		Metadata: make(map[string]interface{}),
	}
	return result, nil
}

// Task 1.4.11.1: Engine Event Bus
func (m *mockScriptEngine) GetEventBus() EventBus {
	return NewDefaultEventBus()
}

// Task 1.4.11.2: Type Conversion Registry
func (m *mockScriptEngine) RegisterTypeConverter(fromType, toType string, converter TypeConverterFunc) error {
	return nil
}

func (m *mockScriptEngine) GetTypeRegistry() TypeRegistry {
	return NewDefaultTypeRegistry()
}

// Task 1.4.11.3: Engine Profiling
func (m *mockScriptEngine) EnableProfiling(config ProfilingConfig) error {
	return nil
}

func (m *mockScriptEngine) DisableProfiling() error {
	return nil
}

func (m *mockScriptEngine) GetProfilingReport() (*ProfilingReport, error) {
	return &ProfilingReport{}, nil
}

// Task 1.4.11.4: Engine API Export
func (m *mockScriptEngine) ExportAPI(format ExportFormat) ([]byte, error) {
	return []byte("{}"), nil
}

func (m *mockScriptEngine) GenerateClientLibrary(language string, options ClientLibraryOptions) ([]byte, error) {
	return []byte("{}"), nil
}

// Mock implementation of Bridge for testing
type mockBridge struct {
	name        string
	methods     []MethodInfo
	initialized bool
}

func newMockBridge(name string) *mockBridge {
	return &mockBridge{
		name: name,
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
	}
}

func (m *mockBridge) GetID() string {
	return m.name
}

func (m *mockBridge) GetMetadata() BridgeMetadata {
	return BridgeMetadata{
		Name:        m.name,
		Version:     "1.0.0",
		Description: "Mock bridge for testing",
	}
}

func (m *mockBridge) Initialize(ctx context.Context) error {
	if m.initialized {
		return errors.New("already initialized")
	}
	m.initialized = true
	return nil
}

func (m *mockBridge) Cleanup(ctx context.Context) error {
	if !m.initialized {
		return errors.New("not initialized")
	}
	m.initialized = false
	return nil
}

func (m *mockBridge) IsInitialized() bool {
	return m.initialized
}

func (m *mockBridge) RegisterWithEngine(engine ScriptEngine) error {
	return engine.RegisterBridge(m)
}

func (m *mockBridge) Methods() []MethodInfo {
	return m.methods
}

func (m *mockBridge) TypeMappings() map[string]TypeMapping {
	return map[string]TypeMapping{
		"string": {
			GoType:     "string",
			ScriptType: "string",
			Converter:  "direct",
		},
	}
}

func (m *mockBridge) ValidateMethod(name string, args []ScriptValue) error {
	for _, method := range m.methods {
		if method.Name == name {
			return nil
		}
	}
	return errors.New("method not found")
}

func (m *mockBridge) ExecuteMethod(ctx context.Context, name string, args []ScriptValue) (ScriptValue, error) {
	if !m.initialized {
		return nil, errors.New("bridge not initialized")
	}
	for _, method := range m.methods {
		if method.Name == name {
			// Simple mock implementation
			if len(args) > 0 {
				return NewStringValue("result: " + args[0].String()), nil
			}
			return NewStringValue("result"), nil
		}
	}
	return nil, errors.New("method not found")
}

func (m *mockBridge) RequiredPermissions() []Permission {
	return []Permission{
		{
			Type:        PermissionNetwork,
			Resource:    "http://api.example.com",
			Actions:     []string{"GET", "POST"},
			Description: "Access to example API",
		},
	}
}

// Mock implementation of TypeConverter for testing
type mockTypeConverter struct{}

func (m *mockTypeConverter) ToBoolean(v ScriptValue) (bool, error) {
	if v == nil || v.IsNil() {
		return false, nil
	}
	switch v.Type() {
	case TypeBool:
		if bv, ok := v.(BoolValue); ok {
			return bv.Value(), nil
		}
	case TypeString:
		if sv, ok := v.(StringValue); ok {
			return sv.Value() == "true", nil
		}
	}
	return false, errors.New("cannot convert to boolean")
}

func (m *mockTypeConverter) ToNumber(v ScriptValue) (float64, error) {
	if v == nil || v.IsNil() {
		return 0, nil
	}
	if v.Type() == TypeNumber {
		if nv, ok := v.(NumberValue); ok {
			return nv.Value(), nil
		}
	}
	return 0, errors.New("cannot convert to number")
}

func (m *mockTypeConverter) ToString(v ScriptValue) (string, error) {
	if v == nil || v.IsNil() {
		return "", nil
	}
	return v.String(), nil
}

func (m *mockTypeConverter) ToArray(v ScriptValue) ([]ScriptValue, error) {
	if v == nil || v.IsNil() {
		return nil, nil
	}
	if v.Type() == TypeArray {
		if av, ok := v.(ArrayValue); ok {
			return av.Elements(), nil
		}
	}
	return nil, errors.New("cannot convert to array")
}

func (m *mockTypeConverter) ToMap(v ScriptValue) (map[string]ScriptValue, error) {
	if v == nil || v.IsNil() {
		return nil, nil
	}
	if v.Type() == TypeObject {
		if ov, ok := v.(ObjectValue); ok {
			return ov.Fields(), nil
		}
	}
	return nil, errors.New("cannot convert to map")
}

func (m *mockTypeConverter) ToStruct(v ScriptValue, target interface{}) error {
	return errors.New("not implemented")
}

func (m *mockTypeConverter) FromStruct(v interface{}) (ScriptValue, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTypeConverter) ToFunction(v ScriptValue) (Function, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTypeConverter) FromFunction(fn Function) (ScriptValue, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTypeConverter) FromInterface(v interface{}) (ScriptValue, error) {
	switch val := v.(type) {
	case nil:
		return NewNilValue(), nil
	case bool:
		return NewBoolValue(val), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		floatVal, _ := toFloat64(val)
		return NewNumberValue(floatVal), nil
	case string:
		return NewStringValue(val), nil
	default:
		return NewCustomValue("unknown", val), nil
	}
}

func (m *mockTypeConverter) ToInterface(v ScriptValue) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	return v.ToGo(), nil
}

func (m *mockTypeConverter) SupportsType(typeName string) bool {
	supportedTypes := []string{"bool", "string", "number", "array", "map"}
	for _, t := range supportedTypes {
		if t == typeName {
			return true
		}
	}
	return false
}

func (m *mockTypeConverter) GetTypeInfo(typeName string) TypeInfo {
	return TypeInfo{
		Name:        typeName,
		Category:    TypeCategoryPrimitive,
		Description: "Test type",
	}
}

// Mock implementation of ScriptContext for testing
type mockScriptContext struct {
	id        string
	variables map[string]interface{}
}

func (m *mockScriptContext) ID() string {
	return m.id
}

func (m *mockScriptContext) SetVariable(name string, value interface{}) error {
	m.variables[name] = value
	return nil
}

func (m *mockScriptContext) GetVariable(name string) (interface{}, error) {
	val, exists := m.variables[name]
	if !exists {
		return nil, errors.New("variable not found")
	}
	return val, nil
}

func (m *mockScriptContext) Execute(script string) (interface{}, error) {
	return "context executed: " + script, nil
}

func (m *mockScriptContext) Destroy() error {
	m.variables = nil
	return nil
}

// Tests for ScriptEngine interface
func TestScriptEngine(t *testing.T) {
	t.Run("Initialize", func(t *testing.T) {
		engine := newMockScriptEngine("test")
		config := EngineConfig{
			MemoryLimit:  1024 * 1024,
			TimeoutLimit: 30 * time.Second,
			SandboxMode:  true,
		}

		err := engine.Initialize(config)
		assert.NoError(t, err)
		assert.True(t, engine.initialized)
		assert.Equal(t, int64(1024*1024), engine.memoryLimit)
		assert.Equal(t, 30*time.Second, engine.timeout)

		// Test double initialization
		err = engine.Initialize(config)
		assert.Error(t, err)
	})

	t.Run("Execute", func(t *testing.T) {
		engine := newMockScriptEngine("test")

		// Test execution before initialization
		_, err := engine.Execute(context.Background(), "test", nil)
		assert.Error(t, err)

		// Initialize and test successful execution
		err = engine.Initialize(EngineConfig{})
		require.NoError(t, err)

		result, err := engine.Execute(context.Background(), "test script", nil)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, TypeString, result.Type())
		assert.Equal(t, "executed: test script", result.String())

		// Test error case
		_, err = engine.Execute(context.Background(), "error", nil)
		assert.Error(t, err)

		var engineErr *EngineError
		assert.True(t, errors.As(err, &engineErr))
		assert.Equal(t, ErrorTypeRuntime, engineErr.Type)
	})

	t.Run("Bridge Management", func(t *testing.T) {
		engine := newMockScriptEngine("test")
		bridge := newMockBridge("testBridge")

		// Register bridge
		err := engine.RegisterBridge(bridge)
		assert.NoError(t, err)

		// Test duplicate registration
		err = engine.RegisterBridge(bridge)
		assert.Error(t, err)

		// Get bridge
		retrievedBridge, err := engine.GetBridge("testBridge")
		assert.NoError(t, err)
		assert.Equal(t, bridge, retrievedBridge)

		// List bridges
		bridges := engine.ListBridges()
		assert.Contains(t, bridges, "testBridge")

		// Unregister bridge
		err = engine.UnregisterBridge("testBridge")
		assert.NoError(t, err)

		// Test getting non-existent bridge
		_, err = engine.GetBridge("testBridge")
		assert.Error(t, err)
	})

	t.Run("Type Conversion", func(t *testing.T) {
		engine := newMockScriptEngine("test")

		// Test ToNative
		scriptVal := NewStringValue("test value")
		native, err := engine.ToNative(scriptVal)
		assert.NoError(t, err)
		assert.Equal(t, "test value", native)

		// Test FromNative
		script, err := engine.FromNative("go value")
		assert.NoError(t, err)
		assert.NotNil(t, script)
		assert.Equal(t, TypeString, script.Type())
		assert.Equal(t, "go value", script.String())
	})

	t.Run("Metadata", func(t *testing.T) {
		engine := newMockScriptEngine("test")

		assert.Equal(t, "test", engine.Name())
		assert.Equal(t, "1.0.0", engine.Version())
		assert.Equal(t, []string{"mock", "test"}, engine.FileExtensions())
		assert.Equal(t, []EngineFeature{FeatureAsync, FeatureDebugging}, engine.Features())
	})

	t.Run("Resource Management", func(t *testing.T) {
		engine := newMockScriptEngine("test")

		// Set memory limit
		err := engine.SetMemoryLimit(1024 * 1024)
		assert.NoError(t, err)
		assert.Equal(t, int64(1024*1024), engine.memoryLimit)

		// Test invalid memory limit
		err = engine.SetMemoryLimit(-1)
		assert.Error(t, err)

		// Set timeout
		err = engine.SetTimeout(10 * time.Second)
		assert.NoError(t, err)
		assert.Equal(t, 10*time.Second, engine.timeout)

		// Set resource limits
		limits := ResourceLimits{
			MaxMemory:     2048,
			MaxGoroutines: 10,
			MaxExecTime:   5 * time.Second,
		}
		err = engine.SetResourceLimits(limits)
		assert.NoError(t, err)
		assert.Equal(t, limits, engine.resourceLimits)

		// Get metrics
		metrics := engine.GetMetrics()
		assert.NotNil(t, metrics)
	})

	t.Run("Context Management", func(t *testing.T) {
		engine := newMockScriptEngine("test")

		// Create context
		ctx, err := engine.CreateContext(ContextOptions{})
		assert.NoError(t, err)
		assert.NotNil(t, ctx)
		assert.NotEmpty(t, ctx.ID())

		// Set and get variable
		err = ctx.SetVariable("test", "value")
		assert.NoError(t, err)

		val, err := ctx.GetVariable("test")
		assert.NoError(t, err)
		assert.Equal(t, "value", val)

		// Execute in context
		result, err := ctx.Execute("context script")
		assert.NoError(t, err)
		assert.Equal(t, "context executed: context script", result)

		// Destroy context
		err = engine.DestroyContext(ctx)
		assert.NoError(t, err)
	})
}

// Tests for Bridge interface
func TestBridge(t *testing.T) {
	t.Run("Bridge Registration", func(t *testing.T) {
		bridge := newMockBridge("test")
		engine := newMockScriptEngine("engine")

		// Initialize bridge
		ctx := context.Background()
		err := bridge.Initialize(ctx)
		assert.NoError(t, err)
		assert.True(t, bridge.initialized)

		// Test double initialization
		err = bridge.Initialize(ctx)
		assert.Error(t, err)

		// Register with engine
		err = bridge.RegisterWithEngine(engine)
		assert.NoError(t, err)

		// Cleanup bridge
		err = bridge.Cleanup(ctx)
		assert.NoError(t, err)
		assert.False(t, bridge.initialized)

		// Test cleanup when not initialized
		err = bridge.Cleanup(ctx)
		assert.Error(t, err)
	})

	t.Run("Bridge Metadata", func(t *testing.T) {
		bridge := newMockBridge("test")

		assert.Equal(t, "test", bridge.GetID())
		metadata := bridge.GetMetadata()
		assert.Equal(t, "test", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)
		assert.Equal(t, "Mock bridge for testing", metadata.Description)

		methods := bridge.Methods()
		assert.Len(t, methods, 1)
		assert.Equal(t, "testMethod", methods[0].Name)
		assert.Equal(t, "A test method", methods[0].Description)
		assert.Len(t, methods[0].Parameters, 1)
		assert.Equal(t, "input", methods[0].Parameters[0].Name)
		assert.True(t, methods[0].Parameters[0].Required)

		mappings := bridge.TypeMappings()
		assert.Len(t, mappings, 1)
		assert.Contains(t, mappings, "string")
	})

	t.Run("Method Validation", func(t *testing.T) {
		bridge := newMockBridge("test")

		// Valid method
		err := bridge.ValidateMethod("testMethod", []ScriptValue{NewStringValue("arg")})
		assert.NoError(t, err)

		// Invalid method
		err = bridge.ValidateMethod("unknownMethod", []ScriptValue{})
		assert.Error(t, err)
	})

	t.Run("Permissions", func(t *testing.T) {
		bridge := newMockBridge("test")

		perms := bridge.RequiredPermissions()
		assert.Len(t, perms, 1)
		assert.Equal(t, PermissionNetwork, perms[0].Type)
		assert.Equal(t, "http://api.example.com", perms[0].Resource)
		assert.Contains(t, perms[0].Actions, "GET")
		assert.Contains(t, perms[0].Actions, "POST")
	})
}

// Tests for TypeConverter interface
func TestTypeConverter(t *testing.T) {
	converter := &mockTypeConverter{}

	t.Run("ToBoolean", func(t *testing.T) {
		// Test bool input
		result, err := converter.ToBoolean(NewBoolValue(true))
		assert.NoError(t, err)
		assert.True(t, result)

		// Test string input
		result, err = converter.ToBoolean(NewStringValue("true"))
		assert.NoError(t, err)
		assert.True(t, result)

		result, err = converter.ToBoolean(NewStringValue("false"))
		assert.NoError(t, err)
		assert.False(t, result)

		// Test unsupported type
		_, err = converter.ToBoolean(NewNumberValue(123))
		assert.Error(t, err)
	})

	t.Run("ToNumber", func(t *testing.T) {
		// Test number input
		result, err := converter.ToNumber(NewNumberValue(3.14))
		assert.NoError(t, err)
		assert.Equal(t, 3.14, result)

		// Test another number
		result, err = converter.ToNumber(NewNumberValue(42))
		assert.NoError(t, err)
		assert.Equal(t, 42.0, result)

		// Test unsupported type
		_, err = converter.ToNumber(NewStringValue("not a number"))
		assert.Error(t, err)
	})

	t.Run("ToString", func(t *testing.T) {
		// Test string input
		result, err := converter.ToString(NewStringValue("hello"))
		assert.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test number type (should work since ToString returns v.String())
		result, err = converter.ToString(NewNumberValue(123))
		assert.NoError(t, err)
		assert.Equal(t, "123", result)
	})

	t.Run("ToArray", func(t *testing.T) {
		// Test array input
		elements := []ScriptValue{NewStringValue("a"), NewStringValue("b"), NewStringValue("c")}
		input := NewArrayValue(elements)
		result, err := converter.ToArray(input)
		assert.NoError(t, err)
		assert.Equal(t, elements, result)

		// Test unsupported type
		_, err = converter.ToArray(NewStringValue("not an array"))
		assert.Error(t, err)
	})

	t.Run("ToMap", func(t *testing.T) {
		// Test map input
		fields := map[string]ScriptValue{"key": NewStringValue("value")}
		input := NewObjectValue(fields)
		result, err := converter.ToMap(input)
		assert.NoError(t, err)
		assert.Equal(t, fields, result)

		// Test unsupported type
		_, err = converter.ToMap(NewStringValue("not a map"))
		assert.Error(t, err)
	})

	t.Run("Type Support", func(t *testing.T) {
		assert.True(t, converter.SupportsType("bool"))
		assert.True(t, converter.SupportsType("string"))
		assert.True(t, converter.SupportsType("number"))
		assert.True(t, converter.SupportsType("array"))
		assert.True(t, converter.SupportsType("map"))
		assert.False(t, converter.SupportsType("function"))
		assert.False(t, converter.SupportsType("custom"))
	})

	t.Run("Type Info", func(t *testing.T) {
		info := converter.GetTypeInfo("string")
		assert.Equal(t, "string", info.Name)
		assert.Equal(t, TypeCategoryPrimitive, info.Category)
		assert.Equal(t, "Test type", info.Description)
	})
}

// Tests for EngineConfig
func TestEngineConfig(t *testing.T) {
	config := EngineConfig{
		MemoryLimit:    1024 * 1024,
		TimeoutLimit:   30 * time.Second,
		GoroutineLimit: 10,
		SandboxMode:    true,
		AllowedModules: []string{"json", "http"},
		FileSystemMode: FSModeReadOnly,
		DebugMode:      true,
		LogLevel:       "debug",
	}

	assert.Equal(t, int64(1024*1024), config.MemoryLimit)
	assert.Equal(t, 30*time.Second, config.TimeoutLimit)
	assert.Equal(t, 10, config.GoroutineLimit)
	assert.True(t, config.SandboxMode)
	assert.Contains(t, config.AllowedModules, "json")
	assert.Contains(t, config.AllowedModules, "http")
	assert.Equal(t, FSModeReadOnly, config.FileSystemMode)
	assert.True(t, config.DebugMode)
	assert.Equal(t, "debug", config.LogLevel)
}

// Tests for EngineError
func TestEngineError(t *testing.T) {
	baseErr := errors.New("underlying error")
	err := &EngineError{
		Type:       ErrorTypeSyntax,
		Message:    "syntax error at line 10",
		ScriptLine: 10,
		ScriptCol:  15,
		StackTrace: []string{"function1", "function2"},
		Cause:      baseErr,
	}

	assert.Equal(t, "syntax error at line 10", err.Error())
	assert.Equal(t, ErrorTypeSyntax, err.Type)
	assert.Equal(t, 10, err.ScriptLine)
	assert.Equal(t, 15, err.ScriptCol)
	assert.Contains(t, err.StackTrace, "function1")
	assert.Contains(t, err.StackTrace, "function2")
	assert.Equal(t, baseErr, err.Unwrap())
}

// Tests for enums and constants
func TestEnums(t *testing.T) {
	t.Run("EngineFeatures", func(t *testing.T) {
		features := []EngineFeature{
			FeatureAsync,
			FeatureCoroutines,
			FeatureModules,
			FeatureDebugging,
			FeatureHotReload,
			FeatureCompilation,
			FeatureInteractive,
			FeatureStreaming,
		}

		assert.Len(t, features, 8)
		assert.Contains(t, features, FeatureAsync)
		assert.Contains(t, features, FeatureDebugging)
	})

	t.Run("FSMode", func(t *testing.T) {
		modes := []FSMode{
			FSModeReadOnly,
			FSModeReadWrite,
			FSModeNone,
			FSModeSandbox,
		}

		assert.Len(t, modes, 4)
		assert.Contains(t, modes, FSModeReadOnly)
		assert.Contains(t, modes, FSModeSandbox)
	})

	t.Run("TypeCategory", func(t *testing.T) {
		categories := []TypeCategory{
			TypeCategoryPrimitive,
			TypeCategoryObject,
			TypeCategoryFunction,
			TypeCategoryArray,
			TypeCategoryMap,
			TypeCategoryCustom,
		}

		assert.Len(t, categories, 6)
		assert.Contains(t, categories, TypeCategoryPrimitive)
		assert.Contains(t, categories, TypeCategoryFunction)
	})

	t.Run("PermissionType", func(t *testing.T) {
		permissions := []PermissionType{
			PermissionFileSystem,
			PermissionNetwork,
			PermissionProcess,
			PermissionMemory,
			PermissionTime,
			PermissionCrypto,
		}

		assert.Len(t, permissions, 6)
		assert.Contains(t, permissions, PermissionFileSystem)
		assert.Contains(t, permissions, PermissionNetwork)
	})

	t.Run("ErrorType", func(t *testing.T) {
		errorTypes := []ErrorType{
			ErrorTypeSyntax,
			ErrorTypeRuntime,
			ErrorTypeType,
			ErrorTypeResource,
			ErrorTypeSecurity,
			ErrorTypeBridge,
			ErrorTypeTimeout,
			ErrorTypeMemory,
			ErrorTypePermission,
		}

		assert.Len(t, errorTypes, 9)
		assert.Contains(t, errorTypes, ErrorTypeSyntax)
		assert.Contains(t, errorTypes, ErrorTypeSecurity)
	})
}
