# Async LLM Example

This example demonstrates how to use promises for asynchronous LLM operations in go-llmspell.

## Features

1. **Parallel LLM Calls**: Shows how to make multiple LLM calls concurrently using promises
2. **Promise Chaining**: Demonstrates sequential operations with `.next()` 
3. **Promise Racing**: Shows how to race multiple providers to get the fastest response
4. **Error Handling**: Demonstrates error recovery with `.catch()`

## Usage

```bash
# Run with default prompts
llmspell run async-llm

# Run with custom prompts
llmspell run async-llm --prompts='["What is AI?", "Explain machine learning", "Define neural networks"]'
```

## How It Works

### Parallel Processing
The example creates a promise for each LLM prompt and then uses `promise.all()` to wait for all of them to complete:

```lua
local promises = {}
for i, prompt in ipairs(prompts) do
    promises[i] = llm_async(prompt)
end

local results = promise.all(promises):await()
```

### Promise Chaining
Shows how to chain operations using `.next()` (renamed from `then` to avoid Lua keyword conflict):

```lua
promise.new(function(resolve, reject)
    resolve("Tell me a joke")
end):next(function(prompt)
    return llm.complete(prompt)
end):next(function(joke)
    return llm.complete("Explain: " .. joke)
end)
```

### Racing Providers
Demonstrates using `promise.race()` to get the fastest response from multiple providers:

```lua
local race_promises = {}
for i, provider in ipairs(providers) do
    race_promises[i] = promise_for_provider(provider)
end
local winner = promise.race(race_promises):await()
```

### Error Handling
Shows how to handle errors gracefully with `.catch()`:

```lua
promise.new(function(resolve, reject)
    -- operation that might fail
end):catch(function(err)
    -- handle error and return recovery value
    return "default response"
end)
```

## Note on Async Behavior

While the promises provide a familiar async-like API, the current implementation executes synchronously due to Lua's single-threaded nature. This still provides benefits:

- Clean error handling
- Composable operations  
- Familiar promise patterns
- Future-ready for true async when/if implemented

The LLM operations themselves may have internal concurrency depending on the provider implementation.

## Important Notes

1. The promise implementation uses `.next()` instead of `.then()` because `then` is a reserved keyword in Lua
2. When using `promise.all()`, the results are returned as a simple array of values
3. Table values passed through promises are converted to/from Go, which may change their structure slightly