// ABOUTME: Tests for the Tool interface and FunctionTool implementation
// ABOUTME: Verifies basic tool functionality and parameter handling

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestFunctionTool(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		description string
		parameters  string
		fn          ToolFunc
		input       map[string]interface{}
		expected    interface{}
		expectError bool
	}{
		{
			name:        "simple calculator add",
			toolName:    "add",
			description: "Adds two numbers",
			parameters:  `{"type":"object","properties":{"a":{"type":"number"},"b":{"type":"number"}},"required":["a","b"]}`,
			fn: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				a, ok := params["a"].(float64)
				if !ok {
					return nil, errors.New("parameter 'a' must be a number")
				}
				b, ok := params["b"].(float64)
				if !ok {
					return nil, errors.New("parameter 'b' must be a number")
				}
				return a + b, nil
			},
			input:       map[string]interface{}{"a": float64(5), "b": float64(3)},
			expected:    float64(8),
			expectError: false,
		},
		{
			name:        "string concatenation",
			toolName:    "concat",
			description: "Concatenates two strings",
			parameters:  `{"type":"object","properties":{"str1":{"type":"string"},"str2":{"type":"string"}},"required":["str1","str2"]}`,
			fn: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				str1, ok := params["str1"].(string)
				if !ok {
					return nil, errors.New("parameter 'str1' must be a string")
				}
				str2, ok := params["str2"].(string)
				if !ok {
					return nil, errors.New("parameter 'str2' must be a string")
				}
				return str1 + str2, nil
			},
			input:       map[string]interface{}{"str1": "hello", "str2": "world"},
			expected:    "helloworld",
			expectError: false,
		},
		{
			name:        "missing parameter",
			toolName:    "multiply",
			description: "Multiplies two numbers",
			parameters:  `{"type":"object","properties":{"x":{"type":"number"},"y":{"type":"number"}},"required":["x","y"]}`,
			fn: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				x, ok := params["x"].(float64)
				if !ok {
					return nil, errors.New("parameter 'x' must be a number")
				}
				y, ok := params["y"].(float64)
				if !ok {
					return nil, errors.New("parameter 'y' must be a number")
				}
				return x * y, nil
			},
			input:       map[string]interface{}{"x": float64(5)}, // missing y
			expected:    nil,
			expectError: true,
		},
		{
			name:        "context cancellation",
			toolName:    "slow_operation",
			description: "A slow operation that respects context",
			parameters:  `{"type":"object","properties":{}}`,
			fn: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return "completed", nil
				}
			},
			input:       map[string]interface{}{},
			expected:    "completed",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewFunctionTool(tt.toolName, tt.description, json.RawMessage(tt.parameters), tt.fn)

			// Test Name
			if tool.Name() != tt.toolName {
				t.Errorf("Name() = %v, want %v", tool.Name(), tt.toolName)
			}

			// Test Description
			if tool.Description() != tt.description {
				t.Errorf("Description() = %v, want %v", tool.Description(), tt.description)
			}

			// Test Parameters
			if string(tool.Parameters()) != tt.parameters {
				t.Errorf("Parameters() = %v, want %v", string(tool.Parameters()), tt.parameters)
			}

			// Test Execute
			ctx := context.Background()
			result, err := tool.Execute(ctx, tt.input)

			if (err != nil) != tt.expectError {
				t.Errorf("Execute() error = %v, expectError %v", err, tt.expectError)
			}

			if !tt.expectError && result != tt.expected {
				t.Errorf("Execute() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	tool := NewFunctionTool(
		"cancellable",
		"A cancellable operation",
		json.RawMessage(`{"type":"object"}`),
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return "should not reach here in test", nil
			}
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := tool.Execute(ctx, map[string]interface{}{})
	if err == nil {
		t.Error("Expected context cancellation error")
	}
}

func TestResult(t *testing.T) {
	// Test successful result
	successResult := Result{
		Success: true,
		Data:    "test data",
		Error:   "",
	}

	if !successResult.Success {
		t.Error("Expected success to be true")
	}

	// Test error result
	errorResult := Result{
		Success: false,
		Data:    nil,
		Error:   "test error",
	}

	if errorResult.Success {
		t.Error("Expected success to be false")
	}

	// Test JSON marshaling
	data, err := json.Marshal(successResult)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var unmarshaledResult Result
	err = json.Unmarshal(data, &unmarshaledResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if unmarshaledResult.Success != successResult.Success {
		t.Error("Success field not preserved through JSON marshaling")
	}
}

func TestMetadata(t *testing.T) {
	metadata := Metadata{
		Name:        "test-tool",
		Description: "A test tool",
		Version:     "1.0.0",
		Author:      "Test Author",
		Tags:        []string{"test", "example"},
		Parameters:  json.RawMessage(`{"type":"object"}`),
	}

	// Test JSON marshaling
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	var unmarshaledMetadata Metadata
	err = json.Unmarshal(data, &unmarshaledMetadata)
	if err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if unmarshaledMetadata.Name != metadata.Name {
		t.Errorf("Name = %v, want %v", unmarshaledMetadata.Name, metadata.Name)
	}

	if len(unmarshaledMetadata.Tags) != len(metadata.Tags) {
		t.Errorf("Tags length = %v, want %v", len(unmarshaledMetadata.Tags), len(metadata.Tags))
	}
}
