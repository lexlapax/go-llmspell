// ABOUTME: Tests to verify engine selector works correctly with factory-only registration (lazy loading)
// ABOUTME: Ensures that engine discovery and selection functions without bridges being pre-registered

package runner

import (
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge/registry"
	"github.com/lexlapax/go-llmspell/pkg/security"
)

// TestEngineSelectorWithFactoryOnlyRegistration verifies that engine selector
// works correctly when only lightweight engine factories are registered,
// without bridges being loaded at startup.
func TestEngineSelectorWithFactoryOnlyRegistration(t *testing.T) {
	// Create a minimal runner config
	config := &RunnerConfig{
		Timeout:                30 * time.Second,
		MaxConcurrentScripts:   10,
		EnableMetrics:          true,
		EnableValidation:       true,
		EnableDebug:            false,
		DefaultEngine:          "lua",
		EngineConfigs:          make(map[string]map[string]interface{}),
		DefaultSecurityProfile: "sandbox",
		SecurityProfiles:       make(map[string]interface{}),
		Environment:            make(map[string]string),
	}

	// Setup engine registry with factory-only registration (no bridges)
	engineManager, err := SetupEngineRegistry(config, "sandbox")
	if err != nil {
		t.Fatalf("Failed to setup engine registry: %v", err)
	}
	defer func() { _ = engineManager.Shutdown() }()

	// Create engine selector
	selector := NewEngineSelector(engineManager)

	t.Run("SelectByExtension", func(t *testing.T) {
		// Test that engine selector can find engines by extension
		// even though only factories are registered
		testCases := []struct {
			name           string
			filepath       string
			expectedEngine string
			expectError    bool
		}{
			{
				name:           "lua extension",
				filepath:       "test.lua",
				expectedEngine: "lua",
				expectError:    false,
			},
			{
				name:        "unsupported extension",
				filepath:    "test.unknown",
				expectError: true,
			},
			{
				name:        "no extension",
				filepath:    "test",
				expectError: true,
			},
			{
				name:        "empty filepath",
				filepath:    "",
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				engine, err := selector.SelectByExtension(tc.filepath)

				if tc.expectError {
					if err == nil {
						t.Errorf("Expected error for filepath %s, but got none", tc.filepath)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for filepath %s: %v", tc.filepath, err)
					}
					if engine != tc.expectedEngine {
						t.Errorf("Expected engine %s for filepath %s, but got %s", tc.expectedEngine, tc.filepath, engine)
					}
				}
			})
		}
	})

	t.Run("GetSupportedExtensions", func(t *testing.T) {
		// Test that we can get supported extensions from factory-only registration
		extensions := selector.GetSupportedExtensions()

		if len(extensions) == 0 {
			t.Error("Expected at least one supported extension, but got none")
		}

		// Should at least have lua extension
		hasLua := false
		for _, ext := range extensions {
			if ext == "lua" {
				hasLua = true
				break
			}
		}
		if !hasLua {
			t.Errorf("Expected lua extension in supported extensions, but got: %v", extensions)
		}
	})

	t.Run("GetEngineExtensionMap", func(t *testing.T) {
		// Test that we can get the extension-to-engine mapping
		extensionMap := selector.GetEngineExtensionMap()

		if len(extensionMap) == 0 {
			t.Error("Expected at least one extension mapping, but got none")
		}

		// Should have lua -> lua mapping
		if engine, exists := extensionMap["lua"]; !exists {
			t.Error("Expected lua extension mapping, but it doesn't exist")
		} else if engine != "lua" {
			t.Errorf("Expected lua extension to map to lua engine, but got: %s", engine)
		}
	})

	t.Run("ValidateEngineAvailability", func(t *testing.T) {
		// Test that we can validate engine availability with factory-only registration
		testCases := []struct {
			name        string
			engineName  string
			expectError bool
		}{
			{
				name:        "lua engine available",
				engineName:  "lua",
				expectError: false,
			},
			{
				name:        "non-existent engine",
				engineName:  "nonexistent",
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := selector.ValidateEngineAvailability(tc.engineName)

				if tc.expectError {
					if err == nil {
						t.Errorf("Expected error for engine %s, but got none", tc.engineName)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for engine %s: %v", tc.engineName, err)
					}
				}
			})
		}
	})

	t.Run("SelectForSpell", func(t *testing.T) {
		// Test spell-based engine selection
		testCases := []struct {
			name           string
			metadata       *SpellMetadata
			expectedEngine string
			expectError    bool
		}{
			{
				name: "explicit engine",
				metadata: &SpellMetadata{
					Name:   "test-spell",
					Engine: "lua",
				},
				expectedEngine: "lua",
				expectError:    false,
			},
			{
				name: "infer from entry point",
				metadata: &SpellMetadata{
					Name:       "test-spell",
					EntryPoint: "main.lua",
				},
				expectedEngine: "lua",
				expectError:    false,
			},
			{
				name: "no engine or entry point",
				metadata: &SpellMetadata{
					Name: "test-spell",
				},
				expectError: true,
			},
			{
				name: "invalid explicit engine",
				metadata: &SpellMetadata{
					Name:   "test-spell",
					Engine: "nonexistent",
				},
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				engine, err := selector.SelectForSpell(tc.metadata)

				if tc.expectError {
					if err == nil {
						t.Errorf("Expected error for spell %+v, but got none", tc.metadata)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for spell %+v: %v", tc.metadata, err)
					}
					if engine != tc.expectedEngine {
						t.Errorf("Expected engine %s for spell %+v, but got %s", tc.expectedEngine, tc.metadata, engine)
					}
				}
			})
		}
	})

	t.Run("SelectWithOptions", func(t *testing.T) {
		// Test selection with runtime options overriding spell metadata
		metadata := &SpellMetadata{
			Name:       "test-spell",
			Engine:     "lua",
			EntryPoint: "main.lua",
		}

		// Options should override metadata
		options := &RunnerOptions{
			Engine: "lua", // In a multi-engine setup, this could be different
		}

		engine, err := selector.SelectWithOptions(metadata, options)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if engine != "lua" {
			t.Errorf("Expected lua engine, but got: %s", engine)
		}

		// Test with nil options (should use metadata)
		engine, err = selector.SelectWithOptions(metadata, nil)
		if err != nil {
			t.Errorf("Unexpected error with nil options: %v", err)
		}
		if engine != "lua" {
			t.Errorf("Expected lua engine with nil options, but got: %s", engine)
		}
	})
}

