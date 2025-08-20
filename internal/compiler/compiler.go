package compiler

import (
	"go/ast"
	"go/token"
	"log"
	"path" // Added for path.Base
	"sort"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/interfaces"
	rulesPkg "github.com/origadmin/adptool/internal/rules"
)

// compiledRules stores the transformation rules compiled from the config.
// It maps original names to a slice of interfaces.RenameRule.
type compiledRules map[string][]interfaces.RenameRule

// priorityRule represents a rule with its priority level and package context
type priorityRule struct {
	rule        interfaces.RenameRule
	priority    int    // 0: global, 1: package level, 2: type specific
	packageName string // Package name for package-level rules
	isWildcard  bool   // Whether this is a wildcard rule
}

// realReplacer implements the interfaces.Replacer interface
// and applies actual transformation rules based on the compiled configuration.
type realReplacer struct {
	config *interfaces.CompiledConfig
	// 包别名字典，用于避免对包别名应用重命名规则
	packageAliases map[string]bool
	// 已处理节点的映射，用于避免重复处理
	processedNodes map[ast.Node]bool
}

// Apply applies the transformation rules to the given AST node.
func (r *realReplacer) Apply(ctx interfaces.Context, node ast.Node) ast.Node {
	// 检查节点是否已经处理过
	if r.processedNodes != nil {
		if r.processedNodes[node] {
			// 节点已经处理过，直接返回
			return node
		}
		// 标记节点已处理
		r.processedNodes[node] = true
	}

	switch n := node.(type) {
	case *ast.Ident:
		log.Printf("Apply: Processing identifier: %s", n.Name)

		// 检查是否为包别名，如果是则不应用重命名规则
		if r.packageAliases != nil && r.packageAliases[n.Name] {
			log.Printf("Apply: Skipping package alias %s", n.Name)
			return node
		}

		// 根据上下文决定是否应用规则
		shouldApplyRules := false
		ruleType := ctx.CurrentNodeType()
		if ruleType == "const_decl_name" {
			// 如果在常量声明名称上下文中，则应用规则
			shouldApplyRules = true
			log.Printf("Apply: Applying const rules for %s in const declaration name context", n.Name)
		} else if ruleType == "type" {
			// 如果在类型声明上下文中，则应用规则
			shouldApplyRules = true
			log.Printf("Apply: Applying type rules for %s in type declaration context", n.Name)
		} else if ruleType == "var_decl_name" {
			// 如果在变量声明名称上下文中，则应用规则
			shouldApplyRules = true
			log.Printf("Apply: Applying var rules for %s in var declaration name context", n.Name)
		} else if ruleType == "func" {
			// 如果在函数声明上下文中，则应用规则
			shouldApplyRules = true
			log.Printf("Apply: Applying func rules for %s in func declaration context", n.Name)
		} else {
			log.Printf("Apply: Skipping rule application for %s as it's not in a rule-applicable context", n.Name)
		}

		if shouldApplyRules {
			// 获取当前节点的包上下文
			// 在这个简化版本中，我们将实现一个基本的匹配逻辑
			// 实际应用中，可能需要通过AST节点的上下文来确定当前所在的包

			// 首先检查是否有针对该标识符的特定规则
			if priorityRulesToApply, ok := r.config.PriorityRules[n.Name]; ok {
				log.Printf("Apply: Found priority rules for %s: %+v", n.Name, priorityRulesToApply)

				// 根据规则类型过滤规则
				var filteredRules []interfaces.PriorityRule

				if ruleType != "" {
					for _, rule := range priorityRulesToApply {
						// 这里需要根据规则的来源判断规则类型
						// 简化处理：假设规则的优先级或其它属性可以标识规则类型
						// 实际实现中应该在编译阶段就标记规则类型
						filteredRules = append(filteredRules, rule)
					}
				} else {
					filteredRules = priorityRulesToApply
				}

				// 应用优先级规则
				if len(filteredRules) > 0 {
					// 获取最高优先级的规则
					highestPriorityRule := filteredRules[0]
					log.Printf("Apply: Using highest priority rule for %s: %+v", n.Name, highestPriorityRule)

					// 只应用这一个规则
					newName, err := rulesPkg.ApplyRules(n.Name, []interfaces.RenameRule{
						{
							Type:    highestPriorityRule.Rule.Type,
							Value:   highestPriorityRule.Rule.Value,
							From:    highestPriorityRule.Rule.From,
							To:      highestPriorityRule.Rule.To,
							Pattern: highestPriorityRule.Rule.Pattern,
							Replace: highestPriorityRule.Rule.Replace,
						},
					})
					if err != nil {
						log.Printf("Error applying priority rule to identifier %s: %v", n.Name, err)
						return node // Return original node on error
					}

					if newName != n.Name {
						n.Name = newName // Modify the identifier's name
						log.Printf("Transformed identifier %s to %s using priority rule", n.Name, newName)
					}
				}
			} else if priorityWildcardRules, ok := r.config.PriorityRules["*"]; ok {
				// 如果没有特定规则，检查是否有通配符规则
				log.Printf("Apply: Found wildcard rules: %+v", priorityWildcardRules)

				// 根据规则类型过滤通配符规则
				var filteredRules []interfaces.PriorityRule

				if ruleType != "" {
					for _, rule := range priorityWildcardRules {
						// 根据上下文类型过滤规则
						// 这里需要更复杂的逻辑来确定规则类型与上下文的匹配
						// 简化处理：假设规则的Value可以标识规则类型
						if (ruleType == "const" && rule.Rule.Value == "Const") ||
							(ruleType == "type" && rule.Rule.Value == "Type") ||
							(ruleType == "var" && rule.Rule.Value == "Var") ||
							(ruleType == "func" && rule.Rule.Value == "Func") {
							filteredRules = append(filteredRules, rule)
						}
					}
				} else {
					filteredRules = priorityWildcardRules
				}

				// 应用通配符规则
				if len(filteredRules) > 0 {
					var rulesToApply []interfaces.RenameRule
					for _, rule := range filteredRules {
						rulesToApply = append(rulesToApply, interfaces.RenameRule{
							Type:    rule.Rule.Type,
							Value:   rule.Rule.Value,
							From:    rule.Rule.From,
							To:      rule.Rule.To,
							Pattern: rule.Rule.Pattern,
							Replace: rule.Rule.Replace,
						})
					}

					newName, err := rulesPkg.ApplyRules(n.Name, rulesToApply)
					if err != nil {
						log.Printf("Error applying wildcard rules to identifier %s: %v", n.Name, err)
						return node // Return original node on error
					}

					if newName != n.Name {
						n.Name = newName
						log.Printf("Transformed identifier %s to %s using wildcard rule", n.Name, newName)
					}
				}
			} else if rulesToApply, ok := r.config.Rules[n.Name]; ok {
				// Fallback to the old rules if no priority rules exist
				log.Printf("Apply: Found rules for %s: %+v", n.Name, rulesToApply)

				// 根据规则类型过滤规则
				var filteredRules []interfaces.RenameRule

				if ruleType != "" {
					for _, rule := range rulesToApply {
						// 根据上下文类型过滤规则
						if (ruleType == "const" && rule.Value == "Const") ||
							(ruleType == "type" && rule.Value == "Type") ||
							(ruleType == "var" && rule.Value == "Var") ||
							(ruleType == "func" && rule.Value == "Func") {
							filteredRules = append(filteredRules, rule)
						}
					}
				} else {
					filteredRules = rulesToApply
				}

				// 为了保持一致性，我们也只应用第一个规则
				if len(filteredRules) > 0 {
					// 只应用第一个规则
					firstRule := filteredRules[0]
					newName, err := rulesPkg.ApplyRules(n.Name, []interfaces.RenameRule{
						{
							Type:    firstRule.Type,
							Value:   firstRule.Value,
							From:    firstRule.From,
							To:      firstRule.To,
							Pattern: firstRule.Pattern,
							Replace: firstRule.Replace,
						},
					})
					if err != nil {
						log.Printf("Error applying rule to identifier %s: %v", n.Name, err)
						return node // Return original node on error
					}

					if newName != n.Name {
						n.Name = newName // Modify the identifier's name
						log.Printf("Transformed identifier %s to %s using fallback rule", n.Name, newName)
					}
				}
			} else {
				log.Printf("Apply: No rules found for identifier: %s", n.Name)
			}
		} else {
			log.Printf("Apply: Skipping rule application for identifier: %s", n.Name)
		}

	// 处理常量声明
	case *ast.GenDecl:
		if n.Tok == token.CONST {
			log.Printf("Apply: Processing constant declaration")
			// 添加常量上下文
			ctx = ctx.Push("const")

			// 处理常量声明中的值规范
			for _, spec := range n.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					// 为常量名称添加声明上下文
					constDeclCtx := ctx.Push("const_decl")
					for _, name := range valueSpec.Names {
						constDeclNameCtx := constDeclCtx.Push("const_decl_name")
						r.Apply(constDeclNameCtx, name)
					}
					// 处理值部分，但不应用规则到引用的标识符
					// 这里我们不递归处理值部分，因为AST遍历会自动处理其中的标识符
				}
			}
		} else if n.Tok == token.VAR {
			log.Printf("Apply: Processing variable declaration")
			// 添加变量上下文
			ctx.Push("var")

			// 处理变量声明中的值规范
			for _, spec := range n.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					// 为变量名称添加声明上下文
					ctx.Push("var_decl")
					for _, name := range valueSpec.Names {
						valueSpecCtx := ctx.Push("var_decl_name")
						r.Apply(valueSpecCtx, name)
					}
				}
			}
		} else if n.Tok == token.TYPE {
			log.Printf("Apply: Processing type declaration")
			// 添加类型上下文
			ctx.Push("type_decl")

			// 处理类型声明中的类型规范
			for _, spec := range n.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					typeCtx := ctx.Push("type")
					r.Apply(typeCtx, typeSpec.Name)
				}
			}
		}

	// 处理函数声明
	case *ast.FuncDecl:
		log.Printf("Apply: Processing function declaration: %s", n.Name.Name)
		// 添加函数上下文
		funcCtx := ctx.Push("func")
		r.Apply(funcCtx, n.Name)
	// 处理类型声明
	case *ast.TypeSpec:
		log.Printf("Apply: Processing type declaration: %s", n.Name.Name)
		// 添加类型上下文
		typeCtx := ctx.Push("type")
		r.Apply(typeCtx, n.Name)

	// TODO: Add more cases for other AST node types if needed
	default:
		// log.Printf("Real replacer applied to node: %T", node) // Too verbose
	}
	return node
}

