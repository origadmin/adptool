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
	"github.com/origadmin/adptool/internal/interfaces"
)

// replacerVisitor implements the ast.Visitor interface to apply replacements.
type replacerVisitor struct {
	replacer interfaces.Replacer
}

// Visit implements the ast.Visitor interface.
func (v replacerVisitor) Visit(n ast.Node) ast.Visitor {
	if n != nil {
		v.replacer.Apply(n)
	}
	return v // Continue traversal
}

// qualifyType recursively qualifies ast.Ident nodes within an ast.Expr
// with the given package alias if they are not already qualified.
// definedTypes is a map of type names that are defined in the current package.
func qualifyType(expr ast.Expr, pkgAlias string, definedTypes map[string]bool) ast.Expr {
	switch t := expr.(type) {
	case *ast.Ident:
		// Only qualify if it's not already qualified (i.e., not a SelectorExpr),
		// it's an exported identifier, and it's not defined in the current package.
		if t.IsExported() && !definedTypes[t.Name] {
			return &ast.SelectorExpr{
				X:   ast.NewIdent(pkgAlias),
				Sel: t,
			}
		}
		return t
	case *ast.StarExpr:
		t.X = qualifyType(t.X, pkgAlias, definedTypes)
		return t
	case *ast.ArrayType:
		t.Elt = qualifyType(t.Elt, pkgAlias, definedTypes)
		return t
	case *ast.MapType:
		t.Key = qualifyType(t.Key, pkgAlias, definedTypes)
		t.Value = qualifyType(t.Value, pkgAlias, definedTypes)
		return t
	case *ast.ChanType:
		t.Value = qualifyType(t.Value, pkgAlias, definedTypes)
		return t
	case *ast.FuncType:
		if t.Params != nil {
			for _, field := range t.Params.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes)
			}
		}
		if t.Results != nil {
			for _, field := range t.Results.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes)
			}
		}
		return t
	case *ast.InterfaceType:
		// Interface methods' types need to be qualified
		if t.Methods != nil {
			for _, field := range t.Methods.List {
				if funcType, ok := field.Type.(*ast.FuncType); ok {
					field.Type = qualifyType(funcType, pkgAlias, definedTypes)
				}
			}
		}
		return t
	case *ast.StructType:
		// Struct fields' types need to be qualified
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes)
			}
		}
		return t
	case *ast.SelectorExpr:
		// If it's already a SelectorExpr, check if it's a reference to a type from the source package
		// that we've defined locally
		if ident, ok := t.X.(*ast.Ident); ok && ident.Name == pkgAlias {
			typeName := t.Sel.Name
			if definedTypes[typeName] {
				// Replace with local alias
				return ast.NewIdent(typeName)
			}
		}
		return t
	default:
		return expr
	}
}

