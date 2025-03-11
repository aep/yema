// Package validator provides validation functions for yema.Type
package validator

import (
	"encoding/json"
	"fmt"
	"github.com/aep/yema"
	"strconv"
)

// Validate checks if a map[string]interface{} matches a given yema.Type
func Validate(data map[string]interface{}, schema *yema.Type) []error {
	if schema == nil || schema.Struct == nil {
		return []error{fmt.Errorf("invalid schema")}
	}

	var errors []error

	// For each field in the schema, validate the corresponding field in the data
	for fieldName, fieldType := range *schema.Struct {
		value, exists := data[fieldName]

		// If the field doesn't exist in the data
		if !exists {
			// Check if it's optional
			if !fieldType.Optional {
				errors = append(errors, fmt.Errorf("required field '%s' is missing", fieldName))
			}
			// Skip validation for optional fields that don't exist
			continue
		}

		// Field exists, validate it against the field type
		if err := validateValue(value, &fieldType, fieldName); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateValue checks if a single value matches a yema.Type specification
func validateValue(value interface{}, schema *yema.Type, path string) error {
	// Handle nil values
	if value == nil {
		if schema.Optional {
			return nil
		}
		return fmt.Errorf("field '%s' is nil but not optional", path)
	}

	switch schema.Kind {
	case yema.Bool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' must be a boolean", path)
		}

	case yema.String:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' must be a string", path)
		}

	case yema.Int, yema.Int8, yema.Int16, yema.Int32, yema.Int64:
		return validateIntValue(value, schema.Kind, path)

	case yema.Uint, yema.Uint8, yema.Uint16, yema.Uint32, yema.Uint64:
		return validateUintValue(value, schema.Kind, path)

	case yema.Float32, yema.Float64:
		return validateFloatValue(value, schema.Kind, path)

	case yema.Array:
		if schema.Array == nil {
			return fmt.Errorf("array type definition for '%s' is nil", path)
		}

		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("field '%s' must be an array", path)
		}

		// Validate each element in the array
		for i, elem := range arr {
			elemPath := path + "[" + strconv.Itoa(i) + "]"
			if err := validateValue(elem, schema.Array, elemPath); err != nil {
				return err
			}
		}

	case yema.Struct:
		if schema.Struct == nil {
			return fmt.Errorf("struct type definition for '%s' is nil", path)
		}

		mapValue, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field '%s' must be a map[string]interface{}", path)
		}

		// For each field in the schema, validate the corresponding field in the data
		for fieldName, fieldType := range *schema.Struct {
			nestedValue, exists := mapValue[fieldName]

			// If the field doesn't exist in the data
			if !exists {
				// Check if it's optional
				if !fieldType.Optional {
					return fmt.Errorf("required field '%s.%s' is missing", path, fieldName)
				}
				// Skip validation for optional fields that don't exist
				continue
			}

			// Field exists, validate it against the field type
			nestedPath := path + "." + fieldName
			if err := validateValue(nestedValue, &fieldType, nestedPath); err != nil {
				return err
			}
		}

	case yema.Bytes:
		// Accept both []byte and string for bytes type
		if _, ok := value.([]byte); !ok {
			if _, ok := value.(string); !ok {
				return fmt.Errorf("field '%s' must be bytes or string", path)
			}
		}

	default:
		return fmt.Errorf("unsupported type %v for field '%s'", schema.Kind, path)
	}

	return nil
}

// validateIntValue handles validation of integer types with proper range checking
func validateIntValue(value interface{}, kind yema.Kind, path string) error {
	// Check for various numeric types from JSON unmarshaling
	var intVal int64
	var isInt bool

	switch v := value.(type) {
	case json.Number:
		{
			var err error
			intVal, err = v.Int64()
			isInt = err == nil
		}
	case int:
		intVal, isInt = int64(v), true
	case int8:
		intVal, isInt = int64(v), true
	case int16:
		intVal, isInt = int64(v), true
	case int32:
		intVal, isInt = int64(v), true
	case int64:
		intVal, isInt = v, true
	case float64: // JSON numbers typically come as float64
		if v == float64(int64(v)) { // Check if it's a whole number
			intVal, isInt = int64(v), true
		}
	}

	if !isInt {
		return fmt.Errorf("field '%s' must be an integer", path)
	}

	// Range validation
	switch kind {
	case yema.Int8:
		if intVal < -128 || intVal > 127 {
			return fmt.Errorf("field '%s' value out of range for int8", path)
		}
	case yema.Int16:
		if intVal < -32768 || intVal > 32767 {
			return fmt.Errorf("field '%s' value out of range for int16", path)
		}
	case yema.Int32:
		if intVal < -2147483648 || intVal > 2147483647 {
			return fmt.Errorf("field '%s' value out of range for int32", path)
		}
	case yema.Int64, yema.Int:
		// No range check needed for int64 (handled by conversion)
		// For yema.Int we also don't check, as it maps to Go's int which can be 32 or 64 bits
	}

	return nil
}

