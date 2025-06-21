// ABOUTME: Tests for the configuration package, covering default config generation and validation.
// ABOUTME: Ensures configuration struct integrity and proper default value handling.

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	t.Run("basic_settings", func(t *testing.T) {
		assert.Equal(t, "0.1.0", config.Version)
		assert.False(t, config.Debug)
		assert.False(t, config.Quiet)
		assert.False(t, config.Verbose)
	})

	t.Run("engine_settings", func(t *testing.T) {
		assert.Equal(t, "lua", config.Engine.Default)
		assert.Equal(t, int64(64*1024*1024), config.Engine.MemoryLimit)
		assert.Equal(t, 30*time.Second, config.Engine.TimeoutLimit)
		assert.Equal(t, 100, config.Engine.GoroutineLimit)
	})

	t.Run("lua_engine_settings", func(t *testing.T) {
		lua := config.Engine.Lua
		assert.Equal(t, 2, lua.PoolMinSize)
		assert.Equal(t, 10, lua.PoolMaxSize)
		assert.Equal(t, 5*time.Minute, lua.PoolIdleTimeout)
		assert.Equal(t, 0.8, lua.HealthThreshold)
		assert.Equal(t, 1*time.Minute, lua.CleanupInterval)
		assert.True(t, lua.CacheEnabled)
		assert.True(t, lua.CompileOptimized)

		expectedModules := []string{
			"core", "promise", "llm", "agent", "state",
			"events", "tools", "data", "errors", "logging",
		}
		assert.Equal(t, expectedModules, lua.StdlibModules)
	})

	t.Run("javascript_engine_settings", func(t *testing.T) {
		js := config.Engine.JavaScript
		assert.Equal(t, "ES2022", js.ESVersion)
		assert.True(t, js.Strict)
		assert.True(t, js.ModuleSupport)
	})

	t.Run("tengo_engine_settings", func(t *testing.T) {
		tengo := config.Engine.Tengo
		assert.Equal(t, 1000000, tengo.MaxAllocs)
		assert.Equal(t, 100, tengo.ImportLimit)
	})

	t.Run("security_settings", func(t *testing.T) {
		security := config.Security
		assert.Equal(t, "sandbox", security.Profile)
		assert.Equal(t, engine.FSModeSandbox, security.FileSystemMode)
		assert.Equal(t, []string{"string", "table", "math", "utf8"}, security.AllowedModules)
		assert.Equal(t, []string{"io", "os", "debug", "package"}, security.DisabledModules)
		assert.True(t, security.EnableValidation)
		assert.False(t, security.StrictValidation)
		assert.True(t, security.SecurityWarnings)
		assert.True(t, security.PerformanceChecks)
		assert.False(t, security.NetworkAccess)
		assert.False(t, security.FileAccess)
		assert.False(t, security.ProcessAccess)
		assert.Empty(t, security.AllowedDomains)
		assert.Empty(t, security.AllowedPaths)
		assert.Contains(t, security.DeniedPatterns, "%.%.%.")
		assert.Equal(t, int64(10*1024*1024), security.MaxFileSize)
		assert.Equal(t, 0, security.MaxNetworkCalls)
	})

	t.Run("logging_settings", func(t *testing.T) {
		logging := config.Logging
		assert.Equal(t, "info", logging.Level)
		assert.Equal(t, "text", logging.Format)
		assert.Equal(t, "stdout", logging.Output)

		file := logging.FileSettings
		assert.Equal(t, 100, file.MaxSize)
		assert.Equal(t, 3, file.MaxBackups)
		assert.Equal(t, 30, file.MaxAge)
		assert.True(t, file.Compress)

		components := logging.Components
		assert.Equal(t, "info", components.Engine)
		assert.Equal(t, "info", components.Bridge)
		assert.Equal(t, "info", components.Runner)
		assert.Equal(t, "warn", components.Security)
		assert.Equal(t, "info", components.REPL)
	})

	t.Run("cli_settings", func(t *testing.T) {
		cli := config.CLI
		assert.True(t, cli.ColorOutput)
		assert.False(t, cli.NoColor)
		assert.Equal(t, "default", cli.Theme)
		assert.True(t, cli.ShowProgress)
		assert.Equal(t, "bar", cli.ProgressStyle)
		assert.Equal(t, "less", cli.PagerCommand)
		assert.Equal(t, "vim", cli.Editor)
		assert.True(t, cli.EnableCompletion)
		assert.Equal(t, "bash", cli.CompletionShell)
	})

	t.Run("repl_settings", func(t *testing.T) {
		repl := config.REPL
		assert.Equal(t, "~/.llmspell_history", repl.HistoryFile)
		assert.Equal(t, 1000, repl.HistorySize)
		assert.True(t, repl.SaveHistory)
		assert.Equal(t, "30d", repl.HistoryDuration)
		assert.Equal(t, "llmspell> ", repl.Prompt)
		assert.Equal(t, "... ", repl.ContinuePrompt)
		assert.True(t, repl.SyntaxHighlight)
		assert.True(t, repl.AutoComplete)
		assert.True(t, repl.MultiLine)
		assert.True(t, repl.AutoIndent)
		assert.True(t, repl.BracketMatch)
		assert.False(t, repl.VimMode)
		assert.Equal(t, "lua", repl.DefaultEngine)
		assert.True(t, repl.AutoSwitch)
	})

	t.Run("runner_settings", func(t *testing.T) {
		runner := config.Runner
		assert.Equal(t, 30*time.Second, runner.DefaultTimeout)
		assert.Equal(t, int64(64*1024*1024), runner.DefaultMemoryLimit)

		expectedPaths := []string{".", "./spells", "~/.llmspell/spells"}
		assert.Equal(t, expectedPaths, runner.SpellPaths)
		assert.True(t, runner.AutoDiscovery)
		assert.True(t, runner.CacheEnabled)
		assert.Equal(t, 1*time.Hour, runner.CacheExpiration)
		assert.False(t, runner.ParallelExecution)
		assert.Equal(t, 4, runner.MaxParallelSpells)
		assert.Equal(t, 3, runner.RetryAttempts)
		assert.Equal(t, 1*time.Second, runner.RetryDelay)
		assert.True(t, runner.CaptureOutput)
		assert.Equal(t, "auto", runner.OutputFormat)
		assert.Equal(t, "detailed", runner.ErrorFormat)
		assert.False(t, runner.TimestampOutput)
		assert.True(t, runner.GracefulShutdown)
		assert.Equal(t, 10*time.Second, runner.ShutdownTimeout)
		assert.Equal(t, "graceful", runner.InterruptBehavior)
		assert.True(t, runner.CleanupTempFiles)
	})

	t.Run("template_settings", func(t *testing.T) {
		templates := config.Templates
		assert.Equal(t, "templates", templates.BuiltinPath)

		expectedPaths := []string{"~/.llmspell/templates", "./templates"}
		assert.Equal(t, expectedPaths, templates.UserPaths)
		assert.Equal(t, "", templates.DefaultAuthor)
		assert.Equal(t, "MIT", templates.DefaultLicense)
		assert.False(t, templates.OverwriteExisting)
		assert.True(t, templates.CreateDirectories)
		assert.True(t, templates.ValidateOnCreate)
	})
}

