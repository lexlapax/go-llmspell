package agent

import (
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// Helper functions to reduce repetitive ScriptValue creation
func sv(value interface{}) types.ScriptValue {
	return testutils.InterfaceToScriptValue(value)
}

func svMap(m map[string]interface{}) types.ScriptValue {
	return testutils.ObjectFromMap(m)
}

func svArray(values ...interface{}) types.ScriptValue {
	return testutils.ArrayFromSlice(values)
}
