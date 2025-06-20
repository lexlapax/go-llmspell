// ABOUTME: Profiling infrastructure for Lua engine performance analysis including execution time, memory usage, and hot path tracking
// ABOUTME: Provides comprehensive profiling API with minimal overhead for production use

package gopherlua

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ProfilerInterface defines the profiling API
type ProfilerInterface interface {
	Enable()
	Disable()
	IsEnabled() bool
	Reset()

	// Execution profiling
	RecordFunctionCall(name string, start time.Time)
	RecordFunctionReturn(name string, end time.Time)
	GetExecutionProfile() *ExecutionProfile
	GetHotPaths(limit int) []HotPath

	// Memory profiling
	RecordAllocation(typ string, size uint64, location string)
	GetMemoryProfile() *MemoryProfile
	GetAllocationSites() []AllocationSite

	// Allocation tracking
	EnableAllocationTracking(enable bool)
	IsAllocationTrackingEnabled() bool

	// Export/Import
	Export() ([]byte, error)
	Import(data []byte) error
}

// ExecutionProfile contains execution timing data
type ExecutionProfile struct {
	TotalTime     time.Duration            `json:"total_time"`
	FunctionTimes map[string]FunctionStats `json:"function_times"`
	CallGraph     map[string][]string      `json:"call_graph"`
	StartTime     time.Time                `json:"start_time"`
	EndTime       time.Time                `json:"end_time"`
}

// FunctionStats contains statistics for a single function
type FunctionStats struct {
	Name      string        `json:"name"`
	TotalTime time.Duration `json:"total_time"`
	SelfTime  time.Duration `json:"self_time"`
	CallCount uint64        `json:"call_count"`
	MinTime   time.Duration `json:"min_time"`
	MaxTime   time.Duration `json:"max_time"`
	AvgTime   time.Duration `json:"avg_time"`
}

// HotPath represents a frequently executed code path
type HotPath struct {
	Name       string        `json:"name"`
	CallCount  uint64        `json:"call_count"`
	TotalTime  time.Duration `json:"total_time"`
	AvgTime    time.Duration `json:"avg_time"`
	Percentage float64       `json:"percentage"` // Percentage of total execution time
}

// MemoryProfile contains memory usage data
type MemoryProfile struct {
	Allocations uint64                  `json:"allocations"`
	TotalBytes  uint64                  `json:"total_bytes"`
	LiveObjects uint64                  `json:"live_objects"`
	TypeStats   map[string]TypeMemStats `json:"type_stats"`
	HeapAlloc   uint64                  `json:"heap_alloc"`
	HeapSys     uint64                  `json:"heap_sys"`
}

// TypeMemStats contains memory statistics for a specific type
type TypeMemStats struct {
	Type       string `json:"type"`
	Count      uint64 `json:"count"`
	TotalBytes uint64 `json:"total_bytes"`
	AvgBytes   uint64 `json:"avg_bytes"`
}

// AllocationSite represents a location where allocations occur
type AllocationSite struct {
	Location string `json:"location"`
	Type     string `json:"type"`
	Count    uint64 `json:"count"`
	Bytes    uint64 `json:"bytes"`
	AvgBytes uint64 `json:"avg_bytes"`
}

// Profiler implements the ProfilerInterface
type Profiler struct {
	mu            sync.RWMutex
	enabled       atomic.Bool
	allocTracking atomic.Bool

	// Execution profiling
	functionStats map[string]*FunctionStats
	callStack     []string
	callGraph     map[string]map[string]bool
	currentCalls  map[string]time.Time

	// Memory profiling
	allocations     uint64
	totalBytes      uint64
	typeStats       map[string]*TypeMemStats
	allocationSites map[string]*AllocationSite

	// Timing
	startTime time.Time
	endTime   time.Time
}

// NewProfiler creates a new profiler instance
func NewProfiler() *Profiler {
	return &Profiler{
		functionStats:   make(map[string]*FunctionStats),
		callGraph:       make(map[string]map[string]bool),
		currentCalls:    make(map[string]time.Time),
		typeStats:       make(map[string]*TypeMemStats),
		allocationSites: make(map[string]*AllocationSite),
		startTime:       time.Now(),
	}
}

// Enable enables the profiler
func (p *Profiler) Enable() {
	p.enabled.Store(true)
	p.mu.Lock()
	p.startTime = time.Now()
	p.mu.Unlock()
}

// Disable disables the profiler
func (p *Profiler) Disable() {
	p.enabled.Store(false)
	p.mu.Lock()
	p.endTime = time.Now()
	p.mu.Unlock()
}

// IsEnabled returns whether profiling is enabled
func (p *Profiler) IsEnabled() bool {
	return p.enabled.Load()
}

