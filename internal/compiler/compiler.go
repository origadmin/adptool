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

	// Find and apply the highest priority rule
	if newName, ok := r.findAndApplyRule(ident.Name, ruleType); ok {
		ident.Name = newName
	}
}

func (r *realReplacer) applyGenDeclRule(ctx interfaces.Context, decl *ast.GenDecl) {
	switch decl.Tok {
	case token.CONST:
		ctx = ctx.Push("const")
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				nameCtx := ctx.Push("const_decl").Push("const_decl_name")
				for _, name := range valueSpec.Names {
					r.Apply(nameCtx, name)
				}
			}
		}
	case token.VAR:
		ctx = ctx.Push("var")
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				nameCtx := ctx.Push("var_decl").Push("var_decl_name")
				for _, name := range valueSpec.Names {
					r.Apply(nameCtx, name)
				}
			}
		}
	case token.TYPE:
		ctx = ctx.Push("type_decl")
		for _, spec := range decl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				r.Apply(ctx.Push("type"), typeSpec.Name)
			}
		}
	case token.IMPORT:
		// Not handling imports for now
	}
}

func (r *realReplacer) applyFuncDeclRule(ctx interfaces.Context, decl *ast.FuncDecl) {
	r.Apply(ctx.Push("func"), decl.Name)
}

func (r *realReplacer) applyTypeSpecRule(ctx interfaces.Context, spec *ast.TypeSpec) {
	r.Apply(ctx.Push("type"), spec.Name)
}

func (r *realReplacer) findAndApplyRule(name, ruleType string) (string, bool) {
	// 1. Check for specific priority rules
	if rules, ok := r.config.PriorityRules[name]; ok {
		if newName, applied := applyFilteredRules(name, rules, ruleType); applied {
			return newName, true
		}
	}

	// 2. Check for wildcard priority rules
	if rules, ok := r.config.PriorityRules["*"]; ok {
		if newName, applied := applyFilteredRules(name, rules, ruleType); applied {
			return newName, true
		}
	}

	// 3. Fallback to old rule structure
	if rules, ok := r.config.Rules[name]; ok {
		var priorityRules []interfaces.PriorityRule
		for _, rule := range rules {
			priorityRules = append(priorityRules, interfaces.PriorityRule{Rule: rule})
		}
		if newName, applied := applyFilteredRules(name, priorityRules, ruleType); applied {
			return newName, true
		}
	}

	return "", false
}

func applyFilteredRules(name string, rules []interfaces.PriorityRule, ruleType string) (string, bool) {
	filtered := filterRulesByContext(rules, ruleType)
	if len(filtered) == 0 {
		return "", false
	}

	// Apply highest priority rule
	highestPriorityRule := filtered[0].Rule
	newName, err := rulesPkg.ApplyRules(name, []interfaces.RenameRule{highestPriorityRule})
	if err != nil {
		return "", false
	}
	return newName, newName != name
}

func filterRulesByContext(rules []interfaces.PriorityRule, ruleType string) []interfaces.PriorityRule {
	var filtered []interfaces.PriorityRule
	for _, r := range rules {
		if (ruleType == "const_decl_name" && r.Rule.Value == "Const") ||
			(ruleType == "type" && r.Rule.Value == "Type") ||
			(ruleType == "var_decl_name" && r.Rule.Value == "Var") ||
			(ruleType == "func" && r.Rule.Value == "Func") {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func isApplicableRuleType(ruleType string) bool {
	switch ruleType {
	case "const_decl_name", "type", "var_decl_name", "func":
		return true
	default:
		return false
	}
}

// Compile takes a configuration and returns a compiled representation of it.
func Compile(cfg *config.Config) (*interfaces.CompiledConfig, error) {
	priorityRules := make(map[string][]internalPriorityRule)

	// Process global rules (priority 0)
	processRuleHolders(priorityRules, cfg.GetGlobalRules(), 0, "")

	// Process package-specific rules (priority 1)
	for _, pkg := range cfg.Packages {
		processRuleHolders(priorityRules, pkg.GetPackageRules(), 1, pkg.Import)
	}

	// Process type-specific rules (priority 2)
	for _, pkg := range cfg.Packages {
		for _, t := range pkg.Types {
			if t.Fields != nil {
				for _, field := range t.Fields {
					processRuleHolder(priorityRules, field, 2, pkg.Import)
				}
			}
			if t.Methods != nil {
				for _, method := range t.Methods {
					processRuleHolder(priorityRules, method, 2, pkg.Import)
				}
			}
		}
	}

	sortPriorityRules(priorityRules)

	compiledPackages := compilePackages(cfg.Packages)

	compiledCfg := &interfaces.CompiledConfig{
		PackageName:   cfg.OutputPackageName,
		Packages:      compiledPackages,
		Rules:         convertPriorityToLegacy(priorityRules),
		PriorityRules: convertToExternalPriorityRules(priorityRules),
	}

	if compiledCfg.PackageName == "" {
		compiledCfg.PackageName = "adapters"
	}

	return compiledCfg, nil
}

// --- Helper types and functions for compilation ---

type internalPriorityRule struct {
	rule        interfaces.RenameRule
	priority    int
	packageName string
	isWildcard  bool
}

func processRuleHolders(priorityRules map[string][]internalPriorityRule, holders []config.RuleHolder, priority int, pkgName string) {
	for _, holder := range holders {
		processRuleHolder(priorityRules, holder, priority, pkgName)
	}
}

func processRuleHolder(priorityRules map[string][]internalPriorityRule, holder config.RuleHolder, priority int, pkgName string) {
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
				Rule:     rule.rule,
				Priority: rule.priority,
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
