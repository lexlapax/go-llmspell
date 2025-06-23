// ABOUTME: Bridge stress tests to verify bridge performance and stability under heavy load
// ABOUTME: Tests concurrent bridge operations, initialization/cleanup cycles, and bridge method calls

package stress

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/bridge/agent"
	"github.com/lexlapax/go-llmspell/pkg/bridge/llm"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestBridgeConcurrentInitialization tests concurrent bridge initialization
func TestBridgeConcurrentInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const numGoroutines = 20
	const bridgesPerGoroutine = 5

	var wg sync.WaitGroup
	results := make(chan error, numGoroutines*bridgesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < bridgesPerGoroutine; j++ {
				// Test agent bridge
				agentBridge := agent.NewAgentBridge()
				ctx := context.Background()
				
				err := agentBridge.Initialize(ctx)
				if err != nil {
					results <- err
					continue
				}
				
				err = agentBridge.Cleanup(ctx)
				if err != nil {
					results <- err
					continue
				}
				
				results <- nil
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Count results
	successCount := 0
	errorCount := 0
	for result := range results {
		if result == nil {
			successCount++
		} else {
			errorCount++
			t.Logf("Bridge initialization error: %v", result)
		}
	}

	totalTests := numGoroutines * bridgesPerGoroutine
	t.Logf("Concurrent bridge initialization: %d successes, %d errors out of %d total", 
		successCount, errorCount, totalTests)

	assert.Equal(t, 0, errorCount, "All bridge initializations should succeed")
	assert.Equal(t, totalTests, successCount, "All bridge initializations should succeed")
}

// TestBridgeRepeatedInitializationCleanup tests repeated init/cleanup cycles
func TestBridgeRepeatedInitializationCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const numCycles = 100

	// Test agent bridge cycles
	agentBridge := agent.NewAgentBridge()
	ctx := context.Background()

	for i := 0; i < numCycles; i++ {
		err := agentBridge.Initialize(ctx)
		assert.NoError(t, err, "Agent bridge initialization %d failed", i)

		// Verify bridge is initialized
		assert.True(t, agentBridge.IsInitialized(), "Agent bridge should be initialized")

		err = agentBridge.Cleanup(ctx)
		assert.NoError(t, err, "Agent bridge cleanup %d failed", i)

		// Log progress
		if i%20 == 0 {
			t.Logf("Completed %d agent bridge init/cleanup cycles", i)
		}
	}

	// Test LLM bridge cycles
	llmBridge := llm.NewLLMBridge()
	
	for i := 0; i < numCycles; i++ {
		err := llmBridge.Initialize(ctx)
		assert.NoError(t, err, "LLM bridge initialization %d failed", i)

		// Verify bridge is initialized
		assert.True(t, llmBridge.IsInitialized(), "LLM bridge should be initialized")

		err = llmBridge.Cleanup(ctx)
		assert.NoError(t, err, "LLM bridge cleanup %d failed", i)

		// Log progress
		if i%20 == 0 {
			t.Logf("Completed %d LLM bridge init/cleanup cycles", i)
		}
	}

	t.Logf("Successfully completed %d initialization/cleanup cycles for each bridge type", numCycles)
}

// TestMockBridgeStress tests stress on mock bridges
func TestMockBridgeStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const numBridges = 50
	const operationsPerBridge = 20

	bridges := make([]*testutils.MockBridge, numBridges)
	
	// Create multiple mock bridges
	for i := 0; i < numBridges; i++ {
		bridge := testutils.NewMockBridge("stress-bridge-" + string(rune('A'+i%26)))
		bridges[i] = bridge
	}

	var wg sync.WaitGroup
	results := make(chan error, numBridges*operationsPerBridge)

	// Stress test each bridge concurrently
	for i, bridge := range bridges {
		wg.Add(1)
		go func(bridgeID int, b *testutils.MockBridge) {
			defer wg.Done()
			
			ctx := context.Background()
			
			for j := 0; j < operationsPerBridge; j++ {
				// Test initialization
				err := b.Initialize(ctx)
				if err != nil {
					results <- err
					continue
				}
				
				// Test metadata access
				metadata := b.GetMetadata()
				if metadata.Name == "" {
					results <- assert.AnError
					continue
				}
				
				// Test cleanup
				err = b.Cleanup(ctx)
				if err != nil {
					results <- err
					continue
				}
				
				results <- nil
			}
		}(i, bridge)
	}

	wg.Wait()
	close(results)

	// Count results
	successCount := 0
	errorCount := 0
	for result := range results {
		if result == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	totalOperations := numBridges * operationsPerBridge
	t.Logf("Mock bridge stress test: %d successes, %d errors out of %d total operations", 
		successCount, errorCount, totalOperations)

	assert.Equal(t, 0, errorCount, "All mock bridge operations should succeed")
	assert.Equal(t, totalOperations, successCount, "All mock bridge operations should succeed")
}

