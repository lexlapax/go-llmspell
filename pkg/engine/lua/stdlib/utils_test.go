// ABOUTME: Tests for the utils Lua module with multi-bridge adapter pattern
// ABOUTME: Validates utility functions through 8 specialized bridges

package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

// setupUtilsBridges sets up all 8 mock bridges for utils module
func setupUtilsBridges(t *testing.T, L *lua.LState) {
	t.Helper()

	// Create bridges table
	bridges := L.NewTable()
	L.SetGlobal("bridges", bridges)

	// Create auth bridge
	authBridge := L.NewTable()
	authBridge.RawSetString("authenticate", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	authBridge.RawSetString("validateToken", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	authBridge.RawSetString("refreshToken", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("new-token"))
		return 1
	}))
	authBridge.RawSetString("generateToken", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("generated-token"))
		return 1
	}))
	authBridge.RawSetString("hashPassword", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("hashed-password"))
		return 1
	}))
	authBridge.RawSetString("verifyPassword", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	bridges.RawSetString("util_auth", authBridge)

	// Create debug bridge
	debugBridge := L.NewTable()
	debugBridge.RawSetString("setLevel", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	debugBridge.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	debugBridge.RawSetString("getConfig", L.NewFunction(func(L *lua.LState) int {
		config := L.NewTable()
		config.RawSetString("level", lua.LString("debug"))
		L.Push(config)
		return 1
	}))
	debugBridge.RawSetString("trace", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("trace-id"))
		return 1
	}))
	debugBridge.RawSetString("profile", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("duration", lua.LNumber(0.123))
		L.Push(result)
		return 1
	}))
	debugBridge.RawSetString("dump", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("dumped"))
		return 1
	}))
	debugBridge.RawSetString("assert", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	bridges.RawSetString("util_debug", debugBridge)

	// Create errors bridge
	errorsBridge := L.NewTable()
	errorsBridge.RawSetString("createError", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		category := L.CheckString(2)
		result := L.NewTable()
		result.RawSetString("message", lua.LString(message))
		result.RawSetString("category", lua.LString(category))
		L.Push(result)
		return 1
	}))
	errorsBridge.RawSetString("wrapError", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("wrapped", lua.LTrue)
		L.Push(result)
		return 1
	}))
	errorsBridge.RawSetString("aggregateErrors", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("count", lua.LNumber(3))
		L.Push(result)
		return 1
	}))
	errorsBridge.RawSetString("categorizeError", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("validation"))
		return 1
	}))
	errorsBridge.RawSetString("wrap", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("wrapped", lua.LTrue)
		L.Push(result)
		return 1
	}))
	errorsBridge.RawSetString("unwrap", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNil)
		return 1
	}))
	errorsBridge.RawSetString("isType", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	errorsBridge.RawSetString("getStack", L.NewFunction(func(L *lua.LState) int {
		stack := L.NewTable()
		stack.Append(lua.LString("stack frame 1"))
		stack.Append(lua.LString("stack frame 2"))
		L.Push(stack)
		return 1
	}))
	bridges.RawSetString("util_errors", errorsBridge)

	// Create JSON bridge
	jsonBridge := L.NewTable()
	jsonBridge.RawSetString("parse", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("key", lua.LString("value"))
		L.Push(result)
		return 1
	}))
	jsonBridge.RawSetString("decode", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("key", lua.LString("value"))
		L.Push(result)
		return 1
	}))
	jsonBridge.RawSetString("toJSON", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(`{"key":"value"}`))
		return 1
	}))
	jsonBridge.RawSetString("encode", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(`{"key":"value"}`))
		return 1
	}))
	jsonBridge.RawSetString("validateJSONSchema", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	jsonBridge.RawSetString("validate", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	jsonBridge.RawSetString("extractStructuredData", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("extracted", lua.LTrue)
		L.Push(result)
		return 1
	}))
	jsonBridge.RawSetString("format", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("{\n  \"key\": \"value\"\n}"))
		return 1
	}))
	bridges.RawSetString("util_json", jsonBridge)

	// Create LLM bridge
	llmBridge := L.NewTable()
	llmBridge.RawSetString("createProvider", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("id", lua.LString("provider-123"))
		L.Push(result)
		return 1
	}))
	llmBridge.RawSetString("generateTyped", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("result", lua.LString("generated"))
		L.Push(result)
		return 1
	}))
	llmBridge.RawSetString("trackCost", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	llmBridge.RawSetString("parseResponse", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("parsed", lua.LTrue)
		L.Push(result)
		return 1
	}))
	llmBridge.RawSetString("formatPrompt", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("formatted prompt"))
		return 1
	}))
	llmBridge.RawSetString("countTokens", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(100))
		return 1
	}))
	llmBridge.RawSetString("splitMessage", L.NewFunction(func(L *lua.LState) int {
		parts := L.NewTable()
		parts.Append(lua.LString("part1"))
		parts.Append(lua.LString("part2"))
		L.Push(parts)
		return 1
	}))
	bridges.RawSetString("util_llm", llmBridge)

	// Create logger bridge
	loggerBridge := L.NewTable()
	loggerBridge.RawSetString("createLogger", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("id", lua.LString("logger-123"))
		L.Push(result)
		return 1
	}))
	loggerBridge.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	loggerBridge.RawSetString("error", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	loggerBridge.RawSetString("warn", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	loggerBridge.RawSetString("info", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	loggerBridge.RawSetString("debug", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	loggerBridge.RawSetString("setLogLevel", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))
	bridges.RawSetString("util_script_logger", loggerBridge)

	// Create slog bridge
	slogBridge := L.NewTable()
	slogBridge.RawSetString("info", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	slogBridge.RawSetString("warn", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	slogBridge.RawSetString("error", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	slogBridge.RawSetString("debug", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	slogBridge.RawSetString("withFields", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("logger", lua.LString("with-fields"))
		L.Push(result)
		return 1
	}))
	bridges.RawSetString("util_slog", slogBridge)

	// Create util bridge
	utilBridge := L.NewTable()
	utilBridge.RawSetString("generateUUID", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("123e4567-e89b-12d3-a456-426614174000"))
		return 1
	}))
	utilBridge.RawSetString("uuid", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("123e4567-e89b-12d3-a456-426614174000"))
		return 1
	}))
	utilBridge.RawSetString("hash", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("hashed-value"))
		return 1
	}))
	utilBridge.RawSetString("retry", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("attempts", lua.LNumber(1))
		result.RawSetString("success", lua.LTrue)
		L.Push(result)
		return 1
	}))
	utilBridge.RawSetString("sleep", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	utilBridge.RawSetString("encode", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("encoded-value"))
		return 1
	}))
	utilBridge.RawSetString("decode", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("decoded-value"))
		return 1
	}))
	bridges.RawSetString("util_core", utilBridge)
}

