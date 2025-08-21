package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"sort"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/interfaces"
	rulesPkg "github.com/origadmin/adptool/internal/rules"
)

// realReplacer implements the interfaces.Replacer interface
// and applies actual transformation rules based on the compiled configuration.
type realReplacer struct {
	config         *interfaces.CompiledConfig
	packageAliases map[string]bool
	processedNodes map[ast.Node]bool
}

// NewReplacer creates a new Replacer instance from a compiled configuration.
func NewReplacer(compiledCfg *interfaces.CompiledConfig) interfaces.Replacer {
	if compiledCfg == nil {
		return nil
	}

	packageAliases := make(map[string]bool)
	for _, pkg := range compiledCfg.Packages {
		packageAliases[pkg.ImportAlias] = true
	}

	return &realReplacer{
		config:         compiledCfg,
		packageAliases: packageAliases,
		processedNodes: make(map[ast.Node]bool),
	}
}

// Apply applies the transformation rules to the given AST node.
func (r *realReplacer) Apply(ctx interfaces.Context, node ast.Node) ast.Node {
	if r.processedNodes[node] {
		return node
	}
	r.processedNodes[node] = true

	switch n := node.(type) {
	case *ast.Ident:
		r.applyIdentRule(ctx, n)
	case *ast.GenDecl:
		r.applyGenDeclRule(ctx, n)
	case *ast.FuncDecl:
		r.applyFuncDeclRule(ctx, n)
	case *ast.TypeSpec:
		r.applyTypeSpecRule(ctx, n)
	}
	return node
}

func (r *realReplacer) applyIdentRule(ctx interfaces.Context, ident *ast.Ident) {
	if r.packageAliases != nil && r.packageAliases[ident.Name] {
		return
	}

	ruleType := ctx.CurrentNodeType()
	if !isApplicableRuleType(ruleType) {
		return
	}

	if newName, ok := r.findAndApplyRule(ident.Name, ruleType, ""); ok {
		ident.Name = newName
	}
}

func (r *realReplacer) applyGenDeclRule(ctx interfaces.Context, decl *ast.GenDecl) {
	switch decl.Tok {
	case token.CONST:
		nameCtx := ctx.Push(interfaces.RuleTypeConst)
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					r.Apply(nameCtx, name)
				}
			}
		}
	case token.VAR:
		nameCtx := ctx.Push(interfaces.RuleTypeVar)
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					r.Apply(nameCtx, name)
				}
			}
		}
	case token.TYPE:
		//nameCtx := ctx.Push(interfaces.RuleTypeType)
		for _, spec := range decl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				pkgImportPath := ""
				nameToFindRuleFor := ""

				if selExpr, ok := typeSpec.Type.(*ast.SelectorExpr); ok {
					if pkgIdent, ok := selExpr.X.(*ast.Ident); ok {
						for _, p := range r.config.Packages {
							if p.ImportAlias == pkgIdent.Name {
								pkgImportPath = p.ImportPath
								break
							}
						}
					}
					nameToFindRuleFor = selExpr.Sel.Name
				} else {
					nameToFindRuleFor = typeSpec.Name.Name
				}

				if newName, ok := r.findAndApplyRule(nameToFindRuleFor, interfaces.RuleTypeType, pkgImportPath); ok {
					typeSpec.Name.Name = newName
				}
			}
		}
	case token.IMPORT:
		// Not handling imports for now
	}
}

func (r *realReplacer) applyFuncDeclRule(ctx interfaces.Context, decl *ast.FuncDecl) {
	r.Apply(ctx.Push(interfaces.RuleTypeFunc), decl.Name)
}

func (r *realReplacer) applyTypeSpecRule(ctx interfaces.Context, spec *ast.TypeSpec) {
	r.Apply(ctx.Push(interfaces.RuleTypeType), spec.Name)
}

