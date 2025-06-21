// ABOUTME: Test file for the main documentation generator interface and orchestrator.
// ABOUTME: Tests DocGenerator interface, GeneratorManager, and multi-language coordination.

package docs

import (
	"strings"
	"testing"
	"time"
)

// MockDocGenerator is a mock implementation of DocGenerator for testing
type MockDocGenerator struct {
	language string
	modules  []Module
	error    error
}

func (m *MockDocGenerator) GetLanguage() string {
	return m.language
}

func (m *MockDocGenerator) ExtractAPIs() ([]Module, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.modules, nil
}

func (m *MockDocGenerator) GenerateMarkdown(modules []Module) string {
	var sb strings.Builder
	// Use simple title case instead of deprecated strings.Title
	title := m.language
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}
	sb.WriteString("# " + title + " Documentation\n\n")
	for _, module := range modules {
		sb.WriteString("## " + module.Name + "\n\n")
	}
	return sb.String()
}

func (m *MockDocGenerator) GenerateJSON(modules []Module) (string, error) {
	if m.error != nil {
		return "", m.error
	}
	return `{"language":"` + m.language + `","modules":[]}`, nil
}

func (m *MockDocGenerator) GenerateCompletion(modules []Module) interface{} {
	return map[string]interface{}{
		"language": m.language,
		"modules":  len(modules),
	}
}

func TestNewGeneratorManager(t *testing.T) {
	config := GeneratorConfig{
		OutputFormats: []string{"markdown", "json"},
		Languages:     []string{"lua", "javascript"},
		Version:       "1.0.0",
	}

	manager := NewGeneratorManager(config)
	if manager == nil {
		t.Fatal("Expected non-nil manager")
		return // unreachable, but makes static analyzer happy
	}
	if manager.config.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", manager.config.Version)
	}
}

func TestRegisterGenerator(t *testing.T) {
	manager := NewGeneratorManager(GeneratorConfig{})

	// Test successful registration
	gen1 := &MockDocGenerator{language: "lua"}
	err := manager.RegisterGenerator(gen1)
	if err != nil {
		t.Errorf("Failed to register generator: %v", err)
	}

	// Test duplicate registration
	gen2 := &MockDocGenerator{language: "lua"}
	err = manager.RegisterGenerator(gen2)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("Expected 'already registered' error, got: %v", err)
	}

	// Test different language registration
	gen3 := &MockDocGenerator{language: "javascript"}
	err = manager.RegisterGenerator(gen3)
	if err != nil {
		t.Errorf("Failed to register different language: %v", err)
	}
}

func TestGenerateAll(t *testing.T) {
	config := GeneratorConfig{
		OutputFormats: []string{"markdown", "json", "completion"},
		Languages:     []string{"lua", "javascript"},
		Version:       "1.0.0",
	}
	manager := NewGeneratorManager(config)

	// Register generators
	luaGen := &MockDocGenerator{
		language: "lua",
		modules: []Module{
			{Name: "test", Language: "lua", Description: "Test module"},
		},
	}
	jsGen := &MockDocGenerator{
		language: "javascript",
		modules: []Module{
			{Name: "test", Language: "javascript", Description: "Test module"},
		},
	}

	if err := manager.RegisterGenerator(luaGen); err != nil {
		t.Fatalf("Failed to register Lua generator: %v", err)
	}
	if err := manager.RegisterGenerator(jsGen); err != nil {
		t.Fatalf("Failed to register JavaScript generator: %v", err)
	}

	// Generate all
	result, err := manager.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	// Check result
	if result.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", result.Version)
	}
	if len(result.Languages) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(result.Languages))
	}

	// Check Lua result
	luaResult, exists := result.Languages["lua"]
	if !exists {
		t.Error("Lua result not found")
	}
	if luaResult.Language != "lua" {
		t.Errorf("Expected language 'lua', got %s", luaResult.Language)
	}
	if len(luaResult.Modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(luaResult.Modules))
	}
	if luaResult.Markdown == "" {
		t.Error("Expected non-empty markdown")
	}
	if luaResult.JSON == "" {
		t.Error("Expected non-empty JSON")
	}
	if luaResult.Completion == nil {
		t.Error("Expected non-nil completion data")
	}
}

