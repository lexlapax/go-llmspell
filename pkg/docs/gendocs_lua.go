// ABOUTME: Lua-specific documentation generator implementing DocGenerator interface.
// ABOUTME: Extracts API information from bridges and generates comprehensive Lua documentation.

package docs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BridgeManager interface for documentation extraction
type BridgeManager interface {
	ListBridges() []string
	GetBridge(id string) interface{}
}

// LuaDocGenerator generates documentation for Lua APIs
type LuaDocGenerator struct {
	BridgeManager BridgeManager
	ModulePaths   []string
	OutputFormats []string
}

// GetLanguage returns "lua" for this generator
func (g *LuaDocGenerator) GetLanguage() string {
	return "lua"
}

// ExtractAPIs extracts API information from bridges and stdlib
func (g *LuaDocGenerator) ExtractAPIs() ([]Module, error) {
	// Extract from bridges
	bridgeModules, err := g.ExtractBridgeAPIs()
	if err != nil {
		return nil, fmt.Errorf("failed to extract bridge APIs: %w", err)
	}

	// Extract from stdlib
	stdlibModules, err := g.ExtractStdlibAPIs()
	if err != nil {
		// Log warning but continue with just bridge modules
		fmt.Printf("Warning: Failed to extract stdlib APIs: %v\n", err)
		stdlibModules = []LuaModule{}
	}

	// Convert to generic modules
	var modules []Module
	for _, luaModule := range append(bridgeModules, stdlibModules...) {
		modules = append(modules, g.convertLuaModuleToGeneric(luaModule))
	}

	return modules, nil
}

// GenerateMarkdown generates markdown documentation
func (g *LuaDocGenerator) GenerateMarkdown(modules []Module) string {
	// Convert back to Lua modules for existing implementation
	var luaModules []LuaModule
	for _, module := range modules {
		luaModules = append(luaModules, g.convertGenericToLuaModule(module))
	}
	return g.GenerateMarkdownDocs(luaModules)
}

// GenerateJSON generates JSON documentation
func (g *LuaDocGenerator) GenerateJSON(modules []Module) (string, error) {
	// Convert back to Lua modules for existing implementation
	var luaModules []LuaModule
	for _, module := range modules {
		luaModules = append(luaModules, g.convertGenericToLuaModule(module))
	}
	return g.GenerateJSONDocs(luaModules)
}

// GenerateCompletion generates IDE completion data
func (g *LuaDocGenerator) GenerateCompletion(modules []Module) interface{} {
	// Convert back to Lua modules for existing implementation
	var luaModules []LuaModule
	for _, module := range modules {
		luaModules = append(luaModules, g.convertGenericToLuaModule(module))
	}
	return g.GenerateCompletionData(luaModules)
}

// APIFunction represents a Lua function in the documentation
type APIFunction struct {
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

// LuaModule represents a Lua module in the documentation
type LuaModule struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Functions   []APIFunction          `json:"functions"`
	Constants   map[string]interface{} `json:"constants"`
	Types       []TypeDefinition       `json:"types"`
	Examples    []ModuleExample        `json:"examples"`
	SeeAlso     []string               `json:"see_also"`
	Since       string                 `json:"since"`
}

// TypeDefinition represents a Lua type definition
type TypeDefinition struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fields      []Field           `json:"fields"`
	Methods     []APIFunction     `json:"methods"`
	Examples    []string          `json:"examples"`
	Metadata    map[string]string `json:"metadata"`
}

// CompletionData represents data for IDE completion
type CompletionData struct {
	Functions []CompletionFunction `json:"functions"`
	Modules   []CompletionModule   `json:"modules"`
	Types     []CompletionType     `json:"types"`
	Keywords  []string             `json:"keywords"`
}

// CompletionFunction represents a function for completion
type CompletionFunction struct {
	Name       string              `json:"name"`
	Module     string              `json:"module"`
	Signature  string              `json:"signature"`
	ReturnType string              `json:"return_type"`
	Parameters []CompletionParam   `json:"parameters"`
	Snippets   []CompletionSnippet `json:"snippets"`
}

// CompletionParam represents a parameter for completion
type CompletionParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
}

// CompletionModule represents a module for completion
type CompletionModule struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Functions   []string `json:"functions"`
}

// CompletionType represents a type for completion
type CompletionType struct {
	Name    string   `json:"name"`
	Fields  []string `json:"fields"`
	Methods []string `json:"methods"`
}

