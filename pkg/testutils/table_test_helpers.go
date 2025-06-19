// ABOUTME: Table-driven test helpers for method and validation testing across bridge packages
// ABOUTME: Provides standardized test case structures and execution functions for consistent testing patterns

package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// MethodTestCase represents a test case for bridge method execution
type MethodTestCase struct {
	Name       string               // Test case name
	Method     string               // Method name to test
	Args       []engine.ScriptValue // Method arguments
	WantResult engine.ScriptValue   // Expected result
	WantErr    bool                 // Whether error is expected
	WantErrMsg string               // Expected error message (partial match)
	Setup      func(t *testing.T)   // Optional setup function
	Teardown   func(t *testing.T)   // Optional teardown function
	SkipReason string               // Optional skip reason
}

// ValidationTestCase represents a test case for ValidateMethod testing
type ValidationTestCase struct {
	Name       string               // Test case name
	Method     string               // Method name to validate
	Args       []engine.ScriptValue // Method arguments
	WantErr    bool                 // Whether validation error is expected
	WantErrMsg string               // Expected error message (partial match)
	Setup      func(t *testing.T)   // Optional setup function
	Teardown   func(t *testing.T)   // Optional teardown function
	SkipReason string               // Optional skip reason
}

// BridgeMethodExecutor defines the interface for executing bridge methods
type BridgeMethodExecutor interface {
	ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error)
}

// BridgeMethodValidator defines the interface for validating bridge methods
type BridgeMethodValidator interface {
	ValidateMethod(method string, args []engine.ScriptValue) error
}

// RunMethodTests executes a series of method test cases against a bridge
func RunMethodTests(t *testing.T, executor BridgeMethodExecutor, testCases []MethodTestCase) {
	t.Helper()

	for _, tc := range testCases {
		tc := tc // capture loop variable
		t.Run(tc.Name, func(t *testing.T) {
			// Skip test if reason provided
			if tc.SkipReason != "" {
				t.Skip(tc.SkipReason)
				return
			}

			// Run setup if provided
			if tc.Setup != nil {
				tc.Setup(t)
			}

			// Run teardown if provided
			if tc.Teardown != nil {
				defer tc.Teardown(t)
			}

			// Execute the method
			ctx := context.Background()
			result, err := executor.ExecuteMethod(ctx, tc.Method, tc.Args)

			// Check error expectations
			if tc.WantErr {
				assert.Error(t, err, "Expected error but got none")
				if tc.WantErrMsg != "" {
					assert.Contains(t, err.Error(), tc.WantErrMsg, "Error message should contain expected text")
				}
				return
			}

			// Check for unexpected error
			require.NoError(t, err, "Unexpected error occurred")

			// Compare results if expected result provided
			if tc.WantResult != nil {
				AssertScriptValueEquals(t, tc.WantResult, result)
			}
		})
	}
}

// RunValidationTests executes a series of validation test cases against a bridge
func RunValidationTests(t *testing.T, validator BridgeMethodValidator, testCases []ValidationTestCase) {
	t.Helper()

	for _, tc := range testCases {
		tc := tc // capture loop variable
		t.Run(tc.Name, func(t *testing.T) {
			// Skip test if reason provided
			if tc.SkipReason != "" {
				t.Skip(tc.SkipReason)
				return
			}

			// Run setup if provided
			if tc.Setup != nil {
				tc.Setup(t)
			}

			// Run teardown if provided
			if tc.Teardown != nil {
				defer tc.Teardown(t)
			}

			// Execute validation
			err := validator.ValidateMethod(tc.Method, tc.Args)

			// Check error expectations
			if tc.WantErr {
				assert.Error(t, err, "Expected validation error but got none")
				if tc.WantErrMsg != "" {
					assert.Contains(t, err.Error(), tc.WantErrMsg, "Validation error message should contain expected text")
				}
				return
			}

			// Check for unexpected error
			assert.NoError(t, err, "Unexpected validation error occurred")
		})
	}
}

// CreateMethodTestCase is a builder function for MethodTestCase
func CreateMethodTestCase(name, method string) *MethodTestCase {
	return &MethodTestCase{
		Name:   name,
		Method: method,
		Args:   []engine.ScriptValue{},
	}
}

// WithArgs sets the arguments for a MethodTestCase
func (tc *MethodTestCase) WithArgs(args ...engine.ScriptValue) *MethodTestCase {
	tc.Args = args
	return tc
}

// WithResult sets the expected result for a MethodTestCase
func (tc *MethodTestCase) WithResult(result engine.ScriptValue) *MethodTestCase {
	tc.WantResult = result
	return tc
}

// WithError sets error expectation for a MethodTestCase
func (tc *MethodTestCase) WithError(wantErr bool, errMsg string) *MethodTestCase {
	tc.WantErr = wantErr
	tc.WantErrMsg = errMsg
	return tc
}

// WithSetup sets the setup function for a MethodTestCase
func (tc *MethodTestCase) WithSetup(setup func(t *testing.T)) *MethodTestCase {
	tc.Setup = setup
	return tc
}

// WithTeardown sets the teardown function for a MethodTestCase
func (tc *MethodTestCase) WithTeardown(teardown func(t *testing.T)) *MethodTestCase {
	tc.Teardown = teardown
	return tc
}

// WithSkip sets the skip reason for a MethodTestCase
func (tc *MethodTestCase) WithSkip(reason string) *MethodTestCase {
	tc.SkipReason = reason
	return tc
}

// CreateValidationTestCase is a builder function for ValidationTestCase
func CreateValidationTestCase(name, method string) *ValidationTestCase {
	return &ValidationTestCase{
		Name:   name,
		Method: method,
		Args:   []engine.ScriptValue{},
	}
}

// WithArgs sets the arguments for a ValidationTestCase
func (tc *ValidationTestCase) WithArgs(args ...engine.ScriptValue) *ValidationTestCase {
	tc.Args = args
	return tc
}

// WithError sets error expectation for a ValidationTestCase
func (tc *ValidationTestCase) WithError(wantErr bool, errMsg string) *ValidationTestCase {
	tc.WantErr = wantErr
	tc.WantErrMsg = errMsg
	return tc
}

// WithSetup sets the setup function for a ValidationTestCase
func (tc *ValidationTestCase) WithSetup(setup func(t *testing.T)) *ValidationTestCase {
	tc.Setup = setup
	return tc
}

// WithTeardown sets the teardown function for a ValidationTestCase
func (tc *ValidationTestCase) WithTeardown(teardown func(t *testing.T)) *ValidationTestCase {
	tc.Teardown = teardown
	return tc
}

// WithSkip sets the skip reason for a ValidationTestCase
func (tc *ValidationTestCase) WithSkip(reason string) *ValidationTestCase {
	tc.SkipReason = reason
	return tc
}
