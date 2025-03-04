package typescript

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/aep/yema"
)

// Options holds configuration options for TypeScript code generation
type Options struct {
	// Namespace is the TypeScript namespace to use (if any)
	Namespace string
	// RootType is the name of the root interface type
	RootType string
	// UseInterfaces determines whether to generate interfaces (true) or types (false)
	UseInterfaces bool
	// ExportAll determines whether to export all types (true) or just the root type (false)
	ExportAll bool
}

// ToTypeScriptWithOptions converts a yema.Type to TypeScript definitions with custom options
func ToTypeScript(t *yema.Type, opts Options) ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type provided")
	}

	if t.Kind != yema.Struct {
		return nil, fmt.Errorf("expected root type to be Struct, got %v", t.Kind)
	}

	// Use default values if not provided
	if opts.RootType == "" {
		opts.RootType = "Root"
	}
	if !opts.UseInterfaces {
		opts.UseInterfaces = true // Default to interfaces
	}

	var buf bytes.Buffer

	// Write namespace if provided
	if opts.Namespace != "" {
		buf.WriteString(fmt.Sprintf("namespace %s {\n\n", opts.Namespace))
	}

	// Process the root struct
	err := generateInterfaces(t, opts.RootType, &buf, make(map[string]bool), opts)
	if err != nil {
		return nil, err
	}

	// Close namespace if needed
	if opts.Namespace != "" {
		buf.WriteString("}\n")
	}

	return buf.Bytes(), nil
}

// generateInterfaces recursively generates TypeScript interface definitions
func generateInterfaces(t *yema.Type, typeName string, buf *bytes.Buffer, generatedTypes map[string]bool, opts Options) error {
	if t.Kind != yema.Struct {
		return fmt.Errorf("expected Struct type, got %v", t.Kind)
	}

	// Don't regenerate types we've already processed
	if generatedTypes[typeName] {
		return nil
	}

	// Mark this type as generated
	generatedTypes[typeName] = true

	// Start type definition
	fmt.Fprintf(buf, "/**\n * %s represents a generated type\n */\n", typeName)

	// Determine export keyword
	exportKeyword := ""
	if opts.ExportAll || typeName == opts.RootType {
		exportKeyword = "export "
	}

	// Use interface or type alias based on options
	if opts.UseInterfaces {
		fmt.Fprintf(buf, "%sinterface %s {\n", exportKeyword, typeName)
	} else {
		fmt.Fprintf(buf, "%stype %s = {\n", exportKeyword, typeName)
	}

	// Track any nested types we need to generate
	nestedTypes := make(map[string]*yema.Type)

	// Process all fields in the struct
	for fieldName, fieldType := range *t.Struct {
		var tsSuffix string
		if fieldType.Optional {
			tsSuffix = "?"
		}
		tsFieldType, nestedName, err := typeToTypeScriptType(&fieldType, typeName, fieldName)
		if err != nil {
			return err
		}

		// Check if this field requires a nested type to be generated
		if nestedName != "" && fieldType.Kind == yema.Struct {
			nestedTypes[nestedName] = &yema.Type{
				Kind:   yema.Struct,
				Struct: fieldType.Struct,
			}
		} else if nestedName != "" && fieldType.Kind == yema.Array && fieldType.Array.Kind == yema.Struct {
			nestedTypes[nestedName] = &yema.Type{
				Kind:   yema.Struct,
				Struct: fieldType.Array.Struct,
			}
		}

		// Write field definition
		fmt.Fprintf(buf, "  %s%s: %s;\n", fieldName, tsSuffix, tsFieldType)
	}

	// Close type definition
	if opts.UseInterfaces {
		fmt.Fprintf(buf, "}\n\n")
	} else {
		fmt.Fprintf(buf, "};\n\n")
	}

	// Generate any nested type definitions
	for nestedName, nestedStruct := range nestedTypes {
		err := generateInterfaces(nestedStruct, nestedName, buf, generatedTypes, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

// typeToTypeScriptType converts a yema.Type to a TypeScript type string
func typeToTypeScriptType(t *yema.Type, parentName, fieldName string) (string, string, error) {
	var tsType string
	var nestedStructName string

	switch t.Kind {
	case yema.Bool:
		tsType = "boolean"
	case yema.Int, yema.Int8, yema.Int16, yema.Int32, yema.Int64,
		yema.Uint, yema.Uint8, yema.Uint16, yema.Uint32, yema.Uint64,
		yema.Float32, yema.Float64:
		tsType = "number"
	case yema.String:
		tsType = "string"
	case yema.Bytes:
		tsType = "Uint8Array"
	case yema.Array:
		if t.Array == nil {
			return "", "", fmt.Errorf("array type with nil Array field")
		}
		elemType, elemNestedName, err := typeToTypeScriptType(t.Array, parentName, fieldName)
		if err != nil {
			return "", "", err
		}
		tsType = elemType + "[]"
		nestedStructName = elemNestedName
	case yema.Struct:
		// Create a name for the nested type
		nestedStructName = parentName + toCamelCase(fieldName)
		tsType = nestedStructName
	default:
		return "", "", fmt.Errorf("unexpected type kind: %v", t.Kind)
	}

	return tsType, nestedStructName, nil
}

// toCamelCase converts a string to CamelCase
func toCamelCase(s string) string {
	var result string
	nextUpper := true

	for _, char := range s {
		if char == '_' || char == '-' || char == ' ' {
			nextUpper = true
			continue
		}

		if nextUpper {
			result += string(unicode.ToUpper(char))
			nextUpper = false
		} else {
			result += string(char)
		}
	}

	return result
}

