// ABOUTME: Tests for bridge registry functionality including bridge set creation and profile management
// ABOUTME: Comprehensive test coverage for bridge factory functions, registration, and error handling

package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestBridgeSetConstants(t *testing.T) {
	tests := []struct {
		name      string
		bridgeSet BridgeSet
		expected  string
	}{
		{"Core bridge set", BridgeSetCore, "core"},
		{"LLM bridge set", BridgeSetLLM, "llm"},
		{"Utility bridge set", BridgeSetUtility, "utility"},
		{"Agent bridge set", BridgeSetAgent, "agent"},
		{"Observability bridge set", BridgeSetObservability, "observability"},
		{"State bridge set", BridgeSetState, "state"},
		{"Structured bridge set", BridgeSetStructured, "structured"},
		{"All bridge set", BridgeSetAll, "all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.bridgeSet))
		})
	}
}

func TestFeatureSetBridgeSets(t *testing.T) {
	tests := []struct {
		name         string
		featureSet   FeatureSet
		expectedSets []BridgeSet
		minSetCount  int
	}{
		{
			name:         "Full feature set",
			featureSet:   FeatureSetFull,
			expectedSets: []BridgeSet{BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetAgent, BridgeSetObservability, BridgeSetState, BridgeSetStructured},
			minSetCount:  7,
		},
		{
			name:         "Minimal feature set",
			featureSet:   FeatureSetMinimal,
			expectedSets: []BridgeSet{BridgeSetCore, BridgeSetUtility},
			minSetCount:  2,
		},
		{
			name:         "LLM feature set",
			featureSet:   FeatureSetLLM,
			expectedSets: []BridgeSet{BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetStructured},
			minSetCount:  4,
		},
		{
			name:         "Agent feature set",
			featureSet:   FeatureSetAgent,
			expectedSets: []BridgeSet{BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetStructured, BridgeSetAgent, BridgeSetState},
			minSetCount:  6,
		},
		{
			name:         "Observable feature set",
			featureSet:   FeatureSetObservable,
			expectedSets: []BridgeSet{BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetObservability},
			minSetCount:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridgeSets := GetBridgeSetsForFeature(tt.featureSet)
			assert.Len(t, bridgeSets, tt.minSetCount)

			// Check that expected sets are present
			for _, expectedSet := range tt.expectedSets {
				assert.Contains(t, bridgeSets, expectedSet)
			}
		})
	}
}

func TestCreateCoreBridges(t *testing.T) {
	bridges, err := createCoreBridges()

	require.NoError(t, err)
	require.NotEmpty(t, bridges)

	// Should contain modelinfo bridge
	found := false
	for _, bridge := range bridges {
		if bridge.GetID() == "llm_modelinfo" {
			found = true
			break
		}
	}
	assert.True(t, found, "Core bridges should include modelinfo bridge")
}

func TestCreateLLMBridges(t *testing.T) {
	bridges, err := createLLMBridges()

	require.NoError(t, err)
	require.NotEmpty(t, bridges)
	assert.GreaterOrEqual(t, len(bridges), 3, "Should have at least 3 LLM bridges")

	// Check for expected bridge IDs
	bridgeIDs := make(map[string]bool)
	for _, bridge := range bridges {
		bridgeIDs[bridge.GetID()] = true
	}

	assert.True(t, bridgeIDs["llm_core"], "Should have llm bridge")
	assert.True(t, bridgeIDs["llm_providers"], "Should have providers bridge")
	assert.True(t, bridgeIDs["llm_pool"], "Should have pool bridge")
}

func TestCreateUtilityBridges(t *testing.T) {
	bridges, err := createUtilityBridges()

	require.NoError(t, err)
	require.NotEmpty(t, bridges)
	assert.GreaterOrEqual(t, len(bridges), 5, "Should have multiple utility bridges")

	// Check that bridges have valid IDs
	for _, bridge := range bridges {
		assert.NotEmpty(t, bridge.GetID(), "Each bridge should have a non-empty ID")
	}
}

func TestCreateAgentBridges(t *testing.T) {
	bridges, err := createAgentBridges()

	require.NoError(t, err)
	require.NotEmpty(t, bridges)
	assert.GreaterOrEqual(t, len(bridges), 5, "Should have multiple agent bridges")

	// Check for expected bridge IDs
	bridgeIDs := make(map[string]bool)
	for _, bridge := range bridges {
		bridgeIDs[bridge.GetID()] = true
	}

	assert.True(t, bridgeIDs["agent_core"], "Should have agent bridge")
	assert.True(t, bridgeIDs["agent_events"], "Should have events bridge")
	assert.True(t, bridgeIDs["agent_tools"], "Should have tools bridge")
}