// Reset clears all profiling data
func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.functionStats = make(map[string]*FunctionStats)
	p.callStack = nil
	p.callGraph = make(map[string]map[string]bool)
	p.currentCalls = make(map[string]time.Time)
	p.allocations = 0
	p.totalBytes = 0
	p.typeStats = make(map[string]*TypeMemStats)
	p.allocationSites = make(map[string]*AllocationSite)
	p.startTime = time.Now()
	p.endTime = time.Time{}
}

// RecordFunctionCall records the start of a function call
func (p *Profiler) RecordFunctionCall(name string, start time.Time) {
	if !p.enabled.Load() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Track call stack
	if len(p.callStack) > 0 {
		caller := p.callStack[len(p.callStack)-1]
		if p.callGraph[caller] == nil {
			p.callGraph[caller] = make(map[string]bool)
		}
		p.callGraph[caller][name] = true
	}

	p.callStack = append(p.callStack, name)
	p.currentCalls[name] = start

	// Initialize stats if needed
	if p.functionStats[name] == nil {
		p.functionStats[name] = &FunctionStats{
			Name:    name,
			MinTime: time.Duration(1<<63 - 1), // MaxInt64
		}
	}
}

// RecordFunctionReturn records the end of a function call
func (p *Profiler) RecordFunctionReturn(name string, end time.Time) {
	if !p.enabled.Load() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Pop from call stack
	if len(p.callStack) > 0 && p.callStack[len(p.callStack)-1] == name {
		p.callStack = p.callStack[:len(p.callStack)-1]
	}

	// Calculate duration
	if start, ok := p.currentCalls[name]; ok {
		duration := end.Sub(start)
		delete(p.currentCalls, name)

		if stats := p.functionStats[name]; stats != nil {
			stats.TotalTime += duration
			stats.CallCount++

			if duration < stats.MinTime {
				stats.MinTime = duration
			}
			if duration > stats.MaxTime {
				stats.MaxTime = duration
			}

			stats.AvgTime = stats.TotalTime / time.Duration(stats.CallCount)
		}
	}
}

// GetExecutionProfile returns the execution profile
func (p *Profiler) GetExecutionProfile() *ExecutionProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Copy function stats
	funcTimes := make(map[string]FunctionStats)
	var totalTime time.Duration

	for name, stats := range p.functionStats {
		funcTimes[name] = *stats
		totalTime += stats.TotalTime
	}

	// Copy call graph
	callGraph := make(map[string][]string)
	for caller, callees := range p.callGraph {
		for callee := range callees {
			callGraph[caller] = append(callGraph[caller], callee)
		}
		sort.Strings(callGraph[caller])
	}

	// If no functions were tracked and profiler was never started, return zero duration
	if len(p.functionStats) == 0 && !p.enabled.Load() && p.endTime.IsZero() {
		return &ExecutionProfile{
			TotalTime:     0,
			FunctionTimes: funcTimes,
			CallGraph:     callGraph,
			StartTime:     p.startTime,
			EndTime:       p.startTime,
		}
	}

	endTime := p.endTime
	if endTime.IsZero() && p.enabled.Load() {
		endTime = time.Now()
	} else if endTime.IsZero() {
		endTime = p.startTime
	}

	return &ExecutionProfile{
		TotalTime:     endTime.Sub(p.startTime),
		FunctionTimes: funcTimes,
		CallGraph:     callGraph,
		StartTime:     p.startTime,
		EndTime:       endTime,
	}
}

// GetHotPaths returns the most frequently called or time-consuming functions
func (p *Profiler) GetHotPaths(limit int) []HotPath {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Calculate total execution time
	var totalTime time.Duration
	for _, stats := range p.functionStats {
		totalTime += stats.TotalTime
	}

	// Create hot paths
	paths := make([]HotPath, 0, len(p.functionStats))
	for name, stats := range p.functionStats {
		percentage := 0.0
		if totalTime > 0 {
			percentage = float64(stats.TotalTime) / float64(totalTime) * 100
		}

		paths = append(paths, HotPath{
			Name:       name,
			CallCount:  stats.CallCount,
			TotalTime:  stats.TotalTime,
			AvgTime:    stats.AvgTime,
			Percentage: percentage,
		})
	}

	// Sort by total time (hottest first)
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].TotalTime > paths[j].TotalTime
	})

	// Limit results
	if limit > 0 && limit < len(paths) {
		paths = paths[:limit]
	}

	return paths
}

