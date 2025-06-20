// ABOUTME: Comprehensive test suite for Logging & Debug Library in Lua standard library
// ABOUTME: Tests logging, debugging, profiling, metrics, audit logging, and hook integration

package stdlib

import (
	"fmt"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// setupLoggingLibrary loads the logging library and sets up required bridges
func setupLoggingLibrary(t *testing.T, L *lua.LState) {
	t.Helper()

	// Set up mock debug bridge
	debugTable := L.NewTable()

	// Track enabled components
	enabledComponents := make(map[string]bool)

	debugTable.RawSetString("debugPrintf", L.NewFunction(func(L *lua.LState) int {
		component := L.CheckString(1)
		format := L.CheckString(2)
		_ = component
		_ = format
		// Just consume the arguments
		return 0
	}))

	debugTable.RawSetString("debugPrintln", L.NewFunction(func(L *lua.LState) int {
		component := L.CheckString(1)
		message := L.CheckString(2)
		_ = component
		_ = message
		return 0
	}))

	debugTable.RawSetString("isDebugEnabled", L.NewFunction(func(L *lua.LState) int {
		component := L.CheckString(1)
		L.Push(lua.LBool(enabledComponents[component]))
		return 1
	}))

	debugTable.RawSetString("enableDebugComponent", L.NewFunction(func(L *lua.LState) int {
		component := L.CheckString(1)
		enabledComponents[component] = true
		return 0
	}))

	debugTable.RawSetString("disableDebugComponent", L.NewFunction(func(L *lua.LState) int {
		component := L.CheckString(1)
		delete(enabledComponents, component)
		return 0
	}))

	L.SetGlobal("util_debug", debugTable)

	// Set up mock slog bridge
	slogTable := L.NewTable()

	// Store hooks for testing
	beforeGenerateHooks := []lua.LValue{}
	afterGenerateHooks := []lua.LValue{}
	beforeToolCallHooks := []lua.LValue{}
	afterToolCallHooks := []lua.LValue{}

	slogTable.RawSetString("info", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.OptTable(2, L.NewTable())
		return 0
	}))

	slogTable.RawSetString("warn", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.OptTable(2, L.NewTable())
		return 0
	}))

	slogTable.RawSetString("error", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.OptTable(2, L.NewTable())
		return 0
	}))

	slogTable.RawSetString("debug", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.OptTable(2, L.NewTable())
		return 0
	}))

	slogTable.RawSetString("registerBeforeGenerateHook", L.NewFunction(func(L *lua.LState) int {
		hook := L.CheckFunction(1)
		beforeGenerateHooks = append(beforeGenerateHooks, hook)
		return 0
	}))

	slogTable.RawSetString("registerAfterGenerateHook", L.NewFunction(func(L *lua.LState) int {
		hook := L.CheckFunction(1)
		afterGenerateHooks = append(afterGenerateHooks, hook)
		return 0
	}))

	slogTable.RawSetString("registerBeforeToolCallHook", L.NewFunction(func(L *lua.LState) int {
		hook := L.CheckFunction(1)
		beforeToolCallHooks = append(beforeToolCallHooks, hook)
		return 0
	}))

	slogTable.RawSetString("registerAfterToolCallHook", L.NewFunction(func(L *lua.LState) int {
		hook := L.CheckFunction(1)
		afterToolCallHooks = append(afterToolCallHooks, hook)
		return 0
	}))

	L.SetGlobal("util_slog", slogTable)

	// Set up mock script logger bridge
	scriptLoggerTable := L.NewTable()

	// Store logged messages for testing
	loggedMessages := []map[string]interface{}{}

	scriptLoggerTable.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		level := L.CheckNumber(2)
		attributes := L.OptTable(3, L.NewTable())

		// Store the log entry
		entry := map[string]interface{}{
			"message": message,
			"level":   int(level),
		}

		// Extract attributes
		attrs := make(map[string]interface{})
		attributes.ForEach(func(k, v lua.LValue) {
			if key, ok := k.(lua.LString); ok {
				attrs[string(key)] = v
			}
		})
		entry["attributes"] = attrs

		loggedMessages = append(loggedMessages, entry)
		return 0
	}))

	scriptLoggerTable.RawSetString("configure", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		return 0
	}))

	scriptLoggerTable.RawSetString("debug", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.OptTable(2, L.NewTable())
		return 0
	}))

	scriptLoggerTable.RawSetString("info", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.OptTable(2, L.NewTable())
		return 0
	}))

	scriptLoggerTable.RawSetString("warn", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.OptTable(2, L.NewTable())
		return 0
	}))

	scriptLoggerTable.RawSetString("error", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.OptTable(2, L.NewTable())
		return 0
	}))

	// Expose logged messages for testing
	scriptLoggerTable.RawSetString("_getLoggedMessages", L.NewFunction(func(L *lua.LState) int {
		table := L.NewTable()
		for _, entry := range loggedMessages {
			entryTable := L.NewTable()
			entryTable.RawSetString("message", lua.LString(entry["message"].(string)))
			entryTable.RawSetString("level", lua.LNumber(entry["level"].(int)))
			table.Append(entryTable)
		}
		L.Push(table)
		return 1
	}))

	L.SetGlobal("util_script_logger", scriptLoggerTable)

	// Load the logging library
	libPath := filepath.Join(".", "logging.lua")
	if err := L.DoFile(libPath); err != nil {
		t.Fatalf("Failed to load logging library: %v", err)
	}
	logging := L.Get(-1)
	L.SetGlobal("logging", logging)
}

func TestLoggingLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	// Test that the logging library was loaded
	script := `
		return type(logging) == "table" and
		       type(logging.create) == "function" and
		       type(logging.default) == "function" and
		       type(logging.configure) == "function" and
		       type(logging.debug) == "table" and
		       type(logging.hooks) == "table" and
		       type(logging.metrics) == "table" and
		       type(logging.audit) == "table"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected logging library to be properly loaded")
	}
}

func TestLoggerCreationAndConfiguration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_logger",
			script: `
				local logger = logging.create("test_logger")
				return type(logger) == "table" and
				       type(logger.debug) == "function" and
				       type(logger.info) == "function" and
				       type(logger.warn) == "function" and
				       type(logger.error) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected logger to be created with logging methods")
				}
			},
		},
		{
			name: "default_logger",
			script: `
				local log1 = logging.default()
				local log2 = logging.default()
				return log1 == log2  -- Should return same instance
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected default logger to return same instance")
				}
			},
		},
		{
			name: "configure_logging",
			script: `
				logging.configure({
					level = "debug",
					format = "json",
					emoji = true,
					components = {"auth", "database"}
				})
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected configuration to succeed")
				}
			},
		},
		{
			name: "logger_with_context",
			script: `
				local logger = logging.create("context_test")
				logger:with_context({
					user_id = "123",
					request_id = "req_456"
				})
				logger:info("Test message")
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected logger context to work")
				}
			},
		},
		{
			name: "child_logger",
			script: `
				local parent = logging.create("parent")
				parent:with_context({env = "production"})
				
				local child = parent:child({module = "auth"})
				child:info("Child log")
				
				return type(child) == "table" and child ~= parent
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected child logger to be created")
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

func TestBasicLogging(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "log_levels",
			script: `
				local logger = logging.create("level_test")
				
				logger:debug("Debug message", {detail = "debug"})
				logger:info("Info message", {detail = "info"})
				logger:warn("Warning message", {detail = "warn"})
				logger:error("Error message", {detail = "error"})
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected all log levels to work")
				}
			},
		},
		{
			name: "conditional_debug",
			script: `
				local logger = logging.create("conditional_test")
				local logged = false
				
				logger:debug_if(true, "This should log")
				logger:debug_if(false, "This should not log")
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected conditional debug to work")
				}
			},
		},
		{
			name: "component_debug",
			script: `
				local logger = logging.create("component_test")
				
				logger:debug_component("auth", "Authentication debug")
				logger:debug_component("database", "Database query: %s", "SELECT * FROM users")
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected component debug to work")
				}
			},
		},
		{
			name: "error_with_stack",
			script: `
				local logger = logging.create("error_test")
				
				local function failing_function()
					error("Something went wrong")
				end
				
				local success, err = pcall(failing_function)
				if not success then
					logger:error_with_stack(err, "Function failed")
				end
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error with stack trace to work")
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

func TestDebugControl(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "enable_disable_debug",
			script: `
				-- Enable and disable global debug
				logging.debug.enable()
				logging.debug.disable()
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected debug enable/disable to work")
				}
			},
		},
		{
			name: "component_debug_control",
			script: `
				-- Enable specific components
				logging.debug.enable_component("auth")
				logging.debug.enable_component("database")
				
				local auth_enabled = logging.debug.is_enabled("auth")
				local db_enabled = logging.debug.is_enabled("database")
				local other_enabled = logging.debug.is_enabled("other")
				
				logging.debug.disable_component("auth")
				local auth_disabled = not logging.debug.is_enabled("auth")
				
				return auth_enabled and db_enabled and not other_enabled and auth_disabled
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected component debug control to work")
				}
			},
		},
		{
			name: "get_debug_components",
			script: `
				local components = logging.debug.get_components()
				return type(components) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected get_components to return table")
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

func TestFormatters(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "register_custom_formatter",
			script: `
				logging.formatters.register("custom", function(entry)
					return string.format("[%s] %s", entry.level, entry.message)
				end)
				
				local logger = logging.create("formatter_test")
				logger:set_formatter("custom")
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected custom formatter registration to work")
				}
			},
		},
		{
			name: "built_in_formatters",
			script: `
				-- Test that built-in formatters exist
				local logger1 = logging.create("json_logger")
				logger1:set_formatter("json")
				
				local logger2 = logging.create("text_logger")
				logger2:set_formatter("text")
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected built-in formatters to work")
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

func TestPerformanceAndProfiling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "timer_usage",
			script: `
				local timer = logging.timer("test_operation")
				
				-- Simulate some work
				local sum = 0
				for i = 1, 1000 do
					sum = sum + i
				end
				
				local duration = timer:stop()
				
				return type(duration) == "number" and duration >= 0
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected timer to return valid duration")
				}
			},
		},
		{
			name: "profile_function",
			script: `
				local result = logging.profile("compute_sum", function()
					local sum = 0
					for i = 1, 1000 do
						sum = sum + i
					end
					return sum
				end)
				
				return result == 500500  -- Sum of 1 to 1000
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected profile to return function result")
				}
			},
		},
		{
			name: "manual_timing",
			script: `
				local start = logging.time()
				
				-- Do some work
				local x = 0
				for i = 1, 100 do x = x + 1 end
				
				local duration = logging.duration("manual_task", start)
				
				return type(start) == "number" and type(duration) == "number"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected manual timing to work")
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

func TestHooks(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "register_hooks",
			script: `
				local hook_called = false
				
				logging.hooks.before_generate(function(params)
					hook_called = true
				end)
				
				logging.hooks.after_generate(function(params, result)
					-- Hook registered
				end)
				
				logging.hooks.before_tool_call(function(tool, args)
					-- Hook registered
				end)
				
				logging.hooks.after_tool_call(function(tool, args, result)
					-- Hook registered
				end)
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected hook registration to work")
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

func TestLoggingErrorHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "catch_exceptions",
			script: `
				local result, err = logging.catch(function()
					return "success"
				end, "test_operation")
				
				local result2, err2 = logging.catch(function()
					error("test error")
				end, "failing_operation")
				
				return result == "success" and err == nil and 
				       result2 == nil and err2 ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected catch to handle both success and error cases")
				}
			},
		},
		{
			name: "assert_with_logging",
			script: `
				logging.assert(true, "This should pass")
				
				local success, err = pcall(function()
					logging.assert(false, "This should fail", {detail = "test"})
				end)
				
				return success == false and string.find(err, "This should fail") ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected assert to work correctly")
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

func TestMetrics(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "metric_counter",
			script: `
				logging.metrics.count("user_login")
				logging.metrics.count("api_request", 1, {endpoint = "/users"})
				logging.metrics.count("errors", 3, {type = "validation"})
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected metric counter to work")
				}
			},
		},
		{
			name: "metric_gauge",
			script: `
				logging.metrics.gauge("queue_size", 42)
				logging.metrics.gauge("memory_usage", 1024.5, {unit = "MB"})
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected metric gauge to work")
				}
			},
		},
		{
			name: "metric_histogram",
			script: `
				logging.metrics.histogram("response_time", 123.45)
				logging.metrics.histogram("file_size", 2048, {type = "upload"})
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected metric histogram to work")
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

func TestLoggingAudit(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "audit_log",
			script: `
				logging.audit.log("user_login", {
					user_id = "123",
					ip_address = "192.168.1.1",
					success = true
				})
				
				logging.audit.log("permission_change", {
					admin = "admin@example.com",
					target_user = "user@example.com",
					old_role = "viewer",
					new_role = "editor"
				})
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected audit logging to work")
				}
			},
		},
		{
			name: "compliance_logging",
			script: `
				logging.audit.compliance("data_export", {
					user = "admin",
					records_count = 1000,
					destination = "external_system",
					timestamp = os.time()
				})
				
				logging.audit.compliance("gdpr_request", {
					request_type = "data_deletion",
					user_id = "user_123",
					processed_by = "system"
				})
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected compliance logging to work")
				}
			},
		},
		{
			name: "audit_handler_registration",
			script: `
				local handler_called = false
				
				logging.audit.register_handler(function(event_type, details)
					handler_called = true
				end)
				
				logging.audit.log("test_event", {test = true})
				
				return handler_called
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected audit handler to be called")
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

func TestIntegrationHelpers(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "search_logs",
			script: `
				local results = logging.search({
					level = {"error", "warn"},
					since = os.time() - 3600
				})
				
				return type(results) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected search to return table")
				}
			},
		},
		{
			name: "log_stats",
			script: `
				local stats = logging.stats({
					period = "1h",
					group_by = "level"
				})
				
				return type(stats) == "table" and
				       type(stats.by_level) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected stats to return proper structure")
				}
			},
		},
		{
			name: "logger_from_bridge",
			script: `
				local bridge_logger = logging.from_bridge(_G.util_script_logger)
				return type(bridge_logger) == "table" and
				       type(bridge_logger.info) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected from_bridge to create logger")
				}
			},
		},
		{
			name: "export_logs",
			script: `
				logging.export({
					format = "json",
					file = "/tmp/test_logs.json",
					filter = {level = "error"}
				})
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected export to work")
				}
			},
		},
		{
			name: "stream_logs",
			script: `
				logging.stream(function(entry)
					-- Process log entry
				end)
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected stream registration to work")
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

func TestSystemInfoAndCleanup(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "system_info",
			script: `
				local info = logging.get_system_info()
				
				return type(info) == "table" and
				       type(info.lua_version) == "string" and
				       type(info.bridges_available) == "table" and
				       info.bridges_available.debug == true and
				       info.bridges_available.slog == true and
				       info.bridges_available.script_logger == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected system info to have correct structure")
				}
			},
		},
		{
			name: "cleanup",
			script: `
				-- Create some loggers
				logging.create("logger1")
				logging.create("logger2")
				
				-- Register some hooks
				logging.hooks.before_generate(function() end)
				logging.audit.register_handler(function() end)
				
				-- Cleanup
				logging.cleanup()
				
				return true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected cleanup to work")
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

func TestLoggingValidation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "missing_required_parameters",
			script: `
				local success, err = pcall(function()
					local logger = logging.create()
					logger:log()  -- Missing level and message
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
					logging.configure("not_a_table")
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
			name: "invalid_log_level",
			script: `
				local logger = logging.create("test")
				local success, err = pcall(function()
					logger:log("invalid_level", "message")
				end)
				return success == false and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected invalid log level validation")
				}
			},
		},
		{
			name: "missing_bridge_graceful",
			script: `
				-- Clear any cached loggers first
				logging.cleanup()
				
				-- Temporarily remove bridge
				local original_bridge = _G.util_script_logger
				_G.util_script_logger = nil
				
				local success, err = pcall(function()
					-- Use a unique logger name to avoid cache
					local logger = logging.create("test_missing_bridge_" .. os.time())
				end)
				
				-- Restore bridge
				_G.util_script_logger = original_bridge
				
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

func TestLoggingIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	script := `
		-- Test comprehensive logging workflow
		local all_tests_passed = true
		
		-- Test 1: Configure logging
		logging.configure({
			level = "info",
			format = "json",
			emoji = false
		})
		
		-- Test 2: Create logger with context
		local app_logger = logging.create("myapp")
		app_logger:with_context({
			version = "1.0.0",
			environment = "test"
		})
		
		-- Test 3: Basic logging
		app_logger:info("Application started")
		app_logger:debug("Debug info")  -- Should not log if level is info
		
		-- Test 4: Child logger
		local auth_logger = app_logger:child({module = "auth"})
		auth_logger:info("Auth module initialized")
		
		-- Test 5: Performance monitoring
		local timer = logging.timer("initialization")
		-- Simulate work
		local sum = 0
		for i = 1, 1000 do sum = sum + i end
		timer:stop()
		
		-- Test 6: Error handling
		local result, err = logging.catch(function()
			return "success"
		end, "safe_operation")
		if result ~= "success" then
			all_tests_passed = false
		end
		
		-- Test 7: Metrics
		logging.metrics.count("app_start", 1)
		logging.metrics.gauge("memory_usage", 512)
		
		-- Test 8: Audit logging
		logging.audit.log("app_started", {
			user = "system",
			timestamp = os.time()
		})
		
		-- Test 9: Debug control
		logging.debug.enable_component("auth")
		if not logging.debug.is_enabled("auth") then
			all_tests_passed = false
		end
		
		-- Test 10: System info
		local info = logging.get_system_info()
		if type(info) ~= "table" then
			all_tests_passed = false
		end
		
		return all_tests_passed
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected comprehensive logging integration to work")
	}
}

