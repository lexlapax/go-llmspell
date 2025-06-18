#!/bin/bash

# ABOUTME: Script to apply helper function conversions to bridge files
# ABOUTME: Uses the validation helpers to replace manual ScriptValue conversion patterns

set -e

apply_helper_conversions() {
    local file="$1"
    local package_name=$(dirname "$file" | sed 's|.*/||')
    
    echo "Applying helper conversions to $file..."
    
    # Backup original file
    cp "$file" "$file.helpers_backup"
    
    # Add import for helpers (add after existing imports)
    if ! grep -q "helpers" "$file"; then
        # Find the last import line and add the helper import
        sed -i '/^import (/,/^)/{
            /^)/{
                i\	"github.com/lexlapax/go-llmspell/pkg/bridge/'$package_name'/helpers"
            }
        }' "$file"
    fi
    
    # Replace common validation patterns with helper calls
    
    # String argument validation pattern
    sed -i '
    /if args\[0\] == nil || args\[0\]\.Type() != engine\.TypeString {/,/args\[0\]\.(engine\.StringValue)\.Value()/ {
        /if args\[0\] == nil/c\
        value, err := helpers.ValidateStringArg(args, 0, "argument")
        /return nil, fmt\.Errorf/c\
        if err != nil {
        /args\[0\]\.(engine\.StringValue)\.Value()/c\
        return nil, err\
        }\
        // value now contains the string value
    }' "$file"
    
    # Object argument validation
    sed -i '
    /if args\[0\] == nil || args\[0\]\.Type() != engine\.TypeObject {/,/}/ {
        /if args\[0\] == nil/c\
        objData, err := helpers.ValidateObjectArg(args, 0, "argument")
        /return nil, fmt\.Errorf/c\
        if err != nil {
        /for k, v := range args\[0\]\.(engine\.ObjectValue)\.Fields()/c\
        return nil, err\
        }\
        // objData now contains map[string]interface{}
    }' "$file"
    
    # Simple return value conversions using helpers
    sed -i 's/return engine\.NewObjectValue(map\[string\]engine\.ScriptValue{\([^}]*\)}), nil/return engine.NewObjectValue(helpers.ConvertMapToScriptValue(\1)), nil/g' "$file"
    
    echo "Applied helper conversions to $file (backup: $file.helpers_backup)"
}

# Main script
if [ $# -eq 0 ]; then
    echo "Usage: $0 <bridge_file1> [bridge_file2] ..."
    echo "   or: $0 --all (applies to all bridge files)"
    exit 1
fi

if [ "$1" = "--all" ]; then
    echo "Applying helper conversions to all bridge files..."
    
    find ./pkg/bridge -name "*.go" -not -name "*_test.go" -not -name "*helpers*" | while read -r file; do
        apply_helper_conversions "$file"
    done
    
    echo "Helper conversions applied to all bridge files!"
else
    for file in "$@"; do
        if [ -f "$file" ]; then
            apply_helper_conversions "$file"
        else
            echo "Warning: File $file not found"
        fi
    done
fi

echo ""
echo "Next steps:"
echo "1. Review the conversion results"
echo "2. Test compilation: go build ./pkg/bridge/..."
echo "3. Fix any remaining manual conversions"
echo "4. Update method names and parameters as needed"