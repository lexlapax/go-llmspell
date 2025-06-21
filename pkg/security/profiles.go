// ABOUTME: Security profiles implementation for controlling script execution permissions and resource limits.
// ABOUTME: Provides sandbox, development, and production profiles with configurable restrictions.

package security

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Permission represents a specific permission type
type Permission string

const (
	PermissionNetwork     Permission = "network"
	PermissionFilesystem  Permission = "filesystem"
	PermissionEnvironment Permission = "environment"
	PermissionExec        Permission = "exec"
	PermissionUnsafe      Permission = "unsafe"
)

// SecurityProfile defines security constraints for script execution
type SecurityProfile struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`

	// Permission flags
	AllowNetwork     bool `json:"allow_network" yaml:"allow_network"`
	AllowFilesystem  bool `json:"allow_filesystem" yaml:"allow_filesystem"`
	AllowEnvironment bool `json:"allow_environment" yaml:"allow_environment"`
	AllowExec        bool `json:"allow_exec" yaml:"allow_exec"`
	AllowUnsafe      bool `json:"allow_unsafe" yaml:"allow_unsafe"`

	// Resource limits
	MemoryLimit    int64 `json:"memory_limit" yaml:"memory_limit"`       // In bytes
	CPULimit       int   `json:"cpu_limit" yaml:"cpu_limit"`             // Percentage (1-100)
	TimeoutSeconds int   `json:"timeout_seconds" yaml:"timeout_seconds"` // Maximum execution time

	// Module restrictions
	AllowedModules     []string `json:"allowed_modules" yaml:"allowed_modules"`
	ForbiddenFunctions []string `json:"forbidden_functions" yaml:"forbidden_functions"`

	// Path restrictions
	AllowedPaths   []string `json:"allowed_paths" yaml:"allowed_paths"`
	ForbiddenPaths []string `json:"forbidden_paths" yaml:"forbidden_paths"`
}

// SandboxProfile returns a highly restrictive security profile
func SandboxProfile() *SecurityProfile {
	return &SecurityProfile{
		Name:        "sandbox",
		Description: "Maximum security restrictions for untrusted scripts",

		// All permissions denied
		AllowNetwork:     false,
		AllowFilesystem:  false,
		AllowEnvironment: false,
		AllowExec:        false,
		AllowUnsafe:      false,

		// Tight resource limits
		MemoryLimit:    64 * 1024 * 1024, // 64MB
		CPULimit:       5,                // 5% CPU
		TimeoutSeconds: 30,               // 30 seconds

		// Limited modules
		AllowedModules: []string{
			"string", "table", "math", "utf8",
			"coroutine", "bit32", "bit",
		},
		ForbiddenFunctions: []string{
			"loadstring", "load", "loadfile",
			"dofile", "require", "module",
			"rawget", "rawset", "rawequal",
			"getmetatable", "setmetatable",
			"getfenv", "setfenv", "debug",
		},

		// No filesystem access
		AllowedPaths:   []string{},
		ForbiddenPaths: []string{"/"},
	}
}

// DevelopmentProfile returns a moderately permissive profile for development
func DevelopmentProfile() *SecurityProfile {
	return &SecurityProfile{
		Name:        "development",
		Description: "Balanced security for development and testing",

		// More permissions
		AllowNetwork:     true,
		AllowFilesystem:  true,
		AllowEnvironment: true,
		AllowExec:        false, // Still restricted
		AllowUnsafe:      false,

		// Higher resource limits
		MemoryLimit:    256 * 1024 * 1024, // 256MB
		CPULimit:       50,                // 50% CPU
		TimeoutSeconds: 300,               // 5 minutes

		// More modules allowed
		AllowedModules: []string{
			"string", "table", "math", "utf8",
			"coroutine", "bit32", "bit",
			"os", "io", "debug", "package",
		},
		ForbiddenFunctions: []string{
			"os.execute", "io.popen",
		},

		// Some filesystem access
		AllowedPaths: []string{
			"/tmp", "/var/tmp",
			"./", // Current directory
		},
		ForbiddenPaths: []string{
			"/etc", "/sys", "/proc",
			"/root", "/home",
		},
	}
}

// ProductionProfile returns a production-ready security profile
func ProductionProfile() *SecurityProfile {
	return &SecurityProfile{
		Name:        "production",
		Description: "Production security settings with network access",

		// Limited permissions
		AllowNetwork:     true,
		AllowFilesystem:  false,
		AllowEnvironment: false,
		AllowExec:        false,
		AllowUnsafe:      false,

		// Production resource limits
		MemoryLimit:    512 * 1024 * 1024, // 512MB
		CPULimit:       25,                // 25% CPU
		TimeoutSeconds: 60,                // 1 minute

		// Production modules
		AllowedModules: []string{
			"string", "table", "math", "utf8",
			"coroutine", "bit32", "bit",
			"json", "http", "crypto",
		},
		ForbiddenFunctions: []string{
			"loadstring", "load", "loadfile",
			"dofile", "debug",
		},

		// No filesystem access in production
		AllowedPaths:   []string{},
		ForbiddenPaths: []string{"/"},
	}
}

// Validate checks if the security profile is valid
func (p *SecurityProfile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if p.MemoryLimit <= 0 {
		return fmt.Errorf("memory limit must be positive")
	}

	if p.CPULimit <= 0 {
		return fmt.Errorf("CPU limit must be positive")
	}
	if p.CPULimit > 100 {
		return fmt.Errorf("CPU limit must not exceed 100")
	}

	if p.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// CheckPermission checks if a permission is allowed
func (p *SecurityProfile) CheckPermission(perm Permission) bool {
	switch perm {
	case PermissionNetwork:
		return p.AllowNetwork
	case PermissionFilesystem:
		return p.AllowFilesystem
	case PermissionEnvironment:
		return p.AllowEnvironment
	case PermissionExec:
		return p.AllowExec
	case PermissionUnsafe:
		return p.AllowUnsafe
	default:
		return false
	}
}

// IsModuleAllowed checks if a module is allowed
func (p *SecurityProfile) IsModuleAllowed(module string) bool {
	for _, allowed := range p.AllowedModules {
		if allowed == module {
			return true
		}
	}
	return false
}

// IsFunctionForbidden checks if a function is forbidden
func (p *SecurityProfile) IsFunctionForbidden(function string) bool {
	for _, forbidden := range p.ForbiddenFunctions {
		if forbidden == function {
			return true
		}
	}
	return false
}

// IsPathAllowed checks if a path is allowed
func (p *SecurityProfile) IsPathAllowed(path string) bool {
	// First check forbidden paths
	for _, forbidden := range p.ForbiddenPaths {
		if strings.HasPrefix(path, forbidden) {
			return false
		}
	}

	// If no allowed paths specified, default to forbidden
	if len(p.AllowedPaths) == 0 {
		return false
	}

	// Check allowed paths
	for _, allowed := range p.AllowedPaths {
		if strings.HasPrefix(path, allowed) {
			return true
		}
	}

	return false
}

// ProfileManager manages security profiles
type ProfileManager struct {
	mu       sync.RWMutex
	profiles map[string]*SecurityProfile
}

// NewProfileManager creates a new profile manager with default profiles
func NewProfileManager() *ProfileManager {
	pm := &ProfileManager{
		profiles: make(map[string]*SecurityProfile),
	}

	// Register default profiles
	pm.profiles["sandbox"] = SandboxProfile()
	pm.profiles["development"] = DevelopmentProfile()
	pm.profiles["production"] = ProductionProfile()

	return pm
}

// GetProfile retrieves a security profile by name
func (pm *ProfileManager) GetProfile(name string) (*SecurityProfile, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	profile, exists := pm.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile not found: %s", name)
	}

	return profile, nil
}

// RegisterProfile registers a new security profile
func (pm *ProfileManager) RegisterProfile(profile *SecurityProfile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.profiles[profile.Name]; exists {
		return fmt.Errorf("profile %s already exists", profile.Name)
	}

	pm.profiles[profile.Name] = profile
	return nil
}

// RemoveProfile removes a security profile
func (pm *ProfileManager) RemoveProfile(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Don't allow removing default profiles
	if name == "sandbox" || name == "development" || name == "production" {
		return fmt.Errorf("cannot remove default profile: %s", name)
	}

	if _, exists := pm.profiles[name]; !exists {
		return fmt.Errorf("profile not found: %s", name)
	}

	delete(pm.profiles, name)
	return nil
}

// ListProfiles returns a list of all profile names
func (pm *ProfileManager) ListProfiles() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, 0, len(pm.profiles))
	for name := range pm.profiles {
		names = append(names, name)
	}
	return names
}

// SecurityViolation represents a security policy violation
type SecurityViolation struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// SecurityContext tracks security state during execution
type SecurityContext struct {
	Profile    *SecurityProfile
	Violations []SecurityViolation
	mu         sync.Mutex
}

// NewSecurityContext creates a new security context
func NewSecurityContext(profile *SecurityProfile) *SecurityContext {
	return &SecurityContext{
		Profile:    profile,
		Violations: make([]SecurityViolation, 0),
	}
}

// RecordViolation records a security violation
func (ctx *SecurityContext) RecordViolation(violationType, description string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.Violations = append(ctx.Violations, SecurityViolation{
		Type:        violationType,
		Description: description,
		Timestamp:   time.Now(),
	})
}

// CheckAndRecord checks a permission and records a violation if denied
func (ctx *SecurityContext) CheckAndRecord(perm Permission, description string) bool {
	allowed := ctx.Profile.CheckPermission(perm)
	if !allowed {
		ctx.RecordViolation(string(perm), description)
	}
	return allowed
}

// HasViolations returns true if any violations have been recorded
func (ctx *SecurityContext) HasViolations() bool {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return len(ctx.Violations) > 0
}

// GetViolationSummary returns a summary of recorded violations
func (ctx *SecurityContext) GetViolationSummary() string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if len(ctx.Violations) == 0 {
		return "No security violations"
	}

	// Count violations by type
	counts := make(map[string]int)
	for _, v := range ctx.Violations {
		counts[v.Type]++
	}

	// Build summary
	var parts []string
	for vtype, count := range counts {
		parts = append(parts, fmt.Sprintf("%s: %d", vtype, count))
	}

	return fmt.Sprintf("%d security violations: %s",
		len(ctx.Violations), strings.Join(parts, ", "))
}
