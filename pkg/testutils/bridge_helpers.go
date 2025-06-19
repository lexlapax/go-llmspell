// ABOUTME: Bridge test helpers provide common setup, teardown, and verification patterns for bridge testing
// ABOUTME: Simplifies bridge initialization, cleanup, and assertion across all bridge test files

package testutils

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupTestBridge initializes a bridge and returns a cleanup function
// Usage:
//
//	cleanup := SetupTestBridge(t, bridge)
//	defer cleanup()
func SetupTestBridge(t *testing.T, bridge engine.Bridge) func() {
	t.Helper()

	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err, "Failed to initialize bridge %s", bridge.GetID())

	return func() {
		if bridge.IsInitialized() {
			err := bridge.Cleanup(ctx)
			assert.NoError(t, err, "Failed to cleanup bridge %s", bridge.GetID())
		}
	}
}

// SetupTestBridgeWithEngine initializes a bridge with a mock engine and returns both
// Usage:
//
//	engine, cleanup := SetupTestBridgeWithEngine(t, bridge)
//	defer cleanup()
func SetupTestBridgeWithEngine(t *testing.T, bridge engine.Bridge) (*MockScriptEngine, func()) {
	t.Helper()

	// Create and initialize mock engine
	mockEngine := NewMockScriptEngine()
	err := mockEngine.Initialize(engine.EngineConfig{
		MemoryLimit:  1024 * 1024,
		TimeoutLimit: 0, // No timeout for tests
	})
	require.NoError(t, err, "Failed to initialize mock engine")

	// Initialize bridge
	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err, "Failed to initialize bridge %s", bridge.GetID())

	// Register bridge with engine
	err = bridge.RegisterWithEngine(mockEngine)
	require.NoError(t, err, "Failed to register bridge %s with engine", bridge.GetID())

	return mockEngine, func() {
		// Cleanup bridge first
		if bridge.IsInitialized() {
			err := bridge.Cleanup(ctx)
			assert.NoError(t, err, "Failed to cleanup bridge %s", bridge.GetID())
		}

		// Then shutdown engine
		err := mockEngine.Shutdown()
		assert.NoError(t, err, "Failed to shutdown mock engine")
	}
}

// AssertBridgeInitialized verifies bridge initialization state
func AssertBridgeInitialized(t *testing.T, bridge engine.Bridge) {
	t.Helper()
	assert.True(t, bridge.IsInitialized(), "Bridge %s should be initialized", bridge.GetID())
}

// AssertBridgeNotInitialized verifies bridge is not initialized
func AssertBridgeNotInitialized(t *testing.T, bridge engine.Bridge) {
	t.Helper()
	assert.False(t, bridge.IsInitialized(), "Bridge %s should not be initialized", bridge.GetID())
}

// AssertBridgeMethod verifies a bridge method exists and has correct info
func AssertBridgeMethod(t *testing.T, bridge engine.Bridge, methodName string, expectedParams int) {
	t.Helper()

	methods := bridge.Methods()
	var found *engine.MethodInfo

	for _, method := range methods {
		if method.Name == methodName {
			found = &method
			break
		}
	}

	require.NotNil(t, found, "Method %s not found in bridge %s", methodName, bridge.GetID())

	if expectedParams >= 0 {
		assert.Len(t, found.Parameters, expectedParams,
			"Method %s should have %d parameters", methodName, expectedParams)
	}
}

// AssertBridgeHasMethod verifies a bridge has a specific method
func AssertBridgeHasMethod(t *testing.T, bridge engine.Bridge, methodName string) {
	t.Helper()
	AssertBridgeMethod(t, bridge, methodName, -1) // -1 means don't check param count
}

// AssertBridgeMethodCount verifies the total number of methods
func AssertBridgeMethodCount(t *testing.T, bridge engine.Bridge, expectedCount int) {
	t.Helper()
	methods := bridge.Methods()
	assert.Len(t, methods, expectedCount,
		"Bridge %s should have %d methods", bridge.GetID(), expectedCount)
}

