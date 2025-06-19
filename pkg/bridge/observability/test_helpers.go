package observability

import (
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// Helper functions to reduce repetitive ScriptValue creation
func sv(value interface{}) engine.ScriptValue {
	return testutils.InterfaceToScriptValue(value)
}

func svMap(m map[string]interface{}) engine.ScriptValue {
	return testutils.ObjectFromMap(m)
}

func svArray(values ...interface{}) engine.ScriptValue {
	return testutils.ArrayFromSlice(values)
}
