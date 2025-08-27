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
	// allPackageDecls is keyed by import path
	allPackageDecls map[string]*packageDecls
	importSpecs     map[string]*ast.ImportSpec
	replacer        interfaces.Replacer
	// pathToAlias maps import path to its generated alias
	pathToAlias map[string]string
}

// NewCollector creates a new Collector.
func NewCollector(replacer interfaces.Replacer) *Collector {
	return &Collector{
		allPackageDecls: make(map[string]*packageDecls),
		importSpecs:     make(map[string]*ast.ImportSpec),
		replacer:        replacer,
		pathToAlias:     make(map[string]string),
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
			// 如果是空导入 (import _ "path")，则跳过
			if importSpec.Name != nil && importSpec.Name.Name == "_" {
				continue
			}
			importPath := strings.Trim(importSpec.Path.Value, "\"")
			if _, exists := c.importSpecs[importPath]; !exists {
				c.importSpecs[importPath] = importSpec
			}
		}
	}
}

func (c *Collector) collectTypeDeclarations(sourcePkg *packages.Package, importPath, importAlias string) {
	for _, file := range sourcePkg.Syntax {
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.IsExported() {
						c.collectTypeDeclaration(typeSpec, importPath, importAlias)
					}
				}
			}
		}
	}
}

func (c *Collector) collectTypeDeclaration(typeSpec *ast.TypeSpec, importPath, importAlias string) {
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

	if c.allPackageDecls[importPath] == nil {
		c.allPackageDecls[importPath] = &packageDecls{}
	}
	c.allPackageDecls[importPath].typeSpecs = append(c.allPackageDecls[importPath].typeSpecs, newSpec)
}

func (c *Collector) collectOtherDeclarations(sourcePkg *packages.Package, importPath, importAlias string) {
	for _, file := range sourcePkg.Syntax {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				c.collectFunctionDeclaration(d, sourcePkg, importPath, importAlias)
			case *ast.GenDecl:
				switch d.Tok {
				case token.CONST:
					c.collectValueDeclaration(d, importPath, importAlias, token.CONST)
				case token.VAR:
					c.collectValueDeclaration(d, importPath, importAlias, token.VAR)
				}
			}
		}
	}
}

func (c *Collector) collectFunctionDeclaration(funcDecl *ast.FuncDecl, sourcePkg *packages.Package, importPath, importAlias string) {
	if funcDecl.Recv == nil && funcDecl.Name.IsExported() {
		if containsInvalidTypes(sourcePkg.TypesInfo, sourcePkg.Types, funcDecl.Type) {
			slog.Debug("Skipping function because it uses unexported or internal types", "func", "Collector.collectFunctionDeclaration", "function", funcDecl.Name.Name)
			return
		}
		originalName := funcDecl.Name.Name

		var args []ast.Expr
		if funcDecl.Type.Params != nil {
			// Collect all existing parameter names to avoid collisions.
			existingNames := make(map[string]bool)
			for _, param := range funcDecl.Type.Params.List {
				for _, name := range param.Names {
					if name.Name != "_" {
						existingNames[name.Name] = true
					}
				}
			}

			unnamedParamCounter := 0
			// generateUniqueName creates a unique parameter name that doesn't conflict with existing ones.
			generateUniqueName := func() string {
				for {
					newName := fmt.Sprintf("p%d", unnamedParamCounter)
					unnamedParamCounter++
					if !existingNames[newName] {
						// Add to existing names to prevent future collisions in the same function.
						existingNames[newName] = true
						return newName
					}
				}
			}

			for _, param := range funcDecl.Type.Params.List {
				if len(param.Names) == 0 {
					// This is an unnamed parameter, generate a unique name.
					newName := generateUniqueName()
					newIdent := ast.NewIdent(newName)
					param.Names = []*ast.Ident{newIdent}
					args = append(args, newIdent)
				} else {
					for i, name := range param.Names {
						if name.Name == "_" {
							// Parameter name is _, generate a unique name.
							newName := generateUniqueName()
							newIdent := ast.NewIdent(newName)
							param.Names[i] = newIdent
							args = append(args, newIdent)
						} else {
							args = append(args, name)
						}
					}
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
			Type: qualifyType(funcDecl.Type, importAlias, nil, nil).(*ast.FuncType),
			Body: &ast.BlockStmt{List: results},
		}

		if c.allPackageDecls[importPath] == nil {
			c.allPackageDecls[importPath] = &packageDecls{}
		}
		c.allPackageDecls[importPath].funcDecls = append(c.allPackageDecls[importPath].funcDecls, newFuncDecl)
	}
}

func (c *Collector) collectValueDeclaration(genDecl *ast.GenDecl, importPath, importAlias string, tok token.Token) {
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

					if c.allPackageDecls[importPath] == nil {
						c.allPackageDecls[importPath] = &packageDecls{}
					}

					if tok == token.VAR {
						c.allPackageDecls[importPath].varDecls = append(c.allPackageDecls[importPath].varDecls, newDecl)
					} else if tok == token.CONST {
						c.allPackageDecls[importPath].constDecls = append(c.allPackageDecls[importPath].constDecls, newDecl)
					}
				}
			}
		}
	}
}

