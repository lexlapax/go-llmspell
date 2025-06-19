// ABOUTME: Tests for Workflow bridge adapter that exposes go-llms workflow functionality to Lua scripts
// ABOUTME: Validates workflow creation, execution, step management, templates, and serialization

package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestWorkflowAdapter_Creation(t *testing.T) {
	t.Run("create_workflow_adapter", func(t *testing.T) {
		// Create workflow bridge mock
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Workflow Bridge",
				Version:     "2.1.0",
				Description: "Enhanced workflow engine bridge with serialization, script steps, and templates (v0.3.5)",
			}).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name:        "createWorkflow",
				Description: "Create a new workflow",
				ReturnType:  "string",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock workflow creation
				id := args[0].(engine.StringValue).Value()
				return engine.NewStringValue(id), nil
			}).
			WithMethod("listWorkflows", engine.MethodInfo{
				Name:        "listWorkflows",
				Description: "List all workflows",
				ReturnType:  "array",
			}, nil).
			WithMethod("executeWorkflow", engine.MethodInfo{
				Name:        "executeWorkflow",
				Description: "Execute a workflow",
				ReturnType:  "object",
			}, nil).
			WithMethod("pauseWorkflow", engine.MethodInfo{
				Name:        "pauseWorkflow",
				Description: "Pause a workflow",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("resumeWorkflow", engine.MethodInfo{
				Name:        "resumeWorkflow",
				Description: "Resume a workflow",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("stopWorkflow", engine.MethodInfo{
				Name:        "stopWorkflow",
				Description: "Stop a workflow",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("addStep", engine.MethodInfo{
				Name:        "addStep",
				Description: "Add a step to workflow",
				ReturnType:  "string",
			}, nil).
			WithMethod("removeStep", engine.MethodInfo{
				Name:        "removeStep",
				Description: "Remove a step from workflow",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("exportWorkflow", engine.MethodInfo{
				Name:        "exportWorkflow",
				Description: "Export workflow",
				ReturnType:  "string",
			}, nil).
			WithMethod("importWorkflow", engine.MethodInfo{
				Name:        "importWorkflow",
				Description: "Import workflow",
				ReturnType:  "object",
			}, nil).
			WithMethod("createWorkflowFromTemplate", engine.MethodInfo{
				Name:        "createWorkflowFromTemplate",
				Description: "Create workflow from template",
				ReturnType:  "string",
			}, nil).
			WithMethod("getWorkflow", engine.MethodInfo{
				Name:        "getWorkflow",
				Description: "Get workflow details",
				ReturnType:  "object",
			}, nil).
			WithMethod("deleteWorkflow", engine.MethodInfo{
				Name:        "deleteWorkflow",
				Description: "Delete a workflow",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("getStep", engine.MethodInfo{
				Name:        "getStep",
				Description: "Get step details",
				ReturnType:  "object",
			}, nil).
			WithMethod("listSteps", engine.MethodInfo{
				Name:        "listSteps",
				Description: "List workflow steps",
				ReturnType:  "array",
			}, nil).
			WithMethod("updateStep", engine.MethodInfo{
				Name:        "updateStep",
				Description: "Update a step",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("getWorkflowStatus", engine.MethodInfo{
				Name:        "getWorkflowStatus",
				Description: "Get workflow status",
				ReturnType:  "string",
			}, nil).
			WithMethod("validateWorkflow", engine.MethodInfo{
				Name:        "validateWorkflow",
				Description: "Validate workflow",
				ReturnType:  "object",
			}, nil).
			WithMethod("listWorkflowTemplates", engine.MethodInfo{
				Name:        "listWorkflowTemplates",
				Description: "List workflow templates",
				ReturnType:  "array",
			}, nil).
			WithMethod("getTemplate", engine.MethodInfo{
				Name:        "getTemplate",
				Description: "Get template details",
				ReturnType:  "object",
			}, nil).
			WithMethod("saveAsTemplate", engine.MethodInfo{
				Name:        "saveAsTemplate",
				Description: "Save workflow as template",
				ReturnType:  "string",
			}, nil).
			WithMethod("setWorkflowVariables", engine.MethodInfo{
				Name:        "setWorkflowVariables",
				Description: "Set workflow variables",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("getWorkflowVariables", engine.MethodInfo{
				Name:        "getWorkflowVariables",
				Description: "Get workflow variables",
				ReturnType:  "object",
			}, nil).
			WithMethod("listWorkflowVariables", engine.MethodInfo{
				Name:        "listWorkflowVariables",
				Description: "List workflow variables",
				ReturnType:  "object",
			}, nil)

		// Create adapter
		adapter := NewWorkflowAdapter(workflowBridge)
		require.NotNil(t, adapter)

		// Should have workflow-specific methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "createWorkflow")
		assert.Contains(t, methods, "executeWorkflow")
		assert.Contains(t, methods, "pauseWorkflow")
		assert.Contains(t, methods, "resumeWorkflow")
		assert.Contains(t, methods, "stopWorkflow")
		assert.Contains(t, methods, "listWorkflows")
		assert.Contains(t, methods, "addStep")
		assert.Contains(t, methods, "removeStep")
		assert.Contains(t, methods, "exportWorkflow")
		assert.Contains(t, methods, "importWorkflow")
	})

	t.Run("workflow_module_structure", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name: "Workflow Bridge",
			}).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name: "createWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("workflow-123"), nil
			}).
			WithMethod("listWorkflows", engine.MethodInfo{
				Name: "listWorkflows",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{}), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Create module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("workflow", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test module structure
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Check basic module properties
			assert(workflow._adapter == "workflow", "should have correct adapter name")
			assert(workflow._version == "2.1.0", "should have correct version")
			
			-- Check workflow types constants
			assert(workflow.TYPES.SEQUENTIAL == "sequential", "should have sequential type")
			assert(workflow.TYPES.PARALLEL == "parallel", "should have parallel type")
			assert(workflow.TYPES.CONDITIONAL == "conditional", "should have conditional type")
			
			-- Check status constants
			assert(workflow.STATUS.CREATED == "created", "should have created status")
			assert(workflow.STATUS.RUNNING == "running", "should have running status")
			assert(workflow.STATUS.PAUSED == "paused", "should have paused status")
			assert(workflow.STATUS.COMPLETED == "completed", "should have completed status")
			assert(workflow.STATUS.FAILED == "failed", "should have failed status")
			
			-- Check export formats
			assert(workflow.FORMATS.JSON == "json", "should have JSON format")
			assert(workflow.FORMATS.YAML == "yaml", "should have YAML format")
		`)
		assert.NoError(t, err)
	})
}

func TestWorkflowAdapter_WorkflowLifecycle(t *testing.T) {
	t.Run("create_and_execute_workflow", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name: "createWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				config := args[1].(engine.ObjectValue).Fields()

				// Validate config
				assert.NotNil(t, config["type"])
				assert.NotNil(t, config["name"])

				return engine.NewStringValue(id), nil
			}).
			WithMethod("executeWorkflow", engine.MethodInfo{
				Name: "executeWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				workflowID := args[0].(engine.StringValue).Value()
				input := args[1].(engine.ObjectValue).Fields()

				// Mock execution result
				result := map[string]engine.ScriptValue{
					"workflowID": engine.NewStringValue(workflowID),
					"status":     engine.NewStringValue("completed"),
					"output":     input["data"], // Echo input data
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test workflow creation and execution
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Create a sequential workflow
			local wfID = workflow.createWorkflow("test-workflow", {
				type = workflow.TYPES.SEQUENTIAL,
				name = "Test Workflow",
				description = "A test workflow"
			})
			assert(wfID == "test-workflow", "should return workflow ID")
			
			-- Execute the workflow
			local result = workflow.executeWorkflow(wfID, {
				data = "test input"
			})
			assert(result.status == "completed", "workflow should complete")
			assert(result.output == "test input", "should return output")
		`)
		assert.NoError(t, err)
	})

	t.Run("workflow_control_operations", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("pauseWorkflow", engine.MethodInfo{
				Name: "pauseWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			}).
			WithMethod("resumeWorkflow", engine.MethodInfo{
				Name: "resumeWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			}).
			WithMethod("stopWorkflow", engine.MethodInfo{
				Name: "stopWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			}).
			WithMethod("getWorkflowStatus", engine.MethodInfo{
				Name: "getWorkflowStatus",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("paused"), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test workflow control
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Pause workflow
			local paused = workflow.pauseWorkflow("test-workflow")
			assert(paused == true, "should pause workflow")
			
			-- Get status
			local status = workflow.getWorkflowStatus("test-workflow")
			assert(status == "paused", "should be paused")
			
			-- Resume workflow
			local resumed = workflow.resumeWorkflow("test-workflow")
			assert(resumed == true, "should resume workflow")
			
			-- Stop workflow
			local stopped = workflow.stopWorkflow("test-workflow")
			assert(stopped == true, "should stop workflow")
		`)
		assert.NoError(t, err)
	})

	t.Run("list_and_get_workflows", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("listWorkflows", engine.MethodInfo{
				Name: "listWorkflows",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock workflow list
				workflows := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":   engine.NewStringValue("wf-1"),
						"name": engine.NewStringValue("Workflow 1"),
						"type": engine.NewStringValue("sequential"),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":   engine.NewStringValue("wf-2"),
						"name": engine.NewStringValue("Workflow 2"),
						"type": engine.NewStringValue("parallel"),
					}),
				}
				return engine.NewArrayValue(workflows), nil
			}).
			WithMethod("getWorkflow", engine.MethodInfo{
				Name: "getWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				workflowID := args[0].(engine.StringValue).Value()
				result := map[string]engine.ScriptValue{
					"id":     engine.NewStringValue(workflowID),
					"name":   engine.NewStringValue("Test Workflow"),
					"type":   engine.NewStringValue("sequential"),
					"status": engine.NewStringValue("created"),
					"steps":  engine.NewNumberValue(3),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test listing and getting workflows
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- List workflows - capture multiple return values
			local workflows = {workflow.listWorkflows()}
			assert(#workflows == 2, "should have 2 workflows")
			assert(workflows[1].id == "wf-1", "first workflow ID")
			assert(workflows[2].type == "parallel", "second workflow type")
			
			-- Get specific workflow
			local wf = workflow.getWorkflow("test-workflow")
			assert(wf.name == "Test Workflow", "should have correct name")
			assert(wf.steps == 3, "should have 3 steps")
		`)
		assert.NoError(t, err)
	})
}

func TestWorkflowAdapter_StepManagement(t *testing.T) {
	t.Run("add_and_list_steps", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("addStep", engine.MethodInfo{
				Name: "addStep",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				step := args[1].(engine.ObjectValue).Fields()

				// Validate step config
				assert.NotNil(t, step["name"])
				assert.NotNil(t, step["type"])

				return engine.NewStringValue("step-123"), nil
			}).
			WithMethod("listSteps", engine.MethodInfo{
				Name: "listSteps",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				steps := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":   engine.NewStringValue("step-1"),
						"name": engine.NewStringValue("Step 1"),
						"type": engine.NewStringValue("action"),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":   engine.NewStringValue("step-2"),
						"name": engine.NewStringValue("Step 2"),
						"type": engine.NewStringValue("condition"),
					}),
				}
				return engine.NewArrayValue(steps), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test step management
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Add a step
			local stepID = workflow.addStep("test-workflow", {
				name = "Process Data",
				type = "action",
				config = {
					action = "transform"
				}
			})
			assert(stepID == "step-123", "should return step ID")
			
			-- List steps - capture multiple return values
			local steps = {workflow.listSteps("test-workflow")}
			assert(#steps == 2, "should have 2 steps")
			assert(steps[1].name == "Step 1", "first step name")
			assert(steps[2].type == "condition", "second step type")
		`)
		assert.NoError(t, err)
	})

	t.Run("update_and_remove_steps", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("updateStep", engine.MethodInfo{
				Name: "updateStep",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			}).
			WithMethod("removeStep", engine.MethodInfo{
				Name: "removeStep",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			}).
			WithMethod("getStep", engine.MethodInfo{
				Name: "getStep",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				result := map[string]engine.ScriptValue{
					"id":   engine.NewStringValue("step-123"),
					"name": engine.NewStringValue("Updated Step"),
					"type": engine.NewStringValue("action"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test step updates
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Update a step
			local updated = workflow.updateStep("test-workflow", "step-123", {
				name = "Updated Step"
			})
			assert(updated == true, "should update step")
			
			-- Get step details
			local step = workflow.getStep("test-workflow", "step-123")
			assert(step.name == "Updated Step", "should have updated name")
			
			-- Remove a step
			local removed = workflow.removeStep("test-workflow", "step-123")
			assert(removed == true, "should remove step")
		`)
		assert.NoError(t, err)
	})
}

func TestWorkflowAdapter_Templates(t *testing.T) {
	t.Run("list_and_use_templates", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("listWorkflowTemplates", engine.MethodInfo{
				Name: "listWorkflowTemplates",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				templates := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":          engine.NewStringValue("tmpl-1"),
						"name":        engine.NewStringValue("Basic Sequential"),
						"description": engine.NewStringValue("A basic sequential workflow"),
						"category":    engine.NewStringValue("basic"),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":          engine.NewStringValue("tmpl-2"),
						"name":        engine.NewStringValue("Parallel Processing"),
						"description": engine.NewStringValue("Process data in parallel"),
						"category":    engine.NewStringValue("advanced"),
					}),
				}
				return engine.NewArrayValue(templates), nil
			}).
			WithMethod("createWorkflowFromTemplate", engine.MethodInfo{
				Name: "createWorkflowFromTemplate",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				templateID := args[0].(engine.StringValue).Value()
				workflowID := args[1].(engine.StringValue).Value()

				// Validate template ID
				assert.Equal(t, "tmpl-1", templateID)

				return engine.NewStringValue(workflowID), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test templates
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- List templates - capture multiple return values
			local templates = {workflow.listWorkflowTemplates()}
			assert(#templates == 2, "should have 2 templates")
			assert(templates[1].name == "Basic Sequential", "first template name")
			assert(templates[2].category == "advanced", "second template category")
			
			-- Create workflow from template
			local wfID = workflow.createWorkflowFromTemplate("tmpl-1", "my-workflow", {
				param1 = "value1",
				param2 = 42
			})
			assert(wfID == "my-workflow", "should create workflow from template")
		`)
		assert.NoError(t, err)
	})

	t.Run("create_and_manage_templates", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("createWorkflowTemplate", engine.MethodInfo{
				Name: "createWorkflowTemplate",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("tmpl-custom"), nil
			}).
			WithMethod("getWorkflowTemplate", engine.MethodInfo{
				Name: "getWorkflowTemplate",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				result := map[string]engine.ScriptValue{
					"id":          engine.NewStringValue("tmpl-custom"),
					"name":        engine.NewStringValue("Custom Template"),
					"description": engine.NewStringValue("A custom workflow template"),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("removeWorkflowTemplate", engine.MethodInfo{
				Name: "removeWorkflowTemplate",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test template management
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Create a template from existing workflow
			local templateID = workflow.createWorkflowTemplate("test-workflow", "Custom Template")
			assert(templateID == "tmpl-custom", "should create template")
			
			-- Get template details
			local template = workflow.getWorkflowTemplate("tmpl-custom")
			assert(template.name == "Custom Template", "should have correct name")
			
			-- Remove template
			local removed = workflow.removeWorkflowTemplate("tmpl-custom")
			assert(removed == true, "should remove template")
		`)
		assert.NoError(t, err)
	})
}

func TestWorkflowAdapter_ImportExport(t *testing.T) {
	t.Run("export_and_import_workflow", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("exportWorkflow", engine.MethodInfo{
				Name: "exportWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				_ = args[0].(engine.StringValue).Value() // workflowID - used for validation in real implementation
				format := "json"
				if len(args) > 1 {
					format = args[1].(engine.StringValue).Value()
				}

				// Mock exported data
				exportData := `{"name":"Test Workflow","type":"sequential","steps":[{"name":"Step 1"}]}`
				if format == "yaml" {
					exportData = "name: Test Workflow\ntype: sequential\nsteps:\n  - name: Step 1"
				}

				return engine.NewStringValue(exportData), nil
			}).
			WithMethod("importWorkflow", engine.MethodInfo{
				Name: "importWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				data := args[0].(engine.StringValue).Value()

				// Validate import data contains expected content
				assert.Contains(t, data, "Test Workflow")

				result := map[string]engine.ScriptValue{
					"id":          engine.NewStringValue("imported-workflow"),
					"name":        engine.NewStringValue("Test Workflow"),
					"description": engine.NewStringValue("Imported workflow"),
					"steps":       engine.NewNumberValue(1),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test import/export
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Export workflow as JSON
			local jsonData = workflow.exportWorkflow("test-workflow", workflow.FORMATS.JSON)
			assert(string.find(jsonData, "sequential") ~= nil, "should contain workflow type")
			
			-- Export workflow as YAML
			local yamlData = workflow.exportWorkflow("test-workflow", workflow.FORMATS.YAML)
			assert(string.find(yamlData, "name: Test Workflow") ~= nil, "should be YAML format")
			
			-- Import workflow
			local imported = workflow.importWorkflow(jsonData, workflow.FORMATS.JSON)
			assert(imported.name == "Test Workflow", "should import workflow")
			assert(imported.steps == 1, "should have correct step count")
		`)
		assert.NoError(t, err)
	})
}

func TestWorkflowAdapter_Variables(t *testing.T) {
	t.Run("workflow_variables", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("setWorkflowVariable", engine.MethodInfo{
				Name: "setWorkflowVariable",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("getWorkflowVariable", engine.MethodInfo{
				Name: "getWorkflowVariable",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				name := args[1].(engine.StringValue).Value()
				if name == "test_var" {
					return engine.NewStringValue("test_value"), nil
				}
				return engine.NewNilValue(), nil
			}).
			WithMethod("listWorkflowVariables", engine.MethodInfo{
				Name: "listWorkflowVariables",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				vars := map[string]engine.ScriptValue{
					"var1": engine.NewStringValue("value1"),
					"var2": engine.NewNumberValue(42),
				}
				return engine.NewObjectValue(vars), nil
			}).
			WithMethod("removeWorkflowVariable", engine.MethodInfo{
				Name: "removeWorkflowVariable",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test variables
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Set a variable
			workflow.setWorkflowVariable("test-workflow", "test_var", "test_value")
			
			-- Get a variable
			local value = workflow.getWorkflowVariable("test-workflow", "test_var")
			assert(value == "test_value", "should get variable value")
			
			-- List variables
			local vars = workflow.listWorkflowVariables("test-workflow")
			assert(vars.var1 == "value1", "should have var1")
			assert(vars.var2 == 42, "should have var2")
			
			-- Remove a variable
			local removed = workflow.removeWorkflowVariable("test-workflow", "test_var")
			assert(removed == true, "should remove variable")
		`)
		assert.NoError(t, err)
	})
}

func TestWorkflowAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(false).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name: "createWorkflow",
			}, nil)

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test error handling
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Try to create workflow when bridge not initialized
			-- Bridge methods return (result, error) in Lua convention
			local result, err = workflow.createWorkflow("test-workflow", {type = "sequential"})
			
			assert(result == nil, "result should be nil when bridge not initialized")
			assert(err ~= nil, "should have error when bridge not initialized")
			assert(string.find(err, "not initialized") ~= nil, "should have initialization error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})

	t.Run("handle_invalid_workflow_type", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name: "createWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				config := args[1].(engine.ObjectValue).Fields()
				if typeVal, ok := config["type"]; ok {
					workflowType := typeVal.(engine.StringValue).Value()
					if workflowType != "sequential" && workflowType != "parallel" && workflowType != "conditional" {
						return nil, fmt.Errorf("unsupported workflow type: %s", workflowType)
					}
				}
				return engine.NewStringValue("workflow-123"), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test invalid workflow type
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Try to create workflow with invalid type
			-- Bridge methods return (result, error) in Lua convention
			local result, err = workflow.createWorkflow("test-workflow", {type = "invalid"})
			
			assert(result == nil, "result should be nil for invalid type")
			assert(err ~= nil, "should have error for invalid type")
			assert(string.find(err, "unsupported workflow type") ~= nil, "should have correct error message: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})
}

func TestWorkflowAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("workflow_builder_pattern", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name: "createWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				config := args[1].(engine.ObjectValue).Fields()

				// Verify builder pattern created correct workflow
				workflowType := config["type"].(engine.StringValue).Value()
				assert.Equal(t, "sequential", workflowType)

				return engine.NewStringValue(id), nil
			}).
			WithMethod("addStep", engine.MethodInfo{
				Name: "addStep",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("step-123"), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test workflow builder pattern
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Create workflow using builder pattern
			local wf = workflow.createBuilder("builder-workflow")
				:withType(workflow.TYPES.SEQUENTIAL)
				:withName("Builder Workflow")
				:withDescription("Created with builder pattern")
				:addStep({
					name = "Step 1",
					type = "action"
				})
				:addStep({
					name = "Step 2",
					type = "condition"
				})
				:build()
			
			assert(wf == "builder-workflow", "should create workflow with builder pattern")
		`)
		assert.NoError(t, err)
	})

	t.Run("workflow_validation", func(t *testing.T) {
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("validateWorkflow", engine.MethodInfo{
				Name: "validateWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				config := args[0].(engine.ObjectValue).Fields()

				errors := []engine.ScriptValue{}

				// Validate required fields
				if _, ok := config["type"]; !ok {
					errors = append(errors, engine.NewStringValue("type is required"))
				}
				if _, ok := config["name"]; !ok {
					errors = append(errors, engine.NewStringValue("name is required"))
				}

				result := map[string]engine.ScriptValue{
					"valid":  engine.NewBoolValue(len(errors) == 0),
					"errors": engine.NewArrayValue(errors),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewWorkflowAdapter(workflowBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()
		L.PreloadModule("workflow", loader)
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("workflow", module)

		// Test validation
		err = L.DoString(`
			local workflow = require("workflow")
			
			-- Validate valid workflow config
			local result1 = workflow.validateWorkflow({
				type = workflow.TYPES.SEQUENTIAL,
				name = "Valid Workflow"
			})
			assert(result1.valid == true, "should be valid")
			assert(#result1.errors == 0, "should have no errors")
			
			-- Validate invalid workflow config
			local result2 = workflow.validateWorkflow({
				-- missing type and name
			})
			assert(result2.valid == false, "should be invalid")
			assert(#result2.errors == 2, "should have 2 errors")
		`)
		assert.NoError(t, err)
	})
}
