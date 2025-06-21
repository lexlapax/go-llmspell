// ABOUTME: Implementation of the engines command for listing available script engines.
// ABOUTME: Shows engine information including features, extensions, and statistics.

package commands

import (
	"context"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// EnginesCmd lists available script engines
type EnginesCmd struct {
	BaseCommand
	Details bool `short:"d" help:"Show detailed engine information"`
}

// Run executes the command
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
			c.Println()

			// Show supported features
			if len(info.Features) > 0 {
				c.Printf("    Features: ")
				for i, feature := range info.Features {
					if i > 0 {
						c.Printf(", ")
					}
					c.Printf("%s", feature)
				}
				c.Println()
			}

			// Show file extensions
			if len(info.FileExtensions) > 0 {
				c.Printf("    Extensions: ")
				for i, ext := range info.FileExtensions {
					if i > 0 {
						c.Printf(", ")
					}
					c.Printf("%s", ext)
				}
				c.Println()
			}
		} else {
			c.Println()
		}
	}

	// Show stats if available and verbose
	if IsVerbose(ctx) && c.Details {
		stats := engineRegistry.GetStats()
		if len(stats) > 0 {
			c.Println("\nEngine Statistics:")
			for name, stat := range stats {
				c.Printf("  %s:\n", name)
				c.Printf("    Executions: %d (Success: %d, Errors: %d)\n",
					stat.SuccessCount+stat.ErrorCount, stat.SuccessCount, stat.ErrorCount)
				if stat.TotalExecTime > 0 && stat.SuccessCount > 0 {
					avgTime := stat.TotalExecTime / time.Duration(stat.SuccessCount)
					c.Printf("    Avg execution time: %v\n", avgTime)
				}
			}
		}
	}

	return nil
}
