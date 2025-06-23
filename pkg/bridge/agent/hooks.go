// ABOUTME: Bridge for go-llms agent hook system enabling script-based lifecycle hooks
// ABOUTME: Provides hook registration, priority ordering, and execution for agent operations

package agent

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
	// Note: domain import is used by the Hook interface in scriptHook
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// scriptHook wraps a script-defined hook implementation.
// It implements the domain.Hook interface and provides lifecycle
// callbacks for agent operations with priority ordering.
type scriptHook struct {
	id             string
	beforeGenerate func(ctx interface{}, messages interface{})
	afterGenerate  func(ctx interface{}, response interface{}, err interface{})
	beforeToolCall func(ctx interface{}, tool interface{}, params interface{})
	afterToolCall  func(ctx interface{}, tool interface{}, result interface{}, err interface{})
	priority       int
	enabled        bool
}

// BeforeGenerate implements domain.Hook interface.
// It's called before the agent generates a response,
// converting messages to script-compatible format.
func (h *scriptHook) BeforeGenerate(ctx context.Context, messages []llmdomain.Message) {
	if h.enabled && h.beforeGenerate != nil {
		// Convert messages to script-compatible format
		scriptMessages := make([]map[string]interface{}, len(messages))
		for i, msg := range messages {
			scriptMessages[i] = map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Content,
			}
		}
		h.beforeGenerate(ctx, scriptMessages)
	}
}

// AfterGenerate implements domain.Hook interface.
// It's called after the agent generates a response,
// providing the response or error to the script hook.
func (h *scriptHook) AfterGenerate(ctx context.Context, response llmdomain.Response, err error) {
	if h.enabled && h.afterGenerate != nil {
		// Convert response to script-compatible format
		scriptResponse := map[string]interface{}{
			"content": response.Content,
		}
		var scriptErr interface{}
		if err != nil {
			scriptErr = err.Error()
		}
		h.afterGenerate(ctx, scriptResponse, scriptErr)
	}
}

// BeforeToolCall implements domain.Hook interface.
// It's called before a tool is invoked, allowing scripts
// to inspect or modify tool parameters.
func (h *scriptHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	if h.enabled && h.beforeToolCall != nil {
		h.beforeToolCall(ctx, tool, params)
	}
}

// AfterToolCall implements domain.Hook interface.
// It's called after a tool completes execution,
// providing the result or error to the script hook.
func (h *scriptHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	if h.enabled && h.afterToolCall != nil {
		var scriptErr interface{}
		if err != nil {
			scriptErr = err.Error()
		}
		h.afterToolCall(ctx, tool, result, scriptErr)
	}
}

// HooksBridge bridges hook functionality to scripts.
// It manages hook registration, priority ordering, and execution
// for script-defined lifecycle callbacks in agent operations.
type HooksBridge struct {
	mu          sync.RWMutex
	initialized bool
	hooks       map[string]*scriptHook
}

// NewHooksBridge creates a new hooks bridge.
// It initializes an empty hook registry for managing
// script-defined lifecycle callbacks.
func NewHooksBridge() *HooksBridge {
	return &HooksBridge{
		hooks: make(map[string]*scriptHook),
	}
}

// GetID returns the bridge identifier.
// It implements the types.Bridge interface.
func (b *HooksBridge) GetID() string {
	return "agent_hooks"
}

// GetMetadata returns bridge metadata.
// It provides information about the bridge including
// name, version, description, and author.
func (b *HooksBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:        "agent_hooks",
		Version:     "1.0.0",
		Description: "Bridge for go-llms agent hook system",
		Author:      "go-llmspell",
	}
}

// Initialize sets up the bridge.
// It marks the bridge as initialized and ready for use.
func (b *HooksBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = true
	return nil
}

// Cleanup releases bridge resources.
// It clears all registered hooks and marks the bridge as uninitialized.
func (b *HooksBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = false
	b.hooks = make(map[string]*scriptHook)
	return nil
}