// CompletionSnippet represents a code snippet
type CompletionSnippet struct {
	Trigger     string `json:"trigger"`
	Description string `json:"description"`
	Body        string `json:"body"`
}

// NewLuaDocGenerator creates a new Lua documentation generator
func NewLuaDocGenerator(bridgeManager BridgeManager) *LuaDocGenerator {
	return &LuaDocGenerator{
		BridgeManager: bridgeManager,
		ModulePaths: []string{
			"pkg/engine/gopherlua/stdlib",
			"pkg/bridge",
		},
		OutputFormats: []string{"markdown", "json", "html"},
	}
}

// ExtractBridgeAPIs extracts API information from bridge implementations
func (g *LuaDocGenerator) ExtractBridgeAPIs() ([]LuaModule, error) {
	var modules []LuaModule

	// Get all registered bridges
	bridges := g.BridgeManager.ListBridges()

	for _, bridgeID := range bridges {
		bridgeInstance := g.BridgeManager.GetBridge(bridgeID)
		if bridgeInstance == nil {
			continue
		}

		module, err := g.extractBridgeModule(bridgeID, bridgeInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to extract bridge %s: %w", bridgeID, err)
		}

		modules = append(modules, module)
	}

	return modules, nil
}

// extractBridgeModule extracts module information from a bridge
func (g *LuaDocGenerator) extractBridgeModule(bridgeID string, bridgeInstance interface{}) (LuaModule, error) {
	module := LuaModule{
		Name:        bridgeID,
		Description: fmt.Sprintf("Bridge for %s functionality from go-llms", bridgeID),
		Functions:   []APIFunction{},
		Constants:   make(map[string]interface{}),
		Types:       []TypeDefinition{},
		Since:       "1.0.0",
	}

	// Use reflection to extract method information
	bridgeType := reflect.TypeOf(bridgeInstance)
	bridgeValue := reflect.ValueOf(bridgeInstance)

	// Handle pointer types
	originalType := bridgeType
	if bridgeType.Kind() == reflect.Ptr {
		// We use originalType for methods, no need to modify bridgeType
		_ = bridgeValue.Elem() // bridgeValue not used currently
	}

	// Extract methods using reflection - use original type for methods
	methodType := originalType
	for i := 0; i < methodType.NumMethod(); i++ {
		method := methodType.Method(i)

		// Skip unexported methods
		if !method.IsExported() {
			continue
		}

		// Skip certain internal methods
		if g.shouldSkipMethod(method.Name) {
			continue
		}

		function, err := g.extractMethodInfo(bridgeID, method)
		if err != nil {
			continue // Skip methods that can't be extracted
		}

		module.Functions = append(module.Functions, function)
	}

	// Sort functions alphabetically
	sort.Slice(module.Functions, func(i, j int) bool {
		return module.Functions[i].Name < module.Functions[j].Name
	})

	return module, nil
}

// shouldSkipMethod determines if a method should be skipped
func (g *LuaDocGenerator) shouldSkipMethod(methodName string) bool {
	skipMethods := []string{
		"String", "GoString", "Error", // Common Go methods
		"GetID", "GetMetadata", "Initialize", "Cleanup", // Bridge internal methods
	}

	for _, skip := range skipMethods {
		if methodName == skip {
			return true
		}
	}

	return false
}

// extractMethodInfo extracts information about a method
func (g *LuaDocGenerator) extractMethodInfo(moduleName string, method reflect.Method) (APIFunction, error) {
	function := APIFunction{
		Name:       g.convertMethodName(method.Name),
		Module:     moduleName,
		Parameters: []Parameter{},
		Returns:    []ReturnValue{},
		Examples:   []string{},
		SeeAlso:    []string{},
		Since:      "1.0.0",
		Tags:       []string{"bridge"},
		Metadata:   make(map[string]interface{}),
	}

	// Analyze method type
	methodType := method.Type

	// Extract parameters (skip receiver)
	for i := 1; i < methodType.NumIn(); i++ {
		paramType := methodType.In(i)
		param := Parameter{
			Name: fmt.Sprintf("arg%d", i),
			Type: g.convertGoTypeToLua(paramType),
		}
		function.Parameters = append(function.Parameters, param)
	}

	// Extract return values
	for i := 0; i < methodType.NumOut(); i++ {
		returnType := methodType.Out(i)
		returnVal := ReturnValue{
			Type: g.convertGoTypeToLua(returnType),
		}
		function.Returns = append(function.Returns, returnVal)
	}

	// Generate description
	function.Description = g.generateMethodDescription(method.Name, function.Parameters, function.Returns)

	// Add example
	example := g.generateMethodExample(moduleName, function.Name, function.Parameters)
	if example != "" {
		function.Examples = append(function.Examples, example)
	}

	return function, nil
}

