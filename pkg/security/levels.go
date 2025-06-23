// ABOUTME: Security levels definition for controlling script execution permissions and resource limits.
// ABOUTME: Single source of truth for all security level enums and configurations.

// Package security provides security levels and permission management.
// It implements security constraints for safe script execution across
// different trust levels: untrusted, trusted, and privileged.
package security

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// SecurityLevel represents the trust level for script execution.
// This is the single source of truth for security level definitions.
type SecurityLevel string

const (
	// SecurityLevelUntrusted is for maximum restrictions (untrusted scripts)
	SecurityLevelUntrusted SecurityLevel = "untrusted"
	// SecurityLevelTrusted is for moderate restrictions (known scripts)
	SecurityLevelTrusted SecurityLevel = "trusted"
	// SecurityLevelPrivileged is for minimal restrictions (system scripts)
	SecurityLevelPrivileged SecurityLevel = "privileged"
)

// IsValidLevel validates if a string represents a valid security level.
// This prevents redefinition of security levels across the codebase.
func IsValidLevel(level string) bool {
	switch SecurityLevel(level) {
	case SecurityLevelUntrusted, SecurityLevelTrusted, SecurityLevelPrivileged:
		return true
	default:
		return false
	}
}

// Permission represents a specific permission type.
// Each permission controls access to a different system resource
// or capability.
type Permission string

const (
	PermissionNetwork     Permission = "network"
	PermissionFilesystem  Permission = "filesystem"
	PermissionEnvironment Permission = "environment"
	PermissionExec        Permission = "exec"
	PermissionUnsafe      Permission = "unsafe"
)

