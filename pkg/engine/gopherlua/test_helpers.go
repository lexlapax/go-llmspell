// ABOUTME: GopherLua test helpers providing ScriptValue creation and bridge testing utilities
// ABOUTME: Avoids import cycles by providing package-local helper functions for consistent testing

package gopherlua

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Helper functions for ScriptValue creation - mirrors testutils pattern
func sv(value interface{}) engine.ScriptValue {
	return engine.ConvertToScriptValue(value)
}

// Note: svMap and svArray helpers available when needed for migrations

// Bridge test helpers

// MockBridge provides a simple bridge implementation for testing
type MockBridge struct {
	id               string
	metadata         engine.BridgeMetadata
	methods          []engine.MethodInfo
	callFunc         func(method string, args ...interface{}) (interface{}, error)
	validateFunc     func(method string, args []engine.ScriptValue) error
	permissionsFunc  func() []engine.Permission
	typeMappingsFunc func() map[string]engine.TypeMapping
}

func NewMockBridge(id string) *MockBridge {
	return &MockBridge{
		id: id,
		metadata: engine.BridgeMetadata{
			Name:        "Mock Bridge",
			Version:     "1.0.0",
			Description: "A mock bridge for testing",
		},
		methods: []engine.MethodInfo{
			{Name: "testMethod", Description: "Test method", ReturnType: "string"},
			{Name: "mathOperation", Description: "A math operation method", ReturnType: "number"},
			{Name: "calculateSum", Description: "Calculate sum", ReturnType: "number"},
			{Name: "getValue", Description: "Get a value", ReturnType: "string"},
		},
		callFunc: func(method string, args ...interface{}) (interface{}, error) {
			switch method {
			case "testMethod":
				return "test result", nil
			case "mathOperation":
				return 42, nil
			case "calculateSum":
				return 100, nil
			case "getValue":
				return "mock value", nil
			default:
				return "mock result", nil
			}
		},
		validateFunc: func(method string, args []engine.ScriptValue) error {
			return nil
		},
		permissionsFunc: func() []engine.Permission {
			return []engine.Permission{}
		},
		typeMappingsFunc: func() map[string]engine.TypeMapping {
			return map[string]engine.TypeMapping{}
		},
	}
}

func (m *MockBridge) GetID() string {
	return m.id
}

func (m *MockBridge) GetMetadata() engine.BridgeMetadata {
	return m.metadata
}

func (m *MockBridge) Methods() []engine.MethodInfo {
	return m.methods
}

func (m *MockBridge) ValidateMethod(method string, args []engine.ScriptValue) error {
	return m.validateFunc(method, args)
}

func (m *MockBridge) ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Convert args to interface{} for callFunc
	interfaceArgs := make([]interface{}, len(args))
	for i, arg := range args {
		interfaceArgs[i] = arg.ToGo()
	}

	result, err := m.callFunc(method, interfaceArgs...)
	if err != nil {
		return sv(nil), err
	}

	return sv(result), nil
}

func (m *MockBridge) Initialize(ctx context.Context) error {
	return nil
}

func (m *MockBridge) Cleanup(ctx context.Context) error {
	return nil
}

func (m *MockBridge) IsInitialized() bool {
	return true
}

func (m *MockBridge) RequiredPermissions() []engine.Permission {
	return m.permissionsFunc()
}

func (m *MockBridge) TypeMappings() map[string]engine.TypeMapping {
	return m.typeMappingsFunc()
}

func (m *MockBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return nil
}

// Builder methods
func (m *MockBridge) WithMethods(methods []engine.MethodInfo) *MockBridge {
	m.methods = methods
	return m
}

func (m *MockBridge) WithCallFunc(f func(method string, args ...interface{}) (interface{}, error)) *MockBridge {
	m.callFunc = f
	return m
}

func (m *MockBridge) WithValidateFunc(f func(method string, args []engine.ScriptValue) error) *MockBridge {
	m.validateFunc = f
	return m
}

func (m *MockBridge) WithMetadata(metadata engine.BridgeMetadata) *MockBridge {
	m.metadata = metadata
	return m
}

// Test execution helpers
func TestBridgeExecution(t *testing.T, bridge engine.Bridge, method string, args []engine.ScriptValue, expectedResult interface{}) {
	t.Helper()

	ctx := context.Background()
	result, err := bridge.ExecuteMethod(ctx, method, args)
	require.NoError(t, err, "Bridge method execution should succeed")

	if expectedResult != nil {
		expected := sv(expectedResult)
		assert.Equal(t, expected.ToGo(), result.ToGo(), "Result should match expected value")
	}
}

func TestBridgeValidation(t *testing.T, bridge engine.Bridge, method string, args []engine.ScriptValue, expectError bool) {
	t.Helper()

	err := bridge.ValidateMethod(method, args)
	if expectError {
		assert.Error(t, err, "Validation should fail")
	} else {
		assert.NoError(t, err, "Validation should succeed")
	}
}

// Assertion helpers for ScriptValues
func AssertScriptValueType(t *testing.T, result engine.ScriptValue, expectedType engine.ScriptValueType) {
	t.Helper()
	require.NotNil(t, result, "ScriptValue should not be nil")
	assert.Equal(t, expectedType, result.Type(), "Expected type %s, got %s", expectedType, result.Type())
}

func AssertScriptValueEquals(t *testing.T, expected, actual engine.ScriptValue) {
	t.Helper()
	require.NotNil(t, expected, "Expected value should not be nil")
	require.NotNil(t, actual, "Actual value should not be nil")
	assert.Equal(t, expected.Type(), actual.Type(), "Types should match")
	assert.Equal(t, expected.ToGo(), actual.ToGo(), "Values should match")
}

// Context helpers
func TestContextWithTimeout(timeout string) context.Context {
	return context.Background() // Simplified for now
}

func TestContext() context.Context {
	return context.Background()
}
