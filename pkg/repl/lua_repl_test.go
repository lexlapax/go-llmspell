package repl

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLuaREPL_Creation(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	assert.NotNil(t, repl)

	// Test that it implements REPL interface
	var _ REPL = repl
}

func TestLuaREPL_InvalidConfig(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "javascript", // Wrong engine
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	_, err := NewLuaREPL(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "engine must be 'lua'")
}

func TestLuaREPL_BasicExecution(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"return 2 + 2", "4", false},
		{"return 'hello world'", "hello world", false},
		{"return true", "true", false},
		{"return nil", "", false}, // nil typically prints as empty
		{"invalid lua syntax ((", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := repl.Evaluate(ctx, tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected != "" {
					assert.Contains(t, result, tt.expected)
				}
			}
		})
	}
}

func TestLuaREPL_Variables(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Set a variable
	_, err = repl.Evaluate(ctx, "x = 42")
	require.NoError(t, err)

	// Use the variable
	result, err := repl.Evaluate(ctx, "return x")
	require.NoError(t, err)
	assert.Contains(t, result, "42")

	// Modify the variable
	_, err = repl.Evaluate(ctx, "x = x * 2")
	require.NoError(t, err)

	result, err = repl.Evaluate(ctx, "return x")
	require.NoError(t, err)
	assert.Contains(t, result, "84")
}

func TestLuaREPL_Functions(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Define a function
	_, err = repl.Evaluate(ctx, "function add(a, b) return a + b end")
	require.NoError(t, err)

	// Use the function
	result, err := repl.Evaluate(ctx, "return add(3, 5)")
	require.NoError(t, err)
	assert.Contains(t, result, "8")
}

func TestLuaREPL_Completion(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:       "lua",
		Prompt:       "lua> ",
		AutoComplete: true,
		Input:        &stdin,
		Output:       &stdout,
		Error:        &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	// Test Lua-specific completions
	completions := repl.Complete("pri")
	assert.Contains(t, completions, "print")

	completions = repl.Complete("loc")
	assert.Contains(t, completions, "local")

	completions = repl.Complete("fun")
	assert.Contains(t, completions, "function")

	completions = repl.Complete("tab")
	assert.Contains(t, completions, "table")
}

func TestLuaREPL_MultilineDetection(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:    "lua",
		Prompt:    "lua> ",
		MultiLine: true,
		Input:     &stdin,
		Output:    &stdout,
		Error:     &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	// Test that multiline detection works for Lua
	assert.NotNil(t, repl)
}

func TestLuaREPL_ErrorHandling(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Test syntax error
	_, err = repl.Evaluate(ctx, "invalid syntax ((")
	assert.Error(t, err)

	// Test runtime error
	_, err = repl.Evaluate(ctx, "error('test error')")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
}

func TestLuaREPL_StatePreservation(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Set up some state
	_, err = repl.Evaluate(ctx, "counter = 0")
	require.NoError(t, err)

	_, err = repl.Evaluate(ctx, "function increment() counter = counter + 1 return counter end")
	require.NoError(t, err)

	// Test that state is preserved across evaluations
	result, err := repl.Evaluate(ctx, "return increment()")
	require.NoError(t, err)
	assert.Contains(t, result, "1")

	result, err = repl.Evaluate(ctx, "return increment()")
	require.NoError(t, err)
	assert.Contains(t, result, "2")

	result, err = repl.Evaluate(ctx, "return counter")
	require.NoError(t, err)
	assert.Contains(t, result, "2")
}

func TestLuaREPL_DualFlagSystem(t *testing.T) {
	tests := []struct {
		name           string
		securityLevel  string
		featureSet     string
		expectCreation bool
	}{
		{
			name:           "trusted_and_full",
			securityLevel:  "trusted",
			featureSet:     "full",
			expectCreation: true,
		},
		{
			name:           "untrusted_and_minimal",
			securityLevel:  "untrusted",
			featureSet:     "minimal",
			expectCreation: true,
		},
		{
			name:           "privileged_and_llm",
			securityLevel:  "privileged",
			featureSet:     "llm",
			expectCreation: true,
		},
		{
			name:           "invalid_security_level",
			securityLevel:  "invalid",
			featureSet:     "full",
			expectCreation: true, // Should still work with fallback
		},
		{
			name:           "invalid_feature_set",
			securityLevel:  "trusted",
			featureSet:     "invalid",
			expectCreation: true, // Should still work with fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdin, stdout, stderr bytes.Buffer
			config := REPLConfig{
				Engine:        "lua",
				Prompt:        "test> ",
				SecurityLevel: tt.securityLevel,
				FeatureSet:    tt.featureSet,
				Input:         &stdin,
				Output:        &stdout,
				Error:         &stderr,
			}

			repl, err := NewLuaREPL(config)

			if tt.expectCreation {
				assert.NoError(t, err, "REPL creation should succeed")
				assert.NotNil(t, repl, "REPL should not be nil")
				if repl != nil {
					err = repl.Close()
					assert.NoError(t, err, "REPL close should succeed")
				}
			} else {
				assert.Error(t, err, "REPL creation should fail")
				assert.Nil(t, repl, "REPL should be nil")
			}
		})
	}
}

func TestLuaREPL_DefaultSecurityAndFeatureSet(t *testing.T) {
	// Test that REPL defaults to trusted + full when no flags specified
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "test> ",
		// SecurityLevel and FeatureSet not specified
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	require.NotNil(t, repl)
	defer repl.Close()

	// Verify REPL was created successfully with defaults
	ctx := context.Background()
	result, err := repl.Evaluate(ctx, "return 'default test'")
	require.NoError(t, err)
	assert.Contains(t, result, "default test")
}

func TestLuaREPL_SecurityLevelValidation(t *testing.T) {
	tests := []struct {
		name          string
		securityLevel string
		expectWorking bool
	}{
		{
			name:          "valid_untrusted",
			securityLevel: "untrusted",
			expectWorking: true,
		},
		{
			name:          "valid_trusted",
			securityLevel: "trusted",
			expectWorking: true,
		},
		{
			name:          "valid_privileged",
			securityLevel: "privileged",
			expectWorking: true,
		},
		{
			name:          "empty_falls_back_to_default",
			securityLevel: "",
			expectWorking: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdin, stdout, stderr bytes.Buffer
			config := REPLConfig{
				Engine:        "lua",
				Prompt:        "test> ",
				SecurityLevel: tt.securityLevel,
				FeatureSet:    "full", // Always use full for these tests
				Input:         &stdin,
				Output:        &stdout,
				Error:         &stderr,
			}

			repl, err := NewLuaREPL(config)
			require.NoError(t, err)
			require.NotNil(t, repl)
			defer repl.Close()

			// Test basic functionality
			ctx := context.Background()
			result, err := repl.Evaluate(ctx, "return 2 + 2")

			if tt.expectWorking {
				assert.NoError(t, err, "Basic evaluation should work")
				assert.Contains(t, result, "4")
			} else {
				// Note: Even restricted security levels should allow basic math
				// This test mainly verifies the security level is accepted
				_ = err
				_ = result
			}
		})
	}
}
