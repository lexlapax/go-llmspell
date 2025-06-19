// ABOUTME: Tests for debug logging bridge functionality with ScriptValue-based API
// ABOUTME: Tests ExecuteMethod dispatcher and individual debug methods

package util

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugBridgeInitialization(t *testing.T) {
	bridge := NewDebugBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "debug", bridge.GetID())
	assert.False(t, bridge.IsInitialized())

	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test double initialization
	err = bridge.Initialize(ctx)
	assert.NoError(t, err)

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestDebugBridgeMetadata(t *testing.T) {
	bridge := NewDebugBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "debug", metadata.Name)
	assert.NotEmpty(t, metadata.Version)
	assert.NotEmpty(t, metadata.Description)
	assert.NotEmpty(t, metadata.Author)
	assert.NotEmpty(t, metadata.License)
}

func TestDebugBridgeMethods(t *testing.T) {
	bridge := NewDebugBridge()
	methods := bridge.Methods()

	// Check that all expected methods are present
	expectedMethods := []string{
		"debugPrintf",
		"debugPrintln",
		"isDebugEnabled",
		"enableDebugComponent",
		"disableDebugComponent",
		"listEnabledComponents",
		"setCustomLogger",
		"getDebugEnvironment",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Method %s not found", expected)
	}
}

func TestDebugBridgeExecuteMethod(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		args        []engine.ScriptValue
		wantErr     bool
		checkResult func(t *testing.T, result engine.ScriptValue, err error)
	}{
		{
			name:   "enableDebugComponent",
			method: "enableDebugComponent",
			args: []engine.ScriptValue{
				sv("test-component"),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				assert.NotNil(t, result)
				assert.True(t, result.IsNil())
			},
		},
		{
			name:   "isDebugEnabled - after enable",
			method: "isDebugEnabled",
			args: []engine.ScriptValue{
				sv("test-component"),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				require.NotNil(t, result)
				assert.Equal(t, engine.TypeBool, result.Type())
				assert.True(t, result.(engine.BoolValue).Value())
			},
		},
		{
			name:   "debugPrintln",
			method: "debugPrintln",
			args: []engine.ScriptValue{
				sv("test-component"),
				sv("test message"),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				assert.NotNil(t, result)
				assert.True(t, result.IsNil())
			},
		},
		{
			name:   "debugPrintf",
			method: "debugPrintf",
			args: []engine.ScriptValue{
				sv("test-component"),
				sv("test %s %d"),
				svArray("hello", 42),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				assert.NotNil(t, result)
				assert.True(t, result.IsNil())
			},
		},
		{
			name:    "listEnabledComponents",
			method:  "listEnabledComponents",
			args:    []engine.ScriptValue{},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				require.NotNil(t, result)
				assert.Equal(t, engine.TypeArray, result.Type())
				array := result.(engine.ArrayValue).Elements()
				assert.GreaterOrEqual(t, len(array), 1) // Should have at least test-component
				found := false
				for _, elem := range array {
					if elem.Type() == engine.TypeString && elem.(engine.StringValue).Value() == "test-component" {
						found = true
						break
					}
				}
				assert.True(t, found, "test-component not found in enabled components")
			},
		},
		{
			name:   "disableDebugComponent",
			method: "disableDebugComponent",
			args: []engine.ScriptValue{
				sv("test-component"),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				assert.NotNil(t, result)
				assert.True(t, result.IsNil())
			},
		},
		{
			name:   "isDebugEnabled - after disable",
			method: "isDebugEnabled",
			args: []engine.ScriptValue{
				sv("test-component"),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				require.NotNil(t, result)
				assert.Equal(t, engine.TypeBool, result.Type())
				assert.False(t, result.(engine.BoolValue).Value())
			},
		},
		{
			name:   "setCustomLogger",
			method: "setCustomLogger",
			args: []engine.ScriptValue{
				svMap(map[string]interface{}{
					"prefix": "[CUSTOM]",
					"flags":  "datetime",
				}),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				assert.NotNil(t, result)
				assert.True(t, result.IsNil())
			},
		},
		{
			name:    "getDebugEnvironment",
			method:  "getDebugEnvironment",
			args:    []engine.ScriptValue{},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				require.NotNil(t, result)
				assert.Equal(t, engine.TypeObject, result.Type())
				obj := result.(engine.ObjectValue).Fields()

				// Check expected fields
				assert.Contains(t, obj, "go_llms_debug_env")
				assert.Contains(t, obj, "enabled_components")
				assert.Contains(t, obj, "compilation_mode")

				// Verify types
				assert.Equal(t, engine.TypeString, obj["go_llms_debug_env"].Type())
				assert.Equal(t, engine.TypeArray, obj["enabled_components"].Type())
				assert.Equal(t, engine.TypeString, obj["compilation_mode"].Type())
			},
		},
		{
			name:    "unknown method",
			method:  "unknownMethod",
			args:    []engine.ScriptValue{},
			wantErr: true,
			checkResult: func(t *testing.T, result engine.ScriptValue, err error) {
				assert.Nil(t, result)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown method")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.checkResult != nil {
				tt.checkResult(t, result, err)
			}
		})
	}
}

