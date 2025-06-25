#!/bin/bash
# run_subset_tests.sh - Run a subset of example tests with rate limiting

set -e

# Configuration
EXAMPLES_DIR="examples/spells/lua"
OUTPUT_DIR="test-output/examples"
RESULTS_DIR="test-output/results"
LOG_DIR="test-output/logs"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
MAIN_LOG="$LOG_DIR/subset-test-run-$TIMESTAMP.log"
RESULTS_FILE="$RESULTS_DIR/subset-results-$TIMESTAMP.json"

# Rate limiting configuration
RATE_LIMIT_DELAY=2  # seconds between API tests
API_TEST_TIMEOUT=30 # shorter timeout for subset tests

# Test phases
declare -a PHASE1_TESTS=(
    "test-examples/minimal-test.lua"
    "test-examples/minimal-working-test.lua"
    "01-tools-usage.lua"  # This will only run with 'full' features
)

declare -a PHASE2_TESTS=(
    "02-basic-llm.lua"
    "03-agent-plain.lua"
)

declare -a PHASE3_TESTS=(
    "04-agent-with-tools.lua"
    "06-complex-workflows.lua"
)

# Limited test configurations for subset
SECURITY_LEVELS=("trusted")  # Start with just trusted
FEATURE_SETS=("minimal" "full")  # Test extremes

# Function to get parameters for an example
get_example_params() {
    case "$1" in
        "minimal-test.lua") echo "" ;;
        "minimal-working-test.lua") echo "" ;;
        "01-tools-usage.lua") echo "output_dir=$OUTPUT_DIR/01-tools" ;;
        "02-basic-llm.lua") echo "model=gpt-3.5-turbo prompt='Say hello in 5 words'" ;;
        "03-agent-plain.lua") echo "task='Write a 3 line haiku'" ;;
        "04-agent-with-tools.lua") echo "task='What is 25 + 37?'" ;;
        "06-complex-workflows.lua") echo "workflow=analyze input='Test'" ;;
        "09-state-management.lua") echo "operations=5" ;;
        "10-hooks.lua") echo "verbose=true" ;;
        "11-debug-usage.lua") echo "debug_level=1" ;;
        "12-custom-tool.lua") echo "tool_name=test_calc" ;;
        *) echo "" ;;
    esac
}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Helper functions
log() {
    echo "$1" | tee -a "$MAIN_LOG"
}

log_phase() {
    echo -e "${BLUE}=== $1 ===${NC}" | tee -a "$MAIN_LOG"
}

log_error() {
    echo -e "${RED}$1${NC}" | tee -a "$MAIN_LOG"
}

log_success() {
    echo -e "${GREEN}$1${NC}" | tee -a "$MAIN_LOG"
}

log_warning() {
    echo -e "${YELLOW}$1${NC}" | tee -a "$MAIN_LOG"
}

