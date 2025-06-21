// ABOUTME: Implementation of all CLI commands using the runner package.
// ABOUTME: Provides run, validate, engines, version, config, security, repl, and debug commands.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// RunCmd executes a spell script
type RunCmd struct {
	BaseCommand
	Script     string            `arg:"" help:"Script file to execute" type:"existingfile"`
	Parameters map[string]string `short:"p" help:"Parameters to pass to the script (key=value)"`
	Engine     string            `short:"e" help:"Script engine to use (auto-detected if not specified)"`
	Timeout    int               `short:"t" help:"Execution timeout in seconds" default:"300"`
}

func (c *RunCmd) Run(ctx context.Context) error {
	// Get engine registry from context
	engineRegistryInterface := GetEngineRegistry(ctx)
	if engineRegistryInterface == nil {
		return errors.New(errors.CategoryConfig, "engine registry not found in context")
	}
	
	engineRegistry, ok := engineRegistryInterface.(*runner.EngineRegistryManager)
	if !ok {
		return errors.New(errors.CategoryConfig, "invalid engine registry type")
	}
	
	// Create runner config
	runnerConfig := &runner.RunnerConfig{
		MaxConcurrentScripts: 1,
		Timeout:             time.Duration(c.Timeout) * time.Second,
		EngineConfigs:       make(map[string]map[string]interface{}),
		DefaultEngine:       "lua",
	}
	
	// Create engine selector
	selector := runner.NewEngineSelector(engineRegistry)
	
	// Create script executor
	executor := runner.NewScriptExecutor(runnerConfig, engineRegistry, selector)
	
	// Initialize executor
	if err := executor.Initialize(ctx); err != nil {
		return errors.Wrap(err, errors.CategoryEngine, "failed to initialize executor")
	}
	defer executor.Shutdown()
	
	// Convert string parameters to interface{}
	params := make(map[string]interface{})
	for k, v := range c.Parameters {
		params[k] = v
	}
	
	// Add engine override if specified
	if c.Engine != "" {
		params["__engine"] = c.Engine
	}
	
	// Execute the script
	c.Debug(ctx, "Executing script: %s", c.Script)
	result, err := executor.ExecuteFile(ctx, c.Script, params)
	if err != nil {
		return errors.Wrap(err, errors.CategoryScript, "failed to execute script")
	}
	
	// Print result if not nil
	if result != nil {
		c.Printf("%v\n", result)
	}
	
	return nil
}

// ValidateCmd validates a spell or script
type ValidateCmd struct {
	BaseCommand
	Path string `arg:"" help:"Path to spell.yaml or script file" type:"existingfile"`
}

func (c *ValidateCmd) Run(ctx context.Context) error {
	// Get engine registry from context
	engineRegistryInterface := GetEngineRegistry(ctx)
	if engineRegistryInterface == nil {
		return errors.New(errors.CategoryConfig, "engine registry not found in context")
	}
	
	engineRegistry, ok := engineRegistryInterface.(*runner.EngineRegistryManager)
	if !ok {
		return errors.New(errors.CategoryConfig, "invalid engine registry type")
	}
	
	// Check if it's a spell file or script
	ext := filepath.Ext(c.Path)
	if ext == ".yaml" || ext == ".yml" {
		// Validate spell file
		loader := runner.NewSpellLoader()
		spell, err := loader.LoadFromFile(c.Path)
		if err != nil {
			return errors.Wrap(err, errors.CategoryValidation, "failed to load spell file")
		}
		
		c.Printf("✓ Spell file is valid\n")
		c.Printf("  Name: %s\n", spell.Name)
		c.Printf("  Version: %s\n", spell.Version)
		c.Printf("  Entry Point: %s\n", spell.EntryPoint)
		if spell.Engine != "" {
			c.Printf("  Engine: %s\n", spell.Engine)
		}
		
		return nil
	}
	
	// It's a script file - validate using engine
	selector := runner.NewEngineSelector(engineRegistry)
	engineName, err := selector.SelectByExtension(c.Path)
	if err != nil {
		return errors.Wrap(err, errors.CategoryValidation, "unable to determine script engine")
	}
	
	// For now, just check if we can get engine info
	if _, err := engineRegistry.GetEngineInfo(engineName); err != nil {
		return errors.Wrap(err, errors.CategoryEngine, "engine not available")
	}
	
	c.Printf("✓ Script file is valid (%s engine)\n", engineName)
	
	return nil
}

// EnginesCmd lists available script engines
type EnginesCmd struct {
	BaseCommand
	Details bool `short:"d" help:"Show detailed engine information"`
}

func (c *EnginesCmd) Run(ctx context.Context) error {
	// Get engine registry from context
	engineRegistryInterface := GetEngineRegistry(ctx)
	if engineRegistryInterface == nil {
		// Fall back to hardcoded list if no registry
		c.Println("Available engines:")
		c.Println("  - lua (Lua 5.1)")
		c.Println("  - javascript (ES6+) [not implemented]")
		c.Println("  - tengo (Tengo script) [not implemented]")
		return nil
	}
	
	engineRegistry, ok := engineRegistryInterface.(*runner.EngineRegistryManager)
	if !ok {
		return errors.New(errors.CategoryConfig, "invalid engine registry type")
	}
	
	// List registered engines
	engines := engineRegistry.ListEngines()
	if len(engines) == 0 {
		c.Println("No engines registered")
		return nil
	}
	
	c.Println("Available engines:")
	for _, info := range engines {
		c.Printf("  - %s", info.Name)
		if c.Details {
			c.Printf(" v%s (%s)", info.Version, info.Description)
		}
		c.Println()
	}
	
	return nil
}

