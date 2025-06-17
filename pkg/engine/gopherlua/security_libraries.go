// ABOUTME: SafeLibraryLoader handles secure loading of Lua standard libraries
// ABOUTME: Removes dangerous functions and provides safe replacements based on security level

package gopherlua

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// SafeLibraryLoader manages library loading with security restrictions
type SafeLibraryLoader struct {
	level           SecurityLevel
	libraryLoaders  map[string]lua.LGFunction
	deniedFunctions map[string][]string // library -> functions to remove
	replacements    map[string]lua.LGFunction
}

// NewSafeLibraryLoader creates a new safe library loader
func NewSafeLibraryLoader(level SecurityLevel) *SafeLibraryLoader {
	loader := &SafeLibraryLoader{
		level: level,
		libraryLoaders: map[string]lua.LGFunction{
			"base":      lua.OpenBase,
			"coroutine": lua.OpenCoroutine,
			"table":     lua.OpenTable,
			"io":        lua.OpenIo,
			"os":        lua.OpenOs,
			"string":    lua.OpenString,
			"math":      lua.OpenMath,
			"package":   lua.OpenPackage,
		},
		deniedFunctions: make(map[string][]string),
		replacements:    make(map[string]lua.LGFunction),
	}

	// Configure denied functions based on security level
	loader.configureDeniedFunctions()

	// Set up safe replacements
	loader.configureReplacements()

	return loader
}

// LoadLibraries loads the specified libraries with security restrictions
func (sll *SafeLibraryLoader) LoadLibraries(L *lua.LState, libraries []string) error {
	for _, libName := range libraries {
		// Skip debug library completely
		if libName == "debug" {
			continue
		}

		// Skip libraries not allowed at this security level
		if !sll.isLibraryAllowed(libName) {
			continue
		}

		// Load the library
		if loader, ok := sll.libraryLoaders[libName]; ok {
			if err := L.CallByParam(lua.P{
				Fn:      L.NewFunction(loader),
				NRet:    0,
				Protect: true,
			}, lua.LString(libName)); err != nil {
				return fmt.Errorf("failed to load %s library: %w", libName, err)
			}
		}
	}

	return nil
}

// RemoveDangerousFunctions removes functions that are considered dangerous
func (sll *SafeLibraryLoader) RemoveDangerousFunctions(L *lua.LState) error {
	for libName, functions := range sll.deniedFunctions {
		lib := L.GetGlobal(libName)
		if lib == lua.LNil {
			continue
		}

		if tbl, ok := lib.(*lua.LTable); ok {
			for _, funcName := range functions {
				tbl.RawSetString(funcName, lua.LNil)
			}
		}
	}

	return nil
}

// ApplyCustomReplacements installs safe replacement functions
func (sll *SafeLibraryLoader) ApplyCustomReplacements(L *lua.LState) error {
	// Replace dangerous global functions with safe versions
	for name, replacement := range sll.replacements {
		L.SetGlobal(name, L.NewFunction(replacement))
	}

	return nil
}

// configureDeniedFunctions sets up the list of functions to remove based on security level
func (sll *SafeLibraryLoader) configureDeniedFunctions() {
	switch sll.level {
	case SecurityLevelMinimal:
		// Minimal restrictions - only the most dangerous functions
		sll.deniedFunctions["os"] = []string{
			"execute",
			"exit",
			"setenv",
		}

	case SecurityLevelStandard:
		// Standard restrictions - no file/system access
		sll.deniedFunctions["os"] = []string{
			"execute",
			"exit",
			"setenv",
			"remove",
			"rename",
			"tmpname",
			"getenv",
		}

	case SecurityLevelStrict:
		// Strict restrictions - minimal functionality
		// Don't load OS/IO at all
		sll.deniedFunctions["os"] = []string{} // Will be blocked at library level
		sll.deniedFunctions["io"] = []string{} // Will be blocked at library level
		sll.deniedFunctions["package"] = []string{
			"loadlib",
			"searchpath",
		}
	}
}

// configureReplacements sets up safe replacement functions
func (sll *SafeLibraryLoader) configureReplacements() {
	// Safe print that captures output
	sll.replacements["print"] = func(L *lua.LState) int {
		n := L.GetTop()
		parts := make([]string, n)

		for i := 1; i <= n; i++ {
			parts[i-1] = L.ToStringMeta(L.Get(i)).String()
		}

		output := strings.Join(parts, "\t")

		// Store in a global table for testing
		testOutput := L.GetGlobal("_test_output")
		if testOutput != lua.LNil {
			if tbl, ok := testOutput.(*lua.LTable); ok {
				tbl.Append(lua.LString(output))
			}
		}

		// In production, this could log to a safe location
		// For now, we just capture it

		return 0
	}

	// Block require in strict mode
	if sll.level == SecurityLevelStrict {
		sll.replacements["require"] = func(L *lua.LState) int {
			L.RaiseError("require is disabled in strict security mode")
			return 0
		}
	}

	// Safe dofile that only allows whitelisted files
	sll.replacements["dofile"] = func(L *lua.LState) int {
		L.RaiseError("dofile is disabled for security")
		return 0
	}

	// Safe loadfile
	sll.replacements["loadfile"] = func(L *lua.LState) int {
		L.RaiseError("loadfile is disabled for security")
		return 0
	}

	// Safe load/loadstring with restrictions
	if sll.level == SecurityLevelStrict {
		sll.replacements["load"] = func(L *lua.LState) int {
			L.RaiseError("load is disabled in strict security mode")
			return 0
		}
		sll.replacements["loadstring"] = func(L *lua.LState) int {
			L.RaiseError("loadstring is disabled in strict security mode")
			return 0
		}
	}
}

// isLibraryAllowed checks if a library is allowed at the current security level
func (sll *SafeLibraryLoader) isLibraryAllowed(libName string) bool {
	switch sll.level {
	case SecurityLevelStrict:
		// In strict mode, no IO or OS libraries
		return libName != "io" && libName != "os" && libName != "debug"

	case SecurityLevelStandard:
		// In standard mode, no IO or debug
		return libName != "io" && libName != "debug"

	case SecurityLevelMinimal:
		// In minimal mode, only debug is blocked
		return libName != "debug"

	default:
		return true
	}
}
