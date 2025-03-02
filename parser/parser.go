package parser

import (
	"fmt"
	"github.com/aep/yema"
	"strings"
	"unicode"
	"unicode/utf8"
)

func isValidFieldName(name string) bool {
	if name == "" {
		return false
	}

	// Check first character
	first, width := utf8.DecodeRuneInString(name)
	if !(unicode.IsLetter(first) || first == '_') {
		return false
	}

	// Check rest of the characters
	for _, r := range name[width:] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}

	return true
}

func From(schema map[string]interface{}) (*yema.Type, error) {
	structType := make(map[string]yema.Type)

	for key, value := range schema {
		isOptional := false
		fieldName := key
		if strings.HasSuffix(key, "?") {
			isOptional = true
			fieldName = key[:len(key)-1]
		}

		if !isValidFieldName(fieldName) {
			return nil, fmt.Errorf("invalid field name: %q", fieldName)
		}

		fieldType, err := parseValueToType(fieldName, value, isOptional)
		if err != nil {
			return nil, err
		}

		structType[fieldName] = fieldType
	}

	return &yema.Type{
		Kind:   yema.Struct,
		Struct: &structType,
	}, nil
}

func parseValueToType(fieldName string, value interface{}, isOptional bool) (yema.Type, error) {
	switch v := value.(type) {
	case string:
		var kind yema.Kind
		switch v {
		case "bool":
			kind = yema.Bool
		case "int":
			kind = yema.Int
		case "int8":
			kind = yema.Int8
		case "int16":
			kind = yema.Int16
		case "int32":
			kind = yema.Int32
		case "int64":
			kind = yema.Int64
		case "uint":
			kind = yema.Uint
		case "uint8":
			kind = yema.Uint8
		case "uint16":
			kind = yema.Uint16
		case "uint32":
			kind = yema.Uint32
		case "uint64":
			kind = yema.Uint64
		case "float32":
			kind = yema.Float32
		case "float64":
			kind = yema.Float64
		case "string":
			kind = yema.String
		case "bytes":
			kind = yema.Bytes
		default:
			return yema.Type{}, fmt.Errorf("failed parsing field '%s', expected type, not: %s", fieldName, v)
		}
		return yema.Type{
			Kind:     kind,
			Optional: isOptional,
		}, nil

	case []interface{}:
		if len(v) == 0 {
			return yema.Type{}, fmt.Errorf("failed parsing field '%s', must declare type of array item", fieldName)
		}

		if len(v) > 1 {
			return yema.Type{}, fmt.Errorf("failed parsing field '%s', can only declare type of array items once", fieldName)
		}

		// Parse the array item type
		itemType, err := parseValueToType(fieldName, v[0], false)
		if err != nil {
			return yema.Type{}, err
		}

		return yema.Type{
			Kind:     yema.Array,
			Optional: isOptional,
			Array:    &itemType,
		}, nil

	case map[string]interface{}:
		nestedStruct := make(map[string]yema.Type)

		for k, val := range v {
			nestedIsOptional := false
			nestedFieldName := k
			if strings.HasSuffix(k, "?") {
				nestedIsOptional = true
				nestedFieldName = k[:len(k)-1]
			}

			if !isValidFieldName(nestedFieldName) {
				return yema.Type{}, fmt.Errorf("invalid field name: %q", nestedFieldName)
			}

			nestedType, err := parseValueToType(nestedFieldName, val, nestedIsOptional)
			if err != nil {
				return yema.Type{}, err
			}

			nestedStruct[nestedFieldName] = nestedType
		}

		return yema.Type{
			Kind:     yema.Struct,
			Optional: isOptional,
			Struct:   &nestedStruct,
		}, nil
	default:
		return yema.Type{}, fmt.Errorf("failed parsing field '%s', expected type, not: %s", fieldName, v)
	}
}