// convertMethodName converts Go method name to Lua naming convention
func (g *LuaDocGenerator) convertMethodName(methodName string) string {
	// Convert PascalCase to snake_case
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(methodName, `${1}_${2}`)
	return strings.ToLower(snake)
}

// convertGoTypeToLua converts Go type to Lua type representation
func (g *LuaDocGenerator) convertGoTypeToLua(goType reflect.Type) string {
	switch goType.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Slice, reflect.Array:
		elemType := g.convertGoTypeToLua(goType.Elem())
		return fmt.Sprintf("table<%s>", elemType)
	case reflect.Map:
		keyType := g.convertGoTypeToLua(goType.Key())
		valueType := g.convertGoTypeToLua(goType.Elem())
		return fmt.Sprintf("table<%s, %s>", keyType, valueType)
	case reflect.Ptr:
		return g.convertGoTypeToLua(goType.Elem())
	case reflect.Interface:
		if goType.String() == "error" {
			return "string|nil"
		}
		return "any"
	case reflect.Struct:
		return "table"
	default:
		return "any"
	}
}

// generateMethodDescription generates a description for a method
func (g *LuaDocGenerator) generateMethodDescription(methodName string, params []Parameter, returns []ReturnValue) string {
	// Convert method name to human readable
	readable := strings.ReplaceAll(methodName, "_", " ")
	readable = toTitle(readable)

	description := fmt.Sprintf("%s operation", readable)

	if len(params) > 0 {
		description += fmt.Sprintf(" with %d parameter(s)", len(params))
	}

	if len(returns) > 0 {
		description += fmt.Sprintf(" returning %d value(s)", len(returns))
	}

	return description + "."
}

// generateMethodExample generates an example for a method
func (g *LuaDocGenerator) generateMethodExample(moduleName, methodName string, params []Parameter) string {
	if len(params) == 0 {
		return fmt.Sprintf("local result = %s.%s()", moduleName, methodName)
	}

	var paramNames []string
	for i, param := range params {
		switch param.Type {
		case "string":
			paramNames = append(paramNames, fmt.Sprintf(`"param%d"`, i+1))
		case "number":
			paramNames = append(paramNames, strconv.Itoa((i+1)*10))
		case "boolean":
			paramNames = append(paramNames, "true")
		default:
			paramNames = append(paramNames, fmt.Sprintf("param%d", i+1))
		}
	}

	return fmt.Sprintf("local result = %s.%s(%s)", moduleName, methodName, strings.Join(paramNames, ", "))
}

// ExtractStdlibAPIs extracts API information from Lua standard library modules
func (g *LuaDocGenerator) ExtractStdlibAPIs() ([]LuaModule, error) {
	var modules []LuaModule

	stdlibPath := "pkg/engine/gopherlua/stdlib"

	// Find all Lua files in stdlib
	luaFiles, err := filepath.Glob(filepath.Join(stdlibPath, "*.lua"))
	if err != nil {
		return nil, fmt.Errorf("failed to find Lua files: %w", err)
	}

	for _, luaFile := range luaFiles {
		module, err := g.parseStdlibModule(luaFile)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		modules = append(modules, module)
	}

	return modules, nil
}

// parseStdlibModule parses a Lua stdlib module file
func (g *LuaDocGenerator) parseStdlibModule(filePath string) (LuaModule, error) {
	// Extract module name from filename
	baseName := filepath.Base(filePath)
	moduleName := strings.TrimSuffix(baseName, ".lua")

	module := LuaModule{
		Name:        moduleName,
		Description: fmt.Sprintf("Lua standard library module for %s functionality", moduleName),
		Functions:   []APIFunction{},
		Constants:   make(map[string]interface{}),
		Types:       []TypeDefinition{},
		Since:       "1.0.0",
	}

	// Parse the Lua file to extract function definitions
	// This is a simplified parser - in a real implementation,
	// you might want to use a proper Lua parser
	functions, err := g.parseLuaFunctions(filePath)
	if err != nil {
		return module, err
	}

	module.Functions = functions
	return module, nil
}

