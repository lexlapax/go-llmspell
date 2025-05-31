// ABOUTME: Wrapper for tools bridge to integrate with Lua engine
// ABOUTME: Provides the Register method required by the engine bridge system

package bridges

import (
	"github.com/lexlapax/go-llmspell/pkg/bridge"
	lua "github.com/yuin/gopher-lua"
)

// ToolsBridgeWrapper wraps the tools bridge for Lua integration
type ToolsBridgeWrapper struct {
	toolBridge *bridge.ToolBridge
}

// NewToolsBridgeWrapper creates a new tools bridge wrapper
func NewToolsBridgeWrapper(tb *bridge.ToolBridge) *ToolsBridgeWrapper {
	return &ToolsBridgeWrapper{
		toolBridge: tb,
	}
}

// Register registers the tools module with the Lua state
func (tbw *ToolsBridgeWrapper) Register(L *lua.LState) error {
	return RegisterToolsModule(L, tbw.toolBridge)
}
