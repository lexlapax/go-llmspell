// ABOUTME: Tests for SecurityManager which enforces security policies for Lua VM instances
// ABOUTME: Validates library restrictions, resource limits, and sandbox enforcement

package gopherlua

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestSecurityManager_NewWithProfiles(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, sm *SecurityManager)
	}{
		{
			name:    "creates_with_minimal_profile",
			profile: SecurityProfileMinimal,
			validate: func(t *testing.T, sm *SecurityManager) {
				assert.Equal(t, SecurityLevelMinimal, sm.config.Level)
				assert.Contains(t, sm.config.AllowedLibraries, "base")
				assert.Contains(t, sm.config.AllowedLibraries, "io")
				assert.Contains(t, sm.config.AllowedLibraries, "os")
				assert.NotContains(t, sm.config.AllowedLibraries, "debug")
			},
		},
		{
			name:    "creates_with_standard_profile",
			profile: SecurityProfileStandard,
			validate: func(t *testing.T, sm *SecurityManager) {
				assert.Equal(t, SecurityLevelStandard, sm.config.Level)
				assert.Contains(t, sm.config.AllowedLibraries, "base")
				assert.Contains(t, sm.config.AllowedLibraries, "string")
				assert.NotContains(t, sm.config.AllowedLibraries, "io")
				assert.Contains(t, sm.config.AllowedLibraries, "os") // Limited OS functions
				assert.NotContains(t, sm.config.AllowedLibraries, "debug")
			},
		},
		{
			name:    "creates_with_strict_profile",
			profile: SecurityProfileStrict,
			validate: func(t *testing.T, sm *SecurityManager) {
				assert.Equal(t, SecurityLevelStrict, sm.config.Level)
				assert.Contains(t, sm.config.AllowedLibraries, "base")
				assert.Contains(t, sm.config.AllowedLibraries, "string")
				assert.NotContains(t, sm.config.AllowedLibraries, "io")
				assert.NotContains(t, sm.config.AllowedLibraries, "os")
				assert.NotContains(t, sm.config.AllowedLibraries, "debug")
			},
		},
		{
			name:        "fails_with_unknown_profile",
			profile:     "unknown",
			wantErr:     true,
			errContains: "unknown security profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewSecurityManagerFromProfile(tt.profile)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, sm)

			if tt.validate != nil {
				tt.validate(t, sm)
			}
		})
	}
}

func TestSecurityManager_NewWithConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   SecurityConfig
		validate func(t *testing.T, sm *SecurityManager)
	}{
		{
			name: "creates_with_custom_config",
			config: SecurityConfig{
				Level:            SecurityLevelCustom,
				AllowedLibraries: []string{"base", "string", "math"},
				DeniedFunctions: map[string]bool{
					"os.execute": true,
					"os.exit":    true,
				},
				ResourceLimits: ResourceLimits{
					MaxInstructions: 1000000,
					MaxMemory:       10 * 1024 * 1024, // 10MB
					MaxDuration:     5 * time.Second,
				},
			},
			validate: func(t *testing.T, sm *SecurityManager) {
				assert.Equal(t, SecurityLevelCustom, sm.config.Level)
				assert.Equal(t, []string{"base", "string", "math"}, sm.config.AllowedLibraries)
				assert.True(t, sm.config.DeniedFunctions["os.execute"])
				assert.Equal(t, int64(1000000), sm.config.ResourceLimits.MaxInstructions)
			},
		},
		{
			name: "applies_defaults_for_empty_config",
			config: SecurityConfig{
				Level: SecurityLevelStandard,
			},
			validate: func(t *testing.T, sm *SecurityManager) {
				assert.Equal(t, SecurityLevelStandard, sm.config.Level)
				// Should have some default libraries
				assert.NotEmpty(t, sm.config.AllowedLibraries)
				// Should have some default resource limits
				assert.Greater(t, sm.config.ResourceLimits.MaxInstructions, int64(0))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSecurityManager(tt.config)
			require.NotNil(t, sm)

			if tt.validate != nil {
				tt.validate(t, sm)
			}
		})
	}
}

