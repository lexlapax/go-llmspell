-- ABOUTME: Model discovery and comparison library for Lua that wraps the llm_modelinfo bridge
-- ABOUTME: Provides model discovery, capability querying, model comparison, and recommendation functionality

local modelinfo = {}

-- Version information
modelinfo._VERSION = "1.0.0"
modelinfo._DESCRIPTION = "Model discovery and comparison library for Lua"

-- Get the modelinfo bridge
local function get_modelinfo_bridge()
    if bridges and bridges.llm_modelinfo then
        return bridges.llm_modelinfo
    end
    return nil
end

-- Check if bridge is available
local function with_modelinfo_bridge(method_name, ...)
    local modelinfo_bridge = get_modelinfo_bridge()
    if not modelinfo_bridge then
        error("ModelInfo bridge not available for " .. method_name)
    end
    return modelinfo_bridge[method_name](...)
end

-- Constants
modelinfo.CAPABILITIES = {
    TEXT_READ = "text.read",
    TEXT_WRITE = "text.write", 
    IMAGE_READ = "image.read",
    IMAGE_WRITE = "image.write",
    AUDIO_READ = "audio.read",
    AUDIO_WRITE = "audio.write",
    VIDEO_READ = "video.read",
    VIDEO_WRITE = "video.write",
    FILE_READ = "file.read",
    FILE_WRITE = "file.write",
    FUNCTION_CALLING = "functionCalling",
    STREAMING = "streaming"
}

modelinfo.RANKING = {
    COST = "cost",
    PERFORMANCE = "performance",
    QUALITY = "quality",
    FEATURES = "features"
}

-- Re-export from adapter PRIORITIES
modelinfo.PRIORITIES = {
    COST = "cost",
    PERFORMANCE = "performance",
    CONTEXT_WINDOW = "context_window",
    CAPABILITY = "capability"
}

-- Re-export from adapter TASKS
modelinfo.TASKS = {
    FUNCTION_CALLING = "function_calling",
    TEXT_GENERATION = "text_generation",
    CODE_GENERATION = "code_generation",
    ANALYSIS = "analysis"
}

-- Discovery methods (snake_case) - Flattened API
function modelinfo.discovery_scan()
    return with_modelinfo_bridge("discoveryScan")
end

function modelinfo.discovery_refresh()
    return with_modelinfo_bridge("discoveryRefresh")
end

function modelinfo.discovery_get_providers()
    return with_modelinfo_bridge("discoveryGetProviders")
end

function modelinfo.discovery_get_models()
    return with_modelinfo_bridge("discoveryGetModels")
end

-- Legacy discovery methods (for compatibility)
function modelinfo.list_models()
    return with_modelinfo_bridge("listModels")
end

function modelinfo.fetch_inventory()
    return with_modelinfo_bridge("fetchModelInventory")
end

function modelinfo.get_model(model_name)
    return with_modelinfo_bridge("getModel", model_name)
end

function modelinfo.list_registries()
    return with_modelinfo_bridge("listRegistries")
end

-- Capabilities methods (snake_case) - Flattened API
function modelinfo.capabilities_check(model_name)
    return with_modelinfo_bridge("capabilitiesCheck", model_name)
end

function modelinfo.capabilities_list()
    return with_modelinfo_bridge("capabilitiesList")
end

function modelinfo.capabilities_compare(capability)
    return with_modelinfo_bridge("capabilitiesCompare", capability)
end

function modelinfo.capabilities_get_details(model_name)
    return with_modelinfo_bridge("capabilitiesGetDetails", model_name)
end

-- Legacy capabilities methods (for compatibility)
function modelinfo.get_model_capabilities(model_name)
    return with_modelinfo_bridge("getModelCapabilities", model_name)
end

function modelinfo.find_models_by_capability(capability)
    return with_modelinfo_bridge("findModelsByCapability", capability)
end

-- Selection methods (snake_case) - Flattened API
function modelinfo.selection_find(requirements)
    return with_modelinfo_bridge("selectionFind", requirements)
end

function modelinfo.selection_rank(criteria)
    return with_modelinfo_bridge("selectionRank", criteria)
end

function modelinfo.selection_filter(filters)
    return with_modelinfo_bridge("selectionFilter", filters)
end

function modelinfo.selection_recommend(task)
    return with_modelinfo_bridge("selectionRecommend", task)
end

-- Legacy selection methods (for compatibility)
function modelinfo.suggest_model(requirements)
    return with_modelinfo_bridge("suggestModel", requirements)
end

function modelinfo.compare_models(model_names)
    return with_modelinfo_bridge("compareModels", model_names)
end

function modelinfo.estimate_cost(model_name, usage)
    return with_modelinfo_bridge("estimateCost", model_name, usage)
