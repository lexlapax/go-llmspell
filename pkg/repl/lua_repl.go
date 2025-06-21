// ABOUTME: Lua-specific REPL implementation with engine integration and Lua syntax support.
// ABOUTME: Provides Lua script execution, state preservation, and Lua-specific completion.

package repl

import (
	"context"
	"fmt"
	"strings"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/errors"
	lua "github.com/yuin/gopher-lua"
)

// LuaREPL provides a Lua-specific REPL implementation
type LuaREPL struct {
	*BaseREPL
	engine   engine.ScriptEngine // Used for capabilities and validation
	luaState *lua.LState         // Persistent state for REPL evaluations
}

// NewLuaREPL creates a new Lua REPL instance
func NewLuaREPL(config REPLConfig) (*LuaREPL, error) {
	if config.Engine != "lua" {
		return nil, errors.Newf(errors.CategoryConfig, "engine must be 'lua', got '%s'", config.Engine)
	}

	// Create base REPL
	baseREPL, err := NewBaseREPL(config)
	if err != nil {
		return nil, errors.Wrap(err, errors.CategoryConfig, "failed to create base REPL")
	}

	// Create Lua engine using factory pattern
	factory := gopherlua.NewLuaEngineFactory()
	engineConfig := factory.GetDefaultConfig()

	// Override some settings for REPL use
	engineConfig.SandboxMode = false // More permissive for interactive use
	engineConfig.DebugMode = false
	engineConfig.MetricsMode = true

	scriptEngine, err := factory.Create(engineConfig)
	if err != nil {
		_ = baseREPL.Close()
		return nil, errors.Wrap(err, errors.CategoryEngine, "failed to create Lua engine")
	}

	// Create persistent Lua state for REPL
	luaState := lua.NewState()

	// Load standard libraries for REPL use
	luaState.OpenLibs()

	luaREPL := &LuaREPL{
		BaseREPL: baseREPL,
		engine:   scriptEngine,
		luaState: luaState,
	}

	return luaREPL, nil
}

// Evaluate executes Lua code and returns the result
func (l *LuaREPL) Evaluate(ctx context.Context, input string) (string, error) {
	input = strings.TrimSpace(input)

	// Check if it's a REPL command first
	if isCommand, command := parseREPLCommand(input); isCommand {
		return l.executeCommand(ctx, input, command)
	}

	// Execute Lua code using persistent state
	if err := l.luaState.DoString(input); err != nil {
		return "", errors.Wrap(err, errors.CategoryScript, "Lua execution failed")
	}

	// Get and format result from stack
	if l.luaState.GetTop() > 0 {
		result := l.luaState.Get(-1)
		l.luaState.Pop(1) // Remove result from stack

		if result != lua.LNil {
			return result.String(), nil
		}
	}

	return "", nil
}

// Complete provides Lua-specific auto-completion
func (l *LuaREPL) Complete(input string) []string {
	// Get base completions first
	completions := l.BaseREPL.Complete(input)

	// Add Lua-specific completions
	if !strings.HasPrefix(input, ".") {
		luaCompletions := l.getLuaCompletions(input)
		completions = append(completions, luaCompletions...)
	}

	return completions
}

// Close shuts down the Lua REPL and cleans up resources
func (l *LuaREPL) Close() error {
	// Close persistent Lua state
	if l.luaState != nil {
		l.luaState.Close()
	}

	// Shutdown Lua engine
	if l.engine != nil {
		_ = l.engine.Shutdown()
	}

	// Close base REPL
	return l.BaseREPL.Close()
}

// executeCommand executes REPL commands with Lua-specific extensions
func (l *LuaREPL) executeCommand(ctx context.Context, input, command string) (string, error) {
	switch command {
	case "load":
		return l.handleLoadCommand(ctx, input)
	case "engines":
		return l.handleEnginesCommand(ctx)
	default:
		// Use base implementation for other commands
		return l.BaseREPL.Evaluate(ctx, input)
	}
}

// handleLoadCommand loads and executes a Lua file
func (l *LuaREPL) handleLoadCommand(_ context.Context, input string) (string, error) {
	args := strings.Fields(input)
	if len(args) < 2 {
		return "", errors.New(errors.CategoryValidation, "usage: .load <filename>")
	}

	filename := args[1]

	// Execute the file using persistent state
	if err := l.luaState.DoFile(filename); err != nil {
		return "", errors.Wrapf(err, errors.CategoryIO, "failed to load file %s", filename)
	}

	response := fmt.Sprintf("Loaded file: %s", filename)

	// Get result from stack if any
	if l.luaState.GetTop() > 0 {
		result := l.luaState.Get(-1)
		l.luaState.Pop(1)
		if result != lua.LNil {
			response += fmt.Sprintf("\nResult: %v", result.String())
		}
	}

	return response, nil
}

// handleEnginesCommand shows Lua engine information
func (l *LuaREPL) handleEnginesCommand(_ context.Context) (string, error) {
	var result strings.Builder
	result.WriteString("Current engine: lua\n")

	if l.engine != nil {
		metrics := l.engine.GetMetrics()
		result.WriteString("Engine metrics:\n")
		result.WriteString(fmt.Sprintf("  Scripts executed: %d\n", metrics.ScriptsExecuted))
		result.WriteString(fmt.Sprintf("  Errors: %d\n", metrics.ErrorCount))
	}

	return result.String(), nil
}

// getLuaCompletions returns Lua-specific completion suggestions
func (l *LuaREPL) getLuaCompletions(input string) []string {
	var completions []string

	// Lua keywords
	luaKeywords := []string{
		"and", "break", "do", "else", "elseif", "end", "false", "for",
		"function", "if", "in", "local", "nil", "not", "or", "repeat",
		"return", "then", "true", "until", "while",
	}

	// Lua built-in functions
	luaBuiltins := []string{
		"print", "type", "tostring", "tonumber", "pairs", "ipairs",
		"next", "getmetatable", "setmetatable", "rawget", "rawset",
		"error", "assert", "pcall", "xpcall", "load", "loadfile",
		"dofile", "require", "module", "package", "select",
	}

	// Lua standard libraries
	luaLibraries := []string{
		"table", "string", "math", "io", "os", "debug", "coroutine",
		"table.insert", "table.remove", "table.concat", "table.sort",
		"string.sub", "string.find", "string.match", "string.gsub",
		"string.format", "string.upper", "string.lower", "string.len",
		"math.abs", "math.ceil", "math.floor", "math.max", "math.min",
		"math.random", "math.sqrt", "math.sin", "math.cos", "math.pi",
		"io.open", "io.read", "io.write", "io.lines", "io.close",
		"os.time", "os.date", "os.clock", "os.getenv", "os.execute",
	}

	// Combine all completions
	allCompletions := append(luaKeywords, luaBuiltins...)
	allCompletions = append(allCompletions, luaLibraries...)

	// Filter by input prefix
	for _, completion := range allCompletions {
		if strings.HasPrefix(completion, input) {
			completions = append(completions, completion)
		}
	}

	return completions
}
