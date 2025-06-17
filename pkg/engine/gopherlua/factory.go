// ABOUTME: LStateFactory creates and configures new Lua VM instances with security and optimization
// ABOUTME: Handles library loading, initialization scripts, and warmup strategies for performance

package gopherlua

import (
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// WarmupFunc is a function that warms up a newly created LState
type WarmupFunc func(L *lua.LState) error

// FactoryConfig configures how LStates are created
type FactoryConfig struct {
	// SecurityManager determines which libraries and functions are available
	SecurityManager *SecurityManager

	// Options for the Lua VM
	Options lua.Options

	// RegistrySize overrides the default registry size (0 uses default)
	RegistrySize int

	// InitScript runs after libraries are loaded
	InitScript string

	// PreloadModules are modules to preload but not require
	PreloadModules map[string]lua.LGFunction

	// WarmupFunc runs after initialization to optimize performance
	WarmupFunc WarmupFunc
}

// LStateFactory creates configured Lua VM instances
type LStateFactory struct {
	mu     sync.RWMutex
	config FactoryConfig
}

// NewLStateFactory creates a new factory with the given configuration
func NewLStateFactory(config FactoryConfig) *LStateFactory {
	// Apply defaults
	if config.RegistrySize > 0 {
		config.Options.RegistrySize = config.RegistrySize
	}

	// Use default SecurityManager if none provided
	if config.SecurityManager == nil {
		config.SecurityManager = NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		})
	}

	return &LStateFactory{
		config: config,
	}
}

// Create creates a new configured LState
func (f *LStateFactory) Create() (*lua.LState, error) {
	f.mu.RLock()
	config := f.config
	f.mu.RUnlock()

	// Create state with options - skip open libs so we can control what's loaded
	opts := config.Options
	opts.SkipOpenLibs = true
	L := lua.NewState(opts)
	if L == nil {
		return nil, fmt.Errorf("failed to create Lua state")
	}

	// Load libraries using SecurityManager
	if err := config.SecurityManager.LoadLibraries(L); err != nil {
		L.Close()
		return nil, fmt.Errorf("failed to load libraries: %w", err)
	}

	// Apply security sandbox
	if err := config.SecurityManager.ApplySandbox(L); err != nil {
		L.Close()
		return nil, fmt.Errorf("failed to apply security sandbox: %w", err)
	}

	// Preload custom modules
	for name, loader := range config.PreloadModules {
		L.PreloadModule(name, loader)
	}

	// Execute init script if provided
	if config.InitScript != "" {
		if err := L.DoString(config.InitScript); err != nil {
			L.Close()
			return nil, fmt.Errorf("init script failed: %w", err)
		}
	}

	// Run warmup function if provided
	if config.WarmupFunc != nil {
		if err := config.WarmupFunc(L); err != nil {
			L.Close()
			return nil, fmt.Errorf("warmup failed: %w", err)
		}
	}

	return L, nil
}

// Reset updates the factory configuration
func (f *LStateFactory) Reset(config FactoryConfig) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if config.RegistrySize > 0 {
		config.Options.RegistrySize = config.RegistrySize
	}

	// Use default SecurityManager if none provided
	if config.SecurityManager == nil {
		config.SecurityManager = NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		})
	}

	f.config = config
}
