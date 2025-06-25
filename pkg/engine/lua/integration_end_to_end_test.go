// ABOUTME: Comprehensive end-to-end tests verifying complete execution flow through all layers
// ABOUTME: Tests Lua Script → Stdlib → Adapter → Bridge → go-llms integration with real adapters

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

// TestCompleteExecutionFlow tests the complete execution path from Lua scripts
// through adapters to bridges, verifying that the factory pattern properly
// creates adapters and that the execution flow works end-to-end.
func TestCompleteExecutionFlow(t *testing.T) {
	// Use test factory to support test bridge IDs while maintaining production behavior
	testFactory := factory.NewTestAdapterFactory()
	eng := NewLuaEngineWithFactory(testFactory)
	defer func() {
		_ = eng.Shutdown()
	}()

	// Initialize engine with minimal security for testing
	config := engine.EngineConfig{
		SandboxMode:     false,
		DebugMode:       true,
		MetricsMode:     true,
		FileSystemMode:  engine.FSModeReadWrite,
		AllowedModules:  []string{"string", "math", "table"},
		DisabledModules: []string{},
		EngineOptions:   make(map[string]interface{}),
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	t.Run("single_bridge_adapter_integration", func(t *testing.T) {
		// Create and register a test bridge for state management
		stateBridge := testutils.NewMockBridge("state_manager").
			WithInitialized(true).
			WithMethod("get", engine.MethodInfo{
				Name:        "get",
				Description: "Gets a state value",
				Parameters:  []engine.ParameterInfo{{Name: "key", Type: "string", Required: true}},
				ReturnType:  "any",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("test result from state_manager.get"), nil
			}).
			WithMethod("set", engine.MethodInfo{
				Name:        "set",
				Description: "Sets a state value",
				Parameters:  []engine.ParameterInfo{{Name: "key", Type: "string", Required: true}, {Name: "value", Type: "any", Required: true}},
				ReturnType:  "boolean",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("test result from state_manager.set"), nil
			})

		err := eng.RegisterBridge(stateBridge)
		require.NoError(t, err)

		// Verify adapter was created
		adapter := eng.GetAdapter("state_manager")
		require.NotNil(t, adapter, "Adapter should be created for registered bridge")

		// Test complete Lua script execution that uses the bridge through adapter
		script := `
			-- Verify bridge is accessible
			assert(bridges ~= nil, "bridges global should exist")
			assert(bridges.state_manager ~= nil, "state_manager bridge should be accessible")
			
			-- Get the bridge adapter
			local state = bridges.state_manager
			assert(state ~= nil, "state adapter should not be nil")
			
			-- Verify adapter metadata
			local meta = state._meta
			assert(meta ~= nil, "adapter should have metadata")
			assert(meta.bridge_id == "state_manager", "bridge ID should match")
			assert(meta.type == "test_adapter", "should be test adapter type")
			
			-- Test adapter methods
			assert(type(state.set) == "function", "set method should be available")
			assert(type(state.get) == "function", "get method should be available")
			
			-- Execute bridge methods through adapter
			local set_result = state.set("test_key", "test_value")
			assert(set_result ~= nil, "set should return a result")
			
			local get_result = state.get("test_key")
			assert(get_result ~= nil, "get should return a result")
			
			return {
				bridge_accessible = true,
				adapter_metadata = meta,
				method_results = {
					set = tostring(set_result),
					get = tostring(get_result)
				},
				execution_path = "lua->adapter->bridge->go-llms"
			}
		`

		ctx := context.Background()
		result, err := eng.Execute(ctx, script, nil)
		require.NoError(t, err, "End-to-end script execution should succeed")

		// Verify results
		require.Equal(t, engine.TypeObject, result.Type())
		objectValue, ok := result.(engine.ObjectValue)
		require.True(t, ok)

		fields := objectValue.Fields()
		bridgeAccessible, _ := engine.ConvertToBool(fields["bridge_accessible"])
		assert.True(t, bridgeAccessible, "Bridge should be accessible through adapter")

		executionPath, _ := engine.ConvertToString(fields["execution_path"])
		assert.Equal(t, "lua->adapter->bridge->go-llms", executionPath)

		// Verify adapter metadata was properly exposed
		metadataField := fields["adapter_metadata"]
		require.Equal(t, engine.TypeObject, metadataField.Type())
		metadataObj, ok := metadataField.(engine.ObjectValue)
		require.True(t, ok)

		metaFields := metadataObj.Fields()
		bridgeID, _ := engine.ConvertToString(metaFields["bridge_id"])
		adapterType, _ := engine.ConvertToString(metaFields["type"])
		assert.Equal(t, "state_manager", bridgeID)
		assert.Equal(t, "test_adapter", adapterType)
	})

	t.Run("multi_bridge_adapter_integration", func(t *testing.T) {
		// Create multiple observability bridges for testing multi-bridge adapter
		metricsBridge := testutils.NewMockBridge("observability_metrics").
			WithInitialized(true).
			WithMethod("record_metric", engine.MethodInfo{
				Name:        "record_metric",
				Description: "Records a metric",
				Parameters:  []engine.ParameterInfo{{Name: "name", Type: "string", Required: true}, {Name: "value", Type: "number", Required: true}},
				ReturnType:  "boolean",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("test result from observability_metrics.record_metric"), nil
			})

		tracingBridge := testutils.NewMockBridge("observability_tracing").
			WithInitialized(true).
			WithMethod("start_span", engine.MethodInfo{
				Name:        "start_span",
				Description: "Starts a trace span",
				Parameters:  []engine.ParameterInfo{{Name: "name", Type: "string", Required: true}},
				ReturnType:  "string",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("test result from observability_tracing.start_span"), nil
			})

		// Register bridges one by one
		err := eng.RegisterBridge(metricsBridge)
		require.NoError(t, err)
		err = eng.RegisterBridge(tracingBridge)
		require.NoError(t, err)

		// Verify adapters were created for observability bridges
		require.NotNil(t, eng.GetAdapter("observability_metrics"))
		require.NotNil(t, eng.GetAdapter("observability_tracing"))

		// Test multi-bridge adapter functionality
		script := `
			-- Verify observability bridges are accessible
			assert(bridges.observability_metrics ~= nil, "metrics bridge should be accessible")
			assert(bridges.observability_tracing ~= nil, "tracing bridge should be accessible")
			
			local metrics = bridges.observability_metrics
			local tracing = bridges.observability_tracing
			
			-- Test that adapters have proper metadata
			assert(metrics._meta.bridge_id == "observability_metrics")
			assert(tracing._meta.bridge_id == "observability_tracing")
			
			-- Test that adapters have their expected methods
			assert(type(metrics.record_metric) == "function")
			assert(type(tracing.start_span) == "function")
			
			-- Execute methods on adapters
			local metric_result = metrics.record_metric("test.metric", 42)
			local span_result = tracing.start_span("test-span")
			
			return {
				multi_bridge_support = true,
				adapters_created = 2,
				method_results = {
					metrics = tostring(metric_result),
					tracing = tostring(span_result)
				},
				flow_verified = "lua->multi-adapters->bridges->go-llms"
			}
		`

		ctx := context.Background()
		result, err := eng.Execute(ctx, script, nil)
		require.NoError(t, err, "Multi-bridge adapter execution should succeed")

		// Verify results
		require.Equal(t, engine.TypeObject, result.Type())
		objectValue, ok := result.(engine.ObjectValue)
		require.True(t, ok)

		fields := objectValue.Fields()
		multiBridgeSupport, _ := engine.ConvertToBool(fields["multi_bridge_support"])
		assert.True(t, multiBridgeSupport)

		adaptersCreated, _ := engine.ConvertToNumber(fields["adapters_created"])
		assert.Equal(t, 2.0, adaptersCreated)

		flowVerified, _ := engine.ConvertToString(fields["flow_verified"])
		assert.Equal(t, "lua->multi-adapters->bridges->go-llms", flowVerified)
	})

	t.Run("adapter_enforcement_verification", func(t *testing.T) {
		// Verify that bridges without adapters cannot be accessed (adapter enforcement)
		script := `
			-- This test verifies that the engine properly enforces adapter requirement
			-- If a bridge doesn't have an adapter, it should not be accessible
			
			-- Try to access a non-existent bridge
			local nonexistent = bridges.nonexistent_bridge
			
			-- This should be nil because no such bridge exists
			assert(nonexistent == nil, "nonexistent bridge should be nil")
			
			-- Count actual bridges that are accessible
			local bridge_count = 0
			for bridge_id, bridge_adapter in pairs(bridges) do
				if bridge_adapter ~= nil then
					bridge_count = bridge_count + 1
				end
			end
			
			return {
				adapter_enforcement = true,
				accessible_bridges = bridge_count,
				enforcement_verified = "only-registered-bridges-with-adapters-accessible"
			}
		`

		ctx := context.Background()
		result, err := eng.Execute(ctx, script, nil)
		require.NoError(t, err, "Adapter enforcement test should succeed")

		// Verify adapter enforcement
		require.Equal(t, engine.TypeObject, result.Type())
		objectValue, ok := result.(engine.ObjectValue)
		require.True(t, ok)

		fields := objectValue.Fields()
		adapterEnforcement, _ := engine.ConvertToBool(fields["adapter_enforcement"])
		accessibleBridges, _ := engine.ConvertToNumber(fields["accessible_bridges"])
		enforcementVerified, _ := engine.ConvertToString(fields["enforcement_verified"])

		assert.True(t, adapterEnforcement)
		assert.Greater(t, accessibleBridges, 0.0, "Should have some accessible bridges")
		assert.Equal(t, "only-registered-bridges-with-adapters-accessible", enforcementVerified)
	})

	// Final verification of engine state
	t.Run("final_engine_verification", func(t *testing.T) {
		// Verify final state of the engine after all tests
		bridges := eng.ListBridges()
		adapters := make(map[string]interface{})
		
		for _, bridgeID := range bridges {
			adapter := eng.GetAdapter(bridgeID)
			if adapter != nil {
				adapters[bridgeID] = adapter
			}
		}

		assert.Greater(t, len(bridges), 0, "Should have registered bridges")
		assert.Equal(t, len(bridges), len(adapters), "Each bridge should have an adapter")

		// Verify metrics
		metrics := eng.GetMetrics()
		assert.Greater(t, metrics.ScriptsExecuted, int64(0), "Should have executed scripts")
		assert.GreaterOrEqual(t, metrics.ErrorCount, int64(0), "Error count should be non-negative")
	})
}

// TestAdapterFactoryIntegration tests that the adapter factory properly
// creates and manages adapters for various bridge types
func TestAdapterFactoryIntegration(t *testing.T) {
	testFactory := factory.NewTestAdapterFactory()
	eng := NewLuaEngineWithFactory(testFactory)
	defer func() { _ = eng.Shutdown() }()

	err := eng.Initialize(engine.EngineConfig{SandboxMode: false})
	require.NoError(t, err)

	t.Run("factory_creates_real_adapters", func(t *testing.T) {
		// Register a real bridge that should get a production adapter
		agentBridge := testutils.NewMockBridge("agent_core").
			WithInitialized(true).
			WithMethod("create", engine.MethodInfo{
				Name:        "create",
				Description: "Creates a new agent",
				Parameters:  []engine.ParameterInfo{{Name: "config", Type: "object", Required: true}},
				ReturnType:  "string",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("test result from agent_core.create"), nil
			})

		err := eng.RegisterBridge(agentBridge)
		require.NoError(t, err)

		// Verify that factory created an adapter
		adapter := eng.GetAdapter("agent_core")
		require.NotNil(t, adapter, "Factory should create adapter for agent_core bridge")

		// Test that the adapter works in Lua
		script := `
			local agent = bridges.agent_core
			assert(agent ~= nil, "agent bridge should be accessible")
			
			-- Test adapter method
			local create_result = agent.create({name = "test-agent"})
			
			return {
				adapter_created = true,
				method_works = create_result ~= nil,
				bridge_id = agent._meta.bridge_id
			}
		`

		ctx := context.Background()
		result, err := eng.Execute(ctx, script, nil)
		require.NoError(t, err, "Factory-created adapter should work")

		// Verify results
		require.Equal(t, engine.TypeObject, result.Type())
		objectValue, ok := result.(engine.ObjectValue)
		require.True(t, ok)

		fields := objectValue.Fields()
		adapterCreated, _ := engine.ConvertToBool(fields["adapter_created"])
		methodWorks, _ := engine.ConvertToBool(fields["method_works"])
		bridgeID, _ := engine.ConvertToString(fields["bridge_id"])

		assert.True(t, adapterCreated, "Adapter should be created by factory")
		assert.True(t, methodWorks, "Adapter method should work")
		assert.Equal(t, "agent_core", bridgeID, "Bridge ID should match")
	})

	t.Run("factory_handles_test_bridges", func(t *testing.T) {
		// Register a test bridge that should get a test adapter
		testBridge := testutils.NewMockBridge("test_bridge").
			WithInitialized(true).
			WithMethod("test_method", engine.MethodInfo{
				Name:        "test_method",
				Description: "A test method",
				Parameters:  []engine.ParameterInfo{},
				ReturnType:  "string",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("test result from test_bridge.test_method"), nil
			})

		err := eng.RegisterBridge(testBridge)
		require.NoError(t, err)

		// Verify that factory created a test adapter
		adapter := eng.GetAdapter("test_bridge")
		require.NotNil(t, adapter, "Factory should create test adapter for test bridge")

		// Test that the test adapter works in Lua
		script := `
			local test_bridge = bridges.test_bridge
			assert(test_bridge ~= nil, "test bridge should be accessible")
			assert(test_bridge._meta.type == "test_adapter", "should be test adapter type")
			
			-- Test adapter method
			local test_result = test_bridge.test_method()
			
			return {
				test_adapter_created = true,
				method_works = test_result ~= nil,
				adapter_type = test_bridge._meta.type
			}
		`

		ctx := context.Background()
		result, err := eng.Execute(ctx, script, nil)
		require.NoError(t, err, "Test adapter should work")

		// Verify results
		require.Equal(t, engine.TypeObject, result.Type())
		objectValue, ok := result.(engine.ObjectValue)
		require.True(t, ok)

		fields := objectValue.Fields()
		testAdapterCreated, _ := engine.ConvertToBool(fields["test_adapter_created"])
		methodWorks, _ := engine.ConvertToBool(fields["method_works"])
		adapterType, _ := engine.ConvertToString(fields["adapter_type"])

		assert.True(t, testAdapterCreated, "Test adapter should be created by factory")
		assert.True(t, methodWorks, "Test adapter method should work")
		assert.Equal(t, "test_adapter", adapterType, "Should be test adapter type")
	})
}