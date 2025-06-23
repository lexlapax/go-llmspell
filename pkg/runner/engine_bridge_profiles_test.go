// ABOUTME: Tests for engine-specific bridge profile mapping and configuration
// ABOUTME: Ensures multi-engine architecture works correctly with custom bridge profile configurations

package runner

import (
	"testing"
	"time"

	bridgeregistry "github.com/lexlapax/go-llmspell/pkg/bridge/registry"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEngineSpecificBridgeProfiles tests the default bridge profile mappings for different engines
func TestEngineSpecificBridgeProfiles(t *testing.T) {
	// Create registry and manager
	registry := engine.NewRegistry(engine.RegistryConfig{})
	err := registry.Initialize()
	require.NoError(t, err)

	config := DefaultRunnerConfig()
	manager := NewEngineRegistryManager(registry, config)

	t.Run("LuaEngineProfiles", func(t *testing.T) {
		testCases := []struct {
			securityProfile string
			expectedProfile bridgeregistry.BridgeProfile
			expectedName    string
		}{
			{"sandbox", bridgeregistry.StandardProfile, "standard"},
			{"development", bridgeregistry.DevelopmentProfile, "development"},
			{"production", bridgeregistry.StandardProfile, "standard"},
			{"minimal", bridgeregistry.MinimalProfile, "minimal"},
			{"llm", bridgeregistry.LLMProfile, "llm"},
			{"unknown", bridgeregistry.StandardProfile, "standard"},
		}

		for _, tc := range testCases {
			t.Run(tc.securityProfile, func(t *testing.T) {
				profile, err := manager.getBridgeProfileForSecurityProfile(tc.securityProfile, "lua")
				require.NoError(t, err)
				assert.Equal(t, tc.expectedProfile.Name, profile.Name)
				assert.Equal(t, tc.expectedName, profile.Name)
			})
		}
	})

	t.Run("JavaScriptEngineProfiles", func(t *testing.T) {
		testCases := []struct {
			securityProfile string
			expectedProfile bridgeregistry.BridgeProfile
			expectedName    string
		}{
			{"sandbox", bridgeregistry.LLMProfile, "llm"},
			{"development", bridgeregistry.DevelopmentProfile, "development"},
			{"production", bridgeregistry.LLMProfile, "llm"},
			{"minimal", bridgeregistry.MinimalProfile, "minimal"},
			{"llm", bridgeregistry.LLMProfile, "llm"},
			{"unknown", bridgeregistry.LLMProfile, "llm"},
		}

		for _, tc := range testCases {
			t.Run(tc.securityProfile, func(t *testing.T) {
				profile, err := manager.getBridgeProfileForSecurityProfile(tc.securityProfile, "javascript")
				require.NoError(t, err)
				assert.Equal(t, tc.expectedProfile.Name, profile.Name)
				assert.Equal(t, tc.expectedName, profile.Name)
			})
		}
	})

	t.Run("TengoEngineProfiles", func(t *testing.T) {
		testCases := []struct {
			securityProfile string
			expectedProfile bridgeregistry.BridgeProfile
			expectedName    string
		}{
			{"sandbox", bridgeregistry.MinimalProfile, "minimal"},
			{"development", bridgeregistry.DevelopmentProfile, "development"},
			{"production", bridgeregistry.MinimalProfile, "minimal"},
			{"minimal", bridgeregistry.MinimalProfile, "minimal"},
			{"llm", bridgeregistry.LLMProfile, "llm"},
			{"unknown", bridgeregistry.MinimalProfile, "minimal"},
		}

		for _, tc := range testCases {
			t.Run(tc.securityProfile, func(t *testing.T) {
				profile, err := manager.getBridgeProfileForSecurityProfile(tc.securityProfile, "tengo")
				require.NoError(t, err)
				assert.Equal(t, tc.expectedProfile.Name, profile.Name)
				assert.Equal(t, tc.expectedName, profile.Name)
			})
		}
	})

	t.Run("UnknownEngineProfiles", func(t *testing.T) {
		// Unknown engines should fall back to Lua behavior
		profile, err := manager.getBridgeProfileForSecurityProfile("sandbox", "unknown-engine")
		require.NoError(t, err)
		assert.Equal(t, bridgeregistry.StandardProfile.Name, profile.Name)
	})
}

// TestCustomEngineBridgeProfiles tests custom bridge profile mappings via configuration
func TestCustomEngineBridgeProfiles(t *testing.T) {
	// Create registry
	registry := engine.NewRegistry(engine.RegistryConfig{})
	err := registry.Initialize()
	require.NoError(t, err)

	t.Run("CustomProfileMapping", func(t *testing.T) {
		// Create config with custom engine bridge profiles
		config := &RunnerConfig{
			Timeout:                30 * time.Second,
			MaxConcurrentScripts:   10,
			DefaultEngine:          "lua",
			DefaultSecurityProfile: "sandbox",
			EngineBridgeProfiles: map[string]map[string]string{
				"lua": {
					"sandbox":     "minimal", // Override: lua+sandbox should use minimal instead of standard
					"development": "llm",     // Override: lua+development should use llm instead of development
				},
				"javascript": {
					"sandbox":    "standard", // Override: js+sandbox should use standard instead of llm
					"production": "minimal",  // Override: js+production should use minimal instead of llm
				},
				"custom-engine": {
					"sandbox": "development", // Custom engine with custom mapping
				},
			},
		}

		manager := NewEngineRegistryManager(registry, config)

		// Test overridden Lua profiles
		profile, err := manager.getBridgeProfileForSecurityProfile("sandbox", "lua")
		require.NoError(t, err)
		assert.Equal(t, "minimal", profile.Name) // Should be minimal, not standard

		profile, err = manager.getBridgeProfileForSecurityProfile("development", "lua")
		require.NoError(t, err)
		assert.Equal(t, "llm", profile.Name) // Should be llm, not development

		// Test non-overridden Lua profile (should use default)
		profile, err = manager.getBridgeProfileForSecurityProfile("production", "lua")
		require.NoError(t, err)
		assert.Equal(t, "standard", profile.Name) // Should use default lua behavior

		// Test overridden JavaScript profiles
		profile, err = manager.getBridgeProfileForSecurityProfile("sandbox", "javascript")
		require.NoError(t, err)
		assert.Equal(t, "standard", profile.Name) // Should be standard, not llm

		profile, err = manager.getBridgeProfileForSecurityProfile("production", "javascript")
		require.NoError(t, err)
		assert.Equal(t, "minimal", profile.Name) // Should be minimal, not llm

		// Test non-overridden JavaScript profile (should use default)
		profile, err = manager.getBridgeProfileForSecurityProfile("development", "javascript")
		require.NoError(t, err)
		assert.Equal(t, "development", profile.Name) // Should use default js behavior

		// Test custom engine with custom mapping
		profile, err = manager.getBridgeProfileForSecurityProfile("sandbox", "custom-engine")
		require.NoError(t, err)
		assert.Equal(t, "development", profile.Name)

		// Test custom engine without mapping (should fall back to default)
		profile, err = manager.getBridgeProfileForSecurityProfile("production", "custom-engine")
		require.NoError(t, err)
		assert.Equal(t, "standard", profile.Name) // Should fall back to lua default
	})

	t.Run("InvalidProfileName", func(t *testing.T) {
		config := &RunnerConfig{
			Timeout:                30 * time.Second,
			MaxConcurrentScripts:   10,
			DefaultEngine:          "lua",
			DefaultSecurityProfile: "sandbox",
			EngineBridgeProfiles: map[string]map[string]string{
				"lua": {
					"sandbox": "invalid-profile-name",
				},
			},
		}

		manager := NewEngineRegistryManager(registry, config)

		// Should return error for invalid profile name
		_, err := manager.getBridgeProfileForSecurityProfile("sandbox", "lua")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown bridge profile: invalid-profile-name")
	})

	t.Run("NilConfiguration", func(t *testing.T) {
		// Manager with nil config should use defaults
		manager := NewEngineRegistryManager(registry, nil)

		profile, err := manager.getBridgeProfileForSecurityProfile("sandbox", "lua")
		require.NoError(t, err)
		assert.Equal(t, "standard", profile.Name) // Should use default behavior
	})

	t.Run("EmptyEngineBridgeProfiles", func(t *testing.T) {
		// Manager with empty EngineBridgeProfiles should use defaults
		config := &RunnerConfig{
			Timeout:                30 * time.Second,
			MaxConcurrentScripts:   10,
			DefaultEngine:          "lua",
			DefaultSecurityProfile: "sandbox",
			EngineBridgeProfiles:   nil, // Explicitly nil
		}

		manager := NewEngineRegistryManager(registry, config)

		profile, err := manager.getBridgeProfileForSecurityProfile("sandbox", "lua")
		require.NoError(t, err)
		assert.Equal(t, "standard", profile.Name) // Should use default behavior
	})
}

// TestGetProfileByName tests the profile name lookup functionality
func TestGetProfileByName(t *testing.T) {
	registry := engine.NewRegistry(engine.RegistryConfig{})
	manager := NewEngineRegistryManager(registry, nil)

	testCases := []struct {
		name            string
		expectedProfile bridgeregistry.BridgeProfile
		expectError     bool
	}{
		{"standard", bridgeregistry.StandardProfile, false},
		{"minimal", bridgeregistry.MinimalProfile, false},
		{"llm", bridgeregistry.LLMProfile, false},
		{"development", bridgeregistry.DevelopmentProfile, false},
		{"invalid-name", bridgeregistry.BridgeProfile{}, true},
		{"", bridgeregistry.BridgeProfile{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile, err := manager.getProfileByName(tc.name)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unknown bridge profile")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedProfile.Name, profile.Name)
				assert.Equal(t, tc.expectedProfile.Description, profile.Description)
				assert.Equal(t, tc.expectedProfile.BridgeSets, profile.BridgeSets)
			}
		})
	}
}

