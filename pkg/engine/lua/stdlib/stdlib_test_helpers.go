// ABOUTME: Test infrastructure helpers for go-llmspell Lua standard library testing
// ABOUTME: Provides utilities for module loading, table comparison, async testing, error assertions, and mock bridges

// Package stdlib provides the Lua standard library implementations for go-llmspell.
// This file contains comprehensive test helpers for testing Lua modules, including
// utilities for module loading, Lua value comparison, async testing, error handling,
// mock bridge creation, and test fixture management.
package stdlib

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
	lua "github.com/yuin/gopher-lua"
)

// ============================================================================
// Lua Module Loading Helpers
// ============================================================================

// LoadModule loads a Lua module from a file and sets it as a global variable.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to load the module into
//   - moduleName: Name of the module (without .lua extension)
//
// Returns:
//   - lua.LValue: The loaded module value
//
// The function looks for a file named moduleName.lua in the current directory
// and fails the test if loading fails.
func LoadModule(t *testing.T, L *lua.LState, moduleName string) lua.LValue {
	t.Helper()

	modulePath := filepath.Join(".", moduleName+".lua")
	if err := L.DoFile(modulePath); err != nil {
		t.Fatalf("Failed to load %s module: %v", moduleName, err)
	}

	module := L.Get(-1)
	L.SetGlobal(moduleName, module)
	L.Pop(1)

	return module
}

// LoadMultipleModules loads multiple Lua modules in sequence.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to load modules into
//   - moduleNames: Variable number of module names to load
//
// Returns:
//   - map[string]lua.LValue: Map of module names to their loaded values
//
// Each module is loaded using LoadModule and stored in the returned map.
func LoadMultipleModules(t *testing.T, L *lua.LState, moduleNames ...string) map[string]lua.LValue {
	t.Helper()

	modules := make(map[string]lua.LValue)
	for _, name := range moduleNames {
		modules[name] = LoadModule(t, L, name)
	}

	return modules
}

// RequireModule loads a module using Lua's require mechanism.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to load the module into
//   - moduleName: Name of the module to require
//
// The module is loaded using Lua's require() function and set as a global
// variable with the same name. Fails the test if require fails.
func RequireModule(t *testing.T, L *lua.LState, moduleName string) {
	t.Helper()

	script := fmt.Sprintf(`
		local %s = require("%s")
		_G.%s = %s
	`, moduleName, moduleName, moduleName, moduleName)

	if err := L.DoString(script); err != nil {
		t.Fatalf("Failed to require %s module: %v", moduleName, err)
	}
}

// LoadModuleWithBridges loads a module along with mock bridges.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to load into
//   - moduleName: Name of the module to load
//   - bridges: Map of bridge names to mock bridge instances
//
// Returns:
//   - lua.LValue: The loaded module value
//
// Sets up a global 'bridge' table containing the mock bridges before loading
// the module, allowing the module to access the bridges during initialization.
func LoadModuleWithBridges(t *testing.T, L *lua.LState, moduleName string, bridges map[string]*testutils.MockBridge) lua.LValue {
	t.Helper()

	// Set up bridge global
	bridgeTable := L.NewTable()
	for name, bridge := range bridges {
		bridgeModule := CreateMockBridgeModule(L, bridge)
		L.SetField(bridgeTable, name, bridgeModule)
	}
	L.SetGlobal("bridge", bridgeTable)

	// Load the module
	return LoadModule(t, L, moduleName)
}

// ============================================================================
// Lua Table Comparison Utilities
// ============================================================================

