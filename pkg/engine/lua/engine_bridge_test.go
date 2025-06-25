// ABOUTME: Tests for LuaEngine bridge registration and management functionality
// ABOUTME: Validates bridge lifecycle, module creation, method wrapping, and Lua-side access

package lua

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/converters"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/factory"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// testBridgeForRegistration is a more complete test bridge for registration testing
type testBridgeForRegistration struct {
	id          string
	meta        engine.BridgeMetadata
	initialized bool
	methods     []engine.MethodInfo
}

func (b *testBridgeForRegistration) GetID() string {
	return b.id
}

func (b *testBridgeForRegistration) GetMetadata() engine.BridgeMetadata {
	return b.meta
}

func (b *testBridgeForRegistration) Initialize(ctx context.Context) error {
	b.initialized = true
	return nil
}

func (b *testBridgeForRegistration) Cleanup(ctx context.Context) error {
	b.initialized = false
	return nil
}

func (b *testBridgeForRegistration) IsInitialized() bool {
	return b.initialized
}

func (b *testBridgeForRegistration) RegisterWithEngine(engine engine.ScriptEngine) error {
	return nil
}

func (b *testBridgeForRegistration) Methods() []engine.MethodInfo {
	return b.methods
}

func (b *testBridgeForRegistration) ValidateMethod(name string, args []engine.ScriptValue) error {
	for _, method := range b.methods {
		if method.Name == name {
			if len(args) == len(method.Parameters) {
				return nil
			}
		}
	}
	return fmt.Errorf("invalid method call: %s", name)
}

func (b *testBridgeForRegistration) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	switch name {
	case "testMethod":
		if len(args) > 0 {
			return engine.NewStringValue("result: " + args[0].String()), nil
		}
		return engine.NewStringValue("result: no input"), nil
	case "mathOperation":
		if len(args) >= 2 {
			a, _ := engine.ConvertToNumber(args[0])
			b, _ := engine.ConvertToNumber(args[1])
			return engine.NewNumberValue(a + b), nil
		}
		return engine.NewNumberValue(0), nil
	case "add":
		if len(args) >= 2 {
			a, _ := engine.ConvertToNumber(args[0])
			b, _ := engine.ConvertToNumber(args[1])
			return engine.NewNumberValue(a + b), nil
		}
		return engine.NewNumberValue(0), nil
	case "multiply":
		if len(args) >= 2 {
			a, _ := engine.ConvertToNumber(args[0])
			b, _ := engine.ConvertToNumber(args[1])
			return engine.NewNumberValue(a * b), nil
		}
		return engine.NewNumberValue(1), nil
	default:
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), fmt.Errorf("unknown method: %s", name)
	}
}

func (b *testBridgeForRegistration) TypeMappings() map[string]engine.TypeMapping {
	return nil
}

func (b *testBridgeForRegistration) RequiredPermissions() []engine.Permission {
	return nil
}

