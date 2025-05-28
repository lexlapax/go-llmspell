-- ABOUTME: Demonstrates promise-based async LLM operations
-- ABOUTME: Shows how to use promises for concurrent LLM calls

-- Get prompts from parameters or use defaults
local prompts = params and params.prompts or {}
if not prompts or type(prompts) ~= "table" or #prompts == 0 then
    prompts = {
        "What is the capital of France?",
        "Explain quantum computing in one sentence",
        "What's 2+2?",
        "Why is the sky blue?"
    }
end

print("Processing " .. #prompts .. " prompts asynchronously...\n")

-- Function to create a promise for an LLM call
function llm_async(prompt)
    return promise.new(function(resolve, reject)
        local result, err = llm.complete(prompt, 100)
        if err then
            reject("Error for prompt '" .. prompt .. "': " .. err)
        else
            resolve(result)
        end
    end)
end

-- Example 1: Process all prompts concurrently
print("=== Example 1: Concurrent Processing with promise.all ===\n")

local promises = {}
for i, prompt in ipairs(prompts) do
    print("Starting request " .. i .. ": " .. prompt)
    promises[i] = llm_async(prompt)
end

print("\nWaiting for all responses...")
local all_promise = promise.all(promises)
local results, err = all_promise:await(30) -- 30 second timeout

if err then
    print("Error waiting for responses: " .. err)
    return
end

print("\nAll responses received:\n")
for i, result in ipairs(results) do
    print("Q" .. i .. ": " .. prompts[i])
    print("A" .. i .. ": " .. result)
    print()
end

-- Example 2: Promise chaining for sequential operations
print("=== Example 2: Promise Chaining ===\n")

local chain_result = promise.new(function(resolve, reject)
    resolve("Tell me a short joke")
end):next(function(prompt)
    print("Step 1: Asking for joke...")
    local result, err = llm.complete(prompt, 50)
    if err then error(err) end
    return result
end):next(function(joke)
    print("Step 2: Got joke: " .. joke)
    print("\nAsking for explanation...")
    local prompt = "Explain why this is funny: " .. joke
    local result, err = llm.complete(prompt, 100)
    if err then error(err) end
    return "JOKE: " .. joke .. "\nEXPLANATION: " .. result
end)

local final_result = chain_result:await()
print("\nFinal result:")
print(final_result)

-- Example 3: Racing multiple approaches
print("\n=== Example 3: Promise Racing ===\n")

local race_promises = {}
local approaches = {
    "Explain AI in technical terms",
    "Explain AI like I'm five",
    "Explain AI with an analogy"
}

for i, approach in ipairs(approaches) do
    race_promises[i] = promise.new(function(resolve, reject)
        print("Starting approach " .. i .. ": " .. approach)
        local result, err = llm.complete(approach, 75)
        if err then
            reject("Approach " .. i .. " failed: " .. err)
        else
            resolve("Approach: " .. approach .. "\nResponse: " .. result)
        end
    end)
end

print("\nRacing to see which approach completes first...")
local race_promise = promise.race(race_promises)
local winner, err = race_promise:await(10)

if err then
    print("Race error: " .. err)
else
    print("\nFirst to complete:")
    print(winner)
end

-- Example 4: Error handling with catch
print("\n=== Example 4: Error Handling with .catch() ===\n")

local error_promise = promise.new(function(resolve, reject)
    -- Intentionally use invalid parameters
    local result, err = llm.complete("", 0)
    if err then
        reject(err)
    else
        resolve(result)
    end
end):catch(function(err)
    print("Caught error: " .. err)
    -- Recover with a default prompt
    local result, err2 = llm.complete("Say hello", 20)
    if err2 then
        return "Failed to recover: " .. err2
    end
    return "Recovered with: " .. result
end)

local recovered = error_promise:await()
print("Final result: " .. recovered)

-- Example 5: Creating resolved/rejected promises
print("\n=== Example 5: Pre-resolved Promises ===\n")

local resolved = promise.resolve("Already done!")
local rejected = promise.reject("Already failed!")

print("Resolved promise: " .. resolved:await())

local handled = rejected:catch(function(err)
    return "Handled rejection: " .. err
end)
print("Rejected promise: " .. handled:await())

print("\n✅ Async LLM demo complete!")