func TestConfig_GetEngineConfig(t *testing.T) {
	config := GetDefaultConfig()

	t.Run("lua_engine_config", func(t *testing.T) {
		engineConfig := config.GetEngineConfig("lua")

		assert.Equal(t, config.Engine.MemoryLimit, engineConfig.MemoryLimit)
		assert.Equal(t, config.Engine.TimeoutLimit, engineConfig.TimeoutLimit)
		assert.Equal(t, config.Engine.GoroutineLimit, engineConfig.GoroutineLimit)
		assert.True(t, engineConfig.SandboxMode) // sandbox profile
		assert.Equal(t, config.Security.AllowedModules, engineConfig.AllowedModules)
		assert.Equal(t, config.Security.DisabledModules, engineConfig.DisabledModules)
		assert.Equal(t, config.Security.FileSystemMode, engineConfig.FileSystemMode)
		assert.Equal(t, config.Debug, engineConfig.DebugMode)
		assert.Equal(t, config.Logging.Level, engineConfig.LogLevel)
		assert.Equal(t, config.Debug || config.Verbose, engineConfig.MetricsMode)
		assert.Equal(t, config.Debug, engineConfig.TracingMode)

		// Check Lua-specific options
		assert.Equal(t, config.Engine.Lua.PoolMinSize, engineConfig.EngineOptions["pool_min_size"])
		assert.Equal(t, config.Engine.Lua.PoolMaxSize, engineConfig.EngineOptions["pool_max_size"])
		assert.Equal(t, config.Engine.Lua.PoolIdleTimeout.String(), engineConfig.EngineOptions["pool_idle_timeout"])
		assert.Equal(t, config.Engine.Lua.HealthThreshold, engineConfig.EngineOptions["health_threshold"])
		assert.Equal(t, config.Engine.Lua.CleanupInterval.String(), engineConfig.EngineOptions["cleanup_interval"])
		assert.Equal(t, config.Engine.Lua.CacheEnabled, engineConfig.EngineOptions["cache_enabled"])
		assert.Equal(t, config.Engine.Lua.CompileOptimized, engineConfig.EngineOptions["compile_optimized"])
		assert.Equal(t, config.Engine.Lua.StdlibModules, engineConfig.EngineOptions["stdlib_modules"])
	})

	t.Run("javascript_engine_config", func(t *testing.T) {
		engineConfig := config.GetEngineConfig("javascript")

		// Check JavaScript-specific options
		assert.Equal(t, config.Engine.JavaScript.ESVersion, engineConfig.EngineOptions["es_version"])
		assert.Equal(t, config.Engine.JavaScript.Strict, engineConfig.EngineOptions["strict"])
		assert.Equal(t, config.Engine.JavaScript.ModuleSupport, engineConfig.EngineOptions["module_support"])
	})

	t.Run("tengo_engine_config", func(t *testing.T) {
		engineConfig := config.GetEngineConfig("tengo")

		// Check Tengo-specific options
		assert.Equal(t, config.Engine.Tengo.MaxAllocs, engineConfig.EngineOptions["max_allocs"])
		assert.Equal(t, config.Engine.Tengo.ImportLimit, engineConfig.EngineOptions["import_limit"])
	})

	t.Run("development_profile", func(t *testing.T) {
		config.Security.Profile = "development"
		engineConfig := config.GetEngineConfig("lua")

		assert.False(t, engineConfig.SandboxMode) // development profile disables sandbox
	})

	t.Run("debug_mode_effects", func(t *testing.T) {
		config.Debug = true
		config.Verbose = false
		engineConfig := config.GetEngineConfig("lua")

		assert.True(t, engineConfig.DebugMode)
		assert.True(t, engineConfig.TracingMode)
		assert.True(t, engineConfig.MetricsMode)
	})

	t.Run("verbose_mode_effects", func(t *testing.T) {
		config.Debug = false
		config.Verbose = true
		engineConfig := config.GetEngineConfig("lua")

		assert.False(t, engineConfig.DebugMode)
		assert.False(t, engineConfig.TracingMode)
		assert.True(t, engineConfig.MetricsMode)
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("default_config_is_valid", func(t *testing.T) {
		config := GetDefaultConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})

	// TODO: Add more validation tests when Validate() is implemented
}

func TestConfigStructFields(t *testing.T) {
	t.Run("config_has_all_required_fields", func(t *testing.T) {
		config := &Config{}

		// Verify main sections exist
		assert.NotNil(t, &config.Engine)
		assert.NotNil(t, &config.Security)
		assert.NotNil(t, &config.Logging)
		assert.NotNil(t, &config.CLI)
		assert.NotNil(t, &config.REPL)
		assert.NotNil(t, &config.Runner)
		assert.NotNil(t, &config.Templates)
	})

	t.Run("engine_config_has_engine_specific_sections", func(t *testing.T) {
		config := &EngineConfig{}

		assert.NotNil(t, &config.Lua)
		assert.NotNil(t, &config.JavaScript)
		assert.NotNil(t, &config.Tengo)
	})

	t.Run("logging_config_has_all_sections", func(t *testing.T) {
		config := &LoggingConfig{}

		assert.NotNil(t, &config.FileSettings)
		assert.NotNil(t, &config.Components)
	})
}

func TestConfigDefaults(t *testing.T) {
	config := GetDefaultConfig()

	t.Run("sensible_memory_limits", func(t *testing.T) {
		// Memory limits should be reasonable for most systems
		assert.True(t, config.Engine.MemoryLimit > 0)
		assert.True(t, config.Engine.MemoryLimit <= 1024*1024*1024) // <= 1GB
		assert.True(t, config.Runner.DefaultMemoryLimit > 0)
		assert.True(t, config.Security.MaxFileSize > 0)
	})

	t.Run("sensible_timeout_limits", func(t *testing.T) {
		// Timeouts should be reasonable
		assert.True(t, config.Engine.TimeoutLimit > 0)
		assert.True(t, config.Engine.TimeoutLimit <= 5*time.Minute)
		assert.True(t, config.Runner.DefaultTimeout > 0)
		assert.True(t, config.Runner.ShutdownTimeout > 0)
	})

	t.Run("sensible_pool_settings", func(t *testing.T) {
		lua := config.Engine.Lua
		assert.True(t, lua.PoolMinSize >= 1)
		assert.True(t, lua.PoolMaxSize >= lua.PoolMinSize)
		assert.True(t, lua.PoolIdleTimeout > 0)
		assert.True(t, lua.HealthThreshold > 0 && lua.HealthThreshold <= 1)
	})

	t.Run("secure_defaults", func(t *testing.T) {
		security := config.Security
		assert.Equal(t, "sandbox", security.Profile)
		assert.False(t, security.NetworkAccess)
		assert.False(t, security.FileAccess)
		assert.False(t, security.ProcessAccess)
		assert.True(t, security.EnableValidation)
		assert.True(t, security.SecurityWarnings)
	})
}