end

function modelinfo.get_best_model_for_task(task)
    return with_modelinfo_bridge("getBestModelForTask", task)
end

-- Namespace-based API (maintains backward compatibility)
modelinfo.discovery = {
    list_models = modelinfo.list_models,
    fetch_inventory = modelinfo.fetch_inventory,
    get_providers = modelinfo.discovery_get_providers,
    get_models = modelinfo.discovery_get_models,
    scan = modelinfo.discovery_scan,
    refresh = modelinfo.discovery_refresh
}

modelinfo.capabilities = {
    get_model_capabilities = modelinfo.get_model_capabilities,
    find_models_by_capability = modelinfo.find_models_by_capability,
    check = modelinfo.capabilities_check,
    list = modelinfo.capabilities_list,
    compare = modelinfo.capabilities_compare,
    get_details = modelinfo.capabilities_get_details
}

modelinfo.selection = {
    suggest_model = modelinfo.suggest_model,
    compare_models = modelinfo.compare_models,
    estimate_cost = modelinfo.estimate_cost,
    get_best_model_for_task = modelinfo.get_best_model_for_task,
    find = modelinfo.selection_find,
    rank = modelinfo.selection_rank,
    filter = modelinfo.selection_filter,
    recommend = modelinfo.selection_recommend
}

-- Helper functions for common patterns
function modelinfo.find_cheapest_model(capabilities_required)
    local requirements = {
        capabilities = capabilities_required or {},
        priority = modelinfo.PRIORITIES.COST
    }
    return modelinfo.selection_find(requirements)
end

function modelinfo.find_largest_context_model(min_context_window)
    local requirements = {
        minContextWindow = min_context_window or 0,
        priority = modelinfo.PRIORITIES.CONTEXT_WINDOW
    }
    return modelinfo.selection_find(requirements)
end

function modelinfo.find_most_capable_model(capabilities_required)
    local requirements = {
        capabilities = capabilities_required or {},
        priority = modelinfo.PRIORITIES.CAPABILITY
    }
    return modelinfo.selection_find(requirements)
end

-- Quick model check functions
function modelinfo.supports_function_calling(model_name)
    local caps = modelinfo.capabilities_check(model_name)
    if caps and caps.functionCalling then
        return true
    end
    return false
end

function modelinfo.supports_streaming(model_name)
    local caps = modelinfo.capabilities_check(model_name)
    if caps and caps.streaming then
        return true
    end
    return false
end

function modelinfo.supports_images(model_name)
    local caps = modelinfo.capabilities_check(model_name)
    if caps and caps.image and (caps.image.read or caps.image.write) then
        return true
    end
    return false
end

-- Cost estimation helper
function modelinfo.estimate_conversation_cost(model_name, num_messages, avg_tokens_per_message)
    local input_tokens = num_messages * avg_tokens_per_message
    local output_tokens = num_messages * avg_tokens_per_message * 0.8 -- Estimate 80% of input as output
    
    return modelinfo.estimate_cost(model_name, {
        inputTokens = input_tokens,
        outputTokens = output_tokens
    })
end

-- Model comparison helper
function modelinfo.compare_all_by_capability(capability)
    local models_with_capability = modelinfo.capabilities_compare(capability)
    if not models_with_capability or #models_with_capability == 0 then
        return nil, "No models found with capability: " .. capability
    end
    
    -- Extract model names
    local model_names = {}
    for _, model in ipairs(models_with_capability) do
        if model.name then
            table.insert(model_names, model.name)
        end
    end
    
    if #model_names == 0 then
        return nil, "No model names could be extracted"
    end
    
    return modelinfo.compare_models(model_names)
end

-- Ranking helper
function modelinfo.get_top_n_models(criteria, n)
    n = n or 5
    local ranked = modelinfo.selection_rank(criteria)
    if not ranked then
        return nil
    end
    
    -- Take top N
    local top_models = {}
    for i = 1, math.min(n, #ranked) do
        table.insert(top_models, ranked[i])
    end
    
    return top_models
end

-- Filter helper
function modelinfo.find_models_matching(filters)
    -- Convert common filter patterns
    local filter_table = {}
    
    if filters.min_context then
        filter_table.minContextWindow = filters.min_context
    end
    
    if filters.max_cost then
        filter_table.maxCost = filters.max_cost
    end
    
    if filters.provider then
        filter_table.provider = filters.provider
    end
    
    if filters.capabilities then
        filter_table.capabilities = filters.capabilities
    end
    
    return modelinfo.selection_filter(filter_table)
end

-- Bridge access to raw methods if available
if bridges and bridges.llm_modelinfo then
    -- Allow access to raw bridge methods
    modelinfo.bridge = bridges.llm_modelinfo
end

return modelinfo