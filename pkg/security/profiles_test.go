// ABOUTME: Tests for security profiles covering sandbox, development, and production modes.
// ABOUTME: Ensures proper permission enforcement and resource restrictions.

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityProfile(t *testing.T) {
	t.Run("sandbox_profile", func(t *testing.T) {
		profile := SandboxProfile()

		assert.Equal(t, "sandbox", profile.Name)
		assert.Equal(t, "Maximum security restrictions for untrusted scripts", profile.Description)

		// Verify restrictions
		assert.False(t, profile.AllowNetwork)
		assert.False(t, profile.AllowFilesystem)
		assert.False(t, profile.AllowEnvironment)
		assert.False(t, profile.AllowExec)
		assert.False(t, profile.AllowUnsafe)

		// Verify resource limits
		assert.Equal(t, int64(64*1024*1024), profile.MemoryLimit)
		assert.Equal(t, 5, profile.CPULimit)
		assert.Equal(t, 30, profile.TimeoutSeconds)

		// Verify allowed modules
		assert.Contains(t, profile.AllowedModules, "string")
		assert.Contains(t, profile.AllowedModules, "table")
		assert.Contains(t, profile.AllowedModules, "math")
		assert.NotContains(t, profile.AllowedModules, "os")
		assert.NotContains(t, profile.AllowedModules, "io")
	})

	t.Run("development_profile", func(t *testing.T) {
		profile := DevelopmentProfile()

		assert.Equal(t, "development", profile.Name)

		// More permissive than sandbox
		assert.True(t, profile.AllowNetwork)
		assert.True(t, profile.AllowFilesystem)
		assert.True(t, profile.AllowEnvironment)
		assert.False(t, profile.AllowExec) // Still restricted
		assert.False(t, profile.AllowUnsafe)

		// Higher resource limits
		assert.Equal(t, int64(256*1024*1024), profile.MemoryLimit)
		assert.Equal(t, 50, profile.CPULimit)
		assert.Equal(t, 300, profile.TimeoutSeconds)

		// More modules allowed
		assert.Contains(t, profile.AllowedModules, "os")
		assert.Contains(t, profile.AllowedModules, "io")
		assert.Contains(t, profile.AllowedModules, "debug")
	})

	t.Run("production_profile", func(t *testing.T) {
		profile := ProductionProfile()

		assert.Equal(t, "production", profile.Name)

		// Balanced permissions
		assert.True(t, profile.AllowNetwork)
		assert.False(t, profile.AllowFilesystem) // Restricted in production
		assert.False(t, profile.AllowEnvironment)
		assert.False(t, profile.AllowExec)
		assert.False(t, profile.AllowUnsafe)

		// Production resource limits
		assert.Equal(t, int64(512*1024*1024), profile.MemoryLimit)
		assert.Equal(t, 25, profile.CPULimit)
		assert.Equal(t, 60, profile.TimeoutSeconds)

		// Limited modules
		assert.NotContains(t, profile.AllowedModules, "debug")
		assert.NotContains(t, profile.AllowedModules, "io")
	})

	t.Run("custom_profile", func(t *testing.T) {
		profile := &SecurityProfile{
			Name:               "custom",
			Description:        "Custom security profile",
			AllowNetwork:       true,
			AllowFilesystem:    false,
			AllowEnvironment:   true,
			AllowExec:          false,
			AllowUnsafe:        false,
			MemoryLimit:        128 * 1024 * 1024,
			CPULimit:           10,
			TimeoutSeconds:     120,
			AllowedModules:     []string{"string", "table"},
			ForbiddenFunctions: []string{"loadstring", "dofile"},
		}

		assert.Equal(t, "custom", profile.Name)
		assert.True(t, profile.AllowNetwork)
		assert.False(t, profile.AllowFilesystem)
		assert.Equal(t, 120, profile.TimeoutSeconds)
	})
}

