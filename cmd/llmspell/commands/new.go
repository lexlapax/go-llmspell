// ABOUTME: Implementation of the new command for creating spell projects from templates.
// ABOUTME: Provides scaffolding generation with various template types and engines.

package commands

import (
	"context"
	"os"
	"path/filepath"

	pkgerrors "github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/template"
)

// NewCmd creates a new spell from a template
type NewCmd struct {
	BaseCommand
	Name        string `arg:"" name:"name" help:"Name of the spell to create"`
	Type        string `short:"t" help:"Template type (basic, advanced, agent, workflow, interactive)" default:"basic"`
	Engine      string `short:"e" help:"Script engine (lua, javascript, tengo)" default:"lua"`
	Description string `short:"d" help:"Spell description" default:"A new spell"`
	Author      string `short:"a" help:"Author name"`
	License     string `short:"l" help:"License type" default:"MIT"`
	OutputDir   string `short:"o" help:"Output directory" default:"."`
	Force       bool   `short:"f" help:"Overwrite existing directory"`
	List        bool   `help:"List available templates"`
}

// Run executes the command
func (c *NewCmd) Run(ctx context.Context) error {
	// Create generator
	gen := template.NewGenerator()

	// If listing templates
	if c.List {
		return c.listTemplates(gen)
	}

	// Validate name
	if c.Name == "" {
		return pkgerrors.New(pkgerrors.CategoryValidation, "spell name is required")
	}

	// Get author from git config if not provided
	if c.Author == "" {
		c.Author = c.getGitAuthor()
	}

	// Convert type string to TemplateType
	tmplType := template.TemplateType(c.Type)

	// Generate spell
	opts := template.GeneratorOptions{
		Name:        c.Name,
		Type:        tmplType,
		Engine:      c.Engine,
		Description: c.Description,
		Author:      c.Author,
		License:     c.License,
		OutputDir:   c.OutputDir,
		Force:       c.Force,
	}

	c.Info(ctx, "Creating new %s spell: %s", c.Type, c.Name)
	c.Debug(ctx, "Template: %s", c.Type)
	c.Debug(ctx, "Engine: %s", c.Engine)
	c.Debug(ctx, "Author: %s", c.Author)
	c.Debug(ctx, "License: %s", c.License)

	if err := gen.Generate(opts); err != nil {
		return pkgerrors.Wrap(err, pkgerrors.CategoryEngine, "failed to generate spell")
	}

	// Calculate output path
	outputPath := filepath.Join(c.OutputDir, c.Name)
	absPath, _ := filepath.Abs(outputPath)

	c.Info(ctx, "✓ Spell created successfully at: %s", absPath)
	c.Info(ctx, "\nNext steps:")
	c.Info(ctx, "  1. cd %s", c.Name)
	c.Info(ctx, "  2. Review and edit spell.yaml")
	c.Info(ctx, "  3. Run: llmspell validate")
	c.Info(ctx, "  4. Run: llmspell run main.%s", c.getExtension())

	return nil
}

// listTemplates lists available templates
func (c *NewCmd) listTemplates(gen *template.Generator) error {
	templates := gen.ListTemplates()

	c.Println("Available Templates:")
	c.Println()

	for _, tmpl := range templates {
		c.Printf("  %-15s - %s\n", tmpl.Type, tmpl.Description)
	}

	c.Println()
	c.Println("Usage: llmspell new <name> --type <template>")
	c.Println()
	c.Println("Examples:")
	c.Println("  llmspell new my-spell")
	c.Println("  llmspell new chat-bot --type interactive")
	c.Println("  llmspell new data-processor --type workflow --engine javascript")
	c.Println("  llmspell new assistant --type agent --author \"John Doe\"")

	return nil
}

// getGitAuthor attempts to get author name from git config
func (c *NewCmd) getGitAuthor() string {
	// Try to get from git config
	if output, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".gitconfig")); err == nil {
		// Simple parsing - look for name =
		lines := string(output)
		for _, line := range splitLines(lines) {
			if len(line) > 7 && line[:7] == "\tname = " {
				return line[7:]
			}
		}
	}

	// Try environment variables
	if author := os.Getenv("USER"); author != "" {
		return author
	}

	return "Unknown Author"
}

// getExtension returns the file extension for the current engine
func (c *NewCmd) getExtension() string {
	switch c.Engine {
	case "javascript", "js":
		return "js"
	case "tengo":
		return "tengo"
	default:
		return "lua"
	}
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	// Always append the last part, even if it's empty
	lines = append(lines, s[start:])
	return lines
}
