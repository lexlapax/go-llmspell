// ABOUTME: Bridge for go-llms agent hook system enabling script-based lifecycle hooks
// ABOUTME: Provides hook registration, priority ordering, and execution for agent operations

package agent

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	// Note: domain import is used by the Hook interface in scriptHook
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// scriptHook wraps a script-defined hook implementation
type scriptHook struct {
	id             string
	beforeGenerate func(ctx interface{}, messages interface{})
	afterGenerate  func(ctx interface{}, response interface{}, err interface{})
	beforeToolCall func(ctx interface{}, tool interface{}, params interface{})
	afterToolCall  func(ctx interface{}, tool interface{}, result interface{}, err interface{})
	priority       int
	enabled        bool
}

// Implement domain.Hook interface
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

func (h *scriptHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	if h.enabled && h.beforeToolCall != nil {
		h.beforeToolCall(ctx, tool, params)
	}
}

func (h *scriptHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	if h.enabled && h.afterToolCall != nil {
		var scriptErr interface{}
		if err != nil {
			scriptErr = err.Error()
		}
		h.afterToolCall(ctx, tool, result, scriptErr)
	}
}

// HooksBridge bridges hook functionality to scripts
type HooksBridge struct {
	mu          sync.RWMutex
	initialized bool
	hooks       map[string]*scriptHook
}

// NewHooksBridge creates a new hooks bridge
func NewHooksBridge() *HooksBridge {
	return &HooksBridge{
		hooks: make(map[string]*scriptHook),
	}
}

// GetID returns the bridge identifier
func (b *HooksBridge) GetID() string {
	return "hooks"
}

// GetMetadata returns bridge metadata
func (b *HooksBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "Hooks Bridge",
		Version:     "1.0.0",
		Description: "Bridge for go-llms agent hook system",
		Author:      "go-llmspell",
	}
}

// Initialize sets up the bridge
func (b *HooksBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = true
	return nil
}

// Cleanup releases bridge resources
func (b *HooksBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = false
	b.hooks = make(map[string]*scriptHook)
	return nil
}

// IsInitialized checks if bridge is ready
func (b *HooksBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// Methods returns available bridge methods
func (b *HooksBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "registerHook",
			Description: "Register a new hook with lifecycle callbacks",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
				{Name: "definition", Type: "object", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "unregisterHook",
			Description: "Remove a registered hook",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "listHooks",
			Description: "List all registered hooks",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "enableHook",
			Description: "Enable a disabled hook",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "disableHook",
			Description: "Disable a hook without removing it",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "getHookInfo",
			Description: "Get information about a specific hook",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "executeHooks",
			Description: "Execute hooks of a specific type",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Required: true},
				{Name: "context", Type: "object", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "clearHooks",
			Description: "Remove all registered hooks",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "number",
		},
	}
}

