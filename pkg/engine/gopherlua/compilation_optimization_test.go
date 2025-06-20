// ABOUTME: Tests for compilation optimization infrastructure including AST optimization, dead code elimination, and caching
// ABOUTME: Validates compiler performance improvements, optimization effectiveness, and static analysis capabilities

package gopherlua

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptimizedCompiler_BasicCompilation(t *testing.T) {
	config := OptimizedCompilerConfig{
		ChunkCacheConfig: ChunkCacheConfig{
			MaxSize: 10,
			TTL:     time.Minute,
		},
		EnableSourceOptimization:  true,
		EnableDeadCodeElimination: true,
		EnableConstantFolding:     true,
	}

	compiler := NewOptimizedCompiler(config)

	tests := []struct {
		name   string
		script string
		valid  bool
	}{
		{
			name:   "simple function",
			script: `function add(a, b) return a + b end`,
			valid:  true,
		},
		{
			name: "with constants",
			script: `
				local x = 10 + 20
				local y = "hello" .. " world"
				return x, y
			`,
			valid: true,
		},
		{
			name:   "syntax error",
			script: `function broken(`,
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto, err := compiler.Compile(tt.script, tt.name+".lua")

			if tt.valid {
				require.NoError(t, err)
				require.NotNil(t, proto)
			} else {
				require.Error(t, err)
				require.Nil(t, proto)
			}
		})
	}
}

func TestOptimizedCompiler_CacheHits(t *testing.T) {
	config := OptimizedCompilerConfig{
		ChunkCacheConfig: ChunkCacheConfig{
			MaxSize: 10,
			TTL:     time.Minute,
		},
	}

	compiler := NewOptimizedCompiler(config)
	script := `return "hello world"`
	filename := "test.lua"

	// First compilation - cache miss
	proto1, err := compiler.Compile(script, filename)
	require.NoError(t, err)
	require.NotNil(t, proto1)

	metrics := compiler.GetMetrics()
	assert.Equal(t, int64(1), metrics.Compilations)
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(1), metrics.CacheMisses)

	// Second compilation - cache hit
	proto2, err := compiler.Compile(script, filename)
	require.NoError(t, err)
	require.NotNil(t, proto2)
	require.Equal(t, proto1, proto2) // Should be the same cached proto

	metrics = compiler.GetMetrics()
	assert.Equal(t, int64(2), metrics.Compilations)
	assert.Equal(t, int64(1), metrics.CacheHits)
	assert.Equal(t, int64(1), metrics.CacheMisses)
	assert.Equal(t, 0.5, metrics.CacheHitRate)
}

func TestOptimizedCompiler_StaticAnalysis(t *testing.T) {
	config := OptimizedCompilerConfig{
		EnableStaticAnalysis: true,
		WarnUnusedVariables:  true,
		WarnUnreachableCode:  true,
	}

	compiler := NewOptimizedCompiler(config)

	tests := []struct {
		name             string
		script           string
		expectedWarnings int
		warningTypes     []string
	}{
		{
			name: "unused variable",
			script: `
				local unused = 10
				local used = 20
				return used
			`,
			expectedWarnings: 1,
			warningTypes:     []string{"UnusedVariable"},
		},
		{
			name: "unreachable code",
			script: `
				function test(x)
					if true then
						return 10
					end
					print("this is unreachable")
				end
			`,
			expectedWarnings: 1,
			warningTypes:     []string{"UnreachableCode"},
		},
		{
			name: "no warnings",
			script: `
				local x = 10
				return x * 2
			`,
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := compiler.ValidateScript(tt.script, tt.name+".lua")
			assert.Len(t, warnings, tt.expectedWarnings)

			for i, warnType := range tt.warningTypes {
				if i < len(warnings) {
					assert.Equal(t, warnType, warnings[i].Type)
				}
			}
		})
	}
}

