// ABOUTME: Tests for the module system which manages registration, loading, and dependency resolution of Lua modules
// ABOUTME: Validates module lifecycle, lazy loading, circular dependency detection, and profile-based loading

package gopherlua

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestModuleSystem_Registration(t *testing.T) {
	t.Run("register_simple_module", func(t *testing.T) {
		ms := NewModuleSystem()

		module := ModuleDefinition{
			Name: "test",
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.SetField(mod, "version", lua.LString("1.0"))
				L.Push(mod)
				return 1
			},
			Priority: 10,
		}

		err := ms.Register(module)
		assert.NoError(t, err)

		// Module should exist
		assert.True(t, ms.Exists("test"))
	})

	t.Run("register_duplicate_module", func(t *testing.T) {
		ms := NewModuleSystem()

		module := ModuleDefinition{
			Name: "test",
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}

		err := ms.Register(module)
		assert.NoError(t, err)

		// Registering duplicate should fail
		err = ms.Register(module)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("register_module_with_dependencies", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register dependency first
		dep := ModuleDefinition{
			Name: "base",
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}
		err := ms.Register(dep)
		require.NoError(t, err)

		// Register module with dependency
		module := ModuleDefinition{
			Name:         "extended",
			Dependencies: []string{"base"},
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}

		err = ms.Register(module)
		assert.NoError(t, err)
	})

	t.Run("register_module_with_forward_dependency", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register module with dependency that doesn't exist yet
		module := ModuleDefinition{
			Name:         "test",
			Dependencies: []string{"forward"},
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}

		// Should succeed - forward references are allowed
		err := ms.Register(module)
		assert.NoError(t, err)

		// But loading should fail
		L := lua.NewState()
		defer L.Close()

		err = ms.LoadModule(L, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing dependency")

		// Now register the dependency
		forward := ModuleDefinition{
			Name: "forward",
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}
		err = ms.Register(forward)
		assert.NoError(t, err)

		// Now loading should succeed
		err = ms.LoadModule(L, "test")
		assert.NoError(t, err)
	})
}

