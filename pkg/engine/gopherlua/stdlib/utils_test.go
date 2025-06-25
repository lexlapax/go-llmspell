// ABOUTME: Tests for the utils Lua module
// ABOUTME: Validates utility functions for file, system, and time operations

package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

// setupTestEnvironment sets up the test environment with required modules
func setupTestEnvironment(t *testing.T, L *lua.LState) {
	t.Helper()
	
	// Load the tools module first (utils depends on it)
	LoadModule(t, L, "tools")
	
	// Load the utils module
	LoadModule(t, L, "utils")
}

func TestUtilsModule(t *testing.T) {
	t.Run("module loads successfully", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Load the module
		setupTestEnvironment(t, L)
		err := L.DoString(`
			local utils = require("utils")
			assert(type(utils) == "table", "utils should be a table")
			
			-- Check main functions exist
			assert(type(utils.file_exists) == "function", "file_exists should be a function")
			assert(type(utils.file_write) == "function", "file_write should be a function")
			assert(type(utils.file_read) == "function", "file_read should be a function")
			assert(type(utils.mkdir) == "function", "mkdir should be a function")
			assert(type(utils.list_files) == "function", "list_files should be a function")
			
			assert(type(utils.sleep) == "function", "sleep should be a function")
			assert(type(utils.env) == "function", "env should be a function")
			assert(type(utils.exec) == "function", "exec should be a function")
			assert(type(utils.system_info) == "function", "system_info should be a function")
			
			assert(type(utils.current_time) == "function", "current_time should be a function")
			assert(type(utils.format_time) == "function", "format_time should be a function")
		`)
		assert.NoError(t, err)
	})

	t.Run("file operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupTestEnvironment(t, L)

		// Test file_exists with empty path
		err := L.DoString(`
			local utils = require("utils")
			
			-- Test empty path
			assert(utils.file_exists("") == false, "empty path should return false")
			assert(utils.file_exists(nil) == false, "nil path should return false")
		`)
		assert.NoError(t, err)

		// Test file operations with mock tools
		setupMockTools(t, L)
		err = L.DoString(`
			local utils = require("utils")
			
			-- Test file_exists
			local exists = utils.file_exists("/test/file.txt")
			assert(type(exists) == "boolean", "file_exists should return boolean")
			
			-- Test file_write
			local result = utils.file_write("/test/output.txt", "Hello, World!")
			assert(type(result) == "table", "file_write should return table")
			
			-- Test file_read
			local content, err = utils.file_read("/test/input.txt")
			assert(content ~= nil or err ~= nil, "file_read should return content or error")
			
			-- Test mkdir
			result = utils.mkdir("/test/newdir")
			assert(type(result) == "table", "mkdir should return table")
			
			-- Test list_files
			local files, err = utils.list_files("/test")
			assert(files ~= nil or err ~= nil, "list_files should return files or error")
		`)
		assert.NoError(t, err)
	})

	t.Run("system operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupTestEnvironment(t, L)
		setupMockTools(t, L)

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test sleep with invalid input
			local result = utils.sleep(-1)
			assert(result.error ~= nil, "negative sleep should return error")
			
			result = utils.sleep("not a number")
			assert(result.error ~= nil, "non-number sleep should return error")
			
			-- Test env
			local value = utils.env("TEST_VAR")
			assert(value == nil or type(value) == "string", "env should return string or nil")
			
			-- Test exec with empty command
			result = utils.exec("")
			assert(result.error ~= nil, "empty command should return error")
			
			-- Test system_info
			local info = utils.system_info(false)
			assert(type(info) == "table", "system_info should return table")
		`)
		assert.NoError(t, err)
	})

	t.Run("time operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupTestEnvironment(t, L)

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test current_time
			local now = utils.current_time()
			assert(type(now) == "number", "current_time should return number")
			assert(now > 0, "current_time should be positive")
			
			-- Test format_time
			local formatted = utils.format_time(now)
			assert(type(formatted) == "string", "format_time should return string")
			assert(#formatted > 0, "formatted time should not be empty")
			
			-- Test with custom format
			formatted = utils.format_time(now, "%Y")
			assert(type(formatted) == "string", "format_time with custom format should return string")
		`)
		assert.NoError(t, err)
	})

	t.Run("utility functions", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupTestEnvironment(t, L)
		setupMockTools(t, L)

		err := L.DoString(`
			local utils = require("utils")
			
			-- Test is_windows
			local is_win = utils.is_windows()
			assert(type(is_win) == "boolean", "is_windows should return boolean")
			
			-- Test join_path
			local path = utils.join_path("dir", "subdir", "file.txt")
			assert(type(path) == "string", "join_path should return string")
			assert(#path > 0, "joined path should not be empty")
			
			-- Test with empty components
			path = utils.join_path("dir", "", "file.txt")
			assert(type(path) == "string", "join_path with empty component should work")
			
			-- Test file_extension
			assert(utils.file_extension("test.txt") == ".txt", "should extract .txt extension")
			assert(utils.file_extension("test.tar.gz") == ".gz", "should extract last extension")
			assert(utils.file_extension("no_extension") == "", "should return empty for no extension")
			assert(utils.file_extension(nil) == "", "should handle nil")
			
			-- Test basename
			assert(utils.basename("/path/to/file.txt") == "file.txt", "should extract basename")
			assert(utils.basename("file.txt") == "file.txt", "should handle no directory")
			assert(utils.basename(nil) == "", "should handle nil")
			
			-- Test dirname
			assert(utils.dirname("/path/to/file.txt") == "/path/to", "should extract directory")
			assert(utils.dirname("file.txt") == ".", "should return . for no directory")
			assert(utils.dirname(nil) == ".", "should handle nil")
		`)
		assert.NoError(t, err)
	})
}

