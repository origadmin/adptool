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

	// 创建一个映射来跟踪所有声明的名称，用于检测冲突
	allNames := make(map[string]int) // name -> count

	// 辅助函数来记录名称
	recordName := func(name string) {
		allNames[name]++
	}

	for alias, pkgDecls := range c.allPackageDecls {
		importPath := c.aliasToPath[alias]
		pkgCtx := interfaces.NewContext().WithValue(interfaces.PackagePathContextKey, importPath)

		// First, process all type declarations and populate the new defined types map.
		for i, spec := range pkgDecls.typeSpecs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				originalName := typeSpec.Name.Name
				recordName(originalName)

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
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							recordName(name.Name)
						}
					}
				}
			}

			replaced := c.replacer.Apply(pkgCtx, decl)
			if replacedDecl, ok := replaced.(*ast.GenDecl); ok {
				pkgDecls.constDecls[i] = replacedDecl
			}
		}

		for i, decl := range pkgDecls.varDecls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							recordName(name.Name)
						}
					}
				}
			}

			replaced := c.replacer.Apply(pkgCtx, decl)
			if replacedDecl, ok := replaced.(*ast.GenDecl); ok {
				pkgDecls.varDecls[i] = replacedDecl
			}
		}

		for i, decl := range pkgDecls.funcDecls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				recordName(funcDecl.Name.Name)
			}

			replaced := c.replacer.Apply(pkgCtx, decl)
			if replacedDecl, ok := replaced.(*ast.FuncDecl); ok {
				replacedDecl.Type = qualifyType(replacedDecl.Type, alias, c.definedTypes, nil).(*ast.FuncType)
				pkgDecls.funcDecls[i] = replacedDecl
			}
		}
	}

	// 处理由replacer引起的名称冲突
	c.resolveReplacerConflicts(allNames)
}

// resolveReplacerConflicts 处理由replacer引起的名称冲突
func (c *Collector) resolveReplacerConflicts(allNames map[string]int) {
	// 创建一个映射来跟踪每个名称的计数
	nameCounters := make(map[string]int)

	// 辅助函数来生成唯一名称
	generateUniqueName := func(name string) string {
		count := nameCounters[name]
		nameCounters[name]++
		if count == 0 {
			return name
		}
		return name + strconv.Itoa(count)
	}

	// 更新名称的辅助函数
	updateValueSpecName := func(spec *ast.ValueSpec, newName string) *ast.ValueSpec {
		newSpec := &ast.ValueSpec{
			Doc:     spec.Doc,
			Names:   make([]*ast.Ident, len(spec.Names)),
			Type:    spec.Type,
			Values:  spec.Values,
			Comment: spec.Comment,
		}
		for i, name := range spec.Names {
			newSpec.Names[i] = &ast.Ident{
				NamePos: name.NamePos,
				Name:    newName,
				Obj:     name.Obj,
			}
		}
		return newSpec
	}

	updateTypeSpecName := func(spec *ast.TypeSpec, newName string) *ast.TypeSpec {
		newSpec := &ast.TypeSpec{
			Doc:        spec.Doc,
			Name:       &ast.Ident{Name: newName},
			Assign:     spec.Assign,
			TypeParams: spec.TypeParams, // Preserve type parameters for generic types
			Type:       spec.Type,
			Comment:    spec.Comment,
		}
		return newSpec
	}

	updateFuncDeclName := func(decl *ast.FuncDecl, newName string) *ast.FuncDecl {
		newDecl := &ast.FuncDecl{
			Doc:  decl.Doc,
			Recv: decl.Recv,
			Name: &ast.Ident{Name: newName},
			Type: decl.Type,
			Body: decl.Body,
		}
		return newDecl
	}

	// 遍历所有包声明，处理冲突
	for _, pkgDecls := range c.allPackageDecls {
		// 处理常量声明
		for i, decl := range pkgDecls.constDecls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				var newSpecs []ast.Spec
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, ident := range valueSpec.Names {
							uniqueName := generateUniqueName(ident.Name)
							if uniqueName != ident.Name {
								newSpec := updateValueSpecName(valueSpec, uniqueName)
								newSpecs = append(newSpecs, newSpec)
							} else {
								newSpecs = append(newSpecs, spec)
							}
						}
					}
				}
				genDecl.Specs = newSpecs
				pkgDecls.constDecls[i] = genDecl
			}
		}

		// 处理变量声明
		for i, decl := range pkgDecls.varDecls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				var newSpecs []ast.Spec
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, ident := range valueSpec.Names {
							uniqueName := generateUniqueName(ident.Name)
							if uniqueName != ident.Name {
								newSpec := updateValueSpecName(valueSpec, uniqueName)
								newSpecs = append(newSpecs, newSpec)
							} else {
								newSpecs = append(newSpecs, spec)
							}
						}
					}
				}
				genDecl.Specs = newSpecs
				pkgDecls.varDecls[i] = genDecl
			}
		}

		// 处理类型声明
		for i, spec := range pkgDecls.typeSpecs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				uniqueName := generateUniqueName(typeSpec.Name.Name)
				if uniqueName != typeSpec.Name.Name {
					newSpec := updateTypeSpecName(typeSpec, uniqueName)
					pkgDecls.typeSpecs[i] = newSpec
				}
			}
		}

		// 处理函数声明
		for i, decl := range pkgDecls.funcDecls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				uniqueName := generateUniqueName(funcDecl.Name.Name)
				if uniqueName != funcDecl.Name.Name {
					newDecl := updateFuncDeclName(funcDecl, uniqueName)
					pkgDecls.funcDecls[i] = newDecl
				}
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

		// Mark this path as processed.
		processedPaths[pkg.ImportPath] = true

		c.collectImports(sourcePkg)
		c.collectTypeDeclarations(sourcePkg, pkg.ImportAlias)
		c.collectOtherDeclarations(sourcePkg, pkg.ImportAlias)
	}

	if c.replacer != nil {
		c.applyReplacements()
	}

	return nil
}
