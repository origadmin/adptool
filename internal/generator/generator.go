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
func qualifyType(expr ast.Expr, pkgAlias string) ast.Expr {
	switch t := expr.(type) {
	case *ast.Ident:
		// Only qualify if it's not already qualified (i.e., not a SelectorExpr)
		// and it's an exported identifier.
		if t.IsExported() {
			return &ast.SelectorExpr{
				X:   ast.NewIdent(pkgAlias),
				Sel: t,
			}
		}
		return t
	case *ast.StarExpr:
		t.X = qualifyType(t.X, pkgAlias)
		return t
	case *ast.ArrayType:
		t.Elt = qualifyType(t.Elt, pkgAlias)
		return t
	case *ast.MapType:
		t.Key = qualifyType(t.Key, pkgAlias)
		t.Value = qualifyType(t.Value, pkgAlias)
		return t
	case *ast.ChanType:
		t.Value = qualifyType(t.Value, pkgAlias)
		return t
	case *ast.FuncType:
		if t.Params != nil {
			for _, field := range t.Params.List {
				field.Type = qualifyType(field.Type, pkgAlias)
			}
		}
		if t.Results != nil {
			for _, field := range t.Results.List {
				field.Type = qualifyType(field.Type, pkgAlias)
			}
		}
		return t
	case *ast.InterfaceType:
		// Interface methods' types need to be qualified
		if t.Methods != nil {
			for _, field := range t.Methods.List {
				if funcType, ok := field.Type.(*ast.FuncType); ok {
					field.Type = qualifyType(funcType, pkgAlias)
				}
			}
		}
		return t
	case *ast.StructType:
		// Struct fields' types need to be qualified
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				field.Type = qualifyType(field.Type, pkgAlias)
			}
		}
		return t
	case *ast.SelectorExpr:
		// If it's already a SelectorExpr, we don't need to qualify it further.
		// However, its X (package part) might need qualification if it's a nested selector.
		// For simplicity, we'll assume top-level qualification for now.
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
							Type: qualifyType(d.Type, pkg.ImportAlias).(*ast.FuncType), // Qualify the function type
							Body: &ast.BlockStmt{List: results},
						}
						if allPackageDecls[pkg.ImportAlias] == nil {
							allPackageDecls[pkg.ImportAlias] = &packageDecls{}
						}
						allPackageDecls[pkg.ImportAlias].funcDecls = append(allPackageDecls[pkg.ImportAlias].funcDecls, funcDecl)
					}
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if s.Name.IsExported() {
								if allPackageDecls[pkg.ImportAlias] == nil {
									allPackageDecls[pkg.ImportAlias] = &packageDecls{}
								}
								allPackageDecls[pkg.ImportAlias].typeSpecs = append(allPackageDecls[pkg.ImportAlias].typeSpecs, &ast.TypeSpec{
									Name: s.Name,
									Type: &ast.SelectorExpr{
										X:   ast.NewIdent(pkg.ImportAlias),
										Sel: s.Name,
									},
								})
							}
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

	// Sort package aliases for consistent output
	var sortedPackageAliases []string
	for alias := range allPackageDecls {
		sortedPackageAliases = append(sortedPackageAliases, alias)
	}
	sort.Strings(sortedPackageAliases)

	for _, alias := range sortedPackageAliases {
		pkgDecls := allPackageDecls[alias]

		// Add grouped value declarations (CONST)
		if len(pkgDecls.constSpecs) > 0 {
			orderedDecls = append(orderedDecls, &ast.GenDecl{Tok: token.CONST, Specs: pkgDecls.constSpecs})
		}

		// Add grouped value declarations (VAR)
		if len(pkgDecls.varSpecs) > 0 {
			orderedDecls = append(orderedDecls, &ast.GenDecl{Tok: token.VAR, Specs: pkgDecls.varSpecs})
		}

		// Add grouped type declarations
		if len(pkgDecls.typeSpecs) > 0 {
			orderedDecls = append(orderedDecls, &ast.GenDecl{Tok: token.TYPE, Specs: pkgDecls.typeSpecs})
		}

		// Add function declarations
		orderedDecls = append(orderedDecls, pkgDecls.funcDecls...)
	}

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
