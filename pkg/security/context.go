// ABOUTME: Secure execution contexts with resource management and policies
// ABOUTME: Provides sandboxed environments for script execution with limits

package security

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Context key types
type contextKey int

const (
	configKey contextKey = iota
	trackerKey
	policyKey
)

// ContextConfig defines the configuration for a secure execution context
type ContextConfig struct {
	// MaxMemory is the maximum memory allowed (in bytes)
	MaxMemory int64

	// MaxCPUTime is the maximum CPU time allowed
	MaxCPUTime time.Duration

	// MaxExecutionTime is the maximum wall-clock time allowed
	MaxExecutionTime time.Duration

	// MaxGoroutines is the maximum number of goroutines allowed
	MaxGoroutines int

	// SecurityPolicy defines additional security restrictions
	SecurityPolicy *SecurityPolicy
}

// Validate checks if the configuration is valid
func (c ContextConfig) Validate() error {
	if c.MaxMemory < 0 {
		return errors.New("MaxMemory must be non-negative")
	}
	if c.MaxExecutionTime <= 0 {
		return errors.New("MaxExecutionTime must be positive")
	}
	if c.MaxGoroutines < 0 {
		return errors.New("MaxGoroutines must be non-negative")
	}
	return nil
}

// SecurityPolicy defines security restrictions for execution
type SecurityPolicy struct {
	// AllowNetworkAccess controls network access
	AllowNetworkAccess bool

	// AllowFileWrite controls file write operations
	AllowFileWrite bool

	// AllowFileRead controls file read operations
	AllowFileRead bool

	// AllowedPaths lists paths that can be accessed
	AllowedPaths []string

	// BlockedPaths lists paths that cannot be accessed
	BlockedPaths []string
}

// IsPathAllowed checks if a path is allowed by the security policy
func (p *SecurityPolicy) IsPathAllowed(path string) bool {
	// Clean the path
	path = filepath.Clean(path)

	// Check blocked paths first
	for _, blocked := range p.BlockedPaths {
		if strings.HasPrefix(path, filepath.Clean(blocked)) {
			return false
		}
	}

	// If no allowed paths specified, allow all (except blocked)
	if len(p.AllowedPaths) == 0 {
		return true
	}

	// Check if path is in allowed list
	for _, allowed := range p.AllowedPaths {
		if strings.HasPrefix(path, filepath.Clean(allowed)) {
			return true
		}
	}

	return false
}

// ResourceLimits defines resource usage limits
type ResourceLimits struct {
	MaxMemory     int64
	MaxCPUTime    time.Duration
	MaxGoroutines int
}

// ResourceTracker tracks resource usage
type ResourceTracker struct {
	limits ResourceLimits

	mu             sync.RWMutex
	memoryUsage    int64
	goroutineCount int32
	cpuStartTime   time.Time
	cpuTime        time.Duration
}

// NewResourceTracker creates a new resource tracker
func NewResourceTracker(limits ResourceLimits) *ResourceTracker {
	return &ResourceTracker{
		limits: limits,
	}
}

// AllocateMemory attempts to allocate memory
func (rt *ResourceTracker) AllocateMemory(size int64) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.limits.MaxMemory > 0 && rt.memoryUsage+size > rt.limits.MaxMemory {
		return fmt.Errorf("memory limit exceeded: requested %d, used %d, limit %d",
			size, rt.memoryUsage, rt.limits.MaxMemory)
	}

	rt.memoryUsage += size
	return nil
}

// FreeMemory frees allocated memory
func (rt *ResourceTracker) FreeMemory(size int64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.memoryUsage -= size
	if rt.memoryUsage < 0 {
		rt.memoryUsage = 0
	}
}

// GetMemoryUsage returns current memory usage
func (rt *ResourceTracker) GetMemoryUsage() int64 {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.memoryUsage
}

// StartGoroutine attempts to start a new goroutine
func (rt *ResourceTracker) StartGoroutine() error {
	count := atomic.AddInt32(&rt.goroutineCount, 1)
	if rt.limits.MaxGoroutines > 0 && int(count) > rt.limits.MaxGoroutines {
		atomic.AddInt32(&rt.goroutineCount, -1)
		return fmt.Errorf("goroutine limit exceeded: limit %d", rt.limits.MaxGoroutines)
	}
	return nil
}

// EndGoroutine marks a goroutine as ended
func (rt *ResourceTracker) EndGoroutine() {
	atomic.AddInt32(&rt.goroutineCount, -1)
}

// GetGoroutineCount returns current goroutine count
func (rt *ResourceTracker) GetGoroutineCount() int {
	return int(atomic.LoadInt32(&rt.goroutineCount))
}

// StartCPUTracking starts tracking CPU time
func (rt *ResourceTracker) StartCPUTracking() {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.cpuStartTime = time.Now()
}

// UpdateCPUTime updates the tracked CPU time
func (rt *ResourceTracker) UpdateCPUTime() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if !rt.cpuStartTime.IsZero() {
		rt.cpuTime += time.Since(rt.cpuStartTime)
		rt.cpuStartTime = time.Now()
	}
}

// GetCPUTime returns the tracked CPU time
func (rt *ResourceTracker) GetCPUTime() time.Duration {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	elapsed := rt.cpuTime
	if !rt.cpuStartTime.IsZero() {
		elapsed += time.Since(rt.cpuStartTime)
	}

	return elapsed
}

// CheckCPULimit checks if CPU limit is exceeded
func (rt *ResourceTracker) CheckCPULimit() error {
	cpuTime := rt.GetCPUTime()
	if rt.limits.MaxCPUTime > 0 && cpuTime > rt.limits.MaxCPUTime {
		return fmt.Errorf("CPU time limit exceeded: used %v, limit %v", cpuTime, rt.limits.MaxCPUTime)
	}
	return nil
}

