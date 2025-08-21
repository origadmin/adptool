package compiler

import (
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
		nameCtx := ctx.Push(interfaces.RuleTypeType)
		for _, spec := range decl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				pkgImportPath := ""
				nameToFindRuleFor := typeSpec.Name.Name

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
				}

				if newName, ok := r.findAndApplyRule(nameToFindRuleFor, interfaces.RuleTypeType, pkgImportPath); ok {
					typeSpec.Name.Name = newName
				} else {
					r.Apply(nameCtx, typeSpec.Name)
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

	// 根据规则类型直接查找对应类型的规则，提高效率
	// 首先查找包级别的规则
	if pkgName != "" {
		switch ruleType {
		case interfaces.RuleTypeType:
			if r.config.PackageTypeRules != nil {
				if rules, ok := r.config.PackageTypeRules[pkgName][name]; ok {
					applicableRules = append(applicableRules, rules...)
				}
				if rules, ok := r.config.PackageTypeRules[pkgName]["*"]; ok {
					applicableRules = append(applicableRules, rules...)
				}
			}
		case interfaces.RuleTypeFunc:
			if r.config.PackageFuncRules != nil {
				if rules, ok := r.config.PackageFuncRules[pkgName][name]; ok {
					applicableRules = append(applicableRules, rules...)
				}
				if rules, ok := r.config.PackageFuncRules[pkgName]["*"]; ok {
					applicableRules = append(applicableRules, rules...)
				}
			}
		case interfaces.RuleTypeVar:
			if r.config.PackageVarRules != nil {
				if rules, ok := r.config.PackageVarRules[pkgName][name]; ok {
					applicableRules = append(applicableRules, rules...)
				}
				if rules, ok := r.config.PackageVarRules[pkgName]["*"]; ok {
					applicableRules = append(applicableRules, rules...)
				}
			}
		case interfaces.RuleTypeConst:
			if r.config.PackageConstRules != nil {
				if rules, ok := r.config.PackageConstRules[pkgName][name]; ok {
					applicableRules = append(applicableRules, rules...)
				}
				if rules, ok := r.config.PackageConstRules[pkgName]["*"]; ok {
					applicableRules = append(applicableRules, rules...)
				}
			}
		}
	}

	// 如果没有找到包级别的规则或者没有指定包，则查找全局规则
	if len(applicableRules) == 0 {
		switch ruleType {
		case interfaces.RuleTypeType:
			if rules, ok := r.config.TypeRules[name]; ok {
				applicableRules = append(applicableRules, rules...)
			}
			if rules, ok := r.config.TypeRules["*"]; ok {
				applicableRules = append(applicableRules, rules...)
			}
		case interfaces.RuleTypeFunc:
			if rules, ok := r.config.FuncRules[name]; ok {
				applicableRules = append(applicableRules, rules...)
			}
			if rules, ok := r.config.FuncRules["*"]; ok {
				applicableRules = append(applicableRules, rules...)
			}
		case interfaces.RuleTypeVar:
			if rules, ok := r.config.VarRules[name]; ok {
				applicableRules = append(applicableRules, rules...)
			}
			if rules, ok := r.config.VarRules["*"]; ok {
				applicableRules = append(applicableRules, rules...)
			}
		case interfaces.RuleTypeConst:
			if rules, ok := r.config.ConstRules[name]; ok {
				applicableRules = append(applicableRules, rules...)
			}
			if rules, ok := r.config.ConstRules["*"]; ok {
				applicableRules = append(applicableRules, rules...)
			}
		}
	}

	// 过滤匹配包名的规则（主要用于全局规则）
	var filtered []interfaces.PriorityRule
	for _, rule := range applicableRules {
		// 如果是包级别的规则，已经经过筛选，直接添加
		if pkgName != "" && isPackageRule(rule, pkgName) {
			filtered = append(filtered, rule)
			continue
		}
		
		// 对于全局规则，检查包名匹配
		if pkgName != "" {
			if rule.PackageName == pkgName || rule.PackageName == "" {
				filtered = append(filtered, rule)
			}
		} else {
			if rule.PackageName == "" {
				filtered = append(filtered, rule)
			}
		}
	}

	if len(filtered) == 0 {
		return "", false
	}

	// 应用优先级最高的规则
	highestPriorityRule := filtered[0].Rule
	newName, err := rulesPkg.ApplyRules(name, []interfaces.RenameRule{highestPriorityRule})
	if err != nil {
		return "", false
	}
	return newName, newName != name
}

// isPackageRule 检查规则是否为指定包的规则
func isPackageRule(rule interfaces.PriorityRule, pkgName string) bool {
	// 包级别的规则存储在Package*Rules中，其PackageName应与pkgName匹配
	return rule.PackageName == pkgName
}

