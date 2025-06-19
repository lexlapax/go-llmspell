// ABOUTME: Tests for Utility bridge adapter that exposes go-llms utility functionality to Lua scripts
// ABOUTME: Validates auth, debug, errors, json, llm utils, logging, and general utility functionality

package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestUtilsAdapter_Creation(t *testing.T) {
	t.Run("create_utils_adapter", func(t *testing.T) {
		// Create utility bridges mock
		authBridge := testutils.NewMockBridge("auth").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Auth Bridge",
				Version:     "1.0.0",
				Description: "Authentication utilities",
			})

		debugBridge := testutils.NewMockBridge("debug").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Debug Bridge",
				Version:     "1.0.0",
				Description: "Debug logging utilities",
			})

		errorsBridge := testutils.NewMockBridge("errors").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Errors Bridge",
				Version:     "1.0.0",
				Description: "Error handling utilities",
			})

		jsonBridge := testutils.NewMockBridge("json").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "JSON Bridge",
				Version:     "1.0.0",
				Description: "JSON processing utilities",
			})

		llmBridge := testutils.NewMockBridge("llm").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "LLM Utils Bridge",
				Version:     "1.0.0",
				Description: "LLM utility functions",
			})

		loggerBridge := testutils.NewMockBridge("script_logger").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Script Logger Bridge",
				Version:     "1.0.0",
				Description: "Unified logging interface",
			})

		slogBridge := testutils.NewMockBridge("slog").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Slog Bridge",
				Version:     "1.0.0",
				Description: "Structured logging utilities",
			})

		utilBridge := testutils.NewMockBridge("util").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "General Utils Bridge",
				Version:     "1.0.0",
				Description: "General utility functions",
			})

		// Create adapter
		adapter := NewUtilsAdapter(authBridge, debugBridge, errorsBridge, jsonBridge, llmBridge, loggerBridge, slogBridge, utilBridge)
		require.NotNil(t, adapter)

		// Should have utility-specific methods
		methods := adapter.GetMethods()

		// Auth methods
		assert.Contains(t, methods, "authenticate")
		assert.Contains(t, methods, "validateToken")
		assert.Contains(t, methods, "refreshToken")

		// Debug methods
		assert.Contains(t, methods, "setDebugLevel")
		assert.Contains(t, methods, "debugLog")
		assert.Contains(t, methods, "getDebugConfig")

		// Error methods
		assert.Contains(t, methods, "createError")
		assert.Contains(t, methods, "wrapError")
		assert.Contains(t, methods, "aggregateErrors")

		// JSON methods
		assert.Contains(t, methods, "parseJSON")
		assert.Contains(t, methods, "toJSON")
		assert.Contains(t, methods, "validateJSONSchema")

		// General utility methods
		assert.Contains(t, methods, "generateUUID")
		assert.Contains(t, methods, "hash")
		assert.Contains(t, methods, "sleep")
	})

	t.Run("utils_module_structure", func(t *testing.T) {
		authBridge := testutils.NewMockBridge("auth").WithInitialized(true)
		adapter := NewUtilsAdapter(authBridge, nil, nil, nil, nil, nil, nil, nil)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Create module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module table
		module := L.Get(-1)
		L.SetGlobal("utils", module)

		// Test module structure
		err = L.DoString(`
			-- Check basic module properties
			assert(utils._adapter == "utils", "should have correct adapter name")
			assert(utils._version == "1.0.0", "should have correct version")
			
			-- Check namespaces exist
			assert(type(utils.auth) == "table", "auth namespace should exist")
			assert(type(utils.debug) == "table", "debug namespace should exist")
			assert(type(utils.errors) == "table", "errors namespace should exist")
			assert(type(utils.json) == "table", "json namespace should exist")
			assert(type(utils.logger) == "table", "logger namespace should exist")
			assert(type(utils.general) == "table", "general namespace should exist")
		`)
		assert.NoError(t, err)
	})
}

