#!/bin/bash
# analyze_results.sh - Analyze test results and generate reports

set -e

RESULTS_DIR="test-output/results"
REPORTS_DIR="test-output/reports"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Create reports directory
mkdir -p "$REPORTS_DIR"

# Find latest results file if not specified
if [ -z "$1" ]; then
    RESULTS_FILE=$(ls -t "$RESULTS_DIR"/results-*.json 2>/dev/null | head -1)
    if [ -z "$RESULTS_FILE" ]; then
        echo -e "${RED}No results files found in $RESULTS_DIR${NC}"
        echo "Run ./test-examples/run_tests.sh first"
        exit 1
    fi
else
    RESULTS_FILE="$1"
fi

echo -e "${BLUE}=== LLMSpell Test Results Analysis ===${NC}"
echo "Analyzing: $RESULTS_FILE"
echo

# Generate markdown report
REPORT_FILE="$REPORTS_DIR/test-report-$TIMESTAMP.md"

cat > "$REPORT_FILE" <<'EOF'
# LLMSpell Example Tests Report

EOF

# Add timestamp
echo "**Generated:** $(date)" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Extract summary using jq (or python if jq not available)
if command -v jq &> /dev/null; then
    # Use jq
    TOTAL=$(jq -r '.summary.total' "$RESULTS_FILE")
    PASSED=$(jq -r '.summary.passed' "$RESULTS_FILE")
    FAILED=$(jq -r '.summary.failed' "$RESULTS_FILE")
    SKIPPED=$(jq -r '.summary.skipped' "$RESULTS_FILE")
    
    # Summary section
    cat >> "$REPORT_FILE" <<EOF
## Summary

- **Total Tests:** $TOTAL
- **Passed:** $PASSED ($(( PASSED * 100 / TOTAL ))%)
- **Failed:** $FAILED ($(( FAILED * 100 / TOTAL ))%)
- **Skipped:** $SKIPPED

EOF
    
    # Detailed results by example
    echo "## Results by Example" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # Get unique examples
    EXAMPLES=$(jq -r '.test_results[].example' "$RESULTS_FILE" | sort -u)
    
    for example in $EXAMPLES; do
        echo "### $example" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "| Security Level | Feature Set | Result | Duration | Error |" >> "$REPORT_FILE"
        echo "|----------------|-------------|--------|----------|-------|" >> "$REPORT_FILE"
        
        # Get results for this example
        jq -r --arg ex "$example" '.test_results[] | select(.example == $ex) | 
            "| \(.security_level) | \(.feature_set) | \(.result) | \(.duration)s | \(.error | if . == "" then "-" else . end) |"' \
            "$RESULTS_FILE" >> "$REPORT_FILE"
        
        echo "" >> "$REPORT_FILE"
    done
    
    # Failed tests section
    FAILED_COUNT=$(jq -r '.test_results[] | select(.result != "passed" and .result != "skipped") | .example' "$RESULTS_FILE" | wc -l)
    if [ "$FAILED_COUNT" -gt 0 ]; then
        echo "## Failed Tests Details" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        
        jq -r '.test_results[] | select(.result != "passed" and .result != "skipped") | 
            "### \(.example) - \(.security_level)/\(.feature_set)\n\n**Error:** \(.error)\n\n**Log:** \(.log_file)\n"' \
            "$RESULTS_FILE" >> "$REPORT_FILE"
    fi
    
    # Compatibility matrix
    echo "## Compatibility Matrix" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "| Example | untrusted/minimal | untrusted/llm | trusted/minimal | trusted/llm | trusted/agent | privileged/full |" >> "$REPORT_FILE"
    echo "|---------|-------------------|---------------|-----------------|-------------|---------------|-----------------|" >> "$REPORT_FILE"
    
    for example in $EXAMPLES; do
        echo -n "| ${example%%.lua} " >> "$REPORT_FILE"
        for security in untrusted trusted privileged; do
            for features in minimal llm agent full; do
                # Skip invalid combinations
                if [[ "$security" == "untrusted" && ("$features" == "agent" || "$features" == "full") ]]; then
                    continue
                fi
                if [[ "$security" == "privileged" && "$features" != "full" ]]; then
                    continue
                fi
                
                RESULT=$(jq -r --arg ex "$example" --arg sec "$security" --arg feat "$features" \
                    '.test_results[] | select(.example == $ex and .security_level == $sec and .feature_set == $feat) | .result' \
                    "$RESULTS_FILE" | head -1)
                
                case "$RESULT" in
                    "passed") echo -n "| ✅ " >> "$REPORT_FILE" ;;
                    "failed") echo -n "| ❌ " >> "$REPORT_FILE" ;;
                    "timeout") echo -n "| ⏱️ " >> "$REPORT_FILE" ;;
                    *) echo -n "| - " >> "$REPORT_FILE" ;;
                esac
            done
        done
        echo "|" >> "$REPORT_FILE"
    done
    
else
    # Fallback to Python
    python3 -c "
import json
import sys

with open('$RESULTS_FILE', 'r') as f:
    data = json.load(f)

summary = data['summary']
print(f'Total: {summary[\"total\"]}')
print(f'Passed: {summary[\"passed\"]}')
print(f'Failed: {summary[\"failed\"]}')
print(f'Skipped: {summary[\"skipped\"]}')
"
fi

# Print summary to console
echo -e "${GREEN}Report generated: $REPORT_FILE${NC}"
echo
echo "Summary:"
grep -A 4 "## Summary" "$REPORT_FILE" | tail -n +2

# Create CSV for spreadsheet analysis
CSV_FILE="$REPORTS_DIR/test-results-$TIMESTAMP.csv"
echo "example,security_level,feature_set,result,duration,error" > "$CSV_FILE"

if command -v jq &> /dev/null; then
    jq -r '.test_results[] | [.example, .security_level, .feature_set, .result, .duration, .error] | @csv' \
        "$RESULTS_FILE" >> "$CSV_FILE"
fi

echo
echo -e "${GREEN}CSV generated: $CSV_FILE${NC}"

# Show failed tests if any
if [ "$FAILED" -gt 0 ]; then
    echo
    echo -e "${RED}Failed tests:${NC}"
    if command -v jq &> /dev/null; then
        jq -r '.test_results[] | select(.result == "failed") | 
            "  - \(.example) [\(.security_level)/\(.feature_set)]: \(.error)"' "$RESULTS_FILE"
    fi
fi

echo
echo "To view detailed logs for a specific test, check the test-output/logs/ directory"