// CompareLuaTables compares two Lua tables for deep equality.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state containing the tables
//   - expected: The expected table value
//   - actual: The actual table value to compare
//
// Returns:
//   - bool: true if tables are equal, false otherwise
//
// Performs deep comparison of table contents, checking that all keys and values
// match. Reports differences via t.Errorf but doesn't fail the test directly.
func CompareLuaTables(t *testing.T, L *lua.LState, expected, actual lua.LValue) bool {
	t.Helper()

	if expected.Type() != actual.Type() {
		t.Errorf("Type mismatch: expected %s, got %s", expected.Type(), actual.Type())
		return false
	}

	if expected.Type() != lua.LTTable {
		return lua.LVAsString(expected) == lua.LVAsString(actual)
	}

	expectedTable := expected.(*lua.LTable)
	actualTable := actual.(*lua.LTable)

	// Compare all keys in expected table
	equal := true
	expectedTable.ForEach(func(key, expectedValue lua.LValue) {
		actualValue := actualTable.RawGet(key)
		if actualValue == lua.LNil {
			t.Errorf("Missing key %s in actual table", lua.LVAsString(key))
			equal = false
			return
		}

		if !CompareLuaValues(t, L, expectedValue, actualValue) {
			t.Errorf("Value mismatch for key %s", lua.LVAsString(key))
			equal = false
		}
	})

	// Check for extra keys in actual table
	actualTable.ForEach(func(key, _ lua.LValue) {
		if expectedTable.RawGet(key) == lua.LNil {
			t.Errorf("Unexpected key %s in actual table", lua.LVAsString(key))
			equal = false
		}
	})

	return equal
}

// CompareLuaValues compares two Lua values for equality.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state containing the values
//   - expected: The expected value
//   - actual: The actual value to compare
//
// Returns:
//   - bool: true if values are equal, false otherwise
//
// Handles all Lua types including nil, bool, number, string, table, and function.
// For tables, performs deep comparison. Functions are compared by reference.
func CompareLuaValues(t *testing.T, L *lua.LState, expected, actual lua.LValue) bool {
	t.Helper()

	if expected.Type() != actual.Type() {
		t.Errorf("Type mismatch: expected %s, got %s", expected.Type(), actual.Type())
		return false
	}

	switch expected.Type() {
	case lua.LTNil:
		return true
	case lua.LTBool:
		return lua.LVAsBool(expected) == lua.LVAsBool(actual)
	case lua.LTNumber:
		return lua.LVAsNumber(expected) == lua.LVAsNumber(actual)
	case lua.LTString:
		return lua.LVAsString(expected) == lua.LVAsString(actual)
	case lua.LTTable:
		return CompareLuaTables(t, L, expected, actual)
	case lua.LTFunction:
		// Functions are compared by reference
		return expected == actual
	default:
		t.Errorf("Cannot compare values of type %s", expected.Type())
		return false
	}
}

// AssertTableHasKey verifies a table has a specific key.
//
// Parameters:
//   - t: The testing context
//   - table: The Lua table to check
//   - key: The key to look for
//
// Fails the test with an error if the key is not present in the table.
func AssertTableHasKey(t *testing.T, table *lua.LTable, key string) {
	t.Helper()

	if table.RawGetString(key) == lua.LNil {
		t.Errorf("Table missing expected key: %s", key)
	}
}

// AssertTableHasKeys verifies a table has all specified keys.
//
// Parameters:
//   - t: The testing context
//   - table: The Lua table to check
//   - keys: Variable number of keys to check for
//
// Calls AssertTableHasKey for each key, failing if any are missing.
func AssertTableHasKeys(t *testing.T, table *lua.LTable, keys ...string) {
	t.Helper()

	for _, key := range keys {
		AssertTableHasKey(t, table, key)
	}
}

// GetTableKeys returns all keys from a Lua table as strings.
//
// Parameters:
//   - table: The Lua table to extract keys from
//
// Returns:
//   - []string: Slice containing all keys converted to strings
//
// The order of keys in the returned slice is not guaranteed.
func GetTableKeys(table *lua.LTable) []string {
	keys := []string{}
	table.ForEach(func(key, _ lua.LValue) {
		keys = append(keys, lua.LVAsString(key))
	})
	return keys
}

// ============================================================================
// Async Test Utilities
// ============================================================================

