package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/generator"
)

func main() {
	outputFile := flag.String("o", "", "Output file path for generated adapters. Can be a file or a directory.")
	configFile := flag.String("f", "", "File-level configuration file (YAML/JSON). If specified, it completely replaces adptool.yaml.")

	flag.Parse()

	inputPaths := flag.Args()
	if len(inputPaths) == 0 {
		fmt.Println("Error: No input files or directories specified.")
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Collect all Go files to process
	var filesToProcess []string
	for _, inputPath := range inputPaths {
		info, err := os.Stat(inputPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if info.IsDir() {
			// Recursively find .go files in the directory
			filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
					filesToProcess = append(filesToProcess, path)
				}
				return nil
			})
		} else {
			filesToProcess = append(filesToProcess, inputPath)
		}
	}

	if len(filesToProcess) == 0 {
		fmt.Println("Error: No Go files found to process.")
		os.Exit(1)
	}

	// Determine output file path
	finalOutputFile := *outputFile
	if finalOutputFile == "" {
		if len(filesToProcess) == 1 {
			// Default output for single input file
			ext := filepath.Ext(filesToProcess[0])
			base := strings.TrimSuffix(filesToProcess[0], ext)
			finalOutputFile = base + ".adapter.go"
		} else {
			fmt.Println("Error: For multiple input files/directories, -o must specify an output directory.")
			flag.Usage()
			os.Exit(1)
		}
	} else {
		// If output is specified, check if it's a directory for multiple inputs
		if len(filesToProcess) > 1 {
			outputInfo, err := os.Stat(finalOutputFile)
			if err != nil || !outputInfo.IsDir() {
				fmt.Println("Error: For multiple input files/directories, -o must specify an existing output directory.")
				os.Exit(1)
			}
		}
	}

	// Call the generator
	if err := generator.Generate(cfg, filesToProcess, finalOutputFile); err != nil {
		fmt.Printf("Error generating adapters: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated adapters to %s\n", finalOutputFile)
}