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
    if not bridges or not bridges.llm_core then
        error("LLM bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return bridges.llm_core
end

-- Helper function to get providers bridge (optional)
local function get_providers_bridge()
    if bridges and bridges.llm_providers then
        return bridges.llm_providers
    end
    return nil
end

-- Helper function to get pool bridge (optional) 
local function get_pool_bridge()
    if bridges and bridges.llm_pool then
        return bridges.llm_pool
    end
    return nil
end

-- High-level LLM operation helpers

-- Quick prompt for simple prompting
function llm.quick_prompt(prompt, options)
    validate_required(prompt, "prompt")

    local opts = merge_options(options)
    local bridge = get_llm_bridge()

    -- Use current provider if set, otherwise use default
    if current_provider then
        return bridge.generateWithProvider(current_provider, prompt, opts)
    else
        return bridge.generate(prompt, opts)
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
            response = bridge.generateWithProvider(current_provider, self.messages, opts)
        else
            response = bridge.generateMessage(self.messages, opts)
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
        stream_id = bridge.streamWithProvider(current_provider, prompt, opts)
    else
        stream_id = bridge.stream(prompt, opts)
    end

    if not stream_id then
        error("Failed to start streaming")
    end

    -- Read stream in chunks and call callback
    local function read_stream()
        while true do
            local chunk = bridge.readStream(stream_id)
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
        bridge.closeStream(stream_id)
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
                response = bridge.generateWithProvider(current_provider, prompt, opts)
            else
                response = bridge.generate(prompt, opts)
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

    -- Set the provider in the bridge
    local success, err = pcall(function()
        if config then
            return bridge.setProvider(name, config)
        else
            -- Try setting with default config
            return bridge.setProvider(name, {})
        end
    end)
    
    if not success then
        error("Failed to set provider '" .. name .. "': " .. tostring(err))
    end

    current_provider = name
    return true
end

-- Get current provider
function llm.get_current_provider()
    return current_provider
end

-- List available providers
function llm.list_providers()
    local bridge = get_llm_bridge()
    return bridge.listProviders()
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
            return bridge.generateWithProvider(provider_name, prompt, opts)
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
    return bridge.setFallbackChain(providers)
end

-- Get current fallback chain
function llm.get_fallback_chain()
    local bridge = get_llm_bridge()
    return bridge.getFallbackChain()
end

-- Generate with fallback chain
function llm.generate_with_fallback(prompt, options)
    validate_required(prompt, "prompt")

    local opts = merge_options(options)
    local bridge = get_llm_bridge()
    local fallback_chain = bridge.getFallbackChain()

    if not fallback_chain or #fallback_chain == 0 then
        error("No fallback chain configured. Use llm.setup_fallback_chain() first.")
    end

    local last_error

    -- Try each provider in the fallback chain
    for _, provider_name in ipairs(fallback_chain) do
        local success, result = pcall(function()
            return bridge.generateWithProvider(provider_name, prompt, opts)
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
    local providers_to_check = requirements.providers or bridge.listProviders()

    local suitable_models = {}

    for _, provider_name in ipairs(providers_to_check) do
        local models = bridge.listModels(provider_name)

        if models then
            for _, model in ipairs(models) do
                local model_info = bridge.getModelInfo(model.id)

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
        return bridge.getModelInfo(model_id, provider)
    else
        -- Get info from current or default provider
        return bridge.getModelInfo(model_id)
    end
end

-- Estimate cost for an operation
function llm.cost_estimate(operation, model, provider)
    validate_required(operation, "operation")

    local bridge = get_llm_bridge()
    local util_bridge = bridges and bridges.util_llm

    if not util_bridge then
        error("LLM utilities bridge not available")
    end

    local model_id = model or "default"
    local provider_name = provider or current_provider

    if not provider_name then
        error("No provider specified and no current provider set")
    end

    -- Get model info for cost calculation
    local model_info = bridge.getModelInfo(model_id, provider_name)

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
    return bridge.getCapabilities(provider_name_to_use)
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
    return bridge.getProviderMetrics(provider_name_to_use)
end

-- Reset provider metrics
function llm.reset_provider_metrics(provider_name)
    local provider_name_to_use = provider_name or current_provider

    if not provider_name_to_use then
        error("No provider specified and no current provider set")
    end

    local bridge = get_llm_bridge()
    return bridge.resetProviderMetrics(provider_name_to_use)
end

-- ===================================================================
-- MISSING LLMADAPTER METHODS - Adding for full compatibility  
-- ===================================================================

-- Core LLM Methods (from base bridge)

-- Generate text from prompt
function llm.generate(prompt, options)
    validate_required(prompt, "prompt")
    
    local bridge = get_llm_bridge()
    local opts = merge_options(options)
    
    return bridge.generate(prompt, opts)
end

-- Generate from message array
function llm.generateMessage(messages, options)
    validate_required(messages, "messages")
    
    if type(messages) ~= "table" then
        error("messages must be a table")
    end
    
    local bridge = get_llm_bridge()
    local opts = merge_options(options)
    
    return bridge.generateMessage(messages, opts)
end

-- Stream response
function llm.stream(prompt, options)
    validate_required(prompt, "prompt")
    
    local bridge = get_llm_bridge()
    local opts = merge_options(options)
    
    return bridge.stream(prompt, opts)
end

-- Count tokens in text
function llm.countTokens(text, model)
    validate_required(text, "text")
    
    local bridge = get_llm_bridge()
    local model_name = model or ""
    
    return bridge.countTokens(text, model_name)
end

-- Create agent
function llm.createAgent(config)
    local bridge = get_llm_bridge()
    local agent_config = config or {}
    
    return bridge.createAgent(agent_config)
end

-- Agent alias
function llm.Agent(config)
    return llm.createAgent(config)
end

-- Agent completion
function llm.agentComplete(agentId, prompt, options)
    validate_required(agentId, "agentId")
    validate_required(prompt, "prompt")
    
    local bridge = get_llm_bridge()
    local opts = merge_options(options)
    
    return bridge.agentComplete(agentId, prompt, opts)
end

-- Agent streaming
function llm.agentStream(agentId, prompt, options)
    validate_required(agentId, "agentId")
    validate_required(prompt, "prompt")
    
    local bridge = get_llm_bridge()
    local opts = merge_options(options)
    
    return bridge.agentStream(agentId, prompt, opts)
end

-- Convenience Methods

-- Quick completion with default model
function llm.quick(prompt)
    validate_required(prompt, "prompt")
    
    return llm.generate(prompt)
end

-- Batch completion
function llm.batchComplete(prompts, options)
    if type(prompts) ~= "table" then
        error("prompts must be a table")
    end
    
    local results = {}
    local opts = merge_options(options)
    
    for i, prompt in ipairs(prompts) do
        if type(prompt) == "string" then
            local success, result = pcall(function()
                return llm.generate(prompt, opts)
            end)
            
            if success then
                results[i] = result
            else
                results[i] = {error = result}
            end
        else
            results[i] = {error = "Invalid prompt type at index " .. i}
        end
    end
    
    return results
end

-- Provider Methods (flattened naming)

-- Create provider
function llm.providersCreate(providerType, name, config)
    validate_required(providerType, "providerType")
    validate_required(name, "name")
    
    local bridge = get_llm_bridge()
    local provider_config = config or {}
    
    return bridge.createProvider(providerType, name, provider_config)
end

-- Get provider
function llm.providersGet(name)
    validate_required(name, "name")
    
    local bridge = get_llm_bridge()
    return bridge.getProvider(name)
end

-- List providers
function llm.providersList()
    local bridge = get_llm_bridge()
    return bridge.listProviders()
end

-- Get provider template
function llm.providersGetTemplate(templateName)
    validate_required(templateName, "templateName")
    
    local bridge = get_llm_bridge()
    return bridge.getProviderTemplate(templateName)
end

-- Create multi-provider
function llm.providersCreateMulti(name, providerList, strategy, config)
    validate_required(name, "name")
    validate_required(providerList, "providerList")
    validate_required(strategy, "strategy")
    
    if type(providerList) ~= "table" then
        error("providerList must be a table")
    end
    
    local bridge = get_llm_bridge()
    local provider_config = config or {}
    
    return bridge.createMultiProvider(name, providerList, strategy, provider_config)
end

-- Provider methods that require providers bridge
local function with_providers_bridge(method_name, ...)
    local providers_bridge = get_providers_bridge()
    if not providers_bridge then
        error("Providers bridge not available for " .. method_name)
    end
    return providers_bridge[method_name](...)
end

-- Create provider from environment
function llm.providersCreateFromEnvironment(providerType, name)
    validate_required(providerType, "providerType")
    validate_required(name, "name")
    
    return with_providers_bridge("createProviderFromEnvironment", providerType, name)
end

-- Remove provider
function llm.providersRemove(name)
    validate_required(name, "name")
    
    return with_providers_bridge("removeProvider", name)
end

-- List provider templates
function llm.providersTemplatesList()
    return with_providers_bridge("listProviderTemplates")
end

-- Validate provider config
function llm.providersTemplatesValidate(providerType, config)
    validate_required(providerType, "providerType")
    validate_required(config, "config")
    
    if type(config) ~= "table" then
        error("config must be a table")
    end
    
    return with_providers_bridge("validateProviderConfig", providerType, config)
end

-- Configure multi-provider
function llm.providersConfigureMulti(name, config)
    validate_required(name, "name")
    validate_required(config, "config")
    
    if type(config) ~= "table" then
        error("config must be a table")
    end
    
    return with_providers_bridge("configureMultiProvider", name, config)
end

-- Get multi-provider
function llm.providersGetMulti(name)
    validate_required(name, "name")
    
    return with_providers_bridge("getMultiProvider", name)
end

-- Create mock provider
function llm.providersCreateMock(name, responses)
    validate_required(name, "name")
    validate_required(responses, "responses")
    
    if type(responses) ~= "table" then
        error("responses must be a table")
    end
    
    return with_providers_bridge("createMockProvider", name, responses)
end

-- Generate with specific provider
function llm.providersGenerateWith(providerName, prompt, options)
    validate_required(providerName, "providerName")
    validate_required(prompt, "prompt")
    
    local opts = merge_options(options)
    return with_providers_bridge("generateWithProvider", providerName, prompt, opts)
end

-- Export provider config
function llm.providersExportConfig()
    return with_providers_bridge("exportProviderConfig")
end

-- Import provider config
function llm.providersImportConfig(config)
    validate_required(config, "config")
    
    if type(config) ~= "table" then
        error("config must be a table")
    end
    
    return with_providers_bridge("importProviderConfig", config)
end

-- Set provider metadata
function llm.providersSetMetadata(providerName, metadata)
    validate_required(providerName, "providerName")
    validate_required(metadata, "metadata")
    
    if type(metadata) ~= "table" then
        error("metadata must be a table")
    end
    
    return with_providers_bridge("setProviderMetadata", providerName, metadata)
end

-- Get provider metadata
function llm.providersGetMetadata(providerName)
    validate_required(providerName, "providerName")
    
    return with_providers_bridge("getProviderMetadata", providerName)
end

-- List providers by capability
function llm.providersListByCapability(capability)
    validate_required(capability, "capability")
    
    return with_providers_bridge("listProvidersByCapability", capability)
end

-- Pool Methods (flattened naming)

-- Create pool
function llm.poolCreate(name, providers, strategy, config)
    validate_required(name, "name")
    validate_required(providers, "providers")
    validate_required(strategy, "strategy")
    
    if type(providers) ~= "table" then
        error("providers must be a table")
    end
    
    local bridge = get_llm_bridge()
    local pool_config = config or {}
    
    return bridge.createPool(name, providers, strategy, pool_config)
end

-- Get pool health
function llm.poolGetHealth(poolName)
    validate_required(poolName, "poolName")
    
    local bridge = get_llm_bridge()
    return bridge.getPoolHealth(poolName)
end

-- Generate with pool
function llm.poolGenerate(poolName, prompt, options)
    validate_required(poolName, "poolName")
    validate_required(prompt, "prompt")
    
    local bridge = get_llm_bridge()
    local opts = merge_options(options)
    
    return bridge.generateWithPool(poolName, prompt, opts)
end

-- Get pool metrics
function llm.poolGetMetrics(poolName)
    validate_required(poolName, "poolName")
    
    local bridge = get_llm_bridge()
    return bridge.getPoolMetrics(poolName)
end

-- Pool methods that require pool bridge
local function with_pool_bridge(method_name, ...)
    local pool_bridge = get_pool_bridge()
    if not pool_bridge then
        error("Pool bridge not available for " .. method_name)
    end
    return pool_bridge[method_name](...)
end

-- Get pool
function llm.poolGet(poolName)
    validate_required(poolName, "poolName")
    
    return with_pool_bridge("getPool", poolName)
end

-- List pools
function llm.poolList()
    return with_pool_bridge("listPools")
end

-- Remove pool
function llm.poolRemove(poolName)
    validate_required(poolName, "poolName")
    
    return with_pool_bridge("removePool", poolName)
end

-- Get provider health in pool
function llm.poolGetProviderHealth(poolName)
    validate_required(poolName, "poolName")
    
    return with_pool_bridge("getProviderHealth", poolName)
end

-- Reset pool metrics
function llm.poolResetMetrics(poolName)
    validate_required(poolName, "poolName")
    
    return with_pool_bridge("resetPoolMetrics", poolName)
end

-- Generate message with pool
function llm.poolGenerateMessage(poolName, messages, options)
    validate_required(poolName, "poolName")
    validate_required(messages, "messages")
    
    if type(messages) ~= "table" then
        error("messages must be a table")
    end
    
    local opts = merge_options(options)
    return with_pool_bridge("generateMessageWithPool", poolName, messages, opts)
end

-- Stream with pool
function llm.poolStream(poolName, prompt, options)
    validate_required(poolName, "poolName")
    validate_required(prompt, "prompt")
    
    local opts = merge_options(options)
    return with_pool_bridge("streamWithPool", poolName, prompt, opts)
end

-- Object pooling methods
function llm.poolGetResponse()
    return with_pool_bridge("getResponseFromPool")
end

function llm.poolReturnResponse(response)
    validate_required(response, "response")
    
    if type(response) ~= "table" then
        error("response must be a table")
    end
    
    return with_pool_bridge("returnResponseToPool", response)
end

function llm.poolGetToken()
    return with_pool_bridge("getTokenFromPool")
end

function llm.poolReturnToken(token)
    validate_required(token, "token")
    
    if type(token) ~= "table" then
        error("token must be a table")
    end
    
    return with_pool_bridge("returnTokenToPool", token)
end

function llm.poolGetChannel()
    return with_pool_bridge("getChannelFromPool")
end

function llm.poolReturnChannel(channel)
    validate_required(channel, "channel")
    
    if type(channel) ~= "table" then
        error("channel must be a table")
    end
    
    return with_pool_bridge("returnChannelToPool", channel)
end

-- Model Methods (flattened naming)

-- List models
function llm.modelsList(provider)
    local bridge = get_llm_bridge()
    local provider_name = provider or ""
    
    return bridge.listModels(provider_name)
end

-- Get model info
function llm.modelsGetInfo(modelName)
    validate_required(modelName, "modelName")
    
    local bridge = get_llm_bridge()
    return bridge.getModelInfo(modelName)
end

-- Check model capabilities
function llm.modelsCheckCapabilities(modelName, capability)
    validate_required(modelName, "modelName")
    validate_required(capability, "capability")
    
    local bridge = get_llm_bridge()
    return bridge.checkModelCapability(modelName, capability)
end

-- Namespace Methods (for organized access)

-- Provider namespace
llm.providers = {
    create = llm.providersCreate,
    get = llm.providersGet,
    list = llm.providersList,
    getTemplate = llm.providersGetTemplate,
    createMulti = llm.providersCreateMulti,
    createFromEnvironment = llm.providersCreateFromEnvironment,
    remove = llm.providersRemove,
    templatesList = llm.providersTemplatesList,
    templatesValidate = llm.providersTemplatesValidate,
    configureMulti = llm.providersConfigureMulti,
    getMulti = llm.providersGetMulti,
    createMock = llm.providersCreateMock,
    generateWith = llm.providersGenerateWith,
    exportConfig = llm.providersExportConfig,
    importConfig = llm.providersImportConfig,
    setMetadata = llm.providersSetMetadata,
    getMetadata = llm.providersGetMetadata,
    listByCapability = llm.providersListByCapability
}

-- Pool namespace
llm.pool = {
    create = llm.poolCreate,
    getHealth = llm.poolGetHealth,
    generate = llm.poolGenerate,
    getMetrics = llm.poolGetMetrics,
    get = llm.poolGet,
    list = llm.poolList,
    remove = llm.poolRemove,
    getProviderHealth = llm.poolGetProviderHealth,
    resetMetrics = llm.poolResetMetrics,
    generateMessage = llm.poolGenerateMessage,
    stream = llm.poolStream,
    getResponse = llm.poolGetResponse,
    returnResponse = llm.poolReturnResponse,
    getToken = llm.poolGetToken,
    returnToken = llm.poolReturnToken,
    getChannel = llm.poolGetChannel,
    returnChannel = llm.poolReturnChannel
}

-- Models namespace
llm.models = {
    list = llm.modelsList,
    getInfo = llm.modelsGetInfo,
    checkCapabilities = llm.modelsCheckCapabilities
}

-- Constants (from LLMAdapter)

-- Model constants
llm.MODELS = {
    GPT4 = "gpt-4",
    GPT35_TURBO = "gpt-3.5-turbo", 
    CLAUDE3 = "claude-3",
    CLAUDE2 = "claude-2"
}

-- Default options
llm.DEFAULTS = {
    temperature = 0.7,
    maxTokens = 1000,
    topP = 1.0
}

-- Error codes
llm.ERRORS = {
    RATE_LIMIT = "rate_limit_exceeded",
    INVALID_MODEL = "invalid_model",
    CONTEXT_LENGTH = "context_length_exceeded"
}

-- Pool strategies
llm.STRATEGIES = {
    ROUND_ROBIN = "round_robin",
    FAILOVER = "failover",
    FASTEST = "fastest",
    WEIGHTED = "weighted",
    LEAST_USED = "least_used"
}

-- Export the module
return llm
