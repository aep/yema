package main

import (
	"fmt"
	ycue "github.com/aep/yema/cue"
	"github.com/aep/yema/parser"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"

	"cuelang.org/go/cue/format"
)

var (
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:  "yema",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var input io.Reader = os.Stdin

		if len(args) > 0 {
			file, err := os.Open(args[0])
			if err != nil {
				log.Fatalf("Error opening file: %v", err)
			}
			defer file.Close()
			input = file
		}

		var ys map[string]interface{}
		err := yaml.NewDecoder(input).Decode(&ys)
		if err != nil {
			log.Fatalf("Error parsing YAML: %v", err)
		}

		yy, err := parser.From(ys)
		if err != nil {
			log.Fatalf("Error parsing schema: %v", err)
		}

		switch outputFormat {
		case "cue":
			value, err := ycue.ToCue(yy)
			if err != nil {
				log.Fatalf("Error parsing schema: %v", err)
			}

			node := value.Syntax()
			bytes, err := format.Node(node)
			if err != nil {
				log.Fatalf("Error formatting CUE: %v", err)
			}

			fmt.Println(string(bytes))
		default:
			log.Fatalf("Unsupported output format: %s", outputFormat)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "cue", "Output format (cue, jsonschema)")
}
