// ABOUTME: Mock adapter for testing LuaEngine adapter management functionality
// ABOUTME: Provides simple mock implementations for testing adapter creation and retrieval

package gopherlua

// mockAdapter is a simple mock adapter for testing
type mockAdapter struct {
	id string
}

// CreateLuaModule creates a mock Lua module
func (ma *mockAdapter) CreateLuaModule() interface{} {
	return func() map[string]interface{} {
		return map[string]interface{}{
			"id":      ma.id,
			"test":    true,
			"methods": []string{"mock_method_1", "mock_method_2"},
		}
	}
}