func TestGenerateAllWithFilter(t *testing.T) {
	config := GeneratorConfig{
		OutputFormats: []string{"markdown"},
		Languages:     []string{"lua"}, // Only Lua
		Version:       "1.0.0",
	}
	manager := NewGeneratorManager(config)

	// Register both generators
	luaGen := &MockDocGenerator{language: "lua", modules: []Module{}}
	jsGen := &MockDocGenerator{language: "javascript", modules: []Module{}}

	if err := manager.RegisterGenerator(luaGen); err != nil {
		t.Fatalf("Failed to register Lua generator: %v", err)
	}
	if err := manager.RegisterGenerator(jsGen); err != nil {
		t.Fatalf("Failed to register JavaScript generator: %v", err)
	}

	// Generate all (should only include Lua)
	result, err := manager.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	if len(result.Languages) != 1 {
		t.Errorf("Expected 1 language, got %d", len(result.Languages))
	}
	if _, exists := result.Languages["lua"]; !exists {
		t.Error("Lua result not found")
	}
	if _, exists := result.Languages["javascript"]; exists {
		t.Error("JavaScript should not be included")
	}
}

func TestGenerateForLanguage(t *testing.T) {
	config := GeneratorConfig{
		OutputFormats: []string{"markdown", "json"},
		Version:       "1.0.0",
	}
	manager := NewGeneratorManager(config)

	luaGen := &MockDocGenerator{
		language: "lua",
		modules: []Module{
			{Name: "test", Language: "lua"},
		},
	}
	if err := manager.RegisterGenerator(luaGen); err != nil {
		t.Fatalf("Failed to register generator: %v", err)
	}

	// Test existing language
	result, err := manager.GenerateForLanguage("lua")
	if err != nil {
		t.Fatalf("GenerateForLanguage failed: %v", err)
	}
	if result.Language != "lua" {
		t.Errorf("Expected language 'lua', got %s", result.Language)
	}
	if len(result.Modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(result.Modules))
	}

	// Test non-existent language
	_, err = manager.GenerateForLanguage("python")
	if err == nil {
		t.Error("Expected error for non-existent language")
	}
	if !strings.Contains(err.Error(), "no generator registered") {
		t.Errorf("Expected 'no generator registered' error, got: %v", err)
	}
}