func (r *realReplacer) findAndApplyRule(name string, ruleType interfaces.RuleType, pkgName string) (string, bool) {
	var applicableRules []interfaces.PriorityRule

	// Helper to collect specific and wildcard rules from a map
	collect := func(ruleMap map[string][]interfaces.PriorityRule) {
		if ruleMap == nil {
			return
		}
		if rules, ok := ruleMap[name]; ok {
			applicableRules = append(applicableRules, rules...)
		}
		if rules, ok := ruleMap["*"]; ok {
			applicableRules = append(applicableRules, rules...)
		}
	}

	// 1. Collect all applicable rules (package and global)
	switch ruleType {
	case interfaces.RuleTypeType:
		if pkgName != "" {
			collect(r.config.PackageTypeRules[pkgName])
		}
		collect(r.config.TypeRules)
	case interfaces.RuleTypeFunc:
		if pkgName != "" {
			collect(r.config.PackageFuncRules[pkgName])
		}
		collect(r.config.FuncRules)
	case interfaces.RuleTypeVar:
		if pkgName != "" {
			collect(r.config.PackageVarRules[pkgName])
		}
		collect(r.config.VarRules)
	case interfaces.RuleTypeConst:
		if pkgName != "" {
			collect(r.config.PackageConstRules[pkgName])
		}
		collect(r.config.ConstRules)
	}

	if len(applicableRules) == 0 {
		return "", false
	}

	// 2. Sort the rules to find the one with the highest priority
	sort.Slice(applicableRules, func(i, j int) bool {
		r1 := applicableRules[i]
		r2 := applicableRules[j]
		if r1.Priority != r2.Priority {
			return r1.Priority > r2.Priority
		}
		if r1.IsWildcard != r2.IsWildcard {
			return !r1.IsWildcard // Non-wildcard rules take precedence
		}
		if (r1.PackageName != "") != (r2.PackageName != "") {
			return r1.PackageName != "" // Package-specific rules take precedence
		}
		return false // Keep original order if equal
	})

	// 3. Apply the highest-priority rule
	highestPriorityRule := applicableRules[0].Rule
	newName, err := rulesPkg.ApplyRules(name, []interfaces.RenameRule{highestPriorityRule})
	if err != nil {
		return "", false
	}
	return newName, newName != name
}

func isApplicableRuleType(ruleType interfaces.RuleType) bool {
	switch ruleType {
	case interfaces.RuleTypeConst, interfaces.RuleTypeType, interfaces.RuleTypeVar, interfaces.RuleTypeFunc:
		return true
	default:
		return false
	}
}

// Compile takes a configuration and returns a compiled representation of it.
func Compile(cfg *config.Config) (*interfaces.CompiledConfig, error) {
	fmt.Printf("Compiling configuration with %d global type rules, %d global function rules, %d global variable rules, %d global constant rules\n",
		len(cfg.Types), len(cfg.Functions), len(cfg.Variables), len(cfg.Constants))
	priorityRules := make(map[string][]internalPriorityRule)

	// 按类型分类的全局规则
	typeRules := make(map[string][]interfaces.PriorityRule)
	funcRules := make(map[string][]interfaces.PriorityRule)
	varRules := make(map[string][]interfaces.PriorityRule)
	constRules := make(map[string][]interfaces.PriorityRule)

	// 按包和类型分类的规则
	packageTypeRules := make(map[string]map[string][]interfaces.PriorityRule)
	packageFuncRules := make(map[string]map[string][]interfaces.PriorityRule)
	packageVarRules := make(map[string]map[string][]interfaces.PriorityRule)
	packageConstRules := make(map[string]map[string][]interfaces.PriorityRule)

	process := func(holder config.RuleHolder, priority int, pkgName string, ruleType interfaces.RuleType) {
		processRuleHolder(priorityRules, holder, priority, pkgName, ruleType)
	}

	for _, r := range cfg.Types {
		fmt.Printf("Processing global type rule: %s\n", r.GetName())
		process(r, 0, "", interfaces.RuleTypeType)
	}
	for _, r := range cfg.Functions {
		fmt.Printf("Processing global function rule: %s\n", r.GetName())
		process(r, 0, "", interfaces.RuleTypeFunc)
	}
	for _, r := range cfg.Variables {
		fmt.Printf("Processing global variable rule: %s\n", r.GetName())
		process(r, 0, "", interfaces.RuleTypeVar)
	}
	for _, r := range cfg.Constants {
		fmt.Printf("Processing global constant rule: %s\n", r.GetName())
		process(r, 0, "", interfaces.RuleTypeConst)
	}

	for _, pkg := range cfg.Packages {
		fmt.Printf("Processing package: %s (alias: %s)\n", pkg.Import, pkg.Alias)

		for _, r := range pkg.Types {
			fmt.Printf("  Processing package type rule: %s\n", r.GetName())
			process(r, 1, pkg.Import, interfaces.RuleTypeType)
		}
		for _, r := range pkg.Functions {
			fmt.Printf("  Processing package function rule: %s\n", r.GetName())
			process(r, 1, pkg.Import, interfaces.RuleTypeFunc)
		}
		for _, r := range pkg.Variables {
			fmt.Printf("  Processing package variable rule: %s\n", r.GetName())
			process(r, 1, pkg.Import, interfaces.RuleTypeVar)
		}
		for _, r := range pkg.Constants {
			fmt.Printf("  Processing package constant rule: %s\n", r.GetName())
			process(r, 1, pkg.Import, interfaces.RuleTypeConst)
		}

		for _, t := range pkg.Types {
			if t.Fields != nil {
				for _, field := range t.Fields {
					fmt.Printf("    Processing type field rule: %s\n", field.GetName())
					process(field, 2, pkg.Import, interfaces.RuleTypeVar)
				}
			}
			if t.Methods != nil {
				for _, method := range t.Methods {
					fmt.Printf("    Processing type method rule: %s\n", method.GetName())
					process(method, 2, pkg.Import, interfaces.RuleTypeFunc)
				}
			}
		}
	}

	sortPriorityRules(priorityRules)

	// 将规则按类型分类
	fmt.Println("Categorizing rules by type...")
	categorizeRules(priorityRules, typeRules, funcRules, varRules, constRules)

	// 将规则按包和类型分类
	fmt.Println("Categorizing rules by package and type...")
	categorizePackageRules(priorityRules, packageTypeRules, packageFuncRules, packageVarRules, packageConstRules)

	compiledPackages := compilePackages(cfg.Packages)

	compiledCfg := &interfaces.CompiledConfig{
		PackageName:       cfg.OutputPackageName,
		Packages:          compiledPackages,
		Rules:             convertPriorityToLegacy(priorityRules),
		PriorityRules:     convertToExternalPriorityRules(priorityRules),
		TypeRules:         typeRules,
		FuncRules:         funcRules,
		VarRules:          varRules,
		ConstRules:        constRules,
		PackageTypeRules:  packageTypeRules,
		PackageFuncRules:  packageFuncRules,
		PackageVarRules:   packageVarRules,
		PackageConstRules: packageConstRules,
	}

	if compiledCfg.PackageName == "" {
		compiledCfg.PackageName = "adapters"
	}

	return compiledCfg, nil
}

