// ABOUTME: Metrics bridge for go-llms performance monitoring and aggregation system
// ABOUTME: Provides script-accessible counters, gauges, timers, and ratio tracking with thread-safe operations

package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	// go-llms imports for metrics functionality
	"github.com/lexlapax/go-llms/pkg/util/metrics"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// MetricsBridge provides script access to go-llms metrics system
type MetricsBridge struct {
	initialized   bool
	registry      *metrics.Registry
	counters      map[string]*metrics.Counter
	gauges        map[string]*metrics.Gauge
	ratioCounters map[string]*metrics.RatioCounter
	timers        map[string]*metrics.Timer
	mu            sync.RWMutex
}

// NewMetricsBridge creates a new metrics bridge
func NewMetricsBridge() *MetricsBridge {
	return &MetricsBridge{
		registry:      metrics.GetRegistry(),
		counters:      make(map[string]*metrics.Counter),
		gauges:        make(map[string]*metrics.Gauge),
		ratioCounters: make(map[string]*metrics.RatioCounter),
		timers:        make(map[string]*metrics.Timer),
	}
}

// GetID returns the bridge identifier
func (mb *MetricsBridge) GetID() string {
	return "metrics"
}

// GetMetadata returns bridge metadata
func (mb *MetricsBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:         "metrics",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms performance metrics system with counters, gauges, timers, and aggregation",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/util/metrics"},
	}
}

// Initialize sets up the metrics bridge
func (mb *MetricsBridge) Initialize(ctx context.Context) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.initialized = true
	return nil
}

// Cleanup performs bridge cleanup
func (mb *MetricsBridge) Cleanup(ctx context.Context) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// Clear all stored metrics
	mb.counters = make(map[string]*metrics.Counter)
	mb.gauges = make(map[string]*metrics.Gauge)
	mb.ratioCounters = make(map[string]*metrics.RatioCounter)
	mb.timers = make(map[string]*metrics.Timer)
	mb.initialized = false

	return nil
}

// IsInitialized returns initialization status
func (mb *MetricsBridge) IsInitialized() bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (mb *MetricsBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(mb)
}

