// ABOUTME: Implementation of the man command for generating manual pages.
// ABOUTME: Creates UNIX man pages for llmspell and its subcommands.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lexlapax/go-llmspell/pkg/docs"
	"github.com/lexlapax/go-llmspell/pkg/errors"
)

// ManCmd generates man pages
type ManCmd struct {
	BaseCommand
	Command string `arg:"" optional:"" help:"Specific command to generate man page for (empty for main page)"`
	Output  string `short:"o" help:"Output file (default: stdout)"`
	Dir     string `short:"d" help:"Output directory for all man pages"`
	All     bool   `short:"a" help:"Generate all man pages"`
	Install bool   `short:"i" help:"Install to system man directory"`
	Format  string `help:"Output format" enum:"troff,text,html" default:"troff"`
}

// Run executes the command
func (c *ManCmd) Run(ctx context.Context) error {
	version := "dev" // This would come from build info in real implementation

	// Generate all man pages
	if c.All {
		return c.generateAll(ctx, version)
	}

	// Install man pages
	if c.Install {
		return c.installManPages(ctx, version)
	}

	// Generate specific man page
	var manPage *docs.ManPage
	var err error

	if c.Command == "" {
		// Generate main man page
		manPage = docs.GenerateLLMSpellManPage(version)
	} else {
		// Generate command-specific man page
		manPage, err = docs.GenerateCommandManPage(c.Command, version)
		if err != nil {
			return errors.Wrap(err, errors.CategoryUsage,
				"failed to generate man page for command").
				WithSuggestion("Use 'llmspell man --list' to see available commands")
		}
	}

	// Generate output
	output := manPage.Generate()

	// Handle different formats
	switch c.Format {
	case "troff":
		// Already in troff format
	case "text":
		// Convert to plain text (simplified for now)
		output = c.convertToText(output)
	case "html":
		// Convert to HTML (simplified for now)
		output = c.convertToHTML(output)
	}

	// Write output
	if c.Output != "" {
		return c.writeToFile(c.Output, output)
	}

	c.Println(output)
	return nil
}

// generateAll generates all man pages
func (c *ManCmd) generateAll(ctx context.Context, version string) error {
	if c.Dir == "" {
		c.Dir = "man"
	}

	// Create output directory
	if err := os.MkdirAll(c.Dir, 0755); err != nil {
		return errors.Wrap(err, errors.CategoryIO,
			"failed to create output directory")
	}

	// Generate main man page
	mainPage := docs.GenerateLLMSpellManPage(version)
	mainFile := filepath.Join(c.Dir, "llmspell.1")
	if err := c.writeToFile(mainFile, mainPage.Generate()); err != nil {
		return err
	}
	c.Info(ctx, "Generated %s", mainFile)

	// Generate command man pages
	for _, cmd := range docs.GetAllCommands() {
		cmdPage, err := docs.GenerateCommandManPage(cmd, version)
		if err != nil {
			c.Error(ctx, "Failed to generate man page for %s: %v", cmd, err)
			continue
		}

		cmdFile := filepath.Join(c.Dir, fmt.Sprintf("llmspell-%s.1", cmd))
		if err := c.writeToFile(cmdFile, cmdPage.Generate()); err != nil {
			return err
		}
		c.Info(ctx, "Generated %s", cmdFile)
	}

	c.Println("\nAll man pages generated successfully!")
	c.Printf("To view: man -l %s\n", filepath.Join(c.Dir, "llmspell.1"))

	return nil
}

// installManPages installs man pages to system directory
func (c *ManCmd) installManPages(ctx context.Context, version string) error {
	// Determine man directory
	manDir := c.findManDirectory()
	if manDir == "" {
		return errors.New(errors.CategoryIO,
			"could not find suitable man directory").
			WithSuggestion("Try specifying --dir or install manually")
	}

	// Check if we have write permission
	if err := c.checkWritePermission(manDir); err != nil {
		return errors.Wrap(err, errors.CategorySecurity,
			"cannot write to man directory").
			WithSuggestion("Try running with sudo or specify a different directory with --dir")
	}

	c.Info(ctx, "Installing man pages to %s", manDir)

	// Generate and install main page
	mainPage := docs.GenerateLLMSpellManPage(version)
	mainFile := filepath.Join(manDir, "llmspell.1")
	if err := c.writeToFile(mainFile, mainPage.Generate()); err != nil {
		return err
	}

	// Generate and install command pages
	for _, cmd := range docs.GetAllCommands() {
		cmdPage, err := docs.GenerateCommandManPage(cmd, version)
		if err != nil {
			continue
		}

		cmdFile := filepath.Join(manDir, fmt.Sprintf("llmspell-%s.1", cmd))
		if err := c.writeToFile(cmdFile, cmdPage.Generate()); err != nil {
			return err
		}
	}

	c.Println("\nMan pages installed successfully!")
	c.Println("You can now use: man llmspell")

	return nil
}

// findManDirectory finds appropriate man directory
func (c *ManCmd) findManDirectory() string {
	// Try common locations
	candidates := []string{
		"/usr/local/share/man/man1",
		"/usr/share/man/man1",
		filepath.Join(os.Getenv("HOME"), ".local/share/man/man1"),
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	// Try to create user directory
	userDir := filepath.Join(os.Getenv("HOME"), ".local/share/man/man1")
	if err := os.MkdirAll(userDir, 0755); err == nil {
		return userDir
	}

	return ""
}

// checkWritePermission checks if we can write to directory
func (c *ManCmd) checkWritePermission(dir string) error {
	testFile := filepath.Join(dir, ".llmspell-test")
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	f.Close()
	os.Remove(testFile)
	return nil
}

// writeToFile writes content to file
func (c *ManCmd) writeToFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, errors.CategoryIO,
			"failed to create directory")
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return errors.Wrap(err, errors.CategoryIO,
			"failed to write file")
	}

	return nil
}

// convertToText converts troff to plain text (simplified)
func (c *ManCmd) convertToText(troff string) string {
	// This is a very simplified conversion
	// In a real implementation, you'd use a proper troff parser
	text := troff

	// Remove common troff commands
	replacements := map[string]string{
		".TH":  "MANUAL PAGE:",
		".SH":  "\n",
		".SS":  "\n  ",
		".TP":  "\n  ",
		".PP":  "\n",
		".B":   "",
		".I":   "",
		".BR":  "",
		".RS":  "",
		".RE":  "",
		".nf":  "",
		".fi":  "",
		"\\fB": "",
		"\\fR": "",
		"\\fI": "",
		"\\-":  "-",
	}

	for old, new := range replacements {
		text = replaceAll(text, old, new)
	}

	return text
}

// convertToHTML converts troff to HTML (simplified)
func (c *ManCmd) convertToHTML(troff string) string {
	// This is a very simplified conversion
	html := "<html><head><title>LLMSpell Manual</title></head><body>\n"
	html += "<pre>\n"
	html += c.convertToText(troff)
	html += "</pre>\n"
	html += "</body></html>\n"
	return html
}

// replaceAll replaces all occurrences of old with new
func replaceAll(s, old, new string) string {
	// Simple implementation since strings.ReplaceAll might not be available
	for {
		i := findString(s, old)
		if i == -1 {
			break
		}
		s = s[:i] + new + s[i+len(old):]
	}
	return s
}

// findString finds the index of substr in s
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
