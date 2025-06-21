// ABOUTME: LLMSpell-specific shell completion setup using Kong CLI metadata.
// ABOUTME: Extracts command structure from Kong to generate accurate completions.

package shell

import (
	"fmt"
	"reflect"
	"strings"
)

// GenerateFromKong generates shell completion from Kong CLI structure
func GenerateFromKong(app interface{}, programName string) (*CompletionGenerator, error) {
	gen := NewCompletionGenerator(programName)

	// Use reflection to extract commands and flags from Kong app
	appType := reflect.TypeOf(app)
	if appType.Kind() == reflect.Ptr {
		appType = appType.Elem()
	}

	// Extract global flags from all fields (not just embedded)
	for i := 0; i < appType.NumField(); i++ {
		field := appType.Field(i)
		// Skip commands
		if field.Tag.Get("cmd") != "" {
			continue
		}
		// Extract flag from field
		if flag := extractFlag(field); flag != nil {
			gen.AddGlobalFlag(*flag)
		}
	}

	// Extract commands
	for i := 0; i < appType.NumField(); i++ {
		field := appType.Field(i)
		kongTag := field.Tag.Get("cmd")

		// Skip if not a command
		if kongTag == "" {
			continue
		}

		cmd := extractCommand(field)
		if cmd != nil {
			gen.AddCommand(*cmd)
		}
	}

	return gen, nil
}

// extractCommand extracts a command from a struct field
func extractCommand(field reflect.StructField) *Command {
	cmd := &Command{
		Name:        toKebabCase(field.Name),
		Description: getHelp(field.Tag),
		Flags:       []Flag{},
		Subcommands: []Command{},
	}

	// Extract flags from the command struct
	cmdType := field.Type
	if cmdType.Kind() == reflect.Ptr {
		cmdType = cmdType.Elem()
	}

	// Check for embedded BaseCommand or similar
	for i := 0; i < cmdType.NumField(); i++ {
		subField := cmdType.Field(i)

		// Skip embedded structs (like BaseCommand)
		if subField.Anonymous {
			continue
		}

		// Check if it's a subcommand
		if subField.Tag.Get("cmd") != "" {
			if subcmd := extractCommand(subField); subcmd != nil {
				cmd.Subcommands = append(cmd.Subcommands, *subcmd)
			}
		} else {
			// It's a flag
			if flag := extractFlag(subField); flag != nil {
				cmd.Flags = append(cmd.Flags, *flag)
			}
		}
	}

	return cmd
}

// extractFlag extracts a flag from a struct field
func extractFlag(field reflect.StructField) *Flag {
	// Skip if it's a command
	if field.Tag.Get("cmd") != "" {
		return nil
	}

	// Skip if it's a positional arg
	if field.Tag.Get("arg") != "" {
		return nil
	}

	// Skip unexported fields
	if !field.IsExported() {
		return nil
	}

	flag := &Flag{
		Long:        toKebabCase(field.Name),
		Description: getHelp(field.Tag),
		HasValue:    field.Type.Kind() != reflect.Bool,
	}

	// Extract short flag
	if short := getShort(field.Tag); short != "" {
		flag.Short = short
	}

	// Extract enum values
	if enum := getEnum(field.Tag); enum != "" {
		flag.Values = strings.Split(enum, ",")
		for i := range flag.Values {
			flag.Values[i] = strings.TrimSpace(flag.Values[i])
		}
	}

	// Handle special types
	switch field.Type.Kind() {
	case reflect.Slice:
		flag.HasValue = true
	case reflect.Map:
		flag.HasValue = true
	}

	return flag
}

// Helper functions for tag parsing

func getHelp(tag reflect.StructTag) string {
	if help := tag.Get("help"); help != "" {
		return help
	}
	return ""
}

func getShort(tag reflect.StructTag) string {
	if short := tag.Get("short"); short != "" {
		return short
	}
	return ""
}

func getEnum(tag reflect.StructTag) string {
	if enum := tag.Get("enum"); enum != "" {
		return enum
	}
	return ""
}