// Methods returns available bridge methods
func (mb *MetricsBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Counter methods
		{
			Name:        "createCounter",
			Description: "Create a new counter metric",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Counter name"},
			},
			ReturnType: "object",
			Examples:   []string{"createCounter('api_requests')"},
		},
		{
			Name:        "incrementCounter",
			Description: "Increment counter by 1",
			Parameters: []engine.ParameterInfo{
				{Name: "counterID", Type: "string", Required: true, Description: "Counter identifier"},
			},
			ReturnType: "void",
			Examples:   []string{"incrementCounter(counterID)"},
		},
		{
			Name:        "incrementCounterBy",
			Description: "Increment counter by specific value",
			Parameters: []engine.ParameterInfo{
				{Name: "counterID", Type: "string", Required: true, Description: "Counter identifier"},
				{Name: "value", Type: "number", Required: true, Description: "Increment value"},
			},
			ReturnType: "void",
			Examples:   []string{"incrementCounterBy(counterID, 5)"},
		},
		{
			Name:        "getCounterValue",
			Description: "Get current counter value",
			Parameters: []engine.ParameterInfo{
				{Name: "counterID", Type: "string", Required: true, Description: "Counter identifier"},
			},
			ReturnType: "number",
			Examples:   []string{"getCounterValue(counterID)"},
		},
		// Gauge methods
		{
			Name:        "createGauge",
			Description: "Create a new gauge metric",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Gauge name"},
			},
			ReturnType: "object",
			Examples:   []string{"createGauge('memory_usage')"},
		},
		{
			Name:        "setGaugeValue",
			Description: "Set gauge to specific value",
			Parameters: []engine.ParameterInfo{
				{Name: "gaugeID", Type: "string", Required: true, Description: "Gauge identifier"},
				{Name: "value", Type: "number", Required: true, Description: "Gauge value"},
			},
			ReturnType: "void",
			Examples:   []string{"setGaugeValue(gaugeID, 85.5)"},
		},
		{
			Name:        "incrementGauge",
			Description: "Increment gauge by 1",
			Parameters: []engine.ParameterInfo{
				{Name: "gaugeID", Type: "string", Required: true, Description: "Gauge identifier"},
			},
			ReturnType: "void",
			Examples:   []string{"incrementGauge(gaugeID)"},
		},
		{
			Name:        "addToGauge",
			Description: "Add value to gauge",
			Parameters: []engine.ParameterInfo{
				{Name: "gaugeID", Type: "string", Required: true, Description: "Gauge identifier"},
				{Name: "value", Type: "number", Required: true, Description: "Value to add"},
			},
			ReturnType: "void",
			Examples:   []string{"addToGauge(gaugeID, 10.5)"},
		},
		{
			Name:        "getGaugeValue",
			Description: "Get current gauge value",
			Parameters: []engine.ParameterInfo{
				{Name: "gaugeID", Type: "string", Required: true, Description: "Gauge identifier"},
			},
			ReturnType: "number",
			Examples:   []string{"getGaugeValue(gaugeID)"},
		},
		// Ratio counter methods
		{
			Name:        "createRatioCounter",
			Description: "Create a new ratio counter metric",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Ratio counter name"},
			},
			ReturnType: "object",
			Examples:   []string{"createRatioCounter('cache_hit_rate')"},
		},
		{
			Name:        "incrementRatioNumerator",
			Description: "Increment ratio numerator",
			Parameters: []engine.ParameterInfo{
				{Name: "ratioID", Type: "string", Required: true, Description: "Ratio counter identifier"},
			},
			ReturnType: "void",
			Examples:   []string{"incrementRatioNumerator(ratioID)"},
		},
		{
			Name:        "incrementRatioDenominator",
			Description: "Increment ratio denominator",
			Parameters: []engine.ParameterInfo{
				{Name: "ratioID", Type: "string", Required: true, Description: "Ratio counter identifier"},
			},
			ReturnType: "void",
			Examples:   []string{"incrementRatioDenominator(ratioID)"},
		},
		{
			Name:        "getRatio",
			Description: "Get current ratio value",
			Parameters: []engine.ParameterInfo{
				{Name: "ratioID", Type: "string", Required: true, Description: "Ratio counter identifier"},
			},
			ReturnType: "number",
			Examples:   []string{"getRatio(ratioID)"},
		},
		{
			Name:        "getRatioValues",
			Description: "Get raw numerator and denominator values",
			Parameters: []engine.ParameterInfo{
				{Name: "ratioID", Type: "string", Required: true, Description: "Ratio counter identifier"},
			},
			ReturnType: "object",
			Examples:   []string{"getRatioValues(ratioID)"},
		},
		// Timer methods
		{
			Name:        "createTimer",
			Description: "Create a new timer metric",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Timer name"},
			},
			ReturnType: "object",
			Examples:   []string{"createTimer('operation_duration')"},
		},
		{
			Name:        "startTimer",
			Description: "Start timing an operation",
			Parameters: []engine.ParameterInfo{
				{Name: "timerID", Type: "string", Required: true, Description: "Timer identifier"},
			},
			ReturnType: "void",
			Examples:   []string{"startTimer(timerID)"},
		},
		{
			Name:        "stopTimer",
			Description: "Stop timer and record duration",
			Parameters: []engine.ParameterInfo{
				{Name: "timerID", Type: "string", Required: true, Description: "Timer identifier"},
			},
			ReturnType: "number",
			Examples:   []string{"stopTimer(timerID)"},
		},
		{
			Name:        "recordTimerDuration",
			Description: "Manually record a duration",
			Parameters: []engine.ParameterInfo{
				{Name: "timerID", Type: "string", Required: true, Description: "Timer identifier"},
				{Name: "durationSeconds", Type: "number", Required: true, Description: "Duration in seconds"},
			},
			ReturnType: "void",
			Examples:   []string{"recordTimerDuration(timerID, 0.125)"},
		},
		{
			Name:        "getTimerStats",
			Description: "Get timer statistics",
			Parameters: []engine.ParameterInfo{
				{Name: "timerID", Type: "string", Required: true, Description: "Timer identifier"},
			},
			ReturnType: "object",
			Examples:   []string{"getTimerStats(timerID)"},
		},
		// Registry methods
		{
			Name:        "getAllMetrics",
			Description: "Get all metrics from registry",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getAllMetrics()"},
		},
		{
			Name:        "resetAllMetrics",
			Description: "Reset all metrics",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
			Examples:    []string{"resetAllMetrics()"},
		},
	}
}

