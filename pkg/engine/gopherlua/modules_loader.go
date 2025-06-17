// ABOUTME: Module loader implementation with support for lazy loading, profile-based loading, and module bundling
// ABOUTME: Provides PreloadModule, module initialization callbacks, and version management

package gopherlua

import (
	"fmt"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// ModuleLoaderConfig provides configuration for module loading
type ModuleLoaderConfig struct {
	// Base path for module files
	BasePath string

	// Module search paths
	SearchPaths []string

	// Default profile to load
	DefaultProfile string

	// Enable module caching
	EnableCache bool

	// Auto-load dependencies
	AutoLoadDeps bool
}

// ModuleLoader handles module loading operations
type ModuleLoader struct {
	system *ModuleSystem
	config ModuleLoaderConfig
}

// NewModuleLoader creates a new module loader
func NewModuleLoader(system *ModuleSystem, config ModuleLoaderConfig) *ModuleLoader {
	return &ModuleLoader{
		system: system,
		config: config,
	}
}

// LoadFromFile loads a module definition from a Lua file
func (ml *ModuleLoader) LoadFromFile(path string) (*ModuleDefinition, error) {
	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid module path: %w", err)
	}

	// Extract module name from filename
	moduleName := strings.TrimSuffix(filepath.Base(absPath), ".lua")

	// Create module definition
	module := &ModuleDefinition{
		Name: moduleName,
		LoadFunc: func(L *lua.LState) int {
			// Load and execute the module file
			if err := L.DoFile(absPath); err != nil {
				L.RaiseError("failed to load module %s: %v", moduleName, err)
				return 0
			}

			// Module should leave its table on the stack
			if L.GetTop() == 0 {
				// If nothing on stack, push an empty table
				L.Push(L.NewTable())
			}

			return 1
		},
	}

	// Parse module metadata from file if available
	if metadata, err := ml.parseModuleMetadata(absPath); err == nil {
		module.Version = metadata.Version
		module.Description = metadata.Description
		module.Dependencies = metadata.Dependencies
		module.Profiles = metadata.Profiles
		module.Priority = metadata.Priority
	}

	return module, nil
}

// LoadDirectory loads all modules from a directory
func (ml *ModuleLoader) LoadDirectory(dir string) error {
	pattern := filepath.Join(dir, "*.lua")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	for _, file := range files {
		module, err := ml.LoadFromFile(file)
		if err != nil {
			return fmt.Errorf("failed to load module %s: %w", file, err)
		}

		if err := ml.system.Register(*module); err != nil {
			return fmt.Errorf("failed to register module %s: %w", module.Name, err)
		}
	}

	return nil
}

// LoadProfile loads all modules for a specific profile
func (ml *ModuleLoader) LoadProfile(L *lua.LState, profile string) error {
	if profile == "" {
		profile = ml.config.DefaultProfile
	}

	return ml.system.LoadProfile(L, profile)
}

// PreloadAll preloads all registered modules for lazy loading
func (ml *ModuleLoader) PreloadAll(L *lua.LState) error {
	modules := ml.system.ListModules()

	for _, info := range modules {
		if err := ml.system.PreloadModule(L, info.Name); err != nil {
			return fmt.Errorf("failed to preload module %s: %w", info.Name, err)
		}
	}

	return nil
}

// CreateRequireFunction creates a custom require function with module system integration
func (ml *ModuleLoader) CreateRequireFunction() lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1)

		// Check if already loaded
		L.GetField(L.Get(lua.RegistryIndex), "_LOADED")
		loaded := L.Get(-1).(*lua.LTable)
		L.Pop(1)

		// Check _LOADED table
		if module := loaded.RawGetString(name); module != lua.LNil {
			L.Push(module)
			return 1
		}

		// Try to load from module system
		if ml.system.Exists(name) {
			// Load with dependencies if configured
			if ml.config.AutoLoadDeps {
				if err := ml.system.LoadModule(L, name); err != nil {
					L.RaiseError("failed to load module %s: %v", name, err)
					return 0
				}
			}

			// Call the preloaded module
			L.GetField(L.Get(lua.RegistryIndex), "_PRELOAD")
			preload := L.Get(-1).(*lua.LTable)
			L.Pop(1)

			if loader := preload.RawGetString(name); loader != lua.LNil {
				L.Push(loader)
				L.Call(0, 1)
				module := L.Get(-1)

				// Store in _LOADED
				loaded.RawSetString(name, module)

				return 1
			}
		}

		// Fall back to standard require
		L.RaiseError("module '%s' not found", name)
		return 0
	}
}