func TestLuaEngine_BridgeRegistration(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	bridge := &testBridgeForRegistration{
		id: "test_bridge",
		meta: engine.BridgeMetadata{
			Name:        "Test Bridge",
			Version:     "1.0.0",
			Description: "Bridge for testing registration",
		},
		methods: []engine.MethodInfo{
			{
				Name:        "testMethod",
				Description: "Test method for bridge",
				Parameters: []engine.ParameterInfo{
					{Name: "input", Type: "string", Required: true},
				},
				ReturnType: "string",
			},
			{
				Name:        "mathOperation",
				Description: "Math operation method",
				Parameters: []engine.ParameterInfo{
					{Name: "a", Type: "number", Required: true},
					{Name: "b", Type: "number", Required: true},
				},
				ReturnType: "number",
			},
		},
	}

	tests := []struct {
		name    string
		bridge  engine.Bridge
		wantErr bool
	}{
		{
			name:    "register_valid_bridge",
			bridge:  bridge,
			wantErr: false,
		},
		{
			name:    "register_duplicate_bridge",
			bridge:  bridge,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := eng.RegisterBridge(tt.bridge)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLuaEngine_BridgeModuleCreation(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	bridge := &testBridgeForRegistration{
		id: "math_bridge",
		meta: engine.BridgeMetadata{
			Name:        "Math Bridge",
			Version:     "1.0.0",
			Description: "Math operations bridge",
		},
		methods: []engine.MethodInfo{
			{
				Name:        "add",
				Description: "Add two numbers",
				Parameters: []engine.ParameterInfo{
					{Name: "a", Type: "number", Required: true},
					{Name: "b", Type: "number", Required: true},
				},
				ReturnType: "number",
			},
			{
				Name:        "multiply",
				Description: "Multiply two numbers",
				Parameters: []engine.ParameterInfo{
					{Name: "x", Type: "number", Required: true},
					{Name: "y", Type: "number", Required: true},
				},
				ReturnType: "number",
			},
		},
	}

	err = eng.RegisterBridge(bridge)
	require.NoError(t, err)

	// Test that the bridge can be retrieved
	retrievedBridge, err := eng.GetBridge("math_bridge")
	require.NoError(t, err)
	assert.Equal(t, bridge, retrievedBridge)

	// Test that the bridge appears in the list
	bridges := eng.ListBridges()
	assert.Contains(t, bridges, "math_bridge")
}

func TestLuaEngine_BridgeMethodWrapping(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	bridge := &testBridgeForRegistration{
		id: "string_bridge",
		meta: engine.BridgeMetadata{
			Name:        "String Bridge",
			Version:     "1.0.0",
			Description: "String operations bridge",
		},
		methods: []engine.MethodInfo{
			{
				Name:        "concat",
				Description: "Concatenate strings",
				Parameters: []engine.ParameterInfo{
					{Name: "str1", Type: "string", Required: true},
					{Name: "str2", Type: "string", Required: true},
				},
				ReturnType: "string",
			},
			{
				Name:        "length",
				Description: "Get string length",
				Parameters: []engine.ParameterInfo{
					{Name: "str", Type: "string", Required: true},
				},
				ReturnType: "number",
			},
		},
	}

	err = eng.RegisterBridge(bridge)
	require.NoError(t, err)

	// Verify bridge methods are accessible
	methods := bridge.Methods()
	assert.Len(t, methods, 2)

	// Check method details
	concatMethod := methods[0]
	assert.Equal(t, "concat", concatMethod.Name)
	assert.Len(t, concatMethod.Parameters, 2)
	assert.Equal(t, "string", concatMethod.ReturnType)

	lengthMethod := methods[1]
	assert.Equal(t, "length", lengthMethod.Name)
	assert.Len(t, lengthMethod.Parameters, 1)
	assert.Equal(t, "number", lengthMethod.ReturnType)
}

func TestLuaEngine_BridgeLifecycleManagement(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	bridge := &testBridgeForRegistration{
		id: "lifecycle_bridge",
		meta: engine.BridgeMetadata{
			Name:        "Lifecycle Bridge",
			Version:     "1.0.0",
			Description: "Bridge for testing lifecycle",
		},
	}

	// Test registration
	err = eng.RegisterBridge(bridge)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized(), "Bridge should be initialized after registration")

	// Test unregistration
	err = eng.UnregisterBridge("lifecycle_bridge")
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized(), "Bridge should be cleaned up after unregistration")

	// Verify bridge is no longer accessible
	_, err = eng.GetBridge("lifecycle_bridge")
	assert.Error(t, err)

	// Verify bridge is not in the list
	bridges := eng.ListBridges()
	assert.NotContains(t, bridges, "lifecycle_bridge")
}

func TestLuaEngine_BridgeMetadataHandling(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	bridge := &testBridgeForRegistration{
		id: "metadata_bridge",
		meta: engine.BridgeMetadata{
			Name:        "Metadata Test Bridge",
			Version:     "2.1.0",
			Description: "Bridge for testing metadata handling",
			Author:      "Test Author",
			License:     "MIT",
		},
	}

	err = eng.RegisterBridge(bridge)
	require.NoError(t, err)

	// Test metadata retrieval
	retrievedBridge, err := eng.GetBridge("metadata_bridge")
	require.NoError(t, err)

	meta := retrievedBridge.GetMetadata()
	assert.Equal(t, "Metadata Test Bridge", meta.Name)
	assert.Equal(t, "2.1.0", meta.Version)
	assert.Equal(t, "Bridge for testing metadata handling", meta.Description)
	assert.Equal(t, "Test Author", meta.Author)
	assert.Equal(t, "MIT", meta.License)
}

func TestLuaEngine_BridgeValidation(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	bridge := &testBridgeForRegistration{
		id: "validation_bridge",
		meta: engine.BridgeMetadata{
			Name:        "Validation Bridge",
			Version:     "1.0.0",
			Description: "Bridge for testing validation",
		},
		methods: []engine.MethodInfo{
			{
				Name:        "validateMe",
				Description: "Method that validates inputs",
				Parameters: []engine.ParameterInfo{
					{Name: "required_param", Type: "string", Required: true},
					{Name: "optional_param", Type: "number", Required: false},
				},
				ReturnType: "boolean",
			},
		},
	}

	err = eng.RegisterBridge(bridge)
	require.NoError(t, err)

	// Test method validation
	validArgs := []engine.ScriptValue{engine.NewStringValue("test_string"), engine.NewNumberValue(42.0)}
	err = bridge.ValidateMethod("validateMe", validArgs)
	assert.NoError(t, err)

	// Test invalid method name
	err = bridge.ValidateMethod("nonExistentMethod", validArgs)
	assert.Error(t, err)

	// Test wrong argument count
	invalidArgs := []engine.ScriptValue{engine.NewStringValue("test_string")}
	err = bridge.ValidateMethod("validateMe", invalidArgs)
	assert.Error(t, err)
}

func TestLuaEngine_MultipleBridgeManagement(t *testing.T) {
	// Use test factory to support test bridge IDs
	testFactory := factory.NewTestAdapterFactory()
	eng := NewLuaEngineWithFactory(testFactory)
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	// Create multiple bridges
	bridges := []*testBridgeForRegistration{
		{
			id: "bridge_1",
			meta: engine.BridgeMetadata{
				Name:    "Bridge One",
				Version: "1.0.0",
			},
		},
		{
			id: "bridge_2",
			meta: engine.BridgeMetadata{
				Name:    "Bridge Two",
				Version: "1.0.0",
			},
		},
		{
			id: "bridge_3",
			meta: engine.BridgeMetadata{
				Name:    "Bridge Three",
				Version: "1.0.0",
			},
		},
	}

	// Register all bridges
	for _, bridge := range bridges {
		err := eng.RegisterBridge(bridge)
		require.NoError(t, err)
	}

	// Verify all bridges are listed
	bridgeList := eng.ListBridges()
	assert.Len(t, bridgeList, 3)
	assert.Contains(t, bridgeList, "bridge_1")
	assert.Contains(t, bridgeList, "bridge_2")
	assert.Contains(t, bridgeList, "bridge_3")

	// Verify each bridge can be retrieved
	for _, bridge := range bridges {
		retrieved, err := eng.GetBridge(bridge.id)
		require.NoError(t, err)
		assert.Equal(t, bridge, retrieved)
	}

	// Unregister one bridge
	err = eng.UnregisterBridge("bridge_2")
	require.NoError(t, err)

	// Verify the remaining bridges
	bridgeList = eng.ListBridges()
	assert.Len(t, bridgeList, 2)
	assert.Contains(t, bridgeList, "bridge_1")
	assert.NotContains(t, bridgeList, "bridge_2")
	assert.Contains(t, bridgeList, "bridge_3")
}

func TestLuaEngine_BridgeErrorHandling(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name      string
		operation func() error
		wantErr   bool
	}{
		{
			name: "get_nonexistent_bridge",
			operation: func() error {
				_, err := eng.GetBridge("nonexistent")
				return err
			},
			wantErr: true,
		},
		{
			name: "unregister_nonexistent_bridge",
			operation: func() error {
				return eng.UnregisterBridge("nonexistent")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBridgeManagerAdapterModuleCreation tests module creation via adapter callback
func TestBridgeManagerAdapterModuleCreation(t *testing.T) {
	t.Run("create module via adapter callback", func(t *testing.T) {
		// Create Lua state and converter
		L := lua.NewState()
		defer L.Close()

		converter := converters.NewLuaTypeConverter()

		// Create mock adapter that returns a test module
		mockAdapter := &mockModuleAdapter{
			moduleFunc: func(L *lua.LState) int {
				module := L.NewTable()
				L.SetField(module, "testMethod", L.NewFunction(func(L *lua.LState) int {
					L.Push(lua.LString("adapter result"))
					return 1
				}))
				L.Push(module)
				return 1
			},
		}

		// Create module creator callback that returns our mock adapter's module
		moduleCreator := func(bridgeID string) (lua.LGFunction, error) {
			if bridgeID == "test_bridge" {
				return mockAdapter.CreateLuaModule(), nil
			}
			return nil, assert.AnError
		}

		// Create BridgeManager with module creator
		bm := NewBridgeManagerWithModuleCreator(converter, moduleCreator)

		// Create and register a test bridge
		bridge := testutils.NewMockBridge("test_bridge").WithInitialized(true)
		err := bm.RegisterBridge(bridge)
		require.NoError(t, err)

		// Create Lua module - should use adapter via callback
		module, err := bm.CreateLuaModule(L, "test_bridge")
		require.NoError(t, err)
		require.NotNil(t, module)

		// Verify the module came from adapter (has testMethod)
		testMethod := L.GetField(module, "testMethod")
		assert.Equal(t, lua.LTFunction, testMethod.Type(), "Module should have adapter's testMethod")

		// Test calling the adapter method
		L.SetGlobal("testModule", module)
		err = L.DoString(`
			local result = testModule.testMethod()
			assert(result == "adapter result", "Should return adapter result")
		`)
		assert.NoError(t, err)
	})

	t.Run("error handling when no adapter exists", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		converter := converters.NewLuaTypeConverter()

		// Create module creator that always returns error (no adapter for bridge)
		moduleCreator := func(bridgeID string) (lua.LGFunction, error) {
			return nil, assert.AnError
		}

		bm := NewBridgeManagerWithModuleCreator(converter, moduleCreator)

		// Register bridge
		bridge := testutils.NewMockBridge("test_bridge").WithInitialized(true)
		err := bm.RegisterBridge(bridge)
		require.NoError(t, err)

		// Try to create module - should fail since every bridge requires an adapter
		module, err := bm.CreateLuaModule(L, "test_bridge")
		assert.Error(t, err, "Should fail when no adapter exists for bridge")
		assert.Nil(t, module, "Should return nil module when adapter creation fails")
		assert.Contains(t, err.Error(), "failed to create adapter module", "Should indicate adapter creation failure")
	})

	t.Run("error when no module creator configured", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		converter := converters.NewLuaTypeConverter()

		// Create BridgeManager without module creator (nil) - this violates architecture
		bm := NewBridgeManager(converter)

		// Register bridge
		bridge := testutils.NewMockBridge("test_bridge").WithInitialized(true)
		err := bm.RegisterBridge(bridge)
		require.NoError(t, err)

		// Try to create module - should fail since no module creator configured
		module, err := bm.CreateLuaModule(L, "test_bridge")
		assert.Error(t, err, "Should fail when no module creator configured")
		assert.Nil(t, module, "Should return nil module")
		assert.Contains(t, err.Error(), "no module creator configured", "Should indicate missing module creator")
		assert.Contains(t, err.Error(), "every bridge requires an adapter", "Should mention adapter requirement")
	})
}

// TestBridgeManagerAdapterIntegration tests complete adapter integration
func TestBridgeManagerAdapterIntegration(t *testing.T) {
	t.Run("adapter module integration with engine", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		converter := converters.NewLuaTypeConverter()

		// Create adapter that provides enhanced functionality
		mockAdapter := &mockModuleAdapter{
			moduleFunc: func(L *lua.LState) int {
				module := L.NewTable()

				// Add adapter-enhanced method
				L.SetField(module, "enhancedMethod", L.NewFunction(func(L *lua.LState) int {
					arg := L.CheckString(1)
					result := "adapter enhanced: " + arg
					L.Push(lua.LString(result))
					return 1
				}))

				// Add adapter metadata
				meta := L.NewTable()
				L.SetField(meta, "source", lua.LString("adapter"))
				L.SetField(meta, "version", lua.LString("1.0.0"))
				L.SetField(module, "_adapter_meta", meta)

				L.Push(module)
				return 1
			},
		}

		// Create module creator
		moduleCreator := func(bridgeID string) (lua.LGFunction, error) {
			if bridgeID == "enhanced_bridge" {
				return mockAdapter.CreateLuaModule(), nil
			}
			return nil, assert.AnError
		}

		bm := NewBridgeManagerWithModuleCreator(converter, moduleCreator)

		// Register bridge
		bridge := testutils.NewMockBridge("enhanced_bridge").WithInitialized(true)
		err := bm.RegisterBridge(bridge)
		require.NoError(t, err)

		// Create module
		module, err := bm.CreateLuaModule(L, "enhanced_bridge")
		require.NoError(t, err)

		// Test adapter integration
		L.SetGlobal("enhancedModule", module)
		err = L.DoString(`
			-- Test adapter method
			local result = enhancedModule.enhancedMethod("test")
			assert(result == "adapter enhanced: test", "Should use adapter method")
			
			-- Test adapter metadata
			assert(enhancedModule._adapter_meta.source == "adapter", "Should have adapter metadata")
			assert(enhancedModule._adapter_meta.version == "1.0.0", "Should have version")
		`)
		assert.NoError(t, err)
	})

	t.Run("adapter error propagation", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		converter := converters.NewLuaTypeConverter()

		// Create adapter that can fail
		mockAdapter := &mockModuleAdapter{
			moduleFunc: func(L *lua.LState) int {
				module := L.NewTable()

				L.SetField(module, "failingMethod", L.NewFunction(func(L *lua.LState) int {
					L.RaiseError("adapter error: operation failed")
					return 0
				}))

				L.Push(module)
				return 1
			},
		}

		moduleCreator := func(bridgeID string) (lua.LGFunction, error) {
			return mockAdapter.CreateLuaModule(), nil
		}

		bm := NewBridgeManagerWithModuleCreator(converter, moduleCreator)

		// Register and create module
		bridge := testutils.NewMockBridge("failing_bridge").WithInitialized(true)
		err := bm.RegisterBridge(bridge)
		require.NoError(t, err)

		module, err := bm.CreateLuaModule(L, "failing_bridge")
		require.NoError(t, err)

		// Test error propagation
		L.SetGlobal("failingModule", module)
		err = L.DoString(`
			local success, err = pcall(function()
				failingModule.failingMethod()
			end)
			assert(success == false, "Should fail")
			assert(string.find(err, "adapter error"), "Should propagate adapter error")
		`)
		assert.NoError(t, err)
	})
}

