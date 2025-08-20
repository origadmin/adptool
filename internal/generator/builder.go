package generator

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// Builder is responsible for building the output file from collected declarations.
type Builder struct {
	fset           *token.FileSet
	outputFilePath string
	aliasFile      *ast.File
}

// NewBuilder creates a new Builder.
func NewBuilder(packageName string, outputFilePath string) *Builder {
	return &Builder{
		fset:           token.NewFileSet(),
		outputFilePath: outputFilePath,
		aliasFile: &ast.File{
			Name:  ast.NewIdent(packageName),
			Decls: []ast.Decl{},
		},
	}
}

// Build builds the output file structure from the collected data.
func (b *Builder) Build(importSpecs map[string]*ast.ImportSpec, allPackageDecls map[string]*packageDecls, definedTypes map[string]bool) {
	var orderedDecls []ast.Decl

	importDecl := b.buildImportDeclaration(importSpecs)
	if len(importDecl.(*ast.GenDecl).Specs) > 0 {
		orderedDecls = append(orderedDecls, importDecl)
	}

	allConstSpecs, allVarSpecs, allTypeSpecs, allFuncDecls := b.collectAllDeclarations(allPackageDecls, definedTypes)

	if len(allConstSpecs) > 0 {
		constDecl := &ast.GenDecl{
			Tok:    token.CONST,
			Lparen: token.Pos(1),
			Specs:  allConstSpecs,
		}
		orderedDecls = append(orderedDecls, constDecl)
	}

	if len(allVarSpecs) > 0 {
		varDecl := &ast.GenDecl{
			Tok:    token.VAR,
			Lparen: token.Pos(1),
			Specs:  allVarSpecs,
		}
		orderedDecls = append(orderedDecls, varDecl)
	}

	if len(allTypeSpecs) > 0 {
		typeDecl := &ast.GenDecl{
			Tok:    token.TYPE,
			Lparen: token.Pos(1),
			Specs:  allTypeSpecs,
		}
		orderedDecls = append(orderedDecls, typeDecl)
	}

	orderedDecls = append(orderedDecls, allFuncDecls...)

	b.aliasFile.Decls = orderedDecls
}

// Write writes the generated code to the output file.
func (b *Builder) Write() error {
	outputDir := filepath.Dir(b.outputFilePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outFile, err := os.Create(b.outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := printer.Fprint(outFile, b.fset, b.aliasFile); err != nil {
		return fmt.Errorf("failed to write generated code: %w", err)
	}

	return nil
}

func (b *Builder) buildImportDeclaration(importSpecs map[string]*ast.ImportSpec) ast.Decl {
	var finalImportSpecs []ast.Spec
	for _, spec := range importSpecs {
		finalImportSpecs = append(finalImportSpecs, spec)
	}

	sort.Slice(finalImportSpecs, func(i, j int) bool {
		var iPath, jPath string
		if imp, ok := finalImportSpecs[i].(*ast.ImportSpec); ok && imp.Path != nil {
			iPath = imp.Path.Value
		}
		if imp, ok := finalImportSpecs[j].(*ast.ImportSpec); ok && imp.Path != nil {
			jPath = imp.Path.Value
		}
		return iPath < jPath
	})

	return &ast.GenDecl{Tok: token.IMPORT, Specs: finalImportSpecs}
}

func (b *Builder) collectAllDeclarations(allPackageDecls map[string]*packageDecls, definedTypes map[string]bool) ([]ast.Spec, []ast.Spec, []ast.Spec, []ast.Decl) {
	log.Printf("collectAllDeclarations: Current definedTypes: %v", definedTypes)

	var allConstSpecs []ast.Spec
	var allVarSpecs []ast.Spec
	var allTypeSpecs []ast.Spec
	var allFuncDecls []ast.Decl

	var sortedPackageAliases []string
	for alias := range allPackageDecls {
		sortedPackageAliases = append(sortedPackageAliases, alias)
	}
	sort.Strings(sortedPackageAliases)

	for _, alias := range sortedPackageAliases {
		pkgDecls := allPackageDecls[alias]

		// Extract specs from const declarations
		for _, decl := range pkgDecls.constDecls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				allConstSpecs = append(allConstSpecs, genDecl.Specs...)
			}
		}

		// Extract specs from var declarations
		for _, decl := range pkgDecls.varDecls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				allVarSpecs = append(allVarSpecs, genDecl.Specs...)
			}
		}

		allTypeSpecs = append(allTypeSpecs, pkgDecls.typeSpecs...)
		allFuncDecls = append(allFuncDecls, pkgDecls.funcDecls...)
	}

	return allConstSpecs, allVarSpecs, allTypeSpecs, allFuncDecls
}
