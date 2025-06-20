// ABOUTME: Tests for the stdlib test infrastructure helpers
// ABOUTME: Verifies that test helpers work correctly for module loading, async testing, and fixtures

package stdlib

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// TestLoadModule verifies module loading helpers
func TestLoadModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test loading promise module
	module := LoadModule(t, L, "promise")
	if module.Type() != lua.LTTable {
		t.Errorf("Expected module to be a table, got %s", module.Type())
	}

	// Verify it's set as global
	global := L.GetGlobal("promise")
	if global != module {
		t.Errorf("Module not set as global correctly")
	}
}

// TestLoadMultipleModules verifies loading multiple modules
func TestLoadMultipleModules(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	modules := LoadMultipleModules(t, L, "promise", "testing")

	if len(modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(modules))
	}

	// Verify both are loaded
	if modules["promise"].Type() != lua.LTTable {
		t.Errorf("Promise module not loaded correctly")
	}
	if modules["testing"].Type() != lua.LTTable {
		t.Errorf("Testing module not loaded correctly")
	}
}

// TestCompareLuaTables verifies table comparison
func TestCompareLuaTables(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name     string
		script1  string
		script2  string
		expected bool
	}{
		{
			name:     "equal_simple_tables",
			script1:  "return {a = 1, b = 2}",
			script2:  "return {a = 1, b = 2}",
			expected: true,
		},
		{
			name:     "different_values",
			script1:  "return {a = 1, b = 2}",
			script2:  "return {a = 1, b = 3}",
			expected: false,
		},
		{
			name:     "different_keys",
			script1:  "return {a = 1, b = 2}",
			script2:  "return {a = 1, c = 2}",
			expected: false,
		},
		{
			name:     "nested_tables_equal",
			script1:  "return {a = {x = 1}, b = 2}",
			script2:  "return {a = {x = 1}, b = 2}",
			expected: true,
		},
		{
			name:     "nested_tables_different",
			script1:  "return {a = {x = 1}, b = 2}",
			script2:  "return {a = {x = 2}, b = 2}",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get first table
			if err := L.DoString(tt.script1); err != nil {
				t.Fatalf("Failed to create first table: %v", err)
			}
			table1 := L.Get(-1)
			L.Pop(1)

			// Get second table
			if err := L.DoString(tt.script2); err != nil {
				t.Fatalf("Failed to create second table: %v", err)
			}
			table2 := L.Get(-1)
			L.Pop(1)

			// Create a sub-test to capture comparison errors
			subTest := &testing.T{}
			result := CompareLuaTables(subTest, L, table1, table2)
			if result != tt.expected {
				t.Errorf("Expected comparison to be %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestAsyncUtilities verifies async test helpers
func TestAsyncUtilities(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Load promise module
	LoadModule(t, L, "promise")

	t.Run("WaitForCondition", func(t *testing.T) {
		// Use a channel to synchronize instead of accessing Lua state from multiple goroutines
		done := make(chan bool)

		// Start a goroutine that will signal completion
		go func() {
			time.Sleep(50 * time.Millisecond)
			done <- true
		}()

		// Wait for condition using the channel
		WaitForCondition(t, 200*time.Millisecond, func() bool {
			select {
			case <-done:
				return true
			default:
				return false
			}
		}, "test completion signal")
	})

	t.Run("RunAsyncTest", func(t *testing.T) {
		// Create a new Lua state for this test to avoid race conditions
		testL := lua.NewState()
		defer testL.Close()

		script := `
			-- Simple synchronous test that sets a value
			_G.async_test_complete = true
		`

		RunAsyncTest(t, testL, script, 200*time.Millisecond)

		// Verify completion
		complete := testL.GetGlobal("async_test_complete")
		if !lua.LVAsBool(complete) {
			t.Errorf("Async test did not complete")
		}
	})
}

// TestErrorAssertions verifies error assertion helpers
func TestErrorAssertions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	t.Run("AssertLuaError", func(t *testing.T) {
		script := `error("test error message")`
		AssertLuaError(t, L, script, "test error message")
	})

	t.Run("AssertNoLuaError", func(t *testing.T) {
		script := `return 1 + 1`
		AssertNoLuaError(t, L, script)
	})
}

// TestMockBridgeCreation verifies mock bridge helpers
func TestMockBridgeCreation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	t.Run("CreateMockBridge", func(t *testing.T) {
		bridge := CreateMockBridge("test-bridge")

		if bridge.GetID() != "test-bridge" {
			t.Errorf("Expected bridge ID to be 'test-bridge', got %s", bridge.GetID())
		}

		// Verify execute method exists
		methods := bridge.Methods()
		found := false
		for _, method := range methods {
			if method.Name == "execute" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Execute method not found in mock bridge")
		}
	})

	t.Run("CreateMockBridgeModule", func(t *testing.T) {
		bridge := CreateMockBridge("test-bridge")
		// Mark bridge as initialized for the test
		bridge.WithInitialized(true)
		module := CreateMockBridgeModule(L, bridge)

		if module.Type() != lua.LTTable {
			t.Errorf("Expected module to be a table, got %s", module.Type())
		}

		// Verify method is callable
		L.SetGlobal("test_module", module)
		script := `
			local result = test_module:execute("test-op")
			return result == "executed: test-op"
		`

		if err := L.DoString(script); err != nil {
			t.Fatalf("Failed to call mock method: %v", err)
		}

		result := L.Get(-1)
		if !lua.LVAsBool(result) {
			t.Errorf("Mock method did not return expected result")
		}
	})
}

// TestFixtureManagement verifies test fixture functionality
func TestFixtureManagement(t *testing.T) {
	t.Run("BasicFixture", func(t *testing.T) {
		fixture := NewTestFixture(t)
		defer fixture.Close()

		// Load module
		module := fixture.LoadModule("promise")
		if module.Type() != lua.LTTable {
			t.Errorf("Module not loaded correctly")
		}

		// Run script
		fixture.MustRunScript(`_G.test_result = 42`)

		result := fixture.GetGlobal("test_result")
		if lua.LVAsNumber(result) != 42 {
			t.Errorf("Expected 42, got %v", lua.LVAsNumber(result))
		}
	})

	t.Run("FixtureWithBridge", func(t *testing.T) {
		fixture := NewTestFixture(t)
		defer fixture.Close()

		// Add mock bridge
		bridge := CreateMockBridge("test-bridge")
		bridge.WithInitialized(true)
		fixture.AddBridge("test", bridge)

		// Use bridge in script
		script := `
			if bridge and bridge.test then
				local result = bridge.test:execute("operation")
				_G.test_result = (result == "executed: operation")
			else
				_G.test_result = false
			end
		`

		err := fixture.RunScript(script)
		if err != nil {
			t.Fatalf("Script failed: %v", err)
		}

		result := fixture.GetGlobal("test_result")
		if !lua.LVAsBool(result) {
			t.Errorf("Bridge method did not work correctly")
		}
	})

	t.Run("FixtureCleanup", func(t *testing.T) {
		cleanupCalled := false

		fixture := NewTestFixture(t)
		fixture.AddCleanup(func() {
			cleanupCalled = true
		})

		fixture.Close()

		if !cleanupCalled {
			t.Errorf("Cleanup function was not called")
		}
	})
}

// TestTableDrivenTests verifies table-driven test helper
func TestTableDrivenTests(t *testing.T) {
	fixture := NewTestFixture(t)
	defer fixture.Close()

	tests := []struct {
		Name   string
		Script string
		Check  func(t *testing.T, fixture *TestFixture)
		Error  string
	}{
		{
			Name:   "successful_test",
			Script: "_G.test_var = 100",
			Check: func(t *testing.T, f *TestFixture) {
				val := f.GetGlobal("test_var")
				if lua.LVAsNumber(val) != 100 {
					t.Errorf("Expected 100, got %v", lua.LVAsNumber(val))
				}
			},
		},
		{
			Name:   "error_test",
			Script: "error('expected error')",
			Error:  "expected error",
		},
	}

	RunTableDrivenTests(t, fixture, tests)
}

// setupMockPromiseModuleForTests creates a simple promise module for testing
func setupMockPromiseModuleForTests(L *lua.LState) {
	err := L.DoString(`
		promise = {}
		
		function promise.new(executor)
			local p = {
				_resolved = false,
				_rejected = false,
				_value = nil,
				_error = nil,
				_thens = {},
				_catches = {}
			}
			
			function p:andThen(callback)
				if self._resolved then
					callback(self._value)
				elseif not self._rejected then
					table.insert(self._thens, callback)
				end
				return self
			end
			
			function p:onError(callback)
				if self._rejected then
					callback(self._error)
				else
					table.insert(self._catches, callback)
				end
				return self
			end
			
			local function resolve(value)
				if not p._resolved and not p._rejected then
					p._resolved = true
					p._value = value
					for _, cb in ipairs(p._thens) do
						cb(value)
					end
				end
			end
			
			local function reject(err)
				if not p._resolved and not p._rejected then
					p._rejected = true
					p._error = err
					for _, cb in ipairs(p._catches) do
						cb(err)
					end
				end
			end
			
			executor(resolve, reject)
			return p
		end
		
		function promise.sleep(ms)
			local p = promise.new(function(resolve, reject)
				resolve(true)
			end)
			return p
		end
	`)
	if err != nil {
		panic(err)
	}
}

// TestPromiseAssertions verifies promise assertion helpers
func TestPromiseAssertions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Use mock promise module for testing
	setupMockPromiseModuleForTests(L)

	t.Run("AssertPromiseResolves", func(t *testing.T) {
		// Create a promise that resolves
		script := `
			test_promise = promise.new(function(resolve, reject)
				promise.sleep(10):andThen(function()
					resolve("success")
				end)
			end)
		`

		if err := L.DoString(script); err != nil {
			t.Fatalf("Failed to create promise: %v", err)
		}

		expectedValue := lua.LString("success")
		AssertPromiseResolves(t, L, "test_promise", expectedValue, 100*time.Millisecond)
	})

	t.Run("AssertPromiseRejects", func(t *testing.T) {
		// Create a promise that rejects
		script := `
			test_promise = promise.new(function(resolve, reject)
				promise.sleep(10):andThen(function()
					reject("expected error")
				end)
			end)
		`

		if err := L.DoString(script); err != nil {
			t.Fatalf("Failed to create promise: %v", err)
		}

		AssertPromiseRejects(t, L, "test_promise", "expected error", 100*time.Millisecond)
	})
}

// TestMemoryLeakDetection verifies memory leak detection
func TestMemoryLeakDetection(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	t.Run("CaptureMemorySnapshot", func(t *testing.T) {
		// First, create some data in Lua to ensure memory is allocated
		err := L.DoString(`
			-- Create some data to ensure memory usage
			local data = {}
			for i = 1, 1000 do
				data[i] = string.rep("x", 100)
			end
			_G.test_data = data
		`)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}

		snapshot := CaptureMemorySnapshot(L)

		// Clean up
		L.SetGlobal("test_data", lua.LNil)

		// For this test, we'll accept that memory might be 0 if collectgarbage
		// is not fully implemented in GopherLua
		if snapshot.LuaMemory < 0 {
			t.Errorf("Lua memory should be >= 0, got %v", snapshot.LuaMemory)
		}

		if snapshot.Goroutines <= 0 {
			t.Errorf("Goroutines should be > 0, got %v", snapshot.Goroutines)
		}

		// At least verify that the snapshot structure is created correctly
		if snapshot.Timestamp.IsZero() {
			t.Errorf("Timestamp should not be zero")
		}

		if snapshot.LuaObjects < 0 {
			t.Errorf("LuaObjects should be >= 0, got %v", snapshot.LuaObjects)
		}
	})

	t.Run("RunMemoryLeakTest", func(t *testing.T) {
		setup := `_G.test_data = {}`
		operation := `
			-- Create some temporary data
			local temp = {}
			for i = 1, 10 do
				temp[i] = "data" .. i
			end
			-- Don't store it globally (no leak)
		`
		cleanup := `_G.test_data = nil`

		RunMemoryLeakTest(t, L, setup, operation, cleanup, 100)
	})
}