// IsInitialized checks if bridge is ready.
// It returns true if the bridge has been initialized and is ready for use.
func (b *HooksBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// Methods returns available bridge methods.
// It provides metadata about all hook management methods
// exposed to scripts through this bridge.
func (b *HooksBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		{
			Name:        "registerHook",
			Description: "Register a new hook with lifecycle callbacks",
			Parameters: []types.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
				{Name: "definition", Type: "object", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "unregisterHook",
			Description: "Remove a registered hook",
			Parameters: []types.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "listHooks",
			Description: "List all registered hooks",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "enableHook",
			Description: "Enable a disabled hook",
			Parameters: []types.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "disableHook",
			Description: "Disable a hook without removing it",
			Parameters: []types.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "getHookInfo",
			Description: "Get information about a specific hook",
			Parameters: []types.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "executeHooks",
			Description: "Execute hooks of a specific type",
			Parameters: []types.ParameterInfo{
				{Name: "type", Type: "string", Required: true},
				{Name: "context", Type: "object", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "clearHooks",
			Description: "Remove all registered hooks",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "number",
		},
	}
}

// ExecuteMethod runs a bridge method.
// It implements the types.Bridge interface, routing method calls
// to the appropriate hook management functions.
func (b *HooksBridge) ExecuteMethod(ctx context.Context, method string, args []types.ScriptValue) (types.ScriptValue, error) {
	if !b.IsInitialized() {
		return types.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
	}

	switch method {
	case "registerHook":
		id, err := b.registerHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewStringValue(id.(string)), nil
	case "unregisterHook":
		exists, err := b.unregisterHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewBoolValue(exists.(bool)), nil
	case "listHooks":
		hooks, err := b.listHooks(ctx)
		if err != nil {
			return nil, err
		}
		return convertHooksListToScriptValue(hooks.([]map[string]interface{})), nil
	case "enableHook":
		ok, err := b.enableHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewBoolValue(ok.(bool)), nil
	case "disableHook":
		ok, err := b.disableHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewBoolValue(ok.(bool)), nil
	case "getHookInfo":
		info, err := b.getHookInfo(ctx, args)
		if err != nil {
			return nil, err
		}
		return convertHookInfoToScriptValue(info.(map[string]interface{})), nil
	case "executeHooks":
		success, err := b.executeHooks(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewBoolValue(success.(bool)), nil
	case "clearHooks":
		count, err := b.clearHooks(ctx)
		if err != nil {
			return nil, err
		}
		return types.NewNumberValue(float64(count.(int))), nil
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

// registerHook registers a new hook with lifecycle callbacks.
// It expects an ID string and a definition object containing hook functions
// for beforeGenerate, afterGenerate, beforeToolCall, and afterToolCall.
func (b *HooksBridge) registerHook(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("registerHook requires id and definition arguments")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(types.StringValue).Value()

	if args[1] == nil || args[1].Type() != types.TypeObject {
		return nil, fmt.Errorf("definition must be an object")
	}
	definitionObj := args[1].(types.ObjectValue).Fields()
	definition := make(map[string]interface{})
	for k, v := range definitionObj {
		definition[k] = v.ToGo()
	}

	hook := &scriptHook{
		id:      id,
		enabled: true,
	}

	// Extract priority
	if priority, ok := definition["priority"].(int); ok {
		hook.priority = priority
	} else if priority, ok := definition["priority"].(float64); ok {
		hook.priority = int(priority)
	}

	// Extract hook functions
	if fn, ok := definition["beforeGenerate"].(func(interface{}, interface{})); ok {
		hook.beforeGenerate = fn
	}
	if fn, ok := definition["afterGenerate"].(func(interface{}, interface{}, interface{})); ok {
		hook.afterGenerate = fn
	}
	if fn, ok := definition["beforeToolCall"].(func(interface{}, interface{}, interface{})); ok {
		hook.beforeToolCall = fn
	}
	if fn, ok := definition["afterToolCall"].(func(interface{}, interface{}, interface{}, interface{})); ok {
		hook.afterToolCall = fn
	}

	b.mu.Lock()
	b.hooks[id] = hook
	b.mu.Unlock()

	return id, nil
}

// unregisterHook removes a registered hook.
// It expects an ID string and returns true if the hook existed and was removed,
// or false if the hook was not found.
func (b *HooksBridge) unregisterHook(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("unregisterHook requires id argument")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(types.StringValue).Value()

	b.mu.Lock()
	_, exists := b.hooks[id]
	if exists {
		delete(b.hooks, id)
	}
	b.mu.Unlock()

	return exists, nil
}

// listHooks returns all registered hooks sorted by priority.
// It returns an array of hook information objects containing
// ID, enabled status, and priority for each hook.
func (b *HooksBridge) listHooks(ctx context.Context) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(b.hooks))
	for _, hook := range b.hooks {
		info := map[string]interface{}{
			"id":       hook.id,
			"enabled":  hook.enabled,
			"priority": hook.priority,
		}
		result = append(result, info)
	}

	// Sort by priority (high to low)
	sort.Slice(result, func(i, j int) bool {
		return result[i]["priority"].(int) > result[j]["priority"].(int)
	})

	return result, nil
}

// enableHook enables a disabled hook.
// It expects an ID string and returns true if the hook was found and enabled,
// or false with an error if the hook was not found.
func (b *HooksBridge) enableHook(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("enableHook requires id argument")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(types.StringValue).Value()

	b.mu.Lock()
	defer b.mu.Unlock()

	if hook, exists := b.hooks[id]; exists {
		hook.enabled = true
		return true, nil
	}

	return false, fmt.Errorf("hook not found: %s", id)
}

// disableHook disables a hook without removing it.
// It expects an ID string and returns true if the hook was found and disabled,
// or false with an error if the hook was not found.
func (b *HooksBridge) disableHook(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("disableHook requires id argument")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(types.StringValue).Value()

	b.mu.Lock()
	defer b.mu.Unlock()

	if hook, exists := b.hooks[id]; exists {
		hook.enabled = false
		return true, nil
	}

	return false, fmt.Errorf("hook not found: %s", id)
}

// getHookInfo returns detailed information about a specific hook.
// It expects an ID string and returns an object with the hook's
// ID, enabled status, priority, and available callbacks.
func (b *HooksBridge) getHookInfo(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getHookInfo requires id argument")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(types.StringValue).Value()

	b.mu.RLock()
	defer b.mu.RUnlock()

	if hook, exists := b.hooks[id]; exists {
		return map[string]interface{}{
			"id":       hook.id,
			"enabled":  hook.enabled,
			"priority": hook.priority,
		}, nil
	}

	return nil, fmt.Errorf("hook not found: %s", id)
}

// executeHooks executes hooks of a specific type.
// It expects a hook type string (beforeGenerate, afterGenerate, beforeToolCall, afterToolCall)
// and a context object containing relevant data for the hook execution.
func (b *HooksBridge) executeHooks(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("executeHooks requires type and context arguments")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("type must be a string")
	}
	hookType := args[0].(types.StringValue).Value()

	if args[1] == nil || args[1].Type() != types.TypeObject {
		return nil, fmt.Errorf("context must be an object")
	}
	hookContextObj := args[1].(types.ObjectValue).Fields()
	hookContext := make(map[string]interface{})
	for k, v := range hookContextObj {
		hookContext[k] = v.ToGo()
	}

	// Get hooks sorted by priority
	sortedHooks := b.getSortedHooks()

	switch hookType {
	case "beforeGenerate":
		return b.executeBeforeGenerate(ctx, sortedHooks, hookContext)
	case "afterGenerate":
		return b.executeAfterGenerate(ctx, sortedHooks, hookContext)
	case "beforeToolCall":
		return b.executeBeforeToolCall(ctx, sortedHooks, hookContext)
	case "afterToolCall":
		return b.executeAfterToolCall(ctx, sortedHooks, hookContext)
	default:
		return nil, fmt.Errorf("unknown hook type: %s", hookType)
	}
}

// getSortedHooks returns all enabled hooks sorted by priority.
// Higher priority hooks are executed first.
func (b *HooksBridge) getSortedHooks() []*scriptHook {
	b.mu.RLock()
	defer b.mu.RUnlock()

	hooks := make([]*scriptHook, 0, len(b.hooks))
	for _, hook := range b.hooks {
		if hook.enabled {
			hooks = append(hooks, hook)
		}
	}

	// Sort by priority (high to low)
	sort.Slice(hooks, func(i, j int) bool {
		return hooks[i].priority > hooks[j].priority
	})

	return hooks
}

// executeBeforeGenerate executes all beforeGenerate hooks in priority order.
// It passes the messages context to each hook's beforeGenerate callback.
func (b *HooksBridge) executeBeforeGenerate(ctx context.Context, hooks []*scriptHook, hookContext map[string]interface{}) (interface{}, error) {
	messages := hookContext["messages"]
	for _, hook := range hooks {
		if hook.beforeGenerate != nil {
			hook.beforeGenerate(ctx, messages)
		}
	}
	return true, nil
}

// executeAfterGenerate executes all afterGenerate hooks in priority order.
// It passes the response and error context to each hook's afterGenerate callback.
func (b *HooksBridge) executeAfterGenerate(ctx context.Context, hooks []*scriptHook, hookContext map[string]interface{}) (interface{}, error) {
	response := hookContext["response"]
	err := hookContext["error"]
	for _, hook := range hooks {
		if hook.afterGenerate != nil {
			hook.afterGenerate(ctx, response, err)
		}
	}
	return true, nil
}

// executeBeforeToolCall executes all beforeToolCall hooks in priority order.
// It passes the tool name and parameters to each hook's beforeToolCall callback.
func (b *HooksBridge) executeBeforeToolCall(ctx context.Context, hooks []*scriptHook, hookContext map[string]interface{}) (interface{}, error) {
	tool := hookContext["tool"]
	params := hookContext["params"]
	for _, hook := range hooks {
		if hook.beforeToolCall != nil {
			hook.beforeToolCall(ctx, tool, params)
		}
	}
	return true, nil
}

// executeAfterToolCall executes all afterToolCall hooks in priority order.
// It passes the tool name, result, and error to each hook's afterToolCall callback.
func (b *HooksBridge) executeAfterToolCall(ctx context.Context, hooks []*scriptHook, hookContext map[string]interface{}) (interface{}, error) {
	tool := hookContext["tool"]
	result := hookContext["result"]
	err := hookContext["error"]
	for _, hook := range hooks {
		if hook.afterToolCall != nil {
			hook.afterToolCall(ctx, tool, result, err)
		}
	}
	return true, nil
}

// clearHooks removes all registered hooks.
// It returns the count of hooks that were removed.
func (b *HooksBridge) clearHooks(ctx context.Context) (interface{}, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	count := len(b.hooks)
	b.hooks = make(map[string]*scriptHook)

	return count, nil
}

// TypeMappings returns type mappings for the bridge.
// It defines how Go types are mapped to script types for
// hooks, hook information, and hook contexts.
func (b *HooksBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
		"Hook": {
			GoType:     "domain.Hook",
			ScriptType: "object",
		},
		"HookInfo": {
			GoType:     "map[string]interface{}",
			ScriptType: "object",
		},
		"HookType": {
			GoType:     "string",
			ScriptType: "string",
		},
		"HookContext": {
			GoType:     "map[string]interface{}",
			ScriptType: "object",
		},
	}
}

// RequiredPermissions returns permissions needed by this bridge.
// It requires process permissions for hook registration,
// execution, and management operations.
func (b *HooksBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionProcess,
			Resource:    "hook",
			Actions:     []string{"register", "execute", "manage"},
			Description: "Hook registration and execution",
		},
	}
}

