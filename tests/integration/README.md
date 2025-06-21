# Integration Tests

This directory contains comprehensive integration tests for the llmspell CLI tool.

## Structure

```
tests/integration/
├── README.md           # This file
├── suite_test.go       # Main test suite runner
├── helpers/            # Test helper utilities
│   └── helpers.go      # TestHelper implementation
└── commands/           # Command-specific tests
    ├── run_test.go     # Tests for run command
    ├── validate_test.go # Tests for validate command
    ├── new_test.go     # Tests for new command
    ├── repl_test.go    # Tests for REPL command
    ├── config_test.go  # Tests for config command
    ├── security_test.go # Tests for security command
    ├── engines_test.go # Tests for engines command
    ├── version_test.go # Tests for version command
    ├── debug_test.go   # Tests for debug command
    └── integration_test.go # Cross-command tests
```

## Running Tests

### Run all integration tests:
```bash
go test ./tests/integration/...
```

### Run with verbose output:
```bash
go test -v ./tests/integration/...
```

### Run specific command tests:
```bash
go test -v ./tests/integration/commands -run TestRunCommand
```

### Skip integration tests (short mode):
```bash
go test -short ./...
```

### Run with custom flags:
```bash
# Keep temporary files for debugging
go test ./tests/integration/... -integration.keep

# Use verbose output
go test ./tests/integration/... -integration.verbose

# Use specific binary
go test ./tests/integration/... -integration.bin=/path/to/llmspell
```

## Test Categories

### Command Tests
Each command has its own test file with comprehensive coverage:
- Basic functionality
- Flag combinations
- Error handling
- Edge cases

### Cross-Command Tests
`integration_test.go` contains tests for:
- Command interactions
- Configuration layering
- Signal handling
- Performance benchmarks
- Cross-platform compatibility

### Test Helpers
The `helpers` package provides:
- `TestHelper`: Main test utility
- Command execution helpers
- File creation utilities
- Assertion helpers
- Cleanup management

## Writing New Tests

### Basic Test Structure
```go
func TestMyFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    h := helpers.NewTestHelper(t)
    defer h.Cleanup()

    t.Run("subtest", func(t *testing.T) {
        // Create test files
        script := h.CreateSpell("test.lua", `print("test")`)
        
        // Run command
        stdout, stderr, err := h.RunCommand("run", script)
        
        // Assert results
        h.AssertSuccess(stdout, stderr, err)
        h.AssertOutput(stdout, "test")
    })
}
```

### Test Best Practices
1. Always check `testing.Short()` to allow skipping
2. Use `TestHelper` for consistency
3. Always call `defer h.Cleanup()`
4. Use subtests for organization
5. Test both success and failure cases
6. Include descriptive test names

## Performance Testing

### Run benchmarks:
```bash
go test -bench=. ./tests/integration/...
```

### Profile tests:
```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./tests/integration/...
```

## Debugging Failed Tests

### Keep temporary files:
```bash
go test -v ./tests/integration/... -integration.keep
```

### Run single test with verbose output:
```bash
go test -v -run TestRunCommand/basic_script_execution ./tests/integration/commands
```

### Enable debug logging in llmspell:
```bash
LLMSPELL_DEBUG=1 go test -v ./tests/integration/...
```

## CI/CD Integration

The integration tests are designed to run in CI environments:
- Automatically builds llmspell binary
- Creates isolated temporary directories
- Cleans up after tests
- Provides clear error messages
- Supports parallel execution

## Test Coverage

Current coverage includes:
- ✅ All CLI commands
- ✅ Configuration management
- ✅ Security profiles
- ✅ Engine selection
- ✅ REPL functionality
- ✅ Debug features
- ✅ Template generation
- ✅ Error handling
- ✅ Signal handling
- ✅ Cross-platform support

## Future Additions

Planned test additions:
- [ ] Watch mode testing
- [ ] Plugin system tests (when implemented)
- [ ] Multi-engine tests (when JS/Tengo added)
- [ ] Performance regression tests
- [ ] Load testing for concurrent execution