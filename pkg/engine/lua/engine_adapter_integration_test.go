// ABOUTME: Integration tests for LuaEngine adapter factory integration
// ABOUTME: Verifies adapter creation, bridge registration, and module creation flow

package lua

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestEngineAdapterFactoryIntegration tests that the engine creates real adapters using the factory
func TestEngineAdapterFactoryIntegration(t *testing.T) {
	t.Run("single_bridge_adapter_creation", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Create and register a state bridge
		stateBridge := testutils.NewMockBridge("state_manager").WithInitialized(true)
		err = e.RegisterBridge(stateBridge)
		require.NoError(t, err)

		// Verify adapter was created
		adapter := e.GetAdapter("state_manager")
		require.NotNil(t, adapter, "adapter should be created for state_manager")

		// Verify adapter has CreateLuaModule method
		type moduleProvider interface {
			CreateLuaModule() lua.LGFunction
		}
		provider, ok := adapter.(moduleProvider)
		require.True(t, ok, "adapter should implement CreateLuaModule")
		require.NotNil(t, provider.CreateLuaModule())
	})

	t.Run("bridge_map_maintenance", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Register multiple bridges
		bridges := []engine.Bridge{
			testutils.NewMockBridge("state_manager").WithInitialized(true),
			testutils.NewMockBridge("agent_core").WithInitialized(true),
			testutils.NewMockBridge("agent_events").WithInitialized(true),
		}

		for _, bridge := range bridges {
			err = e.RegisterBridge(bridge)
			require.NoError(t, err)
		}

		// Verify all adapters were created
		for _, bridge := range bridges {
			adapter := e.GetAdapter(bridge.GetID())
			require.NotNil(t, adapter, "adapter should exist for %s", bridge.GetID())
		}

		// Unregister a bridge
		err = e.UnregisterBridge("agent_core")
		require.NoError(t, err)

		// Verify adapter was removed
		adapter := e.GetAdapter("agent_core")
		assert.Nil(t, adapter, "adapter should be removed after unregister")

		// Verify other adapters still exist
		assert.NotNil(t, e.GetAdapter("state_manager"))
		assert.NotNil(t, e.GetAdapter("agent_events"))
	})

	t.Run("multi_bridge_adapter_creation", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Register LLM bridges (LLMAdapter needs 3 bridges)
		coreBridge := testutils.NewMockBridge("llm_core").WithInitialized(true)
		providersBridge := testutils.NewMockBridge("llm_providers").WithInitialized(true)
		poolBridge := testutils.NewMockBridge("llm_pool").WithInitialized(true)

		// Register all bridges
		err = e.RegisterBridge(coreBridge)
		require.NoError(t, err)
		err = e.RegisterBridge(providersBridge)
		require.NoError(t, err)
		err = e.RegisterBridge(poolBridge)
		require.NoError(t, err)

		// Verify adapter was created for core bridge
		adapter := e.GetAdapter("llm_core")
		require.NotNil(t, adapter, "LLM adapter should be created")

		// The same adapter instance should handle all LLM functionality
		// (factory creates one LLMAdapter with all three bridges)
	})

	t.Run("bridge_manager_uses_adapters", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Create test Lua state
		L := lua.NewState()
		defer L.Close()

		// Register a bridge
		stateBridge := testutils.NewMockBridge("state_manager").WithInitialized(true)
		err = e.RegisterBridge(stateBridge)
		require.NoError(t, err)

		// Load bridge modules into Lua state
		err = e.LoadBridgeModulesIntoState(L)
		if err != nil {
			t.Logf("LoadBridgeModulesIntoState error: %v", err)
		}
		require.NoError(t, err)

		// Debug: Check bridges table directly through Lua
		err = L.DoString(`
			if bridges ~= nil then
				print("bridges exists and has type:", type(bridges))
				for k,v in pairs(bridges) do
					print("  bridge:", k, type(v))
				end
			else
				print("bridges is nil")
			end
		`)
		require.NoError(t, err)
		
		// Use Lua script to verify bridges table functionality
		err = L.DoString(`
			assert(bridges ~= nil, "bridges table should exist")
			assert(type(bridges) == "table", "bridges should be a table")
			assert(bridges.state_manager ~= nil, "state_manager module should exist")
			assert(type(bridges.state_manager) == "table", "state_manager should be a table")
			return true -- Signal success
		`)
		require.NoError(t, err)
	})

	t.Run("adapter_creation_error_handling", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Try to register a bridge with unknown ID
		unknownBridge := testutils.NewMockBridge("unknown_bridge_id").WithInitialized(true)
		err = e.RegisterBridge(unknownBridge)
		// Factory should return error for unknown bridge ID
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown bridge ID")

		// Verify no adapter was created
		adapter := e.GetAdapter("unknown_bridge_id")
		assert.Nil(t, adapter)
	})

	t.Run("partial_multi_bridge_adapter", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Register only core LLM bridge (without providers and pool)
		coreBridge := testutils.NewMockBridge("llm_core").WithInitialized(true)
		err = e.RegisterBridge(coreBridge)
		require.NoError(t, err)

		// Adapter should still be created with nil optional bridges
		adapter := e.GetAdapter("llm_core")
		require.NotNil(t, adapter, "LLM adapter should be created even with optional bridges missing")
	})
}

