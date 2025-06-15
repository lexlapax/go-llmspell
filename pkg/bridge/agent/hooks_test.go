// ABOUTME: Tests for hooks bridge providing access to go-llms agent hook system
// ABOUTME: Verifies hook registration, execution, priority ordering, and lifecycle integration

package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/helpers"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestHooksBridge_BasicOperations(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *HooksBridge)
	}{
		{
			name: "GetID returns correct identifier",
			test: func(t *testing.T, b *HooksBridge) {
				assert.Equal(t, "hooks", b.GetID())
			},
		},
		{
			name: "GetMetadata returns valid metadata",
			test: func(t *testing.T, b *HooksBridge) {
				metadata := b.GetMetadata()
				assert.Equal(t, "Hooks Bridge", metadata.Name)
				assert.NotEmpty(t, metadata.Version)
				assert.NotEmpty(t, metadata.Description)
				assert.Equal(t, "go-llmspell", metadata.Author)
			},
		},
		{
			name: "Initialize and cleanup work correctly",
			test: func(t *testing.T, b *HooksBridge) {
				ctx := context.Background()

				// Initial state
				assert.False(t, b.IsInitialized())

				// Initialize
				err := b.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, b.IsInitialized())

				// Double initialize should be safe
				err = b.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, b.IsInitialized())

				// Cleanup
				err = b.Cleanup(ctx)
				require.NoError(t, err)
				assert.False(t, b.IsInitialized())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewHooksBridge()
			tt.test(t, bridge)
		})
	}
}

func TestHooksBridge_Methods(t *testing.T) {
	bridge := NewHooksBridge()
	methods := bridge.Methods()

	// Check that we have the expected hook-related methods
	expectedMethods := []string{
		"registerHook",
		"unregisterHook",
		"listHooks",
		"enableHook",
		"disableHook",
		"getHookInfo",
		"executeHooks",
		"clearHooks",
	}

	methodMap := make(map[string]engine.MethodInfo)
	for _, m := range methods {
		methodMap[m.Name] = m
	}

	for _, expected := range expectedMethods {
		t.Run("has_method_"+expected, func(t *testing.T) {
			method, exists := methodMap[expected]
			assert.True(t, exists, "Missing method: %s", expected)
			assert.NotEmpty(t, method.Description)
			assert.NotEmpty(t, method.ReturnType)
		})
	}
}

func TestHooksBridge_HookRegistration(t *testing.T) {
	ctx := context.Background()
	bridge := NewHooksBridge()
	require.NoError(t, bridge.Initialize(ctx))

	tests := []struct {
		name     string
		method   string
		args     []interface{}
		validate func(t *testing.T, result interface{}, err error)
	}{
		{
			name:   "registerHook adds new hook",
			method: "registerHook",
			args: []interface{}{
				"test_hook",
				map[string]interface{}{
					"beforeGenerate": func(ctx interface{}, messages interface{}) {
						// Mock implementation
					},
					"afterGenerate": func(ctx interface{}, response interface{}, err interface{}) {
						// Mock implementation
					},
					"priority": 10,
				},
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, "test_hook", result)
			},
		},
		{
			name:   "listHooks returns registered hooks",
			method: "listHooks",
			args:   []interface{}{},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				hooks, ok := result.([]map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, hooks)

				// Find our test hook
				found := false
				for _, hook := range hooks {
					if hook["id"] == "test_hook" {
						found = true
						assert.Equal(t, true, hook["enabled"])
						assert.Equal(t, 10, hook["priority"])
						break
					}
				}
				assert.True(t, found, "Test hook not found in list")
			},
		},
		{
			name:   "getHookInfo returns hook details",
			method: "getHookInfo",
			args:   []interface{}{"test_hook"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				info, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test_hook", info["id"])
				assert.Equal(t, true, info["enabled"])
				assert.Equal(t, 10, info["priority"])
			},
		},
		{
			name:   "disableHook disables the hook",
			method: "disableHook",
			args:   []interface{}{"test_hook"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, true, result)
			},
		},
		{
			name:   "enableHook re-enables the hook",
			method: "enableHook",
			args:   []interface{}{"test_hook"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, true, result)
			},
		},
		{
			name:   "unregisterHook removes the hook",
			method: "unregisterHook",
			args:   []interface{}{"test_hook"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, true, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			tt.validate(t, result, err)
		})
	}
}

