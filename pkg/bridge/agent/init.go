// ABOUTME: Package initialization for agent bridge
// ABOUTME: Ensures tools are available by importing them

package agent

import (
	// Import tool packages to ensure they register themselves
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)