// AssertBridgeMetadata verifies bridge metadata fields
func AssertBridgeMetadata(t *testing.T, bridge engine.Bridge, expectedName, expectedVersion string) {
	t.Helper()

	metadata := bridge.GetMetadata()
	assert.Equal(t, expectedName, metadata.Name, "Bridge name mismatch")
	assert.Equal(t, expectedVersion, metadata.Version, "Bridge version mismatch")
	assert.NotEmpty(t, metadata.Description, "Bridge description should not be empty")
}

// AssertBridgePermissions verifies bridge required permissions
func AssertBridgePermissions(t *testing.T, bridge engine.Bridge, expectedTypes ...engine.PermissionType) {
	t.Helper()

	permissions := bridge.RequiredPermissions()
	permTypes := make(map[engine.PermissionType]bool)

	for _, perm := range permissions {
		permTypes[perm.Type] = true
	}

	for _, expectedType := range expectedTypes {
		assert.True(t, permTypes[expectedType],
			"Bridge %s should require permission %s", bridge.GetID(), expectedType)
	}
}

// TestBridgeMethodExecution is a helper for testing method execution
type TestBridgeMethodExecution struct {
	Bridge      engine.Bridge
	MethodName  string
	Args        []engine.ScriptValue
	Expected    engine.ScriptValue
	ExpectErr   bool
	ErrContains string
}

// ExecuteBridgeMethodTest runs a single bridge method test
func ExecuteBridgeMethodTest(t *testing.T, test TestBridgeMethodExecution) {
	t.Helper()

	ctx := context.Background()
	result, err := test.Bridge.ExecuteMethod(ctx, test.MethodName, test.Args)

	if test.ExpectErr {
		require.Error(t, err, "Expected error for method %s", test.MethodName)
		if test.ErrContains != "" {
			assert.Contains(t, err.Error(), test.ErrContains,
				"Error message should contain expected text")
		}
	} else {
		require.NoError(t, err, "Unexpected error for method %s", test.MethodName)
		if test.Expected != nil {
			assert.Equal(t, test.Expected, result,
				"Method %s returned unexpected result", test.MethodName)
		}
	}
}

// ValidateBridgeInterface performs comprehensive validation of a bridge implementation
func ValidateBridgeInterface(t *testing.T, bridge engine.Bridge) {
	t.Helper()

	// Check basic interface compliance
	assert.NotEmpty(t, bridge.GetID(), "Bridge ID should not be empty")

	metadata := bridge.GetMetadata()
	assert.NotEmpty(t, metadata.Name, "Bridge name should not be empty")
	assert.NotEmpty(t, metadata.Version, "Bridge version should not be empty")

	// Check methods
	methods := bridge.Methods()
	assert.NotNil(t, methods, "Methods should not return nil")

	// Check that each method has required fields
	for _, method := range methods {
		assert.NotEmpty(t, method.Name, "Method name should not be empty")
		assert.NotEmpty(t, method.Description, "Method %s should have description", method.Name)
		assert.NotEmpty(t, method.ReturnType, "Method %s should have return type", method.Name)

		// Validate parameters
		for i, param := range method.Parameters {
			assert.NotEmpty(t, param.Name,
				"Parameter %d of method %s should have name", i, method.Name)
			assert.NotEmpty(t, param.Type,
				"Parameter %s of method %s should have type", param.Name, method.Name)
		}
	}

	// Check permissions and type mappings (can be empty but not nil)
	assert.NotNil(t, bridge.RequiredPermissions(), "RequiredPermissions should not return nil")
	assert.NotNil(t, bridge.TypeMappings(), "TypeMappings should not return nil")
}

// BridgeTestConfig provides configuration for bridge tests
type BridgeTestConfig struct {
	Bridge          engine.Bridge
	RequiresInit    bool
	ExpectedMethods []string
	SkipValidation  bool
}

