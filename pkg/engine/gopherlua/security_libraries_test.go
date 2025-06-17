// ABOUTME: Tests for library restriction functionality in the SecurityManager
// ABOUTME: Validates safe library loading and dangerous function removal

package gopherlua

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestSafeLibraryLoader_LoadLibraries(t *testing.T) {
	tests := []struct {
		name      string
		libraries []string
		level     SecurityLevel
		validate  func(t *testing.T, L *lua.LState)
	}{
		{
			name:      "loads_base_library",
			libraries: []string{"base"},
			level:     SecurityLevelStandard,
			validate: func(t *testing.T, L *lua.LState) {
				// Check base functions exist
				assert.NotEqual(t, lua.LNil, L.GetGlobal("print"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("type"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("pairs"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("ipairs"))
			},
		},
		{
			name:      "loads_string_library",
			libraries: []string{"string"},
			level:     SecurityLevelStandard,
			validate: func(t *testing.T, L *lua.LState) {
				stringLib := L.GetGlobal("string")
				assert.NotEqual(t, lua.LNil, stringLib)
				if tbl, ok := stringLib.(*lua.LTable); ok {
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("upper"))
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("lower"))
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("find"))
				}
			},
		},
		{
			name:      "loads_table_library",
			libraries: []string{"table"},
			level:     SecurityLevelStandard,
			validate: func(t *testing.T, L *lua.LState) {
				tableLib := L.GetGlobal("table")
				assert.NotEqual(t, lua.LNil, tableLib)
				if tbl, ok := tableLib.(*lua.LTable); ok {
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("insert"))
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("remove"))
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("sort"))
				}
			},
		},
		{
			name:      "loads_math_library",
			libraries: []string{"math"},
			level:     SecurityLevelStandard,
			validate: func(t *testing.T, L *lua.LState) {
				mathLib := L.GetGlobal("math")
				assert.NotEqual(t, lua.LNil, mathLib)
				if tbl, ok := mathLib.(*lua.LTable); ok {
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("floor"))
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("ceil"))
					assert.NotEqual(t, lua.LNil, tbl.RawGetString("random"))
				}
			},
		},
		{
			name:      "skips_debug_library",
			libraries: []string{"debug"},
			level:     SecurityLevelStandard,
			validate: func(t *testing.T, L *lua.LState) {
				// Debug should never be loaded
				assert.Equal(t, lua.LNil, L.GetGlobal("debug"))
			},
		},
		{
			name:      "loads_multiple_libraries",
			libraries: []string{"base", "string", "table", "math"},
			level:     SecurityLevelStandard,
			validate: func(t *testing.T, L *lua.LState) {
				assert.NotEqual(t, lua.LNil, L.GetGlobal("print"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("string"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("table"))
				assert.NotEqual(t, lua.LNil, L.GetGlobal("math"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewSafeLibraryLoader(tt.level)

			opts := lua.Options{SkipOpenLibs: true}
			L := lua.NewState(opts)
			defer L.Close()

			err := loader.LoadLibraries(L, tt.libraries)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, L)
			}
		})
	}
}

func TestSafeLibraryLoader_RemoveDangerousFunctions(t *testing.T) {
	tests := []struct {
		name           string
		level          SecurityLevel
		checkFunctions map[string]bool // library.function -> should exist
	}{
		{
			name:  "standard_removes_dangerous_os_functions",
			level: SecurityLevelStandard,
			checkFunctions: map[string]bool{
				"os.execute": false,
				"os.exit":    false,
				"os.setenv":  false,
				"os.remove":  false,
				"os.rename":  false,
				"os.time":    true,
				"os.date":    true,
				"os.clock":   true,
			},
		},
		{
			name:  "minimal_keeps_most_os_functions",
			level: SecurityLevelMinimal,
			checkFunctions: map[string]bool{
				"os.execute": false, // Still dangerous
				"os.exit":    false,
				"os.setenv":  false,
				"os.remove":  true, // Allowed in minimal
				"os.rename":  true,
				"os.time":    true,
				"os.date":    true,
			},
		},
		{
			name:  "strict_removes_io_completely",
			level: SecurityLevelStrict,
			checkFunctions: map[string]bool{
				"io": false, // Entire library removed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewSafeLibraryLoader(tt.level)

			opts := lua.Options{SkipOpenLibs: true}
			L := lua.NewState(opts)
			defer L.Close()

			// Load OS library first
			if tt.level != SecurityLevelStrict {
				err := loader.LoadLibraries(L, []string{"os"})
				require.NoError(t, err)
			}

			// Apply security restrictions
			err := loader.RemoveDangerousFunctions(L)
			require.NoError(t, err)

			// Check functions
			for funcPath, shouldExist := range tt.checkFunctions {
				parts := splitFunctionPath(funcPath)
				if len(parts) == 1 {
					// Check entire library
					lib := L.GetGlobal(parts[0])
					if shouldExist {
						assert.NotEqual(t, lua.LNil, lib, "%s should exist", funcPath)
					} else {
						assert.Equal(t, lua.LNil, lib, "%s should not exist", funcPath)
					}
				} else if len(parts) == 2 {
					// Check specific function
					lib := L.GetGlobal(parts[0])
					if lib == lua.LNil {
						if shouldExist {
							t.Errorf("%s library should exist for function %s", parts[0], funcPath)
						}
						continue
					}

					if tbl, ok := lib.(*lua.LTable); ok {
						fn := tbl.RawGetString(parts[1])
						if shouldExist {
							assert.NotEqual(t, lua.LNil, fn, "%s should exist", funcPath)
						} else {
							assert.Equal(t, lua.LNil, fn, "%s should not exist", funcPath)
						}
					}
				}
			}
		})
	}
}

func TestSafeLibraryLoader_CustomReplacements(t *testing.T) {
	tests := []struct {
		name        string
		level       SecurityLevel
		testCode    string
		expectError bool
		validate    func(t *testing.T, L *lua.LState)
	}{
		{
			name:  "safe_print_replacement",
			level: SecurityLevelStrict,
			testCode: `
				_G._test_output = {}
				print("hello", "world")
				return table.concat(_G._test_output, " ")
			`,
			validate: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1)
				if str, ok := result.(lua.LString); ok {
					assert.Contains(t, string(str), "hello\tworld")
				}
			},
		},
		{
			name:  "safe_require_blocked",
			level: SecurityLevelStrict,
			testCode: `
				local ok, err = pcall(require, "dangerous")
				return ok
			`,
			validate: func(t *testing.T, L *lua.LState) {
				result := L.Get(-1)
				assert.Equal(t, lua.LFalse, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewSafeLibraryLoader(tt.level)

			opts := lua.Options{SkipOpenLibs: true}
			L := lua.NewState(opts)
			defer L.Close()

			// Load base library with replacements
			err := loader.LoadLibraries(L, []string{"base", "table"})
			require.NoError(t, err)

			// Apply custom replacements
			err = loader.ApplyCustomReplacements(L)
			require.NoError(t, err)

			// Run test code
			err = L.DoString(tt.testCode)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, L)
				}
			}
		})
	}
}

// Helper function to split function paths
func splitFunctionPath(path string) []string {
	parts := []string{}
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
