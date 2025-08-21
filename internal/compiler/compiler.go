package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"path"
	"regexp"
	"sort"
	"strings"

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
	case *ast.ValueSpec:
		r.applyValueSpecRule(ctx, n)
	}
	return node
}

func (r *realReplacer) applyValueSpecRule(ctx interfaces.Context, spec *ast.ValueSpec) {
	for _, ident := range spec.Names {
		r.Apply(ctx, ident)
	}
}

func (r *realReplacer) applyIdentRule(ctx interfaces.Context, ident *ast.Ident) {
	if r.packageAliases != nil && r.packageAliases[ident.Name] {
		return
	}

	ruleType := ctx.CurrentNodeType()
	if !isApplicableRuleType(ruleType) {
		return
	}

	// Get package path from context
	pkgPath, _ := ctx.Value(interfaces.PackagePathContextKey).(string)

	if newName, ok := r.findAndApplyRule(ident.Name, ruleType, pkgPath); ok {
		ident.Name = newName
	}
}

func (r *realReplacer) applyGenDeclRule(ctx interfaces.Context, decl *ast.GenDecl) {
	var ruleType interfaces.RuleType
	switch decl.Tok {
	case token.CONST:
		ruleType = interfaces.RuleTypeConst
	case token.VAR:
		ruleType = interfaces.RuleTypeVar
	case token.TYPE:
		ruleType = interfaces.RuleTypeType
	case token.IMPORT:
		return // Not handling imports for now
	default:
		return
	}

	// Create a new context with the appropriate rule type
	nameCtx := ctx.Push(ruleType)
	for _, spec := range decl.Specs {
		// The Apply function will dispatch to the correct handler for the spec type
		r.Apply(nameCtx, spec)
	}
}

func (r *realReplacer) applyFuncDeclRule(ctx interfaces.Context, decl *ast.FuncDecl) {
	// Apply rules to the function name
	r.Apply(ctx.Push(interfaces.RuleTypeFunc), decl.Name)
}

func (r *realReplacer) applyTypeSpecRule(ctx interfaces.Context, spec *ast.TypeSpec) {
	// Apply rules to the type name (Ident)
	r.Apply(ctx, spec.Name) // The context already has RuleTypeType from applyGenDeclRule
}

