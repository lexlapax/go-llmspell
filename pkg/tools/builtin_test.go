package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

func TestLLMSToolAdapter(t *testing.T) {
	// Create a WebFetch tool from go-llms
	webFetchTool := tools.MustGetTool("web_fetch")

	// Wrap it with our adapter
	adapter := NewLLMSToolAdapter(webFetchTool)

	// Test Name
	if adapter.Name() != "web_fetch" {
		t.Errorf("Expected name 'web_fetch', got %s", adapter.Name())
	}

	// Test Description
	expectedDesc := "Fetches content from a URL with customizable timeout"
	if adapter.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got %s", expectedDesc, adapter.Description())
	}

	// Test Parameters
	params := adapter.Parameters()
	var paramSchema map[string]interface{}
	if err := json.Unmarshal(params, &paramSchema); err != nil {
		t.Fatalf("Failed to unmarshal parameters: %v", err)
	}

	// Check schema structure
	if paramSchema["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", paramSchema["type"])
	}

	props, ok := paramSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	urlProp, ok := props["url"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected url property")
	}

	if urlProp["type"] != "string" {
		t.Errorf("Expected url type 'string', got %v", urlProp["type"])
	}

	if urlProp["format"] != "uri" {
		t.Errorf("Expected url format 'uri', got %v", urlProp["format"])
	}

	required, ok := paramSchema["required"].([]interface{})
	if !ok {
		t.Fatal("Expected required to be an array")
	}

	if len(required) != 1 || required[0] != "url" {
		t.Errorf("Expected required ['url'], got %v", required)
	}
}

func TestRegisterBuiltinTools(t *testing.T) {
	tests := []struct {
		name          string
		config        *BuiltinToolConfig
		expectedTools []string
		expectError   bool
	}{
		{
			name:          "Default config",
			config:        nil,
			expectedTools: []string{"web_fetch"},
			expectError:   false,
		},
		{
			name: "All safe tools enabled",
			config: &BuiltinToolConfig{
				EnableWebFetch:       true,
				EnableSearch:         false,
				EnableExecuteCommand: false,
				EnableReadFile:       false,
				EnableWriteFile:      false,
			},
			expectedTools: []string{"web_fetch"},
			expectError:   false,
		},
		{
			name: "All tools enabled except search",
			config: &BuiltinToolConfig{
				EnableWebFetch:       true,
				EnableSearch:         false,
				EnableExecuteCommand: true,
				EnableReadFile:       true,
				EnableWriteFile:      true,
			},
			expectedTools: []string{"web_fetch", "execute_command", "file_read", "file_write"},
			expectError:   false,
		},
		{
			name: "Search enabled (should error)",
			config: &BuiltinToolConfig{
				EnableWebFetch:       false,
				EnableSearch:         true,
				EnableExecuteCommand: false,
				EnableReadFile:       false,
				EnableWriteFile:      false,
			},
			expectedTools: []string{},
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()

			err := RegisterBuiltinTools(registry, tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check registered tools
			registeredTools := registry.List()

			// Create a map for easier checking
			toolMap := make(map[string]bool)
			for _, tool := range registeredTools {
				toolMap[tool.Name()] = true
			}

			// Check expected tools are present
			for _, expectedTool := range tt.expectedTools {
				if !toolMap[expectedTool] {
					t.Errorf("Expected tool %s not found in registry", expectedTool)
				}
			}

			// Check no unexpected tools are present
			if len(registeredTools) != len(tt.expectedTools) {
				t.Errorf("Expected %d tools, got %d", len(tt.expectedTools), len(registeredTools))
			}
		})
	}
}

func TestBuiltinToolExecution(t *testing.T) {
	// Test WebFetch tool execution (this will fail without network)
	registry := NewRegistry()

	config := &BuiltinToolConfig{
		EnableWebFetch: true,
	}

	if err := RegisterBuiltinTools(registry, config); err != nil {
		t.Fatalf("Failed to register builtin tools: %v", err)
	}

	webFetchTool, err := registry.Get("web_fetch")
	if err != nil {
		t.Fatalf("Failed to get web_fetch tool: %v", err)
	}

	// Test with invalid URL (should fail)
	ctx := context.Background()
	params := map[string]interface{}{
		"url": "not-a-valid-url",
	}

	_, err = webFetchTool.Execute(ctx, params)
	if err == nil {
		t.Error("Expected error for invalid URL, got none")
	}
}

func TestSchemaConversion(t *testing.T) {
	// Test all builtin tools have valid schema conversion
	registry := NewRegistry()

	// Register all tools
	config := &BuiltinToolConfig{
		EnableWebFetch:       true,
		EnableExecuteCommand: true,
		EnableReadFile:       true,
		EnableWriteFile:      true,
		EnableSearch:         false,
	}

	if err := RegisterBuiltinTools(registry, config); err != nil {
		t.Fatalf("Failed to register tools: %v", err)
	}

	// Test each registered tool
	for _, tool := range registry.List() {
		t.Run(tool.Name(), func(t *testing.T) {
			// Check that Parameters() returns valid JSON
			params := tool.Parameters()
			var schema map[string]interface{}
			if err := json.Unmarshal(params, &schema); err != nil {
				t.Errorf("Failed to unmarshal parameters for %s: %v", tool.Name(), err)
			}

			// Basic schema validation
			if schema["type"] != "object" {
				t.Errorf("Expected schema type 'object' for %s, got %v", tool.Name(), schema["type"])
			}
		})
	}
}
