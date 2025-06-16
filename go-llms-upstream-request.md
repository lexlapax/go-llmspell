# go-llms Upstream Feature Requests

This document tracks features that are missing from go-llms but needed for go-llmspell's bridge architecture. These should be implemented in go-llms first, then bridged in go-llmspell.

## Current Date: 2025-06-16

## Bridge-First Principle Compliance

**Rule**: If it's not in go-llms, we don't implement it in go-llmspell.

When we identify missing functionality during bridge implementation, we document it here for upstream contribution to go-llms.

---

## 1. Model Performance Analytics System

**Context**: Phase 1.4.6.1 - Add Model Performance Analytics  
**Status**: Missing from go-llms  
**Priority**: High  

### 1.1 Model Performance Tracking
- [ ] Implement model-specific performance metrics collection
- [ ] Add latency tracking per model/provider
- [ ] Implement token usage analytics per model
- [ ] Add cost tracking and calculation per model
- [ ] Create performance trend analysis over time
- [ ] Add anomaly detection for model performance

### 1.2 Suggested Implementation Location
- **Package**: `pkg/util/llmutil/analytics/`
- **Files**: 
  - `performance_tracker.go` - Core performance tracking
  - `model_metrics.go` - Model-specific metrics
  - `cost_calculator.go` - Token cost calculations
  - `trend_analyzer.go` - Performance trend analysis
  - `anomaly_detector.go` - Performance anomaly detection

### 1.3 Interface Design
```go
type ModelPerformanceTracker interface {
    TrackRequest(modelID string, latency time.Duration, inputTokens, outputTokens int, cost float64)
    GetModelPerformance(modelID string) (*ModelPerformanceReport, error)
    GetPerformanceTrends(modelID string, timeRange time.Duration) (*PerformanceTrends, error)
    DetectAnomalies(modelID string) ([]PerformanceAnomaly, error)
}
```

---

## 2. Model Recommendation Engine

**Context**: Phase 1.4.6.2 - Add Model Recommendation Engine  
**Status**: Missing from go-llms  
**Priority**: High  

### 2.1 Model Selection Algorithms
- [ ] Implement capability-based model matching
- [ ] Add task-specific model recommendations
- [ ] Create cost/performance optimization algorithms
- [ ] Implement multi-criteria decision making (MCDM)
- [ ] Add recommendation explanations and reasoning
- [ ] Support A/B testing for model selection

### 2.2 Suggested Implementation Location
- **Package**: `pkg/util/llmutil/recommendation/`
- **Files**:
  - `recommender.go` - Core recommendation engine
  - `capability_matcher.go` - Capability-based matching
  - `cost_optimizer.go` - Cost/performance optimization
  - `task_classifier.go` - Task-specific recommendations
  - `explanation_generator.go` - Recommendation explanations

### 2.3 Interface Design
```go
type ModelRecommender interface {
    FindModelsWithCapabilities(capabilities []string) ([]ModelRecommendation, error)
    RecommendForTask(taskType string, constraints ModelConstraints) ([]ModelRecommendation, error)
    OptimizeForCostPerformance(requirements PerformanceRequirements) ([]ModelRecommendation, error)
    ExplainRecommendation(recommendation ModelRecommendation) (string, error)
}
```

---

## 3. Model Catalog Export System

**Context**: Phase 1.4.6.3 - Add Model Catalog Export  
**Status**: Missing from go-llms  
**Priority**: Medium  

### 3.1 Documentation Generation
- [ ] Implement OpenAPI specification export for model catalog
- [ ] Add interactive documentation generation
- [ ] Include pricing information in exports
- [ ] Generate capability matrices and comparison charts
- [ ] Support multiple export formats (JSON, YAML, Markdown)
- [ ] Add custom export format support

### 3.2 Suggested Implementation Location
- **Package**: `pkg/util/llmutil/catalog/`
- **Files**:
  - `exporter.go` - Core catalog export functionality
  - `openapi_generator.go` - OpenAPI specification generation
  - `docs_generator.go` - Interactive documentation
  - `comparison_generator.go` - Capability comparison charts
  - `format_converter.go` - Multi-format export support