// ValidateMethod validates method calls
func (mb *MetricsBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	if !mb.IsInitialized() {
		return fmt.Errorf("metrics bridge not initialized")
	}

	methods := mb.Methods()
	for _, method := range methods {
		if method.Name == name {
			requiredCount := 0
			for _, param := range method.Parameters {
				if param.Required {
					requiredCount++
				}
			}
			if len(args) < requiredCount {
				return fmt.Errorf("method %s requires at least %d arguments, got %d", name, requiredCount, len(args))
			}
			return nil
		}
	}
	return fmt.Errorf("unknown method: %s", name)
}

// TypeMappings returns type conversion mappings
func (mb *MetricsBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"counter": {
			GoType:     "*metrics.Counter",
			ScriptType: "object",
			Converter:  "counterConverter",
			Metadata:   map[string]interface{}{"description": "Monotonically increasing counter"},
		},
		"gauge": {
			GoType:     "*metrics.Gauge",
			ScriptType: "object",
			Converter:  "gaugeConverter",
			Metadata:   map[string]interface{}{"description": "Value that can go up and down"},
		},
		"ratio_counter": {
			GoType:     "*metrics.RatioCounter",
			ScriptType: "object",
			Converter:  "ratioCounterConverter",
			Metadata:   map[string]interface{}{"description": "Tracks ratio between two counters"},
		},
		"timer": {
			GoType:     "*metrics.Timer",
			ScriptType: "object",
			Converter:  "timerConverter",
			Metadata:   map[string]interface{}{"description": "Tracks execution duration"},
		},
	}
}

