// ABOUTME: Adapter to ensure bridge.ToolBridge implements ToolBridgeInterface
// ABOUTME: This is mainly for documentation since ToolBridge already has all methods

package bridges

import (
	"github.com/lexlapax/go-llmspell/pkg/bridge"
)

// ToolBridgeAdapter is a compile-time check that bridge.ToolBridge implements ToolBridgeInterface
var _ ToolBridgeInterface = (*bridge.ToolBridge)(nil)
