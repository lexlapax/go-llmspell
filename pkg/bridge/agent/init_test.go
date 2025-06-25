// ABOUTME: Test initialization to ensure tools are loaded for all tests
// ABOUTME: Imports tool packages so they register themselves before tests run

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