// ExecuteMethod runs a bridge method
func (b *HooksBridge) ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if !b.IsInitialized() {
		return engine.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
	}

	switch method {
	case "registerHook":
		id, err := b.registerHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewStringValue(id.(string)), nil
	case "unregisterHook":
		exists, err := b.unregisterHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewBoolValue(exists.(bool)), nil
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
		return engine.NewBoolValue(ok.(bool)), nil
	case "disableHook":
		ok, err := b.disableHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewBoolValue(ok.(bool)), nil
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
		return engine.NewBoolValue(success.(bool)), nil
	case "clearHooks":
		count, err := b.clearHooks(ctx)
		if err != nil {
			return nil, err
		}
		return engine.NewNumberValue(float64(count.(int))), nil
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

func (b *HooksBridge) registerHook(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("registerHook requires id and definition arguments")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeObject {
		return nil, fmt.Errorf("definition must be an object")
	}
	definitionObj := args[1].(engine.ObjectValue).Fields()
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

func (b *HooksBridge) unregisterHook(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("unregisterHook requires id argument")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(engine.StringValue).Value()

	b.mu.Lock()
	_, exists := b.hooks[id]
	if exists {
		delete(b.hooks, id)
	}
	b.mu.Unlock()

	return exists, nil
}

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

func (b *HooksBridge) enableHook(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("enableHook requires id argument")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(engine.StringValue).Value()

	b.mu.Lock()
	defer b.mu.Unlock()

	if hook, exists := b.hooks[id]; exists {
		hook.enabled = true
		return true, nil
	}

	return false, fmt.Errorf("hook not found: %s", id)
}

func (b *HooksBridge) disableHook(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("disableHook requires id argument")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(engine.StringValue).Value()

	b.mu.Lock()
	defer b.mu.Unlock()

	if hook, exists := b.hooks[id]; exists {
		hook.enabled = false
		return true, nil
	}

	return false, fmt.Errorf("hook not found: %s", id)
}

func (b *HooksBridge) getHookInfo(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getHookInfo requires id argument")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("id must be a string")
	}
	id := args[0].(engine.StringValue).Value()

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

func (b *HooksBridge) executeHooks(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("executeHooks requires type and context arguments")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("type must be a string")
	}
	hookType := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be an object")
	}
	hookContextObj := args[1].(engine.ObjectValue).Fields()
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

func (b *HooksBridge) executeBeforeGenerate(ctx context.Context, hooks []*scriptHook, hookContext map[string]interface{}) (interface{}, error) {
	messages := hookContext["messages"]
	for _, hook := range hooks {
		if hook.beforeGenerate != nil {
			hook.beforeGenerate(ctx, messages)
		}
	}
	return true, nil
}

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

func (b *HooksBridge) clearHooks(ctx context.Context) (interface{}, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	count := len(b.hooks)
	b.hooks = make(map[string]*scriptHook)

	return count, nil
}

// TypeMappings returns type mappings for the bridge
func (b *HooksBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
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

// RequiredPermissions returns permissions needed by this bridge
func (b *HooksBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionProcess,
			Resource:    "hook",
			Actions:     []string{"register", "execute", "manage"},
			Description: "Hook registration and execution",
		},
	}
}

// Validate checks if the bridge is properly configured
func (b *HooksBridge) Validate() error {
	return nil
}

// GetDependencies returns bridge dependencies
func (b *HooksBridge) GetDependencies() []string {
	return []string{}
}

// RegisterWithEngine registers the bridge with a script engine
func (b *HooksBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	// No special registration needed for this bridge
	return nil
}

// ValidateMethod validates method arguments before execution
func (b *HooksBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	switch name {
	case "registerHook":
		if len(args) < 2 {
			return fmt.Errorf("registerHook requires id and definition arguments")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return fmt.Errorf("id must be a string")
		}
		if args[1] == nil || args[1].Type() != engine.TypeObject {
			return fmt.Errorf("definition must be an object")
		}
	case "unregisterHook", "enableHook", "disableHook", "getHookInfo":
		if len(args) < 1 {
			return fmt.Errorf("%s requires id argument", name)
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return fmt.Errorf("id must be a string")
		}
	case "executeHooks":
		if len(args) < 2 {
			return fmt.Errorf("executeHooks requires type and context arguments")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return fmt.Errorf("type must be a string")
		}
		if args[1] == nil || args[1].Type() != engine.TypeObject {
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
var _ engine.Bridge = (*HooksBridge)(nil)

// Helper functions for ScriptValue conversions
func convertHooksListToScriptValue(hooks []map[string]interface{}) engine.ScriptValue {
	result := make([]engine.ScriptValue, len(hooks))
	for i, hook := range hooks {
		result[i] = convertHookInfoToScriptValue(hook)
	}
	return engine.NewArrayValue(result)
}

func convertHookInfoToScriptValue(info map[string]interface{}) engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)
	for k, v := range info {
		result[k] = engine.ConvertToScriptValue(v)
	}
	return engine.NewObjectValue(result)
}


// NOTE: Duplicate conversion function removed - using centralized engine.ConvertToScriptValue() instead
