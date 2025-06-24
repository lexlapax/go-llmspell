-- testspell
-- A new spell

local llm = require("llm")

-- Get parameters
local prompt = params.prompt or error("Prompt is required")
local model = params.model or "gpt-3.5-turbo"

-- Set the provider with model
llm.setProvider("openai", {
    model = model
})

-- Generate response
local response, err = llm.generate(prompt, {
    temperature = 0.7
})

if err then
    error("LLM request failed: " .. tostring(err))
end

-- Return response for further processing
return response
