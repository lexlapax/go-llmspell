// ABOUTME: Main entry point for the llmspell CLI using Kong for command parsing.
// ABOUTME: Provides spell execution, validation, REPL, and management commands.

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/lexlapax/go-llmspell/cmd/llmspell/commands"
	"github.com/lexlapax/go-llmspell/pkg/config"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// Version information set during build
var (
	version   = "dev"
	buildDate = ""
	gitCommit = ""
)

// CLI represents the command-line interface structure
type CLI struct {
	// Global flags
	DebugMode  bool   `help:"Enable debug mode" env:"LLMSPELL_DEBUG" name:"debug"`
	ConfigFile string `help:"Config file path" type:"path" env:"LLMSPELL_CONFIG" name:"config"`
	Quiet      bool   `help:"Suppress non-error output" short:"q"`
	Verbose    bool   `help:"Enable verbose output" short:"v"`
	Profile    string `help:"Security profile to use" default:"sandbox" enum:"sandbox,development,production"`

	// Commands
	Run        commands.RunCmd        `cmd:"" help:"Execute a spell script"`
	Validate   commands.ValidateCmd   `cmd:"" help:"Validate a spell or script"`
	Engines    commands.EnginesCmd    `cmd:"" help:"List available script engines"`
	Version    commands.VersionCmd    `cmd:"" help:"Show version information"`
	Config     commands.ConfigCmd     `cmd:"" help:"Manage configuration"`
	Security   commands.SecurityCmd   `cmd:"" help:"Manage security profiles"`
	REPL       commands.REPLCmd       `cmd:"" help:"Start interactive REPL"`
	Debug      commands.DebugCmd      `cmd:"" help:"Debug a spell script"`
	New        commands.NewCmd        `cmd:"" help:"Create a new spell from a template"`
	Completion commands.CompletionCmd `cmd:"" help:"Generate shell completion script"`
	Man        commands.ManCmd        `cmd:"" help:"Generate man pages"`
	GenDocs    commands.GenDocsCmd    `cmd:"" help:"Generate Lua API documentation"`
}

// osExit allows testing of exit behavior
var osExit = os.Exit

func main() {
	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Parse CLI
	cli := &CLI{}
	parser, err := kong.New(cli,
		kong.Name("llmspell"),
		kong.Description("Scriptable LLM interactions via Lua, JavaScript, and Tengo\n\n"+
			"Examples:\n"+
			"  llmspell run hello.lua                    # Run a spell script\n"+
			"  llmspell repl                             # Start interactive REPL\n"+
			"  llmspell new myspell --type agent         # Create new spell from template\n"+
			"  llmspell validate script.lua              # Validate script syntax\n"+
			"  llmspell completion bash                  # Generate shell completions\n\n"+
			"For more help on a command, use: llmspell <command> --help"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{
			"version": formatVersion(),
		},
	)
	if err != nil {
		panic(err)
	}

	kongCtx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
		osExit(1)
		return
	}

	// Load configuration
	cfg := loadConfig(cli.ConfigFile)

	// Apply CLI flags to config
	if cli.DebugMode {
		cfg.Debug = true
	}

	// Create engine registry with default configuration
	registryConfig := engine.RegistryConfig{
		MaxEngines:        10,
		DefaultTimeout:    30 * time.Second,
		HealthCheckPeriod: 60 * time.Second,
		PoolingEnabled:    true,
		MaxPoolSize:       5,
		IdleTimeout:       10 * time.Minute,
		MetricsEnabled:    true,
		LoggingEnabled:    cli.Verbose,
		TracingEnabled:    cli.DebugMode,
	}
	registry := engine.NewRegistry(registryConfig)

	// Initialize registry first
	if err := registry.Initialize(); err != nil {
		parser.Fatalf("failed to initialize engine registry: %v", err)
		osExit(1)
		return
	}

	// Register Lua engine factory
	luaFactory := gopherlua.NewLuaEngineFactory()
	if err := registry.Register(luaFactory); err != nil {
		parser.Fatalf("failed to register Lua engine: %v", err)
		osExit(1)
		return
	}

	// Create engine registry manager for runner
	engineRegistry := runner.NewEngineRegistryManager(registry)

	// TODO: Register JavaScript and Tengo engines when implemented

	// Create command context
	cmdCtx := createCommandContext(ctx, cfg, cli, engineRegistry)

	// Set up error handler
	errorHandler := setupErrorHandler(cfg)

	// Execute command
	kongCtx.BindTo(cmdCtx, (*context.Context)(nil))
	if err := kongCtx.Run(); err != nil {
		errorHandler.Handle(err)
		osExit(1)
		return
	}
}

// loadConfig loads configuration from file or defaults
func loadConfig(configPath string) *config.Config {
	// Set up loader options
	options := config.LoaderOptions{
		ConfigFile:     configPath,
		EnvPrefix:      "LLMSPELL",
		EnvDelimiter:   "_",
		ValidateOnLoad: true,
	}

	// Use the config loader
	loader := config.NewLoader(options)

	// Load configuration
	cfg, err := loader.LoadConfig()
	if err != nil {
		// Return default config on error
		return config.GetDefaultConfig()
	}

	return cfg
}

// formatVersion formats version information
func formatVersion() string {
	v := version
	if gitCommit != "" {
		v += " (" + gitCommit[:7] + ")"
	}
	if buildDate != "" {
		v += " built " + buildDate
	}
	return v
}

// createCommandContext creates context for command execution
func createCommandContext(ctx context.Context, cfg *config.Config, cli *CLI, engineRegistry *runner.EngineRegistryManager) context.Context {
	ctx = context.WithValue(ctx, commands.ConfigKey, cfg)
	ctx = context.WithValue(ctx, commands.DebugKey, cli.DebugMode)
	ctx = context.WithValue(ctx, commands.VerboseKey, cli.Verbose)
	ctx = context.WithValue(ctx, commands.ProfileKey, cli.Profile)
	ctx = context.WithValue(ctx, commands.EngineRegistryKey, engineRegistry)
	return ctx
}

// setupErrorHandler sets up error handling
func setupErrorHandler(cfg *config.Config) *errors.ErrorHandler {
	// Initialize global error handler
	errors.InitializeErrorHandler(cfg.Debug, true)
	return errors.GetErrorHandler()
}