// NewSecureContext creates a new secure execution context
func NewSecureContext(parent context.Context, config ContextConfig) (context.Context, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(parent, config.MaxExecutionTime)

	// Store config
	ctx = context.WithValue(ctx, configKey, &config)

	// Create and store resource tracker
	tracker := NewResourceTracker(ResourceLimits{
		MaxMemory:     config.MaxMemory,
		MaxCPUTime:    config.MaxCPUTime,
		MaxGoroutines: config.MaxGoroutines,
	})
	ctx = context.WithValue(ctx, trackerKey, tracker)

	// Store security policy if provided
	if config.SecurityPolicy != nil {
		ctx = context.WithValue(ctx, policyKey, config.SecurityPolicy)
	}

	// Start a goroutine to monitor context cancellation
	go func() {
		<-ctx.Done()
		cancel()
	}()

	return ctx, nil
}

// GetContextConfig retrieves the config from a context
func GetContextConfig(ctx context.Context) *ContextConfig {
	if v := ctx.Value(configKey); v != nil {
		return v.(*ContextConfig)
	}
	return nil
}

// GetResourceTracker retrieves the resource tracker from a context
func GetResourceTracker(ctx context.Context) *ResourceTracker {
	if v := ctx.Value(trackerKey); v != nil {
		return v.(*ResourceTracker)
	}
	return nil
}

// GetSecurityPolicy retrieves the security policy from a context
func GetSecurityPolicy(ctx context.Context) *SecurityPolicy {
	if v := ctx.Value(policyKey); v != nil {
		return v.(*SecurityPolicy)
	}
	return nil
}

// ResourceMonitor monitors resource usage
type ResourceMonitor struct {
	ctx        context.Context
	interval   time.Duration
	violations []ResourceViolation
	mu         sync.Mutex
	stop       chan struct{}
	wg         sync.WaitGroup
}

// ResourceViolation represents a resource limit violation
type ResourceViolation struct {
	Type      string
	Message   string
	Timestamp time.Time
}

// StartResourceMonitor starts monitoring resources
func StartResourceMonitor(ctx context.Context, interval time.Duration) *ResourceMonitor {
	monitor := &ResourceMonitor{
		ctx:      ctx,
		interval: interval,
		stop:     make(chan struct{}),
	}

	monitor.wg.Add(1)
	go monitor.run()

	return monitor
}

// run is the main monitoring loop
func (m *ResourceMonitor) run() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkResources()
		}
	}
}

// checkResources checks current resource usage
func (m *ResourceMonitor) checkResources() {
	tracker := GetResourceTracker(m.ctx)
	if tracker == nil {
		return
	}

	config := GetContextConfig(m.ctx)
	if config == nil {
		return
	}

	// Check memory usage
	memUsage := tracker.GetMemoryUsage()
	if config.MaxMemory > 0 && memUsage > int64(float64(config.MaxMemory)*0.9) {
		m.addViolation(ResourceViolation{
			Type:      "memory",
			Message:   fmt.Sprintf("Memory usage high: %d/%d bytes", memUsage, config.MaxMemory),
			Timestamp: time.Now(),
		})
	}

	// Check CPU time
	if err := tracker.CheckCPULimit(); err != nil {
		m.addViolation(ResourceViolation{
			Type:      "cpu",
			Message:   err.Error(),
			Timestamp: time.Now(),
		})
	}

	// Check goroutine count
	goroutines := tracker.GetGoroutineCount()
	if config.MaxGoroutines > 0 && goroutines > int(float64(config.MaxGoroutines)*0.9) {
		m.addViolation(ResourceViolation{
			Type:      "goroutines",
			Message:   fmt.Sprintf("Goroutine count high: %d/%d", goroutines, config.MaxGoroutines),
			Timestamp: time.Now(),
		})
	}

	// Check actual memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	if config.MaxMemory > 0 && int64(memStats.Alloc) > config.MaxMemory {
		m.addViolation(ResourceViolation{
			Type:      "system_memory",
			Message:   fmt.Sprintf("System memory exceeded: %d/%d bytes", memStats.Alloc, config.MaxMemory),
			Timestamp: time.Now(),
		})
	}
}

// addViolation adds a violation to the list
func (m *ResourceMonitor) addViolation(v ResourceViolation) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.violations = append(m.violations, v)
}

// GetViolations returns all recorded violations
func (m *ResourceMonitor) GetViolations() []ResourceViolation {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]ResourceViolation, len(m.violations))
	copy(result, m.violations)
	return result
}

// Stop stops the monitor
func (m *ResourceMonitor) Stop() {
	close(m.stop)
	m.wg.Wait()
}

// CheckResourceLimits performs an immediate resource check
func CheckResourceLimits(ctx context.Context) error {
	tracker := GetResourceTracker(ctx)
	if tracker == nil {
		return nil
	}

	// Check CPU limit
	if err := tracker.CheckCPULimit(); err != nil {
		return err
	}

	// Check memory
	config := GetContextConfig(ctx)
	if config != nil && config.MaxMemory > 0 {
		usage := tracker.GetMemoryUsage()
		if usage > config.MaxMemory {
			return fmt.Errorf("memory limit exceeded: used %d, limit %d", usage, config.MaxMemory)
		}
	}

	// Check goroutines
	if config != nil && config.MaxGoroutines > 0 {
		count := tracker.GetGoroutineCount()
		if count > config.MaxGoroutines {
			return fmt.Errorf("goroutine limit exceeded: count %d, limit %d", count, config.MaxGoroutines)
		}
	}

	return nil
}
