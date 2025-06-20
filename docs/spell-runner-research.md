# Spell Runner Research & Implementation Plan

## Overview
The spell runner is the command-line interface for executing go-llmspell scripts. It needs to detect script types, load appropriate engines, and provide various development tools.

## Script Detection Strategy

### File Extension Based
- `.lua` → Lua engine
- `.js`, `.mjs` → JavaScript engine  
- `.tengo` → Tengo engine
- `spell.yaml` → Read engine type from metadata

### Spell Directory Structure
```
spell-name/
├── spell.yaml     # Metadata (required)
├── main.lua       # Entry point (configurable)
├── lib/          # Optional libraries
└── README.md     # Optional documentation
```

## Command Structure

### Core Commands
1. **run** (default) - Execute a spell
   ```bash
   llmspell [run] <spell-path> [args...]
   llmspell spell.yaml
   llmspell main.lua --param value
   ```

2. **validate** - Syntax and security validation
   ```bash
   llmspell validate <spell-path>
   ```

3. **engines** - List available engines
   ```bash
   llmspell engines [--verbose]
   ```

4. **repl** - Interactive mode
   ```bash
   llmspell repl [--engine lua|js|tengo]
   ```

5. **help** - Show help
   ```bash
   llmspell help [command]
   ```

6. **version** - Version info
   ```bash
   llmspell version
   ```

### Development Commands
1. **debug** - Run with debugging
   ```bash
   llmspell debug <spell-path> [--breakpoint line]
   ```

2. **security** - Show security profile
   ```bash
   llmspell security [--profile name]
   ```

3. **config** - Manage configuration
   ```bash
   llmspell config [get|set|list] [key] [value]
   ```

## Library Recommendations

### CLI Framework: Kong (github.com/alecthomas/kong)
**Reasons:**
- Modern struct-based approach with minimal boilerplate
- Type-safe with automatic validation
- Auto-generated help with excellent formatting
- Supports complex command hierarchies naturally
- Created by the author of Kingpin as its successor
- Developers migrating from Cobra praise its simplicity
- Excellent interface design that matches Go idioms

**Example Structure:**
```go
var CLI struct {
    Config struct {
        File string `type:"path" help:"Config file path"`
    } `embed:"" prefix:"config-"`
    
    Run struct {
        Script string   `arg:"" required:"" help:"Script to run"`
        Args   []string `arg:"" optional:"" help:"Arguments to pass"`
        Debug  bool     `help:"Enable debug mode"`
    } `cmd:"" default:"1" help:"Run a spell"`
    
    Repl struct {
        Engine string `enum:"lua,js,tengo" default:"lua" help:"Script engine"`
    } `cmd:"" help:"Start interactive REPL"`
    
    Engines struct{} `cmd:"" help:"List available engines"`
}
```

