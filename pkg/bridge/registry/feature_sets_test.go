// ABOUTME: Unit tests for feature set definitions and bridge mappings.
// ABOUTME: Tests feature set validation, bridge set retrieval, and descriptions.

package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureSetConstants(t *testing.T) {
	// Test that feature set constants are properly defined
	assert.Equal(t, FeatureSet("minimal"), FeatureSetMinimal)
	assert.Equal(t, FeatureSet("llm"), FeatureSetLLM)
	assert.Equal(t, FeatureSet("agent"), FeatureSetAgent)
	assert.Equal(t, FeatureSet("observable"), FeatureSetObservable)
	assert.Equal(t, FeatureSet("full"), FeatureSetFull)
}

func TestIsValidFeatureSet(t *testing.T) {
	tests := []struct {
		name     string
		fs       string
		expected bool
	}{
		{"valid minimal", "minimal", true},
		{"valid llm", "llm", true},
		{"valid agent", "agent", true},
		{"valid observable", "observable", true},
		{"valid full", "full", true},
		{"invalid empty", "", false},
		{"invalid unknown", "unknown", false},
		{"invalid standard", "standard", false},       // Old profile name
		{"invalid development", "development", false}, // Old profile name
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidFeatureSet(tt.fs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBridgeSetsForFeature(t *testing.T) {
	tests := []struct {
		name         string
		featureSet   FeatureSet
		expectedSets []BridgeSet
	}{
		{
			name:       "minimal feature set",
			featureSet: FeatureSetMinimal,
			expectedSets: []BridgeSet{
				BridgeSetCore,
				BridgeSetUtility,
			},
		},
		{
			name:       "llm feature set",
			featureSet: FeatureSetLLM,
			expectedSets: []BridgeSet{
				BridgeSetCore,
				BridgeSetUtility,
				BridgeSetLLM,
				BridgeSetStructured,
			},
		},
		{
			name:       "agent feature set",
			featureSet: FeatureSetAgent,
			expectedSets: []BridgeSet{
				BridgeSetCore,
				BridgeSetUtility,
				BridgeSetLLM,
				BridgeSetStructured,
				BridgeSetAgent,
				BridgeSetState,
			},
		},
		{
			name:       "observable feature set",
			featureSet: FeatureSetObservable,
			expectedSets: []BridgeSet{
				BridgeSetCore,
				BridgeSetUtility,
				BridgeSetLLM,
				BridgeSetObservability,
			},
		},
		{
			name:       "full feature set",
			featureSet: FeatureSetFull,
			expectedSets: []BridgeSet{
				BridgeSetCore,
				BridgeSetUtility,
				BridgeSetLLM,
				BridgeSetStructured,
				BridgeSetAgent,
				BridgeSetState,
				BridgeSetObservability,
			},
		},
		{
			name:       "unknown feature set defaults to minimal",
			featureSet: FeatureSet("unknown"),
			expectedSets: []BridgeSet{
				BridgeSetCore,
				BridgeSetUtility,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sets := GetBridgeSetsForFeature(tt.featureSet)
			assert.Equal(t, tt.expectedSets, sets)
		})
	}
}

func TestFeatureSetBridgesMapping(t *testing.T) {
	// Ensure all defined feature sets have mappings
	featureSets := []FeatureSet{
		FeatureSetMinimal,
		FeatureSetLLM,
		FeatureSetAgent,
		FeatureSetObservable,
		FeatureSetFull,
	}

	for _, fs := range featureSets {
		t.Run(string(fs), func(t *testing.T) {
			sets, exists := FeatureSetBridges[fs]
			assert.True(t, exists, "Feature set %s should have bridge mapping", fs)
			assert.NotEmpty(t, sets, "Feature set %s should have at least one bridge set", fs)

			// All feature sets should include Core and Utility
			assert.Contains(t, sets, BridgeSetCore, "Feature set %s should include Core bridges", fs)
			assert.Contains(t, sets, BridgeSetUtility, "Feature set %s should include Utility bridges", fs)
		})
	}
}

func TestGetFeatureSetDescription(t *testing.T) {
	tests := []struct {
		name        string
		featureSet  FeatureSet
		expectEmpty bool
	}{
		{"minimal has description", FeatureSetMinimal, false},
		{"llm has description", FeatureSetLLM, false},
		{"agent has description", FeatureSetAgent, false},
		{"observable has description", FeatureSetObservable, false},
		{"full has description", FeatureSetFull, false},
		{"unknown returns default", FeatureSet("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := GetFeatureSetDescription(tt.featureSet)
			if tt.expectEmpty {
				assert.Empty(t, desc)
			} else {
				assert.NotEmpty(t, desc)
				if tt.featureSet == FeatureSet("unknown") {
					assert.Equal(t, "Unknown feature set", desc)
				}
			}
		})
	}
}

func TestGetBridgesForFeatureSet(t *testing.T) {
	// This test verifies that GetBridgesForFeatureSet returns bridges
	// Note: This test will need the actual bridge factories to be defined

	t.Run("minimal feature set", func(t *testing.T) {
		bridges, err := GetBridgesForFeatureSet(FeatureSetMinimal)
		// If bridge factories are not yet defined, this should not error but return empty
		assert.NoError(t, err)
		assert.NotNil(t, bridges)
	})

	t.Run("full feature set", func(t *testing.T) {
		bridges, err := GetBridgesForFeatureSet(FeatureSetFull)
		assert.NoError(t, err)
		assert.NotNil(t, bridges)
	})
}

func TestFeatureSetHierarchy(t *testing.T) {
	// Test that feature sets build on each other properly
	minimalSets := GetBridgeSetsForFeature(FeatureSetMinimal)
	llmSets := GetBridgeSetsForFeature(FeatureSetLLM)
	agentSets := GetBridgeSetsForFeature(FeatureSetAgent)
	fullSets := GetBridgeSetsForFeature(FeatureSetFull)

	// LLM should include everything from Minimal
	for _, set := range minimalSets {
		assert.Contains(t, llmSets, set, "LLM should include all Minimal bridge sets")
	}

	// Agent should include Core, Utility, LLM, Structured (from LLM)
	assert.Contains(t, agentSets, BridgeSetCore)
	assert.Contains(t, agentSets, BridgeSetUtility)
	assert.Contains(t, agentSets, BridgeSetLLM)
	assert.Contains(t, agentSets, BridgeSetStructured)

	// Full should include everything
	assert.Contains(t, fullSets, BridgeSetCore)
	assert.Contains(t, fullSets, BridgeSetUtility)
	assert.Contains(t, fullSets, BridgeSetLLM)
	assert.Contains(t, fullSets, BridgeSetStructured)
	assert.Contains(t, fullSets, BridgeSetAgent)
	assert.Contains(t, fullSets, BridgeSetState)
	assert.Contains(t, fullSets, BridgeSetObservability)
}

func TestFeatureSetConsistency(t *testing.T) {
	// Ensure feature set mappings are consistent

	t.Run("no duplicate bridge sets", func(t *testing.T) {
		for fs, sets := range FeatureSetBridges {
			seen := make(map[BridgeSet]bool)
			for _, set := range sets {
				assert.False(t, seen[set], "Feature set %s has duplicate bridge set %s", fs, set)
				seen[set] = true
			}
		}
	})

	t.Run("all bridge sets are valid", func(t *testing.T) {
		validSets := map[BridgeSet]bool{
			BridgeSetCore:          true,
			BridgeSetLLM:           true,
			BridgeSetUtility:       true,
			BridgeSetAgent:         true,
			BridgeSetObservability: true,
			BridgeSetState:         true,
			BridgeSetStructured:    true,
		}

		for fs, sets := range FeatureSetBridges {
			for _, set := range sets {
				assert.True(t, validSets[set], "Feature set %s contains invalid bridge set %s", fs, set)
			}
		}
	})
}
