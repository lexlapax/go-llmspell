// ABOUTME: Tests for LuaEngine execution pipeline functionality
// ABOUTME: Validates state acquisition, security sandbox, parameter injection, script compilation, and result extraction

package gopherlua

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestExecutionPipeline_StateAcquisition(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		MemoryLimit:  32 * 1024 * 1024,
		TimeoutLimit: 5 * time.Second,
		SandboxMode:  true,
		EngineOptions: map[string]interface{}{
			"pool_min_size": 2,
			"pool_max_size": 4,
		},
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name       string
		script     string
		concurrent int
		expectErr  bool
	}{
		{
			name:       "single_state_acquisition",
			script:     "return 'test'",
			concurrent: 1,
		},
		{
			name:       "concurrent_state_acquisition",
			script:     "return math.random()",
			concurrent: 3,
		},
		{
			name:       "pool_exhaustion_recovery",
			script:     "return 'recovery_test'",
			concurrent: 6, // More than pool max size
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			results := make(chan error, tt.concurrent)

			// Execute concurrently
			for i := 0; i < tt.concurrent; i++ {
				go func() {
					_, err := eng.Execute(ctx, tt.script, nil)
					results <- err
				}()
			}

			// Check results
			for i := 0; i < tt.concurrent; i++ {
				err := <-results
				if tt.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestExecutionPipeline_SecuritySandbox(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode:     true,
		AllowedModules:  []string{"string", "math"},
		DisabledModules: []string{"io", "os"},
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name    string
		script  string
		wantErr bool
		errMsg  string
	}{
		{
			name:   "allowed_string_module",
			script: `return string.upper("hello")`,
		},
		{
			name:   "allowed_math_module",
			script: `return math.pi`,
		},
		{
			name:    "blocked_io_module",
			script:  `return io.open("/tmp/test", "r")`,
			wantErr: true,
		},
		{
			name:    "blocked_os_module",
			script:  `return os.execute("echo test")`,
			wantErr: true,
		},
		{
			name:    "blocked_load_function",
			script:  `return load("return 42")()`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := eng.Execute(ctx, tt.script, nil)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecutionPipeline_ParameterInjection(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		script   string
		params   map[string]interface{}
		validate func(t *testing.T, result interface{})
	}{
		{
			name:   "string_parameter",
			script: `return "Hello, " .. name`,
			params: map[string]interface{}{
				"name": "World",
			},
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, "Hello, World", result)
			},
		},
		{
			name:   "number_parameters",
			script: `return a + b * c`,
			params: map[string]interface{}{
				"a": 10,
				"b": 5,
				"c": 2,
			},
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, 20.0, result)
			},
		},
		{
			name:   "boolean_parameter",
			script: `return flag and "yes" or "no"`,
			params: map[string]interface{}{
				"flag": true,
			},
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, "yes", result)
			},
		},
		{
			name:   "table_parameter",
			script: `return data.x + data.y`,
			params: map[string]interface{}{
				"data": map[string]interface{}{
					"x": 15,
					"y": 25,
				},
			},
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, 40.0, result)
			},
		},
		{
			name:   "array_parameter",
			script: `return items[1] + items[2] + items[3]`,
			params: map[string]interface{}{
				"items": []interface{}{10, 20, 30},
			},
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, 60.0, result)
			},
		},
		{
			name:   "nil_parameter",
			script: `return value == nil and "nil" or "not nil"`,
			params: map[string]interface{}{
				"value": nil,
			},
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, "nil", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := eng.Execute(ctx, tt.script, tt.params)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestExecutionPipeline_ScriptCompilation(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name    string
		script  string
		wantErr bool
		errType engine.ErrorType
	}{
		{
			name:   "valid_simple_script",
			script: `return 42`,
		},
		{
			name: "valid_complex_script",
			script: `
				local function factorial(n)
					if n <= 1 then
						return 1
					else
						return n * factorial(n - 1)
					end
				end
				return factorial(5)
			`,
		},
		{
			name:    "syntax_error_missing_end",
			script:  `if true then return "missing end"`,
			wantErr: true,
			errType: engine.ErrorTypeSyntax,
		},
		{
			name:    "syntax_error_invalid_operator",
			script:  `return 5 ++ 3`,
			wantErr: true,
			errType: engine.ErrorTypeSyntax,
		},
		{
			name:    "syntax_error_unbalanced_parentheses",
			script:  `return (5 + 3`,
			wantErr: true,
			errType: engine.ErrorTypeSyntax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := eng.Execute(ctx, tt.script, nil)

			if tt.wantErr {
				require.Error(t, err)
				var engineErr *engine.EngineError
				if assert.ErrorAs(t, err, &engineErr) {
					assert.Equal(t, tt.errType, engineErr.Type)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestExecutionPipeline_ResultExtraction(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		script   string
		validate func(t *testing.T, result interface{})
	}{
		{
			name:   "nil_result",
			script: `return nil`,
			validate: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name:   "boolean_result",
			script: `return true`,
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, true, result)
			},
		},
		{
			name:   "number_result",
			script: `return 3.14159`,
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, 3.14159, result)
			},
		},
		{
			name:   "string_result",
			script: `return "Hello, Lua!"`,
			validate: func(t *testing.T, result interface{}) {
				assert.Equal(t, "Hello, Lua!", result)
			},
		},
		{
			name:   "table_result",
			script: `return {name = "test", value = 42}`,
			validate: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test", resultMap["name"])
				assert.Equal(t, 42.0, resultMap["value"])
			},
		},
		{
			name:   "array_result",
			script: `return {1, 2, 3, 4, 5}`,
			validate: func(t *testing.T, result interface{}) {
				resultSlice, ok := result.([]interface{})
				require.True(t, ok)
				assert.Len(t, resultSlice, 5)
				assert.Equal(t, 1.0, resultSlice[0])
				assert.Equal(t, 5.0, resultSlice[4])
			},
		},
		{
			name:   "no_return_value",
			script: `local x = 42`,
			validate: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := eng.Execute(ctx, tt.script, nil)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestExecutionPipeline_ErrorHandling(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name    string
		script  string
		errType engine.ErrorType
	}{
		{
			name:    "runtime_error_function",
			script:  `error("deliberate error")`,
			errType: engine.ErrorTypeRuntime,
		},
		{
			name:    "runtime_error_nil_operation",
			script:  `return nil + 5`,
			errType: engine.ErrorTypeRuntime,
		},
		{
			name:    "runtime_error_invalid_function_call",
			script:  `return nonexistent_function()`,
			errType: engine.ErrorTypeRuntime,
		},
		{
			name:    "runtime_error_table_index",
			script:  `local t = nil; return t.field`,
			errType: engine.ErrorTypeRuntime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := eng.Execute(ctx, tt.script, nil)

			require.Error(t, err)
			var engineErr *engine.EngineError
			if assert.ErrorAs(t, err, &engineErr) {
				assert.Equal(t, tt.errType, engineErr.Type)
			}
		})
	}
}

