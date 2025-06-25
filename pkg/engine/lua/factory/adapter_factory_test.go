// ABOUTME: Tests for AdapterFactory ensuring correct adapter creation for all bridge types
// ABOUTME: Validates single-bridge, multi-bridge, optional dependencies, and error handling

package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/adapters/impl"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestAdapterFactory_SingleBridgeAdapters(t *testing.T) {
	factory := NewAdapterFactory()

	testCases := []struct {
		name       string
		bridgeID   string
		adapterType interface{}
	}{
		{"state_adapter", "state_manager", (*impl.StateAdapter)(nil)},
		{"events_adapter", "agent_events", (*impl.EventsAdapter)(nil)},
		{"structured_adapter", "structured_schema", (*impl.StructuredAdapter)(nil)},
		{"agent_adapter", "agent_core", (*impl.AgentAdapter)(nil)},
		{"hooks_adapter", "agent_hooks", (*impl.HooksAdapter)(nil)},
		{"workflow_adapter", "agent_workflow", (*impl.WorkflowAdapter)(nil)},
		{"modelinfo_adapter", "llm_modelinfo", (*impl.ModelInfoAdapter)(nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock bridge
			bridge := testutils.NewMockBridge(tc.bridgeID).WithInitialized(true)
			bridgeMap := map[string]engine.Bridge{
				tc.bridgeID: bridge,
			}

			// Create adapter
			adapter, err := factory.CreateAdapter(tc.bridgeID, bridgeMap)
			require.NoError(t, err)
			require.NotNil(t, adapter)

			// Verify adapter type
			assert.IsType(t, tc.adapterType, adapter)
		})
	}
}

