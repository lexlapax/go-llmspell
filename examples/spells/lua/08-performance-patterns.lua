-- ABOUTME: Example demonstrating performance optimization patterns for LLM spells
-- ABOUTME: Shows caching, batching, parallel processing, and resource management

-- Required modules
local core = require("core")
local agent = require("agent")
local data = require("data")
local utils = require("utils")  -- Add utils for file operations

-- Performance Patterns Example
-- This spell demonstrates performance optimization techniques:
-- 1. Response caching
-- 2. Request batching
-- 3. Parallel processing
-- 4. Token optimization
-- 5. Resource pooling
-- 6. Performance monitoring

-- Parameters
local test_size = tonumber(params.test_size) or 10
local model = params.model or "gpt-3.5-turbo"  -- Use faster model for performance tests
local output_dir = params.output_dir or "./performance-output"

print("=== Performance Patterns Example ===")
print("Test size: " .. test_size)
print("Model: " .. model)
print()

-- Ensure output directory exists
if not utils.file_exists(output_dir) then
    utils.mkdir(output_dir)
end

-- Performance monitoring utilities
local PerformanceMonitor = {}

function PerformanceMonitor:new()
    local obj = {
        metrics = {},
        timers = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function PerformanceMonitor:start_timer(name)
    self.timers[name] = os.clock()
end

function PerformanceMonitor:end_timer(name)
    if self.timers[name] then
        local duration = os.clock() - self.timers[name]
        if not self.metrics[name] then
            self.metrics[name] = {
                count = 0,
                total_time = 0,
                min_time = math.huge,
                max_time = 0
            }
        end
        
        local metric = self.metrics[name]
        metric.count = metric.count + 1
        metric.total_time = metric.total_time + duration
        metric.min_time = math.min(metric.min_time, duration)
        metric.max_time = math.max(metric.max_time, duration)
        
        self.timers[name] = nil
        return duration
    end
    return 0
end

function PerformanceMonitor:get_stats(name)
    local metric = self.metrics[name]
    if metric and metric.count > 0 then
        return {
            count = metric.count,
            total_time = metric.total_time,
            avg_time = metric.total_time / metric.count,
            min_time = metric.min_time,
            max_time = metric.max_time
        }
    end
    return nil
end

function PerformanceMonitor:report()
    print("\n📊 Performance Report:")
    for name, _ in pairs(self.metrics) do
        local stats = self:get_stats(name)
        if stats then
            print(string.format(
                "  %s: Count=%d, Avg=%.3fs, Min=%.3fs, Max=%.3fs, Total=%.3fs",
                name, stats.count, stats.avg_time, stats.min_time, 
                stats.max_time, stats.total_time
            ))
        end
    end
end

-- Create global monitor
local monitor = PerformanceMonitor:new()

-- Example 1: Response Caching
print("=== Example 1: Response Caching ===")

local ResponseCache = {}

function ResponseCache:new(max_size, ttl)
    local obj = {
        cache = {},
        access_times = {},
        max_size = max_size or 100,
        ttl = ttl or 3600,  -- 1 hour default
        hits = 0,
        misses = 0
    }
    setmetatable(obj, {__index = self})
    return obj
end

function ResponseCache:make_key(request)
    -- Create cache key from request
    local key = request.model .. "|" .. request.temperature .. "|"
    for _, msg in ipairs(request.messages) do
        key = key .. msg.role .. ":" .. msg.content .. "|"
    end
    -- Simple hash
    local hash = 0
    for i = 1, #key do
        hash = (hash * 31 + string.byte(key, i)) % 2147483647
    end
    return tostring(hash)
end

function ResponseCache:get(request)
    local key = self:make_key(request)
    local entry = self.cache[key]
    
    if entry then
        -- Check TTL
        if os.time() - entry.timestamp < self.ttl then
            self.hits = self.hits + 1
            self.access_times[key] = os.time()
            return entry.response
        else
            -- Expired
            self.cache[key] = nil
            self.access_times[key] = nil
        end
    end
    
    self.misses = self.misses + 1
    return nil
end

function ResponseCache:set(request, response)
    local key = self:make_key(request)
    
    -- Evict LRU if cache full
    if self:size() >= self.max_size then
        self:evict_lru()
    end
    
    self.cache[key] = {
        response = response,
        timestamp = os.time()
    }
    self.access_times[key] = os.time()
end

function ResponseCache:size()
    local count = 0
    for _ in pairs(self.cache) do count = count + 1 end
    return count
end

function ResponseCache:evict_lru()
    local oldest_key = nil
    local oldest_time = math.huge
    
    for key, time in pairs(self.access_times) do
        if time < oldest_time then
            oldest_time = time
            oldest_key = key
        end
    end
    
    if oldest_key then
        self.cache[oldest_key] = nil
        self.access_times[oldest_key] = nil
    end
end

function ResponseCache:stats()
    local total = self.hits + self.misses
    local hit_rate = total > 0 and (self.hits / total * 100) or 0
    return {
        hits = self.hits,
        misses = self.misses,
        hit_rate = hit_rate,
        size = self:size()
    }
end

-- Test caching
local cache = ResponseCache:new(50, 3600)

local function cached_llm_complete(request)
    monitor:start_timer("cached_llm_call")
    
    -- Check cache
    local cached_response = cache:get(request)
    if cached_response then
        monitor:end_timer("cached_llm_call")
        return cached_response
    end
    
    -- Cache miss - make actual call
    monitor:start_timer("actual_llm_call")
    local response = llm.complete(request)
    monitor:end_timer("actual_llm_call")
    
    -- Store in cache
    cache:set(request, response)
    
    monitor:end_timer("cached_llm_call")
    return response
end

print("\nTesting cache performance:")
local test_prompts = {
    "What is 2+2?",
    "What is the capital of France?",
    "What is 2+2?",  -- Duplicate
    "Explain quantum computing",
    "What is the capital of France?",  -- Duplicate
    "What is 2+2?"  -- Duplicate
}

for i, prompt in ipairs(test_prompts) do
    print("  Request " .. i .. ": " .. prompt:sub(1, 30) .. "...")
    local response = cached_llm_complete({
        model = model,
        messages = {{role = "user", content = prompt}},
        temperature = 0
    })
end

local cache_stats = cache:stats()
print(string.format("\n  Cache Stats: Hits=%d, Misses=%d, Hit Rate=%.1f%%, Size=%d",
    cache_stats.hits, cache_stats.misses, cache_stats.hit_rate, cache_stats.size))
print()

-- Example 2: Request Batching
print("=== Example 2: Request Batching ===")

local RequestBatcher = {}

function RequestBatcher:new(batch_size, flush_interval)
    local obj = {
        batch_size = batch_size or 5,
        flush_interval = flush_interval or 2,  -- seconds
        queue = {},
        callbacks = {},
        last_flush = os.time()
    }
    setmetatable(obj, {__index = self})
    return obj
end

function RequestBatcher:add(request, callback)
    table.insert(self.queue, request)
    table.insert(self.callbacks, callback)
    
    if #self.queue >= self.batch_size then
        self:flush()
    end
end

function RequestBatcher:flush()
    if #self.queue == 0 then return end
    
    monitor:start_timer("batch_processing")
    print("  📦 Processing batch of " .. #self.queue .. " requests")
    
    -- Process batch in parallel
    local promises = {}
    for i, request in ipairs(self.queue) do
        promises[i] = promise.new(function(resolve)
            core.async(function()
                local response = llm.complete(request)
                resolve({index = i, response = response})
            end)
        end)
    end
    
    -- Wait for all
    local results = promise.all(promises):await()
    
    -- Call callbacks
    for _, result in ipairs(results) do
        if self.callbacks[result.index] then
            self.callbacks[result.index](result.response)
        end
    end
    
    -- Clear queue
    self.queue = {}
    self.callbacks = {}
    self.last_flush = os.time()
    
    monitor:end_timer("batch_processing")
end

function RequestBatcher:auto_flush()
    core.async(function()
        while true do
            utils.sleep(0.5)
            if os.time() - self.last_flush >= self.flush_interval and #self.queue > 0 then
                self:flush()
            end
        end
    end)
end

-- Test batching
local batcher = RequestBatcher:new(3, 2)
batcher:auto_flush()

print("\nTesting request batching:")
local batch_results = {}

for i = 1, 7 do
    print("  Adding request " .. i)
    batcher:add({
        model = model,
        messages = {{role = "user", content = "Count to " .. i}},
        temperature = 0,
        max_tokens = 50
    }, function(response)
        batch_results[i] = response
        print("  ✓ Received response for request " .. i)
    end)
    utils.sleep(0.2)  -- Simulate time between requests
end

-- Ensure final flush
utils.sleep(2.5)
print("  Completed " .. #batch_results .. " batched requests")
print()

-- Example 3: Parallel Processing
print("=== Example 3: Parallel Processing ===")

local function process_sequential(items, processor)
    monitor:start_timer("sequential_processing")
    local results = {}
    
    for i, item in ipairs(items) do
        results[i] = processor(item)
    end
    
    monitor:end_timer("sequential_processing")
    return results
end

local function process_parallel(items, processor, max_concurrent)
    monitor:start_timer("parallel_processing")
    max_concurrent = max_concurrent or 5
    local results = {}
    local promises = {}
    
    -- Process in chunks
    for i = 1, #items, max_concurrent do
        local chunk_promises = {}
        
        for j = 0, max_concurrent - 1 do
            local index = i + j
            if index <= #items then
                chunk_promises[j + 1] = promise.new(function(resolve)
                    core.async(function()
                        local result = processor(items[index])
                        resolve({index = index, result = result})
                    end)
                end)
            end
        end
        
        -- Wait for chunk
        local chunk_results = promise.all(chunk_promises):await()
        for _, r in ipairs(chunk_results) do
            results[r.index] = r.result
        end
    end
    
    monitor:end_timer("parallel_processing")
    return results
end

-- Test processor function
local function analyze_text(text)
    return llm.complete({
        model = model,
        messages = {{role = "user", content = "Summarize in 5 words: " .. text}},
        temperature = 0,
        max_tokens = 20
    })
end

-- Generate test data
local test_texts = {}
for i = 1, test_size do
    test_texts[i] = "Sample text number " .. i .. " about various topics"
end

print("\nComparing sequential vs parallel processing (" .. test_size .. " items):")

print("\n  Sequential processing...")
local seq_results = process_sequential(test_texts, analyze_text)

print("\n  Parallel processing (max 3 concurrent)...")
local par_results = process_parallel(test_texts, analyze_text, 3)

print()

-- Example 4: Token Optimization
print("=== Example 4: Token Optimization ===")

local TokenOptimizer = {}

function TokenOptimizer.estimate_tokens(text)
    -- Rough estimate: ~1 token per 4 characters
    return math.ceil(#text / 4)
end

function TokenOptimizer.truncate_to_tokens(text, max_tokens)
    local estimated_chars = max_tokens * 4
    if #text <= estimated_chars then
        return text
    end
    return text:sub(1, estimated_chars - 3) .. "..."
end

function TokenOptimizer.optimize_messages(messages, max_context_tokens)
    local total_tokens = 0
    local optimized = {}
    
    -- Always keep system message
    if messages[1] and messages[1].role == "system" then
        table.insert(optimized, messages[1])
        total_tokens = TokenOptimizer.estimate_tokens(messages[1].content)
    end
    
    -- Add messages from newest to oldest (keep recent context)
    for i = #messages, 1, -1 do
        local msg = messages[i]
        local msg_tokens = TokenOptimizer.estimate_tokens(msg.content)
        
        if total_tokens + msg_tokens <= max_context_tokens then
            table.insert(optimized, 1, msg)  -- Insert at beginning
            total_tokens = total_tokens + msg_tokens
        else
            -- Truncate if this is the last message we can fit
            local remaining = max_context_tokens - total_tokens
            if remaining > 100 then  -- Only truncate if meaningful amount remains
                msg.content = TokenOptimizer.truncate_to_tokens(msg.content, remaining)
                table.insert(optimized, 1, msg)
            end
            break
        end
    end
    
    return optimized, total_tokens
end

-- Test token optimization
print("\nTesting token optimization:")
local long_conversation = {
    {role = "system", content = "You are a helpful assistant."},
    {role = "user", content = string.rep("This is a very long message. ", 50)},
    {role = "assistant", content = string.rep("This is a long response. ", 50)},
    {role = "user", content = "Short question?"},
    {role = "assistant", content = "Short answer."}
}

print("  Original conversation:")
local original_tokens = 0
for i, msg in ipairs(long_conversation) do
    local tokens = TokenOptimizer.estimate_tokens(msg.content)
    original_tokens = original_tokens + tokens
    print(string.format("    Message %d (%s): %d tokens", i, msg.role, tokens))
end
print("  Total tokens: " .. original_tokens)

local optimized, optimized_tokens = TokenOptimizer.optimize_messages(long_conversation, 500)
print("\n  Optimized conversation (max 500 tokens):")
for i, msg in ipairs(optimized) do
    local tokens = TokenOptimizer.estimate_tokens(msg.content)
    print(string.format("    Message %d (%s): %d tokens", i, msg.role, tokens))
end
print("  Total tokens: " .. optimized_tokens)
print()

-- Example 5: Resource Pooling
print("=== Example 5: Resource Pooling ===")

local AgentPool = {}

function AgentPool:new(config)
    local obj = {
        pool = {},
        available = {},
        in_use = {},
        config = config,
        created = 0,
        max_size = config.max_size or 5
    }
    setmetatable(obj, {__index = self})
    
    -- Pre-create some agents
    for i = 1, config.min_size or 2 do
        self:create_agent()
    end
    
    return obj
end

function AgentPool:create_agent()
    if self.created >= self.max_size then
        return nil
    end
    
    local agent = agent.create({
        name = self.config.name .. "_" .. (self.created + 1),
        model = self.config.model,
        system = self.config.system,
        temperature = self.config.temperature
    })
    
    self.created = self.created + 1
    self.pool[self.created] = agent
    self.available[self.created] = true
    
    return self.created
end

function AgentPool:acquire()
    -- Find available agent
    for id, is_available in pairs(self.available) do
        if is_available then
            self.available[id] = false
            self.in_use[id] = true
            return self.pool[id], id
        end
    end
    
    -- Create new if possible
    if self.created < self.max_size then
        local id = self:create_agent()
        if id then
            self.available[id] = false
            self.in_use[id] = true
            return self.pool[id], id
        end
    end
    
    -- No agents available
    return nil, nil
end

function AgentPool:release(id)
    if self.in_use[id] then
        self.in_use[id] = false
        self.available[id] = true
    end
end

function AgentPool:stats()
    local available_count = 0
    local in_use_count = 0
    
    for _, v in pairs(self.available) do
        if v then available_count = available_count + 1 end
    end
    
    for _, v in pairs(self.in_use) do
        if v then in_use_count = in_use_count + 1 end
    end
    
    return {
        total = self.created,
        available = available_count,
        in_use = in_use_count,
        utilization = (in_use_count / self.created) * 100
    }
end

-- Test agent pooling
local pool = AgentPool:new({
    name = "PooledAnalyst",
    model = model,
    system = "You are an analyst. Be concise.",
    temperature = 0.3,
    min_size = 2,
    max_size = 4
})

print("\nTesting agent pool:")
local pool_tasks = {}

-- Simulate concurrent agent usage
for i = 1, 6 do
    pool_tasks[i] = promise.new(function(resolve)
        core.async(function()
            print("  Task " .. i .. ": Requesting agent...")
            local agent, id = pool:acquire()
            
            if agent then
                print("  Task " .. i .. ": Got agent " .. id)
                
                -- Use agent
                local result = agent:run("Analyze the number " .. i)
                
                -- Simulate work
                utils.sleep(math.random() * 2)
                
                -- Release agent
                pool:release(id)
                print("  Task " .. i .. ": Released agent " .. id)
                
                resolve({task = i, result = result})
            else
                print("  Task " .. i .. ": No agent available!")
                resolve({task = i, error = "No agent available"})
            end
        end)
    end)
end

-- Wait for all tasks
local pool_results = promise.all(pool_tasks):await()

local pool_stats = pool:stats()
print(string.format("\n  Pool Stats: Total=%d, Available=%d, In Use=%d, Utilization=%.1f%%",
    pool_stats.total, pool_stats.available, pool_stats.in_use, pool_stats.utilization))
print()

-- Example 6: Performance Monitoring Dashboard
print("=== Example 6: Performance Dashboard ===")

local function create_performance_report()
    local report = {
        timestamp = os.date(),
        cache_performance = cache:stats(),
        processing_stats = {
            sequential = monitor:get_stats("sequential_processing"),
            parallel = monitor:get_stats("parallel_processing"),
            batch = monitor:get_stats("batch_processing")
        },
        agent_pool = pool:stats(),
        optimizations = {
            token_reduction = string.format("%.1f%%", 
                (1 - optimized_tokens / original_tokens) * 100),
            cache_hit_rate = string.format("%.1f%%", cache_stats.hit_rate),
            parallel_speedup = "See timing comparison"
        }
    }
    
    return report
end

-- Generate performance report
local perf_report = create_performance_report()
local report_json = data.to_json(perf_report, {pretty = true})
utils.file_write(output_dir .. "/performance-report.json", report_json)

print("\nPerformance Dashboard:")
print("  📊 Cache Performance:")
print("    - Hit Rate: " .. perf_report.optimizations.cache_hit_rate)
print("    - Cache Size: " .. perf_report.cache_performance.size)
print()
print("  ⚡ Processing Performance:")
if perf_report.processing_stats.sequential then
    print(string.format("    - Sequential: %.3fs avg", 
        perf_report.processing_stats.sequential.avg_time))
end
if perf_report.processing_stats.parallel then
    print(string.format("    - Parallel: %.3fs avg", 
        perf_report.processing_stats.parallel.avg_time))
end
print()
print("  🔧 Optimizations:")
print("    - Token Reduction: " .. perf_report.optimizations.token_reduction)
print("    - Agent Pool Utilization: " .. string.format("%.1f%%", 
    perf_report.agent_pool.utilization))

-- Show detailed performance metrics
monitor:report()

-- Summary
print("\n=== Summary ===")
print("This example demonstrated:")
print("1. Response caching with LRU eviction")
print("2. Request batching for efficiency")
print("3. Parallel vs sequential processing")
print("4. Token optimization strategies")
print("5. Resource pooling for agents")
print("6. Performance monitoring and reporting")
print()
print("Key insights:")
print("- Caching dramatically reduces redundant API calls")
print("- Batching improves throughput for multiple requests")
print("- Parallel processing provides significant speedup")
print("- Token optimization reduces costs and improves speed")
print("- Resource pooling manages agent lifecycle efficiently")
print("- Monitoring helps identify bottlenecks")
print()

-- List generated files
print("Files created in " .. output_dir .. ":")
local files = utils.list_files(output_dir)
for _, file in ipairs(files) do
    print("  - " .. file)
end

-- Return performance summary
return {
    cache_hit_rate = cache_stats.hit_rate,
    parallel_speedup_factor = monitor:get_stats("sequential_processing") and 
                             monitor:get_stats("parallel_processing") and
                             (monitor:get_stats("sequential_processing").avg_time / 
                              monitor:get_stats("parallel_processing").avg_time) or "N/A",
    token_savings_percent = (1 - optimized_tokens / original_tokens) * 100,
    agent_pool_size = pool_stats.total,
    total_llm_calls = (cache_stats.hits + cache_stats.misses) + test_size * 2,
    patterns_demonstrated = {
        "response_caching",
        "request_batching", 
        "parallel_processing",
        "token_optimization",
        "resource_pooling",
        "performance_monitoring"
    }
}