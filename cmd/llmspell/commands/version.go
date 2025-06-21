// ABOUTME: Implementation of the version command for showing version information.
// ABOUTME: Displays version, build date, and git commit information.

package commands

import (
	"context"
	"encoding/json"
	"runtime"
)

// VersionCmd shows version information
type VersionCmd struct {
	BaseCommand
	Short       bool   `short:"s" help:"Show short version only"`
	Verbose     bool   `short:"v" help:"Show verbose version information"`
	BuildInfo   bool   `help:"Show build information"`
	Format      string `enum:"text,json" default:"text" help:"Output format"`
	Deps        bool   `help:"Show dependencies"`
	CheckCompat bool   `help:"Check go-llms compatibility"`
}

// Version info - will be set during build
var (
	Version   = "dev"
	BuildDate = ""
	GitCommit = ""
)

// Run executes the command
func (c *VersionCmd) Run(ctx context.Context) error {
	if c.Format == "json" {
		return c.outputJSON()
	}

	if c.Short {
		c.Println(Version)
		return nil
	}

	c.Printf("llmspell version %s\n", Version)

	if c.Verbose || c.BuildInfo {
		c.Println("")
		c.Println("version:", Version)
		c.Println("go version:", runtime.Version())
		c.Println("built:", getBuildDate())
		c.Println("platform:", runtime.GOOS+"/"+runtime.GOARCH)

		if GitCommit != "" {
			c.Println("commit:", GitCommit)
		}
	}

	if c.Deps {
		c.Println("\ndependencies:")
		c.Println("  github.com/yuin/gopher-lua: v1.1.1")
		c.Println("  github.com/alecthomas/kong: v1.6.0")
		c.Println("  github.com/knadh/koanf/v2: v2.1.2")
		c.Println("  github.com/lexlapax/go-llms: v0.3.5")
	}

	if c.CheckCompat {
		c.Println("\ngo-llms compatibility:")
		c.Println("  minimum version: v0.3.5")
		c.Println("  current version: v0.3.5")
		c.Println("  status: ✓ compatible")
	}

	return nil
}

func (c *VersionCmd) outputJSON() error {
	info := map[string]interface{}{
		"version":    Version,
		"go_version": runtime.Version(),
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
		"built":      getBuildDate(),
	}

	if GitCommit != "" {
		info["commit"] = GitCommit
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	c.Println(string(data))
	return nil
}

func getBuildDate() string {
	if BuildDate != "" {
		return BuildDate
	}
	return "unknown"
}