// TestMultiEngineArchitectureReadiness tests that the architecture is ready for multiple engines
func TestMultiEngineArchitectureReadiness(t *testing.T) {
	t.Run("SetupWithMultipleEngines", func(t *testing.T) {
		// Test that registry setup works with the multi-engine architecture
		config := &RunnerConfig{
			Timeout:                30 * time.Second,
			MaxConcurrentScripts:   10,
			DefaultEngine:          "lua",
			DefaultSecurityProfile: "sandbox",
			EngineBridgeProfiles: map[string]map[string]string{
				"lua": {
					"sandbox": "standard",
				},
				"javascript": {
					"sandbox": "llm",
				},
				"tengo": {
					"sandbox": "minimal",
				},
			},
		}

		// This should work without errors
		engineManager, err := SetupEngineRegistry(config, "sandbox")
		require.NoError(t, err)
		assert.NotNil(t, engineManager)

		// Verify that Lua engine is available
		engines := engineManager.ListEngines()
		assert.Greater(t, len(engines), 0)

		// Verify that engine-specific bridge profiles work
		luaProfile, err := engineManager.getBridgeProfileForSecurityProfile("sandbox", "lua")
		require.NoError(t, err)
		assert.Equal(t, "standard", luaProfile.Name)

		jsProfile, err := engineManager.getBridgeProfileForSecurityProfile("sandbox", "javascript")
		require.NoError(t, err)
		assert.Equal(t, "llm", jsProfile.Name)

		tengoProfile, err := engineManager.getBridgeProfileForSecurityProfile("sandbox", "tengo")
		require.NoError(t, err)
		assert.Equal(t, "minimal", tengoProfile.Name)

		// Cleanup
		err = engineManager.Shutdown()
		require.NoError(t, err)
	})

	t.Run("LazyBridgeLoadingWithConfiguration", func(t *testing.T) {
		config := &RunnerConfig{
			Timeout:                30 * time.Second,
			MaxConcurrentScripts:   10,
			DefaultEngine:          "lua",
			DefaultSecurityProfile: "sandbox",
			EngineBridgeProfiles: map[string]map[string]string{
				"lua": {
					"sandbox": "minimal", // Use minimal instead of standard for testing
				},
			},
		}

		engineManager, err := SetupEngineRegistry(config, "sandbox")
		require.NoError(t, err)
		defer func() { _ = engineManager.Shutdown() }()

		// Note: SetupEngineRegistry already registers the Lua engine, so no need to register again

		// Test lazy bridge loading with custom profile
		engineConfig := BuildEngineConfig(config, nil)
		scriptEngine, err := engineManager.GetEngine("lua", engineConfig, "sandbox")
		require.NoError(t, err)
		assert.NotNil(t, scriptEngine)

		// Verify the correct bridge profile was used
		// The engine should have bridges loaded according to minimal profile
		// (This is tested by verifying no errors occurred during bridge loading)
	})
}