func (c *Collector) applyReplacements() {
	for importPath, pkgDecls := range c.allPackageDecls {
		alias := c.pathToAlias[importPath]
		pkgCtx := interfaces.NewContext().WithValue(interfaces.PackagePathContextKey, importPath)

		// First, process all type declarations.
		for i, spec := range pkgDecls.typeSpecs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				typeCtx := pkgCtx.Push(interfaces.RuleTypeType)
				replaced := c.replacer.Apply(typeCtx, typeSpec)
				if replacedSpec, ok := replaced.(*ast.TypeSpec); ok {
					pkgDecls.typeSpecs[i] = replacedSpec
					slog.Debug("Applied replacer to type", "func", "Collector.applyReplacements", "type", replacedSpec.Name.Name)
				}
			}
		}

		// Now, process other declarations.
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
				replacedDecl.Type = qualifyType(replacedDecl.Type, alias, nil, nil).(*ast.FuncType)
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

	// Handle case where the name is only special characters
	if strings.TrimFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) == "" {
		return "pkg"
	}

	// Convert hyphens to camelCase
	parts := strings.Split(name, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			// Capitalize the first letter of each part after the first
			runes := []rune(parts[i])
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
	}
	name = strings.Join(parts, "")

	// Process each character to build a valid Go identifier
	var result strings.Builder
	firstChar := true
	for _, r := range name {
		// For the first character, ensure it's a letter or underscore
		if firstChar {
			if unicode.IsLetter(r) {
				result.WriteRune(unicode.ToLower(r))
			} else if r == '_' {
				result.WriteRune(r)
			} else {
				// If first char is not a letter or underscore, prepend 'p'
				result.WriteRune('p')
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					result.WriteRune(unicode.ToLower(r))
				}
			}
			firstChar = false
			continue
		}

		// For subsequent characters
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			result.WriteRune(r)
		}
		// Skip other special characters
	}

	cleaned := result.String()

	// Handle empty result
	if cleaned == "" {
		return "pkg"
	}

	// If it's a Go keyword, append "Pkg"
	if token.Lookup(cleaned).IsKeyword() {
		return cleaned + "Pkg"
	}

	return cleaned
}

// Collect method to use the new alias manager
func (c *Collector) Collect(packages []*PackageInfo) error {
	aliasMgr := newAliasManager()
	processedPaths := make(map[string]bool) // Keep track of processed package paths

	for _, pkg := range packages {
		// If we have already processed this package path, skip it.
		if processedPaths[pkg.ImportPath] {
			continue
		}

		// Generate or use the specified alias
		importAlias := aliasMgr.generateAlias(pkg.ImportPath, pkg.ImportAlias)

		c.pathToAlias[pkg.ImportPath] = importAlias
		c.importSpecs[pkg.ImportPath] = &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", pkg.ImportPath)},
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

		// Mark this path as processed.
		processedPaths[pkg.ImportPath] = true

		c.collectImports(sourcePkg)
		c.collectTypeDeclarations(sourcePkg, pkg.ImportPath, importAlias)
		c.collectOtherDeclarations(sourcePkg, pkg.ImportPath, importAlias)
	}

	if c.replacer != nil {
		c.applyReplacements()
	}

	return nil
}