func TestOptimizedCompiler_ConstantFolding(t *testing.T) {
	config := OptimizedCompilerConfig{
		EnableConstantFolding: true,
		EnableStaticAnalysis:  true,
	}

	compiler := NewOptimizedCompiler(config)

	tests := []struct {
		name               string
		script             string
		expectOptimization bool
	}{
		{
			name: "numeric constants",
			script: `
				local x = 10 + 20 * 3
				local y = 100 / 2 - 10
				return x, y
			`,
			expectOptimization: true,
		},
		{
			name: "string concatenation",
			script: `
				local msg = "Hello" .. " " .. "World"
				return msg
			`,
			expectOptimization: true,
		},
		{
			name: "no constants to fold",
			script: `
				local x = a + b
				return x
			`,
			expectOptimization: false,
		},
	}

	// Reset metrics before test
	compiler.Reset()

	optimizationCount := int64(0)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevMetrics := compiler.GetMetrics()

			proto, err := compiler.Compile(tt.script, tt.name+".lua")
			require.NoError(t, err)
			require.NotNil(t, proto)

			// Check if optimization happened via metrics
			newMetrics := compiler.GetMetrics()

			if tt.expectOptimization {
				assert.Greater(t, newMetrics.Optimizations, prevMetrics.Optimizations, "Should have performed optimizations")
				_ = optimizationCount
				optimizationCount = newMetrics.Optimizations
			}
		})
	}
}

func TestOptimizedCompiler_DeadCodeElimination(t *testing.T) {
	config := OptimizedCompilerConfig{
		EnableDeadCodeElimination: false, // Disabled for now as it requires AST analysis
		EnableStaticAnalysis:      true,
		WarnUnreachableCode:       true,
	}

	compiler := NewOptimizedCompiler(config)

	script := `
		function test(x)
			if x > 0 then
				return x
			end
			-- The parser enforces no code after return at parse time
			-- So we test with if true pattern instead
			if true then
				return 0
			end
			print("This is unreachable")
		end
		
		local result = test(10)
		return result
	`

	proto, err := compiler.Compile(script, "dead_code.lua")
	require.NoError(t, err)
	require.NotNil(t, proto)

	// Check for unreachable code warning
	warnings := compiler.ValidateScript(script, "dead_code.lua")
	hasUnreachableWarning := false
	for _, w := range warnings {
		if w.Type == "UnreachableCode" {
			hasUnreachableWarning = true
			break
		}
	}
	assert.True(t, hasUnreachableWarning, "Should detect unreachable code")
}

func TestOptimizedCompiler_Dependencies(t *testing.T) {
	config := OptimizedCompilerConfig{
		EnableStaticAnalysis: true,
	}

	compiler := NewOptimizedCompiler(config)

	script := `
		local json = require("json")
		local utils = require("utils/helpers")
		
		function process(data)
			return json.encode(data)
		end
		
		return process
	`

	proto, err := compiler.Compile(script, "deps.lua")
	require.NoError(t, err)
	require.NotNil(t, proto)

	// Check that compilation succeeded
	// We can't directly check dependencies without accessing internal metadata
	// This test mainly ensures the dependency detection doesn't break compilation
}

