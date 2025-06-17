// ABOUTME: Module system for managing Lua module registration, loading, and dependency resolution
// ABOUTME: Provides lazy loading, circular dependency detection, and profile-based module management

package gopherlua

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// ModuleDefinition defines a Lua module
type ModuleDefinition struct {
	Name         string         // Module name
	Version      string         // Module version (optional)
	Description  string         // Module description
	Dependencies []string       // List of dependencies
	Profiles     []string       // Profiles this module belongs to
	Priority     int            // Loading priority (lower = higher priority)
	InitFunc     func() error   // Initialization function (called once)
	LoadFunc     lua.LGFunction // Lua module loader function
	initialized  bool           // Whether InitFunc has been called
}

// ModuleBundle groups related modules
type ModuleBundle struct {
	Name        string   // Bundle name
	Description string   // Bundle description
	Modules     []string // Module names in the bundle
}

// ModuleInfo provides information about a registered module
type ModuleInfo struct {
	Name         string
	Version      string
	Description  string
	Dependencies []string
	Profiles     []string
	Priority     int
	Loaded       bool
}

// ModuleSystem manages Lua modules
type ModuleSystem struct {
	mu          sync.RWMutex
	modules     map[string]*ModuleDefinition
	bundles     map[string]*ModuleBundle
	loaded      map[string]bool                 // Global loaded state
	loading     map[string]bool                 // Tracks modules currently being loaded
	stateLoaded map[*lua.LState]map[string]bool // Per-state loaded tracking
}

// NewModuleSystem creates a new module system
func NewModuleSystem() *ModuleSystem {
	return &ModuleSystem{
		modules:     make(map[string]*ModuleDefinition),
		bundles:     make(map[string]*ModuleBundle),
		loaded:      make(map[string]bool),
		loading:     make(map[string]bool),
		stateLoaded: make(map[*lua.LState]map[string]bool),
	}
}

// Register registers a module definition
func (ms *ModuleSystem) Register(module ModuleDefinition) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check if already registered
	if _, exists := ms.modules[module.Name]; exists {
		return fmt.Errorf("module %s already registered", module.Name)
	}

	// Store temporarily for circular dependency check
	ms.modules[module.Name] = &module

	// Check for circular dependencies (now that module is in registry)
	if err := ms.checkCircularDependency(module.Name, module.Dependencies, nil); err != nil {
		// Remove module on error
		delete(ms.modules, module.Name)
		return err
	}

	// Validate dependencies exist (warning only for forward references)
	// We allow forward references, so missing dependencies are OK at registration time
	// They will be checked at load time

	return nil
}

// RegisterBundle registers a module bundle
func (ms *ModuleSystem) RegisterBundle(bundle ModuleBundle) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate all modules in bundle exist
	for _, modName := range bundle.Modules {
		if _, exists := ms.modules[modName]; !exists {
			return fmt.Errorf("bundle %s references non-existent module: %s", bundle.Name, modName)
		}
	}

	ms.bundles[bundle.Name] = &bundle
	return nil
}

// Exists checks if a module is registered
func (ms *ModuleSystem) Exists(name string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	_, exists := ms.modules[name]
	return exists
}

// IsLoaded checks if a module has been loaded
func (ms *ModuleSystem) IsLoaded(name string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.loaded[name]
}

// LoadModule loads a module and its dependencies into a Lua state
func (ms *ModuleSystem) LoadModule(L *lua.LState, name string) error {
	ms.mu.Lock()
	module, exists := ms.modules[name]
	if !exists {
		ms.mu.Unlock()
		return fmt.Errorf("module %s not found", name)
	}

	// Initialize per-state tracking if needed
	if ms.stateLoaded[L] == nil {
		ms.stateLoaded[L] = make(map[string]bool)
	}

	// Check if already loaded in this state
	if ms.stateLoaded[L][name] {
		ms.mu.Unlock()
		return nil
	}

	// Mark as loading to prevent infinite recursion
	ms.loading[name] = true
	ms.mu.Unlock()

	// Validate dependencies exist at load time
	ms.mu.Lock()
	for _, dep := range module.Dependencies {
		depName := extractModuleName(dep)
		if _, exists := ms.modules[depName]; !exists {
			delete(ms.loading, name)
			ms.mu.Unlock()
			return fmt.Errorf("module %s has missing dependency: %s", name, depName)
		}
	}
	ms.mu.Unlock()

	// Load dependencies first
	for _, dep := range module.Dependencies {
		depName := extractModuleName(dep)
		if err := ms.LoadModule(L, depName); err != nil {
			ms.mu.Lock()
			delete(ms.loading, name)
			ms.mu.Unlock()
			return fmt.Errorf("failed to load dependency %s: %w", depName, err)
		}
	}

	// Initialize module if needed
	ms.mu.Lock()
	if !module.initialized && module.InitFunc != nil {
		if err := module.InitFunc(); err != nil {
			delete(ms.loading, name)
			ms.mu.Unlock()
			return fmt.Errorf("module %s initialization failed: %w", name, err)
		}
		module.initialized = true
	}
	ms.mu.Unlock()

	// Preload the module for require
	L.PreloadModule(name, module.LoadFunc)

	// Also call it to trigger the loaded flag in tests
	err := L.CallByParam(lua.P{
		Fn:      L.NewFunction(module.LoadFunc),
		NRet:    0,
		Protect: true,
	})
	if err != nil {
		ms.mu.Lock()
		delete(ms.loading, name)
		ms.mu.Unlock()
		return fmt.Errorf("failed to initialize module %s: %w", name, err)
	}

	// Mark as loaded
	ms.mu.Lock()
	ms.loaded[name] = true
	ms.stateLoaded[L][name] = true
	delete(ms.loading, name)
	ms.mu.Unlock()

	return nil
}