// toKebabCase converts PascalCase to kebab-case
func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteByte('-')
		}
		result.WriteRune(rune(strings.ToLower(string(r))[0]))
	}
	return result.String()
}

// LLMSpellCommands defines the llmspell command structure for completion
func LLMSpellCommands() *CompletionGenerator {
	gen := NewCompletionGenerator("llmspell")

	// Global flags
	gen.AddGlobalFlag(Flag{Long: "debug", Short: "d", Description: "Enable debug mode"})
	gen.AddGlobalFlag(Flag{Long: "config", Short: "c", Description: "Config file path", HasValue: true})
	gen.AddGlobalFlag(Flag{Long: "quiet", Short: "q", Description: "Suppress non-essential output"})
	gen.AddGlobalFlag(Flag{Long: "verbose", Short: "v", Description: "Show detailed information"})
	gen.AddGlobalFlag(Flag{Long: "profile", Short: "p", Description: "Security profile", HasValue: true, Values: []string{"sandbox", "development", "production"}})
	gen.AddGlobalFlag(Flag{Long: "help", Short: "h", Description: "Show help"})

	// run command
	gen.AddCommand(Command{
		Name:        "run",
		Description: "Execute a spell script",
		Flags: []Flag{
			{Long: "engine", Short: "e", Description: "Script engine", HasValue: true, Values: []string{"lua", "javascript", "tengo"}},
			{Long: "param", Short: "p", Description: "Parameters (key=value)", HasValue: true},
			{Long: "timeout", Short: "t", Description: "Execution timeout", HasValue: true},
			{Long: "env", Description: "Environment variables", HasValue: true},
			{Long: "dry-run", Description: "Show execution plan"},
			{Long: "watch", Short: "w", Description: "Watch for changes"},
			{Long: "progress", Description: "Show progress"},
		},
	})

	// repl command
	gen.AddCommand(Command{
		Name:        "repl",
		Description: "Interactive REPL",
		Flags: []Flag{
			{Long: "engine", Short: "e", Description: "Script engine", HasValue: true, Values: []string{"lua", "javascript", "tengo"}},
			{Long: "no-history", Description: "Disable history"},
			{Long: "no-highlight", Description: "Disable syntax highlighting"},
			{Long: "history-file", Description: "History file path", HasValue: true},
		},
	})

	// new command
	gen.AddCommand(Command{
		Name:        "new",
		Description: "Create spell from template",
		Flags: []Flag{
			{Long: "type", Short: "t", Description: "Template type", HasValue: true, Values: []string{"basic", "advanced", "agent", "workflow", "interactive"}},
			{Long: "engine", Short: "e", Description: "Script engine", HasValue: true, Values: []string{"lua", "javascript", "tengo"}},
			{Long: "author", Description: "Author name", HasValue: true},
			{Long: "description", Description: "Spell description", HasValue: true},
			{Long: "license", Description: "License type", HasValue: true},
			{Long: "force", Short: "f", Description: "Overwrite existing"},
			{Long: "list", Short: "l", Description: "List templates"},
		},
	})

	// validate command
	gen.AddCommand(Command{
		Name:        "validate",
		Description: "Validate scripts",
		Flags: []Flag{
			{Long: "engine", Short: "e", Description: "Force engine", HasValue: true, Values: []string{"lua", "javascript", "tengo"}},
			{Long: "strict", Description: "Strict validation"},
			{Long: "security", Description: "Include security check"},
		},
	})

	// config command with subcommands
	gen.AddCommand(Command{
		Name:        "config",
		Description: "Configuration management",
		Subcommands: []Command{
			{Name: "view", Description: "View configuration"},
			{Name: "get", Description: "Get config value"},
			{Name: "set", Description: "Set config value"},
			{Name: "reset", Description: "Reset to default"},
			{Name: "init", Description: "Initialize config"},
			{Name: "validate", Description: "Validate config"},
			{Name: "list", Description: "List all keys"},
			{Name: "export", Description: "Export config"},
			{Name: "import", Description: "Import config"},
		},
		Flags: []Flag{
			{Long: "format", Short: "f", Description: "Output format", HasValue: true, Values: []string{"json", "yaml", "text"}},
		},
	})

	// security command with subcommands
	gen.AddCommand(Command{
		Name:        "security",
		Description: "Security profiles",
		Subcommands: []Command{
			{Name: "list", Description: "List profiles"},
			{Name: "view", Description: "View profile"},
			{Name: "validate", Description: "Validate profile"},
			{Name: "check", Description: "Check permission"},
			{Name: "compare", Description: "Compare profiles"},
			{Name: "export", Description: "Export profile"},
		},
		Flags: []Flag{
			{Long: "verbose", Short: "v", Description: "Verbose output"},
		},
	})

	// engines command with subcommands
	gen.AddCommand(Command{
		Name:        "engines",
		Description: "Engine information",
		Subcommands: []Command{
			{Name: "list", Description: "List engines"},
			{Name: "info", Description: "Engine details"},
			{Name: "capabilities", Description: "Engine features"},
			{Name: "detect", Description: "Detect engine"},
			{Name: "check", Description: "Health check"},
			{Name: "benchmark", Description: "Performance test"},
		},
		Flags: []Flag{
			{Long: "format", Description: "Output format", HasValue: true, Values: []string{"text", "json", "yaml"}},
			{Long: "verbose", Short: "v", Description: "Detailed info"},
		},
	})

	// debug command
	gen.AddCommand(Command{
		Name:        "debug",
		Description: "Interactive debugger",
		Flags: []Flag{
			{Long: "breakpoint", Short: "b", Description: "Set breakpoint", HasValue: true},
			{Long: "watch", Short: "w", Description: "Watch expression", HasValue: true},
		},
	})

	// version command
	gen.AddCommand(Command{
		Name:        "version",
		Description: "Version information",
		Flags: []Flag{
			{Long: "short", Short: "s", Description: "Short version"},
			{Long: "build-info", Description: "Build details"},
			{Long: "format", Description: "Output format", HasValue: true, Values: []string{"text", "json"}},
			{Long: "deps", Description: "Show dependencies"},
			{Long: "check-compat", Description: "Check compatibility"},
		},
	})

	// completion command (meta!)
	gen.AddCommand(Command{
		Name:        "completion",
		Description: "Generate shell completion",
		Flags: []Flag{
			{Long: "shell", Short: "s", Description: "Target shell", HasValue: true, Values: []string{"bash", "zsh", "fish", "powershell", "sh"}},
		},
	})

	// man command
	gen.AddCommand(Command{
		Name:        "man",
		Description: "Generate man pages",
		Flags: []Flag{
			{Long: "output", Short: "o", Description: "Output file", HasValue: true},
			{Long: "dir", Short: "d", Description: "Output directory", HasValue: true},
			{Long: "all", Short: "a", Description: "Generate all man pages"},
			{Long: "install", Short: "i", Description: "Install to system"},
			{Long: "format", Description: "Output format", HasValue: true, Values: []string{"troff", "text", "html"}},
		},
	})

	return gen
}

// InstallInstructions returns installation instructions for each shell
func InstallInstructions(shell Shell) string {
	switch shell {
	case Bash:
		return `# Add to ~/.bashrc or ~/.bash_profile:
source <(llmspell completion bash)

# Or save to file:
llmspell completion bash > ~/.local/share/bash-completion/completions/llmspell`

	case Zsh:
		return `# Add to ~/.zshrc:
source <(llmspell completion zsh)

# Or add to fpath (before compinit):
llmspell completion zsh > ~/.zsh/completions/_llmspell`

	case Fish:
		return `# Save to fish completions:
llmspell completion fish > ~/.config/fish/completions/llmspell.fish`

	case PowerShell:
		return `# Add to PowerShell profile:
llmspell completion powershell | Out-String | Invoke-Expression

# To find profile location:
echo $PROFILE`

	case Sh:
		return `# POSIX sh doesn't support programmable completion.
# View the command reference:
llmspell completion sh`

	default:
		return fmt.Sprintf("# Unknown shell: %s", shell)
	}
}