func filterRulesByContext(rules []interfaces.PriorityRule, ruleType interfaces.RuleType, pkgName string) []interfaces.PriorityRule {
	var filtered []interfaces.PriorityRule
	for _, r := range rules {
		isCorrectType := (ruleType == interfaces.RuleTypeConst && r.Rule.RuleType == interfaces.RuleTypeConst) ||
			(ruleType == interfaces.RuleTypeType && r.Rule.RuleType == interfaces.RuleTypeType) ||
			(ruleType == interfaces.RuleTypeVar && r.Rule.RuleType == interfaces.RuleTypeVar) ||
			(ruleType == interfaces.RuleTypeFunc && r.Rule.RuleType == interfaces.RuleTypeFunc)

		if !isCorrectType {
			continue
		}

		if pkgName != "" {
			if r.PackageName == pkgName || r.PackageName == "" {
				filtered = append(filtered, r)
			}
		} else {
			if r.PackageName == "" {
				filtered = append(filtered, r)
			}
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Priority != filtered[j].Priority {
			return filtered[i].Priority > filtered[j].Priority
		}
		if filtered[i].PackageName != filtered[j].PackageName {
			return filtered[i].PackageName != ""
		}
		return false
	})

	return filtered
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
		process(r, 0, "", interfaces.RuleTypeType)
	}
	for _, r := range cfg.Functions {
		process(r, 0, "", interfaces.RuleTypeFunc)
	}
	for _, r := range cfg.Variables {
		process(r, 0, "", interfaces.RuleTypeVar)
	}
	for _, r := range cfg.Constants {
		process(r, 0, "", interfaces.RuleTypeConst)
	}

	for _, pkg := range cfg.Packages {
		for _, r := range pkg.Types {
			process(r, 1, pkg.Import, interfaces.RuleTypeType)
		}
		for _, r := range pkg.Functions {
			process(r, 1, pkg.Import, interfaces.RuleTypeFunc)
		}
		for _, r := range pkg.Variables {
			process(r, 1, pkg.Import, interfaces.RuleTypeVar)
		}
		for _, r := range pkg.Constants {
			process(r, 1, pkg.Import, interfaces.RuleTypeConst)
		}

		for _, t := range pkg.Types {
			if t.Fields != nil {
				for _, field := range t.Fields {
					process(field, 2, pkg.Import, interfaces.RuleTypeVar)
				}
			}
			if t.Methods != nil {
				for _, method := range t.Methods {
					process(method, 2, pkg.Import, interfaces.RuleTypeFunc)
				}
			}
		}
	}

	sortPriorityRules(priorityRules)
	
	// 将规则按类型分类
	categorizeRules(priorityRules, typeRules, funcRules, varRules, constRules)
	
	// 将规则按包和类型分类
	categorizePackageRules(priorityRules, packageTypeRules, packageFuncRules, packageVarRules, packageConstRules)

	compiledPackages := compilePackages(cfg.Packages)

	compiledCfg := &interfaces.CompiledConfig{
		PackageName:   cfg.OutputPackageName,
		Packages:      compiledPackages,
		Rules:         convertPriorityToLegacy(priorityRules),
		PriorityRules: convertToExternalPriorityRules(priorityRules),
		TypeRules:     typeRules,
		FuncRules:     funcRules,
		VarRules:      varRules,
		ConstRules:    constRules,
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
	for name, rules := range priorityRules {
		for _, rule := range rules {
			externalRule := interfaces.PriorityRule{
				Rule:        rule.rule,
				Priority:    rule.priority,
				PackageName: rule.packageName,
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
	sortCategorizedRules(typeRules)
	sortCategorizedRules(funcRules)
	sortCategorizedRules(varRules)
	sortCategorizedRules(constRules)
}

// categorizePackageRules 将规则按包和类型分类存储，提高运行时效率
func categorizePackageRules(priorityRules map[string][]internalPriorityRule, 
	packageTypeRules, packageFuncRules, packageVarRules, packageConstRules map[string]map[string][]interfaces.PriorityRule) {
	for name, rules := range priorityRules {
		for _, rule := range rules {
			// 只处理包级别的规则
			if rule.packageName == "" {
				continue
			}
			
			externalRule := interfaces.PriorityRule{
				Rule:        rule.rule,
				Priority:    rule.priority,
				PackageName: rule.packageName,
			}
			
			// 确保包的map已初始化
			ensurePackageMapInitialized(packageTypeRules, packageFuncRules, packageVarRules, packageConstRules, rule.packageName)
			
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
	
	// 对每种类型的规则进行排序
	for _, rules := range packageTypeRules {
		sortCategorizedRules(rules)
	}
	for _, rules := range packageFuncRules {
		sortCategorizedRules(rules)
	}
	for _, rules := range packageVarRules {
		sortCategorizedRules(rules)
	}
	for _, rules := range packageConstRules {
		sortCategorizedRules(rules)
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

// sortCategorizedRules 对分类后的规则进行排序
func sortCategorizedRules(rules map[string][]interfaces.PriorityRule) {
	for name, prules := range rules {
		sort.Slice(prules, func(i, j int) bool {
			if prules[i].Priority != prules[j].Priority {
				return prules[i].Priority > prules[j].Priority
			}
			if prules[i].PackageName != prules[j].PackageName {
				return prules[i].PackageName != ""
			}
			return false
		})
		rules[name] = prules
	}
}