// WaitForCondition waits for a condition to become true within a timeout.
//
// Parameters:
//   - t: The testing context
//   - timeout: Maximum time to wait
//   - check: Function that returns true when condition is met
//   - message: Description of what we're waiting for
//
// Polls the check function every 10ms. Fails the test with a timeout error
// if the condition doesn't become true within the specified duration.
func WaitForCondition(t *testing.T, timeout time.Duration, check func() bool, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if check() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("Timeout waiting for condition: %s", message)
		}
	}
}

// RunAsyncTest runs a Lua script with async operations and timeout handling.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to run the script in
//   - script: The Lua script to execute
//   - timeout: Maximum time allowed for execution
//
// Executes the script in a goroutine and waits for completion or timeout.
// Fails the test if the script errors or doesn't complete within timeout.
func RunAsyncTest(t *testing.T, L *lua.LState, script string, timeout time.Duration) {
	t.Helper()

	done := make(chan error, 1)

	go func() {
		err := L.DoString(script)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Async test failed: %v", err)
		}
	case <-time.After(timeout):
		t.Fatalf("Async test timed out after %v", timeout)
	}
}

// WaitForLuaValue waits for a Lua global variable to have a specific value.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state containing the global
//   - globalName: Name of the global variable to monitor
//   - expectedValue: The expected value
//   - timeout: Maximum time to wait
//
// Uses WaitForCondition to poll the global variable until it matches the
// expected value or the timeout expires.
func WaitForLuaValue(t *testing.T, L *lua.LState, globalName string, expectedValue lua.LValue, timeout time.Duration) {
	t.Helper()

	WaitForCondition(t, timeout, func() bool {
		actual := L.GetGlobal(globalName)
		return CompareLuaValues(t, L, expectedValue, actual)
	}, fmt.Sprintf("waiting for %s to equal expected value", globalName))
}

// ============================================================================
// Error Assertion Helpers
// ============================================================================

// AssertLuaError verifies that a Lua script produces an expected error.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to run the script in
//   - script: The Lua script expected to error
//   - expectedError: Substring that should appear in the error message
//
// Fails the test if the script doesn't error or if the error doesn't contain
// the expected substring.
func AssertLuaError(t *testing.T, L *lua.LState, script string, expectedError string) {
	t.Helper()

	err := L.DoString(script)
	if err == nil {
		t.Errorf("Expected error containing '%s', but got no error", expectedError)
		return
	}

	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', but got: %v", expectedError, err)
	}
}

// AssertNoLuaError verifies that a Lua script runs without error.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to run the script in
//   - script: The Lua script to execute
//
// Fails the test if the script produces any error.
func AssertNoLuaError(t *testing.T, L *lua.LState, script string) {
	t.Helper()

	if err := L.DoString(script); err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

// AssertLuaErrorMatch verifies that a Lua script error matches a pattern.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to run the script in
//   - script: The Lua script expected to error
//   - pattern: Pattern that should match the error message (currently substring match)
//
// Similar to AssertLuaError but intended for pattern matching (currently
// implements simple substring matching).
func AssertLuaErrorMatch(t *testing.T, L *lua.LState, script string, pattern string) {
	t.Helper()

	err := L.DoString(script)
	if err == nil {
		t.Errorf("Expected error matching pattern '%s', but got no error", pattern)
		return
	}

	// Simple pattern matching (could be enhanced with regex)
	if !strings.Contains(err.Error(), pattern) {
		t.Errorf("Error '%v' does not match pattern '%s'", err, pattern)
	}
}

// CaptureError runs a Lua script and returns the error if any.
//
// Parameters:
//   - L: The Lua state to run the script in
//   - script: The Lua script to execute
//
// Returns:
//   - error: The error from script execution, or nil if successful
//
// Unlike assertion functions, this simply returns the error without failing tests.
func CaptureError(L *lua.LState, script string) error {
	return L.DoString(script)
}

// ============================================================================
// Mock Bridge Creation Utilities
// ============================================================================

// CreateMockBridge creates a mock bridge with standard configuration.
//
// Parameters:
//   - id: Identifier for the mock bridge
//
// Returns:
//   - *testutils.MockBridge: A mock bridge with an "execute" method pre-configured
//
// The bridge includes a standard "execute" method that takes an operation string
// and returns "executed: <operation>".
func CreateMockBridge(id string) *testutils.MockBridge {
	bridge := testutils.NewMockBridge(id)

	// Add common methods
	bridge.WithMethod("execute", engine.MethodInfo{
		Name:        "execute",
		Description: "Execute operation",
		Parameters: []engine.ParameterInfo{
			{Name: "operation", Type: "string", Required: true},
		},
		ReturnType: "any",
	}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("missing operation argument")
		}
		return engine.NewStringValue(fmt.Sprintf("executed: %s", args[0])), nil
	})

	return bridge
}

