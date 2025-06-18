#!/bin/bash

# ABOUTME: Quick script to fix the most common type assertion patterns in schema.go
# ABOUTME: Targets specific lines that need conversion from old interface{} to ScriptValue

set -e

file="pkg/bridge/structured/schema.go"

echo "Fixing schema.go type assertions..."

# Fix string type assertions with validation pattern
sed -i 's/\([a-zA-Z_][a-zA-Z0-9_]*\), ok := args\[\([0-9]\)\]\.\(string\)/\1, err := ValidateStringArg(args, \2, "\1")\n\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}/g' "$file"

# Fix remaining direct type assertions that weren't caught
sed -i 's/args\[\([0-9]\)\]\.\(string\)/args[\1].(engine.StringValue).Value()/g' "$file"

# Fix bool type assertions  
sed -i 's/\([a-zA-Z_][a-zA-Z0-9_]*\), ok := args\[\([0-9]\)\]\.\(bool\)/\1, err := ValidateBoolArg(args, \2, "\1")\n\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}/g' "$file"

# Fix remaining direct bool assertions
sed -i 's/args\[\([0-9]\)\]\.\(bool\)/args[\1].(engine.BoolValue).Value()/g' "$file"

# Fix object type assertions
sed -i 's/\([a-zA-Z_][a-zA-Z0-9_]*\), ok := args\[\([0-9]\)\]\.\(map\[string\]interface{}\)/\1, err := ValidateObjectArg(args, \2, "\1")\n\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}/g' "$file"

echo "Fixed common type assertion patterns in schema.go"