// ABOUTME: Integration tests that combine security profiles with validation.
// ABOUTME: Tests the end-to-end validation flow with security constraints.

package validator

import (
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityIntegration(t *testing.T) {
	t.Run("sandbox_profile_validation", func(t *testing.T) {
		// Create sandbox profile
		profile := security.SandboxProfile()

		// Create validation config based on profile
		config := &ValidationConfig{
			EnableSecurityCheck: true,
			SecurityProfile:     profile.Name,
			AllowedModules:      profile.AllowedModules,
			ForbiddenFunctions:  profile.ForbiddenFunctions,
		}

		validator := NewSecurityValidator(config)

		// Test allowed code
		allowedScript := `
local function add(a, b)
    return a + b
end

local result = add(1, 2)
print(result)
`
		result, err := validator.ValidateScript(allowedScript, "test.lua")
		require.NoError(t, err)
		assert.True(t, result.IsValid())

		// Test forbidden code
		forbiddenScript := `
local f = io.open("test.txt", "w")
f:write("test")
f:close()
`
		result, err = validator.ValidateScript(forbiddenScript, "test.lua")
		require.NoError(t, err)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Errors[0].Message, "forbidden module")
	})

	t.Run("development_profile_validation", func(t *testing.T) {
		// Create development profile
		profile := security.DevelopmentProfile()

		config := &ValidationConfig{
			EnableSecurityCheck: true,
			SecurityProfile:     profile.Name,
			AllowedModules:      profile.AllowedModules,
			ForbiddenFunctions:  profile.ForbiddenFunctions,
			ForbiddenPatterns:   []string{`os\.execute`},
		}

		validator := NewSecurityValidator(config)

		// String operations should be allowed
		devScript := `
local str = string.upper("hello")
local tbl = table.concat({"a", "b"}, ",")
print(str, tbl)
`
		result, err := validator.ValidateScript(devScript, "test.lua")
		require.NoError(t, err)
		assert.True(t, result.IsValid())

		// But os.execute should still be forbidden due to pattern
		dangerousScript := `os.execute("rm -rf /")`
		result, err = validator.ValidateScript(dangerousScript, "test.lua")
		require.NoError(t, err)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Errors[0].Message, "forbidden pattern")
	})

	t.Run("production_profile_validation", func(t *testing.T) {
		profile := security.ProductionProfile()

		config := &ValidationConfig{
			EnableSecurityCheck: true,
			SecurityProfile:     profile.Name,
			AllowedModules:      profile.AllowedModules,
			ForbiddenFunctions:  profile.ForbiddenFunctions,
		}

		validator := NewSecurityValidator(config)

		// Math operations should be allowed
		mathScript := `
local result = math.sqrt(16)
local rounded = math.floor(3.7)
print(result, rounded)
`
		result, err := validator.ValidateScript(mathScript, "test.lua")
		require.NoError(t, err)
		assert.True(t, result.IsValid())

		// But filesystem access should be forbidden
		fsScript := `
local f = io.open("config.json", "r")
`
		result, err = validator.ValidateScript(fsScript, "test.lua")
		require.NoError(t, err)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Errors[0].Message, "forbidden module")
	})
}

func TestValidationWithSecurityContext(t *testing.T) {
	t.Run("track_violations", func(t *testing.T) {
		profile := security.SandboxProfile()
		securityCtx := security.NewSecurityContext(profile)

		// Simulate validation that would trigger violations
		if !profile.CheckPermission(security.PermissionNetwork) {
			securityCtx.RecordViolation("network", "attempted network access")
		}

		if !profile.CheckPermission(security.PermissionFilesystem) {
			securityCtx.RecordViolation("filesystem", "attempted file access")
		}

		assert.True(t, securityCtx.HasViolations())
		summary := securityCtx.GetViolationSummary()
		assert.Contains(t, summary, "2 security violations")
	})

	t.Run("profile_based_chain", func(t *testing.T) {
		// Create a validation chain that includes security
		chain := NewValidationChain()

		// Add security validator with proper forbidden patterns
		secConfig := DefaultValidationConfig()
		secConfig.SecurityProfile = "sandbox"
		secConfig.ForbiddenPatterns = []string{`os\.execute`}
		chain.AddValidator(NewSecurityValidator(secConfig))

		// Add style validator
		styleConfig := DefaultValidationConfig()
		chain.AddValidator(NewStyleValidator(styleConfig))

		// Test script with both security and style issues
		script := `os.execute("ls")                                                                                        -- very long line with more text`

		result, err := chain.Validate(script, "test.lua")
		require.NoError(t, err)

		assert.False(t, result.IsValid())
		assert.True(t, result.HasErrors())   // Security error
		assert.True(t, result.HasWarnings()) // Style warning
	})
}

func TestProfileManager(t *testing.T) {
	t.Run("get_profile_for_validation", func(t *testing.T) {
		manager := security.NewProfileManager()

		// Get sandbox profile
		profile, err := manager.GetProfile("sandbox")
		require.NoError(t, err)

		// Create validator with profile settings
		config := &ValidationConfig{
			EnableSecurityCheck: true,
			SecurityProfile:     profile.Name,
			AllowedModules:      profile.AllowedModules,
			ForbiddenFunctions:  profile.ForbiddenFunctions,
		}

		validator := NewSecurityValidator(config)
		assert.NotNil(t, validator)

		// Verify settings match profile
		assert.Equal(t, profile.AllowedModules, config.AllowedModules)
	})
}
