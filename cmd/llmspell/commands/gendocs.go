// ABOUTME: Generate documentation command for extracting API documentation.
// ABOUTME: Thin wrapper that delegates to pkg/docs/gendocs for actual generation.

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/lexlapax/go-llmspell/pkg/docs"
)

// GenDocsCmd generates API documentation for script languages
type GenDocsCmd struct {
	BaseCommand
	Output   string `short:"o" default:"docs/api" help:"Output directory for documentation"`
	Format   string `short:"f" enum:"markdown,json,completion,all" default:"all" help:"Output format"`
	Language string `short:"l" enum:"lua,javascript,tengo,all" default:"all" help:"Target language"`
	Version  string `short:"V" default:"1.0.0" help:"Documentation version"`
}

// Run executes the documentation generation
func (cmd *GenDocsCmd) Run(ctx *kong.Context) error {
	// Create output directory
	if err := os.MkdirAll(cmd.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine output formats
	var formats []string
	if cmd.Format == "all" {
		formats = docs.GetSupportedFormats()
	} else {
		formats = []string{cmd.Format}
	}

	// Determine languages
	var languages []string
	if cmd.Language == "all" {
		languages = docs.GetSupportedLanguages()
	} else {
		languages = []string{cmd.Language}
	}

	// Create generator configuration
	config := docs.GeneratorConfig{
		OutputFormats: formats,
		Languages:     languages,
		IncludeStdlib: true,
		IncludeBridge: true,
		Version:       cmd.Version,
	}

	// Create generator manager
	manager := docs.NewGeneratorManager(config)

	// Get bridge manager from context if available
	bridgeManager := GetBridgeManagerFromContext(ctx)

	// Register language-specific generators
	for _, lang := range languages {
		switch lang {
		case "lua":
			luaGen := docs.NewLuaDocGenerator(bridgeManager)
			if err := manager.RegisterGenerator(luaGen); err != nil {
				return fmt.Errorf("failed to register Lua generator: %w", err)
			}
		case "javascript":
			fmt.Fprintf(ctx.Stdout, "JavaScript documentation generator not yet implemented\n")
		case "tengo":
			fmt.Fprintf(ctx.Stdout, "Tengo documentation generator not yet implemented\n")
		}
	}

	// Generate documentation
	fmt.Fprintln(ctx.Stdout, "Generating API documentation...")
	result, err := manager.GenerateAll()
	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Write output files
	for lang, langResult := range result.Languages {
		fmt.Fprintf(ctx.Stdout, "\nGenerating %s documentation:\n", toTitle(lang))

		if langResult.Markdown != "" && containsString(formats, "markdown") {
			outputPath := docs.GetOutputPath(cmd.Output, lang, "markdown")
			if err := os.WriteFile(outputPath, []byte(langResult.Markdown), 0644); err != nil {
				return fmt.Errorf("failed to write markdown: %w", err)
			}
			fmt.Fprintf(ctx.Stdout, "  Written: %s (%d bytes)\n", outputPath, len(langResult.Markdown))
		}

		if langResult.JSON != "" && containsString(formats, "json") {
			outputPath := docs.GetOutputPath(cmd.Output, lang, "json")
			if err := os.WriteFile(outputPath, []byte(langResult.JSON), 0644); err != nil {
				return fmt.Errorf("failed to write JSON: %w", err)
			}
			fmt.Fprintf(ctx.Stdout, "  Written: %s (%d bytes)\n", outputPath, len(langResult.JSON))
		}

		if langResult.Completion != nil && containsString(formats, "completion") {
			outputPath := docs.GetOutputPath(cmd.Output, lang, "completion")
			jsonBytes, err := json.Marshal(langResult.Completion)
			if err != nil {
				return fmt.Errorf("failed to marshal completion data: %w", err)
			}
			if err := os.WriteFile(outputPath, jsonBytes, 0644); err != nil {
				return fmt.Errorf("failed to write completion data: %w", err)
			}
			fmt.Fprintf(ctx.Stdout, "  Written: %s (%d bytes)\n", outputPath, len(jsonBytes))
		}
	}

	// Generate combined markdown if multiple languages
	if cmd.Language == "all" && containsString(formats, "markdown") {
		combinedPath := filepath.Join(cmd.Output, "api-combined.md")
		combinedMarkdown := docs.GenerateCombinedMarkdown(result)
		if err := os.WriteFile(combinedPath, []byte(combinedMarkdown), 0644); err != nil {
			return fmt.Errorf("failed to write combined markdown: %w", err)
		}
		fmt.Fprintf(ctx.Stdout, "\n  Written: %s (%d bytes)\n", combinedPath, len(combinedMarkdown))
	}

	fmt.Fprintln(ctx.Stdout, "\nDocumentation generation complete!")
	fmt.Fprintf(ctx.Stdout, "Output directory: %s\n", cmd.Output)

	return nil
}

// GetBridgeManagerFromContext retrieves the bridge manager from context
func GetBridgeManagerFromContext(ctx *kong.Context) docs.BridgeManager {
	// TODO: Implement actual bridge manager retrieval from Kong context
	// For now, return a mock implementation
	return &mockBridgeManager{bridges: make(map[string]interface{})}
}

// mockBridgeManager is a temporary implementation
type mockBridgeManager struct {
	bridges map[string]interface{}
}

func (m *mockBridgeManager) ListBridges() []string {
	var ids []string
	for id := range m.bridges {
		ids = append(ids, id)
	}
	return ids
}

func (m *mockBridgeManager) GetBridge(id string) interface{} {
	return m.bridges[id]
}

// containsString checks if a string slice contains a value
func containsString(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// toTitle converts a string to title case
func toTitle(s string) string {
	if s == "" {
		return ""
	}
	// Simple implementation for ASCII strings
	return strings.ToUpper(s[:1]) + s[1:]
}
