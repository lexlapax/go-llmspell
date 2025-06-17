// ABOUTME: BridgeConverter handles conversion between Go bridge objects and Lua userdata
// ABOUTME: Provides metatable generation, method wrapping, type safety checks, and bridge type registry

package gopherlua

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

// BridgeConverter handles conversion of bridge objects to/from Lua
type BridgeConverter struct {
	mu                 sync.RWMutex
	registeredTypes    map[string]engine.Bridge
	primitiveConverter *PrimitiveConverter
}

// NewBridgeConverter creates a new bridge converter
func NewBridgeConverter() *BridgeConverter {
	return &BridgeConverter{
		registeredTypes:    make(map[string]engine.Bridge),
		primitiveConverter: NewPrimitiveConverter(),
	}
}

// Bridge to Lua conversion

// BridgeToLua converts a Go bridge object to Lua userdata with metatable
func (bc *BridgeConverter) BridgeToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	bridge, ok := value.(engine.Bridge)
	if !ok {
		return nil, fmt.Errorf("expected bridge object, got %T", value)
	}

	// Create userdata
	userdata := L.NewUserData()
	userdata.Value = bridge

	// Generate and set metatable
	metatable := bc.GenerateMetatable(L, bridge)
	L.SetMetatable(userdata, metatable)

	return userdata, nil
}

// GenerateMetatable creates a metatable for a bridge object with its methods
func (bc *BridgeConverter) GenerateMetatable(L *lua.LState, bridge engine.Bridge) *lua.LTable {
	metatable := L.NewTable()

	// Add bridge type information
	metatable.RawSetString("__type", lua.LString("bridge"))
	metatable.RawSetString("__bridge_id", lua.LString(bridge.GetID()))

	// Add bridge metadata
	metadata := bridge.GetMetadata()
	metatable.RawSetString("__bridge_name", lua.LString(metadata.Name))
	metatable.RawSetString("__bridge_version", lua.LString(metadata.Version))

	// Add __tostring method
	metatable.RawSetString("__tostring", L.NewFunction(func(L *lua.LState) int {
		userdata := L.CheckUserData(1)
		if bridge, ok := userdata.Value.(engine.Bridge); ok {
			result := fmt.Sprintf("Bridge<%s:%s>", bridge.GetID(), bridge.GetMetadata().Name)
			L.Push(lua.LString(result))
		} else {
			L.Push(lua.LString("Bridge<invalid>"))
		}
		return 1
	}))

	// Add __index method to handle method calls
	metatable.RawSetString("__index", L.NewFunction(func(L *lua.LState) int {
		userdata := L.CheckUserData(1)
		methodName := L.CheckString(2)

		bridge, ok := userdata.Value.(engine.Bridge)
		if !ok {
			L.RaiseError("invalid bridge object")
			return 0
		}

		// Find method in bridge
		for _, method := range bridge.Methods() {
			if method.Name == methodName {
				wrappedMethod := bc.WrapMethod(L, bridge, method)
				L.Push(wrappedMethod)
				return 1
			}
		}

		// Method not found
		L.Push(lua.LNil)
		return 1
	}))

	// Add all methods directly to metatable for easier access
	for _, method := range bridge.Methods() {
		wrappedMethod := bc.WrapMethod(L, bridge, method)
		metatable.RawSetString(method.Name, wrappedMethod)
	}

	return metatable
}

// WrapMethod wraps a bridge method for Lua calling
func (bc *BridgeConverter) WrapMethod(L *lua.LState, bridge engine.Bridge, methodInfo engine.MethodInfo) lua.LValue {
	return L.NewFunction(func(L *lua.LState) int {
		// Skip the first argument if it's the userdata (self)
		startArg := 1
		if L.GetTop() > 0 {
			if _, ok := L.Get(1).(*lua.LUserData); ok {
				startArg = 2
			}
		}

		// Collect arguments
		args := make([]lua.LValue, 0)
		for i := startArg; i <= L.GetTop(); i++ {
			args = append(args, L.Get(i))
		}

		// Validate method call
		err := bc.ValidateMethodCall(methodInfo.Name, args, methodInfo)
		if err != nil {
			L.RaiseError("method validation failed: %s", err.Error())
			return 0
		}

		// For now, return a placeholder since we're focusing on type conversion
		// In a real implementation, this would call the actual bridge method
		switch methodInfo.ReturnType {
		case "string":
			L.Push(lua.LString(fmt.Sprintf("result_of_%s", methodInfo.Name)))
		case "number":
			L.Push(lua.LNumber(42))
		case "boolean":
			L.Push(lua.LBool(true))
		default:
			L.Push(lua.LNil)
		}

		return 1
	})
}

