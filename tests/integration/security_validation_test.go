// ABOUTME: Integration tests that combine security levels with validation.
// ABOUTME: Tests the end-to-end validation flow with security constraints.

package integration

import (
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/security"
	"github.com/lexlapax/go-llmspell/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityIntegration(t *testing.T) {
	t.Run("untrusted_level_validation", func(t *testing.T) {
		// Create untrusted security config
		secConfig := security.UntrustedLevel()

		// Create validation config based on security level
		config := &validator.ValidationConfig{
			EnableSecurityCheck: true,
			SecurityProfile:     string(secConfig.Level),
			AllowedModules:      secConfig.AllowedModules,
			ForbiddenFunctions:  secConfig.ForbiddenFunctions,
		}

		val := validator.NewSecurityValidator(config)

		// Test allowed code
		allowedScript := `
local function add(a, b)
    return a + b
end

local result = add(1, 2)
print(result)
`
		result, err := val.ValidateScript(allowedScript, "test.lua")
		require.NoError(t, err)
		assert.True(t, result.IsValid())

		// Test forbidden code (io module not allowed in untrusted)
		forbiddenScript := `
local f = io.open("test.txt", "w")
f:write("test")
f:close()
`
		result, err = val.ValidateScript(forbiddenScript, "test.lua")
		require.NoError(t, err)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Errors[0].Message, "forbidden module")
	})

	t.Run("trusted_level_validation", func(t *testing.T) {
		// Create trusted security config
		secConfig := security.TrustedLevel()

		config := &validator.ValidationConfig{
			EnableSecurityCheck: true,
			SecurityProfile:     string(secConfig.Level),
			AllowedModules:      secConfig.AllowedModules,
			ForbiddenFunctions:  secConfig.ForbiddenFunctions,
			ForbiddenPatterns:   []string{`os\.execute`},
		}

		val := validator.NewSecurityValidator(config)

		// String operations should be allowed
		devScript := `
local str = string.upper("hello")
local tbl = table.concat({"a", "b"}, ",")
print(str, tbl)
`
		result, err := val.ValidateScript(devScript, "test.lua")
		require.NoError(t, err)
		assert.True(t, result.IsValid())

		// But os.execute should still be forbidden due to pattern
		dangerousScript := `os.execute("rm -rf /")`
		result, err = val.ValidateScript(dangerousScript, "test.lua")
		require.NoError(t, err)
		assert.False(t, result.IsValid())
		assert.Contains(t, result.Errors[0].Message, "forbidden pattern")
	})

	t.Run("privileged_level_validation", func(t *testing.T) {
		secConfig := security.PrivilegedLevel()

		config := &validator.ValidationConfig{
			EnableSecurityCheck: true,
			SecurityProfile:     string(secConfig.Level),
			AllowedModules:      secConfig.AllowedModules,
			ForbiddenFunctions:  secConfig.ForbiddenFunctions,
		}

		val := validator.NewSecurityValidator(config)

		// Math operations should be allowed
		mathScript := `
local result = math.sqrt(16)
local rounded = math.floor(3.7)
print(result, rounded)
`
		result, err := val.ValidateScript(mathScript, "test.lua")
		require.NoError(t, err)
		assert.True(t, result.IsValid())

		// Privileged level allows io module
		fsScript := `
local io = io
print(io)`
		result, err = val.ValidateScript(fsScript, "test.lua")
		require.NoError(t, err)
		if !result.IsValid() && len(result.Errors) > 0 {
			t.Logf("Validation failed with error: %s", result.Errors[0].Message)
		}
		assert.True(t, result.IsValid()) // Should be allowed in privileged
	})
}

func TestValidationWithSecurityContext(t *testing.T) {
	t.Run("track_violations", func(t *testing.T) {
		config := security.UntrustedLevel()
		securityCtx := security.NewSecurityContext(config)

		// Simulate validation that would trigger violations
		if !config.CheckPermission(security.PermissionNetwork) {
			securityCtx.RecordViolation("network", "attempted network access")
		}

		if !config.CheckPermission(security.PermissionFilesystem) {
			securityCtx.RecordViolation("filesystem", "attempted file access")
		}

		assert.True(t, securityCtx.HasViolations())
		summary := securityCtx.GetViolationSummary()
		assert.Contains(t, summary, "2 security violations")
	})

	t.Run("security_level_based_chain", func(t *testing.T) {
		// Create a validation chain that includes security
		chain := validator.NewValidationChain()

		// Add security validator with proper forbidden patterns
		secConfig := validator.DefaultValidationConfig()
		secConfig.SecurityProfile = "untrusted"
		secConfig.ForbiddenPatterns = []string{`os\.execute`}
		chain.AddValidator(validator.NewSecurityValidator(secConfig))

		// Add style validator
		styleConfig := validator.DefaultValidationConfig()
		chain.AddValidator(validator.NewStyleValidator(styleConfig))

		// Test script with both security and style issues
		script := `os.execute("ls")                                                                                        -- very long line with more text`

		result, err := chain.Validate(script, "test.lua")
		require.NoError(t, err)

		assert.False(t, result.IsValid())
		assert.True(t, result.HasErrors())   // Security error
		assert.True(t, result.HasWarnings()) // Style warning
	})
}

func TestLevelManager(t *testing.T) {
	t.Run("get_level_for_validation", func(t *testing.T) {
		manager := security.NewLevelManager()

		// Get untrusted level
		config, err := manager.GetLevel(security.SecurityLevelUntrusted)
		require.NoError(t, err)

		// Create validator with level settings
		valConfig := &validator.ValidationConfig{
			EnableSecurityCheck: true,
			SecurityProfile:     string(config.Level),
			AllowedModules:      config.AllowedModules,
			ForbiddenFunctions:  config.ForbiddenFunctions,
		}

		val := validator.NewSecurityValidator(valConfig)
		assert.NotNil(t, val)

		// Verify settings match level
		assert.Equal(t, config.AllowedModules, valConfig.AllowedModules)
	})
}