func TestModuleSystem_Loading(t *testing.T) {
	t.Run("load_simple_module", func(t *testing.T) {
		ms := NewModuleSystem()

		loaded := false
		module := ModuleDefinition{
			Name: "test",
			LoadFunc: func(L *lua.LState) int {
				loaded = true
				mod := L.NewTable()
				L.SetField(mod, "loaded", lua.LBool(true))
				L.Push(mod)
				return 1
			},
		}

		err := ms.Register(module)
		require.NoError(t, err)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Load module
		err = ms.LoadModule(L, "test")
		assert.NoError(t, err)
		assert.True(t, loaded)

		// Module should be available via require
		err = L.DoString(`
			local test = require("test")
			assert(test ~= nil, "module should not be nil")
			assert(test.loaded == true, "module should have loaded field")
		`)
		require.NoError(t, err)
	})

	t.Run("load_module_with_dependencies", func(t *testing.T) {
		ms := NewModuleSystem()

		loadOrder := []string{}

		// Register base module
		base := ModuleDefinition{
			Name: "base",
			LoadFunc: func(L *lua.LState) int {
				loadOrder = append(loadOrder, "base")
				mod := L.NewTable()
				L.SetField(mod, "name", lua.LString("base"))
				L.Push(mod)
				return 1
			},
			Priority: 1,
		}
		err := ms.Register(base)
		require.NoError(t, err)

		// Register dependent module
		dependent := ModuleDefinition{
			Name:         "dependent",
			Dependencies: []string{"base"},
			LoadFunc: func(L *lua.LState) int {
				loadOrder = append(loadOrder, "dependent")
				mod := L.NewTable()
				L.SetField(mod, "name", lua.LString("dependent"))
				L.Push(mod)
				return 1
			},
			Priority: 10,
		}
		err = ms.Register(dependent)
		require.NoError(t, err)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Load dependent module (should load base first)
		err = ms.LoadModule(L, "dependent")
		assert.NoError(t, err)

		// Check load order
		assert.Equal(t, []string{"base", "dependent"}, loadOrder)
	})

	t.Run("lazy_loading", func(t *testing.T) {
		ms := NewModuleSystem()

		loaded := false
		module := ModuleDefinition{
			Name: "lazy",
			LoadFunc: func(L *lua.LState) int {
				loaded = true
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}

		err := ms.Register(module)
		require.NoError(t, err)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Preload module (should not load immediately)
		err = ms.PreloadModule(L, "lazy")
		assert.NoError(t, err)
		assert.False(t, loaded)

		// Module loads on require
		err = L.DoString(`local lazy = require("lazy")`)
		assert.NoError(t, err)
		assert.True(t, loaded)
	})

	t.Run("load_nonexistent_module", func(t *testing.T) {
		ms := NewModuleSystem()

		L := lua.NewState()
		defer L.Close()

		err := ms.LoadModule(L, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestModuleSystem_CircularDependencies(t *testing.T) {
	t.Run("detect_direct_circular_dependency", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register module A depending on B
		moduleA := ModuleDefinition{
			Name:         "moduleA",
			Dependencies: []string{"moduleB"},
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}
		err := ms.Register(moduleA)
		require.NoError(t, err)

		// Try to register module B depending on A (circular)
		moduleB := ModuleDefinition{
			Name:         "moduleB",
			Dependencies: []string{"moduleA"},
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}
		err = ms.Register(moduleB)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency")
	})

	t.Run("detect_indirect_circular_dependency", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register A -> B
		moduleA := ModuleDefinition{
			Name:         "A",
			Dependencies: []string{"B"},
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}
		err := ms.Register(moduleA)
		require.NoError(t, err)

		// Register B -> C
		moduleB := ModuleDefinition{
			Name:         "B",
			Dependencies: []string{"C"},
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}
		err = ms.Register(moduleB)
		require.NoError(t, err)

		// Try to register C -> A (creates A->B->C->A cycle)
		moduleC := ModuleDefinition{
			Name:         "C",
			Dependencies: []string{"A"},
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}
		err = ms.Register(moduleC)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency")
	})
}

func TestModuleSystem_Priority(t *testing.T) {
	t.Run("load_modules_by_priority", func(t *testing.T) {
		ms := NewModuleSystem()

		loadOrder := []string{}

		// Register modules with different priorities
		modules := []ModuleDefinition{
			{
				Name:     "low",
				Priority: 100,
				LoadFunc: func(L *lua.LState) int {
					loadOrder = append(loadOrder, "low")
					return 0
				},
			},
			{
				Name:     "high",
				Priority: 1,
				LoadFunc: func(L *lua.LState) int {
					loadOrder = append(loadOrder, "high")
					return 0
				},
			},
			{
				Name:     "medium",
				Priority: 50,
				LoadFunc: func(L *lua.LState) int {
					loadOrder = append(loadOrder, "medium")
					return 0
				},
			},
		}

		for _, mod := range modules {
			err := ms.Register(mod)
			require.NoError(t, err)
		}

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Load all modules
		err := ms.LoadAll(L)
		assert.NoError(t, err)

		// Should load in priority order (lower number = higher priority)
		assert.Equal(t, []string{"high", "medium", "low"}, loadOrder)
	})
}

func TestModuleSystem_Profiles(t *testing.T) {
	t.Run("profile_based_loading", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register modules with profiles
		coreModule := ModuleDefinition{
			Name:     "core",
			Profiles: []string{"minimal", "standard", "full"},
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}
		err := ms.Register(coreModule)
		require.NoError(t, err)

		extendedModule := ModuleDefinition{
			Name:     "extended",
			Profiles: []string{"standard", "full"},
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}
		err = ms.Register(extendedModule)
		require.NoError(t, err)

		advancedModule := ModuleDefinition{
			Name:     "advanced",
			Profiles: []string{"full"},
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}
		err = ms.Register(advancedModule)
		require.NoError(t, err)

		// Test minimal profile
		L := lua.NewState()
		defer L.Close()

		err = ms.LoadProfile(L, "minimal")
		assert.NoError(t, err)

		// Only core should be loaded
		assert.True(t, ms.IsLoaded("core"))
		assert.False(t, ms.IsLoaded("extended"))
		assert.False(t, ms.IsLoaded("advanced"))
	})
}

func TestModuleSystem_Bundle(t *testing.T) {
	t.Run("module_bundling", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register individual modules
		modules := []ModuleDefinition{
			{
				Name: "llm",
				LoadFunc: func(L *lua.LState) int {
					mod := L.NewTable()
					L.Push(mod)
					return 1
				},
			},
			{
				Name: "tools",
				LoadFunc: func(L *lua.LState) int {
					mod := L.NewTable()
					L.Push(mod)
					return 1
				},
			},
			{
				Name: "state",
				LoadFunc: func(L *lua.LState) int {
					mod := L.NewTable()
					L.Push(mod)
					return 1
				},
			},
		}

		for _, mod := range modules {
			err := ms.Register(mod)
			require.NoError(t, err)
		}

		// Create bundle
		bundle := ModuleBundle{
			Name:    "bridges",
			Modules: []string{"llm", "tools", "state"},
		}
		err := ms.RegisterBundle(bundle)
		assert.NoError(t, err)

		// Load bundle
		L := lua.NewState()
		defer L.Close()

		err = ms.LoadBundle(L, "bridges")
		assert.NoError(t, err)

		// All modules in bundle should be loaded
		assert.True(t, ms.IsLoaded("llm"))
		assert.True(t, ms.IsLoaded("tools"))
		assert.True(t, ms.IsLoaded("state"))
	})
}

func TestModuleSystem_InitCallbacks(t *testing.T) {
	t.Run("module_initialization", func(t *testing.T) {
		ms := NewModuleSystem()

		initialized := false
		module := ModuleDefinition{
			Name: "test",
			InitFunc: func() error {
				initialized = true
				return nil
			},
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}

		err := ms.Register(module)
		require.NoError(t, err)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Loading should trigger initialization
		err = ms.LoadModule(L, "test")
		assert.NoError(t, err)
		assert.True(t, initialized)
	})

	t.Run("initialization_failure", func(t *testing.T) {
		ms := NewModuleSystem()

		module := ModuleDefinition{
			Name: "failing",
			InitFunc: func() error {
				return assert.AnError
			},
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}

		err := ms.Register(module)
		require.NoError(t, err)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Loading should fail due to init error
		err = ms.LoadModule(L, "failing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initialization failed")
	})
}

func TestModuleSystem_VersionManagement(t *testing.T) {
	t.Run("module_versioning", func(t *testing.T) {
		ms := NewModuleSystem()

		module := ModuleDefinition{
			Name:    "versioned",
			Version: "1.2.3",
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.SetField(mod, "version", lua.LString("1.2.3"))
				L.Push(mod)
				return 1
			},
		}

		err := ms.Register(module)
		require.NoError(t, err)

		// Get module info
		info, err := ms.GetModuleInfo("versioned")
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", info.Version)
	})

	t.Run("dependency_version_constraints", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register base module with version
		base := ModuleDefinition{
			Name:    "base",
			Version: "1.0.0",
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}
		err := ms.Register(base)
		require.NoError(t, err)

		// Register dependent with version constraint
		dependent := ModuleDefinition{
			Name:         "dependent",
			Dependencies: []string{"base@>=1.0.0"},
			LoadFunc: func(L *lua.LState) int {
				return 0
			},
		}
		err = ms.Register(dependent)
		assert.NoError(t, err)
	})
}

func TestModuleSystem_Concurrency(t *testing.T) {
	t.Run("concurrent_registration", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register modules concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				module := ModuleDefinition{
					Name: fmt.Sprintf("module%d", idx),
					LoadFunc: func(L *lua.LState) int {
						return 0
					},
				}
				err := ms.Register(module)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all registrations
		for i := 0; i < 10; i++ {
			<-done
		}

		// All modules should exist
		for i := 0; i < 10; i++ {
			assert.True(t, ms.Exists(fmt.Sprintf("module%d", i)))
		}
	})

	t.Run("concurrent_loading", func(t *testing.T) {
		ms := NewModuleSystem()

		// Register a module
		module := ModuleDefinition{
			Name: "shared",
			LoadFunc: func(L *lua.LState) int {
				mod := L.NewTable()
				L.Push(mod)
				return 1
			},
		}
		err := ms.Register(module)
		require.NoError(t, err)

		// Load from multiple states concurrently
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				L := lua.NewState()
				defer L.Close()

				err := ms.LoadModule(L, "shared")
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all loads
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}
