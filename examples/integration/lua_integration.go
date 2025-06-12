//go:build to_be_migrated

// ABOUTME: Example demonstrating Lua engine integration with LLM bridge
// ABOUTME: Shows how to run Lua scripts that interact with LLMs

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua"
)

func main() {
	// Create Lua engine
	config := &engine.Config{
		MaxExecutionTime: 30,               // 30 seconds
		MaxMemory:        64 * 1024 * 1024, // 64MB
	}

	eng, err := lua.NewLuaEngine(config)
	if err != nil {
		log.Fatalf("Failed to create Lua engine: %v", err)
	}
	defer eng.Close()

	// Create LLM bridge (this will auto-detect available providers)
	llmBridge, err := bridge.NewLLMBridge()
	if err != nil {
		// If no API keys are set, create a mock for demonstration
		fmt.Printf("Note: %v\n", err)
		fmt.Println("Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY to use real providers.")
		fmt.Println("\nFor this demo, we'll show the Lua integration without actual LLM calls.")

		// For demo purposes, we'll just register mock functions
		registerMockLLMFunctions(eng)
	} else {
		// For this demo, we'll just show that the bridge was created
		fmt.Printf("LLM Bridge initialized with provider: %s\n", llmBridge.GetCurrentProvider())
		fmt.Println("(Note: Full LLM integration requires additional setup)")
	}

	// Example 1: Simple Lua script
	fmt.Println("=== Example 1: Simple Lua Script ===")
	simpleScript := `
		print("Hello from Lua!")
		local x = 10
		local y = 20
		print("x + y = " .. (x + y))
	`

	err = eng.LoadScript(strings.NewReader(simpleScript))
	if err != nil {
		log.Fatalf("Failed to load script: %v", err)
	}

	err = eng.Execute(context.Background())
	if err != nil {
		log.Fatalf("Failed to execute script: %v", err)
	}

	// Example 2: Using registered Go functions
	fmt.Println("\n=== Example 2: Calling Go Functions from Lua ===")

	// Register a simple Go function
	err = eng.RegisterFunction("greet", func(name string) string {
		return "Hello, " + name + "!"
	})
	if err != nil {
		log.Fatalf("Failed to register function: %v", err)
	}

	functionScript := `
		local greeting = greet("Lua User")
		print(greeting)
	`

	err = eng.LoadScript(strings.NewReader(functionScript))
	if err != nil {
		log.Fatalf("Failed to load script: %v", err)
	}

	err = eng.Execute(context.Background())
	if err != nil {
		log.Fatalf("Failed to execute script: %v", err)
	}

	// Example 3: Variable exchange
	fmt.Println("\n=== Example 3: Variable Exchange ===")

	// Set variables from Go (using simple types for now)
	err = eng.SetVariable("apiKey", "demo-key")
	if err != nil {
		log.Fatalf("Failed to set apiKey: %v", err)
	}
	err = eng.SetVariable("model", "gpt-3.5-turbo")
	if err != nil {
		log.Fatalf("Failed to set model: %v", err)
	}
	err = eng.SetVariable("temperature", 0.7)
	if err != nil {
		log.Fatalf("Failed to set temperature: %v", err)
	}

	variableScript := `
		print("Configuration:")
		print("  API Key: " .. apiKey)
		print("  Model: " .. model)
		print("  Temperature: " .. temperature)
		
		-- Set a result variable
		result = "Script completed successfully"
	`

	err = eng.LoadScript(strings.NewReader(variableScript))
	if err != nil {
		log.Fatalf("Failed to load script: %v", err)
	}

	err = eng.Execute(context.Background())
	if err != nil {
		log.Fatalf("Failed to execute script: %v", err)
	}

	// Get variable from Lua
	result, err := eng.GetVariable("result")
	if err != nil {
		log.Fatalf("Failed to get variable: %v", err)
	}
	fmt.Printf("Result from Lua: %v\n", result)

	// Example 4: Security sandbox demonstration
	fmt.Println("\n=== Example 4: Security Sandbox ===")

	securityScript := `
		-- These should fail due to security restrictions
		print("Attempting to use restricted functions...")
		
		-- Try to use io (should fail)
		local ok, err = pcall(function()
			io.open("test.txt", "w")
		end)
		if not ok then
			print("✓ io.open blocked: " .. tostring(err))
		end
		
		-- Try to use os.execute (should fail)
		ok, err = pcall(function()
			os.execute("ls")
		end)
		if not ok then
			print("✓ os.execute blocked: " .. tostring(err))
		end
		
		-- Try to use dofile (should fail)
		ok, err = pcall(function()
			dofile("malicious.lua")
		end)
		if not ok then
			print("✓ dofile blocked: " .. tostring(err))
		end
		
		print("Security sandbox is working correctly!")
	`

	err = eng.LoadScript(strings.NewReader(securityScript))
	if err != nil {
		log.Fatalf("Failed to load script: %v", err)
	}

	err = eng.Execute(context.Background())
	if err != nil {
		log.Fatalf("Failed to execute script: %v", err)
	}

	fmt.Println("\n=== Demo Complete ===")
}

// registerMockLLMFunctions registers mock LLM functions for demonstration
func registerMockLLMFunctions(eng *lua.LuaEngine) {
	// Mock chat function
	eng.RegisterFunction("llm_chat", func(prompt string) string {
		return "[Mock Response] You said: " + prompt
	})

	// Mock completion function
	eng.RegisterFunction("llm_complete", func(prompt string, maxTokens int) string {
		return prompt + " [Mock Completion]"
	})

	fmt.Println("Registered mock LLM functions for demonstration.")
}
