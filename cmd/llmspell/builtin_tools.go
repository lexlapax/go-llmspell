// ABOUTME: Imports built-in tools from go-llms to make them available in scripts
// ABOUTME: Uses direct imports to avoid build tag import cycle issues

package main

import (
	// Import all built-in tool packages to register them
	// This avoids the import cycle issue when using -tags tools
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)