// SecurityConfig defines security constraints for script execution.
// It specifies permissions, resource limits, module restrictions,
// and path access controls for safe script execution.
type SecurityConfig struct {
	Level       SecurityLevel `json:"level" yaml:"level"`
	Description string        `json:"description" yaml:"description"`

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

// GetLevelConfig returns the security configuration for a given level.
// This is the single source for security level configurations.
func GetLevelConfig(level SecurityLevel) *SecurityConfig {
	switch level {
	case SecurityLevelUntrusted:
		return UntrustedLevel()
	case SecurityLevelTrusted:
		return TrustedLevel()
	case SecurityLevelPrivileged:
		return PrivilegedLevel()
	default:
		// Default to most restrictive
		return UntrustedLevel()
	}
}

// UntrustedLevel returns a highly restrictive security configuration.
// It denies all system access and enforces tight resource limits,
// suitable for running untrusted scripts.
func UntrustedLevel() *SecurityConfig {
	return &SecurityConfig{
		Level:       SecurityLevelUntrusted,
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

// TrustedLevel returns a moderately permissive configuration.
// It allows filesystem and network access with reasonable resource limits,
// suitable for development and known scripts.
func TrustedLevel() *SecurityConfig {
	return &SecurityConfig{
		Level:       SecurityLevelTrusted,
		Description: "Balanced security for development and trusted scripts",

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

// PrivilegedLevel returns a minimally restrictive configuration.
// It allows all permissions with high resource limits,
// suitable for system administration scripts.
func PrivilegedLevel() *SecurityConfig {
	return &SecurityConfig{
		Level:       SecurityLevelPrivileged,
		Description: "Minimal restrictions for system administration",

		// All permissions granted
		AllowNetwork:     true,
		AllowFilesystem:  true,
		AllowEnvironment: true,
		AllowExec:        true,
		AllowUnsafe:      true,

		// High resource limits
		MemoryLimit:    1024 * 1024 * 1024, // 1GB
		CPULimit:       100,                // 100% CPU
		TimeoutSeconds: 0,                  // No timeout

		// All modules allowed
		AllowedModules: []string{
			"string", "table", "math", "utf8",
			"coroutine", "bit32", "bit",
			"os", "io", "debug", "package",
			"ffi", // Allow FFI for privileged scripts
		},
		ForbiddenFunctions: []string{}, // No restrictions

		// Full filesystem access
		AllowedPaths:   []string{"/"}, // All paths
		ForbiddenPaths: []string{},    // No restrictions
	}
}

// Validate checks if the security configuration is valid.
// It ensures all resource limits are within acceptable ranges
// and required fields are present.
func (c *SecurityConfig) Validate() error {
	if !IsValidLevel(string(c.Level)) {
		return fmt.Errorf("invalid security level: %s", c.Level)
	}

	if c.MemoryLimit <= 0 {
		return fmt.Errorf("memory limit must be positive")
	}

	if c.CPULimit <= 0 {
		return fmt.Errorf("CPU limit must be positive")
	}
	if c.CPULimit > 100 {
		return fmt.Errorf("CPU limit must not exceed 100")
	}

	if c.TimeoutSeconds < 0 {
		return fmt.Errorf("timeout must be non-negative")
	}

	return nil
}

// CheckPermission checks if a permission is allowed.
// It returns true if the config grants the specified permission,
// false otherwise.
func (c *SecurityConfig) CheckPermission(perm Permission) bool {
	switch perm {
	case PermissionNetwork:
		return c.AllowNetwork
	case PermissionFilesystem:
		return c.AllowFilesystem
	case PermissionEnvironment:
		return c.AllowEnvironment
	case PermissionExec:
		return c.AllowExec
	case PermissionUnsafe:
		return c.AllowUnsafe
	default:
		return false
	}
}

// IsModuleAllowed checks if a module is allowed.
// It returns true if the module is in the allowed modules list,
// false otherwise.
func (c *SecurityConfig) IsModuleAllowed(module string) bool {
	for _, allowed := range c.AllowedModules {
		if allowed == module {
			return true
		}
	}
	return false
}

// IsFunctionForbidden checks if a function is forbidden.
// It returns true if the function is in the forbidden functions list,
// preventing its use in scripts.
func (c *SecurityConfig) IsFunctionForbidden(function string) bool {
	for _, forbidden := range c.ForbiddenFunctions {
		if forbidden == function {
			return true
		}
	}
	return false
}

// IsPathAllowed checks if a path is allowed.
// It first checks forbidden paths, then allowed paths,
// implementing a deny-first security model.
func (c *SecurityConfig) IsPathAllowed(path string) bool {
	// First check forbidden paths
	for _, forbidden := range c.ForbiddenPaths {
		if strings.HasPrefix(path, forbidden) {
			return false
		}
	}

	// If no allowed paths specified, default to forbidden
	if len(c.AllowedPaths) == 0 {
		return false
	}

	// Check allowed paths
	for _, allowed := range c.AllowedPaths {
		if strings.HasPrefix(path, allowed) {
			return true
		}
	}

	return false
}

// LevelManager manages security levels.
// It provides thread-safe storage and retrieval of security configurations.
type LevelManager struct {
	mu     sync.RWMutex
	levels map[SecurityLevel]*SecurityConfig
}

// NewLevelManager creates a new level manager with default levels.
// It automatically registers untrusted, trusted, and privileged levels
// for immediate use.
func NewLevelManager() *LevelManager {
	lm := &LevelManager{
		levels: make(map[SecurityLevel]*SecurityConfig),
	}

	// Register default levels
	lm.levels[SecurityLevelUntrusted] = UntrustedLevel()
	lm.levels[SecurityLevelTrusted] = TrustedLevel()
	lm.levels[SecurityLevelPrivileged] = PrivilegedLevel()

	return lm
}

// GetLevel retrieves a security configuration by level.
// It returns an error if the level doesn't exist.
func (lm *LevelManager) GetLevel(level SecurityLevel) (*SecurityConfig, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	config, exists := lm.levels[level]
	if !exists {
		return nil, fmt.Errorf("security level not found: %s", level)
	}

	return config, nil
}

// SecurityViolation represents a security policy violation.
// It records the type of violation, description, and when it occurred
// for security auditing and monitoring.
type SecurityViolation struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// SecurityContext tracks security state during execution.
// It maintains the active security configuration and records any
// security violations that occur during script execution.
type SecurityContext struct {
	Config     *SecurityConfig
	Violations []SecurityViolation
	mu         sync.Mutex
}

// NewSecurityContext creates a new security context.
// It initializes the context with the specified security configuration
// and an empty violations list.
func NewSecurityContext(config *SecurityConfig) *SecurityContext {
	return &SecurityContext{
		Config:     config,
		Violations: []SecurityViolation{},
	}
}

// RecordViolation adds a security violation to the context.
// It is thread-safe and records the current timestamp.
func (sc *SecurityContext) RecordViolation(violationType, description string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	violation := SecurityViolation{
		Type:        violationType,
		Description: description,
		Timestamp:   time.Now(),
	}
	sc.Violations = append(sc.Violations, violation)
}

// GetViolations returns a copy of all recorded violations.
// It is thread-safe and returns a new slice to prevent external modification.
func (sc *SecurityContext) GetViolations() []SecurityViolation {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	violations := make([]SecurityViolation, len(sc.Violations))
	copy(violations, sc.Violations)
	return violations
}

// CheckAndRecord checks a permission and records a violation if denied.
// It returns true if the permission is allowed, false otherwise.
func (sc *SecurityContext) CheckAndRecord(perm Permission, description string) bool {
	allowed := sc.Config.CheckPermission(perm)
	if !allowed {
		sc.RecordViolation(string(perm), description)
	}
	return allowed
}

// HasViolations returns true if any violations have been recorded.
// This can be used to determine if script execution should be terminated
// due to security policy violations.
func (sc *SecurityContext) HasViolations() bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return len(sc.Violations) > 0
}

// GetViolationSummary returns a summary of recorded violations.
// It groups violations by type and provides a count of each,
// useful for security reporting and debugging.
func (sc *SecurityContext) GetViolationSummary() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if len(sc.Violations) == 0 {
		return "No security violations"
	}

	// Count violations by type
	counts := make(map[string]int)
	for _, v := range sc.Violations {
		counts[v.Type]++
	}

	// Build summary
	var parts []string
	for vtype, count := range counts {
		parts = append(parts, fmt.Sprintf("%s: %d", vtype, count))
	}

	return fmt.Sprintf("%d security violations: %s",
		len(sc.Violations), strings.Join(parts, ", "))
}