// setupMockTools sets up mock implementations of tools for testing
func setupMockTools(t *testing.T, L *lua.LState) {
	// Create a mock tools module
	err := L.DoString(`
		-- Override tools.execute_safe for testing
		local original_tools = require("tools")
		local mock_tools = {}
		
		-- Copy all original functions
		for k, v in pairs(original_tools) do
			mock_tools[k] = v
		end
		
		-- Mock execute_safe
		function mock_tools.execute_safe(tool_name, params, opts)
			-- Return mock responses based on tool name
			if tool_name == "file_read" then
				if params.path == "/test/file.txt" then
					return {content = "test content"}
				else
					return {error = "file not found"}
				end
			elseif tool_name == "file_write" then
				return {success = true}
			elseif tool_name == "file_list" then
				return {files = {"file1.txt", "file2.txt"}}
			elseif tool_name == "system_execute" then
				return {stdout = "output", stderr = "", exit_code = 0, success = true}
			elseif tool_name == "system_env_var" then
				return {variables = {{name = params.name, value = "test_value"}}}
			elseif tool_name == "system_info" then
				return {os = {name = "test", platform = "Test OS"}, architecture = "test_arch"}
			else
				return {error = "unknown tool: " .. tostring(tool_name)}
			end
		end
		
		-- Replace the tools module
		package.loaded["tools"] = mock_tools
	`)
	require.NoError(t, err)
}

func TestUtilsPathOperations(t *testing.T) {
	// Test path operations in isolation
	L := lua.NewState()
	defer L.Close()

	setupTestEnvironment(t, L)

	testCases := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "join path unix style",
			script: `
				local utils = require("utils")
				-- Mock is_windows to return false
				utils.is_windows = function() return false end
				return utils.join_path("home", "user", "file.txt")
			`,
			expected: "home/user/file.txt",
		},
		{
			name: "file extension variations",
			script: `
				local utils = require("utils")
				local results = {}
				table.insert(results, utils.file_extension("file.txt"))
				table.insert(results, utils.file_extension("archive.tar.gz"))
				table.insert(results, utils.file_extension(".hidden"))
				table.insert(results, utils.file_extension("no-ext"))
				return table.concat(results, ",")
			`,
			expected: ".txt,.gz,,",
		},
		{
			name: "basename variations",
			script: `
				local utils = require("utils")
				local results = {}
				table.insert(results, utils.basename("/usr/local/bin/tool"))
				table.insert(results, utils.basename("C:\\Windows\\System32\\cmd.exe"))
				table.insert(results, utils.basename("file.txt"))
				return table.concat(results, ",")
			`,
			expected: "tool,cmd.exe,file.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := L.DoString(tc.script)
			require.NoError(t, err)
			
			result := L.Get(-1)
			L.Pop(1)
			
			assert.Equal(t, tc.expected, result.String())
		})
	}
}

func TestUtilsFileOperationsWithTempDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "utils_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	L := lua.NewState()
	defer L.Close()

	setupTestEnvironment(t, L)

	// Test with actual file operations (if tools are available)
	testFile := filepath.Join(tempDir, "test.txt")
	
	// Set the test file path as a global
	L.SetGlobal("TEST_FILE", lua.LString(testFile))
	L.SetGlobal("TEST_DIR", lua.LString(tempDir))

	err = L.DoString(`
		local utils = require("utils")
		
		-- These tests will only work if actual file tools are available
		-- Otherwise they'll use our mocked tools
		
		local test_file = TEST_FILE
		local test_dir = TEST_DIR
		
		-- Test file operations
		local exists = utils.file_exists(test_file)
		assert(type(exists) == "boolean", "file_exists should return boolean")
		
		-- The actual behavior depends on whether real tools are available
		-- So we just verify the API contract
	`)
	assert.NoError(t, err)
}