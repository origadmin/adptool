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
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/origadmin/adptool/internal/interfaces"
)

// packageDecls holds declarations for a single package
type packageDecls struct {
	typeSpecs  []ast.Spec
	varSpecs   []ast.Spec
	constSpecs []ast.Spec
	funcDecls  []ast.Decl
}

// Generator holds the state and configuration for code generation.
type Generator struct {
	packageName     string
	outputFilePath  string
	fset            *token.FileSet
	aliasFile       *ast.File
	importSpecs     map[string]*ast.ImportSpec
	allPackageDecls map[string]*packageDecls
	definedTypes    map[string]bool
	replacer        interfaces.Replacer
}

// NewGenerator creates a new Generator instance.
func NewGenerator(packageName string, outputFilePath string, replacer interfaces.Replacer) *Generator {
	return &Generator{
		packageName:    packageName,
		outputFilePath: outputFilePath,
		fset:           token.NewFileSet(),
		aliasFile: &ast.File{
			Name:  ast.NewIdent(packageName),
			Decls: []ast.Decl{},
		},
		importSpecs:     make(map[string]*ast.ImportSpec),
		allPackageDecls: make(map[string]*packageDecls),
		definedTypes:    make(map[string]bool),
		replacer:        replacer,
	}
}

// Generate generates the output code.
func (g *Generator) Generate(packages []*PackageInfo) error {
	// Initialize generator state
	g.initializeState()

	// Process each source package
	if err := g.processPackages(packages); err != nil {
		return err
	}

	// Apply replacements using the Replacer to all collected declarations
	if g.replacer != nil {
		g.applyReplacements()
	}

	// Build the output file structure
	g.buildOutputFile()

	// Write the output file
	return g.writeOutputFile()
}

// initializeState initializes the generator's state.
func (g *Generator) initializeState() {
	g.allPackageDecls = make(map[string]*packageDecls)
	g.definedTypes = make(map[string]bool)
}

// processPackages processes each source package and collects declarations.
func (g *Generator) processPackages(packages []*PackageInfo) error {
	for _, pkg := range packages {
		// Add the primary package being adapted to the import list
		g.importSpecs[pkg.ImportPath] = &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", pkg.ImportPath)},
			Name: &ast.Ident{Name: pkg.ImportAlias},
		}

		// Load the package
		sourcePkg, err := g.loadPackage(pkg.ImportPath)
		if err != nil {
			return err
		}
		if sourcePkg == nil {
			continue // Skip if package not found
		}

		// Collect imports from the source package
		g.collectImports(sourcePkg)

		// Collect declarations from the source package
		g.collectTypeDeclarations(sourcePkg, pkg.ImportAlias)
		g.collectOtherDeclarations(sourcePkg, pkg.ImportAlias)
	}
	return nil
}

// loadPackage loads a package by its import path.
func (g *Generator) loadPackage(importPath string) (*packages.Package, error) {
	loadCfg := &packages.Config{
		Mode: packages.LoadSyntax | packages.LoadTypes,
	}
	pkgs, err := packages.Load(loadCfg, importPath)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, nil // Package not found
	}
	return pkgs[0], nil
}

// collectImports collects imports from the source package.
func (g *Generator) collectImports(sourcePkg *packages.Package) {
	for importPath := range sourcePkg.Imports {
		if _, exists := g.importSpecs[importPath]; !exists {
			g.importSpecs[importPath] = &ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", importPath)},
				Name: nil, // Default to no alias
			}
		}
	}
}

// collectTypeDeclarations collects type declarations from the source package.
func (g *Generator) collectTypeDeclarations(sourcePkg *packages.Package, importAlias string) {
	for _, file := range sourcePkg.Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.IsExported() {
						g.collectTypeDeclaration(typeSpec, importAlias)
					}
				}
			}
		}
	}
}

// collectTypeDeclaration collects a single type declaration.
func (g *Generator) collectTypeDeclaration(typeSpec *ast.TypeSpec, importAlias string) {
	if typeSpec.Name.IsExported() {
		// Extract original name by removing prefixes
		originalName := g.extractOriginalName(typeSpec.Name.Name)

		// Create new type specification as an alias
		newSpec := &ast.TypeSpec{
			Name: typeSpec.Name,
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent(importAlias),
				Sel: ast.NewIdent(originalName),
			},
			Assign: 1, // Make it an alias with '='
		}

		// Add to package declarations
		if g.allPackageDecls[importAlias] == nil {
			g.allPackageDecls[importAlias] = &packageDecls{}
		}
		g.allPackageDecls[importAlias].typeSpecs = append(g.allPackageDecls[importAlias].typeSpecs, newSpec)

		// Add to definedTypes map using the original name
		g.definedTypes[originalName] = true

		log.Printf("collectTypeDeclaration: Added %s (original: %s) to definedTypes",
			typeSpec.Name.Name, originalName)
	}
}