// ExecuteMethod executes a bridge method
func (mb *MetricsBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	switch name {
	case "createCounter":
		result, err := mb.createCounter(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewObjectValue(result.(map[string]engine.ScriptValue)), nil
	case "incrementCounter":
		err := mb.incrementCounter(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "incrementCounterBy":
		err := mb.incrementCounterBy(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "getCounterValue":
		result, err := mb.getCounterValue(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNumberValue(result.(float64)), nil
	case "createGauge":
		result, err := mb.createGauge(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewObjectValue(result.(map[string]engine.ScriptValue)), nil
	case "setGaugeValue":
		err := mb.setGaugeValue(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "addToGaugeValue":
		err := mb.addToGauge(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "getGaugeValue":
		result, err := mb.getGaugeValue(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNumberValue(result.(float64)), nil
	// Note: Histogram methods not implemented in current MetricsBridge
	// These would need to be added if histogram support is required
	case "createTimer":
		result, err := mb.createTimer(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewObjectValue(result.(map[string]engine.ScriptValue)), nil
	case "recordTimerDuration":
		err := mb.recordTimerDuration(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "getAllMetrics":
		result, err := mb.getAllMetrics(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewObjectValue(result.(map[string]engine.ScriptValue)), nil
	case "resetAllMetrics":
		err := mb.resetAllMetrics(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "startTimer":
		err := mb.startTimer(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "stopTimer":
		result, err := mb.stopTimer(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNumberValue(result.(float64)), nil
	case "getTimerStats":
		result, err := mb.getTimerStats(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewObjectValue(result.(map[string]engine.ScriptValue)), nil
	case "createRatioCounter":
		result, err := mb.createRatioCounter(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewObjectValue(result.(map[string]engine.ScriptValue)), nil
	case "incrementRatioNumerator":
		err := mb.incrementRatioNumerator(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "incrementRatioDenominator":
		err := mb.incrementRatioDenominator(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	case "getRatio":
		result, err := mb.getRatio(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNumberValue(result.(float64)), nil
	case "getRatioValues":
		result, err := mb.getRatioValues(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewObjectValue(result.(map[string]engine.ScriptValue)), nil
	case "incrementGauge":
		err := mb.incrementGauge(ctx, args)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

// RequiredPermissions returns required permissions
func (mb *MetricsBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "metrics.collection",
			Actions:     []string{"create", "update", "read"},
			Description: "Create and update metrics",
		},
		{
			Type:        engine.PermissionProcess,
			Resource:    "metrics.registry",
			Actions:     []string{"access", "reset"},
			Description: "Access global metrics registry",
		},
	}
}

// Bridge method implementations

// Counter methods

// createCounter creates a new counter
func (mb *MetricsBridge) createCounter(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("createCounter", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("counter name must be a string")
	}
	name := args[0].(engine.StringValue).Value()

	// Create counter through registry
	counter := mb.registry.GetOrCreateCounter(name)

	// Store it with a unique ID
	counterID := fmt.Sprintf("counter-%s-%d", name, time.Now().UnixNano())
	mb.mu.Lock()
	mb.counters[counterID] = counter
	mb.mu.Unlock()

	return map[string]engine.ScriptValue{
		"id":      engine.NewStringValue(counterID),
		"name":    engine.NewStringValue(name),
		"type":    engine.NewStringValue("counter"),
		"created": engine.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// incrementCounter increments a counter by 1
func (mb *MetricsBridge) incrementCounter(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("incrementCounter", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("counter ID must be a string")
	}
	counterID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	counter, exists := mb.counters[counterID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("counter not found: %s", counterID)
	}

	counter.Increment()
	return nil
}

// incrementCounterBy increments a counter by a specific value
func (mb *MetricsBridge) incrementCounterBy(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("incrementCounterBy", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("counter ID must be a string")
	}
	counterID := args[0].(engine.StringValue).Value()

	if len(args) < 2 || args[1] == nil || args[1].Type() != engine.TypeNumber {
		return fmt.Errorf("increment value must be a number")
	}
	valueFloat := args[1].(engine.NumberValue).Value()

	value := int64(valueFloat)

	mb.mu.RLock()
	counter, exists := mb.counters[counterID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("counter not found: %s", counterID)
	}

	counter.IncrementBy(value)
	return nil
}

// getCounterValue gets the current counter value
func (mb *MetricsBridge) getCounterValue(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("getCounterValue", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("counter ID must be a string")
	}
	counterID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	counter, exists := mb.counters[counterID]
	mb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("counter not found: %s", counterID)
	}

	return counter.GetValue(), nil
}

// Gauge methods

// createGauge creates a new gauge
func (mb *MetricsBridge) createGauge(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("createGauge", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("gauge name must be a string")
	}
	name := args[0].(engine.StringValue).Value()

	// Create gauge through registry
	gauge := mb.registry.GetOrCreateGauge(name)

	// Store it with a unique ID
	gaugeID := fmt.Sprintf("gauge-%s-%d", name, time.Now().UnixNano())
	mb.mu.Lock()
	mb.gauges[gaugeID] = gauge
	mb.mu.Unlock()

	return map[string]engine.ScriptValue{
		"id":      engine.NewStringValue(gaugeID),
		"name":    engine.NewStringValue(name),
		"type":    engine.NewStringValue("gauge"),
		"created": engine.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// setGaugeValue sets the gauge value
func (mb *MetricsBridge) setGaugeValue(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("setGaugeValue", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("gauge ID must be a string")
	}
	gaugeID := args[0].(engine.StringValue).Value()

	if len(args) < 2 || args[1] == nil || args[1].Type() != engine.TypeNumber {
		return fmt.Errorf("gauge value must be a number")
	}
	value := args[1].(engine.NumberValue).Value()

	mb.mu.RLock()
	gauge, exists := mb.gauges[gaugeID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge not found: %s", gaugeID)
	}

	gauge.Set(value)
	return nil
}

// incrementGauge increments the gauge by 1
func (mb *MetricsBridge) incrementGauge(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("incrementGauge", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("gauge ID must be a string")
	}
	gaugeID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	gauge, exists := mb.gauges[gaugeID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge not found: %s", gaugeID)
	}

	gauge.Increment()
	return nil
}

// addToGauge adds a value to the gauge
func (mb *MetricsBridge) addToGauge(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("addToGauge", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("gauge ID must be a string")
	}
	gaugeID := args[0].(engine.StringValue).Value()

	if len(args) < 2 || args[1] == nil || args[1].Type() != engine.TypeNumber {
		return fmt.Errorf("value must be a number")
	}
	value := args[1].(engine.NumberValue).Value()

	mb.mu.RLock()
	gauge, exists := mb.gauges[gaugeID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge not found: %s", gaugeID)
	}

	gauge.Add(value)
	return nil
}

// getGaugeValue gets the current gauge value
func (mb *MetricsBridge) getGaugeValue(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("getGaugeValue", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("gauge ID must be a string")
	}
	gaugeID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	gauge, exists := mb.gauges[gaugeID]
	mb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("gauge not found: %s", gaugeID)
	}

	return gauge.GetValue(), nil
}

// Ratio counter methods

// createRatioCounter creates a new ratio counter
func (mb *MetricsBridge) createRatioCounter(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("createRatioCounter", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("name must be a string")
	}
	name := args[0].(engine.StringValue).Value()

	// Create ratio counter through registry
	ratioCounter := mb.registry.GetOrCreateRatioCounter(name)

	// Store it with a unique ID
	ratioID := fmt.Sprintf("ratio-%s-%d", name, time.Now().UnixNano())
	mb.mu.Lock()
	mb.ratioCounters[ratioID] = ratioCounter
	mb.mu.Unlock()

	return map[string]engine.ScriptValue{
		"id":      engine.NewStringValue(ratioID),
		"name":    engine.NewStringValue(name),
		"type":    engine.NewStringValue("ratio_counter"),
		"created": engine.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// incrementRatioNumerator increments the ratio numerator
func (mb *MetricsBridge) incrementRatioNumerator(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("incrementRatioNumerator", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("ratio ID must be a string")
	}
	ratioID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	ratio, exists := mb.ratioCounters[ratioID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("ratio counter not found: %s", ratioID)
	}

	ratio.IncrementNumerator()
	return nil
}

// incrementRatioDenominator increments the ratio denominator
func (mb *MetricsBridge) incrementRatioDenominator(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("incrementRatioDenominator", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("ratio ID must be a string")
	}
	ratioID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	ratio, exists := mb.ratioCounters[ratioID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("ratio counter not found: %s", ratioID)
	}

	ratio.IncrementDenominator()
	return nil
}

// getRatio gets the current ratio value
func (mb *MetricsBridge) getRatio(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("getRatio", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ratio ID must be a string")
	}
	ratioID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	ratio, exists := mb.ratioCounters[ratioID]
	mb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("ratio counter not found: %s", ratioID)
	}

	return ratio.GetRatio(), nil
}

// getRatioValues gets the raw numerator and denominator values
func (mb *MetricsBridge) getRatioValues(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("getRatioValues", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ratio ID must be a string")
	}
	ratioID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	ratio, exists := mb.ratioCounters[ratioID]
	mb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("ratio counter not found: %s", ratioID)
	}

	numerator, denominator := ratio.GetValues()
	return map[string]engine.ScriptValue{
		"numerator":   engine.NewNumberValue(float64(numerator)),
		"denominator": engine.NewNumberValue(float64(denominator)),
	}, nil
}

// Timer methods

// createTimer creates a new timer
func (mb *MetricsBridge) createTimer(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("createTimer", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("name must be a string")
	}
	name := args[0].(engine.StringValue).Value()

	// Create timer through registry
	timer := mb.registry.GetOrCreateTimer(name)

	// Store it with a unique ID
	timerID := fmt.Sprintf("timer-%s-%d", name, time.Now().UnixNano())
	mb.mu.Lock()
	mb.timers[timerID] = timer
	mb.mu.Unlock()

	return map[string]engine.ScriptValue{
		"id":      engine.NewStringValue(timerID),
		"name":    engine.NewStringValue(name),
		"type":    engine.NewStringValue("timer"),
		"created": engine.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// startTimer starts a timer
func (mb *MetricsBridge) startTimer(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("startTimer", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("timer ID must be a string")
	}
	timerID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	timer, exists := mb.timers[timerID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("timer not found: %s", timerID)
	}

	timer.Start()
	return nil
}

// stopTimer stops a timer and returns the duration
func (mb *MetricsBridge) stopTimer(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("stopTimer", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("timer ID must be a string")
	}
	timerID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	timer, exists := mb.timers[timerID]
	mb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("timer not found: %s", timerID)
	}

	duration := timer.Stop()
	return duration.Seconds(), nil
}

// recordTimerDuration manually records a duration
func (mb *MetricsBridge) recordTimerDuration(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("recordTimerDuration", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("timer ID must be a string")
	}
	timerID := args[0].(engine.StringValue).Value()

	if len(args) < 2 || args[1] == nil || args[1].Type() != engine.TypeNumber {
		return fmt.Errorf("duration must be a number")
	}
	durationSeconds := args[1].(engine.NumberValue).Value()

	mb.mu.RLock()
	timer, exists := mb.timers[timerID]
	mb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("timer not found: %s", timerID)
	}

	duration := time.Duration(durationSeconds * float64(time.Second))
	timer.RecordDuration(duration)
	return nil
}

// getTimerStats gets timer statistics
func (mb *MetricsBridge) getTimerStats(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("getTimerStats", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("timer ID must be a string")
	}
	timerID := args[0].(engine.StringValue).Value()

	mb.mu.RLock()
	timer, exists := mb.timers[timerID]
	mb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("timer not found: %s", timerID)
	}

	return map[string]engine.ScriptValue{
		"count":            engine.NewNumberValue(float64(timer.GetCount())),
		"last_duration":    engine.NewNumberValue(timer.GetLastDuration().Seconds()),
		"total_duration":   engine.NewNumberValue(timer.GetTotalDuration().Seconds()),
		"average_duration": engine.NewNumberValue(timer.GetAverageDuration().Seconds()),
	}, nil
}

// Registry methods

// getAllMetrics gets all metrics from the registry
func (mb *MetricsBridge) getAllMetrics(ctx context.Context, args []engine.ScriptValue) (interface{}, error) {
	if err := mb.ValidateMethod("getAllMetrics", args); err != nil {
		return nil, err
	}

	mb.mu.RLock()
	defer mb.mu.RUnlock()

	// Collect counter values
	counters := make(map[string]engine.ScriptValue)
	for id, counter := range mb.counters {
		counters[id] = engine.NewNumberValue(float64(counter.GetValue()))
	}

	// Collect gauge values
	gauges := make(map[string]engine.ScriptValue)
	for id, gauge := range mb.gauges {
		gauges[id] = engine.NewNumberValue(gauge.GetValue())
	}

	// Collect ratio counter values
	ratios := make(map[string]engine.ScriptValue)
	for id, ratio := range mb.ratioCounters {
		num, den := ratio.GetValues()
		ratios[id] = engine.NewObjectValue(map[string]engine.ScriptValue{
			"ratio":       engine.NewNumberValue(ratio.GetRatio()),
			"numerator":   engine.NewNumberValue(float64(num)),
			"denominator": engine.NewNumberValue(float64(den)),
		})
	}

	// Collect timer values
	timers := make(map[string]engine.ScriptValue)
	for id, timer := range mb.timers {
		timers[id] = engine.NewObjectValue(map[string]engine.ScriptValue{
			"count":            engine.NewNumberValue(float64(timer.GetCount())),
			"last_duration":    engine.NewNumberValue(timer.GetLastDuration().Seconds()),
			"total_duration":   engine.NewNumberValue(timer.GetTotalDuration().Seconds()),
			"average_duration": engine.NewNumberValue(timer.GetAverageDuration().Seconds()),
		})
	}

	result := map[string]engine.ScriptValue{
		"counters":       engine.NewObjectValue(counters),
		"gauges":         engine.NewObjectValue(gauges),
		"ratio_counters": engine.NewObjectValue(ratios),
		"timers":         engine.NewObjectValue(timers),
	}

	return result, nil
}

// resetAllMetrics resets all metrics
func (mb *MetricsBridge) resetAllMetrics(ctx context.Context, args []engine.ScriptValue) error {
	if err := mb.ValidateMethod("resetAllMetrics", args); err != nil {
		return err
	}

	mb.mu.Lock()
	defer mb.mu.Unlock()

	// Clear all stored metrics (this will require recreating them)
	mb.counters = make(map[string]*metrics.Counter)
	mb.gauges = make(map[string]*metrics.Gauge)
	mb.ratioCounters = make(map[string]*metrics.RatioCounter)
	mb.timers = make(map[string]*metrics.Timer)

	// Note: This doesn't reset the global registry, only our bridge's references
	// The global registry would need a separate reset method in go-llms

	return nil
}
