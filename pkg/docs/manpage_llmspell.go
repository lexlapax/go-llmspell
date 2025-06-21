// ABOUTME: LLMSpell-specific man page definitions and generators.
// ABOUTME: Provides man page content for the main llmspell command and all subcommands.

package docs

import (
	"fmt"
)

// GenerateLLMSpellManPage generates the main llmspell man page
func GenerateLLMSpellManPage(version string) *ManPage {
	man := NewManPage("llmspell", 1, version)
	man.Description = "scriptable LLM interactions via Lua, JavaScript, and Tengo"
	man.Synopsis = "[\\fIGLOBAL-OPTIONS\\fR] \\fICOMMAND\\fR [\\fICOMMAND-OPTIONS\\fR] [\\fIARGS\\fR]..."

	// Global options
	man.Options = []Option{
		{Short: "h", Long: "help", Description: "Show context-sensitive help"},
		{Long: "debug", Description: "Enable debug mode (also via $LLMSPELL_DEBUG)"},
		{Long: "config", Arg: "PATH", Description: "Config file path (also via $LLMSPELL_CONFIG)"},
		{Short: "q", Long: "quiet", Description: "Suppress non-error output"},
		{Short: "v", Long: "verbose", Description: "Enable verbose output"},
		{Long: "profile", Arg: "NAME", Description: "Security profile to use", Default: "sandbox"},
	}

	// Commands
	man.Commands = []Command{
		{
			Name:        "run",
			Description: "Execute a spell script with optional parameters",
			Options: []Option{
				{Short: "p", Long: "parameters", Arg: "KEY=VALUE", Description: "Parameters to pass to the script (can be repeated)"},
				{Short: "e", Long: "engine", Arg: "NAME", Description: "Script engine to use (auto-detected if not specified)"},
				{Short: "t", Long: "timeout", Arg: "SECONDS", Description: "Execution timeout", Default: "300"},
			},
			Examples: []Example{
				{Command: "llmspell run hello.lua", Description: "Run a simple Lua script"},
				{Command: "llmspell run process.lua -p input=data.txt -p count=10", Description: "Run with parameters"},
			},
		},
		{
			Name:        "repl",
			Description: "Start an interactive Read-Eval-Print Loop",
			Options: []Option{
				{Short: "e", Long: "engine", Arg: "NAME", Description: "Script engine to use", Default: "lua"},
				{Long: "no-history", Description: "Disable command history"},
				{Long: "no-highlight", Description: "Disable syntax highlighting"},
				{Long: "history-file", Arg: "PATH", Description: "Custom history file location"},
			},
		},
		{
			Name:        "new",
			Description: "Create a new spell from a template",
			Options: []Option{
				{Short: "t", Long: "type", Arg: "TYPE", Description: "Template type (basic, advanced, agent, workflow, interactive)", Default: "basic"},
				{Short: "e", Long: "engine", Arg: "NAME", Description: "Script engine", Default: "lua"},
				{Short: "a", Long: "author", Arg: "NAME", Description: "Author name"},
				{Short: "d", Long: "description", Arg: "TEXT", Description: "Spell description"},
				{Short: "f", Long: "force", Description: "Overwrite existing directory"},
				{Long: "list", Description: "List available templates"},
			},
		},
		{
			Name:        "validate",
			Description: "Validate a spell or script file",
			Options: []Option{
				{Short: "e", Long: "engine", Arg: "NAME", Description: "Force specific engine for validation"},
				{Long: "strict", Description: "Enable strict validation mode"},
				{Long: "security", Description: "Include security analysis"},
			},
		},
		{
			Name:        "config",
			Description: "Manage configuration settings",
			Examples: []Example{
				{Command: "llmspell config show", Description: "Display current configuration"},
				{Command: "llmspell config get engine.default", Description: "Get specific value"},
				{Command: "llmspell config set engine.default javascript", Description: "Set configuration value"},
			},
		},
		{
			Name:        "security",
			Description: "Manage security profiles",
			Examples: []Example{
				{Command: "llmspell security list", Description: "List available profiles"},
				{Command: "llmspell security show sandbox", Description: "Show profile details"},
			},
		},
		{
			Name:        "engines",
			Description: "Display information about available script engines",
			Options: []Option{
				{Short: "d", Long: "details", Description: "Show detailed engine information"},
			},
		},
		{
			Name:        "debug",
			Description: "Debug a spell script with breakpoints and step execution",
			Options: []Option{
				{Short: "b", Long: "breakpoint", Arg: "LINE", Description: "Set initial breakpoint"},
				{Short: "w", Long: "watch", Arg: "EXPR", Description: "Add watch expression"},
			},
		},
		{
			Name:        "version",
			Description: "Show version information",
			Options: []Option{
				{Short: "s", Long: "short", Description: "Show version only"},
				{Long: "build-info", Description: "Show build details"},
				{Long: "deps", Description: "Show dependencies"},
				{Long: "check-compat", Description: "Check go-llms compatibility"},
			},
		},
		{
			Name:        "completion",
			Description: "Generate shell completion scripts",
			Options: []Option{
				{Short: "l", Long: "list", Description: "List supported shells"},
			},
			Examples: []Example{
				{Command: "llmspell completion bash > ~/.bash_completion.d/llmspell", Description: "Install bash completion"},
				{Command: "llmspell completion zsh > ~/.zsh/completions/_llmspell", Description: "Install zsh completion"},
			},
		},
	}

	// Examples
	man.Examples = []Example{
		{
			Command:     "llmspell run hello.lua",
			Description: "Execute a simple Lua script",
		},
		{
			Command:     "llmspell repl",
			Description: "Start interactive REPL with default Lua engine",
		},
		{
			Command:     "llmspell new myspell --type agent --author \"John Doe\"",
			Description: "Create a new agent-based spell from template",
		},
		{
			Command:     "llmspell validate script.lua --security",
			Description: "Validate script with security analysis",
		},
		{
			Command:     "llmspell run complex.lua --profile development -p api_key=$API_KEY",
			Description: "Run script with development profile and environment variable",
		},
	}

	// Files
	man.Files = []string{
		"~/.llmspell/config.yaml  User configuration file",
		"~/.llmspell_history      REPL command history",
		"/etc/llmspell/config.yaml  System-wide configuration",
	}

	// See also
	man.SeeAlso = []string{
		"lua(1)",
		"node(1)",
	}

	// Authors
	man.Authors = []string{
		"LLMSpell Contributors",
	}

	// Bugs
	man.Bugs = "Report bugs at https://github.com/lexlapax/go-llmspell/issues"

	return man
}