func (r *realReplacer) findAndApplyRule(name string, ruleType interfaces.RuleType, pkgName string) (string, bool) {
	var applicableRules []interfaces.CompiledRenameRule

	// Collect package-specific rules
	if pkgRules, ok := r.config.RulesByPackageAndType[pkgName]; ok {
		if rules, ok := pkgRules[ruleType]; ok {
			applicableRules = append(applicableRules, rules...)
		}
	}

	// Collect global rules
	if globalRules, ok := r.config.RulesByPackageAndType[""]; ok {
		if rules, ok := globalRules[ruleType]; ok {
			applicableRules = append(applicableRules, rules...)
		}
	}

	if len(applicableRules) == 0 {
		return "", false
	}

	// Rules are already sorted by priority during compilation.
	// We need to find the highest priority rule that applies to the current name.
	// For explicit rules, we prioritize exact matches over wildcards.
	for _, rule := range applicableRules {
		log.Printf("Considering rule: Type=%s, OriginalName=%s, Pattern=%s, IsWildcard=%t for name=%s",
			rule.Type, rule.OriginalName, rule.Pattern, rule.IsWildcard, name)
		if rule.Type == "explicit" {
			if rule.From == name || rule.From == "*" {
				// If it's an explicit rule, and it matches, it's the highest priority.
				// If there are multiple explicit rules, the one with higher priority (already sorted) or non-wildcard 'From' takes precedence.
				newName, err := rulesPkg.ApplyRules(name, []interfaces.CompiledRenameRule{rule})
				if err != nil {
					return "", false
				}
				return newName, newName != name
			}
		} else { // For prefix, suffix, regex rules
			// First, check if the rule's 'Name' (OriginalName) matches the current 'name'
			// This is the filtering step based on the rule's scope
			nameMatchesRuleScope := false
			if rule.IsWildcard { // Name is "*"
				nameMatchesRuleScope = true
			} else { // For prefix, suffix, and regex rules, OriginalName is the scope
				// Check if OriginalName is a regex pattern
				if strings.HasPrefix(rule.OriginalName, "^") && strings.HasSuffix(rule.OriginalName, "$") {
					// Attempt to compile OriginalName as a regex for matching
					// This assumes OriginalName is intended to be a regex for scope matching
					scopeRegex, err := regexp.Compile(rule.OriginalName)
					if err == nil && scopeRegex.MatchString(name) {
						nameMatchesRuleScope = true
					}
				} else if rule.OriginalName == name { // Literal match
					nameMatchesRuleScope = true
				}
			}

			if nameMatchesRuleScope {
				// Now, apply the transformation based on the rule's type
				newName, err := rulesPkg.ApplyRules(name, []interfaces.CompiledRenameRule{rule})
				if err != nil {
					return "", false
				}
				if newName != name {
					return newName, true
				}
			}
		}
	}

	return "", false
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


func processRule(holder config.RuleHolder, priority int, pkgName string, ruleType interfaces.RuleType) ([]interfaces.CompiledRenameRule, error) {
	if holder.IsDisabled() {
		return nil, nil
	}
	ruleSet := holder.GetRuleSet()
	if ruleSet == nil {
		return nil, nil
	}

	var compiledRules []interfaces.CompiledRenameRule
	isWildcard := holder.GetName() == "*"

	// Process explicit rules
	if len(ruleSet.Explicit) > 0 {
		for _, explicit := range ruleSet.Explicit {
			compiledRules = append(compiledRules, interfaces.CompiledRenameRule{
				Type:       "explicit",
				RuleType:   ruleType,
				From:       explicit.From,
				To:         explicit.To,
				Priority:   priority,
				IsWildcard: isWildcard,
			})
		}
		return compiledRules, nil // Explicit rules override all others
	}

	// Process regex rules
	if len(ruleSet.Regex) > 0 {
		for _, regex := range ruleSet.Regex {
			re, err := regexp.Compile(regex.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern '%s': %w", regex.Pattern, err)
			}
			compiledRules = append(compiledRules, interfaces.CompiledRenameRule{
				Type:          "regex",
				RuleType:      ruleType,
				OriginalName:  holder.GetName(),
				Pattern:       regex.Pattern,
				Replace:       regex.Replace,
				CompiledRegex: re,
				Priority:      priority,
				IsWildcard:    isWildcard,
			})
		}
		return compiledRules, nil // Regex rules override prefix/suffix
	}

	// Process prefix rule
	if ruleSet.Prefix != "" {
		compiledRules = append(compiledRules, interfaces.CompiledRenameRule{
			Type:       "prefix",
			RuleType:   ruleType,
			OriginalName: holder.GetName(),
			Value:      ruleSet.Prefix,
			Priority:   priority,
			IsWildcard: isWildcard,
		})
	}

	// Process suffix rule
	if ruleSet.Suffix != "" {
		compiledRules = append(compiledRules, interfaces.CompiledRenameRule{
			Type:       "suffix",
			RuleType:   ruleType,
			OriginalName: holder.GetName(),
			Value:      ruleSet.Suffix,
			Priority:   priority,
			IsWildcard: isWildcard,
		})
	}

	return compiledRules, nil
}

// Compile takes a configuration and returns a compiled representation of it.
func Compile(cfg *config.Config) (*interfaces.CompiledConfig, error) {
	compiledCfg := &interfaces.CompiledConfig{
		PackageName:           cfg.OutputPackageName,
		Packages:              compilePackages(cfg.Packages),
		RulesByPackageAndType: make(map[string]map[interfaces.RuleType][]interfaces.CompiledRenameRule),
	}

	// Helper to add rules to the main map and sort them
	addAndSortRules := func(pkgName string, rType interfaces.RuleType, rules []interfaces.CompiledRenameRule) {
		if _, ok := compiledCfg.RulesByPackageAndType[pkgName]; !ok {
			compiledCfg.RulesByPackageAndType[pkgName] = make(map[interfaces.RuleType][]interfaces.CompiledRenameRule)
		}
		compiledCfg.RulesByPackageAndType[pkgName][rType] = append(compiledCfg.RulesByPackageAndType[pkgName][rType], rules...)
		sort.Slice(compiledCfg.RulesByPackageAndType[pkgName][rType], func(i, j int) bool {
			r1 := compiledCfg.RulesByPackageAndType[pkgName][rType][i]
			r2 := compiledCfg.RulesByPackageAndType[pkgName][rType][j]
			if r1.Priority != r2.Priority {
				return r1.Priority > r2.Priority
			}
			if r1.IsWildcard != r2.IsWildcard {
				return !r1.IsWildcard // Non-wildcard rules take precedence
			}
			// For explicit rules, specific 'From' values take precedence over '*' wildcards
			if r1.Type == "explicit" && r2.Type == "explicit" {
				if r1.From == "*" && r2.From != "*" {
					return false // r2 (non-wildcard) takes precedence
				}
				if r1.From != "*" && r2.From == "*" {
					return true // r1 (non-wildcard) takes precedence
				}
			}
			return false // Keep original order if equal
		})
	}

	// Process global rules
	for _, r := range cfg.Types {
		rules, err := processRule(r, 0, "", interfaces.RuleTypeType)
		if err != nil {
			return nil, err
		}
		addAndSortRules("", interfaces.RuleTypeType, rules)
	}
	for _, r := range cfg.Functions {
		rules, err := processRule(r, 0, "", interfaces.RuleTypeFunc)
		if err != nil {
			return nil, err
		}
		addAndSortRules("", interfaces.RuleTypeFunc, rules)
	}
	for _, r := range cfg.Variables {
		rules, err := processRule(r, 0, "", interfaces.RuleTypeVar)
		if err != nil {
			return nil, err
		}
		addAndSortRules("", interfaces.RuleTypeVar, rules)
	}
	for _, r := range cfg.Constants {
		rules, err := processRule(r, 0, "", interfaces.RuleTypeConst)
		if err != nil {
			return nil, err
		}
		addAndSortRules("", interfaces.RuleTypeConst, rules)
	}

	// Process package-specific rules
	for _, pkg := range cfg.Packages {
		for _, r := range pkg.Types {
			rules, err := processRule(r, 1, pkg.Import, interfaces.RuleTypeType)
			if err != nil {
				return nil, err
			}
			addAndSortRules(pkg.Import, interfaces.RuleTypeType, rules)
		}
		for _, r := range pkg.Functions {
			rules, err := processRule(r, 1, pkg.Import, interfaces.RuleTypeFunc)
			if err != nil {
				return nil, err
			}
			addAndSortRules(pkg.Import, interfaces.RuleTypeFunc, rules)
		}
		for _, r := range pkg.Variables {
			rules, err := processRule(r, 1, pkg.Import, interfaces.RuleTypeVar)
			if err != nil {
				return nil, err
			}
			addAndSortRules(pkg.Import, interfaces.RuleTypeVar, rules)
		}
		for _, r := range pkg.Constants {
			rules, err := processRule(r, 1, pkg.Import, interfaces.RuleTypeConst)
			if err != nil {
				return nil, err
			}
			addAndSortRules(pkg.Import, interfaces.RuleTypeConst, rules)
		}

		for _, t := range pkg.Types {
			if t.Fields != nil {
				for _, field := range t.Fields {
					rules, err := processRule(field, 2, pkg.Import, interfaces.RuleTypeVar)
					if err != nil {
						return nil, err
					}
					addAndSortRules(pkg.Import, interfaces.RuleTypeVar, rules)
				}
			}
			if t.Methods != nil {
				for _, method := range t.Methods {
					rules, err := processRule(method, 2, pkg.Import, interfaces.RuleTypeFunc)
					if err != nil {
						return nil, err
					}
					addAndSortRules(pkg.Import, interfaces.RuleTypeFunc, rules)
				}
			}
		}
	}

	return compiledCfg, nil
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
