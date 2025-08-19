package generator

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"sort"

	"golang.org/x/tools/go/packages"

	"github.com/origadmin/adptool/internal/config"
)

// packageDecls struct (already exists, but will be part of Generator's context)
type packageDecls struct {
	typeSpecs  []ast.Spec
	varSpecs   []ast.Spec
	constSpecs []ast.Spec
	funcDecls  []ast.Decl
}

// Generator holds the state and configuration for code generation.
type Generator struct {
	compiledCfg     *config.CompiledConfig
	outputFilePath  string
	fset            *token.FileSet
	aliasFile       *ast.File
	importSpecs     map[string]*ast.ImportSpec
	allPackageDecls map[string]*packageDecls
	definedTypes    map[string]bool
}

// NewGenerator creates a new Generator instance.
func NewGenerator(compiledCfg *config.CompiledConfig, outputFilePath string) *Generator {
	return &Generator{
		compiledCfg:    compiledCfg,
		outputFilePath: outputFilePath,
		fset:           token.NewFileSet(),
		aliasFile: &ast.File{
			Name:  ast.NewIdent(compiledCfg.PackageName),
			Decls: []ast.Decl{},
		},
		importSpecs:     make(map[string]*ast.ImportSpec),
		allPackageDecls: make(map[string]*packageDecls),
		definedTypes:    make(map[string]bool),
	}
}

// Generate generates the output code.
func (g *Generator) Generate() error {
	outFile, err := os.Create(g.outputFilePath)
	fmt.Printf("Attempting to create file: %s, Error: %v\n", g.outputFilePath, err)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Initialize maps
	g.allPackageDecls = make(map[string]*packageDecls)
	g.definedTypes = make(map[string]bool)

	// Process each source package and generate declarations
	for _, pkg := range g.compiledCfg.Packages {
		// Add the primary package being adapted to the import list
		g.importSpecs[pkg.ImportPath] = &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", pkg.ImportPath)},
			Name: ast.NewIdent(pkg.ImportAlias),
		}

		loadCfg := &packages.Config{
			Mode: packages.LoadSyntax | packages.LoadTypes,
		}
		pkgs, err := packages.Load(loadCfg, pkg.ImportPath)
		if err != nil {
			return err
		}
		if len(pkgs) == 0 {
			continue // Skip if package not found
		}
		sourcePkg := pkgs[0]

		// Collect imports from the source package
		for importPath := range sourcePkg.Imports {
			if _, exists := g.importSpecs[importPath]; !exists {
				g.importSpecs[importPath] = &ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", importPath)},
					Name: nil, // Default to no alias
				}
			}
		}

		// First pass: collect all type declarations
		for _, file := range sourcePkg.Syntax {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.IsExported() {
							if g.allPackageDecls[pkg.ImportAlias] == nil {
								g.allPackageDecls[pkg.ImportAlias] = &packageDecls{}
							}
							g.allPackageDecls[pkg.ImportAlias].typeSpecs = append(g.allPackageDecls[pkg.ImportAlias].typeSpecs, &ast.TypeSpec{
								Name: typeSpec.Name,
								Type: &ast.SelectorExpr{
									X:   ast.NewIdent(pkg.ImportAlias),
									Sel: typeSpec.Name,
								},
								Assign: token.Pos(1), // Set non-zero position to make it an alias type (type A = B)
							})
							// Add to definedTypes map
							g.definedTypes[typeSpec.Name.Name] = true
						}
					}
				}
			}
		}

		// Second pass: collect other declarations
		for _, file := range sourcePkg.Syntax {
			fileInfo := g.fset.File(file.Pos())
			if fileInfo != nil {
				// fmt.Printf("Processing file: %s", fileInfo.Name())
			} else {
				// fmt.Printf("Processing file: <unknown> (file.Pos() returned no file info)")
			}
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					// fmt.Printf("  Found FuncDecl: %s, Exported: %t, Recv: %v", d.Name.Name, d.Name.IsExported(), d.Recv)
					if d.Recv == nil && d.Name.IsExported() {
						// fmt.Printf("    Generating alias for function: %s", d.Name.Name)
						var args []ast.Expr
						if d.Type.Params != nil {
							for _, param := range d.Type.Params.List {
								for _, name := range param.Names {
									args = append(args, name)
								}
							}
						}
						callExpr := &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent(pkg.ImportAlias),
								Sel: d.Name,
							},
							Args: args,
						}
						var results []ast.Stmt
						if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
							results = []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{callExpr}}}
						} else {
							results = []ast.Stmt{&ast.ExprStmt{X: callExpr}}
						}
						funcDecl := &ast.FuncDecl{
							Name: d.Name,
							Type: qualifyType(d.Type, pkg.ImportAlias, g.definedTypes).(*ast.FuncType), // Qualify the function type
							Body: &ast.BlockStmt{List: results},
						}
						if g.allPackageDecls[pkg.ImportAlias] == nil {
							g.allPackageDecls[pkg.ImportAlias] = &packageDecls{}
						}
						g.allPackageDecls[pkg.ImportAlias].funcDecls = append(g.allPackageDecls[pkg.ImportAlias].funcDecls, funcDecl)
					}
				case *ast.GenDecl:
					if d.Tok == token.TYPE {
						// Already processed in the first pass
						continue
					}
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.IsExported() {
									newSpec := &ast.ValueSpec{
										Names: []*ast.Ident{name},
										Values: []ast.Expr{
											&ast.SelectorExpr{
												X:   ast.NewIdent(pkg.ImportAlias),
												Sel: name,
											},
										},
									}
									if g.allPackageDecls[pkg.ImportAlias] == nil {
										g.allPackageDecls[pkg.ImportAlias] = &packageDecls{}
									}
									if d.Tok == token.VAR {
										g.allPackageDecls[pkg.ImportAlias].varSpecs = append(g.allPackageDecls[pkg.ImportAlias].varSpecs, newSpec)
									} else if d.Tok == token.CONST {
										g.allPackageDecls[pkg.ImportAlias].constSpecs = append(g.allPackageDecls[pkg.ImportAlias].constSpecs, newSpec)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	var orderedDecls []ast.Decl

	// Add imports first
	var finalImportSpecs []ast.Spec
	for _, spec := range g.importSpecs {
		finalImportSpecs = append(finalImportSpecs, spec)
	}
	if len(finalImportSpecs) > 0 {
		orderedDecls = append(orderedDecls, &ast.GenDecl{Tok: token.IMPORT, Specs: finalImportSpecs})
	}

	// Collect all declarations by type across all packages
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

		// Update function declarations to use local type references
		for _, funcDecl := range pkgDecls.funcDecls {
			if fd, ok := funcDecl.(*ast.FuncDecl); ok {
				// Re-qualify the function type with defined types
				fd.Type = qualifyType(fd.Type, alias, g.definedTypes).(*ast.FuncType)
			}
			allFuncDecls = append(allFuncDecls, funcDecl)
		}
	}

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

	// Apply replacements to each declaration
	for i, decl := range g.aliasFile.Decls {
		g.aliasFile.Decls[i] = g.compiledCfg.Replacer.Apply(decl).(ast.Decl)
	}

	return printer.Fprint(outFile, g.fset, g.aliasFile)
}