func TestCreateObservabilityBridges(t *testing.T) {
	bridges, err := createObservabilityBridges()

	require.NoError(t, err)
	require.NotEmpty(t, bridges)
	assert.GreaterOrEqual(t, len(bridges), 3, "Should have multiple observability bridges")

	// Check for expected bridge IDs
	bridgeIDs := make(map[string]bool)
	for _, bridge := range bridges {
		bridgeIDs[bridge.GetID()] = true
	}

	assert.True(t, bridgeIDs["observability_metrics"], "Should have metrics bridge")
	assert.True(t, bridgeIDs["observability_tracing"], "Should have tracing bridge")
	assert.True(t, bridgeIDs["observability_guardrails"], "Should have guardrails bridge")
}

func TestCreateStateBridges(t *testing.T) {
	bridges, err := createStateBridges()

	require.NoError(t, err)
	require.NotEmpty(t, bridges)

	// Check for state context bridge
	found := false
	for _, bridge := range bridges {
		if bridge.GetID() == "state_context" {
			found = true
			break
		}
	}
	assert.True(t, found, "State bridges should include state_context bridge")
}

func TestCreateStructuredBridges(t *testing.T) {
	bridges, err := createStructuredBridges()

	require.NoError(t, err)
	require.NotEmpty(t, bridges)

	// Check for schema bridge
	found := false
	for _, bridge := range bridges {
		if bridge.GetID() == "structured_schema" {
			found = true
			break
		}
	}
	assert.True(t, found, "Structured bridges should include schema bridge")
}

func TestRegisterBridgeSets(t *testing.T) {
	tests := []struct {
		name        string
		bridgeSets  []BridgeSet
		expectError bool
		description string
	}{
		{
			name:        "Register core bridges",
			bridgeSets:  []BridgeSet{BridgeSetCore},
			expectError: false,
			description: "Should successfully register core bridges",
		},
		{
			name:        "Register multiple bridge sets",
			bridgeSets:  []BridgeSet{BridgeSetCore, BridgeSetUtility},
			expectError: false,
			description: "Should successfully register multiple bridge sets",
		},
		{
			name:        "Register all bridges",
			bridgeSets:  []BridgeSet{BridgeSetAll},
			expectError: false,
			description: "Should successfully register all bridge sets",
		},
		{
			name:        "Register unknown bridge set",
			bridgeSets:  []BridgeSet{"unknown"},
			expectError: true,
			description: "Should fail when registering unknown bridge set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := testutils.NewMockScriptEngine()
			err := mockEngine.Initialize(types.EngineConfig{})
			require.NoError(t, err)

			err = RegisterBridgeSets(mockEngine, tt.bridgeSets)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)

				// Verify bridges were registered
				bridges := mockEngine.ListBridges()
				assert.NotEmpty(t, bridges, "Should have registered bridges")
			}
		})
	}
}

func TestRegisterBridgesByFeatureSet(t *testing.T) {
	tests := []struct {
		name        string
		featureSet  FeatureSet
		expectError bool
	}{
		{
			name:        "Register full feature set",
			featureSet:  FeatureSetFull,
			expectError: false,
		},
		{
			name:        "Register minimal feature set",
			featureSet:  FeatureSetMinimal,
			expectError: false,
		},
		{
			name:        "Register LLM feature set",
			featureSet:  FeatureSetLLM,
			expectError: false,
		},
		{
			name:        "Register agent feature set",
			featureSet:  FeatureSetAgent,
			expectError: false,
		},
		{
			name:        "Register observable feature set",
			featureSet:  FeatureSetObservable,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := testutils.NewMockScriptEngine()
			err := mockEngine.Initialize(types.EngineConfig{})
			require.NoError(t, err)

			err = RegisterBridgesByFeatureSet(mockEngine, tt.featureSet)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify bridges were registered
				bridges := mockEngine.ListBridges()
				assert.NotEmpty(t, bridges, "Should have registered bridges")
			}
		})
	}
}

func TestRegisterStandardBridges(t *testing.T) {
	mockEngine := testutils.NewMockScriptEngine()
	err := mockEngine.Initialize(types.EngineConfig{})
	require.NoError(t, err)

	err = RegisterStandardBridges(mockEngine)
	assert.NoError(t, err)

	// Verify bridges were registered
	bridges := mockEngine.ListBridges()
	assert.NotEmpty(t, bridges, "Should have registered standard bridges")

	// Should have bridges from all sets in full feature set
	assert.GreaterOrEqual(t, len(bridges), 10, "Full feature set should register many bridges")
}

