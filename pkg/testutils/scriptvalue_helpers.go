// ABOUTME: Test helper functions for ScriptValue conversion and validation
// ABOUTME: Provides utilities to simplify test assertions after ScriptValue refactoring

package testutils

import (
	"fmt"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/require"
)

// ExtractScriptValue extracts the underlying Go value from a ScriptValue result
// This helper is used to fix tests after ScriptValue refactoring
func ExtractScriptValue(t *testing.T, result any) any {
	t.Helper()

	if result == nil {
		return nil
	}

	sv, ok := result.(engine.ScriptValue)
	if !ok {
		// If it's not a ScriptValue, return as-is
		return result
	}

	// Handle special case for NilValue
	if sv.Type() == engine.TypeNil {
		return nil
	}

	return sv.ToGo()
}

// AssertScriptValueInterface asserts that a ScriptValue result equals expected value
func AssertScriptValueInterface(t *testing.T, expected any, result any) {
	t.Helper()

	actual := ExtractScriptValue(t, result)
	require.Equal(t, expected, actual)
}

// AssertScriptValueNil asserts that a ScriptValue result is nil
func AssertScriptValueNil(t *testing.T, result any) {
	t.Helper()

	if result == nil {
		return
	}

	sv, ok := result.(engine.ScriptValue)
	if !ok {
		t.Errorf("expected ScriptValue, got %T", result)
		return
	}

	if sv.Type() != engine.TypeNil {
		t.Errorf("expected nil ScriptValue, got %v", sv)
	}
}

// ExtractScriptValueMap extracts a map from a ScriptValue result
func ExtractScriptValueMap(t *testing.T, result any) map[string]any {
	t.Helper()

	sv, ok := result.(engine.ScriptValue)
	require.True(t, ok, "expected ScriptValue, got %T", result)

	m, ok := sv.ToGo().(map[string]any)
	require.True(t, ok, "expected map, got %T", sv.ToGo())

	return m
}

// ExtractScriptValueSlice extracts a slice from a ScriptValue result
func ExtractScriptValueSlice(t *testing.T, result any) []any {
	t.Helper()

	sv, ok := result.(engine.ScriptValue)
	require.True(t, ok, "expected ScriptValue, got %T", result)

	// Handle both array and table representations
	goVal := sv.ToGo()

	// Try as slice first
	if s, ok := goVal.([]any); ok {
		return s
	}

	// Try as map with numeric keys (Lua table)
	if m, ok := goVal.(map[string]any); ok {
		// Convert Lua 1-based table to slice
		size := len(m)
		slice := make([]any, size)
		for i := 1; i <= size; i++ {
			key := fmt.Sprintf("%d", i)
			if val, exists := m[key]; exists {
				slice[i-1] = val
			}
		}
		return slice
	}

	t.Fatalf("expected array or table, got %T", goVal)
	return nil
}
