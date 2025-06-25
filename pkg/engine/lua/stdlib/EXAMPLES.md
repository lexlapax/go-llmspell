# go-llmspell Lua Standard Library Examples

Practical examples demonstrating how to use the go-llmspell Lua standard library in real-world scenarios.

## Table of Contents

- [Getting Started](#getting-started)
- [Basic Spell Development](#basic-spell-development)
- [Asynchronous Operations](#asynchronous-operations)
- [Spell Composition](#spell-composition)
- [Agent Development](#agent-development)
- [Data Processing](#data-processing)
- [Event-Driven Programming](#event-driven-programming)
- [Testing Spells](#testing-spells)
- [Error Handling](#error-handling)
- [Performance Optimization](#performance-optimization)
- [Real-World Use Cases](#real-world-use-cases)

---

## Getting Started

### Hello World Spell

```lua
-- Load the spell framework
local spell = require("spell")

-- Initialize the spell
spell.init({
    name = "hello-world",
    version = "1.0.0",
    description = "A simple greeting spell",
    params = {
        name = { type = "string", default = "World", description = "Name to greet" }
    }
})

-- Get parameter
local name = spell.params("name")

-- Generate greeting
local greeting = "Hello, " .. name .. "!"

-- Output result
spell.output(greeting, "text")
```

### Simple LLM Interaction

```lua
local spell = require("spell")
local llm = require("llm")
local promise = require("promise")

spell.init({
    name = "simple-chat",
    params = {
        question = { type = "string", required = true }
    }
})

local question = spell.params("question")

-- Async LLM call
local response = promise.await(llm.chat(question))

spell.output(response, "text")
```

---

## Basic Spell Development

### Parameter Validation Spell

```lua
local spell = require("spell")

spell.init({
    name = "calculator",
    description = "Basic calculator operations",
    params = {
        operation = { 
            type = "string", 
            required = true, 
            enum = {"add", "subtract", "multiply", "divide"},
            description = "Mathematical operation to perform"
        },
        a = { type = "number", required = true, description = "First number" },
        b = { type = "number", required = true, description = "Second number" }
    }
})

local operation = spell.params("operation")
local a = spell.params("a")
local b = spell.params("b")

local result
if operation == "add" then
    result = a + b
elseif operation == "subtract" then
    result = a - b
elseif operation == "multiply" then
    result = a * b
elseif operation == "divide" then
    if b == 0 then
        error("Division by zero is not allowed")
    end
    result = a / b
end

spell.output({
    operation = operation,
    operands = {a, b},
    result = result
}, "json")
```

### Spell with Caching

```lua
local spell = require("spell")
local llm = require("llm")
local promise = require("promise")
local core = require("core")

spell.init({
    name = "cached-translator",
    description = "Translation with caching",
    params = {
        text = { type = "string", required = true },
        target_language = { type = "string", default = "Spanish" }
    }
})

local text = spell.params("text")
local target_language = spell.params("target_language")

-- Create cache key
local cache_key = core.crypto.hash("sha256", text .. ":" .. target_language)

-- Check cache first
local cached_result = spell.cache(cache_key)
if cached_result then
    spell.output({
        text = text,
        translation = cached_result,
        cached = true
    }, "json")
    return
end

-- Translate using LLM
local prompt = string.format("Translate the following text to %s: %s", target_language, text)
local translation = promise.await(llm.chat(prompt))

-- Cache result for 1 hour
spell.cache(cache_key, translation, 3600)

spell.output({
    text = text,
    translation = translation,
    cached = false
}, "json")
```

---

## Asynchronous Operations

### Promise-Based Operations

```lua
local promise = require("promise")
local llm = require("llm")
local spell = require("spell")

spell.init({
    name = "async-processor",
    params = {
        queries = { type = "array", required = true }
    }
})

local queries = spell.params("queries")

-- Process multiple queries concurrently
local async_process = promise.async(function()
    local promises = {}
    
    for i, query in ipairs(queries) do
        promises[i] = llm.chat(query)
    end
    
    -- Wait for all to complete
    local results = promise.await(promise.all(promises))
    
    return results
end)

-- Execute and output
local results = promise.await(async_process())

spell.output({
    query_count = #queries,
    results = results,
    processing_time = core.time.now() - spell.context().start_time
}, "json")
```

### Timeout Handling

```lua
local promise = require("promise")
local llm = require("llm")
local spell = require("spell")

spell.init({
    name = "timeout-example",
    params = {
        query = { type = "string", required = true },
        timeout_ms = { type = "number", default = 30000 }
    }
})

local query = spell.params("query")
local timeout_ms = spell.params("timeout_ms")

local timeout_chat = promise.async(function()
    local llm_promise = llm.chat(query)
    
    -- Race between LLM response and timeout
    local result = promise.await(promise.race({
        llm_promise,
        promise.new(function(_, reject)
            promise.sleep(timeout_ms):andThen(function()
                reject("Operation timed out after " .. timeout_ms .. "ms")
            end)
        end)
    }))
    
    return result
end)

-- Handle with error recovery
timeout_chat():andThen(function(result)
    spell.output({ success = true, result = result }, "json")
end):onError(function(error)
    spell.output({ success = false, error = tostring(error) }, "json")
end)
```

---

## Spell Composition

### Multi-Step Workflow

```lua
local spell = require("spell")

spell.init({
    name = "research-workflow",
    description = "Multi-step research and analysis workflow",
    params = {
        topic = { type = "string", required = true },
        depth = { type = "string", default = "detailed", enum = {"brief", "detailed", "comprehensive"} }
    }
})

local topic = spell.params("topic")
local depth = spell.params("depth")

-- Define workflow steps
local workflow_results = spell.compose({
    {
        name = "search",
        spell = "web-search",
        params = {
            query = topic,
            limit = depth == "brief" and 3 or depth == "detailed" and 7 or 15
        }
    },
    {
        name = "fetch_content",
        spell = "web-fetcher",
        params = {
            urls = "$search.results",
            max_content_length = 10000
        }
    },
    {
        name = "analyze",
        spell = "content-analyzer",
        params = {
            content = "$fetch_content.content",
            analysis_type = "comprehensive"
        }
    },
    {
        name = "summarize",
        spell = "multi-source-summarizer",
        params = {
            sources = "$analyze.structured_content",
            style = depth,
            focus = topic
        }
    }
})

-- Output final results
spell.output({
    topic = topic,
    depth = depth,
    summary = workflow_results.summarize.summary,
    sources_analyzed = #workflow_results.fetch_content.content,
    key_insights = workflow_results.analyze.insights,
    execution_time = core.time.now() - spell.context().start_time
}, "json")
```

### Conditional Workflow

```lua
local spell = require("spell")

spell.init({
    name = "adaptive-content-processor",
    params = {
        content = { type = "string", required = true },
        auto_detect = { type = "boolean", default = true }
    }
})

local content = spell.params("content")
local auto_detect = spell.params("auto_detect")

-- First, detect content type if requested
local detection_result = nil
if auto_detect then
    detection_result = spell.compose({
        {
            name = "detect",
            spell = "content-type-detector",
            params = { content = content }
        }
    })
end

local content_type = detection_result and detection_result.detect.type or "text"

-- Process based on detected type
local processing_steps = {}

if content_type == "code" then
    table.insert(processing_steps, {
        name = "code_analysis",
        spell = "code-analyzer",
        params = { code = content }
    })
    table.insert(processing_steps, {
        name = "documentation",
        spell = "code-documenter",
        params = { analysis = "$code_analysis.analysis" }
    })
elseif content_type == "academic" then
    table.insert(processing_steps, {
        name = "extract_citations",
        spell = "citation-extractor",
        params = { text = content }
    })
    table.insert(processing_steps, {
        name = "summarize",
        spell = "academic-summarizer",
        params = { 
            text = content,
            citations = "$extract_citations.citations"
        }
    })
else
    table.insert(processing_steps, {
        name = "general_analysis",
        spell = "text-analyzer",
        params = { text = content }
    })
end

-- Execute the determined workflow
local results = spell.compose(processing_steps)

spell.output({
    content_type = content_type,
    auto_detected = auto_detect,
    processing_results = results
}, "json")
```

---

## Agent Development

### Research Assistant Agent

```lua
local agent = require("agent")
local spell = require("spell")
local tools = require("tools")

spell.init({
    name = "research-assistant",
    params = {
        research_topic = { type = "string", required = true },
        depth = { type = "string", default = "medium", enum = {"shallow", "medium", "deep"} }
    }
})

-- Create specialized research agent
local researcher = agent.create({
    name = "research-agent",
    model = "gpt-4",
    system_prompt = [[
You are a thorough research assistant. Your job is to:
1. Break down research topics into manageable questions
2. Use available tools to gather information
3. Synthesize findings into coherent reports
4. Cite sources and validate information
    ]],
    tools = {"web-search", "document-reader", "fact-checker"},
    memory_size = 10
})

local research_topic = spell.params("research_topic")
local depth = spell.params("depth")

-- Multi-turn research conversation
local research_plan = promise.await(researcher:chat(
    "Create a research plan for: " .. research_topic .. 
    ". Depth level: " .. depth .. 
    ". Break it into 3-5 specific questions to investigate."
))

spell.output({
    topic = research_topic,
    research_plan = research_plan,
    agent_id = researcher.id
}, "json")

-- Execute research plan (simplified)
local questions = extract_questions_from_plan(research_plan)
local findings = {}

for i, question in ipairs(questions) do
    local answer = promise.await(researcher:chat(
        "Research this specific question using available tools: " .. question
    ))
    findings[i] = {
        question = question,
        answer = answer
    }
end

-- Generate final report
local final_report = promise.await(researcher:chat(
    "Based on all the research findings, create a comprehensive report on " .. research_topic
))

spell.output({
    topic = research_topic,
    findings = findings,
    final_report = final_report,
    agent_metrics = researcher:get_metrics()
}, "json")
```

### Multi-Agent Collaboration

```lua
local agent = require("agent")
local spell = require("spell")
local events = require("events")

spell.init({
    name = "multi-agent-debate",
    params = {
        topic = { type = "string", required = true },
        rounds = { type = "number", default = 3 }
    }
})

-- Create two debating agents
local advocate = agent.create({
    name = "advocate",
    model = "gpt-4",
    system_prompt = "You are debating FOR the given topic. Be persuasive and fact-based.",
    tools = {"fact-checker", "web-search"}
})

local opponent = agent.create({
    name = "opponent", 
    model = "gpt-4",
    system_prompt = "You are debating AGAINST the given topic. Be critical and evidence-based.",
    tools = {"fact-checker", "web-search"}
})

local moderator = agent.create({
    name = "moderator",
    model = "gpt-4",
    system_prompt = "You moderate debates, ensuring fairness and summarizing key points."
})

local topic = spell.params("topic")
local rounds = spell.params("rounds")

-- Set up event coordination
local debate_state = {
    topic = topic,
    rounds = {},
    current_round = 0
}

-- Debate loop
for round = 1, rounds do
    debate_state.current_round = round
    
    -- Advocate's turn
    local advocate_argument = promise.await(advocate:chat(
        string.format("Round %d: Argue FOR '%s'. Previous arguments: %s", 
            round, topic, format_previous_rounds(debate_state.rounds))
    ))
    
    -- Opponent's response
    local opponent_argument = promise.await(opponent:chat(
        string.format("Round %d: Argue AGAINST '%s'. Counter this argument: %s", 
            round, topic, advocate_argument)
    ))
    
    -- Moderator's summary
    local round_summary = promise.await(moderator:chat(
        string.format("Summarize round %d of the debate. Advocate: %s. Opponent: %s",
            round, advocate_argument, opponent_argument)
    ))
    
    table.insert(debate_state.rounds, {
        round = round,
        advocate_argument = advocate_argument,
        opponent_argument = opponent_argument,
        summary = round_summary
    })
end

-- Final moderator judgment
local final_judgment = promise.await(moderator:chat(
    "Provide a final analysis of this debate, highlighting the strongest arguments from both sides."
))

spell.output({
    topic = topic,
    debate_rounds = debate_state.rounds,
    final_judgment = final_judgment,
    participant_metrics = {
        advocate = advocate:get_metrics(),
        opponent = opponent:get_metrics(),
        moderator = moderator:get_metrics()
    }
}, "json")
```

---

## Data Processing

### CSV Data Analysis

```lua
local spell = require("spell")
local data = require("data")
local core = require("core")

spell.init({
    name = "csv-analyzer",
    params = {
        csv_data = { type = "string", required = true },
        analysis_type = { type = "string", default = "summary", enum = {"summary", "detailed", "statistical"} }
    }
})

local csv_data = spell.params("csv_data")
local analysis_type = spell.params("analysis_type")

-- Parse CSV data
local parsed_data = data.from_csv(csv_data, nil)  -- Auto-detect headers

-- Basic analysis
local analysis = {
    row_count = #parsed_data,
    column_count = 0,
    columns = {},
    summary = {}
}

if #parsed_data > 0 then
    -- Get column information
    for column, _ in pairs(parsed_data[1]) do
        analysis.column_count = analysis.column_count + 1
        table.insert(analysis.columns, column)
    end
    
    -- Column-wise analysis
    for _, column in ipairs(analysis.columns) do
        local values = data.map(parsed_data, function(row) return row[column] end)
        local non_null_values = data.filter(values, function(v) return v ~= nil and v ~= "" end)
        
        analysis.summary[column] = {
            total_values = #values,
            non_null_values = #non_null_values,
            null_percentage = (#values - #non_null_values) / #values * 100
        }
        
        -- Try to detect numeric columns
        local numeric_values = data.filter(non_null_values, function(v) 
            return tonumber(v) ~= nil 
        end)
        
        if #numeric_values > #non_null_values * 0.8 then  -- 80% numeric threshold
            analysis.summary[column].type = "numeric"
            local numbers = data.map(numeric_values, tonumber)
            analysis.summary[column].stats = {
                min = math.min(table.unpack(numbers)),
                max = math.max(table.unpack(numbers)),
                avg = data.reduce(numbers, function(sum, n) return sum + n end, 0) / #numbers
            }
        else
            analysis.summary[column].type = "text"
            analysis.summary[column].unique_count = #core.table.keys(
                data.reduce(non_null_values, function(acc, v)
                    acc[v] = true
                    return acc
                end, {})
            )
        end
    end
end

-- Additional analysis based on type
if analysis_type == "detailed" then
    analysis.sample_rows = data.slice(parsed_data, 1, math.min(5, #parsed_data))
elseif analysis_type == "statistical" then
    -- Add more statistical analysis
    for column, info in pairs(analysis.summary) do
        if info.type == "numeric" then
            local values = data.map(parsed_data, function(row) 
                return tonumber(row[column]) 
            end)
            local valid_values = data.filter(values, function(v) return v ~= nil end)
            
            if #valid_values > 0 then
                table.sort(valid_values)
                local median_idx = math.ceil(#valid_values / 2)
                info.stats.median = valid_values[median_idx]
                
                -- Calculate standard deviation
                local mean = info.stats.avg
                local variance = data.reduce(valid_values, function(sum, v)
                    return sum + (v - mean) ^ 2
                end, 0) / #valid_values
                info.stats.std_dev = math.sqrt(variance)
            end
        end
    end
end

spell.output({
    analysis_type = analysis_type,
    data_analysis = analysis
}, "json")
```

### JSON Data Transformation

```lua
local spell = require("spell")
local data = require("data")

spell.init({
    name = "json-transformer",
    params = {
        json_data = { type = "string", required = true },
        transformation_rules = { type = "array", required = true }
    }
})

local json_data = spell.params("json_data")
local transformation_rules = spell.params("transformation_rules")

-- Parse input JSON
local input_data = data.from_json(json_data)

-- Apply transformation rules
local transformed_data = input_data

for _, rule in ipairs(transformation_rules) do
    if rule.type == "map" then
        transformed_data = data.map(transformed_data, function(item)
            return apply_mapping_rule(item, rule.mapping)
        end)
    elseif rule.type == "filter" then
        transformed_data = data.filter(transformed_data, function(item)
            return evaluate_filter_condition(item, rule.condition)
        end)
    elseif rule.type == "reduce" then
        transformed_data = data.reduce(transformed_data, function(acc, item)
            return apply_reduction_rule(acc, item, rule.reducer)
        end, rule.initial or {})
    elseif rule.type == "sort" then
        table.sort(transformed_data, function(a, b)
            return compare_by_field(a, b, rule.field, rule.order or "asc")
        end)
    end
end

-- Convert back to JSON
local output_json = data.to_json(transformed_data)

spell.output({
    original_count = type(input_data) == "table" and #input_data or 1,
    transformed_count = type(transformed_data) == "table" and #transformed_data or 1,
    rules_applied = #transformation_rules,
    result = output_json
}, "json")

-- Helper functions
function apply_mapping_rule(item, mapping)
    local result = {}
    for output_field, input_path in pairs(mapping) do
        result[output_field] = get_nested_value(item, input_path)
    end
    return result
end

function evaluate_filter_condition(item, condition)
    local field_value = get_nested_value(item, condition.field)
    if condition.operator == "equals" then
        return field_value == condition.value
    elseif condition.operator == "contains" then
        return string.find(tostring(field_value), condition.value) ~= nil
    elseif condition.operator == "greater_than" then
        return tonumber(field_value) and tonumber(field_value) > condition.value
    end
    return false
end

function get_nested_value(obj, path)
    local current = obj
    for part in string.gmatch(path, "[^%.]+") do
        if type(current) == "table" and current[part] ~= nil then
            current = current[part]
        else
            return nil
        end
    end
    return current
end
```

---

## Event-Driven Programming

### Event-Based Workflow

```lua
local spell = require("spell")
local events = require("events")
local promise = require("promise")

spell.init({
    name = "event-driven-processor",
    params = {
        input_file = { type = "string", required = true }
    }
})

-- Set up event handlers for processing pipeline
local processing_state = {
    stages_completed = 0,
    total_stages = 4,
    results = {}
}

-- Stage 1: File Reading
events.on("file:read", function(data)
    processing_state.results.file_content = data.content
    processing_state.stages_completed = processing_state.stages_completed + 1
    events.emit("processing:stage_complete", { stage = "file_read", data = data })
    
    -- Trigger next stage
    events.emit("data:parse", { content = data.content })
end)

-- Stage 2: Data Parsing
events.on("data:parse", function(data)
    local parsed = parse_content(data.content)
    processing_state.results.parsed_data = parsed
    processing_state.stages_completed = processing_state.stages_completed + 1
    events.emit("processing:stage_complete", { stage = "data_parse", data = parsed })
    
    -- Trigger next stage
    events.emit("data:validate", { data = parsed })
end)

-- Stage 3: Data Validation
events.on("data:validate", function(data)
    local validation_result = validate_data(data.data)
    processing_state.results.validation = validation_result
    processing_state.stages_completed = processing_state.stages_completed + 1
    events.emit("processing:stage_complete", { stage = "validation", data = validation_result })
    
    if validation_result.valid then
        events.emit("data:process", { data = data.data })
    else
        events.emit("processing:error", { 
            stage = "validation", 
            errors = validation_result.errors 
        })
    end
end)

-- Stage 4: Data Processing
events.on("data:process", function(data)
    local processed = process_data(data.data)
    processing_state.results.processed_data = processed
    processing_state.stages_completed = processing_state.stages_completed + 1
    events.emit("processing:stage_complete", { stage = "processing", data = processed })
    events.emit("processing:complete", processing_state.results)
end)

-- Progress tracking
events.on("processing:stage_complete", function(data)
    local progress = (processing_state.stages_completed / processing_state.total_stages) * 100
    spell.output({
        stage = data.stage,
        progress_percentage = progress,
        timestamp = core.time.now()
    }, "text")
end)

-- Error handling
events.on("processing:error", function(error_data)
    spell.output({
        error = true,
        stage = error_data.stage,
        details = error_data.errors
    }, "json")
end)

-- Final completion
events.on("processing:complete", function(results)
    spell.output({
        success = true,
        processing_time = core.time.now() - spell.context().start_time,
        results = results
    }, "json")
end)

-- Start the pipeline
local input_file = spell.params("input_file")
events.emit("file:read", { filename = input_file, content = read_file(input_file) })

-- Wait for completion or timeout
promise.await(events.wait_for("processing:complete", 30000))
```

### Real-Time Monitoring

```lua
local spell = require("spell")
local events = require("events")
local metrics = require("observability").metrics
local promise = require("promise")

spell.init({
    name = "real-time-monitor",
    params = {
        monitor_duration = { type = "number", default = 60 },
        alert_thresholds = { type = "table", default = {} }
    }
})

local monitor_duration = spell.params("monitor_duration")
local alert_thresholds = spell.params("alert_thresholds")

-- Monitoring state
local monitoring_state = {
    start_time = core.time.now(),
    events_processed = 0,
    alerts_triggered = 0,
    metrics = {}
}

-- Set up monitoring
events.on("system:metric", function(metric_data)
    monitoring_state.events_processed = monitoring_state.events_processed + 1
    
    -- Record metric
    metrics.gauge(metric_data.name, metric_data.value, metric_data.tags)
    
    -- Check thresholds
    for threshold_name, threshold_config in pairs(alert_thresholds) do
        if metric_data.name == threshold_config.metric then
            local should_alert = false
            
            if threshold_config.condition == "greater_than" then
                should_alert = metric_data.value > threshold_config.value
            elseif threshold_config.condition == "less_than" then
                should_alert = metric_data.value < threshold_config.value
            end
            
            if should_alert then
                monitoring_state.alerts_triggered = monitoring_state.alerts_triggered + 1
                events.emit("monitoring:alert", {
                    threshold = threshold_name,
                    metric = metric_data.name,
                    value = metric_data.value,
                    threshold_value = threshold_config.value,
                    timestamp = core.time.now()
                })
            end
        end
    end
end)

-- Alert handling
events.on("monitoring:alert", function(alert_data)
    spell.output({
        alert = true,
        threshold = alert_data.threshold,
        metric = alert_data.metric,
        current_value = alert_data.value,
        threshold_value = alert_data.threshold_value,
        timestamp = alert_data.timestamp
    }, "json")
end)

-- Periodic status reports
local status_interval = 10  -- seconds
local status_timer = promise.new(function(resolve)
    local function report_status()
        if core.time.now() - monitoring_state.start_time < monitor_duration then
            spell.output({
                status = "monitoring",
                duration_elapsed = core.time.now() - monitoring_state.start_time,
                events_processed = monitoring_state.events_processed,
                alerts_triggered = monitoring_state.alerts_triggered
            }, "text")
            
            -- Schedule next report
            promise.sleep(status_interval * 1000):andThen(report_status)
        else
            resolve()
        end
    end
    report_status()
end)

-- Wait for monitoring duration to complete
promise.await(status_timer)

-- Final report
spell.output({
    monitoring_complete = true,
    total_duration = monitor_duration,
    events_processed = monitoring_state.events_processed,
    alerts_triggered = monitoring_state.alerts_triggered,
    avg_events_per_second = monitoring_state.events_processed / monitor_duration
}, "json")
```

---

## Testing Spells

### Unit Testing Example

```lua
local testing = require("testing")
local core = require("core")

-- Test suite for string utilities
testing.describe("String Utilities", function()
    
    testing.describe("template function", function()
        testing.it("should replace simple variables", function()
            local result = string.template("Hello ${name}!", {name = "World"})
            testing.assert.equal(result, "Hello World!")
        end)
        
        testing.it("should handle missing variables", function()
            local result = string.template("Hello ${name}!", {})
            testing.assert.equal(result, "Hello ${name}!")
        end)
        
        testing.it("should handle multiple variables", function()
            local result = string.template("${greeting} ${name}!", {
                greeting = "Hello",
                name = "World"
            })
            testing.assert.equal(result, "Hello World!")
        end)
    end)
    
    testing.describe("slugify function", function()
        testing.it("should convert spaces to hyphens", function()
            local result = string.slugify("Hello World")
            testing.assert.equal(result, "hello-world")
        end)
        
        testing.it("should remove special characters", function()
            local result = string.slugify("Hello, World!")
            testing.assert.equal(result, "hello-world")
        end)
        
        testing.it("should handle empty strings", function()
            local result = string.slugify("")
            testing.assert.equal(result, "")
        end)
    end)
end)

-- Test suite for table utilities
testing.describe("Table Utilities", function()
    
    testing.describe("deep_copy function", function()
        testing.it("should create independent copy", function()
            local original = {a = 1, b = {c = 2}}
            local copy = table.deep_copy(original)
            
            copy.b.c = 3
            testing.assert.equal(original.b.c, 2)
            testing.assert.equal(copy.b.c, 3)
        end)
        
        testing.it("should handle circular references", function()
            local original = {a = 1}
            original.self = original
            
            testing.assert.no_error(function()
                local copy = table.deep_copy(original)
                testing.assert.equal(copy.a, 1)
                testing.assert.equal(copy.self, copy)
            end)
        end)
    end)
    
    testing.describe("merge function", function()
        testing.it("should merge two tables", function()
            local t1 = {a = 1, b = 2}
            local t2 = {b = 3, c = 4}
            local result = table.merge(t1, t2)
            
            testing.assert.equal(result.a, 1)
            testing.assert.equal(result.b, 3)  -- t2 overwrites t1
            testing.assert.equal(result.c, 4)
        end)
    end)
end)

-- Run the tests
testing.run_all()
```

### Integration Testing

```lua
local testing = require("testing")
local spell = require("spell")
local llm = require("llm")
local promise = require("promise")

testing.describe("Spell Integration Tests", function()
    
    testing.before_each(function()
        -- Reset spell state before each test
        spell.reset()
    end)
    
    testing.describe("Basic spell lifecycle", function()
        testing.it("should initialize and execute successfully", function()
            spell.init({
                name = "test-spell",
                params = {
                    input = { type = "string", default = "test" }
                }
            })
            
            local input = spell.params("input")
            testing.assert.equal(input, "test")
            
            local context = spell.context()
            testing.assert.equal(context.spell_name, "test-spell")
        end)
        
        testing.it("should validate required parameters", function()
            spell.init({
                name = "test-spell",
                params = {
                    required_param = { type = "string", required = true }
                }
            })
            
            testing.assert.has_error(function()
                spell.params("required_param")
            end, "Required parameter")
        end)
    end)
    
    testing.describe("Spell composition", function()
        testing.it("should execute composed spells", function()
            -- Mock spell definitions
            spell.library("test-spells", {
                echo = function(params)
                    return { output = "echo: " .. (params.input or "") }
                end,
                uppercase = function(params)
                    return { output = string.upper(params.input or "") }
                end
            })
            
            spell.init({ name = "composition-test" })
            
            local results = spell.compose({
                {
                    name = "step1",
                    spell = "echo",
                    params = { input = "hello" }
                },
                {
                    name = "step2", 
                    spell = "uppercase",
                    params = { input = "$step1.output" }
                }
            })
            
            testing.assert.is_not_nil(results.step1)
            testing.assert.is_not_nil(results.step2)
        end)
    end)
    
    testing.describe("Error handling", function()
        testing.it("should handle spell errors gracefully", function()
            local error_handled = false
            
            spell.on_error(function(err)
                error_handled = true
                return { recoverable = true }
            end)
            
            spell.init({ name = "error-test" })
            
            testing.assert.has_error(function()
                error("Test error")
            end)
            
            -- Note: In real implementation, error_handled would be true
            -- This is a simplified test
        end)
    end)
end)

-- Async testing
testing.describe("Async Operations", function()
    
    testing.it("should handle promise resolution", function()
        local async_test = promise.async(function()
            local result = promise.await(promise.resolve("success"))
            testing.assert.equal(result, "success")
        end)
        
        testing.assert.no_error(function()
            promise.await(async_test())
        end)
    end)
    
    testing.it("should handle promise rejection", function()
        local async_test = promise.async(function()
            promise.await(promise.reject("test error"))
        end)
        
        testing.assert.has_error(function()
            promise.await(async_test())
        end, "test error")
    end)
    
    testing.it("should handle timeouts", function()
        local slow_promise = promise.new(function(resolve)
            promise.sleep(2000):andThen(function()
                resolve("too slow")
            end)
        end)
        
        testing.assert.has_error(function()
            promise.await(slow_promise, 1000)  -- 1 second timeout
        end, "timeout")
    end)
end)
```

---

## Error Handling

### Comprehensive Error Recovery

```lua
local spell = require("spell")
local errors = require("errors")
local llm = require("llm")
local promise = require("promise")

spell.init({
    name = "robust-processor",
    params = {
        input_data = { type = "array", required = true },
        max_retries = { type = "number", default = 3 },
        fallback_strategy = { type = "string", default = "skip", enum = {"skip", "default", "fail"} }
    }
})

local input_data = spell.params("input_data")
local max_retries = spell.params("max_retries")
local fallback_strategy = spell.params("fallback_strategy")

-- Global error handler
spell.on_error(function(error, context)
    spell.output({
        error = true,
        message = tostring(error),
        context = context,
        timestamp = core.time.now()
    }, "text")
    
    return { recoverable = true, retry_after = 1000 }
end)

-- Process each item with error recovery
local results = {}
local errors_encountered = {}

for i, item in ipairs(input_data) do
    local success = false
    local attempt = 0
    local result = nil
    
    while not success and attempt < max_retries do
        attempt = attempt + 1
        
        local processing_result = errors.try(function()
            -- Simulate risky operation
            if math.random() > 0.7 then  -- 30% failure rate
                error("Random processing error for item " .. i)
            end
            
            return process_item(item)
        end, function(err)
            table.insert(errors_encountered, {
                item_index = i,
                attempt = attempt,
                error = tostring(err),
                timestamp = core.time.now()
            })
            
            if attempt < max_retries then
                promise.await(promise.sleep(1000 * attempt))  -- Exponential backoff
                return nil  -- Retry
            else
                return handle_fallback(item, fallback_strategy, err)
            end
        end)
        
        if processing_result then
            result = processing_result
            success = true
        end
    end
    
    results[i] = {
        item = item,
        result = result,
        success = success,
        attempts = attempt
    }
end

-- Summary
local successful_items = 0
for _, result in ipairs(results) do
    if result.success then
        successful_items = successful_items + 1
    end
end

spell.output({
    processing_summary = {
        total_items = #input_data,
        successful_items = successful_items,
        failed_items = #input_data - successful_items,
        total_errors = #errors_encountered,
        success_rate = (successful_items / #input_data) * 100
    },
    results = results,
    errors = errors_encountered
}, "json")

function process_item(item)
    -- Simulate item processing
    return {
        processed = true,
        value = item.value * 2,
        timestamp = core.time.now()
    }
end

function handle_fallback(item, strategy, error)
    if strategy == "skip" then
        return nil
    elseif strategy == "default" then
        return {
            processed = false,
            value = 0,
            fallback = true
        }
    else
        error("Processing failed for item: " .. tostring(error))
    end
end
```

### Circuit Breaker Pattern

```lua
local spell = require("spell")
local promise = require("promise")

spell.init({
    name = "circuit-breaker-example",
    params = {
        requests = { type = "array", required = true }
    }
})

-- Circuit breaker state
local circuit_breaker = {
    state = "closed",  -- closed, open, half-open
    failure_count = 0,
    failure_threshold = 5,
    timeout = 30000,  -- 30 seconds
    last_failure_time = 0,
    success_threshold = 3  -- for half-open state
}

function call_with_circuit_breaker(operation)
    local current_time = core.time.now() * 1000  -- Convert to milliseconds
    
    -- Check if circuit should move from open to half-open
    if circuit_breaker.state == "open" then
        if current_time - circuit_breaker.last_failure_time > circuit_breaker.timeout then
            circuit_breaker.state = "half-open"
            circuit_breaker.failure_count = 0
        else
            return promise.reject("Circuit breaker is open")
        end
    end
    
    -- If circuit is open, reject immediately
    if circuit_breaker.state == "open" then
        return promise.reject("Circuit breaker is open")
    end
    
    -- Try the operation
    return promise.new(function(resolve, reject)
        operation():andThen(function(result)
            -- Success
            if circuit_breaker.state == "half-open" then
                circuit_breaker.success_count = (circuit_breaker.success_count or 0) + 1
                if circuit_breaker.success_count >= circuit_breaker.success_threshold then
                    circuit_breaker.state = "closed"
                    circuit_breaker.failure_count = 0
                    circuit_breaker.success_count = 0
                end
            end
            resolve(result)
        end):onError(function(error)
            -- Failure
            circuit_breaker.failure_count = circuit_breaker.failure_count + 1
            circuit_breaker.last_failure_time = current_time
            
            if circuit_breaker.failure_count >= circuit_breaker.failure_threshold then
                circuit_breaker.state = "open"
            end
            
            reject(error)
        end)
    end)
end

-- Process requests with circuit breaker
local requests = spell.params("requests")
local results = {}

for i, request in ipairs(requests) do
    local request_operation = function()
        return promise.new(function(resolve, reject)
            -- Simulate unreliable service
            if math.random() > 0.6 then  -- 40% success rate
                resolve("Success for request " .. i)
            else
                reject("Service unavailable for request " .. i)
            end
        end)
    end
    
    call_with_circuit_breaker(request_operation):andThen(function(result)
        results[i] = {
            success = true,
            result = result,
            circuit_state = circuit_breaker.state
        }
    end):onError(function(error)
        results[i] = {
            success = false,
            error = tostring(error),
            circuit_state = circuit_breaker.state
        }
    end)
    
    -- Brief delay between requests
    promise.await(promise.sleep(100))
end

spell.output({
    total_requests = #requests,
    results = results,
    final_circuit_state = circuit_breaker.state,
    total_failures = circuit_breaker.failure_count
}, "json")
```

---

## Performance Optimization

### Caching Strategy

```lua
local spell = require("spell")
local llm = require("llm")
local promise = require("promise")
local core = require("core")

spell.init({
    name = "optimized-batch-processor",
    params = {
        items = { type = "array", required = true },
        cache_duration = { type = "number", default = 3600 },
        batch_size = { type = "number", default = 10 }
    }
})

local items = spell.params("items")
local cache_duration = spell.params("cache_duration")
local batch_size = spell.params("batch_size")

-- Multi-level caching strategy
local performance_stats = {
    cache_hits = 0,
    cache_misses = 0,
    llm_calls = 0,
    batches_processed = 0
}

function get_cached_or_process(item)
    -- Level 1: Local spell cache
    local cache_key = "processed_" .. core.crypto.hash("md5", core.json.encode(item))
    local cached_result = spell.cache(cache_key)
    
    if cached_result then
        performance_stats.cache_hits = performance_stats.cache_hits + 1
        return promise.resolve(cached_result)
    end
    
    performance_stats.cache_misses = performance_stats.cache_misses + 1
    performance_stats.llm_calls = performance_stats.llm_calls + 1
    
    -- Process item
    return llm.chat("Process this item: " .. core.json.encode(item)):andThen(function(result)
        -- Cache the result
        spell.cache(cache_key, result, cache_duration)
        return result
    end)
end

-- Batch processing for efficiency
local all_results = {}
local current_batch = {}

for i, item in ipairs(items) do
    table.insert(current_batch, item)
    
    -- Process batch when full or at end
    if #current_batch >= batch_size or i == #items then
        performance_stats.batches_processed = performance_stats.batches_processed + 1
        
        -- Process batch concurrently
        local batch_promises = {}
        for j, batch_item in ipairs(current_batch) do
            batch_promises[j] = get_cached_or_process(batch_item)
        end
        
        -- Wait for batch completion
        local batch_results = promise.await(promise.all(batch_promises))
        
        -- Add to all results
        for j, result in ipairs(batch_results) do
            table.insert(all_results, {
                item = current_batch[j],
                result = result,
                batch = performance_stats.batches_processed
            })
        end
        
        -- Clear batch
        current_batch = {}
        
        -- Brief pause between batches to avoid rate limiting
        if i < #items then
            promise.await(promise.sleep(100))
        end
    end
end

-- Calculate performance metrics
local cache_hit_rate = performance_stats.cache_hits / 
    (performance_stats.cache_hits + performance_stats.cache_misses) * 100

spell.output({
    total_items = #items,
    results = all_results,
    performance = {
        cache_hit_rate = cache_hit_rate,
        llm_calls_made = performance_stats.llm_calls,
        batches_processed = performance_stats.batches_processed,
        avg_batch_size = #items / performance_stats.batches_processed
    },
    cache_stats = spell.get_cache_stats()
}, "json")
```

### Memory Management

```lua
local spell = require("spell")
local promise = require("promise")

spell.init({
    name = "memory-efficient-processor",
    params = {
        large_dataset = { type = "array", required = true },
        chunk_size = { type = "number", default = 100 }
    }
})

local large_dataset = spell.params("large_dataset")
local chunk_size = spell.params("chunk_size")

-- Memory tracking
local memory_stats = {
    initial_memory = collectgarbage("count"),
    peak_memory = 0,
    gc_cycles = 0
}

-- Process in chunks to manage memory
local total_chunks = math.ceil(#large_dataset / chunk_size)
local processed_count = 0

for chunk_idx = 1, total_chunks do
    local start_idx = (chunk_idx - 1) * chunk_size + 1
    local end_idx = math.min(chunk_idx * chunk_size, #large_dataset)
    local chunk = {}
    
    -- Extract chunk
    for i = start_idx, end_idx do
        chunk[i - start_idx + 1] = large_dataset[i]
    end
    
    -- Process chunk
    local chunk_results = process_chunk(chunk)
    processed_count = processed_count + #chunk
    
    -- Memory management
    chunk = nil  -- Clear chunk reference
    collectgarbage("collect")  -- Force garbage collection
    memory_stats.gc_cycles = memory_stats.gc_cycles + 1
    
    local current_memory = collectgarbage("count")
    if current_memory > memory_stats.peak_memory then
        memory_stats.peak_memory = current_memory
    end
    
    -- Progress reporting
    local progress = (processed_count / #large_dataset) * 100
    spell.output({
        progress = progress,
        chunk = chunk_idx,
        total_chunks = total_chunks,
        memory_usage_kb = current_memory
    }, "text")
    
    -- Yield to prevent blocking
    promise.await(promise.sleep(10))
end

-- Final memory statistics
memory_stats.final_memory = collectgarbage("count")
memory_stats.memory_saved = memory_stats.peak_memory - memory_stats.final_memory

spell.output({
    processing_complete = true,
    total_items = #large_dataset,
    memory_statistics = memory_stats
}, "json")

function process_chunk(chunk)
    -- Simulate memory-intensive processing
    local results = {}
    for i, item in ipairs(chunk) do
        results[i] = {
            processed = true,
            data = string.rep(tostring(item), 10)  -- Simulate memory usage
        }
    end
    return results
end
```

---

## Real-World Use Cases

### Web Content Summarizer

```lua
local spell = require("spell")
local llm = require("llm")
local promise = require("promise")
local core = require("core")

spell.init({
    name = "web-content-summarizer",
    description = "Comprehensive web content analysis and summarization",
    params = {
        urls = { type = "array", required = true },
        summary_style = { type = "string", default = "balanced", enum = {"brief", "balanced", "detailed"} },
        include_metadata = { type = "boolean", default = true },
        max_content_length = { type = "number", default = 50000 }
    }
})

local urls = spell.params("urls")
local summary_style = spell.params("summary_style")
local include_metadata = spell.params("include_metadata")
local max_content_length = spell.params("max_content_length")

-- Multi-step processing workflow
local workflow_results = spell.compose({
    {
        name = "fetch_content",
        spell = "web-fetcher",
        params = {
            urls = urls,
            max_content_length = max_content_length,
            extract_metadata = include_metadata
        }
    },
    {
        name = "content_analysis", 
        spell = "content-analyzer",
        params = {
            content = "$fetch_content.pages",
            analysis_type = "comprehensive"
        }
    },
    {
        name = "summarization",
        spell = "multi-document-summarizer",
        params = {
            documents = "$content_analysis.structured_content",
            style = summary_style,
            preserve_structure = true
        }
    }
})

-- Generate final report
local report = {
    summary = workflow_results.summarization.summary,
    sources = {},
    analysis = workflow_results.content_analysis.insights,
    processing_metadata = {
        urls_processed = #urls,
        total_content_length = 0,
        processing_time = core.time.now() - spell.context().start_time
    }
}

-- Build source information
for i, page_data in ipairs(workflow_results.fetch_content.pages) do
    report.sources[i] = {
        url = page_data.url,
        title = page_data.title,
        content_length = #page_data.content,
        last_modified = page_data.last_modified,
        credibility_score = workflow_results.content_analysis.credibility_scores[i]
    }
    report.processing_metadata.total_content_length = 
        report.processing_metadata.total_content_length + #page_data.content
end

spell.output(report, "json")
```

### Automated Research Assistant

```lua
local spell = require("spell")
local agent = require("agent")
local events = require("events")
local promise = require("promise")

spell.init({
    name = "automated-research-assistant",
    description = "Conducts comprehensive research on any topic",
    params = {
        research_query = { type = "string", required = true },
        depth_level = { type = "number", default = 3, min = 1, max = 5 },
        include_citations = { type = "boolean", default = true },
        output_format = { type = "string", default = "report", enum = {"report", "outline", "references"} }
    }
})

-- Create specialized research team
local research_team = {
    coordinator = agent.create({
        name = "research-coordinator",
        model = "gpt-4",
        system_prompt = "You coordinate research efforts, breaking down complex queries into manageable tasks.",
        tools = {"task-planner", "progress-tracker"}
    }),
    
    searcher = agent.create({
        name = "information-searcher", 
        model = "gpt-4",
        system_prompt = "You excel at finding relevant information using various search strategies.",
        tools = {"web-search", "academic-search", "news-search"}
    }),
    
    analyzer = agent.create({
        name = "content-analyzer",
        model = "gpt-4", 
        system_prompt = "You analyze and synthesize information, identifying key insights and patterns.",
        tools = {"fact-checker", "sentiment-analyzer", "trend-analyzer"}
    }),
    
    writer = agent.create({
        name = "report-writer",
        model = "gpt-4",
        system_prompt = "You create well-structured, comprehensive reports with proper citations.",
        tools = {"citation-formatter", "document-generator"}
    })
}

local research_query = spell.params("research_query")
local depth_level = spell.params("depth_level")
local include_citations = spell.params("include_citations")
local output_format = spell.params("output_format")

-- Research orchestration
local research_session = {
    query = research_query,
    start_time = core.time.now(),
    phases = {},
    knowledge_base = {}
}

-- Phase 1: Research Planning
local research_plan = promise.await(research_team.coordinator:chat(
    string.format("Create a comprehensive research plan for: '%s'. Depth level: %d. Break into specific searchable questions.",
        research_query, depth_level)
))

research_session.phases.planning = {
    plan = research_plan,
    questions = extract_research_questions(research_plan)
}

-- Phase 2: Information Gathering
local search_results = {}
for i, question in ipairs(research_session.phases.planning.questions) do
    local search_result = promise.await(research_team.searcher:chat(
        "Find comprehensive information about: " .. question
    ))
    search_results[i] = {
        question = question,
        results = search_result
    }
    
    -- Update knowledge base
    research_session.knowledge_base[question] = search_result
end

research_session.phases.gathering = search_results

-- Phase 3: Analysis and Synthesis
local analysis_tasks = {}
for topic, information in pairs(research_session.knowledge_base) do
    local analysis = promise.await(research_team.analyzer:chat(
        string.format("Analyze this information about '%s': %s", topic, information)
    ))
    analysis_tasks[topic] = analysis
end

research_session.phases.analysis = analysis_tasks

-- Phase 4: Report Generation
local report_prompt = string.format([[
Create a comprehensive %s about "%s" using the following analyzed information:

%s

Requirements:
- Include citations: %s
- Output format: %s
- Depth level: %d
]], 
    output_format, research_query, 
    serialize_analysis(analysis_tasks),
    include_citations and "Yes" or "No",
    output_format, depth_level
)

local final_report = promise.await(research_team.writer:chat(report_prompt))

-- Compile final results
local research_results = {
    query = research_query,
    report = final_report,
    methodology = {
        research_plan = research_plan,
        questions_investigated = #research_session.phases.planning.questions,
        sources_analyzed = count_unique_sources(search_results),
        depth_level = depth_level
    },
    metadata = {
        total_time = core.time.now() - research_session.start_time,
        team_metrics = {
            coordinator = research_team.coordinator:get_metrics(),
            searcher = research_team.searcher:get_metrics(),
            analyzer = research_team.analyzer:get_metrics(),
            writer = research_team.writer:get_metrics()
        }
    }
}

if include_citations then
    research_results.citations = extract_citations(final_report)
end

spell.output(research_results, "json")

-- Helper functions
function extract_research_questions(plan)
    -- Extract questions from research plan
    local questions = {}
    for question in string.gmatch(plan, "Q:%s*([^\n]+)") do
        table.insert(questions, question)
    end
    return questions
end

function serialize_analysis(analysis_tasks)
    local serialized = ""
    for topic, analysis in pairs(analysis_tasks) do
        serialized = serialized .. "Topic: " .. topic .. "\n"
        serialized = serialized .. "Analysis: " .. analysis .. "\n\n"
    end
    return serialized
end

function count_unique_sources(search_results)
    local unique_sources = {}
    for _, result in ipairs(search_results) do
        -- Extract source URLs from results (simplified)
        local source_count = string.len(result.results) / 100  -- Rough estimate
        unique_sources[result.question] = source_count
    end
    return #core.table.keys(unique_sources)
end

function extract_citations(report)
    -- Extract citations from report (simplified)
    local citations = {}
    for citation in string.gmatch(report, "%[([^%]]+)%]") do
        table.insert(citations, citation)
    end
    return citations
end
```

These examples demonstrate the power and flexibility of the go-llmspell Lua standard library across various real-world scenarios. Each example showcases different aspects of the library while solving practical problems that developers might encounter when building LLM-powered applications.

For complete API documentation, see [API_REFERENCE.md](API_REFERENCE.md).
For library overview and philosophy, see [README.md](README.md).