// ABOUTME: Comprehensive test suite for Observability & Monitoring Library in Lua standard library
// ABOUTME: Tests metrics, tracing, logging, performance monitoring, health checks, and safety guardrails

package stdlib

import (
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// setupObservabilityLibrary loads the observability library and sets up required bridges
func setupObservabilityLibrary(t *testing.T, L *lua.LState) {
	t.Helper()

	// Set up mock metrics bridge
	metricsTable := L.NewTable()

	// Mock counter methods
	metricsTable.RawSetString("createCounter", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		// Store counter creation
		L.Push(lua.LString("counter_" + name))
		return 1
	}))

	metricsTable.RawSetString("incrementCounter", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		value := L.CheckNumber(2)
		_ = name
		_ = value
		L.Push(lua.LTrue)
		return 1
	}))

	metricsTable.RawSetString("getCounter", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LNumber(42))
		return 1
	}))

	metricsTable.RawSetString("resetCounter", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LTrue)
		return 1
	}))

	// Mock gauge methods
	metricsTable.RawSetString("createGauge", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		L.Push(lua.LString("gauge_" + name))
		return 1
	}))

	metricsTable.RawSetString("setGauge", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckNumber(2)
		L.Push(lua.LTrue)
		return 1
	}))

	metricsTable.RawSetString("increaseGauge", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckNumber(2)
		L.Push(lua.LTrue)
		return 1
	}))

	metricsTable.RawSetString("decreaseGauge", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckNumber(2)
		L.Push(lua.LTrue)
		return 1
	}))

	metricsTable.RawSetString("getGauge", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LNumber(100))
		return 1
	}))

	// Mock timer methods
	metricsTable.RawSetString("createTimer", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		L.Push(lua.LString("timer_" + name))
		return 1
	}))

	metricsTable.RawSetString("recordTimer", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckNumber(2)
		L.Push(lua.LTrue)
		return 1
	}))

	metricsTable.RawSetString("getTimerStats", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		stats := L.NewTable()
		stats.RawSetString("count", lua.LNumber(5))
		stats.RawSetString("total", lua.LNumber(1000))
		stats.RawSetString("average", lua.LNumber(200))
		stats.RawSetString("min", lua.LNumber(50))
		stats.RawSetString("max", lua.LNumber(500))
		L.Push(stats)
		return 1
	}))

	// Mock ratio counter methods
	metricsTable.RawSetString("createRatioCounter", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		L.Push(lua.LString("ratio_" + name))
		return 1
	}))

	metricsTable.RawSetString("incrementRatioNumerator", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckNumber(2)
		L.Push(lua.LTrue)
		return 1
	}))

	metricsTable.RawSetString("incrementRatioDenominator", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckNumber(2)
		L.Push(lua.LTrue)
		return 1
	}))

	metricsTable.RawSetString("getRatio", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LNumber(0.85))
		return 1
	}))

	metricsTable.RawSetString("getRatioCounts", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		counts := L.NewTable()
		counts.RawSetString("numerator", lua.LNumber(85))
		counts.RawSetString("denominator", lua.LNumber(100))
		L.Push(counts)
		return 1
	}))

	metricsTable.RawSetString("getAllMetrics", L.NewFunction(func(L *lua.LState) int {
		summary := L.NewTable()
		summary.RawSetString("counters", lua.LNumber(3))
		summary.RawSetString("gauges", lua.LNumber(2))
		summary.RawSetString("timers", lua.LNumber(4))
		summary.RawSetString("ratios", lua.LNumber(1))
		L.Push(summary)
		return 1
	}))

	L.SetGlobal("metrics", metricsTable)

	// Set up mock tracing bridge
	tracingTable := L.NewTable()

	tracingTable.RawSetString("startSpan", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		_ = name
		spanID := "span_" + name + "_123"
		L.Push(lua.LString(spanID))
		return 1
	}))

	tracingTable.RawSetString("endSpan", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LTrue)
		return 1
	}))

	tracingTable.RawSetString("setSpanAttribute", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckString(2)
		_ = L.Get(3)
		L.Push(lua.LTrue)
		return 1
	}))

	tracingTable.RawSetString("addSpanEvent", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckString(2)
		L.Push(lua.LTrue)
		return 1
	}))

	tracingTable.RawSetString("setSpanStatus", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckString(2)
		L.Push(lua.LTrue)
		return 1
	}))

	tracingTable.RawSetString("recordError", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckString(2)
		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("tracing", tracingTable)

	// Set up mock slog bridge
	slogTable := L.NewTable()

	slogTable.RawSetString("debug", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LTrue)
		return 1
	}))

	slogTable.RawSetString("info", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LTrue)
		return 1
	}))

	slogTable.RawSetString("warn", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LTrue)
		return 1
	}))

	slogTable.RawSetString("error", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("slog", slogTable)

	// Set up mock events bridge
	eventsTable := L.NewTable()

	eventsTable.RawSetString("subscribe", L.NewFunction(func(L *lua.LState) int {
		pattern := L.CheckString(1)
		_ = pattern
		L.Push(lua.LString("subscription_123"))
		return 1
	}))

	eventsTable.RawSetString("unsubscribe", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("events", eventsTable)

	// Set up optional mock guardrails bridge
	guardrailsTable := L.NewTable()

	guardrailsTable.RawSetString("createGuardrail", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		guardrailID := "guardrail_" + name + "_456"
		L.Push(lua.LString(guardrailID))
		return 1
	}))

	guardrailsTable.RawSetString("validateGuardrail", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.Get(2)
		result := L.NewTable()
		result.RawSetString("valid", lua.LTrue)
		result.RawSetString("violations", L.NewTable())
		L.Push(result)
		return 1
	}))

	guardrailsTable.RawSetString("getGuardrailMetrics", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		metrics := L.NewTable()
		metrics.RawSetString("validations", lua.LNumber(10))
		metrics.RawSetString("violations", lua.LNumber(2))
		L.Push(metrics)
		return 1
	}))

	L.SetGlobal("guardrails", guardrailsTable)

	// Load the observability library
	observabilityPath := filepath.Join(".", "observability.lua")
	err := L.DoFile(observabilityPath)
	if err != nil {
		t.Fatalf("Failed to load observability library: %v", err)
	}
	observability := L.Get(-1)
	L.SetGlobal("observability", observability)
}

