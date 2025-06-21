// ABOUTME: This file defines the main configuration structure for go-llmspell CLI.
// ABOUTME: It provides comprehensive configuration options for engines, security, logging, and runtime behavior.

package config

import (
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Config represents the complete configuration for go-llmspell
type Config struct {
	// Core settings
	Version string `yaml:"version" json:"version" env:"LLMSPELL_VERSION"`
	Debug   bool   `yaml:"debug" json:"debug" env:"LLMSPELL_DEBUG"`
	Quiet   bool   `yaml:"quiet" json:"quiet" env:"LLMSPELL_QUIET"`
	Verbose bool   `yaml:"verbose" json:"verbose" env:"LLMSPELL_VERBOSE"`

	// Engine configuration
	Engine EngineConfig `yaml:"engine" json:"engine"`

	// Security settings
	Security SecurityConfig `yaml:"security" json:"security"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`

	// CLI-specific settings
	CLI CLIConfig `yaml:"cli" json:"cli"`

	// REPL configuration
	REPL REPLConfig `yaml:"repl" json:"repl"`

	// Spell runner settings
	Runner RunnerConfig `yaml:"runner" json:"runner"`

	// Template settings
	Templates TemplateConfig `yaml:"templates" json:"templates"`
}

// EngineConfig holds configuration for script engines
type EngineConfig struct {
	// Default engine to use when not specified
	Default string `yaml:"default" json:"default" env:"LLMSPELL_ENGINE_DEFAULT"`

	// Resource limits
	MemoryLimit    int64         `yaml:"memory_limit" json:"memory_limit" env:"LLMSPELL_ENGINE_MEMORY_LIMIT"`
	TimeoutLimit   time.Duration `yaml:"timeout_limit" json:"timeout_limit" env:"LLMSPELL_ENGINE_TIMEOUT_LIMIT"`
	GoroutineLimit int           `yaml:"goroutine_limit" json:"goroutine_limit" env:"LLMSPELL_ENGINE_GOROUTINE_LIMIT"`

	// Engine-specific configurations
	Lua        LuaEngineConfig        `yaml:"lua" json:"lua"`
	JavaScript JavaScriptEngineConfig `yaml:"javascript" json:"javascript"`
	Tengo      TengoEngineConfig      `yaml:"tengo" json:"tengo"`
}

// LuaEngineConfig holds Lua-specific configuration
type LuaEngineConfig struct {
	// Pool settings
	PoolMinSize     int           `yaml:"pool_min_size" json:"pool_min_size"`
	PoolMaxSize     int           `yaml:"pool_max_size" json:"pool_max_size"`
	PoolIdleTimeout time.Duration `yaml:"pool_idle_timeout" json:"pool_idle_timeout"`

	// Performance settings
	HealthThreshold  float64       `yaml:"health_threshold" json:"health_threshold"`
	CleanupInterval  time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	CacheEnabled     bool          `yaml:"cache_enabled" json:"cache_enabled"`
	CompileOptimized bool          `yaml:"compile_optimized" json:"compile_optimized"`

	// Standard library settings
	StdlibModules []string `yaml:"stdlib_modules" json:"stdlib_modules"`
}

// JavaScriptEngineConfig holds JavaScript-specific configuration
type JavaScriptEngineConfig struct {
	// ES version support
	ESVersion string `yaml:"es_version" json:"es_version"`

	// Runtime settings
	Strict bool `yaml:"strict" json:"strict"`

	// Module support
	ModuleSupport bool `yaml:"module_support" json:"module_support"`
}

// TengoEngineConfig holds Tengo-specific configuration
type TengoEngineConfig struct {
	// VM settings
	MaxAllocs int `yaml:"max_allocs" json:"max_allocs"`

	// Module settings
	ImportLimit int `yaml:"import_limit" json:"import_limit"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// Security profile (sandbox, development, production)
	Profile string `yaml:"profile" json:"profile" env:"LLMSPELL_SECURITY_PROFILE"`

	// Filesystem access mode
	FileSystemMode engine.FSMode `yaml:"filesystem_mode" json:"filesystem_mode"`

	// Allowed and disabled modules
	AllowedModules  []string `yaml:"allowed_modules" json:"allowed_modules"`
	DisabledModules []string `yaml:"disabled_modules" json:"disabled_modules"`

	// Validation settings
	EnableValidation  bool `yaml:"enable_validation" json:"enable_validation"`
	StrictValidation  bool `yaml:"strict_validation" json:"strict_validation"`
	SecurityWarnings  bool `yaml:"security_warnings" json:"security_warnings"`
	PerformanceChecks bool `yaml:"performance_checks" json:"performance_checks"`

	// Sandbox settings
	NetworkAccess   bool     `yaml:"network_access" json:"network_access"`
	FileAccess      bool     `yaml:"file_access" json:"file_access"`
	ProcessAccess   bool     `yaml:"process_access" json:"process_access"`
	AllowedDomains  []string `yaml:"allowed_domains" json:"allowed_domains"`
	AllowedPaths    []string `yaml:"allowed_paths" json:"allowed_paths"`
	DeniedPatterns  []string `yaml:"denied_patterns" json:"denied_patterns"`
	MaxFileSize     int64    `yaml:"max_file_size" json:"max_file_size"`
	MaxNetworkCalls int      `yaml:"max_network_calls" json:"max_network_calls"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	// Log level (debug, info, warn, error)
	Level string `yaml:"level" json:"level" env:"LLMSPELL_LOG_LEVEL"`

	// Output format (text, json)
	Format string `yaml:"format" json:"format" env:"LLMSPELL_LOG_FORMAT"`

	// Output destination (stdout, stderr, file path)
	Output string `yaml:"output" json:"output" env:"LLMSPELL_LOG_OUTPUT"`

	// Log file settings (when output is a file)
	FileSettings LogFileSettings `yaml:"file" json:"file"`

	// Component-specific logging
	Components ComponentLogging `yaml:"components" json:"components"`
}

// LogFileSettings holds log file configuration
type LogFileSettings struct {
	MaxSize    int  `yaml:"max_size" json:"max_size"`       // MB
	MaxBackups int  `yaml:"max_backups" json:"max_backups"` // Number of backup files
	MaxAge     int  `yaml:"max_age" json:"max_age"`         // Days
	Compress   bool `yaml:"compress" json:"compress"`       // Compress rotated files
}

// ComponentLogging holds component-specific logging levels
type ComponentLogging struct {
	Engine   string `yaml:"engine" json:"engine"`
	Bridge   string `yaml:"bridge" json:"bridge"`
	Runner   string `yaml:"runner" json:"runner"`
	Security string `yaml:"security" json:"security"`
	REPL     string `yaml:"repl" json:"repl"`
}

// CLIConfig holds CLI-specific configuration
type CLIConfig struct {
	// Output formatting
	ColorOutput bool   `yaml:"color_output" json:"color_output"`
	NoColor     bool   `yaml:"no_color" json:"no_color" env:"NO_COLOR"`
	Theme       string `yaml:"theme" json:"theme"`

	// Progress indicators
	ShowProgress  bool   `yaml:"show_progress" json:"show_progress"`
	ProgressStyle string `yaml:"progress_style" json:"progress_style"`

	// Help and documentation
	PagerCommand string `yaml:"pager_command" json:"pager_command" env:"PAGER"`
	Editor       string `yaml:"editor" json:"editor" env:"EDITOR"`

	// Shell integration
	EnableCompletion bool   `yaml:"enable_completion" json:"enable_completion"`
	CompletionShell  string `yaml:"completion_shell" json:"completion_shell"`
}

// REPLConfig holds REPL-specific configuration
type REPLConfig struct {
	// History settings
	HistoryFile     string `yaml:"history_file" json:"history_file"`
	HistorySize     int    `yaml:"history_size" json:"history_size"`
	SaveHistory     bool   `yaml:"save_history" json:"save_history"`
	HistoryDuration string `yaml:"history_duration" json:"history_duration"`

	// Display settings
	Prompt          string `yaml:"prompt" json:"prompt"`
	ContinuePrompt  string `yaml:"continue_prompt" json:"continue_prompt"`
	SyntaxHighlight bool   `yaml:"syntax_highlight" json:"syntax_highlight"`
	AutoComplete    bool   `yaml:"auto_complete" json:"auto_complete"`

	// Behavior settings
	MultiLine    bool `yaml:"multi_line" json:"multi_line"`
	AutoIndent   bool `yaml:"auto_indent" json:"auto_indent"`
	BracketMatch bool `yaml:"bracket_match" json:"bracket_match"`
	VimMode      bool `yaml:"vim_mode" json:"vim_mode"`

	// Engine settings
	DefaultEngine string `yaml:"default_engine" json:"default_engine"`
	AutoSwitch    bool   `yaml:"auto_switch" json:"auto_switch"`
}

// RunnerConfig holds spell runner configuration
type RunnerConfig struct {
	// Default behavior
	DefaultTimeout     time.Duration `yaml:"default_timeout" json:"default_timeout"`
	DefaultMemoryLimit int64         `yaml:"default_memory_limit" json:"default_memory_limit"`

	// Spell discovery
	SpellPaths      []string      `yaml:"spell_paths" json:"spell_paths"`
	AutoDiscovery   bool          `yaml:"auto_discovery" json:"auto_discovery"`
	CacheEnabled    bool          `yaml:"cache_enabled" json:"cache_enabled"`
	CacheExpiration time.Duration `yaml:"cache_expiration" json:"cache_expiration"`

	// Execution settings
	ParallelExecution bool          `yaml:"parallel_execution" json:"parallel_execution"`
	MaxParallelSpells int           `yaml:"max_parallel_spells" json:"max_parallel_spells"`
	RetryAttempts     int           `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay        time.Duration `yaml:"retry_delay" json:"retry_delay"`

	// Output settings
	CaptureOutput   bool   `yaml:"capture_output" json:"capture_output"`
	OutputFormat    string `yaml:"output_format" json:"output_format"`
	ErrorFormat     string `yaml:"error_format" json:"error_format"`
	TimestampOutput bool   `yaml:"timestamp_output" json:"timestamp_output"`

	// Signal handling
	GracefulShutdown  bool          `yaml:"graceful_shutdown" json:"graceful_shutdown"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
	InterruptBehavior string        `yaml:"interrupt_behavior" json:"interrupt_behavior"`
	CleanupTempFiles  bool          `yaml:"cleanup_temp_files" json:"cleanup_temp_files"`
}

// TemplateConfig holds template-related configuration
type TemplateConfig struct {
	// Template directories
	BuiltinPath string   `yaml:"builtin_path" json:"builtin_path"`
	UserPaths   []string `yaml:"user_paths" json:"user_paths"`

	// Template settings
	DefaultAuthor  string `yaml:"default_author" json:"default_author" env:"LLMSPELL_TEMPLATE_AUTHOR"`
	DefaultLicense string `yaml:"default_license" json:"default_license" env:"LLMSPELL_TEMPLATE_LICENSE"`

	// Generation settings
	OverwriteExisting bool `yaml:"overwrite_existing" json:"overwrite_existing"`
	CreateDirectories bool `yaml:"create_directories" json:"create_directories"`
	ValidateOnCreate  bool `yaml:"validate_on_create" json:"validate_on_create"`
}

// GetDefaultConfig returns a configuration with sensible defaults
func GetDefaultConfig() *Config {
	return &Config{
		Version: "0.1.0",
		Debug:   false,
		Quiet:   false,
		Verbose: false,

		Engine: EngineConfig{
			Default:        "lua",
			MemoryLimit:    64 * 1024 * 1024, // 64MB
			TimeoutLimit:   30 * time.Second,
			GoroutineLimit: 100,

			Lua: LuaEngineConfig{
				PoolMinSize:      2,
				PoolMaxSize:      10,
				PoolIdleTimeout:  5 * time.Minute,
				HealthThreshold:  0.8,
				CleanupInterval:  1 * time.Minute,
				CacheEnabled:     true,
				CompileOptimized: true,
				StdlibModules: []string{
					"core", "promise", "llm", "agent", "state",
					"events", "tools", "data", "errors", "logging",
				},
			},

			JavaScript: JavaScriptEngineConfig{
				ESVersion:     "ES2022",
				Strict:        true,
				ModuleSupport: true,
			},

			Tengo: TengoEngineConfig{
				MaxAllocs:   1000000,
				ImportLimit: 100,
			},
		},

		Security: SecurityConfig{
			Profile:           "sandbox",
			FileSystemMode:    engine.FSModeSandbox,
			AllowedModules:    []string{"string", "table", "math", "utf8"},
			DisabledModules:   []string{"io", "os", "debug", "package"},
			EnableValidation:  true,
			StrictValidation:  false,
			SecurityWarnings:  true,
			PerformanceChecks: true,
			NetworkAccess:     false,
			FileAccess:        false,
			ProcessAccess:     false,
			AllowedDomains:    []string{},
			AllowedPaths:      []string{},
			DeniedPatterns:    []string{"%.%.%.", "/etc/", "/proc/", "/sys/"},
			MaxFileSize:       10 * 1024 * 1024, // 10MB
			MaxNetworkCalls:   0,
		},

		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
			FileSettings: LogFileSettings{
				MaxSize:    100, // 100MB
				MaxBackups: 3,
				MaxAge:     30, // 30 days
				Compress:   true,
			},
			Components: ComponentLogging{
				Engine:   "info",
				Bridge:   "info",
				Runner:   "info",
				Security: "warn",
				REPL:     "info",
			},
		},

		CLI: CLIConfig{
			ColorOutput:      true,
			NoColor:          false,
			Theme:            "default",
			ShowProgress:     true,
			ProgressStyle:    "bar",
			PagerCommand:     "less",
			Editor:           "vim",
			EnableCompletion: true,
			CompletionShell:  "bash",
		},

		REPL: REPLConfig{
			HistoryFile:     "~/.llmspell_history",
			HistorySize:     1000,
			SaveHistory:     true,
			HistoryDuration: "30d",
			Prompt:          "llmspell> ",
			ContinuePrompt:  "... ",
			SyntaxHighlight: true,
			AutoComplete:    true,
			MultiLine:       true,
			AutoIndent:      true,
			BracketMatch:    true,
			VimMode:         false,
			DefaultEngine:   "lua",
			AutoSwitch:      true,
		},

		Runner: RunnerConfig{
			DefaultTimeout:     30 * time.Second,
			DefaultMemoryLimit: 64 * 1024 * 1024, // 64MB
			SpellPaths:         []string{".", "./spells", "~/.llmspell/spells"},
			AutoDiscovery:      true,
			CacheEnabled:       true,
			CacheExpiration:    1 * time.Hour,
			ParallelExecution:  false,
			MaxParallelSpells:  4,
			RetryAttempts:      3,
			RetryDelay:         1 * time.Second,
			CaptureOutput:      true,
			OutputFormat:       "auto",
			ErrorFormat:        "detailed",
			TimestampOutput:    false,
			GracefulShutdown:   true,
			ShutdownTimeout:    10 * time.Second,
			InterruptBehavior:  "graceful",
			CleanupTempFiles:   true,
		},

		Templates: TemplateConfig{
			BuiltinPath:       "templates",
			UserPaths:         []string{"~/.llmspell/templates", "./templates"},
			DefaultAuthor:     "",
			DefaultLicense:    "MIT",
			OverwriteExisting: false,
			CreateDirectories: true,
			ValidateOnCreate:  true,
		},
	}
}