// parseLuaFunctions parses Lua functions from a file
func (g *LuaDocGenerator) parseLuaFunctions(filePath string) ([]APIFunction, error) {
	// This is a simplified implementation
	// In practice, you'd want a more sophisticated Lua parser

	var functions []APIFunction

	// For now, we'll extract functions based on common patterns
	// This could be enhanced to parse actual Lua AST

	// Add some common functions based on module patterns
	baseName := filepath.Base(filePath)
	moduleName := strings.TrimSuffix(baseName, ".lua")

	// Add placeholder functions based on known stdlib modules
	switch moduleName {
	case "promise":
		functions = append(functions, g.createPromiseFunctions()...)
	case "llm":
		functions = append(functions, g.createLLMFunctions()...)
	case "agent":
		functions = append(functions, g.createAgentFunctions()...)
	case "tools":
		functions = append(functions, g.createToolsFunctions()...)
	case "state":
		functions = append(functions, g.createStateFunctions()...)
	case "events":
		functions = append(functions, g.createEventsFunctions()...)
	case "data":
		functions = append(functions, g.createDataFunctions()...)
	case "observability":
		functions = append(functions, g.createObservabilityFunctions()...)
	case "auth":
		functions = append(functions, g.createAuthFunctions()...)
	case "errors":
		functions = append(functions, g.createErrorsFunctions()...)
	case "logging":
		functions = append(functions, g.createLoggingFunctions()...)
	case "testing":
		functions = append(functions, g.createTestingFunctions()...)
	case "core":
		functions = append(functions, g.createCoreFunctions()...)
	case "spell":
		functions = append(functions, g.createSpellFunctions()...)
	}

	return functions, nil
}

// Helper functions to create function definitions for known modules
func (g *LuaDocGenerator) createPromiseFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "new",
			Description: "Create a new Promise with an executor function",
			Parameters: []Parameter{
				{Name: "executor", Type: "function", Description: "Function that receives resolve and reject callbacks"},
			},
			Returns: []ReturnValue{
				{Type: "Promise", Description: "A new Promise instance"},
			},
			Examples: []string{
				`local promise = Promise.new(function(resolve, reject)
    -- async work here
    resolve("success")
end)`,
			},
		},
		{
			Name:        "resolve",
			Description: "Create a resolved Promise with the given value",
			Parameters: []Parameter{
				{Name: "value", Type: "any", Description: "Value to resolve with", Optional: true},
			},
			Returns: []ReturnValue{
				{Type: "Promise", Description: "A resolved Promise"},
			},
			Examples: []string{
				`local promise = Promise.resolve("success")`,
			},
		},
		{
			Name:        "reject",
			Description: "Create a rejected Promise with the given error",
			Parameters: []Parameter{
				{Name: "error", Type: "any", Description: "Error to reject with"},
			},
			Returns: []ReturnValue{
				{Type: "Promise", Description: "A rejected Promise"},
			},
			Examples: []string{
				`local promise = Promise.reject("error message")`,
			},
		},
	}
}

func (g *LuaDocGenerator) createLLMFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "quick_prompt",
			Description: "Send a quick prompt to an LLM and get a response",
			Parameters: []Parameter{
				{Name: "prompt", Type: "string", Description: "The prompt to send"},
				{Name: "options", Type: "table", Description: "Optional configuration", Optional: true},
			},
			Returns: []ReturnValue{
				{Type: "string", Description: "The LLM response"},
			},
			Examples: []string{
				`local response = llm.quick_prompt("What is 2+2?")`,
			},
		},
	}
}

func (g *LuaDocGenerator) createAgentFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "create",
			Description: "Create a new agent with the specified configuration",
			Parameters: []Parameter{
				{Name: "name", Type: "string", Description: "Agent name"},
				{Name: "config", Type: "table", Description: "Agent configuration"},
			},
			Returns: []ReturnValue{
				{Type: "Agent", Description: "The created agent"},
			},
			Examples: []string{
				`local agent = agent.create("my-agent", {model = "gpt-4"})`,
			},
		},
	}
}