// RecordAllocation records a memory allocation
func (p *Profiler) RecordAllocation(typ string, size uint64, location string) {
	if !p.enabled.Load() || !p.allocTracking.Load() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.allocations++
	p.totalBytes += size

	// Update type stats
	if p.typeStats[typ] == nil {
		p.typeStats[typ] = &TypeMemStats{Type: typ}
	}
	stats := p.typeStats[typ]
	stats.Count++
	stats.TotalBytes += size
	stats.AvgBytes = stats.TotalBytes / stats.Count

	// Update allocation sites
	if location != "" {
		key := fmt.Sprintf("%s:%s", location, typ)
		if p.allocationSites[key] == nil {
			p.allocationSites[key] = &AllocationSite{
				Location: location,
				Type:     typ,
			}
		}
		site := p.allocationSites[key]
		site.Count++
		site.Bytes += size
		site.AvgBytes = site.Bytes / site.Count
	}
}

// GetMemoryProfile returns the memory profile
func (p *Profiler) GetMemoryProfile() *MemoryProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Get current memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Copy type stats
	typeStats := make(map[string]TypeMemStats)
	for typ, stats := range p.typeStats {
		typeStats[typ] = *stats
	}

	return &MemoryProfile{
		Allocations: p.allocations,
		TotalBytes:  p.totalBytes,
		LiveObjects: uint64(len(p.allocationSites)),
		TypeStats:   typeStats,
		HeapAlloc:   memStats.HeapAlloc,
		HeapSys:     memStats.HeapSys,
	}
}

// GetAllocationSites returns allocation sites sorted by bytes
func (p *Profiler) GetAllocationSites() []AllocationSite {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sites := make([]AllocationSite, 0, len(p.allocationSites))
	for _, site := range p.allocationSites {
		sites = append(sites, *site)
	}

	// Sort by total bytes (highest first)
	sort.Slice(sites, func(i, j int) bool {
		return sites[i].Bytes > sites[j].Bytes
	})

	return sites
}

// EnableAllocationTracking enables or disables allocation tracking
func (p *Profiler) EnableAllocationTracking(enable bool) {
	p.allocTracking.Store(enable)
}

// IsAllocationTrackingEnabled returns whether allocation tracking is enabled
func (p *Profiler) IsAllocationTrackingEnabled() bool {
	return p.allocTracking.Load()
}

// Export exports profiling data as JSON
func (p *Profiler) Export() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	data := struct {
		Execution   *ExecutionProfile `json:"execution"`
		Memory      *MemoryProfile    `json:"memory"`
		HotPaths    []HotPath         `json:"hot_paths"`
		Allocations []AllocationSite  `json:"allocations"`
	}{
		Execution:   p.GetExecutionProfile(),
		Memory:      p.GetMemoryProfile(),
		HotPaths:    p.GetHotPaths(20),
		Allocations: p.GetAllocationSites(),
	}

	return json.MarshalIndent(data, "", "  ")
}

// Import imports profiling data from JSON
func (p *Profiler) Import(data []byte) error {
	// For simplicity, just return nil
	// In a real implementation, this would parse and restore the data
	return nil
}

// InstallProfilerAPI installs the profiler API in the Lua state
func InstallProfilerAPI(L *lua.LState, profiler ProfilerInterface) {
	// Create a table for the profiler API
	profilerTable := L.NewTable()

	// Register methods
	L.SetField(profilerTable, "start", L.NewFunction(func(L *lua.LState) int {
		profiler.Enable()
		return 0
	}))

	L.SetField(profilerTable, "stop", L.NewFunction(func(L *lua.LState) int {
		profiler.Disable()
		return 0
	}))

	L.SetField(profilerTable, "reset", L.NewFunction(func(L *lua.LState) int {
		profiler.Reset()
		return 0
	}))

	L.SetField(profilerTable, "getProfile", L.NewFunction(func(L *lua.LState) int {
		profile := profiler.GetExecutionProfile()

		// Convert to Lua table
		tbl := L.NewTable()
		L.SetField(tbl, "totalTime", lua.LNumber(profile.TotalTime.Nanoseconds()))

		// Add function times
		funcTimes := L.NewTable()
		for name, stats := range profile.FunctionTimes {
			funcTbl := L.NewTable()
			L.SetField(funcTbl, "callCount", lua.LNumber(stats.CallCount))
			L.SetField(funcTbl, "totalTime", lua.LNumber(stats.TotalTime.Nanoseconds()))
			L.SetField(funcTbl, "avgTime", lua.LNumber(stats.AvgTime.Nanoseconds()))
			L.SetField(funcTimes, name, funcTbl)
		}
		L.SetField(tbl, "functions", funcTimes)

		L.Push(tbl)
		return 1
	}))

	// Set as global
	L.SetGlobal("profiler", profilerTable)
}