// CreateMockBridgeModule creates a Lua table that acts as a bridge module.
//
// Parameters:
//   - L: The Lua state to create the module in
//   - bridge: The mock bridge to wrap
//
// Returns:
//   - *lua.LTable: A Lua table with methods corresponding to the bridge methods
//
// The created module handles both dot and colon syntax for method calls,
// automatically converting between Lua values and script values.
func CreateMockBridgeModule(L *lua.LState, bridge *testutils.MockBridge) *lua.LTable {
	module := L.NewTable()

	// Add methods from the bridge
	for _, method := range bridge.Methods() {
		// Create a closure to capture the method name
		methodName := method.Name
		fn := L.NewFunction(func(L *lua.LState) int {
			// Collect arguments
			args := []engine.ScriptValue{}
			nargs := L.GetTop()

			// Handle colon syntax (first arg is self)
			startIdx := 1
			if nargs > 0 && L.Get(1).Type() == lua.LTTable {
				// Check if first arg is the module itself
				if L.Get(1) == module {
					startIdx = 2 // Skip self
				}
			}

			for i := startIdx; i <= nargs; i++ {
				args = append(args, LuaValueToScriptValue(L.Get(i)))
			}

			// Execute method
			ctx := context.Background()
			result, err := bridge.ExecuteMethod(ctx, methodName, args)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			// Convert result
			L.Push(ScriptValueToLuaValue(L, result))
			return 1
		})
		L.SetField(module, methodName, fn)
	}

	return module
}

// CreateMockBridgeWithHandlers creates a mock bridge with custom method handlers.
//
// Parameters:
//   - id: Identifier for the mock bridge
//   - handlers: Map of method names to their handler functions
//
// Returns:
//   - *testutils.MockBridge: A mock bridge configured with the provided handlers
//
// Each handler is registered as a method with a generic description and "any" return type.
func CreateMockBridgeWithHandlers(id string, handlers map[string]testutils.MethodHandler) *testutils.MockBridge {
	bridge := testutils.NewMockBridge(id)

	for method, handler := range handlers {
		bridge.WithMethod(method, engine.MethodInfo{
			Name:        method,
			Description: fmt.Sprintf("Mock %s method", method),
			ReturnType:  "any",
		}, handler)
	}

	return bridge
}

// ============================================================================
// Test Fixture Management
// ============================================================================

// TestFixture represents a test environment with Lua state, loaded modules,
// mock bridges, and cleanup functions. It provides a convenient way to manage
// test resources and ensure proper cleanup.
type TestFixture struct {
	T       *testing.T
	L       *lua.LState
	Modules map[string]lua.LValue
	Bridges map[string]*testutils.MockBridge
	Cleanup []func()
	mu      sync.Mutex
}

// NewTestFixture creates a new test fixture with initialized Lua state.
//
// Parameters:
//   - t: The testing context
//
// Returns:
//   - *TestFixture: A new fixture ready for use
//
// The fixture should be closed using Close() to ensure proper cleanup.
func NewTestFixture(t *testing.T) *TestFixture {
	return &TestFixture{
		T:       t,
		L:       lua.NewState(),
		Modules: make(map[string]lua.LValue),
		Bridges: make(map[string]*testutils.MockBridge),
		Cleanup: []func(){},
	}
}

