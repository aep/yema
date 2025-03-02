package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aep/yema/parser"
	"github.com/aep/yema/validator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var validateCmd = &cobra.Command{
	Use:   "validate [schema] [subject]",
	Short: "Validate data against a Yema schema",
	Long: `Validate JSON or YAML data against a Yema schema.
This command checks if the provided data conforms to the specified schema.
Unknown fields not defined in the schema are ignored during validation.

Example:
  yema validate data.json --schema schema.yaml`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		schemaData, err := os.ReadFile(args[0])
		if err != nil {
			log.Fatalf("Error reading schema file: %v", err)
		}

		var schemaMap map[string]interface{}
		err = yaml.Unmarshal(schemaData, &schemaMap)
		if err != nil {
			log.Fatalf("Error parsing schema file: %v", err)
		}

		// Convert schema to yema.Type
		schema, err := parser.From(schemaMap)
		if err != nil {
			log.Fatalf("Error parsing schema: %v", err)
		}

		// Read input data (from file or stdin)
		var input io.Reader = os.Stdin
		if len(args) > 1 {
			file, err := os.Open(args[1])
			if err != nil {
				log.Fatalf("Error opening data file: %v", err)
			}
			defer file.Close()
			input = file
		}

		// Read all data from input
		inputData, err := io.ReadAll(input)
		if err != nil {
			log.Fatalf("Error reading input data: %v", err)
		}

		// Parse input data based on extension or try both formats
		var dataMap map[string]interface{}
		err = yaml.Unmarshal(inputData, &dataMap)
		if err != nil {
			log.Fatalf("Error parsing input data: %v", err)
		}

		// Validate the data against the schema
		if err := validator.Validate(dataMap, schema); len(err) != 0 {
			fmt.Println("Validation failed")
			for _, e := range err {
				fmt.Printf("  %s\n", e)
			}
			os.Exit(1)
		}

		fmt.Println("Validation successful! ✓")
	},
}

func init() {
	validateCmd.MarkFlagRequired("schema")
	rootCmd.AddCommand(validateCmd)
}