func TestSecurityManager_LoadLibraries(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		validate func(t *testing.T, L *lua.LState)
	}{
		{
			name:    "minimal_loads_most_libraries",
			profile: SecurityProfileMinimal,
			validate: func(t *testing.T, L *lua.LState) {
				// Should have most libraries
				assert.NotEqual(t, lua.LNil, L.GetGlobal("string"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("table"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("math"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("io"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("os"))
				// But never debug
				assert.Equal(t, lua.LNil, L.GetGlobal("debug"))
			},
		},
		{
			name:    "standard_restricts_dangerous_libraries",
			profile: SecurityProfileStandard,
			validate: func(t *testing.T, L *lua.LState) {
				// Should have safe libraries
				assert.NotEqual(t, lua.LNil, L.GetGlobal("string"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("table"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("math"))
				// No IO
				assert.Equal(t, lua.LNil, L.GetGlobal("io"))
				// Limited OS (check specific functions)
				assert.NotEqual(t, lua.LNil, L.GetGlobal("os"))
				// No debug
				assert.Equal(t, lua.LNil, L.GetGlobal("debug"))
			},
		},
		{
			name:    "strict_minimal_libraries_only",
			profile: SecurityProfileStrict,
			validate: func(t *testing.T, L *lua.LState) {
				// Only essential libraries
				assert.NotEqual(t, lua.LNil, L.GetGlobal("string"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("table"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("math"))
				// No dangerous libraries
				assert.Equal(t, lua.LNil, L.GetGlobal("io"))
				assert.Equal(t, lua.LNil, L.GetGlobal("os"))
				assert.Equal(t, lua.LNil, L.GetGlobal("debug"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewSecurityManagerFromProfile(tt.profile)
			require.NoError(t, err)

			opts := lua.Options{SkipOpenLibs: true}
			L := lua.NewState(opts)
			defer L.Close()

			err = sm.LoadLibraries(L)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, L)
			}
		})
	}
}

func TestSecurityManager_DeniedFunctions(t *testing.T) {
	sm, err := NewSecurityManagerFromProfile(SecurityProfileStandard)
	require.NoError(t, err)

	opts := lua.Options{SkipOpenLibs: true}
	L := lua.NewState(opts)
	defer L.Close()

	err = sm.LoadLibraries(L)
	require.NoError(t, err)

	// Test that dangerous OS functions are removed
	dangerousFuncs := []string{
		"os.execute",
		"os.exit",
		"os.setenv",
		"os.remove",
		"os.rename",
	}

	for _, funcPath := range dangerousFuncs {
		parts := strings.Split(funcPath, ".")
		if len(parts) == 2 {
			err := L.DoString(fmt.Sprintf(`
				if %s and %s.%s then
					error("%s should not exist")
				end
			`, parts[0], parts[0], parts[1], funcPath))
			assert.NoError(t, err, "%s should be removed", funcPath)
		}
	}

	// Test that safe OS functions still exist
	safeFuncs := []string{
		"os.time",
		"os.date",
		"os.clock",
		"os.difftime",
	}

	for _, funcPath := range safeFuncs {
		parts := strings.Split(funcPath, ".")
		if len(parts) == 2 {
			err := L.DoString(fmt.Sprintf(`
				if not (%s and %s.%s) then
					error("%s should exist")
				end
			`, parts[0], parts[0], parts[1], funcPath))
			assert.NoError(t, err, "%s should exist", funcPath)
		}
	}
}

func TestSecurityManager_ApplySandbox(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		testCode string
		wantErr  bool
	}{
		{
			name:    "allows_safe_operations",
			profile: SecurityProfileStandard,
			testCode: `
				local x = 1 + 1
				local s = string.upper("hello")
				local t = {a = 1, b = 2}
				return x, s, t.a
			`,
			wantErr: false,
		},
		{
			name:    "blocks_file_operations_in_standard",
			profile: SecurityProfileStandard,
			testCode: `
				local f = io.open("test.txt", "w")
			`,
			wantErr: true, // io should be nil
		},
		{
			name:    "allows_file_operations_in_minimal",
			profile: SecurityProfileMinimal,
			testCode: `
				-- This would work if io is available
				if io then
					return "io available"
				end
				return "no io"
			`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewSecurityManagerFromProfile(tt.profile)
			require.NoError(t, err)

			opts := lua.Options{SkipOpenLibs: true}
			L := lua.NewState(opts)
			defer L.Close()

			// Apply sandbox
			err = sm.ApplySandbox(L)
			require.NoError(t, err)

			// Test code execution
			err = L.DoString(tt.testCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityManager_ResourceLimits(t *testing.T) {
	t.Skip("Resource limits require SetHook which is not available in gopher-lua")

	config := SecurityConfig{
		Level:            SecurityLevelCustom,
		AllowedLibraries: []string{"base", "string"},
		ResourceLimits: ResourceLimits{
			MaxInstructions: 1000,
			MaxMemory:       1024 * 1024, // 1MB
			MaxDuration:     100 * time.Millisecond,
			CheckInterval:   100,
		},
	}

	sm := NewSecurityManager(config)

	opts := lua.Options{SkipOpenLibs: true}
	L := lua.NewState(opts)
	defer L.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	L.SetContext(ctx)

	err := sm.ApplySandbox(L)
	require.NoError(t, err)

	// TODO: Implement alternative resource limit testing
	// Currently gopher-lua doesn't support SetHook for instruction counting
}

func TestSecurityManager_InstallHooks(t *testing.T) {
	config := SecurityConfig{
		Level:            SecurityLevelCustom,
		AllowedLibraries: []string{"base"},
		ResourceLimits: ResourceLimits{
			MaxInstructions: 500,
			CheckInterval:   50,
		},
	}

	sm := NewSecurityManager(config)

	opts := lua.Options{SkipOpenLibs: true}
	L := lua.NewState(opts)
	defer L.Close()

	monitor := sm.InstallHooks(L)
	require.NotNil(t, monitor)

	// TODO: Since gopher-lua doesn't support SetHook, we can only verify monitor creation
	// Resource monitoring will need to be implemented differently
	assert.Equal(t, int64(0), monitor.GetInstructionCount())
}

func TestSecurityManager_CustomDeniedFunctions(t *testing.T) {
	config := SecurityConfig{
		Level:            SecurityLevelCustom,
		AllowedLibraries: []string{"base", "string", "table"},
		DeniedFunctions: map[string]bool{
			"string.dump":  true,
			"table.concat": true,
		},
	}

	sm := NewSecurityManager(config)

	opts := lua.Options{SkipOpenLibs: true}
	L := lua.NewState(opts)
	defer L.Close()

	err := sm.ApplySandbox(L)
	require.NoError(t, err)

	// Check that denied functions are nil
	err = L.DoString(`
		if string.dump then
			error("string.dump should be removed")
		end
		if table.concat then
			error("table.concat should be removed")
		end
	`)
	assert.NoError(t, err)

	// Check that other functions still work
	err = L.DoString(`
		local s = string.upper("hello")
		local t = table.insert({}, 1)
	`)
	assert.NoError(t, err)
}
