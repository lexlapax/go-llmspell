// ABOUTME: LStateFactory creates and configures new Lua VM instances with security and optimization
// ABOUTME: Handles library loading, initialization scripts, and warmup strategies for performance

package gopherlua

// waiting to be implemented after task 2.2.1 and 2.2.2 are done

// import (
// 	"fmt"
// 	"sync"

// 	lua "github.com/yuin/gopher-lua"
// )

// // SecurityLevel defines the security restrictions for Lua states
// type SecurityLevel int

// const (
// 	// SecurityLevelMinimal provides basic restrictions with most libraries available
// 	SecurityLevelMinimal SecurityLevel = iota
// 	// SecurityLevelStandard removes file/network access, safe libraries only
// 	SecurityLevelStandard
// 	// SecurityLevelStrict provides minimal libraries with aggressive limits
// 	SecurityLevelStrict
// 	// SecurityLevelCustom allows custom security configuration
// 	SecurityLevelCustom
// )

// // WarmupFunc is a function that warms up a newly created LState
// type WarmupFunc func(L *lua.LState) error

// // FactoryConfig configures how LStates are created
// type FactoryConfig struct {
// 	// SecurityLevel determines which libraries and functions are available
// 	SecurityLevel SecurityLevel

// 	// Options for the Lua VM
// 	Options lua.Options

// 	// RegistrySize overrides the default registry size (0 uses default)
// 	RegistrySize int

// 	// InitScript runs after libraries are loaded
// 	InitScript string

// 	// PreloadModules are modules to preload but not require
// 	PreloadModules map[string]lua.LGFunction

// 	// WarmupFunc runs after initialization to optimize performance
// 	WarmupFunc WarmupFunc
// }

// // LStateFactory creates configured Lua VM instances
// type LStateFactory struct {
// 	mu     sync.RWMutex
// 	config FactoryConfig
// }

// // NewLStateFactory creates a new factory with the given configuration
// func NewLStateFactory(config FactoryConfig) *LStateFactory {
// 	// Apply defaults
// 	if config.RegistrySize > 0 {
// 		config.Options.RegistrySize = config.RegistrySize
// 	}

// 	return &LStateFactory{
// 		config: config,
// 	}
// }

// // Create creates a new configured LState
// func (f *LStateFactory) Create() (*lua.LState, error) {
// 	f.mu.RLock()
// 	config := f.config
// 	f.mu.RUnlock()

// 	// Create state with options - skip open libs so we can control what's loaded
// 	opts := config.Options
// 	opts.SkipOpenLibs = true
// 	L := lua.NewState(opts)
// 	if L == nil {
// 		return nil, fmt.Errorf("failed to create Lua state")
// 	}

// 	// Load libraries based on security level
// 	if err := f.loadLibraries(L, config.SecurityLevel); err != nil {
// 		L.Close()
// 		return nil, fmt.Errorf("failed to load libraries: %w", err)
// 	}

// 	// Preload custom modules
// 	for name, loader := range config.PreloadModules {
// 		L.PreloadModule(name, loader)
// 	}

// 	// Execute init script if provided
// 	if config.InitScript != "" {
// 		if err := L.DoString(config.InitScript); err != nil {
// 			L.Close()
// 			return nil, fmt.Errorf("init script failed: %w", err)
// 		}
// 	}

// 	// Run warmup function if provided
// 	if config.WarmupFunc != nil {
// 		if err := config.WarmupFunc(L); err != nil {
// 			L.Close()
// 			return nil, fmt.Errorf("warmup failed: %w", err)
// 		}
// 	}

// 	return L, nil
// }

// // Reset updates the factory configuration
// func (f *LStateFactory) Reset(config FactoryConfig) {
// 	f.mu.Lock()
// 	defer f.mu.Unlock()

// 	if config.RegistrySize > 0 {
// 		config.Options.RegistrySize = config.RegistrySize
// 	}
// 	f.config = config
// }

