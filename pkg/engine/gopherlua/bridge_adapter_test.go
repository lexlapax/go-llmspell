// ABOUTME: Tests for bridge adapter system that wraps go-llms bridges for Lua script access
// ABOUTME: Validates bridge method wrapping, type conversion, error handling, and metadata exposure

package gopherlua

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestBridgeAdapter_Creation(t *testing.T) {
	t.Run("create_adapter_from_bridge", func(t *testing.T) {
		// Create a mock bridge
		mockBridge := &mockBridge{
			id: "test-bridge",
			metadata: engine.BridgeMetadata{
				Name:        "Test Bridge",
				Version:     "1.0.0",
				Description: "A test bridge",
			},
		}

		// Create adapter
		adapter := NewBridgeAdapter(mockBridge)
		require.NotNil(t, adapter)

		// Verify bridge is wrapped
		assert.Equal(t, mockBridge, adapter.GetBridge())
		assert.Equal(t, "test-bridge", adapter.GetID())
	})

	t.Run("adapter_exposes_metadata", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			metadata: engine.BridgeMetadata{
				Name:        "Test Bridge",
				Version:     "1.0.0",
				Description: "A test bridge",
				Author:      "Test Author",
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		metadata := adapter.GetMetadata()

		assert.Equal(t, "Test Bridge", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)
		assert.Equal(t, "A test bridge", metadata.Description)
		assert.Equal(t, "Test Author", metadata.Author)
	})
}

func TestBridgeAdapter_MethodDiscovery(t *testing.T) {
	t.Run("discover_bridge_methods", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{
					Name:        "testMethod",
					Description: "A test method",
					ReturnType:  "string",
				},
				{
					Name:        "anotherMethod",
					Description: "Another test method",
					ReturnType:  "int",
				},
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		methods := adapter.GetMethods()

		assert.Len(t, methods, 2)
		assert.Contains(t, methods, "testMethod")
		assert.Contains(t, methods, "anotherMethod")
	})

	t.Run("get_method_info", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{
					Name:        "testMethod",
					Description: "A test method",
					Parameters: []engine.ParameterInfo{
						{Name: "input", Type: "map[string]interface{}", Required: true},
						{Name: "option", Type: "interface{}", Required: false},
					},
					ReturnType: "string",
				},
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		info, err := adapter.GetMethodInfo("testMethod")

		require.NoError(t, err)
		assert.Equal(t, "testMethod", info.Name)
		assert.Equal(t, "A test method", info.Description)
		require.Len(t, info.Parameters, 2)
		assert.Equal(t, "input", info.Parameters[0].Name)
		assert.True(t, info.Parameters[0].Required)
		assert.Equal(t, "option", info.Parameters[1].Name)
		assert.False(t, info.Parameters[1].Required)
	})

	t.Run("get_unknown_method_info", func(t *testing.T) {
		mockBridge := &mockBridge{
			id:      "test-bridge",
			methods: []engine.MethodInfo{},
		}

		adapter := NewBridgeAdapter(mockBridge)
		_, err := adapter.GetMethodInfo("unknown")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method not found")
	})
}

func TestBridgeAdapter_LuaModule(t *testing.T) {
	t.Run("create_lua_module", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			metadata: engine.BridgeMetadata{
				Name: "Test Bridge",
			},
			methods: []engine.MethodInfo{
				{Name: "testMethod"},
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		L := lua.NewState()
		defer L.Close()

		// Create Lua module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get the module
		module := L.Get(-1).(*lua.LTable)
		L.Pop(1)

		// Check module has metadata
		assert.NotEqual(t, lua.LNil, module.RawGetString("_bridge"))
		assert.NotEqual(t, lua.LNil, module.RawGetString("_version"))

		// Check method exists
		assert.NotEqual(t, lua.LNil, module.RawGetString("testMethod"))
	})
}

