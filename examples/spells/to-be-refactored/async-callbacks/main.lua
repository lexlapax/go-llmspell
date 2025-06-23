-- ABOUTME: Demonstrates real async callbacks for parallel LLM execution
-- ABOUTME: Shows how to use llm.chat_async and llm.complete_async

local mode = params.mode or "simple"

print("Async Callbacks Demonstration")
print("=============================")
print("Mode: " .. mode)
print("")

if mode == "simple" then
    print("=== Simple Async Example ===")
    print("Starting three async LLM calls...\n")

    local results = {}
    local completed = 0

    -- Start multiple async operations
    local id1 = llm.chat_async("What is 2+2?", function(result)
        results[1] = result
        completed = completed + 1
        print("Got result 1: " .. result)
    end, function(err)
        results[1] = "Error: " .. err
        completed = completed + 1
        print("Error 1: " .. err)
    end)

    local id2 = llm.chat_async("What is the capital of France?", function(result)
        results[2] = result
        completed = completed + 1
        print("Got result 2: " .. result)
    end, function(err)
        results[2] = "Error: " .. err
        completed = completed + 1
        print("Error 2: " .. err)
    end)

    local id3 = llm.complete_async("The meaning of life is", 50, function(result)
        results[3] = result
        completed = completed + 1
        print("Got result 3: " .. result)
    end, function(err)
        results[3] = "Error: " .. err
        completed = completed + 1
        print("Error 3: " .. err)
    end)

    print("All calls started. Callback IDs: " .. id1 .. ", " .. id2 .. ", " .. id3)
    print("Processing callbacks...\n")

    -- Process callbacks
    local iterations = 0
    while completed < 3 and iterations < 100 do
        local processed = async.process_callbacks()
        if processed > 0 then
            print("Processed " .. processed .. " callbacks")
        end
        iterations = iterations + 1
        -- Delay
        for i = 1, 1000000 do
        end
    end

    print("\nAll callbacks processed!")
    print("Results:")
    for i = 1, 3 do
        print("  [" .. i .. "] " .. (results[i] or "No result"))
    end
elseif mode == "parallel" then
    print("=== Parallel Execution Demonstration ===")
    print("Running multiple LLM queries in parallel...\n")

    local results = {}
    local completed = 0
    local total = 5

    local prompts = {
        "Count to 5",
        "Name 3 colors",
        "What is AI?",
        "Capital of Japan",
        "2 + 2 equals?",
    }

    -- Start all operations in parallel
    print("Starting " .. total .. " parallel operations...")

    for i, prompt in ipairs(prompts) do
        llm.chat_async(prompt, function(result)
            completed = completed + 1
            results[i] = result
            print("  [" .. completed .. "/" .. total .. "] Completed: " .. prompt)
        end, function(err)
            completed = completed + 1
            results[i] = "ERROR: " .. err
            print("  [" .. completed .. "/" .. total .. "] Failed: " .. prompt)
        end)
    end

    print("\nWaiting for all operations to complete...")

    -- Event loop
    local iterations = 0
    while completed < total and iterations < 150 do
        local processed = async.process_callbacks()
        iterations = iterations + 1
        -- Delay
        for i = 1, 1000000 do
        end
    end

    print("\nAll operations completed in " .. iterations .. " iterations!")
    print("\nResults:")
    for i, prompt in ipairs(prompts) do
        print(i .. ". " .. prompt)
        print("   → " .. (results[i] or "No response"))
    end
elseif mode == "promise" then
    print("=== Promise Integration ===")
    print("Using promises with async callbacks\n")

    -- Helper to create promise from async call
    function llm_promise(prompt)
        return promise.new(function(resolve, reject)
            llm.chat_async(prompt, resolve, reject)
        end)
    end

    print("Creating promise-wrapped async calls...")

    local promises = {
        llm_promise("First question: What is 1+1?"),
        llm_promise("Second question: Name a planet"),
        llm_promise("Third question: What color is the sky?"),
    }

    print("Processing promises with callbacks...")

    -- Collect results
    local results = {}
    local done = 0

    for i, p in ipairs(promises) do
        p:next(function(result)
            results[i] = result
            done = done + 1
            print("Promise " .. i .. " resolved")
        end):catch(function(err)
            results[i] = "Error: " .. err
            done = done + 1
            print("Promise " .. i .. " rejected")
        end)
    end

    -- Wait for all
    local iterations = 0
    while done < #promises and iterations < 100 do
        async.process_callbacks()
        iterations = iterations + 1
        for i = 1, 1000000 do
        end
    end

    print("\nAll promises resolved!")
    for i, result in ipairs(results) do
        print("Result " .. i .. ": " .. result)
    end
elseif mode == "streaming" then
    print("=== Streaming with Callbacks ===")
    print("Combining streaming and async patterns\n")

    local stream_chunks = {}
    local stream_done = false

    -- Regular async call
    print("1. Starting async chat...")
    llm.chat_async("Tell me a very short joke", function(result)
        print("   Async result: " .. result)
    end)

    -- Streaming call (synchronous but with callback for each chunk)
    print("\n2. Starting streaming chat...")
    local err = llm.stream_chat("Tell me another very short joke", function(chunk)
        table.insert(stream_chunks, chunk)
        -- io.write not available in sandbox, just collect chunks
        return nil
    end)

    if err then
        print("\n   Streaming error: " .. err)
    else
        stream_done = true
        print("\n   Streaming complete!")
        print("   Collected chunks: " .. table.concat(stream_chunks, ""))
    end

    -- Process async callbacks
    print("\n3. Processing async callbacks...")
    local iterations = 0
    while iterations < 50 do
        local processed = async.process_callbacks()
        if processed > 0 then
            break
        end
        iterations = iterations + 1
        for i = 1, 1000000 do
        end
    end

    print("\nDemo complete!")
end

print("\n" .. string.rep("=", 60))
print("KEY INSIGHTS")
print(string.rep("=", 60))
print([[
Real async callbacks enable true parallel LLM execution:

1. **Non-blocking**: llm.chat_async returns immediately
2. **Parallel execution**: Multiple LLM calls run concurrently
3. **Event loop**: async.process_callbacks() handles results
4. **Error handling**: Separate callbacks for success/failure

This is much more efficient than sequential execution!
]])

print("\n✅ Demo complete!")
