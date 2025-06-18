# Benchmarks

This directory contains performance benchmarks for go-llmspell.

## Running Benchmarks

To run all benchmarks:
```bash
make bench
```

To run specific benchmarks:
```bash
make bench-run BENCH=BenchmarkTypeChecking
```

To run benchmarks directly:
```bash
go test -bench=. -benchmem -tags=bench ./tests/benchmarks/...
```

## Benchmark Files

- `scriptvalue_benchmark_test.go` - Compares performance of interface{} vs ScriptValue type system

## Build Tag

All benchmark files use the `bench` build tag to exclude them from regular test runs. This ensures:
- `make test` doesn't run benchmarks
- Benchmarks are only executed when explicitly requested
- Better separation of concerns between unit tests and performance tests

## Writing New Benchmarks

When adding new benchmarks:

1. Place them in this directory
2. Add the build tag:
   ```go
   //go:build bench
   // +build bench
   ```
3. Follow the naming convention: `*_benchmark_test.go`
4. Import the package being benchmarked:
   ```go
   import "github.com/lexlapax/go-llmspell/pkg/..."
   ```