### 3.3 Interface Design
```go
type CatalogExporter interface {
    ExportToOpenAPI(models []Model, version string) (*OpenAPISpec, error)
    GenerateInteractiveDocs(models []Model) (*InteractiveDocs, error)
    ExportComparisonMatrix(models []Model) (*ComparisonMatrix, error)
    ExportToFormat(models []Model, format ExportFormat) ([]byte, error)
}
```

---

## 4. Enhanced Metrics Integration

**Context**: Extend existing metrics system for model-specific tracking  
**Status**: Partially exists in go-llms  
**Priority**: Medium  

### 4.1 Model-Specific Metrics
- [ ] Extend existing `pkg/util/metrics/` with model-specific counters
- [ ] Add model performance histograms
- [ ] Implement provider-specific metric aggregation
- [ ] Add real-time metric streaming capabilities

### 4.2 Suggested Enhancement Location
- **Package**: `pkg/util/metrics/` (extend existing)
- **Files**:
  - `model_metrics.go` - Model-specific metric types
  - `provider_aggregator.go` - Provider-level aggregation
  - `streaming_metrics.go` - Real-time metric streaming

---

## 5. Integration with Existing Systems

**Context**: Ensure new features integrate with existing go-llms architecture  
**Status**: Design consideration  
**Priority**: High  

### 5.1 Integration Points
- [ ] Integrate with existing `ModelInfoService`
- [ ] Leverage existing metrics registry
- [ ] Extend model inventory with performance data
- [ ] Integrate with provider metadata system
- [ ] Use existing event system for metric collection

### 5.2 Backward Compatibility
- [ ] Ensure all new features are optional and don't break existing APIs
- [ ] Provide configuration options to enable/disable analytics
- [ ] Maintain existing model info interfaces

---

## 6. Testing Requirements

**Context**: Comprehensive testing for new features  
**Status**: Required for all new features  
**Priority**: High  

### 6.1 Test Coverage Requirements
- [ ] Unit tests for all new interfaces and implementations
- [ ] Integration tests with existing model info system
- [ ] Performance benchmarks for analytics overhead
- [ ] Mock implementations for testing
- [ ] Example usage in `cmd/examples/`

### 6.2 Test Location
- **Package**: Follow existing go-llms testing patterns
- **Files**: `*_test.go` files alongside implementations
- **Benchmarks**: `tests/benchmarks/` directory
- **Integration**: `tests/integration/` directory

---

## 7. Implementation Priority

**Recommended Order**:

1. **Model Performance Analytics** (Phase 1.4.6.1)
   - Foundation for other features
   - Extends existing metrics system
   - High value for users

2. **Model Recommendation Engine** (Phase 1.4.6.2)
   - Builds on performance analytics
   - Complex algorithms requiring careful design
   - High impact on user experience

3. **Model Catalog Export** (Phase 1.4.6.3)
   - Documentation and tooling feature
   - Lower complexity
   - Can be implemented independently

4. **Enhanced Metrics Integration** (Phase 4)
   - Supports all other features
   - Extends existing system
   - Continuous improvement

---

## 8. Notes for go-llms Contributors

### 8.1 Design Principles
- Follow existing go-llms patterns and conventions
- Maintain backward compatibility
- Use dependency injection where appropriate
- Leverage existing interfaces and abstractions

### 8.2 Dependencies
- Build on existing `pkg/util/metrics/` system
- Integrate with `pkg/util/llmutil/modelinfo/` 
- Use existing provider metadata where available
- Follow existing error handling patterns

### 8.3 Documentation
- Add comprehensive godoc comments
- Include usage examples
- Update main README.md with new features
- Add to `docs/` if complex features require detailed documentation

---

## 9. go-llmspell Bridge Implementation Plan

**After go-llms Implementation**:

Once these features are available in go-llms, go-llmspell will implement corresponding bridges:

- `ModelPerformanceBridge` - Bridge performance analytics
- `ModelRecommendationBridge` - Bridge recommendation engine  
- `ModelCatalogBridge` - Bridge catalog export functionality

**Bridge Location**: `pkg/bridge/modelinfo/` (extend existing)

---

## 10. Review and Approval Process

### 10.1 go-llms Review
- [ ] Feature design review with go-llms maintainers
- [ ] API design approval
- [ ] Implementation review
- [ ] Testing and documentation review

### 10.2 go-llmspell Integration
- [ ] Bridge implementation after go-llms release
- [ ] Integration testing
- [ ] Documentation updates
- [ ] Phase 1.4.6 completion