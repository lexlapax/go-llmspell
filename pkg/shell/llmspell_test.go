// ABOUTME: Tests for LLMSpell-specific shell completion functionality.
// ABOUTME: Verifies Kong integration and llmspell command completion generation.

package shell

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"RunCommand", "run-command"},
		{"REPLMode", "r-e-p-l-mode"},
		{"NewSpell", "new-spell"},
		{"Version", "version"},
		{"ConfigFile", "config-file"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toKebabCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractFlag(t *testing.T) {
	// This would require actual struct definitions to test properly
	// For now, we'll test the helper functions

	t.Run("getHelp", func(t *testing.T) {
		type TestStruct struct {
			Field string `help:"This is help text"`
		}
		field, _ := reflect.TypeOf(TestStruct{}).FieldByName("Field")
		help := getHelp(field.Tag)
		assert.Equal(t, "This is help text", help)
	})

	t.Run("getShort", func(t *testing.T) {
		type TestStruct struct {
			Field string `short:"f"`
		}
		field, _ := reflect.TypeOf(TestStruct{}).FieldByName("Field")
		short := getShort(field.Tag)
		assert.Equal(t, "f", short)
	})

	t.Run("getEnum", func(t *testing.T) {
		type TestStruct struct {
			Field string `enum:"one,two,three"`
		}
		field, _ := reflect.TypeOf(TestStruct{}).FieldByName("Field")
		enum := getEnum(field.Tag)
		assert.Equal(t, "one,two,three", enum)
	})
}

func TestLLMSpellCommands(t *testing.T) {
	gen := LLMSpellCommands()
	assert.NotNil(t, gen)

	// Test that all major commands are present
	commandNames := []string{
		"run", "repl", "new", "validate", "config",
		"security", "engines", "debug", "version", "completion",
	}

	// Generate bash completion and check for commands
	script, err := gen.Generate(Bash)
	require.NoError(t, err)

	for _, cmd := range commandNames {
		assert.Contains(t, script, cmd, "Command %s should be in completion", cmd)
	}

	// Check for global flags
	globalFlags := []string{"--debug", "--config", "--quiet", "--verbose", "--profile"}
	for _, flag := range globalFlags {
		assert.Contains(t, script, flag, "Global flag %s should be in completion", flag)
	}
}

func TestLLMSpellSubcommands(t *testing.T) {
	gen := LLMSpellCommands()

	// Test config subcommands
	t.Run("config subcommands", func(t *testing.T) {
		script, err := gen.Generate(Zsh)
		require.NoError(t, err)

		configSubcmds := []string{"view", "get", "set", "reset", "init", "validate", "list", "export", "import"}
		for _, subcmd := range configSubcmds {
			assert.Contains(t, script, subcmd)
		}
	})

	// Test security subcommands
	t.Run("security subcommands", func(t *testing.T) {
		script, err := gen.Generate(Fish)
		require.NoError(t, err)

		securitySubcmds := []string{"list", "view", "validate", "check", "compare", "export"}
		for _, subcmd := range securitySubcmds {
			assert.Contains(t, script, subcmd)
		}
	})
}

func TestLLMSpellEnumValues(t *testing.T) {
	gen := LLMSpellCommands()

	// Test engine enum values
	t.Run("engine values", func(t *testing.T) {
		script, err := gen.Generate(Zsh)
		require.NoError(t, err)
		assert.Contains(t, script, "lua")
		assert.Contains(t, script, "javascript")
		assert.Contains(t, script, "tengo")
	})

	// Test profile enum values
	t.Run("profile values", func(t *testing.T) {
		script, err := gen.Generate(Fish)
		require.NoError(t, err)
		assert.Contains(t, script, "sandbox")
		assert.Contains(t, script, "development")
		assert.Contains(t, script, "production")
	})

	// Test template type values
	t.Run("template types", func(t *testing.T) {
		script, err := gen.Generate(Zsh)
		require.NoError(t, err)
		assert.Contains(t, script, "basic")
		assert.Contains(t, script, "advanced")
		assert.Contains(t, script, "agent")
		assert.Contains(t, script, "workflow")
		assert.Contains(t, script, "interactive")
	})
}

func TestInstallInstructions(t *testing.T) {
	tests := []struct {
		shell    Shell
		contains []string
	}{
		{
			shell:    Bash,
			contains: []string{"~/.bashrc", "source <(llmspell completion bash)"},
		},
		{
			shell:    Zsh,
			contains: []string{"~/.zshrc", "source <(llmspell completion zsh)", "fpath"},
		},
		{
			shell:    Fish,
			contains: []string{"~/.config/fish/completions/llmspell.fish"},
		},
		{
			shell:    PowerShell,
			contains: []string{"$PROFILE", "Invoke-Expression"},
		},
		{
			shell:    Sh,
			contains: []string{"POSIX sh doesn't support"},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.shell), func(t *testing.T) {
			instructions := InstallInstructions(tt.shell)
			for _, expected := range tt.contains {
				assert.Contains(t, instructions, expected)
			}
		})
	}
}