func TestRegisterBridgesWithRegistry(t *testing.T) {
	tests := []struct {
		name        string
		featureSet  FeatureSet
		expectError bool
	}{
		{
			name:        "Register with full feature set",
			featureSet:  FeatureSetFull,
			expectError: false,
		},
		{
			name:        "Register with minimal feature set",
			featureSet:  FeatureSetMinimal,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mock engine directly without registry
			mockEngine := testutils.NewMockScriptEngine()
			err := mockEngine.Initialize(types.EngineConfig{})
			require.NoError(t, err)

			err = RegisterBridgesByFeatureSet(mockEngine, tt.featureSet)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify bridges were registered
				bridges := mockEngine.ListBridges()
				assert.NotEmpty(t, bridges, "Should have registered bridges")
			}
		})
	}
}

func TestRegisterBridgesWithAllEngines(t *testing.T) {
	// Test with multiple engines directly
	mockEngine1 := testutils.NewMockScriptEngine()
	err := mockEngine1.Initialize(types.EngineConfig{})
	require.NoError(t, err)

	mockEngine2 := testutils.NewMockScriptEngine()
	err = mockEngine2.Initialize(types.EngineConfig{})
	require.NoError(t, err)

	// Register with specific feature sets per engine
	err = RegisterBridgesByFeatureSet(mockEngine1, FeatureSetMinimal)
	assert.NoError(t, err)

	err = RegisterBridgesByFeatureSet(mockEngine2, FeatureSetFull)
	assert.NoError(t, err)

	// Verify both engines have bridges registered
	bridges1 := mockEngine1.ListBridges()
	bridges2 := mockEngine2.ListBridges()

	assert.NotEmpty(t, bridges1, "Engine1 should have bridges")
	assert.NotEmpty(t, bridges2, "Engine2 should have bridges")

	// Engine2 (full) should have more bridges than Engine1 (minimal)
	assert.Greater(t, len(bridges2), len(bridges1), "Full feature set should have more bridges than minimal")
}

func TestBridgeFactoryFunctions(t *testing.T) {
	factoryTests := []struct {
		name        string
		factory     func() ([]types.Bridge, error)
		minBridges  int
		description string
	}{
		{
			name:        "Core bridge factory",
			factory:     createCoreBridges,
			minBridges:  1,
			description: "Should create at least 1 core bridge",
		},
		{
			name:        "LLM bridge factory",
			factory:     createLLMBridges,
			minBridges:  3,
			description: "Should create at least 3 LLM bridges",
		},
		{
			name:        "Utility bridge factory",
			factory:     createUtilityBridges,
			minBridges:  5,
			description: "Should create at least 5 utility bridges",
		},
		{
			name:        "Agent bridge factory",
			factory:     createAgentBridges,
			minBridges:  5,
			description: "Should create at least 5 agent bridges",
		},
		{
			name:        "Observability bridge factory",
			factory:     createObservabilityBridges,
			minBridges:  3,
			description: "Should create at least 3 observability bridges",
		},
		{
			name:        "State bridge factory",
			factory:     createStateBridges,
			minBridges:  1,
			description: "Should create at least 1 state bridge",
		},
		{
			name:        "Structured bridge factory",
			factory:     createStructuredBridges,
			minBridges:  1,
			description: "Should create at least 1 structured bridge",
		},
	}

	for _, tt := range factoryTests {
		t.Run(tt.name, func(t *testing.T) {
			bridges, err := tt.factory()

			assert.NoError(t, err, "Factory should not return error")
			assert.GreaterOrEqual(t, len(bridges), tt.minBridges, tt.description)

			// Verify all bridges have valid IDs and metadata
			bridgeIDs := make(map[string]bool)
			for _, bridge := range bridges {
				id := bridge.GetID()
				assert.NotEmpty(t, id, "Bridge should have non-empty ID")
				assert.False(t, bridgeIDs[id], "Bridge IDs should be unique within a set")
				bridgeIDs[id] = true

				metadata := bridge.GetMetadata()
				assert.NotEmpty(t, metadata.Name, "Bridge should have metadata name")
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	t.Run("Uninitialized engine registration", func(t *testing.T) {
		mockEngine := testutils.NewMockScriptEngine()
		// Don't initialize the engine

		err := RegisterBridgeSets(mockEngine, []BridgeSet{BridgeSetCore})
		assert.Error(t, err, "Should fail to register bridges on uninitialized engine")
	})

	t.Run("Empty bridge sets", func(t *testing.T) {
		mockEngine := testutils.NewMockScriptEngine()
		err := mockEngine.Initialize(types.EngineConfig{})
		require.NoError(t, err)

		err = RegisterBridgeSets(mockEngine, []BridgeSet{})
		assert.NoError(t, err, "Should handle empty bridge sets gracefully")

		bridges := mockEngine.ListBridges()
		assert.Empty(t, bridges, "No bridges should be registered")
	})
}