// TestObservabilityLibraryLoading tests that the observability library can be loaded
func TestObservabilityLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupObservabilityLibrary(t, L)

	// Check that observability table exists and has expected functions
	script := `
		if type(observability) ~= "table" then
			error("Observability module should be a table")
		end
		
		local required_functions = {
			"counter", "gauge", "timer", "ratio_counter",
			"track", "start_span", "trace",
			"logger", "debug", "info", "warn", "error",
			"health_check", "monitor_events", "guardrail",
			"get_metrics_summary", "get_system_info", "cleanup"
		}
		
		for _, func_name in ipairs(required_functions) do
			if type(observability[func_name]) ~= "function" then
				error("Function " .. func_name .. " should be available")
			end
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Observability library structure test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected library structure test to pass")
	}
}

// TestMetricsManagement tests metrics creation and usage
func TestMetricsManagement(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_and_use_counter",
			script: `
				local counter = observability.counter("test_counter", "Test counter metric")
				counter.increment(5)
				local value = counter.get()
				counter.reset()
				return value == 42  -- Mock returns 42
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected counter operations to work, got %v", result)
				}
			},
		},
		{
			name: "create_and_use_gauge",
			script: `
				local gauge = observability.gauge("test_gauge", "Test gauge metric")
				gauge.set(100)
				gauge.increase(10)
				gauge.decrease(5)
				local value = gauge.get()
				return value == 100  -- Mock returns 100
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected gauge operations to work, got %v", result)
				}
			},
		},
		{
			name: "create_and_use_timer",
			script: `
				local timer = observability.timer("test_timer", "Test timer metric")
				timer.record(250)
				
				local stopwatch = timer.start()
				local duration = stopwatch.stop()
				
				local stats = timer.get_stats()
				return stats.count == 5 and stats.average == 200
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected timer operations to work, got %v", result)
				}
			},
		},
		{
			name: "create_and_use_ratio_counter",
			script: `
				local ratio = observability.ratio_counter("test_ratio", "Test ratio metric")
				ratio.increment_numerator(85)
				ratio.increment_denominator(100)
				
				local ratio_value = ratio.get_ratio()
				local counts = ratio.get_counts()
				
				return ratio_value == 0.85 and counts.numerator == 85 and counts.denominator == 100
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected ratio counter operations to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupObservabilityLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestPerformanceMonitoring tests function tracking and performance monitoring
func TestPerformanceMonitoring(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "track_function_execution",
			script: `
				local function test_func(x, y)
					return x + y
				end
				
				local tracked_func = observability.track(test_func, "add_numbers")
				local result = tracked_func(5, 3)
				
				return result == 8
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected function tracking to work, got %v", result)
				}
			},
		},
		{
			name: "track_function_with_options",
			script: `
				local function test_func(msg)
					return "processed: " .. msg
				end
				
				local tracked_func = observability.track(test_func, "process_message", {
					include_args = true,
					include_result = true,
					auto_metric = true
				})
				local result = tracked_func("hello")
				
				return result == "processed: hello"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected function tracking with options to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupObservabilityLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestDistributedTracing tests tracing capabilities
func TestDistributedTracing(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "start_and_finish_span",
			script: `
				local span = observability.start_span("test_operation")
				span.add_attribute("operation_type", "test")
				span.add_event("processing_started")
				span.set_status("ok")
				span.finish()
				
				return span.name == "test_operation"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected span operations to work, got %v", result)
				}
			},
		},
		{
			name: "trace_function",
			script: `
				local function test_func(x)
					return x * 2
				end
				
				local traced_func = observability.trace(test_func, "multiply_operation")
				local result = traced_func(21)
				
				return result == 42
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected function tracing to work, got %v", result)
				}
			},
		},
		{
			name: "trace_function_with_error",
			script: `
				local function error_func()
					error("Test error")
				end
				
				local traced_func = observability.trace(error_func, "error_operation")
				
				local success, err = pcall(traced_func)
				return not success and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error tracing to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupObservabilityLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestStructuredLogging tests logging capabilities
func TestStructuredLogging(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "basic_logging",
			script: `
				observability.debug("Debug message", {key = "value"})
				observability.info("Info message", {count = 42})
				observability.warn("Warning message")
				observability.error("Error message", {error_code = 500})
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected basic logging to work, got %v", result)
				}
			},
		},
		{
			name: "custom_logger",
			script: `
				local logger = observability.logger("test_service", {
					level = "info",
					context = {service = "test", version = "1.0"}
				})
				
				logger.info("Service started", {port = 8080})
				logger.warn("High memory usage", {memory_mb = 512})
				
				local contextual_logger = logger.with_context({request_id = "123"})
				contextual_logger.info("Request processed")
				
				return logger.name == "test_service"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected custom logger to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupObservabilityLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestHealthChecks tests health monitoring capabilities
func TestHealthChecks(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "healthy_check",
			script: `
				local health_check = observability.health_check("database", function()
					return {status = "connected", connections = 5}
				end)
				
				local result = health_check.check()
				local metrics = health_check.get_metrics()
				
				return result.healthy == true and result.status == "ok" and metrics ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected healthy check to work, got %v", result)
				}
			},
		},
		{
			name: "unhealthy_check",
			script: `
				local health_check = observability.health_check("api", function()
					error("Connection timeout")
				end)
				
				local result = health_check.check()
				
				return result.healthy == false and result.status == "error" and type(result.error) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected unhealthy check to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupObservabilityLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestEventMonitoring tests event monitoring capabilities
func TestEventMonitoring(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupObservabilityLibrary(t, L)

	script := `
		local event_count = 0
		local monitor = observability.monitor_events("user.*", function(event_data)
			event_count = event_count + 1
			return true
		end)
		
		-- Simulate some events being processed
		-- In real usage, these would be actual events from the event bridge
		
		local count = monitor.get_count()
		monitor.unsubscribe()
		
		return monitor.pattern == "user.*" and type(count) == "number"
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Event monitoring test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected event monitoring to work correctly, got %v", result)
	}
}

// TestGuardrails tests safety monitoring capabilities
func TestGuardrails(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "bridge_based_guardrail",
			script: `
				local guardrail = observability.guardrail("content_filter", function(data)
					return not string.find(data.content or "", "blocked_word")
				end)
				
				local valid_result = guardrail.validate({content = "safe content"})
				local metrics = guardrail.get_metrics()
				
				return guardrail.name == "content_filter" and metrics ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected bridge-based guardrail to work, got %v", result)
				}
			},
		},
		{
			name: "local_guardrail_fallback",
			script: `
				-- Test with guardrails bridge disabled
				_G.guardrails = nil
				
				local guardrail = observability.guardrail("local_filter", function(data)
					return data.value > 0
				end)
				
				local valid_result, err = guardrail.validate({value = 10})
				local invalid_result, err2 = guardrail.validate({value = -1})
				local metrics = guardrail.get_metrics()
				
				return valid_result == true and invalid_result == false and metrics ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected local guardrail fallback to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupObservabilityLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestObservabilityUtilityFunctions tests utility and system information functions
func TestObservabilityUtilityFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupObservabilityLibrary(t, L)

	script := `
		-- Test metrics summary
		local summary = observability.get_metrics_summary()
		
		-- Test system info
		local system_info = observability.get_system_info()
		
		-- Test cleanup
		observability.cleanup()
		
		return summary.counters == 3 and 
		       system_info.bridges_available.metrics == true and
		       type(system_info.lua_version) == "string"
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Utility functions test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected utility functions to work correctly, got %v", result)
	}
}

