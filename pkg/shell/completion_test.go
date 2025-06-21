// ABOUTME: Tests for shell completion generation functionality.
// ABOUTME: Verifies completion script generation for all supported shells.

package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseShell(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Shell
		wantErr bool
	}{
		{
			name:  "bash lowercase",
			input: "bash",
			want:  Bash,
		},
		{
			name:  "bash uppercase",
			input: "BASH",
			want:  Bash,
		},
		{
			name:  "zsh",
			input: "zsh",
			want:  Zsh,
		},
		{
			name:  "fish",
			input: "fish",
			want:  Fish,
		},
		{
			name:  "powershell",
			input: "powershell",
			want:  PowerShell,
		},
		{
			name:  "pwsh alias",
			input: "pwsh",
			want:  PowerShell,
		},
		{
			name:  "sh",
			input: "sh",
			want:  Sh,
		},
		{
			name:    "unknown shell",
			input:   "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseShell(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetSupportedShells(t *testing.T) {
	shells := GetSupportedShells()
	assert.Len(t, shells, 5)
	assert.Contains(t, shells, Bash)
	assert.Contains(t, shells, Zsh)
	assert.Contains(t, shells, Fish)
	assert.Contains(t, shells, PowerShell)
	assert.Contains(t, shells, Sh)
}

func TestCompletionGenerator(t *testing.T) {
	// Create a test generator with sample commands
	gen := NewCompletionGenerator("llmspell")

	// Add global flags
	gen.AddGlobalFlag(Flag{
		Long:        "debug",
		Short:       "d",
		Description: "Enable debug mode",
		HasValue:    false,
	})
	gen.AddGlobalFlag(Flag{
		Long:        "config",
		Short:       "c",
		Description: "Config file path",
		HasValue:    true,
	})

	// Add run command
	gen.AddCommand(Command{
		Name:        "run",
		Description: "Execute a spell script",
		Flags: []Flag{
			{
				Long:        "engine",
				Short:       "e",
				Description: "Script engine",
				HasValue:    true,
				Values:      []string{"lua", "javascript", "tengo"},
			},
			{
				Long:        "timeout",
				Short:       "t",
				Description: "Execution timeout",
				HasValue:    true,
			},
		},
	})

	// Add config command with subcommands
	gen.AddCommand(Command{
		Name:        "config",
		Description: "Configuration management",
		Subcommands: []Command{
			{
				Name:        "view",
				Description: "View configuration",
			},
			{
				Name:        "set",
				Description: "Set configuration value",
			},
		},
		Flags: []Flag{
			{
				Long:        "format",
				Short:       "f",
				Description: "Output format",
				HasValue:    true,
				Values:      []string{"json", "yaml", "text"},
			},
		},
	})

	t.Run("bash completion", func(t *testing.T) {
		script, err := gen.Generate(Bash)
		require.NoError(t, err)
		assert.NotEmpty(t, script)

		// Check for expected content
		assert.Contains(t, script, "_llmspell_complete()")
		assert.Contains(t, script, "commands=(\"run\" \"config\"")
		assert.Contains(t, script, "global_flags=(--debug -d --config -c")
		assert.Contains(t, script, "_llmspell_run_complete()")
		assert.Contains(t, script, "_llmspell_config_complete()")
		assert.Contains(t, script, "complete -F _llmspell_complete llmspell")
	})

	t.Run("zsh completion", func(t *testing.T) {
		script, err := gen.Generate(Zsh)
		require.NoError(t, err)
		assert.NotEmpty(t, script)

		// Check for expected content
		assert.Contains(t, script, "#compdef llmspell")
		assert.Contains(t, script, "_llmspell()")
		assert.Contains(t, script, "'run:Execute a spell script'")
		assert.Contains(t, script, "'config:Configuration management'")
		assert.Contains(t, script, "_llmspell_run()")
		assert.Contains(t, script, "_llmspell_config()")
		assert.Contains(t, script, "_arguments")
	})

	t.Run("fish completion", func(t *testing.T) {
		script, err := gen.Generate(Fish)
		require.NoError(t, err)
		assert.NotEmpty(t, script)

		// Check for expected content
		assert.Contains(t, script, "complete -c llmspell")
		assert.Contains(t, script, "-n '__fish_use_subcommand' -a run")
		assert.Contains(t, script, "-n '__fish_use_subcommand' -a config")
		assert.Contains(t, script, "-l debug -s d")
		assert.Contains(t, script, "__fish_seen_subcommand_from")
	})

	t.Run("powershell completion", func(t *testing.T) {
		script, err := gen.Generate(PowerShell)
		require.NoError(t, err)
		assert.NotEmpty(t, script)

		// Check for expected content
		assert.Contains(t, script, "Register-ArgumentCompleter -Native -CommandName llmspell")
		assert.Contains(t, script, "$commands = @('run', 'config')")
		assert.Contains(t, script, "[System.Management.Automation.CompletionResult]")
		assert.Contains(t, script, "switch ($command)")
	})

	t.Run("sh reference", func(t *testing.T) {
		script, err := gen.Generate(Sh)
		require.NoError(t, err)
		assert.NotEmpty(t, script)

		// Check for expected content
		assert.Contains(t, script, "llmspell - LLM spell runner")
		assert.Contains(t, script, "COMMANDS:")
		assert.Contains(t, script, "run - Execute a spell script")
		assert.Contains(t, script, "config - Configuration management")
		assert.Contains(t, script, "GLOBAL FLAGS:")
	})

	t.Run("unknown shell", func(t *testing.T) {
		_, err := gen.Generate("unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported shell")
	})
}

func TestCompletionGeneratorEmpty(t *testing.T) {
	// Test with empty generator
	gen := NewCompletionGenerator("test")

	shells := []Shell{Bash, Zsh, Fish, PowerShell, Sh}
	for _, shell := range shells {
		t.Run(string(shell), func(t *testing.T) {
			script, err := gen.Generate(shell)
			require.NoError(t, err)
			assert.NotEmpty(t, script)
			assert.Contains(t, script, "test")
		})
	}
}

func TestEnumFlagCompletion(t *testing.T) {
	gen := NewCompletionGenerator("test")
	gen.AddCommand(Command{
		Name: "cmd",
		Flags: []Flag{
			{
				Long:     "type",
				HasValue: true,
				Values:   []string{"basic", "advanced", "expert"},
			},
		},
	})

	t.Run("bash enum values", func(t *testing.T) {
		script, err := gen.Generate(Bash)
		require.NoError(t, err)
		// Bash doesn't directly support enum completion in our basic template
		assert.Contains(t, script, "--type")
	})

	t.Run("zsh enum values", func(t *testing.T) {
		script, err := gen.Generate(Zsh)
		require.NoError(t, err)
		assert.Contains(t, script, "(basic advanced expert)")
	})

	t.Run("fish enum values", func(t *testing.T) {
		script, err := gen.Generate(Fish)
		require.NoError(t, err)
		assert.Contains(t, script, "-xa 'basic advanced expert'")
	})
}

func TestComplexCommandHierarchy(t *testing.T) {
	gen := NewCompletionGenerator("complex")

	// Add a complex command with nested subcommands
	gen.AddCommand(Command{
		Name:        "cluster",
		Description: "Cluster management",
		Subcommands: []Command{
			{
				Name:        "create",
				Description: "Create cluster",
				Flags: []Flag{
					{Long: "name", HasValue: true},
					{Long: "size", HasValue: true},
				},
			},
			{
				Name:        "delete",
				Description: "Delete cluster",
				Flags: []Flag{
					{Long: "force", Short: "f"},
				},
			},
			{
				Name:        "list",
				Description: "List clusters",
				Flags: []Flag{
					{Long: "format", HasValue: true, Values: []string{"json", "table"}},
				},
			},
		},
	})

	// Test that all shells can handle complex hierarchies
	shells := []Shell{Bash, Zsh, Fish, PowerShell}
	for _, shell := range shells {
		t.Run(string(shell), func(t *testing.T) {
			script, err := gen.Generate(shell)
			require.NoError(t, err)
			assert.NotEmpty(t, script)

			// Check that subcommands are present
			assert.Contains(t, script, "create")
			assert.Contains(t, script, "delete")
			assert.Contains(t, script, "list")
		})
	}
}

func TestSpecialCharacterEscaping(t *testing.T) {
	gen := NewCompletionGenerator("test")
	gen.AddCommand(Command{
		Name:        "test",
		Description: "Test with 'quotes' and \"double quotes\"",
		Flags: []Flag{
			{
				Long:        "message",
				Description: "A flag with $special <characters>",
			},
		},
	})

	// Generate for all shells and ensure no syntax errors
	shells := []Shell{Bash, Zsh, Fish, PowerShell}
	for _, shell := range shells {
		t.Run(string(shell), func(t *testing.T) {
			script, err := gen.Generate(shell)
			require.NoError(t, err)
			assert.NotEmpty(t, script)

			// Basic check that script was generated
			assert.Contains(t, script, "test")
		})
	}
}

func TestFileCompletionCommands(t *testing.T) {
	gen := NewCompletionGenerator("llmspell")

	// Commands that should have file completion
	fileCommands := []string{"run", "validate", "debug"}
	for _, cmd := range fileCommands {
		gen.AddCommand(Command{
			Name:        cmd,
			Description: "Command that accepts files",
		})
	}

	t.Run("bash file completion", func(t *testing.T) {
		script, err := gen.Generate(Bash)
		require.NoError(t, err)
		assert.Contains(t, script, "_filedir")
	})

	t.Run("fish file completion", func(t *testing.T) {
		script, err := gen.Generate(Fish)
		require.NoError(t, err)
		assert.Contains(t, script, "-n '__fish_seen_subcommand_from run validate debug' -F")
	})
}

func TestShortFlagGeneration(t *testing.T) {
	gen := NewCompletionGenerator("test")

	// Test flag with both long and short forms
	gen.AddGlobalFlag(Flag{
		Long:  "verbose",
		Short: "v",
	})

	// Test flag with only long form
	gen.AddGlobalFlag(Flag{
		Long: "quiet",
	})

	scripts := map[Shell][]string{
		Bash: {"--verbose", "-v", "--quiet"},
		Zsh:  {"-v[--verbose]", "--quiet"},
		Fish: {"-l verbose -s v", "-l quiet"},
	}

	for shell, expected := range scripts {
		t.Run(string(shell), func(t *testing.T) {
			script, err := gen.Generate(shell)
			require.NoError(t, err)
			for _, exp := range expected {
				assert.Contains(t, script, exp)
			}
		})
	}
}
