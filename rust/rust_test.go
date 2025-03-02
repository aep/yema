package rust

import (
	"testing"

	"github.com/aep/yema"
)

func TestToRustSimple(t *testing.T) {
	// Create a simple yema.Type with a struct
	person := map[string]yema.Type{
		"name": {
			Kind: yema.String,
		},
		"age": {
			Kind: yema.Int,
		},
		"isActive": {
			Kind: yema.Bool,
		},
		"tags": {
			Kind: yema.Array,
			Array: &yema.Type{
				Kind: yema.String,
			},
		},
	}

	yemaType := &yema.Type{
		Kind:   yema.Struct,
		Struct: &person,
	}

	// Convert to Rust
	result, err := ToRust(yemaType)
	if err != nil {
		t.Fatalf("ToRust failed: %v", err)
	}

	// Print the result for inspection
	t.Logf("Generated Rust code:\n%s", string(result))
}

func TestToRustNested(t *testing.T) {
	// Create address type
	address := map[string]yema.Type{
		"street": {
			Kind: yema.String,
		},
		"city": {
			Kind: yema.String,
		},
		"zipCode": {
			Kind: yema.String,
		},
	}

	// Create a nested yema.Type with a struct containing another struct
	person := map[string]yema.Type{
		"name": {
			Kind: yema.String,
		},
		"age": {
			Kind: yema.Int,
		},
		"isActive": {
			Kind: yema.Bool,
		},
		"address": {
			Kind:   yema.Struct,
			Struct: &address,
		},
		"email": {
			Kind:     yema.String,
			Optional: true,
		},
	}

	yemaType := &yema.Type{
		Kind:   yema.Struct,
		Struct: &person,
	}

	// Convert to Rust with custom options
	options := Options{
		Module:       "test_module",
		RootType:     "Person",
		DeriveTraits: []string{"Debug", "Clone", "Serialize", "Deserialize", "PartialEq"},
	}

	result, err := ToRustWithOptions(yemaType, options)
	if err != nil {
		t.Fatalf("ToRustWithOptions failed: %v", err)
	}

	// Print the result for inspection
	t.Logf("Generated Rust code with options:\n%s", string(result))
}