func TestHooksBridge_HookExecution(t *testing.T) {
	ctx := context.Background()
	bridge := NewHooksBridge()
	require.NoError(t, bridge.Initialize(ctx))

	// Track hook executions
	executions := []string{}

	// Register multiple hooks with different priorities
	hooks := []struct {
		id       string
		priority int
	}{
		{"hook_low", 1},
		{"hook_medium", 5},
		{"hook_high", 10},
	}

	for _, h := range hooks {
		_, err := bridge.ExecuteMethod(ctx, "registerHook", []interface{}{
			h.id,
			map[string]interface{}{
				"beforeGenerate": func(ctx interface{}, messages interface{}) {
					executions = append(executions, h.id+"_before")
				},
				"afterGenerate": func(ctx interface{}, response interface{}, err interface{}) {
					executions = append(executions, h.id+"_after")
				},
				"priority": h.priority,
			},
		})
		require.NoError(t, err)
	}

	// Test executeHooks for BeforeGenerate
	t.Run("executeHooks BeforeGenerate respects priority", func(t *testing.T) {
		executions = []string{} // Reset

		result, err := bridge.ExecuteMethod(ctx, "executeHooks", []interface{}{
			"beforeGenerate",
			map[string]interface{}{
				"messages": []map[string]interface{}{
					{"role": "user", "content": "test"},
				},
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Check execution order (high priority first)
		assert.Equal(t, []string{
			"hook_high_before",
			"hook_medium_before",
			"hook_low_before",
		}, executions)
	})

	// Test executeHooks for AfterGenerate
	t.Run("executeHooks AfterGenerate respects priority", func(t *testing.T) {
		executions = []string{} // Reset

		result, err := bridge.ExecuteMethod(ctx, "executeHooks", []interface{}{
			"afterGenerate",
			map[string]interface{}{
				"response": map[string]interface{}{
					"content": "Generated response",
				},
				"error": nil,
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Check execution order (high priority first)
		assert.Equal(t, []string{
			"hook_high_after",
			"hook_medium_after",
			"hook_low_after",
		}, executions)
	})

	// Test with disabled hook
	t.Run("disabled hooks are not executed", func(t *testing.T) {
		// Disable medium priority hook
		_, err := bridge.ExecuteMethod(ctx, "disableHook", []interface{}{"hook_medium"})
		require.NoError(t, err)

		executions = []string{} // Reset

		_, err = bridge.ExecuteMethod(ctx, "executeHooks", []interface{}{
			"beforeGenerate",
			map[string]interface{}{
				"messages": []map[string]interface{}{
					{"role": "user", "content": "test"},
				},
			},
		})
		require.NoError(t, err)

		// Medium hook should not execute
		assert.Equal(t, []string{
			"hook_high_before",
			"hook_low_before",
		}, executions)
	})
}

func TestHooksBridge_ToolHooks(t *testing.T) {
	ctx := context.Background()
	bridge := NewHooksBridge()
	require.NoError(t, bridge.Initialize(ctx))

	// Track tool hook executions
	toolCalls := []string{}

	// Register a tool hook
	_, err := bridge.ExecuteMethod(ctx, "registerHook", []interface{}{
		"tool_monitor",
		map[string]interface{}{
			"beforeToolCall": func(ctx interface{}, tool interface{}, params interface{}) {
				toolName := tool.(string)
				toolCalls = append(toolCalls, "before_"+toolName)
			},
			"afterToolCall": func(ctx interface{}, tool interface{}, result interface{}, err interface{}) {
				toolName := tool.(string)
				toolCalls = append(toolCalls, "after_"+toolName)
			},
			"priority": 5,
		},
	})
	require.NoError(t, err)

	// Test tool hook execution
	t.Run("tool hooks execute correctly", func(t *testing.T) {
		toolCalls = []string{} // Reset

		// Execute beforeToolCall
		_, err := bridge.ExecuteMethod(ctx, "executeHooks", []interface{}{
			"beforeToolCall",
			map[string]interface{}{
				"tool": "calculator",
				"params": map[string]interface{}{
					"operation": "add",
					"a":         5,
					"b":         3,
				},
			},
		})
		require.NoError(t, err)

		// Execute afterToolCall
		_, err = bridge.ExecuteMethod(ctx, "executeHooks", []interface{}{
			"afterToolCall",
			map[string]interface{}{
				"tool":   "calculator",
				"result": 8,
				"error":  nil,
			},
		})
		require.NoError(t, err)

		assert.Equal(t, []string{"before_calculator", "after_calculator"}, toolCalls)
	})
}

func TestHooksBridge_ClearHooks(t *testing.T) {
	ctx := context.Background()
	bridge := NewHooksBridge()
	require.NoError(t, bridge.Initialize(ctx))

	// Register some hooks
	for i := 0; i < 3; i++ {
		_, err := bridge.ExecuteMethod(ctx, "registerHook", []interface{}{
			string(rune('a'+i)) + "_hook",
			map[string]interface{}{
				"beforeGenerate": func(ctx interface{}, messages interface{}) {},
				"priority":       i,
			},
		})
		require.NoError(t, err)
	}

	// Verify hooks exist
	result, err := bridge.ExecuteMethod(ctx, "listHooks", []interface{}{})
	require.NoError(t, err)
	hooks := result.([]map[string]interface{})
	assert.Len(t, hooks, 3)

	// Clear all hooks
	result, err = bridge.ExecuteMethod(ctx, "clearHooks", []interface{}{})
	require.NoError(t, err)
	assert.Equal(t, 3, result) // Should return count of cleared hooks

	// Verify hooks are cleared
	result, err = bridge.ExecuteMethod(ctx, "listHooks", []interface{}{})
	require.NoError(t, err)
	hooks = result.([]map[string]interface{})
	assert.Empty(t, hooks)
}

func TestHooksBridge_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	bridge := NewHooksBridge()

	t.Run("methods fail when not initialized", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "registerHook", []interface{}{"test", map[string]interface{}{}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	require.NoError(t, bridge.Initialize(ctx))

	t.Run("unknown method returns error", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "unknownMethod", []interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method not found")
	})

	t.Run("getHookInfo with non-existent hook", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "getHookInfo", []interface{}{"non_existent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("invalid arguments return error", func(t *testing.T) {
		// Missing required arguments
		_, err := bridge.ExecuteMethod(ctx, "registerHook", []interface{}{})
		assert.Error(t, err)

		// Wrong argument type
		_, err = bridge.ExecuteMethod(ctx, "registerHook", []interface{}{123, "not a map"})
		assert.Error(t, err)
	})
}

func TestHooksBridge_TypeMappings(t *testing.T) {
	bridge := NewHooksBridge()
	mappings := bridge.TypeMappings()

	// Check that we have the expected type mappings
	expectedTypes := []string{
		"Hook",
		"HookInfo",
		"HookType",
		"HookContext",
	}

	for _, typeName := range expectedTypes {
		t.Run("has_type_"+typeName, func(t *testing.T) {
			mapping, exists := mappings[typeName]
			assert.True(t, exists, "Missing type mapping: %s", typeName)
			assert.NotEmpty(t, mapping.GoType)
			assert.NotEmpty(t, mapping.ScriptType)
		})
	}
}

func TestHooksBridge_Permissions(t *testing.T) {
	bridge := NewHooksBridge()
	permissions := bridge.RequiredPermissions()

	// Should require permissions for hook operations
	assert.NotEmpty(t, permissions)

	// Check for essential permissions
	hasHookPermission := false
	for _, perm := range permissions {
		if perm.Type == engine.PermissionProcess && perm.Resource == "hook" {
			hasHookPermission = true
			assert.Contains(t, perm.Actions, "register")
			assert.Contains(t, perm.Actions, "execute")
		}
	}
	assert.True(t, hasHookPermission, "Missing hook management permission")
}

// Test with go-llms testutils mock hooks
func TestHooksBridge_WithMockHooks(t *testing.T) {
	ctx := context.Background()
	bridge := NewHooksBridge()
	require.NoError(t, bridge.Initialize(ctx))

	// Create a mock hook using testutils
	// Track executions
	var beforeGenerateCalled bool
	var afterGenerateCalled bool

	// We'll test by registering hooks that track their calls

	// Register the mock hook through the bridge
	// Note: We need to wrap it in a script-compatible format
	_, err := bridge.ExecuteMethod(ctx, "registerHook", []interface{}{
		"mock_hook",
		map[string]interface{}{
			"beforeGenerate": func(ctx interface{}, messages interface{}) {
				// Track that the hook was called
				beforeGenerateCalled = true
			},
			"afterGenerate": func(ctx interface{}, response interface{}, err interface{}) {
				afterGenerateCalled = true
			},
			"priority": 10,
		},
	})
	require.NoError(t, err)

	// Execute hooks
	_, err = bridge.ExecuteMethod(ctx, "executeHooks", []interface{}{
		"beforeGenerate",
		map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "user", "content": "test message"},
			},
		},
	})
	require.NoError(t, err)
	assert.True(t, beforeGenerateCalled)

	_, err = bridge.ExecuteMethod(ctx, "executeHooks", []interface{}{
		"afterGenerate",
		map[string]interface{}{
			"response": map[string]interface{}{
				"content": "generated text",
			},
			"error": nil,
		},
	})
	require.NoError(t, err)
	assert.True(t, afterGenerateCalled)
}

// Test hook integration with agent helpers
func TestHooksBridge_WithAgentHelpers(t *testing.T) {
	ctx := context.Background()
	bridge := NewHooksBridge()
	require.NoError(t, bridge.Initialize(ctx))

	// Create a mock agent with delay to test hook timing
	agent := helpers.CreateMockAgentWithDelay("test_agent", 100)

	// Track hook execution timing
	var hookOrder []string

	// Register hooks that track execution order
	_, err := bridge.ExecuteMethod(ctx, "registerHook", []interface{}{
		"timing_hook",
		map[string]interface{}{
			"beforeGenerate": func(ctx interface{}, messages interface{}) {
				hookOrder = append(hookOrder, "before_start")
				// Simulate agent starting
				go func() {
					state := domain.NewState()
					_, _ = agent.Run(context.Background(), state)
					hookOrder = append(hookOrder, "agent_complete")
				}()
			},
			"afterGenerate": func(ctx interface{}, response interface{}, err interface{}) {
				hookOrder = append(hookOrder, "after_complete")
			},
			"priority": 1,
		},
	})
	require.NoError(t, err)

	// Execute hooks
	_, err = bridge.ExecuteMethod(ctx, "executeHooks", []interface{}{
		"beforeGenerate",
		map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "user", "content": "test"},
			},
		},
	})
	require.NoError(t, err)

	// Verify execution started
	assert.Contains(t, hookOrder, "before_start")
}