// Generate creates an alias package based on the compiled configuration.
func Generate(compiledCfg *config.CompiledConfig, outputFilePath string) error {
	fset := token.NewFileSet()

	aliasFile := &ast.File{
		Name:  ast.NewIdent(compiledCfg.PackageName),
		Decls: []ast.Decl{},
	}

	// Removed aliasFile.Doc assignment

	outFile, err := os.Create(outputFilePath)
	fmt.Printf("Attempting to create file: %s, Error: %v\n", outputFilePath, err)
	if err != nil {
		return err
	}
	defer outFile.Close()

	importSpecs := make(map[string]*ast.ImportSpec)

	// Separate declarations by type
	type packageDecls struct {
		typeSpecs  []ast.Spec
		varSpecs   []ast.Spec
		constSpecs []ast.Spec
		funcDecls  []ast.Decl
	}
	var allPackageDecls = make(map[string]*packageDecls)
	
	// Initialize definedTypes map to collect all defined types
	definedTypes := make(map[string]bool)

	// Process each source package and generate declarations
	for _, pkg := range compiledCfg.Packages {
		// Add the primary package being adapted to the import list
		importSpecs[pkg.ImportPath] = &ast.ImportSpec{
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
			if _, exists := importSpecs[importPath]; !exists {
				importSpecs[importPath] = &ast.ImportSpec{
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
							if allPackageDecls[pkg.ImportAlias] == nil {
								allPackageDecls[pkg.ImportAlias] = &packageDecls{}
							}
							allPackageDecls[pkg.ImportAlias].typeSpecs = append(allPackageDecls[pkg.ImportAlias].typeSpecs, &ast.TypeSpec{
								Name: typeSpec.Name,
								Type: &ast.SelectorExpr{
									X:   ast.NewIdent(pkg.ImportAlias),
									Sel: typeSpec.Name,
								},
								Assign: token.Pos(1), // Set non-zero position to make it an alias type (type A = B)
							})
							// Add to definedTypes map
							definedTypes[typeSpec.Name.Name] = true
						}
					}
				}
			}
		}

		// Second pass: collect other declarations
		for _, file := range sourcePkg.Syntax {
			fileInfo := fset.File(file.Pos())
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
							Type: qualifyType(d.Type, pkg.ImportAlias, definedTypes).(*ast.FuncType), // Qualify the function type
							Body: &ast.BlockStmt{List: results},
						}
						if allPackageDecls[pkg.ImportAlias] == nil {
							allPackageDecls[pkg.ImportAlias] = &packageDecls{}
						}
						allPackageDecls[pkg.ImportAlias].funcDecls = append(allPackageDecls[pkg.ImportAlias].funcDecls, funcDecl)
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
									if allPackageDecls[pkg.ImportAlias] == nil {
										allPackageDecls[pkg.ImportAlias] = &packageDecls{}
									}
									if d.Tok == token.VAR {
										allPackageDecls[pkg.ImportAlias].varSpecs = append(allPackageDecls[pkg.ImportAlias].varSpecs, newSpec)
									} else if d.Tok == token.CONST {
										allPackageDecls[pkg.ImportAlias].constSpecs = append(allPackageDecls[pkg.ImportAlias].constSpecs, newSpec)
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
	for _, spec := range importSpecs {
		finalImportSpecs = append(finalImportSpecs, spec)
	}
	if len(finalImportSpecs) > 0 {
		orderedDecls = append(orderedDecls, &ast.GenDecl{Tok: token.IMPORT, Specs: finalImportSpecs})
	}

	// definedTypes has already been populated in the first pass

	// Collect all declarations by type across all packages
	var allConstSpecs []ast.Spec
	var allVarSpecs []ast.Spec
	var allTypeSpecs []ast.Spec
	var allFuncDecls []ast.Decl

	// Sort package aliases for consistent output
	var sortedPackageAliases []string
	for alias := range allPackageDecls {
		sortedPackageAliases = append(sortedPackageAliases, alias)
	}
	sort.Strings(sortedPackageAliases)

	for _, alias := range sortedPackageAliases {
		pkgDecls := allPackageDecls[alias]
		
		// Collect all declarations
		allConstSpecs = append(allConstSpecs, pkgDecls.constSpecs...)
		allVarSpecs = append(allVarSpecs, pkgDecls.varSpecs...)
		allTypeSpecs = append(allTypeSpecs, pkgDecls.typeSpecs...)
		
		// Update function declarations to use local type references
		for _, funcDecl := range pkgDecls.funcDecls {
			if fd, ok := funcDecl.(*ast.FuncDecl); ok {
				// Re-qualify the function type with defined types
				fd.Type = qualifyType(fd.Type, alias, definedTypes).(*ast.FuncType)
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

	aliasFile.Decls = orderedDecls

	// Apply replacements to each declaration
	for i, decl := range aliasFile.Decls {
		aliasFile.Decls[i] = compiledCfg.Replacer.Apply(decl).(ast.Decl)
	}

	// Debug print final declarations
	// fmt.Println("Final aliasFile.Decls:")
	// for _, decl := range aliasFile.Decls {
	// 	switch d := decl.(type) {
	// 	case *ast.GenDecl:
	// 		if d.Tok == token.IMPORT {
	// 			for _, spec := range d.Specs {
	// 				if s, ok := spec.(*ast.ImportSpec); ok {
	// 					fmt.Printf("  Import: %s", s.Path.Value)
	// 				}
	// 			}
	// 		} else if len(d.Specs) > 0 {
	// 			// Handle TypeSpec and ValueSpec within GenDecl
	// 			for _, spec := range d.Specs {
	// 				switch s := spec.(type) {
	// 				case *ast.TypeSpec:
	// 					fmt.Printf("  Type: %s", s.Name.Name)
	// 				case *ast.ValueSpec:
	// 					for _, name := range s.Names {
	// 						fmt.Printf("  Value: %s", name.Name)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	case *ast.FuncDecl:
	// 		fmt.Printf("  Func: %s", d.Name.Name)
	// 	default:
	// 		fmt.Printf("  Other Decl Type: %T", decl)
	// 	}
	// }

	return printer.Fprint(outFile, fset, aliasFile)
}
