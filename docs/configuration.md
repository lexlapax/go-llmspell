# LLMSpell Configuration Guide

LLMSpell uses a flexible, layered configuration system that allows you to customize behavior through configuration files, environment variables, and command-line flags.

## Configuration Hierarchy

Configuration values are resolved in the following order (highest priority first):

1. **Command-line flags** - Override all other settings
2. **Environment variables** - Override config file settings
3. **Configuration file** - User-defined settings
4. **Default values** - Built-in defaults

## Configuration File

### Location

LLMSpell looks for configuration files in these locations (in order):

1. Path specified by `--config` flag
2. `$LLMSPELL_CONFIG` environment variable
3. `.llmspell.yaml` in current directory
4. `$HOME/.llmspell/config.yaml` (default)
5. `/etc/llmspell/config.yaml` (system-wide)

### Format

Configuration files use YAML format. Here's a complete example:

```yaml
# Engine Configuration
engine:
  default: lua                    # Default script engine
  timeout: 60s                    # Execution timeout
  max_memory: 512MB              # Memory limit
  max_cpu: 2                     # CPU cores limit
  registry:
    search_paths:                # Additional engine search paths
      - ~/.llmspell/engines
      - /usr/local/lib/llmspell/engines

# REPL Configuration  
repl:
  prompt: "lua> "                # Interactive prompt
  continuation_prompt: "... "    # Multi-line continuation
  save_history: true            # Save command history
  history_file: ~/.llmspell_history
  history_size: 1000            # Max history entries
  syntax_highlight: true        # Enable syntax coloring
  auto_complete: true           # Enable tab completion
  colors:
    keyword: blue
    string: green
    number: yellow
    comment: gray

# Security Configuration
security:
  profile: development          # Default security profile
  allow_override: true          # Allow profile override
  custom_profiles:              # Define custom profiles
    restricted:
      file_system: none
      network: none
      external_commands: false
      resource_limits:
        max_memory: 128MB
        max_cpu: 1
        max_time: 30s

# Debug Configuration
debug:
  enabled: false                # Global debug mode
  verbose: false                # Verbose output
  trace: false                  # Trace execution
  log_file: ~/.llmspell/debug.log
  log_level: info              # debug, info, warn, error
  breakpoints:                  # Default breakpoints
    - file: main.lua
      line: 10
      condition: "x > 5"

# Runner Configuration
runner:
  parallel: false               # Parallel execution
  max_workers: 4                # Worker threads
  progress: true                # Show progress bars
  cache:
    enabled: true
    directory: ~/.llmspell/cache
    max_size: 1GB
    ttl: 24h

# Template Configuration
templates:
  author: "Your Name"           # Default author
  license: MIT                  # Default license
  organization: ""              # Organization name
  repository: ""                # Default repository

# Network Configuration
network:
  proxy: ""                     # HTTP proxy
  timeout: 30s                  # Network timeout
  retry: 3                      # Retry attempts
  retry_delay: 1s              # Delay between retries

# Logging Configuration
logging:
  level: info                   # Logging level
  format: text                  # text or json
  output: stderr                # stdout, stderr, or file path
  timestamp: true               # Include timestamps
  caller: false                 # Include caller info

# Plugin Configuration (future)
plugins:
  enabled: false
  directory: ~/.llmspell/plugins
  auto_load: []                 # Plugins to auto-load
```

## Environment Variables

All configuration options can be set via environment variables. The pattern is:

```
LLMSPELL_<SECTION>_<KEY>=value
```

### Common Environment Variables

```bash
# Engine settings
export LLMSPELL_ENGINE_DEFAULT=javascript
export LLMSPELL_ENGINE_TIMEOUT=120s

# REPL settings
export LLMSPELL_REPL_PROMPT="js> "
export LLMSPELL_REPL_SAVE_HISTORY=false

# Security settings
export LLMSPELL_SECURITY_PROFILE=sandbox

# Debug settings
export LLMSPELL_DEBUG_ENABLED=true
export LLMSPELL_DEBUG_VERBOSE=true

# Special variables
export LLMSPELL_CONFIG=/path/to/config.yaml
export LLMSPELL_HOME=/custom/llmspell/home
```

### Nested Configuration

For nested configuration values, use underscores:

```bash
# Sets security.custom_profiles.restricted.file_system
export LLMSPELL_SECURITY_CUSTOM_PROFILES_RESTRICTED_FILE_SYSTEM=read_only
```

## Command-Line Configuration

### Global Flags

These flags are available for all commands:

```bash
# Override config file
llmspell --config /path/to/config.yaml run script.lua

# Set debug mode
llmspell --debug run script.lua

# Set verbosity
llmspell --verbose run script.lua

# Set security profile
llmspell --profile sandbox run script.lua
```

### Command-Specific Configuration

Many commands accept configuration overrides:

```bash
# Override engine timeout
llmspell run script.lua --timeout 30s

# Override REPL prompt
llmspell repl --engine javascript

# Disable history
llmspell repl --no-history
```

## Configuration Commands

### View Configuration

```bash
# View current configuration
llmspell config view

# View with defaults shown
llmspell config view --show-defaults

# View specific section
llmspell config view engine
```

### Get/Set Values

```bash
# Get specific value
llmspell config get engine.default

# Set value
llmspell config set engine.default javascript

# Set with validation
llmspell config set engine.timeout 30s --validate

# Reset to default
llmspell config reset engine.timeout
```

### Configuration Management

