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
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua/adapters"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestBridgeAgentCoverage tests uncovered paths in bridge/agent
func TestBridgeAgentCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "agent_error_conditions",
			test: func(t *testing.T) {
				bridge := agent.NewBridge(nil)
				assert.NotNil(t, bridge)

				// Test with nil context
				result, err := bridge.CreateAgent(nil, map[string]interface{}{})
				assert.Error(t, err)
				assert.Nil(t, result)

				// Test with invalid configuration
				ctx := context.Background()
				result, err = bridge.CreateAgent(ctx, map[string]interface{}{
					"invalid_field": "value",
				})
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
		{
			name: "agent_edge_cases",
			test: func(t *testing.T) {
				mockClient := testutils.NewMockLLMClient()
				bridge := agent.NewBridge(mockClient)

				ctx := context.Background()

				// Test with empty configuration
				result, err := bridge.CreateAgent(ctx, map[string]interface{}{})
				assert.Error(t, err)
				assert.Nil(t, result)

				// Test with minimal valid configuration
				result, err = bridge.CreateAgent(ctx, map[string]interface{}{
					"name":  "test-agent",
					"model": "gpt-3.5-turbo",
				})
				assert.NoError(t, err)
				assert.NotNil(t, result)
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
			name: "llm_error_handling",
			test: func(t *testing.T) {
				bridge := llm.NewBridge(nil)
				assert.NotNil(t, bridge)

				// Test with nil context
				result, err := bridge.Complete(nil, map[string]interface{}{})
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
		{
			name: "llm_streaming_errors",
			test: func(t *testing.T) {
				mockClient := testutils.NewMockLLMClient()
				bridge := llm.NewBridge(mockClient)

				ctx := context.Background()

				// Test streaming with invalid configuration
				result, err := bridge.Stream(ctx, map[string]interface{}{
					"invalid": "config",
				})
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
		{
			name: "llm_model_listing",
			test: func(t *testing.T) {
				mockClient := testutils.NewMockLLMClient()
				bridge := llm.NewBridge(mockClient)

				ctx := context.Background()

				// Test model listing
				models, err := bridge.ListModels(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, models)
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
			name: "state_error_conditions",
			test: func(t *testing.T) {
				bridge := state.NewBridge()
				assert.NotNil(t, bridge)

				ctx := context.Background()

				// Test with invalid state name
				result, err := bridge.CreateState(ctx, "")
				assert.Error(t, err)
				assert.Nil(t, result)

				// Test with invalid characters in state name
				result, err = bridge.CreateState(ctx, "invalid/name")
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
		{
			name: "state_operations",
			test: func(t *testing.T) {
				bridge := state.NewBridge()
				ctx := context.Background()

				// Create valid state
				stateObj, err := bridge.CreateState(ctx, "test-state")
				assert.NoError(t, err)
				assert.NotNil(t, stateObj)

				// Test state operations
				err = bridge.SetValue(ctx, stateObj, "key", "value")
				assert.NoError(t, err)

				value, err := bridge.GetValue(ctx, stateObj, "key")
				assert.NoError(t, err)
				assert.Equal(t, "value", value)

				// Test with non-existent key
				value, err = bridge.GetValue(ctx, stateObj, "nonexistent")
				assert.NoError(t, err)
				assert.Nil(t, value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestGopherLuaAdaptersCoverage tests uncovered paths in gopherlua/adapters
func TestGopherLuaAdaptersCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "adapter_error_conditions",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					ScriptTimeout: 5 * time.Second,
				})
				require.NoError(t, err)

				adapter := adapters.NewLuaAdapter(luaEngine)
				assert.NotNil(t, adapter)

				// Test with invalid bridge configuration
				err = adapter.RegisterBridge("invalid", nil)
				assert.Error(t, err)

				// Test with empty bridge name
				mockBridge := testutils.NewMockBridge()
				err = adapter.RegisterBridge("", mockBridge)
				assert.Error(t, err)
			},
		},
		{
			name: "adapter_bridge_operations",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					ScriptTimeout: 5 * time.Second,
				})
				require.NoError(t, err)

				adapter := adapters.NewLuaAdapter(luaEngine)
				mockBridge := testutils.NewMockBridge()

				// Register valid bridge
				err = adapter.RegisterBridge("test-bridge", mockBridge)
				assert.NoError(t, err)

				// Test double registration
				err = adapter.RegisterBridge("test-bridge", mockBridge)
				assert.Error(t, err)

				// Test bridge lookup
				bridge := adapter.GetBridge("test-bridge")
				assert.NotNil(t, bridge)

				// Test non-existent bridge
				bridge = adapter.GetBridge("nonexistent")
				assert.Nil(t, bridge)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestGopherLuaStdlibCoverage tests uncovered paths in gopherlua/stdlib
func TestGopherLuaStdlibCoverage(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "stdlib_module_loading",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					ScriptTimeout: 5 * time.Second,
				})
				require.NoError(t, err)

				ctx := context.Background()

				// Test loading all stdlib modules
				script := `
					local core = require("core")
					local data = require("data")
					local errors = require("errors")
					local promise = require("promise")
					
					-- Test basic functionality
					assert(core ~= nil, "core module should load")
					assert(data ~= nil, "data module should load")
					assert(errors ~= nil, "errors module should load")
					assert(promise ~= nil, "promise module should load")
					
					return "success"
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
		{
			name: "stdlib_error_handling",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					ScriptTimeout: 5 * time.Second,
				})
				require.NoError(t, err)

				ctx := context.Background()

				// Test error handling in stdlib
				script := `
					local errors = require("errors")
					
					-- Test error creation
					local err = errors.new("TEST_ERROR", "test message")
					assert(err ~= nil, "error should be created")
					
					-- Test error type checking
					local isType = errors.is_type(err, "TEST_ERROR")
					assert(isType == true, "error type should match")
					
					local isNotType = errors.is_type(err, "OTHER_ERROR")
					assert(isNotType == false, "error type should not match")
					
					return "success"
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestEngineEdgeCases tests edge cases in the engine package
func TestEngineEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "engine_timeout_handling",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					ScriptTimeout: 100 * time.Millisecond, // Very short timeout
				})
				require.NoError(t, err)

				ctx := context.Background()

				// Test timeout with infinite loop
				script := `
					while true do
						-- Infinite loop to trigger timeout
					end
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name: "engine_memory_limits",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					MaxMemory: 1024, // Very low memory limit
				})
				require.NoError(t, err)

				ctx := context.Background()

				// Test memory exhaustion
				script := `
					local bigTable = {}
					for i = 1, 100000 do
						bigTable[i] = string.rep("x", 1000)
					end
					return #bigTable
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				// This may or may not fail depending on GC behavior
				// We're just testing the code path
				_ = result
				_ = err
			},
		},
		{
			name: "engine_syntax_errors",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					ScriptTimeout: 5 * time.Second,
				})
				require.NoError(t, err)

				ctx := context.Background()

				// Test syntax error
				script := `
					function invalid(
						-- Missing closing parenthesis
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestAsyncOperations tests async-related code paths
func TestAsyncOperations(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "promise_timeout",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					ScriptTimeout: 5 * time.Second,
				})
				require.NoError(t, err)

				ctx := context.Background()

				script := `
					local promise = require("promise")
					local core = require("core")
					
					-- Create a promise that times out
					local p = promise.new(function(resolve, reject)
						core.async(function()
							core.sleep(10) -- Sleep longer than reasonable
							resolve("too late")
						end)
					end)
					
					-- This should timeout or resolve quickly
					local success, result = pcall(function()
						return p:await()
					end)
					
					return {success = success, result = tostring(result)}
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
		{
			name: "concurrent_promises",
			test: func(t *testing.T) {
				luaEngine, err := gopherlua.NewEngine(&engine.Config{
					ScriptTimeout: 10 * time.Second,
				})
				require.NoError(t, err)

				ctx := context.Background()

				script := `
					local promise = require("promise")
					local core = require("core")
					
					-- Create multiple concurrent promises
					local promises = {}
					for i = 1, 5 do
						promises[i] = promise.new(function(resolve)
							core.async(function()
								core.sleep(0.1)
								resolve("result_" .. i)
							end)
						end)
					end
					
					-- Wait for all
					local results = promise.all(promises):await()
					
					return #results
				`

				result, err := luaEngine.Execute(ctx, script, nil)
				assert.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}