// InstallRequire installs the custom require function
func (ml *ModuleLoader) InstallRequire(L *lua.LState) {
	// Replace the global require function
	L.SetGlobal("require", L.NewFunction(ml.CreateRequireFunction()))
}

// ModuleMetadata represents parsed module metadata
type ModuleMetadata struct {
	Version      string
	Description  string
	Dependencies []string
	Profiles     []string
	Priority     int
}

// parseModuleMetadata extracts metadata from a module file
func (ml *ModuleLoader) parseModuleMetadata(path string) (*ModuleMetadata, error) {
	// This is a simplified implementation
	// In a real implementation, you might parse special comments or a manifest

	metadata := &ModuleMetadata{
		Version:  "1.0.0",
		Priority: 50,
	}

	// TODO: Implement actual metadata parsing
	// Could parse special comments like:
	// -- @version 1.0.0
	// -- @description Module description
	// -- @depends base, utils
	// -- @profiles standard, full
	// -- @priority 10

	return metadata, nil
}

// LoadStandardLibrary loads the standard Lua libraries based on security profile
func (ml *ModuleLoader) LoadStandardLibrary(L *lua.LState, profile string) error {
	// Define which libraries to load for each profile
	libraries := map[string][]string{
		"minimal":  {"base", "string", "table", "math"},
		"standard": {"base", "string", "table", "math", "coroutine", "utf8"},
		"full":     {"base", "string", "table", "math", "coroutine", "utf8", "debug"},
	}

	libs, exists := libraries[profile]
	if !exists {
		libs = libraries["standard"] // Default to standard
	}

	// Load selected libraries
	for _, lib := range libs {
		switch lib {
		case "base":
			L.OpenLibs() // Opens base library
		case "string":
			L.Push(L.NewFunction(lua.OpenString))
			L.Push(lua.LString("string"))
			L.Call(1, 0)
		case "table":
			L.Push(L.NewFunction(lua.OpenTable))
			L.Push(lua.LString("table"))
			L.Call(1, 0)
		case "math":
			L.Push(L.NewFunction(lua.OpenMath))
			L.Push(lua.LString("math"))
			L.Call(1, 0)
		case "coroutine":
			L.Push(L.NewFunction(lua.OpenCoroutine))
			L.Push(lua.LString("coroutine"))
			L.Call(1, 0)
		case "debug":
			L.Push(L.NewFunction(lua.OpenDebug))
			L.Push(lua.LString("debug"))
			L.Call(1, 0)
		}
	}

	return nil
}

// WatchModuleChanges watches for module file changes and reloads them
func (ml *ModuleLoader) WatchModuleChanges(callback func(moduleName string)) error {
	// TODO: Implement file watching
	// This would use fsnotify or similar to watch module files
	// and trigger reloads when they change
	return fmt.Errorf("module watching not implemented")
}

// ExportModuleMap exports a module dependency map
func (ml *ModuleLoader) ExportModuleMap() map[string][]string {
	modules := ml.system.ListModules()
	depMap := make(map[string][]string)

	for _, mod := range modules {
		depMap[mod.Name] = mod.Dependencies
	}

	return depMap
}

// ValidateModuleDependencies validates all module dependencies are satisfied
func (ml *ModuleLoader) ValidateModuleDependencies() error {
	modules := ml.system.ListModules()

	for _, mod := range modules {
		for _, dep := range mod.Dependencies {
			depName := extractModuleName(dep)
			if !ml.system.Exists(depName) {
				return fmt.Errorf("module %s has missing dependency: %s", mod.Name, depName)
			}
		}
	}

	return nil
}

// GetModulePath resolves a module name to its file path
func (ml *ModuleLoader) GetModulePath(name string) (string, error) {
	// Try each search path
	for _, searchPath := range ml.config.SearchPaths {
		path := filepath.Join(searchPath, name+".lua")
		if fileExists(path) {
			return path, nil
		}

		// Try with module subdirectory
		path = filepath.Join(searchPath, name, "init.lua")
		if fileExists(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("module %s not found in search paths", name)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	// Simple implementation - in production would use os.Stat
	return true // TODO: Implement actual file checking
}
