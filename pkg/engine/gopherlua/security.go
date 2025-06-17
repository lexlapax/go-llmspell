// ABOUTME: SecurityManager enforces security policies for Lua VM instances
// ABOUTME: Provides configurable library restrictions, resource limits, and sandboxing

package gopherlua

import (
	"fmt"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// SecurityLevel defines the security restrictions level
type SecurityLevel int

const (
	// SecurityLevelMinimal provides basic restrictions with most libraries available
	SecurityLevelMinimal SecurityLevel = iota
	// SecurityLevelStandard removes file/network access, safe libraries only
	SecurityLevelStandard
	// SecurityLevelStrict provides minimal libraries with aggressive limits
	SecurityLevelStrict
	// SecurityLevelCustom allows custom configuration
	SecurityLevelCustom
)

// Security profile names
const (
	SecurityProfileMinimal  = "minimal"
	SecurityProfileStandard = "standard"
	SecurityProfileStrict   = "strict"
)

// ResourceLimits defines execution resource limits
type ResourceLimits struct {
	// MaxInstructions is the maximum number of VM instructions allowed
	MaxInstructions int64
	// MaxMemory is the maximum memory in bytes (not strictly enforced by Lua)
	MaxMemory int64
	// MaxDuration is the maximum execution time
	MaxDuration time.Duration
	// MaxStackDepth is the maximum call stack depth
	MaxStackDepth int
	// CheckInterval is how often to check limits (in instructions)
	CheckInterval int
}

// SecurityConfig defines the security configuration
type SecurityConfig struct {
	// Level is the security level
	Level SecurityLevel
	// AllowedLibraries is the list of allowed Lua libraries
	AllowedLibraries []string
	// DeniedFunctions is a map of function paths to deny (e.g., "os.execute")
	DeniedFunctions map[string]bool
	// ResourceLimits defines execution limits
	ResourceLimits ResourceLimits
}

// SecurityManager manages security policies for Lua states
type SecurityManager struct {
	mu     sync.RWMutex
	config SecurityConfig
}

// ResourceMonitor monitors resource usage
type ResourceMonitor struct {
	instructionCount int64
	memUsed          int64
	startTime        time.Time
	limits           ResourceLimits
	mu               sync.Mutex
}

// Predefined security profiles
var securityProfiles = map[string]SecurityConfig{
	SecurityProfileMinimal: {
		Level:            SecurityLevelMinimal,
		AllowedLibraries: []string{"base", "coroutine", "table", "io", "os", "string", "math", "package"},
		DeniedFunctions: map[string]bool{
			// Still deny some dangerous functions
			"os.execute": true,
			"os.exit":    true,
			"os.setenv":  true,
		},
		ResourceLimits: ResourceLimits{
			MaxInstructions: 100_000_000,
			MaxMemory:       100 * 1024 * 1024, // 100MB
			MaxDuration:     5 * time.Minute,
			MaxStackDepth:   1000,
			CheckInterval:   10000,
		},
	},
	SecurityProfileStandard: {
		Level:            SecurityLevelStandard,
		AllowedLibraries: []string{"base", "coroutine", "table", "string", "math", "os", "package"},
		DeniedFunctions: map[string]bool{
			"os.execute": true,
			"os.exit":    true,
			"os.setenv":  true,
			"os.remove":  true,
			"os.rename":  true,
			"os.tmpname": true,
			"os.getenv":  true,
		},
		ResourceLimits: ResourceLimits{
			MaxInstructions: 10_000_000,
			MaxMemory:       50 * 1024 * 1024, // 50MB
			MaxDuration:     30 * time.Second,
			MaxStackDepth:   500,
			CheckInterval:   5000,
		},
	},
	SecurityProfileStrict: {
		Level:            SecurityLevelStrict,
		AllowedLibraries: []string{"base", "coroutine", "table", "string", "math", "package"},
		DeniedFunctions:  map[string]bool{}, // OS library not loaded at all
		ResourceLimits: ResourceLimits{
			MaxInstructions: 1_000_000,
			MaxMemory:       10 * 1024 * 1024, // 10MB
			MaxDuration:     5 * time.Second,
			MaxStackDepth:   100,
			CheckInterval:   1000,
		},
	},
}

// NewSecurityManager creates a new security manager with the given config
func NewSecurityManager(config SecurityConfig) *SecurityManager {
	// Apply defaults based on security level if not provided
	if config.Level != SecurityLevelCustom {
		var profileName string
		switch config.Level {
		case SecurityLevelMinimal:
			profileName = SecurityProfileMinimal
		case SecurityLevelStandard:
			profileName = SecurityProfileStandard
		case SecurityLevelStrict:
			profileName = SecurityProfileStrict
		default:
			profileName = SecurityProfileStandard
		}

		if profile, ok := securityProfiles[profileName]; ok {
			// Apply profile defaults only if not explicitly set
			if config.ResourceLimits.CheckInterval == 0 {
				config.ResourceLimits = profile.ResourceLimits
			}
			if len(config.AllowedLibraries) == 0 {
				config.AllowedLibraries = profile.AllowedLibraries
			}
			if len(config.DeniedFunctions) == 0 {
				config.DeniedFunctions = profile.DeniedFunctions
			}
		}
	}

	// Apply minimum defaults
	if config.ResourceLimits.CheckInterval == 0 {
		config.ResourceLimits.CheckInterval = 1000
	}

	return &SecurityManager{
		config: config,
	}
}

// NewSecurityManagerFromProfile creates a security manager from a named profile
func NewSecurityManagerFromProfile(profile string) (*SecurityManager, error) {
	config, ok := securityProfiles[profile]
	if !ok {
		return nil, fmt.Errorf("unknown security profile: %s", profile)
	}
	return NewSecurityManager(config), nil
}

// LoadLibraries loads allowed libraries into the Lua state
func (sm *SecurityManager) LoadLibraries(L *lua.LState) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create safe library loader
	loader := NewSafeLibraryLoader(sm.config.Level)

	// Load libraries
	if err := loader.LoadLibraries(L, sm.config.AllowedLibraries); err != nil {
		return err
	}

	// Remove dangerous functions
	if err := loader.RemoveDangerousFunctions(L); err != nil {
		return err
	}

	// Apply custom safe replacements
	if err := loader.ApplyCustomReplacements(L); err != nil {
		return err
	}

	// Remove additional denied functions from config
	for funcPath := range sm.config.DeniedFunctions {
		sm.removeFunctionPath(L, funcPath)
	}

	return nil
}

