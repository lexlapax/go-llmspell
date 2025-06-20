// ABOUTME: Comprehensive tests for Event & Hooks Library in Lua standard library
// ABOUTME: Tests event emission, subscription, hooks, filtering, and advanced features

package stdlib

import (
	"path/filepath"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// setupEventsLibrary loads the events library and dependencies
func setupEventsLibrary(t testing.TB, L *lua.LState) {
	t.Helper()

	// Load promise library first (dependency)
	promisePath := filepath.Join(".", "promise.lua")
	err := L.DoFile(promisePath)
	if err != nil {
		t.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)
	L.Pop(1)

	// Load events library
	eventsPath := filepath.Join(".", "events.lua")
	err = L.DoFile(eventsPath)
	if err != nil {
		t.Fatalf("Failed to load events library: %v", err)
	}
}

// TestEventsLibraryLoading tests that the events library can be loaded
func TestEventsLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(t, L)

	script := `
		local events = require("events")
		
		if type(events) ~= "table" then
			error("Events module should be a table")
		end
		
		if type(events.emit) ~= "function" then
			error("events.emit should be a function")
		end
		
		if type(events.on) ~= "function" then
			error("events.on should be a function")
		end
		
		if type(events.EventEmitter) ~= "table" then
			error("EventEmitter class should be available")
		end
		
		if type(events.hooks) ~= "table" then
			error("Hooks module should be available")
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Events library structure test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected library structure test to pass")
	}
}

// TestEventEmitter tests EventEmitter functionality
func TestEventEmitter(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, L *lua.LState)
	}{
		{
			name: "basic_emit_and_on",
			script: `
				local events = require("events")
				local emitter = events.create_emitter()
				
				local called = false
				local received_data = nil
				
				emitter:on("test", function(data)
					called = true
					received_data = data
				end)
				
				emitter:emit("test", "hello")
				
				return {called = called, data = received_data}
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1).(*lua.LTable)
				if !lua.LVAsBool(result.RawGetString("called")) {
					t.Error("Expected handler to be called")
				}
				if result.RawGetString("data").String() != "hello" {
					t.Error("Expected data to be 'hello'")
				}
			},
		},
		{
			name: "multiple_handlers",
			script: `
				local events = require("events")
				local emitter = events.create_emitter()
				
				local count = 0
				
				emitter:on("test", function() count = count + 1 end)
				emitter:on("test", function() count = count + 10 end)
				emitter:on("test", function() count = count + 100 end)
				
				emitter:emit("test")
				
				return count
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1)
				if lua.LVAsNumber(result) != 111 {
					t.Errorf("Expected count to be 111, got %v", result)
				}
			},
		},
		{
			name: "once_handler",
			script: `
				local events = require("events")
				local emitter = events.create_emitter()
				
				local count = 0
				
				emitter:once("test", function() count = count + 1 end)
				
				emitter:emit("test")
				emitter:emit("test")
				emitter:emit("test")
				
				return count
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1)
				if lua.LVAsNumber(result) != 1 {
					t.Errorf("Expected once handler to be called only once, got %v calls", result)
				}
			},
		},
		{
			name: "off_removes_handler",
			script: `
				local events = require("events")
				local emitter = events.create_emitter()
				
				local count = 0
				local handler = function() count = count + 1 end
				
				emitter:on("test", handler)
				emitter:emit("test") -- count = 1
				
				emitter:off("test", handler)
				emitter:emit("test") -- count should still be 1
				
				return count
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1)
				if lua.LVAsNumber(result) != 1 {
					t.Errorf("Expected handler to be removed, got %v calls", result)
				}
			},
		},
		{
			name: "listener_count",
			script: `
				local events = require("events")
				local emitter = events.create_emitter()
				
				local initial = emitter:listenerCount("test")
				
				emitter:on("test", function() end)
				emitter:on("test", function() end)
				local after_add = emitter:listenerCount("test")
				
				emitter:removeAllListeners("test")
				local after_remove = emitter:listenerCount("test")
				
				return {
					initial = initial,
					after_add = after_add,
					after_remove = after_remove
				}
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1).(*lua.LTable)

				initial := lua.LVAsNumber(result.RawGetString("initial"))
				afterAdd := lua.LVAsNumber(result.RawGetString("after_add"))
				afterRemove := lua.LVAsNumber(result.RawGetString("after_remove"))

				if initial != 0 {
					t.Errorf("Expected initial count to be 0, got %v", initial)
				}
				if afterAdd != 2 {
					t.Errorf("Expected count after add to be 2, got %v", afterAdd)
				}
				if afterRemove != 0 {
					t.Errorf("Expected count after remove to be 0, got %v", afterRemove)
				}
			},
		},
		{
			name: "error_event",
			script: `
				local events = require("events")
				local emitter = events.create_emitter()
				
				local state = {
					error_caught = false,
					error_msg = nil
				}
				
				-- Set up error handler first
				emitter:on("error", function(err, event, handler)
					state.error_caught = true
					state.error_msg = tostring(err)
				end)
				
				-- Add test handler that will error
				emitter:on("test", function()
					error("Handler error")
				end)
				
				-- Check handlers are set up correctly
				local error_count = emitter:listenerCount("error")
				local test_count = emitter:listenerCount("test")
				
				-- Emit the test event
				local has_listeners = emitter:emit("test")
				
				
				return {
					caught = state.error_caught, 
					msg = state.error_msg, 
					has_listeners = has_listeners,
					error_count = error_count,
					test_count = test_count
				}
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1).(*lua.LTable)

				if !lua.LVAsBool(result.RawGetString("caught")) {
					t.Error("Expected error event to be emitted")
				}

				msg := result.RawGetString("msg")
				if msg == lua.LNil {
					t.Error("Expected error message to be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupEventsLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}

			tt.check(t, L)
		})
	}
}

// TestGlobalEvents tests global event system
func TestGlobalEvents(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(t, L)

	script := `
		local events = require("events")
		
		local results = {}
		
		-- Test global event emission
		events.on("global.test", function(data)
			results.global_received = data
		end)
		
		events.emit("global.test", "global data")
		
		-- Test once on global
		local once_count = 0
		events.once("global.once", function()
			once_count = once_count + 1
		end)
		
		events.emit("global.once")
		events.emit("global.once")
		
		results.once_count = once_count
		
		-- Test off on global
		local handler = function() results.should_not_run = true end
		events.on("global.off", handler)
		events.off("global.off", handler)
		events.emit("global.off")
		
		return results
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := L.Get(-1).(*lua.LTable)

	globalReceived := result.RawGetString("global_received")
	if globalReceived.String() != "global data" {
		t.Error("Expected global event to be received")
	}

	onceCount := lua.LVAsNumber(result.RawGetString("once_count"))
	if onceCount != 1 {
		t.Errorf("Expected once to be called once, got %v", onceCount)
	}

	shouldNotRun := result.RawGetString("should_not_run")
	if shouldNotRun != lua.LNil {
		t.Error("Expected handler to be removed")
	}
}

