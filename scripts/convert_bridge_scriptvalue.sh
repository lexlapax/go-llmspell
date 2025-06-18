#!/bin/bash

# ABOUTME: Batch script to automate ScriptValue conversion patterns in bridge files
# ABOUTME: Converts common patterns from []interface{} to []engine.ScriptValue usage

set -e

# Function to convert a single file
convert_file() {
    local file="$1"
    echo "Converting $file..."
    
    # Backup original file
    cp "$file" "$file.backup"
    
    # Common argument type assertion patterns
    sed -i 's/args\[0\]\.\(string\)/args[0].(engine.StringValue).Value()/g' "$file"
    sed -i 's/args\[1\]\.\(string\)/args[1].(engine.StringValue).Value()/g' "$file"
    sed -i 's/args\[2\]\.\(string\)/args[2].(engine.StringValue).Value()/g' "$file"
    sed -i 's/args\[3\]\.\(string\)/args[3].(engine.StringValue).Value()/g' "$file"
    
    # Number conversions
    sed -i 's/args\[0\]\.\(float64\)/args[0].(engine.NumberValue).Value()/g' "$file"
    sed -i 's/args\[1\]\.\(float64\)/args[1].(engine.NumberValue).Value()/g' "$file"
    sed -i 's/args\[2\]\.\(float64\)/args[2].(engine.NumberValue).Value()/g' "$file"
    
    # Boolean conversions
    sed -i 's/args\[0\]\.\(bool\)/args[0].(engine.BoolValue).Value()/g' "$file"
    sed -i 's/args\[1\]\.\(bool\)/args[1].(engine.BoolValue).Value()/g' "$file"
    sed -i 's/args\[2\]\.\(bool\)/args[2].(engine.BoolValue).Value()/g' "$file"
    
    # Add string validation patterns before type assertions
    # This is more complex and needs manual review, but we can add placeholders
    
    # Simple return value conversions
    sed -i 's/return nil, nil$/return engine.NewNilValue(), nil/g' "$file"
    sed -i 's/return \([a-zA-Z_][a-zA-Z0-9_]*\), nil$/return engine.NewStringValue(\1), nil/g' "$file"
    
    # Common object return patterns (these need manual review)
    # sed -i 's/return \([a-zA-Z_][a-zA-Z0-9_]*\)\[\(.*\)\], nil$/return convertToScriptValue(\1[\2]), nil/g' "$file"
    
    echo "Converted $file (backup saved as $file.backup)"
}

# Function to add validation patterns
add_validations() {
    local file="$1"
    echo "Adding validation patterns to $file..."
    
    # This is complex and should be done with a more sophisticated tool
    # For now, we'll create a template
    cat > "$file.validation_template" << 'EOF'
// Common validation patterns to replace manual type assertions:

// String validation:
if args[0] == nil || args[0].Type() != engine.TypeString {
    return nil, fmt.Errorf("argument must be string")
}
value := args[0].(engine.StringValue).Value()

// Number validation:
if args[0] == nil || args[0].Type() != engine.TypeNumber {
    return nil, fmt.Errorf("argument must be number")
}
value := args[0].(engine.NumberValue).Value()

// Object validation:
if args[0] == nil || args[0].Type() != engine.TypeObject {
    return nil, fmt.Errorf("argument must be object")
}
objData := make(map[string]interface{})
for k, v := range args[0].(engine.ObjectValue).Fields() {
    objData[k] = v.ToGo()
}

// Array validation:
if args[0] == nil || args[0].Type() != engine.TypeArray {
    return nil, fmt.Errorf("argument must be array")
}
arrayData := make([]interface{}, 0)
for _, v := range args[0].(engine.ArrayValue).Values() {
    arrayData = append(arrayData, v.ToGo())
}

// Return value conversions:
// String: return engine.NewStringValue(result), nil
// Number: return engine.NewNumberValue(float64(result)), nil  
// Boolean: return engine.NewBoolValue(result), nil
// Object: return engine.NewObjectValue(convertMapToScriptValue(result)), nil
// Array: return engine.NewArrayValue(convertSliceToScriptValue(result)), nil
// Nil: return engine.NewNilValue(), nil
EOF
}

# Function to create helper functions file
create_helpers() {
    local dir="$1"
    cat > "$dir/scriptvalue_helpers.go" << 'EOF'
// ABOUTME: Helper functions for ScriptValue conversions in bridge implementations
// ABOUTME: Reduces boilerplate code when converting between Go types and ScriptValues

package bridge

import "github.com/lexlapax/go-llmspell/pkg/engine"

// convertMapToScriptValue converts a map[string]interface{} to map[string]engine.ScriptValue
func convertMapToScriptValue(data map[string]interface{}) map[string]engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)
	for k, v := range data {
		result[k] = convertInterfaceToScriptValue(v)
	}
	return result
}

// convertSliceToScriptValue converts a []interface{} to []engine.ScriptValue
func convertSliceToScriptValue(data []interface{}) []engine.ScriptValue {
	result := make([]engine.ScriptValue, len(data))
	for i, v := range data {
		result[i] = convertInterfaceToScriptValue(v)
	}
	return result
}

