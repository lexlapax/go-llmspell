# LLMSpell CLI Usage Guide

`llmspell` is a command-line tool for executing scripts that interact with Large Language Models (LLMs) through the go-llms library.

## Installation

```bash
go install github.com/lexlapax/go-llmspell/cmd/llmspell@latest
```

Or build from source:
```bash
git clone https://github.com/lexlapax/go-llmspell.git
cd go-llmspell
make build
```

## Quick Start

```bash
# Run a Lua script
llmspell run hello.lua

# Start interactive REPL
llmspell repl

# Create a new spell from template
llmspell new myspell --type basic

# Validate a script
llmspell validate script.lua
```

## Global Flags

These flags are available for all commands:

- `--debug` - Enable debug output with detailed error information
- `--config PATH` - Specify configuration file (default: `~/.llmspell/config.yaml`)
- `--quiet` - Suppress non-essential output
- `--verbose` - Show detailed execution information
- `--profile NAME` - Security profile to use (sandbox/development/production)
- `--help` - Show help for any command

## Commands

### run - Execute a spell script

Execute a script file with optional parameters.

**Usage:**
```bash
llmspell run <script> [flags]
```

**Flags:**
- `-e, --engine NAME` - Script engine to use (default: auto-detect)
- `-p, --param KEY=VALUE` - Parameters to pass to script (can be repeated)
- `-t, --timeout DURATION` - Execution timeout (default: 60s)
- `--env KEY=VALUE` - Environment variables for script
- `--dry-run` - Show what would be executed without running
- `--watch` - Watch script for changes and re-run
- `--progress` - Show execution progress

**Examples:**
```bash
# Run with parameters
llmspell run process.lua --param input=data.txt --param output=result.txt

# Run with specific engine
llmspell run script.txt --engine lua

# Run with timeout
llmspell run long-task.lua --timeout 5m

# Dry run to see execution plan
llmspell run complex.lua --dry-run

# Watch mode for development
llmspell run dev.lua --watch
```

### repl - Interactive REPL

Start an interactive Read-Eval-Print Loop for script development.

**Usage:**
```bash
llmspell repl [flags]
```

**Flags:**
- `-e, --engine NAME` - Script engine to use (default: lua)
- `--no-history` - Disable command history
- `--no-highlight` - Disable syntax highlighting
- `--history-file PATH` - Custom history file location

**REPL Commands:**
- `.help` - Show available commands
- `.exit` / `.quit` - Exit REPL
- `.clear` - Clear current context
- `.save <file>` - Save session to file
- `.load <file>` - Load session from file
- `.engines` - List available engines
- `.mode <mode>` - Switch input mode

**Examples:**
```bash
# Start REPL with Lua engine
llmspell repl

# Start with custom prompt from config
llmspell repl --config myconfig.yaml

# Start without history
llmspell repl --no-history
```

### new - Create spell from template

Generate a new spell project from predefined templates.

**Usage:**
```bash
llmspell new <name> [flags]
```

**Flags:**
- `-t, --type TYPE` - Template type (basic/advanced/agent/workflow/interactive)
- `-e, --engine NAME` - Script engine (default: lua)
- `--author NAME` - Spell author name
- `--description TEXT` - Spell description
- `--license NAME` - License type (default: MIT)
- `--force` - Overwrite existing directory
- `--list` - List available templates

**Template Types:**
- `basic` - Simple single-file spell
- `advanced` - Multi-file spell with libraries
- `agent` - LLM agent with tools
- `workflow` - Multi-step workflow automation
- `interactive` - Interactive CLI spell

**Examples:**
```bash
# Create basic spell
llmspell new hello-world

# Create agent spell
llmspell new my-agent --type agent --author "Jane Doe"

# Create JavaScript spell
llmspell new js-app --engine javascript

# List templates
llmspell new --list
```

### validate - Validate scripts

Check script syntax and spell configuration validity.

**Usage:**
```bash
llmspell validate <file...> [flags]
```

**Flags:**
- `-e, --engine NAME` - Force specific engine for validation
- `--strict` - Enable strict validation mode
- `--security` - Include security analysis

**Validates:**
- Script syntax correctness
- spell.yaml schema compliance
- Security profile compatibility
- Dependencies availability

**Examples:**
```bash
# Validate single script
llmspell validate script.lua

# Validate multiple files
llmspell validate *.lua

# Validate with security check
llmspell validate script.lua --security --profile sandbox

# Validate spell directory
llmspell validate ./myspell/
```

### config - Configuration management