// TestHooksSystem tests hooks functionality
func TestHooksSystem(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, L *lua.LState)
	}{
		{
			name: "before_hook",
			script: `
				local events = require("events")
				local hooks = events.hooks
				
				local call_order = {}
				
				-- Add before hook
				hooks.before("action", function(arg)
					table.insert(call_order, "before: " .. arg)
					return arg .. " modified"
				end)
				
				-- Execute with hooks
				local result = hooks.execute("action", function(arg)
					table.insert(call_order, "main: " .. arg)
					return "result: " .. arg
				end, "test")
				
				return {order = call_order, result = result}
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1).(*lua.LTable)

				order := result.RawGetString("order").(*lua.LTable)
				if order.Len() != 2 {
					t.Errorf("Expected 2 calls, got %d", order.Len())
				}

				// Check order
				first := order.RawGetInt(1).String()
				second := order.RawGetInt(2).String()

				if first != "before: test" {
					t.Errorf("Expected before hook first, got %s", first)
				}
				if second != "main: test modified" {
					t.Errorf("Expected main with modified arg, got %s", second)
				}
			},
		},
		{
			name: "after_hook",
			script: `
				local events = require("events")
				local hooks = events.hooks
				
				local call_order = {}
				
				-- Add after hook
				hooks.after("action", function(result)
					table.insert(call_order, "after: " .. result)
					return result .. " post-processed"
				end)
				
				-- Execute with hooks
				local result = hooks.execute("action", function(arg)
					table.insert(call_order, "main: " .. arg)
					return "result"
				end, "test")
				
				return {order = call_order, result = result}
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1).(*lua.LTable)

				order := result.RawGetString("order").(*lua.LTable)
				if order.Len() != 2 {
					t.Errorf("Expected 2 calls, got %d", order.Len())
				}

				finalResult := result.RawGetString("result").String()
				if finalResult != "result post-processed" {
					t.Errorf("Expected post-processed result, got %s", finalResult)
				}
			},
		},
		{
			name: "around_hook",
			script: `
				local events = require("events")
				local hooks = events.hooks
				
				local call_order = {}
				
				-- Add around hook
				hooks.around("action", function(fn, arg)
					table.insert(call_order, "around before: " .. arg)
					local result = fn(arg .. " wrapped")
					table.insert(call_order, "around after: " .. result)
					return result .. " wrapped"
				end)
				
				-- Execute with hooks
				local result = hooks.execute("action", function(arg)
					table.insert(call_order, "main: " .. arg)
					return "result"
				end, "test")
				
				return {order = call_order, result = result}
			`,
			check: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1).(*lua.LTable)

				order := result.RawGetString("order").(*lua.LTable)
				if order.Len() != 3 {
					t.Errorf("Expected 3 calls, got %d", order.Len())
				}

				finalResult := result.RawGetString("result").String()
				if finalResult != "result wrapped" {
					t.Errorf("Expected wrapped result, got %s", finalResult)
				}
			},
		},
		{
			name: "multiple_hooks_order",
			script: `
				local events = require("events")
				local hooks = events.hooks
				
				local call_order = {}
				
				-- Add multiple hooks
				hooks.before("action", function(arg)
					table.insert(call_order, "before1")
				end)
				
				hooks.before("action", function(arg)
					table.insert(call_order, "before2")
				end)
				
				hooks.after("action", function(result)
					table.insert(call_order, "after1")
				end)
				
				hooks.after("action", function(result)
					table.insert(call_order, "after2")
				end)
				
				hooks.around("action", function(fn, arg)
					table.insert(call_order, "around1_before")
					local result = fn(arg)
					table.insert(call_order, "around1_after")
					return result
				end)
				
				hooks.around("action", function(fn, arg)
					table.insert(call_order, "around2_before")
					local result = fn(arg)
					table.insert(call_order, "around2_after")
					return result
				end)
				
				-- Execute with hooks
				hooks.execute("action", function(arg)
					table.insert(call_order, "main")
					return "result"
				end, "test")
				
				return call_order
			`,
			check: func(t *testing.T, L *lua.LState) {
				order := L.Get(-1).(*lua.LTable)

				// Expected order:
				// 1. before1, before2 (in registration order)
				// 2. around1_before (first registered wraps last)
				// 3. around2_before
				// 4. main
				// 5. around2_after
				// 6. around1_after
				// 7. after1, after2 (in registration order)

				expected := []string{
					"before1", "before2",
					"around1_before", "around2_before",
					"main",
					"around2_after", "around1_after",
					"after1", "after2",
				}

				if order.Len() != len(expected) {
					t.Errorf("Expected %d calls, got %d", len(expected), order.Len())
				}

				for i, exp := range expected {
					actual := order.RawGetInt(i + 1).String()
					if actual != exp {
						t.Errorf("Expected order[%d] to be %s, got %s", i+1, exp, actual)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupEventsLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}

			tt.check(t, L)
		})
	}
}