// TestBridgeManagerModuleCreatorLifecycle tests module creator lifecycle management
func TestBridgeManagerModuleCreatorLifecycle(t *testing.T) {
	t.Run("module creator state management", func(t *testing.T) {
		converter := converters.NewLuaTypeConverter()

		callCount := 0
		moduleCreator := func(bridgeID string) (lua.LGFunction, error) {
			callCount++
			return func(L *lua.LState) int {
				module := L.NewTable()
				L.SetField(module, "callCount", lua.LNumber(callCount))
				L.Push(module)
				return 1
			}, nil
		}

		bm := NewBridgeManagerWithModuleCreator(converter, moduleCreator)

		// Register bridge
		bridge := testutils.NewMockBridge("state_bridge").WithInitialized(true)
		err := bm.RegisterBridge(bridge)
		require.NoError(t, err)

		// Create module multiple times
		L1 := lua.NewState()
		defer L1.Close()

		L2 := lua.NewState()
		defer L2.Close()

		// First call
		module1, err := bm.CreateLuaModule(L1, "state_bridge")
		require.NoError(t, err)

		// Second call (should be cached, not call creator again)
		module2, err := bm.CreateLuaModule(L2, "state_bridge")
		require.NoError(t, err)

		// Module creator should only be called once due to caching
		assert.Equal(t, 1, callCount, "Module creator should be called only once due to caching")

		// Both modules should be the same cached instance
		assert.Equal(t, module1, module2, "Should return same cached module")
	})

	t.Run("concurrent module creation", func(t *testing.T) {
		converter := converters.NewLuaTypeConverter()

		callCount := 0
		moduleCreator := func(bridgeID string) (lua.LGFunction, error) {
			callCount++
			return func(L *lua.LState) int {
				module := L.NewTable()
				L.SetField(module, "concurrentTest", lua.LBool(true))
				L.Push(module)
				return 1
			}, nil
		}

		bm := NewBridgeManagerWithModuleCreator(converter, moduleCreator)

		// Register bridge
		bridge := testutils.NewMockBridge("concurrent_bridge").WithInitialized(true)
		err := bm.RegisterBridge(bridge)
		require.NoError(t, err)

		// Test concurrent access
		done := make(chan *lua.LTable, 10)
		errors := make(chan error, 10)

		// Start multiple goroutines
		for i := 0; i < 10; i++ {
			go func() {
				L := lua.NewState()
				defer L.Close()

				module, err := bm.CreateLuaModule(L, "concurrent_bridge")
				if err != nil {
					errors <- err
					return
				}
				done <- module
			}()
		}

		// Collect results
		var modules []*lua.LTable
		var errs []error

		for i := 0; i < 10; i++ {
			select {
			case module := <-done:
				modules = append(modules, module)
			case err := <-errors:
				errs = append(errs, err)
			}
		}

		// Should have no errors
		assert.Empty(t, errs, "Should have no errors in concurrent access")
		assert.Equal(t, 10, len(modules), "Should have all modules")

		// All modules should be the same due to caching
		for i := 1; i < len(modules); i++ {
			assert.Equal(t, modules[0], modules[i], "All modules should be the same cached instance")
		}
	})
}

// mockModuleAdapter implements a test adapter that creates Lua modules
type mockModuleAdapter struct {
	moduleFunc lua.LGFunction
}

func (ma *mockModuleAdapter) CreateLuaModule() lua.LGFunction {
	return ma.moduleFunc
}
