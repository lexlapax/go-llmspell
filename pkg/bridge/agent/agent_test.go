// ABOUTME: Tests for the agent bridge that exposes go-llms agent functionality to scripts
// ABOUTME: Verifies agent creation, configuration, tool registration, and execution bridging

package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockScriptEngine implements engine.ScriptEngine for testing
type MockScriptEngine struct {
	bridges map[string]engine.Bridge
}

func NewMockScriptEngine() *MockScriptEngine {
	return &MockScriptEngine{
		bridges: make(map[string]engine.Bridge),
	}
}

func (m *MockScriptEngine) RegisterBridge(bridge engine.Bridge) error {
	m.bridges[bridge.GetID()] = bridge
	return nil
}

func (m *MockScriptEngine) UnregisterBridge(id string) error {
	delete(m.bridges, id)
	return nil
}

func (m *MockScriptEngine) GetBridge(id string) (engine.Bridge, error) {
	bridge, ok := m.bridges[id]
	if !ok {
		return nil, fmt.Errorf("bridge %s not found", id)
	}
	return bridge, nil
}

func (m *MockScriptEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockScriptEngine) ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockScriptEngine) Name() string                     { return "mock" }
func (m *MockScriptEngine) Version() string                  { return "1.0.0" }
func (m *MockScriptEngine) FileExtensions() []string         { return []string{".mock"} }
func (m *MockScriptEngine) Features() []engine.EngineFeature { return nil }

func (m *MockScriptEngine) Initialize(config engine.EngineConfig) error          { return nil }
func (m *MockScriptEngine) Shutdown() error                                      { return nil }
func (m *MockScriptEngine) SetMemoryLimit(bytes int64) error                     { return nil }
func (m *MockScriptEngine) SetTimeout(duration time.Duration) error              { return nil }
func (m *MockScriptEngine) SetResourceLimits(limits engine.ResourceLimits) error { return nil }
func (m *MockScriptEngine) GetMetrics() engine.EngineMetrics                     { return engine.EngineMetrics{} }
func (m *MockScriptEngine) CreateContext(options engine.ContextOptions) (engine.ScriptContext, error) {
	return nil, nil
}
func (m *MockScriptEngine) DestroyContext(ctx engine.ScriptContext) error { return nil }
func (m *MockScriptEngine) ExecuteScript(ctx context.Context, script string, options engine.ExecutionOptions) (*engine.ExecutionResult, error) {
	return nil, nil
}
func (m *MockScriptEngine) ToNative(scriptValue interface{}) (interface{}, error) {
	return scriptValue, nil
}
func (m *MockScriptEngine) FromNative(goValue interface{}) (interface{}, error) {
	return goValue, nil
}
func (m *MockScriptEngine) ListBridges() []string {
	var ids []string
	for id := range m.bridges {
		ids = append(ids, id)
	}
	return ids
}

// Task 1.4.11.1: Engine Event Bus
func (m *MockScriptEngine) GetEventBus() engine.EventBus {
	return engine.NewDefaultEventBus()
}

// Task 1.4.11.2: Type Conversion Registry
func (m *MockScriptEngine) RegisterTypeConverter(fromType, toType string, converter engine.TypeConverterFunc) error {
	return nil
}

func (m *MockScriptEngine) GetTypeRegistry() engine.TypeRegistry {
	return engine.NewDefaultTypeRegistry()
}

// Task 1.4.11.3: Engine Profiling
func (m *MockScriptEngine) EnableProfiling(config engine.ProfilingConfig) error {
	return nil
}

func (m *MockScriptEngine) DisableProfiling() error {
	return nil
}

func (m *MockScriptEngine) GetProfilingReport() (*engine.ProfilingReport, error) {
	return &engine.ProfilingReport{}, nil
}

// Task 1.4.11.4: Engine API Export
func (m *MockScriptEngine) ExportAPI(format engine.ExportFormat) ([]byte, error) {
	return []byte("{}"), nil
}

func (m *MockScriptEngine) GenerateClientLibrary(language string, options engine.ClientLibraryOptions) ([]byte, error) {
	return []byte("{}"), nil
}

func TestNewAgentBridge(t *testing.T) {
	bridge := NewAgentBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "agent", bridge.GetID())
}

func TestAgentBridgeMetadata(t *testing.T) {
	bridge := NewAgentBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "agent", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "state serialization")
	assert.Contains(t, metadata.Description, "event replay")
	assert.Contains(t, metadata.Description, "performance profiling")
	assert.NotEmpty(t, metadata.Author)
	assert.NotEmpty(t, metadata.License)
}