func TestGenerateFromKong(t *testing.T) {
	// Test with a simple Kong app structure
	type TestApp struct {
		Debug  bool   `short:"d" help:"Enable debug"`
		Config string `short:"c" help:"Config file"`

		Run struct {
			Engine  string `short:"e" help:"Script engine" enum:"lua,js"`
			Timeout string `short:"t" help:"Timeout"`
		} `cmd:"" help:"Run a script"`

		Version struct {
			Short bool `short:"s" help:"Short version"`
		} `cmd:"" help:"Show version"`
	}

	app := TestApp{}
	gen, err := GenerateFromKong(app, "test")
	require.NoError(t, err)
	assert.NotNil(t, gen)

	// Generate completion and verify structure
	script, err := gen.Generate(Bash)
	require.NoError(t, err)

	// Check for commands
	assert.Contains(t, script, "run")
	assert.Contains(t, script, "version")

	// Check for global flags
	assert.Contains(t, script, "--debug")
	assert.Contains(t, script, "--config")
}

func TestFileCompletionForSpecificCommands(t *testing.T) {
	gen := LLMSpellCommands()

	// Commands that should trigger file completion
	// fileCommands := []string{"run", "validate", "debug"}

	t.Run("bash file completion", func(t *testing.T) {
		script, err := gen.Generate(Bash)
		require.NoError(t, err)

		// Check that run command uses _filedir
		assert.Contains(t, script, "_llmspell_run_complete")
		assert.Contains(t, script, "_filedir")
	})

	t.Run("fish file completion directive", func(t *testing.T) {
		script, err := gen.Generate(Fish)
		require.NoError(t, err)

		// Check for file completion enablement
		assert.Contains(t, script, "__fish_seen_subcommand_from run validate debug")
		assert.Contains(t, script, "-F")
	})
}

func TestCompletionCommandItself(t *testing.T) {
	gen := LLMSpellCommands()

	// The completion command should be able to complete shell names
	script, err := gen.Generate(Bash)
	require.NoError(t, err)

	assert.Contains(t, script, "completion")

	// Generate zsh to check enum values for shell flag
	zshScript, err := gen.Generate(Zsh)
	require.NoError(t, err)
	assert.Contains(t, zshScript, "bash")
	assert.Contains(t, zshScript, "zsh")
	assert.Contains(t, zshScript, "fish")
	assert.Contains(t, zshScript, "powershell")
}

func TestRealWorldCompletion(t *testing.T) {
	gen := LLMSpellCommands()

	// Test a realistic command line scenario
	t.Run("complex run command", func(t *testing.T) {
		// User types: llmspell run --engine lua --timeout 30s --param key=value script.lua
		script, err := gen.Generate(Bash)
		require.NoError(t, err)

		// Should have all the flags
		assert.Contains(t, script, "--engine")
		assert.Contains(t, script, "--timeout")
		assert.Contains(t, script, "--param")
	})

	t.Run("config subcommand completion", func(t *testing.T) {
		// User types: llmspell config set <TAB>
		script, err := gen.Generate(Zsh)
		require.NoError(t, err)

		// Should show config subcommands
		assert.Contains(t, script, "_llmspell_config")
		assert.Contains(t, script, "set:Set config value")
	})
}