// Validate checks if the bridge is properly configured.
// It currently performs no validation as the bridge has no
// required configuration.
func (b *HooksBridge) Validate() error {
	return nil
}

// GetDependencies returns bridge dependencies.
// The hooks bridge has no dependencies on other bridges.
func (b *HooksBridge) GetDependencies() []string {
	return []string{}
}

// RegisterWithEngine registers the bridge with a script types.
// The hooks bridge requires no special engine registration.
func (b *HooksBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// No special registration needed for this bridge
	return nil
}

// ValidateMethod validates method arguments before execution.
// It ensures that each method receives the correct number and
// types of arguments before processing.
func (b *HooksBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	switch name {
	case "registerHook":
		if len(args) < 2 {
			return fmt.Errorf("registerHook requires id and definition arguments")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return fmt.Errorf("id must be a string")
		}
		if args[1] == nil || args[1].Type() != types.TypeObject {
			return fmt.Errorf("definition must be an object")
		}
	case "unregisterHook", "enableHook", "disableHook", "getHookInfo":
		if len(args) < 1 {
			return fmt.Errorf("%s requires id argument", name)
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return fmt.Errorf("id must be a string")
		}
	case "executeHooks":
		if len(args) < 2 {
			return fmt.Errorf("executeHooks requires type and context arguments")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return fmt.Errorf("type must be a string")
		}
		if args[1] == nil || args[1].Type() != types.TypeObject {
			return fmt.Errorf("context must be an object")
		}
	case "listHooks", "clearHooks":
		// No arguments required
	default:
		return fmt.Errorf("unknown method: %s", name)
	}
	return nil
}

// Ensure HooksBridge implements the Bridge interface
var _ types.Bridge = (*HooksBridge)(nil)

// Helper functions for ScriptValue conversions

// convertHooksListToScriptValue converts a slice of hook information maps
// to a ScriptValue array for returning to scripts.
func convertHooksListToScriptValue(hooks []map[string]interface{}) types.ScriptValue {
	result := make([]types.ScriptValue, len(hooks))
	for i, hook := range hooks {
		result[i] = convertHookInfoToScriptValue(hook)
	}
	return types.NewArrayValue(result)
}

// convertHookInfoToScriptValue converts a hook information map
// to a ScriptValue object for returning to scripts.
func convertHookInfoToScriptValue(info map[string]interface{}) types.ScriptValue {
	result := make(map[string]types.ScriptValue)
	for k, v := range info {
		result[k] = types.ConvertToScriptValue(v)
	}
	return types.NewObjectValue(result)
}

// NOTE: Duplicate conversion function removed - using centralized types.ConvertToScriptValue() instead