// TestObservabilityErrorHandling tests error handling in observability operations
func TestObservabilityErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "missing_required_parameters",
			script: `
				local success, err = pcall(function()
					observability.counter(nil)
				end)
				return not success and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing parameters, got %v", result)
				}
			},
		},
		{
			name: "invalid_parameter_types",
			script: `
				local success, err = pcall(function()
					observability.track("not a function", "test")
				end)
				return not success and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid parameter types, got %v", result)
				}
			},
		},
		{
			name: "missing_bridge_error",
			script: `
				-- Temporarily remove bridge
				local original_bridge = _G.metrics
				_G.metrics = nil
				
				local success, err = pcall(function()
					observability.counter("test")
				end)
				
				-- Restore bridge
				_G.metrics = original_bridge
				
				return not success and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing bridge, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupObservabilityLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestObservabilityIntegration tests integration between different observability components
func TestObservabilityIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupObservabilityLibrary(t, L)

	script := `
		-- Create a comprehensive monitoring setup
		local request_counter = observability.counter("http_requests", "HTTP request count")
		local response_timer = observability.timer("http_response_time", "HTTP response time")
		local logger = observability.logger("api_server", {context = {service = "api"}})
		
		-- Create a monitored function that uses multiple observability features
		local function process_request(request_id, data)
			local span = observability.start_span("process_request")
			span.add_attribute("request_id", request_id)
			
			request_counter.increment(1)
			local timer = response_timer.start()
			
			logger.info("Processing request", {request_id = request_id})
			
			-- Simulate processing
			local result = "processed: " .. data
			
			local duration = timer.stop()
			span.add_attribute("duration_ms", duration)
			span.set_status("ok")
			span.finish()
			
			logger.info("Request completed", {request_id = request_id, duration_ms = duration})
			
			return result
		end
		
		-- Process some requests
		local result1 = process_request("req_1", "data1")
		local result2 = process_request("req_2", "data2")
		
		-- Verify results
		local counter_value = request_counter.get()
		local timer_stats = response_timer.get_stats()
		local system_info = observability.get_system_info()
		
		return result1 == "processed: data1" and 
		       result2 == "processed: data2" and
		       counter_value == 42 and  -- Mock returns 42
		       timer_stats.count == 5 and  -- Mock returns stats
		       system_info.bridges_available.metrics == true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected integration test to work correctly, got %v", result)
	}
}

