// ABOUTME: Tests for spell loading functionality, covering spell.yaml parsing and validation.
// ABOUTME: Ensures proper loading of spell metadata, dependencies, and configuration.

package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpellMetadata(t *testing.T) {
	t.Run("valid_metadata", func(t *testing.T) {
		meta := &SpellMetadata{
			Name:         "test-spell",
			Version:      "1.0.0",
			Description:  "A test spell",
			Author:       "Test Author",
			Engine:       "lua",
			EntryPoint:   "main.lua",
			Dependencies: []string{"http", "json"},
			Parameters: []SpellParameter{
				{
					Name:        "input",
					Type:        "string",
					Description: "Input data",
					Required:    true,
				},
				{
					Name:        "verbose",
					Type:        "boolean",
					Description: "Enable verbose output",
					Required:    false,
					Default:     false,
				},
			},
			SecurityProfile: "sandbox",
			Tags:            []string{"utility", "testing"},
		}

		assert.Equal(t, "test-spell", meta.Name)
		assert.Equal(t, "1.0.0", meta.Version)
		assert.Equal(t, "lua", meta.Engine)
		assert.Len(t, meta.Parameters, 2)
		assert.True(t, meta.Parameters[0].Required)
		assert.False(t, meta.Parameters[1].Required)
	})

	t.Run("validate_metadata", func(t *testing.T) {
		tests := []struct {
			name    string
			meta    *SpellMetadata
			wantErr bool
			errMsg  string
		}{
			{
				name: "valid_metadata",
				meta: &SpellMetadata{
					Name:       "valid-spell",
					Version:    "1.0.0",
					Engine:     "lua",
					EntryPoint: "main.lua",
				},
				wantErr: false,
			},
			{
				name: "missing_name",
				meta: &SpellMetadata{
					Version:    "1.0.0",
					Engine:     "lua",
					EntryPoint: "main.lua",
				},
				wantErr: true,
				errMsg:  "spell name is required",
			},
			{
				name: "missing_version",
				meta: &SpellMetadata{
					Name:       "test",
					Engine:     "lua",
					EntryPoint: "main.lua",
				},
				wantErr: true,
				errMsg:  "spell version is required",
			},
			{
				name: "missing_engine",
				meta: &SpellMetadata{
					Name:       "test",
					Version:    "1.0.0",
					EntryPoint: "main.lua",
				},
				wantErr: true,
				errMsg:  "spell engine is required",
			},
			{
				name: "missing_entrypoint",
				meta: &SpellMetadata{
					Name:    "test",
					Version: "1.0.0",
					Engine:  "lua",
				},
				wantErr: true,
				errMsg:  "spell entry point is required",
			},
			{
				name: "invalid_parameter_type",
				meta: &SpellMetadata{
					Name:       "test",
					Version:    "1.0.0",
					Engine:     "lua",
					EntryPoint: "main.lua",
					Parameters: []SpellParameter{
						{
							Name: "param",
							Type: "invalid",
						},
					},
				},
				wantErr: true,
				errMsg:  "invalid parameter type",
			},
			{
				name: "duplicate_parameter_names",
				meta: &SpellMetadata{
					Name:       "test",
					Version:    "1.0.0",
					Engine:     "lua",
					EntryPoint: "main.lua",
					Parameters: []SpellParameter{
						{Name: "param", Type: "string"},
						{Name: "param", Type: "number"},
					},
				},
				wantErr: true,
				errMsg:  "duplicate parameter name",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.meta.Validate()
				if tt.wantErr {
					assert.Error(t, err)
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestSpellLoader_LoadFromFile(t *testing.T) {
	loader := NewSpellLoader()

	t.Run("load_valid_yaml", func(t *testing.T) {
		// Create a temporary spell.yaml file
		tmpDir := t.TempDir()
		spellFile := filepath.Join(tmpDir, "spell.yaml")

		content := `
name: test-spell
version: 1.0.0
description: A test spell for unit testing
author: Test Author
engine: lua
entry_point: main.lua

parameters:
  - name: message
    type: string
    description: Message to display
    required: true
  - name: count
    type: number
    description: Number of times to repeat
    default: 1

dependencies:
  - http
  - json

security_profile: sandbox

tags:
  - test
  - example

metadata:
  license: MIT
  repository: https://github.com/example/test-spell
`
		err := os.WriteFile(spellFile, []byte(content), 0644)
		require.NoError(t, err)

		// Load the spell
		spell, err := loader.LoadFromFile(spellFile)
		require.NoError(t, err)
		assert.NotNil(t, spell)

		// Verify loaded data
		assert.Equal(t, "test-spell", spell.Name)
		assert.Equal(t, "1.0.0", spell.Version)
		assert.Equal(t, "lua", spell.Engine)
		assert.Equal(t, "main.lua", spell.EntryPoint)
		assert.Len(t, spell.Parameters, 2)
		assert.Equal(t, "message", spell.Parameters[0].Name)
		assert.Equal(t, "string", spell.Parameters[0].Type)
		assert.True(t, spell.Parameters[0].Required)
		assert.Equal(t, 1, spell.Parameters[1].Default)
		assert.Contains(t, spell.Dependencies, "http")
		assert.Contains(t, spell.Dependencies, "json")
		assert.Equal(t, "sandbox", spell.SecurityProfile)
		assert.Contains(t, spell.Tags, "test")
		assert.Equal(t, "MIT", spell.Metadata["license"])
	})

	t.Run("load_minimal_yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		spellFile := filepath.Join(tmpDir, "spell.yaml")

		content := `
name: minimal-spell
version: 0.1.0
engine: lua
entry_point: main.lua
`
		err := os.WriteFile(spellFile, []byte(content), 0644)
		require.NoError(t, err)

		spell, err := loader.LoadFromFile(spellFile)
		require.NoError(t, err)
		assert.NotNil(t, spell)
		assert.Equal(t, "minimal-spell", spell.Name)
		assert.Empty(t, spell.Parameters)
		assert.Empty(t, spell.Dependencies)
	})

	t.Run("load_nonexistent_file", func(t *testing.T) {
		spell, err := loader.LoadFromFile("/nonexistent/spell.yaml")
		assert.Error(t, err)
		assert.Nil(t, spell)
		assert.Contains(t, err.Error(), "failed to read spell file")
	})

	t.Run("load_invalid_yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		spellFile := filepath.Join(tmpDir, "spell.yaml")

		content := `
invalid yaml content
  - this is not valid
    yaml: syntax
`
		err := os.WriteFile(spellFile, []byte(content), 0644)
		require.NoError(t, err)

		spell, err := loader.LoadFromFile(spellFile)
		assert.Error(t, err)
		assert.Nil(t, spell)
		assert.Contains(t, err.Error(), "failed to parse")
	})

	t.Run("load_invalid_metadata", func(t *testing.T) {
		tmpDir := t.TempDir()
		spellFile := filepath.Join(tmpDir, "spell.yaml")

		content := `
name: ""
version: 1.0.0
engine: lua
entry_point: main.lua
`
		err := os.WriteFile(spellFile, []byte(content), 0644)
		require.NoError(t, err)

		spell, err := loader.LoadFromFile(spellFile)
		assert.Error(t, err)
		assert.Nil(t, spell)
		assert.Contains(t, err.Error(), "spell name is required")
	})
}

func TestSpellLoader_LoadFromDirectory(t *testing.T) {
	loader := NewSpellLoader()

	t.Run("load_from_directory", func(t *testing.T) {
		// Create a spell directory structure
		tmpDir := t.TempDir()
		spellDir := filepath.Join(tmpDir, "my-spell")
		err := os.MkdirAll(spellDir, 0755)
		require.NoError(t, err)

		// Create spell.yaml
		spellFile := filepath.Join(spellDir, "spell.yaml")
		content := `
name: dir-spell
version: 1.0.0
engine: lua
entry_point: main.lua
`
		err = os.WriteFile(spellFile, []byte(content), 0644)
		require.NoError(t, err)

		// Create main.lua
		mainFile := filepath.Join(spellDir, "main.lua")
		err = os.WriteFile(mainFile, []byte("print('Hello from spell')"), 0644)
		require.NoError(t, err)

		// Load the spell
		spell, err := loader.LoadFromDirectory(spellDir)
		require.NoError(t, err)
		assert.NotNil(t, spell)
		assert.Equal(t, "dir-spell", spell.Name)
		assert.Equal(t, spellDir, spell.RootDir)
	})

	t.Run("load_from_directory_no_spell_yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		spell, err := loader.LoadFromDirectory(tmpDir)
		assert.Error(t, err)
		assert.Nil(t, spell)
		assert.Contains(t, err.Error(), "spell.yaml not found")
	})

	t.Run("load_from_nonexistent_directory", func(t *testing.T) {
		spell, err := loader.LoadFromDirectory("/nonexistent/directory")
		assert.Error(t, err)
		assert.Nil(t, spell)
	})
}

func TestSpellLoader_ResolveEntryPoint(t *testing.T) {
	loader := NewSpellLoader()

	t.Run("resolve_absolute_path", func(t *testing.T) {
		spell := &SpellMetadata{
			RootDir:    "/home/user/spells/my-spell",
			EntryPoint: "/absolute/path/to/script.lua",
		}

		path := loader.ResolveEntryPoint(spell)
		assert.Equal(t, "/absolute/path/to/script.lua", path)
	})

	t.Run("resolve_relative_path", func(t *testing.T) {
		spell := &SpellMetadata{
			RootDir:    "/home/user/spells/my-spell",
			EntryPoint: "src/main.lua",
		}

		path := loader.ResolveEntryPoint(spell)
		assert.Equal(t, "/home/user/spells/my-spell/src/main.lua", path)
	})

	t.Run("resolve_with_empty_root", func(t *testing.T) {
		spell := &SpellMetadata{
			RootDir:    "",
			EntryPoint: "main.lua",
		}

		path := loader.ResolveEntryPoint(spell)
		assert.Equal(t, "main.lua", path)
	})
}

func TestSpellParameter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		param   SpellParameter
		wantErr bool
	}{
		{
			name: "valid_string_param",
			param: SpellParameter{
				Name:        "message",
				Type:        "string",
				Description: "A message",
			},
			wantErr: false,
		},
		{
			name: "valid_number_param",
			param: SpellParameter{
				Name:    "count",
				Type:    "number",
				Default: 42,
			},
			wantErr: false,
		},
		{
			name: "valid_boolean_param",
			param: SpellParameter{
				Name:    "verbose",
				Type:    "boolean",
				Default: true,
			},
			wantErr: false,
		},
		{
			name: "valid_array_param",
			param: SpellParameter{
				Name: "items",
				Type: "array",
			},
			wantErr: false,
		},
		{
			name: "valid_object_param",
			param: SpellParameter{
				Name: "config",
				Type: "object",
			},
			wantErr: false,
		},
		{
			name: "empty_name",
			param: SpellParameter{
				Name: "",
				Type: "string",
			},
			wantErr: true,
		},
		{
			name: "empty_type",
			param: SpellParameter{
				Name: "param",
				Type: "",
			},
			wantErr: true,
		},
		{
			name: "invalid_type",
			param: SpellParameter{
				Name: "param",
				Type: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParameter(tt.param)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkSpellLoader_LoadFromFile(b *testing.B) {
	loader := NewSpellLoader()

	// Create a test file
	tmpDir := b.TempDir()
	spellFile := filepath.Join(tmpDir, "spell.yaml")
	content := `
name: bench-spell
version: 1.0.0
engine: lua
entry_point: main.lua
parameters:
  - name: param1
    type: string
  - name: param2
    type: number
`
	err := os.WriteFile(spellFile, []byte(content), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.LoadFromFile(spellFile)
	}
}

func BenchmarkSpellMetadata_Validate(b *testing.B) {
	meta := &SpellMetadata{
		Name:       "bench-spell",
		Version:    "1.0.0",
		Engine:     "lua",
		EntryPoint: "main.lua",
		Parameters: []SpellParameter{
			{Name: "param1", Type: "string"},
			{Name: "param2", Type: "number"},
			{Name: "param3", Type: "boolean"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = meta.Validate()
	}
}