// TestBridgeMetadataStress tests repeated metadata access
func TestBridgeMetadataStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	bridges := []interface{}{
		agent.NewAgentBridge(),
		llm.NewLLMBridge(),
		testutils.NewMockBridge("metadata-test"),
	}

	const accessesPerBridge = 1000

	for i, bridge := range bridges {
		switch b := bridge.(type) {
		case *agent.AgentBridge:
			for j := 0; j < accessesPerBridge; j++ {
				id := b.GetID()
				assert.NotEmpty(t, id, "Agent bridge ID should not be empty")
				
				metadata := b.GetMetadata()
				assert.NotEmpty(t, metadata.Name, "Agent bridge metadata name should not be empty")
			}
			
		case *llm.LLMBridge:
			for j := 0; j < accessesPerBridge; j++ {
				id := b.GetID()
				assert.NotEmpty(t, id, "LLM bridge ID should not be empty")
				
				metadata := b.GetMetadata()
				assert.NotEmpty(t, metadata.Name, "LLM bridge metadata name should not be empty")
			}
			
		case *testutils.MockBridge:
			for j := 0; j < accessesPerBridge; j++ {
				id := b.GetID()
				assert.NotEmpty(t, id, "Mock bridge ID should not be empty")
				
				metadata := b.GetMetadata()
				assert.NotEmpty(t, metadata.Name, "Mock bridge metadata name should not be empty")
			}
		}
		
		t.Logf("Completed %d metadata accesses for bridge %d", accessesPerBridge, i)
	}
}

// TestBridgeMethodListingStress tests repeated method listing calls
func TestBridgeMethodListingStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create bridges that support method listing
	llmBridge := llm.NewLLMBridge()
	ctx := context.Background()
	err := llmBridge.Initialize(ctx)
	require.NoError(t, err)
	defer llmBridge.Cleanup(ctx)

	const methodListingCalls = 500

	// Test LLM bridge method listing
	for i := 0; i < methodListingCalls; i++ {
		methods := llmBridge.Methods()
		assert.NotEmpty(t, methods, "LLM bridge should have methods")
		
		// Verify method structure
		for _, method := range methods {
			assert.NotEmpty(t, method.Name, "Method name should not be empty")
			assert.NotEmpty(t, method.Description, "Method description should not be empty")
		}
		
		if i%100 == 0 {
			t.Logf("Completed %d method listing calls", i)
		}
	}

	t.Logf("Successfully completed %d method listing stress tests", methodListingCalls)
}

// TestBridgeConcurrentAccess tests concurrent access to bridge methods
func TestBridgeConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	agentBridge := agent.NewAgentBridge()
	llmBridge := llm.NewLLMBridge()
	
	ctx := context.Background()
	require.NoError(t, agentBridge.Initialize(ctx))
	require.NoError(t, llmBridge.Initialize(ctx))
	
	defer func() {
		agentBridge.Cleanup(ctx)
		llmBridge.Cleanup(ctx)
	}()

	const numGoroutines = 20
	const operationsPerGoroutine = 50

	var wg sync.WaitGroup
	results := make(chan error, numGoroutines*operationsPerGoroutine*2) // *2 for both bridges

	// Test concurrent access to multiple bridges
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				// Test agent bridge
				id := agentBridge.GetID()
				if id != "agent" {
					results <- assert.AnError
				} else {
					results <- nil
				}
				
				// Test LLM bridge
				id = llmBridge.GetID()
				if id != "llm" {
					results <- assert.AnError
				} else {
					results <- nil
				}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Count results
	successCount := 0
	errorCount := 0
	for result := range results {
		if result == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	totalOperations := numGoroutines * operationsPerGoroutine * 2
	t.Logf("Concurrent bridge access: %d successes, %d errors out of %d total operations", 
		successCount, errorCount, totalOperations)

	assert.Equal(t, 0, errorCount, "All concurrent bridge accesses should succeed")
	assert.Equal(t, totalOperations, successCount, "All concurrent bridge accesses should succeed")
}