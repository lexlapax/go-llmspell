// ABOUTME: Tests for LuaEngine which implements the ScriptEngine interface for Lua script execution
// ABOUTME: Validates engine lifecycle, script execution, bridge integration, and resource management

package gopherlua

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestLuaEngine_Lifecycle(t *testing.T) {
	tests := []struct {
		name    string
		config  engine.EngineConfig
		wantErr bool
	}{
		{
			name: "default_config_initialization",
			config: engine.EngineConfig{
				MemoryLimit:  64 * 1024 * 1024, // 64MB
				TimeoutLimit: 30 * time.Second,
				SandboxMode:  true,
			},
		},
		{
			name: "custom_pool_config",
			config: engine.EngineConfig{
				MemoryLimit:    32 * 1024 * 1024, // 32MB
				TimeoutLimit:   10 * time.Second,
				SandboxMode:    true,
				GoroutineLimit: 5,
				EngineOptions: map[string]interface{}{
					"pool_min_size":     2,
					"pool_max_size":     8,
					"pool_idle_timeout": "5m",
					"health_threshold":  0.8,
					"cleanup_interval":  "1m",
					"security_level":    "strict",
				},
			},
		},
		{
			name: "minimal_config",
			config: engine.EngineConfig{
				SandboxMode: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng := NewLuaEngine()
			require.NotNil(t, eng)

			// Test initialization
			err := eng.Initialize(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Test engine metadata
			assert.Equal(t, "lua", eng.Name())
			assert.NotEmpty(t, eng.Version())
			assert.Contains(t, eng.FileExtensions(), ".lua")

			// Test supported features
			features := eng.Features()
			assert.Contains(t, features, engine.FeatureCoroutines)
			assert.Contains(t, features, engine.FeatureModules)

			// Test metrics
			metrics := eng.GetMetrics()
			assert.GreaterOrEqual(t, metrics.ScriptsExecuted, int64(0))

			// Test shutdown
			err = eng.Shutdown()
			assert.NoError(t, err)
		})
	}
}

func TestLuaEngine_BasicExecution(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		MemoryLimit:  32 * 1024 * 1024,
		TimeoutLimit: 10 * time.Second,
		SandboxMode:  true,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		script   string
		params   map[string]interface{}
		wantErr  bool
		validate func(t *testing.T, result interface{})
	}{
		{
			name:   "simple_arithmetic",
			script: "return 2 + 3",
			validate: func(t *testing.T, result interface{}) {
				sv, ok := result.(engine.ScriptValue)
				require.True(t, ok)
				assert.Equal(t, 5.0, sv.ToGo())
			},
		},
		{
			name:   "string_operation",
			script: `return "Hello, " .. "World!"`,
			validate: func(t *testing.T, result interface{}) {
				sv, ok := result.(engine.ScriptValue)
				require.True(t, ok)
				assert.Equal(t, "Hello, World!", sv.ToGo())
			},
		},
		{
			name:   "parameter_usage",
			script: "return name .. ' is ' .. age .. ' years old'",
			params: map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
			validate: func(t *testing.T, result interface{}) {
				sv, ok := result.(engine.ScriptValue)
				require.True(t, ok)
				assert.Equal(t, "Alice is 30 years old", sv.ToGo())
			},
		},
		{
			name:   "table_creation",
			script: "local x, y = 10, 20; return {x = x, y = y, z = x + y}",
			validate: func(t *testing.T, result interface{}) {
				sv, ok := result.(engine.ScriptValue)
				require.True(t, ok)
				resultMap, ok := sv.ToGo().(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, 10.0, resultMap["x"])
				assert.Equal(t, 20.0, resultMap["y"])
				assert.Equal(t, 30.0, resultMap["z"])
			},
		},
		{
			name:   "array_creation",
			script: "return {1, 2, 3, 4, 5}",
			validate: func(t *testing.T, result interface{}) {
				sv, ok := result.(engine.ScriptValue)
				require.True(t, ok)
				// Lua tables with numeric indices can be either arrays or objects
				resultGo := sv.ToGo()
				if resultSlice, ok := resultGo.([]interface{}); ok {
					assert.Len(t, resultSlice, 5)
					assert.Equal(t, 1.0, resultSlice[0])
					assert.Equal(t, 5.0, resultSlice[4])
				} else if resultMap, ok := resultGo.(map[string]interface{}); ok {
					// Lua uses 1-based indexing
					assert.Equal(t, 1.0, resultMap["1"])
					assert.Equal(t, 5.0, resultMap["5"])
				} else {
					t.Fatalf("unexpected result type: %T", resultGo)
				}
			},
		},
		{
			name:    "syntax_error",
			script:  "return 2 +",
			wantErr: true,
		},
		{
			name:    "runtime_error",
			script:  "error('test error')",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := eng.Execute(ctx, tt.script, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestLuaEngine_FileExecution(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		MemoryLimit:  32 * 1024 * 1024,
		TimeoutLimit: 10 * time.Second,
		SandboxMode:  true,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	// Note: In a real test, you'd write this to a temp file
	// For now, we'll test the basic functionality
	t.Run("non_existent_file", func(t *testing.T) {
		ctx := context.Background()
		_, err := eng.ExecuteFile(ctx, "/tmp/nonexistent.lua", nil)
		assert.Error(t, err)
	})
}

func TestLuaEngine_TypeConversion(t *testing.T) {
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
		goValue  interface{}
		validate func(t *testing.T, scriptValue engine.ScriptValue)
	}{
		{
			name:    "boolean_true",
			goValue: true,
			validate: func(t *testing.T, scriptValue engine.ScriptValue) {
				converted, err := eng.ToNative(scriptValue)
				require.NoError(t, err)
				assert.Equal(t, true, converted)
			},
		},
		{
			name:    "number_int",
			goValue: 42,
			validate: func(t *testing.T, scriptValue engine.ScriptValue) {
				converted, err := eng.ToNative(scriptValue)
				require.NoError(t, err)
				assert.Equal(t, 42.0, converted) // Lua numbers are float64
			},
		},
		{
			name:    "string_value",
			goValue: "test string",
			validate: func(t *testing.T, scriptValue engine.ScriptValue) {
				converted, err := eng.ToNative(scriptValue)
				require.NoError(t, err)
				assert.Equal(t, "test string", converted)
			},
		},
		{
			name: "map_value",
			goValue: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			validate: func(t *testing.T, scriptValue engine.ScriptValue) {
				converted, err := eng.ToNative(scriptValue)
				require.NoError(t, err)
				convertedMap, ok := converted.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "value1", convertedMap["key1"])
				assert.Equal(t, 123.0, convertedMap["key2"])
			},
		},
		{
			name:    "slice_value",
			goValue: []interface{}{1, 2, 3},
			validate: func(t *testing.T, scriptValue engine.ScriptValue) {
				converted, err := eng.ToNative(scriptValue)
				require.NoError(t, err)
				convertedSlice, ok := converted.([]interface{})
				require.True(t, ok)
				assert.Len(t, convertedSlice, 3)
				assert.Equal(t, 1.0, convertedSlice[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Go -> Script conversion
			scriptValue, err := eng.FromNative(tt.goValue)
			require.NoError(t, err)
			require.NotNil(t, scriptValue)

			// Test Script -> Go conversion
			if tt.validate != nil {
				tt.validate(t, scriptValue)
			}
		})
	}
}

func TestLuaEngine_ResourceLimits(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		MemoryLimit:  1024 * 1024, // 1MB - very small
		TimeoutLimit: 100 * time.Millisecond,
		SandboxMode:  true,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	t.Run("memory_limit_test", func(t *testing.T) {
		// Test setting memory limit
		err := eng.SetMemoryLimit(2 * 1024 * 1024) // 2MB
		assert.NoError(t, err)
	})

	t.Run("timeout_limit_test", func(t *testing.T) {
		// Test setting timeout
		err := eng.SetTimeout(5 * time.Second)
		assert.NoError(t, err)
	})

	t.Run("resource_limits_test", func(t *testing.T) {
		limits := engine.ResourceLimits{
			MaxMemory:     4 * 1024 * 1024, // 4MB
			MaxExecTime:   2 * time.Second,
			MaxGoroutines: 10,
		}
		err := eng.SetResourceLimits(limits)
		assert.NoError(t, err)
	})

	t.Run("timeout_execution", func(t *testing.T) {
		// Set a very short timeout
		err := eng.SetTimeout(50 * time.Millisecond)
		require.NoError(t, err)

		ctx := context.Background()
		// Script that should timeout
		longScript := `
			local start = os.clock()
			while os.clock() - start < 1 do
				-- busy wait
			end
			return "completed"
		`

		_, err = eng.Execute(ctx, longScript, nil)
		// Should timeout or be cancelled
		assert.Error(t, err)
	})
}

func TestLuaEngine_Security(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode:     true,
		AllowedModules:  []string{"string", "table", "math"},
		DisabledModules: []string{"io", "os"},
		FileSystemMode:  engine.FSModeNone,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name    string
		script  string
		wantErr bool
		errType string
	}{
		{
			name:   "allowed_string_lib",
			script: `return string.upper("hello")`,
		},
		{
			name:   "allowed_math_lib",
			script: `return math.pi`,
		},
		{
			name:    "blocked_io_access",
			script:  `return io.open("/tmp/test", "r")`,
			wantErr: true,
		},
		{
			name:    "blocked_os_execute",
			script:  `return os.execute("echo hello")`,
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
			result, err := eng.Execute(ctx, tt.script, nil)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != "" {
					assert.Contains(t, strings.ToLower(err.Error()), tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestLuaEngine_Concurrency(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		MemoryLimit:    64 * 1024 * 1024,
		TimeoutLimit:   10 * time.Second,
		SandboxMode:    true,
		GoroutineLimit: 20,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	t.Run("concurrent_execution", func(t *testing.T) {
		const numGoroutines = 10
		const numExecutions = 5

		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < numExecutions; j++ {
					ctx := context.Background()
					script := `return math.random(1, 100)`

					_, err := eng.Execute(ctx, script, nil)
					if err != nil {
						results <- err
						return
					}
				}
				results <- nil
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}

		// Check metrics
		metrics := eng.GetMetrics()
		assert.GreaterOrEqual(t, metrics.ScriptsExecuted, int64(numGoroutines*numExecutions))
	})
}

func TestLuaEngine_ErrorHandling(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: true,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	tests := []struct {
		name      string
		script    string
		errorType engine.ErrorType
	}{
		{
			name:      "syntax_error",
			script:    "return 2 +",
			errorType: engine.ErrorTypeSyntax,
		},
		{
			name:      "runtime_error",
			script:    "error('runtime error')",
			errorType: engine.ErrorTypeRuntime,
		},
		{
			name:      "type_error",
			script:    "return nil + 5",
			errorType: engine.ErrorTypeRuntime, // In Lua, this is actually a runtime error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := eng.Execute(ctx, tt.script, nil)

			require.Error(t, err)

			var engineErr *engine.EngineError
			if assert.ErrorAs(t, err, &engineErr) {
				assert.Equal(t, tt.errorType, engineErr.Type)
			}
		})
	}
}

// Mock bridge for testing
type testBridge struct {
	id   string
	meta engine.BridgeMetadata
}

func (b *testBridge) GetID() string {
	return b.id
}

func (b *testBridge) GetMetadata() engine.BridgeMetadata {
	return b.meta
}

func (b *testBridge) Initialize(ctx context.Context) error {
	return nil
}

func (b *testBridge) Cleanup(ctx context.Context) error {
	return nil
}

func (b *testBridge) IsInitialized() bool {
	return true
}

func (b *testBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return nil
}

func (b *testBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "testMethod",
			Description: "Test method",
			Parameters: []engine.ParameterInfo{
				{Name: "input", Type: "string", Required: true},
			},
			ReturnType: "string",
		},
	}
}

func (b *testBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	if name == "testMethod" && len(args) == 1 {
		return nil
	}
	return errors.New("invalid method call")
}

func (b *testBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	switch name {
	case "testMethod":
		if len(args) > 0 {
			return engine.NewStringValue("result: " + args[0].String()), nil
		}
		return engine.NewStringValue("result: no input"), nil
	default:
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), fmt.Errorf("unknown method: %s", name)
	}
}

func (b *testBridge) TypeMappings() map[string]engine.TypeMapping {
	return nil
}

func (b *testBridge) RequiredPermissions() []engine.Permission {
	return nil
}

func TestLuaEngine_BridgeManagement(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	bridge := &testBridge{
		id: "test_bridge",
		meta: engine.BridgeMetadata{
			Name:        "Test Bridge",
			Version:     "1.0.0",
			Description: "Bridge for testing",
		},
	}

	t.Run("register_bridge", func(t *testing.T) {
		err := eng.RegisterBridge(bridge)
		assert.NoError(t, err)

		// Check bridge is registered
		bridges := eng.ListBridges()
		assert.Contains(t, bridges, "test_bridge")

		// Get bridge
		retrieved, err := eng.GetBridge("test_bridge")
		assert.NoError(t, err)
		assert.Equal(t, bridge, retrieved)
	})

	t.Run("unregister_bridge", func(t *testing.T) {
		err := eng.UnregisterBridge("test_bridge")
		assert.NoError(t, err)

		// Check bridge is unregistered
		bridges := eng.ListBridges()
		assert.NotContains(t, bridges, "test_bridge")

		// Should not be able to get it
		_, err = eng.GetBridge("test_bridge")
		assert.Error(t, err)
	})
}
