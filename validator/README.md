# Yema Validator

This package provides validation functionality for Yema schemas. It allows you to check if data (in the form of a `map[string]interface{}`) conforms to a given schema defined as a `yema.Type`.

## Features

- Validates data against Yema schema definitions
- Ignores unknown fields that are not defined in the schema
- Comprehensive type checking with proper range validation
- Detailed error messages for failed validations

## Using the CLI

The Yema CLI includes a `validate` subcommand that allows you to validate JSON or YAML data against a Yema schema:

```bash
# Validate a JSON file against a schema
yema validate data.json --schema schema.yaml

# Validate YAML data from stdin
cat data.yaml | yema validate --schema schema.yaml
```

## Examples

### Valid Data

```bash
$ yema validate examples/valid-data.json --schema examples/schema.yaml
Validation successful! ✓
```

### Invalid Data

```bash
$ yema validate examples/invalid-data.yaml --schema examples/schema.yaml
Validation failed: required field 'age' is missing
```

## Using the Validator in Code

You can also use the validator directly in your Go code:

```go
import (
    "fmt"
    "github.com/aep/yema"
    "github.com/aep/yema/validator"
)

func main() {
    // Define your schema
    schema := &yema.Type{
        Kind: yema.Struct,
        Struct: &map[string]yema.Type{
            "name": {Kind: yema.String},
            "age":  {Kind: yema.Int},
        },
    }
    
    // Your data
    data := map[string]interface{}{
        "name": "John",
        "age": 30,
        "unknown_field": "this will be ignored",
    }
    
    // Validate
    if err := validator.ValidateMap(data, schema); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    } else {
        fmt.Println("Data is valid!")
    }
}
```

## Performance

The validator is designed to be fast and efficient, with minimal memory allocations. It's suitable for validating large data structures in production environments.