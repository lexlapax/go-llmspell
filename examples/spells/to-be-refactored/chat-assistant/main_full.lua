-- ABOUTME: Interactive chat assistant spell with conversation history
-- ABOUTME: Maintains context across multiple interactions

-- Get parameters with defaults
local system_prompt = params.system_prompt or "You are a helpful assistant."
local max_history = params.max_history or 10

-- Initialize conversation history
local history = {}

-- Load previous history if available
if storage and storage.exists("chat_history.json") then
    local saved_history, err = storage.read("chat_history.json")
    if not err and saved_history then
        history = json.decode(saved_history) or {}
    end
end

-- Add system message if history is empty
if #history == 0 then
    table.insert(history, {
        role = "system",
        content = system_prompt
    })
end

-- Function to trim history to max size
local function trim_history()
    -- Keep system message + last max_history exchanges
    if #history > (max_history * 2 + 1) then
        local new_history = {history[1]} -- Keep system message
        local start_idx = #history - (max_history * 2) + 1
        for i = start_idx, #history do
            table.insert(new_history, history[i])
        end
        history = new_history
    end
end

-- Function to save history
local function save_history()
    if storage then
        local json_data = json.encode(history)
        storage.write("chat_history.json", json_data)
    end
end

-- Main chat loop
print("Chat Assistant Started")
print("System: " .. system_prompt)
print("Type 'exit' to quit, 'clear' to reset history\n")

while true do
    -- Get user input
    io.write("You: ")
    io.flush()
    local input = io.read()
    
    -- Check for commands
    if input == "exit" then
        print("Goodbye!")
        save_history()
        break
    elseif input == "clear" then
        history = {{role = "system", content = system_prompt}}
        save_history()
        print("History cleared.\n")
        goto continue
    elseif input == "" then
        goto continue
    end
    
    -- Add user message to history
    table.insert(history, {
        role = "user",
        content = input
    })
    
    -- Stream the response
    io.write("\nAssistant: ")
    io.flush()
    
    local response_text = ""
    local function stream_callback(chunk)
        io.write(chunk)
        io.flush()
        response_text = response_text .. chunk
        return nil
    end
    
    -- Send chat with history
    local err = llm.stream_chat_with_history(history, stream_callback)
    
    if err then
        print("\nError: " .. err)
    else
        -- Add assistant response to history
        table.insert(history, {
            role = "assistant",
            content = response_text
        })
        
        -- Trim and save history
        trim_history()
        save_history()
        
        print("\n") -- Add spacing
    end
    
    ::continue::
end