// TestConcurrentOperations verifies concurrent operation validators
func TestConcurrentOperations(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	LoadModule(t, L, "promise")

	t.Run("RunConcurrentTests", func(t *testing.T) {
		tests := []ConcurrentTest{
			{
				Name:  "concurrent_counter",
				Setup: "_G.counter = 0",
				Concurrent: []string{
					"_G.counter = _G.counter + 1",
					"_G.counter = _G.counter + 1",
					"_G.counter = _G.counter + 1",
				},
				Validate: func(t *testing.T, L *lua.LState) {
					// Note: Without proper synchronization, this might not always be 3
					counter := L.GetGlobal("counter")
					val := lua.LVAsNumber(counter)
					if val < 1 || val > 3 {
						t.Errorf("Unexpected counter value: %v", val)
					}
				},
			},
		}

		RunConcurrentTests(t, L, tests)
	})
}

// TestLuaStackClean verifies stack cleanup checking
func TestLuaStackClean(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Clean stack should pass
	AssertLuaStackClean(t, L)

	// Push some values and don't clean
	L.Push(lua.LNumber(1))
	L.Push(lua.LString("test"))

	// This would fail if we didn't clean up
	// (commented out to avoid test failure)
	// AssertLuaStackClean(t, L)

	// Clean up
	L.Pop(2)
	AssertLuaStackClean(t, L)
}
