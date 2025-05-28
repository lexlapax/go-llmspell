# Lua Standard Library Reference

This document describes the standard library modules available to Lua spells in go-llmspell.

## Security Considerations

For security reasons, the following standard Lua libraries are **disabled**:
- `io` - File I/O operations (use `storage` module instead)
- `os` - Operating system interface (no direct OS access)
- `debug` - Debug library (security risk)
- `dofile`, `loadfile`, `load`, `loadstring` - Dynamic code loading (security risk)
- `require` - Module loading (controlled environment)

## Available Modules

### JSON Module

The `json` module provides JSON encoding and decoding functionality.

```lua
-- Encode a Lua table to JSON
local data = {name = "Alice", age = 30, tags = {"developer", "golang"}}
local json_string = json.encode(data)
print(json_string) -- {"age":30,"name":"Alice","tags":["developer","golang"]}

-- Decode JSON to Lua table
local decoded = json.decode(json_string)
print(decoded.name) -- Alice
```

**Functions:**
- `json.encode(value)` - Converts Lua value to JSON string
- `json.decode(string)` - Parses JSON string to Lua value

### Storage Module

The `storage` module provides sandboxed file storage operations.

```lua
-- Key-value storage
storage.set("user:123", "Alice")
local name = storage.get("user:123") -- "Alice"

-- File operations (sandboxed to storage directory)
storage.write("data.txt", "Hello, World!")
local content = storage.read("data.txt") -- "Hello, World!"

-- Check existence
if storage.exists("data.txt") then
    print("File exists")
end
```

**Functions:**
- `storage.get(key)` - Get value by key
- `storage.set(key, value)` - Set key-value pair
- `storage.exists(path)` - Check if file exists
- `storage.read(path)` - Read file contents
- `storage.write(path, content)` - Write file contents

**Security:** All paths are sandboxed to a storage directory. Path traversal attempts are blocked.

### HTTP Module

The `http` module provides HTTP client functionality with security restrictions.

```lua
-- Simple GET request
local response, err = http.get("https://api.example.com/data")
if err then
    log.error("HTTP error", {error = err})
else
    print(response)
end

-- POST with data
local data = json.encode({message = "Hello"})
local response, err = http.post("https://api.example.com/messages", data, {
    ["Content-Type"] = "application/json"
})

-- Full request control
local response, err = http.request({
    url = "https://api.example.com/data",
    method = "PUT",
    headers = {["Authorization"] = "Bearer token"},
    body = "data",
    timeout = 10 -- seconds
})
```

**Functions:**
- `http.get(url, headers)` - Perform GET request
- `http.post(url, body, headers)` - Perform POST request
- `http.request(options)` - Full request control

**Options for `http.request`:**
- `url` (required) - Target URL
- `method` - HTTP method (default: "GET")
- `headers` - Table of headers
- `body` - Request body
- `timeout` - Timeout in seconds

**Security:** Configurable domain allowlisting, default timeout of 30 seconds.

### Log Module

The `log` module provides structured logging using slog.

```lua
-- Basic logging
log.debug("Debug message")
log.info("User logged in", {user_id = 123})
log.warn("High memory usage", {percent = 85})
log.error("Failed to connect", {error = "timeout"})

-- Log with multiple fields
log.info("Processing complete", {
    duration = 1.23,
    records = 1000,
    status = "success"
})
```

**Functions:**
- `log.debug(message, fields)` - Debug level log
- `log.info(message, fields)` - Info level log
- `log.warn(message, fields)` - Warning level log
- `log.error(message, fields)` - Error level log

**Features:**
- Structured logging with key-value pairs
- Automatic spell name inclusion
- Output to stderr
- Configurable log levels

### Promise Module

The `promise` module provides promise-like patterns for async operations.

**Note:** Due to Lua's single-threaded nature, promises execute synchronously but provide a clean API for handling async patterns.

```lua
-- Create a promise
local p = promise.new(function(resolve, reject)
    local result, err = llm.complete("Hello", 50)
    if err then
        reject(err)
    else
        resolve(result)
    end
end)

-- Handle results
local value, err = p:await()

-- Promise chaining (using 'next' instead of 'then' due to Lua keyword)
promise.resolve(5)
    :next(function(x) return x * 2 end)
    :next(function(x) return x + 1 end)
    :await() -- Returns 11

-- Error handling
promise.reject("error")
    :catch(function(err) 
        return "recovered from: " .. err 
    end)
    :await() -- Returns "recovered from: error"

-- Wait for multiple promises
local all_results = promise.all({p1, p2, p3}):await()

-- Race promises
local first_result = promise.race({p1, p2, p3}):await()
```

**Functions:**
- `promise.new(executor)` - Create new promise
- `promise.resolve(value)` - Create resolved promise
- `promise.reject(reason)` - Create rejected promise
- `promise.all(promises)` - Wait for all promises
- `promise.race(promises)` - Get first settled promise

**Methods:**
- `p:next(onResolve, onReject)` - Chain handlers (note: not `then`)
- `p:catch(onReject)` - Handle rejection
- `p:await(timeout)` - Block until settled

## LLM Module

The `llm` module is provided by the LLM bridge and offers these functions:

```lua
-- Basic chat
local response, err = llm.chat("What is AI?")

-- Completion with max tokens
local response, err = llm.complete("The future of AI is", 100)

-- Streaming response
llm.stream_chat("Tell me a story", function(chunk)
    io.write(chunk)
    io.flush()
    return nil -- Return error to stop streaming
end)

-- Provider management
local providers = llm.list_providers() -- {"openai", "anthropic", "gemini"}
local current = llm.get_provider() -- "openai"
llm.set_provider("anthropic") -- Switch provider

-- Model listing
local models = llm.list_models() -- All available models
```

## Example Usage

Here's a complete example using multiple modules:

```lua
-- Async LLM calls with error handling
local prompts = {"What is AI?", "What is ML?", "What is DL?"}
local promises = {}

for i, prompt in ipairs(prompts) do
    promises[i] = promise.new(function(resolve, reject)
        log.info("Sending prompt", {index = i, prompt = prompt})
        local response, err = llm.complete(prompt, 100)
        if err then
            reject(err)
        else
            resolve(response)
        end
    end)
end

-- Wait for all responses
local results, err = promise.all(promises):await(30)
if err then
    log.error("Failed to get all responses", {error = err})
else
    -- Save results
    local data = json.encode({
        timestamp = os.time(),
        prompts = prompts,
        responses = results
    })
    storage.write("ai_responses.json", data)
    log.info("Saved responses", {count = #results})
end
```

## Best Practices

1. **Error Handling**: Always check for errors from I/O operations
2. **Logging**: Use structured logging with meaningful fields
3. **Storage**: Use the storage module for all file operations
4. **Promises**: Use promises for clean async code patterns
5. **Security**: Never try to bypass sandbox restrictions

## Limitations

- No direct file I/O outside storage directory
- No OS command execution
- No dynamic code loading
- HTTP requests may be restricted by domain allowlist
- Promise execution is synchronous (no true parallelism)