// GenerateCommandManPage generates a man page for a specific command
func GenerateCommandManPage(command, version string) (*ManPage, error) {
	name := fmt.Sprintf("llmspell-%s", command)
	man := NewManPage(name, 1, version)

	switch command {
	case "run":
		man.Description = "execute a spell script"
		man.Synopsis = "\\fISCRIPT\\fR [\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "p", Long: "parameters", Arg: "KEY=VALUE", Description: "Parameters to pass to the script (can be repeated)"},
			{Short: "e", Long: "engine", Arg: "NAME", Description: "Script engine to use (auto-detected if not specified)"},
			{Short: "t", Long: "timeout", Arg: "SECONDS", Description: "Execution timeout in seconds", Default: "300"},
		}
		man.Examples = []Example{
			{Command: "llmspell run hello.lua", Description: "Run a simple script"},
			{Command: "llmspell run process.lua -p input=data.txt -p output=result.txt", Description: "Run with parameters"},
			{Command: "llmspell run script.txt --engine lua", Description: "Run with specific engine"},
		}

	case "repl":
		man.Description = "start interactive REPL (Read-Eval-Print Loop)"
		man.Synopsis = "[\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "e", Long: "engine", Arg: "NAME", Description: "Script engine to use", Default: "lua"},
			{Long: "no-history", Description: "Disable command history"},
			{Long: "no-highlight", Description: "Disable syntax highlighting"},
			{Long: "history-file", Arg: "PATH", Description: "Custom history file location"},
		}
		man.Description += "\n.PP\nREPL Commands:\n" +
			".TP\n.B .help\nShow available commands\n" +
			".TP\n.B .exit, .quit\nExit REPL\n" +
			".TP\n.B .clear\nClear current context\n" +
			".TP\n.B .save <file>\nSave session to file\n" +
			".TP\n.B .load <file>\nLoad and execute file\n" +
			".TP\n.B .engines\nList available engines\n" +
			".TP\n.B .mode <mode>\nSwitch input mode\n"

	case "new":
		man.Description = "create a new spell from a template"
		man.Synopsis = "\\fINAME\\fR [\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "t", Long: "type", Arg: "TYPE", Description: "Template type", Default: "basic"},
			{Short: "e", Long: "engine", Arg: "NAME", Description: "Script engine", Default: "lua"},
			{Short: "a", Long: "author", Arg: "NAME", Description: "Author name"},
			{Short: "d", Long: "description", Arg: "TEXT", Description: "Spell description"},
			{Short: "l", Long: "license", Arg: "NAME", Description: "License type", Default: "MIT"},
			{Short: "f", Long: "force", Description: "Overwrite existing directory"},
			{Long: "list", Description: "List available templates"},
		}
		man.Description += "\n.PP\nTemplate Types:\n" +
			".TP\n.B basic\nSimple single-file spell\n" +
			".TP\n.B advanced\nMulti-file spell with libraries\n" +
			".TP\n.B agent\nLLM agent with tools\n" +
			".TP\n.B workflow\nMulti-step workflow automation\n" +
			".TP\n.B interactive\nInteractive CLI spell\n"

	case "validate":
		man.Description = "validate a spell or script file"
		man.Synopsis = "\\fIPATH\\fR [\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "e", Long: "engine", Arg: "NAME", Description: "Force specific engine for validation"},
			{Long: "strict", Description: "Enable strict validation mode"},
			{Long: "security", Description: "Include security analysis"},
		}

	case "config":
		man.Description = "manage configuration settings"
		man.Synopsis = "[\\fIACTION\\fR [\\fIKEY\\fR [\\fIVALUE\\fR]]]"
		man.Examples = []Example{
			{Command: "llmspell config show", Description: "Display all configuration"},
			{Command: "llmspell config get engine.default", Description: "Get specific value"},
			{Command: "llmspell config set engine.default javascript", Description: "Set configuration value"},
			{Command: "llmspell config path", Description: "Show config file location"},
		}

	case "security":
		man.Description = "manage security profiles"
		man.Synopsis = "[\\fIACTION\\fR [\\fIPROFILE\\fR]]"
		man.Examples = []Example{
			{Command: "llmspell security list", Description: "List all profiles"},
			{Command: "llmspell security show sandbox", Description: "Show profile details"},
			{Command: "llmspell security validate production", Description: "Validate profile"},
		}

	case "engines":
		man.Description = "display information about available script engines"
		man.Synopsis = "[\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "d", Long: "details", Description: "Show detailed engine information"},
		}

	case "debug":
		man.Description = "debug a spell script with breakpoints"
		man.Synopsis = "\\fISCRIPT\\fR [\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "b", Long: "breakpoint", Arg: "LINE", Description: "Set initial breakpoint"},
			{Short: "w", Long: "watch", Arg: "EXPR", Description: "Add watch expression"},
		}

	case "version":
		man.Description = "show version information"
		man.Synopsis = "[\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "s", Long: "short", Description: "Show version only"},
			{Long: "build-info", Description: "Show build details"},
			{Long: "deps", Description: "Show dependencies"},
			{Long: "check-compat", Description: "Check go-llms compatibility"},
		}

	case "completion":
		man.Description = "generate shell completion scripts"
		man.Synopsis = "[\\fISHELL\\fR] [\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "l", Long: "list", Description: "List supported shells"},
		}
		man.Examples = []Example{
			{Command: "eval \"$(llmspell completion bash)\"", Description: "Enable bash completion"},
			{Command: "llmspell completion zsh > ~/.zsh/completions/_llmspell", Description: "Install zsh completion"},
		}

	case "man":
		man.Description = "generate man pages"
		man.Synopsis = "[\\fICOMMAND\\fR] [\\fIOPTIONS\\fR]"
		man.Options = []Option{
			{Short: "o", Long: "output", Arg: "FILE", Description: "Output file"},
			{Short: "d", Long: "dir", Arg: "DIR", Description: "Output directory"},
			{Short: "a", Long: "all", Description: "Generate all man pages"},
			{Short: "i", Long: "install", Description: "Install to system"},
			{Long: "format", Arg: "FORMAT", Description: "Output format (troff, text, html)", Default: "troff"},
		}

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}

	man.Files = []string{
		"~/.llmspell/config.yaml  User configuration file",
	}
	man.SeeAlso = []string{
		"llmspell(1)",
	}

	return man, nil
}

// GetAllCommands returns all available commands for man page generation
func GetAllCommands() []string {
	return []string{
		"run",
		"repl",
		"new",
		"validate",
		"config",
		"security",
		"engines",
		"debug",
		"version",
		"completion",
		"man",
	}
}