func TestDebugBridgeValidateMethod(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		args        []engine.ScriptValue
		shouldError bool
	}{
		{
			name:   "valid debugPrintf with all args",
			method: "debugPrintf",
			args: []engine.ScriptValue{
				sv("component"),
				sv("format"),
				svArray(),
			},
			shouldError: false,
		},
		{
			name:   "valid debugPrintf without optional args",
			method: "debugPrintf",
			args: []engine.ScriptValue{
				sv("component"),
				sv("format"),
			},
			shouldError: false,
		},
		{
			name:        "invalid debugPrintf - missing required args",
			method:      "debugPrintf",
			args:        []engine.ScriptValue{sv("component")},
			shouldError: true,
		},
		{
			name:   "valid debugPrintln",
			method: "debugPrintln",
			args: []engine.ScriptValue{
				sv("component"),
				sv("message"),
			},
			shouldError: false,
		},
		{
			name:        "invalid debugPrintln - missing args",
			method:      "debugPrintln",
			args:        []engine.ScriptValue{},
			shouldError: true,
		},
		{
			name:        "unknown method",
			method:      "unknownMethod",
			args:        []engine.ScriptValue{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDebugBridgeTypeConversions(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test debugPrintf with various argument types
	testCases := []struct {
		name   string
		method string
		args   []engine.ScriptValue
	}{
		{
			name:   "string format args",
			method: "debugPrintf",
			args: []engine.ScriptValue{
				sv("test"),
				sv("String: %s"),
				svArray("hello"),
			},
		},
		{
			name:   "number format args",
			method: "debugPrintf",
			args: []engine.ScriptValue{
				sv("test"),
				sv("Number: %d, Float: %f"),
				svArray(42, 3.14),
			},
		},
		{
			name:   "mixed format args",
			method: "debugPrintf",
			args: []engine.ScriptValue{
				sv("test"),
				sv("Mixed: %s %d %v"),
				svArray("hello", 42, true),
			},
		},
	}

	// Enable test component
	_, err = bridge.ExecuteMethod(ctx, "enableDebugComponent", []engine.ScriptValue{
		sv("test"),
	})
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tc.method, tc.args)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsNil())
		})
	}
}

func TestDebugBridgeRequiredPermissions(t *testing.T) {
	bridge := NewDebugBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 2)

	// Check for expected permissions
	hasStorage := false
	hasMemory := false

	for _, perm := range permissions {
		if perm.Type == engine.PermissionStorage && perm.Resource == "debug.logging" {
			hasStorage = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
		if perm.Type == engine.PermissionMemory && perm.Resource == "debug.components" {
			hasMemory = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
	}

	assert.True(t, hasStorage, "Storage permission not found")
	assert.True(t, hasMemory, "Memory permission not found")
}

func TestDebugBridgeTypeMappings(t *testing.T) {
	bridge := NewDebugBridge()
	mappings := bridge.TypeMappings()

	// Check expected type mappings
	assert.Contains(t, mappings, "debug_logger")
	assert.Contains(t, mappings, "debug_config")

	// Verify mapping properties
	loggerMapping := mappings["debug_logger"]
	assert.Equal(t, "*log.Logger", loggerMapping.GoType)
	assert.Equal(t, "object", loggerMapping.ScriptType)

	configMapping := mappings["debug_config"]
	assert.Equal(t, "map[string]interface{}", configMapping.GoType)
	assert.Equal(t, "object", configMapping.ScriptType)
}
