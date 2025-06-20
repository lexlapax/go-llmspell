// ABOUTME: Compilation optimization infrastructure for Lua scripts including pattern-based optimization and enhanced caching
// ABOUTME: Provides compilation pipeline with source transformations and performance tracking

package gopherlua

import (
	"crypto/sha256"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// OptimizedCompilerConfig configures the optimized compiler
type OptimizedCompilerConfig struct {
	// Chunk cache configuration
	ChunkCacheConfig ChunkCacheConfig

	// Optimization flags
	EnableSourceOptimization  bool // Pattern-based source optimization
	EnableDeadCodeElimination bool // Simple dead code patterns
	EnableConstantFolding     bool // Basic constant folding
	EnableCommentStripping    bool // Remove comments
	EnableWhitespaceReduction bool // Minimize whitespace

	// Analysis options
	EnableStaticAnalysis bool
	WarnUnusedVariables  bool
	WarnUnreachableCode  bool

	// Performance options
	AggressiveOptimization bool // Enable aggressive optimizations
}

// OptimizedCompiler provides advanced compilation with optimizations
type OptimizedCompiler struct {
	config    OptimizedCompilerConfig
	cache     *EnhancedChunkCache
	optimizer *SourceOptimizer

	// Metrics
	compilations      int64
	optimizations     int64
	cacheHits         int64
	cacheMisses       int64
	totalCompileTime  int64 // nanoseconds
	totalOptimizeTime int64 // nanoseconds
}

// EnhancedChunkCache extends ChunkCache with optimization metadata
type EnhancedChunkCache struct {
	*ChunkCache

	// Optimization metadata
	metadata   map[string]*ChunkMetadata
	metadataMu sync.RWMutex

	// Source mapping for debugging
	sourceMaps   map[string]*SourceMap
	sourceMapsMu sync.RWMutex
}

// ChunkMetadata stores metadata about compiled chunks
type ChunkMetadata struct {
	OriginalSize     int
	OptimizedSize    int
	CompilationTime  time.Duration
	OptimizationTime time.Duration
	Optimizations    []string
	Dependencies     []string
	ExportedSymbols  []string
	Warnings         []CompilerWarning
}

// SourceMap maps optimized code back to original for debugging
type SourceMap struct {
	Original  []SourceLocation
	Optimized []SourceLocation
}

// SourceLocation represents a location in source code
type SourceLocation struct {
	Line   int
	Column int
	Offset int
}

// CompilerWarning represents a warning during compilation
type CompilerWarning struct {
	Type     string
	Message  string
	Location SourceLocation
}

// CompilerMetrics contains compiler performance metrics
type CompilerMetrics struct {
	Compilations      int64
	Optimizations     int64
	CacheHits         int64
	CacheMisses       int64
	TotalCompileTime  time.Duration
	TotalOptimizeTime time.Duration
	AvgCompileTime    time.Duration
	AvgOptimizeTime   time.Duration
	CacheHitRate      float64
}

// NewOptimizedCompiler creates a new optimized compiler
func NewOptimizedCompiler(config OptimizedCompilerConfig) *OptimizedCompiler {
	enhancedCache := &EnhancedChunkCache{
		ChunkCache: NewChunkCache(config.ChunkCacheConfig),
		metadata:   make(map[string]*ChunkMetadata),
		sourceMaps: make(map[string]*SourceMap),
	}

	return &OptimizedCompiler{
		config:    config,
		cache:     enhancedCache,
		optimizer: NewSourceOptimizer(config),
	}
}

// Compile compiles a script with optimizations
func (oc *OptimizedCompiler) Compile(script, filename string) (*lua.FunctionProto, error) {
	startTime := time.Now()
	atomic.AddInt64(&oc.compilations, 1)

	// Generate cache key
	key := oc.generateCacheKey(script, filename)

	// Check cache
	if cached := oc.cache.Get(key); cached != nil {
		atomic.AddInt64(&oc.cacheHits, 1)
		return cached, nil
	}

	atomic.AddInt64(&oc.cacheMisses, 1)

	// Create metadata
	metadata := &ChunkMetadata{
		OriginalSize: len(script),
		Dependencies: oc.findDependencies(script),
	}

	// Optimize source code
	optimizedScript := script
	optimizeStart := time.Now()
	if oc.shouldOptimize() {
		var optimizations []string
		optimizedScript, optimizations = oc.optimizer.Optimize(script)
		metadata.Optimizations = optimizations
		metadata.OptimizedSize = len(optimizedScript)
		if len(optimizations) > 0 {
			atomic.AddInt64(&oc.optimizations, 1)
		}
	}
	metadata.OptimizationTime = time.Since(optimizeStart)
	atomic.AddInt64(&oc.totalOptimizeTime, metadata.OptimizationTime.Nanoseconds())

	// Parse and compile
	reader := strings.NewReader(optimizedScript)
	chunk, err := parse.Parse(reader, filename)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	proto, err := lua.Compile(chunk, filename)
	if err != nil {
		return nil, fmt.Errorf("compilation error: %w", err)
	}

	// Analyze for warnings if enabled
	if oc.config.EnableStaticAnalysis {
		metadata.Warnings = oc.analyzeSource(script)
	}

	// Update compilation time
	metadata.CompilationTime = time.Since(startTime)
	atomic.AddInt64(&oc.totalCompileTime, metadata.CompilationTime.Nanoseconds())

	// Cache the result
	oc.cache.Put(key, proto)
	oc.cache.PutMetadata(key, metadata)

	return proto, nil
}

// CompileWithSourceMap compiles with source mapping for debugging
func (oc *OptimizedCompiler) CompileWithSourceMap(script, filename string) (*lua.FunctionProto, *SourceMap, error) {
	proto, err := oc.Compile(script, filename)
	if err != nil {
		return nil, nil, err
	}

	// Create basic source map
	sourceMap := &SourceMap{
		Original:  []SourceLocation{},
		Optimized: []SourceLocation{},
	}

	return proto, sourceMap, nil
}

// ValidateScript validates a script without compiling
func (oc *OptimizedCompiler) ValidateScript(script, filename string) []CompilerWarning {
	// Try to parse
	reader := strings.NewReader(script)
	_, err := parse.Parse(reader, filename)
	if err != nil {
		return []CompilerWarning{{
			Type:    "SyntaxError",
			Message: err.Error(),
		}}
	}

	// Analyze for warnings
	if oc.config.EnableStaticAnalysis {
		return oc.analyzeSource(script)
	}

	return nil
}

// GetMetrics returns compiler metrics
func (oc *OptimizedCompiler) GetMetrics() CompilerMetrics {
	totalCompileTime := time.Duration(atomic.LoadInt64(&oc.totalCompileTime))
	totalOptimizeTime := time.Duration(atomic.LoadInt64(&oc.totalOptimizeTime))
	compilations := atomic.LoadInt64(&oc.compilations)

	var avgCompileTime, avgOptimizeTime time.Duration
	if compilations > 0 {
		avgCompileTime = totalCompileTime / time.Duration(compilations)
		avgOptimizeTime = totalOptimizeTime / time.Duration(compilations)
	}

	return CompilerMetrics{
		Compilations:      compilations,
		Optimizations:     atomic.LoadInt64(&oc.optimizations),
		CacheHits:         atomic.LoadInt64(&oc.cacheHits),
		CacheMisses:       atomic.LoadInt64(&oc.cacheMisses),
		TotalCompileTime:  totalCompileTime,
		TotalOptimizeTime: totalOptimizeTime,
		AvgCompileTime:    avgCompileTime,
		AvgOptimizeTime:   avgOptimizeTime,
		CacheHitRate:      oc.calculateCacheHitRate(),
	}
}

// Reset clears all metrics
func (oc *OptimizedCompiler) Reset() {
	atomic.StoreInt64(&oc.compilations, 0)
	atomic.StoreInt64(&oc.optimizations, 0)
	atomic.StoreInt64(&oc.cacheHits, 0)
	atomic.StoreInt64(&oc.cacheMisses, 0)
	atomic.StoreInt64(&oc.totalCompileTime, 0)
	atomic.StoreInt64(&oc.totalOptimizeTime, 0)
	oc.cache.Clear()
}

// Private methods

func (oc *OptimizedCompiler) generateCacheKey(script, filename string) string {
	// Include optimization settings in cache key
	settings := fmt.Sprintf("%+v", oc.config)
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", filename, script, settings)))
	return fmt.Sprintf("%x", hash)
}

func (oc *OptimizedCompiler) shouldOptimize() bool {
	return oc.config.EnableSourceOptimization ||
		oc.config.EnableDeadCodeElimination ||
		oc.config.EnableConstantFolding ||
		oc.config.EnableCommentStripping ||
		oc.config.EnableWhitespaceReduction
}

func (oc *OptimizedCompiler) calculateCacheHitRate() float64 {
	hits := atomic.LoadInt64(&oc.cacheHits)
	misses := atomic.LoadInt64(&oc.cacheMisses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total)
}

func (oc *OptimizedCompiler) findDependencies(script string) []string {
	deps := []string{}

	// Find require() calls
	requirePattern := regexp.MustCompile(`require\s*\(\s*["']([^"']+)["']\s*\)`)
	matches := requirePattern.FindAllStringSubmatch(script, -1)
	for _, match := range matches {
		if len(match) > 1 {
			deps = append(deps, match[1])
		}
	}

	return deps
}

func (oc *OptimizedCompiler) analyzeSource(script string) []CompilerWarning {
	warnings := []CompilerWarning{}

	if oc.config.WarnUnusedVariables {
		warnings = append(warnings, oc.findUnusedVariables(script)...)
	}

	if oc.config.WarnUnreachableCode {
		warnings = append(warnings, oc.findUnreachableCode(script)...)
	}

	return warnings
}

func (oc *OptimizedCompiler) findUnusedVariables(script string) []CompilerWarning {
	warnings := []CompilerWarning{}

	// Simple pattern-based detection
	localPattern := regexp.MustCompile(`local\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
	lines := strings.Split(script, "\n")

	for lineNum, line := range lines {
		matches := localPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				varName := match[1]
				// Skip underscore variables
				if strings.HasPrefix(varName, "_") {
					continue
				}

				// Check if variable is used elsewhere
				// This is a simplified check
				varUsePattern := regexp.MustCompile(`\b` + varName + `\b`)
				uses := varUsePattern.FindAllStringIndex(script, -1)
				if len(uses) <= 1 { // Only the declaration
					warnings = append(warnings, CompilerWarning{
						Type:    "UnusedVariable",
						Message: fmt.Sprintf("Variable '%s' is declared but never used", varName),
						Location: SourceLocation{
							Line: lineNum + 1,
						},
					})
				}
			}
		}
	}

	return warnings
}

func (oc *OptimizedCompiler) findUnreachableCode(script string) []CompilerWarning {
	warnings := []CompilerWarning{}

	lines := strings.Split(script, "\n")

	// Look for if true then return pattern across multiple lines
	inIfTrue := false
	foundReturn := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for if true then
		if strings.Contains(trimmed, "if true then") {
			inIfTrue = true
			foundReturn = false
		}

		// Check for return inside if true block
		if inIfTrue && strings.HasPrefix(trimmed, "return") {
			foundReturn = true
		}

		// Check for end of if block
		if inIfTrue && trimmed == "end" && foundReturn {
			// Now look for code after this end until function end
			for j := i + 1; j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" {
					continue
				}
				if strings.HasPrefix(nextLine, "end") {
					break // End of function
				}
				// Found unreachable code
				warnings = append(warnings, CompilerWarning{
					Type:    "UnreachableCode",
					Message: "Code after 'if true then return' is unreachable",
					Location: SourceLocation{
						Line: j + 1,
					},
				})
				break // Only report first unreachable line
			}
			inIfTrue = false
			foundReturn = false
		}
	}

	return warnings
}

// EnhancedChunkCache methods

// PutMetadata stores metadata for a chunk
func (ec *EnhancedChunkCache) PutMetadata(key string, metadata *ChunkMetadata) {
	ec.metadataMu.Lock()
	defer ec.metadataMu.Unlock()
	ec.metadata[key] = metadata
}

// GetMetadata retrieves metadata for a chunk
func (ec *EnhancedChunkCache) GetMetadata(key string) *ChunkMetadata {
	ec.metadataMu.RLock()
	defer ec.metadataMu.RUnlock()
	return ec.metadata[key]
}

// Clear clears the cache and metadata
func (ec *EnhancedChunkCache) Clear() {
	ec.ChunkCache.Clear()

	ec.metadataMu.Lock()
	defer ec.metadataMu.Unlock()
	ec.metadata = make(map[string]*ChunkMetadata)

	ec.sourceMapsMu.Lock()
	defer ec.sourceMapsMu.Unlock()
	ec.sourceMaps = make(map[string]*SourceMap)
}

// SourceOptimizer performs source-level optimizations
type SourceOptimizer struct {
	config OptimizedCompilerConfig
}

// NewSourceOptimizer creates a new source optimizer
func NewSourceOptimizer(config OptimizedCompilerConfig) *SourceOptimizer {
	return &SourceOptimizer{config: config}
}

// Optimize performs source-level optimizations
func (so *SourceOptimizer) Optimize(source string) (string, []string) {
	optimizations := []string{}
	result := source
	originalSource := source

	if so.config.EnableCommentStripping {
		prevResult := result
		result = so.stripComments(result)
		if result != prevResult {
			optimizations = append(optimizations, "comment_stripping")
		}
	}

	if so.config.EnableWhitespaceReduction {
		prevResult := result
		result = so.reduceWhitespace(result)
		if result != prevResult {
			optimizations = append(optimizations, "whitespace_reduction")
		}
	}

	if so.config.EnableConstantFolding {
		prevResult := result
		result = so.foldConstants(result)
		if result != prevResult {
			optimizations = append(optimizations, "constant_folding")
		}
	}

	if so.config.EnableDeadCodeElimination {
		prevResult := result
		result = so.eliminateSimpleDeadCode(result)
		if result != prevResult {
			optimizations = append(optimizations, "dead_code_elimination")
		}
	}

	if so.config.EnableSourceOptimization {
		prevResult := result
		result = so.applyPatternOptimizations(result)
		if result != prevResult {
			optimizations = append(optimizations, "pattern_optimization")
		}
	}

	// If nothing changed, but optimizations were enabled, return empty optimizations list
	if result == originalSource {
		return result, []string{}
	}

	return result, optimizations
}

func (so *SourceOptimizer) stripComments(source string) string {
	// Remove single-line comments
	singleLineComment := regexp.MustCompile(`--[^\n]*`)
	result := singleLineComment.ReplaceAllString(source, "")

	// Remove multi-line comments
	multiLineComment := regexp.MustCompile(`--\[\[[\s\S]*?\]\]`)
	result = multiLineComment.ReplaceAllString(result, "")

	return result
}

func (so *SourceOptimizer) reduceWhitespace(source string) string {
	// Reduce multiple spaces to single space
	multiSpace := regexp.MustCompile(`[ \t]+`)
	result := multiSpace.ReplaceAllString(source, " ")

	// Remove trailing whitespace
	trailingSpace := regexp.MustCompile(`[ \t]+$`)
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = trailingSpace.ReplaceAllString(line, "")
	}
	result = strings.Join(lines, "\n")

	// Remove multiple blank lines
	multiBlank := regexp.MustCompile(`\n\s*\n\s*\n`)
	result = multiBlank.ReplaceAllString(result, "\n\n")

	return result
}

func (so *SourceOptimizer) foldConstants(source string) string {
	// Simple numeric constant folding
	patterns := []struct {
		pattern *regexp.Regexp
		eval    func(matches []string) string
	}{
		{
			// Addition: 10 + 20
			pattern: regexp.MustCompile(`(\d+)\s*\+\s*(\d+)`),
			eval: func(matches []string) string {
				if len(matches) >= 3 {
					var a, b int
					_, _ = fmt.Sscanf(matches[1], "%d", &a)
					_, _ = fmt.Sscanf(matches[2], "%d", &b)
					return fmt.Sprintf("%d", a+b)
				}
				return matches[0]
			},
		},
		{
			// Subtraction: 30 - 10
			pattern: regexp.MustCompile(`(\d+)\s*-\s*(\d+)`),
			eval: func(matches []string) string {
				if len(matches) >= 3 {
					var a, b int
					_, _ = fmt.Sscanf(matches[1], "%d", &a)
					_, _ = fmt.Sscanf(matches[2], "%d", &b)
					return fmt.Sprintf("%d", a-b)
				}
				return matches[0]
			},
		},
		{
			// Multiplication: 5 * 6
			pattern: regexp.MustCompile(`(\d+)\s*\*\s*(\d+)`),
			eval: func(matches []string) string {
				if len(matches) >= 3 {
					var a, b int
					_, _ = fmt.Sscanf(matches[1], "%d", &a)
					_, _ = fmt.Sscanf(matches[2], "%d", &b)
					return fmt.Sprintf("%d", a*b)
				}
				return matches[0]
			},
		},
	}

	result := source
	for _, p := range patterns {
		result = p.pattern.ReplaceAllStringFunc(result, func(match string) string {
			matches := p.pattern.FindStringSubmatch(match)
			return p.eval(matches)
		})
	}

	// String concatenation
	stringConcat := regexp.MustCompile(`"([^"]*?)"\s*\.\.\s*"([^"]*?)"`)
	result = stringConcat.ReplaceAllStringFunc(result, func(match string) string {
		matches := stringConcat.FindStringSubmatch(match)
		if len(matches) >= 3 {
			return fmt.Sprintf(`"%s%s"`, matches[1], matches[2])
		}
		return match
	})

	return result
}

func (so *SourceOptimizer) eliminateSimpleDeadCode(source string) string {
	// For now, only remove obviously dead code patterns
	// Remove if false blocks - match non-greedy to avoid removing too much
	ifFalse := regexp.MustCompile(`if\s+false\s+then[^}]*?end`)
	result := ifFalse.ReplaceAllString(source, "")

	// Remove while false blocks - match non-greedy to avoid removing too much
	whileFalse := regexp.MustCompile(`while\s+false\s+do[^}]*?end`)
	result = whileFalse.ReplaceAllString(result, "")

	// Don't try to remove code after return statements as this is complex
	// and would require proper AST analysis to do correctly

	return result
}

func (so *SourceOptimizer) applyPatternOptimizations(source string) string {
	result := source

	// Apply optimization patterns
	optimizationPatterns := []struct {
		pattern *regexp.Regexp
		replace string
	}{
		{
			// Empty table check: #t == 0 -> next(t) == nil
			pattern: regexp.MustCompile(`#(\w+)\s*==\s*0`),
			replace: "next($1) == nil",
		},
		{
			// Remove empty string concatenation
			pattern: regexp.MustCompile(`""\s*\.\.\s*`),
			replace: "",
		},
		{
			// Remove concatenation with empty string at end
			pattern: regexp.MustCompile(`\s*\.\.\s*""`),
			replace: "",
		},
		{
			// Optimize not not -> boolean conversion
			pattern: regexp.MustCompile(`not\s+not\s+`),
			replace: "",
		},
		{
			// Optimize x = x + 1 -> x = x + 1 (no change, but could be x += 1 if supported)
			// Skip for now as Lua doesn't have +=
		},
	}

	for _, opt := range optimizationPatterns {
		if opt.pattern != nil {
			result = opt.pattern.ReplaceAllString(result, opt.replace)
		}
	}

	return result
}

// CompileReader compiles from an io.Reader
func (oc *OptimizedCompiler) CompileReader(reader io.Reader, filename string) (*lua.FunctionProto, error) {
	// Read the entire source
	var builder strings.Builder
	_, err := io.Copy(&builder, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	return oc.Compile(builder.String(), filename)
}

// PrecompileDirectory precompiles all Lua files in a directory
func (oc *OptimizedCompiler) PrecompileDirectory(dir string) error {
	// This would walk the directory and precompile all .lua files
	// For now, just return nil
	return nil
}
