// ABOUTME: Wrapper for agents bridge to integrate with Lua engine
// ABOUTME: Provides the Register method required by the engine bridge system

package bridges

import (
	"github.com/lexlapax/go-llmspell/pkg/bridge"
	lua "github.com/yuin/gopher-lua"
)

// AgentsBridgeWrapper wraps the agents bridge for Lua integration
type AgentsBridgeWrapper struct {
	agentBridge bridge.AgentBridge
}

// NewAgentsBridgeWrapper creates a new agents bridge wrapper
func NewAgentsBridgeWrapper(ab bridge.AgentBridge) *AgentsBridgeWrapper {
	return &AgentsBridgeWrapper{
		agentBridge: ab,
	}
}

// Register registers the agents module with the Lua state
func (abw *AgentsBridgeWrapper) Register(L *lua.LState) error {
	return RegisterAgentsModule(L, abw.agentBridge)
}