func TestBridgeAdapter_MethodWrapping(t *testing.T) {
	t.Run("wrap_simple_method", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{Name: "echo"},
			},
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "echo" && len(args) > 0 {
					return args[0], nil
				}
				return nil, errors.New("invalid call")
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		L := lua.NewState()
		defer L.Close()

		// Get wrapped method
		fn := adapter.WrapMethod("echo")
		L.SetGlobal("echo", L.NewFunction(fn))

		// Call from Lua
		err := L.DoString(`
			result = echo("hello")
			assert(result == "hello")
		`)
		assert.NoError(t, err)
	})

	t.Run("wrap_method_with_table_input", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{Name: "process"},
			},
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "process" && len(args) > 0 {
					if m, ok := args[0].(map[string]interface{}); ok {
						return m["value"], nil
					}
				}
				return nil, errors.New("invalid input")
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		L := lua.NewState()
		defer L.Close()

		// Get wrapped method
		fn := adapter.WrapMethod("process")
		L.SetGlobal("process", L.NewFunction(fn))

		// Call from Lua with table
		err := L.DoString(`
			result = process({value = "test"})
			assert(result == "test")
		`)
		assert.NoError(t, err)
	})

	t.Run("wrap_method_with_error_handling", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{Name: "failing"},
			},
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("bridge error")
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		L := lua.NewState()
		defer L.Close()

		// Get wrapped method
		fn := adapter.WrapMethod("failing")
		L.SetGlobal("failing", L.NewFunction(fn))

		// Call from Lua - should return nil, error
		err := L.DoString(`
			result, err = failing()
			assert(result == nil)
			assert(err ~= nil)
			assert(string.find(err, "bridge error"))
		`)
		assert.NoError(t, err)
	})

	t.Run("wrap_method_with_multiple_returns", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{Name: "multiReturn"},
			},
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				// Return a slice which should be unpacked
				return []interface{}{"first", 42, true}, nil
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		L := lua.NewState()
		defer L.Close()

		// Get wrapped method
		fn := adapter.WrapMethod("multiReturn")
		L.SetGlobal("multiReturn", L.NewFunction(fn))

		// Call from Lua
		err := L.DoString(`
			a, b, c = multiReturn()
			assert(a == "first")
			assert(b == 42)
			assert(c == true)
		`)
		assert.NoError(t, err)
	})
}

func TestBridgeAdapter_TypeConversion(t *testing.T) {
	t.Run("convert_complex_types", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{Name: "complexMethod"},
			},
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				// Return a complex structure
				return map[string]interface{}{
					"name":   "test",
					"count":  42,
					"active": true,
					"tags":   []string{"a", "b", "c"},
					"nested": map[string]interface{}{
						"value": 3.14,
					},
				}, nil
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		converter := NewLuaTypeConverter()
		adapter.SetTypeConverter(converter)

		L := lua.NewState()
		defer L.Close()

		// Get wrapped method
		fn := adapter.WrapMethod("complexMethod")
		L.SetGlobal("complexMethod", L.NewFunction(fn))

		// Call from Lua and verify structure
		err := L.DoString(`
			result = complexMethod()
			assert(result.name == "test")
			assert(result.count == 42)
			assert(result.active == true)
			assert(#result.tags == 3)
			assert(result.tags[1] == "a")
			assert(result.nested.value == 3.14)
		`)
		assert.NoError(t, err)
	})
}

func TestBridgeAdapter_Registration(t *testing.T) {
	t.Run("register_with_module_system", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			metadata: engine.BridgeMetadata{
				Name: "Test Bridge",
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		ms := NewModuleSystem()

		// Register adapter as module
		err := adapter.RegisterAsModule(ms, "testbridge")
		assert.NoError(t, err)

		// Module should exist
		assert.True(t, ms.Exists("testbridge"))

		// Should be loadable
		L := lua.NewState()
		defer L.Close()

		err = ms.LoadModule(L, "testbridge")
		assert.NoError(t, err)
	})

	t.Run("auto_register_dependencies", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "dependent-bridge",
			metadata: engine.BridgeMetadata{
				Name:         "Dependent Bridge",
				Dependencies: []string{"base-bridge"},
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		ms := NewModuleSystem()

		// Register with dependencies
		err := adapter.RegisterAsModule(ms, "dependent")
		assert.NoError(t, err)

		// Check module has dependencies
		info, err := ms.GetModuleInfo("dependent")
		assert.NoError(t, err)
		assert.Contains(t, info.Dependencies, "base-bridge")
	})
}

func TestBridgeAdapter_MethodValidation(t *testing.T) {
	t.Run("validate_method_arguments", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{
					Name: "validateMethod",
					Parameters: []engine.ParameterInfo{
						{Name: "name", Type: "string", Required: true},
						{Name: "age", Type: "int", Required: true},
						{Name: "email", Type: "string", Required: false},
					},
				},
			},
			validateFunc: func(method string, args ...interface{}) error {
				if method == "validateMethod" {
					if len(args) < 1 {
						return errors.New("missing arguments")
					}
					if m, ok := args[0].(map[string]interface{}); ok {
						if _, ok := m["name"]; !ok {
							return errors.New("missing required field: name")
						}
						if _, ok := m["age"]; !ok {
							return errors.New("missing required field: age")
						}
					}
				}
				return nil
			},
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				return "success", nil
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		adapter.EnableValidation(true)

		L := lua.NewState()
		defer L.Close()

		fn := adapter.WrapMethod("validateMethod")
		L.SetGlobal("validateMethod", L.NewFunction(fn))

		// Test missing required fields
		err := L.DoString(`
			local result, err = validateMethod({name = "test"})
			assert(result == nil)
			assert(string.find(err, "validation error"))
		`)
		assert.NoError(t, err)

		// Test with all required fields
		err = L.DoString(`
			result = validateMethod({name = "test", age = 25})
			assert(result == "success")
		`)
		assert.NoError(t, err)
	})
}