// isWildcardConstRuleOnly 检查是否只有常量通配符规则
func (r *realReplacer) isWildcardConstRuleOnly() bool {
	// 检查是否有通配符规则
	if wildcardRules, ok := r.config.PriorityRules["*"]; ok && len(wildcardRules) > 0 {
		// 检查是否所有通配符规则都是一样的（前缀为"Const"的规则）
		for _, rule := range wildcardRules {
			if rule.Rule.Type != "prefix" || rule.Rule.Value != "Const" {
				// 如果有任何规则不是Const前缀规则，则不是纯常量规则
				return false
			}
		}
		// 所有通配符规则都是Const前缀规则
		return true
	}
	return false
}

// NewReplacer creates a new Replacer instance from a compiled configuration.
// This function is used to create a new realReplacer instance that will be used
// to apply transformation rules to AST nodes during code generation.
func NewReplacer(compiledCfg *interfaces.CompiledConfig) interfaces.Replacer {
	if compiledCfg == nil {
		return nil
	}

	// 初始化包别名字典
	packageAliases := make(map[string]bool)
	for _, pkg := range compiledCfg.Packages {
		packageAliases[pkg.ImportAlias] = true
	}

	return &realReplacer{
		config:         compiledCfg,
		packageAliases: packageAliases,
		processedNodes: make(map[ast.Node]bool), // 初始化已处理节点映射
	}
}

