package parser

import (
	"errors"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"strings"
)

// parseNameValue parses an argument string into a name and value.
// Expected format: "name value"
func parseNameValue(argument string) (name, value string, err error) {
	parts := strings.SplitN(argument, " ", 2)
	if len(parts) != 2 {
		return "", "", errors.New("argument must be in 'name value' format")
	}
	return parts[0], parts[1], nil
}

// loadGoFile loads a Go file and returns the AST and file set.
func loadGoFile(filePath string) (*goast.File, *gotoken.FileSet, error) {
	fset := gotoken.NewFileSet()
	node, err := goparser.ParseFile(fset, filePath, nil, goparser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Go file %s: %w", filePath, err)
	}
	return node, fset, nil
}