// TestEventFiltering tests event filtering functionality
func TestEventFiltering(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(t, L)

	script := `
		local events = require("events")
		
		local matched_events = {}
		
		-- Set up filter for user.* events
		local unsubscribe = events.filter("user.*", function(event, data)
			table.insert(matched_events, {event = event, data = data})
		end)
		
		-- Emit various events
		events.emit("user.login", "alice")
		events.emit("user.logout", "bob")
		events.emit("system.startup", "test") -- Should not match
		events.emit("user.profile.update", "charlie")
		
		-- Unsubscribe and emit more
		unsubscribe()
		events.emit("user.deleted", "david") -- Should not be captured
		
		return matched_events
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := L.Get(-1).(*lua.LTable)

	// Should have captured 3 user.* events
	if result.Len() != 3 {
		t.Errorf("Expected 3 matched events, got %d", result.Len())
	}

	// Check first event
	first := result.RawGetInt(1).(*lua.LTable)
	if first.RawGetString("event").String() != "user.login" {
		t.Error("Expected first event to be user.login")
	}
	if first.RawGetString("data").String() != "alice" {
		t.Error("Expected first data to be alice")
	}
}

// TestEventNamespacing tests namespaced events
func TestEventNamespacing(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(t, L)

	script := `
		local events = require("events")
		
		-- Create namespaced emitters
		local auth = events.namespace("auth")
		local api = events.namespace("api")
		
		local results = {}
		
		-- Subscribe to namespaced events
		auth:on("login", function(user)
			results.auth_login = user
		end)
		
		api:on("request", function(endpoint)
			results.api_request = endpoint
		end)
		
		-- Emit on namespaces
		auth:emit("login", "alice")
		api:emit("request", "/users")
		
		-- These should not trigger handlers (different namespace)
		auth:emit("request", "/auth")
		api:emit("login", "bob")
		
		return results
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := L.Get(-1).(*lua.LTable)

	authLogin := result.RawGetString("auth_login")
	if authLogin.String() != "alice" {
		t.Error("Expected auth login event to be received")
	}

	apiRequest := result.RawGetString("api_request")
	if apiRequest.String() != "/users" {
		t.Error("Expected api request event to be received")
	}
}

