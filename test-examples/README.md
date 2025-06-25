# LLMSpell Example Testing Suite

This directory contains scripts for comprehensive testing of all example spells with various security level and feature set combinations.

## Prerequisites

1. **Build llmspell**: 
   ```bash
   go build -o llmspell ./cmd/llmspell
   ```

2. **Set API Keys**:
   ```bash
   export OPENAI_API_KEY="your-openai-api-key"
   # Optional for other providers:
   export ANTHROPIC_API_KEY="your-anthropic-api-key"
   export GOOGLE_API_KEY="your-google-api-key"
   ```

## Usage

### 1. Check Environment

First, verify your environment is ready:

```bash
./test-examples/check_env.sh
```

This will check for:
- llmspell binary
- Required API keys
- Example files
- Directory permissions
- Network connectivity

### 2. Run Tests

Execute all tests:

```bash
./test-examples/run_tests.sh
```

This will:
- Test each example (01-13) with different security/feature combinations
- Create detailed logs for each test
- Generate a JSON results file
- Show progress in real-time
- Handle timeouts (60s per test)

Test combinations:
- Security levels: `untrusted`, `trusted`, `privileged`
- Feature sets: `minimal`, `llm`, `agent`, `full`
- Some combinations are skipped (e.g., untrusted+agent)

### 3. Analyze Results

Generate reports from test results:

```bash
./test-examples/analyze_results.sh
```

Or analyze a specific results file:

```bash
./test-examples/analyze_results.sh test-output/results/results-TIMESTAMP.json
```

This generates:
- Markdown report with summary and detailed results
- CSV file for spreadsheet analysis
- Compatibility matrix showing which examples work with which configurations
- Failed test details with error messages

## Output Structure

```
test-output/
├── examples/       # Output from examples that create files
├── logs/          # Detailed logs for each test run
├── results/       # JSON results files
└── reports/       # Generated analysis reports
```

## Troubleshooting

### Common Issues

1. **"Environment check failed"**
   - Make sure API keys are set
   - Verify llmspell is built
   - Check file permissions

2. **"Timeout" errors**
   - Some examples may take longer with complex tasks
   - Check if the example needs more time
   - Network issues can cause timeouts

3. **"Permission denied" errors**
   - Some security levels restrict file/network access
   - This is expected behavior for certain combinations

### Viewing Logs

To see why a specific test failed:

```bash
# Find the log file from the results
grep -l "error" test-output/logs/*.log

# View a specific test log
cat test-output/logs/01_untrusted_minimal-*.log
```

## Example Parameters

Each example has default parameters that can be customized in `run_tests.sh`:

- `01-tools-usage.lua`: File operations output directory
- `02-basic-llm.lua`: Model and prompt
- `03-agent-plain.lua`: Task description
- `04-agent-with-tools.lua`: Calculation task
- `05-agent-as-tool.lua`: Research task
- `06-complex-workflows.lua`: Workflow type and input
- `07-event-driven.lua`: Number of events
- `08-performance-patterns.lua`: Iterations
- `09-state-management.lua`: Operations count
- `10-hooks.lua`: Verbosity
- `11-debug-usage.lua`: Debug level
- `12-custom-tool.lua`: Tool name
- `13-agent-handoff.lua`: Initial task

## Expected Results

Not all combinations will pass:
- **untrusted + agent/full**: Expected to fail (agents need more permissions)
- **File operations**: May fail with untrusted security
- **Network operations**: May fail with restricted security
- **LLM operations**: Require valid API keys

The goal is to document which examples work with which security/feature combinations.