func TestProfileValidation(t *testing.T) {
	t.Run("valid_profile", func(t *testing.T) {
		profile := SandboxProfile()
		err := profile.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid_name", func(t *testing.T) {
		profile := &SecurityProfile{
			Name: "",
		}
		err := profile.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profile name is required")
	})

	t.Run("invalid_memory_limit", func(t *testing.T) {
		profile := &SecurityProfile{
			Name:        "test",
			MemoryLimit: -1,
		}
		err := profile.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "memory limit must be positive")
	})

	t.Run("invalid_cpu_limit", func(t *testing.T) {
		profile := &SecurityProfile{
			Name:        "test",
			MemoryLimit: 1024,
			CPULimit:    -1,
		}
		err := profile.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU limit must be positive")
	})

	t.Run("invalid_timeout", func(t *testing.T) {
		profile := &SecurityProfile{
			Name:           "test",
			MemoryLimit:    1024,
			CPULimit:       10,
			TimeoutSeconds: 0,
		}
		err := profile.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout must be positive")
	})
}

func TestProfileManager(t *testing.T) {
	t.Run("new_manager", func(t *testing.T) {
		manager := NewProfileManager()
		assert.NotNil(t, manager)

		// Should have default profiles
		profiles := manager.ListProfiles()
		assert.Contains(t, profiles, "sandbox")
		assert.Contains(t, profiles, "development")
		assert.Contains(t, profiles, "production")
	})

	t.Run("get_profile", func(t *testing.T) {
		manager := NewProfileManager()

		// Get existing profile
		profile, err := manager.GetProfile("sandbox")
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "sandbox", profile.Name)

		// Get non-existent profile
		_, err = manager.GetProfile("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profile not found")
	})

	t.Run("register_profile", func(t *testing.T) {
		manager := NewProfileManager()

		custom := &SecurityProfile{
			Name:           "custom",
			MemoryLimit:    1024,
			CPULimit:       5,
			TimeoutSeconds: 10,
		}

		err := manager.RegisterProfile(custom)
		assert.NoError(t, err)

		// Should be able to get it
		profile, err := manager.GetProfile("custom")
		require.NoError(t, err)
		assert.Equal(t, "custom", profile.Name)

		// Should be in list
		profiles := manager.ListProfiles()
		assert.Contains(t, profiles, "custom")
	})

	t.Run("register_duplicate", func(t *testing.T) {
		manager := NewProfileManager()

		custom := &SecurityProfile{
			Name:           "sandbox", // Already exists
			MemoryLimit:    1024,
			CPULimit:       5,
			TimeoutSeconds: 10,
		}

		err := manager.RegisterProfile(custom)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("register_invalid", func(t *testing.T) {
		manager := NewProfileManager()

		invalid := &SecurityProfile{
			Name: "", // Invalid
		}

		err := manager.RegisterProfile(invalid)
		assert.Error(t, err)
	})

	t.Run("remove_profile", func(t *testing.T) {
		manager := NewProfileManager()

		// Register custom profile
		custom := &SecurityProfile{
			Name:           "removable",
			MemoryLimit:    1024,
			CPULimit:       5,
			TimeoutSeconds: 10,
		}

		err := manager.RegisterProfile(custom)
		require.NoError(t, err)

		// Remove it
		err = manager.RemoveProfile("removable")
		assert.NoError(t, err)

		// Should not exist
		_, err = manager.GetProfile("removable")
		assert.Error(t, err)
	})

	t.Run("cannot_remove_default", func(t *testing.T) {
		manager := NewProfileManager()

		err := manager.RemoveProfile("sandbox")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove default profile")
	})
}

func TestPermissionChecking(t *testing.T) {
	t.Run("check_permission", func(t *testing.T) {
		profile := &SecurityProfile{
			AllowNetwork:     true,
			AllowFilesystem:  false,
			AllowEnvironment: true,
			AllowExec:        false,
			AllowUnsafe:      false,
		}

		assert.True(t, profile.CheckPermission(PermissionNetwork))
		assert.False(t, profile.CheckPermission(PermissionFilesystem))
		assert.True(t, profile.CheckPermission(PermissionEnvironment))
		assert.False(t, profile.CheckPermission(PermissionExec))
		assert.False(t, profile.CheckPermission(PermissionUnsafe))
	})

	t.Run("check_module_allowed", func(t *testing.T) {
		profile := &SecurityProfile{
			AllowedModules: []string{"string", "table", "math"},
		}

		assert.True(t, profile.IsModuleAllowed("string"))
		assert.True(t, profile.IsModuleAllowed("table"))
		assert.True(t, profile.IsModuleAllowed("math"))
		assert.False(t, profile.IsModuleAllowed("os"))
		assert.False(t, profile.IsModuleAllowed("io"))
	})

	t.Run("check_function_forbidden", func(t *testing.T) {
		profile := &SecurityProfile{
			ForbiddenFunctions: []string{"loadstring", "dofile", "load"},
		}

		assert.True(t, profile.IsFunctionForbidden("loadstring"))
		assert.True(t, profile.IsFunctionForbidden("dofile"))
		assert.True(t, profile.IsFunctionForbidden("load"))
		assert.False(t, profile.IsFunctionForbidden("print"))
		assert.False(t, profile.IsFunctionForbidden("require"))
	})

	t.Run("check_path_allowed", func(t *testing.T) {
		profile := &SecurityProfile{
			AllowedPaths:   []string{"/tmp", "/home/user/scripts"},
			ForbiddenPaths: []string{"/etc", "/sys", "/proc"},
		}

		// Allowed paths
		assert.True(t, profile.IsPathAllowed("/tmp/test.txt"))
		assert.True(t, profile.IsPathAllowed("/home/user/scripts/main.lua"))

		// Forbidden paths
		assert.False(t, profile.IsPathAllowed("/etc/passwd"))
		assert.False(t, profile.IsPathAllowed("/sys/class"))

		// Path not in either list - default to forbidden
		assert.False(t, profile.IsPathAllowed("/usr/local/bin"))
	})
}

func TestSecurityContext(t *testing.T) {
	t.Run("create_context", func(t *testing.T) {
		profile := SandboxProfile()
		ctx := NewSecurityContext(profile)

		assert.NotNil(t, ctx)
		assert.Equal(t, profile, ctx.Profile)
		assert.NotNil(t, ctx.Violations)
		assert.Equal(t, 0, len(ctx.Violations))
	})

	t.Run("record_violation", func(t *testing.T) {
		profile := SandboxProfile()
		ctx := NewSecurityContext(profile)

		ctx.RecordViolation("network", "attempted to access network")
		ctx.RecordViolation("filesystem", "attempted to read /etc/passwd")

		assert.Equal(t, 2, len(ctx.Violations))
		assert.Equal(t, "network", ctx.Violations[0].Type)
		assert.Equal(t, "attempted to access network", ctx.Violations[0].Description)
		assert.NotZero(t, ctx.Violations[0].Timestamp)
	})

	t.Run("check_and_record", func(t *testing.T) {
		profile := &SecurityProfile{
			AllowNetwork:    false,
			AllowFilesystem: true,
		}
		ctx := NewSecurityContext(profile)

		// Should pass
		allowed := ctx.CheckAndRecord(PermissionFilesystem, "read file")
		assert.True(t, allowed)
		assert.Equal(t, 0, len(ctx.Violations))

		// Should fail and record
		allowed = ctx.CheckAndRecord(PermissionNetwork, "connect to server")
		assert.False(t, allowed)
		assert.Equal(t, 1, len(ctx.Violations))
		assert.Equal(t, "network", ctx.Violations[0].Type)
	})

	t.Run("has_violations", func(t *testing.T) {
		profile := SandboxProfile()
		ctx := NewSecurityContext(profile)

		assert.False(t, ctx.HasViolations())

		ctx.RecordViolation("test", "test violation")
		assert.True(t, ctx.HasViolations())
	})

	t.Run("get_violation_summary", func(t *testing.T) {
		profile := SandboxProfile()
		ctx := NewSecurityContext(profile)

		ctx.RecordViolation("network", "violation 1")
		ctx.RecordViolation("filesystem", "violation 2")
		ctx.RecordViolation("network", "violation 3")

		summary := ctx.GetViolationSummary()
		assert.Contains(t, summary, "3 security violations")
		assert.Contains(t, summary, "network")
		assert.Contains(t, summary, "filesystem")
	})
}

// Benchmark tests
func BenchmarkProfileValidation(b *testing.B) {
	profile := SandboxProfile()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.Validate()
	}
}

func BenchmarkPermissionCheck(b *testing.B) {
	profile := SandboxProfile()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.CheckPermission(PermissionNetwork)
	}
}

func BenchmarkModuleCheck(b *testing.B) {
	profile := SandboxProfile()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.IsModuleAllowed("string")
	}
}
