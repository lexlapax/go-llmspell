// ABOUTME: Template generator for creating new spell projects with scaffolding.
// ABOUTME: Provides various templates (basic, advanced, agent-based) for quick spell creation.

package template

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lexlapax/go-llmspell/pkg/errors"
)

// SpellTemplate represents a spell template
type SpellTemplate struct {
	Name        string
	Description string
	Type        TemplateType
	Files       map[string]FileTemplate
}

// TemplateType represents the type of template
type TemplateType string

const (
	TemplateTypeBasic       TemplateType = "basic"
	TemplateTypeAdvanced    TemplateType = "advanced"
	TemplateTypeAgent       TemplateType = "agent"
	TemplateTypeWorkflow    TemplateType = "workflow"
	TemplateTypeInteractive TemplateType = "interactive"
)

// FileTemplate represents a file to be generated
type FileTemplate struct {
	Path     string
	Content  string
	Template bool // If true, content is a Go template
}

// GeneratorOptions contains options for generating a spell
type GeneratorOptions struct {
	Name        string       // Name of the spell
	Type        TemplateType // Type of template to use
	Engine      string       // Script engine (lua, javascript, tengo)
	Description string       // Spell description
	Author      string       // Author name
	License     string       // License type
	OutputDir   string       // Output directory
	Force       bool         // Overwrite existing files
}

// Generator generates spell scaffolding
type Generator struct {
	templates map[TemplateType]*SpellTemplate
}

// NewGenerator creates a new template generator
func NewGenerator() *Generator {
	g := &Generator{
		templates: make(map[TemplateType]*SpellTemplate),
	}
	g.registerBuiltinTemplates()
	return g
}

// Generate creates a new spell from a template
func (g *Generator) Generate(opts GeneratorOptions) error {
	// Validate options
	if err := g.validateOptions(opts); err != nil {
		return err
	}

	// Get template
	tmpl, exists := g.templates[opts.Type]
	if !exists {
		return errors.Newf(errors.CategoryValidation, "unknown template type: %s", opts.Type)
	}

	// Create output directory
	outputPath := filepath.Join(opts.OutputDir, opts.Name)
	if err := g.createOutputDir(outputPath, opts.Force); err != nil {
		return err
	}

	// Generate files
	for _, fileTmpl := range tmpl.Files {
		if err := g.generateFile(outputPath, fileTmpl, opts); err != nil {
			return errors.Wrapf(err, errors.CategoryIO, "failed to generate file: %s", fileTmpl.Path)
		}
	}

	return nil
}

// ListTemplates returns available template types
func (g *Generator) ListTemplates() []TemplateInfo {
	var templates []TemplateInfo
	for typ, tmpl := range g.templates {
		templates = append(templates, TemplateInfo{
			Type:        typ,
			Name:        tmpl.Name,
			Description: tmpl.Description,
		})
	}
	return templates
}

// TemplateInfo contains information about a template
type TemplateInfo struct {
	Type        TemplateType
	Name        string
	Description string
}

// validateOptions validates generation options
func (g *Generator) validateOptions(opts GeneratorOptions) error {
	if opts.Name == "" {
		return errors.New(errors.CategoryValidation, "spell name is required")
	}

	if opts.Engine == "" {
		opts.Engine = "lua" // Default to Lua
	}

	validEngines := map[string]bool{
		"lua":        true,
		"javascript": true,
		"js":         true,
		"tengo":      true,
	}

	if !validEngines[opts.Engine] {
		return errors.Newf(errors.CategoryValidation, "invalid engine: %s", opts.Engine)
	}

	return nil
}

// createOutputDir creates the output directory
func (g *Generator) createOutputDir(path string, force bool) error {
	// Check if directory exists
	if _, err := os.Stat(path); err == nil {
		if !force {
			return errors.Newf(errors.CategoryIO, "directory already exists: %s", path)
		}
	}

	// Create directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.Wrap(err, errors.CategoryIO, "failed to create directory")
	}

	return nil
}

// generateFile generates a single file
func (g *Generator) generateFile(outputPath string, fileTmpl FileTemplate, opts GeneratorOptions) error {
	// Calculate full file path
	fullPath := filepath.Join(outputPath, fileTmpl.Path)

	// Substitute engine-specific extensions
	fullPath = g.substituteExtension(fullPath, opts.Engine)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, errors.CategoryIO, "failed to create directory")
	}

	// Generate content
	var content string
	if fileTmpl.Template {
		// Process as template
		tmpl, err := template.New(fileTmpl.Path).Parse(fileTmpl.Content)
		if err != nil {
			return errors.Wrap(err, errors.CategoryValidation, "failed to parse template")
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, opts); err != nil {
			return errors.Wrap(err, errors.CategoryEngine, "failed to execute template")
		}
		content = buf.String()
	} else {
		content = fileTmpl.Content
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return errors.Wrap(err, errors.CategoryIO, "failed to write file")
	}

	return nil
}

// substituteExtension substitutes script file extensions based on engine
func (g *Generator) substituteExtension(path string, engine string) string {
	if strings.HasSuffix(path, ".script") {
		ext := g.getEngineExtension(engine)
		return strings.TrimSuffix(path, ".script") + ext
	}
	return path
}

// getEngineExtension returns the file extension for an engine
func (g *Generator) getEngineExtension(engine string) string {
	switch engine {
	case "lua":
		return ".lua"
	case "javascript", "js":
		return ".js"
	case "tengo":
		return ".tengo"
	default:
		return ".lua"
	}
}

// registerBuiltinTemplates registers the built-in templates
func (g *Generator) registerBuiltinTemplates() {
	g.templates[TemplateTypeBasic] = g.createBasicTemplate()
	g.templates[TemplateTypeAdvanced] = g.createAdvancedTemplate()
	g.templates[TemplateTypeAgent] = g.createAgentTemplate()
	g.templates[TemplateTypeWorkflow] = g.createWorkflowTemplate()
	g.templates[TemplateTypeInteractive] = g.createInteractiveTemplate()
}

// Template creation methods follow...