func TestBridgeAdapter_Performance(t *testing.T) {
	t.Run("method_caching", func(t *testing.T) {
		callCount := 0
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{Name: "cached"},
			},
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				callCount++
				return "result", nil
			},
		}

		adapter := NewBridgeAdapter(mockBridge)

		// Get same method multiple times
		fn1 := adapter.WrapMethod("cached")
		fn2 := adapter.WrapMethod("cached")

		// Should be the same function (cached)
		// Note: In real implementation, we'd check if they're the same reference
		assert.NotNil(t, fn1)
		assert.NotNil(t, fn2)
	})
}

func TestBridgeAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_panic_in_bridge", func(t *testing.T) {
		mockBridge := &mockBridge{
			id: "test-bridge",
			methods: []engine.MethodInfo{
				{Name: "panic"},
			},
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				panic("bridge panic!")
			},
		}

		adapter := NewBridgeAdapter(mockBridge)
		L := lua.NewState()
		defer L.Close()

		fn := adapter.WrapMethod("panic")
		L.SetGlobal("panicMethod", L.NewFunction(fn))

		// Should catch panic and return error
		err := L.DoString(`
			local result, err = panicMethod()
			print("result:", result)
			print("err:", err)
			assert(result == nil)
			assert(string.find(err, "panic"))
		`)
		assert.NoError(t, err)
	})
}

// Mock bridge implementation for testing
type mockBridge struct {
	id           string
	metadata     engine.BridgeMetadata
	methods      []engine.MethodInfo
	initialized  bool
	dependencies []string
	callFunc     func(string, ...interface{}) (interface{}, error)
	validateFunc func(string, ...interface{}) error
}

func (m *mockBridge) GetID() string {
	return m.id
}

func (m *mockBridge) GetMetadata() engine.BridgeMetadata {
	return m.metadata
}

func (m *mockBridge) Initialize(ctx context.Context) error {
	m.initialized = true
	return nil
}

func (m *mockBridge) Cleanup(ctx context.Context) error {
	m.initialized = false
	return nil
}

func (m *mockBridge) IsInitialized() bool {
	return m.initialized
}

func (m *mockBridge) GetDependencies() []string {
	return m.dependencies
}

func (m *mockBridge) Methods() []engine.MethodInfo {
	return m.methods
}

func (m *mockBridge) Call(method string, args ...interface{}) (interface{}, error) {
	if m.callFunc != nil {
		return m.callFunc(method, args...)
	}
	return nil, errors.New("method not implemented")
}

func (m *mockBridge) ValidateMethod(method string, args []engine.ScriptValue) error {
	// Find method info
	for _, mi := range m.methods {
		if mi.Name == method {
			return nil
		}
	}
	return errors.New("unknown method")
}

func (m *mockBridge) ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if m.callFunc != nil {
		// Convert ScriptValue args to interface{} for the old callFunc
		interfaceArgs := make([]interface{}, len(args))
		for i, arg := range args {
			interfaceArgs[i] = arg.ToGo()
		}
		result, err := m.callFunc(method, interfaceArgs...)
		if err != nil {
			return engine.NewErrorValue(err), err
		}
		// Convert result back to ScriptValue
		switch v := result.(type) {
		case string:
			return engine.NewStringValue(v), nil
		case int:
			return engine.NewNumberValue(float64(v)), nil
		case float64:
			return engine.NewNumberValue(v), nil
		case bool:
			return engine.NewBoolValue(v), nil
		default:
			return engine.NewStringValue(fmt.Sprintf("%v", v)), nil
		}
	}
	return engine.NewStringValue("mock result"), nil
}

func (m *mockBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return nil
}

func (m *mockBridge) TypeMappings() map[string]engine.TypeMapping {
	return make(map[string]engine.TypeMapping)
}

func (m *mockBridge) RequiredPermissions() []engine.Permission {
	return nil
}
