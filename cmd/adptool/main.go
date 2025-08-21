package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/origadmin/adptool/internal/compiler"
	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/generator"
	"github.com/origadmin/adptool/internal/loader"
	adpparser "github.com/origadmin/adptool/internal/parser"
)

func main() {
	configFile := flag.String("f", "", "Configuration file (YAML/JSON). If specified, it completely replaces adptool.yaml.")
	flag.Parse()

	// Get the input Go file from command line arguments
	args := flag.Args()
	if len(args) == 0 {
		slog.Error("No input Go file specified")
		os.Exit(1)
	}

	inputFile := args[0]

	// Get absolute path to the input file
	abspath, err := filepath.Abs(inputFile)
	if err != nil {
		slog.Error("Failed to get absolute path", "error", err)
		os.Exit(1)
	}

	// Initialize config with defaults
	cfg := config.New()
	
	// Load config from file if provided
	if *configFile != "" {
		fileCfg, err := loader.LoadConfigFile(*configFile)
		if err != nil {
			slog.Error("Failed to load config file", "file", *configFile, "error", err)
			os.Exit(1)
		}
		// Merge the loaded config with defaults
		cfg = fileCfg
	}

	// Parse the Go file to get the AST
	file, fset, err := loader.LoadGoFile(abspath)
	if err != nil {
		slog.Error("Failed to load Go file", "file", abspath, "error", err)
		os.Exit(1)
	}
	// Parse file directives using the loaded config
	pkgConfig, err := adpparser.ParseFileDirectives(cfg, file, fset)
	if err != nil {
		slog.Error("Failed to parse file directives", "file", abspath, "error", err)
		os.Exit(1)
	}

	// Compile the configuration into a compiled config and replacer
	compiledCfg, err := compiler.Compile(pkgConfig)
	if err != nil {
		slog.Error("Error compiling config", "error", err)
		os.Exit(1)
	}

	replacer := compiler.NewReplacer(compiledCfg)

	// Set output file path (same directory as input file with .adapter.go suffix)
	dir := filepath.Dir(abspath)
	outputFile := filepath.Join(dir, filepath.Base(dir)+".adapter.go")

	// Convert PackageConfig to PackageInfo
	var packageInfos []*generator.PackageInfo
	for _, pkg := range pkgConfig.Packages {
		packageInfos = append(packageInfos, &generator.PackageInfo{
			ImportPath:  pkg.Import,
			ImportAlias: pkg.Alias,
		})
	}

	// Initialize and call the generator
	gen := generator.NewGenerator(pkgConfig.PackageName, outputFile, replacer).
		WithNoEditHeader(true)

	if err := gen.Generate(packageInfos); err != nil {
		slog.Error("Error generating adapter file", "error", err)
		os.Exit(1)
	}

	slog.Info("Successfully generated adapter file", "path", outputFile)
}