```bash
# Initialize default config
llmspell config init

# Validate configuration
llmspell config validate

# Export configuration
llmspell config export backup.yaml

# Import configuration
llmspell config import custom.yaml

# List all keys
llmspell config list
```

## Per-Project Configuration

You can have project-specific configuration by placing `.llmspell.yaml` in your project directory:

```yaml
# .llmspell.yaml in project root
engine:
  default: javascript    # This project uses JavaScript

parameters:
  api_key: ${API_KEY}   # Environment variable expansion
  base_url: https://api.example.com

# Project-specific security
security:
  profile: production
```

## Configuration Profiles

You can define multiple configuration profiles:

```yaml
# ~/.llmspell/profiles/dev.yaml
debug:
  enabled: true
  verbose: true
security:
  profile: development

# ~/.llmspell/profiles/prod.yaml  
debug:
  enabled: false
security:
  profile: production
```

Load profiles using:
```bash
llmspell --config ~/.llmspell/profiles/dev.yaml run script.lua
```

## Security Profiles

### Built-in Profiles

**Sandbox:**
```yaml
security:
  profile: sandbox
  # Implies:
  # - No file system access
  # - No network access
  # - No external commands
  # - Strict resource limits
```

**Development:**
```yaml
security:
  profile: development
  # Implies:
  # - Full file system access
  # - Network access allowed
  # - External commands allowed
  # - Relaxed resource limits
```

**Production:**
```yaml
security:
  profile: production
  # Implies:
  # - Read-only file access
  # - Restricted network access
  # - No external commands
  # - Moderate resource limits
```

### Custom Security Profiles

Define custom profiles in configuration:

```yaml
security:
  custom_profiles:
    api_only:
      file_system: none
      network: restricted
      allowed_hosts:
        - api.openai.com
        - api.anthropic.com
      external_commands: false
      resource_limits:
        max_memory: 256MB
        max_time: 60s
```

## Advanced Configuration

### Variable Expansion

Configuration supports environment variable expansion:

```yaml
network:
  proxy: ${HTTP_PROXY}
  
parameters:
  api_key: ${OPENAI_API_KEY}
  model: ${LLM_MODEL:-gpt-4}  # With default value
```

### Configuration Validation

LLMSpell validates configuration on load:

```bash
# Validate current config
llmspell config validate

# Validate specific file
llmspell config validate --config custom.yaml

# Validate with verbose output
llmspell config validate --verbose
```

### Configuration Schema

```yaml
# Validation rules example
engine:
  default:
    type: string
    enum: [lua, javascript, tengo]
  timeout:
    type: duration
    min: 1s
    max: 24h
    
security:
  profile:
    type: string
    enum: [sandbox, development, production]
    required: true
```

## Best Practices

1. **Use project-specific config** for team consistency:
   ```bash
   echo '.llmspell.yaml' >> .gitignore  # Don't commit secrets
   ```

2. **Layer configurations** for different environments:
   ```bash
   # Base config
   cp ~/.llmspell/config.yaml ~/.llmspell/config.base.yaml
   
   # Environment-specific
   llmspell --config ~/.llmspell/config.dev.yaml run script.lua
   ```

3. **Secure sensitive values** using environment variables:
   ```yaml
   parameters:
     api_key: ${API_KEY}  # Never hardcode
   ```

4. **Validate before deployment**:
   ```bash
   llmspell config validate --config production.yaml
   ```

5. **Use appropriate security profiles**:
   - Development: `development` profile for local work
   - Testing: `sandbox` profile for CI/CD
   - Production: `production` or custom restrictive profile

## Troubleshooting

### Debug Configuration Loading

```bash
# Show configuration resolution
LLMSPELL_DEBUG=true llmspell config view

# Trace configuration sources
llmspell config view --trace

# Show effective configuration
llmspell config view --effective
```

### Common Issues

1. **Configuration not loading:**
   ```bash
   # Check file location
   llmspell config view --show-source
   
   # Verify file syntax
   llmspell config validate --config myconfig.yaml
   ```

2. **Environment variables not working:**
   ```bash
   # Check variable format
   echo $LLMSPELL_ENGINE_DEFAULT
   
   # Debug with trace
   LLMSPELL_DEBUG=true llmspell run script.lua
   ```

3. **Precedence confusion:**
   ```bash
   # Show value sources
   llmspell config get engine.timeout --verbose
   ```

## Migration Guide

### From v0.2.x

If upgrading from an older version:

```bash
# Backup old config
cp ~/.llmspell.conf ~/.llmspell.conf.backup

# Migrate to new format
llmspell config migrate ~/.llmspell.conf

# Validate new config
llmspell config validate
```

## Reference

### Configuration Types

- `string` - Text values
- `integer` - Whole numbers
- `float` - Decimal numbers
- `boolean` - true/false
- `duration` - Time durations (e.g., "30s", "5m", "1h")
- `size` - Memory sizes (e.g., "512MB", "1GB")
- `array` - Lists of values
- `map` - Key-value pairs

### Duration Format

- `s` - seconds (e.g., "30s")
- `m` - minutes (e.g., "5m")
- `h` - hours (e.g., "2h")
- `d` - days (e.g., "1d")

### Size Format

- `B` - bytes
- `KB` - kilobytes
- `MB` - megabytes  
- `GB` - gigabytes
- `TB` - terabytes

## See Also

- [CLI Usage Guide](cli-usage.md)
- [REPL Guide](repl-guide.md)
- [Security Documentation](../security/README.md)