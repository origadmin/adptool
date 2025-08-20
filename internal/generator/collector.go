package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/origadmin/adptool/internal/interfaces"
)

// packageDecls holds declarations for a single package.
type packageDecls struct {
	typeSpecs  []ast.Spec
	varDecls   []ast.Decl // Changed from varSpecs to store GenDecls
	constDecls []ast.Decl // Changed from constSpecs to store GenDecls
	funcDecls  []ast.Decl
}

// Collector is responsible for collecting declarations from source packages.
type Collector struct {
	allPackageDecls map[string]*packageDecls
	definedTypes    map[string]bool
	importSpecs     map[string]*ast.ImportSpec
	replacer        interfaces.Replacer
}

// NewCollector creates a new Collector.
func NewCollector(replacer interfaces.Replacer) *Collector {
	return &Collector{
		allPackageDecls: make(map[string]*packageDecls),
		definedTypes:    make(map[string]bool),
		importSpecs:     make(map[string]*ast.ImportSpec),
		replacer:        replacer,
	}
}

// Collect processes each source package and collects declarations.
func (c *Collector) Collect(packages []*PackageInfo) error {
	for _, pkg := range packages {
		c.importSpecs[pkg.ImportPath] = &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", pkg.ImportPath)},
			Name: &ast.Ident{Name: pkg.ImportAlias},
		}

		sourcePkg, err := c.loadPackage(pkg.ImportPath)
		if err != nil {
			return err
		}
		if sourcePkg == nil {
			continue // Skip if package not found
		}

		c.collectImports(sourcePkg)
		c.collectTypeDeclarations(sourcePkg, pkg.ImportAlias)
		c.collectOtherDeclarations(sourcePkg, pkg.ImportAlias)
	}

	if c.replacer != nil {
		c.applyReplacements()
	}

	return nil
}

func (c *Collector) loadPackage(importPath string) (*packages.Package, error) {
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

func (c *Collector) collectImports(sourcePkg *packages.Package) {
	for _, file := range sourcePkg.Syntax {
		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, "")
			if _, exists := c.importSpecs[importPath]; !exists {
				c.importSpecs[importPath] = importSpec
			}
		}
	}
}

func (c *Collector) collectTypeDeclarations(sourcePkg *packages.Package, importAlias string) {
	for _, file := range sourcePkg.Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.IsExported() {
						c.collectTypeDeclaration(typeSpec, importAlias)
					}
				}
			}
		}
	}
}

func (c *Collector) collectTypeDeclaration(typeSpec *ast.TypeSpec, importAlias string) {
	if typeSpec.Name.IsExported() {
		originalName := typeSpec.Name.Name

		newSpec := &ast.TypeSpec{
			Name: typeSpec.Name,
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent(importAlias),
				Sel: ast.NewIdent(originalName),
			},
			Assign: 1, // Make it an alias with '='
		}

		if c.allPackageDecls[importAlias] == nil {
			c.allPackageDecls[importAlias] = &packageDecls{}
		}
		c.allPackageDecls[importAlias].typeSpecs = append(c.allPackageDecls[importAlias].typeSpecs, newSpec)

		// Removed: c.definedTypes[originalName] = true
		// Removed: log.Printf("collectTypeDeclaration: Added %s (original: %s) to definedTypes",
		// 	typeSpec.Name.Name, originalName)
	}
}

func (c *Collector) collectOtherDeclarations(sourcePkg *packages.Package, importAlias string) {
	for _, file := range sourcePkg.Syntax {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				c.collectFunctionDeclaration(d, importAlias)
			case *ast.GenDecl:
				switch d.Tok {
				case token.CONST:
					c.collectValueDeclaration(d, importAlias, token.CONST)
				case token.VAR:
					c.collectValueDeclaration(d, importAlias, token.VAR)
				}
			}
		}
	}
}

