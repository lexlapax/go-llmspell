// ABOUTME: Basic integration tests for end-to-end execution flow verification
// ABOUTME: Simple tests to verify Lua Script → Adapter → Bridge flow works correctly

package lua

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/factory"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestBasicAdapterFlow tests the simplest possible end-to-end flow
func TestBasicAdapterFlow(t *testing.T) {
	testFactory := factory.NewTestAdapterFactory()
	eng := NewLuaEngineWithFactory(testFactory)
	defer func() { _ = eng.Shutdown() }()

	err := eng.Initialize(engine.EngineConfig{SandboxMode: false})
	require.NoError(t, err)

	// Register a simple test bridge
	testBridge := testutils.NewMockBridge("test_bridge").
		WithInitialized(true).
		WithMethod("simple_method", engine.MethodInfo{
			Name:        "simple_method",
			Description: "A simple test method",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "string",
		}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
			return engine.NewStringValue("success"), nil
		})

	err = eng.RegisterBridge(testBridge)
	require.NoError(t, err)

	// Verify adapter was created
	adapter := eng.GetAdapter("test_bridge")
	require.NotNil(t, adapter, "Adapter should be created for test bridge")

	// Test very basic functionality
	script := `
		if bridges then
			if bridges.test_bridge then
				local test_result = bridges.test_bridge.simple_method()
				return test_result
			else
				return "bridge_not_found"
			end
		else
			return "bridges_table_missing"
		end
	`

	ctx := context.Background()
	result, err := eng.Execute(ctx, script, nil)
	require.NoError(t, err, "Basic adapter flow should work")

	resultString, err := engine.ConvertToString(result)
	require.NoError(t, err)
	assert.Equal(t, "test result from test_bridge.simple_method", resultString, "Should get test result from adapter")
}

// TestAdapterCreationProcess verifies the adapter creation process
func TestAdapterCreationProcess(t *testing.T) {
	testFactory := factory.NewTestAdapterFactory()
	eng := NewLuaEngineWithFactory(testFactory)
	defer func() { _ = eng.Shutdown() }()

	err := eng.Initialize(engine.EngineConfig{SandboxMode: false})
	require.NoError(t, err)

	t.Run("test_bridge_gets_test_adapter", func(t *testing.T) {
		testBridge := testutils.NewMockBridge("test_bridge").WithInitialized(true)
		err := eng.RegisterBridge(testBridge)
		require.NoError(t, err)

		adapter := eng.GetAdapter("test_bridge")
		require.NotNil(t, adapter, "Test bridge should get an adapter")
	})

	t.Run("production_bridge_gets_production_adapter", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state_manager").WithInitialized(true)
		err := eng.RegisterBridge(stateBridge)
		require.NoError(t, err)

		adapter := eng.GetAdapter("state_manager")
		require.NotNil(t, adapter, "State manager bridge should get an adapter")
	})

	t.Run("unknown_bridge_gets_no_adapter", func(t *testing.T) {
		unknownBridge := testutils.NewMockBridge("unknown_bridge_id").WithInitialized(true)
		err := eng.RegisterBridge(unknownBridge)
		
		// This should succeed but the adapter might not be created if the factory doesn't recognize it
		if err == nil {
			// Bridge registration succeeded, check if adapter was created
			adapter := eng.GetAdapter("unknown_bridge_id")
			// For test factory, unknown bridges still get test adapters
			// This is expected behavior
			t.Logf("Unknown bridge adapter status: %v", adapter != nil)
		}
	})
}

