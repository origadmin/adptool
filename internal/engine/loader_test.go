package engine

import (
	"context"
	"go/ast"
	"go/token"
	"io"
	"log/slog"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/origadmin/adptool/internal/config"
)

type mockParser struct {
	file *ast.File
	fset *token.FileSet
}

func (m *mockParser) ParseFile(filePath string) (*ast.File, *token.FileSet, error) {
	return m.file, m.fset, nil
}

func TestLoader_New(t *testing.T) {
	fsys := fstest.MapFS{}
	parser := &mockParser{}
	cfg := config.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // Use Discard to avoid output

	loader := NewLoader(fsys, parser, cfg, logger)
	if loader == nil {
		t.Error("Expected loader to be created, got nil")
	}
}

func TestLoader_Load(t *testing.T) {
	// Create a mock file system with a Go file containing //go:adapter directive
	fsys := fstest.MapFS{
		"test.go": &fstest.MapFile{
			Data: []byte(`//go:adapter type:MyType prefix:Adapted
package main

type MyType struct {
	Name string
}`),
		},
		"regular.go": &fstest.MapFile{
			Data: []byte(`package main

func main() {
	println("Hello, World!")
}`),
		},
	}

	parser := &mockParser{
		file: &ast.File{
			Name: ast.NewIdent("main"),
		},
		fset: token.NewFileSet(),
	}

	cfg := config.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // Use Discard to avoid output

	loader := NewLoader(fsys, parser, cfg, logger)
	ctx := context.Background()

	loadCtx, err := loader.Load(ctx, []string{"."}) // Load all files in the mock filesystem
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	if loadCtx == nil {
		t.Error("Expected LoadContext to be returned, got nil")
	}

	// The test is checking that the loader correctly identifies files with //go:adapter directive
	// In the current implementation, it checks all .go files in the directory
	// Since we're using a mock filesystem, all files are checked
	// But only those with //go:adapter directive should be added to loadCtx.Files
	if len(loadCtx.Files) != 1 {
		t.Errorf("Expected 1 file with //go:adapter directive, got %d", len(loadCtx.Files))
	}
	
	// Check that we got the right file
	if _, exists := loadCtx.Files["test.go"]; !exists {
		t.Error("Expected test.go to be loaded (it has //go:adapter directive)")
	}
}

func TestLoader_LoadConfig(t *testing.T) {
	fsys := fstest.MapFS{}
	parser := &mockParser{}
	cfg := config.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // Use Discard to avoid output

	loader := NewLoader(fsys, parser, cfg, logger)

	// Try to load a non-existent config file - this should return an error
	_, err := loader.LoadConfig("non-existent.yaml")
	// This should return an error since the file doesn't exist
	if err == nil {
		t.Error("Expected LoadConfig to return error for non-existent file")
	}
	
	// Check that the error message contains expected text
	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("Expected error to contain 'failed to read config file', got: %v", err)
	}
}