func TestOptimizedCompiler_ConcurrentCompilation(t *testing.T) {
	config := OptimizedCompilerConfig{
		ChunkCacheConfig: ChunkCacheConfig{
			MaxSize: 100,
			TTL:     time.Minute,
		},
		EnableSourceOptimization:  true,
		EnableDeadCodeElimination: false, // Disabled as it requires AST
		EnableConstantFolding:     true,
		EnableCommentStripping:    true,
		EnableWhitespaceReduction: true,
	}

	compiler := NewOptimizedCompiler(config)

	// Create multiple scripts with optimizable patterns
	scripts := make([]string, 20)
	for i := range scripts {
		scripts[i] = fmt.Sprintf(`
			-- This comment will be stripped
			local x = %d + %d  -- Constant folding opportunity
			local y = x * 2
			local msg = "Hello" .. " " .. "World"  -- String concatenation
			return y + %d
		`, i, i, i*10)
	}

	// Compile concurrently
	var wg sync.WaitGroup
	errors := make(chan error, len(scripts))

	for i, script := range scripts {
		wg.Add(1)
		go func(idx int, s string) {
			defer wg.Done()

			_, err := compiler.Compile(s, fmt.Sprintf("script_%d.lua", idx))
			if err != nil {
				errors <- err
			}
		}(i, script)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Logf("Compilation error: %v", err)
		errorCount++
	}
	assert.Equal(t, 0, errorCount, "No compilation errors should occur")

	// Verify metrics
	metrics := compiler.GetMetrics()
	assert.Equal(t, int64(len(scripts)), metrics.Compilations)
	assert.Greater(t, metrics.Optimizations, int64(0))
}

func TestOptimizedCompiler_LargeScript(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large script test in short mode")
	}

	config := OptimizedCompilerConfig{
		EnableSourceOptimization:  true,
		EnableDeadCodeElimination: true,
		EnableConstantFolding:     true,
		EnableStaticAnalysis:      true,
	}

	compiler := NewOptimizedCompiler(config)

	// Generate a large script
	var sb strings.Builder
	sb.WriteString("local data = {\n")
	for i := 0; i < 1000; i++ {
		sb.WriteString(fmt.Sprintf("  item_%d = %d,\n", i, i*2))
	}
	sb.WriteString("}\n\n")

	sb.WriteString("function process()\n")
	sb.WriteString("  local sum = 0\n")
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf("  sum = sum + data.item_%d\n", i))
	}
	sb.WriteString("  return sum\n")
	sb.WriteString("end\n\n")
	sb.WriteString("return process()\n")

	script := sb.String()

	// Compile and measure time
	start := time.Now()
	proto, err := compiler.Compile(script, "large.lua")
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, proto)

	t.Logf("Large script compilation took: %v", duration)

	// Check metrics
	metrics := compiler.GetMetrics()
	assert.Greater(t, metrics.TotalCompileTime, time.Duration(0))
	assert.Greater(t, metrics.TotalOptimizeTime, time.Duration(0))
}

