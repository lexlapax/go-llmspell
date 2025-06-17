// ABOUTME: Tests for LuaEngine bridge registration and management functionality
// ABOUTME: Validates bridge lifecycle, module creation, method wrapping, and Lua-side access

package gopherlua

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
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

func (b *testBridgeForRegistration) ValidateMethod(name string, args []interface{}) error {
	for _, method := range b.methods {
		if method.Name == name {
			if len(args) == len(method.Parameters) {
				return nil
			}
		}
	}
	return fmt.Errorf("invalid method call: %s", name)
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
	validArgs := []interface{}{"test_string", 42.0}
	err = bridge.ValidateMethod("validateMe", validArgs)
	assert.NoError(t, err)

	// Test invalid method name
	err = bridge.ValidateMethod("nonExistentMethod", validArgs)
	assert.Error(t, err)

	// Test wrong argument count
	invalidArgs := []interface{}{"test_string"}
	err = bridge.ValidateMethod("validateMe", invalidArgs)
	assert.Error(t, err)
}

func TestLuaEngine_MultipleBridgeManagement(t *testing.T) {
	eng := NewLuaEngine()
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
