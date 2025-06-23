// ABOUTME: Tests for stdlib module loader functionality
// ABOUTME: Verifies module loading, caching, dependencies, and Lua integration

package stdlib

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestModuleLoader(t *testing.T) {
	t.Run("LoadEmbeddedModule", func(t *testing.T) {
		// Test loading an existing module
		loader, err := LoadEmbeddedModule("core")
		if err != nil {
			t.Fatalf("Failed to load core module: %v", err)
		}

		if loader == nil {
			t.Error("Loader function is nil")
		}

		// Test loading non-existent module
		_, err = LoadEmbeddedModule("nonexistent")
		if err == nil {
			t.Error("Expected error when loading non-existent module")
		}
	})

	t.Run("GetAllStdlibLoaders", func(t *testing.T) {
		loaders, err := GetAllStdlibLoaders()
		if err != nil {
			t.Fatalf("Failed to get all stdlib loaders: %v", err)
		}

		// Check that we have loaders for expected modules
		expectedModules := []string{"core", "logging", "llm", "agent", "data"}
		for _, module := range expectedModules {
			if _, exists := loaders[module]; !exists {
				t.Errorf("No loader found for module '%s'", module)
			}
		}

		// Verify each loader is not nil
		for name, loader := range loaders {
			if loader == nil {
				t.Errorf("Loader for module '%s' is nil", name)
			}
		}
	})

	t.Run("ModuleLoadingInLua", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Get all loaders
		loaders, err := GetAllStdlibLoaders()
		if err != nil {
			t.Fatalf("Failed to get loaders: %v", err)
		}

		// Preload all modules
		for name, loader := range loaders {
			L.PreloadModule(name, loader)
		}

		// Test requiring a module
		testScript := `
			local core = require("core")
			assert(core ~= nil, "core module is nil")
			assert(type(core) == "table", "core module is not a table")
			
			local log = require("logging")
			assert(log ~= nil, "logging module is nil")
			assert(type(log) == "table", "logging module is not a table")
			
			return true
		`

		if err := L.DoString(testScript); err != nil {
			t.Fatalf("Failed to execute test script: %v", err)
		}

		result := L.Get(-1)
		if result != lua.LTrue {
			t.Error("Test script did not return true")
		}
	})

	t.Run("ModuleCaching", func(t *testing.T) {
		// Clear cache first
		ClearModuleCache()

		L := lua.NewState()
		defer L.Close()

		loader, err := LoadEmbeddedModule("core")
		if err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		// First load - should compile
		L.PreloadModule("core", loader)
		if err := L.DoString(`local core1 = require("core")`); err != nil {
			t.Fatalf("First require failed: %v", err)
		}

		// Second load - should use cache
		if err := L.DoString(`local core2 = require("core")`); err != nil {
			t.Fatalf("Second require failed: %v", err)
		}

		// Verify cache has the module
		moduleCacheMu.RLock()
		_, cached := moduleCache["core"]
		moduleCacheMu.RUnlock()

		if !cached {
			t.Error("Module was not cached")
		}
	})

	t.Run("GetOrderedStdlibLoaders", func(t *testing.T) {
		ordered, err := GetOrderedStdlibLoaders()
		if err != nil {
			t.Fatalf("Failed to get ordered loaders: %v", err)
		}

		// Verify core modules come first
		coreModules := []string{"core", "errors", "logging", "data"}
		for i, expected := range coreModules {
			if i >= len(ordered) {
				t.Fatalf("Not enough modules in ordered list")
			}
			if ordered[i].Name != expected {
				t.Errorf("Expected module at position %d to be '%s', got '%s'", i, expected, ordered[i].Name)
			}
		}

		// Verify all loaders are present
		loaderMap := make(map[string]bool)
		for _, item := range ordered {
			if item.Loader == nil {
				t.Errorf("Loader for module '%s' is nil", item.Name)
			}
			loaderMap[item.Name] = true
		}

		// Check that all expected modules are in the ordered list
		allLoaders, _ := GetAllStdlibLoaders()
		for name := range allLoaders {
			if !loaderMap[name] {
				t.Errorf("Module '%s' missing from ordered list", name)
			}
		}
	})

	t.Run("ModuleDependencies", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Load modules in order
		ordered, err := GetOrderedStdlibLoaders()
		if err != nil {
			t.Fatalf("Failed to get ordered loaders: %v", err)
		}

		for _, item := range ordered {
			L.PreloadModule(item.Name, item.Loader)
		}

		// Test that modules can access their dependencies
		testScript := `
			-- Agent module depends on llm module
			local agent = require("agent")
			assert(agent ~= nil, "agent module is nil")
			
			-- State module might use data module
			local state = require("state")
			assert(state ~= nil, "state module is nil")
			
			return true
		`

		if err := L.DoString(testScript); err != nil {
			t.Fatalf("Module dependency test failed: %v", err)
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Test that loader handles errors gracefully
		loader, err := LoadEmbeddedModule("core")
		if err != nil {
			t.Fatalf("Failed to create loader: %v", err)
		}

		// Preload the module
		L.PreloadModule("core", loader)

		// This should work
		if err := L.DoString(`local core = require("core")`); err != nil {
			t.Errorf("Failed to require module: %v", err)
		}

		// Test requiring non-existent module
		err = L.DoString(`local fake = require("fake_module")`)
		if err == nil {
			t.Error("Expected error when requiring non-existent module")
		}
	})
}
