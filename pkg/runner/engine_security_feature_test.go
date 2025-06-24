// ABOUTME: Tests for engine-specific security level and feature set configuration
// ABOUTME: Ensures multi-engine architecture works correctly with new dual-flag system

package runner

import (
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge/registry"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEngineSecurityLevelAndFeatureSet tests the new security level + feature set system
func TestEngineSecurityLevelAndFeatureSet(t *testing.T) {
	// Create registry and manager
	engineRegistry := engine.NewRegistry(engine.RegistryConfig{})
	err := engineRegistry.Initialize()
	require.NoError(t, err)

	config := DefaultRunnerConfig()
	config.DefaultSecurityLevel = "trusted"
	config.DefaultFeatureSet = "full"
	_ = NewEngineRegistryManager(engineRegistry, config)

	t.Run("ValidSecurityLevelsAndFeatureSets", func(t *testing.T) {
		testCases := []struct {
			securityLevel string
			featureSet    string
			expectError   bool
		}{
			{"untrusted", "minimal", false},
			{"trusted", "llm", false},
			{"privileged", "full", false},
			{"invalid-level", "full", true},
			{"trusted", "invalid-set", true},
		}

		for _, tc := range testCases {
			t.Run(tc.securityLevel+"_"+tc.featureSet, func(t *testing.T) {
				// Test validation
				_ = security.SecurityLevel(tc.securityLevel)
				featSet := registry.FeatureSet(tc.featureSet)

				validLevel := security.IsValidLevel(tc.securityLevel)
				validFeatureSet := registry.IsValidFeatureSet(tc.featureSet)

				if tc.expectError {
					assert.False(t, validLevel && validFeatureSet)
				} else {
					assert.True(t, validLevel)
					assert.True(t, validFeatureSet)

					// Test that we can get feature sets for valid combinations
					bridgeSets := registry.GetBridgeSetsForFeature(featSet)
					assert.NotEmpty(t, bridgeSets)
				}
			})
		}
	})

	t.Run("FeatureSetBridgeMapping", func(t *testing.T) {
		testCases := []struct {
			featureSet         registry.FeatureSet
			expectedBridgeSets []registry.BridgeSet
			minExpectedCount   int
		}{
			{
				featureSet:         registry.FeatureSetMinimal,
				expectedBridgeSets: []registry.BridgeSet{registry.BridgeSetCore, registry.BridgeSetUtility},
				minExpectedCount:   2,
			},
			{
				featureSet:         registry.FeatureSetLLM,
				expectedBridgeSets: []registry.BridgeSet{registry.BridgeSetCore, registry.BridgeSetUtility, registry.BridgeSetLLM, registry.BridgeSetStructured},
				minExpectedCount:   4,
			},
			{
				featureSet:         registry.FeatureSetFull,
				expectedBridgeSets: []registry.BridgeSet{registry.BridgeSetCore, registry.BridgeSetUtility, registry.BridgeSetLLM, registry.BridgeSetStructured, registry.BridgeSetAgent, registry.BridgeSetState, registry.BridgeSetObservability},
				minExpectedCount:   7,
			},
		}

		for _, tc := range testCases {
			t.Run(string(tc.featureSet), func(t *testing.T) {
				bridgeSets := registry.GetBridgeSetsForFeature(tc.featureSet)
				assert.GreaterOrEqual(t, len(bridgeSets), tc.minExpectedCount)

				// Check that expected bridge sets are present
				for _, expectedSet := range tc.expectedBridgeSets {
					assert.Contains(t, bridgeSets, expectedSet)
				}
			})
		}
	})
}

// TestSecurityLevelValidation tests security level validation
func TestSecurityLevelValidation(t *testing.T) {
	testCases := []struct {
		level    string
		expected bool
	}{
		{"untrusted", true},
		{"trusted", true},
		{"privileged", true},
		{"sandbox", false},     // Old profile name
		{"development", false}, // Old profile name
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			result := security.IsValidLevel(tc.level)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestFeatureSetValidation tests feature set validation
func TestFeatureSetValidation(t *testing.T) {
	testCases := []struct {
		featureSet string
		expected   bool
	}{
		{"minimal", true},
		{"llm", true},
		{"agent", true},
		{"observable", true},
		{"full", true},
		{"standard", false},    // Old profile name
		{"development", false}, // Old profile name
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.featureSet, func(t *testing.T) {
			result := registry.IsValidFeatureSet(tc.featureSet)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestBackwardCompatibility tests backward compatibility mapping
func TestBackwardCompatibility(t *testing.T) {
	testCases := []struct {
		oldProfile         string
		expectedLevel      string
		expectedFeatureSet string
	}{
		{"sandbox", "untrusted", "full"},
		{"development", "trusted", "full"},
		{"production", "trusted", "full"},
		{"minimal", "trusted", "minimal"},
		{"llm", "trusted", "llm"},
		{"unknown", "trusted", "full"}, // Default fallback
	}

	for _, tc := range testCases {
		t.Run(tc.oldProfile, func(t *testing.T) {
			level, featureSet := mapProfileToLevelAndFeatureSet(tc.oldProfile)
			assert.Equal(t, tc.expectedLevel, level)
			assert.Equal(t, tc.expectedFeatureSet, featureSet)
		})
	}
}

// TestRunnerConfigDefaults tests that runner config has correct defaults
func TestRunnerConfigDefaults(t *testing.T) {
	config := DefaultRunnerConfig()

	assert.Equal(t, "trusted", config.DefaultSecurityLevel)
	assert.Equal(t, "full", config.DefaultFeatureSet)
	assert.Equal(t, "sandbox", config.DefaultSecurityProfile) // Deprecated but still present
}

// TestEngineConfigurationWithNewSystem tests engine configuration with new system
func TestEngineConfigurationWithNewSystem(t *testing.T) {
	config := &RunnerConfig{
		Timeout:              30 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
		DefaultSecurityLevel: "trusted",
		DefaultFeatureSet:    "llm",
	}

	engineConfig := BuildEngineConfig(config, nil)

	// Trusted level should not enable sandbox mode
	assert.False(t, engineConfig.SandboxMode)

	// Test with untrusted level
	config.DefaultSecurityLevel = "untrusted"
	engineConfig = BuildEngineConfig(config, nil)

	// Untrusted level should enable sandbox mode
	assert.True(t, engineConfig.SandboxMode)
}