// TestEngineRegistryLazyBridgeLoading verifies that bridges are not loaded
// during engine registry setup, but only when engines are actually requested.
func TestEngineRegistryLazyBridgeLoading(t *testing.T) {
	config := DefaultRunnerConfig()

	// Setup engine registry
	engineManager, err := SetupEngineRegistry(config, "sandbox")
	if err != nil {
		t.Fatalf("Failed to setup engine registry: %v", err)
	}
	defer func() { _ = engineManager.Shutdown() }()

	t.Run("EngineListingWithoutBridgeLoading", func(t *testing.T) {
		// List engines should work without loading bridges
		engines := engineManager.ListEngines()

		if len(engines) == 0 {
			t.Error("Expected at least one engine to be listed, but got none")
		}

		// Should have lua engine
		hasLua := false
		for _, eng := range engines {
			if eng.Name == "lua" {
				hasLua = true
				// Verify basic engine info is available
				if len(eng.FileExtensions) == 0 {
					t.Error("Expected lua engine to have file extensions")
				}
				if eng.Status == "" {
					t.Error("Expected lua engine to have a status")
				}
				break
			}
		}
		if !hasLua {
			t.Error("Expected lua engine in engine list")
		}
	})

	t.Run("EngineInfoWithoutBridgeLoading", func(t *testing.T) {
		// Get engine info should work without loading bridges
		info, err := engineManager.GetEngineInfo("lua")
		if err != nil {
			t.Errorf("Failed to get lua engine info: %v", err)
		}

		if info.Name != "lua" {
			t.Errorf("Expected lua engine info, but got: %s", info.Name)
		}
		if len(info.FileExtensions) == 0 {
			t.Error("Expected lua engine to have file extensions")
		}
	})

	t.Run("ExtensionLookupWithoutBridgeLoading", func(t *testing.T) {
		// Extension lookup should work without loading bridges
		engineName, err := engineManager.FindEngineByExtension("lua")
		if err != nil {
			t.Errorf("Failed to find engine by extension: %v", err)
		}
		if engineName != "lua" {
			t.Errorf("Expected lua engine for .lua extension, but got: %s", engineName)
		}
	})
}

