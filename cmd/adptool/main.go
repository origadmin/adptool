package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/engine"
	"github.com/origadmin/adptool/internal/loader"
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

	// Create and use the new engine
	e := engine.New()
	if err := e.ExecuteFile(abspath, cfg); err != nil {
		slog.Error("Failed to execute engine", "error", err)
		os.Exit(1)
	}

	slog.Info("Successfully generated adapter file", "path", abspath)
}
