// ABOUTME: Tests for complex type conversion handlers - map, slice, struct, interface{} converters
// ABOUTME: Validates nested structures, struct tags, field mapping, and complex type handling

package gopherlua

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestComplexConverter_MapConversions(t *testing.T) {
	converter := NewComplexConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		validate    func(t *testing.T, result lua.LValue)
	}{
		{
			name:  "string_map_to_table",
			input: map[string]interface{}{"name": "test", "value": 42, "active": true},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, "test", table.RawGetString("name").String())
				assert.Equal(t, float64(42), float64(table.RawGetString("value").(lua.LNumber)))
				assert.Equal(t, true, bool(table.RawGetString("active").(lua.LBool)))
			},
		},
		{
			name:  "int_key_map_to_table",
			input: map[int]string{1: "first", 2: "second", 3: "third"},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, "first", table.RawGetString("1").String())
				assert.Equal(t, "second", table.RawGetString("2").String())
				assert.Equal(t, "third", table.RawGetString("3").String())
			},
		},
		{
			name:  "nested_map_to_table",
			input: map[string]interface{}{"outer": map[string]interface{}{"inner": "value"}},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				outerTable := table.RawGetString("outer").(*lua.LTable)
				assert.Equal(t, "value", outerTable.RawGetString("inner").String())
			},
		},
		{
			name:  "empty_map_to_table",
			input: map[string]interface{}{},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, 0, table.Len())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.MapToLua(L, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, lua.LTTable, result.Type())
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestComplexConverter_SliceConversions(t *testing.T) {
	converter := NewComplexConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		validate    func(t *testing.T, result lua.LValue)
	}{
		{
			name:  "string_slice_to_table",
			input: []string{"first", "second", "third"},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, 3, table.Len())
				assert.Equal(t, "first", table.RawGetInt(1).String())
				assert.Equal(t, "second", table.RawGetInt(2).String())
				assert.Equal(t, "third", table.RawGetInt(3).String())
			},
		},
		{
			name:  "int_slice_to_table",
			input: []int{10, 20, 30},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, 3, table.Len())
				assert.Equal(t, float64(10), float64(table.RawGetInt(1).(lua.LNumber)))
				assert.Equal(t, float64(20), float64(table.RawGetInt(2).(lua.LNumber)))
				assert.Equal(t, float64(30), float64(table.RawGetInt(3).(lua.LNumber)))
			},
		},
		{
			name:  "interface_slice_to_table",
			input: []interface{}{"string", 42, true, nil},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				// In Lua, nil values in tables are handled specially
				// The table length might be affected by trailing nils
				assert.GreaterOrEqual(t, table.Len(), 3) // At least 3 elements
				assert.Equal(t, "string", table.RawGetInt(1).String())
				assert.Equal(t, float64(42), float64(table.RawGetInt(2).(lua.LNumber)))
				assert.Equal(t, true, bool(table.RawGetInt(3).(lua.LBool)))
				assert.Equal(t, lua.LNil, table.RawGetInt(4))
			},
		},
		{
			name:  "nested_slice_to_table",
			input: [][]string{{"a", "b"}, {"c", "d"}},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, 2, table.Len())
				innerTable1 := table.RawGetInt(1).(*lua.LTable)
				innerTable2 := table.RawGetInt(2).(*lua.LTable)
				assert.Equal(t, "a", innerTable1.RawGetInt(1).String())
				assert.Equal(t, "b", innerTable1.RawGetInt(2).String())
				assert.Equal(t, "c", innerTable2.RawGetInt(1).String())
				assert.Equal(t, "d", innerTable2.RawGetInt(2).String())
			},
		},
		{
			name:  "empty_slice_to_table",
			input: []string{},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, 0, table.Len())
			},
		},
		{
			name:  "array_to_table",
			input: [3]int{1, 2, 3},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, 3, table.Len())
				assert.Equal(t, float64(1), float64(table.RawGetInt(1).(lua.LNumber)))
				assert.Equal(t, float64(2), float64(table.RawGetInt(2).(lua.LNumber)))
				assert.Equal(t, float64(3), float64(table.RawGetInt(3).(lua.LNumber)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.SliceToLua(L, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, lua.LTTable, result.Type())
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestComplexConverter_StructConversions(t *testing.T) {
	converter := NewComplexConverter()
	L := lua.NewState()
	defer L.Close()

	type SimpleStruct struct {
		Name   string
		Value  int
		Active bool
	}

	type StructWithTags struct {
		Name     string `lua:"title"`
		Value    int    `lua:"amount"`
		Hidden   string `lua:"-"`
		Internal string `lua:"internal,omitempty"`
	}

	type NestedStruct struct {
		ID     int
		Simple SimpleStruct
		Data   map[string]interface{}
	}

	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		validate    func(t *testing.T, result lua.LValue)
	}{
		{
			name:  "simple_struct_to_table",
			input: SimpleStruct{Name: "test", Value: 42, Active: true},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, "test", table.RawGetString("Name").String())
				assert.Equal(t, float64(42), float64(table.RawGetString("Value").(lua.LNumber)))
				assert.Equal(t, true, bool(table.RawGetString("Active").(lua.LBool)))
			},
		},
		{
			name:  "struct_with_tags_to_table",
			input: StructWithTags{Name: "test", Value: 42, Hidden: "secret", Internal: "internal"},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, "test", table.RawGetString("title").String())
				assert.Equal(t, float64(42), float64(table.RawGetString("amount").(lua.LNumber)))
				assert.Equal(t, lua.LNil, table.RawGetString("Hidden")) // Should be hidden
				assert.Equal(t, "internal", table.RawGetString("internal").String())
			},
		},
		{
			name: "nested_struct_to_table",
			input: NestedStruct{
				ID:     1,
				Simple: SimpleStruct{Name: "nested", Value: 100, Active: false},
				Data:   map[string]interface{}{"key": "value"},
			},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, float64(1), float64(table.RawGetString("ID").(lua.LNumber)))

				simpleTable := table.RawGetString("Simple").(*lua.LTable)
				assert.Equal(t, "nested", simpleTable.RawGetString("Name").String())
				assert.Equal(t, float64(100), float64(simpleTable.RawGetString("Value").(lua.LNumber)))
				assert.Equal(t, false, bool(simpleTable.RawGetString("Active").(lua.LBool)))

				dataTable := table.RawGetString("Data").(*lua.LTable)
				assert.Equal(t, "value", dataTable.RawGetString("key").String())
			},
		},
		{
			name:  "struct_with_omitempty_nil",
			input: StructWithTags{Name: "test", Value: 42, Internal: ""},
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, "test", table.RawGetString("title").String())
				assert.Equal(t, float64(42), float64(table.RawGetString("amount").(lua.LNumber)))
				assert.Equal(t, lua.LNil, table.RawGetString("internal")) // Should be omitted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.StructToLua(L, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, lua.LTTable, result.Type())
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestComplexConverter_InterfaceHandling(t *testing.T) {
	converter := NewComplexConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		validate    func(t *testing.T, result lua.LValue)
	}{
		{
			name:  "interface_with_string",
			input: interface{}("hello"),
			validate: func(t *testing.T, result lua.LValue) {
				assert.Equal(t, "hello", result.String())
			},
		},
		{
			name:  "interface_with_map",
			input: interface{}(map[string]interface{}{"key": "value"}),
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, "value", table.RawGetString("key").String())
			},
		},
		{
			name:  "interface_with_slice",
			input: interface{}([]string{"a", "b", "c"}),
			validate: func(t *testing.T, result lua.LValue) {
				table := result.(*lua.LTable)
				assert.Equal(t, 3, table.Len())
				assert.Equal(t, "a", table.RawGetInt(1).String())
			},
		},
		{
			name:  "nil_interface",
			input: interface{}(nil),
			validate: func(t *testing.T, result lua.LValue) {
				assert.Equal(t, lua.LNil, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.InterfaceToLua(L, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestComplexConverter_FromLuaConversions(t *testing.T) {
	converter := NewComplexConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name        string
		setupTable  func() lua.LValue
		expectError bool
		validate    func(t *testing.T, result interface{})
	}{
		{
			name: "table_to_map",
			setupTable: func() lua.LValue {
				table := L.NewTable()
				table.RawSetString("name", lua.LString("test"))
				table.RawSetString("value", lua.LNumber(42))
				return table
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test", m["name"])
				assert.Equal(t, float64(42), m["value"])
			},
		},
		{
			name: "table_to_slice",
			setupTable: func() lua.LValue {
				table := L.NewTable()
				table.RawSetInt(1, lua.LString("first"))
				table.RawSetInt(2, lua.LString("second"))
				table.RawSetInt(3, lua.LString("third"))
				return table
			},
			validate: func(t *testing.T, result interface{}) {
				slice, ok := result.([]interface{})
				require.True(t, ok)
				assert.Len(t, slice, 3)
				assert.Equal(t, "first", slice[0])
				assert.Equal(t, "second", slice[1])
				assert.Equal(t, "third", slice[2])
			},
		},
		{
			name: "mixed_table_to_map",
			setupTable: func() lua.LValue {
				table := L.NewTable()
				table.RawSetInt(1, lua.LString("indexed"))
				table.RawSetString("key", lua.LString("named"))
				return table
			},
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "indexed", m["1"])
				assert.Equal(t, "named", m["key"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			luaValue := tt.setupTable()
			result, err := converter.FromLua(luaValue)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestComplexConverter_StructTagParsing(t *testing.T) {
	converter := NewComplexConverter()

	type TaggedStruct struct {
		Name     string `lua:"title"`
		Value    int    `lua:"amount,omitempty"`
		Hidden   string `lua:"-"`
		Default  string
		Required string `lua:",required"`
	}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected StructTagInfo
	}{
		{
			name:  "simple_tag",
			field: reflect.TypeOf(TaggedStruct{}).Field(0), // Name
			expected: StructTagInfo{
				Name:      "title",
				Omitempty: false,
				Skip:      false,
				Required:  false,
			},
		},
		{
			name:  "omitempty_tag",
			field: reflect.TypeOf(TaggedStruct{}).Field(1), // Value
			expected: StructTagInfo{
				Name:      "amount",
				Omitempty: true,
				Skip:      false,
				Required:  false,
			},
		},
		{
			name:  "skip_tag",
			field: reflect.TypeOf(TaggedStruct{}).Field(2), // Hidden
			expected: StructTagInfo{
				Name:      "",
				Omitempty: false,
				Skip:      true,
				Required:  false,
			},
		},
		{
			name:  "default_tag",
			field: reflect.TypeOf(TaggedStruct{}).Field(3), // Default
			expected: StructTagInfo{
				Name:      "Default",
				Omitempty: false,
				Skip:      false,
				Required:  false,
			},
		},
		{
			name:  "required_tag",
			field: reflect.TypeOf(TaggedStruct{}).Field(4), // Required
			expected: StructTagInfo{
				Name:      "Required",
				Omitempty: false,
				Skip:      false,
				Required:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ParseStructTag(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComplexConverter_ErrorHandling(t *testing.T) {
	converter := NewComplexConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name      string
		operation func() error
		errorText string
	}{
		{
			name: "unsupported_map_key_type",
			operation: func() error {
				_, err := converter.MapToLua(L, map[chan int]string{make(chan int): "test"})
				return err
			},
			errorText: "unsupported map key type",
		},
		{
			name: "unsupported_slice_element_type",
			operation: func() error {
				_, err := converter.SliceToLua(L, []chan int{make(chan int)})
				return err
			},
			errorText: "unsupported type",
		},
		{
			name: "non_struct_to_struct_conversion",
			operation: func() error {
				_, err := converter.StructToLua(L, "not a struct")
				return err
			},
			errorText: "expected struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorText)
		})
	}
}

func TestComplexConverter_CircularReferenceHandling(t *testing.T) {
	converter := NewComplexConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("circular_map_reference", func(t *testing.T) {
		parent := make(map[string]interface{})
		child := make(map[string]interface{})
		parent["child"] = child
		child["parent"] = parent

		_, err := converter.MapToLua(L, parent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular reference")
	})

	t.Run("circular_slice_reference", func(t *testing.T) {
		parent := make(map[string]interface{})
		slice := []interface{}{parent}
		parent["slice"] = slice

		_, err := converter.MapToLua(L, parent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular reference")
	})
}

func TestComplexConverter_PerformanceAndEdgeCases(t *testing.T) {
	converter := NewComplexConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("large_map_conversion", func(t *testing.T) {
		largeMap := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeMap[string(rune('a'+i%26))+string(rune('a'+(i/26)%26))] = i
		}

		result, err := converter.MapToLua(L, largeMap)
		require.NoError(t, err)
		assert.Equal(t, lua.LTTable, result.Type())
	})

	t.Run("large_slice_conversion", func(t *testing.T) {
		largeSlice := make([]int, 1000)
		for i := range largeSlice {
			largeSlice[i] = i
		}

		result, err := converter.SliceToLua(L, largeSlice)
		require.NoError(t, err)
		table := result.(*lua.LTable)
		assert.Equal(t, 1000, table.Len())
	})

	t.Run("deeply_nested_structure", func(t *testing.T) {
		data := make(map[string]interface{})
		current := data
		for i := 0; i < 10; i++ {
			next := make(map[string]interface{})
			current["next"] = next
			current = next
		}
		current["value"] = "deep"

		result, err := converter.MapToLua(L, data)
		require.NoError(t, err)
		assert.Equal(t, lua.LTTable, result.Type())
	})
}
