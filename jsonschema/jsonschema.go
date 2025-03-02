package jsonschema

import (
	"encoding/json"
	"fmt"
	"github.com/aep/yema"
)

// SchemaVersion is the JSON Schema version to use
const SchemaVersion = "http://json-schema.org/draft-07/schema#"

// JSONSchema represents a JSON Schema document
type JSONSchema struct {
	Schema      string                 `json:"$schema,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// ToJSONSchema converts an abstract Type to a JSON Schema document
func ToJSONSchema(t *yema.Type) ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type provided")
	}

	if t.Kind != yema.Struct {
		return nil, fmt.Errorf("expected root type to be Struct, got %v", t.Kind)
	}

	schema := &JSONSchema{
		Schema: SchemaVersion,
	}

	err := typeToJSONSchema(t, schema)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(schema, "", "  ")
}

func typeToJSONSchema(t *yema.Type, schema *JSONSchema) error {
	switch t.Kind {
	case yema.Bool:
		schema.Type = "boolean"
	case yema.Int, yema.Int8, yema.Int16, yema.Int32, yema.Int64,
		yema.Uint, yema.Uint8, yema.Uint16, yema.Uint32, yema.Uint64:
		schema.Type = "integer"
	case yema.Float32, yema.Float64:
		schema.Type = "number"
	case yema.String, yema.Bytes:
		schema.Type = "string"
	case yema.Array:
		schema.Type = "array"
		if t.Array != nil {
			itemSchema := &JSONSchema{}
			err := typeToJSONSchema(t.Array, itemSchema)
			if err != nil {
				return err
			}
			schema.Items = itemSchema
		}
	case yema.Struct:
		schema.Type = "object"
		if t.Struct == nil {
			return fmt.Errorf("struct type with nil Struct field")
		}

		schema.Properties = make(map[string]*JSONSchema)
		schema.Required = []string{}

		for fieldName, fieldType := range *t.Struct {
			propSchema := &JSONSchema{}
			err := typeToJSONSchema(&fieldType, propSchema)
			if err != nil {
				return err
			}

			schema.Properties[fieldName] = propSchema

			// Add to required list if not optional
			if !fieldType.Optional {
				schema.Required = append(schema.Required, fieldName)
			}
		}

		// If no required fields, omit the required array
		if len(schema.Required) == 0 {
			schema.Required = nil
		}

	default:
		return fmt.Errorf("unexpected type kind: %v", t.Kind)
	}

	return nil
}

