# LLMSpell REPL Guide

The LLMSpell REPL (Read-Eval-Print Loop) provides an interactive environment for developing and testing LLM scripts. It supports multiple scripting languages, syntax highlighting, command history, and debugging capabilities.

## Getting Started

### Basic Usage

Start the REPL with default settings (Lua engine):
```bash
llmspell repl
```

Start with a specific engine:
```bash
llmspell repl --engine javascript
llmspell repl --engine tengo
```

### REPL Interface

When you start the REPL, you'll see:
```
LLMSpell REPL v1.0.0 (Lua 5.1)
Type .help for commands, .exit to quit

lua> 
```

The prompt shows the current engine. You can type expressions or statements:
```lua
lua> print("Hello, LLMSpell!")
Hello, LLMSpell!

lua> x = 42
lua> print(x * 2)
84
```

## REPL Commands

All REPL commands start with a dot (`.`). Type `.help` to see available commands:

### Core Commands

- `.help` - Display help information
- `.exit` / `.quit` - Exit the REPL
- `.clear` - Clear the current context (reset variables)
- `.reset` - Full reset (clear + reload engine)

### Session Management

- `.save <filename>` - Save current session to file
- `.load <filename>` - Load and execute a session file
- `.history` - Show command history
- `.history clear` - Clear command history

### Engine Control

- `.engines` - List available script engines
- `.engine <name>` - Switch to a different engine
- `.info` - Show current engine information

### Display Control

- `.mode <mode>` - Change input mode (normal/multiline/paste)
- `.prompt <string>` - Change the prompt
- `.colors on|off` - Toggle syntax highlighting
- `.width <n>` - Set display width

### Development Tools

- `.type <expr>` - Show type of expression
- `.time <expr>` - Time expression execution
- `.mem` - Show memory usage
- `.gc` - Run garbage collector

## Multi-line Input

The REPL automatically detects when you're entering multi-line code:

```lua
lua> function factorial(n)
...   if n <= 1 then
...     return 1
...   else
...     return n * factorial(n - 1)
...   end
... end
lua> print(factorial(5))
120
```

The `...` prompt indicates continuation lines. Press Enter on an empty line to execute.

### Paste Mode

For pasting large code blocks:
```
lua> .mode paste
-- Entering paste mode. Press Ctrl+D when done --
function complex_function()
  -- Your pasted code here
  local result = {}
  for i = 1, 10 do
    table.insert(result, i * i)
  end
  return result
end
^D
-- Paste mode ended --
lua> 
```

## Working with LLM Bridges

The REPL provides access to all LLMSpell bridges:

### Basic LLM Interaction

```lua
lua> llm = require("llm")
lua> response = llm.complete("Explain quantum computing in one sentence")
lua> print(response)
Quantum computing uses quantum mechanical phenomena like superposition and 
entanglement to process information in ways classical computers cannot.
```

### Using Tools

```lua
lua> tools = require("tools")
lua> calc = tools.create("calculator")
lua> result = calc:execute("sqrt(144)")
lua> print(result)
12
```

### State Management

```lua
lua> state = require("state")
lua> state.set("user_name", "Alice")
lua> state.set("conversation_id", "12345")
lua> print(state.get("user_name"))
Alice
```

## Syntax Highlighting

The REPL includes syntax highlighting for better code readability:

- **Keywords** - Blue (if, then, function, etc.)
- **Strings** - Green ("text", 'text')
- **Numbers** - Yellow (42, 3.14)
- **Comments** - Gray (-- comment)
- **Functions** - Cyan (print, require)
- **Operators** - White (+, -, *, /)

Toggle highlighting:
```
lua> .colors off
Syntax highlighting disabled
lua> .colors on
Syntax highlighting enabled
```

## Auto-completion

The REPL supports tab completion for:

- Built-in functions
- Keywords
- Variables in scope
- Module names
- Method names

Examples:
```
lua> pri<TAB>
print

lua> string.<TAB>
string.byte    string.char    string.find    string.format
string.gmatch  string.gsub    string.len     string.lower
...

lua> llm.com<TAB>
llm.complete   llm.complete_async
```

## Command History

The REPL saves command history between sessions:

- **Up/Down arrows** - Navigate history
- **Ctrl+R** - Reverse search history
- **Ctrl+P/Ctrl+N** - Previous/next command
- `.history` - Show full history
- `.history search <pattern>` - Search history

History file location: `~/.llmspell_history`

## Error Handling

The REPL provides helpful error messages:

```lua
lua> print(undefined_variable)
Error: attempt to call a nil value (global 'undefined_variable')
  at line 1

lua> 1 / 0
Error: division by zero
  at line 1

lua> function bad() error("Something went wrong") end
lua> bad()
Error: Something went wrong
  stack traceback:
    [string "bad"]:1: in function 'bad'
    [string "<repl>"]:1: in main chunk
```

## Working with Files

Load and execute files:

```lua
lua> .load utils.lua
Loading utils.lua... done

lua> dofile("script.lua")
Executing script.lua...
```

Save your session:
```lua
lua> -- Define some functions and variables
lua> x = 42
lua> function greet(name)
...   return "Hello, " .. name
... end
lua> .save session.lua
Session saved to session.lua
```

## REPL Configuration

### Via Configuration File

```yaml
# ~/.llmspell/config.yaml
repl:
  prompt: ">> "
  continuation_prompt: ".. "
  save_history: true
  history_file: ~/.llmspell_history
  history_size: 10000
  syntax_highlight: true
  auto_complete: true
  colors:
    keyword: blue
    string: green
    number: yellow
    comment: gray
    function: cyan
```

