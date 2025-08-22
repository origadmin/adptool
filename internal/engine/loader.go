package engine

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/loader"
	adpparser "github.com/origadmin/adptool/internal/parser"
)

// Loader loads source files and configurations.
type Loader struct {
	fs       fs.FS
	parser   Parser
	config   *config.Config
	logger   *slog.Logger
}

// Parser parses Go source files.
type Parser interface {
	ParseFile(filePath string) (*ast.File, *token.FileSet, error)
}

// NewLoader creates a new Loader.
func NewLoader(fsys fs.FS, parser Parser, cfg *config.Config, logger *slog.Logger) *Loader {
	// Ensure we have a logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	
	return &Loader{
		fs:     fsys,
		parser: parser,
		config: cfg,
		logger: logger,
	}
}

// Load loads the source files and configurations.
func (l *Loader) Load(ctx context.Context, paths []string) (*LoadContext, error) {
	l.logger.Info("Loading files", "paths", paths)

	loadCtx := &LoadContext{
		Files:    make(map[string]*ast.File),
		FileSets: make(map[string]*token.FileSet),
		Config:   l.config,
	}

	for _, path := range paths {
		// When using a mock filesystem, we need to handle "." differently
		if path == "." {
			// Walk all files in the filesystem
			err := fs.WalkDir(l.fs, ".", func(filePath string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Skip directories and non-Go files
				if d.IsDir() || filepath.Ext(filePath) != ".go" {
					return nil
				}

				// Check if file contains //go:adapter directive
				hasAdapter, err := l.hasAdapterDirective(filePath)
				if err != nil {
					l.logger.Warn("Failed to check adapter directive", "file", filePath, "error", err)
					return nil
				}

				if !hasAdapter {
					return nil
				}

				// Parse the Go file
				file, fset, err := l.parser.ParseFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to parse file %s: %w", filePath, err)
				}

				loadCtx.Files[filePath] = file
				loadCtx.FileSets[filePath] = fset

				// Parse file directives
				_, err = adpparser.ParseFileDirectives(loadCtx.Config, file, fset)
				if err != nil {
					return fmt.Errorf("failed to parse directives in %s: %w", filePath, err)
				}

				l.logger.Info("Loaded file", "path", filePath)
				return nil
			})
			
			if err != nil {
				return nil, fmt.Errorf("error walking path %s: %w", path, err)
			}
		} else {
			// Handle specific file paths
			err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Skip directories and non-Go files
				if d.IsDir() || filepath.Ext(filePath) != ".go" {
					return nil
				}

				// Check if file contains //go:adapter directive
				hasAdapter, err := hasAdapterDirective(filePath)
				if err != nil {
					l.logger.Warn("Failed to check adapter directive", "file", filePath, "error", err)
					return nil
				}

				if !hasAdapter {
					return nil
				}

				// Parse the Go file
				file, fset, err := l.parser.ParseFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to parse file %s: %w", filePath, err)
				}

				loadCtx.Files[filePath] = file
				loadCtx.FileSets[filePath] = fset

				// Parse file directives
				_, err = adpparser.ParseFileDirectives(loadCtx.Config, file, fset)
				if err != nil {
					return fmt.Errorf("failed to parse directives in %s: %w", filePath, err)
				}

				l.logger.Info("Loaded file", "path", filePath)
				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("error walking path %s: %w", path, err)
			}
		}
	}

	return loadCtx, nil
}

// LoadConfig loads configuration from a file.
func (l *Loader) LoadConfig(path string) (*config.Config, error) {
	return loader.LoadConfigFile(path)
}

// hasAdapterDirective checks if a file contains the //go:adapter directive.
func (l *Loader) hasAdapterDirective(filePath string) (bool, error) {
	// When using a mock filesystem, we need to read from the fs.FS
	file, err := l.fs.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "//go:adapter") {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// hasAdapterDirective checks if a file contains the //go:adapter directive.
func hasAdapterDirective(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "//go:adapter") {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// FileSystemParser implements the Parser interface using the file system.
type FileSystemParser struct{}

// NewFileSystemParser creates a new FileSystemParser.
func NewFileSystemParser() *FileSystemParser {
	return &FileSystemParser{}
}

// ParseFile parses a Go source file.
func (p *FileSystemParser) ParseFile(filePath string) (*ast.File, *token.FileSet, error) {
	return loader.LoadGoFile(filePath)
}