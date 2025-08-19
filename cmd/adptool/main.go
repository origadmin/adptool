package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/origadmin/adptool/internal/compiler" // Updated import
	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/generator"
)

func main() {
	configFile := flag.String("f", "", "File-level configuration file (YAML/JSON). If specified, it completely replaces adptool.yaml.")

	flag.Parse()

	// Load config
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Compile the configuration into a replacer
	replacer := compiler.Compile(cfg) // Call the compiler

	// Determine source package import path
	sourcePackageImportPath := "github.com/origadmin/adptool/alias_source/sourcepkg"
	aliasPackageName := "aliaspkg"
	outputAliasFilePath := "tools/adptool/generated_alias/aliaspkg.go"

	// Call the generator
	if err := generator.Generate(replacer, sourcePackageImportPath, aliasPackageName, outputAliasFilePath); err != nil {
		fmt.Printf("Error generating alias package: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated alias package to %s\n", outputAliasFilePath)
}