Manage llmspell configuration settings.

**Usage:**
```bash
llmspell config <command> [flags]
```

**Subcommands:**
- `view` - Display current configuration
- `get <key>` - Get specific config value
- `set <key> <value>` - Set config value
- `reset <key>` - Reset to default value
- `init` - Create default config file
- `validate` - Check config validity
- `list` - List all config keys
- `export <file>` - Export configuration
- `import <file>` - Import configuration

**Examples:**
```bash
# View current config
llmspell config view

# Set default engine
llmspell config set engine.default javascript

# Get specific value
llmspell config get repl.prompt

# Initialize config file
llmspell config init

# Export config
llmspell config export backup.yaml
```

### security - Security profile management

View and manage security profiles for script execution.

**Usage:**
```bash
llmspell security <command> [flags]
```

**Subcommands:**
- `list` - List available profiles
- `view <profile>` - Show profile details
- `validate <profile>` - Check profile validity
- `check <profile> <permission>` - Check specific permission
- `compare <profile1> <profile2>` - Compare profiles
- `export <profile>` - Export profile as YAML

**Security Profiles:**
- `sandbox` - Restricted environment, no file/network access
- `development` - Permissive for local development
- `production` - Balanced security for production use

**Examples:**
```bash
# List all profiles
llmspell security list

# View sandbox profile
llmspell security view sandbox

# Check permission
llmspell security check sandbox file_read

# Compare profiles
llmspell security compare sandbox development
```

### engines - Script engine information

Display information about available script engines.

**Usage:**
```bash
llmspell engines <command> [flags]
```

**Subcommands:**
- `list` - List all engines
- `info <engine>` - Show engine details
- `capabilities <engine>` - List engine features
- `detect <file>` - Detect engine for file
- `check <engine>` - Health check
- `benchmark <engine>` - Performance test

**Flags:**
- `--format FORMAT` - Output format (text/json/yaml)
- `--verbose` - Show detailed information

**Examples:**
```bash
# List engines
llmspell engines list

# Show Lua engine info
llmspell engines info lua

# Detect engine for file
llmspell engines detect script.lua

# Benchmark engine
llmspell engines benchmark lua
```

### debug - Interactive debugger

Debug scripts with breakpoints and step execution.

**Usage:**
```bash
llmspell debug <script> [flags]
```

**Debug Commands:**
- `break <line> [condition]` - Set breakpoint
- `clear <id>` - Clear breakpoint
- `continue` / `c` - Continue execution
- `step` / `s` - Step to next line
- `next` / `n` - Step over function
- `out` / `o` - Step out of function
- `print <expr>` / `p` - Evaluate expression
- `where` / `w` - Show call stack
- `locals` / `l` - Show local variables
- `watch <expr>` - Add watch expression
- `help` / `h` - Show debug help
- `quit` / `q` - Exit debugger

**Examples:**
```bash
# Debug with initial breakpoint
llmspell debug script.lua

# Debug with spell.yaml config
llmspell debug ./myspell/
```

### version - Version information

Display version and build information.

**Usage:**
```bash
llmspell version [flags]
```

**Flags:**
- `-s, --short` - Show version only
- `--build-info` - Show build details
- `--format FORMAT` - Output format (text/json)
- `--deps` - Show dependencies
- `--check-compat` - Check go-llms compatibility

**Examples:**
```bash
# Show version
llmspell version

# Short version only
llmspell version --short

# Detailed with dependencies
llmspell version --build-info --deps
```

### completion - Shell completion scripts

Generate shell completion scripts for command-line tab completion.

**Usage:**
```bash
llmspell completion [shell] [flags]
```

**Flags:**
- `-l, --list` - List supported shells

**Supported Shells:**
- `bash` - Bash shell (requires bash-completion)
- `zsh` - Z shell
- `fish` - Fish shell
- `powershell` - PowerShell
- `sh` - POSIX shell (reference only)

**Installation Examples:**

**Bash:**
```bash
# Add to ~/.bashrc
source <(llmspell completion bash)

# Or save to completion directory
llmspell completion bash > ~/.local/share/bash-completion/completions/llmspell
```

**Zsh:**
```bash
# Add to ~/.zshrc (before compinit)
source <(llmspell completion zsh)

# Or save to fpath directory
llmspell completion zsh > ~/.zsh/completions/_llmspell
```

**Fish:**
```bash
# Save to fish completions directory
llmspell completion fish > ~/.config/fish/completions/llmspell.fish
```

