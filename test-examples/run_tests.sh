#!/bin/bash
# run_tests.sh - Run all example spells with various security/feature combinations

set -e

# Configuration
EXAMPLES_DIR="examples/spells/lua"
OUTPUT_DIR="test-output/examples"
RESULTS_DIR="test-output/results"
LOG_DIR="test-output/logs"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
MAIN_LOG="$LOG_DIR/test-run-$TIMESTAMP.log"
RESULTS_FILE="$RESULTS_DIR/results-$TIMESTAMP.json"

# Test configurations
SECURITY_LEVELS=("untrusted" "trusted" "privileged")
FEATURE_SETS=("minimal" "llm" "agent" "full")

# Special parameters for specific examples
declare -A EXAMPLE_PARAMS
EXAMPLE_PARAMS["01-tools-usage.lua"]="output_dir=$OUTPUT_DIR/01-tools"
EXAMPLE_PARAMS["02-basic-llm.lua"]="model=gpt-3.5-turbo prompt='Tell me about Lua'"
EXAMPLE_PARAMS["03-agent-plain.lua"]="task='Write a haiku about programming'"
EXAMPLE_PARAMS["04-agent-with-tools.lua"]="task='Calculate the sum of 1 to 100'"
EXAMPLE_PARAMS["05-agent-as-tool.lua"]="task='Research and summarize Lua history'"
EXAMPLE_PARAMS["06-complex-workflows.lua"]="workflow=analyze input='Sample text for analysis'"
EXAMPLE_PARAMS["07-event-driven.lua"]="events=10"
EXAMPLE_PARAMS["08-performance-patterns.lua"]="iterations=5"
EXAMPLE_PARAMS["09-state-management.lua"]="operations=10"
EXAMPLE_PARAMS["10-hooks.lua"]="verbose=true"
EXAMPLE_PARAMS["11-debug-usage.lua"]="debug_level=2"
EXAMPLE_PARAMS["12-custom-tool.lua"]="tool_name=custom_calculator"
EXAMPLE_PARAMS["13-agent-handoff.lua"]="initial_task='Plan a simple program'"

# Initialize results
echo "{" > "$RESULTS_FILE"
echo "  \"timestamp\": \"$TIMESTAMP\"," >> "$RESULTS_FILE"
echo "  \"test_results\": [" >> "$RESULTS_FILE"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Helper functions
log() {
    echo "$1" | tee -a "$MAIN_LOG"
}

log_error() {
    echo -e "${RED}$1${NC}" | tee -a "$MAIN_LOG"
}

log_success() {
    echo -e "${GREEN}$1${NC}" | tee -a "$MAIN_LOG"
}

log_info() {
    echo -e "${BLUE}$1${NC}" | tee -a "$MAIN_LOG"
}

# Start testing
log "=== LLMSpell Example Testing Suite ==="
log "Started at: $(date)"
log "Main log: $MAIN_LOG"
log "Results: $RESULTS_FILE"
log ""

# Check environment first
if ! ./test-examples/check_env.sh > "$LOG_DIR/env-check-$TIMESTAMP.log" 2>&1; then
    log_error "Environment check failed! See $LOG_DIR/env-check-$TIMESTAMP.log"
    exit 1
fi
log_success "Environment check passed"
log ""

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Run tests
FIRST_RESULT=true
for example in $EXAMPLES_DIR/*.lua; do
    EXAMPLE_NAME=$(basename "$example")
    EXAMPLE_NUM="${EXAMPLE_NAME%%-*}"
    
    log "=== Testing: $EXAMPLE_NAME ==="
    
    # Get example-specific parameters
    PARAMS="${EXAMPLE_PARAMS[$EXAMPLE_NAME]:-}"
    
    for security in "${SECURITY_LEVELS[@]}"; do
        for features in "${FEATURE_SETS[@]}"; do
            TOTAL_TESTS=$((TOTAL_TESTS + 1))
            TEST_NAME="${EXAMPLE_NUM}_${security}_${features}"
            TEST_LOG="$LOG_DIR/$TEST_NAME-$TIMESTAMP.log"
            
            # Skip certain combinations that are known to be incompatible
            # (e.g., untrusted security with agent features)
            if [[ "$security" == "untrusted" && "$features" == "agent" ]]; then
                log "  Skipping $security/$features (incompatible combination)"
                SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
                continue
            fi
            
            log -n "  Testing $security/$features... "
            
            # Build command
            CMD="./llmspell run \"$example\" \
                --security-level \"$security\" \
                --feature-set \"$features\""
            
            # Add parameters if any
            if [ ! -z "$PARAMS" ]; then
                CMD="$CMD --parameters \"$PARAMS\""
            fi
            
            # Add timeout to prevent hanging
            CMD="timeout 60s $CMD"
            
            # Run the test
            START_TIME=$(date +%s)
            if eval "$CMD" > "$TEST_LOG" 2>&1; then
                END_TIME=$(date +%s)
                DURATION=$((END_TIME - START_TIME))
                log_success "PASSED (${DURATION}s)"
                PASSED_TESTS=$((PASSED_TESTS + 1))
                RESULT="passed"
                ERROR=""
            else
                EXIT_CODE=$?
                END_TIME=$(date +%s)
                DURATION=$((END_TIME - START_TIME))
                
                if [ $EXIT_CODE -eq 124 ]; then
                    log_error "TIMEOUT (60s)"
                    RESULT="timeout"
                    ERROR="Test exceeded 60 second timeout"
                else
                    # Extract error message
                    ERROR=$(tail -n 5 "$TEST_LOG" | tr '\n' ' ' | sed 's/"/\\"/g')
                    log_error "FAILED"
                    RESULT="failed"
                fi
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
            
            # Write result to JSON
            if [ "$FIRST_RESULT" = true ]; then
                FIRST_RESULT=false
            else
                echo "," >> "$RESULTS_FILE"
            fi
            
            cat >> "$RESULTS_FILE" <<EOF
    {
      "example": "$EXAMPLE_NAME",
      "security_level": "$security",
      "feature_set": "$features",
      "result": "$RESULT",
      "duration": $DURATION,
      "log_file": "$TEST_LOG",
      "error": "$ERROR"
    }
EOF
        done
    done
    
    log ""
done

# Close JSON
echo "" >> "$RESULTS_FILE"
echo "  ]," >> "$RESULTS_FILE"
echo "  \"summary\": {" >> "$RESULTS_FILE"
echo "    \"total\": $TOTAL_TESTS," >> "$RESULTS_FILE"
echo "    \"passed\": $PASSED_TESTS," >> "$RESULTS_FILE"
echo "    \"failed\": $FAILED_TESTS," >> "$RESULTS_FILE"
echo "    \"skipped\": $SKIPPED_TESTS" >> "$RESULTS_FILE"
echo "  }" >> "$RESULTS_FILE"
echo "}" >> "$RESULTS_FILE"

# Summary
log "=== Test Summary ==="
log "Total tests: $TOTAL_TESTS"
log_success "Passed: $PASSED_TESTS"
log_error "Failed: $FAILED_TESTS"
log "Skipped: $SKIPPED_TESTS"
log ""
log "Results saved to: $RESULTS_FILE"
log "Logs saved to: $LOG_DIR"
log ""

# Exit with appropriate code
if [ $FAILED_TESTS -gt 0 ]; then
    log_error "Some tests failed!"
    exit 1
else
    log_success "All tests passed!"
    exit 0
fi