// ABOUTME: Tests for table-driven test helpers ensuring proper test case execution and validation
// ABOUTME: Validates MethodTestCase and ValidationTestCase functionality with mock implementations

package testutils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Mock implementation for testing
type mockMethodExecutor struct {
	results map[string]engine.ScriptValue
	errors  map[string]error
}

func (m *mockMethodExecutor) ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err, exists := m.errors[method]; exists {
		return nil, err
	}
	if result, exists := m.results[method]; exists {
		return result, nil
	}
	return engine.NewStringValue("default"), nil
}

type mockMethodValidator struct {
	errors map[string]error
}

func (m *mockMethodValidator) ValidateMethod(method string, args []engine.ScriptValue) error {
	if err, exists := m.errors[method]; exists {
		return err
	}
	return nil
}

func TestCreateMethodTestCase(t *testing.T) {
	tc := CreateMethodTestCase("test_case", "test_method")

	assert.Equal(t, "test_case", tc.Name)
	assert.Equal(t, "test_method", tc.Method)
	assert.Empty(t, tc.Args)
	assert.Nil(t, tc.WantResult)
	assert.False(t, tc.WantErr)
	assert.Empty(t, tc.WantErrMsg)
}

func TestMethodTestCaseBuilder(t *testing.T) {
	setupCalled := false
	teardownCalled := false

	tc := CreateMethodTestCase("builder_test", "test_method").
		WithArgs(engine.NewStringValue("arg1"), engine.NewNumberValue(42)).
		WithResult(engine.NewStringValue("expected")).
		WithError(true, "expected error").
		WithSetup(func(t *testing.T) { setupCalled = true }).
		WithTeardown(func(t *testing.T) { teardownCalled = true }).
		WithSkip("skip reason")

	assert.Equal(t, "builder_test", tc.Name)
	assert.Equal(t, "test_method", tc.Method)
	assert.Len(t, tc.Args, 2)
	assert.Equal(t, "arg1", tc.Args[0].(engine.StringValue).Value())
	assert.Equal(t, float64(42), tc.Args[1].(engine.NumberValue).Value())
	assert.Equal(t, "expected", tc.WantResult.(engine.StringValue).Value())
	assert.True(t, tc.WantErr)
	assert.Equal(t, "expected error", tc.WantErrMsg)
	assert.Equal(t, "skip reason", tc.SkipReason)
	assert.NotNil(t, tc.Setup)
	assert.NotNil(t, tc.Teardown)

	// Test setup/teardown functions
	tc.Setup(t)
	tc.Teardown(t)
	assert.True(t, setupCalled)
	assert.True(t, teardownCalled)
}

func TestCreateValidationTestCase(t *testing.T) {
	tc := CreateValidationTestCase("validation_test", "validate_method")

	assert.Equal(t, "validation_test", tc.Name)
	assert.Equal(t, "validate_method", tc.Method)
	assert.Empty(t, tc.Args)
	assert.False(t, tc.WantErr)
	assert.Empty(t, tc.WantErrMsg)
}

func TestValidationTestCaseBuilder(t *testing.T) {
	setupCalled := false
	teardownCalled := false

	tc := CreateValidationTestCase("validation_builder", "validate_method").
		WithArgs(engine.NewStringValue("arg1")).
		WithError(true, "validation error").
		WithSetup(func(t *testing.T) { setupCalled = true }).
		WithTeardown(func(t *testing.T) { teardownCalled = true }).
		WithSkip("skip validation")

	assert.Equal(t, "validation_builder", tc.Name)
	assert.Equal(t, "validate_method", tc.Method)
	assert.Len(t, tc.Args, 1)
	assert.True(t, tc.WantErr)
	assert.Equal(t, "validation error", tc.WantErrMsg)
	assert.Equal(t, "skip validation", tc.SkipReason)

	tc.Setup(t)
	tc.Teardown(t)
	assert.True(t, setupCalled)
	assert.True(t, teardownCalled)
}

func TestRunMethodTests(t *testing.T) {
	executor := &mockMethodExecutor{
		results: map[string]engine.ScriptValue{
			"success_method": engine.NewStringValue("success"),
		},
		errors: map[string]error{
			"error_method": errors.New("test error message"),
		},
	}

	testCases := []MethodTestCase{
		{
			Name:       "successful_execution",
			Method:     "success_method",
			Args:       []engine.ScriptValue{},
			WantResult: engine.NewStringValue("success"),
			WantErr:    false,
		},
		{
			Name:       "error_execution",
			Method:     "error_method",
			Args:       []engine.ScriptValue{},
			WantErr:    true,
			WantErrMsg: "test error",
		},
		{
			Name:       "skipped_test",
			Method:     "any_method",
			SkipReason: "testing skip functionality",
		},
	}

	// This should run without panicking or failing
	RunMethodTests(t, executor, testCases)
}

func TestRunValidationTests(t *testing.T) {
	validator := &mockMethodValidator{
		errors: map[string]error{
			"invalid_method": errors.New("validation failed"),
		},
	}

	testCases := []ValidationTestCase{
		{
			Name:    "valid_method",
			Method:  "valid_method",
			Args:    []engine.ScriptValue{},
			WantErr: false,
		},
		{
			Name:       "invalid_method",
			Method:     "invalid_method",
			Args:       []engine.ScriptValue{},
			WantErr:    true,
			WantErrMsg: "validation failed",
		},
		{
			Name:       "skipped_validation",
			Method:     "any_method",
			SkipReason: "testing skip in validation",
		},
	}

	// This should run without panicking or failing
	RunValidationTests(t, validator, testCases)
}

func TestMethodTestExecutionWithSetupTeardown(t *testing.T) {
	executor := &mockMethodExecutor{
		results: map[string]engine.ScriptValue{
			"test_method": engine.NewStringValue("result"),
		},
	}

	setupExecuted := false
	teardownExecuted := false

	testCases := []MethodTestCase{
		{
			Name:   "setup_teardown_test",
			Method: "test_method",
			Setup: func(t *testing.T) {
				setupExecuted = true
			},
			Teardown: func(t *testing.T) {
				teardownExecuted = true
			},
			WantResult: engine.NewStringValue("result"),
		},
	}

	RunMethodTests(t, executor, testCases)

	assert.True(t, setupExecuted, "Setup function should have been called")
	assert.True(t, teardownExecuted, "Teardown function should have been called")
}

func TestValidationTestExecutionWithSetupTeardown(t *testing.T) {
	validator := &mockMethodValidator{
		errors: map[string]error{},
	}

	setupExecuted := false
	teardownExecuted := false

	testCases := []ValidationTestCase{
		{
			Name:   "validation_setup_teardown",
			Method: "valid_method",
			Setup: func(t *testing.T) {
				setupExecuted = true
			},
			Teardown: func(t *testing.T) {
				teardownExecuted = true
			},
		},
	}

	RunValidationTests(t, validator, testCases)

	assert.True(t, setupExecuted, "Setup function should have been called")
	assert.True(t, teardownExecuted, "Teardown function should have been called")
}
