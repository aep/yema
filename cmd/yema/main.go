package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aep/yema/cue"
	"github.com/aep/yema/golang"
	"github.com/aep/yema/jsonschema"
	"github.com/aep/yema/parser"
	"github.com/aep/yema/rust"
	"github.com/aep/yema/typescript"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
)

var (
	outputFormat     string
	codePackage      string
	codeModuleName   string
	codeTypeName     string
	tsNamespace      string
	tsUseInterfaces  bool
	tsExportAll      bool
	rustDeriveTraits string
	rustUseRename    bool
)

var rootCmd = &cobra.Command{
	Use:   "yema",
	Short: "Yema schema processing tool",
	Long: `Yema is a tool for working with schema definitions.
It can convert Yema schemas to various formats and validate data against schemas.`,
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
			value, err := cue.ToCue(cuecontext.New(), yy)
			if err != nil {
				log.Fatalf("Error parsing schema: %v", err)
			}

			node := value.Syntax()
			bytes, err := format.Node(node)
			if err != nil {
				log.Fatalf("Error formatting CUE: %v", err)
			}

			fmt.Println(string(bytes))
		case "jsonschema":
			jsonBytes, err := jsonschema.ToJSONSchema(yy)
			if err != nil {
				log.Fatalf("Error generating JSON Schema: %v", err)
			}
			fmt.Println(string(jsonBytes))
		case "golang":
			goBytes, err := golang.ToGolangWithOptions(yy, golang.Options{
				Package:  codePackage,
				RootType: codeTypeName,
			})
			if err != nil {
				log.Fatalf("Error generating Go structs: %v", err)
			}
			fmt.Println(string(goBytes))
		case "typescript":
			tsBytes, err := typescript.ToTypeScriptWithOptions(yy, typescript.Options{
				Namespace:     tsNamespace,
				RootType:      codeTypeName,
				UseInterfaces: tsUseInterfaces,
				ExportAll:     tsExportAll,
			})
			if err != nil {
				log.Fatalf("Error generating TypeScript definitions: %v", err)
			}
			fmt.Println(string(tsBytes))
		case "rust":
			// Parse the derive traits string into a slice
			var deriveTraits []string
			if rustDeriveTraits != "" {
				deriveTraits = strings.Split(rustDeriveTraits, ",")
				for i := range deriveTraits {
					deriveTraits[i] = strings.TrimSpace(deriveTraits[i])
				}
			}

			rustBytes, err := rust.ToRustWithOptions(yy, rust.Options{
				Module:         codeModuleName,
				RootType:       codeTypeName,
				DeriveTraits:   deriveTraits,
				UseSerdeRename: rustUseRename,
			})
			if err != nil {
				log.Fatalf("Error generating Rust structs: %v", err)
			}
			fmt.Println(string(rustBytes))
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
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "cue", "Output format (cue, jsonschema, golang, typescript, rust)")
	rootCmd.PersistentFlags().StringVar(&codePackage, "package", "generated", "Package name for generated code (golang)")
	rootCmd.PersistentFlags().StringVar(&codeModuleName, "module", "generated", "Module name for generated code (rust)")
	rootCmd.PersistentFlags().StringVar(&codeTypeName, "type", "Type", "Root type name for generated code")
	rootCmd.PersistentFlags().StringVar(&tsNamespace, "namespace", "", "Namespace for TypeScript code (typescript)")
	rootCmd.PersistentFlags().BoolVar(&tsUseInterfaces, "interfaces", true, "Use interfaces instead of type aliases (typescript)")
	rootCmd.PersistentFlags().BoolVar(&tsExportAll, "export-all", true, "Export all TypeScript types (typescript)")
	rootCmd.PersistentFlags().StringVar(&rustDeriveTraits, "derive", "Debug,Clone,Serialize,Deserialize", "Comma-separated list of traits to derive (rust)")
	rootCmd.PersistentFlags().BoolVar(&rustUseRename, "serde-rename", true, "Use serde rename attributes for JSON field names (rust)")
}

