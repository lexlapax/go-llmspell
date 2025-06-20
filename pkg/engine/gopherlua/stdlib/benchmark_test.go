// ABOUTME: Performance benchmarks for go-llmspell Lua standard library modules
// ABOUTME: Tests promise creation/resolution, module loading, memory usage, concurrency, events, state, and tools

package stdlib

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
	lua "github.com/yuin/gopher-lua"
)

// ============================================================================
// Promise Creation/Resolution Benchmarks
// ============================================================================

// BenchmarkPromiseCreationPerf measures promise creation performance
func BenchmarkPromiseCreationPerf(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	setupMockPromiseModule(L)

	script := `
		function create_promise()
			return promise.new(function(resolve, reject)
				resolve("test")
			end)
		end
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := L.DoString("create_promise()"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPromiseResolution measures promise resolution performance
func BenchmarkPromiseResolution(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	setupMockPromiseModule(L)

	script := `
		function test_promise_resolution()
			local resolved = false
			promise.new(function(resolve, reject)
				resolve("test")
			end):andThen(function(value)
				resolved = true
			end)
			return resolved
		end
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := L.DoString("test_promise_resolution()"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPromiseChaining measures promise chain performance
func BenchmarkPromiseChaining(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	setupMockPromiseModule(L)

	script := `
		function test_promise_chain()
			local result = nil
			promise.new(function(resolve, reject)
				resolve(1)
			end):andThen(function(value)
				return value + 1
			end):andThen(function(value)
				return value + 1
			end):andThen(function(value)
				result = value
			end)
			return result
		end
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := L.DoString("test_promise_chain()"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPromiseAll measures Promise.all performance
func BenchmarkPromiseAll(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	setupMockPromiseModule(L)

	for _, count := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("promises_%d", count), func(b *testing.B) {
			script := fmt.Sprintf(`
				function test_promise_all()
					local promises = {}
					for i = 1, %d do
						promises[i] = promise.new(function(resolve, reject)
							resolve(i)
						end)
					end
					return promise.all(promises)
				end
			`, count)

			if err := L.DoString(script); err != nil {
				b.Fatalf("Setup failed: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := L.DoString("test_promise_all()"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ============================================================================
// Module Loading Time Benchmarks
// ============================================================================

// BenchmarkModuleLoading measures module loading performance
func BenchmarkModuleLoading(b *testing.B) {
	modules := []string{"promise", "llm", "agent", "state", "events", "tools", "errors"}

	for _, module := range modules {
		b.Run(module, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				L := lua.NewState()
				setupMockModule(L, module)
				L.Close()
			}
		})
	}
}

// BenchmarkAllModulesLoading measures loading all modules
func BenchmarkAllModulesLoading(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		L := lua.NewState()
		modules := []string{"promise", "llm", "agent", "state", "events", "tools", "errors", "data", "logging"}
		for _, module := range modules {
			setupMockModule(L, module)
		}
		L.Close()
	}
}

// ============================================================================
// Memory Usage Profiling
// ============================================================================

// BenchmarkMemoryUsage profiles memory usage for various operations
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("promise_creation", func(b *testing.B) {
		L := lua.NewState()
		defer L.Close()
		setupMockPromiseModule(L)

		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		allocBefore := m.Alloc

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = L.DoString(`
				local p = promise.new(function(resolve, reject)
					resolve("test")
				end)
			`)
		}

		runtime.GC()
		runtime.ReadMemStats(&m)
		allocAfter := m.Alloc

		b.ReportMetric(float64(allocAfter-allocBefore)/float64(b.N), "bytes/op")
	})

	b.Run("large_state_storage", func(b *testing.B) {
		L := lua.NewState()
		defer L.Close()
		setupMockStateModule(L)

		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		allocBefore := m.Alloc

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = L.DoString(fmt.Sprintf(`
				state.set("key_%d", string.rep("x", 1000))
			`, i))
		}

		runtime.GC()
		runtime.ReadMemStats(&m)
		allocAfter := m.Alloc

		b.ReportMetric(float64(allocAfter-allocBefore)/float64(b.N), "bytes/op")
	})
}

// BenchmarkMemoryLeaks tests for memory leaks in promise chains
func BenchmarkMemoryLeaks(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	setupMockPromiseModule(L)

	script := `
		function create_leaky_chain()
			local data = string.rep("x", 10000)
			return promise.new(function(resolve)
				resolve(data)
			end):andThen(function(value)
				return value .. "a"
			end):andThen(function(value)
				return value .. "b"
			end)
		end
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	// Force GC before starting
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocBefore := m.Alloc

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = L.DoString("create_leaky_chain()")
		if i%100 == 0 {
			_ = L.DoString("collectgarbage('step')")
		}
	}

	// Force GC and measure
	runtime.GC()
	runtime.ReadMemStats(&m)
	allocAfter := m.Alloc

	avgBytesPerOp := float64(allocAfter-allocBefore) / float64(b.N)
	b.ReportMetric(avgBytesPerOp, "bytes/op")

	// Report if there's potential leak (>1KB per operation)
	if avgBytesPerOp > 1024 {
		b.Logf("WARNING: Potential memory leak detected: %.2f bytes/op", avgBytesPerOp)
	}
}

// ============================================================================
// Concurrent Operation Stress Tests
// ============================================================================

// BenchmarkConcurrentPromises measures concurrent promise performance
func BenchmarkConcurrentPromises(b *testing.B) {
	for _, concurrency := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("concurrent_%d", concurrency), func(b *testing.B) {
			L := lua.NewState()
			defer L.Close()
			setupMockPromiseModule(L)

			script := fmt.Sprintf(`
				function concurrent_promises()
					local promises = {}
					for i = 1, %d do
						promises[i] = promise.new(function(resolve, reject)
							-- Simulate async work
							promise.sleep(1):andThen(function()
								resolve(i)
							end)
						end)
					end
					return promise.all(promises)
				end
			`, concurrency)

			if err := L.DoString(script); err != nil {
				b.Fatalf("Setup failed: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := L.DoString("concurrent_promises()"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCoroutineStress tests coroutine performance under load
func BenchmarkCoroutineStress(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	script := `
		function coroutine_stress(count)
			local coroutines = {}
			for i = 1, count do
				coroutines[i] = coroutine.create(function()
					for j = 1, 10 do
						coroutine.yield(j)
					end
				end)
			end
			
			-- Resume all coroutines
			local results = 0
			for i = 1, count do
				while coroutine.status(coroutines[i]) ~= "dead" do
					local ok, val = coroutine.resume(coroutines[i])
					if ok then
						results = results + 1
					end
				end
			end
			return results
		end
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	for _, count := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("coroutines_%d", count), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := L.DoString(fmt.Sprintf("coroutine_stress(%d)", count)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ============================================================================
// Event System Throughput Tests
// ============================================================================

// BenchmarkEventEmissionPerf measures event emission performance
func BenchmarkEventEmissionPerf(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	setupMockEventsModule(L)

	script := `
		-- Set up listeners
		local counter = 0
		events.on("test.event", function(data)
			counter = counter + 1
		end)
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := L.DoString(`events.emit("test.event", {value = 1})`); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEventListeners measures performance with many listeners
func BenchmarkEventListeners(b *testing.B) {
	for _, listeners := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("listeners_%d", listeners), func(b *testing.B) {
			L := lua.NewState()
			defer L.Close()
			setupMockEventsModule(L)

			// Add many listeners
			script := fmt.Sprintf(`
				local counter = 0
				for i = 1, %d do
					events.on("test.event", function(data)
						counter = counter + 1
					end)
				end
			`, listeners)

			if err := L.DoString(script); err != nil {
				b.Fatalf("Setup failed: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := L.DoString(`events.emit("test.event", {value = 1})`); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkEventThroughput measures maximum event throughput
func BenchmarkEventThroughput(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	setupMockEventsModule(L)

	script := `
		-- Simple listener
		events.on("throughput.test", function(data) end)
		
		function emit_burst(count)
			for i = 1, count do
				events.emit("throughput.test", {index = i})
			end
		end
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	start := time.Now()
	totalEvents := 0

	for i := 0; i < b.N; i++ {
		if err := L.DoString("emit_burst(1000)"); err != nil {
			b.Fatal(err)
		}
		totalEvents += 1000
	}

	elapsed := time.Since(start)
	eventsPerSecond := float64(totalEvents) / elapsed.Seconds()
	b.ReportMetric(eventsPerSecond, "events/sec")
}

// ============================================================================
// State Management Scalability Tests
// ============================================================================

// BenchmarkStateOperations measures state get/set performance
func BenchmarkStateOperations(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	setupMockStateModule(L)

	b.Run("set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := L.DoString(fmt.Sprintf(`state.set("key_%d", %d)`, i%1000, i)); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("get", func(b *testing.B) {
		// Pre-populate state
		for i := 0; i < 1000; i++ {
			_ = L.DoString(fmt.Sprintf(`state.set("key_%d", %d)`, i, i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := L.DoString(fmt.Sprintf(`state.get("key_%d")`, i%1000)); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("mixed", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				if err := L.DoString(fmt.Sprintf(`state.set("key_%d", %d)`, i%1000, i)); err != nil {
					b.Fatal(err)
				}
			} else {
				if err := L.DoString(fmt.Sprintf(`state.get("key_%d")`, i%1000)); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

// BenchmarkStateScalability tests state with many keys
func BenchmarkStateScalability(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("keys_%d", size), func(b *testing.B) {
			L := lua.NewState()
			defer L.Close()
			setupMockStateModule(L)

			// Pre-populate state
			for i := 0; i < size; i++ {
				_ = L.DoString(fmt.Sprintf(`state.set("key_%d", "value_%d")`, i, i))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := i % size
				if err := L.DoString(fmt.Sprintf(`state.get("key_%d")`, key)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ============================================================================
// Tool Execution Performance Tests
// ============================================================================

// BenchmarkToolExecution measures tool execution performance
func BenchmarkToolExecution(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	setupMockToolsModule(L)

	script := `
		-- Define a simple tool
		local simple_tool = tools.define({
			name = "benchmark_tool",
			description = "Tool for benchmarking",
			parameters = {
				{name = "input", type = "string", required = true}
			},
			func = function(params)
				return {result = string.upper(params.input)}
			end
		})
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := L.DoString(`tools.execute("benchmark_tool", {input = "test"})`); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkToolPipeline measures pipeline execution performance
func BenchmarkToolPipeline(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	setupMockToolsModule(L)

	script := `
		-- Define tools for pipeline
		local tool1 = tools.define({
			name = "uppercase",
			func = function(params)
				return {text = string.upper(params.text or params)}
			end
		})
		
		local tool2 = tools.define({
			name = "prefix",
			func = function(params)
				local text = params.text or params
				return {text = "PREFIX_" .. text}
			end
		})
		
		local tool3 = tools.define({
			name = "suffix",
			func = function(params)
				local text = params.text or params
				return {text = text .. "_SUFFIX"}
			end
		})
		
		-- Create pipeline
		test_pipeline = tools.pipeline({tool1, tool2, tool3})
	`

	if err := L.DoString(script); err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := L.DoString(`tools.execute(test_pipeline, {text = "benchmark"})`); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBridgeMethodCalls measures bridge method call performance
func BenchmarkBridgeMethodCalls(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	// Create and setup mock bridge
	bridge := testutils.NewMockBridge("benchmark-bridge")
	bridge.WithMethod("test_method", engine.MethodInfo{
		Name:        "test_method",
		Description: "Test method for benchmarking",
		Parameters: []engine.ParameterInfo{
			{Name: "input", Type: "string", Required: true},
		},
		ReturnType: "string",
	}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
		if len(args) > 0 {
			return engine.NewStringValue("Result: " + args[0].String()), nil
		}
		return engine.NewStringValue("Result: empty"), nil
	})

	bridge.WithInitialized(true)

	// Set up bridge in Lua
	bridgeTable := L.NewTable()
	bridgeModule := CreateMockBridgeModule(L, bridge)
	L.SetField(bridgeTable, "test", bridgeModule)
	L.SetGlobal("bridge", bridgeTable)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := L.DoString(`bridge.test:test_method("benchmark")`); err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// Performance Report Generation
// ============================================================================

// BenchmarkGenerateReport generates a performance report
func BenchmarkGenerateReport(b *testing.B) {
	// This is not a traditional benchmark but generates a report
	b.Skip("Use TestGeneratePerformanceReport instead")
}

// TestGeneratePerformanceReport generates a comprehensive performance report
func TestGeneratePerformanceReport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance report in short mode")
	}

	report := PerformanceReport{
		Timestamp: time.Now(),
		System: SystemInfo{
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			CPUs:      runtime.NumCPU(),
			GoVersion: runtime.Version(),
			MaxProcs:  runtime.GOMAXPROCS(0),
		},
	}

	// Run mini benchmarks to collect data
	L := lua.NewState()
	defer L.Close()

	// Promise performance
	setupMockPromiseModule(L)
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_ = L.DoString(`promise.new(function(resolve) resolve(1) end)`)
	}
	report.PromiseMetrics.CreationTime = time.Since(start) / 1000

	// Module loading
	moduleStart := time.Now()
	modules := []string{"promise", "llm", "agent", "state", "events"}
	for _, mod := range modules {
		setupMockModule(L, mod)
	}
	report.ModuleMetrics.TotalLoadTime = time.Since(moduleStart)
	report.ModuleMetrics.AverageLoadTime = report.ModuleMetrics.TotalLoadTime / time.Duration(len(modules))

	// Memory snapshot
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	report.MemoryMetrics.HeapAlloc = m.HeapAlloc
	report.MemoryMetrics.HeapObjects = m.HeapObjects
	report.MemoryMetrics.NumGC = m.NumGC

	// Event throughput
	setupMockEventsModule(L)
	eventStart := time.Now()
	_ = L.DoString(`
		events.on("perf.test", function() end)
		for i = 1, 10000 do
			events.emit("perf.test", {i = i})
		end
	`)
	elapsed := time.Since(eventStart)
	report.EventMetrics.ThroughputPerSecond = int(10000.0 / elapsed.Seconds())

	// State operations
	setupMockStateModule(L)
	stateStart := time.Now()
	for i := 0; i < 1000; i++ {
		_ = L.DoString(fmt.Sprintf(`state.set("k%d", %d)`, i, i))
	}
	report.StateMetrics.WriteOpsPerSecond = int(1000.0 / time.Since(stateStart).Seconds())

	// Generate report
	fmt.Printf("\n=== Go-LLMSpell Lua Standard Library Performance Report ===\n")
	fmt.Printf("Generated: %s\n", report.Timestamp.Format(time.RFC3339))
	fmt.Printf("\nSystem Information:\n")
	fmt.Printf("  OS/Arch: %s/%s\n", report.System.OS, report.System.Arch)
	fmt.Printf("  CPUs: %d (GOMAXPROCS: %d)\n", report.System.CPUs, report.System.MaxProcs)
	fmt.Printf("  Go Version: %s\n", report.System.GoVersion)
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("  Promise Creation: %.2fµs/op\n", float64(report.PromiseMetrics.CreationTime.Nanoseconds())/1000)
	fmt.Printf("  Module Load Time: %.2fms (avg: %.2fms)\n",
		float64(report.ModuleMetrics.TotalLoadTime.Nanoseconds())/1e6,
		float64(report.ModuleMetrics.AverageLoadTime.Nanoseconds())/1e6)
	fmt.Printf("  Event Throughput: %d events/sec\n", report.EventMetrics.ThroughputPerSecond)
	fmt.Printf("  State Write Ops: %d ops/sec\n", report.StateMetrics.WriteOpsPerSecond)
	fmt.Printf("\nMemory Usage:\n")
	fmt.Printf("  Heap Allocated: %.2f MB\n", float64(report.MemoryMetrics.HeapAlloc)/1024/1024)
	fmt.Printf("  Heap Objects: %d\n", report.MemoryMetrics.HeapObjects)
	fmt.Printf("  GC Runs: %d\n", report.MemoryMetrics.NumGC)
	fmt.Printf("\n=== End Report ===\n")
}

// ============================================================================
// Helper Types and Functions
// ============================================================================

// PerformanceReport holds performance test results
type PerformanceReport struct {
	Timestamp      time.Time
	System         SystemInfo
	PromiseMetrics PromisePerformance
	ModuleMetrics  ModulePerformance
	MemoryMetrics  MemoryUsage
	EventMetrics   EventPerformance
	StateMetrics   StatePerformance
}

// SystemInfo holds system information
type SystemInfo struct {
	OS        string
	Arch      string
	CPUs      int
	GoVersion string
	MaxProcs  int
}

// PromisePerformance holds promise performance metrics
type PromisePerformance struct {
	CreationTime   time.Duration
	ResolutionTime time.Duration
	ChainTime      time.Duration
	ConcurrentOps  int
	AllPerformance time.Duration
}

// ModulePerformance holds module loading metrics
type ModulePerformance struct {
	TotalLoadTime   time.Duration
	AverageLoadTime time.Duration
	ModuleCount     int
}

// MemoryUsage holds memory usage metrics
type MemoryUsage struct {
	HeapAlloc   uint64
	HeapObjects uint64
	NumGC       uint32
	BytesPerOp  float64
}

// EventPerformance holds event system metrics
type EventPerformance struct {
	ThroughputPerSecond int
	ListenerCount       int
	EmitLatency         time.Duration
}

// StatePerformance holds state management metrics
type StatePerformance struct {
	ReadOpsPerSecond  int
	WriteOpsPerSecond int
	KeyCount          int
}

// setupMockModule sets up a mock module for benchmarking
func setupMockModule(L *lua.LState, moduleName string) {
	switch moduleName {
	case "promise":
		setupMockPromiseModule(L)
	case "llm":
		setupMockLLMModule(L)
	case "agent":
		setupMockAgentModule(L)
	case "state":
		setupMockStateModule(L)
	case "events":
		setupMockEventsModule(L)
	case "tools":
		setupMockToolsModule(L)
	case "errors":
		setupMockErrorsModule(L)
	case "data":
		setupMockDataModule(L)
	case "logging":
		setupMockLoggingModule(L)
	}
}

// RunBenchmarkSuite runs all benchmarks and generates a report
func RunBenchmarkSuite(t *testing.T) {
	// This would run all benchmarks programmatically and generate a report
	// For now, users should run: go test -bench=. -benchmem
}
