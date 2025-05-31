// ABOUTME: Tests for the tool registry implementation
// ABOUTME: Verifies thread-safety, registration, lookup, and removal operations

package tools

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
)

func createTestTool(name, description string) Tool {
	return NewFunctionTool(
		name,
		description,
		json.RawMessage(`{"type":"object"}`),
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	)
}

func TestRegistry(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		reg := NewRegistry()

		// Test registering a tool
		tool1 := createTestTool("tool1", "Test tool 1")
		err := reg.Register(tool1)
		if err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		// Test getting a registered tool
		retrieved, err := reg.Get("tool1")
		if err != nil {
			t.Fatalf("Failed to get tool: %v", err)
		}
		if retrieved.Name() != "tool1" {
			t.Errorf("Retrieved tool has wrong name: %v", retrieved.Name())
		}

		// Test listing tools
		tools := reg.List()
		if len(tools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(tools))
		}

		// Test removing a tool
		err = reg.Remove("tool1")
		if err != nil {
			t.Fatalf("Failed to remove tool: %v", err)
		}

		// Verify tool is removed
		_, err = reg.Get("tool1")
		if err == nil {
			t.Error("Expected error when getting removed tool")
		}
	})

	t.Run("error cases", func(t *testing.T) {
		reg := NewRegistry()

		// Test registering nil tool
		err := reg.Register(nil)
		if err == nil {
			t.Error("Expected error when registering nil tool")
		}

		// Test registering tool with empty name
		emptyNameTool := NewFunctionTool(
			"",
			"Empty name tool",
			json.RawMessage(`{}`),
			func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				return nil, nil
			},
		)
		err = reg.Register(emptyNameTool)
		if err == nil {
			t.Error("Expected error when registering tool with empty name")
		}

		// Test duplicate registration
		tool := createTestTool("duplicate", "Duplicate tool")
		err = reg.Register(tool)
		if err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}
		err = reg.Register(tool)
		if err == nil {
			t.Error("Expected error when registering duplicate tool")
		}

		// Test getting non-existent tool
		_, err = reg.Get("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent tool")
		}

		// Test removing non-existent tool
		err = reg.Remove("nonexistent")
		if err == nil {
			t.Error("Expected error when removing non-existent tool")
		}
	})

	t.Run("multiple tools", func(t *testing.T) {
		reg := NewRegistry()

		// Register multiple tools
		for i := 0; i < 5; i++ {
			tool := createTestTool(
				string(rune('a'+i)),
				"Test tool",
			)
			err := reg.Register(tool)
			if err != nil {
				t.Fatalf("Failed to register tool %d: %v", i, err)
			}
		}

		// Verify all tools are listed
		tools := reg.List()
		if len(tools) != 5 {
			t.Errorf("Expected 5 tools, got %d", len(tools))
		}

		// Verify each tool can be retrieved
		for i := 0; i < 5; i++ {
			name := string(rune('a' + i))
			tool, err := reg.Get(name)
			if err != nil {
				t.Errorf("Failed to get tool %s: %v", name, err)
			}
			if tool.Name() != name {
				t.Errorf("Wrong tool name: got %s, want %s", tool.Name(), name)
			}
		}
	})
}

func TestRegistryConcurrency(t *testing.T) {
	reg := NewRegistry()
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent registrations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				name := string(rune('a' + (id*numOperations+j)%26))
				tool := createTestTool(name, "Concurrent tool")
				_ = reg.Register(tool) // Ignore errors from duplicates
			}
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				_ = reg.List()
				name := string(rune('a' + j%26))
				_, _ = reg.Get(name)
			}
		}()
	}
	wg.Wait()

	// Concurrent removals
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				name := string(rune('a' + (id*numOperations+j)%26))
				_ = reg.Remove(name) // Ignore errors from non-existent tools
			}
		}(i)
	}
	wg.Wait()
}

func TestDefaultRegistry(t *testing.T) {
	// Clear any existing tools in default registry
	for _, tool := range List() {
		_ = Remove(tool.Name())
	}

	// Test global functions
	tool := createTestTool("global", "Global tool")

	err := Register(tool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	retrieved, err := Get("global")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}
	if retrieved.Name() != "global" {
		t.Errorf("Retrieved tool has wrong name: %v", retrieved.Name())
	}

	tools := List()
	found := false
	for _, t := range tools {
		if t.Name() == "global" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Global tool not found in list")
	}

	err = Remove("global")
	if err != nil {
		t.Fatalf("Failed to remove tool: %v", err)
	}
}