// convertInterfaceToScriptValue converts interface{} to appropriate ScriptValue type
func convertInterfaceToScriptValue(v interface{}) engine.ScriptValue {
	switch val := v.(type) {
	case string:
		return engine.NewStringValue(val)
	case float64:
		return engine.NewNumberValue(val)
	case int:
		return engine.NewNumberValue(float64(val))
	case int64:
		return engine.NewNumberValue(float64(val))
	case bool:
		return engine.NewBoolValue(val)
	case nil:
		return engine.NewNilValue()
	case []interface{}:
		return engine.NewArrayValue(convertSliceToScriptValue(val))
	case map[string]interface{}:
		return engine.NewObjectValue(convertMapToScriptValue(val))
	default:
		// Fallback to string representation
		return engine.NewStringValue(fmt.Sprintf("%v", v))
	}
}

// validateStringArg validates that args[index] is a string and returns its value
func validateStringArg(args []engine.ScriptValue, index int, name string) (string, error) {
	if len(args) <= index {
		return "", fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeString {
		return "", fmt.Errorf("%s must be string", name)
	}
	return args[index].(engine.StringValue).Value(), nil
}

// validateNumberArg validates that args[index] is a number and returns its value
func validateNumberArg(args []engine.ScriptValue, index int, name string) (float64, error) {
	if len(args) <= index {
		return 0, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeNumber {
		return 0, fmt.Errorf("%s must be number", name)
	}
	return args[index].(engine.NumberValue).Value(), nil
}

// validateBoolArg validates that args[index] is a boolean and returns its value
func validateBoolArg(args []engine.ScriptValue, index int, name string) (bool, error) {
	if len(args) <= index {
		return false, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeBool {
		return false, fmt.Errorf("%s must be boolean", name)
	}
	return args[index].(engine.BoolValue).Value(), nil
}

// validateObjectArg validates that args[index] is an object and returns its fields as map[string]interface{}
func validateObjectArg(args []engine.ScriptValue, index int, name string) (map[string]interface{}, error) {
	if len(args) <= index {
		return nil, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeObject {
		return nil, fmt.Errorf("%s must be object", name)
	}
	
	result := make(map[string]interface{})
	for k, v := range args[index].(engine.ObjectValue).Fields() {
		result[k] = v.ToGo()
	}
	return result, nil
}

// validateArrayArg validates that args[index] is an array and returns its values as []interface{}
func validateArrayArg(args []engine.ScriptValue, index int, name string) ([]interface{}, error) {
	if len(args) <= index {
		return nil, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeArray {
		return nil, fmt.Errorf("%s must be array", name)
	}
	
	result := make([]interface{}, 0)
	for _, v := range args[index].(engine.ArrayValue).Values() {
		result = append(result, v.ToGo())
	}
	return result, nil
}
EOF
}

# Main script
echo "Starting ScriptValue conversion..."

# Check if we have arguments
if [ $# -eq 0 ]; then
    echo "Usage: $0 <bridge_file1> [bridge_file2] ..."
    echo "   or: $0 --all (converts all bridge files)"
    exit 1
fi

if [ "$1" = "--all" ]; then
    # Create helpers directory for structured package
    mkdir -p "./pkg/bridge/structured/helpers"
    create_helpers "./pkg/bridge/structured"
    
    # Create helpers for other packages as needed
    mkdir -p "./pkg/bridge/observability/helpers"
    create_helpers "./pkg/bridge/observability"
    mkdir -p "./pkg/bridge/agent/helpers"
    create_helpers "./pkg/bridge/agent"
    echo "Converting all bridge files..."
    
    # Find all bridge Go files
    find ./pkg/bridge -name "*.go" -not -name "*_test.go" -not -name "scriptvalue_helpers.go" | while read -r file; do
        convert_file "$file"
        add_validations "$file"
    done
    
    echo "All bridge files converted!"
    echo "Helper functions created in pkg/bridge/helpers/scriptvalue_helpers.go"
    echo "Validation templates created as *.validation_template files"
    echo "Original files backed up as *.backup files"
    
else
    # Convert specific files
    # Create helpers directory for the first file's package
    first_file_dir=$(dirname "$1")
    mkdir -p "$first_file_dir/helpers"
    create_helpers "$first_file_dir"
    
    for file in "$@"; do
        if [ -f "$file" ]; then
            convert_file "$file"
            add_validations "$file"
        else
            echo "Warning: File $file not found"
        fi
    done
    
    echo "Conversion complete!"
    echo "Helper functions created in $first_file_dir/helpers/scriptvalue_helpers.go"
    echo "Review the .validation_template files for manual conversion patterns"
fi

echo ""
echo "Next steps:"
echo "1. Review and manually apply validation patterns from .validation_template files"
echo "2. Add imports for the helper functions where needed"
echo "3. Replace manual conversions with helper function calls"
echo "4. Test compilation and fix any remaining issues"
echo "5. Run tests to ensure functionality is preserved"