// collectOtherDeclarations collects function, variable, and constant declarations.
func (g *Generator) collectOtherDeclarations(sourcePkg *packages.Package, importAlias string) {
	for _, file := range sourcePkg.Syntax {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				g.collectFunctionDeclaration(d, importAlias)
			case *ast.GenDecl:
				switch d.Tok {
				case token.CONST:
					g.collectValueDeclaration(d, importAlias, token.CONST)
				case token.VAR:
					g.collectValueDeclaration(d, importAlias, token.VAR)
				}
			}
		}
	}
}

// collectFunctionDeclaration collects a function declaration.
func (g *Generator) collectFunctionDeclaration(funcDecl *ast.FuncDecl, importAlias string) {
	// Only process exported functions without receivers (not methods)
	if funcDecl.Recv == nil && funcDecl.Name.IsExported() {
		// Extract original name by removing prefixes
		originalName := g.extractOriginalName(funcDecl.Name.Name)

		// Create argument list for the function call
		var args []ast.Expr
		if funcDecl.Type.Params != nil {
			for _, param := range funcDecl.Type.Params.List {
				for _, name := range param.Names {
					args = append(args, name)
				}
			}
		}

		// Create function call expression
		callExpr := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(importAlias),
				Sel: ast.NewIdent(originalName),
			},
			Args: args,
		}

		// Create function body with appropriate return statement
		var results []ast.Stmt
		if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
			results = []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{callExpr}}}
		} else {
			results = []ast.Stmt{&ast.ExprStmt{X: callExpr}}
		}

		// Create function declaration with qualified types
		newFuncDecl := &ast.FuncDecl{
			Name: funcDecl.Name,
			Type: qualifyType(funcDecl.Type, importAlias, g.definedTypes).(*ast.FuncType),
			Body: &ast.BlockStmt{List: results},
		}

		// Add to package declarations
		if g.allPackageDecls[importAlias] == nil {
			g.allPackageDecls[importAlias] = &packageDecls{}
		}
		g.allPackageDecls[importAlias].funcDecls = append(g.allPackageDecls[importAlias].funcDecls, newFuncDecl)
	}
}

// collectValueDeclaration collects variable and constant declarations.
func (g *Generator) collectValueDeclaration(genDecl *ast.GenDecl, importAlias string, tok token.Token) {
	for _, spec := range genDecl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			for _, name := range valueSpec.Names {
				if name.IsExported() {
					// Extract original name by removing prefixes
					originalName := g.extractOriginalName(name.Name)

					// Apply replacer to get the correct name if available
					finalName := name
					if g.replacer != nil {
						if renamed, ok := g.replacer.Apply(name).(*ast.Ident); ok {
							finalName = renamed
						}
					}

					// Create new value specification
					newSpec := &ast.ValueSpec{
						Names: []*ast.Ident{finalName},
						Values: []ast.Expr{
							&ast.SelectorExpr{
								X:   ast.NewIdent(importAlias),
								Sel: ast.NewIdent(originalName),
							},
						},
					}

					// Add to package declarations based on token type
					if g.allPackageDecls[importAlias] == nil {
						g.allPackageDecls[importAlias] = &packageDecls{}
					}

					if tok == token.VAR {
						g.allPackageDecls[importAlias].varSpecs = append(g.allPackageDecls[importAlias].varSpecs, newSpec)
					} else if tok == token.CONST {
						g.allPackageDecls[importAlias].constSpecs = append(g.allPackageDecls[importAlias].constSpecs, newSpec)
					}
				}
			}
		}
	}
}

// applyReplacements applies the replacer to all collected declarations.
func (g *Generator) applyReplacements() {
	// Apply replacer to all collected declarations across all packages
	for alias, pkgDecls := range g.allPackageDecls {
		// Apply to type declarations
		for i, spec := range pkgDecls.typeSpecs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				replaced := g.replacer.Apply(typeSpec)
				if replacedSpec, ok := replaced.(*ast.TypeSpec); ok {
					pkgDecls.typeSpecs[i] = replacedSpec

					// Update definedTypes with the new type name
					originalName := g.extractOriginalName(typeSpec.Name.Name)
					newName := g.extractOriginalName(replacedSpec.Name.Name)

					// Keep the type in definedTypes with its original name
					if _, exists := g.definedTypes[originalName]; !exists {
						g.definedTypes[originalName] = true
					}
					log.Printf("applyReplacements: Applied replacer to type %s (original: %s, new: %s) in package %s",
						typeSpec.Name.Name, originalName, newName, alias)
				}
			}
		}

		// Apply to const declarations
		for i, spec := range pkgDecls.constSpecs {
			replaced := g.replacer.Apply(spec)
			if replacedSpec, ok := replaced.(*ast.ValueSpec); ok {
				pkgDecls.constSpecs[i] = replacedSpec
			}
		}

		// Apply to var declarations
		for i, spec := range pkgDecls.varSpecs {
			replaced := g.replacer.Apply(spec)
			if replacedSpec, ok := replaced.(*ast.ValueSpec); ok {
				pkgDecls.varSpecs[i] = replacedSpec
			}
		}

		// Apply to function declarations
		for i, decl := range pkgDecls.funcDecls {
			replaced := g.replacer.Apply(decl)
			if replacedDecl, ok := replaced.(*ast.FuncDecl); ok {
				// Make sure function types are qualified properly
				replacedDecl.Type = qualifyType(replacedDecl.Type, alias, g.definedTypes).(*ast.FuncType)
				pkgDecls.funcDecls[i] = replacedDecl
			}
		}
	}
}