### Via Command Line

```bash
# Custom prompt
llmspell repl --prompt "myrepl> "

# Disable history
llmspell repl --no-history

# Disable syntax highlighting
llmspell repl --no-highlight

# Custom history file
llmspell repl --history-file ~/myhistory
```

### Via Environment Variables

```bash
export LLMSPELL_REPL_PROMPT=">> "
export LLMSPELL_REPL_SAVE_HISTORY=false
export LLMSPELL_REPL_SYNTAX_HIGHLIGHT=true
```

## Advanced Features

### Async Operations

```lua
lua> async = require("async")
lua> future = llm.complete_async("Tell me a joke")
lua> -- Do other work while waiting
lua> result = future:wait()
lua> print(result)
Why don't scientists trust atoms? Because they make up everything!
```

### Debugging in REPL

```lua
lua> .debug on
Debug mode enabled

lua> function buggy(x)
...   local y = x * 2
...   print("y =", y)  -- This will show in debug
...   return y + undefined
... end
[DEBUG] Function 'buggy' defined at line 1

lua> buggy(5)
[DEBUG] Entering function 'buggy' with args: 5
y = 10
[DEBUG] Error in function 'buggy': attempt to perform arithmetic on nil
Error: attempt to perform arithmetic on a nil value (global 'undefined')
```

### Performance Profiling

```lua
lua> .profile on
Profiling enabled

lua> for i = 1, 1000000 do
...   local x = math.sqrt(i)
... end

Profile results:
  Total time: 0.123s
  Lines executed: 1000001
  Memory allocated: 15.2MB
```

## Engine-Specific Features

### Lua Engine

- Full Lua 5.1 compatibility
- Coroutine support
- Debug library access
- Custom metatables

```lua
lua> co = coroutine.create(function()
...   for i = 1, 3 do
...     print("Step", i)
...     coroutine.yield(i)
...   end
... end)
lua> print(coroutine.resume(co))
Step 1
true    1
```

### JavaScript Engine (when available)

- ES6+ syntax support
- Promise handling
- JSON manipulation

```javascript
js> const arr = [1, 2, 3, 4, 5]
js> const doubled = arr.map(x => x * 2)
js> console.log(doubled)
[2, 4, 6, 8, 10]
```

### Tengo Engine (when available)

- Simple, fast scripting
- Built-in immutability
- Go-like syntax

```tengo
tengo> arr := [1, 2, 3, 4, 5]
tengo> doubled := arr.map(func(x) { return x * 2 })
tengo> fmt.println(doubled)
[2, 4, 6, 8, 10]
```

## Tips and Tricks

### Quick Testing

Test LLM prompts quickly:
```lua
lua> function test_prompt(p) return llm.complete(p, {max_tokens=50}) end
lua> test_prompt("Write a haiku about coding")
```

### Create Shortcuts

```lua
lua> c = llm.complete  -- Shortcut for complete
lua> c("What is 2+2?")
The answer is 4.
```

### Persistent Utilities

Create a `.llmspell_init.lua` file in your home directory:
```lua
-- ~/.llmspell_init.lua
function pp(t)
  -- Pretty print function
  for k,v in pairs(t) do
    print(k, "=>", v)
  end
end

-- Auto-load common modules
llm = require("llm")
tools = require("tools")
```

### REPL as Calculator

```lua
lua> = 2 + 2
4
lua> = math.sqrt(16)
4
lua> = 10 * 20 / 5
40
```

### Explore APIs

```lua
lua> for k,v in pairs(llm) do print(k, type(v)) end
complete         function
complete_async   function
stream          function
models          table
...
```

## Troubleshooting

### Common Issues

**Issue**: REPL won't start
```bash
# Check if engine is available
llmspell engines list

# Try with explicit engine
llmspell repl --engine lua --debug
```

**Issue**: History not saving
```bash
# Check history file permissions
ls -la ~/.llmspell_history

# Use custom history file
llmspell repl --history-file /tmp/llmspell_history
```

**Issue**: Syntax highlighting not working
```bash
# Check terminal support
echo $TERM

# Disable if terminal doesn't support colors
llmspell repl --no-highlight
```

### Debug Mode

Enable debug output:
```bash
llmspell repl --debug
```

In REPL:
```
lua> .debug on
Debug mode enabled
lua> -- Now all operations show debug info
```

## Integration with External Editors

### Vim Integration

```vim
" ~/.vimrc
" Send selection to LLMSpell REPL
vnoremap <leader>r :w !llmspell repl --engine lua<CR>
```

### VS Code Integration

Use the "Run Selection/Line in Terminal" feature with:
```json
{
  "key": "ctrl+enter",
  "command": "workbench.action.terminal.runSelectedText"
}
```

## Best Practices

1. **Use `.save` regularly** to preserve work
2. **Create init files** for common utilities
3. **Use `.clear` between experiments** to avoid variable pollution
4. **Enable debug mode** when troubleshooting
5. **Use paste mode** for large code blocks
6. **Leverage tab completion** to explore APIs
7. **Keep history size reasonable** to avoid slowdowns

## See Also

- [CLI Usage Guide](cli-usage.md)
- [Configuration Guide](configuration.md)
- [Lua Engine Documentation](../engines/lua.md)
- [API Reference](../api/README.md)