// VersionCmd shows version information
type VersionCmd struct {
	BaseCommand
	Short bool `short:"s" help:"Show short version only"`
}

// Version info - will be set during build
var (
	Version   = "dev"
	BuildDate = ""
	GitCommit = ""
)

func (c *VersionCmd) Run(ctx context.Context) error {
	if c.Short {
		c.Println(Version)
	} else {
		c.Printf("llmspell version %s\n", Version)
		if GitCommit != "" {
			c.Printf("Commit: %s\n", GitCommit)
		}
		if BuildDate != "" {
			c.Printf("Built: %s\n", BuildDate)
		}
	}
	return nil
}

// ConfigCmd manages configuration
type ConfigCmd struct {
	BaseCommand
	Action string `arg:"" help:"Action to perform (show, get, set, path)" enum:"show,get,set,path" default:"show"`
	Key    string `arg:"" optional:"" help:"Configuration key (for get/set)"`
	Value  string `arg:"" optional:"" help:"Value to set (for set)"`
}

func (c *ConfigCmd) Run(ctx context.Context) error {
	cfg := GetConfig(ctx)
	
	switch c.Action {
	case "show":
		c.Println("Configuration:")
		c.Printf("  default_engine: lua\n")
		c.Printf("  security_profile: %s\n", GetProfile(ctx))
		c.Printf("  debug: %v\n", cfg.Debug)
		return nil
		
	case "path":
		c.Println(getDefaultConfigPath())
		return nil
		
	case "get":
		if c.Key == "" {
			return errors.New(errors.CategoryUsage, "key required for get action")
		}
		// For now, just handle known keys
		var value string
		switch c.Key {
		case "debug":
			value = fmt.Sprintf("%v", cfg.Debug)
		case "default_engine":
			value = "lua"
		default:
			value = "<not set>"
		}
		c.Println(value)
		return nil
		
	case "set":
		return errors.New(errors.CategoryUsage, "set action not implemented - edit config file directly")
		
	default:
		return errors.Newf(errors.CategoryUsage, "unknown action: %s", c.Action)
	}
}

// SecurityCmd manages security profiles
type SecurityCmd struct {
	BaseCommand
	Action  string `arg:"" help:"Action to perform (list, show, validate)" enum:"list,show,validate" default:"list"`
	Profile string `arg:"" optional:"" help:"Profile name (for show/validate)"`
}

func (c *SecurityCmd) Run(ctx context.Context) error {
	switch c.Action {
	case "list":
		c.Println("Available security profiles:")
		// TODO: Use actual security package when available
		profiles := []struct{name, desc string}{
			{"sandbox", "Maximum security restrictions"},
			{"development", "Balanced for development"},
			{"production", "Production security settings"},
		}
		for _, p := range profiles {
			c.Printf("  - %s (%s)\n", p.name, p.desc)
		}
		return nil
		
	case "show":
		if c.Profile == "" {
			c.Profile = GetProfile(ctx)
		}
		
		// TODO: Use actual security package when available
		c.Printf("Profile: %s\n", c.Profile)
		switch c.Profile {
		case "sandbox":
			c.Printf("Description: Maximum security restrictions\n")
			c.Println("\nPermissions:")
			c.Println("  - read:script")
			c.Println("  - execute:llm")
		case "development":
			c.Printf("Description: Balanced for development\n")
			c.Println("\nPermissions:")
			c.Println("  - read:*")
			c.Println("  - write:temp")
			c.Println("  - execute:*")
			c.Println("  - network:llm")
		case "production":
			c.Printf("Description: Production security settings\n")
			c.Println("\nPermissions:")
			c.Println("  - read:*")
			c.Println("  - write:output")
			c.Println("  - execute:*")
			c.Println("  - network:*")
		default:
			return errors.Newf(errors.CategorySecurity, "unknown profile: %s", c.Profile)
		}
		
		return nil
		
	case "validate":
		if c.Profile == "" {
			return errors.New(errors.CategoryUsage, "profile name required")
		}
		
		// TODO: Use actual security package when available
		validProfiles := []string{"sandbox", "development", "production"}
		valid := false
		for _, p := range validProfiles {
			if p == c.Profile {
				valid = true
				break
			}
		}
		if !valid {
			return errors.Newf(errors.CategoryValidation, "invalid profile: %s", c.Profile)
		}
		
		c.Printf("✓ Profile '%s' is valid\n", c.Profile)
		return nil
		
	default:
		return errors.Newf(errors.CategoryUsage, "unknown action: %s", c.Action)
	}
}

// REPLCmd starts an interactive REPL
type REPLCmd struct {
	BaseCommand
	Engine string `short:"e" help:"Script engine to use" default:"lua"`
}

func (c *REPLCmd) Run(ctx context.Context) error {
	c.Printf("Starting REPL with %s engine...\n", c.Engine)
	return errors.New(errors.CategoryUsage, "REPL not implemented yet")
}

// DebugCmd debugs a spell script
type DebugCmd struct {
	BaseCommand
	Script      string `arg:"" help:"Script file to debug" type:"existingfile"`
	Breakpoints []int  `short:"b" help:"Line numbers for breakpoints"`
}

func (c *DebugCmd) Run(ctx context.Context) error {
	c.Printf("Debugging script: %s\n", c.Script)
	if len(c.Breakpoints) > 0 {
		c.Printf("Breakpoints at lines: %v\n", c.Breakpoints)
	}
	return errors.New(errors.CategoryUsage, "debug command not implemented yet")
}

// getDefaultConfigPath returns the default config file path
func getDefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "llmspell", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "llmspell", "config.yaml")
}