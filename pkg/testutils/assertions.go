// ABOUTME: ScriptValue assertion helpers provide comprehensive type checking and validation utilities
// ABOUTME: Simplifies test assertions for ScriptValue types with clear error messages and common patterns

package testutils

import (
	"strings"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertScriptValueType checks if result is of expected type
func AssertScriptValueType(t *testing.T, result engine.ScriptValue, expectedType engine.ScriptValueType) {
	t.Helper()
	require.NotNil(t, result, "ScriptValue should not be nil")
	assert.Equal(t, expectedType, result.Type(),
		"Expected type %s, got %s", expectedType, result.Type())
}

// AssertIsString asserts that a ScriptValue is a string and returns the value
func AssertIsString(t *testing.T, result engine.ScriptValue) string {
	t.Helper()
	AssertScriptValueType(t, result, engine.TypeString)
	sv, ok := result.(engine.StringValue)
	require.True(t, ok, "Failed to cast to StringValue")
	return sv.Value()
}

// AssertStringEquals asserts that a ScriptValue is a string with expected value
func AssertStringEquals(t *testing.T, result engine.ScriptValue, expected string) {
	t.Helper()
	actual := AssertIsString(t, result)
	assert.Equal(t, expected, actual, "String value mismatch")
}

// AssertStringContains asserts that a ScriptValue is a string containing substring
func AssertStringContains(t *testing.T, result engine.ScriptValue, substring string) {
	t.Helper()
	actual := AssertIsString(t, result)
	assert.Contains(t, actual, substring,
		"String '%s' should contain '%s'", actual, substring)
}

// AssertIsNumber asserts that a ScriptValue is a number and returns the value
func AssertIsNumber(t *testing.T, result engine.ScriptValue) float64 {
	t.Helper()
	AssertScriptValueType(t, result, engine.TypeNumber)
	nv, ok := result.(engine.NumberValue)
	require.True(t, ok, "Failed to cast to NumberValue")
	return nv.Value()
}

// AssertNumberEquals asserts that a ScriptValue is a number with expected value
func AssertNumberEquals(t *testing.T, result engine.ScriptValue, expected float64) {
	t.Helper()
	actual := AssertIsNumber(t, result)
	assert.Equal(t, expected, actual, "Number value mismatch")
}

// AssertNumberInRange asserts that a ScriptValue is a number within range
func AssertNumberInRange(t *testing.T, result engine.ScriptValue, min, max float64) {
	t.Helper()
	actual := AssertIsNumber(t, result)
	assert.GreaterOrEqual(t, actual, min, "Number %f should be >= %f", actual, min)
	assert.LessOrEqual(t, actual, max, "Number %f should be <= %f", actual, max)
}

// AssertIsBool asserts that a ScriptValue is a boolean and returns the value
func AssertIsBool(t *testing.T, result engine.ScriptValue) bool {
	t.Helper()
	AssertScriptValueType(t, result, engine.TypeBool)
	bv, ok := result.(engine.BoolValue)
	require.True(t, ok, "Failed to cast to BoolValue")
	return bv.Value()
}

// AssertBoolEquals asserts that a ScriptValue is a boolean with expected value
func AssertBoolEquals(t *testing.T, result engine.ScriptValue, expected bool) {
	t.Helper()
	actual := AssertIsBool(t, result)
	assert.Equal(t, expected, actual, "Boolean value mismatch")
}

// AssertIsNil asserts that a ScriptValue is nil
func AssertIsNil(t *testing.T, result engine.ScriptValue) {
	t.Helper()
	if result == nil {
		return // nil interface is acceptable
	}
	assert.True(t, result.IsNil(), "Expected nil value, got %s", result.Type())
}

// AssertNotNil asserts that a ScriptValue is not nil
func AssertNotNil(t *testing.T, result engine.ScriptValue) {
	t.Helper()
	require.NotNil(t, result, "ScriptValue should not be nil")
	assert.False(t, result.IsNil(), "ScriptValue should not be nil type")
}

// AssertErrorValue checks if result is an ErrorValue with optional message check
func AssertErrorValue(t *testing.T, result engine.ScriptValue, expectedMessage string) {
	t.Helper()
	AssertScriptValueType(t, result, engine.TypeError)

	ev, ok := result.(engine.ErrorValue)
	require.True(t, ok, "Failed to cast to ErrorValue")

	err := ev.Error()
	require.NotNil(t, err, "ErrorValue should contain an error")

	if expectedMessage != "" {
		assert.Contains(t, err.Error(), expectedMessage,
			"Error message should contain '%s'", expectedMessage)
	}
}

// AssertIsError asserts that a ScriptValue is an error and returns it
func AssertIsError(t *testing.T, result engine.ScriptValue) error {
	t.Helper()
	AssertScriptValueType(t, result, engine.TypeError)
	ev, ok := result.(engine.ErrorValue)
	require.True(t, ok, "Failed to cast to ErrorValue")
	return ev.Error()
}

// AssertObjectValue asserts that a ScriptValue is an object and returns fields
func AssertObjectValue(t *testing.T, result engine.ScriptValue) map[string]engine.ScriptValue {
	t.Helper()
	AssertScriptValueType(t, result, engine.TypeObject)
	ov, ok := result.(engine.ObjectValue)
	require.True(t, ok, "Failed to cast to ObjectValue")
	return ov.Fields()
}

// AssertObjectHasFields verifies an ObjectValue has expected fields
func AssertObjectHasFields(t *testing.T, result engine.ScriptValue, expectedFields ...string) {
	t.Helper()
	fields := AssertObjectValue(t, result)

	for _, fieldName := range expectedFields {
		_, exists := fields[fieldName]
		assert.True(t, exists, "Object should have field '%s'", fieldName)
	}
}

// AssertObjectFieldEquals checks a specific field value in an object
func AssertObjectFieldEquals(t *testing.T, result engine.ScriptValue, fieldName string, expected interface{}) {
	t.Helper()
	fields := AssertObjectValue(t, result)

	fieldValue, exists := fields[fieldName]
	require.True(t, exists, "Object should have field '%s'", fieldName)

	expectedValue := InterfaceToScriptValue(expected)
	assert.Equal(t, expectedValue, fieldValue,
		"Object field '%s' mismatch", fieldName)
}

// AssertObjectFieldCount checks the number of fields in an object
func AssertObjectFieldCount(t *testing.T, result engine.ScriptValue, expectedCount int) {
	t.Helper()
	fields := AssertObjectValue(t, result)
	assert.Len(t, fields, expectedCount,
		"Object should have %d fields", expectedCount)
}

// AssertArrayValue asserts that a ScriptValue is an array and returns elements
func AssertArrayValue(t *testing.T, result engine.ScriptValue) []engine.ScriptValue {
	t.Helper()
	AssertScriptValueType(t, result, engine.TypeArray)
	av, ok := result.(engine.ArrayValue)
	require.True(t, ok, "Failed to cast to ArrayValue")
	return av.Elements()
}

// AssertArrayLength verifies an ArrayValue has expected length
func AssertArrayLength(t *testing.T, result engine.ScriptValue, expectedLength int) {
	t.Helper()
	elements := AssertArrayValue(t, result)
	assert.Len(t, elements, expectedLength,
		"Array should have %d elements", expectedLength)
}

// AssertArrayElementEquals checks a specific element in an array
func AssertArrayElementEquals(t *testing.T, result engine.ScriptValue, index int, expected interface{}) {
	t.Helper()
	elements := AssertArrayValue(t, result)

	require.Less(t, index, len(elements),
		"Array index %d out of bounds (length %d)", index, len(elements))
	require.GreaterOrEqual(t, index, 0, "Array index must be non-negative")

	expectedValue := InterfaceToScriptValue(expected)
	assert.Equal(t, expectedValue, elements[index],
		"Array element at index %d mismatch", index)
}

// AssertArrayContains checks if array contains a specific value
func AssertArrayContains(t *testing.T, result engine.ScriptValue, expected interface{}) {
	t.Helper()
	elements := AssertArrayValue(t, result)
	expectedValue := InterfaceToScriptValue(expected)

	found := false
	for _, elem := range elements {
		if elem.Equals(expectedValue) {
			found = true
			break
		}
	}

	assert.True(t, found, "Array should contain value: %v", expected)
}

// AssertCustomValue asserts that a ScriptValue is a custom type
func AssertCustomValue(t *testing.T, result engine.ScriptValue, expectedType string) interface{} {
	t.Helper()
	AssertScriptValueType(t, result, engine.TypeCustom)
	cv, ok := result.(engine.CustomValue)
	require.True(t, ok, "Failed to cast to CustomValue")

	if expectedType != "" {
		assert.Equal(t, expectedType, cv.TypeName(),
			"Custom type mismatch: expected %s, got %s", expectedType, cv.TypeName())
	}

	return cv.Value()
}

// RequireNoGoError asserts no Go error (for methods that should return ErrorValue)
func RequireNoGoError(t *testing.T, err error, msg string) {
	t.Helper()
	require.NoError(t, err, msg)
}

// RequireGoError asserts a Go error occurred
func RequireGoError(t *testing.T, err error, msg string) {
	t.Helper()
	require.Error(t, err, msg)
}

// AssertScriptValueEquals compares two ScriptValues for equality
func AssertScriptValueEquals(t *testing.T, expected, actual engine.ScriptValue) {
	t.Helper()

	if expected == nil && actual == nil {
		return
	}

	require.NotNil(t, actual, "Actual value is nil but expected is not")
	require.NotNil(t, expected, "Expected value is nil but actual is not")

	assert.Equal(t, expected.Type(), actual.Type(),
		"Type mismatch: expected %s, got %s", expected.Type(), actual.Type())

	// Use the Equals method if available
	type equaler interface {
		Equals(other engine.ScriptValue) bool
	}

	if eq, ok := expected.(equaler); ok {
		assert.True(t, eq.Equals(actual),
			"Values not equal: expected %v, got %v", expected, actual)
		return
	}

	// Fallback to direct comparison
	assert.Equal(t, expected, actual, "ScriptValue mismatch")
}

// AssertScriptValuesEqual compares two slices of ScriptValues
func AssertScriptValuesEqual(t *testing.T, expected, actual []engine.ScriptValue) {
	t.Helper()

	require.Len(t, actual, len(expected),
		"Length mismatch: expected %d, got %d", len(expected), len(actual))

	for i := range expected {
		AssertScriptValueEquals(t, expected[i], actual[i])
	}
}

// ComprehensiveValueAssertion performs multiple assertions on a ScriptValue
type ComprehensiveValueAssertion struct {
	t            *testing.T
	value        engine.ScriptValue
	expectedType engine.ScriptValueType
}

// NewComprehensiveAssertion creates a new comprehensive assertion helper
func NewComprehensiveAssertion(t *testing.T, value engine.ScriptValue) *ComprehensiveValueAssertion {
	t.Helper()
	return &ComprehensiveValueAssertion{
		t:     t,
		value: value,
	}
}

// WithType sets the expected type and verifies it
func (a *ComprehensiveValueAssertion) WithType(expectedType engine.ScriptValueType) *ComprehensiveValueAssertion {
	a.t.Helper()
	a.expectedType = expectedType
	AssertScriptValueType(a.t, a.value, expectedType)
	return a
}

// IsString asserts string type and returns string assertion
func (a *ComprehensiveValueAssertion) IsString() *StringAssertion {
	a.t.Helper()
	a.WithType(engine.TypeString)
	return &StringAssertion{
		t:     a.t,
		value: AssertIsString(a.t, a.value),
	}
}

// IsNumber asserts number type and returns number assertion
func (a *ComprehensiveValueAssertion) IsNumber() *NumberAssertion {
	a.t.Helper()
	a.WithType(engine.TypeNumber)
	return &NumberAssertion{
		t:     a.t,
		value: AssertIsNumber(a.t, a.value),
	}
}

// IsObject asserts object type and returns object assertion
func (a *ComprehensiveValueAssertion) IsObject() *ObjectAssertion {
	a.t.Helper()
	a.WithType(engine.TypeObject)
	return &ObjectAssertion{
		t:      a.t,
		fields: AssertObjectValue(a.t, a.value),
	}
}

// IsArray asserts array type and returns array assertion
func (a *ComprehensiveValueAssertion) IsArray() *ArrayAssertion {
	a.t.Helper()
	a.WithType(engine.TypeArray)
	return &ArrayAssertion{
		t:        a.t,
		elements: AssertArrayValue(a.t, a.value),
	}
}

// StringAssertion provides string-specific assertions
type StringAssertion struct {
	t     *testing.T
	value string
}

// Equals asserts string equality
func (a *StringAssertion) Equals(expected string) *StringAssertion {
	a.t.Helper()
	assert.Equal(a.t, expected, a.value)
	return a
}

// Contains asserts string contains substring
func (a *StringAssertion) Contains(substring string) *StringAssertion {
	a.t.Helper()
	assert.Contains(a.t, a.value, substring)
	return a
}

// HasPrefix asserts string has prefix
func (a *StringAssertion) HasPrefix(prefix string) *StringAssertion {
	a.t.Helper()
	assert.True(a.t, strings.HasPrefix(a.value, prefix),
		"String '%s' should have prefix '%s'", a.value, prefix)
	return a
}

// NumberAssertion provides number-specific assertions
type NumberAssertion struct {
	t     *testing.T
	value float64
}

// Equals asserts number equality
func (a *NumberAssertion) Equals(expected float64) *NumberAssertion {
	a.t.Helper()
	assert.Equal(a.t, expected, a.value)
	return a
}

// InRange asserts number is within range
func (a *NumberAssertion) InRange(min, max float64) *NumberAssertion {
	a.t.Helper()
	assert.GreaterOrEqual(a.t, a.value, min)
	assert.LessOrEqual(a.t, a.value, max)
	return a
}

// GreaterThan asserts number is greater than value
func (a *NumberAssertion) GreaterThan(min float64) *NumberAssertion {
	a.t.Helper()
	assert.Greater(a.t, a.value, min)
	return a
}

// ObjectAssertion provides object-specific assertions
type ObjectAssertion struct {
	t      *testing.T
	fields map[string]engine.ScriptValue
}

// HasFields asserts object has specific fields
func (a *ObjectAssertion) HasFields(fieldNames ...string) *ObjectAssertion {
	a.t.Helper()
	for _, name := range fieldNames {
		_, exists := a.fields[name]
		assert.True(a.t, exists, "Object should have field '%s'", name)
	}
	return a
}

// WithFieldCount asserts object has specific number of fields
func (a *ObjectAssertion) WithFieldCount(count int) *ObjectAssertion {
	a.t.Helper()
	assert.Len(a.t, a.fields, count)
	return a
}

// ArrayAssertion provides array-specific assertions
type ArrayAssertion struct {
	t        *testing.T
	elements []engine.ScriptValue
}

// WithLength asserts array has specific length
func (a *ArrayAssertion) WithLength(length int) *ArrayAssertion {
	a.t.Helper()
	assert.Len(a.t, a.elements, length)
	return a
}

// Contains asserts array contains value
func (a *ArrayAssertion) Contains(value interface{}) *ArrayAssertion {
	a.t.Helper()
	expected := InterfaceToScriptValue(value)
	found := false
	for _, elem := range a.elements {
		if elem.Equals(expected) {
			found = true
			break
		}
	}
	assert.True(a.t, found, "Array should contain value: %v", value)
	return a
}

// AssertMethodResult provides a convenient way to assert method execution results
func AssertMethodResult(t *testing.T, result engine.ScriptValue, err error) *MethodResultAssertion {
	t.Helper()
	return &MethodResultAssertion{
		t:      t,
		result: result,
		err:    err,
	}
}

// MethodResultAssertion helps assert method execution results
type MethodResultAssertion struct {
	t      *testing.T
	result engine.ScriptValue
	err    error
}

// NoError asserts no error occurred
func (a *MethodResultAssertion) NoError() *MethodResultAssertion {
	a.t.Helper()
	require.NoError(a.t, a.err)
	return a
}

// HasError asserts an error occurred
func (a *MethodResultAssertion) HasError() *MethodResultAssertion {
	a.t.Helper()
	require.Error(a.t, a.err)
	return a
}

// ErrorContains asserts error contains text
func (a *MethodResultAssertion) ErrorContains(text string) *MethodResultAssertion {
	a.t.Helper()
	require.Error(a.t, a.err)
	assert.Contains(a.t, a.err.Error(), text)
	return a
}

// Result returns comprehensive assertion for the result value
func (a *MethodResultAssertion) Result() *ComprehensiveValueAssertion {
	a.t.Helper()
	require.NoError(a.t, a.err, "Cannot assert on result when error occurred")
	return NewComprehensiveAssertion(a.t, a.result)
}
