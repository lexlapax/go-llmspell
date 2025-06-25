// ABOUTME: Tests for bridge type conversion handlers - Bridge to LUserData, metatable generation, method wrapping
// ABOUTME: Validates bridge object conversions, method exposure, type safety, and bridge type registry

package gopherlua

import (
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

// MockBridge implementation moved to test_helpers.go

func TestBridgeConverter_BasicOperations(t *testing.T) {
	converter := NewBridgeConverter()
	L := lua.NewState()
	defer L.Close()

	bridge := NewMockBridge("test_bridge").WithMetadata(engine.BridgeMetadata{
		Name:        "Test Bridge",
		Version:     "1.0.0",
		Description: "A bridge for testing",
	})

	t.Run("bridge_to_userdata", func(t *testing.T) {
		result, err := converter.BridgeToLua(L, bridge)
		require.NoError(t, err)

		userdata, ok := result.(*lua.LUserData)
		require.True(t, ok, "Result should be LUserData")
		assert.Equal(t, bridge, userdata.Value)
		assert.NotNil(t, userdata.Metatable)
	})

	t.Run("userdata_to_bridge", func(t *testing.T) {
		userdata := L.NewUserData()
		userdata.Value = bridge

		result, err := converter.FromLua(userdata)
		require.NoError(t, err)

		convertedBridge, ok := result.(*MockBridge)
		require.True(t, ok, "Result should be MockBridge")
		assert.Equal(t, bridge.GetID(), convertedBridge.GetID())
	})

	t.Run("non_bridge_to_userdata_error", func(t *testing.T) {
		_, err := converter.BridgeToLua(L, "not a bridge")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected bridge")
	})
}

func TestBridgeConverter_MetatableGeneration(t *testing.T) {
	converter := NewBridgeConverter()
	L := lua.NewState()
	defer L.Close()

	bridge := NewMockBridge("test_bridge").WithMetadata(engine.BridgeMetadata{
		Name:        "Test Bridge",
		Version:     "1.0.0",
		Description: "A bridge for testing",
	})

	t.Run("metatable_contains_methods", func(t *testing.T) {
		metatable := converter.GenerateMetatable(L, bridge)
		require.NotNil(t, metatable)

		// Check if testMethod is available
		testMethod := metatable.RawGetString("testMethod")
		assert.NotEqual(t, lua.LNil, testMethod)
		assert.Equal(t, lua.LTFunction, testMethod.Type())

		// Check if calculateSum is available
		calculateSum := metatable.RawGetString("calculateSum")
		assert.NotEqual(t, lua.LNil, calculateSum)
		assert.Equal(t, lua.LTFunction, calculateSum.Type())
	})

	t.Run("metatable_has_type_info", func(t *testing.T) {
		metatable := converter.GenerateMetatable(L, bridge)
		require.NotNil(t, metatable)

		// Check bridge type info
		bridgeType := metatable.RawGetString("__type")
		assert.Equal(t, "bridge", bridgeType.String())

		bridgeID := metatable.RawGetString("__bridge_id")
		assert.Equal(t, "test_bridge", bridgeID.String())
	})

	t.Run("metatable_tostring_method", func(t *testing.T) {
		metatable := converter.GenerateMetatable(L, bridge)
		require.NotNil(t, metatable)

		tostring := metatable.RawGetString("__tostring")
		assert.NotEqual(t, lua.LNil, tostring)
		assert.Equal(t, lua.LTFunction, tostring.Type())
	})
}

func TestBridgeConverter_MethodWrapping(t *testing.T) {
	converter := NewBridgeConverter()
	L := lua.NewState()
	defer L.Close()

	bridge := NewMockBridge("test_bridge").WithMetadata(engine.BridgeMetadata{
		Name:        "Test Bridge",
		Version:     "1.0.0",
		Description: "A bridge for testing",
	})

	t.Run("wrap_method_basic", func(t *testing.T) {
		methodInfo := engine.MethodInfo{
			Name:        "testMethod",
			Description: "A test method",
			Parameters: []engine.ParameterInfo{
				{Name: "input", Type: "string", Required: true},
			},
			ReturnType: "string",
		}

		wrappedFn := converter.WrapMethod(L, bridge, methodInfo)
		assert.NotNil(t, wrappedFn)
		assert.Equal(t, lua.LTFunction, wrappedFn.Type())
	})

	t.Run("validate_method_parameters", func(t *testing.T) {
		methodInfo := engine.MethodInfo{
			Name: "calculateSum",
			Parameters: []engine.ParameterInfo{
				{Name: "a", Type: "number", Required: true},
				{Name: "b", Type: "number", Required: true},
			},
		}

		// Valid parameters
		err := converter.ValidateMethodCall("calculateSum", []lua.LValue{
			lua.LNumber(10),
			lua.LNumber(20),
		}, methodInfo)
		assert.NoError(t, err)

		// Missing parameters
		err = converter.ValidateMethodCall("calculateSum", []lua.LValue{
			lua.LNumber(10),
		}, methodInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected 2 arguments")

		// Wrong type
		err = converter.ValidateMethodCall("calculateSum", []lua.LValue{
			lua.LString("not a number"),
			lua.LNumber(20),
		}, methodInfo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected number")
	})
}

func TestBridgeConverter_TypeSafety(t *testing.T) {
	converter := NewBridgeConverter()
	L := lua.NewState()
	defer L.Close()

	bridge := &MockBridge{
		id: "test_bridge",
	}

	t.Run("type_safety_checks", func(t *testing.T) {
		// Test type checking for bridge objects
		assert.True(t, converter.IsBridge(bridge))
		assert.False(t, converter.IsBridge("not a bridge"))
		assert.False(t, converter.IsBridge(nil))

		// Test bridge validation
		err := converter.ValidateBridge(bridge)
		assert.NoError(t, err)

		err = converter.ValidateBridge(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bridge cannot be nil")
	})

	t.Run("userdata_type_checking", func(t *testing.T) {
		userdata := L.NewUserData()
		userdata.Value = bridge

		// Valid bridge userdata
		assert.True(t, converter.IsValidBridgeUserData(userdata))

		// Invalid userdata (wrong value type)
		invalidUserdata := L.NewUserData()
		invalidUserdata.Value = "not a bridge"
		assert.False(t, converter.IsValidBridgeUserData(invalidUserdata))

		// Nil userdata
		assert.False(t, converter.IsValidBridgeUserData(nil))
	})
}

func TestBridgeConverter_BridgeRegistry(t *testing.T) {
	converter := NewBridgeConverter()
	L := lua.NewState()
	defer L.Close()

	bridge1 := &MockBridge{id: "bridge1"}
	bridge2 := &MockBridge{id: "bridge2"}

	t.Run("register_bridge_type", func(t *testing.T) {
		err := converter.RegisterBridgeType("MockBridge", bridge1)
		assert.NoError(t, err)

		// Try to register the same type again
		err = converter.RegisterBridgeType("MockBridge", bridge2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("get_registered_bridge_type", func(t *testing.T) {
		bridge, exists := converter.GetBridgeType("MockBridge")
		assert.True(t, exists)
		assert.Equal(t, "bridge1", bridge.GetID())

		_, exists = converter.GetBridgeType("NonExistentBridge")
		assert.False(t, exists)
	})

	t.Run("list_bridge_types", func(t *testing.T) {
		types := converter.ListBridgeTypes()
		assert.Contains(t, types, "MockBridge")
	})

	t.Run("unregister_bridge_type", func(t *testing.T) {
		err := converter.UnregisterBridgeType("MockBridge")
		assert.NoError(t, err)

		_, exists := converter.GetBridgeType("MockBridge")
		assert.False(t, exists)

		// Try to unregister non-existent type
		err = converter.UnregisterBridgeType("NonExistentBridge")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not registered")
	})
}

func TestBridgeConverter_ErrorHandling(t *testing.T) {
	converter := NewBridgeConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name      string
		operation func() error
		errorText string
	}{
		{
			name: "invalid_bridge_conversion",
			operation: func() error {
				_, err := converter.BridgeToLua(L, 123)
				return err
			},
			errorText: "expected bridge",
		},
		{
			name: "invalid_userdata_conversion",
			operation: func() error {
				_, err := converter.FromLua(lua.LString("not userdata"))
				return err
			},
			errorText: "expected userdata",
		},
		{
			name: "invalid_bridge_userdata",
			operation: func() error {
				userdata := L.NewUserData()
				userdata.Value = "not a bridge"
				_, err := converter.FromLua(userdata)
				return err
			},
			errorText: "does not contain a valid bridge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorText)
		})
	}
}

func TestBridgeConverter_ConcurrentAccess(t *testing.T) {
	converter := NewBridgeConverter()
	L := lua.NewState()
	defer L.Close()

	bridge := &MockBridge{id: "concurrent_bridge"}

	t.Run("concurrent_bridge_registration", func(t *testing.T) {
		// Test concurrent registration/access
		numGoroutines := 10
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				testBridge := &MockBridge{id: string(rune('a' + id))}
				err := converter.RegisterBridgeType(string(rune('A'+id)), testBridge)
				errors <- err
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numGoroutines; i++ {
			err := <-errors
			if err == nil {
				successCount++
			}
		}

		assert.Equal(t, numGoroutines, successCount, "All registrations should succeed with unique types")
	})

	t.Run("concurrent_bridge_conversion", func(t *testing.T) {
		numGoroutines := 10
		results := make(chan lua.LValue, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				result, err := converter.BridgeToLua(L, bridge)
				results <- result
				errors <- err
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < numGoroutines; i++ {
			err := <-errors
			<-results
			if err == nil {
				successCount++
			}
		}

		assert.Equal(t, numGoroutines, successCount, "All conversions should succeed")
	})
}