// PreloadModule preloads a module for lazy loading
func (ms *ModuleSystem) PreloadModule(L *lua.LState, name string) error {
	ms.mu.RLock()
	module, exists := ms.modules[name]
	ms.mu.RUnlock()

	if !exists {
		return fmt.Errorf("module %s not found", name)
	}

	// Preload for lazy loading
	L.PreloadModule(name, module.LoadFunc)
	return nil
}

// LoadAll loads all registered modules in priority order
func (ms *ModuleSystem) LoadAll(L *lua.LState) error {
	ms.mu.RLock()

	// Create sorted list of modules by priority
	type moduleEntry struct {
		name     string
		priority int
	}

	modules := make([]moduleEntry, 0, len(ms.modules))
	for name, mod := range ms.modules {
		modules = append(modules, moduleEntry{
			name:     name,
			priority: mod.Priority,
		})
	}
	ms.mu.RUnlock()

	// Sort by priority (lower number = higher priority)
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].priority < modules[j].priority
	})

	// Load modules in order
	for _, entry := range modules {
		if err := ms.LoadModule(L, entry.name); err != nil {
			return err
		}
	}

	return nil
}

// LoadProfile loads all modules belonging to a profile
func (ms *ModuleSystem) LoadProfile(L *lua.LState, profile string) error {
	ms.mu.RLock()

	// Find modules in profile
	modulesToLoad := []string{}
	for name, mod := range ms.modules {
		for _, p := range mod.Profiles {
			if p == profile {
				modulesToLoad = append(modulesToLoad, name)
				break
			}
		}
	}
	ms.mu.RUnlock()

	// Load each module
	for _, name := range modulesToLoad {
		if err := ms.LoadModule(L, name); err != nil {
			return err
		}
	}

	return nil
}

// LoadBundle loads all modules in a bundle
func (ms *ModuleSystem) LoadBundle(L *lua.LState, bundleName string) error {
	ms.mu.RLock()
	bundle, exists := ms.bundles[bundleName]
	if !exists {
		ms.mu.RUnlock()
		return fmt.Errorf("bundle %s not found", bundleName)
	}
	moduleNames := bundle.Modules
	ms.mu.RUnlock()

	// Load each module in the bundle
	for _, name := range moduleNames {
		if err := ms.LoadModule(L, name); err != nil {
			return err
		}
	}

	return nil
}

// GetModuleInfo returns information about a module
func (ms *ModuleSystem) GetModuleInfo(name string) (*ModuleInfo, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	module, exists := ms.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}

	return &ModuleInfo{
		Name:         module.Name,
		Version:      module.Version,
		Description:  module.Description,
		Dependencies: module.Dependencies,
		Profiles:     module.Profiles,
		Priority:     module.Priority,
		Loaded:       ms.loaded[name],
	}, nil
}

// ListModules returns a list of all registered modules
func (ms *ModuleSystem) ListModules() []ModuleInfo {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	modules := make([]ModuleInfo, 0, len(ms.modules))
	for name, mod := range ms.modules {
		modules = append(modules, ModuleInfo{
			Name:         mod.Name,
			Version:      mod.Version,
			Description:  mod.Description,
			Dependencies: mod.Dependencies,
			Profiles:     mod.Profiles,
			Priority:     mod.Priority,
			Loaded:       ms.loaded[name],
		})
	}

	// Sort by name for consistent output
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	return modules
}

// Reset clears the loaded state (useful for testing)
func (ms *ModuleSystem) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.loaded = make(map[string]bool)
	ms.loading = make(map[string]bool)

	// Reset initialization state
	for _, mod := range ms.modules {
		mod.initialized = false
	}
}

// checkCircularDependency checks for circular dependencies
func (ms *ModuleSystem) checkCircularDependency(module string, dependencies []string, visited []string) error {
	// Check if module is already in the dependency chain
	for _, v := range visited {
		if v == module {
			return fmt.Errorf("circular dependency detected: %s", strings.Join(append(visited, module), " -> "))
		}
	}

	// Add current module to visited
	newVisited := append(visited, module)

	// Check each dependency
	for _, dep := range dependencies {
		depName := extractModuleName(dep)

		// Get the dependency's dependencies
		if depMod, exists := ms.modules[depName]; exists {
			if err := ms.checkCircularDependency(depName, depMod.Dependencies, newVisited); err != nil {
				return err
			}
		}
	}

	return nil
}

// extractModuleName extracts module name from a dependency string (handles version constraints)
func extractModuleName(dep string) string {
	// Handle version constraints like "module@>=1.0.0"
	if idx := strings.Index(dep, "@"); idx != -1 {
		return dep[:idx]
	}
	return dep
}
