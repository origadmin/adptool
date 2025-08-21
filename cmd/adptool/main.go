package main

import (
	"flag"
	"log/slog"
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
		slog.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	// Compile the configuration into a compiled config and replacer
	compiledCfg, err := compiler.Compile(cfg) // Call the compiler
	if err != nil {
		slog.Error("Error compiling config", "error", err)
		os.Exit(1)
	}
	
	// Create replacer from compiled config
	replacer := compiler.NewReplacer(compiledCfg)
	
	// Convert CompiledPackage to PackageInfo
	var packageInfos []*generator.PackageInfo
	for _, pkg := range compiledCfg.Packages {
		packageInfos = append(packageInfos, &generator.PackageInfo{
			ImportPath:  pkg.ImportPath,
			ImportAlias: pkg.ImportAlias,
		})
	}

	// Determine output file path
	outputAliasFilePath := "tools/adptool/generated_alias/aliaspkg.go"

	// Initialize and call the generator
	gen := generator.NewGenerator(compiledCfg.PackageName, outputAliasFilePath, replacer)
	if err := gen.Generate(packageInfos); err != nil {
		slog.Error("Error generating alias package", "error", err)
		os.Exit(1)
	}

	slog.Info("Successfully generated alias package", "output", outputAliasFilePath)
}