// ValidateMethodCall validates the arguments for a method call
func (bc *BridgeConverter) ValidateMethodCall(methodName string, args []lua.LValue, methodInfo engine.MethodInfo) error {
	// Check argument count
	requiredCount := 0
	for _, param := range methodInfo.Parameters {
		if param.Required {
			requiredCount++
		}
	}

	if len(args) < requiredCount {
		return fmt.Errorf("method %s expected %d arguments, got %d", methodName, requiredCount, len(args))
	}

	// Validate argument types
	for i, param := range methodInfo.Parameters {
		if i >= len(args) {
			if param.Required {
				return fmt.Errorf("missing required parameter %s", param.Name)
			}
			continue
		}

		arg := args[i]
		if !bc.validateLuaType(arg, param.Type) {
			return fmt.Errorf("parameter %s expected %s, got %s", param.Name, param.Type, arg.Type().String())
		}
	}

	return nil
}

func (bc *BridgeConverter) validateLuaType(value lua.LValue, expectedType string) bool {
	switch expectedType {
	case "string":
		return value.Type() == lua.LTString
	case "number":
		return value.Type() == lua.LTNumber
	case "boolean":
		return value.Type() == lua.LTBool
	case "table":
		return value.Type() == lua.LTTable
	case "nil":
		return value.Type() == lua.LTNil
	case "function":
		return value.Type() == lua.LTFunction
	case "userdata":
		return value.Type() == lua.LTUserData
	default:
		// For unknown types, allow anything
		return true
	}
}

// From Lua conversion

// FromLua converts Lua userdata back to a Go bridge object
func (bc *BridgeConverter) FromLua(value lua.LValue) (interface{}, error) {
	userdata, ok := value.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("expected userdata, got %s", value.Type().String())
	}

	bridge, ok := userdata.Value.(engine.Bridge)
	if !ok {
		return nil, fmt.Errorf("userdata does not contain a valid bridge object, got %T", userdata.Value)
	}

	return bridge, nil
}

// Type safety and validation

// IsBridge checks if a value is a bridge object
func (bc *BridgeConverter) IsBridge(value interface{}) bool {
	_, ok := value.(engine.Bridge)
	return ok
}

// ValidateBridge validates that a bridge object is valid
func (bc *BridgeConverter) ValidateBridge(bridge interface{}) error {
	if bridge == nil {
		return fmt.Errorf("bridge cannot be nil")
	}

	bridgeObj, ok := bridge.(engine.Bridge)
	if !ok {
		return fmt.Errorf("object does not implement Bridge interface: %T", bridge)
	}

	if bridgeObj.GetID() == "" {
		return fmt.Errorf("bridge ID cannot be empty")
	}

	return nil
}

// IsValidBridgeUserData checks if userdata contains a valid bridge
func (bc *BridgeConverter) IsValidBridgeUserData(userdata *lua.LUserData) bool {
	if userdata == nil {
		return false
	}

	_, ok := userdata.Value.(engine.Bridge)
	return ok
}

// Bridge type registry

// RegisterBridgeType registers a bridge type for later lookup
func (bc *BridgeConverter) RegisterBridgeType(typeName string, bridge engine.Bridge) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if _, exists := bc.registeredTypes[typeName]; exists {
		return fmt.Errorf("bridge type %s is already registered", typeName)
	}

	bc.registeredTypes[typeName] = bridge
	return nil
}

// GetBridgeType retrieves a registered bridge type
func (bc *BridgeConverter) GetBridgeType(typeName string) (engine.Bridge, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	bridge, exists := bc.registeredTypes[typeName]
	return bridge, exists
}

// ListBridgeTypes returns all registered bridge type names
func (bc *BridgeConverter) ListBridgeTypes() []string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	types := make([]string, 0, len(bc.registeredTypes))
	for typeName := range bc.registeredTypes {
		types = append(types, typeName)
	}
	return types
}

// UnregisterBridgeType removes a bridge type from the registry
func (bc *BridgeConverter) UnregisterBridgeType(typeName string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if _, exists := bc.registeredTypes[typeName]; !exists {
		return fmt.Errorf("bridge type %s is not registered", typeName)
	}

	delete(bc.registeredTypes, typeName)
	return nil
}

// Helper methods for integration with existing type converter

// ExtractBridgeFromValue attempts to extract a bridge from any value
func (bc *BridgeConverter) ExtractBridgeFromValue(value interface{}) (engine.Bridge, bool) {
	// Direct bridge object
	if bridge, ok := value.(engine.Bridge); ok {
		return bridge, true
	}

	// Check if it's a pointer to a bridge
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		if bridge, ok := rv.Elem().Interface().(engine.Bridge); ok {
			return bridge, true
		}
	}

	return nil, false
}