func TestUtilsAdapter_Auth(t *testing.T) {
	t.Run("authenticate", func(t *testing.T) {
		authBridge := testutils.NewMockBridge("auth").
			WithInitialized(true).
			WithMethod("authenticate", engine.MethodInfo{
				Name: "authenticate",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				scheme := args[1].(engine.StringValue).Value()

				// Mock authentication
				if scheme == "oauth2" {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"success":      engine.NewBoolValue(true),
						"accessToken":  engine.NewStringValue("access-token-123"),
						"refreshToken": engine.NewStringValue("refresh-token-456"),
						"expiresIn":    engine.NewNumberValue(3600),
					}), nil
				}
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"success": engine.NewBoolValue(false),
					"error":   engine.NewStringValue("unsupported scheme"),
				}), nil
			})

		adapter := NewUtilsAdapter(authBridge, nil, nil, nil, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Authenticate with OAuth2
			local result, err = utils.auth.authenticate({
				clientId = "test-client",
				clientSecret = "test-secret"
			}, "oauth2")
			assert(err == nil, "should not error")
			assert(result.success == true, "should authenticate successfully")
			assert(result.accessToken == "access-token-123", "should have access token")
			assert(result.expiresIn == 3600, "should have expiry time")
		`)
		assert.NoError(t, err)
	})

	t.Run("validate_token", func(t *testing.T) {
		authBridge := testutils.NewMockBridge("auth").
			WithInitialized(true).
			WithMethod("validateToken", engine.MethodInfo{
				Name: "validateToken",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				token := args[0].(engine.StringValue).Value()

				// Mock token validation
				if token == "valid-token" {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"valid":     engine.NewBoolValue(true),
						"userId":    engine.NewStringValue("user-123"),
						"expiresAt": engine.NewStringValue("2024-12-31T23:59:59Z"),
					}), nil
				}
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"valid": engine.NewBoolValue(false),
					"error": engine.NewStringValue("invalid token"),
				}), nil
			})

		adapter := NewUtilsAdapter(authBridge, nil, nil, nil, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Validate valid token
			local result, err = utils.auth.validateToken("valid-token", {})
			assert(err == nil, "should not error")
			assert(result.valid == true, "should be valid")
			assert(result.userId == "user-123", "should have user ID")
			
			-- Validate invalid token
			local invalid, err2 = utils.auth.validateToken("invalid-token", {})
			assert(err2 == nil, "should not error")
			assert(invalid.valid == false, "should be invalid")
		`)
		assert.NoError(t, err)
	})
}