// Compile takes a configuration and returns a compiled representation of it.
func Compile(cfg *config.Config) (*interfaces.CompiledConfig, error) {
	log.Printf("Compile: Received config: %+v", cfg)

	// --- Compile Renaming Rules ---
	rules := make(compiledRules)
	priorityRules := make(map[string][]priorityRule)

	processConfigRules := func(cfgRules interface {
		IsDisabled() bool
		GetName() string
		GetRuleSet() *config.RuleSet
	}, priority int, packageName string) {
		if cfgRules.IsDisabled() {
			return
		}
		name := cfgRules.GetName()
		ruleSet := cfgRules.GetRuleSet()
		log.Printf("Compile: Processing rule for %s, RuleSet: %+v, Priority: %d, Package: %s", name, ruleSet, priority, packageName)
		if ruleSet != nil {
			// Convert rules to config.RenameRule format
			renameRules := rulesPkg.ConvertRuleSetToRenameRules(ruleSet)
			rules[name] = renameRules

			// 添加规则到priorityRules，支持通配符
			isWildcard := name == "*"
			for _, rule := range renameRules {
				priorityRules[name] = append(priorityRules[name], priorityRule{
					rule:        rule,
					priority:    priority,
					packageName: packageName,
					isWildcard:  isWildcard,
				})
			}
		}
	}

	// Process global rules (lowest priority: 0)
	for _, t := range cfg.Types {
		processConfigRules(t, 0, "")
	}
	for _, f := range cfg.Functions {
		processConfigRules(f, 0, "")
	}
	for _, v := range cfg.Variables {
		processConfigRules(v, 0, "")
	}
	for _, c := range cfg.Constants {
		processConfigRules(c, 0, "")
	}

	// Process package-specific rules (medium priority: 1)
	for _, pkg := range cfg.Packages {
		for _, t := range pkg.Types {
			processConfigRules(t, 1, pkg.Import)
		}
		for _, f := range pkg.Functions {
			processConfigRules(f, 1, pkg.Import)
		}
		for _, v := range pkg.Variables {
			processConfigRules(v, 1, pkg.Import)
		}
		for _, c := range pkg.Constants {
			processConfigRules(c, 1, pkg.Import)
		}

		// 添加包级通配符规则
		// 这里可以添加针对整个包的通配符规则处理
	}

	// Process type-specific rules (highest priority: 2)
	// This would typically involve processing rules that are specific to certain types
	// For example, rules that apply only to specific struct fields or method parameters
	for _, pkg := range cfg.Packages {
		for _, t := range pkg.Types {
			// Process type-specific fields if they exist
			if t.Fields != nil {
				for _, field := range t.Fields {
					processConfigRules(field, 2, pkg.Import) // Highest priority
				}
			}

			// Process methods if they exist
			if t.Methods != nil {
				for _, method := range t.Methods {
					processConfigRules(method, 2, pkg.Import) // Highest priority
				}
			}
		}
	}

	// Sort priority rules by priority (highest first)
	for name, prules := range priorityRules {
		sort.Slice(prules, func(i, j int) bool {
			// 首先按优先级排序（数字越大优先级越高）
			if prules[i].priority != prules[j].priority {
				return prules[i].priority > prules[j].priority
			}
			// 对于相同优先级，非通配符规则优先于通配符规则
			if prules[i].isWildcard != prules[j].isWildcard {
				return !prules[i].isWildcard // 非通配符优先
			}
			// 相同优先级和通配符属性时，保持原有顺序
			return false
		})
		priorityRules[name] = prules
	}

	// --- Compile Package Information ---
	var compiledPackages []*interfaces.CompiledPackage

	// Process packages to create compiled packages
	for _, pkg := range cfg.Packages {
		finalAlias := pkg.Alias
		if finalAlias == "" {
			finalAlias = path.Base(pkg.Import)
		}

		compiledPkg := &interfaces.CompiledPackage{
			ImportPath:  pkg.Import,
			ImportAlias: finalAlias,
		}
		compiledPackages = append(compiledPackages, compiledPkg)
	}

	// Convert rules to config.RenameRule format
	configRules := make(map[string][]interfaces.RenameRule)
	for name, ruleList := range rules {
		configRules[name] = ruleList
	}

	compiledCfg := &interfaces.CompiledConfig{
		PackageName:   cfg.OutputPackageName,
		Packages:      compiledPackages,
		Rules:         configRules,
		PriorityRules: convertPriorityRules(priorityRules),
	}

	// Use a default output package name if not provided
	if compiledCfg.PackageName == "" {
		compiledCfg.PackageName = "adapters"
	}

	log.Printf("Successfully compiled config: %+v", compiledCfg)
	return compiledCfg, nil
}

// convertPriorityRules converts the internal priorityRule map to the format used in CompiledConfig
func convertPriorityRules(prules map[string][]priorityRule) map[string][]interfaces.PriorityRule {
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
