package rust

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/aep/yema"
)

// Options holds configuration options for Rust code generation
type Options struct {
	// Module is the name of the Rust module to generate
	Module string
	// RootType is the name of the root struct type
	RootType string
	// DeriveTraits specifies which traits to automatically derive for structs
	DeriveTraits []string
	// UseSerdeRename determines whether to use serde rename attributes for JSON field names
	UseSerdeRename bool
}

// ToRustWithOptions converts a yema.Type to Rust struct definitions with custom options
func ToRust(t *yema.Type, opts Options) ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type provided")
	}

	if t.Kind != yema.Struct {
		return nil, fmt.Errorf("expected root type to be Struct, got %v", t.Kind)
	}

	// Use default values if not provided
	if opts.Module == "" {
		opts.Module = "generated"
	}
	if opts.RootType == "" {
		opts.RootType = "Root"
	}
	if len(opts.DeriveTraits) == 0 {
		opts.DeriveTraits = []string{"Debug", "Clone", "Serialize", "Deserialize"}
	}

	var buf bytes.Buffer

	// Add module declaration
	if opts.Module != "" {
		buf.WriteString(fmt.Sprintf("pub mod %s {\n", opts.Module))
		// Add serde import if we're using it
		if containsTrait(opts.DeriveTraits, "Serialize") || containsTrait(opts.DeriveTraits, "Deserialize") {
			buf.WriteString("    use serde::{Serialize, Deserialize};\n\n")
		}
	}

	// Process the root struct
	err := generateStructs(t, opts.RootType, &buf, make(map[string]bool), opts, 1)
	if err != nil {
		return nil, err
	}

	// Close module if needed
	if opts.Module != "" {
		buf.WriteString("}\n")
	}

	return buf.Bytes(), nil
}

// generateStructs recursively generates Rust struct definitions
func generateStructs(t *yema.Type, structName string, buf *bytes.Buffer, generatedStructs map[string]bool, opts Options, indentLevel int) error {
	if t.Kind != yema.Struct {
		return fmt.Errorf("expected Struct type, got %v", t.Kind)
	}

	// Don't regenerate structs we've already processed
	if generatedStructs[structName] {
		return nil
	}

	// Mark this struct as generated
	generatedStructs[structName] = true

	indent := strings.Repeat("    ", indentLevel)

	// Add derive attributes if any provided
	if len(opts.DeriveTraits) > 0 {
		fmt.Fprintf(buf, "%s#[derive(%s)]\n", indent, strings.Join(opts.DeriveTraits, ", "))
	}

	// Start struct definition
	fmt.Fprintf(buf, "%s/// %s represents a generated struct\n", indent, structName)
	fmt.Fprintf(buf, "%spub struct %s {\n", indent, structName)

	// Track any nested structs we need to generate
	nestedStructs := make(map[string]*yema.Type)

	// Process all fields in the struct
	for fieldName, fieldType := range *t.Struct {
		rustFieldName := toSnakeCase(fieldName)
		rustFieldType, nestedName, err := typeToRustType(&fieldType, structName, fieldName)
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

		// Add field documentation
		fmt.Fprintf(buf, "%s    /// %s field\n", indent, fieldName)

		// Add serde rename attribute if the field name is different from JSON field
		if opts.UseSerdeRename && rustFieldName != fieldName {
			if fieldType.Optional {
				fmt.Fprintf(buf, "%s    #[serde(rename = \"%s\", skip_serializing_if = \"Option::is_none\")]\n", indent, fieldName)
			} else {
				fmt.Fprintf(buf, "%s    #[serde(rename = \"%s\")]\n", indent, fieldName)
			}
		} else if opts.UseSerdeRename && fieldType.Optional {
			fmt.Fprintf(buf, "%s    #[serde(skip_serializing_if = \"Option::is_none\")]\n", indent)
		}

		// Write field definition
		fmt.Fprintf(buf, "%s    pub %s: %s,\n", indent, rustFieldName, rustFieldType)
	}

	// Close struct definition
	fmt.Fprintf(buf, "%s}\n\n", indent)

	// Generate any nested struct definitions
	for nestedName, nestedStruct := range nestedStructs {
		err := generateStructs(nestedStruct, nestedName, buf, generatedStructs, opts, indentLevel)
		if err != nil {
			return err
		}
	}

	return nil
}

// typeToRustType converts a yema.Type to a Rust type string
func typeToRustType(t *yema.Type, parentName, fieldName string) (string, string, error) {
	var rustType string
	var nestedStructName string

	switch t.Kind {
	case yema.Bool:
		rustType = "bool"
	case yema.Int:
		rustType = "i32"
	case yema.Int8:
		rustType = "i8"
	case yema.Int16:
		rustType = "i16"
	case yema.Int32:
		rustType = "i32"
	case yema.Int64:
		rustType = "i64"
	case yema.Uint:
		rustType = "u32"
	case yema.Uint8:
		rustType = "u8"
	case yema.Uint16:
		rustType = "u16"
	case yema.Uint32:
		rustType = "u32"
	case yema.Uint64:
		rustType = "u64"
	case yema.Float32:
		rustType = "f32"
	case yema.Float64:
		rustType = "f64"
	case yema.String:
		rustType = "String"
	case yema.Bytes:
		rustType = "Vec<u8>"
	case yema.Array:
		if t.Array == nil {
			return "", "", fmt.Errorf("array type with nil Array field")
		}
		elemType, elemNestedName, err := typeToRustType(t.Array, parentName, fieldName)
		if err != nil {
			return "", "", err
		}
		rustType = "Vec<" + elemType + ">"
		nestedStructName = elemNestedName
	case yema.Struct:
		// Create a name for the nested struct
		nestedStructName = parentName + toCamelCase(fieldName)
		rustType = nestedStructName
	default:
		return "", "", fmt.Errorf("unexpected type kind: %v", t.Kind)
	}

	if t.Optional {
		rustType = "Option<" + rustType + ">"
	}

	return rustType, nestedStructName, nil
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

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	var result string
	for i, char := range s {
		if unicode.IsUpper(char) {
			if i > 0 {
				result += "_"
			}
			result += string(unicode.ToLower(char))
		} else {
			result += string(char)
		}
	}
	return result
}

// containsTrait checks if a trait is in the derive list
func containsTrait(traits []string, target string) bool {
	for _, t := range traits {
		if t == target {
			return true
		}
	}
	return false
}
