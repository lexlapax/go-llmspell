// ABOUTME: Main entry point for the llmspell CLI using Kong for command parsing.
// ABOUTME: Provides spell execution, validation, REPL, and management commands.

// Package main implements the llmspell command-line interface.
// It provides commands for executing LLM spell scripts, managing configurations,
// and interacting with script engines through a unified CLI experience.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/lexlapax/go-llmspell/cmd/llmspell/commands"
	"github.com/lexlapax/go-llmspell/pkg/config"
	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// Version information set during build
var (
	// version is the semantic version of the build
	version = "dev"
	// buildDate is the ISO8601 date of the build
	buildDate = ""
	// gitCommit is the git commit hash of the build
	gitCommit = ""
)

// CLI represents the command-line interface structure.
// It defines global flags and available commands using Kong tags
// for automatic CLI parsing and help generation.
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

// main is the entry point for the llmspell CLI.
// It sets up signal handling, parses command-line arguments using Kong,
// loads configuration, and executes the appropriate command.
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

	// Create runner configuration
	runnerConfig := runner.DefaultRunnerConfig()
	runnerConfig.EnableDebug = cli.DebugMode
	runnerConfig.EnableMetrics = true

	// Setup engine registry with bridges
	engineManager, err := runner.SetupEngineRegistry(runnerConfig, cli.Profile)
	if err != nil {
		parser.Fatalf("failed to setup engine registry: %v", err)
		osExit(1)
		return
	}

	// Create engine selector
	selector := runner.NewEngineSelector(engineManager)

	// Create script executor with proper architecture
	scriptRunner := runner.NewScriptExecutor(runnerConfig, engineManager, selector)

	// Create command context
	cmdCtx := createCommandContext(ctx, cfg, cli, scriptRunner)

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

// loadConfig loads configuration from file or defaults.
// It uses the config loader with environment variable support
// and returns default configuration if loading fails.
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

// formatVersion formats version information.
// It combines version, git commit hash, and build date
// into a human-readable string for display.
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

// createCommandContext creates context for command execution.
// It enriches the context with configuration, flags, and runner
// information needed by command implementations.
func createCommandContext(ctx context.Context, cfg *config.Config, cli *CLI, scriptRunner runner.Runner) context.Context {
	ctx = context.WithValue(ctx, commands.ConfigKey, cfg)
	ctx = context.WithValue(ctx, commands.DebugKey, cli.DebugMode)
	ctx = context.WithValue(ctx, commands.VerboseKey, cli.Verbose)
	ctx = context.WithValue(ctx, commands.ProfileKey, cli.Profile)
	ctx = context.WithValue(ctx, commands.RunnerKey, scriptRunner)
	return ctx
}

// setupErrorHandler sets up error handling.
// It initializes the global error handler with debug and interactive
// settings from the configuration.
func setupErrorHandler(cfg *config.Config) *errors.ErrorHandler {
	// Initialize global error handler
	errors.InitializeErrorHandler(cfg.Debug, true)
	return errors.GetErrorHandler()
}