func TestAgentBridgeInitialization(t *testing.T) {
	tests := []struct {
		name    string
		bridge  *AgentBridge
		wantErr bool
	}{
		{
			name:    "successful initialization",
			bridge:  NewAgentBridge(),
			wantErr: false,
		},
		{
			name:    "double initialization",
			bridge:  &AgentBridge{initialized: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bridge.Initialize(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.bridge.IsInitialized())
			}
		})
	}
}

func TestAgentBridgeMethods(t *testing.T) {
	bridge := NewAgentBridge()
	methods := bridge.Methods()

	// Essential agent methods
	expectedMethods := []string{
		"createAgent",
		"createLLMAgent",
		"registerTool",
		"runAgent",
		"runAgentAsync",
		"addSubAgent",
		"getAgentState",
		"setAgentState",
		"listAgents",
		"getAgent",
		"removeAgent",
		// State serialization methods (v2.0.0)
		"exportAgentState",
		"importAgentState",
		"saveAgentSnapshot",
		"loadAgentSnapshot",
		"listAgentSnapshots",
		"deleteAgentSnapshot",
		// Event replay methods (v2.0.0)
		"replayAgentEvents",
		"startEventRecording",
		"stopEventRecording",
		"getEventHistory",
		"clearEventHistory",
		// Performance profiling methods (v2.0.0)
		"startAgentProfiling",
		"stopAgentProfiling",
		"getAgentPerformanceReport",
		"clearAgentProfilingData",
		"exportAgentProfilingData",
		"setAgentProfilingConfig",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Missing expected method: %s", expected)
	}

	// Verify method details
	for _, method := range methods {
		assert.NotEmpty(t, method.Description)
		assert.NotEmpty(t, method.ReturnType)

		// Check specific methods have correct parameters
		switch method.Name {
		case "createAgent":
			assert.GreaterOrEqual(t, len(method.Parameters), 2) // id, config
		case "registerTool":
			assert.GreaterOrEqual(t, len(method.Parameters), 2) // agentID, tool
		case "runAgent":
			assert.GreaterOrEqual(t, len(method.Parameters), 2) // agentID, input
		}
	}
}