// LoadModule loads a Lua module into the fixture.
//
// Parameters:
//   - name: Name of the module to load
//
// Returns:
//   - lua.LValue: The loaded module value
//
// The module is stored in the fixture's Modules map for later access.
func (f *TestFixture) LoadModule(name string) lua.LValue {
	f.mu.Lock()
	defer f.mu.Unlock()

	module := LoadModule(f.T, f.L, name)
	f.Modules[name] = module
	return module
}

// AddBridge adds a mock bridge to the fixture.
//
// Parameters:
//   - name: Name to register the bridge under
//   - bridge: The mock bridge instance
//
// The bridge is made available in Lua via the global 'bridge' table.
func (f *TestFixture) AddBridge(name string, bridge *testutils.MockBridge) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.Bridges[name] = bridge

	// Update bridge global
	bridgeTable := f.L.GetGlobal("bridge")
	if bridgeTable.Type() != lua.LTTable {
		bridgeTable = f.L.NewTable()
		f.L.SetGlobal("bridge", bridgeTable)
	}

	bridgeModule := CreateMockBridgeModule(f.L, bridge)
	f.L.SetField(bridgeTable.(*lua.LTable), name, bridgeModule)
}

// AddCleanup adds a cleanup function to be called when the fixture is closed.
//
// Parameters:
//   - fn: Function to call during cleanup
//
// Cleanup functions are called in reverse order (LIFO) when Close() is called.
func (f *TestFixture) AddCleanup(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.Cleanup = append(f.Cleanup, fn)
}

// RunScript executes a Lua script in the fixture.
//
// Parameters:
//   - script: The Lua script to execute
//
// Returns:
//   - error: Any error from script execution
func (f *TestFixture) RunScript(script string) error {
	return f.L.DoString(script)
}

// MustRunScript executes a Lua script and fails the test on error.
//
// Parameters:
//   - script: The Lua script to execute
//
// This is a convenience method that calls RunScript and uses t.Fatalf on error.
func (f *TestFixture) MustRunScript(script string) {
	if err := f.RunScript(script); err != nil {
		f.T.Fatalf("Script execution failed: %v", err)
	}
}

// GetGlobal gets a global value from the Lua state.
//
// Parameters:
//   - name: Name of the global variable
//
// Returns:
//   - lua.LValue: The value of the global variable
func (f *TestFixture) GetGlobal(name string) lua.LValue {
	return f.L.GetGlobal(name)
}

// SetGlobal sets a global value in the Lua state.
//
// Parameters:
//   - name: Name of the global variable
//   - value: Value to set
func (f *TestFixture) SetGlobal(name string, value lua.LValue) {
	f.L.SetGlobal(name, value)
}

// Close cleans up the fixture by running cleanup functions and closing the Lua state.
//
// Cleanup functions are run in reverse order of registration (LIFO).
// This method should be called using defer after creating the fixture.
func (f *TestFixture) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Run cleanup functions in reverse order
	for i := len(f.Cleanup) - 1; i >= 0; i-- {
		f.Cleanup[i]()
	}

	// Close Lua state
	f.L.Close()
}

// ============================================================================
// Conversion Utilities
// ============================================================================

// LuaValueToScriptValue converts a Lua value to an engine.ScriptValue.
//
// Parameters:
//   - lv: The Lua value to convert
//
// Returns:
//   - engine.ScriptValue: The converted value
//
// Handles nil, bool, number, string, and table types. Tables are converted
// to object values with type metadata.
func LuaValueToScriptValue(lv lua.LValue) engine.ScriptValue {
	switch lv.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return engine.NewBoolValue(lua.LVAsBool(lv))
	case lua.LTNumber:
		return engine.NewNumberValue(float64(lua.LVAsNumber(lv)))
	case lua.LTString:
		return engine.NewStringValue(lua.LVAsString(lv))
	case lua.LTTable:
		// Simple conversion - could be enhanced
		// Convert table to object value
		fields := make(map[string]engine.ScriptValue)
		fields["_type"] = engine.NewStringValue("table")
		fields["_ref"] = engine.NewStringValue(fmt.Sprintf("%p", lv))
		return engine.NewObjectValue(fields)
	default:
		return engine.NewStringValue(fmt.Sprintf("<%s>", lv.Type()))
	}
}