// TestEventWaitFor tests promise-based event waiting
func TestEventWaitFor(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(t, L)

	script := `
		local events = require("events")
		local promise = require("promise")
		
		results = {}
		
		-- Test successful wait
		local wait1 = events.wait_for("data.ready")
		
		wait1:andThen(function(args)
			results.success = args[1]
		end)
		
		wait1:onError(function(err)
			results.error = err
		end)
		
		-- Emit event immediately  
		events.emit("data.ready", "success")
		
		return results
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := L.Get(-1).(*lua.LTable)

	success := result.RawGetString("success")
	if success.String() != "success" {
		t.Error("Expected successful wait to receive data")
	}

	errorValue := result.RawGetString("error")
	if errorValue != lua.LNil {
		t.Errorf("Unexpected error: %v", errorValue.String())
	}
}

// TestEventAggregation tests event aggregation
func TestEventAggregation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(t, L)

	script := `
		local events = require("events")
		local promise = require("promise")
		
		results = {}
		
		-- Aggregate multiple events
		local agg = events.aggregate({"step1", "step2", "step3"}, 1000)
		
		agg:andThen(function(aggregated_events)
			results.aggregated = aggregated_events
		end)
		
		agg:onError(function(err)
			results.error = err
		end)
		
		-- Emit events synchronously
		events.emit("step2", "data2")
		events.emit("step1", "data1")
		events.emit("step3", "data3")
		
		return results
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	// Give time for async operations
	time.Sleep(150 * time.Millisecond)

	// Check results
	checkScript := `
		return results
	`

	err = L.DoString(checkScript)
	if err != nil {
		t.Fatalf("Check script failed: %v", err)
	}

	result := L.Get(-1).(*lua.LTable)

	aggregated := result.RawGetString("aggregated")
	if aggregated.Type() != lua.LTTable {
		t.Error("Expected aggregated events to be a table")
	}

	aggTable := aggregated.(*lua.LTable)

	// Check each step
	step1 := aggTable.RawGetString("step1")
	if step1.Type() != lua.LTTable {
		t.Error("Expected step1 data to be present")
	}

	step2 := aggTable.RawGetString("step2")
	if step2.Type() != lua.LTTable {
		t.Error("Expected step2 data to be present")
	}

	step3 := aggTable.RawGetString("step3")
	if step3.Type() != lua.LTTable {
		t.Error("Expected step3 data to be present")
	}
}

// TestConcurrentEventHandling tests sequential event handling (Lua is not thread-safe)
func TestConcurrentEventHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(t, L)

	// Test sequential event handling with multiple emitters
	script := `
		local events = require("events")
		local counter = 0
		
		-- Create multiple emitters
		local emitters = {}
		for i = 1, 10 do
			emitters[i] = events.create_emitter()
			emitters[i]:on("increment", function()
				counter = counter + 1
			end)
		end
		
		-- Emit events on all emitters
		for i = 1, 10 do
			emitters[i]:emit("increment")
		end
		
		return counter
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	counter := lua.LVAsNumber(L.Get(-1))
	if counter != 10 {
		t.Errorf("Expected counter to be 10, got %v", counter)
	}
}

// TestMemoryLeakPrevention tests max listener warnings
func TestMemoryLeakPrevention(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(t, L)

	script := `
		local events = require("events")
		local emitter = events.create_emitter()
		
		-- Set max listeners to 3
		emitter:setMaxListeners(3)
		
		-- Add handlers beyond the limit
		for i = 1, 5 do
			emitter:on("test", function() end)
		end
		
		-- Check listener count
		return emitter:listenerCount("test")
	`

	// Capture print output to check for warning
	// Note: In a real implementation, you might want to redirect stdout
	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	count := lua.LVAsNumber(L.Get(-1))
	if count != 5 {
		t.Errorf("Expected 5 listeners to be added despite warning, got %v", count)
	}
}

// BenchmarkEventEmission benchmarks event emission performance
func BenchmarkEventEmission(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(b, L)

	// Prepare the emitter and handler
	err := L.DoString(`
		local events = require("events")
		test_emitter = events.create_emitter()
		test_emitter:on("bench", function(data) end)
	`)
	if err != nil {
		b.Fatalf("Failed to prepare benchmark: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := L.DoString(`test_emitter:emit("bench", "data")`)
		if err != nil {
			b.Fatalf("Benchmark iteration failed: %v", err)
		}
	}
}

// BenchmarkHookExecution benchmarks hook execution
func BenchmarkHookExecution(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	setupEventsLibrary(b, L)

	// Prepare hooks
	err := L.DoString(`
		local events = require("events")
		local hooks = events.hooks
		
		hooks.before("bench", function(x) return x end)
		hooks.after("bench", function(x) return x end)
		
		bench_func = function()
			return hooks.execute("bench", function(x) return x * 2 end, 10)
		end
	`)
	if err != nil {
		b.Fatalf("Failed to prepare benchmark: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		L.GetGlobal("bench_func")
		L.Call(0, 1)
		L.Pop(1)
	}
}
