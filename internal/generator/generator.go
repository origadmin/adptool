package generator

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"

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

// Generate creates an alias package based on the compiled configuration.
func Generate(compiledCfg *config.CompiledConfig, outputFilePath string) error {
	fset := token.NewFileSet()

	aliasFile := &ast.File{
		Name:  ast.NewIdent(compiledCfg.PackageName),
		Decls: []ast.Decl{},
	}

	for _, pkg := range compiledCfg.Packages {
		aliasFile.Decls = append(aliasFile.Decls, &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", pkg.ImportPath)},
					Name: ast.NewIdent(pkg.ImportAlias),
				},
			},
		})

		loadCfg := &packages.Config{
			Mode: packages.LoadSyntax | packages.LoadTypes,
		}
		pkgs, err := packages.Load(loadCfg, pkg.ImportPath)
		if err != nil {
			return err
		}
		if len(pkgs) == 0 {
			return nil // Should not happen in test, but good practice
		}
		sourcePkg := pkgs[0]

		for _, file := range sourcePkg.Syntax {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Recv == nil && d.Name.IsExported() {
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
						aliasFile.Decls = append(aliasFile.Decls, &ast.FuncDecl{
							Name: d.Name,
							Type: d.Type,
							Body: &ast.BlockStmt{List: results},
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
											Assign: s.Pos(),
											Type: &ast.SelectorExpr{
												X:   ast.NewIdent(pkg.ImportAlias),
												Sel: s.Name,
											},
										},
									},
								})
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.IsExported() {
									aliasFile.Decls = append(aliasFile.Decls, &ast.GenDecl{
										Tok: d.Tok,
										Specs: []ast.Spec{
											&ast.ValueSpec{
												Names: []*ast.Ident{name},
												Values: []ast.Expr{
													&ast.SelectorExpr{
														X:   ast.NewIdent(pkg.ImportAlias),
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
	}

	outFile, err := os.Create(outputFilePath)
	if err != nil {
			return err
		}
	defer outFile.Close()

	return printer.Fprint(outFile, fset, aliasFile)
}