// // loadLibraries loads appropriate libraries based on security level
// func (f *LStateFactory) loadLibraries(L *lua.LState, level SecurityLevel) error {
// 	switch level {
// 	case SecurityLevelMinimal:
// 		// Load most libraries except debug
// 		for _, pair := range []struct {
// 			name string
// 			fn   lua.LGFunction
// 		}{
// 			{lua.BaseLibName, lua.OpenBase},
// 			{lua.TabLibName, lua.OpenTable},
// 			{lua.StringLibName, lua.OpenString},
// 			{lua.MathLibName, lua.OpenMath},
// 			{lua.CoroutineLibName, lua.OpenCoroutine},
// 			{lua.IoLibName, lua.OpenIo},
// 			{lua.OsLibName, lua.OpenOs},
// 			{lua.LoadLibName, lua.OpenPackage}, // Required for module loading
// 		} {
// 			if err := L.CallByParam(lua.P{
// 				Fn:      L.NewFunction(pair.fn),
// 				NRet:    0,
// 				Protect: true,
// 			}, lua.LString(pair.name)); err != nil {
// 				return fmt.Errorf("failed to load %s library: %w", pair.name, err)
// 			}
// 		}

// 	case SecurityLevelStandard:
// 		// Load safe libraries only
// 		for _, pair := range []struct {
// 			name string
// 			fn   lua.LGFunction
// 		}{
// 			{lua.BaseLibName, lua.OpenBase},
// 			{lua.TabLibName, lua.OpenTable},
// 			{lua.StringLibName, lua.OpenString},
// 			{lua.MathLibName, lua.OpenMath},
// 			{lua.CoroutineLibName, lua.OpenCoroutine},
// 			{lua.OsLibName, lua.OpenOs},        // Will sanitize below
// 			{lua.LoadLibName, lua.OpenPackage}, // Required for module loading
// 		} {
// 			if err := L.CallByParam(lua.P{
// 				Fn:      L.NewFunction(pair.fn),
// 				NRet:    0,
// 				Protect: true,
// 			}, lua.LString(pair.name)); err != nil {
// 				return fmt.Errorf("failed to load %s library: %w", pair.name, err)
// 			}
// 		}

// 		// Sanitize OS library
// 		f.sanitizeOSLibrary(L)

// 	case SecurityLevelStrict:
// 		// Load minimal libraries only
// 		for _, pair := range []struct {
// 			name string
// 			fn   lua.LGFunction
// 		}{
// 			{lua.BaseLibName, lua.OpenBase},
// 			{lua.TabLibName, lua.OpenTable},
// 			{lua.StringLibName, lua.OpenString},
// 			{lua.MathLibName, lua.OpenMath},
// 			{lua.CoroutineLibName, lua.OpenCoroutine},
// 			{lua.LoadLibName, lua.OpenPackage}, // Required for module loading
// 		} {
// 			if err := L.CallByParam(lua.P{
// 				Fn:      L.NewFunction(pair.fn),
// 				NRet:    0,
// 				Protect: true,
// 			}, lua.LString(pair.name)); err != nil {
// 				return fmt.Errorf("failed to load %s library: %w", pair.name, err)
// 			}
// 		}

// 	default:
// 		return fmt.Errorf("unknown security level: %d", level)
// 	}

// 	return nil
// }

// // sanitizeOSLibrary removes dangerous functions from the OS library
// func (f *LStateFactory) sanitizeOSLibrary(L *lua.LState) {
// 	osTable := L.GetGlobal("os")
// 	if osTable == lua.LNil {
// 		return
// 	}

// 	if tbl, ok := osTable.(*lua.LTable); ok {
// 		// Remove dangerous functions
// 		dangerousFuncs := []string{
// 			"execute",
// 			"exit",
// 			"setenv",
// 			"remove",
// 			"rename",
// 		}

// 		for _, fn := range dangerousFuncs {
// 			tbl.RawSetString(fn, lua.LNil)
// 		}
// 	}
// }