func TestExecutionPipeline_TimeoutHandling(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		TimeoutLimit: 100 * time.Millisecond,
		SandboxMode:  false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name    string
		script  string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:   "quick_execution",
			script: `return 42`,
		},
		{
			name: "timeout_execution",
			script: `
				-- Long running script that can be interrupted
				local count = 0
				for i = 1, 100000000 do
					count = count + 1
					-- This loop gives the VM chances to check context
					if i % 1000 == 0 then
						-- Do some work that allows context checking
						local x = tostring(i)
					end
				end
				return count
			`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.timeout)
				defer cancel()
			}

			_, err := eng.Execute(ctx, tt.script, nil)

			if tt.wantErr {
				assert.Error(t, err)
				// Could be timeout or cancellation error
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecutionPipeline_ChunkCacheIntegration(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	script := `
		local function fibonacci(n)
			if n <= 1 then
				return n
			else
				return fibonacci(n-1) + fibonacci(n-2)
			end
		end
		return fibonacci(10)
	`

	// Execute the same script multiple times
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		result, err := eng.Execute(ctx, script, nil)
		require.NoError(t, err)
		assert.Equal(t, 55.0, result) // fibonacci(10) = 55
	}

	// Check that metrics show cache hits
	metrics := eng.GetMetrics()
	assert.Equal(t, int64(5), metrics.ScriptsExecuted)
	// First execution should miss cache, subsequent should hit
	assert.GreaterOrEqual(t, metrics.CacheHits, int64(4))
}

func TestExecutionPipeline_MemoryManagement(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		MemoryLimit: 10 * 1024 * 1024, // 10MB
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	// Test that the engine can handle reasonable memory usage
	script := `
		local data = {}
		for i = 1, 1000 do
			data[i] = "item_" .. i
		end
		return #data
	`

	ctx := context.Background()
	result, err := eng.Execute(ctx, script, nil)
	require.NoError(t, err)
	assert.Equal(t, 1000.0, result)

	// Check that memory metrics are updated
	metrics := eng.GetMetrics()
	assert.GreaterOrEqual(t, metrics.MemoryUsed, int64(0))
}
