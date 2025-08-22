package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"path"
	"strconv"
	"strings"
	"unicode"

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
	aliasToPath     map[string]string // Map from import alias to import path
}

// NewCollector creates a new Collector.
func NewCollector(replacer interfaces.Replacer) *Collector {
	return &Collector{
		allPackageDecls: make(map[string]*packageDecls),
		definedTypes:    make(map[string]bool),
		importSpecs:     make(map[string]*ast.ImportSpec),
		replacer:        replacer,
		aliasToPath:     make(map[string]string),
	}
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
			importPath := strings.Trim(importSpec.Path.Value, "\"")
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
	if !typeSpec.Name.IsExported() {
		return
	}

	originalName := typeSpec.Name.Name
	newSpec := &ast.TypeSpec{
		Name:   typeSpec.Name, // This will be replaced later
		Assign: 1,             // Make it an alias with '='
	}

	// Handle generics in type declarations
	if typeSpec.TypeParams != nil {
		newSpec.TypeParams = typeSpec.TypeParams

		var indices []ast.Expr
		for _, list := range typeSpec.TypeParams.List {
			for _, name := range list.Names {
				indices = append(indices, ast.NewIdent(name.Name))
			}
		}

		baseType := &ast.SelectorExpr{
			X:   ast.NewIdent(importAlias),
			Sel: ast.NewIdent(originalName),
		}

		if len(indices) == 1 {
			newSpec.Type = &ast.IndexExpr{
				X:     baseType,
				Index: indices[0],
			}
		} else {
			newSpec.Type = &ast.IndexListExpr{
				X:       baseType,
				Indices: indices,
			}
		}
	} else {
		newSpec.Type = &ast.SelectorExpr{
			X:   ast.NewIdent(importAlias),
			Sel: ast.NewIdent(originalName),
		}
	}

	if c.allPackageDecls[importAlias] == nil {
		c.allPackageDecls[importAlias] = &packageDecls{}
	}
	c.allPackageDecls[importAlias].typeSpecs = append(c.allPackageDecls[importAlias].typeSpecs, newSpec)
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
		if hasUnexportedTypes(funcDecl.Type) {
			slog.Debug("Skipping function because it uses unexported types", "func", "Collector.collectFunctionDeclaration", "function", funcDecl.Name.Name)
			return
		}
		originalName := funcDecl.Name.Name

		var args []ast.Expr
		if funcDecl.Type.Params != nil {
			for _, param := range funcDecl.Type.Params.List {
				for _, name := range param.Names {
					args = append(args, name)
				}
			}
		}

		var callFun ast.Expr = &ast.SelectorExpr{
			X:   ast.NewIdent(importAlias),
			Sel: ast.NewIdent(originalName),
		}

		// Handle generic function calls
		if funcDecl.Type.TypeParams != nil {
			var typeArgs []ast.Expr
			for _, param := range funcDecl.Type.TypeParams.List {
				for _, name := range param.Names {
					typeArgs = append(typeArgs, ast.NewIdent(name.Name))
				}
			}
			if len(typeArgs) > 0 {
				if len(typeArgs) == 1 {
					callFun = &ast.IndexExpr{
						X:     callFun,
						Index: typeArgs[0],
					}
				} else {
					callFun = &ast.IndexListExpr{
						X:       callFun,
						Indices: typeArgs,
					}
				}
			}
		}

		callExpr := &ast.CallExpr{
			Fun:  callFun,
			Args: args,
		}

		// Check if the original function is variadic
		if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
			lastParam := funcDecl.Type.Params.List[len(funcDecl.Type.Params.List)-1]
			if _, ok := lastParam.Type.(*ast.Ellipsis); ok {
				// The original function is variadic, so set Ellipsis for the call
				callExpr.Ellipsis = callExpr.Rparen - 1 // A valid position, just before Rparen
			}
		}

		var results []ast.Stmt
		if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
			results = []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{callExpr}}}
		} else {
			results = []ast.Stmt{&ast.ExprStmt{X: callExpr}}
		}

		newFuncDecl := &ast.FuncDecl{
			Name: funcDecl.Name,
			Type: qualifyType(funcDecl.Type, importAlias, c.definedTypes, nil).(*ast.FuncType),
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

	for alias, pkgDecls := range c.allPackageDecls {
		importPath := c.aliasToPath[alias]
		pkgCtx := interfaces.NewContext().WithValue(interfaces.PackagePathContextKey, importPath)

		// First, process all type declarations and populate the new defined types map.
		for i, spec := range pkgDecls.typeSpecs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				typeCtx := pkgCtx.Push(interfaces.RuleTypeType)
				replaced := c.replacer.Apply(typeCtx, typeSpec)
				if replacedSpec, ok := replaced.(*ast.TypeSpec); ok {
					pkgDecls.typeSpecs[i] = replacedSpec
					newDefinedTypes[replacedSpec.Name.Name] = true
					slog.Debug("Applied replacer to type", "func", "Collector.applyReplacements", "type", replacedSpec.Name.Name)
				}
			}
		}

		c.definedTypes = newDefinedTypes

		// Now, process other declarations using the correct context and definedTypes map.
		for i, decl := range pkgDecls.constDecls {
			replaced := c.replacer.Apply(pkgCtx, decl)
			if replacedDecl, ok := replaced.(*ast.GenDecl); ok {
				pkgDecls.constDecls[i] = replacedDecl
			}
		}

		for i, decl := range pkgDecls.varDecls {
			replaced := c.replacer.Apply(pkgCtx, decl)
			if replacedDecl, ok := replaced.(*ast.GenDecl); ok {
				pkgDecls.varDecls[i] = replacedDecl
			}
		}

		for i, decl := range pkgDecls.funcDecls {
			replaced := c.replacer.Apply(pkgCtx, decl)
			if replacedDecl, ok := replaced.(*ast.FuncDecl); ok {
				replacedDecl.Type = qualifyType(replacedDecl.Type, alias, c.definedTypes, nil).(*ast.FuncType)
				pkgDecls.funcDecls[i] = replacedDecl
			}
		}
	}
}