func (g *LuaDocGenerator) createToolsFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "define",
			Description: "Define a new tool with schema and implementation",
			Parameters: []Parameter{
				{Name: "name", Type: "string", Description: "Tool name"},
				{Name: "description", Type: "string", Description: "Tool description"},
				{Name: "schema", Type: "table", Description: "Parameter schema"},
				{Name: "func", Type: "function", Description: "Tool implementation"},
			},
			Returns: []ReturnValue{
				{Type: "Tool", Description: "The defined tool"},
			},
			Examples: []string{
				`local tool = tools.define("calculator", "Math calculator", schema, function(a, b) return a + b end)`,
			},
		},
	}
}

func (g *LuaDocGenerator) createStateFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "create",
			Description: "Create a new state container",
			Parameters: []Parameter{
				{Name: "initial_data", Type: "table", Description: "Initial state data", Optional: true},
			},
			Returns: []ReturnValue{
				{Type: "State", Description: "The created state"},
			},
			Examples: []string{
				`local state = state.create({counter = 0})`,
			},
		},
	}
}

func (g *LuaDocGenerator) createEventsFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "emit",
			Description: "Emit an event with optional data",
			Parameters: []Parameter{
				{Name: "event", Type: "string", Description: "Event name"},
				{Name: "data", Type: "any", Description: "Event data", Optional: true},
			},
			Returns: []ReturnValue{
				{Type: "boolean", Description: "true if event was emitted"},
			},
			Examples: []string{
				`events.emit("user-login", {user_id = 123})`,
			},
		},
	}
}

func (g *LuaDocGenerator) createDataFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "parse_json",
			Description: "Parse JSON string with optional schema validation",
			Parameters: []Parameter{
				{Name: "text", Type: "string", Description: "JSON text to parse"},
				{Name: "schema", Type: "table", Description: "Validation schema", Optional: true},
			},
			Returns: []ReturnValue{
				{Type: "table", Description: "Parsed data"},
			},
			Examples: []string{
				`local data = data.parse_json('{"name": "John"}')`,
			},
		},
	}
}

func (g *LuaDocGenerator) createObservabilityFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "counter",
			Description: "Create or increment a counter metric",
			Parameters: []Parameter{
				{Name: "name", Type: "string", Description: "Counter name"},
				{Name: "description", Type: "string", Description: "Counter description"},
				{Name: "tags", Type: "table", Description: "Metric tags", Optional: true},
			},
			Returns: []ReturnValue{
				{Type: "Counter", Description: "Counter metric"},
			},
			Examples: []string{
				`local counter = observability.counter("requests", "HTTP requests")`,
			},
		},
	}
}

func (g *LuaDocGenerator) createAuthFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "create_config",
			Description: "Create authentication configuration",
			Parameters: []Parameter{
				{Name: "type", Type: "string", Description: "Auth type (api_key, oauth2, etc.)"},
				{Name: "credentials", Type: "table", Description: "Authentication credentials"},
			},
			Returns: []ReturnValue{
				{Type: "AuthConfig", Description: "Authentication configuration"},
			},
			Examples: []string{
				`local config = auth.create_config("api_key", {key = "sk-..."})`,
			},
		},
	}
}

func (g *LuaDocGenerator) createErrorsFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "try",
			Description: "Execute function with try-catch-finally semantics",
			Parameters: []Parameter{
				{Name: "func", Type: "function", Description: "Function to execute"},
				{Name: "catch_func", Type: "function", Description: "Error handler", Optional: true},
				{Name: "finally_func", Type: "function", Description: "Cleanup function", Optional: true},
			},
			Returns: []ReturnValue{
				{Type: "any", Description: "Function result or error"},
			},
			Examples: []string{
				`local result = errors.try(risky_function, handle_error, cleanup)`,
			},
		},
	}
}

func (g *LuaDocGenerator) createLoggingFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "info",
			Description: "Log an info message",
			Parameters: []Parameter{
				{Name: "message", Type: "string", Description: "Log message"},
				{Name: "context", Type: "table", Description: "Additional context", Optional: true},
			},
			Returns: []ReturnValue{},
			Examples: []string{
				`log.info("User logged in", {user_id = 123})`,
			},
		},
	}
}

func (g *LuaDocGenerator) createTestingFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "describe",
			Description: "Create a test group",
			Parameters: []Parameter{
				{Name: "name", Type: "string", Description: "Test group name"},
				{Name: "tests", Type: "function", Description: "Test functions"},
			},
			Returns: []ReturnValue{},
			Examples: []string{
				`testing.describe("Math tests", function() ... end)`,
			},
		},
	}
}

