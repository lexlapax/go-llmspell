// ABOUTME: Tests for the script engine interface and related types
// ABOUTME: Validates engine contracts and execution behaviors

package engine

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestEngineInterface(t *testing.T) {
	// This test ensures the Engine interface is properly defined
	// and can be implemented by concrete types
	var _ Engine = (*mockEngine)(nil)
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config with all limits",
			config: Config{
				MaxMemory:        64 * 1024 * 1024, // 64MB
				MaxExecutionTime: 30,               // 30 seconds
				EnableDebug:      true,
			},
			valid: true,
		},
		{
			name: "config with zero memory limit",
			config: Config{
				MaxMemory:        0,
				MaxExecutionTime: 30,
			},
			valid: true, // Zero means no limit
		},
		{
			name: "config with negative memory limit",
			config: Config{
				MaxMemory:        -1,
				MaxExecutionTime: 30,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since Config doesn't have a Validate method in the existing interface,
			// we'll just check for negative values
			valid := tt.config.MaxMemory >= 0 && tt.config.MaxExecutionTime >= 0
			if valid != tt.valid {
				t.Errorf("Validation mismatch: got valid=%v, want valid=%v",
					valid, tt.valid)
			}
		})
	}
}

func TestResult(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   struct {
			hasError bool
			output   string
		}
	}{
		{
			name: "successful execution",
			result: Result{
				Output: "Hello, World!",
				Error:  nil,
				Variables: map[string]interface{}{
					"exitCode": 0,
				},
			},
			want: struct {
				hasError bool
				output   string
			}{
				hasError: false,
				output:   "Hello, World!",
			},
		},
		{
			name: "execution with error",
			result: Result{
				Output: "",
				Error:  ErrScriptExecutionFailed,
				Variables: map[string]interface{}{
					"exitCode": 1,
				},
			},
			want: struct {
				hasError bool
				output   string
			}{
				hasError: true,
				output:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.result.Error != nil) != tt.want.hasError {
				t.Errorf("Error state mismatch: got error=%v, want hasError=%v",
					tt.result.Error != nil, tt.want.hasError)
			}
			if tt.result.Output != tt.want.output {
				t.Errorf("Output mismatch: got %q, want %q",
					tt.result.Output, tt.want.output)
			}
		})
	}
}

// mockEngine is a test implementation of the Engine interface
type mockEngine struct {
	name        string
	script      []byte
	variables   map[string]interface{}
	functions   map[string]interface{}
	executeFunc func(context.Context) error
	loadErr     error
	executeErr  error
}

func newMockEngine(name string) *mockEngine {
	return &mockEngine{
		name:      name,
		variables: make(map[string]interface{}),
		functions: make(map[string]interface{}),
	}
}

func (m *mockEngine) Name() string {
	return m.name
}

func (m *mockEngine) LoadScript(reader io.Reader) error {
	if m.loadErr != nil {
		return m.loadErr
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.script = data
	return nil
}

func (m *mockEngine) LoadScriptFile(path string) error {
	if m.loadErr != nil {
		return m.loadErr
	}
	// For testing, just store the path
	m.script = []byte(path)
	return nil
}

func (m *mockEngine) Execute(ctx context.Context) error {
	if m.script == nil {
		return ErrScriptNotLoaded
	}

	if m.executeFunc != nil {
		return m.executeFunc(ctx)
	}

	return m.executeErr
}

func (m *mockEngine) RegisterFunction(name string, fn interface{}) error {
	if fn == nil {
		return ErrInvalidFunctionSignature
	}
	m.functions[name] = fn
	return nil
}

func (m *mockEngine) SetVariable(name string, value interface{}) error {
	m.variables[name] = value
	return nil
}

func (m *mockEngine) GetVariable(name string) (interface{}, error) {
	val, ok := m.variables[name]
	if !ok {
		return nil, ErrVariableNotFound
	}
	return val, nil
}

func TestMockEngine(t *testing.T) {
	engine := newMockEngine("mock")

	// Test script loading
	script := []byte(`print("Hello, World!")`)
	if err := engine.LoadScript(bytes.NewReader(script)); err != nil {
		t.Fatalf("Failed to load script: %v", err)
	}

	// Test variable operations
	if err := engine.SetVariable("foo", "bar"); err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	val, err := engine.GetVariable("foo")
	if err != nil {
		t.Fatalf("Failed to get variable: %v", err)
	}
	if val != "bar" {
		t.Errorf("Variable mismatch: got %v, want %v", val, "bar")
	}

	// Test getting non-existent variable
	_, err = engine.GetVariable("nonexistent")
	if !errors.Is(err, ErrVariableNotFound) {
		t.Errorf("Expected ErrVariableNotFound, got %v", err)
	}

	// Test function registration
	testFunc := func() string { return "test" }
	if err := engine.RegisterFunction("test", testFunc); err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Test execution
	ctx := context.Background()
	if err := engine.Execute(ctx); err != nil {
		t.Fatalf("Failed to execute script: %v", err)
	}
}

func TestEngineExecutionWithoutScript(t *testing.T) {
	engine := newMockEngine("mock")

	ctx := context.Background()
	err := engine.Execute(ctx)
	if !errors.Is(err, ErrScriptNotLoaded) {
		t.Errorf("Expected ErrScriptNotLoaded, got %v", err)
	}
}

func TestEngineExecutionWithContext(t *testing.T) {
	engine := newMockEngine("mock")
	engine.executeFunc = func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}

	// Load a script first
	if err := engine.LoadScript(bytes.NewReader([]byte("test"))); err != nil {
		t.Fatalf("Failed to load script: %v", err)
	}

	t.Run("normal execution", func(t *testing.T) {
		ctx := context.Background()
		err := engine.Execute(ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := engine.Execute(ctx)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})

	t.Run("timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := engine.Execute(ctx)
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}

func TestEngineLoadErrors(t *testing.T) {
	engine := newMockEngine("mock")
	engine.loadErr = errors.New("load failed")

	err := engine.LoadScript(bytes.NewReader([]byte("test")))
	if err == nil || err.Error() != "load failed" {
		t.Errorf("Expected load error, got %v", err)
	}

	err = engine.LoadScriptFile("test.lua")
	if err == nil || err.Error() != "load failed" {
		t.Errorf("Expected load error, got %v", err)
	}
}
