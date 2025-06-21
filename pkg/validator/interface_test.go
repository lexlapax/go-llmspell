// ABOUTME: Tests for the validator interface wrapper that unifies validation across engines.
// ABOUTME: Ensures consistent validation behavior and security profile integration.

package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator(t *testing.T) {
	t.Run("validator_result", func(t *testing.T) {
		// Valid result
		result := &ValidationResult{
			Valid:    true,
			Errors:   []ValidationError{},
			Warnings: []ValidationWarning{},
		}

		assert.True(t, result.IsValid())
		assert.False(t, result.HasErrors())
		assert.False(t, result.HasWarnings())

		// Result with errors
		result.Errors = append(result.Errors, ValidationError{
			Type:    "syntax",
			Message: "unexpected EOF",
			Line:    10,
			Column:  5,
		})
		result.Valid = false

		assert.False(t, result.IsValid())
		assert.True(t, result.HasErrors())
		assert.Equal(t, 1, len(result.Errors))

		// Result with warnings
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    "style",
			Message: "line too long",
			Line:    20,
		})

		assert.True(t, result.HasWarnings())
		assert.Equal(t, 1, len(result.Warnings))
	})

	t.Run("validation_config", func(t *testing.T) {
		config := DefaultValidationConfig()

		assert.True(t, config.EnableSyntaxCheck)
		assert.True(t, config.EnableSecurityCheck)
		assert.True(t, config.EnableStyleCheck)
		assert.False(t, config.EnableTypeCheck)
		assert.Equal(t, 10, config.MaxErrors)
		assert.Equal(t, 20, config.MaxWarnings)
	})

	t.Run("base_validator", func(t *testing.T) {
		config := DefaultValidationConfig()
		validator := NewBaseValidator(config)

		assert.NotNil(t, validator)
		assert.Equal(t, config, validator.GetConfig())

		// Test script validation
		result, err := validator.ValidateScript("print('hello')", "test.lua")
		require.NoError(t, err)
		assert.NotNil(t, result)
		// Base validator doesn't implement actual validation
		assert.True(t, result.IsValid())
	})

	t.Run("validation_context", func(t *testing.T) {
		ctx := NewValidationContext("test.lua", map[string]interface{}{
			"strict":  true,
			"profile": "sandbox",
		})

		assert.Equal(t, "test.lua", ctx.Filename)
		assert.Equal(t, true, ctx.Options["strict"])
		assert.Equal(t, "sandbox", ctx.Options["profile"])
		assert.NotNil(t, ctx.Metadata)
		assert.NotZero(t, ctx.StartTime)
	})
}

func TestValidationChain(t *testing.T) {
	t.Run("chain_execution", func(t *testing.T) {
		chain := NewValidationChain()

		// Add mock validators
		validator1 := &mockValidator{
			name: "syntax",
			validate: func(script, filename string) (*ValidationResult, error) {
				return &ValidationResult{
					Valid:  true,
					Errors: []ValidationError{},
				}, nil
			},
		}

		validator2 := &mockValidator{
			name: "security",
			validate: func(script, filename string) (*ValidationResult, error) {
				return &ValidationResult{
					Valid: false,
					Errors: []ValidationError{
						{Type: "security", Message: "forbidden function"},
					},
				}, nil
			},
		}

		chain.AddValidator(validator1)
		chain.AddValidator(validator2)

		result, err := chain.Validate("test script", "test.lua")
		require.NoError(t, err)

		// Chain should aggregate results
		assert.False(t, result.IsValid())
		assert.Equal(t, 1, len(result.Errors))
		assert.Equal(t, "security", result.Errors[0].Type)
	})

	t.Run("chain_short_circuit", func(t *testing.T) {
		chain := NewValidationChain()
		chain.SetShortCircuit(true)

		callCount := 0

		validator1 := &mockValidator{
			name: "first",
			validate: func(script, filename string) (*ValidationResult, error) {
				callCount++
				return &ValidationResult{
					Valid: false,
					Errors: []ValidationError{
						{Type: "error", Message: "first error"},
					},
				}, nil
			},
		}

		validator2 := &mockValidator{
			name: "second",
			validate: func(script, filename string) (*ValidationResult, error) {
				callCount++
				return &ValidationResult{Valid: true}, nil
			},
		}

		chain.AddValidator(validator1)
		chain.AddValidator(validator2)

		result, err := chain.Validate("test", "test.lua")
		require.NoError(t, err)

		// Should stop after first validator fails
		assert.False(t, result.IsValid())
		assert.Equal(t, 1, callCount)
	})
}