// ApplySandbox applies security restrictions to a Lua state
func (sm *SecurityManager) ApplySandbox(L *lua.LState) error {
	// Load libraries with restrictions
	if err := sm.LoadLibraries(L); err != nil {
		return err
	}

	// Install resource monitoring hooks
	sm.InstallHooks(L)

	return nil
}

// InstallHooks installs resource monitoring hooks
// Note: GopherLua doesn't support SetHook, so this returns a monitor for manual checking
func (sm *SecurityManager) InstallHooks(L *lua.LState) *ResourceMonitor {
	sm.mu.RLock()
	limits := sm.config.ResourceLimits
	sm.mu.RUnlock()

	monitor := &ResourceMonitor{
		startTime: time.Now(),
		limits:    limits,
	}

	// TODO: Implement alternative resource monitoring without SetHook
	// Options:
	// 1. Use context with timeout for execution limits
	// 2. Periodic checking in long-running operations
	// 3. Manual instrumentation in bridge methods

	return monitor
}

// GetInstructionCount returns the current instruction count
func (rm *ResourceMonitor) GetInstructionCount() int64 {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.instructionCount
}

// removeFunctionPath removes a function at the given path (e.g., "os.execute")
func (sm *SecurityManager) removeFunctionPath(L *lua.LState, path string) {
	parts := []string{}
	current := ""
	for _, ch := range path {
		if ch == '.' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) < 2 {
		return
	}

	// Get the table
	table := L.GetGlobal(parts[0])
	if table == lua.LNil {
		return
	}

	// Navigate to parent table if nested
	for i := 1; i < len(parts)-1; i++ {
		if tbl, ok := table.(*lua.LTable); ok {
			table = tbl.RawGetString(parts[i])
			if table == lua.LNil {
				return
			}
		} else {
			return
		}
	}

	// Remove the function
	if tbl, ok := table.(*lua.LTable); ok {
		tbl.RawSetString(parts[len(parts)-1], lua.LNil)
	}
}