// TestObservabilityPackageRequire tests that the module can be required as a package
func TestObservabilityPackageRequire(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Set up mock bridges
	setupObservabilityLibrary(t, L)

	script := `
		-- Test that observability is available globally
		if type(observability) ~= "table" then
			error("Observability module should be available globally")
		end
		
		-- Test basic functionality
		local counter = observability.counter("require_test", "Test counter")
		counter.increment(1)
		
		local logger = observability.logger("test")
		logger.info("Module test successful")
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Package availability test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected package test to pass, got %v", result)
	}
}

// BenchmarkObservabilityOperations benchmarks key observability operations
func BenchmarkObservabilityOperations(b *testing.B) {
	benchmarks := []struct {
		name   string
		script string
	}{
		{
			name: "counter_operations",
			script: `
				local counter = observability.counter("bench_counter", "Benchmark counter")
				counter.increment(1)
				counter.get()
			`,
		},
		{
			name: "timer_operations",
			script: `
				local timer = observability.timer("bench_timer", "Benchmark timer")
				timer.record(100)
				timer.get_stats()
			`,
		},
		{
			name: "span_operations",
			script: `
				local span = observability.start_span("bench_span")
				span.add_attribute("test", "value")
				span.set_status("ok")
				span.finish()
			`,
		},
		{
			name: "logging_operations",
			script: `
				observability.info("Benchmark log message", {key = "value"})
			`,
		},
		{
			name: "function_tracking",
			script: `
				local function test_func(x) return x * 2 end
				local tracked = observability.track(test_func, "bench_func")
				tracked(42)
			`,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			L := lua.NewState()
			defer L.Close()

			setupObservabilityLibrary(nil, L) // Skip t.Helper in benchmark

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := L.DoString(bm.script)
				if err != nil {
					b.Fatalf("Benchmark failed: %v", err)
				}
				L.Pop(1) // Clean stack
			}
		})
	}
}