type internalPriorityRule struct {
	rule        interfaces.RenameRule
	priority    int
	packageName string
	isWildcard  bool
}

func processRuleHolder(priorityRules map[string][]internalPriorityRule, holder config.RuleHolder, priority int, pkgName string, ruleType interfaces.RuleType) {
	if holder.IsDisabled() {
		return
	}
	name := holder.GetName()
	ruleSet := holder.GetRuleSet()
	if ruleSet == nil {
		return
	}

	renameRules := rulesPkg.ConvertRuleSetToRenameRules(ruleSet)
	isWildcard := name == "*"
	for _, rule := range renameRules {
		// Set the RuleType for the rule
		rule.RuleType = ruleType
		priorityRules[name] = append(priorityRules[name], internalPriorityRule{
			rule:        rule,
			priority:    priority,
			packageName: pkgName,
			isWildcard:  isWildcard,
		})
	}
}

func sortPriorityRules(rules map[string][]internalPriorityRule) {
	for name, prules := range rules {
		sort.Slice(prules, func(i, j int) bool {
			if prules[i].priority != prules[j].priority {
				return prules[i].priority > prules[j].priority
			}
			if prules[i].isWildcard != prules[j].isWildcard {
				return !prules[i].isWildcard
			}
			if prules[i].packageName != prules[j].packageName {
				return prules[i].packageName != ""
			}
			return false
		})
		rules[name] = prules
	}
}

func compilePackages(pkgs []*config.Package) []*interfaces.CompiledPackage {
	var compiledPackages []*interfaces.CompiledPackage
	for _, pkg := range pkgs {
		finalAlias := pkg.Alias
		if finalAlias == "" {
			finalAlias = path.Base(pkg.Import)
		}
		compiledPackages = append(compiledPackages, &interfaces.CompiledPackage{
			ImportPath:  pkg.Import,
			ImportAlias: finalAlias,
		})
	}
	return compiledPackages
}

func convertToExternalPriorityRules(prules map[string][]internalPriorityRule) map[string][]interfaces.PriorityRule {
	result := make(map[string][]interfaces.PriorityRule)
	for name, rules := range prules {
		converted := make([]interfaces.PriorityRule, len(rules))
		for i, rule := range rules {
			converted[i] = interfaces.PriorityRule{
				Rule:        rule.rule,
				Priority:    rule.priority,
				PackageName: rule.packageName,
				IsWildcard:  rule.isWildcard,
			}
		}
		result[name] = converted
	}
	return result
}

func convertPriorityToLegacy(prules map[string][]internalPriorityRule) map[string][]interfaces.RenameRule {
	legacyRules := make(map[string][]interfaces.RenameRule)
	for name, rules := range prules {
		for _, r := range rules {
			legacyRules[name] = append(legacyRules[name], r.rule)
		}
	}
	return legacyRules
}

