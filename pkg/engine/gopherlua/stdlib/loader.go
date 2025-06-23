// ABOUTME: Loads embedded Lua stdlib modules and provides them as preloadable functions
// ABOUTME: Handles module caching, dependencies, and conversion to lua.LGFunction format

package stdlib

import (
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// moduleCache caches compiled Lua modules to avoid re-parsing.
var (
	moduleCache   = make(map[string]*lua.FunctionProto)
	moduleCacheMu sync.RWMutex
)

// LoadEmbeddedModule creates a lua.LGFunction that loads an embedded module.
// The returned function can be used with L.PreloadModule.
func LoadEmbeddedModule(moduleName string) (lua.LGFunction, error) {
	// Check if module exists
	if !ModuleExists(moduleName) {
		return nil, fmt.Errorf("module '%s' not found in embedded stdlib", moduleName)
	}

	// Return a loader function
	return func(L *lua.LState) int {
		// Check cache first
		moduleCacheMu.RLock()
		proto, cached := moduleCache[moduleName]
		moduleCacheMu.RUnlock()

		if !cached {
			// Read module content
			content, err := ReadEmbeddedModule(moduleName)
			if err != nil {
				L.RaiseError("failed to read module '%s': %v", moduleName, err)
				return 0
			}

			// Compile the module
			fn, err := L.LoadString(string(content))
			if err != nil {
				L.RaiseError("failed to compile module '%s': %v", moduleName, err)
				return 0
			}

			proto = fn.Proto

			// Cache the compiled proto
			moduleCacheMu.Lock()
			moduleCache[moduleName] = proto
			moduleCacheMu.Unlock()
		}

		// Create and push the function
		fn := L.NewFunctionFromProto(proto)
		L.Push(fn)

		// Call the function to get the module table
		L.Call(0, 1)

		// The module should have left its exports on the stack
		return 1
	}, nil
}

// moduleAliases defines common aliases for modules
var moduleAliases = map[string]string{
	"log": "logging", // Examples use 'log' but the file is 'logging.lua'
}

// GetAllStdlibLoaders returns a map of all stdlib module loaders.
// The map keys are module names and values are loader functions.
func GetAllStdlibLoaders() (map[string]lua.LGFunction, error) {
	modules, err := GetEmbeddedModules()
	if err != nil {
		return nil, fmt.Errorf("failed to get embedded modules: %w", err)
	}

	loaders := make(map[string]lua.LGFunction)

	for _, moduleName := range modules {
		loader, err := LoadEmbeddedModule(moduleName)
		if err != nil {
			return nil, fmt.Errorf("failed to create loader for module '%s': %w", moduleName, err)
		}
		loaders[moduleName] = loader
	}

	// Add aliases
	for alias, target := range moduleAliases {
		if targetLoader, exists := loaders[target]; exists {
			loaders[alias] = targetLoader
		}
	}

	return loaders, nil
}

// ClearModuleCache clears the module cache.
// This is mainly useful for testing.
func ClearModuleCache() {
	moduleCacheMu.Lock()
	defer moduleCacheMu.Unlock()
	moduleCache = make(map[string]*lua.FunctionProto)
}

// moduleLoadOrder defines the order in which modules should be loaded
// to handle dependencies correctly.
var moduleLoadOrder = []string{
	// Core modules first
	"core",
	"errors",
	"logging",
	"data",

	// Then other modules
	"auth",
	"state",
	"events",
	"tools",
	"llm",
	"agent",
	"observability",
	"spell",
	"promise",
	"testing",
}

// GetOrderedStdlibLoaders returns stdlib loaders in dependency order.
// Modules not in the defined order are added at the end.
func GetOrderedStdlibLoaders() ([]struct {
	Name   string
	Loader lua.LGFunction
}, error) {
	allLoaders, err := GetAllStdlibLoaders()
	if err != nil {
		return nil, err
	}

	ordered := make([]struct {
		Name   string
		Loader lua.LGFunction
	}, 0, len(allLoaders))
	loaded := make(map[string]bool)

	// First, load modules in defined order
	for _, moduleName := range moduleLoadOrder {
		if loader, exists := allLoaders[moduleName]; exists {
			ordered = append(ordered, struct {
				Name   string
				Loader lua.LGFunction
			}{
				Name:   moduleName,
				Loader: loader,
			})
			loaded[moduleName] = true
		}
	}

	// Then, load any remaining modules
	for moduleName, loader := range allLoaders {
		if !loaded[moduleName] {
			ordered = append(ordered, struct {
				Name   string
				Loader lua.LGFunction
			}{
				Name:   moduleName,
				Loader: loader,
			})
		}
	}

	return ordered, nil
}