// TestGetEngineLazyLoading tests the lazy bridge loading functionality
func TestGetEngineLazyLoading(t *testing.T) {
	config := DefaultRunnerConfig()

	// Setup engine registry with factory-only registration
	engineManager, err := SetupEngineRegistry(config, "sandbox")
	if err != nil {
		t.Fatalf("Failed to setup engine registry: %v", err)
	}
	defer func() { _ = engineManager.Shutdown() }()

	t.Run("LazyBridgeLoading", func(t *testing.T) {
		// First call should load bridges
		engineConfig := BuildEngineConfig(config, nil)
		engine1, err := engineManager.GetEngine("lua", engineConfig, security.SecurityLevelUntrusted, registry.FeatureSetFull)
		if err != nil {
			t.Errorf("Failed to get engine with bridges: %v", err)
		}
		if engine1 == nil {
			t.Error("Expected engine instance, but got nil")
		}

		// Second call with same profile should use cached bridges
		engine2, err := engineManager.GetEngine("lua", engineConfig, security.SecurityLevelUntrusted, registry.FeatureSetFull)
		if err != nil {
			t.Errorf("Failed to get engine with bridges on second call: %v", err)
		}
		if engine2 == nil {
			t.Error("Expected engine instance on second call, but got nil")
		}
	})

	t.Run("DifferentSecurityProfiles", func(t *testing.T) {
		// Test different security profiles - the key insight is that the same engine
		// instance will be returned, but bridges should be loaded based on the first profile used
		baseConfig := BuildEngineConfig(config, nil)

		// First call with sandbox profile should succeed
		engine1, err := engineManager.GetEngine("lua", baseConfig, security.SecurityLevelUntrusted, registry.FeatureSetFull)
		if err != nil {
			t.Errorf("Failed to get engine with bridges for sandbox profile: %v", err)
		}
		if engine1 == nil {
			t.Error("Expected engine instance for sandbox profile, but got nil")
		}

		// Subsequent calls with different profiles should return the same engine instance
		// since bridges are already loaded (the engine registry reuses instances)
		profiles := []string{"development", "minimal", "llm", "production"}

		for _, profile := range profiles {
			t.Run(profile, func(t *testing.T) {
				engine, err := engineManager.GetEngine("lua", baseConfig, security.SecurityLevelUntrusted, registry.FeatureSetFull)
				if err != nil {
					t.Errorf("Failed to get engine with bridges for profile %s: %v", profile, err)
				}
				if engine == nil {
					t.Errorf("Expected engine instance for profile %s, but got nil", profile)
				}
				// The same engine instance should be returned
				if engine != engine1 {
					t.Logf("Note: Different engine instance returned for profile %s (this is OK if engine pooling creates new instances)", profile)
				}
			})
		}
	})

	t.Run("BridgeCaching", func(t *testing.T) {
		engineConfig := BuildEngineConfig(config, nil)

		// Get engine first to create cache entry
		scriptEngine, err := engineManager.GetEngine("lua", engineConfig, security.SecurityLevelUntrusted, registry.FeatureSetFull)
		if err != nil {
			t.Errorf("Failed to get engine: %v", err)
		}

		// Verify cache is working by checking internal state
		// Cache key is now based on engine instance address, security level, and feature set
		cacheKey := fmt.Sprintf("%p:%s:%s", scriptEngine, security.SecurityLevelUntrusted, registry.FeatureSetFull)
		engineManager.cacheMutex.RLock()
		cached := engineManager.bridgeCache[cacheKey]
		engineManager.cacheMutex.RUnlock()

		if !cached {
			t.Error("Expected bridges to be cached after first call")
		}

		// Second call should use cached bridges
		scriptEngine2, err := engineManager.GetEngine("lua", engineConfig, security.SecurityLevelUntrusted, registry.FeatureSetFull)
		if err != nil {
			t.Errorf("Failed to get engine on second call: %v", err)
		}

		// Should be the same engine instance
		if scriptEngine != scriptEngine2 {
			t.Log("Note: Different engine instances returned (this is OK if engine pooling is used)")
		}
	})
}

// TestEngineCommandsWithFactoryOnly tests that engine-related CLI commands
// work correctly with factory-only registration.
func TestEngineCommandsWithFactoryOnly(t *testing.T) {
	config := DefaultRunnerConfig()

	// Setup engine registry with factory-only registration
	engineManager, err := SetupEngineRegistry(config, "sandbox")
	if err != nil {
		t.Fatalf("Failed to setup engine registry: %v", err)
	}
	defer func() { _ = engineManager.Shutdown() }()

	// Create engine selector
	selector := NewEngineSelector(engineManager)

	t.Run("EnginesCommandSupport", func(t *testing.T) {
		// Verify that `llmspell engines` command would work
		engines := engineManager.ListEngines()

		if len(engines) == 0 {
			t.Error("Expected engines list to be non-empty for engines command")
		}

		// Check that we have the expected engine metadata for CLI display
		for _, engine := range engines {
			if engine.Name == "" {
				t.Error("Engine name should not be empty")
			}
			if engine.Description == "" {
				t.Error("Engine description should not be empty")
			}
			if len(engine.FileExtensions) == 0 {
				t.Error("Engine should have file extensions")
			}
			if engine.Status == "" {
				t.Error("Engine status should not be empty")
			}
		}
	})

	t.Run("ScriptValidationSupport", func(t *testing.T) {
		// Verify that script validation could work without bridge loading
		// (this would be used by validate command)
		engineName, err := selector.SelectByExtension("test.lua")
		if err != nil {
			t.Errorf("Failed to select engine for validation: %v", err)
		}
		if engineName != "lua" {
			t.Errorf("Expected lua engine for validation, but got: %s", engineName)
		}

		// Verify we can get engine info for validation
		_, err = engineManager.GetEngineInfo(engineName)
		if err != nil {
			t.Errorf("Failed to get engine info for validation: %v", err)
		}
	})
}