func (g *LuaDocGenerator) createCoreFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "is_callable",
			Description: "Check if a value is callable (function)",
			Parameters: []Parameter{
				{Name: "value", Type: "any", Description: "Value to check"},
			},
			Returns: []ReturnValue{
				{Type: "boolean", Description: "true if value is callable"},
			},
			Examples: []string{
				`local callable = core.is_callable(my_function)`,
			},
		},
	}
}

func (g *LuaDocGenerator) createSpellFunctions() []APIFunction {
	return []APIFunction{
		{
			Name:        "init",
			Description: "Initialize spell with configuration",
			Parameters: []Parameter{
				{Name: "config", Type: "table", Description: "Spell configuration"},
			},
			Returns: []ReturnValue{},
			Examples: []string{
				`spell.init({name = "my-spell", version = "1.0.0"})`,
			},
		},
	}
}

// GenerateMarkdownDocs generates markdown documentation
func (g *LuaDocGenerator) GenerateMarkdownDocs(modules []LuaModule) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# Lua API Documentation\n\n")
	sb.WriteString(fmt.Sprintf("Generated on %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Table of contents
	sb.WriteString("## Table of Contents\n\n")
	for _, module := range modules {
		sb.WriteString(fmt.Sprintf("- [%s](#%s)\n", module.Name, strings.ToLower(module.Name)))
	}
	sb.WriteString("\n")

	// Module documentation
	for _, module := range modules {
		sb.WriteString(fmt.Sprintf("## %s\n\n", module.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", module.Description))

		if len(module.Functions) > 0 {
			sb.WriteString("### Functions\n\n")
			for _, function := range module.Functions {
				g.writeMarkdownFunction(&sb, function)
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// writeMarkdownFunction writes a function in markdown format
func (g *LuaDocGenerator) writeMarkdownFunction(sb *strings.Builder, function APIFunction) {
	// Function signature
	params := make([]string, len(function.Parameters))
	for i, param := range function.Parameters {
		if param.Optional {
			params[i] = fmt.Sprintf("[%s]", param.Name)
		} else {
			params[i] = param.Name
		}
	}

	signature := fmt.Sprintf("%s.%s(%s)", function.Module, function.Name, strings.Join(params, ", "))
	sb.WriteString(fmt.Sprintf("#### %s\n\n", signature))

	// Description
	sb.WriteString(fmt.Sprintf("%s\n\n", function.Description))

	// Parameters
	if len(function.Parameters) > 0 {
		sb.WriteString("**Parameters:**\n\n")
		for _, param := range function.Parameters {
			optional := ""
			if param.Optional {
				optional = " (optional)"
			}
			sb.WriteString(fmt.Sprintf("- `%s` (%s)%s: %s\n", param.Name, param.Type, optional, param.Description))
		}
		sb.WriteString("\n")
	}

	// Returns
	if len(function.Returns) > 0 {
		sb.WriteString("**Returns:**\n\n")
		for _, ret := range function.Returns {
			sb.WriteString(fmt.Sprintf("- (%s): %s\n", ret.Type, ret.Description))
		}
		sb.WriteString("\n")
	}

	// Examples
	if len(function.Examples) > 0 {
		sb.WriteString("**Example:**\n\n")
		for _, example := range function.Examples {
			sb.WriteString(fmt.Sprintf("```lua\n%s\n```\n\n", example))
		}
	}
}

// GenerateJSONDocs generates JSON documentation
func (g *LuaDocGenerator) GenerateJSONDocs(modules []LuaModule) (string, error) {
	docs := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"version":      "1.0.0",
		"modules":      modules,
	}

	jsonData, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonData), nil
}

// GenerateCompletionData generates IDE completion data
func (g *LuaDocGenerator) GenerateCompletionData(modules []LuaModule) CompletionData {
	completion := CompletionData{
		Functions: []CompletionFunction{},
		Modules:   []CompletionModule{},
		Types:     []CompletionType{},
		Keywords:  []string{"local", "function", "end", "if", "then", "else", "for", "while", "do", "return"},
	}

	for _, module := range modules {
		// Add module
		moduleCompletion := CompletionModule{
			Name:        module.Name,
			Description: module.Description,
			Functions:   []string{},
		}

		// Add functions
		for _, function := range module.Functions {
			// Add to module function list
			moduleCompletion.Functions = append(moduleCompletion.Functions, function.Name)

			// Create completion function
			params := make([]CompletionParam, len(function.Parameters))
			for i, param := range function.Parameters {
				params[i] = CompletionParam{
					Name:     param.Name,
					Type:     param.Type,
					Optional: param.Optional,
				}
			}

			returnType := "any"
			if len(function.Returns) > 0 {
				returnType = function.Returns[0].Type
			}

			completionFunc := CompletionFunction{
				Name:       function.Name,
				Module:     module.Name,
				Signature:  fmt.Sprintf("%s.%s", module.Name, function.Name),
				ReturnType: returnType,
				Parameters: params,
				Snippets:   []CompletionSnippet{},
			}

			// Add snippet
			if len(function.Examples) > 0 {
				snippet := CompletionSnippet{
					Trigger:     fmt.Sprintf("%s.%s", module.Name, function.Name),
					Description: function.Description,
					Body:        function.Examples[0],
				}
				completionFunc.Snippets = append(completionFunc.Snippets, snippet)
			}

			completion.Functions = append(completion.Functions, completionFunc)
		}

		completion.Modules = append(completion.Modules, moduleCompletion)
	}

	return completion
}

// Generate generates all documentation formats
func (g *LuaDocGenerator) Generate() error {
	// Extract APIs from bridges
	bridgeModules, err := g.ExtractBridgeAPIs()
	if err != nil {
		return fmt.Errorf("failed to extract bridge APIs: %w", err)
	}

	// Extract APIs from stdlib
	stdlibModules, err := g.ExtractStdlibAPIs()
	if err != nil {
		return fmt.Errorf("failed to extract stdlib APIs: %w", err)
	}

	// Combine all modules
	allModules := append(bridgeModules, stdlibModules...)

	// Sort modules by name
	sort.Slice(allModules, func(i, j int) bool {
		return allModules[i].Name < allModules[j].Name
	})

	// Generate markdown documentation
	markdownDocs := g.GenerateMarkdownDocs(allModules)
	fmt.Printf("Generated Markdown documentation (%d bytes)\n", len(markdownDocs))

	// Generate JSON documentation
	jsonDocs, err := g.GenerateJSONDocs(allModules)
	if err != nil {
		return fmt.Errorf("failed to generate JSON docs: %w", err)
	}
	fmt.Printf("Generated JSON documentation (%d bytes)\n", len(jsonDocs))

	// Generate completion data
	completionData := g.GenerateCompletionData(allModules)
	completionJSON, err := json.MarshalIndent(completionData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal completion data: %w", err)
	}
	fmt.Printf("Generated completion data (%d bytes)\n", len(completionJSON))

	return nil
}

// convertLuaModuleToGeneric converts a Lua-specific module to generic module
func (g *LuaDocGenerator) convertLuaModuleToGeneric(luaModule LuaModule) Module {
	module := Module{
		Name:        luaModule.Name,
		Language:    "lua",
		Description: luaModule.Description,
		Constants:   luaModule.Constants,
		SeeAlso:     luaModule.SeeAlso,
		Since:       luaModule.Since,
	}

	// Convert functions
	for _, luaFunc := range luaModule.Functions {
		function := Function{
			Name:        luaFunc.Name,
			Module:      luaFunc.Module,
			Description: luaFunc.Description,
			Examples:    luaFunc.Examples,
			SeeAlso:     luaFunc.SeeAlso,
			Since:       luaFunc.Since,
			Deprecated:  luaFunc.Deprecated,
			Tags:        luaFunc.Tags,
			Metadata:    luaFunc.Metadata,
		}

		// Convert parameters
		for _, param := range luaFunc.Parameters {
			function.Parameters = append(function.Parameters, Parameter{
				Name:        param.Name,
				Type:        param.Type,
				Description: param.Description,
				Optional:    param.Optional,
				Default:     param.Default,
			})
		}

		// Convert returns
		for _, ret := range luaFunc.Returns {
			function.Returns = append(function.Returns, ReturnValue{
				Type:        ret.Type,
				Description: ret.Description,
			})
		}

		module.Functions = append(module.Functions, function)
	}

	// Convert types
	for _, luaType := range luaModule.Types {
		typ := Type{
			Name:        luaType.Name,
			Description: luaType.Description,
			Examples:    luaType.Examples,
			Metadata:    luaType.Metadata,
		}

		// Convert fields
		for _, field := range luaType.Fields {
			typ.Fields = append(typ.Fields, Field{
				Name:        field.Name,
				Type:        field.Type,
				Description: field.Description,
				Optional:    field.Optional,
				Default:     field.Default,
			})
		}

		// Convert methods
		for _, method := range luaType.Methods {
			typ.Methods = append(typ.Methods, g.convertLuaFunctionToGeneric(method))
		}

		module.Types = append(module.Types, typ)
	}

	// Convert examples
	for _, example := range luaModule.Examples {
		module.Examples = append(module.Examples, ModuleExample{
			Title:       example.Title,
			Description: example.Description,
			Code:        example.Code,
			Output:      example.Output,
		})
	}

	return module
}

// convertGenericToLuaModule converts a generic module to Lua-specific module
func (g *LuaDocGenerator) convertGenericToLuaModule(module Module) LuaModule {
	luaModule := LuaModule{
		Name:        module.Name,
		Description: module.Description,
		Constants:   module.Constants,
		SeeAlso:     module.SeeAlso,
		Since:       module.Since,
	}

	// Convert functions
	for _, function := range module.Functions {
		luaModule.Functions = append(luaModule.Functions, g.convertGenericFunctionToLua(function))
	}

	// Convert types
	for _, typ := range module.Types {
		luaType := TypeDefinition{
			Name:        typ.Name,
			Description: typ.Description,
			Examples:    typ.Examples,
			Metadata:    typ.Metadata,
		}

		// Convert fields
		for _, field := range typ.Fields {
			luaType.Fields = append(luaType.Fields, Field{
				Name:        field.Name,
				Type:        field.Type,
				Description: field.Description,
				Optional:    field.Optional,
				Default:     field.Default,
			})
		}

		// Convert methods
		for _, method := range typ.Methods {
			luaType.Methods = append(luaType.Methods, g.convertGenericFunctionToLua(method))
		}

		luaModule.Types = append(luaModule.Types, luaType)
	}

	// Convert examples
	for _, example := range module.Examples {
		luaModule.Examples = append(luaModule.Examples, ModuleExample{
			Title:       example.Title,
			Description: example.Description,
			Code:        example.Code,
			Output:      example.Output,
		})
	}

	return luaModule
}

// convertLuaFunctionToGeneric converts a Lua function to generic function
func (g *LuaDocGenerator) convertLuaFunctionToGeneric(luaFunc APIFunction) Function {
	function := Function{
		Name:        luaFunc.Name,
		Module:      luaFunc.Module,
		Description: luaFunc.Description,
		Examples:    luaFunc.Examples,
		SeeAlso:     luaFunc.SeeAlso,
		Since:       luaFunc.Since,
		Deprecated:  luaFunc.Deprecated,
		Tags:        luaFunc.Tags,
		Metadata:    luaFunc.Metadata,
	}

	// Convert parameters
	for _, param := range luaFunc.Parameters {
		function.Parameters = append(function.Parameters, Parameter{
			Name:        param.Name,
			Type:        param.Type,
			Description: param.Description,
			Optional:    param.Optional,
			Default:     param.Default,
		})
	}

	// Convert returns
	for _, ret := range luaFunc.Returns {
		function.Returns = append(function.Returns, ReturnValue{
			Type:        ret.Type,
			Description: ret.Description,
		})
	}

	return function
}

// convertGenericFunctionToLua converts a generic function to Lua function
func (g *LuaDocGenerator) convertGenericFunctionToLua(function Function) APIFunction {
	luaFunc := APIFunction{
		Name:        function.Name,
		Module:      function.Module,
		Description: function.Description,
		Examples:    function.Examples,
		SeeAlso:     function.SeeAlso,
		Since:       function.Since,
		Deprecated:  function.Deprecated,
		Tags:        function.Tags,
		Metadata:    function.Metadata,
	}

	// Convert parameters
	for _, param := range function.Parameters {
		luaFunc.Parameters = append(luaFunc.Parameters, Parameter{
			Name:        param.Name,
			Type:        param.Type,
			Description: param.Description,
			Optional:    param.Optional,
			Default:     param.Default,
		})
	}

	// Convert returns
	for _, ret := range function.Returns {
		luaFunc.Returns = append(luaFunc.Returns, ReturnValue{
			Type:        ret.Type,
			Description: ret.Description,
		})
	}

	return luaFunc
}
