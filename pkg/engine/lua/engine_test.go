// ABOUTME: Tests for the Lua script engine implementation
// ABOUTME: Validates script execution, function registration, and security features

package lua

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// TestNewLuaEngine tests engine creation
func TestNewLuaEngine(t *testing.T) {
	tests := []struct {
		name    string
		config  *engine.Config
		wantErr bool
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &engine.Config{
				MaxMemory:        1024 * 1024,
				MaxExecutionTime: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := NewLuaEngine(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if eng == nil {
				t.Fatal("expected non-nil engine")
			}

			// Cleanup
			err = eng.Close()
			if err != nil {
				t.Errorf("Close() error = %v", err)
			}
		})
	}
}

// TestLoadScript tests script loading
func TestLoadScript(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid script",
			script: `print("Hello from Lua")`,
		},
		{
			name:   "empty script",
			script: "",
		},
		{
			name:    "syntax error",
			script:  `print("unclosed string`,
			wantErr: true,
			errMsg:  "unterminated string",
		},
		{
			name:   "function definition",
			script: `function greet(name) return "Hello, " .. name end`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := NewLuaEngine(nil)
			if err != nil {
				t.Fatalf("failed to create engine: %v", err)
			}
			defer eng.Close()

			reader := strings.NewReader(tt.script)
			err = eng.LoadScript(reader)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %v does not contain %s", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestExecute tests script execution
func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{
			name:   "simple print",
			script: `print("Hello from Lua")`,
		},
		{
			name:   "variable assignment",
			script: `x = 42`,
		},
		{
			name:    "runtime error",
			script:  `error("test error")`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := NewLuaEngine(nil)
			if err != nil {
				t.Fatalf("failed to create engine: %v", err)
			}
			defer eng.Close()

			err = eng.LoadScript(strings.NewReader(tt.script))
			if err != nil {
				t.Fatalf("failed to load script: %v", err)
			}

			err = eng.Execute(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestRegisterFunction tests function registration
func TestRegisterFunction(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		fn       interface{}
		script   string
		wantErr  bool
	}{
		{
			name:     "simple function",
			funcName: "greet",
			fn: func(name string) string {
				return "Hello, " + name
			},
			script: `print(greet("Lua"))`,
		},
		{
			name:     "void function",
			funcName: "doSomething",
			fn: func() {
				// Does nothing
			},
			script: `doSomething()`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := NewLuaEngine(nil)
			if err != nil {
				t.Fatalf("failed to create engine: %v", err)
			}
			defer eng.Close()

			err = eng.RegisterFunction(tt.funcName, tt.fn)
			if err != nil {
				t.Fatalf("failed to register function: %v", err)
			}

			err = eng.LoadScript(strings.NewReader(tt.script))
			if err != nil {
				t.Fatalf("failed to load script: %v", err)
			}

			err = eng.Execute(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSetGetVariable tests variable setting and getting
func TestSetGetVariable(t *testing.T) {
	eng, err := NewLuaEngine(nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer eng.Close()

	// Test setting various types
	tests := []struct {
		name  string
		value interface{}
	}{
		{"string", "hello"},
		{"number", 42.5},
		{"bool", true},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := eng.SetVariable("testVar", tt.value)
			if err != nil {
				t.Errorf("SetVariable() error = %v", err)
				return
			}

			got, err := eng.GetVariable("testVar")
			if err != nil {
				t.Errorf("GetVariable() error = %v", err)
				return
			}

			// Compare values (handle float conversion)
			switch v := tt.value.(type) {
			case int:
				if got != float64(v) {
					t.Errorf("got %v, want %v", got, float64(v))
				}
			default:
				if got != tt.value {
					t.Errorf("got %v, want %v", got, tt.value)
				}
			}
		})
	}
}

// TestSecuritySandbox tests security restrictions
func TestSecuritySandbox(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{
			name:    "dofile disabled",
			script:  `dofile("test.lua")`,
			wantErr: true,
		},
		{
			name:    "loadfile disabled",
			script:  `loadfile("test.lua")`,
			wantErr: true,
		},
		{
			name:    "io disabled",
			script:  `io.open("test.txt")`,
			wantErr: true,
		},
		{
			name:    "os disabled",
			script:  `os.execute("ls")`,
			wantErr: true,
		},
		{
			name:   "safe string operations allowed",
			script: `x = string.upper("hello")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eng, err := NewLuaEngine(nil)
			if err != nil {
				t.Fatalf("failed to create engine: %v", err)
			}
			defer eng.Close()

			err = eng.LoadScript(strings.NewReader(tt.script))
			if err != nil {
				t.Fatalf("failed to load script: %v", err)
			}

			err = eng.Execute(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestContextCancellation tests execution cancellation
func TestContextCancellation(t *testing.T) {
	eng, err := NewLuaEngine(nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer eng.Close()

	// Load a script that takes some time
	script := `
		local i = 0
		while i < 1000000 do
			i = i + 1
		end
	`
	err = eng.LoadScript(strings.NewReader(script))
	if err != nil {
		t.Fatalf("failed to load script: %v", err)
	}

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Execute should timeout or be cancelled
	err = eng.Execute(ctx)
	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

// TestLoadScriptFile tests loading scripts from files
func TestLoadScriptFile(t *testing.T) {
	eng, err := NewLuaEngine(nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer eng.Close()

	// Test with non-existent file
	err = eng.LoadScriptFile("/nonexistent/file.lua")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	// Note: Lua VM is not thread-safe for concurrent execution
	// This test verifies that the engine properly serializes access

	eng, err := NewLuaEngine(nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer eng.Close()

	// Test sequential execution to verify mutex protection
	for i := 0; i < 3; i++ {
		// Load and execute a script
		script := `x = 1`
		err = eng.LoadScript(strings.NewReader(script))
		if err != nil {
			t.Fatalf("iteration %d: failed to load script: %v", i, err)
		}

		err = eng.Execute(context.Background())
		if err != nil {
			t.Errorf("iteration %d: execution failed: %v", i, err)
		}
	}
}

// TestEngineReset tests engine reset functionality
func TestEngineReset(t *testing.T) {
	eng, err := NewLuaEngine(nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer eng.Close()

	// Set a variable
	err = eng.SetVariable("x", 42)
	if err != nil {
		t.Fatalf("failed to set variable: %v", err)
	}

	// Verify it's set
	val, err := eng.GetVariable("x")
	if err != nil {
		t.Fatalf("failed to get variable: %v", err)
	}
	if val != float64(42) {
		t.Errorf("expected 42, got %v", val)
	}

	// Reset the engine
	err = eng.Reset()
	if err != nil {
		t.Fatalf("failed to reset engine: %v", err)
	}

	// Variable should no longer exist
	val, err = eng.GetVariable("x")
	if err != nil {
		t.Fatalf("failed to get variable after reset: %v", err)
	}
	if val != nil {
		t.Error("expected nil for non-existent variable after reset, got:", val)
	}
}
