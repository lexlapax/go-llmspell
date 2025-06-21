// ABOUTME: Tests for the template generator ensuring proper spell scaffolding generation.
// ABOUTME: Validates template creation, file generation, and various template types.

package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	assert.NotNil(t, g)
	assert.NotNil(t, g.templates)

	// Check that all built-in templates are registered
	assert.Contains(t, g.templates, TemplateTypeBasic)
	assert.Contains(t, g.templates, TemplateTypeAdvanced)
	assert.Contains(t, g.templates, TemplateTypeAgent)
	assert.Contains(t, g.templates, TemplateTypeWorkflow)
	assert.Contains(t, g.templates, TemplateTypeInteractive)
}

func TestGenerator_ListTemplates(t *testing.T) {
	g := NewGenerator()
	templates := g.ListTemplates()

	assert.Len(t, templates, 5)

	// Check template names
	templateNames := make(map[TemplateType]bool)
	for _, tmpl := range templates {
		templateNames[tmpl.Type] = true
	}

	assert.True(t, templateNames[TemplateTypeBasic])
	assert.True(t, templateNames[TemplateTypeAdvanced])
	assert.True(t, templateNames[TemplateTypeAgent])
	assert.True(t, templateNames[TemplateTypeWorkflow])
	assert.True(t, templateNames[TemplateTypeInteractive])
}

func TestGenerator_Generate_Basic(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	opts := GeneratorOptions{
		Name:        "test-spell",
		Type:        TemplateTypeBasic,
		Engine:      "lua",
		Description: "Test spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
	}

	err := g.Generate(opts)
	require.NoError(t, err)

	// Check generated files
	spellDir := filepath.Join(tmpDir, "test-spell")
	assert.DirExists(t, spellDir)

	// Check spell.yaml
	spellYaml := filepath.Join(spellDir, "spell.yaml")
	assert.FileExists(t, spellYaml)

	content, err := os.ReadFile(spellYaml)
	require.NoError(t, err)
	assert.Contains(t, string(content), "name: test-spell")
	assert.Contains(t, string(content), "description: Test spell")
	assert.Contains(t, string(content), "author: Test Author")
	assert.Contains(t, string(content), "engine: lua")

	// Check main script
	mainScript := filepath.Join(spellDir, "main.lua")
	assert.FileExists(t, mainScript)

	content, err = os.ReadFile(mainScript)
	require.NoError(t, err)
	assert.Contains(t, string(content), "-- test-spell")
	assert.Contains(t, string(content), "local llm = require(\"llm\")")

	// Check README
	readme := filepath.Join(spellDir, "README.md")
	assert.FileExists(t, readme)

	content, err = os.ReadFile(readme)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# test-spell")
	assert.Contains(t, string(content), "Test spell")
}

func TestGenerator_Generate_JavaScript(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	opts := GeneratorOptions{
		Name:        "js-spell",
		Type:        TemplateTypeBasic,
		Engine:      "javascript",
		Description: "JavaScript spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
	}

	err := g.Generate(opts)
	require.NoError(t, err)

	// Check generated files
	spellDir := filepath.Join(tmpDir, "js-spell")

	// Check main script has .js extension
	mainScript := filepath.Join(spellDir, "main.js")
	assert.FileExists(t, mainScript)

	content, err := os.ReadFile(mainScript)
	require.NoError(t, err)
	assert.Contains(t, string(content), "// js-spell")
	assert.Contains(t, string(content), "const llm = require('llm')")
	assert.Contains(t, string(content), "await client.complete(prompt)")
}

func TestGenerator_Generate_Tengo(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	opts := GeneratorOptions{
		Name:        "tengo-spell",
		Type:        TemplateTypeBasic,
		Engine:      "tengo",
		Description: "Tengo spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
	}

	err := g.Generate(opts)
	require.NoError(t, err)

	// Check generated files
	spellDir := filepath.Join(tmpDir, "tengo-spell")

	// Check main script has .tengo extension
	mainScript := filepath.Join(spellDir, "main.tengo")
	assert.FileExists(t, mainScript)

	content, err := os.ReadFile(mainScript)
	require.NoError(t, err)
	assert.Contains(t, string(content), "// tengo-spell")
	assert.Contains(t, string(content), "llm := import(\"llm\")")
}