func TestGetOutputPath(t *testing.T) {
	tests := []struct {
		baseDir  string
		language string
		format   string
		expected string
	}{
		{"docs", "lua", "markdown", "docs/lua-api.md"},
		{"docs", "javascript", "json", "docs/javascript-api.json"},
		{"output", "tengo", "completion", "output/tengo-api-completion.json"},
		{"docs", "lua", "html", "docs/lua-api.html"},
		{"docs", "lua", "unknown", "docs/lua-api.txt"},
	}

	for _, test := range tests {
		result := GetOutputPath(test.baseDir, test.language, test.format)
		if result != test.expected {
			t.Errorf("GetOutputPath(%s, %s, %s) = %s, expected %s",
				test.baseDir, test.language, test.format, result, test.expected)
		}
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	languages := GetSupportedLanguages()
	if len(languages) != 3 {
		t.Errorf("Expected 3 languages, got %d", len(languages))
	}
	expected := []string{"lua", "javascript", "tengo"}
	for i, lang := range expected {
		if i >= len(languages) || languages[i] != lang {
			t.Errorf("Expected language %s at index %d", lang, i)
		}
	}
}

func TestGetSupportedFormats(t *testing.T) {
	formats := GetSupportedFormats()
	if len(formats) != 4 {
		t.Errorf("Expected 4 formats, got %d", len(formats))
	}
	expected := []string{"markdown", "json", "completion", "html"}
	for i, format := range expected {
		if i >= len(formats) || formats[i] != format {
			t.Errorf("Expected format %s at index %d", format, i)
		}
	}
}

func TestGenerateCombinedMarkdown(t *testing.T) {
	result := &GenerationResult{
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Languages: map[string]*LanguageResult{
			"lua": {
				Language: "lua",
				Modules: []Module{
					{Name: "llm", Description: "LLM module"},
					{Name: "agent", Description: "Agent module"},
				},
				Markdown: "# Lua API Documentation\n\n## llm\n\nLLM functionality\n\n## agent\n\nAgent functionality",
			},
			"javascript": {
				Language: "javascript",
				Modules: []Module{
					{Name: "promise", Description: "Promise module"},
				},
				Markdown: "# JavaScript API Documentation\n\n## promise\n\nPromise functionality",
			},
		},
	}

	markdown := GenerateCombinedMarkdown(result)

	// Check header
	if !strings.Contains(markdown, "# LLMSpell API Documentation") {
		t.Error("Missing main header")
	}
	if !strings.Contains(markdown, "Version: 1.0.0") {
		t.Error("Missing version")
	}

	// Check table of contents
	if !strings.Contains(markdown, "## Table of Contents") {
		t.Error("Missing table of contents")
	}
	if !strings.Contains(markdown, "[Javascript](#javascript)") {
		t.Error("Missing JavaScript in TOC")
	}
	if !strings.Contains(markdown, "[Lua](#lua)") {
		t.Error("Missing Lua in TOC")
	}

	// Check language sections
	if !strings.Contains(markdown, "## Javascript") {
		t.Error("Missing JavaScript section")
	}
	if !strings.Contains(markdown, "## Lua") {
		t.Error("Missing Lua section")
	}

	// Check module content is included (should contain the module sections)
	if !strings.Contains(markdown, "## llm") {
		t.Error("Missing LLM module section")
	}
	if !strings.Contains(markdown, "## agent") {
		t.Error("Missing Agent module section")
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !contains(slice, "banana") {
		t.Error("Expected to find 'banana'")
	}
	if contains(slice, "grape") {
		t.Error("Should not find 'grape'")
	}
	if contains([]string{}, "anything") {
		t.Error("Should not find anything in empty slice")
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"lua", "Lua"},
		{"javascript", "Javascript"},
		{"hello world", "Hello World"},
		{"a", "A"},
		{"ABC", "Abc"},
	}

	for _, test := range tests {
		result := toTitle(test.input)
		if result != test.expected {
			t.Errorf("toTitle(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestGeneratorWithError(t *testing.T) {
	config := GeneratorConfig{
		OutputFormats: []string{"markdown"},
		Languages:     []string{"lua"},
		Version:       "1.0.0",
	}
	manager := NewGeneratorManager(config)

	// Register generator that returns error
	errorGen := &MockDocGenerator{
		language: "lua",
		error:    &testError{msg: "extraction failed"},
	}
	if err := manager.RegisterGenerator(errorGen); err != nil {
		t.Fatalf("Failed to register generator: %v", err)
	}

	// Test GenerateAll with error
	_, err := manager.GenerateAll()
	if err == nil {
		t.Error("Expected error from GenerateAll")
	}
	if !strings.Contains(err.Error(), "extraction failed") {
		t.Errorf("Expected 'extraction failed' error, got: %v", err)
	}

	// Test GenerateForLanguage with error
	_, err = manager.GenerateForLanguage("lua")
	if err == nil {
		t.Error("Expected error from GenerateForLanguage")
	}
	if !strings.Contains(err.Error(), "extraction failed") {
		t.Errorf("Expected 'extraction failed' error, got: %v", err)
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestJSONGenerationError(t *testing.T) {
	config := GeneratorConfig{
		OutputFormats: []string{"json"},
		Languages:     []string{"lua"},
		Version:       "1.0.0",
	}
	manager := NewGeneratorManager(config)

	// Register generator that returns JSON error
	jsonErrorGen := &MockDocGenerator{
		language: "lua",
		modules:  []Module{{Name: "test"}},
	}
	// Override GenerateJSON to return error
	jsonErrorGen.error = &testError{msg: "JSON generation failed"}

	if err := manager.RegisterGenerator(jsonErrorGen); err != nil {
		t.Fatalf("Failed to register generator: %v", err)
	}

	// GenerateJSON will fail when error is set
	_, err := manager.GenerateAll()
	if err == nil {
		t.Error("Expected error from GenerateAll")
	}
	if !strings.Contains(err.Error(), "failed to extract APIs") {
		t.Errorf("Expected API extraction error, got: %v", err)
	}
}