func TestAgentBridgeTypeMappings(t *testing.T) {
	bridge := NewAgentBridge()
	mappings := bridge.TypeMappings()

	expectedTypes := []string{
		"Agent",
		"Tool",
		"State",
		"AgentConfig",
		"LLMConfig",
		"AgentType",
		"AgentEvent",
		"Message",
		"Artifact",
	}

	for _, typeName := range expectedTypes {
		mapping, exists := mappings[typeName]
		assert.True(t, exists, "Missing type mapping for %s", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestAgentBridgeRequiredPermissions(t *testing.T) {
	bridge := NewAgentBridge()
	permissions := bridge.RequiredPermissions()

	assert.NotEmpty(t, permissions)

	// Should require at least agent access permission
	hasAgentPermission := false
	for _, perm := range permissions {
		if perm.Type == engine.PermissionNetwork && perm.Resource == "agent" {
			hasAgentPermission = true
			assert.Contains(t, perm.Actions, "create")
			assert.Contains(t, perm.Actions, "execute")
		}
	}
	assert.True(t, hasAgentPermission, "Missing agent permission")
}

func TestAgentBridgeValidateMethod(t *testing.T) {
	bridge := NewAgentBridge()

	tests := []struct {
		name    string
		method  string
		args    []interface{}
		wantErr bool
	}{
		{
			name:    "valid createAgent",
			method:  "createAgent",
			args:    []interface{}{"test-agent", map[string]interface{}{"type": "llm"}},
			wantErr: false,
		},
		{
			name:    "createAgent missing args",
			method:  "createAgent",
			args:    []interface{}{},
			wantErr: false, // Validation is delegated to engine
		},
		{
			name:    "unknown method",
			method:  "unknownMethod",
			args:    []interface{}{},
			wantErr: false, // Validation is delegated to engine
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgentBridgeEngineRegistration(t *testing.T) {
	bridge := NewAgentBridge()
	engine := NewMockScriptEngine()

	err := bridge.RegisterWithEngine(engine)
	require.NoError(t, err)

	// Verify bridge was registered
	registered, err := engine.GetBridge("agent")
	assert.NoError(t, err)
	assert.Equal(t, bridge, registered)
}

func TestAgentBridgeCleanup(t *testing.T) {
	bridge := NewAgentBridge()

	// Initialize first
	err := bridge.Initialize(context.Background())
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Cleanup
	err = bridge.Cleanup(context.Background())
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestAgentBridgeConcurrentAccess(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Concurrent operations
	done := make(chan bool, 3)

	// Reader 1
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.IsInitialized()
			_ = bridge.GetID()
			_ = bridge.Methods()
		}
		done <- true
	}()

	// Reader 2
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.TypeMappings()
			_ = bridge.RequiredPermissions()
		}
		done <- true
	}()

	// Writer
	go func() {
		for i := 0; i < 50; i++ {
			_ = bridge.Initialize(ctx)
			_ = bridge.Cleanup(ctx)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestAgentBridge_StateSerializationMethods tests that state serialization methods are properly registered
func TestAgentBridge_StateSerializationMethods(t *testing.T) {
	bridge := NewAgentBridge()
	methods := bridge.Methods()

	// Create a map for quick lookup
	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	// Test that all state serialization methods are present
	stateSerializationMethods := []string{
		"exportAgentState",
		"importAgentState",
		"saveAgentSnapshot",
		"loadAgentSnapshot",
		"listAgentSnapshots",
		"deleteAgentSnapshot",
	}

	for _, methodName := range stateSerializationMethods {
		t.Run(fmt.Sprintf("method_%s_registered", methodName), func(t *testing.T) {
			assert.True(t, methodMap[methodName], "Method %s should be registered", methodName)
		})
	}

	// Test method signatures for key state serialization methods
	for _, method := range methods {
		switch method.Name {
		case "exportAgentState":
			assert.GreaterOrEqual(t, len(method.Parameters), 2, "exportAgentState should have at least 2 parameters (agentID, format)")
			assert.NotEmpty(t, method.Description, "exportAgentState should have a description")
		case "saveAgentSnapshot":
			assert.GreaterOrEqual(t, len(method.Parameters), 2, "saveAgentSnapshot should have at least 2 parameters (agentID, snapshotID)")
			assert.NotEmpty(t, method.Description, "saveAgentSnapshot should have a description")
		}
	}
}

// TestAgentBridge_EventReplayMethods tests that event replay methods are properly registered
func TestAgentBridge_EventReplayMethods(t *testing.T) {
	bridge := NewAgentBridge()
	methods := bridge.Methods()

	// Create a map for quick lookup
	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	// Test that all event replay methods are present
	eventReplayMethods := []string{
		"replayAgentEvents",
		"startEventRecording",
		"stopEventRecording",
		"getEventHistory",
		"clearEventHistory",
	}

	for _, methodName := range eventReplayMethods {
		t.Run(fmt.Sprintf("method_%s_registered", methodName), func(t *testing.T) {
			assert.True(t, methodMap[methodName], "Method %s should be registered", methodName)
		})
	}

	// Test method signatures for key event replay methods
	for _, method := range methods {
		switch method.Name {
		case "replayAgentEvents":
			assert.GreaterOrEqual(t, len(method.Parameters), 2, "replayAgentEvents should have at least 2 parameters (agentID, speed)")
			assert.NotEmpty(t, method.Description, "replayAgentEvents should have a description")
		case "getEventHistory":
			assert.GreaterOrEqual(t, len(method.Parameters), 2, "getEventHistory should have at least 2 parameters (agentID, limit)")
			assert.NotEmpty(t, method.Description, "getEventHistory should have a description")
		}
	}
}

// TestAgentBridge_PerformanceProfilingMethods tests that performance profiling methods are properly registered
func TestAgentBridge_PerformanceProfilingMethods(t *testing.T) {
	bridge := NewAgentBridge()
	methods := bridge.Methods()

	// Create a map for quick lookup
	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	// Test that all performance profiling methods are present
	performanceProfilingMethods := []string{
		"startAgentProfiling",
		"stopAgentProfiling",
		"getAgentPerformanceReport",
		"clearAgentProfilingData",
		"exportAgentProfilingData",
		"setAgentProfilingConfig",
	}

	for _, methodName := range performanceProfilingMethods {
		t.Run(fmt.Sprintf("method_%s_registered", methodName), func(t *testing.T) {
			assert.True(t, methodMap[methodName], "Method %s should be registered", methodName)
		})
	}

	// Test method signatures for key performance profiling methods
	for _, method := range methods {
		switch method.Name {
		case "getAgentPerformanceReport":
			assert.GreaterOrEqual(t, len(method.Parameters), 1, "getAgentPerformanceReport should have at least 1 parameter (agentID)")
			assert.NotEmpty(t, method.Description, "getAgentPerformanceReport should have a description")
		case "exportAgentProfilingData":
			assert.GreaterOrEqual(t, len(method.Parameters), 2, "exportAgentProfilingData should have at least 2 parameters (agentID, format)")
			assert.NotEmpty(t, method.Description, "exportAgentProfilingData should have a description")
		}
	}
}
