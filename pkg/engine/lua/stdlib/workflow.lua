-- ABOUTME: Workflow orchestration library for Lua that wraps the agent_workflow bridge
-- ABOUTME: Provides workflow creation, execution, step management, templates, and serialization capabilities

local workflow = {}

-- Version information
workflow._VERSION = "2.1.0"
workflow._DESCRIPTION = "Workflow orchestration library for Lua"

-- Get the workflow bridge
local function get_workflow_bridge()
    if bridges and bridges.agent_workflow then
        return bridges.agent_workflow
    end
    return nil
end

-- Check if bridge is available
local function with_workflow_bridge(method_name, ...)
    local workflow_bridge = get_workflow_bridge()
    if not workflow_bridge then
        error("Workflow bridge not available for " .. method_name)
    end
    return workflow_bridge[method_name](...)
end

-- Constants
workflow.TYPES = {
    SEQUENTIAL = "sequential",
    PARALLEL = "parallel",
    CONDITIONAL = "conditional",
    LOOP = "loop",
    CUSTOM = "custom"
}

-- Using STATUS instead of STATES per adapter
workflow.STATUS = {
    CREATED = "created",
    PENDING = "pending",
    RUNNING = "running",
    PAUSED = "paused",
    COMPLETED = "completed",
    FAILED = "failed",
    CANCELLED = "cancelled"
}

workflow.FORMATS = {
    JSON = "json",
    YAML = "yaml"
}

workflow.STEP_TYPES = {
    AGENT = "agent",
    SCRIPT = "script",
    CONDITION = "condition",
    PARALLEL = "parallel",
    WAIT = "wait"
}

-- Workflow lifecycle methods (snake_case)
function workflow.create_workflow(id, config)
    return with_workflow_bridge("createWorkflow", id, config)
end

function workflow.execute_workflow(workflow_id, input)
    if input then
        return with_workflow_bridge("executeWorkflow", workflow_id, input)
    else
        return with_workflow_bridge("executeWorkflow", workflow_id)
    end
end

function workflow.pause_workflow(workflow_id)
    return with_workflow_bridge("pauseWorkflow", workflow_id)
end

function workflow.resume_workflow(workflow_id)
    return with_workflow_bridge("resumeWorkflow", workflow_id)
end

function workflow.stop_workflow(workflow_id)
    return with_workflow_bridge("stopWorkflow", workflow_id)
end

function workflow.get_workflow_status(workflow_id)
    return with_workflow_bridge("getWorkflowStatus", workflow_id)
end

function workflow.list_workflows()
    return with_workflow_bridge("listWorkflows")
end

function workflow.get_workflow(workflow_id)
    return with_workflow_bridge("getWorkflow", workflow_id)
end

function workflow.delete_workflow(workflow_id)
    return with_workflow_bridge("deleteWorkflow", workflow_id)
end

-- Step management methods (snake_case)
function workflow.add_step(workflow_id, step)
    return with_workflow_bridge("addStep", workflow_id, step)
end

function workflow.remove_step(workflow_id, step_id)
    return with_workflow_bridge("removeStep", workflow_id, step_id)
end

function workflow.update_step(workflow_id, step_id, updates)
    return with_workflow_bridge("updateStep", workflow_id, step_id, updates)
end

function workflow.get_step(workflow_id, step_id)
    return with_workflow_bridge("getStep", workflow_id, step_id)
end

function workflow.list_steps(workflow_id)
    return with_workflow_bridge("listSteps", workflow_id)
end

function workflow.reorder_steps(workflow_id, step_ids)
    return with_workflow_bridge("reorderSteps", workflow_id, step_ids)
end

-- Template methods (snake_case)
function workflow.list_templates()
    return with_workflow_bridge("listTemplates")
end

function workflow.list_workflow_templates()
    return with_workflow_bridge("listWorkflowTemplates")
end

function workflow.get_template(template_id)
    return with_workflow_bridge("getTemplate", template_id)
end

function workflow.get_workflow_template(template_id)
    return with_workflow_bridge("getWorkflowTemplate", template_id)
end

function workflow.create_workflow_template(workflow_id, template_name)
    return with_workflow_bridge("createWorkflowTemplate", workflow_id, template_name)
end

function workflow.remove_workflow_template(template_id)
    return with_workflow_bridge("removeWorkflowTemplate", template_id)
end

function workflow.create_workflow_from_template(template_id, workflow_id, variables)
    if variables then
        return with_workflow_bridge("createWorkflowFromTemplate", template_id, workflow_id, variables)
    else
        return with_workflow_bridge("createWorkflowFromTemplate", template_id, workflow_id)
    end
end

function workflow.save_as_template(workflow_id, template_config)
    return with_workflow_bridge("saveAsTemplate", workflow_id, template_config)
end

-- Import/Export methods (snake_case)
function workflow.export_workflow(workflow_id, format)
    format = format or "json"
    return with_workflow_bridge("exportWorkflow", workflow_id, format)
end

function workflow.import_workflow(data, format)
    format = format or "json"
    return with_workflow_bridge("importWorkflow", data, format)
end

-- Variable management (snake_case)
function workflow.set_workflow_variable(workflow_id, name, value)
    return with_workflow_bridge("setWorkflowVariable", workflow_id, name, value)
end

function workflow.get_workflow_variable(workflow_id, name)
    return with_workflow_bridge("getWorkflowVariable", workflow_id, name)
end

function workflow.list_workflow_variables(workflow_id)
    return with_workflow_bridge("listWorkflowVariables", workflow_id)
end

function workflow.remove_workflow_variable(workflow_id, name)
    return with_workflow_bridge("removeWorkflowVariable", workflow_id, name)
end

-- Error handling (snake_case)
function workflow.get_workflow_errors(workflow_id)
    return with_workflow_bridge("getWorkflowErrors", workflow_id)
end

function workflow.clear_workflow_errors(workflow_id)
    return with_workflow_bridge("clearWorkflowErrors", workflow_id)
end

-- Convenience methods (snake_case)
function workflow.create_builder(workflow_id)
    return with_workflow_bridge("createBuilder", workflow_id)
end

function workflow.validate_workflow(workflow_or_id)
    return with_workflow_bridge("validateWorkflow", workflow_or_id)
end

-- Builder pattern helper
workflow.builder = {}
workflow.builder.__index = workflow.builder

function workflow.builder:new(workflow_id)
    local builder = {
        _workflow_id = workflow_id,
        _config = {}
    }
    setmetatable(builder, self)
    return builder
end

function workflow.builder:with_type(workflow_type)
    self._config.type = workflow_type
    return self
end

function workflow.builder:with_name(name)
    self._config.name = name
    return self
end

function workflow.builder:with_description(description)
    self._config.description = description
    return self
end

function workflow.builder:add_step(step)
    if not self._config.steps then
        self._config.steps = {}
    end
    table.insert(self._config.steps, step)
    return self
end

function workflow.builder:build()
    return workflow.create_workflow(self._workflow_id, self._config)
end

-- Alternative constructor for builder
function workflow.new_builder(workflow_id)
    return workflow.builder:new(workflow_id)
end

-- Bridge access to raw methods if available
if bridges and bridges.agent_workflow then
    -- Allow access to raw bridge methods
    workflow.bridge = bridges.agent_workflow
end

return workflow