func TestUtilsModule(t *testing.T) {
	t.Run("module loads with bridges", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Setup bridges
		setupUtilsBridges(t, L)
		
		// Load the utils module
		LoadModule(t, L, "utils")
		
		err := L.DoString(`
			local utils = require("utils")
			assert(type(utils) == "table", "utils should be a table")
			
			-- Check auth methods exist
			assert(type(utils.auth_authenticate) == "function", "auth_authenticate should be a function")
			assert(type(utils.auth_validate_token) == "function", "auth_validate_token should be a function")
			
			-- Check debug methods exist
			assert(type(utils.debug_set_level) == "function", "debug_set_level should be a function")
			assert(type(utils.debug_log) == "function", "debug_log should be a function")
			
			-- Check error methods exist
			assert(type(utils.errors_create_error) == "function", "errors_create_error should be a function")
			assert(type(utils.errors_wrap_error) == "function", "errors_wrap_error should be a function")
			
			-- Check JSON methods exist
			assert(type(utils.json_parse) == "function", "json_parse should be a function")
			assert(type(utils.json_encode) == "function", "json_encode should be a function")
			
			-- Check constants exist
			assert(type(utils.LOG_LEVELS) == "table", "LOG_LEVELS should be a table")
			assert(utils.LOG_LEVELS.DEBUG == "debug", "LOG_LEVELS.DEBUG should be 'debug'")
			assert(type(utils.AUTH_SCHEMES) == "table", "AUTH_SCHEMES should be a table")
			assert(type(utils.HASH_ALGORITHMS) == "table", "HASH_ALGORITHMS should be a table")
			assert(type(utils.ERROR_CATEGORIES) == "table", "ERROR_CATEGORIES should be a table")
		`)
		assert.NoError(t, err)
	})

	t.Run("auth operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test authenticate
			local result = utils.auth_authenticate({username = "test", password = "pass"})
			assert(result == true, "authenticate should return true")
			
			-- Test validate token
			result = utils.auth_validate_token("test-token")
			assert(result == true, "validate_token should return true")
			
			-- Test refresh token
			result = utils.auth_refresh_token("old-token")
			assert(result == "new-token", "refresh_token should return new token")
			
			-- Test generate token
			result = utils.auth_generate_token({user_id = 123})
			assert(result == "generated-token", "generate_token should return token")
			
			-- Test hash password
			result = utils.auth_hash_password("password", utils.HASH_ALGORITHMS.BCRYPT)
			assert(result == "hashed-password", "hash_password should return hash")
			
			-- Test verify password
			result = utils.auth_verify_password("password", "hash")
			assert(result == true, "verify_password should return true")
		`)
		assert.NoError(t, err)
	})

	t.Run("debug operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test set level
			local result = utils.debug_set_level(utils.LOG_LEVELS.DEBUG)
			assert(result == true, "set_level should return true")
			
			-- Test get config
			result = utils.debug_get_config()
			assert(type(result) == "table", "get_config should return table")
			assert(result.level == "debug", "config level should be debug")
			
			-- Test trace
			result = utils.debug_trace("operation")
			assert(result == "trace-id", "trace should return trace id")
			
			-- Test dump
			result = utils.debug_dump({test = "value"}, "test_var")
			assert(result == "dumped", "dump should return dumped")
		`)
		assert.NoError(t, err)
	})

	t.Run("error operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test create error
			local err = utils.errors_create_error(
				"validation failed", 
				utils.ERROR_CATEGORIES.VALIDATION,
				{field = "email"}
			)
			assert(type(err) == "table", "create_error should return table")
			assert(err.message == "validation failed", "error message should match")
			assert(err.category == utils.ERROR_CATEGORIES.VALIDATION, "error category should match")
			
			-- Test wrap error
			local wrapped = utils.errors_wrap_error(err, "additional context")
			assert(type(wrapped) == "table", "wrap_error should return table")
			assert(wrapped.wrapped == true, "error should be wrapped")
			
			-- Test categorize error
			local category = utils.errors_categorize_error(err)
			assert(category == "validation", "categorize_error should return validation")
			
			-- Test is type
			local is_validation = utils.errors_is_type(err, utils.ERROR_CATEGORIES.VALIDATION)
			assert(is_validation == true, "is_type should return true")
			
			-- Test get stack
			local stack = utils.errors_get_stack(err)
			assert(type(stack) == "table", "get_stack should return table")
			assert(#stack == 2, "stack should have 2 frames")
		`)
		assert.NoError(t, err)
	})

	t.Run("json operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test parse
			local obj = utils.json_parse('{"key": "value"}')
			assert(type(obj) == "table", "parse should return table")
			assert(obj.key == "value", "parsed object should have correct value")
			
			-- Test encode
			local json = utils.json_encode({key = "value"})
			assert(type(json) == "string", "encode should return string")
			assert(json == '{"key":"value"}', "encoded JSON should match")
			
			-- Test validate
			local valid = utils.json_validate('{"key": "value"}')
			assert(valid == true, "validate should return true for valid JSON")
			
			-- Test format
			local formatted = utils.json_format('{"key":"value"}', 2)
			assert(type(formatted) == "string", "format should return string")
			assert(formatted:find("\n") ~= nil, "formatted JSON should have newlines")
		`)
		assert.NoError(t, err)
	})

	t.Run("llm operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test create provider
			local provider = utils.llm_create_provider("openai", {api_key = "test"})
			assert(type(provider) == "table", "create_provider should return table")
			assert(provider.id == "provider-123", "provider should have id")
			
			-- Test count tokens
			local count = utils.llm_count_tokens("test text", "gpt-4")
			assert(count == 100, "count_tokens should return 100")
			
			-- Test format prompt
			local prompt = utils.llm_format_prompt("Hello {name}", {name = "World"})
			assert(prompt == "formatted prompt", "format_prompt should return formatted")
			
			-- Test split message
			local parts = utils.llm_split_message("long text", 10, "gpt-4")
			assert(type(parts) == "table", "split_message should return table")
			assert(#parts == 2, "should have 2 parts")
		`)
		assert.NoError(t, err)
	})

	t.Run("logger operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test create logger
			local logger = utils.logger_create_logger("test", {level = "info"})
			assert(type(logger) == "table", "create_logger should return table")
			
			-- Test set log level
			local result = utils.logger_set_log_level(utils.LOG_LEVELS.WARN)
			assert(result == true, "set_log_level should return true")
			
			-- Test logging methods (should not error)
			utils.logger_error("error message", {code = 500})
			utils.logger_warn("warning message", {threshold = 0.8})
			utils.logger_info("info message", {user = "test"})
			utils.logger_debug("debug message", {verbose = true})
		`)
		assert.NoError(t, err)
	})

	t.Run("slog operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test structured logging methods (should not error)
			utils.slog_info("user logged in", {user_id = 123, ip = "1.2.3.4"})
			utils.slog_warn("high memory usage", {percent = 85})
			utils.slog_error("database connection failed", {host = "db.example.com"})
			utils.slog_debug("cache hit", {key = "user:123"})
			
			-- Test with fields
			local logger = utils.slog_with_fields({service = "api", version = "1.0"})
			assert(type(logger) == "table", "with_fields should return table")
		`)
		assert.NoError(t, err)
	})

	t.Run("general utilities", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test UUID generation
			local uuid1 = utils.general_generate_uuid()
			local uuid2 = utils.general_uuid()
			assert(type(uuid1) == "string", "generate_uuid should return string")
			assert(type(uuid2) == "string", "uuid should return string")
			assert(#uuid1 == 36, "UUID should be 36 characters")
			
			-- Test hash
			local hash = utils.general_hash("data", utils.HASH_ALGORITHMS.SHA256)
			assert(hash == "hashed-value", "hash should return hashed value")
			
			-- Test encode/decode
			local encoded = utils.general_encode("data", "base64")
			assert(encoded == "encoded-value", "encode should return encoded")
			
			local decoded = utils.general_decode("ZGF0YQ==", "base64")
			assert(decoded == "decoded-value", "decode should return decoded")
			
			-- Test sleep (should not error)
			utils.general_sleep(100)
		`)
		assert.NoError(t, err)
	})

	t.Run("missing bridges graceful failure", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Don't set up bridges - test error handling
		L.SetGlobal("bridges", L.NewTable())
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			local success, err = pcall(function()
				utils.auth_authenticate({})
			end)
			assert(not success, "should fail without auth bridge")
			assert(err:find("Auth bridge not available"), "should have correct error message")
		`)
		assert.NoError(t, err)
	})
}

func TestUtilsIntegration(t *testing.T) {
	t.Run("cross-bridge operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Simulate a workflow using multiple bridges
			
			-- 1. Authenticate user
			local auth_result = utils.auth_authenticate({
				username = "testuser",
				password = "testpass"
			})
			assert(auth_result == true, "authentication should succeed")
			
			-- 2. Generate token
			local token = utils.auth_generate_token({user_id = 123})
			assert(token == "generated-token", "should generate token")
			
			-- 3. Log the action
			utils.logger_info("User authenticated", {
				user = "testuser",
				token_issued = true
			})
			
			-- 4. Create structured log entry
			utils.slog_info("auth.success", {
				user_id = 123,
				timestamp = os.time()
			})
			
			-- 5. Track in debug
			local trace_id = utils.debug_trace("auth_flow")
			assert(trace_id == "trace-id", "should get trace id")
			
			-- 6. Format response as JSON
			local response = utils.json_encode({
				success = true,
				token = token,
				trace_id = trace_id
			})
			assert(type(response) == "string", "should encode response")
		`)
		assert.NoError(t, err)
	})

	t.Run("error handling workflow", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupUtilsBridges(t, L)
		LoadModule(t, L, "utils")

		err := L.DoString(`
			local utils = require("utils")
			
			-- Create and handle errors across bridges
			
			-- 1. Create an error
			local err = utils.errors_create_error(
				"Invalid input", 
				utils.ERROR_CATEGORIES.VALIDATION,
				{field = "email", value = "not-an-email"}
			)
			
			-- 2. Log the error
			utils.logger_error("Validation failed", {
				error = err,
				request_id = "req-123"
			})
			
			-- 3. Wrap with context
			local wrapped = utils.errors_wrap_error(err, "User registration failed")
			
			-- 4. Get stack trace
			local stack = utils.errors_get_stack(wrapped)
			assert(#stack > 0, "should have stack trace")
			
			-- 5. Debug dump the error
			utils.debug_dump(wrapped, "registration_error")
			
			-- 6. Return error response
			local error_response = utils.json_encode({
				error = true,
				message = "Registration failed",
				category = utils.errors_categorize_error(wrapped),
				trace_id = utils.debug_trace("error_flow")
			})
			assert(type(error_response) == "string", "should encode error response")
		`)
		assert.NoError(t, err)
	})
}