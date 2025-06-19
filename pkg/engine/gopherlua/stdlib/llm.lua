-- ABOUTME: LLM Operations Library for go-llmspell Lua standard library
-- ABOUTME: Provides high-level LLM operation helpers, provider management, and model discovery utilities

local llm = {}

-- Import promise library for async operations
local promise = _G.promise or require("promise")

-- Internal state for current provider and configuration
local current_provider = nil
local default_options = {
    temperature = 0.7,
    max_tokens = 1000,
    timeout = 30,
}

-- Helper function to merge options with defaults
local function merge_options(options)
    local merged = {}
    for k, v in pairs(default_options) do
        merged[k] = v
    end
    if options then
        for k, v in pairs(options) do
            merged[k] = v
        end
    end
    return merged
end

-- Helper function to validate required parameters
local function validate_required(param, name)
    if not param or param == "" then
        error(name .. " is required")
    end
end

-- Helper function to get LLM bridge (assumes global access to bridges)
local function get_llm_bridge()
    if not _G.llm_bridge then
        error("LLM bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.llm_bridge
end

-- High-level LLM operation helpers

-- Quick prompt for simple prompting
function llm.quick_prompt(prompt, options)
    validate_required(prompt, "prompt")

    local opts = merge_options(options)
    local bridge = get_llm_bridge()

    -- Use current provider if set, otherwise use default
    if current_provider then
        return bridge:generateWithProvider(current_provider, prompt, opts)
    else
        return bridge:generate(prompt, opts)
    end
end

-- Async version of quick_prompt
function llm.quick_prompt_async(prompt, options)
    validate_required(prompt, "prompt")

    return promise.async(function()
        return llm.quick_prompt(prompt, options)
    end)()
end

-- Chat session for conversation management
function llm.chat_session(system_prompt)
    local session = {
        messages = {},
        system_prompt = system_prompt,
    }

    -- Add system message if provided
    if system_prompt then
        table.insert(session.messages, {
            role = "system",
            content = system_prompt,
        })
    end

    -- Add user message and get response
    function session:send(user_message, options)
        validate_required(user_message, "user_message")

        -- Add user message
        table.insert(self.messages, {
            role = "user",
            content = user_message,
        })

        local opts = merge_options(options)
        local bridge = get_llm_bridge()

        local response
        if current_provider then
            response = bridge:generateWithProvider(current_provider, self.messages, opts)
        else
            response = bridge:generateMessage(self.messages, opts)
        end

        -- Add assistant response to conversation
        if response and response.content then
            table.insert(self.messages, {
                role = "assistant",
                content = response.content,
            })
        end

        return response
    end

    -- Async version of send
    function session:send_async(user_message, options)
        return promise.async(function()
            return self:send(user_message, options)
        end)()
    end

    -- Get conversation history
    function session:get_history()
        return self.messages
    end

    -- Clear conversation (keeping system prompt)
    function session:clear()
        self.messages = {}
        if self.system_prompt then
            table.insert(self.messages, {
                role = "system",
                content = self.system_prompt,
            })
        end
    end

    -- Export conversation
    function session:export()
        return {
            system_prompt = self.system_prompt,
            messages = self.messages,
        }
    end

    return session
end

-- Streaming response with callback
function llm.streaming_response(prompt, callback, options)
    validate_required(prompt, "prompt")
    validate_required(callback, "callback")

    if type(callback) ~= "function" then
        error("callback must be a function")
    end

    local opts = merge_options(options)
    local bridge = get_llm_bridge()

    -- Start streaming
    local stream_id
    if current_provider then
        stream_id = bridge:streamWithProvider(current_provider, prompt, opts)
    else
        stream_id = bridge:stream(prompt, opts)
    end

    if not stream_id then
        error("Failed to start streaming")
    end

    -- Read stream in chunks and call callback
    local function read_stream()
        while true do
            local chunk = bridge:readStream(stream_id)
            if not chunk then
                break -- Stream ended
            end

            -- Call user callback with chunk
            local continue = callback(chunk)
            if continue == false then
                break -- User requested stop
            end
        end

        -- Close stream
        bridge:closeStream(stream_id)
    end

    -- Return promise that resolves when streaming completes
    return promise.spawn(read_stream)
end

-- Batch process multiple prompts
function llm.batch_process(prompts, options)
    if type(prompts) ~= "table" then
        error("prompts must be a table")
    end

    local opts = merge_options(options)
    local results = {}
    local bridge = get_llm_bridge()

    -- Process sequentially for now (could be made parallel later)
    for i, prompt in ipairs(prompts) do
        if type(prompt) == "string" then
            local response
            if current_provider then
                response = bridge:generateWithProvider(current_provider, prompt, opts)
            else
                response = bridge:generate(prompt, opts)
            end
            results[i] = response
        else
            error("All prompts must be strings, got " .. type(prompt) .. " at index " .. i)
        end
    end

    return results
end

-- Async batch processing with concurrent execution
function llm.batch_process_async(prompts, options)
    if type(prompts) ~= "table" then
        error("prompts must be a table")
    end

    local promises = {}

    -- Create promises for each prompt
    for i, prompt in ipairs(prompts) do
        promises[i] = promise.async(function()
            return llm.quick_prompt(prompt, options)
        end)()
    end

    -- Return promise that resolves when all complete
    return promise.Promise.all(promises)
end

-- Provider management utilities

-- Use a specific provider
function llm.use_provider(name, config)
    validate_required(name, "provider name")

    local bridge = get_llm_bridge()

    -- Set provider configuration if provided
    if config then
        bridge:setProvider(name, config)
    end

    current_provider = name

    -- Verify provider is working
    local success, err = pcall(function()
        return bridge:testProviderConnection(name)
    end)

    if not success then
        current_provider = nil
        error("Failed to connect to provider '" .. name .. "': " .. tostring(err))
    end

    return true
end

-- Get current provider
function llm.get_current_provider()
    return current_provider
end

-- List available providers
function llm.list_providers()
    local bridge = get_llm_bridge()
    return bridge:listProviders()
end

-- Compare providers with the same prompt
function llm.compare_providers(prompt, providers, options)
    validate_required(prompt, "prompt")

    if type(providers) ~= "table" then
        error("providers must be a table")
    end

    local opts = merge_options(options)
    local results = {}
    local bridge = get_llm_bridge()

    -- Test each provider
    for i, provider_name in ipairs(providers) do
        local start_time = os.clock()
        local success, response = pcall(function()
            return bridge:generateWithProvider(provider_name, prompt, opts)
        end)
        local end_time = os.clock()

        results[i] = {
            provider = provider_name,
            success = success,
            response = success and response or nil,
            error = success and nil or response,
            duration = end_time - start_time,
        }
    end

    return results
end

-- Setup fallback chain for reliability
function llm.setup_fallback_chain(providers)
    if type(providers) ~= "table" then
        error("providers must be a table")
    end

    local bridge = get_llm_bridge()
    return bridge:setFallbackChain(providers)
end

-- Get current fallback chain
function llm.get_fallback_chain()
    local bridge = get_llm_bridge()
    return bridge:getFallbackChain()
end

-- Generate with fallback chain
function llm.generate_with_fallback(prompt, options)
    validate_required(prompt, "prompt")

    local opts = merge_options(options)
    local bridge = get_llm_bridge()
    local fallback_chain = bridge:getFallbackChain()

    if not fallback_chain or #fallback_chain == 0 then
        error("No fallback chain configured. Use llm.setup_fallback_chain() first.")
    end

    local last_error

    -- Try each provider in the fallback chain
    for _, provider_name in ipairs(fallback_chain) do
        local success, result = pcall(function()
            return bridge:generateWithProvider(provider_name, prompt, opts)
        end)

        if success then
            return result
        else
            last_error = result
        end
    end

    -- All providers failed
    error("All providers in fallback chain failed. Last error: " .. tostring(last_error))
end

-- Model discovery helpers

-- Find model based on requirements
function llm.find_model(requirements)
    if type(requirements) ~= "table" then
        error("requirements must be a table")
    end

    local bridge = get_llm_bridge()

    -- Get all providers if none specified
    local providers_to_check = requirements.providers or bridge:listProviders()

    local suitable_models = {}

    for _, provider_name in ipairs(providers_to_check) do
        local models = bridge:listModels(provider_name)

        if models then
            for _, model in ipairs(models) do
                local model_info = bridge:getModelInfo(model.id)

                -- Check requirements
                local suitable = true

                if
                    requirements.min_context_length
                    and model_info.context_length
                    and model_info.context_length < requirements.min_context_length
                then
                    suitable = false
                end

                if requirements.supports_streaming and not model_info.supports_streaming then
                    suitable = false
                end

                if requirements.supports_tools and not model_info.supports_tools then
                    suitable = false
                end

                if
                    requirements.max_cost_per_token
                    and model_info.cost_per_token
                    and model_info.cost_per_token > requirements.max_cost_per_token
                then
                    suitable = false
                end

                if suitable then
                    table.insert(suitable_models, {
                        provider = provider_name,
                        model = model,
                        info = model_info,
                    })
                end
            end
        end
    end

    -- Sort by preference (could be enhanced with scoring)
    return suitable_models
end

-- Get detailed model information
function llm.model_info(model_id, provider)
    validate_required(model_id, "model_id")

    local bridge = get_llm_bridge()

    if provider then
        -- Get info for specific provider
        return bridge:getModelInfo(model_id, provider)
    else
        -- Get info from current or default provider
        return bridge:getModelInfo(model_id)
    end
end

-- Estimate cost for an operation
function llm.cost_estimate(operation, model, provider)
    validate_required(operation, "operation")

    local bridge = get_llm_bridge()
    local util_bridge = _G.llm_util_bridge

    if not util_bridge then
        error("LLM utilities bridge not available")
    end

    local model_id = model or "default"
    local provider_name = provider or current_provider

    if not provider_name then
        error("No provider specified and no current provider set")
    end

    -- Get model info for cost calculation
    local model_info = bridge:getModelInfo(model_id, provider_name)

    if not model_info or not model_info.cost_per_token then
        return {
            estimated_cost = 0,
            currency = "USD",
            note = "Cost information not available for this model",
        }
    end

    -- Estimate token count (simplified)
    local estimated_tokens
    if type(operation) == "string" then
        -- Rough estimation: ~4 characters per token
        estimated_tokens = math.ceil(#operation / 4)
    elseif type(operation) == "table" and operation.estimated_tokens then
        estimated_tokens = operation.estimated_tokens
    else
        estimated_tokens = 100 -- Default estimate
    end

    local estimated_cost = estimated_tokens * model_info.cost_per_token

    return {
        estimated_cost = estimated_cost,
        currency = model_info.currency or "USD",
        estimated_tokens = estimated_tokens,
        cost_per_token = model_info.cost_per_token,
        model = model_id,
        provider = provider_name,
    }
end

-- Get provider capabilities
function llm.get_provider_capabilities(provider_name)
    local provider_name_to_use = provider_name or current_provider

    if not provider_name_to_use then
        error("No provider specified and no current provider set")
    end

    local bridge = get_llm_bridge()
    return bridge:getCapabilities(provider_name_to_use)
end

-- Utility functions

-- Set default options for all operations
function llm.set_defaults(options)
    if type(options) ~= "table" then
        error("options must be a table")
    end

    for k, v in pairs(options) do
        default_options[k] = v
    end
end

-- Get current default options
function llm.get_defaults()
    local copy = {}
    for k, v in pairs(default_options) do
        copy[k] = v
    end
    return copy
end

-- Reset to original defaults
function llm.reset_defaults()
    default_options = {
        temperature = 0.7,
        max_tokens = 1000,
        timeout = 30,
    }
end

-- Get provider metrics
function llm.get_provider_metrics(provider_name)
    local provider_name_to_use = provider_name or current_provider

    if not provider_name_to_use then
        error("No provider specified and no current provider set")
    end

    local bridge = get_llm_bridge()
    return bridge:getProviderMetrics(provider_name_to_use)
end

-- Reset provider metrics
function llm.reset_provider_metrics(provider_name)
    local provider_name_to_use = provider_name or current_provider

    if not provider_name_to_use then
        error("No provider specified and no current provider set")
    end

    local bridge = get_llm_bridge()
    return bridge:resetProviderMetrics(provider_name_to_use)
end

-- Export the module
return llm
