// ABOUTME: Main entry point for the llmspell CLI
// ABOUTME: Provides commands to run spells and manage the spell environment

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/bridges"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/stdlib"
	"github.com/lexlapax/go-llmspell/pkg/tools"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't exit on error
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: spell path required")
			fmt.Println("Usage: llmspell run <spell-path> [param=value ...]")
			os.Exit(1)
		}
		runSpell(os.Args[2], os.Args[3:])
	case "help", "-h", "--help":
		printUsage()
	case "version", "-v", "--version":
		fmt.Println("llmspell v0.1.0")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("llmspell - Cast scripting spells to animate LLM golems")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  llmspell run <spell-path> [param=value ...]  Run a spell")
	fmt.Println("  llmspell help                                 Show this help")
	fmt.Println("  llmspell version                              Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  llmspell run examples/spells/hello-llm")
	fmt.Println("  llmspell run examples/spells/tool-example")
	fmt.Println("  llmspell run my-spell.lua topic=\"AI safety\"")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  OPENAI_API_KEY      OpenAI API key")
	fmt.Println("  ANTHROPIC_API_KEY   Anthropic API key")
	fmt.Println("  GEMINI_API_KEY      Google Gemini API key")
	fmt.Println("  MOCK_LLM            Set to 'true' to use mock LLM for testing")
}

func runSpell(spellPath string, args []string) {
	// Determine if it's a directory or file
	info, err := os.Stat(spellPath)
	if err != nil {
		log.Fatalf("Cannot access spell: %v", err)
	}

	var mainScript string
	var spellName string

	if info.IsDir() {
		// Look for main.lua in the directory
		mainScript = filepath.Join(spellPath, "main.lua")
		spellName = filepath.Base(spellPath)
	} else {
		// Single file spell
		mainScript = spellPath
		spellName = strings.TrimSuffix(filepath.Base(spellPath), filepath.Ext(spellPath))
	}

	// Check if the script exists
	if _, err := os.Stat(mainScript); err != nil {
		log.Fatalf("Cannot find spell script: %v", err)
	}

	fmt.Printf("🧙 Running spell: %s\n\n", spellName)

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

	// Initialize bridges
	initializeBridges(eng, spellName)

	// Set up parameters
	setupParams(eng, args)

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

func initializeBridges(eng *lua.LuaEngine, spellName string) {
	// Register standard library
	stdlibConfig := &stdlib.Config{
		SpellName: spellName,
		LogLevel:  slog.LevelInfo,
		Storage:   stdlib.DefaultStorageConfig(),
		HTTP:      stdlib.DefaultHTTPConfig(),
	}

	luaState := eng.GetLuaState()
	if err := stdlib.RegisterAll(luaState, stdlibConfig); err != nil {
		log.Fatalf("Failed to register stdlib: %v", err)
	}

	// Register tools bridge with built-in tools
	toolRegistry := tools.NewRegistry()
	toolBridge, err := bridge.NewToolBridgeWithBuiltins(toolRegistry, tools.DefaultBuiltinToolConfig())
	if err != nil {
		log.Printf("Warning: Failed to create tool bridge with builtins: %v", err)
		// Fallback to bridge without builtins
		toolBridge = bridge.NewToolBridge(toolRegistry)
	}
	if err := bridges.RegisterToolsModule(luaState, toolBridge); err != nil {
		log.Printf("Warning: Failed to register tools module: %v", err)
	}

	// Register agents bridge
	agentBridge, err := bridge.NewAgentBridge(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to create agent bridge: %v", err)
	} else {
		if err := bridges.RegisterAgentsModule(luaState, agentBridge); err != nil {
			log.Printf("Warning: Failed to register agents module: %v", err)
		}
	}

	// Register LLM bridge
	if os.Getenv("MOCK_LLM") == "true" {
		fmt.Println("🎭 Using mock LLM for demonstration")
		registerMockLLM(eng)
	} else {
		llmBridge, err := bridge.NewLLMBridge()
		if err != nil {
			fmt.Printf("⚠️  LLM Bridge not available: %v\n", err)
			fmt.Println("   Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY to enable LLM features.")
			fmt.Println("   Running with mock LLM functions...")
			registerMockLLM(eng)
		} else {
			fmt.Printf("✅ LLM Bridge initialized with provider: %s\n\n", llmBridge.GetCurrentProvider())
			adapter := bridges.NewLLMBridgeAdapter(llmBridge)
			luaBridge := bridges.NewLLMBridge(adapter)
			if err := luaBridge.Register(luaState); err != nil {
				log.Fatalf("Failed to register LLM bridge: %v", err)
			}
		}
	}
}

func setupParams(eng *lua.LuaEngine, args []string) {
	// Parse parameters
	params := make(map[string]string)
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				params[parts[0]] = parts[1]
			}
		}
	}

	// Create params table
	paramsScript := "params = {"
	for k, v := range params {
		paramsScript += fmt.Sprintf("\n\t%s = %q,", k, v)
	}
	paramsScript += "\n}"

	// Execute setup script
	err := eng.LoadScript(strings.NewReader(paramsScript))
	if err != nil {
		log.Printf("Warning: Failed to set up parameters: %v", err)
	}

	err = eng.Execute(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to execute parameter setup: %v", err)
	}
}

func registerMockLLM(eng *lua.LuaEngine) {
	// Create mock LLM module
	mockScript := `
llm = {
	chat = function(prompt)
		return "[Mock LLM Response] I received your prompt: '" .. prompt .. "'. This is a mock response for demonstration."
	end,
	complete = function(prompt, maxTokens)
		return prompt .. "... [Mock completion with max " .. tostring(maxTokens) .. " tokens]"
	end,
	get_provider = function()
		return "mock"
	end,
	list_providers = function()
		return {"mock"}
	end,
	set_provider = function(name)
		return "[Mock] Would switch to provider: " .. name
	end,
	stream_chat = function(prompt, callback)
		-- Mock streaming by calling callback with chunks
		callback("[Mock streaming: ")
		callback(prompt)
		callback(" - completed]")
		return nil
	end
}
`
	err := eng.LoadScript(strings.NewReader(mockScript))
	if err != nil {
		log.Printf("Warning: Failed to register mock LLM: %v", err)
		return
	}

	err = eng.Execute(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to execute mock LLM setup: %v", err)
	}
}
