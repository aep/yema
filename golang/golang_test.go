package golang

import (
	"testing"

	"github.com/aep/yema"
)

func TestToGolang(t *testing.T) {
	// Create a test struct
	testStruct := &yema.Type{
		Kind: yema.Struct,
		Struct: &map[string]yema.Type{
			"name": {
				Kind: yema.String,
			},
			"age": {
				Kind: yema.Int,
			},
			"optional_field": {
				Kind:     yema.String,
				Optional: true,
			},
			"numbers": {
				Kind: yema.Array,
				Array: &yema.Type{
					Kind: yema.Int,
				},
			},
			"address": {
				Kind: yema.Struct,
				Struct: &map[string]yema.Type{
					"street": {
						Kind: yema.String,
					},
					"city": {
						Kind: yema.String,
					},
					"zipCode": {
						Kind: yema.String,
					},
				},
			},
		},
	}

	// Generate Go struct
	result, err := ToGolang(testStruct)
	if err != nil {
		t.Fatalf("Error generating Go struct: %v", err)
	}

	// Verify result
	if len(result) == 0 {
		t.Errorf("Generated Go struct is empty")
	}

	t.Logf("Generated Go struct:\n%s", string(result))
}