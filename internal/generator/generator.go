package generator

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"

	"golang.org/x/tools/go/packages"

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

// Generate creates an alias package for a source package, applying renaming rules.
// replacer: The Replacer instance to apply renaming rules.
// sourcePackageImportPath: The import path of the original package (e.g., "fmt", "log").
// aliasPackageName: The desired name for the new alias package (e.g., "myfmt").
// outputFilePath: The full path to the output .go file for the alias package.
func Generate(replacer interfaces.Replacer, sourcePackageImportPath, aliasPackageName, outputFilePath string) error {
	fset := token.NewFileSet()

	// 1. Load the source package
	cfg := &packages.Config{
		Mode: packages.LoadSyntax | packages.LoadTypes, // Load AST and type information
	}
	pkgs, err := packages.Load(cfg, sourcePackageImportPath)
	if err != nil {
		return fmt.Errorf("failed to load source package %s: %w", sourcePackageImportPath, err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("package loading had errors for %s", sourcePackageImportPath)
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found for import path %s", sourcePackageImportPath)
	}
	sourcePkg := pkgs[0] // Assuming we're interested in the first package found

	// 2. Create a new AST for the alias package
	aliasFile := &ast.File{
		Name:  ast.NewIdent(aliasPackageName),
		Decls: []ast.Decl{},
	}

	// Add import for the original package
	aliasFile.Decls = append(aliasFile.Decls, &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", sourcePackageImportPath)},
				Name: ast.NewIdent("original"), // Alias the original package to avoid name conflicts
			},
		},
	})

	// 3. Iterate through exported symbols of the source package and generate wrappers
	for _, file := range sourcePkg.Syntax {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && d.Name.IsExported() {
					// Create a list of argument expressions for the call
					var args []ast.Expr
					if d.Type.Params != nil {
						for _, param := range d.Type.Params.List {
							for _, name := range param.Names {
								args = append(args, name)
							}
						}
					}

					// Create the call expression
					callExpr := &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("original"),
							Sel: d.Name,
						},
						Args: args,
					}

					// Create the return statement
					var results []ast.Stmt
					if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
						results = []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{callExpr},
							},
						}
					} else {
						results = []ast.Stmt{
							&ast.ExprStmt{X: callExpr},
						}
					}

					aliasFile.Decls = append(aliasFile.Decls, &ast.FuncDecl{
						Name: d.Name,
						Type: d.Type,
						Body: &ast.BlockStmt{
							List: results,
						},
					})
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							aliasFile.Decls = append(aliasFile.Decls, &ast.GenDecl{
								Tok: token.TYPE,
								Specs: []ast.Spec{
									&ast.TypeSpec{
										Name:   s.Name,
										Assign: s.Pos(), // The '=' in type T = original.T
										Type: &ast.SelectorExpr{
											X:   ast.NewIdent("original"),
											Sel: s.Name,
										},
									},
								},
							})
						}
					case *ast.ValueSpec: // For variables and constants
						for _, name := range s.Names {
							if name.IsExported() {
								aliasFile.Decls = append(aliasFile.Decls, &ast.GenDecl{
									Tok: d.Tok, // Use the original token (VAR or CONST)
									Specs: []ast.Spec{
										&ast.ValueSpec{
											Names: []*ast.Ident{name},
											Values: []ast.Expr{
												&ast.SelectorExpr{
													X:   ast.NewIdent("original"),
													Sel: name,
												},
											},
										},
									},
								})
							}
						}
					}
				}
			}
		}
	}

	// 4. Print the new AST to the output file
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputFilePath, err)
	}
	defer outFile.Close()

	err = printer.Fprint(outFile, fset, aliasFile)
	if err != nil {
		return fmt.Errorf("failed to print AST to file %s: %w", outputFilePath, err)
	}
	fmt.Printf("Successfully generated alias package to %s\n", outputFilePath)

	return nil
}
