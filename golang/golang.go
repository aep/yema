package golang

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/aep/yema"
)

// Options holds configuration options for Go code generation
type Options struct {
	// Package is the name of the Go package to generate
	Package string
	// RootType is the name of the root struct type
	RootType string
}

// ToGolangWithOptions converts a yema.Type to Go struct definitions with custom options
func ToGolang(t *yema.Type, opts Options) ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type provided")
	}

	if t.Kind != yema.Struct {
		return nil, fmt.Errorf("expected root type to be Struct, got %v", t.Kind)
	}

	// Use default values if not provided
	if opts.Package == "" {
		opts.Package = "generated"
	}
	if opts.RootType == "" {
		opts.RootType = "Root"
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("package %s\n\n", opts.Package))

	// Process the root struct
	err := generateStructs(t, opts.RootType, &buf, make(map[string]bool))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// generateStructs recursively generates Go struct definitions
func generateStructs(t *yema.Type, structName string, buf *bytes.Buffer, generatedStructs map[string]bool) error {
	if t.Kind != yema.Struct {
		return fmt.Errorf("expected Struct type, got %v", t.Kind)
	}

	// Don't regenerate structs we've already processed
	if generatedStructs[structName] {
		return nil
	}

	// Mark this struct as generated
	generatedStructs[structName] = true

	// Start struct definition
	fmt.Fprintf(buf, "// %s represents a generated struct\n", structName)
	fmt.Fprintf(buf, "type %s struct {\n", structName)

	// Track any nested structs we need to generate
	nestedStructs := make(map[string]*yema.Type)

	// Process all fields in the struct
	for fieldName, fieldType := range *t.Struct {
		goFieldName := toCamelCase(fieldName)
		goFieldType, nestedName, err := typeToGoType(&fieldType, structName, fieldName)
		if err != nil {
			return err
		}

		// Check if this field requires a nested struct to be generated
		if nestedName != "" && fieldType.Kind == yema.Struct {
			nestedStructs[nestedName] = &yema.Type{
				Kind:   yema.Struct,
				Struct: fieldType.Struct,
			}
		} else if nestedName != "" && fieldType.Kind == yema.Array && fieldType.Array.Kind == yema.Struct {
			nestedStructs[nestedName] = &yema.Type{
				Kind:   yema.Struct,
				Struct: fieldType.Array.Struct,
			}
		}

		// Add json tag
		jsonTag := fieldName
		if fieldType.Optional {
			jsonTag += ",omitempty"
		}

		// Write field definition
		fmt.Fprintf(buf, "\t%s %s `json:\"%s\"`\n", goFieldName, goFieldType, jsonTag)
	}

	// Close struct definition
	fmt.Fprintf(buf, "}\n\n")

	// Generate any nested struct definitions
	for nestedName, nestedStruct := range nestedStructs {
		err := generateStructs(nestedStruct, nestedName, buf, generatedStructs)
		if err != nil {
			return err
		}
	}

	return nil
}

// typeToGoType converts a yema.Type to a Go type string
func typeToGoType(t *yema.Type, parentName, fieldName string) (string, string, error) {
	var goType string
	var nestedStructName string

	switch t.Kind {
	case yema.Bool:
		goType = "bool"
	case yema.Int:
		goType = "int"
	case yema.Int8:
		goType = "int8"
	case yema.Int16:
		goType = "int16"
	case yema.Int32:
		goType = "int32"
	case yema.Int64:
		goType = "int64"
	case yema.Uint:
		goType = "uint"
	case yema.Uint8:
		goType = "uint8"
	case yema.Uint16:
		goType = "uint16"
	case yema.Uint32:
		goType = "uint32"
	case yema.Uint64:
		goType = "uint64"
	case yema.Float32:
		goType = "float32"
	case yema.Float64:
		goType = "float64"
	case yema.String:
		goType = "string"
	case yema.Bytes:
		goType = "[]byte"
	case yema.Array:
		if t.Array == nil {
			return "", "", fmt.Errorf("array type with nil Array field")
		}
		elemType, elemNestedName, err := typeToGoType(t.Array, parentName, fieldName)
		if err != nil {
			return "", "", err
		}
		goType = "[]" + elemType
		nestedStructName = elemNestedName
	case yema.Struct:
		// Create a name for the nested struct
		nestedStructName = parentName + toCamelCase(fieldName)
		goType = nestedStructName
	default:
		return "", "", fmt.Errorf("unexpected type kind: %v", t.Kind)
	}

	if t.Optional {
		// For optional fields (except slices which are already nullable)
		if t.Kind != yema.Array && t.Kind != yema.Bytes {
			goType = "*" + goType
		}
	}

	return goType, nestedStructName, nil
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

