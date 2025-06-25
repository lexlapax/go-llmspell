#!/bin/bash
# check_env.sh - Verify test environment is ready for example spell testing

set -e

echo "=== LLMSpell Example Testing Environment Check ==="
echo "Timestamp: $(date)"
echo

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check function
check_requirement() {
    local name="$1"
    local check_command="$2"
    local required="$3"
    
    echo -n "Checking $name... "
    if eval "$check_command"; then
        echo -e "${GREEN}✓ OK${NC}"
        return 0
    else
        if [ "$required" = "true" ]; then
            echo -e "${RED}✗ FAILED (REQUIRED)${NC}"
            return 1
        else
            echo -e "${YELLOW}⚠ WARNING (Optional)${NC}"
            return 0
        fi
    fi
}

# Track failures
FAILED=0

# Check for llmspell binary
if ! check_requirement "llmspell binary" "[ -f ./llmspell ]" "true"; then
    echo "  Please build llmspell first: go build -o llmspell ./cmd/llmspell"
    FAILED=1
fi

# Check for API keys
echo
echo "=== API Keys ==="
if ! check_requirement "OPENAI_API_KEY" "[ ! -z \"\$OPENAI_API_KEY\" ]" "true"; then
    echo "  Required for examples using OpenAI models (gpt-3.5-turbo, gpt-4)"
    echo "  Set with: export OPENAI_API_KEY='your-key-here'"
    FAILED=1
fi

# Optional API keys for other providers
check_requirement "ANTHROPIC_API_KEY" "[ ! -z \"\$ANTHROPIC_API_KEY\" ]" "false"
check_requirement "GOOGLE_API_KEY" "[ ! -z \"\$GOOGLE_API_KEY\" ]" "false"

# Check for example files
echo
echo "=== Example Files ==="
if ! check_requirement "examples directory" "[ -d examples/spells/lua ]" "true"; then
    echo "  Examples directory not found!"
    FAILED=1
else
    EXAMPLE_COUNT=$(ls examples/spells/lua/*.lua 2>/dev/null | wc -l | tr -d ' ')
    echo "  Found $EXAMPLE_COUNT example spell(s)"
fi

# Create test directories
echo
echo "=== Test Directories ==="
for dir in test-output test-output/examples test-output/logs test-output/results; do
    if [ ! -d "$dir" ]; then
        echo "Creating directory: $dir"
        mkdir -p "$dir"
    else
        echo "Directory exists: $dir"
    fi
done

# Check file system permissions
echo
echo "=== File System Permissions ==="
TEST_FILE="test-output/.permission-test"
if touch "$TEST_FILE" 2>/dev/null; then
    rm -f "$TEST_FILE"
    echo -e "${GREEN}✓ Write permissions OK${NC}"
else
    echo -e "${RED}✗ Cannot write to test-output directory${NC}"
    FAILED=1
fi

# Check network connectivity (for examples that fetch URLs)
echo
echo "=== Network Connectivity ==="
if check_requirement "GitHub API" "curl -s -o /dev/null -w '%{http_code}' https://api.github.com | grep -q '200'" "false"; then
    :
else
    echo "  Some examples may fail if they require network access"
fi

# Summary
echo
echo "=== Summary ==="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All required checks passed! Environment is ready for testing.${NC}"
    echo
    echo "Next steps:"
    echo "1. Run the test suite: ./test-examples/run_tests.sh"
    echo "2. View results: ./test-examples/analyze_results.sh"
    exit 0
else
    echo -e "${RED}Some required checks failed. Please fix the issues above.${NC}"
    exit 1
fi