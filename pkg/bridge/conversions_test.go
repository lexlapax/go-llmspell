// ABOUTME: Tests for type conversion utilities
// ABOUTME: Validates safe conversions between Go and script types

package bridge

import (
	"reflect"
	"testing"
)

// Test struct for conversion tests
type TestStruct struct {
	Name     string                 `json:"name"`
	Age      int                    `json:"age"`
	Active   bool                   `json:"active"`
	Score    float64                `json:"score"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
	private  string                 // Should be ignored
}

type NestedStruct struct {
	ID     string     `json:"id"`
	Person TestStruct `json:"person"`
}

func TestBaseConverter_ToScript(t *testing.T) {
	converter := &BaseConverter{}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "int",
			input:    42,
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "float",
			input:    3.14,
			expected: 3.14,
			wantErr:  false,
		},
		{
			name:     "bool",
			input:    true,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "slice",
			input:    []int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "map",
			input:    map[string]int{"a": 1, "b": 2},
			expected: map[string]interface{}{"a": 1, "b": 2},
			wantErr:  false,
		},
		{
			name: "struct",
			input: TestStruct{
				Name:   "John",
				Age:    30,
				Active: true,
				Score:  95.5,
				Tags:   []string{"go", "test"},
				Metadata: map[string]interface{}{
					"level": "advanced",
				},
				private: "ignored",
			},
			expected: map[string]interface{}{
				"name":   "John",
				"age":    30,
				"active": true,
				"score":  95.5,
				"tags":   []interface{}{"go", "test"},
				"metadata": map[string]interface{}{
					"level": "advanced",
				},
			},
			wantErr: false,
		},
		{
			name: "pointer to struct",
			input: &TestStruct{
				Name: "Jane",
				Age:  25,
			},
			expected: map[string]interface{}{
				"name":     "Jane",
				"age":      25,
				"active":   false,
				"score":    float64(0),
				"tags":     []interface{}(nil),
				"metadata": map[string]interface{}(nil),
			},
			wantErr: false,
		},
		{
			name:     "nil pointer",
			input:    (*TestStruct)(nil),
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.ToScript(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !deepEqualIgnoreMapOrder(got, tt.expected) {
				t.Errorf("ToScript() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseConverter_FromScript(t *testing.T) {
	converter := &BaseConverter{}

	tests := []struct {
		name       string
		value      interface{}
		targetType reflect.Type
		expected   interface{}
		wantErr    bool
	}{
		{
			name:       "nil to string",
			value:      nil,
			targetType: reflect.TypeOf(""),
			expected:   "",
			wantErr:    false,
		},
		{
			name:       "string to string",
			value:      "hello",
			targetType: reflect.TypeOf(""),
			expected:   "hello",
			wantErr:    false,
		},
		{
			name:       "int to string",
			value:      42,
			targetType: reflect.TypeOf(""),
			expected:   "42",
			wantErr:    false,
		},
		{
			name:       "string to bool",
			value:      "true",
			targetType: reflect.TypeOf(false),
			expected:   true,
			wantErr:    false,
		},
		{
			name:       "int to bool",
			value:      1,
			targetType: reflect.TypeOf(false),
			expected:   true,
			wantErr:    false,
		},
		{
			name:       "zero to bool",
			value:      0,
			targetType: reflect.TypeOf(false),
			expected:   false,
			wantErr:    false,
		},
		{
			name:       "float to int",
			value:      42.7,
			targetType: reflect.TypeOf(0),
			expected:   42,
			wantErr:    false,
		},
		{
			name:       "string to int",
			value:      "123",
			targetType: reflect.TypeOf(0),
			expected:   123,
			wantErr:    false,
		},
		{
			name:       "slice conversion",
			value:      []interface{}{"a", "b", "c"},
			targetType: reflect.TypeOf([]string{}),
			expected:   []string{"a", "b", "c"},
			wantErr:    false,
		},
		{
			name: "map conversion",
			value: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			targetType: reflect.TypeOf(map[string]string{}),
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "map to struct",
			value: map[string]interface{}{
				"name":   "John",
				"age":    30,
				"active": true,
				"score":  95.5,
				"tags":   []interface{}{"go", "test"},
				"metadata": map[string]interface{}{
					"level": "advanced",
				},
			},
			targetType: reflect.TypeOf(TestStruct{}),
			expected: TestStruct{
				Name:   "John",
				Age:    30,
				Active: true,
				Score:  95.5,
				Tags:   []string{"go", "test"},
				Metadata: map[string]interface{}{
					"level": "advanced",
				},
			},
			wantErr: false,
		},
		{
			name:       "interface{} type",
			value:      "anything",
			targetType: reflect.TypeOf((*interface{})(nil)).Elem(),
			expected:   "anything",
			wantErr:    false,
		},
		{
			name:       "pointer type",
			value:      "hello",
			targetType: reflect.TypeOf((*string)(nil)),
			expected:   func() interface{} { s := "hello"; return &s }(),
			wantErr:    false,
		},
		{
			name:       "invalid conversion",
			value:      "not a number",
			targetType: reflect.TypeOf(0),
			expected:   0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.FromScript(tt.value, tt.targetType)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("FromScript() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseConverter_ComplexConversions(t *testing.T) {
	converter := &BaseConverter{}

	t.Run("nested struct to script", func(t *testing.T) {
		input := NestedStruct{
			ID: "123",
			Person: TestStruct{
				Name:   "Alice",
				Age:    28,
				Active: true,
				Tags:   []string{"nested", "test"},
			},
		}

		result, err := converter.ToScript(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := map[string]interface{}{
			"id": "123",
			"person": map[string]interface{}{
				"name":     "Alice",
				"age":      28,
				"active":   true,
				"score":    float64(0),
				"tags":     []interface{}{"nested", "test"},
				"metadata": map[string]interface{}(nil),
			},
		}

		if !deepEqualIgnoreMapOrder(result, expected) {
			t.Errorf("got %v, want %v", result, expected)
		}
	})

	t.Run("script to nested struct", func(t *testing.T) {
		input := map[string]interface{}{
			"id": "456",
			"person": map[string]interface{}{
				"name":   "Bob",
				"age":    35,
				"active": false,
				"tags":   []interface{}{"from", "script"},
			},
		}

		result, err := converter.FromScript(input, reflect.TypeOf(NestedStruct{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := NestedStruct{
			ID: "456",
			Person: TestStruct{
				Name:   "Bob",
				Age:    35,
				Active: false,
				Tags:   []string{"from", "script"},
			},
		}

		if !deepEqualIgnoreMapOrder(result, expected) {
			t.Errorf("got %v, want %v", result, expected)
		}
	})

	t.Run("slice of structs", func(t *testing.T) {
		input := []TestStruct{
			{Name: "First", Age: 20},
			{Name: "Second", Age: 30},
		}

		// To script
		scriptVal, err := converter.ToScript(input)
		if err != nil {
			t.Fatalf("ToScript error: %v", err)
		}

		// From script back
		result, err := converter.FromScript(scriptVal, reflect.TypeOf([]TestStruct{}))
		if err != nil {
			t.Fatalf("FromScript error: %v", err)
		}

		// For struct slices, we need custom comparison
		resultSlice := result.([]TestStruct)
		if len(resultSlice) != len(input) {
			t.Errorf("round trip failed: different lengths got %d, want %d", len(resultSlice), len(input))
		}
		for i := range input {
			if resultSlice[i].Name != input[i].Name || resultSlice[i].Age != input[i].Age {
				t.Errorf("round trip failed at index %d: got %v, want %v", i, resultSlice[i], input[i])
			}
		}
	})
}

func TestBaseConverter_EdgeCases(t *testing.T) {
	converter := &BaseConverter{}

	t.Run("empty string to number", func(t *testing.T) {
		_, err := converter.FromScript("", reflect.TypeOf(0))
		if err == nil {
			t.Error("expected error for empty string to number conversion")
		}
	})

	t.Run("overflow handling", func(t *testing.T) {
		// Large number to int8
		result, err := converter.FromScript(256.0, reflect.TypeOf(int8(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Go will truncate, not error
		if result.(int8) != 0 {
			t.Errorf("expected overflow truncation, got %v", result)
		}
	})

	t.Run("json tag with options", func(t *testing.T) {
		type TaggedStruct struct {
			Field1 string `json:"field,omitempty"`
			Field2 int    `json:"-"`
			Field3 bool   `json:"active"`
		}

		input := TaggedStruct{
			Field1: "value",
			Field2: 42, // Should be ignored
			Field3: true,
		}

		result, err := converter.ToScript(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := result.(map[string]interface{})
		if m["field"] != "value" {
			t.Errorf("expected field1 to be mapped to 'field'")
		}
		if _, exists := m["Field2"]; exists {
			t.Error("Field2 should be ignored due to json:\"-\" tag")
		}
		if m["active"] != true {
			t.Errorf("expected field3 to be mapped to 'active'")
		}
	})
}

// deepEqualIgnoreMapOrder compares values, ignoring map ordering
func deepEqualIgnoreMapOrder(a, b interface{}) bool {
	// For maps, convert to JSON and back to normalize ordering
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	switch va.Kind() {
	case reflect.Map:
		if va.Len() != vb.Len() {
			return false
		}

		// Check all keys exist and values match
		for _, key := range va.MapKeys() {
			aVal := va.MapIndex(key)
			bVal := vb.MapIndex(key)
			if !bVal.IsValid() {
				return false
			}
			if !deepEqualIgnoreMapOrder(aVal.Interface(), bVal.Interface()) {
				return false
			}
		}
		return true

	case reflect.Slice, reflect.Array:
		if va.Len() != vb.Len() {
			return false
		}
		for i := 0; i < va.Len(); i++ {
			if !deepEqualIgnoreMapOrder(va.Index(i).Interface(), vb.Index(i).Interface()) {
				return false
			}
		}
		return true

	default:
		// For other types, use regular DeepEqual
		return reflect.DeepEqual(a, b)
	}
}
