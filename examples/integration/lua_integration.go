// ABOUTME: Example demonstrating Lua engine integration
// ABOUTME: Shows how to run Lua scripts using the current engine architecture

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

func main() {
	// Create Lua engine
	config := engine.EngineConfig{
		TimeoutLimit: 30 * time.Second,
		MemoryLimit:  64 * 1024 * 1024, // 64MB
		SandboxMode:  true,
	}

	eng := gopherlua.NewLuaEngine()
	if err := eng.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize Lua engine: %v", err)
	}
	defer func() {
		if err := eng.Shutdown(); err != nil {
			log.Printf("Failed to shutdown engine: %v", err)
		}
	}()

	fmt.Println("=== Go-LLMSpell Lua Engine Integration Demo ===")
	fmt.Println("This demo shows basic Lua script execution using the current engine.")

	// Example 1: Simple Lua script
	fmt.Println("\n=== Example 1: Simple Lua Script ===")
	simpleScript := `
		local x = 10
		local y = 20
		return "Result: " .. (x + y)
	`

	ctx := context.Background()
	result, err := eng.Execute(ctx, simpleScript, nil)
	if err != nil {
		log.Fatalf("Failed to execute script: %v", err)
	}

	if result != nil {
		fmt.Printf("Script result: %v\n", result.ToGo())
	}

	// Example 2: Parameter passing
	fmt.Println("\n=== Example 2: Parameter Passing ===")
	paramScript := `
		return "Hello, " .. name .. "! You are " .. age .. " years old."
	`

	params := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}

	result, err = eng.Execute(ctx, paramScript, params)
	if err != nil {
		log.Fatalf("Failed to execute script with params: %v", err)
	}

	fmt.Printf("Script with params result: %v\n", result.ToGo())

	// Example 3: Error handling
	fmt.Println("\n=== Example 3: Error Handling ===")
	errorScript := `
		error("This is a test error")
	`

	_, err = eng.Execute(ctx, errorScript, nil)
	if err != nil {
		fmt.Printf("Caught expected error: %v\n", err)
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("The Lua engine integration is working correctly!")
}