// buildOutputFile builds the output file structure.
func (g *Generator) buildOutputFile() {
	var orderedDecls []ast.Decl

	// Add imports first
	importDecl := g.buildImportDeclaration()
	if len(importDecl.(*ast.GenDecl).Specs) > 0 {
		orderedDecls = append(orderedDecls, importDecl)
	}

	// Collect all declarations by type across all packages
	allConstSpecs, allVarSpecs, allTypeSpecs, allFuncDecls := g.collectAllDeclarations()

	// Add all const declarations in a single group with parentheses
	if len(allConstSpecs) > 0 {
		constDecl := &ast.GenDecl{
			Tok:    token.CONST,
			Lparen: token.Pos(1), // Set non-zero position to include parentheses
			Specs:  allConstSpecs,
		}
		orderedDecls = append(orderedDecls, constDecl)
	}

	// Add all var declarations in a single group with parentheses
	if len(allVarSpecs) > 0 {
		varDecl := &ast.GenDecl{
			Tok:    token.VAR,
			Lparen: token.Pos(1), // Set non-zero position to include parentheses
			Specs:  allVarSpecs,
		}
		orderedDecls = append(orderedDecls, varDecl)
	}

	// Add all type declarations in a single group with parentheses
	if len(allTypeSpecs) > 0 {
		typeDecl := &ast.GenDecl{
			Tok:    token.TYPE,
			Lparen: token.Pos(1), // Set non-zero position to include parentheses
			Specs:  allTypeSpecs,
		}
		orderedDecls = append(orderedDecls, typeDecl)
	}

	// Add all function declarations
	orderedDecls = append(orderedDecls, allFuncDecls...)

	g.aliasFile.Decls = orderedDecls
}

// buildImportDeclaration builds the import declaration.
func (g *Generator) buildImportDeclaration() ast.Decl {
	var finalImportSpecs []ast.Spec
	for _, spec := range g.importSpecs {
		finalImportSpecs = append(finalImportSpecs, spec)
	}

	// Sort imports for consistent output
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

// collectAllDeclarations collects all declarations from all packages.
func (g *Generator) collectAllDeclarations() ([]ast.Spec, []ast.Spec, []ast.Spec, []ast.Decl) {
	log.Printf("collectAllDeclarations: Current definedTypes: %v", g.definedTypes)

	var allConstSpecs []ast.Spec
	var allVarSpecs []ast.Spec
	var allTypeSpecs []ast.Spec
	var allFuncDecls []ast.Decl

	// Sort package aliases for consistent output
	var sortedPackageAliases []string
	for alias := range g.allPackageDecls {
		sortedPackageAliases = append(sortedPackageAliases, alias)
	}
	sort.Strings(sortedPackageAliases)

	for _, alias := range sortedPackageAliases {
		pkgDecls := g.allPackageDecls[alias]

		// Collect all declarations
		allConstSpecs = append(allConstSpecs, pkgDecls.constSpecs...)
		allVarSpecs = append(allVarSpecs, pkgDecls.varSpecs...)
		allTypeSpecs = append(allTypeSpecs, pkgDecls.typeSpecs...)

		// Add function declarations
		for _, funcDecl := range pkgDecls.funcDecls {
			allFuncDecls = append(allFuncDecls, funcDecl)
		}
	}

	return allConstSpecs, allVarSpecs, allTypeSpecs, allFuncDecls
}

// writeOutputFile writes the output file.
func (g *Generator) writeOutputFile() error {
	outFile, err := os.Create(g.outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Ensure the directory exists
	outputDir := filepath.Dir(g.outputFilePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the file content
	if err := printer.Fprint(outFile, g.fset, g.aliasFile); err != nil {
		return fmt.Errorf("failed to write generated code: %w", err)
	}

	return nil
}

// extractOriginalName removes prefixes like Const, Type, Var, Func from a name.
func (g *Generator) extractOriginalName(name string) string {
	originalName := strings.TrimPrefix(name, "Const")
	originalName = strings.TrimPrefix(originalName, "Type")
	originalName = strings.TrimPrefix(originalName, "Var")
	originalName = strings.TrimPrefix(originalName, "Func")
	return originalName
}

// ApplyRules applies a set of rename rules to a given name and returns the result.
// This is a wrapper around rules.ApplyRules for backward compatibility.
func ApplyRules(name string, rulesList []interfaces.RenameRule) (string, error) {
	// Placeholder - actual implementation would depend on the rules package
	return name, nil
}