// validateUintValue handles validation of unsigned integer types with range checking
func validateUintValue(value interface{}, kind yema.Kind, path string) error {
	// Check for various numeric types from JSON unmarshaling
	var uintVal uint64
	var isUint bool

	switch v := value.(type) {
	case json.Number:
		{
			intVal, err := v.Int64()
			if intVal >= 0 {
				uintVal = uint64(intVal)
				isUint = err == nil
			}
		}
	case uint:
		uintVal, isUint = uint64(v), true
	case uint8:
		uintVal, isUint = uint64(v), true
	case uint16:
		uintVal, isUint = uint64(v), true
	case uint32:
		uintVal, isUint = uint64(v), true
	case uint64:
		uintVal, isUint = v, true
	case int:
		if v >= 0 {
			uintVal, isUint = uint64(v), true
		}
	case int8:
		if v >= 0 {
			uintVal, isUint = uint64(v), true
		}
	case int16:
		if v >= 0 {
			uintVal, isUint = uint64(v), true
		}
	case int32:
		if v >= 0 {
			uintVal, isUint = uint64(v), true
		}
	case int64:
		if v >= 0 {
			uintVal, isUint = uint64(v), true
		}
	case float64: // JSON numbers typically come as float64
		if v >= 0 && v == float64(uint64(v)) { // Check if it's a non-negative whole number
			uintVal, isUint = uint64(v), true
		}
	}

	if !isUint {
		return fmt.Errorf("field '%s' must be a non-negative integer", path)
	}

	// Range validation
	switch kind {
	case yema.Uint8:
		if uintVal > 255 {
			return fmt.Errorf("field '%s' value out of range for uint8", path)
		}
	case yema.Uint16:
		if uintVal > 65535 {
			return fmt.Errorf("field '%s' value out of range for uint16", path)
		}
	case yema.Uint32:
		if uintVal > 4294967295 {
			return fmt.Errorf("field '%s' value out of range for uint32", path)
		}
	case yema.Uint64, yema.Uint:
		// No range check needed for uint64 (handled by conversion)
		// For yema.Uint we also don't check, as it maps to Go's uint which can be 32 or 64 bits
	}

	return nil
}

// validateFloatValue handles validation of float types
func validateFloatValue(value interface{}, kind yema.Kind, path string) error {
	var floatVal float64
	var isFloat bool

	switch v := value.(type) {
	case json.Number:
		{
			var err error
			floatVal, err = v.Float64()
			isFloat = err == nil
		}
	case float32:
		floatVal, isFloat = float64(v), true
	case float64:
		floatVal, isFloat = v, true
	case int:
		floatVal, isFloat = float64(v), true
	case int8:
		floatVal, isFloat = float64(v), true
	case int16:
		floatVal, isFloat = float64(v), true
	case int32:
		floatVal, isFloat = float64(v), true
	case int64:
		floatVal, isFloat = float64(v), true
	case uint:
		floatVal, isFloat = float64(v), true
	case uint8:
		floatVal, isFloat = float64(v), true
	case uint16:
		floatVal, isFloat = float64(v), true
	case uint32:
		floatVal, isFloat = float64(v), true
	case uint64:
		floatVal, isFloat = float64(v), true
	}

	if !isFloat {
		return fmt.Errorf("field '%s' must be a number", path)
	}

	// Float32 range check (approximation)
	if kind == yema.Float32 {
		if floatVal > 3.4e38 || floatVal < -3.4e38 {
			return fmt.Errorf("field '%s' value out of range for float32", path)
		}
	}

	return nil
}
