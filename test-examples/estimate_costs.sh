#!/bin/bash
# estimate_costs.sh - Estimate API costs for running example tests

# OpenAI pricing (as of 2024)
# GPT-3.5-turbo: $0.0005 per 1K input tokens, $0.0015 per 1K output tokens
# GPT-4: $0.03 per 1K input tokens, $0.06 per 1K output tokens

echo "=== LLMSpell Test Suite Cost Estimation ==="
echo

# Estimate tokens per example
declare -A EXAMPLE_ESTIMATES
EXAMPLE_ESTIMATES["01-tools-usage.lua"]="0|0|No API calls"
EXAMPLE_ESTIMATES["02-basic-llm.lua"]="100|200|Single completion"
EXAMPLE_ESTIMATES["03-agent-plain.lua"]="200|300|Simple agent task"
EXAMPLE_ESTIMATES["04-agent-with-tools.lua"]="500|1000|Agent with tool calls"
EXAMPLE_ESTIMATES["05-agent-as-tool.lua"]="1000|2000|Complex agent interactions"
EXAMPLE_ESTIMATES["06-complex-workflows.lua"]="800|1500|Multi-step workflow"
EXAMPLE_ESTIMATES["07-event-driven.lua"]="300|500|Event processing"
EXAMPLE_ESTIMATES["08-performance-patterns.lua"]="500|800|Performance tests"
EXAMPLE_ESTIMATES["09-state-management.lua"]="200|300|State operations"
EXAMPLE_ESTIMATES["10-hooks.lua"]="150|250|Hook demonstrations"
EXAMPLE_ESTIMATES["11-debug-usage.lua"]="0|0|No API calls"
EXAMPLE_ESTIMATES["12-custom-tool.lua"]="0|0|No API calls"
EXAMPLE_ESTIMATES["13-agent-handoff.lua"]="1500|3000|Multiple agents"

# Calculate estimates
echo "Per-Example Token Estimates (GPT-3.5-turbo):"
echo "============================================"
echo "Example                    | Input | Output | Notes"
echo "---------------------------|-------|--------|------------------"

TOTAL_INPUT=0
TOTAL_OUTPUT=0

for example in examples/spells/lua/*.lua; do
    BASENAME=$(basename "$example")
    if [[ -n "${EXAMPLE_ESTIMATES[$BASENAME]}" ]]; then
        IFS='|' read -r INPUT OUTPUT NOTES <<< "${EXAMPLE_ESTIMATES[$BASENAME]}"
        printf "%-26s | %5s | %6s | %s\n" "${BASENAME%%.lua}" "$INPUT" "$OUTPUT" "$NOTES"
        TOTAL_INPUT=$((TOTAL_INPUT + INPUT))
        TOTAL_OUTPUT=$((TOTAL_OUTPUT + OUTPUT))
    fi
done

echo
echo "Totals per run:            | $TOTAL_INPUT | $TOTAL_OUTPUT |"
echo

# Test configurations
SECURITY_LEVELS=3
FEATURE_SETS=4
VALID_COMBINATIONS=10  # Some combinations are skipped

echo "Test Configuration Multipliers:"
echo "=============================="
echo "Security levels: $SECURITY_LEVELS (untrusted, trusted, privileged)"
echo "Feature sets: $FEATURE_SETS (minimal, llm, agent, full)"
echo "Valid combinations: ~$VALID_COMBINATIONS per example"
echo

# Calculate total tokens
TOTAL_RUNS=$VALID_COMBINATIONS
FULL_INPUT=$((TOTAL_INPUT * TOTAL_RUNS))
FULL_OUTPUT=$((TOTAL_OUTPUT * TOTAL_RUNS))

echo "Full Test Suite Estimates:"
echo "========================="
echo "Total test runs: $TOTAL_RUNS per example"
echo "Total input tokens: $(printf "%'d" $FULL_INPUT)"
echo "Total output tokens: $(printf "%'d" $FULL_OUTPUT)"
echo

# Cost calculations
GPT35_INPUT_COST=$(echo "scale=4; $FULL_INPUT * 0.0005 / 1000" | bc)
GPT35_OUTPUT_COST=$(echo "scale=4; $FULL_OUTPUT * 0.0015 / 1000" | bc)
GPT35_TOTAL=$(echo "scale=4; $GPT35_INPUT_COST + $GPT35_OUTPUT_COST" | bc)

GPT4_INPUT_COST=$(echo "scale=4; $FULL_INPUT * 0.03 / 1000" | bc)
GPT4_OUTPUT_COST=$(echo "scale=4; $FULL_OUTPUT * 0.06 / 1000" | bc)
GPT4_TOTAL=$(echo "scale=4; $GPT4_INPUT_COST + $GPT4_OUTPUT_COST" | bc)

echo "Estimated Costs:"
echo "==============="
echo "GPT-3.5-turbo:"
echo "  Input:  \$$GPT35_INPUT_COST"
echo "  Output: \$$GPT35_OUTPUT_COST"
echo "  Total:  \$$GPT35_TOTAL"
echo
echo "GPT-4 (if used):"
echo "  Input:  \$$GPT4_INPUT_COST"
echo "  Output: \$$GPT4_OUTPUT_COST"
echo "  Total:  \$$GPT4_TOTAL"
echo

echo "Subset Test Estimates (run_subset_tests.sh):"
echo "==========================================="
echo "Phase 1 (non-API): \$0.00"
echo "Phase 2 (basic):   ~\$0.01"
echo "Phase 3 (complex): ~\$0.05"
echo "Subset total:      ~\$0.06"
echo

echo "Notes:"
echo "======"
echo "- Estimates assume GPT-3.5-turbo for all tests"
echo "- Actual costs may vary based on:"
echo "  - Prompt complexity"
echo "  - Response length"
echo "  - Number of retries"
echo "  - Model selection"
echo "- Some examples may fail quickly, reducing costs"
echo "- Rate limiting helps prevent runaway costs"