// TestBridgeManagerModuleCreator tests that BridgeManager properly uses the moduleCreator
func TestBridgeManagerModuleCreator(t *testing.T) {
	t.Run("module_creator_called_for_adapters", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Register a bridge
		bridge := testutils.NewMockBridge("agent_events").WithInitialized(true)
		err = e.RegisterBridge(bridge)
		require.NoError(t, err)

		// Create Lua state for testing
		L := lua.NewState()
		defer L.Close()

		// Get the module through BridgeManager
		bridgeManager := e.bridgeManager
		module, err := bridgeManager.CreateLuaModule(L, "agent_events")
		require.NoError(t, err)
		require.NotNil(t, module)
		require.Equal(t, lua.LTTable, module.Type())
	})

	t.Run("module_creator_error_no_adapter", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Create Lua state for testing
		L := lua.NewState()
		defer L.Close()

		// Try to create module without registering bridge/adapter
		bridgeManager := e.bridgeManager
		_, err = bridgeManager.CreateLuaModule(L, "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no adapter found")
	})
}

// TestComplexAdapterDependencies tests scenarios with complex multi-bridge adapters
func TestComplexAdapterDependencies(t *testing.T) {
	t.Run("utils_adapter_with_multiple_bridges", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Register utility bridges
		utilBridges := []engine.Bridge{
			testutils.NewMockBridge("auth").WithInitialized(true),
			testutils.NewMockBridge("util_debug").WithInitialized(true),
			testutils.NewMockBridge("util_json").WithInitialized(true),
		}

		for _, bridge := range utilBridges {
			err = e.RegisterBridge(bridge)
			require.NoError(t, err)
		}

		// Utils adapter should be created for any util bridge
		adapter := e.GetAdapter("auth")
		require.NotNil(t, adapter, "utils adapter should be created")

		// Same adapter should handle all util functionality
		adapter2 := e.GetAdapter("util_json")
		require.NotNil(t, adapter2)
	})

	t.Run("observability_adapter_with_optional_bridges", func(t *testing.T) {
		// Create engine
		e := NewLuaEngine()
		err := e.Initialize(engine.EngineConfig{})
		require.NoError(t, err)
		defer e.Shutdown()

		// Register only metrics bridge (tracing and guardrails are optional)
		metricsBridge := testutils.NewMockBridge("observability_metrics").WithInitialized(true)
		err = e.RegisterBridge(metricsBridge)
		require.NoError(t, err)

		// Adapter should be created with nil optional bridges
		adapter := e.GetAdapter("observability_metrics")
		require.NotNil(t, adapter, "observability adapter should work with only metrics bridge")

		// Now add optional bridges
		tracingBridge := testutils.NewMockBridge("observability_tracing").WithInitialized(true)
		err = e.RegisterBridge(tracingBridge)
		require.NoError(t, err)

		// Adapter should still exist and now have access to tracing
		adapter2 := e.GetAdapter("observability_metrics")
		require.NotNil(t, adapter2)
	})
}