// ScriptValueToLuaValue converts an engine.ScriptValue to a Lua value.
//
// Parameters:
//   - L: The Lua state (needed for creating Lua values)
//   - sv: The script value to convert
//
// Returns:
//   - lua.LValue: The converted Lua value
//
// Uses reflection to handle various types, falling back to string representation
// for complex types.
func ScriptValueToLuaValue(L *lua.LState, sv engine.ScriptValue) lua.LValue {
	if sv == nil {
		return lua.LNil
	}

	// Try to use the String() method if available
	if stringer, ok := sv.(fmt.Stringer); ok {
		return lua.LString(stringer.String())
	}

	// Use reflection to handle different types
	v := reflect.ValueOf(sv)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		return lua.LBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(v.Uint())
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(v.Float())
	case reflect.String:
		return lua.LString(v.String())
	default:
		// For complex types, return a string representation
		return lua.LString(fmt.Sprintf("%v", sv))
	}
}

// ============================================================================
// Test Helpers for Common Patterns
// ============================================================================

// TestModuleStructure verifies a module has the expected function exports.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state containing the module
//   - moduleName: Name of the module global variable
//   - expectedFunctions: List of function names that should exist in the module
//
// Fails the test if the module is not a table or if any expected function is missing.
func TestModuleStructure(t *testing.T, L *lua.LState, moduleName string, expectedFunctions []string) {
	t.Helper()

	module := L.GetGlobal(moduleName)
	if module.Type() != lua.LTTable {
		t.Fatalf("Module %s should be a table, got %s", moduleName, module.Type())
	}

	moduleTable := module.(*lua.LTable)
	for _, funcName := range expectedFunctions {
		fn := moduleTable.RawGetString(funcName)
		if fn.Type() != lua.LTFunction {
			t.Errorf("Module %s missing function %s or it's not a function (got %s)",
				moduleName, funcName, fn.Type())
		}
	}
}

// RunTableDrivenTests runs a set of table-driven tests using a test fixture.
//
// Parameters:
//   - t: The testing context
//   - fixture: The test fixture to run tests in
//   - tests: Array of test cases with Name, Script, Check function, and Error fields
//
// Each test is run as a subtest. If Error is specified, the test expects the script
// to fail with an error containing that string. Otherwise, Check is called to verify results.
func RunTableDrivenTests(t *testing.T, fixture *TestFixture, tests []struct {
	Name   string
	Script string
	Check  func(t *testing.T, fixture *TestFixture)
	Error  string
}) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := fixture.RunScript(tt.Script)

			if tt.Error != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got none", tt.Error)
				} else if !strings.Contains(err.Error(), tt.Error) {
					t.Errorf("Expected error containing '%s', got: %v", tt.Error, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if tt.Check != nil {
					tt.Check(t, fixture)
				}
			}
		})
	}
}

// AssertLuaStackClean verifies the Lua stack is clean with no leaked values.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to check
//
// Fails the test if any values remain on the stack, logging details about
// each leaked value to help with debugging.
func AssertLuaStackClean(t *testing.T, L *lua.LState) {
	t.Helper()

	if L.GetTop() != 0 {
		t.Errorf("Lua stack not clean: %d values remaining", L.GetTop())
		for i := 1; i <= L.GetTop(); i++ {
			t.Logf("  Stack[%d]: %v (%s)", i, L.Get(i), L.Get(i).Type())
		}
	}
}