// aliasManager handles package alias generation and deduplication
type aliasManager struct {
	usedAliases map[string]string // alias -> importPath
}

func newAliasManager() *aliasManager {
	return &aliasManager{
		usedAliases: make(map[string]string),
	}
}

func (m *aliasManager) generateAlias(importPath, preferredAlias string) string {
	// 1. Use preferred alias if provided and available
	if preferredAlias != "" {
		if existingPath, exists := m.usedAliases[preferredAlias]; !exists || existingPath == importPath {
			m.usedAliases[preferredAlias] = importPath
			return preferredAlias
		}
	}

	// 2. Extract base name from import path
	baseName := path.Base(importPath)

	// 3. Sanitize the name
	baseName = sanitizePackageName(baseName)

	// 4. Handle conflicts
	alias := baseName
	counter := 1
	for {
		if existingPath, exists := m.usedAliases[alias]; !exists || existingPath == importPath {
			m.usedAliases[alias] = importPath
			return alias
		}
		alias = baseName + strconv.Itoa(counter)
		counter++
	}
}

func sanitizePackageName(name string) string {
	if name == "" {
		return "pkg"
	}

	var result strings.Builder
	for i, r := range name {
		if i == 0 && !unicode.IsLetter(r) {
			result.WriteRune('p')
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	cleaned := result.String()
	if token.Lookup(cleaned).IsKeyword() {
		return cleaned + "_"
	}

	return cleaned
}

// Collect method to use the new alias manager
func (c *Collector) Collect(packages []*PackageInfo) error {
	aliasMgr := newAliasManager()

	for _, pkg := range packages {
		// Generate or use the specified alias
		importAlias := aliasMgr.generateAlias(pkg.ImportPath, pkg.ImportAlias)

		c.aliasToPath[importAlias] = pkg.ImportPath
		c.importSpecs[pkg.ImportPath] = &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", pkg.ImportPath)},
			Name: &ast.Ident{Name: importAlias},
		}
		pkg.ImportAlias = importAlias

		sourcePkg, err := c.loadPackage(pkg.ImportPath)
		if err != nil {
			return err
		}
		if sourcePkg == nil {
			continue
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
