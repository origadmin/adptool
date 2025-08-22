package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/engine"
	"github.com/origadmin/adptool/internal/loader"
	"github.com/origadmin/adptool/internal/parser"
)

// processFile now acts as a simple wrapper around the core engine.
func processFile(filePath string, cfg *config.Config) error {
	e := engine.New()
	return e.Execute(filePath, cfg)
}

// findGoFiles finds all .go files in the given directory that contain //go:adapter directive
func findGoFiles(dir string) ([]string, error) {
	// Handle current directory (.) case
	if dir == "." {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, test files, and non-Go files
		if d.IsDir() ||
			strings.HasSuffix(d.Name(), "_test.go") ||
			!strings.HasSuffix(d.Name(), ".go") ||
			strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Check if file contains //go:adapter directive
		hasAdapter, err := parser.HasAdapterDirective(path)
		if err != nil {
			slog.Warn("Failed to check adapter directive", "file", path, "error", err)
			return nil
		}

		if hasAdapter {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", dir, err)
	}

	return files, nil
}

func main() {
	configFile := flag.String("c", "", "Configuration file (YAML/JSON). If specified, it completely replaces adptool.yaml.")
	flag.Parse()

	// Get the input path from command line arguments
	args := flag.Args()
	if len(args) == 0 {
		slog.Error("No input path specified")
		os.Exit(1)
	}

	inputPath := args[0]

	// Get absolute path to the input path
	abspath, err := filepath.Abs(inputPath)
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
		// Use the loaded config
		cfg = fileCfg
	}

	// Check if the input is a directory or a file
	fileInfo, err := os.Stat(abspath)
	if err != nil {
		slog.Error("Failed to get file info", "path", abspath, "error", err)
		os.Exit(1)
	}

	var filesToProcess []string

	if fileInfo.IsDir() {
		// If it's a directory, find all .go files
		files, err := findGoFiles(abspath)
		if err != nil {
			slog.Error("Failed to find Go files in directory", "directory", abspath, "error", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			slog.Info("No Go files found in directory", "directory", abspath)
			return
		}

		filesToProcess = files
	} else {
		// If it's a single file, just process that file
		if !strings.HasSuffix(abspath, ".go") {
			slog.Error("Input file is not a Go file", "file", abspath)
			os.Exit(1)
		}
		filesToProcess = []string{abspath}
	}

	// Process each file
	var hasErrors bool
	for _, file := range filesToProcess {
		if err := processFile(file, cfg); err != nil {
			slog.Error("Error processing file", "file", file, "error", err)
			hasErrors = true
		}
	}

	if hasErrors {
		slog.Error("Failed to process some files")
		os.Exit(1)
	}
}