func TestLoggingConcurrency(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	// Test logging from coroutines
	script := `
		local function concurrent_logging()
			local logger = logging.create("concurrent")
			
			local co1 = coroutine.create(function()
				for i = 1, 5 do
					logger:info("Coroutine 1", {iteration = i})
					coroutine.yield()
				end
			end)
			
			local co2 = coroutine.create(function()
				for i = 1, 5 do
					logger:info("Coroutine 2", {iteration = i})
					coroutine.yield()
				end
			end)
			
			-- Interleave execution
			for i = 1, 5 do
				coroutine.resume(co1)
				coroutine.resume(co2)
			end
			
			return true
		end
		
		return concurrent_logging()
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Concurrency test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected concurrent logging to work")
	}
}

func TestLoggingPerformance(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	// Test logging performance
	script := `
		local logger = logging.create("perf_test")
		
		-- Configure for minimal overhead
		logging.configure({
			level = "warn",  -- Skip debug/info
			format = "text"
		})
		
		-- Time bulk logging
		local start = os.clock()
		
		for i = 1, 1000 do
			-- These should be skipped due to level
			logger:debug("Debug message", {index = i})
			logger:info("Info message", {index = i})
		end
		
		local duration = os.clock() - start
		
		-- Should be very fast since messages are filtered
		return duration < 1.0  -- Less than 1 second for 2000 skipped messages
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Performance test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected logging to have good performance")
	}
}

