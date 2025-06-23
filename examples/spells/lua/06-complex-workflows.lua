-- ABOUTME: Example demonstrating complex multi-step workflows with conditions and branching
-- ABOUTME: Shows state management, parallel execution, error handling, and workflow patterns

-- Required modules
local agent = require("agent")
local core = require("core")

-- Complex Workflows Example
-- This spell demonstrates advanced workflow patterns:
-- 1. Sequential workflows with state passing
-- 2. Conditional branching
-- 3. Parallel execution
-- 4. Error handling and rollback
-- 5. Workflow composition
-- 6. State management across steps

-- Parameters
local workflow_type = params.workflow_type or "content_pipeline"
local topic = params.topic or "The Future of Renewable Energy"
local model = params.model or "gpt-4"
local output_dir = params.output_dir or "./workflow-output"

print("=== Complex Workflows Example ===")
print("Workflow type: " .. workflow_type)
print("Topic: " .. topic)
print()

-- Ensure output directory exists
if not tools.file_exists(output_dir) then
    tools.create_directory(output_dir)
end

-- Table copy helper (defined before use)
local function table_copy(t)
    local copy = {}
    for k, v in pairs(t) do
        if type(v) == "table" then
            copy[k] = table_copy(v)
        else
            copy[k] = v
        end
    end
    return copy
end

-- Workflow state management
local WorkflowState = {}