**PowerShell:**
```powershell
# Add to PowerShell profile
llmspell completion powershell | Out-String | Invoke-Expression

# Find profile location
echo $PROFILE
```

**Features:**
- Command name completion
- Flag and option completion
- Subcommand completion
- Enum value completion (e.g., --engine lua/javascript/tengo)
- File path completion for relevant commands

**Examples:**
```bash
# List supported shells
llmspell completion --list

# Generate bash completion
llmspell completion bash

# Auto-detect current shell
llmspell completion
```

### man - Manual page generation

Generate UNIX manual pages for llmspell and its commands.

**Usage:**
```bash
llmspell man [command] [flags]
```

**Flags:**
- `-o, --output FILE` - Output to file instead of stdout
- `-d, --dir DIR` - Output directory for generating all man pages
- `-a, --all` - Generate all man pages
- `-i, --install` - Install man pages to system directory
- `--format FORMAT` - Output format (troff/text/html, default: troff)

**Examples:**
```bash
# View main man page
llmspell man | man -l -

# Generate man page for specific command
llmspell man run > llmspell-run.1

# Generate all man pages to directory
llmspell man --all --dir ./man

# Install man pages (may require sudo)
llmspell man --install

# Generate as plain text
llmspell man --format text

# Generate HTML documentation
llmspell man --format html > llmspell.html
```

**Generated Man Pages:**
- `llmspell.1` - Main program manual
- `llmspell-run.1` - Run command manual
- `llmspell-repl.1` - REPL command manual
- `llmspell-new.1` - New command manual
- And man pages for all other commands

**Installation Locations:**
The `--install` flag will try to install to:
1. `/usr/local/share/man/man1` (if writable)
2. `/usr/share/man/man1` (if writable)
3. `~/.local/share/man/man1` (user directory)

After installation, you can use standard man commands:
```bash
man llmspell
man llmspell-run
man llmspell-repl
```

## Configuration

Configuration uses a layered approach:
1. Default values
2. Config file (`~/.llmspell/config.yaml`)
3. Environment variables (`LLMSPELL_*`)
4. Command-line flags

Example config file:
```yaml
engine:
  default: lua
  timeout: 60s

repl:
  prompt: "lua> "
  save_history: true
  history_file: ~/.llmspell_history
  syntax_highlight: true

security:
  profile: development

debug:
  enabled: false
  verbose: false
```

## Script Parameters

Scripts receive parameters through the `params` global:

```lua
-- Access parameters
local input = params.input or "default.txt"
local count = tonumber(params.count) or 10

-- Use in script
print("Processing " .. input .. " with count " .. count)
```

Pass parameters via CLI:
```bash
llmspell run script.lua --param input=data.txt --param count=20
```

## Security Profiles

Security profiles control what scripts can access:

**Sandbox Profile:**
- No file system access
- No network access
- No external commands
- Limited CPU/memory

**Development Profile:**
- Full file system access
- Network access allowed
- External commands allowed
- Resource limits relaxed

**Production Profile:**
- Read-only file access
- Restricted network access
- No external commands
- Moderate resource limits

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Invalid arguments
- `3` - Script execution error
- `4` - Timeout
- `5` - Security violation
- `127` - Command not found

## Environment Variables

- `LLMSPELL_CONFIG` - Config file path
- `LLMSPELL_ENGINE_DEFAULT` - Default engine
- `LLMSPELL_ENGINE_TIMEOUT` - Default timeout
- `LLMSPELL_REPL_PROMPT` - REPL prompt
- `LLMSPELL_SECURITY_PROFILE` - Default security profile
- `LLMSPELL_DEBUG` - Enable debug mode

## Tips and Best Practices

1. **Use spell.yaml** for complex projects:
   ```yaml
   name: my-spell
   version: 1.0.0
   engine: lua
   entry_point: main.lua
   parameters:
     input:
       type: string
       required: true
       description: Input file path
   ```

2. **Leverage templates** for new projects:
   ```bash
   llmspell new myproject --type workflow
   ```

3. **Test with sandbox profile** before production:
   ```bash
   llmspell run script.lua --profile sandbox
   ```

4. **Use debug mode** for troubleshooting:
   ```bash
   llmspell run script.lua --debug
   ```

5. **Watch mode** for development:
   ```bash
   llmspell run dev.lua --watch --verbose
   ```

## See Also

- [Configuration Guide](configuration.md)
- [REPL Guide](repl-guide.md)
- [Security Documentation](../security/README.md)
- [API Reference](../api/README.md)