// ABOUTME: Tests for man page generation functionality.
// ABOUTME: Verifies correct formatting and content of generated man pages.

package docs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManPage(t *testing.T) {
	man := NewManPage("test", 1, "1.0.0")

	assert.Equal(t, "test", man.Name)
	assert.Equal(t, 1, man.Section)
	assert.Equal(t, "1.0.0", man.Version)
	assert.Equal(t, "TEST", man.Title)
	assert.NotEmpty(t, man.Date)
}

func TestManPageGenerate_Basic(t *testing.T) {
	man := NewManPage("test", 1, "1.0.0")
	man.Description = "test program"
	man.Synopsis = "[OPTIONS] FILE..."

	output := man.Generate()

	// Check header
	assert.Contains(t, output, ".TH TEST 1")
	assert.Contains(t, output, "1.0.0")

	// Check sections
	assert.Contains(t, output, ".SH NAME")
	assert.Contains(t, output, "test \\- test program")
	assert.Contains(t, output, ".SH SYNOPSIS")
	assert.Contains(t, output, ".B test")
	assert.Contains(t, output, "[OPTIONS] FILE...")
	assert.Contains(t, output, ".SH DESCRIPTION")
}

func TestManPageGenerate_Options(t *testing.T) {
	man := NewManPage("test", 1, "1.0.0")
	man.Description = "test program"

	man.Options = []Option{
		{Short: "h", Long: "help", Description: "Show help"},
		{Long: "config", Arg: "FILE", Description: "Config file", Default: "config.yaml"},
		{Short: "v", Description: "Verbose output"},
	}

	output := man.Generate()

	assert.Contains(t, output, ".SH OPTIONS")
	assert.Contains(t, output, "\\fB\\-h\\fR, \\fB\\-\\-help\\fR")
	assert.Contains(t, output, "Show help")
	assert.Contains(t, output, "\\fB\\-\\-config\\fR=\\fIFILE\\fR")
	assert.Contains(t, output, "(default: config.yaml)")
	assert.Contains(t, output, "\\fB\\-v\\fR")
}

func TestManPageGenerate_Commands(t *testing.T) {
	man := NewManPage("test", 1, "1.0.0")
	man.Description = "test program"

	man.Commands = []Command{
		{
			Name:        "run",
			Description: "Run a command",
			Options: []Option{
				{Short: "f", Long: "force", Description: "Force execution"},
			},
			Examples: []Example{
				{Command: "test run file.txt", Description: "Run on file"},
			},
		},
	}

	output := man.Generate()

	assert.Contains(t, output, ".SH COMMANDS")
	assert.Contains(t, output, ".SS run")
	assert.Contains(t, output, "Run a command")
	assert.Contains(t, output, "\\fB\\-f\\fR, \\fB\\-\\-force\\fR")
	assert.Contains(t, output, "test run file.txt")
}

func TestManPageGenerate_Examples(t *testing.T) {
	man := NewManPage("test", 1, "1.0.0")
	man.Description = "test program"

	man.Examples = []Example{
		{Command: "test foo", Description: "Basic usage"},
		{Command: "test bar --verbose", Description: "Verbose mode"},
	}

	output := man.Generate()

	assert.Contains(t, output, ".SH EXAMPLES")
	assert.Contains(t, output, "Basic usage")
	assert.Contains(t, output, "test foo")
	assert.Contains(t, output, "Verbose mode")
	assert.Contains(t, output, "test bar --verbose")
}

func TestManPageGenerate_AllSections(t *testing.T) {
	man := NewManPage("test", 1, "1.0.0")
	man.Description = "test program"
	man.Files = []string{"/etc/test.conf"}
	man.SeeAlso = []string{"bash(1)", "sh(1)"}
	man.Authors = []string{"Test Author"}
	man.Bugs = "Report bugs to test@example.com"

	output := man.Generate()

	assert.Contains(t, output, ".SH FILES")
	assert.Contains(t, output, ".I /etc/test.conf")
	assert.Contains(t, output, ".SH SEE ALSO")
	assert.Contains(t, output, ".BR bash(1)")
	assert.Contains(t, output, ".BR sh(1)")
	assert.Contains(t, output, ".SH AUTHORS")
	assert.Contains(t, output, "Test Author")
	assert.Contains(t, output, ".SH BUGS")
	assert.Contains(t, output, "Report bugs to test@example.com")
}

func TestFormatDescription(t *testing.T) {
	man := NewManPage("test", 1, "1.0.0")

	desc := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."
	man.Description = desc

	output := man.Generate()

	// Should have paragraph breaks
	assert.Contains(t, output, "First paragraph.\n.PP\nSecond paragraph.\n.PP\nThird paragraph.")
}

func TestManPage_GenerateLLMSpellManPage(t *testing.T) {
	man := GenerateLLMSpellManPage("1.0.0")

	assert.Equal(t, "llmspell", man.Name)
	assert.Equal(t, 1, man.Section)
	assert.Equal(t, "1.0.0", man.Version)
	assert.Contains(t, man.Description, "scriptable LLM interactions")

	// Check some commands exist
	commandNames := make([]string, len(man.Commands))
	for i, cmd := range man.Commands {
		commandNames[i] = cmd.Name
	}
	assert.Contains(t, commandNames, "run")
	assert.Contains(t, commandNames, "repl")
	assert.Contains(t, commandNames, "new")
	assert.Contains(t, commandNames, "validate")

	// Check output format
	output := man.Generate()
	assert.Contains(t, output, ".TH LLMSPELL 1")
	assert.Contains(t, output, ".SH COMMANDS")
	assert.Contains(t, output, ".SH EXAMPLES")
}

func TestManPage_GenerateCommandManPage(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"run command", "run", false},
		{"repl command", "repl", false},
		{"new command", "new", false},
		{"unknown command", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			man, err := GenerateCommandManPage(tt.command, "1.0.0")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, man)
			} else {
				require.NoError(t, err)
				require.NotNil(t, man)

				assert.Equal(t, "llmspell-"+tt.command, man.Name)
				assert.Contains(t, man.SeeAlso, "llmspell(1)")

				// Check output
				output := man.Generate()
				assert.Contains(t, output, ".TH LLMSPELL-"+strings.ToUpper(tt.command))
			}
		})
	}
}

func TestManPage_GetAllCommands(t *testing.T) {
	commands := GetAllCommands()

	assert.NotEmpty(t, commands)
	assert.Contains(t, commands, "run")
	assert.Contains(t, commands, "repl")
	assert.Contains(t, commands, "new")
	assert.Contains(t, commands, "validate")
	assert.Contains(t, commands, "completion")
}

func TestManPageSpecialCharacters(t *testing.T) {
	man := NewManPage("test", 1, "1.0.0")
	man.Description = "test program"

	// Test that hyphens in options are escaped
	man.Options = []Option{
		{Long: "foo-bar", Description: "Test option with hyphen"},
	}

	output := man.Generate()

	// Should escape the hyphens in the option name
	assert.Contains(t, output, "\\-\\-foo\\-bar")
}
