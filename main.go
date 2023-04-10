package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Read generator files from the "generators" folder
	generatorFiles, err := ioutil.ReadDir("generators")
	if err != nil {
		log.Fatal(err)
	}

	// Process each generator file
	for _, file := range generatorFiles {
		if filepath.Ext(file.Name()) != ".go" {
			continue
		}

		// Parse the file
		fset := token.NewFileSet()
		parsedFile, err := parser.ParseFile(fset, filepath.Join("generators", file.Name()), nil, 0)
		if err != nil {
			log.Fatal(err)
		}

		// Extract the package name
		packageName := parsedFile.Name.Name

		os.MkdirAll("./generators/mocks", 0755)

		// Iterate through declarations and find structs
		for _, decl := range parsedFile.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// Generate mock functions for the struct
				mockFunctions := generateMockFunctions(packageName, typeSpec.Name.Name, structType.Fields.List)

				// Format the generated code
				formattedCode, err := format.Source([]byte(mockFunctions))
				if err != nil {
					log.Fatal(err)
				}

				// Write the generated functions to a new file
				outputFile := filepath.Join("generators", "mocks", strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))+"_mock.go")
				err = ioutil.WriteFile(outputFile, formattedCode, 0644)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func generateMockFunctions(packageName, structName string, fields []*ast.Field) string {
	var builder strings.Builder

	// Write package declaration
	builder.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Write import statement for the original user types
	builder.WriteString(fmt.Sprintf("import \"%s\"\n\n", packageName))

	// Generate Set and Get methods for each field
	for _, field := range fields {
		fieldName := field.Names[0].Name
		fieldType := fmt.Sprintf("%s", field.Type)

		// Generate Set method
		builder.WriteString(fmt.Sprintf("func (s *%s.%s) Set%s(value %s) {\n\ts.%s = value\n}\n\n", packageName, structName, fieldName, fieldType, fieldName))

		// Generate Get method
		builder.WriteString(fmt.Sprintf("func (s *%s.%s) Get%s() %s {\n\treturn s.%s\n}\n\n", packageName, structName, fieldName, fieldType, fieldName))
	}
	// Generate Query and Execute methods
	builder.WriteString(fmt.Sprintf("func (s *%s) Query() string {\n\t// Implement your query generation logic here\n\treturn \"\"\n}\n\n", structName))
	builder.WriteString(fmt.Sprintf("func (s *%s) Execute() error {\n\t// Implement your query execution logic here\n\treturn nil\n}\n", structName))

	return builder.String()
}