function WorkflowState:new()
    local obj = {
        data = {},
        history = {},
        errors = {},
        checkpoints = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function WorkflowState:set(key, value)
    self.data[key] = value
    table.insert(self.history, {
        timestamp = os.time(),
        action = "set",
        key = key,
        value = value
    })
end

function WorkflowState:get(key)
    return self.data[key]
end

function WorkflowState:checkpoint(name)
    self.checkpoints[name] = {
        data = table_copy(self.data),
        timestamp = os.time()
    }
    print("  ✓ Checkpoint: " .. name)
end

function WorkflowState:rollback(checkpoint_name)
    local checkpoint = self.checkpoints[checkpoint_name]
    if checkpoint then
        self.data = table_copy(checkpoint.data)
        print("  ⏪ Rolled back to: " .. checkpoint_name)
        return true
    end
    return false
end

function WorkflowState:add_error(step, error)
    table.insert(self.errors, {
        step = step,
        error = error,
        timestamp = os.time()
    })
end

-- Example 1: Sequential Workflow with State
print("=== Example 1: Sequential Content Pipeline ===")

local function content_pipeline(topic)
    local state = WorkflowState:new()
    state:set("topic", topic)
    state:set("start_time", os.time())
    
    -- Step 1: Research
    print("Step 1: Research Phase")
    local researcher = agent.create({
        name = "Researcher",
        model = model,
        system = "You are a thorough researcher. Find key facts and recent developments.",
        tools = {"web_search"},
        temperature = 0.3
    })
    
    local research = researcher:run("Research: " .. topic .. ". Find 5 key facts.")
    state:set("research", research)
    state:checkpoint("after_research")
    tools.file_write(output_dir .. "/1_research.txt", research)
    print("  ✓ Research completed")
    
    -- Step 2: Outline Creation
    print("\nStep 2: Outline Creation")
    local outliner = agent.create({
        name = "Outliner",
        model = model,
        system = "You create structured outlines for articles.",
        temperature = 0.4
    })
    
    local outline = outliner:run(
        "Create a detailed outline for an article about: " .. topic .. 
        "\nBased on this research: " .. research
    )
    state:set("outline", outline)
    state:checkpoint("after_outline")
    tools.file_write(output_dir .. "/2_outline.txt", outline)
    print("  ✓ Outline created")
    
    -- Step 3: Content Writing
    print("\nStep 3: Content Writing")
    local writer = agent.create({
        name = "Writer",
        model = model,
        system = "You are a professional content writer. Write engaging, informative content.",
        temperature = 0.6
    })
    
    local content = writer:run(
        "Write an article following this outline:\n" .. outline ..
        "\nMake it approximately 500 words."
    )
    state:set("content", content)
    state:checkpoint("after_writing")
    tools.file_write(output_dir .. "/3_content.md", content)
    print("  ✓ Content written")
    
    -- Step 4: Editing
    print("\nStep 4: Editing")
    local editor = agent.create({
        name = "Editor",
        model = model,
        system = "You are a professional editor. Improve clarity, fix errors, enhance flow.",
        temperature = 0.3
    })
    
    local edited = editor:run("Edit and improve this article:\n" .. content)
    state:set("edited_content", edited)
    state:checkpoint("after_editing")
    tools.file_write(output_dir .. "/4_edited.md", edited)
    print("  ✓ Content edited")
    
    -- Step 5: Final Review
    print("\nStep 5: Final Review")
    local reviewer = agent.create({
        name = "Reviewer",
        model = model,
        system = "You are a quality reviewer. Check for accuracy, completeness, and quality.",
        temperature = 0.2
    })
    
    local review = reviewer:run(
        "Review this article and provide a quality score (1-10) and feedback:\n" .. edited
    )
    state:set("review", review)
    tools.file_write(output_dir .. "/5_review.txt", review)
    print("  ✓ Review completed")
    
    -- Calculate duration
    local duration = os.time() - state:get("start_time")
    state:set("duration", duration)
    
    print("\nWorkflow completed in " .. duration .. " seconds")
    return state
end

-- Run the sequential workflow
local pipeline_state = content_pipeline(topic)
print()

-- Example 2: Conditional Workflow
print("=== Example 2: Conditional Workflow ===")

local function quality_control_workflow(content, min_quality_score)
    local state = WorkflowState:new()
    state:set("content", content)
    state:set("iterations", 0)
    
    local max_iterations = 3
    local quality_score = 0
    
    while quality_score < min_quality_score and state:get("iterations") < max_iterations do
        local iteration = state:get("iterations") + 1
        state:set("iterations", iteration)
        print("\nIteration " .. iteration)
        
        -- Quality check
        print("  Checking quality...")
        local checker = agent.create({
            name = "Quality Checker",
            model = model,
            system = "You evaluate content quality. Rate from 1-10 and explain.",
            temperature = 0.2
        })
        
        local check_result = checker:run(
            "Rate this content from 1-10 and explain why:\n" .. content
        )
        
        -- Extract score (simplified - in production use better parsing)
        quality_score = tonumber(check_result:match("(%d+)/10") or check_result:match("(%d+)")) or 0
        state:set("quality_score_" .. iteration, quality_score)
        print("  Quality score: " .. quality_score)
        
        if quality_score < min_quality_score then
            -- Improve content
            print("  Score too low, improving content...")
            local improver = agent.create({
                name = "Content Improver",
                model = model,
                system = "You improve content based on feedback.",
                temperature = 0.5
            })
            
            content = improver:run(
                "Improve this content based on the feedback:\n" ..
                "Feedback: " .. check_result .. "\n" ..
                "Content: " .. content
            )
            state:set("content", content)
            state:checkpoint("iteration_" .. iteration)
        end
    end
    
    -- Final decision
    if quality_score >= min_quality_score then
        state:set("status", "approved")
        print("\n✅ Content approved with score: " .. quality_score)
    else
        state:set("status", "rejected")
        print("\n❌ Content rejected after " .. max_iterations .. " iterations")
    end
    
    return state
end

-- Run conditional workflow
local test_content = "Renewable energy is good. Solar panels and wind turbines are popular. They help the environment."
local qc_state = quality_control_workflow(test_content, 7)
print()

-- Example 3: Parallel Execution Workflow
print("=== Example 3: Parallel Execution ===")

local function parallel_analysis_workflow(topic)
    local state = WorkflowState:new()
    state:set("topic", topic)
    
    print("Starting parallel analysis of: " .. topic)
    
    -- Define parallel tasks
    local analysts = {
        {
            name = "Technical Analyst",
            system = "Analyze technical aspects and innovations.",
            output_key = "technical_analysis"
        },
        {
            name = "Market Analyst", 
            system = "Analyze market trends and opportunities.",
            output_key = "market_analysis"
        },
        {
            name = "Environmental Analyst",
            system = "Analyze environmental impact and benefits.",
            output_key = "environmental_analysis"
        },
        {
            name = "Economic Analyst",
            system = "Analyze economic implications and costs.",
            output_key = "economic_analysis"
        }
    }
    
    -- Execute analyses in parallel
    local promises = {}
    
    for _, analyst_config in ipairs(analysts) do
        local p = promise.new(function(resolve, reject)
            core.async(function()
                print("  🔄 Starting: " .. analyst_config.name)
                
                local analyst = agent.create({
                    name = analyst_config.name,
                    model = model,
                    system = analyst_config.system,
                    temperature = 0.4
                })
                
                local result = analyst:run("Analyze " .. topic .. " from your perspective. Be concise.")
                
                print("  ✓ Completed: " .. analyst_config.name)
                resolve({
                    key = analyst_config.output_key,
                    value = result
                })
            end)
        end)
        
        table.insert(promises, p)
    end
    
    -- Wait for all to complete
    print("\n  ⏳ Waiting for all analysts to complete...")
    local results = promise.all(promises):await()
    
    -- Store results
    for _, result in ipairs(results) do
        state:set(result.key, result.value)
    end
    
    -- Synthesize results
    print("\n  📊 Synthesizing results...")
    local synthesizer = agent.create({
        name = "Synthesizer",
        model = model,
        system = "You synthesize multiple analyses into a coherent summary.",
        temperature = 0.3
    })
    
    local synthesis_input = ""
    for _, analyst in ipairs(analysts) do
        synthesis_input = synthesis_input .. analyst.name .. ":\n" .. 
                         state:get(analyst.output_key) .. "\n\n"
    end
    
    local synthesis = synthesizer:run(
        "Synthesize these analyses into a comprehensive summary:\n" .. synthesis_input
    )
    state:set("synthesis", synthesis)
    tools.file_write(output_dir .. "/parallel_synthesis.md", synthesis)
    
    print("  ✓ Parallel workflow completed")
    return state
end

-- Run parallel workflow
local parallel_state = parallel_analysis_workflow(topic)
print()

-- Example 4: Error Handling Workflow
print("=== Example 4: Error Handling Workflow ===")

local function robust_workflow(tasks)
    local state = WorkflowState:new()
    state:set("tasks", tasks)
    state:set("completed_tasks", {})
    
    for i, task in ipairs(tasks) do
        print("\nTask " .. i .. ": " .. task.name)
        
        -- Try to execute task
        local success, result = pcall(function()
            -- Simulate potential failures
            if task.fail_probability and math.random() < task.fail_probability then
                error("Simulated failure for " .. task.name)
            end
            
            -- Execute task
            local agent = agent.create({
                name = task.name,
                model = model,
                system = task.system or "You are a helpful assistant.",
                temperature = 0.4
            })
            
            return agent:run(task.prompt)
        end)
        
        if success then
            print("  ✓ Success")
            state:set(task.name, result)
            table.insert(state:get("completed_tasks"), task.name)
            state:checkpoint("after_" .. task.name)
        else
            print("  ❌ Failed: " .. tostring(result))
            state:add_error(task.name, tostring(result))
            
            -- Try recovery strategy
            if task.recovery_strategy then
                print("  🔧 Attempting recovery: " .. task.recovery_strategy)
                
                if task.recovery_strategy == "retry" then
                    -- Simple retry
                    core.sleep(1)
                    success, result = pcall(function()
                        local agent = agent.create({
                            name = task.name .. "_retry",
                            model = model,
                            system = task.system,
                            temperature = 0.4
                        })
                        return agent:run(task.prompt)
                    end)
                    
                    if success then
                        print("  ✓ Recovery successful")
                        state:set(task.name, result)
                    else
                        print("  ❌ Recovery failed")
                    end
                    
                elseif task.recovery_strategy == "fallback" then
                    -- Use fallback
                    state:set(task.name, task.fallback_value)
                    print("  ✓ Used fallback value")
                    
                elseif task.recovery_strategy == "skip" then
                    -- Skip and continue
                    print("  ⏭️  Skipping task")
                end
            end
        end
    end
    
    -- Summary
    print("\n📋 Workflow Summary:")
    print("  Completed: " .. #state:get("completed_tasks") .. "/" .. #tasks)
    print("  Errors: " .. #state.errors)
    
    return state
end

-- Define tasks with error handling
local robust_tasks = {
    {
        name = "data_fetch",
        prompt = "Fetch latest renewable energy statistics",
        fail_probability = 0.3,
        recovery_strategy = "retry"
    },
    {
        name = "analysis",
        prompt = "Analyze the data trends",
        fail_probability = 0.2,
        recovery_strategy = "fallback",
        fallback_value = "Analysis unavailable - using cached data"
    },
    {
        name = "visualization",
        prompt = "Create a visualization description",
        fail_probability = 0.4,
        recovery_strategy = "skip"
    }
}

-- Run robust workflow
local robust_state = robust_workflow(robust_tasks)
print()

-- Example 5: Workflow Composition
print("=== Example 5: Workflow Composition ===")

-- Define reusable workflow components
local WorkflowComponents = {}

function WorkflowComponents.research_component(topic, state)
    print("  📚 Research Component")
    local researcher = agent.create({
        name = "Researcher",
        model = model,
        system = "You are a research specialist.",
        tools = {"web_search"},
        temperature = 0.3
    })
    
    local research = researcher:run("Research: " .. topic)
    state:set("research", research)
    return research
end

function WorkflowComponents.validation_component(data, criteria, state)
    print("  ✅ Validation Component")
    local validator = agent.create({
        name = "Validator",
        model = model,
        system = "You validate data against criteria.",
        temperature = 0.2
    })
    
    local validation = validator:run(
        "Validate this data against the criteria:\n" ..
        "Data: " .. data .. "\n" ..
        "Criteria: " .. table.concat(criteria, ", ")
    )
    
    local is_valid = validation:lower():find("valid") or validation:lower():find("pass")
    state:set("validation_result", validation)
    state:set("is_valid", is_valid)
    
    return is_valid
end

function WorkflowComponents.transformation_component(data, format, state)
    print("  🔄 Transformation Component")
    local transformer = agent.create({
        name = "Transformer",
        model = model,
        system = "You transform data into different formats.",
        temperature = 0.3
    })
    
    local transformed = transformer:run(
        "Transform this data into " .. format .. " format:\n" .. data
    )
    state:set("transformed_data", transformed)
    return transformed
end

-- Compose a workflow from components
local function composed_workflow(topic)
    local state = WorkflowState:new()
    print("Running composed workflow for: " .. topic)
    
    -- Phase 1: Research
    print("\nPhase 1: Research")
    local research = WorkflowComponents.research_component(topic, state)
    
    -- Phase 2: Validation
    print("\nPhase 2: Validation")
    local criteria = {
        "Contains recent information (2023-2024)",
        "Includes statistical data",
        "Mentions key players or technologies"
    }
    local is_valid = WorkflowComponents.validation_component(research, criteria, state)
    
    if not is_valid then
        print("  ⚠️ Validation failed, enhancing research...")
        -- Re-run research with more specific prompt
        research = WorkflowComponents.research_component(
            topic .. " (focus on 2024 statistics and key technologies)",
            state
        )
    end
    
    -- Phase 3: Transform to multiple formats
    print("\nPhase 3: Multi-format Transformation")
    local formats = {"executive summary", "bullet points", "JSON structure"}
    
    for _, format in ipairs(formats) do
        print("  Transforming to: " .. format)
        local transformed = WorkflowComponents.transformation_component(research, format, state)
        tools.file_write(
            output_dir .. "/composed_" .. format:gsub(" ", "_") .. ".txt",
            transformed
        )
    end
    
    print("\n✓ Composed workflow completed")
    return state
end

-- Run composed workflow
local composed_state = composed_workflow(topic)
print()

-- Example 6: Event-Driven Workflow Preview
print("=== Example 6: Event-Driven Workflow Preview ===")

local function event_workflow_preview()
    local state = WorkflowState:new()
    
    -- Define workflow events
    local workflow_events = {
        "data_received",
        "processing_complete", 
        "quality_check_passed",
        "output_generated"
    }
    
    print("Event-driven workflow capabilities:")
    for _, event in ipairs(workflow_events) do
        print("  📡 " .. event)
    end
    
    print("\nThis pattern is demonstrated in detail in 07-event-driven.lua")
    
    return state
end

event_workflow_preview()

-- Summary
print("\n=== Summary ===")
print("This example demonstrated:")
print("1. Sequential workflows with state management")
print("2. Conditional workflows with quality control")
print("3. Parallel execution patterns")
print("4. Error handling and recovery strategies")
print("5. Workflow composition from reusable components")
print("6. Preview of event-driven patterns")
print()
print("Key insights:")
print("- State management is crucial for complex workflows")
print("- Checkpoints enable rollback capabilities")
print("- Parallel execution improves performance")
print("- Error handling makes workflows robust")
print("- Composition enables reusable workflow building blocks")
print()

-- List generated files
print("Files created in " .. output_dir .. ":")
local files = tools.list_files(output_dir)
for _, file in ipairs(files) do
    print("  - " .. file)
end

-- Return workflow statistics
return {
    workflows_demonstrated = 6,
    total_agents_created = 20,  -- Approximate count
    checkpoints_created = #pipeline_state.checkpoints + #qc_state.checkpoints,
    parallel_tasks = 4,
    files_generated = #files,
    patterns = {
        "sequential_state_management",
        "conditional_branching",
        "parallel_execution",
        "error_recovery",
        "workflow_composition"
    }
}