func TestUtilsAdapter_Debug(t *testing.T) {
	t.Run("set_debug_level", func(t *testing.T) {
		debugBridge := testutils.NewMockBridge("debug").
			WithInitialized(true).
			WithMethod("setDebugLevel", engine.MethodInfo{
				Name: "setDebugLevel",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				component := args[0].(engine.StringValue).Value()
				level := args[1].(engine.StringValue).Value()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"component": engine.NewStringValue(component),
					"level":     engine.NewStringValue(level),
					"set":       engine.NewBoolValue(true),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, debugBridge, nil, nil, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Set debug level
			local result, err = utils.debug.setLevel("engine", "DEBUG")
			assert(err == nil, "should not error")
			assert(result.set == true, "should be set")
			assert(result.component == "engine", "should have component")
			assert(result.level == "DEBUG", "should have level")
		`)
		assert.NoError(t, err)
	})

	t.Run("debug_log", func(t *testing.T) {
		debugBridge := testutils.NewMockBridge("debug").
			WithInitialized(true).
			WithMethod("debugLog", engine.MethodInfo{
				Name: "debugLog",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				component := args[0].(engine.StringValue).Value()
				message := args[1].(engine.StringValue).Value()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"logged":    engine.NewBoolValue(true),
					"component": engine.NewStringValue(component),
					"message":   engine.NewStringValue(message),
					"timestamp": engine.NewStringValue("2024-01-01T00:00:00Z"),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, debugBridge, nil, nil, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Log debug message
			local result, err = utils.debug.log("bridge", "test debug message", {})
			assert(err == nil, "should not error")
			assert(result.logged == true, "should be logged")
			assert(result.component == "bridge", "should have component")
			assert(result.message == "test debug message", "should have message")
		`)
		assert.NoError(t, err)
	})
}

func TestUtilsAdapter_Errors(t *testing.T) {
	t.Run("create_error", func(t *testing.T) {
		errorsBridge := testutils.NewMockBridge("errors").
			WithInitialized(true).
			WithMethod("createError", engine.MethodInfo{
				Name: "createError",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				message := args[0].(engine.StringValue).Value()
				code := args[1].(engine.StringValue).Value()
				category := args[2].(engine.StringValue).Value()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"message":  engine.NewStringValue(message),
					"code":     engine.NewStringValue(code),
					"category": engine.NewStringValue(category),
					"created":  engine.NewBoolValue(true),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, nil, errorsBridge, nil, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Create error
			local result, err = utils.errors.createError("test error", "E001", "validation")
			assert(err == nil, "should not error")
			assert(result.created == true, "should be created")
			assert(result.message == "test error", "should have message")
			assert(result.code == "E001", "should have code")
			assert(result.category == "validation", "should have category")
		`)
		assert.NoError(t, err)
	})

	t.Run("wrap_error", func(t *testing.T) {
		errorsBridge := testutils.NewMockBridge("errors").
			WithInitialized(true).
			WithMethod("wrapError", engine.MethodInfo{
				Name: "wrapError",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				originalError := args[0].(engine.ObjectValue).Fields()
				context := args[1].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"wrapped":  engine.NewBoolValue(true),
					"original": engine.NewObjectValue(originalError),
					"context":  engine.NewObjectValue(context),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, nil, errorsBridge, nil, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Wrap error
			local result, err = utils.errors.wrapError({
				message = "original error"
			}, {
				operation = "test",
				component = "adapter"
			})
			assert(err == nil, "should not error")
			assert(result.wrapped == true, "should be wrapped")
			assert(result.original.message == "original error", "should have original error")
			assert(result.context.operation == "test", "should have context")
		`)
		assert.NoError(t, err)
	})
}

func TestUtilsAdapter_JSON(t *testing.T) {
	t.Run("parse_json", func(t *testing.T) {
		jsonBridge := testutils.NewMockBridge("json").
			WithInitialized(true).
			WithMethod("parseJSON", engine.MethodInfo{
				Name: "parseJSON",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				text := args[0].(engine.StringValue).Value()

				if text == `{"name":"test","value":42}` {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"parsed": engine.NewBoolValue(true),
						"data": engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":  engine.NewStringValue("test"),
							"value": engine.NewNumberValue(42),
						}),
					}), nil
				}
				return nil, fmt.Errorf("invalid JSON")
			})

		adapter := NewUtilsAdapter(nil, nil, nil, jsonBridge, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Parse JSON
			local result, err = utils.json.parse('{"name":"test","value":42}', {})
			assert(err == nil, "should not error")
			assert(result.parsed == true, "should be parsed")
			assert(result.data.name == "test", "should have name")
			assert(result.data.value == 42, "should have value")
		`)
		assert.NoError(t, err)
	})

	t.Run("to_json", func(t *testing.T) {
		jsonBridge := testutils.NewMockBridge("json").
			WithInitialized(true).
			WithMethod("toJSON", engine.MethodInfo{
				Name: "toJSON",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"serialized": engine.NewBoolValue(true),
					"json":       engine.NewStringValue(`{"name":"test","value":42}`),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, nil, nil, jsonBridge, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Serialize to JSON
			local result, err = utils.json.toJSON({
				name = "test",
				value = 42
			}, {})
			assert(err == nil, "should not error")
			assert(result.serialized == true, "should be serialized")
			assert(type(result.json) == "string", "should have JSON string")
		`)
		assert.NoError(t, err)
	})
}