func (c *Collector) collectFunctionDeclaration(funcDecl *ast.FuncDecl, importAlias string) {
	if funcDecl.Recv == nil && funcDecl.Name.IsExported() {
		originalName := funcDecl.Name.Name

		var args []ast.Expr
		if funcDecl.Type.Params != nil {
			for _, param := range funcDecl.Type.Params.List {
				for _, name := range param.Names {
					args = append(args, name)
				}
			}
		}

		callExpr := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(importAlias),
				Sel: ast.NewIdent(originalName),
			},
			Args: args,
		}

		var results []ast.Stmt
		if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
			results = []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{callExpr}}}
		} else {
			results = []ast.Stmt{&ast.ExprStmt{X: callExpr}}
		}

		newFuncDecl := &ast.FuncDecl{
			Name: funcDecl.Name,
			Type: qualifyType(funcDecl.Type, importAlias, c.definedTypes).(*ast.FuncType),
			Body: &ast.BlockStmt{List: results},
		}

		if c.allPackageDecls[importAlias] == nil {
			c.allPackageDecls[importAlias] = &packageDecls{}
		}
		c.allPackageDecls[importAlias].funcDecls = append(c.allPackageDecls[importAlias].funcDecls, newFuncDecl)
	}
}

func (c *Collector) collectValueDeclaration(genDecl *ast.GenDecl, importAlias string, tok token.Token) {
	for _, spec := range genDecl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			for _, name := range valueSpec.Names {
				if name.IsExported() {
					originalName := name.Name

					newSpec := &ast.ValueSpec{
						Names: []*ast.Ident{name},
						Values: []ast.Expr{
							&ast.SelectorExpr{
								X:   ast.NewIdent(importAlias),
								Sel: ast.NewIdent(originalName),
							},
						},
					}
					newDecl := &ast.GenDecl{Tok: tok, Specs: []ast.Spec{newSpec}}

					if c.allPackageDecls[importAlias] == nil {
						c.allPackageDecls[importAlias] = &packageDecls{}
					}

					if tok == token.VAR {
						c.allPackageDecls[importAlias].varDecls = append(c.allPackageDecls[importAlias].varDecls, newDecl)
					} else if tok == token.CONST {
						c.allPackageDecls[importAlias].constDecls = append(c.allPackageDecls[importAlias].constDecls, newDecl)
					}
				}
			}
		}
	}
}

func (c *Collector) applyReplacements() {
	newDefinedTypes := make(map[string]bool)
	initialCtx := interfaces.NewContext()

	for alias, pkgDecls := range c.allPackageDecls {
		// First, process all type declarations and populate the new defined types map.
		for i, spec := range pkgDecls.typeSpecs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				replaced := c.replacer.Apply(initialCtx, typeSpec)
				if replacedSpec, ok := replaced.(*ast.TypeSpec); ok {
					pkgDecls.typeSpecs[i] = replacedSpec
					// Add the new, final type name to the map.
					newDefinedTypes[replacedSpec.Name.Name] = true
					log.Printf("applyReplacements: Applied replacer to type %s", replacedSpec.Name.Name)
				}
			}
		}

		// BEFORE processing functions and other declarations, replace the collector's
		// definedTypes map with the newly created, correct one.
		c.definedTypes = newDefinedTypes

		// Now, process other declarations using the correct definedTypes map.
		for i, decl := range pkgDecls.constDecls {
			replaced := c.replacer.Apply(initialCtx, decl)
			if replacedDecl, ok := replaced.(*ast.GenDecl); ok {
				pkgDecls.constDecls[i] = replacedDecl
			}
		}

		for i, decl := range pkgDecls.varDecls {
			replaced := c.replacer.Apply(initialCtx, decl)
			if replacedDecl, ok := replaced.(*ast.GenDecl); ok {
				pkgDecls.varDecls[i] = replacedDecl
			}
		}

		for i, decl := range pkgDecls.funcDecls {
			replaced := c.replacer.Apply(initialCtx, decl)
			if replacedDecl, ok := replaced.(*ast.FuncDecl); ok {
				// This call will now use the correct definedTypes map.
				replacedDecl.Type = qualifyType(replacedDecl.Type, alias, c.definedTypes).(*ast.FuncType)
				pkgDecls.funcDecls[i] = replacedDecl
			}
		}
	}
}
