-- ABOUTME: Agent Management Library for go-llmspell Lua standard library
-- ABOUTME: Provides high-level agent lifecycle, communication, tools, and workflow orchestration

local agent = {}

-- Import promise library for async operations
local promise = _G.promise or require("promise")

-- Internal state for tracking agents and workflows
local active_agents = {}
local active_workflows = {}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if not param or param == "" then
        error(name .. " is required")
    end
end

-- Helper function to get agent bridge
local function get_agent_bridge()
    if not _G.agent_bridge then
        error("Agent bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.agent_bridge
end

-- Helper function to get workflow bridge
local function get_workflow_bridge()
    if not _G.workflow_bridge then
        error("Workflow bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.workflow_bridge
end

-- Helper function to merge options with defaults
local function merge_options(options, defaults)
    local merged = {}
    if defaults then
        for k, v in pairs(defaults) do
            merged[k] = v
        end
    end
    if options then
        for k, v in pairs(options) do
            merged[k] = v
        end
    end
    return merged
end

-- Agent Lifecycle Management

-- Create a new agent
function agent.create(name, config)
    validate_required(name, "agent name")

    local bridge = get_agent_bridge()
    local agent_config = config or {}

    -- Create agent through bridge
    local agent_obj = bridge:lifecycleCreate(name, agent_config)

    if agent_obj and agent_obj.id then
        -- Track active agent locally
        active_agents[agent_obj.id] = {
            name = name,
            config = agent_config,
            created_at = os.time(),
            state = "created",
        }
    end

    return agent_obj
end

-- Configure an existing agent
function agent.configure(agent_id, settings)
    validate_required(agent_id, "agent_id")
    validate_required(settings, "settings")

    local bridge = get_agent_bridge()

    -- Update agent configuration
    local success = bridge:stateSet(agent_id, settings)

    if success and active_agents[agent_id] then
        -- Update local tracking
        active_agents[agent_id].config = merge_options(settings, active_agents[agent_id].config)
        active_agents[agent_id].updated_at = os.time()
    end

    return success
end

-- Clone an agent with modifications
function agent.clone(source_agent_id, new_name, modifications)
    validate_required(source_agent_id, "source_agent_id")
    validate_required(new_name, "new_name")

    local bridge = get_agent_bridge()

    -- Get source agent state
    local source_state = bridge:stateGet(source_agent_id)
    if not source_state then
        error("Source agent not found: " .. tostring(source_agent_id))
    end

    -- Merge modifications with source state
    local new_config = merge_options(modifications or {}, source_state)

    -- Create new agent with merged configuration
    return agent.create(new_name, new_config)
end

-- Get agent information
function agent.get(agent_id)
    validate_required(agent_id, "agent_id")

    local bridge = get_agent_bridge()
    return bridge:lifecycleGet(agent_id)
end

-- List all agents
function agent.list()
    local bridge = get_agent_bridge()
    return bridge:lifecycleList()
end

-- Remove an agent
function agent.remove(agent_id)
    validate_required(agent_id, "agent_id")

    local bridge = get_agent_bridge()
    local success = bridge:lifecycleRemove(agent_id)

    if success and active_agents[agent_id] then
        active_agents[agent_id] = nil
    end

    return success
end

-- Agent Execution

-- Run agent synchronously
function agent.run(agent_id, input, options)
    validate_required(agent_id, "agent_id")
    validate_required(input, "input")

    local bridge = get_agent_bridge()
    local opts = options or {}

    -- Update agent state to running
    if active_agents[agent_id] then
        active_agents[agent_id].state = "running"
        active_agents[agent_id].last_run = os.time()
    end

    local result = bridge:run(agent_id, input, opts)

    -- Update agent state based on result
    if active_agents[agent_id] then
        active_agents[agent_id].state = result and "completed" or "error"
    end

    return result
end

-- Run agent asynchronously
function agent.run_async(agent_id, input, options)
    validate_required(agent_id, "agent_id")
    validate_required(input, "input")

    return promise.async(function()
        local bridge = get_agent_bridge()
        local opts = options or {}

        -- Update agent state to running
        if active_agents[agent_id] then
            active_agents[agent_id].state = "running"
            active_agents[agent_id].last_run = os.time()
        end

        local result = bridge:runAsync(agent_id, input, opts)

        -- Update agent state based on result
        if active_agents[agent_id] then
            active_agents[agent_id].state = result and "completed" or "error"
        end

        return result
    end)()
end

-- Agent Communication Helpers

-- Create a conversation session with an agent
function agent.conversation(agent_id, system_prompt)
    validate_required(agent_id, "agent_id")

    local session = {
        agent_id = agent_id,
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

    -- Send message to agent and get response
    function session:send(user_message, options)
        validate_required(user_message, "user_message")

        -- Add user message
        table.insert(self.messages, {
            role = "user",
            content = user_message,
        })

        -- Run agent with conversation context
        local input = {
            type = "conversation",
            messages = self.messages,
            options = options or {},
        }

        local response = agent.run(self.agent_id, input)

        -- Add agent response to conversation
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
        validate_required(user_message, "user_message")

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
            agent_id = self.agent_id,
            system_prompt = self.system_prompt,
            messages = self.messages,
        }
    end

    return session
end

-- Delegate a task from one agent to another
function agent.delegate(from_agent_id, to_agent_id, task, options)
    validate_required(from_agent_id, "from_agent_id")
    validate_required(to_agent_id, "to_agent_id")
    validate_required(task, "task")

    local opts = options or {}

    -- Prepare delegation input
    local delegation_input = {
        type = "delegation",
        delegating_agent = from_agent_id,
        task = task,
        options = opts,
    }

    -- Execute task on target agent
    local result = agent.run(to_agent_id, delegation_input)

    -- Notify source agent of completion if callback provided
    if opts.callback_to_source then
        local callback_input = {
            type = "delegation_result",
            original_task = task,
            target_agent = to_agent_id,
            result = result,
        }
        agent.run(from_agent_id, callback_input)
    end

    return result
end

-- Collaborate multiple agents on a task
function agent.collaborate(agent_ids, task, options)
    if type(agent_ids) ~= "table" or #agent_ids == 0 then
        error("agent_ids must be a non-empty table")
    end
    validate_required(task, "task")

    local opts = merge_options(options, {
        coordination_method = "sequential", -- sequential, parallel, or leader
        share_context = true,
    })

    local results = {}
    local shared_context = {}

    if opts.coordination_method == "parallel" then
        -- Run all agents in parallel
        for i, agent_id in ipairs(agent_ids) do
            local collaboration_input = {
                type = "collaboration",
                task = task,
                participant_agents = agent_ids,
                agent_index = i,
                shared_context = opts.share_context and shared_context or nil,
            }

            -- Run synchronously for simplicity in this version
            -- In a real implementation, this would be properly async
            local result = agent.run(agent_id, collaboration_input)
            results[agent_id] = result
        end
    elseif opts.coordination_method == "leader" then
        -- First agent acts as leader, coordinates others
        local leader_id = agent_ids[1]
        local follower_ids = {}
        for i = 2, #agent_ids do
            table.insert(follower_ids, agent_ids[i])
        end

        local leader_input = {
            type = "leader_coordination",
            task = task,
            follower_agents = follower_ids,
            shared_context = shared_context,
        }

        results[leader_id] = agent.run(leader_id, leader_input)
    else -- sequential
        -- Run agents sequentially, each building on previous results
        for i, agent_id in ipairs(agent_ids) do
            local collaboration_input = {
                type = "collaboration",
                task = task,
                participant_agents = agent_ids,
                agent_index = i,
                shared_context = opts.share_context and shared_context or nil,
                previous_results = results,
            }

            local result = agent.run(agent_id, collaboration_input)
            results[agent_id] = result

            -- Update shared context with result
            if opts.share_context and result then
                shared_context[agent_id] = result
            end
        end
    end

    return {
        results = results,
        shared_context = shared_context,
        coordination_method = opts.coordination_method,
    }
end

-- Agent Tool Integration

-- Add tools to an agent
function agent.add_tools(agent_id, tools)
    validate_required(agent_id, "agent_id")
    if type(tools) ~= "table" then
        error("tools must be a table")
    end

    local bridge = get_agent_bridge()
    local success_count = 0

    for _, tool in ipairs(tools) do
        local success = bridge:registerTool(agent_id, tool)
        if success then
            success_count = success_count + 1
        end
    end

    return {
        total_tools = #tools,
        successful_registrations = success_count,
        success = success_count == #tools,
    }
end

-- Create a custom tool
function agent.create_tool(name, func, schema)
    validate_required(name, "tool name")
    validate_required(func, "tool function")

    if type(func) ~= "function" then
        error("func must be a function")
    end

    local tool = {
        name = name,
        description = schema and schema.description or ("Custom tool: " .. name),
        parameters = schema and schema.parameters or {},
        execute = func,
        schema = schema or {},
    }

    return tool
end

-- Get tools assigned to an agent
function agent.get_tools(agent_id)
    validate_required(agent_id, "agent_id")

    local bridge = get_agent_bridge()
    return bridge:listTools(agent_id)
end

-- Create a tool chain (pipeline of tools)
function agent.tool_chain(tools, initial_data, options)
    if type(tools) ~= "table" or #tools == 0 then
        error("tools must be a non-empty table")
    end

    local opts = merge_options(options, {
        stop_on_error = true,
        pass_previous_output = true,
    })

    local results = {}
    local current_data = initial_data

    for i, tool in ipairs(tools) do
        local tool_input = opts.pass_previous_output and current_data or initial_data

        local success, result = pcall(function()
            if type(tool) == "function" then
                return tool(tool_input)
            elseif type(tool) == "table" and tool.execute then
                return tool.execute(tool_input)
            else
                error(
                    "Invalid tool at index "
                        .. i
                        .. ": must be function or table with execute method"
                )
            end
        end)

        if success then
            results[i] = result
            current_data = result
        else
            results[i] = { error = result }
            if opts.stop_on_error then
                break
            end
        end
    end

    return {
        results = results,
        final_output = current_data,
        success = #results == #tools,
    }
end

-- Workflow Orchestration Helpers

-- Create a workflow
function agent.workflow_create(name, steps, options)
    validate_required(name, "workflow name")
    if type(steps) ~= "table" or #steps == 0 then
        error("steps must be a non-empty table")
    end

    local bridge = get_workflow_bridge()
    local workflow_config = merge_options(options, {
        type = "sequential", -- sequential, parallel, conditional, loop
    })

    local workflow = bridge:create(name, workflow_config)

    if workflow and workflow.id then
        -- Add steps to workflow
        for _, step in ipairs(steps) do
            bridge:addStep(workflow.id, step)
        end

        -- Track workflow locally
        active_workflows[workflow.id] = {
            name = name,
            steps = steps,
            config = workflow_config,
            created_at = os.time(),
            state = "created",
        }
    end

    return workflow
end

-- Run a workflow
function agent.workflow_run(workflow_id, input, options)
    validate_required(workflow_id, "workflow_id")

    local bridge = get_workflow_bridge()
    local opts = options or {}

    -- Update workflow state
    if active_workflows[workflow_id] then
        active_workflows[workflow_id].state = "running"
        active_workflows[workflow_id].last_run = os.time()
    end

    local result = bridge:execute(workflow_id, input, opts)

    -- Update workflow state based on result
    if active_workflows[workflow_id] then
        active_workflows[workflow_id].state = result and "completed" or "error"
    end

    return result
end

-- Run workflow steps in parallel
function agent.workflow_parallel(steps, input, options)
    if type(steps) ~= "table" or #steps == 0 then
        error("steps must be a non-empty table")
    end

    local _ = options or {} -- options for future use
    local promises = {}

    -- Execute each step as a promise
    for i, step in ipairs(steps) do
        promises[i] = promise.async(function()
            if type(step) == "function" then
                return step(input)
            elseif type(step) == "string" then
                -- Assume it's an agent ID
                return agent.run(step, input)
            elseif type(step) == "table" and step.agent_id then
                return agent.run(step.agent_id, input, step.options)
            else
                error("Invalid step format at index " .. i)
            end
        end)()
    end

    -- Wait for all steps to complete
    return promise.Promise.all(promises)
end

-- Conditional workflow execution
function agent.workflow_conditional(condition, then_step, else_step, input, options)
    validate_required(condition, "condition")
    validate_required(then_step, "then_step")

    local _ = options or {} -- options for future use
    local condition_result

    -- Evaluate condition
    if type(condition) == "function" then
        condition_result = condition(input)
    elseif type(condition) == "boolean" then
        condition_result = condition
    else
        error("condition must be a function or boolean")
    end

    -- Execute appropriate step
    local step_to_execute = condition_result and then_step or else_step
    if not step_to_execute then
        return {
            condition_result = condition_result,
            executed = false,
            result = nil,
        }
    end

    local result
    if type(step_to_execute) == "function" then
        result = step_to_execute(input)
    elseif type(step_to_execute) == "string" then
        result = agent.run(step_to_execute, input)
    elseif type(step_to_execute) == "table" and step_to_execute.agent_id then
        result = agent.run(step_to_execute.agent_id, input, step_to_execute.options)
    else
        error("Invalid step format")
    end

    return {
        condition_result = condition_result,
        executed = true,
        result = result,
    }
end

-- Utility Functions

-- Get agent status
function agent.get_status(agent_id)
    validate_required(agent_id, "agent_id")

    local bridge = get_agent_bridge()
    local agent_info = bridge:lifecycleGet(agent_id)
    local local_info = active_agents[agent_id]

    return {
        bridge_info = agent_info,
        local_info = local_info,
        exists = agent_info ~= nil,
    }
end

-- Get workflow status
function agent.get_workflow_status(workflow_id)
    validate_required(workflow_id, "workflow_id")

    local bridge = get_workflow_bridge()
    local workflow_info = bridge:get(workflow_id)
    local local_info = active_workflows[workflow_id]

    return {
        bridge_info = workflow_info,
        local_info = local_info,
        exists = workflow_info ~= nil,
    }
end

-- List active agents (locally tracked)
function agent.list_active()
    return active_agents
end

-- List active workflows (locally tracked)
function agent.list_active_workflows()
    return active_workflows
end

-- Export the module
return agent
