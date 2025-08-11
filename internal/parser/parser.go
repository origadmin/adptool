package parser

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// ParseFile parses a Go source file and builds a config.Config object from the adptool directives found.
func ParseFile(filePath string) (*config.Config, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	cfg := config.New()

	// State trackers for the parser
	var lastTypeRule *config.TypeRule
	var lastFuncRule *config.FuncRule
	var lastVarRule *config.VarRule
	var lastConstRule *config.ConstRule
	var lastMemberRule *config.MemberRule

	for _, commentGroup := range node.Comments {
		for _, comment := range commentGroup.List {
			if !strings.HasPrefix(comment.Text, "//go:adapter:") {
				continue
			}

			rawDirective := strings.TrimPrefix(comment.Text, "//go:adapter:")
			parts := strings.SplitN(rawDirective, " ", 2)
			command := parts[0]
			argument := ""
			if len(parts) > 1 {
				argument = parts[1]
			}

			cmdParts := strings.Split(command, ":")
			baseCmd := cmdParts[0]

			// Handle top-level ignore directive separately as it doesn't set a target
			if baseCmd == "ignore" {
				// This is a simplified logic. A real implementation would need to parse the argument
				// to determine if it's a func, type, etc., and add it to the correct global ignore list.
				// For now, we assume it applies to functions as an example.
				getGlobalFuncRule(cfg).Ignore = append(getGlobalFuncRule(cfg).Ignore, argument)
				continue
			}

			switch baseCmd {
			case "type":
				if len(cmdParts) == 1 {
					rule := &config.TypeRule{Name: argument, RuleSet: &config.RuleSet{}}
					cfg.Types = append(cfg.Types, rule)
					lastTypeRule = rule
					lastFuncRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
				} else if len(cmdParts) == 2 && cmdParts[1] == "struct" {
					if lastTypeRule == nil {
						return nil, fmt.Errorf("line %d: 'type:struct' must follow a 'type' directive", fset.Position(comment.Pos()).Line)
					}
					lastTypeRule.Pattern = argument
				} else if len(cmdParts) == 2 {
					if lastTypeRule != nil {
						handleRule(lastTypeRule, nil, cmdParts[1], argument)
					}
				}

			case "func":
				if len(cmdParts) == 1 {
					rule := &config.FuncRule{Name: argument, RuleSet: &config.RuleSet{}}
					cfg.Functions = append(cfg.Functions, rule)
					lastFuncRule = rule
					lastTypeRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
				} else if len(cmdParts) == 2 && lastFuncRule != nil {
					handleRule(lastFuncRule, nil, cmdParts[1], argument)
				}

			case "var":
				if len(cmdParts) == 1 {
					rule := &config.VarRule{Name: argument, RuleSet: &config.RuleSet{}}
					cfg.Variables = append(cfg.Variables, rule)
					lastVarRule = rule
					lastTypeRule, lastFuncRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
				} else if len(cmdParts) == 2 && lastVarRule != nil {
					handleRule(lastVarRule, nil, cmdParts[1], argument)
				}

			case "const":
				if len(cmdParts) == 1 {
					rule := &config.ConstRule{Name: argument, RuleSet: &config.RuleSet{}}
					cfg.Constants = append(cfg.Constants, rule)
					lastConstRule = rule
					lastTypeRule, lastFuncRule, lastVarRule, lastMemberRule = nil, nil, nil, nil
				} else if len(cmdParts) == 2 && lastConstRule != nil {
					handleRule(lastConstRule, nil, cmdParts[1], argument)
				}

			case "method", "field":
				if lastTypeRule == nil {
					return nil, fmt.Errorf("line %d: '%s' directive must follow a 'type' directive", fset.Position(comment.Pos()).Line, baseCmd)
				}
				if len(cmdParts) == 1 {
					member := &config.MemberRule{Name: argument, RuleSet: &config.RuleSet{}}
					if baseCmd == "method" {
						lastTypeRule.Methods = append(lastTypeRule.Methods, member)
					} else {
						lastTypeRule.Fields = append(lastTypeRule.Fields, member)
					}
					lastMemberRule = member
				} else if len(cmdParts) == 2 && lastMemberRule != nil {
					handleRule(nil, lastMemberRule, cmdParts[1], argument)
				}
			}
		}
	}

	return cfg, nil
}

// getGlobalFuncRule finds or creates the global rule for functions.
func getGlobalFuncRule(cfg *config.Config) *config.FuncRule {
	for _, r := range cfg.Functions {
		if r.Name == "*" {
			return r
		}
	}
	// Not found, create it
	globalRule := &config.FuncRule{Name: "*", RuleSet: &config.RuleSet{}}
	cfg.Functions = append(cfg.Functions, globalRule)
	return globalRule
}

// handleRule applies a sub-rule to the appropriate rule object.
func handleRule(target interface{}, member *config.MemberRule, ruleName, argument string) {
	switch r := target.(type) {
	case *config.TypeRule:
		applyToRuleSet(r.RuleSet, r.Name, ruleName, argument)
	case *config.FuncRule:
		applyToRuleSet(r.RuleSet, r.Name, ruleName, argument)
	case *config.VarRule:
		applyToRuleSet(r.RuleSet, r.Name, ruleName, argument)
	case *config.ConstRule:
		applyToRuleSet(r.RuleSet, r.Name, ruleName, argument)
	case *config.MemberRule:
		applyToRuleSet(r.RuleSet, r.Name, ruleName, argument)
	}
	if member != nil {
		applyToRuleSet(member.RuleSet, member.Name, ruleName, argument)
	}
}

func applyToRuleSet(ruleset *config.RuleSet, fromName, ruleName, argument string) {
	if ruleset == nil {
		return // Should not happen
	}

	switch ruleName {
	case "rename":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: fromName, To: argument})
	case "disabled":
		// This property is on the rule struct itself, not the ruleset.
		// The main switch should handle this by setting the Disabled field on the target/member rule.
	case "explicit":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		var explicitRules []*config.ExplicitRule
		if err := json.Unmarshal([]byte(argument), &explicitRules); err == nil {
			ruleset.Explicit = append(ruleset.Explicit, explicitRules...)
		}
	}
}