func TestGenerator_Generate_Advanced(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	opts := GeneratorOptions{
		Name:        "advanced-spell",
		Type:        TemplateTypeAdvanced,
		Engine:      "lua",
		Description: "Advanced spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
	}

	err := g.Generate(opts)
	require.NoError(t, err)

	spellDir := filepath.Join(tmpDir, "advanced-spell")

	// Check additional files for advanced template
	assert.FileExists(t, filepath.Join(spellDir, "lib", "utils.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "lib", "prompts.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "config", "default.yaml"))

	// Check utils content
	content, err := os.ReadFile(filepath.Join(spellDir, "lib", "utils.lua"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "function utils.read_file")
	assert.Contains(t, string(content), "function utils.write_file")
}

func TestGenerator_Generate_Agent(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	opts := GeneratorOptions{
		Name:        "agent-spell",
		Type:        TemplateTypeAgent,
		Engine:      "lua",
		Description: "Agent spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
	}

	err := g.Generate(opts)
	require.NoError(t, err)

	spellDir := filepath.Join(tmpDir, "agent-spell")

	// Check tool files
	assert.FileExists(t, filepath.Join(spellDir, "tools", "calculator.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "tools", "web_search.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "tools", "file_reader.lua"))

	// Check calculator tool
	content, err := os.ReadFile(filepath.Join(spellDir, "tools", "calculator.lua"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "tool.name = \"calculator\"")
	assert.Contains(t, string(content), "function tool.execute(params)")
}

func TestGenerator_Generate_Workflow(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	opts := GeneratorOptions{
		Name:        "workflow-spell",
		Type:        TemplateTypeWorkflow,
		Engine:      "lua",
		Description: "Workflow spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
	}

	err := g.Generate(opts)
	require.NoError(t, err)

	spellDir := filepath.Join(tmpDir, "workflow-spell")

	// Check workflow files
	assert.FileExists(t, filepath.Join(spellDir, "workflows", "process_document.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "workflows", "generate_report.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "workflows", "analyze_data.lua"))

	// Check process_document workflow
	content, err := os.ReadFile(filepath.Join(spellDir, "workflows", "process_document.lua"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "workflow.name = \"Process Document\"")
	assert.Contains(t, string(content), "function workflow.execute(state)")
}

func TestGenerator_Generate_Interactive(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	opts := GeneratorOptions{
		Name:        "interactive-spell",
		Type:        TemplateTypeInteractive,
		Engine:      "lua",
		Description: "Interactive spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
	}

	err := g.Generate(opts)
	require.NoError(t, err)

	spellDir := filepath.Join(tmpDir, "interactive-spell")

	// Check main script
	content, err := os.ReadFile(filepath.Join(spellDir, "main.lua"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "-- Command handlers")
	assert.Contains(t, string(content), "local commands = {")
	assert.Contains(t, string(content), "[\"/help\"]")
}

func TestGenerator_Generate_ExistingDirectory(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	// Create existing directory
	spellDir := filepath.Join(tmpDir, "existing-spell")
	err := os.MkdirAll(spellDir, 0755)
	require.NoError(t, err)

	opts := GeneratorOptions{
		Name:        "existing-spell",
		Type:        TemplateTypeBasic,
		Engine:      "lua",
		Description: "Test spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
		Force:       false,
	}

	// Should fail without force
	err = g.Generate(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory already exists")

	// Should succeed with force
	opts.Force = true
	err = g.Generate(opts)
	assert.NoError(t, err)
}

func TestGenerator_ValidateOptions(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		name    string
		opts    GeneratorOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			opts: GeneratorOptions{
				Name:   "test-spell",
				Engine: "lua",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			opts: GeneratorOptions{
				Engine: "lua",
			},
			wantErr: true,
			errMsg:  "spell name is required",
		},
		{
			name: "invalid engine",
			opts: GeneratorOptions{
				Name:   "test-spell",
				Engine: "python",
			},
			wantErr: true,
			errMsg:  "invalid engine: python",
		},
		{
			name: "default engine",
			opts: GeneratorOptions{
				Name:   "test-spell",
				Engine: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.validateOptions(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerator_SubstituteExtension(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		path     string
		engine   string
		expected string
	}{
		{"main.script", "lua", "main.lua"},
		{"main.script", "javascript", "main.js"},
		{"main.script", "js", "main.js"},
		{"main.script", "tengo", "main.tengo"},
		{"lib/utils.script", "lua", "lib/utils.lua"},
		{"main.lua", "lua", "main.lua"},     // No substitution
		{"script.txt", "lua", "script.txt"}, // No substitution
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.engine, func(t *testing.T) {
			result := g.substituteExtension(tt.path, tt.engine)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerator_TemplateProcessing(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	// Test template variable substitution
	opts := GeneratorOptions{
		Name:        "template-test",
		Type:        TemplateTypeBasic,
		Engine:      "lua",
		Description: "Template {{Test}} Description",
		Author:      "Test {{Author}}",
		License:     "Custom {{License}}",
		OutputDir:   tmpDir,
	}

	err := g.Generate(opts)
	require.NoError(t, err)

	// Check that template variables are properly substituted
	spellYaml := filepath.Join(tmpDir, "template-test", "spell.yaml")
	content, err := os.ReadFile(spellYaml)
	require.NoError(t, err)

	// Should contain actual values, not template syntax
	assert.Contains(t, string(content), "name: template-test")
	assert.Contains(t, string(content), "description: Template {{Test}} Description")
	assert.Contains(t, string(content), "author: Test {{Author}}")
	assert.Contains(t, string(content), "license: Custom {{License}}")
	assert.NotContains(t, string(content), "{{.Name}}")
	assert.NotContains(t, string(content), "{{.Description}}")
}

func TestGenerator_AllEnginesAllTemplates(t *testing.T) {
	// Test that all combinations of engines and templates work
	g := NewGenerator()

	engines := []string{"lua", "javascript", "tengo"}
	templates := []TemplateType{
		TemplateTypeBasic,
		TemplateTypeAdvanced,
		TemplateTypeAgent,
		TemplateTypeWorkflow,
		TemplateTypeInteractive,
	}

	for _, engine := range engines {
		for _, tmplType := range templates {
			t.Run(engine+"_"+string(tmplType), func(t *testing.T) {
				tmpDir := t.TempDir()

				opts := GeneratorOptions{
					Name:        strings.ToLower(string(tmplType)) + "-" + engine,
					Type:        tmplType,
					Engine:      engine,
					Description: "Test spell",
					Author:      "Test Author",
					License:     "MIT",
					OutputDir:   tmpDir,
				}

				err := g.Generate(opts)
				assert.NoError(t, err)

				// Basic validation
				spellDir := filepath.Join(tmpDir, opts.Name)
				assert.DirExists(t, spellDir)
				assert.FileExists(t, filepath.Join(spellDir, "spell.yaml"))

				// Check main script has correct extension
				ext := ".lua"
				if engine == "javascript" || engine == "js" {
					ext = ".js"
				} else if engine == "tengo" {
					ext = ".tengo"
				}
				assert.FileExists(t, filepath.Join(spellDir, "main"+ext))
			})
		}
	}
}
