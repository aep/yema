package typescript

import (
	"testing"

	"github.com/aep/yema"
)

func TestToTypeScript(t *testing.T) {
	// Create a sample type structure
	address := map[string]yema.Type{
		"street": {Kind: yema.String},
		"city":   {Kind: yema.String},
		"zip":    {Kind: yema.String},
	}

	contacts := map[string]yema.Type{
		"email": {Kind: yema.String, Optional: true},
		"phone": {Kind: yema.String},
	}

	userType := &yema.Type{
		Kind: yema.Struct,
		Struct: &map[string]yema.Type{
			"id":        {Kind: yema.Int},
			"name":      {Kind: yema.String},
			"is_active": {Kind: yema.Bool},
			"tags":      {Kind: yema.Array, Array: &yema.Type{Kind: yema.String}},
			"address":   {Kind: yema.Struct, Struct: &address},
			"contacts":  {Kind: yema.Struct, Struct: &contacts, Optional: true},
			"scores":    {Kind: yema.Array, Array: &yema.Type{Kind: yema.Float64}},
		},
	}

	// Generate TypeScript
	ts, err := ToTypeScript(userType)
	if err != nil {
		t.Fatalf("Failed to generate TypeScript: %v", err)
	}

	// Print the generated TypeScript for inspection
	t.Logf("Generated TypeScript:\n%s", string(ts))

	// With custom options
	customOpts := Options{
		Namespace:    "API",
		RootType:     "User",
		UseInterfaces: true,
		ExportAll:    false,
	}

	tsWithOpts, err := ToTypeScriptWithOptions(userType, customOpts)
	if err != nil {
		t.Fatalf("Failed to generate TypeScript with options: %v", err)
	}

	// Print the generated TypeScript with custom options for inspection
	t.Logf("Generated TypeScript with custom options:\n%s", string(tsWithOpts))
}