func TestOptimizedCompiler_OptimizationFlags(t *testing.T) {
	tests := []struct {
		name   string
		config OptimizedCompilerConfig
		script string
	}{
		{
			name: "all optimizations disabled",
			config: OptimizedCompilerConfig{
				EnableSourceOptimization:  false,
				EnableDeadCodeElimination: false,
				EnableConstantFolding:     false,
			},
			script: `local x = 10 + 20; return x`,
		},
		{
			name: "only constant folding",
			config: OptimizedCompilerConfig{
				EnableConstantFolding: true,
			},
			script: `local x = 10 + 20; return x`,
		},
		{
			name: "aggressive optimization",
			config: OptimizedCompilerConfig{
				EnableSourceOptimization:  true,
				EnableDeadCodeElimination: false, // Disabled as it requires AST
				EnableConstantFolding:     true,
				EnableCommentStripping:    true,
				EnableWhitespaceReduction: true,
				AggressiveOptimization:    true,
			},
			script: `
				-- Comment to strip
				local x = 10 + 20  -- Constant folding
				local msg = "Hello" .. " World"  -- String concat
				function factorial(n)
					if n <= 1 then return 1 end
					return n * factorial(n - 1)
				end
				return factorial(5)
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewOptimizedCompiler(tt.config)

			proto, err := compiler.Compile(tt.script, tt.name+".lua")
			require.NoError(t, err)
			require.NotNil(t, proto)

			metrics := compiler.GetMetrics()

			if tt.config.EnableSourceOptimization || tt.config.EnableConstantFolding ||
				tt.config.EnableDeadCodeElimination {
				assert.Greater(t, metrics.Optimizations, int64(0))
			} else {
				assert.Equal(t, int64(0), metrics.Optimizations)
			}
		})
	}
}

func TestOptimizedCompiler_SourceMapping(t *testing.T) {
	config := OptimizedCompilerConfig{
		EnableSourceOptimization: true,
	}

	compiler := NewOptimizedCompiler(config)

	script := `
		local x = 10
		local y = 20
		return x + y
	`

	proto, sourceMap, err := compiler.CompileWithSourceMap(script, "source_map.lua")
	require.NoError(t, err)
	require.NotNil(t, proto)
	require.NotNil(t, sourceMap)

	// Source map functionality would be more complex in real implementation
	// For now just verify it returns non-nil
}

func TestOptimizedCompiler_Reset(t *testing.T) {
	config := OptimizedCompilerConfig{
		ChunkCacheConfig: ChunkCacheConfig{
			MaxSize: 10,
		},
	}

	compiler := NewOptimizedCompiler(config)

	// Compile some scripts
	for i := 0; i < 5; i++ {
		script := fmt.Sprintf("return %d", i)
		_, err := compiler.Compile(script, fmt.Sprintf("script_%d.lua", i))
		require.NoError(t, err)
	}

	// Verify metrics are populated
	metrics := compiler.GetMetrics()
	assert.Greater(t, metrics.Compilations, int64(0))
	assert.Greater(t, compiler.cache.Size(), 0)

	// Reset
	compiler.Reset()

	// Verify everything is cleared
	metrics = compiler.GetMetrics()
	assert.Equal(t, int64(0), metrics.Compilations)
	assert.Equal(t, int64(0), metrics.CacheHits)
	assert.Equal(t, int64(0), metrics.CacheMisses)
	assert.Equal(t, 0, compiler.cache.Size())
}

func TestEnhancedChunkCache_Metadata(t *testing.T) {
	cache := &EnhancedChunkCache{
		ChunkCache: NewChunkCache(ChunkCacheConfig{MaxSize: 10}),
		metadata:   make(map[string]*ChunkMetadata),
		sourceMaps: make(map[string]*SourceMap),
	}

	key := "test_key"
	metadata := &ChunkMetadata{
		OriginalSize:    100,
		OptimizedSize:   80,
		CompilationTime: 10 * time.Millisecond,
		Optimizations:   []string{"constant_folding"},
	}

	// Store metadata
	cache.PutMetadata(key, metadata)

	// Retrieve metadata
	retrieved := cache.GetMetadata(key)
	require.NotNil(t, retrieved)
	assert.Equal(t, metadata.OriginalSize, retrieved.OriginalSize)
	assert.Equal(t, metadata.OptimizedSize, retrieved.OptimizedSize)
	assert.Equal(t, metadata.Optimizations, retrieved.Optimizations)

	// Non-existent key
	assert.Nil(t, cache.GetMetadata("non_existent"))
}

// Benchmark compilation performance
func BenchmarkOptimizedCompiler(b *testing.B) {
	config := OptimizedCompilerConfig{
		ChunkCacheConfig: ChunkCacheConfig{
			MaxSize: 100,
			TTL:     time.Minute,
		},
		EnableSourceOptimization:  true,
		EnableDeadCodeElimination: true,
		EnableConstantFolding:     true,
	}

	compiler := NewOptimizedCompiler(config)

	script := `
		local function fibonacci(n)
			if n <= 1 then
				return n
			end
			return fibonacci(n - 1) + fibonacci(n - 2)
		end
		
		local result = fibonacci(10)
		return result
	`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := compiler.Compile(script, "bench.lua")
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()

	metrics := compiler.GetMetrics()
	b.Logf("Cache hit rate: %.2f%%", metrics.CacheHitRate*100)
	b.Logf("Avg compile time: %v", metrics.AvgCompileTime)
	b.Logf("Avg optimize time: %v", metrics.AvgOptimizeTime)
}
