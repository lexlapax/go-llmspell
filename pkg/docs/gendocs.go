// ABOUTME: Main documentation generator interface and orchestrator for all script engines.
// ABOUTME: Coordinates language-specific generators to produce unified API documentation.

package docs

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DocGenerator is the interface for language-specific documentation generators
type DocGenerator interface {
	// GetLanguage returns the script language this generator supports
	GetLanguage() string

	// ExtractAPIs extracts API information from bridges and stdlib
	ExtractAPIs() ([]Module, error)

	// GenerateMarkdown generates markdown documentation
	GenerateMarkdown(modules []Module) string

	// GenerateJSON generates JSON documentation
	GenerateJSON(modules []Module) (string, error)

	// GenerateCompletion generates IDE completion data
	GenerateCompletion(modules []Module) interface{}
}

// Module represents a generic module in documentation (language-agnostic)
type Module struct {
	Name        string                 `json:"name"`
	Language    string                 `json:"language"`
	Description string                 `json:"description"`
	Functions   []Function             `json:"functions"`
	Constants   map[string]interface{} `json:"constants"`
	Types       []Type                 `json:"types"`
	Examples    []ModuleExample        `json:"examples"`
	SeeAlso     []string               `json:"see_also"`
	Since       string                 `json:"since"`
}

// Parameter represents a function parameter
type Parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Optional    bool        `json:"optional"`
	Default     interface{} `json:"default"`
}

// ReturnValue represents a function return value
type ReturnValue struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Field represents a type field
type Field struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Optional    bool        `json:"optional"`
	Default     interface{} `json:"default"`
}

