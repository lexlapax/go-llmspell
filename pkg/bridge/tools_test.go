// ABOUTME: Tests for the tool bridge implementation
// ABOUTME: Verifies tool registration, execution, and validation from scripts

package bridge

import (
	"context"
	"errors"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/tools"
)

func TestToolBridge(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		registry := tools.NewRegistry()
		bridge := NewToolBridge(registry)

		// Test registering a tool from script
		params := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"x": map[string]interface{}{"type": "number"},
				"y": map[string]interface{}{"type": "number"},
			},
			"required": []interface{}{"x", "y"},
		}

		err := bridge.RegisterTool(
			"add",
			"Adds two numbers",
			params,
			func(p map[string]interface{}) (interface{}, error) {
				x, ok1 := p["x"].(float64)
				y, ok2 := p["y"].(float64)
				if !ok1 || !ok2 {
					return nil, errors.New("invalid parameters")
				}
				return x + y, nil
			},
		)
		if err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		// Test executing the tool
		ctx := context.Background()
		result, err := bridge.ExecuteTool(ctx, "add", map[string]interface{}{
			"x": float64(5),
			"y": float64(3),
		})
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}
		if result != float64(8) {
			t.Errorf("Expected 8, got %v", result)
		}

		// Test getting tool info
		info, err := bridge.GetTool("add")
		if err != nil {
			t.Fatalf("Failed to get tool info: %v", err)
		}
		if info["name"] != "add" {
			t.Errorf("Expected name 'add', got %v", info["name"])
		}

		// Test listing tools
		tools := bridge.ListTools()
		if len(tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(tools))
		}

		// Test removing tool
		err = bridge.RemoveTool("add")
		if err != nil {
			t.Fatalf("Failed to remove tool: %v", err)
		}

		// Verify tool is removed
		_, err = bridge.GetTool("add")
		if err == nil {
			t.Error("Expected error when getting removed tool")
		}
	})

	t.Run("parameter validation", func(t *testing.T) {
		registry := tools.NewRegistry()
		bridge := NewToolBridge(registry)

		// Register a tool with parameter schema
		params := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name":   map[string]interface{}{"type": "string"},
				"age":    map[string]interface{}{"type": "number"},
				"active": map[string]interface{}{"type": "boolean"},
			},
			"required": []interface{}{"name", "age"},
		}

		err := bridge.RegisterTool(
			"user_info",
			"Process user information",
			params,
			func(p map[string]interface{}) (interface{}, error) {
				return "processed", nil
			},
		)
		if err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		// Test valid parameters
		err = bridge.ValidateParameters("user_info", map[string]interface{}{
			"name":   "Alice",
			"age":    float64(30),
			"active": true,
		})
		if err != nil {
			t.Errorf("Valid parameters failed validation: %v", err)
		}

		// Test missing required parameter
		err = bridge.ValidateParameters("user_info", map[string]interface{}{
			"name": "Bob",
			// missing age
		})
		if err == nil {
			t.Error("Expected error for missing required parameter")
		}

		// Test wrong type
		err = bridge.ValidateParameters("user_info", map[string]interface{}{
			"name": "Charlie",
			"age":  "thirty", // should be number
		})
		if err == nil {
			t.Error("Expected error for wrong type")
		}
	})

	t.Run("error handling", func(t *testing.T) {
		registry := tools.NewRegistry()
		bridge := NewToolBridge(registry)

		// Test executing non-existent tool
		ctx := context.Background()
		_, err := bridge.ExecuteTool(ctx, "nonexistent", map[string]interface{}{})
		if err == nil {
			t.Error("Expected error when executing non-existent tool")
		}

		// Test getting non-existent tool
		_, err = bridge.GetTool("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent tool")
		}

		// Test removing non-existent tool
		err = bridge.RemoveTool("nonexistent")
		if err == nil {
			t.Error("Expected error when removing non-existent tool")
		}

		// Test validating parameters for non-existent tool
		err = bridge.ValidateParameters("nonexistent", map[string]interface{}{})
		if err == nil {
			t.Error("Expected error when validating parameters for non-existent tool")
		}
	})

	t.Run("nil registry", func(t *testing.T) {
		// Test that nil registry uses default
		bridge := NewToolBridge(nil)
		if bridge.registry == nil {
			t.Error("Expected bridge to use default registry when nil provided")
		}
	})
}

func TestValidateType(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		expectedType string
		shouldError  bool
	}{
		// String tests
		{"valid string", "hello", "string", false},
		{"invalid string", 123, "string", true},

		// Number tests
		{"valid float64", float64(3.14), "number", false},
		{"valid int", 42, "number", false},
		{"invalid number", "123", "number", true},

		// Boolean tests
		{"valid boolean", true, "boolean", false},
		{"invalid boolean", 1, "boolean", true},

		// Object tests
		{"valid object", map[string]interface{}{"key": "value"}, "object", false},
		{"invalid object", []string{"a", "b"}, "object", true},

		// Array tests
		{"valid array interface", []interface{}{1, 2, 3}, "array", false},
		{"valid array string", []string{"a", "b", "c"}, "array", false},
		{"invalid array", "not an array", "array", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateType(tt.value, tt.expectedType)
			if (err != nil) != tt.shouldError {
				t.Errorf("validateType() error = %v, shouldError %v", err, tt.shouldError)
			}
		})
	}
}
