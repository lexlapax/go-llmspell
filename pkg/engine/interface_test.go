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

// Using test helper mock from test_helpers.go
func newMockScriptEngine(name string) *testMockScriptEngine {
	mock := newTestMockScriptEngine(name)
	// Add custom execute behavior for this test
	mock.executeFunc = func(ctx context.Context, script string, params map[string]interface{}) (ScriptValue, error) {
		if script == "error" {
			return nil, &EngineError{
				Type:    ErrorTypeRuntime,
				Message: "runtime error",
			}
		}
		return NewStringValue("executed: " + script), nil
	}
	return mock
}

// Using test helper mock from test_helpers.go
func newMockBridge(name string) *testMockBridge {
	return newTestMockBridge(name)
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

// Using test helper mock from test_helpers.go for ScriptContext

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
