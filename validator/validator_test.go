package validator

import (
	"testing"

	"github.com/aep/yema"
)

func TestValidate(t *testing.T) {
	// Define schema: person with name (string), age (int), scores (array of floats),
	// and optional address (struct with street and city)
	addressSchema := map[string]yema.Type{
		"street": {Kind: yema.String},
		"city":   {Kind: yema.String},
	}

	personSchema := &yema.Type{
		Kind: yema.Struct,
		Struct: &map[string]yema.Type{
			"name":    {Kind: yema.String},
			"age":     {Kind: yema.Int},
			"scores":  {Kind: yema.Array, Array: &yema.Type{Kind: yema.Float64}},
			"address": {Kind: yema.Struct, Struct: &addressSchema, Optional: true},
		},
	}

	tests := []struct {
		name    string
		data    map[string]interface{}
		schema  *yema.Type
		wantErr bool
	}{
		{
			name: "valid basic data",
			data: map[string]interface{}{
				"name":   "John",
				"age":    30,
				"scores": []interface{}{85.5, 90.0, 77.5},
			},
			schema:  personSchema,
			wantErr: false,
		},
		{
			name: "valid data with nested struct",
			data: map[string]interface{}{
				"name":   "Jane",
				"age":    25,
				"scores": []interface{}{90.5, 92.0},
				"address": map[string]interface{}{
					"street": "123 Main St",
					"city":   "Springfield",
				},
			},
			schema:  personSchema,
			wantErr: false,
		},
		{
			name: "valid data with unknown fields",
			data: map[string]interface{}{
				"name":        "Bob",
				"age":         40,
				"scores":      []interface{}{75.0, 80.0},
				"occupation":  "Engineer", // Unknown field, should be ignored
				"phoneNumber": "555-1234", // Unknown field, should be ignored
			},
			schema:  personSchema,
			wantErr: false,
		},
		{
			name: "missing required field",
			data: map[string]interface{}{
				"name": "Alice",
				// age is missing
				"scores": []interface{}{95.0, 98.0},
			},
			schema:  personSchema,
			wantErr: true,
		},
		{
			name: "wrong type for field",
			data: map[string]interface{}{
				"name":   "Charlie",
				"age":    "thirty", // string instead of int
				"scores": []interface{}{70.0, 75.0},
			},
			schema:  personSchema,
			wantErr: true,
		},
		{
			name: "wrong type in array",
			data: map[string]interface{}{
				"name":   "Dave",
				"age":    35,
				"scores": []interface{}{80.0, "not a number", 85.0}, // string in float array
			},
			schema:  personSchema,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIntegerRanges(t *testing.T) {
	schema := &yema.Type{
		Kind: yema.Struct,
		Struct: &map[string]yema.Type{
			"int8Val":  {Kind: yema.Int8},
			"int16Val": {Kind: yema.Int16},
			"int32Val": {Kind: yema.Int32},
			"int64Val": {Kind: yema.Int64},
		},
	}

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid integer values",
			data: map[string]interface{}{
				"int8Val":  127,
				"int16Val": 32767,
				"int32Val": 2147483647,
				"int64Val": 9223372036854775807,
			},
			wantErr: false,
		},
		{
			name: "int8 out of range",
			data: map[string]interface{}{
				"int8Val":  128, // out of range for int8
				"int16Val": 32767,
				"int32Val": 2147483647,
				"int64Val": 9223372036854775807,
			},
			wantErr: true,
		},
		{
			name: "int16 out of range",
			data: map[string]interface{}{
				"int8Val":  127,
				"int16Val": 32768, // out of range for int16
				"int32Val": 2147483647,
				"int64Val": 9223372036854775807,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUnsignedIntegerRanges(t *testing.T) {
	schema := &yema.Type{
		Kind: yema.Struct,
		Struct: &map[string]yema.Type{
			"uint8Val":  {Kind: yema.Uint8},
			"uint16Val": {Kind: yema.Uint16},
			"uint32Val": {Kind: yema.Uint32},
			"uint64Val": {Kind: yema.Uint64},
		},
	}

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid unsigned integer values",
			data: map[string]interface{}{
				"uint8Val":  255,
				"uint16Val": 65535,
				"uint32Val": 4294967295,
				"uint64Val": uint64(9223372036854775807), // Max int64 value - within uint64 range
			},
			wantErr: false,
		},
		{
			name: "uint8 out of range",
			data: map[string]interface{}{
				"uint8Val":  256, // out of range for uint8
				"uint16Val": 65535,
				"uint32Val": 4294967295,
				"uint64Val": uint64(9223372036854775807),
			},
			wantErr: true,
		},
		{
			name: "negative value for unsigned",
			data: map[string]interface{}{
				"uint8Val":  -1, // negative, invalid for uint
				"uint16Val": 65535,
				"uint32Val": 4294967295,
				"uint64Val": uint64(9223372036854775807),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkValidateMap(b *testing.B) {
	// Define a complex schema for benchmarking
	addressSchema := map[string]yema.Type{
		"street":     {Kind: yema.String},
		"city":       {Kind: yema.String},
		"postalCode": {Kind: yema.String},
		"country":    {Kind: yema.String},
	}

	contactSchema := map[string]yema.Type{
		"email":    {Kind: yema.String},
		"phone":    {Kind: yema.String},
		"isActive": {Kind: yema.Bool},
	}

	personSchema := &yema.Type{
		Kind: yema.Struct,
		Struct: &map[string]yema.Type{
			"name":      {Kind: yema.String},
			"age":       {Kind: yema.Int},
			"height":    {Kind: yema.Float64},
			"isStudent": {Kind: yema.Bool},
			"scores":    {Kind: yema.Array, Array: &yema.Type{Kind: yema.Float64}},
			"tags":      {Kind: yema.Array, Array: &yema.Type{Kind: yema.String}},
			"address":   {Kind: yema.Struct, Struct: &addressSchema},
			"contact":   {Kind: yema.Struct, Struct: &contactSchema},
		},
	}

	// Create a valid data instance
	data := map[string]interface{}{
		"name":      "John Doe",
		"age":       30,
		"height":    185.5,
		"isStudent": false,
		"scores":    []interface{}{85.5, 90.0, 77.5, 82.0},
		"tags":      []interface{}{"student", "engineering", "graduate"},
		"address": map[string]interface{}{
			"street":     "123 Main St",
			"city":       "Springfield",
			"postalCode": "12345",
			"country":    "USA",
		},
		"contact": map[string]interface{}{
			"email":    "john.doe@example.com",
			"phone":    "555-1234",
			"isActive": true,
		},
		"extraField1": "This field is not in the schema",
		"extraField2": 42,
		"extraField3": true,
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Validate(data, personSchema)
	}
}