func TestValidatorRegistry(t *testing.T) {
	t.Run("register_validator", func(t *testing.T) {
		registry := NewValidatorRegistry()

		validator := &mockValidator{name: "test"}
		err := registry.Register("test", validator)
		assert.NoError(t, err)

		// Get registered validator
		retrieved, err := registry.Get("test")
		require.NoError(t, err)
		assert.Equal(t, validator, retrieved)

		// List validators
		list := registry.List()
		assert.Contains(t, list, "test")
	})

	t.Run("register_duplicate", func(t *testing.T) {
		registry := NewValidatorRegistry()

		validator1 := &mockValidator{name: "test"}
		validator2 := &mockValidator{name: "test2"}

		err := registry.Register("test", validator1)
		assert.NoError(t, err)

		// Should error on duplicate
		err = registry.Register("test", validator2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("get_unknown", func(t *testing.T) {
		registry := NewValidatorRegistry()

		_, err := registry.Get("unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("unregister", func(t *testing.T) {
		registry := NewValidatorRegistry()

		validator := &mockValidator{name: "test"}
		err := registry.Register("test", validator)
		require.NoError(t, err)

		// Unregister
		err = registry.Unregister("test")
		assert.NoError(t, err)

		// Should not be found
		_, err = registry.Get("test")
		assert.Error(t, err)
	})
}

func TestSecurityValidator(t *testing.T) {
	t.Run("security_validation", func(t *testing.T) {
		config := DefaultValidationConfig()
		config.SecurityProfile = "sandbox"

		validator := NewSecurityValidator(config)

		// Test forbidden pattern
		script := `os.execute("rm -rf /")`
		result, err := validator.ValidateScript(script, "test.lua")
		require.NoError(t, err)

		assert.False(t, result.IsValid())
		assert.True(t, result.HasErrors())
		assert.Contains(t, result.Errors[0].Message, "forbidden")
	})

	t.Run("module_restrictions", func(t *testing.T) {
		config := DefaultValidationConfig()
		config.SecurityProfile = "sandbox"
		config.AllowedModules = []string{"string", "table"}

		validator := NewSecurityValidator(config)

		// Test using forbidden module
		script := `local f = io.open("test.txt")`
		result, err := validator.ValidateScript(script, "test.lua")
		require.NoError(t, err)

		assert.False(t, result.IsValid())
		assert.Contains(t, result.Errors[0].Message, "module")
	})
}

func TestStyleValidator(t *testing.T) {
	t.Run("line_length", func(t *testing.T) {
		config := DefaultValidationConfig()
		config.MaxLineLength = 80

		validator := NewStyleValidator(config)

		// Test long line
		script := "local very_long_variable_name = 'this is a very long string that exceeds the maximum line length limit'"
		result, err := validator.ValidateScript(script, "test.lua")
		require.NoError(t, err)

		assert.True(t, result.HasWarnings())
		assert.Contains(t, result.Warnings[0].Message, "line too long")
	})

	t.Run("trailing_whitespace", func(t *testing.T) {
		config := DefaultValidationConfig()
		validator := NewStyleValidator(config)

		// Test trailing whitespace
		script := "local x = 1  \nlocal y = 2"
		result, err := validator.ValidateScript(script, "test.lua")
		require.NoError(t, err)

		assert.True(t, result.HasWarnings())
		assert.Contains(t, result.Warnings[0].Message, "trailing whitespace")
	})
}

// Mock validator for testing
type mockValidator struct {
	name     string
	validate func(script, filename string) (*ValidationResult, error)
}

func (m *mockValidator) ValidateScript(script, filename string) (*ValidationResult, error) {
	if m.validate != nil {
		return m.validate(script, filename)
	}
	return &ValidationResult{Valid: true}, nil
}

func (m *mockValidator) ValidateFile(filename string) (*ValidationResult, error) {
	return &ValidationResult{Valid: true}, nil
}

func (m *mockValidator) GetConfig() *ValidationConfig {
	return DefaultValidationConfig()
}
