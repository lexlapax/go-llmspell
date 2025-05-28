// ABOUTME: Simple spell runner for testing example spells
// ABOUTME: Demonstrates how to load and execute spell files

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/bridges"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run run_spell.go <spell-directory> [param=value ...]")
		fmt.Println("Example: go run run_spell.go spells/hello-llm")
		fmt.Println("Example: go run run_spell.go spells/web-summarizer url=https://example.com")
		os.Exit(1)
	}

	spellDir := os.Args[1]
	mainScript := filepath.Join(spellDir, "main.lua")

	fmt.Printf("🧙 Running spell from %s\n\n", spellDir)

	// Create Lua engine
	config := &engine.Config{
		MaxExecutionTime: 30,
		MaxMemory:        64 * 1024 * 1024,
	}

	eng, err := lua.NewLuaEngine(config)
	if err != nil {
		log.Fatalf("Failed to create Lua engine: %v", err)
	}
	defer eng.Close()

	// Check if we should use mock LLM
	if os.Getenv("MOCK_LLM") == "true" {
		fmt.Println("🎭 Using mock LLM for demonstration")
		// Register mock functions for demo
		registerMockLLM(eng)
	} else {
		// Try to create LLM bridge
		llmBridge, err := bridge.NewLLMBridge()
		if err != nil {
			fmt.Printf("⚠️  LLM Bridge not available: %v\n", err)
			fmt.Println("   Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY to enable LLM features.")
			fmt.Println("   Running with mock LLM functions...")

			// Register mock functions for demo
			registerMockLLM(eng)
		} else {
			fmt.Printf("✅ LLM Bridge initialized with provider: %s\n\n", llmBridge.GetCurrentProvider())

			// Use the proper Lua bridge with conversions
			luaBridge := bridges.NewLLMBridge(llmBridge)
			luaState := eng.GetLuaState()
			if err := luaBridge.Register(luaState); err != nil {
				log.Fatalf("Failed to register LLM bridge: %v", err)
			}
		}
	}

	// Register standard library
	spellName := filepath.Base(spellDir)
	stdlibConfig := &stdlib.Config{
		SpellName: spellName,
		LogLevel:  slog.LevelInfo,
		Storage:   stdlib.DefaultStorageConfig(),
		HTTP:      stdlib.DefaultHTTPConfig(),
	}
	if err := stdlib.RegisterAll(eng.GetLuaState(), stdlibConfig); err != nil {
		log.Fatalf("Failed to register stdlib: %v", err)
	}

	// First, set up the global modules
	setupScript := ""

	// Only set up llm module for mock mode
	if os.Getenv("MOCK_LLM") == "true" {
		setupScript = `
-- Set up global llm module (mock)
llm = {
	chat = llm_chat,
	complete = llm_complete,
	get_provider = llm_get_provider,
	list_providers = llm_list_providers,
	set_provider = llm_set_provider,
	stream_chat = function(prompt, callback)
		-- Mock streaming by calling callback with chunks
		callback("[Mock streaming: ")
		callback(prompt)
		callback(" - completed]")
		return nil
	end
}
`
	}

	// Set up params from command line
	paramsTable := "params = {"
	for i := 2; i < len(os.Args); i++ {
		if strings.Contains(os.Args[i], "=") {
			parts := strings.SplitN(os.Args[i], "=", 2)
			if len(parts) == 2 {
				paramsTable += fmt.Sprintf("\n\t%s = %q,", parts[0], parts[1])
			}
		}
	}
	paramsTable += "\n}"

	setupScript += paramsTable

	err = eng.LoadScript(strings.NewReader(setupScript))
	if err != nil {
		log.Fatalf("Failed to load setup script: %v", err)
	}

	err = eng.Execute(context.Background())
	if err != nil {
		log.Fatalf("Failed to execute setup script: %v", err)
	}

	// Load and execute the spell
	err = eng.LoadScriptFile(mainScript)
	if err != nil {
		log.Fatalf("Failed to load spell: %v", err)
	}

	fmt.Println("=== Spell Output ===")
	err = eng.Execute(context.Background())
	if err != nil {
		log.Fatalf("Failed to execute spell: %v", err)
	}
	fmt.Println("\n=== Spell Complete ===")
}

// registerMockLLM registers mock LLM functions for demo
func registerMockLLM(eng *lua.LuaEngine) {
	eng.RegisterFunction("llm_chat", func(prompt string) string {
		return fmt.Sprintf("[Mock LLM Response] I received your prompt: '%s'. This is a mock response for demonstration.", prompt)
	})

	eng.RegisterFunction("llm_complete", func(prompt string, maxTokens int) string {
		return fmt.Sprintf("%s... [Mock completion with max %d tokens]", prompt, maxTokens)
	})

	eng.RegisterFunction("llm_get_provider", func() string {
		return "mock"
	})

	eng.RegisterFunction("llm_list_providers", func() []string {
		return []string{"mock"}
	})

	eng.RegisterFunction("llm_set_provider", func(name string) string {
		return fmt.Sprintf("[Mock] Would switch to provider: %s", name)
	})
}
