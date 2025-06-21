// ABOUTME: Implementation of the version command for showing version information.
// ABOUTME: Displays version, build date, and git commit information.

package commands

import (
	"context"
)

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

// Run executes the command
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