func TestLoggingRealWorldExample(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupLoggingLibrary(t, L)

	// Real-world usage example
	script := `
		-- Configure application logging
		logging.configure({
			level = "info",
			format = "json",
			components = {"api", "database"}
		})
		
		-- Create main application logger
		local log = logging.create("webapp")
		log:with_context({
			service = "user-api",
			version = "2.1.0"
		})
		
		-- API request handler simulation
		local function handle_request(request)
			local request_logger = log:child({
				request_id = request.id,
				method = request.method,
				path = request.path
			})
			
			request_logger:info("Request started")
			
			-- Time the request
			local timer = logging.timer("request_" .. request.id)
			
			-- Simulate authentication
			logging.audit.log("api_access", {
				user = request.user,
				endpoint = request.path,
				ip = request.ip
			})
			
			-- Simulate database query with debug logging
			request_logger:debug_component("database", "Executing query: SELECT * FROM users WHERE id = ?", request.user_id)
			
			-- Simulate potential error
			if request.path == "/error" then
				local success, result = logging.catch(function()
					error("Simulated error")
				end, "request_handler")
				
				if not success then
					request_logger:error("Request failed", {
						error = tostring(result)
					})
					return false
				end
			end
			
			-- Stop timer and log metrics
			local duration = timer:stop()
			logging.metrics.histogram("api_request_duration", duration, {
				endpoint = request.path,
				method = request.method
			})
			
			request_logger:info("Request completed", {
				status = 200,
				duration_ms = duration
			})
			
			return true
		end
		
		-- Simulate requests
		local requests = {
			{id = "req_001", method = "GET", path = "/users", user = "alice", user_id = 123, ip = "10.0.0.1"},
			{id = "req_002", method = "POST", path = "/users", user = "bob", user_id = 456, ip = "10.0.0.2"},
			{id = "req_003", method = "GET", path = "/error", user = "charlie", user_id = 789, ip = "10.0.0.3"}
		}
		
		local all_handled = true
		for _, request in ipairs(requests) do
			local success = handle_request(request)
			if request.path ~= "/error" and not success then
				all_handled = false
			end
		end
		
		-- Get system info at end
		local info = logging.get_system_info()
		
		return all_handled and type(info) == "table"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Real-world example test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected real-world logging example to work")
	}
}

// setupBenchmarkBridges sets up minimal bridges for benchmarking
func setupBenchmarkBridges(L *lua.LState) {
	// Minimal debug bridge
	debugTable := L.NewTable()
	debugTable.RawSetString("enableDebugComponent", L.NewFunction(func(L *lua.LState) int { return 0 }))
	debugTable.RawSetString("disableDebugComponent", L.NewFunction(func(L *lua.LState) int { return 0 }))
	debugTable.RawSetString("isDebugEnabled", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LFalse)
		return 1
	}))
	debugTable.RawSetString("debugPrintf", L.NewFunction(func(L *lua.LState) int { return 0 }))
	L.SetGlobal("util_debug", debugTable)

	// Minimal slog bridge
	slogTable := L.NewTable()
	slogTable.RawSetString("info", L.NewFunction(func(L *lua.LState) int { return 0 }))
	slogTable.RawSetString("warn", L.NewFunction(func(L *lua.LState) int { return 0 }))
	slogTable.RawSetString("error", L.NewFunction(func(L *lua.LState) int { return 0 }))
	slogTable.RawSetString("debug", L.NewFunction(func(L *lua.LState) int { return 0 }))
	L.SetGlobal("util_slog", slogTable)

	// Minimal script logger bridge
	scriptLoggerTable := L.NewTable()
	scriptLoggerTable.RawSetString("log", L.NewFunction(func(L *lua.LState) int { return 0 }))
	scriptLoggerTable.RawSetString("configure", L.NewFunction(func(L *lua.LState) int { return 0 }))
	scriptLoggerTable.RawSetString("debug", L.NewFunction(func(L *lua.LState) int { return 0 }))
	scriptLoggerTable.RawSetString("info", L.NewFunction(func(L *lua.LState) int { return 0 }))
	scriptLoggerTable.RawSetString("warn", L.NewFunction(func(L *lua.LState) int { return 0 }))
	scriptLoggerTable.RawSetString("error", L.NewFunction(func(L *lua.LState) int { return 0 }))
	L.SetGlobal("util_script_logger", scriptLoggerTable)

	// Load the logging library
	libPath := filepath.Join(".", "logging.lua")
	if err := L.DoFile(libPath); err != nil {
		panic(fmt.Sprintf("Failed to load logging library: %v", err))
	}
	logging := L.Get(-1)
	L.SetGlobal("logging", logging)
}

// BenchmarkLogging benchmarks logging performance
func BenchmarkLogging(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	// Can't use setupLoggingLibrary here as it needs *testing.T
	// Set up bridges manually for benchmark
	setupBenchmarkBridges(L)

	// Setup
	if err := L.DoString(`
		local logger = logging.create("bench")
		logging.configure({level = "info"})
		_G.bench_logger = logger
	`); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := L.DoString(`
			bench_logger:info("Benchmark message", {
				iteration = ` + fmt.Sprintf("%d", i) + `,
				timestamp = os.time()
			})
		`); err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