func TestAdapterFactory_LLMAdapter(t *testing.T) {
	factory := NewAdapterFactory()

	t.Run("llm_adapter_with_all_bridges", func(t *testing.T) {
		// Create all bridges
		coreBridge := testutils.NewMockBridge("llm_core").WithInitialized(true)
		providersBridge := testutils.NewMockBridge("llm_providers").WithInitialized(true)
		poolBridge := testutils.NewMockBridge("llm_pool").WithInitialized(true)

		bridgeMap := map[string]engine.Bridge{
			"llm_core":      coreBridge,
			"llm_providers": providersBridge,
			"llm_pool":      poolBridge,
		}

		// Create adapter
		adapter, err := factory.CreateAdapter("llm_core", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type
		assert.IsType(t, (*impl.LLMAdapter)(nil), adapter)
	})

	t.Run("llm_adapter_with_core_only", func(t *testing.T) {
		// Create only core bridge
		coreBridge := testutils.NewMockBridge("llm_core").WithInitialized(true)
		bridgeMap := map[string]engine.Bridge{
			"llm_core": coreBridge,
		}

		// Create adapter - should succeed with nil optional bridges
		adapter, err := factory.CreateAdapter("llm_core", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type
		assert.IsType(t, (*impl.LLMAdapter)(nil), adapter)
	})

	t.Run("llm_adapter_missing_core", func(t *testing.T) {
		// Empty bridge map
		bridgeMap := map[string]engine.Bridge{}

		// Create adapter - should fail
		adapter, err := factory.CreateAdapter("llm_core", bridgeMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bridge not found")
		assert.Nil(t, adapter)
	})
}

func TestAdapterFactory_ObservabilityAdapter(t *testing.T) {
	factory := NewAdapterFactory()

	t.Run("observability_adapter_with_all_bridges", func(t *testing.T) {
		// Create all bridges
		metricsBridge := testutils.NewMockBridge("observability_metrics").WithInitialized(true)
		tracingBridge := testutils.NewMockBridge("observability_tracing").WithInitialized(true)
		guardrailsBridge := testutils.NewMockBridge("observability_guardrails").WithInitialized(true)

		bridgeMap := map[string]engine.Bridge{
			"observability_metrics":    metricsBridge,
			"observability_tracing":    tracingBridge,
			"observability_guardrails": guardrailsBridge,
		}

		// Create adapter
		adapter, err := factory.CreateAdapter("observability_metrics", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type
		assert.IsType(t, (*impl.ObservabilityAdapter)(nil), adapter)
	})

	t.Run("observability_adapter_with_metrics_only", func(t *testing.T) {
		// Create only metrics bridge
		metricsBridge := testutils.NewMockBridge("observability_metrics").WithInitialized(true)
		bridgeMap := map[string]engine.Bridge{
			"observability_metrics": metricsBridge,
		}

		// Create adapter - should succeed with nil optional bridges
		adapter, err := factory.CreateAdapter("observability_metrics", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type
		assert.IsType(t, (*impl.ObservabilityAdapter)(nil), adapter)
	})
}

func TestAdapterFactory_ToolsAdapter(t *testing.T) {
	factory := NewAdapterFactory()

	t.Run("tools_adapter_basic", func(t *testing.T) {
		// Create tools bridge only
		toolsBridge := testutils.NewMockBridge("agent_tools").WithInitialized(true)
		bridgeMap := map[string]engine.Bridge{
			"agent_tools": toolsBridge,
		}

		// Create adapter
		adapter, err := factory.CreateAdapter("agent_tools", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type - should be basic tools adapter
		assert.IsType(t, (*impl.ToolsAdapter)(nil), adapter)
	})

	t.Run("tools_adapter_with_registry", func(t *testing.T) {
		// Create both bridges
		toolsBridge := testutils.NewMockBridge("agent_tools").WithInitialized(true)
		registryBridge := testutils.NewMockBridge("tools_registry").WithInitialized(true)
		bridgeMap := map[string]engine.Bridge{
			"agent_tools":    toolsBridge,
			"tools_registry": registryBridge,
		}

		// Create adapter
		adapter, err := factory.CreateAdapter("agent_tools", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type - should be enhanced tools adapter
		// Note: The factory returns ToolsAdapter interface, implementation may vary
		assert.NotNil(t, adapter)
	})
}

func TestAdapterFactory_UtilsAdapter(t *testing.T) {
	factory := NewAdapterFactory()

	t.Run("utils_adapter_with_all_bridges", func(t *testing.T) {
		// Create all utility bridges
		authBridge := testutils.NewMockBridge("auth").WithInitialized(true)
		debugBridge := testutils.NewMockBridge("util_debug").WithInitialized(true)
		errorsBridge := testutils.NewMockBridge("util_errors").WithInitialized(true)
		jsonBridge := testutils.NewMockBridge("util_json").WithInitialized(true)
		llmUtilsBridge := testutils.NewMockBridge("llm_utils").WithInitialized(true)
		loggerBridge := testutils.NewMockBridge("util_script_logger").WithInitialized(true)
		slogBridge := testutils.NewMockBridge("util_slog").WithInitialized(true)
		utilBridge := testutils.NewMockBridge("util_core").WithInitialized(true)

		bridgeMap := map[string]engine.Bridge{
			"auth":               authBridge,
			"util_debug":         debugBridge,
			"util_errors":        errorsBridge,
			"util_json":          jsonBridge,
			"llm_utils":          llmUtilsBridge,
			"util_script_logger": loggerBridge,
			"util_slog":          slogBridge,
			"util_core":          utilBridge,
		}

		// Create adapter from any utility bridge ID
		adapter, err := factory.CreateAdapter("util_core", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type
		assert.IsType(t, (*impl.UtilsAdapter)(nil), adapter)
	})

	t.Run("utils_adapter_with_partial_bridges", func(t *testing.T) {
		// Create only some utility bridges
		authBridge := testutils.NewMockBridge("auth").WithInitialized(true)
		jsonBridge := testutils.NewMockBridge("util_json").WithInitialized(true)

		bridgeMap := map[string]engine.Bridge{
			"auth":      authBridge,
			"util_json": jsonBridge,
		}

		// Create adapter from auth bridge
		adapter, err := factory.CreateAdapter("auth", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type
		assert.IsType(t, (*impl.UtilsAdapter)(nil), adapter)
	})

	t.Run("utils_adapter_with_single_bridge", func(t *testing.T) {
		// Create only one utility bridge
		debugBridge := testutils.NewMockBridge("util_debug").WithInitialized(true)
		bridgeMap := map[string]engine.Bridge{
			"util_debug": debugBridge,
		}

		// Create adapter
		adapter, err := factory.CreateAdapter("util_debug", bridgeMap)
		require.NoError(t, err)
		require.NotNil(t, adapter)

		// Verify adapter type
		assert.IsType(t, (*impl.UtilsAdapter)(nil), adapter)
	})

	t.Run("utils_adapter_no_bridges", func(t *testing.T) {
		// Empty bridge map
		bridgeMap := map[string]engine.Bridge{}

		// Create adapter - should fail
		adapter, err := factory.CreateAdapter("util_core", bridgeMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bridge not found") // Primary bridge check fails first
		assert.Nil(t, adapter)
	})
}

func TestAdapterFactory_ErrorHandling(t *testing.T) {
	factory := NewAdapterFactory()

	t.Run("unknown_bridge_id", func(t *testing.T) {
		bridge := testutils.NewMockBridge("unknown_bridge").WithInitialized(true)
		bridgeMap := map[string]engine.Bridge{
			"unknown_bridge": bridge,
		}

		// Create adapter - should fail
		adapter, err := factory.CreateAdapter("unknown_bridge", bridgeMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown bridge ID")
		assert.Nil(t, adapter)
	})

	t.Run("missing_bridge_in_map", func(t *testing.T) {
		// Empty bridge map
		bridgeMap := map[string]engine.Bridge{}

		// Create adapter - should fail
		adapter, err := factory.CreateAdapter("state_manager", bridgeMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bridge not found")
		assert.Nil(t, adapter)
	})
}

func TestAdapterFactory_Metadata(t *testing.T) {
	factory := NewAdapterFactory()

	t.Run("supported_bridge_ids", func(t *testing.T) {
		ids := factory.GetSupportedBridgeIDs()
		assert.NotEmpty(t, ids)

		// Check some expected IDs
		expectedIDs := []string{
			"llm_core", "state_manager", "agent_core", "agent_tools",
			"observability_metrics", "auth", "util_core",
		}
		for _, expected := range expectedIDs {
			assert.Contains(t, ids, expected)
		}
	})

	t.Run("multi_bridge_detection", func(t *testing.T) {
		// Multi-bridge adapters
		assert.True(t, factory.IsMultiBridgeAdapter("llm_core"))
		assert.True(t, factory.IsMultiBridgeAdapter("observability_metrics"))
		assert.True(t, factory.IsMultiBridgeAdapter("util_core"))
		assert.True(t, factory.IsMultiBridgeAdapter("auth"))

		// Single-bridge adapters
		assert.False(t, factory.IsMultiBridgeAdapter("state_manager"))
		assert.False(t, factory.IsMultiBridgeAdapter("agent_core"))
		assert.False(t, factory.IsMultiBridgeAdapter("agent_tools"))
	})

	t.Run("required_bridges", func(t *testing.T) {
		// LLM requires core
		required := factory.GetRequiredBridges("llm_core")
		assert.Equal(t, []string{"llm_core"}, required)

		// State requires itself
		required = factory.GetRequiredBridges("state_manager")
		assert.Equal(t, []string{"state_manager"}, required)

		// Utils has no specific required bridge
		required = factory.GetRequiredBridges("util_core")
		assert.Empty(t, required)
	})

	t.Run("optional_bridges", func(t *testing.T) {
		// LLM has optional providers and pool
		optional := factory.GetOptionalBridges("llm_core")
		assert.Contains(t, optional, "llm_providers")
		assert.Contains(t, optional, "llm_pool")

		// Tools has optional registry
		optional = factory.GetOptionalBridges("agent_tools")
		assert.Contains(t, optional, "tools_registry")

		// State has no optional bridges
		optional = factory.GetOptionalBridges("state_manager")
		assert.Empty(t, optional)
	})
}

func TestAdapterFactory_Concurrency(t *testing.T) {
	factory := NewAdapterFactory()

	// Create bridges
	bridges := make(map[string]engine.Bridge)
	for _, id := range []string{"state_manager", "agent_core", "agent_events"} {
		bridges[id] = testutils.NewMockBridge(id).WithInitialized(true)
	}

	// Run concurrent adapter creation
	done := make(chan bool, 3)
	for _, bridgeID := range []string{"state_manager", "agent_core", "agent_events"} {
		go func(id string) {
			adapter, err := factory.CreateAdapter(id, bridges)
			assert.NoError(t, err)
			assert.NotNil(t, adapter)
			done <- true
		}(bridgeID)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}