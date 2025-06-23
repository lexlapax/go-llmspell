// ABOUTME: Tests for embedded stdlib Lua modules
// ABOUTME: Verifies embedding functionality, module listing, and content access

package stdlib

import (
	"testing"
)

func TestEmbeddedModules(t *testing.T) {
	t.Run("GetEmbeddedModules", func(t *testing.T) {
		modules, err := GetEmbeddedModules()
		if err != nil {
			t.Fatalf("Failed to get embedded modules: %v", err)
		}

		// Check that we have the expected modules
		expectedModules := []string{
			"core", "errors", "logging", "data", "auth",
			"state", "events", "tools", "llm", "agent",
			"observability", "spell", "promise", "testing",
		}

		moduleMap := make(map[string]bool)
		for _, m := range modules {
			moduleMap[m] = true
		}

		for _, expected := range expectedModules {
			if !moduleMap[expected] {
				t.Errorf("Expected module '%s' not found in embedded modules", expected)
			}
		}

		// Ensure we have at least the expected number of modules
		if len(modules) < len(expectedModules) {
			t.Errorf("Expected at least %d modules, got %d", len(expectedModules), len(modules))
		}
	})

	t.Run("ReadEmbeddedModule", func(t *testing.T) {
		testCases := []struct {
			name     string
			module   string
			contains string // Expected content substring
		}{
			{
				name:     "core module",
				module:   "core",
				contains: "return core",
			},
			{
				name:     "logging module",
				module:   "logging",
				contains: "return log",
			},
			{
				name:     "llm module",
				module:   "llm",
				contains: "return llm",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				content, err := ReadEmbeddedModule(tc.module)
				if err != nil {
					t.Fatalf("Failed to read module '%s': %v", tc.module, err)
				}

				if len(content) == 0 {
					t.Errorf("Module '%s' content is empty", tc.module)
				}

				contentStr := string(content)
				if tc.contains != "" && !contains(contentStr, tc.contains) {
					t.Errorf("Module '%s' does not contain expected string '%s'", tc.module, tc.contains)
				}
			})
		}
	})

	t.Run("ModuleExists", func(t *testing.T) {
		// Test existing modules
		existingModules := []string{"core", "logging", "llm", "agent"}
		for _, module := range existingModules {
			if !ModuleExists(module) {
				t.Errorf("Expected module '%s' to exist", module)
			}
		}

		// Test non-existing modules
		nonExistingModules := []string{"nonexistent", "fake", "missing"}
		for _, module := range nonExistingModules {
			if ModuleExists(module) {
				t.Errorf("Module '%s' should not exist", module)
			}
		}
	})

	t.Run("ReadNonExistentModule", func(t *testing.T) {
		_, err := ReadEmbeddedModule("nonexistent")
		if err == nil {
			t.Error("Expected error when reading non-existent module")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
