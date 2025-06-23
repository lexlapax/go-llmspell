// ABOUTME: Unit tests for security levels implementation.
// ABOUTME: Tests security level validation, configuration retrieval, and permissions.

package security

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityLevelConstants(t *testing.T) {
	// Test that security level constants are properly defined
	assert.Equal(t, SecurityLevel("untrusted"), SecurityLevelUntrusted)
	assert.Equal(t, SecurityLevel("trusted"), SecurityLevelTrusted)
	assert.Equal(t, SecurityLevel("privileged"), SecurityLevelPrivileged)
}

func TestIsValidLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{"valid untrusted", "untrusted", true},
		{"valid trusted", "trusted", true},
		{"valid privileged", "privileged", true},
		{"invalid empty", "", false},
		{"invalid unknown", "unknown", false},
		{"invalid sandbox", "sandbox", false},         // Old profile name
		{"invalid development", "development", false}, // Old profile name
		{"invalid production", "production", false},   // Old profile name
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLevelConfig(t *testing.T) {
	tests := []struct {
		name  string
		level SecurityLevel
		check func(t *testing.T, config *SecurityConfig)
	}{
		{
			name:  "untrusted level",
			level: SecurityLevelUntrusted,
			check: func(t *testing.T, config *SecurityConfig) {
				assert.Equal(t, SecurityLevelUntrusted, config.Level)
				assert.False(t, config.AllowNetwork)
				assert.False(t, config.AllowFilesystem)
				assert.False(t, config.AllowEnvironment)
				assert.False(t, config.AllowExec)
				assert.False(t, config.AllowUnsafe)
				assert.Equal(t, int64(64*1024*1024), config.MemoryLimit)
				assert.Equal(t, 5, config.CPULimit)
				assert.Equal(t, 30, config.TimeoutSeconds)
			},
		},
		{
			name:  "trusted level",
			level: SecurityLevelTrusted,
			check: func(t *testing.T, config *SecurityConfig) {
				assert.Equal(t, SecurityLevelTrusted, config.Level)
				assert.True(t, config.AllowNetwork)
				assert.True(t, config.AllowFilesystem)
				assert.True(t, config.AllowEnvironment)
				assert.False(t, config.AllowExec)
				assert.False(t, config.AllowUnsafe)
				assert.Equal(t, int64(256*1024*1024), config.MemoryLimit)
				assert.Equal(t, 50, config.CPULimit)
				assert.Equal(t, 300, config.TimeoutSeconds)
			},
		},
		{
			name:  "privileged level",
			level: SecurityLevelPrivileged,
			check: func(t *testing.T, config *SecurityConfig) {
				assert.Equal(t, SecurityLevelPrivileged, config.Level)
				assert.True(t, config.AllowNetwork)
				assert.True(t, config.AllowFilesystem)
				assert.True(t, config.AllowEnvironment)
				assert.True(t, config.AllowExec)
				assert.True(t, config.AllowUnsafe)
				assert.Equal(t, int64(1024*1024*1024), config.MemoryLimit)
				assert.Equal(t, 100, config.CPULimit)
				assert.Equal(t, 0, config.TimeoutSeconds)
			},
		},
		{
			name:  "invalid level defaults to untrusted",
			level: SecurityLevel("invalid"),
			check: func(t *testing.T, config *SecurityConfig) {
				assert.Equal(t, SecurityLevelUntrusted, config.Level)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetLevelConfig(tt.level)
			require.NotNil(t, config)
			tt.check(t, config)
		})
	}
}

func TestSecurityConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *SecurityConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			config:  UntrustedLevel(),
			wantErr: false,
		},
		{
			name: "invalid level",
			config: &SecurityConfig{
				Level:          SecurityLevel("invalid"),
				MemoryLimit:    1024,
				CPULimit:       50,
				TimeoutSeconds: 30,
			},
			wantErr: true,
			errMsg:  "invalid security level",
		},
		{
			name: "zero memory limit",
			config: &SecurityConfig{
				Level:          SecurityLevelTrusted,
				MemoryLimit:    0,
				CPULimit:       50,
				TimeoutSeconds: 30,
			},
			wantErr: true,
			errMsg:  "memory limit must be positive",
		},
		{
			name: "zero CPU limit",
			config: &SecurityConfig{
				Level:          SecurityLevelTrusted,
				MemoryLimit:    1024,
				CPULimit:       0,
				TimeoutSeconds: 30,
			},
			wantErr: true,
			errMsg:  "CPU limit must be positive",
		},
		{
			name: "CPU limit over 100",
			config: &SecurityConfig{
				Level:          SecurityLevelTrusted,
				MemoryLimit:    1024,
				CPULimit:       101,
				TimeoutSeconds: 30,
			},
			wantErr: true,
			errMsg:  "CPU limit must not exceed 100",
		},
		{
			name: "negative timeout",
			config: &SecurityConfig{
				Level:          SecurityLevelTrusted,
				MemoryLimit:    1024,
				CPULimit:       50,
				TimeoutSeconds: -1,
			},
			wantErr: true,
			errMsg:  "timeout must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSecurityConfigCheckPermission(t *testing.T) {
	config := &SecurityConfig{
		AllowNetwork:     true,
		AllowFilesystem:  true,
		AllowEnvironment: false,
		AllowExec:        false,
		AllowUnsafe:      false,
	}

	tests := []struct {
		name       string
		permission Permission
		expected   bool
	}{
		{"network allowed", PermissionNetwork, true},
		{"filesystem allowed", PermissionFilesystem, true},
		{"environment denied", PermissionEnvironment, false},
		{"exec denied", PermissionExec, false},
		{"unsafe denied", PermissionUnsafe, false},
		{"unknown permission", Permission("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.CheckPermission(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityConfigIsModuleAllowed(t *testing.T) {
	config := &SecurityConfig{
		AllowedModules: []string{"string", "table", "math"},
	}

	tests := []struct {
		name     string
		module   string
		expected bool
	}{
		{"allowed string", "string", true},
		{"allowed table", "table", true},
		{"allowed math", "math", true},
		{"not allowed io", "io", false},
		{"not allowed os", "os", false},
		{"empty module", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.IsModuleAllowed(tt.module)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityConfigIsFunctionForbidden(t *testing.T) {
	config := &SecurityConfig{
		ForbiddenFunctions: []string{"loadstring", "load", "dofile"},
	}

	tests := []struct {
		name     string
		function string
		expected bool
	}{
		{"forbidden loadstring", "loadstring", true},
		{"forbidden load", "load", true},
		{"forbidden dofile", "dofile", true},
		{"allowed print", "print", false},
		{"allowed require", "require", false},
		{"empty function", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.IsFunctionForbidden(tt.function)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityConfigIsPathAllowed(t *testing.T) {
	config := &SecurityConfig{
		AllowedPaths:   []string{"/tmp", "/var/tmp", "./"},
		ForbiddenPaths: []string{"/etc", "/sys", "/proc"},
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"allowed /tmp", "/tmp/file.txt", true},
		{"allowed /var/tmp", "/var/tmp/data", true},
		{"allowed current dir", "./script.lua", true},
		{"forbidden /etc", "/etc/passwd", false},
		{"forbidden /sys", "/sys/kernel", false},
		{"forbidden /proc", "/proc/1/status", false},
		{"not in allowed", "/home/user", false},
		{"empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.IsPathAllowed(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityConfigIsPathAllowedNoAllowedPaths(t *testing.T) {
	config := &SecurityConfig{
		AllowedPaths:   []string{}, // No allowed paths
		ForbiddenPaths: []string{"/etc"},
	}

	// When no allowed paths are specified, all paths should be denied
	assert.False(t, config.IsPathAllowed("/tmp/file.txt"))
	assert.False(t, config.IsPathAllowed("/home/user"))
}

func TestLevelManager(t *testing.T) {
	manager := NewLevelManager()
	require.NotNil(t, manager)

	t.Run("get existing levels", func(t *testing.T) {
		config, err := manager.GetLevel(SecurityLevelUntrusted)
		require.NoError(t, err)
		assert.Equal(t, SecurityLevelUntrusted, config.Level)

		config, err = manager.GetLevel(SecurityLevelTrusted)
		require.NoError(t, err)
		assert.Equal(t, SecurityLevelTrusted, config.Level)

		config, err = manager.GetLevel(SecurityLevelPrivileged)
		require.NoError(t, err)
		assert.Equal(t, SecurityLevelPrivileged, config.Level)
	})

	t.Run("get non-existent level", func(t *testing.T) {
		_, err := manager.GetLevel(SecurityLevel("nonexistent"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "security level not found")
	})
}

func TestLevelManagerConcurrency(t *testing.T) {
	manager := NewLevelManager()

	// Test concurrent reads
	done := make(chan bool, 3)
	levels := []SecurityLevel{
		SecurityLevelUntrusted,
		SecurityLevelTrusted,
		SecurityLevelPrivileged,
	}

	for _, level := range levels {
		go func(l SecurityLevel) {
			for i := 0; i < 100; i++ {
				config, err := manager.GetLevel(l)
				assert.NoError(t, err)
				assert.Equal(t, l, config.Level)
			}
			done <- true
		}(level)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestSecurityContext(t *testing.T) {
	t.Run("create context", func(t *testing.T) {
		config := TrustedLevel()
		ctx := NewSecurityContext(config)

		assert.NotNil(t, ctx)
		assert.Equal(t, config, ctx.Config)
		assert.Empty(t, ctx.Violations)
	})

	t.Run("record violation", func(t *testing.T) {
		config := UntrustedLevel()
		ctx := NewSecurityContext(config)

		ctx.RecordViolation("network", "attempted network access")

		assert.True(t, ctx.HasViolations())
		violations := ctx.GetViolations()
		assert.Len(t, violations, 1)
		assert.Equal(t, "network", violations[0].Type)
		assert.Equal(t, "attempted network access", violations[0].Description)
		assert.False(t, violations[0].Timestamp.IsZero())
	})

	t.Run("check and record allowed", func(t *testing.T) {
		config := TrustedLevel()
		ctx := NewSecurityContext(config)

		// Trusted level allows network
		allowed := ctx.CheckAndRecord(PermissionNetwork, "accessing API")

		assert.True(t, allowed)
		assert.False(t, ctx.HasViolations())
	})

	t.Run("check and record denied", func(t *testing.T) {
		config := UntrustedLevel()
		ctx := NewSecurityContext(config)

		// Untrusted level denies network
		allowed := ctx.CheckAndRecord(PermissionNetwork, "accessing API")

		assert.False(t, allowed)
		assert.True(t, ctx.HasViolations())
		violations := ctx.GetViolations()
		assert.Len(t, violations, 1)
		assert.Equal(t, "network", violations[0].Type)
	})

	t.Run("violation summary", func(t *testing.T) {
		config := UntrustedLevel()
		ctx := NewSecurityContext(config)

		// No violations
		summary := ctx.GetViolationSummary()
		assert.Equal(t, "No security violations", summary)

		// Add violations
		ctx.RecordViolation("network", "attempted network access")
		ctx.RecordViolation("filesystem", "attempted file write")
		ctx.RecordViolation("network", "attempted network access again")

		summary = ctx.GetViolationSummary()
		assert.Contains(t, summary, "3 security violations")
		assert.Contains(t, summary, "network: 2")
		assert.Contains(t, summary, "filesystem: 1")
	})

	t.Run("concurrent violations", func(t *testing.T) {
		config := UntrustedLevel()
		ctx := NewSecurityContext(config)

		// Record violations concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				ctx.RecordViolation("test", fmt.Sprintf("violation %d", idx))
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		violations := ctx.GetViolations()
		assert.Len(t, violations, 10)
	})

	t.Run("get violations returns copy", func(t *testing.T) {
		config := UntrustedLevel()
		ctx := NewSecurityContext(config)

		ctx.RecordViolation("test", "violation 1")

		// Get violations and modify the returned slice
		violations1 := ctx.GetViolations()
		assert.Len(t, violations1, 1)

		// Modify the returned slice
		violations1[0].Type = "modified"

		// Get violations again - should be unchanged
		violations2 := ctx.GetViolations()
		assert.Equal(t, "test", violations2[0].Type)
		assert.NotEqual(t, "modified", violations2[0].Type)
	})
}
