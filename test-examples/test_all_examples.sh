#!/bin/bash
# ABOUTME: Test all example scripts and report status
# ABOUTME: Run each example and capture success/failure

echo "=== Testing All Example Scripts ==="
echo "Date: $(date)"
echo

# Array to store results
declare -A results

# Test each example
for example in examples/spells/lua/*.lua; do
    echo "Testing: $example"
    
    # Run the example with timeout and capture exit code
    timeout 10s ./llmspell run "$example" > /tmp/test_output.txt 2>&1
    exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        echo "  ✅ SUCCESS"
        results["$example"]="SUCCESS"
    else
        echo "  ❌ FAILED (exit code: $exit_code)"
        # Show first few lines of error
        echo "  Error output:"
        head -5 /tmp/test_output.txt | sed 's/^/    /'
        results["$example"]="FAILED"
    fi
    echo
done

# Summary
echo "=== SUMMARY ==="
success_count=0
fail_count=0

for example in "${!results[@]}"; do
    if [ "${results[$example]}" = "SUCCESS" ]; then
        ((success_count++))
    else
        ((fail_count++))
    fi
done

echo "Total: $((success_count + fail_count)) examples"
echo "Success: $success_count"
echo "Failed: $fail_count"

# List failed examples
if [ $fail_count -gt 0 ]; then
    echo
    echo "Failed examples:"
    for example in "${!results[@]}"; do
        if [ "${results[$example]}" = "FAILED" ]; then
            echo "  - $example"
        fi
    done | sort
fi