// RunBridgeTestSuite runs a standard suite of tests for a bridge
func RunBridgeTestSuite(t *testing.T, config BridgeTestConfig) {
	t.Helper()

	if !config.SkipValidation {
		t.Run("ValidateInterface", func(t *testing.T) {
			ValidateBridgeInterface(t, config.Bridge)
		})
	}

	if config.RequiresInit {
		t.Run("Initialization", func(t *testing.T) {
			ctx := context.Background()

			// Should not be initialized initially
			AssertBridgeNotInitialized(t, config.Bridge)

			// Initialize
			err := config.Bridge.Initialize(ctx)
			require.NoError(t, err)
			AssertBridgeInitialized(t, config.Bridge)

			// Double init should fail
			err = config.Bridge.Initialize(ctx)
			assert.Error(t, err)

			// Cleanup
			err = config.Bridge.Cleanup(ctx)
			require.NoError(t, err)
			AssertBridgeNotInitialized(t, config.Bridge)
		})
	}

	if len(config.ExpectedMethods) > 0 {
		t.Run("Methods", func(t *testing.T) {
			for _, methodName := range config.ExpectedMethods {
				AssertBridgeHasMethod(t, config.Bridge, methodName)
			}
			AssertBridgeMethodCount(t, config.Bridge, len(config.ExpectedMethods))
		})
	}
}

// CreateTestContext creates a context with common test values
func CreateTestContext(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()
	// Add common test context values if needed
	return ctx
}

// SetupMultipleBridges initializes multiple bridges and returns a cleanup function
func SetupMultipleBridges(t *testing.T, bridges ...engine.Bridge) func() {
	t.Helper()

	ctx := context.Background()
	initialized := make([]engine.Bridge, 0, len(bridges))

	// Initialize all bridges
	for _, bridge := range bridges {
		err := bridge.Initialize(ctx)
		if err != nil {
			// Cleanup already initialized bridges
			for _, b := range initialized {
				_ = b.Cleanup(ctx)
			}
			require.NoError(t, err, "Failed to initialize bridge %s", bridge.GetID())
		}
		initialized = append(initialized, bridge)
	}

	// Return cleanup function
	return func() {
		for _, bridge := range initialized {
			if bridge.IsInitialized() {
				err := bridge.Cleanup(ctx)
				assert.NoError(t, err, "Failed to cleanup bridge %s", bridge.GetID())
			}
		}
	}
}

// AssertMethodValidation checks that a method properly validates its arguments
func AssertMethodValidation(t *testing.T, bridge engine.Bridge, methodName string, args []engine.ScriptValue, shouldFail bool) {
	t.Helper()

	err := bridge.ValidateMethod(methodName, args)
	if shouldFail {
		assert.Error(t, err, "Method %s should fail validation with given args", methodName)
	} else {
		assert.NoError(t, err, "Method %s should pass validation with given args", methodName)
	}
}

// BridgeMethodTestCase represents a test case for bridge method execution
type BridgeMethodTestCase struct {
	Name        string
	Method      string
	Args        []engine.ScriptValue
	Setup       func(t *testing.T)
	ExpectError bool
	ExpectType  engine.ScriptValueType
	Validate    func(t *testing.T, result engine.ScriptValue)
}

// RunBridgeMethodTests executes a table of method test cases
func RunBridgeMethodTests(t *testing.T, bridge engine.Bridge, tests []BridgeMethodTestCase) {
	t.Helper()

	cleanup := SetupTestBridge(t, bridge)
	defer cleanup()

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Setup != nil {
				tt.Setup(t)
			}

			result, err := bridge.ExecuteMethod(ctx, tt.Method, tt.Args)

			if tt.ExpectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.ExpectType != 0 {
					require.NotNil(t, result)
					assert.Equal(t, tt.ExpectType, result.Type(),
						"Expected result type %v, got %v", tt.ExpectType, result.Type())
				}

				if tt.Validate != nil {
					tt.Validate(t, result)
				}
			}
		})
	}
}