func TestUtilsAdapter_General(t *testing.T) {
	t.Run("generate_uuid", func(t *testing.T) {
		utilBridge := testutils.NewMockBridge("util").
			WithInitialized(true).
			WithMethod("generateUUID", engine.MethodInfo{
				Name: "generateUUID",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"uuid": engine.NewStringValue("123e4567-e89b-12d3-a456-426614174000"),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, nil, nil, nil, nil, nil, nil, utilBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Generate UUID
			local result, err = utils.general.generateUUID()
			assert(err == nil, "should not error")
			assert(type(result.uuid) == "string", "should have UUID string")
			assert(string.len(result.uuid) == 36, "should be valid UUID length")
		`)
		assert.NoError(t, err)
	})

	t.Run("hash", func(t *testing.T) {
		utilBridge := testutils.NewMockBridge("util").
			WithInitialized(true).
			WithMethod("hash", engine.MethodInfo{
				Name: "hash",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				data := args[0].(engine.StringValue).Value()
				algorithm := args[1].(engine.StringValue).Value()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"hash":      engine.NewStringValue("abcdef123456789"),
					"algorithm": engine.NewStringValue(algorithm),
					"data":      engine.NewStringValue(data),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, nil, nil, nil, nil, nil, nil, utilBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Hash data
			local result, err = utils.general.hash("test data", "sha256")
			assert(err == nil, "should not error")
			assert(result.hash == "abcdef123456789", "should have hash")
			assert(result.algorithm == "sha256", "should have algorithm")
			assert(result.data == "test data", "should have original data")
		`)
		assert.NoError(t, err)
	})

	t.Run("sleep", func(t *testing.T) {
		utilBridge := testutils.NewMockBridge("util").
			WithInitialized(true).
			WithMethod("sleep", engine.MethodInfo{
				Name: "sleep",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				duration := args[0].(engine.NumberValue).Value()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"slept":    engine.NewBoolValue(true),
					"duration": engine.NewNumberValue(duration),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, nil, nil, nil, nil, nil, nil, utilBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Sleep for duration
			local result, err = utils.general.sleep(100)
			assert(err == nil, "should not error")
			assert(result.slept == true, "should have slept")
			assert(result.duration == 100, "should have correct duration")
		`)
		assert.NoError(t, err)
	})
}

func TestUtilsAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		authBridge := testutils.NewMockBridge("auth").
			WithInitialized(true).
			WithMethod("authenticate", engine.MethodInfo{
				Name: "authenticate",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("auth service unavailable")
			})

		adapter := NewUtilsAdapter(authBridge, nil, nil, nil, nil, nil, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Try authentication with error
			local result, err = utils.auth.authenticate({}, "oauth2")
			assert(err ~= nil, "should have error")
			assert(string.find(err, "auth service unavailable"), "error should contain message")
			assert(result == nil, "result should be nil on error")
		`)
		assert.NoError(t, err)
	})
}

func TestUtilsAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("retry_operation", func(t *testing.T) {
		utilBridge := testutils.NewMockBridge("util").
			WithInitialized(true).
			WithMethod("retry", engine.MethodInfo{
				Name: "retry",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				operation := args[0].(engine.ObjectValue).Fields()
				options := args[1].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"success":  engine.NewBoolValue(true),
					"attempts": engine.NewNumberValue(2),
					"result":   engine.NewObjectValue(operation),
					"options":  engine.NewObjectValue(options),
				}), nil
			})

		adapter := NewUtilsAdapter(nil, nil, nil, nil, nil, nil, nil, utilBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "utils")
		require.NoError(t, err)

		err = ms.LoadModule(L, "utils")
		require.NoError(t, err)

		err = L.DoString(`
			local utils = require("utils")
			
			-- Retry operation
			local result, err = utils.general.retry({
				action = "test"
			}, {
				maxAttempts = 3,
				backoff = "exponential"
			})
			assert(err == nil, "should not error")
			assert(result.success == true, "should succeed")
			assert(result.attempts == 2, "should have attempt count")
		`)
		assert.NoError(t, err)
	})
}
