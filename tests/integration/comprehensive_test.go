// ABOUTME: Comprehensive test suite to improve overall code coverage
// ABOUTME: Tests edge cases, error conditions, and uncovered code paths

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/bridge/agent"
	"github.com/lexlapax/go-llmspell/pkg/bridge/llm"
	"github.com/lexlapax/go-llmspell/pkg/bridge/state"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestBridgeAgentCoverage tests uncovered paths in bridge/agent
func TestBridgeAgentCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "agent_bridge_basics",
			test: func(t *testing.T) {
				bridge := agent.NewAgentBridge()
				assert.NotNil(t, bridge)

				// Test bridge ID
				id := bridge.GetID()
				assert.Equal(t, "agent_core", id)

				// Test metadata
				metadata := bridge.GetMetadata()
				assert.Equal(t, "agent_core", metadata.Name)
				assert.NotEmpty(t, metadata.Version)

				// Test initialization
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				assert.NoError(t, err)

				// Test cleanup
				err = bridge.Cleanup(ctx)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestBridgeLLMCoverage tests uncovered paths in bridge/llm
func TestBridgeLLMCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "llm_bridge_basics",
			test: func(t *testing.T) {
				bridge := llm.NewLLMBridge()
				assert.NotNil(t, bridge)

				// Test bridge ID
				id := bridge.GetID()
				assert.Equal(t, "llm_core", id)

				// Test metadata
				metadata := bridge.GetMetadata()
				assert.Equal(t, "llm_core", metadata.Name)

				// Test initialization
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				assert.NoError(t, err)

				// Test that bridge is properly initialized before cleanup
				assert.True(t, bridge.IsInitialized())

				// Test cleanup
				err = bridge.Cleanup(ctx)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestBridgeStateCoverage tests uncovered paths in bridge/state
func TestBridgeStateCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "state_bridge_nil_manager",
			test: func(t *testing.T) {
				// Test nil manager error
				bridge, err := state.NewStateManagerBridge(nil)
				assert.Error(t, err)
				assert.Nil(t, bridge)
				assert.Contains(t, err.Error(), "cannot be nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestGopherLuaEngineCoverage tests basic gopherlua engine functionality
func TestGopherLuaEngineCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "engine_creation",
			test: func(t *testing.T) {
				luaEngine := gopherlua.NewLuaEngine()
				require.NotNil(t, luaEngine)

				// Test engine name and version
				name := luaEngine.Name()
				assert.Equal(t, "lua", name)

				version := luaEngine.Version()
				assert.NotEmpty(t, version)

				// Initialize with basic config
				err := luaEngine.Initialize(engine.EngineConfig{
					TimeoutLimit: 5 * time.Second,
					MemoryLimit:  1024 * 1024,
				})
				assert.NoError(t, err)

				// Test execution
				ctx := context.Background()
				result, err := luaEngine.Execute(ctx, "return 'hello'", nil)
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Cleanup
				err = luaEngine.Shutdown()
				assert.NoError(t, err)
			},
		},
		{
			name: "engine_timeout_handling",
			test: func(t *testing.T) {
				luaEngine := gopherlua.NewLuaEngine()
				require.NotNil(t, luaEngine)

				// Initialize with very short timeout
				err := luaEngine.Initialize(engine.EngineConfig{
					TimeoutLimit: 100 * time.Millisecond,
					MemoryLimit:  1024 * 1024,
				})
				assert.NoError(t, err)

				ctx := context.Background()

				// Test timeout with infinite loop
				script := `
					while true do
						-- Infinite loop to trigger timeout
					end
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.Error(t, err)
				// Result should be an ErrorValue, not nil
				assert.NotNil(t, result)

				// Cleanup
				err = luaEngine.Shutdown()
				assert.NoError(t, err)
			},
		},
		{
			name: "engine_syntax_errors",
			test: func(t *testing.T) {
				luaEngine := gopherlua.NewLuaEngine()
				require.NotNil(t, luaEngine)

				err := luaEngine.Initialize(engine.EngineConfig{
					TimeoutLimit: 5 * time.Second,
					MemoryLimit:  1024 * 1024,
				})
				assert.NoError(t, err)

				ctx := context.Background()

				// Test syntax error
				script := `
					function invalid(
						-- Missing closing parenthesis
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.Error(t, err)
				// Result should be an ErrorValue, not nil
				assert.NotNil(t, result)

				// Cleanup
				err = luaEngine.Shutdown()
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestMockBridgeCoverage tests mock bridge functionality
func TestMockBridgeCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "mock_bridge_operations",
			test: func(t *testing.T) {
				bridge := testutils.NewMockBridge("test-bridge")
				assert.NotNil(t, bridge)

				// Test bridge ID
				assert.Equal(t, "test-bridge", bridge.GetID())

				// Test metadata
				metadata := bridge.GetMetadata()
				assert.Equal(t, "test-bridge", metadata.Name)

				// Test initialization
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				assert.NoError(t, err)

				// Test cleanup
				err = bridge.Cleanup(ctx)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestStdlibModuleCoverage tests stdlib module loading
func TestStdlibModuleCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "stdlib_module_loading",
			test: func(t *testing.T) {
				luaEngine := gopherlua.NewLuaEngine()
				require.NotNil(t, luaEngine)

				err := luaEngine.Initialize(engine.EngineConfig{
					TimeoutLimit: 5 * time.Second,
					MemoryLimit:  1024 * 1024,
				})
				require.NoError(t, err)

				ctx := context.Background()

				// Test basic Lua functionality
				script := `
					-- Test basic Lua operations
					local x = 2 + 2
					local y = "hello" .. " world"
					
					-- Return result
					return {
						math_result = x,
						string_result = y,
						success = true
					}
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Cleanup
				err = luaEngine.Shutdown()
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