// categorizeRules 将规则按类型分类存储，提高运行时效率
func categorizeRules(priorityRules map[string][]internalPriorityRule, typeRules, funcRules, varRules, constRules map[string][]interfaces.PriorityRule) {
	fmt.Println("Starting to categorize global rules by type")
	for name, rules := range priorityRules {
		for _, rule := range rules {
			// 只处理全局规则
			if rule.packageName != "" {
				continue
			}
			externalRule := interfaces.PriorityRule{
				Rule:        rule.rule,
				Priority:    rule.priority,
				PackageName: rule.packageName,
				IsWildcard:  rule.isWildcard,
			}

			switch rule.rule.RuleType {
			case interfaces.RuleTypeType:
				typeRules[name] = append(typeRules[name], externalRule)
			case interfaces.RuleTypeFunc:
				funcRules[name] = append(funcRules[name], externalRule)
			case interfaces.RuleTypeVar:
				varRules[name] = append(varRules[name], externalRule)
			case interfaces.RuleTypeConst:
				constRules[name] = append(constRules[name], externalRule)
			}
		}
	}

	// 对每种类型的规则进行排序
	sortRules(typeRules)
	sortRules(funcRules)
	sortRules(varRules)
	sortRules(constRules)
}

// categorizePackageRules 将规则按包和类型分类存储，提高运行时效率
func categorizePackageRules(priorityRules map[string][]internalPriorityRule,
	packageTypeRules, packageFuncRules, packageVarRules, packageConstRules map[string]map[string][]interfaces.PriorityRule) {
	fmt.Println("Starting to categorize rules by package and type")
	// Track processed packages to avoid redundant initialization and sorting
	processedPackages := make(map[string]bool)

	for name, rules := range priorityRules {
		for _, rule := range rules {
			// 只处理包级别的规则，并且确保规则属于当前正在处理的包
			if rule.packageName == "" {
				continue
			}

			externalRule := interfaces.PriorityRule{
				Rule:        rule.rule,
				Priority:    rule.priority,
				PackageName: rule.packageName,
				IsWildcard:  rule.isWildcard,
			}

			// 初始化包的map（仅在首次遇到该包时）
			if !processedPackages[rule.packageName] {
				ensurePackageMapInitialized(packageTypeRules, packageFuncRules, packageVarRules, packageConstRules, rule.packageName)
				processedPackages[rule.packageName] = true
			}

			switch rule.rule.RuleType {
			case interfaces.RuleTypeType:
				packageTypeRules[rule.packageName][name] = append(packageTypeRules[rule.packageName][name], externalRule)
			case interfaces.RuleTypeFunc:
				packageFuncRules[rule.packageName][name] = append(packageFuncRules[rule.packageName][name], externalRule)
			case interfaces.RuleTypeVar:
				packageVarRules[rule.packageName][name] = append(packageVarRules[rule.packageName][name], externalRule)
			case interfaces.RuleTypeConst:
				packageConstRules[rule.packageName][name] = append(packageConstRules[rule.packageName][name], externalRule)
			}
		}
	}

	// 在添加完规则后，对每种类型的规则进行排序（按包）
	for pkg := range packageTypeRules {
		sortRules(packageTypeRules[pkg])
	}
	for pkg := range packageFuncRules {
		sortRules(packageFuncRules[pkg])
	}
	for pkg := range packageVarRules {
		sortRules(packageVarRules[pkg])
	}
	for pkg := range packageConstRules {
		sortRules(packageConstRules[pkg])
	}
}

// ensurePackageMapInitialized 确保包的map已初始化
func ensurePackageMapInitialized(packageTypeRules, packageFuncRules, packageVarRules, packageConstRules map[string]map[string][]interfaces.PriorityRule, packageName string) {
	if _, ok := packageTypeRules[packageName]; !ok {
		packageTypeRules[packageName] = make(map[string][]interfaces.PriorityRule)
	}
	if _, ok := packageFuncRules[packageName]; !ok {
		packageFuncRules[packageName] = make(map[string][]interfaces.PriorityRule)
	}
	if _, ok := packageVarRules[packageName]; !ok {
		packageVarRules[packageName] = make(map[string][]interfaces.PriorityRule)
	}
	if _, ok := packageConstRules[packageName]; !ok {
		packageConstRules[packageName] = make(map[string][]interfaces.PriorityRule)
	}
}

// sortRules 对分类后的规则进行排序
func sortRules(rules map[string][]interfaces.PriorityRule) {
	for name, prules := range rules {
		sort.Slice(prules, func(i, j int) bool {
			if prules[i].Priority != prules[j].Priority {
				return prules[i].Priority > prules[j].Priority
			}
			if prules[i].IsWildcard != prules[j].IsWildcard {
				return !prules[i].IsWildcard // 非通配符优先
			}
			if prules[i].PackageName != prules[j].PackageName {
				return prules[i].PackageName != "" // 包规则优先
			}
			return false
		})
		rules[name] = prules
	}
}
