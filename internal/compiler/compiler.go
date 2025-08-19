package compiler

import (
	"fmt" // Re-added fmt import for error messages in applyRules
	"go/ast"
	"log"
	"path"   // Added for path.Base
	"regexp" // Re-added regexp import for applyRules
	"sort"

	"github.com/origadmin/adptool/internal/config"
	rulesPkg "github.com/origadmin/adptool/internal/rules"
)

// compiledRules stores the transformation rules compiled from the config.
// It maps original names to a slice of rulesPkg.RenameRule.
type compiledRules map[string][]rulesPkg.RenameRule

// priorityRule represents a rule with its priority level
type priorityRule struct {
	rule     rulesPkg.RenameRule
	priority int // 0: global, 1: package level, 2: type specific
}

// realReplacer implements the interfaces.Replacer interface
// and applies actual transformation rules based on the compiled configuration.
type realReplacer struct {
	rules        compiledRules
	priorityRules map[string][]priorityRule
}

// Apply applies the transformation rules to the given AST node.
func (r *realReplacer) Apply(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Ident:
		log.Printf("Apply: Processing identifier: %s", n.Name)
		
		// Check if we have priority rules for this identifier
		if priorityRulesToApply, ok := r.priorityRules[n.Name]; ok {
			log.Printf("Apply: Found priority rules for %s: %+v", n.Name, priorityRulesToApply)
			
			// 规则已经按优先级排序（最高优先级在前）
			// 我们只应用第一个匹配的规则（最高优先级）
			if len(priorityRulesToApply) > 0 {
				// 获取最高优先级的规则
				highestPriorityRule := priorityRulesToApply[0]
				log.Printf("Apply: Using highest priority rule for %s: %+v", n.Name, highestPriorityRule)
				
				// 只应用这一个规则
				newName, err := applyRules(n.Name, []rulesPkg.RenameRule{highestPriorityRule.rule})
				if err != nil {
					log.Printf("Error applying priority rule to identifier %s: %v", n.Name, err)
					return node // Return original node on error
				}
				
				if newName != n.Name {
					n.Name = newName // Modify the identifier's name
					log.Printf("Transformed identifier %s to %s using priority rule", node, newName)
				}
			}
		} else if rulesToApply, ok := r.rules[n.Name]; ok {
			// Fallback to the old rules if no priority rules exist
			log.Printf("Apply: Found rules for %s: %+v", n.Name, rulesToApply)
			
			// 为了保持一致性，我们也只应用第一个规则
			if len(rulesToApply) > 0 {
				// 只应用第一个规则
				firstRule := rulesToApply[0]
				newName, err := applyRules(n.Name, []rulesPkg.RenameRule{firstRule})
				if err != nil {
					log.Printf("Error applying rule to identifier %s: %v", n.Name, err)
					return node // Return original node on error
				}
				
				if newName != n.Name {
					n.Name = newName // Modify the identifier's name
					log.Printf("Transformed identifier %s to %s using fallback rule", node, newName)
				}
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
	priorityRules := make(map[string][]priorityRule)
	
	processConfigRules := func(cfgRules interface {
		IsDisabled() bool
		GetName() string
		GetRuleSet() *config.RuleSet
	}, priority int) {
		if cfgRules.IsDisabled() {
			return
		}
		name := cfgRules.GetName()
		ruleSet := cfgRules.GetRuleSet()
		log.Printf("Compile: Processing rule for %s, RuleSet: %+v, Priority: %d", name, ruleSet, priority)
		if ruleSet != nil {
			renameRules := rulesPkg.ConvertRuleSetToRenameRules(ruleSet)
			rules[name] = renameRules
			
			// 检查是否已经有更高优先级的规则
			hasHigherPriority := false
			if existingRules, ok := priorityRules[name]; ok {
				for _, pr := range existingRules {
					if pr.priority > priority {
						hasHigherPriority = true
						break
					}
				}
			}
			
			// 只有在没有更高优先级的规则时才添加
			if !hasHigherPriority {
				// 如果有相同优先级的规则，先清除它们
				newRules := []priorityRule{}
				if existingRules, ok := priorityRules[name]; ok {
					for _, pr := range existingRules {
						if pr.priority != priority {
							newRules = append(newRules, pr)
						}
					}
				}
				
				// 添加新规则
				for _, rule := range renameRules {
					newRules = append(newRules, priorityRule{
						rule:     rule,
						priority: priority,
					})
				}
				priorityRules[name] = newRules
			}
		}
	}

	// Process global rules (lowest priority: 0)
	for _, t := range cfg.Types {
		processConfigRules(t, 0)
	}
	for _, f := range cfg.Functions {
		processConfigRules(f, 0)
	}
	for _, v := range cfg.Variables {
		processConfigRules(v, 0)
	}
	for _, c := range cfg.Constants {
		processConfigRules(c, 0)
	}

	// Process package-specific rules (medium priority: 1)
	for _, pkg := range cfg.Packages {
		for _, t := range pkg.Types {
			processConfigRules(t, 1)
		}
		for _, f := range pkg.Functions {
			processConfigRules(f, 1)
		}
		for _, v := range pkg.Variables {
			processConfigRules(v, 1)
		}
		for _, c := range pkg.Constants {
			processConfigRules(c, 1)
		}
	}
	
 	// Process type-specific rules (highest priority: 2)
	// This would typically involve processing rules that are specific to certain types
	// For example, rules that apply only to specific struct fields or method parameters
	for _, pkg := range cfg.Packages {
		for _, t := range pkg.Types {
			// Process type-specific fields if they exist
			if t.Fields != nil {
				for _, field := range t.Fields {
					processConfigRules(field, 2) // Highest priority
				}
			}
			
			// Process methods if they exist
			if t.Methods != nil {
				for _, method := range t.Methods {
					processConfigRules(method, 2) // Highest priority
				}
			}
		}
	}
	
	// Sort priority rules by priority (highest first)
	for name, prules := range priorityRules {
		sort.Slice(prules, func(i, j int) bool {
			return prules[i].priority > prules[j].priority
		})
		priorityRules[name] = prules
	}
	
	replacer := &realReplacer{
		rules:        rules,
		priorityRules: priorityRules,
	}

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
			Types:       pkg.Types,
			Functions:   pkg.Functions,
			Variables:   pkg.Variables,
			Constants:   pkg.Constants,
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
