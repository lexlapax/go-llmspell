// ABOUTME: Tests for the engine-agnostic agent interface that provides lifecycle, metadata, and extension points
// ABOUTME: Ensures agents can work across all script engines (Lua, JavaScript, Tengo) with consistent behavior

package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAgent implements Agent interface for testing
type mockAgent struct {
	id           string
	name         string
	description  string
	version      string
	capabilities map[string]interface{}
	metadata     map[string]interface{}

	initCalled    bool
	runCalled     bool
	cleanupCalled bool

	initError    error
	runError     error
	cleanupError error

	runResult interface{}
}

func (m *mockAgent) ID() string {
	return m.id
}

func (m *mockAgent) Name() string {
	return m.name
}

func (m *mockAgent) Description() string {
	return m.description
}

func (m *mockAgent) Version() string {
	return m.version
}

func (m *mockAgent) Capabilities() map[string]interface{} {
	return m.capabilities
}

func (m *mockAgent) Metadata() map[string]interface{} {
	return m.metadata
}

func (m *mockAgent) Init(ctx context.Context) error {
	m.initCalled = true
	return m.initError
}

func (m *mockAgent) Run(ctx context.Context, input interface{}) (interface{}, error) {
	m.runCalled = true
	if m.runError != nil {
		return nil, m.runError
	}
	if m.runResult != nil {
		return m.runResult, nil
	}
	// Return default result if none specified
	return "default result", nil
}

func (m *mockAgent) Cleanup(ctx context.Context) error {
	m.cleanupCalled = true
	return m.cleanupError
}

func TestAgentInterface_Lifecycle(t *testing.T) {
	tests := []struct {
		name           string
		agent          *mockAgent
		wantInitErr    bool
		wantRunErr     bool
		wantCleanupErr bool
	}{
		{
			name: "successful lifecycle",
			agent: &mockAgent{
				id:          "test-agent-1",
				name:        "Test Agent",
				description: "A test agent",
				version:     "1.0.0",
				runResult:   "success",
			},
			wantInitErr:    false,
			wantRunErr:     false,
			wantCleanupErr: false,
		},
		{
			name: "init error",
			agent: &mockAgent{
				id:        "test-agent-2",
				initError: errors.New("init failed"),
			},
			wantInitErr: true,
		},
		{
			name: "run error",
			agent: &mockAgent{
				id:       "test-agent-3",
				runError: errors.New("run failed"),
			},
			wantInitErr: false,
			wantRunErr:  true,
		},
		{
			name: "cleanup error",
			agent: &mockAgent{
				id:           "test-agent-4",
				cleanupError: errors.New("cleanup failed"),
			},
			wantInitErr:    false,
			wantRunErr:     false,
			wantCleanupErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var agent Agent = tt.agent

			// Test Init
			err := agent.Init(ctx)
			if tt.wantInitErr {
				assert.Error(t, err)
				assert.True(t, tt.agent.initCalled)
				return
			}
			require.NoError(t, err)
			assert.True(t, tt.agent.initCalled)

			// Test Run
			result, err := agent.Run(ctx, "test input")
			if tt.wantRunErr {
				assert.Error(t, err)
				assert.True(t, tt.agent.runCalled)
			} else {
				require.NoError(t, err)
				assert.True(t, tt.agent.runCalled)
				if tt.agent.runResult != nil {
					assert.Equal(t, tt.agent.runResult, result)
				}
			}

			// Test Cleanup
			err = agent.Cleanup(ctx)
			if tt.wantCleanupErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.True(t, tt.agent.cleanupCalled)
		})
	}
}

func TestAgentInterface_Metadata(t *testing.T) {
	agent := &mockAgent{
		id:          "metadata-test",
		name:        "Metadata Test Agent",
		description: "Tests agent metadata",
		version:     "2.0.0",
		capabilities: map[string]interface{}{
			"streaming": true,
			"tools":     []string{"file_read", "file_write"},
			"maxTokens": 4096,
		},
		metadata: map[string]interface{}{
			"author":    "test",
			"license":   "MIT",
			"tags":      []string{"test", "example"},
			"createdAt": time.Now(),
		},
	}

	var a Agent = agent

	// Test metadata accessors
	assert.Equal(t, "metadata-test", a.ID())
	assert.Equal(t, "Metadata Test Agent", a.Name())
	assert.Equal(t, "Tests agent metadata", a.Description())
	assert.Equal(t, "2.0.0", a.Version())

	// Test capabilities
	caps := a.Capabilities()
	assert.NotNil(t, caps)
	assert.Equal(t, true, caps["streaming"])
	assert.Equal(t, []string{"file_read", "file_write"}, caps["tools"])
	assert.Equal(t, 4096, caps["maxTokens"])

	// Test metadata
	meta := a.Metadata()
	assert.NotNil(t, meta)
	assert.Equal(t, "test", meta["author"])
	assert.Equal(t, "MIT", meta["license"])
	assert.Equal(t, []string{"test", "example"}, meta["tags"])
}

func TestAgentInterface_ContextCancellation(t *testing.T) {
	t.Run("init cancelled", func(t *testing.T) {
		agent := &mockAgent{
			id:        "cancel-test-1",
			initError: context.Canceled,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		var a Agent = agent
		err := a.Init(ctx)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
	})

	t.Run("run cancelled", func(t *testing.T) {
		agent := &mockAgent{
			id:       "cancel-test-2",
			runError: context.Canceled,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		var a Agent = agent
		_, err := a.Run(ctx, nil)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
	})

	t.Run("cleanup cancelled", func(t *testing.T) {
		agent := &mockAgent{
			id:           "cancel-test-3",
			cleanupError: context.Canceled,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		var a Agent = agent
		err := a.Cleanup(ctx)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
	})
}

// customAgentImpl for extension test
type customAgentImpl struct {
	*mockAgent
	customValue string
}

func (c *customAgentImpl) CustomMethod() string {
	return c.customValue
}

func TestAgentInterface_ExtensionPoints(t *testing.T) {
	// Test that agents can be extended with custom interfaces
	type CustomAgent interface {
		Agent
		CustomMethod() string
	}

	agent := &customAgentImpl{
		mockAgent: &mockAgent{
			id:   "custom-agent",
			name: "Custom Agent",
		},
		customValue: "custom result",
	}

	// Should work as regular Agent
	var a Agent = agent
	assert.Equal(t, "custom-agent", a.ID())
	assert.Equal(t, "Custom Agent", a.Name())

	// Should also work as CustomAgent
	var ca CustomAgent = agent
	assert.Equal(t, "custom result", ca.CustomMethod())
}

func TestAgentInterface_EngineIndependence(t *testing.T) {
	// Test that agent interface doesn't depend on any specific engine
	agents := []Agent{
		&mockAgent{id: "lua-agent", metadata: map[string]interface{}{"engine": "lua"}},
		&mockAgent{id: "js-agent", metadata: map[string]interface{}{"engine": "javascript"}},
		&mockAgent{id: "tengo-agent", metadata: map[string]interface{}{"engine": "tengo"}},
	}

	ctx := context.Background()

	for _, agent := range agents {
		// All agents should work the same way regardless of engine
		err := agent.Init(ctx)
		assert.NoError(t, err)

		result, err := agent.Run(ctx, "test")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		err = agent.Cleanup(ctx)
		assert.NoError(t, err)

		// Check engine metadata
		meta := agent.Metadata()
		assert.Contains(t, meta, "engine")
	}
}