# Initialize results JSON
init_results() {
    cat > "$RESULTS_FILE" <<EOF
{
  "timestamp": "$TIMESTAMP",
  "test_type": "subset",
  "rate_limit_delay": $RATE_LIMIT_DELAY,
  "test_results": [
EOF
}

# Run a single test
run_test() {
    local example="$1"
    local security="$2"
    local features="$3"
    local timeout="${4:-60}"
    
    local EXAMPLE_NAME=$(basename "$example")
    local TEST_NAME="${EXAMPLE_NAME%%.lua}_${security}_${features}"
    local TEST_LOG="$LOG_DIR/$TEST_NAME-$TIMESTAMP.log"
    
    log -n "  Testing $EXAMPLE_NAME [$security/$features]... "
    
    # Build command
    local CMD="./llmspell run \"$example\" \
        --security-level \"$security\" \
        --feature-set \"$features\""
    
    # Add parameters if any
    local PARAMS=$(get_example_params "$EXAMPLE_NAME")
    if [ ! -z "$PARAMS" ]; then
        CMD="$CMD --parameters \"$PARAMS\""
    fi
    
    # Add timeout
    CMD="timeout ${timeout}s $CMD"
    
    # Run the test
    local START_TIME=$(date +%s)
    if eval "$CMD" > "$TEST_LOG" 2>&1; then
        local END_TIME=$(date +%s)
        local DURATION=$((END_TIME - START_TIME))
        log_success "PASSED (${DURATION}s)"
        echo "passed|$DURATION|" >> "$MAIN_LOG.results"
        return 0
    else
        local EXIT_CODE=$?
        local END_TIME=$(date +%s)
        local DURATION=$((END_TIME - START_TIME))
        
        if [ $EXIT_CODE -eq 124 ]; then
            log_error "TIMEOUT (${timeout}s)"
            echo "timeout|$DURATION|Exceeded ${timeout}s timeout" >> "$MAIN_LOG.results"
        else
            local ERROR=$(tail -n 3 "$TEST_LOG" | tr '\n' ' ')
            log_error "FAILED"
            echo "failed|$DURATION|$ERROR" >> "$MAIN_LOG.results"
        fi
        return 1
    fi
}

# Start testing
log "=== LLMSpell Subset Testing with Rate Limiting ==="
log "Started at: $(date)"
log "Rate limit: ${RATE_LIMIT_DELAY}s between API tests"
log "Test configurations: ${SECURITY_LEVELS[*]} × ${FEATURE_SETS[*]}"
log ""

# Check environment
if ! ./test-examples/check_env.sh > "$LOG_DIR/env-check-subset-$TIMESTAMP.log" 2>&1; then
    log_error "Environment check failed! See $LOG_DIR/env-check-subset-$TIMESTAMP.log"
    exit 1
fi
log_success "Environment check passed"
log ""

# Initialize results
init_results
FIRST_RESULT=true

# Test counters
TOTAL=0
PASSED=0
FAILED=0

# Phase 1: Non-API tests
log_phase "Phase 1: Non-API Tests (No External Dependencies)"
log "These tests should run quickly without API calls"
log ""

for test in "${PHASE1_TESTS[@]}"; do
    for security in "${SECURITY_LEVELS[@]}"; do
        for features in "${FEATURE_SETS[@]}"; do
            # Determine test path
            if [[ "$test" == test-examples/* ]]; then
                TEST_PATH="$test"
            else
                TEST_PATH="$EXAMPLES_DIR/$test"
            fi
            
            # Skip tests that require tools with minimal features
            if [[ "$features" == "minimal" && "$test" == "01-tools-usage.lua" ]]; then
                log "  Skipping $test with minimal features (requires tools)"
                continue
            fi
            
            TOTAL=$((TOTAL + 1))
            if run_test "$TEST_PATH" "$security" "$features" 10; then
                PASSED=$((PASSED + 1))
            else
                FAILED=$((FAILED + 1))
            fi
        done
    done
done

log ""
log_phase "Phase 1 Complete"
log "Passed: $PASSED/$TOTAL"
log ""

# Ask to continue to Phase 2
log_warning "Phase 2 will make API calls. Continue? (y/n)"
read -r CONTINUE
if [[ ! "$CONTINUE" =~ ^[Yy]$ ]]; then
    log "Stopping at Phase 1"
    echo "  ]}" >> "$RESULTS_FILE"
    exit 0
fi

# Phase 2: Basic API tests
log ""
log_phase "Phase 2: Basic API Tests (Minimal Cost)"
log "Rate limiting: ${RATE_LIMIT_DELAY}s between tests"
log ""

PHASE2_START=$TOTAL
for test in "${PHASE2_TESTS[@]}"; do
    for security in "${SECURITY_LEVELS[@]}"; do
        for features in "${FEATURE_SETS[@]}"; do
            # Skip minimal feature set for LLM tests
            if [[ "$features" == "minimal" && "$test" =~ "llm" ]]; then
                log "  Skipping $test with minimal features (no LLM bridge)"
                continue
            fi
            
            TOTAL=$((TOTAL + 1))
            if run_test "$EXAMPLES_DIR/$test" "$security" "$features" "$API_TEST_TIMEOUT"; then
                PASSED=$((PASSED + 1))
            else
                FAILED=$((FAILED + 1))
            fi
            
            # Rate limiting
            log "  Waiting ${RATE_LIMIT_DELAY}s (rate limit)..."
            sleep $RATE_LIMIT_DELAY
        done
    done
done

log ""
log_phase "Phase 2 Complete"
log "Phase 2 passed: $((PASSED - PHASE2_START))/$((TOTAL - PHASE2_START))"
log ""

# Ask to continue to Phase 3
if [ $FAILED -gt 0 ]; then
    log_warning "Some tests failed. Continue to Phase 3? (y/n)"
else
    log_warning "Phase 3 will test complex examples. Continue? (y/n)"
fi
read -r CONTINUE
if [[ ! "$CONTINUE" =~ ^[Yy]$ ]]; then
    log "Stopping at Phase 2"
    echo "  ]}" >> "$RESULTS_FILE"
    exit 0
fi

# Phase 3: Complex API tests
log ""
log_phase "Phase 3: Complex API Tests (Higher Cost)"
log "These may make multiple API calls"
log ""

PHASE3_START=$TOTAL
for test in "${PHASE3_TESTS[@]}"; do
    # Only test with full features for complex examples
    for security in "${SECURITY_LEVELS[@]}"; do
        TOTAL=$((TOTAL + 1))
        if run_test "$EXAMPLES_DIR/$test" "$security" "full" "$API_TEST_TIMEOUT"; then
            PASSED=$((PASSED + 1))
        else
            FAILED=$((FAILED + 1))
        fi
        
        # Rate limiting
        log "  Waiting ${RATE_LIMIT_DELAY}s (rate limit)..."
        sleep $RATE_LIMIT_DELAY
    done
done

# Close JSON
echo "  ]," >> "$RESULTS_FILE"
echo "  \"summary\": {" >> "$RESULTS_FILE"
echo "    \"total\": $TOTAL," >> "$RESULTS_FILE"
echo "    \"passed\": $PASSED," >> "$RESULTS_FILE"
echo "    \"failed\": $FAILED," >> "$RESULTS_FILE"
echo "    \"phases\": {" >> "$RESULTS_FILE"
echo "      \"phase1\": ${#PHASE1_TESTS[@]}," >> "$RESULTS_FILE"
echo "      \"phase2\": ${#PHASE2_TESTS[@]}," >> "$RESULTS_FILE"
echo "      \"phase3\": ${#PHASE3_TESTS[@]}" >> "$RESULTS_FILE"
echo "    }" >> "$RESULTS_FILE"
echo "  }" >> "$RESULTS_FILE"
echo "}" >> "$RESULTS_FILE"

# Final summary
log ""
log_phase "Final Summary"
log "Total tests: $TOTAL"
log_success "Passed: $PASSED"
if [ $FAILED -gt 0 ]; then
    log_error "Failed: $FAILED"
else
    log "Failed: 0"
fi
log ""
log "Results saved to: $RESULTS_FILE"
log "Logs saved to: $LOG_DIR"
log ""

# Estimate for full test suite
FULL_SUITE_TESTS=$((13 * 3 * 4))  # 13 examples × 3 security × 4 features
ESTIMATED_TIME=$((FULL_SUITE_TESTS * 10 / 60))  # Assume 10s average per test
log_phase "Full Test Suite Estimate"
log "Total tests: ~$FULL_SUITE_TESTS"
log "Estimated time: ~${ESTIMATED_TIME} minutes"
log "Estimated API calls: ~50-100 (depending on example complexity)"
log ""
log "Run ./test-examples/analyze_results.sh $RESULTS_FILE to see detailed analysis"