// Validate checks the configuration for consistency and required values
func (c *Config) Validate() error {
	// TODO: Implement comprehensive validation
	// This will be implemented as part of the config validation task
	return nil
}

// GetEngineConfig returns the engine configuration for the specified engine
func (c *Config) GetEngineConfig(engineName string) engine.EngineConfig {
	base := engine.EngineConfig{
		MemoryLimit:     c.Engine.MemoryLimit,
		TimeoutLimit:    c.Engine.TimeoutLimit,
		GoroutineLimit:  c.Engine.GoroutineLimit,
		SandboxMode:     c.Security.Profile != "development",
		AllowedModules:  c.Security.AllowedModules,
		DisabledModules: c.Security.DisabledModules,
		FileSystemMode:  c.Security.FileSystemMode,
		DebugMode:       c.Debug,
		LogLevel:        c.Logging.Level,
		MetricsMode:     c.Debug || c.Verbose,
		TracingMode:     c.Debug,
		EngineOptions:   make(map[string]interface{}),
	}

	// Add engine-specific options
	switch engineName {
	case "lua":
		base.EngineOptions["pool_min_size"] = c.Engine.Lua.PoolMinSize
		base.EngineOptions["pool_max_size"] = c.Engine.Lua.PoolMaxSize
		base.EngineOptions["pool_idle_timeout"] = c.Engine.Lua.PoolIdleTimeout.String()
		base.EngineOptions["health_threshold"] = c.Engine.Lua.HealthThreshold
		base.EngineOptions["cleanup_interval"] = c.Engine.Lua.CleanupInterval.String()
		base.EngineOptions["cache_enabled"] = c.Engine.Lua.CacheEnabled
		base.EngineOptions["compile_optimized"] = c.Engine.Lua.CompileOptimized
		base.EngineOptions["stdlib_modules"] = c.Engine.Lua.StdlibModules

	case "javascript":
		base.EngineOptions["es_version"] = c.Engine.JavaScript.ESVersion
		base.EngineOptions["strict"] = c.Engine.JavaScript.Strict
		base.EngineOptions["module_support"] = c.Engine.JavaScript.ModuleSupport

	case "tengo":
		base.EngineOptions["max_allocs"] = c.Engine.Tengo.MaxAllocs
		base.EngineOptions["import_limit"] = c.Engine.Tengo.ImportLimit
	}

	return base
}