### Configuration Management: Koanf (github.com/knadh/koanf/v2)
**Reasons:**
- Lightweight alternative to Viper with better modularity
- Supports multiple formats (JSON, YAML, TOML, HCL, env, flags)
- Modular design - only import what you need
- Respects language specs (doesn't force lowercase keys like Viper)
- Watch() support for auto-reloading configs
- Clean provider/parser interfaces for extensibility
- Minimal dependencies

**Key Features:**
- Layer multiple config sources (defaults → file → env → flags)
- Thread-safe with proper mutex usage
- Struct provider for loading from Go structs
- Extensible for custom formats/sources

### REPL Library: chzyer/readline
**Reasons:**
- Pure Go implementation of GNU readline
- Cross-platform including Windows
- Rich feature set for interactive shells
- Used by many Go REPLs and interactive tools
- Well-maintained and stable

**Key Features:**
- History with file persistence
- Auto-completion with custom completers
- ANSI color support in prompts
- Vim mode support
- Password input with masking
- Context-aware suffix removal
- Prefill next input line

**REPL Integration Example:**
```go
rl, err := readline.NewEx(&readline.Config{
    Prompt:          "llmspell> ",
    HistoryFile:     filepath.Join(homeDir, ".llmspell_history"),
    AutoComplete:    completer,
    InterruptPrompt: "^C",
    EOFPrompt:       "exit",
    VimMode:         vimMode,
})
```

### Alternative Considerations

1. **CLI: urfave/cli/v2**
   - Still a solid choice if preferring callback-style over struct tags
   - More traditional approach
   - Larger community

2. **Config: Viper**
   - More features but heavier
   - Larger ecosystem
   - Consider if need advanced features

3. **REPL: reeflective/readline**
   - Modern alternative to chzyer/readline
   - Full .inputrc support
   - Fuzzy search in history
   - More advanced but newer

## Architecture

### Package Structure
```
/cmd/llmspell/
├── main.go              # Entry point
├── commands/            # Command implementations
│   ├── run.go
│   ├── validate.go
│   ├── engines.go
│   ├── repl.go
│   └── ...
└── version.go          # Version info

/pkg/runner/            # Core runner logic
├── runner.go           # Main runner interface
├── spell_loader.go     # Load spell metadata
├── engine_selector.go  # Select appropriate engine
├── config.go          # Configuration management
└── security.go        # Security profiles

/pkg/repl/             # REPL implementation
├── repl.go            # REPL interface
├── lua_repl.go        # Lua-specific REPL
├── js_repl.go         # JS-specific REPL
└── tengo_repl.go      # Tengo-specific REPL
```

## Implementation Plan

### Phase 1: Core Runner (Task 2.4.2.2)
1. Create `/pkg/runner/` package structure
2. Implement spell loader (reads spell.yaml)
3. Implement engine selector (file extension → engine)
4. Create basic runner that can execute scripts
5. Add parameter passing to scripts

### Phase 2: CLI Structure with Kong
1. Set up Kong in `/cmd/llmspell/main.go`
   ```go
   type CLI struct {
       Globals
       Run     RunCmd     `cmd:"" default:"1"`
       Validate ValidateCmd `cmd:""`
       Engines EnginesCmd  `cmd:""`
       Repl    ReplCmd    `cmd:""`
       Config  ConfigCmd  `cmd:""`
       Version VersionCmd `cmd:""`
   }
   ```
2. Implement command handlers in `/cmd/llmspell/commands/`
3. Add global flags (debug, config-file, etc.)
4. Set up Kong configuration with help formatting

### Phase 3: Configuration with Koanf
1. Create `/pkg/config/` package using koanf/v2
2. Set up layered configuration:
   - Default values
   - Config file (~/.llmspell/config.yaml)
   - Environment variables (LLMSPELL_*)
   - Command-line flags
3. Implement config providers for each source
4. Add Watch() support for config reloading
5. Create config schema validation

### Phase 4: REPL Mode with readline
1. Create `/pkg/repl/` package
2. Implement base REPL using chzyer/readline
3. Add engine-specific REPL implementations:
   - Lua REPL with auto-completion
   - JavaScript REPL with syntax highlighting
   - Tengo REPL with history
4. Implement custom completers for each engine
5. Add REPL commands (.help, .exit, .clear, etc.)

### Phase 5: Validation & Security
1. Implement `validate` command
2. Add syntax checking per engine
3. Add security profile validation
4. Implement `security` command
5. Create security profile management

### Phase 6: Advanced Features
1. Add `debug` command with breakpoints
2. Add spell template generation
3. Add spell packaging/distribution
4. Add performance profiling
5. Implement spell dependency management

## Configuration Management

### Sources (in priority order)
1. Command-line flags
2. Environment variables (LLMSPELL_*)
3. Config file (~/.llmspell/config.yaml)
4. Project config (.llmspellrc)
5. Default values

### Configuration Schema
```yaml
# ~/.llmspell/config.yaml
engines:
  lua:
    memory_limit: 64MB
    timeout: 30s
  javascript:
    memory_limit: 128MB
    
providers:
  openai:
    api_key: ${OPENAI_API_KEY}
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    
security:
  default_profile: sandbox
  allow_network: false
  allow_filesystem: readonly
```

## Security Profiles

### Predefined Profiles
1. **sandbox** - Maximum restrictions
   - No network access
   - Read-only filesystem
   - Limited memory/CPU
   
2. **development** - Relaxed for development
   - Network access allowed
   - Read/write to project directory
   - Higher resource limits
   
3. **production** - Balanced security
   - Controlled network access
   - Specific directory access
   - Monitoring enabled

## Error Handling

### User-Friendly Errors
- Clear error messages
- Suggestions for fixes
- Links to documentation
- Exit codes following conventions

### Debug Mode
- Stack traces
- Engine internal errors
- Bridge call traces
- Performance metrics

## Testing Strategy

### Unit Tests
- Each command tested individually
- Mock engine registry
- Test configuration loading
- Test security profiles

### Integration Tests
- Full CLI execution tests
- Real spell execution
- Multi-engine tests
- Error scenarios

### Example Test Spells
- Create test spells for each feature
- Cover success and failure cases
- Document expected behaviors

## Documentation Requirements

### User Documentation
- Getting started guide
- Command reference
- Configuration guide
- Security best practices
- Troubleshooting guide

### Developer Documentation
- Architecture overview
- Adding new commands
- Extending the runner
- Contributing guidelines

## Future Enhancements

### Potential Features
- Spell marketplace/registry
- Remote spell execution
- Distributed execution
- Performance monitoring
- Spell dependency management
- Version management
- Hot reload during development
- Integration with IDEs

### Extensibility Points
- Plugin system for commands
- Custom security profiles
- Engine extensions
- Output formatters
- Progress indicators

## Next Steps

1. Create basic runner package structure
2. Implement spell loading and engine selection
3. Set up CLI with urfave/cli/v2
4. Implement core commands
5. Add tests and documentation
6. Iterate based on user feedback