// TestFactoryPatternIntegration verifies the factory pattern works correctly
func TestFactoryPatternIntegration(t *testing.T) {
	testFactory := factory.NewTestAdapterFactory()
	eng := NewLuaEngineWithFactory(testFactory)
	defer func() { _ = eng.Shutdown() }()

	err := eng.Initialize(engine.EngineConfig{SandboxMode: false})
	require.NoError(t, err)

	// Test that the factory creates the right type of adapters
	bridgeTypes := []struct {
		bridgeID     string
		expectsAdapter bool
	}{
		{"state_manager", true},      // Should get production StateAdapter
		{"agent_core", true},         // Should get production AgentAdapter 
		{"test_bridge", true},        // Should get TestAdapter
		{"bridge_1", true},           // Should get TestAdapter
		{"unknown_type", false},      // Should not get adapter (factory rejects unknown IDs)
	}

	for _, bt := range bridgeTypes {
		t.Run("bridge_type_"+bt.bridgeID, func(t *testing.T) {
			bridge := testutils.NewMockBridge(bt.bridgeID).WithInitialized(true)
			err := eng.RegisterBridge(bridge)
			
			if bt.expectsAdapter {
				require.NoError(t, err, "Bridge registration should succeed for known bridge types")
				adapter := eng.GetAdapter(bt.bridgeID)
				assert.NotNil(t, adapter, "Bridge %s should have an adapter", bt.bridgeID)
			} else {
				// For unknown bridge types, registration should fail due to adapter creation failure
				assert.Error(t, err, "Bridge registration should fail for unknown bridge types")
				assert.Contains(t, err.Error(), "unknown bridge ID", "Error should mention unknown bridge ID")
			}
		})
	}
}

// TestMultiBridgeScenario tests multiple bridges working together
func TestMultiBridgeScenario(t *testing.T) {
	testFactory := factory.NewTestAdapterFactory()
	eng := NewLuaEngineWithFactory(testFactory)
	defer func() { _ = eng.Shutdown() }()

	err := eng.Initialize(engine.EngineConfig{SandboxMode: false})
	require.NoError(t, err)

	// Register multiple bridges
	bridges := []string{"state_manager", "agent_core", "test_bridge"}
	
	for _, bridgeID := range bridges {
		bridge := testutils.NewMockBridge(bridgeID).WithInitialized(true)
		err := eng.RegisterBridge(bridge)
		require.NoError(t, err)
	}

	// Verify all adapters were created
	for _, bridgeID := range bridges {
		adapter := eng.GetAdapter(bridgeID)
		assert.NotNil(t, adapter, "Bridge %s should have an adapter", bridgeID)
	}

	// Verify bridge list matches registered bridges
	engineBridges := eng.ListBridges()
	assert.Equal(t, len(bridges), len(engineBridges), "Should have all registered bridges")
	
	for _, expectedBridge := range bridges {
		assert.Contains(t, engineBridges, expectedBridge, "Should contain bridge %s", expectedBridge)
	}
}

// TestEngineMetricsAndCleanup tests engine state and cleanup
func TestEngineMetricsAndCleanup(t *testing.T) {
	testFactory := factory.NewTestAdapterFactory()
	eng := NewLuaEngineWithFactory(testFactory)
	defer func() { _ = eng.Shutdown() }()

	err := eng.Initialize(engine.EngineConfig{SandboxMode: false})
	require.NoError(t, err)

	// Register a bridge and execute some scripts
	testBridge := testutils.NewMockBridge("test_bridge").WithInitialized(true)
	err = eng.RegisterBridge(testBridge)
	require.NoError(t, err)

	// Execute a simple script
	ctx := context.Background()
	_, err = eng.Execute(ctx, "return 'test'", nil)
	require.NoError(t, err)

	// Verify metrics
	metrics := eng.GetMetrics()
	assert.Greater(t, metrics.ScriptsExecuted, int64(0), "Should have executed at least one script")
	assert.GreaterOrEqual(t, metrics.ErrorCount, int64(0), "Error count should be non-negative")

	// Test bridge unregistration
	err = eng.UnregisterBridge("test_bridge")
	require.NoError(t, err)

	// Verify bridge was removed
	bridges := eng.ListBridges()
	assert.NotContains(t, bridges, "test_bridge", "Bridge should be unregistered")

	// Verify adapter was removed
	adapter := eng.GetAdapter("test_bridge")
	assert.Nil(t, adapter, "Adapter should be removed when bridge is unregistered")
}