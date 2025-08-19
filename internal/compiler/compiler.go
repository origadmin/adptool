package compiler

import (
	"fmt" // Re-added fmt import for error messages in applyRules
	"go/ast"
	"log"
	"path"   // Added for path.Base
	"regexp" // Re-added regexp import for applyRules

	"github.com/origadmin/adptool/internal/config"
	rulesPkg "github.com/origadmin/adptool/internal/rules"
)

// compiledRules stores the transformation rules compiled from the config.
// It maps original names to a slice of rulesPkg.RenameRule.
type compiledRules map[string][]rulesPkg.RenameRule

// realReplacer implements the interfaces.Replacer interface
// and applies actual transformation rules based on the compiled configuration.
type realReplacer struct {
	rules compiledRules
}

// Apply applies the transformation rules to the given AST node.
func (r *realReplacer) Apply(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Ident:
		log.Printf("Apply: Processing identifier: %s", n.Name)
		if rulesToApply, ok := r.rules[n.Name]; ok {
			log.Printf("Apply: Found rules for %s: %+v", n.Name, rulesToApply)
			// Apply rules to the identifier's name
			newName, err := applyRules(n.Name, rulesToApply) // Use applyRules
			if err != nil {
				log.Printf("Error applying rules to identifier %s: %v", n.Name, err)
				return node // Return original node on error
			}
			if newName != n.Name {
				n.Name = newName // Modify the identifier's name
				log.Printf("Transformed identifier %s to %s", node, newName)
			}
		} else {
			log.Printf("Apply: No rules found for identifier: %s", n.Name)
		}
	// TODO: Add more cases for other AST node types if needed (e.g., *ast.FuncDecl, *ast.TypeSpec)
	// For now, we focus on ast.Ident as it's the most common target for renaming.
	default:
		// log.Printf("Real replacer applied to node: %T", node) // Too verbose
	}
	return node
}

// Compile takes a configuration and returns a compiled representation of it.
func Compile(cfg *config.Config) (*config.CompiledConfig, error) {
	log.Printf("Compile: Received config: %+v", cfg)

	// --- Compile Renaming Rules ---
	rules := make(compiledRules)
	processConfigRules := func(cfgRules interface {
		IsDisabled() bool
		GetName() string
		GetRuleSet() *config.RuleSet
	}) {
		if cfgRules.IsDisabled() {
			return
		}
		name := cfgRules.GetName()
		ruleSet := cfgRules.GetRuleSet()
		log.Printf("Compile: Processing rule for %s, RuleSet: %+v", name, ruleSet)
		if ruleSet != nil {
			rules[name] = rulesPkg.ConvertRuleSetToRenameRules(ruleSet)
		}
	}

	for _, t := range cfg.Types {
		processConfigRules(t)
	}
	for _, f := range cfg.Functions {
		processConfigRules(f)
	}
	for _, v := range cfg.Variables {
		processConfigRules(v)
	}
	for _, c := range cfg.Constants {
		processConfigRules(c)
	}

	for _, pkg := range cfg.Packages {
		for _, t := range pkg.Types {
			processConfigRules(t)
		}
		for _, f := range pkg.Functions {
			processConfigRules(f)
		}
		for _, v := range pkg.Variables {
			processConfigRules(v)
		}
		for _, c := range pkg.Constants {
			processConfigRules(c)
		}
	}
	replacer := &realReplacer{rules: rules}

	// --- Compile Package Information ---
	var compiledPackages []*config.CompiledPackage
	for _, pkg := range cfg.Packages {
		finalAlias := pkg.Alias
		if finalAlias == "" {
			finalAlias = path.Base(pkg.Import)
		}
		compiledPackages = append(compiledPackages, &config.CompiledPackage{
			ImportPath:  pkg.Import,
			ImportAlias: finalAlias,
		})
	}

	// --- Assemble Compiled Config ---
	compiledCfg := &config.CompiledConfig{
		PackageName: cfg.OutputPackageName,
		Packages:    compiledPackages,
		Replacer:    replacer,
	}

	// Use a default output package name if not provided
	if compiledCfg.PackageName == "" {
		compiledCfg.PackageName = "adapters"
	}

	log.Printf("Successfully compiled config: %+v", compiledCfg)
	return compiledCfg, nil
}

// applyRules applies a set of rename rules to a given name and returns the result.
// This function is copied from internal/generator/generator.go for now.
// Ideally, this logic would be in a shared utility package or the compiler would
// directly implement the rule application without copying.
func applyRules(name string, rules []rulesPkg.RenameRule) (string, error) {
	currentName := name
	for _, rule := range rules {
		switch rule.Type {
		case "explicit":
			if name == rule.From {
				return rule.To, nil // Explicit rule is final
			}
		case "prefix":
			currentName = rule.Value + currentName
		case "suffix":
			currentName = currentName + rule.Value
		case "regex":
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return "", fmt.Errorf("invalid regex pattern '%s': %w", rule.Pattern, err)
			}
			currentName = re.ReplaceAllString(currentName, rule.Replace)
		}
	}
	return currentName, nil
}