// Function represents a generic function in documentation
type Function struct {
	Name        string                 `json:"name"`
	Module      string                 `json:"module"`
	Description string                 `json:"description"`
	Parameters  []Parameter            `json:"parameters"`
	Returns     []ReturnValue          `json:"returns"`
	Examples    []string               `json:"examples"`
	SeeAlso     []string               `json:"see_also"`
	Since       string                 `json:"since"`
	Deprecated  bool                   `json:"deprecated"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Type represents a type definition
type Type struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fields      []Field           `json:"fields"`
	Methods     []Function        `json:"methods"`
	Examples    []string          `json:"examples"`
	Metadata    map[string]string `json:"metadata"`
}

// ModuleExample represents a usage example
type ModuleExample struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Output      string `json:"output"`
}

// GeneratorManager manages multiple language-specific generators
type GeneratorManager struct {
	generators map[string]DocGenerator
	config     GeneratorConfig
}

// GeneratorConfig holds configuration for documentation generation
type GeneratorConfig struct {
	OutputFormats []string // markdown, json, html, completion
	Languages     []string // lua, javascript, tengo
	IncludeStdlib bool
	IncludeBridge bool
	Version       string
}

// NewGeneratorManager creates a new generator manager
func NewGeneratorManager(config GeneratorConfig) *GeneratorManager {
	return &GeneratorManager{
		generators: make(map[string]DocGenerator),
		config:     config,
	}
}

// RegisterGenerator registers a language-specific generator
func (m *GeneratorManager) RegisterGenerator(generator DocGenerator) error {
	lang := generator.GetLanguage()
	if _, exists := m.generators[lang]; exists {
		return fmt.Errorf("generator for language %s already registered", lang)
	}
	m.generators[lang] = generator
	return nil
}

// GenerateAll generates documentation for all registered languages
func (m *GeneratorManager) GenerateAll() (*GenerationResult, error) {
	result := &GenerationResult{
		Timestamp: time.Now(),
		Version:   m.config.Version,
		Languages: make(map[string]*LanguageResult),
	}

	// Process each language
	for lang, generator := range m.generators {
		// Skip if language not requested
		if len(m.config.Languages) > 0 && !contains(m.config.Languages, lang) {
			continue
		}

		langResult := &LanguageResult{
			Language: lang,
			Modules:  []Module{},
		}

		// Extract APIs
		modules, err := generator.ExtractAPIs()
		if err != nil {
			return nil, fmt.Errorf("failed to extract APIs for %s: %w", lang, err)
		}

		langResult.Modules = modules

		// Generate requested formats
		for _, format := range m.config.OutputFormats {
			switch format {
			case "markdown":
				langResult.Markdown = generator.GenerateMarkdown(modules)
			case "json":
				jsonStr, err := generator.GenerateJSON(modules)
				if err != nil {
					return nil, fmt.Errorf("failed to generate JSON for %s: %w", lang, err)
				}
				langResult.JSON = jsonStr
			case "completion":
				langResult.Completion = generator.GenerateCompletion(modules)
			}
		}

		result.Languages[lang] = langResult
	}

	return result, nil
}

// GenerateForLanguage generates documentation for a specific language
func (m *GeneratorManager) GenerateForLanguage(language string) (*LanguageResult, error) {
	generator, exists := m.generators[language]
	if !exists {
		return nil, fmt.Errorf("no generator registered for language: %s", language)
	}

	result := &LanguageResult{
		Language: language,
	}

	// Extract APIs
	modules, err := generator.ExtractAPIs()
	if err != nil {
		return nil, fmt.Errorf("failed to extract APIs: %w", err)
	}

	result.Modules = modules

	// Generate requested formats
	for _, format := range m.config.OutputFormats {
		switch format {
		case "markdown":
			result.Markdown = generator.GenerateMarkdown(modules)
		case "json":
			jsonStr, err := generator.GenerateJSON(modules)
			if err != nil {
				return nil, fmt.Errorf("failed to generate JSON: %w", err)
			}
			result.JSON = jsonStr
		case "completion":
			result.Completion = generator.GenerateCompletion(modules)
		}
	}

	return result, nil
}

// GenerationResult holds the complete generation result
type GenerationResult struct {
	Timestamp time.Time                  `json:"timestamp"`
	Version   string                     `json:"version"`
	Languages map[string]*LanguageResult `json:"languages"`
}

// LanguageResult holds generation result for a specific language
type LanguageResult struct {
	Language   string      `json:"language"`
	Modules    []Module    `json:"modules"`
	Markdown   string      `json:"-"`
	JSON       string      `json:"-"`
	Completion interface{} `json:"-"`
}

// GetOutputPath returns the output path for a specific format and language
func GetOutputPath(baseDir, language, format string) string {
	filename := fmt.Sprintf("%s-api", language)
	switch format {
	case "markdown":
		filename += ".md"
	case "json":
		filename += ".json"
	case "completion":
		filename += "-completion.json"
	case "html":
		filename += ".html"
	default:
		filename += ".txt"
	}
	return filepath.Join(baseDir, filename)
}

// GetSupportedLanguages returns all supported script languages
func GetSupportedLanguages() []string {
	return []string{"lua", "javascript", "tengo"}
}

// GetSupportedFormats returns all supported output formats
func GetSupportedFormats() []string {
	return []string{"markdown", "json", "completion", "html"}
}

// GenerateCombinedMarkdown generates a combined markdown document for all languages
func GenerateCombinedMarkdown(result *GenerationResult) string {
	var md strings.Builder

	md.WriteString("# LLMSpell API Documentation\n\n")
	md.WriteString(fmt.Sprintf("Generated on %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("Version: %s\n\n", result.Version))

	// Table of contents
	md.WriteString("## Table of Contents\n\n")

	// Sort languages for consistent output
	var languages []string
	for lang := range result.Languages {
		languages = append(languages, lang)
	}
	sort.Strings(languages)

	for _, lang := range languages {
		langResult := result.Languages[lang]
		md.WriteString(fmt.Sprintf("- [%s](#%s)\n", toTitle(lang), lang))
		for _, module := range langResult.Modules {
			md.WriteString(fmt.Sprintf("  - [%s](#%s-%s)\n", module.Name, lang, module.Name))
		}
	}
	md.WriteString("\n")

	// Language sections
	for _, lang := range languages {
		langResult := result.Languages[lang]
		md.WriteString(fmt.Sprintf("## %s\n\n", toTitle(lang)))

		// Use the language-specific markdown if available
		if langResult.Markdown != "" {
			// Remove the header from language-specific markdown
			lines := strings.Split(langResult.Markdown, "\n")
			startIdx := 0
			for i, line := range lines {
				if strings.HasPrefix(line, "## ") && i > 0 {
					startIdx = i
					break
				}
			}
			if startIdx > 0 {
				md.WriteString(strings.Join(lines[startIdx:], "\n"))
			} else {
				md.WriteString(langResult.Markdown)
			}
		}
		md.WriteString("\n")
	}

	return md.String()
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
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
	// Title case each word
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
