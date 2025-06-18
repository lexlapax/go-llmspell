#!/usr/bin/env python3

"""
ABOUTME: Python script to systematically fix ScriptValue conversions in schema.go
ABOUTME: Handles the most common patterns to reduce manual work
"""

import re
import sys

def fix_string_validations(content):
    """Fix string argument validations"""
    patterns = [
        # Fix args[0].(string) pattern
        (r'(\w+), ok := args\[(\d+)\]\.\(string\)', r'if args[\2] == nil || args[\2].Type() != engine.TypeString {\n\t\t\treturn nil, fmt.Errorf("argument must be string")\n\t\t}\n\t\t\1 := args[\2].(engine.StringValue).Value()'),
        
        # Fix direct type assertions
        (r'args\[(\d+)\]\.\(string\)', r'args[\1].(engine.StringValue).Value()'),
        (r'args\[(\d+)\]\.\(float64\)', r'args[\1].(engine.NumberValue).Value()'),
        (r'args\[(\d+)\]\.\(bool\)', r'args[\1].(engine.BoolValue).Value()'),
    ]
    
    for pattern, replacement in patterns:
        content = re.sub(pattern, replacement, content)
    
    return content

def fix_return_values(content):
    """Fix return value conversions"""
    patterns = [
        # Simple return patterns
        (r'return ([a-zA-Z_][a-zA-Z0-9_]*), nil$', r'return engine.NewStringValue(\1), nil'),
        (r'return (\d+), nil$', r'return engine.NewNumberValue(float64(\1)), nil'),
        (r'return (true|false), nil$', r'return engine.NewBoolValue(\1), nil'),
        
        # Complex return patterns that need manual review
        (r'return schemaToScript\(([^)]+)\), nil', r'return convertSchemaToScriptValue(\1), nil'),
        (r'return ([a-zA-Z_][a-zA-Z0-9_]*)\[([^\]]+)\], nil', r'// TODO: Convert array/slice return\n\t\treturn engine.NewArrayValue(helpers.ConvertSliceToScriptValue(\1[\2])), nil'),
    ]
    
    for pattern, replacement in patterns:
        content = re.sub(pattern, replacement, content, flags=re.MULTILINE)
    
    return content

def fix_object_validations(content):
    """Fix object argument validations"""
    # This is complex - create template replacements
    object_pattern = r'(\w+), ok := args\[(\d+)\]\.\(map\[string\]interface\{\}\)'
    replacement = r'''if args[\2] == nil || args[\2].Type() != engine.TypeObject {
\t\t\treturn nil, fmt.Errorf("argument must be object")
\t\t}
\t\t\1 := make(map[string]interface{})
\t\tfor k, v := range args[\2].(engine.ObjectValue).Fields() {
\t\t\t\1[k] = v.ToGo()
\t\t}'''
    
    content = re.sub(object_pattern, replacement, content)
    return content

def add_helper_import(content):
    """Add import for helpers package"""
    if 'helpers' not in content:
        # Find import block and add helper import
        import_pattern = r'(import \(\n(?:[^\)]*\n)*)'
        replacement = r'\1\t"github.com/lexlapax/go-llmspell/pkg/bridge/structured/helpers"\n'
        content = re.sub(import_pattern, replacement, content)
    
    return content

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 fix_schema_conversions.py <schema.go>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    
    # Read file
    with open(file_path, 'r') as f:
        content = f.read()
    
    # Backup original
    with open(file_path + '.py_backup', 'w') as f:
        f.write(content)
    
    print(f"Processing {file_path}...")
    
    # Apply fixes
    content = add_helper_import(content)
    content = fix_string_validations(content)
    content = fix_object_validations(content)
    content = fix_return_values(content)
    
    # Write back
    with open(file_path, 'w') as f:
        f.write(content)
    
    print(f"Fixed {file_path} (backup: {file_path}.py_backup)")
    print("Manual review recommended for complex patterns")

if __name__ == "__main__":
    main()