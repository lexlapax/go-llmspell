// ABOUTME: Implementation of the config command for managing configuration.
// ABOUTME: Supports show, get, set, and path actions for configuration management.

package commands

import (
	"context"

	"github.com/lexlapax/go-llmspell/pkg/config"
	"github.com/lexlapax/go-llmspell/pkg/errors"
)

// ConfigCmd manages configuration
type ConfigCmd struct {
	BaseCommand
	Action string `arg:"" help:"Action to perform: show (display all), get (retrieve value), set (update value), path (show config file location)" enum:"show,get,set,path" default:"show"`
	Key    string `arg:"" optional:"" help:"Configuration key (e.g., engine.default, repl.prompt)"`
	Value  string `arg:"" optional:"" help:"Value to set (required for 'set' action)"`
}

// Run executes the command
func (c *ConfigCmd) Run(ctx context.Context) error {
	cfg := GetConfig(ctx)

	switch c.Action {
	case "show":
		c.Println("Configuration:")
		c.Printf("  config_path: %s\n", getDefaultConfigPath())
		c.Printf("  debug: %v\n", cfg.Debug)
		c.Printf("  security_profile: %s\n", GetProfile(ctx))

		// Show engine settings
		c.Println("  engine:")
		c.Printf("    default: %s\n", cfg.Engine.Default)
		if cfg.Engine.MemoryLimit > 0 {
			c.Printf("    memory_limit: %d bytes\n", cfg.Engine.MemoryLimit)
		}
		if cfg.Engine.TimeoutLimit > 0 {
			c.Printf("    timeout_limit: %v\n", cfg.Engine.TimeoutLimit)
		}

		// Show security settings
		c.Println("  security:")
		c.Printf("    profile: %s\n", cfg.Security.Profile)
		if cfg.Security.FileSystemMode != "" {
			c.Printf("    filesystem_mode: %s\n", cfg.Security.FileSystemMode)
		}

		// Show runner settings if verbose
		if IsVerbose(ctx) {
			c.Println("  runner:")
			c.Printf("    default_timeout: %v\n", cfg.Runner.DefaultTimeout)
			if cfg.Runner.MaxParallelSpells > 0 {
				c.Printf("    max_parallel_spells: %d\n", cfg.Runner.MaxParallelSpells)
			}
		}

		return nil

	case "path":
		c.Println(getDefaultConfigPath())
		return nil

	case "get":
		if c.Key == "" {
			return errors.New(errors.CategoryUsage, "key required for get action")
		}

		// Use the config loader to get the actual value
		options := config.LoaderOptions{
			EnvPrefix:    "LLMSPELL",
			EnvDelimiter: "_",
		}
		loader := config.NewLoader(options)

		// Load config to get access to values
		_, err := loader.LoadConfig()
		if err != nil {
			return errors.Wrap(err, errors.CategoryConfig, "failed to load config")
		}

		// Use the loader to get the raw value by key
		value := loader.GetRaw(c.Key)

		if value == nil {
			c.Println("<not set>")
		} else {
			c.Printf("%v\n", value)
		}

		return nil

	case "set":
		if c.Key == "" {
			return errors.New(errors.CategoryUsage, "key required for set action")
		}
		if c.Value == "" {
			return errors.New(errors.CategoryUsage, "value required for set action")
		}

		// For safety, we'll still recommend editing the config file
		c.Printf("To set '%s' to '%s', edit the config file at:\n", c.Key, c.Value)
		c.Printf("  %s\n", getDefaultConfigPath())
		c.Println("\nExample:")
		c.Printf("  %s: %s\n", c.Key, c.Value)

		return nil

	default:
		return errors.Newf(errors.CategoryUsage, "unknown action: %s", c.Action)
	}
}
