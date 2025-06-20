// ABOUTME: Comprehensive test suite for Error Handling & Recovery Library in Lua standard library
// ABOUTME: Tests try-catch-finally, retry logic, circuit breakers, error categorization, and recovery strategies

package stdlib

import (
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// setupErrorsLibrary loads the errors library and sets up required bridges
func setupErrorsLibrary(t *testing.T, L *lua.LState) {
	t.Helper()

	// Set up mock error utilities bridge
	errorsTable := L.NewTable()

	// Mock error creation methods
	errorsTable.RawSetString("createError", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		code := L.OptString(2, "CUSTOM_ERROR")
		context := L.OptTable(3, L.NewTable())
		_ = context

		error_obj := L.NewTable()
		error_obj.RawSetString("message", lua.LString(message))
		error_obj.RawSetString("code", lua.LString(code))
		error_obj.RawSetString("timestamp", lua.LNumber(1234567890))
		L.Push(error_obj)
		return 1
	}))

	errorsTable.RawSetString("wrapError", L.NewFunction(func(L *lua.LState) int {
		originalError := L.CheckTable(1)
		context := L.CheckTable(2)
		_ = context

		wrapped := L.NewTable()
		wrapped.RawSetString("original", originalError)
		wrapped.RawSetString("wrapped", lua.LTrue)
		wrapped.RawSetString("context", context)
		L.Push(wrapped)
		return 1
	}))

	errorsTable.RawSetString("chainErrors", L.NewFunction(func(L *lua.LState) int {
		errorsArray := L.CheckTable(1)
		_ = errorsArray

		chained := L.NewTable()
		chained.RawSetString("type", lua.LString("chained"))
		chained.RawSetString("count", lua.LNumber(3))
		L.Push(chained)
		return 1
	}))

	// Mock backoff strategy methods
	errorsTable.RawSetString("createExponentialBackoffStrategy", L.NewFunction(func(L *lua.LState) int {
		initialDelay := L.CheckNumber(1)
		maxDelay := L.CheckNumber(2)
		jitter := L.CheckBool(3)
		_ = initialDelay
		_ = maxDelay
		_ = jitter

		strategy := L.NewTable()
		strategy.RawSetString("type", lua.LString("exponential_backoff"))
		strategy.RawSetString("initial_delay", lua.LNumber(initialDelay))
		strategy.RawSetString("max_delay", lua.LNumber(maxDelay))
		strategy.RawSetString("jitter", lua.LBool(jitter))
		L.Push(strategy)
		return 1
	}))

	errorsTable.RawSetString("createLinearBackoffStrategy", L.NewFunction(func(L *lua.LState) int {
		initialDelay := L.CheckNumber(1)
		maxDelay := L.CheckNumber(2)

		strategy := L.NewTable()
		strategy.RawSetString("type", lua.LString("linear_backoff"))
		strategy.RawSetString("initial_delay", lua.LNumber(initialDelay))
		strategy.RawSetString("max_delay", lua.LNumber(maxDelay))
		L.Push(strategy)
		return 1
	}))

	errorsTable.RawSetString("createCircuitBreakerStrategy", L.NewFunction(func(L *lua.LState) int {
		threshold := L.CheckNumber(1)
		timeout := L.CheckNumber(2)
		resetTimeout := L.CheckNumber(3)

		breaker := L.NewTable()
		breaker.RawSetString("type", lua.LString("circuit_breaker"))
		breaker.RawSetString("failure_threshold", lua.LNumber(threshold))
		breaker.RawSetString("timeout", lua.LNumber(timeout))
		breaker.RawSetString("reset_timeout", lua.LNumber(resetTimeout))
		breaker.RawSetString("state", lua.LString("closed"))
		L.Push(breaker)
		return 1
	}))

	errorsTable.RawSetString("createFallbackStrategy", L.NewFunction(func(L *lua.LState) int {
		fallbackValue := L.CheckAny(1)

		strategy := L.NewTable()
		strategy.RawSetString("type", lua.LString("fallback"))
		strategy.RawSetString("fallback_value", fallbackValue)
		L.Push(strategy)
		return 1
	}))

	// Mock error inspection methods
	errorsTable.RawSetString("isRetryableError", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckAny(1)
		L.Push(lua.LTrue) // Default to retryable for testing
		return 1
	}))

	errorsTable.RawSetString("isFatalError", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckAny(1)
		L.Push(lua.LFalse) // Default to non-fatal for testing
		return 1
	}))

	errorsTable.RawSetString("categorizeError", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckAny(1)
		L.Push(lua.LString("network")) // Default category
		return 1
	}))

	// Mock error aggregation methods
	errorsTable.RawSetString("createErrorAggregator", L.NewFunction(func(L *lua.LState) int {
		aggregator := L.NewTable()
		aggregator.RawSetString("type", lua.LString("aggregator"))
		aggregator.RawSetString("errors", L.NewTable())
		L.Push(aggregator)
		return 1
	}))

	errorsTable.RawSetString("addErrorToAggregator", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		_ = L.CheckAny(2)
		L.Push(lua.LTrue)
		return 1
	}))

	errorsTable.RawSetString("finalizeErrorAggregator", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)

		aggregated := L.NewTable()
		aggregated.RawSetString("type", lua.LString("aggregated"))
		aggregated.RawSetString("count", lua.LNumber(3))
		L.Push(aggregated)
		return 1
	}))

	// Mock serialization methods
	errorsTable.RawSetString("errorToJSON", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckAny(1)
		L.Push(lua.LString(`{"message":"test error","code":"TEST_ERROR","timestamp":1234567890}`))
		return 1
	}))

	errorsTable.RawSetString("errorFromJSON", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)

		error_obj := L.NewTable()
		error_obj.RawSetString("message", lua.LString("test error"))
		error_obj.RawSetString("code", lua.LString("TEST_ERROR"))
		error_obj.RawSetString("timestamp", lua.LNumber(1234567890))
		L.Push(error_obj)
		return 1
	}))

	// Mock context methods
	errorsTable.RawSetString("getErrorContext", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckAny(1)

		context := L.NewTable()
		context.RawSetString("user_id", lua.LString("test_user"))
		context.RawSetString("request_id", lua.LString("req_123"))
		L.Push(context)
		return 1
	}))

	errorsTable.RawSetString("addErrorContext", L.NewFunction(func(L *lua.LState) int {
		errorObj := L.CheckTable(1)
		key := L.CheckString(2)
		value := L.CheckAny(3)

		errorObj.RawSetString(key, value)
		L.Push(errorObj)
		return 1
	}))

	// Mock event methods
	errorsTable.RawSetString("emitErrorEvent", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckTable(2)
		L.Push(lua.LTrue)
		return 1
	}))

	errorsTable.RawSetString("subscribeToErrorEvents", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		_ = L.CheckFunction(2)
		L.Push(lua.LString("subscription_123"))
		return 1
	}))

	// Mock circuit breaker execution
	errorsTable.RawSetString("executeWithCircuitBreaker", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		fn := L.CheckFunction(2)

		// Execute the function
		L.Push(fn)
		L.Call(0, 1)
		return 1
	}))

	// Mock backoff delay calculation
	errorsTable.RawSetString("calculateBackoffDelay", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		attempt := L.CheckNumber(2)

		// Simple delay calculation for testing
		delay := attempt * 1000
		L.Push(lua.LNumber(delay))
		return 1
	}))

	// Mock error category registration
	errorsTable.RawSetString("registerErrorCategory", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckFunction(2)
		L.Push(lua.LTrue)
		return 1
	}))

	// Set the bridge as global
	L.SetGlobal("util_errors", errorsTable)

	// Load the errors library
	libPath := filepath.Join(".", "errors.lua")
	if err := L.DoFile(libPath); err != nil {
		t.Fatalf("Failed to load errors library: %v", err)
	}
	errors := L.Get(-1)
	L.SetGlobal("errors", errors)
}

func TestErrorsLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	// Test that the errors library was loaded
	script := `
		return type(errors) == "table" and
		       type(errors.try) == "function" and
		       type(errors.retry) == "function" and
		       type(errors.circuit_breaker) == "function"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected errors library to be properly loaded")
	}
}

func TestErrorsHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "try_catch_finally_success",
			script: `
				local result = ""
				local final_result = errors.try(
					function()
						result = result .. "try"
						return "success"
					end,
					function(err)
						result = result .. "catch"
						return "caught"
					end,
					function()
						result = result .. "finally"
					end
				)
				return result == "tryfinally" and final_result == "success"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected try-catch-finally to work correctly")
				}
			},
		},
		{
			name: "try_catch_finally_error",
			script: `
				-- Test that catch function is called and returns correct value
				local final_result = errors.try(
					function()
						error("test error")
					end,
					function(err)
						-- Catch function should receive the error and return a value
						return "caught"
					end,
					function()
						-- Finally function should execute
					end
				)
				
				-- Since error was caught and handled, result should be "caught"
				return final_result == "caught"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected try-catch-finally error handling to work, got %v (%T)", result, result)
				}
			},
		},
		{
			name: "create_custom_error",
			script: `
				local err = errors.create("Test error message", "TEST_CODE", {context = "test"})
				return type(err) == "table" and 
				       err.message == "Test error message" and
				       err.code == "TEST_CODE"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected custom error creation to work")
				}
			},
		},
		{
			name: "wrap_error",
			script: `
				local original = errors.create("Original error", "ORIG_CODE")
				local wrapped = errors.wrap(original, {additional = "context"})
				return type(wrapped) == "table" and 
				       wrapped.wrapped == true and
				       wrapped.original == original
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error wrapping to work")
				}
			},
		},
		{
			name: "chain_errors",
			script: `
				local err1 = errors.create("Error 1", "ERR1")
				local err2 = errors.create("Error 2", "ERR2")
				local err3 = errors.create("Error 3", "ERR3")
				
				local chained = errors.chain({err1, err2, err3})
				return type(chained) == "table" and 
				       chained.type == "chained" and
				       chained.count == 3
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error chaining to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestRetryAndRecovery(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "retry_success_first_attempt",
			script: `
				local attempts = 0
				local result = errors.retry(function()
					attempts = attempts + 1
					return "success"
				end, {max_attempts = 3})
				
				return result == "success" and attempts == 1
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected retry to succeed on first attempt")
				}
			},
		},
		{
			name: "circuit_breaker_creation",
			script: `
				local counter = 0
				local protected_func = errors.circuit_breaker(function()
					counter = counter + 1
					return "result_" .. counter
				end, {
					failure_threshold = 3,
					timeout = 30000,
					reset_timeout = 15000
				})
				
				local result1 = protected_func()
				local result2 = protected_func()
				
				return result1 == "result_1" and 
				       result2 == "result_2" and
				       counter == 2
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected circuit breaker to work correctly")
				}
			},
		},
		{
			name: "fallback_strategy",
			script: `
				-- Test that fallback function is called when primary fails
				local protected_func = errors.fallback(
					function()
						error("primary failed")
					end,
					function()
						return "fallback_result"
					end
				)
				
				local result = protected_func()
				
				-- If we get here, the fallback worked
				return result == "fallback_result"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected fallback strategy to work")
				}
			},
		},
		{
			name: "create_recovery_strategies",
			script: `
				local exp_strategy = errors.create_recovery_strategy("exponential_backoff", {
					initial_delay = 1000,
					max_delay = 30000,
					jitter = true
				})
				
				local lin_strategy = errors.create_recovery_strategy("linear_backoff", {
					initial_delay = 500,
					max_delay = 10000
				})
				
				local cb_strategy = errors.create_recovery_strategy("circuit_breaker", {
					failure_threshold = 5,
					timeout = 60000,
					reset_timeout = 30000
				})
				
				return exp_strategy.type == "exponential_backoff" and
				       lin_strategy.type == "linear_backoff" and
				       cb_strategy.type == "circuit_breaker"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected recovery strategy creation to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestErrorCategorizationAndReporting(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "categorize_error",
			script: `
				local err = errors.create("Network timeout", "TIMEOUT")
				local category = errors.categorize(err)
				return category == "network"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error categorization to work")
				}
			},
		},
		{
			name: "check_error_properties",
			script: `
				local err = errors.create("Test error", "TEST")
				local is_retryable = errors.is_retryable(err)
				local is_fatal = errors.is_fatal(err)
				
				return is_retryable == true and is_fatal == false
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error property checks to work")
				}
			},
		},
		{
			name: "aggregate_errors",
			script: `
				local err1 = errors.create("Error 1", "ERR1")
				local err2 = errors.create("Error 2", "ERR2")
				local err3 = errors.create("Error 3", "ERR3")
				
				local aggregated = errors.aggregate({err1, err2, err3})
				
				return aggregated.type == "aggregated" and aggregated.count == 3
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error aggregation to work")
				}
			},
		},
		{
			name: "register_error_category",
			script: `
				local registered = errors.register_category("custom", function(err)
					return err.code == "CUSTOM"
				end)
				
				return registered == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error category registration to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestSerializationAndContext(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "error_serialization",
			script: `
				local err = errors.create("Test error", "TEST_CODE")
				local json = errors.to_json(err)
				local deserialized = errors.from_json(json)
				
				return type(json) == "string" and
				       type(deserialized) == "table" and
				       deserialized.message == "test error"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error serialization to work")
				}
			},
		},
		{
			name: "error_context_management",
			script: `
				local err = errors.create("Test error", "TEST")
				local context = errors.get_context(err)
				local updated = errors.add_context(err, "operation", "test_operation")
				
				return type(context) == "table" and
				       context.user_id == "test_user" and
				       updated.operation == "test_operation"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected context management to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestEventHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "log_error_event",
			script: `
				local logged = errors.log_error("test_error", {
					component = "test",
					severity = "high"
				})
				
				return logged == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error event logging to work")
				}
			},
		},
		{
			name: "subscribe_to_error_events",
			script: `
				local subscription_id = errors.subscribe_to_errors(
					{"network", "validation"},
					function(event)
						-- Handle error event
					end
				)
				
				return type(subscription_id) == "string" and subscription_id == "subscription_123"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error event subscription to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestErrorsUtilityFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "safe_function_wrapper",
			script: `
				local safe_func = errors.safe(function()
					error("something went wrong")
				end, "default_value")
				
				local result = safe_func()
				return result == "default_value"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected safe function wrapper to work")
				}
			},
		},
		{
			name: "timeout_wrapper",
			script: `
				local timeout_func = errors.timeout(function()
					return "completed"
				end, 5000)
				
				local result = timeout_func()
				return result == "completed"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected timeout wrapper to work")
				}
			},
		},
		{
			name: "system_info",
			script: `
				local info = errors.get_system_info()
				return type(info) == "table" and
				       type(info.lua_version) == "string" and
				       type(info.bridges_available) == "table" and
				       info.bridges_available.errors == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected system info to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestErrorsValidation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "missing_required_parameters",
			script: `
				local success, err = pcall(function()
					errors.create()
				end)
				return success == false and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected missing parameter validation")
				}
			},
		},
		{
			name: "invalid_parameter_types",
			script: `
				local success, err = pcall(function()
					errors.create(123, "CODE") -- message should be string
				end)
				return success == false and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter type validation")
				}
			},
		},
		{
			name: "invalid_retry_options",
			script: `
				local success, err = pcall(function()
					errors.retry("not_a_function", {max_attempts = 3})
				end)
				return success == false and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected function type validation")
				}
			},
		},
		{
			name: "missing_bridge_graceful_handling",
			script: `
				-- Temporarily remove bridge
				local original_bridge = _G.util_errors
				_G.util_errors = nil
				
				local success, err = pcall(function()
					errors.create("test", "TEST")
				end)
				
				-- Restore bridge
				_G.util_errors = original_bridge
				
				return success == false and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected graceful handling when bridge is missing")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestErrorsIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	script := `
		-- Test comprehensive error handling workflow
		local all_tests_passed = true
		
		-- Test 1: Error creation and categorization
		local err = errors.create("Network timeout", "NETWORK_TIMEOUT")
		local category = errors.categorize(err)
		if category ~= "network" then
			all_tests_passed = false
		end
		
		-- Test 2: Try-catch-finally flow
		local try_result = errors.try(
			function()
				return "success"
			end,
			function(err)
				return "caught"
			end,
			function()
				-- Finally block
			end
		)
		if try_result ~= "success" then
			all_tests_passed = false
		end
		
		-- Test 3: Fallback strategy
		local fallback_func = errors.fallback(
			function() return "primary_success" end,
			function() return "fallback_value" end
		)
		local fallback_result = fallback_func()
		if fallback_result ~= "primary_success" then
			all_tests_passed = false
		end
		
		-- Test 4: Error serialization
		local test_error = errors.create("Serialization test", "SERIAL_TEST")
		local json = errors.to_json(test_error)
		local deserialized = errors.from_json(json)
		if type(json) ~= "string" or type(deserialized) ~= "table" then
			all_tests_passed = false
		end
		
		-- Test 5: Error aggregation
		local err1 = errors.create("Error 1", "ERR1")
		local err2 = errors.create("Error 2", "ERR2")
		local aggregated = errors.aggregate({err1, err2})
		if aggregated.type ~= "aggregated" then
			all_tests_passed = false
		end
		
		-- Test 6: Recovery strategy creation
		local strategy = errors.create_recovery_strategy("exponential_backoff", {
			initial_delay = 1000,
			max_delay = 30000
		})
		if strategy.type ~= "exponential_backoff" then
			all_tests_passed = false
		end
		
		return all_tests_passed
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected comprehensive error handling integration to work")
	}
}

func TestErrorsPackageRequire(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupErrorsLibrary(t, L)

	script := `
		-- Test that errors can be required as a module
		local errors_module = errors
		
		return type(errors_module) == "table" and
		       type(errors_module.try) == "function" and
		       type(errors_module.retry) == "function" and
		       type(errors_module.circuit_breaker) == "function" and
		       type(errors_module.create) == "function" and
		       type(errors_module.wrap) == "function" and
		       type(errors_module.categorize) == "function" and
		       type(errors_module.to_json) == "function" and
		       type(errors_module.from_json) == "function"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Package require test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected errors module to be properly exported")
	}
}
