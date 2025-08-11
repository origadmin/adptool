package parser

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	gotoken "go/token"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

const directivePrefix = "//go:adapter:"

// ParseFileDirectives parses a Go source file (provided as an AST) and builds a config.Config object
// containing only the adptool directives found in that file.
// It does not perform any merging with global configurations.
func ParseFileDirectives(file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	cfg := config.New()

	// State trackers for the parser
	var lastTypeRule *config.TypeRule
	var lastFuncRule *config.FuncRule
	var lastVarRule *config.VarRule
	var lastConstRule *config.ConstRule
	var lastMemberRule *config.MemberRule

	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if !strings.HasPrefix(comment.Text, directivePrefix) {
				continue
			}

			rawDirective := strings.TrimPrefix(comment.Text, directivePrefix)
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
				// This parser only extracts directives from *this* file.
				// The top-level 'ignore' directive is for global rules, which are handled by the compiler.
				// We need to add this to a global list in the config.Config object.
				// For now, we'll add it to a dummy list or ignore it.
				// This part needs to be refined in a real compiler/resolver.
				// For now, we'll just add it to a dummy list or ignore it.
				// This is a placeholder for future implementation.
				continue
			}

			switch baseCmd {
			case "type":
				if len(cmdParts) == 1 {
					rule := &config.TypeRule{Name: argument, RuleSet: config.RuleSet{}}
					cfg.Types = append(cfg.Types, rule)
					lastTypeRule = rule
					lastFuncRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
				} else if len(cmdParts) == 2 && cmdParts[1] == "struct" {
					if lastTypeRule == nil {
						return nil, fmt.Errorf("line %d: 'type:struct' must follow a 'type' directive", fset.Position(comment.Pos()).Line)
					}
					lastTypeRule.Pattern = argument
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastTypeRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a 'type' directive", fset.Position(comment.Pos()).Line)
					}
					lastTypeRule.Disabled = argument == "true"
				} else if len(cmdParts) == 2 {
					// Generic sub-rule for type (e.g., :rename, :explicit)
					if lastTypeRule != nil {
						handleRule(&lastTypeRule.RuleSet, lastTypeRule.Name, cmdParts[1], argument)
					}
				}

			case "func":
				if len(cmdParts) == 1 {
					rule := &config.FuncRule{Name: argument, RuleSet: config.RuleSet{}}
					cfg.Functions = append(cfg.Functions, rule)
					lastFuncRule = rule
					lastTypeRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastFuncRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a 'func' directive", fset.Position(comment.Pos()).Line)
					}
					lastFuncRule.Disabled = argument == "true"
				} else if len(cmdParts) == 2 && lastFuncRule != nil {
					// Generic sub-rule for func (e.g., :rename, :explicit)
					handleRule(&lastFuncRule.RuleSet, lastFuncRule.Name, cmdParts[1], argument)
				}

			case "var":
				if len(cmdParts) == 1 {
					rule := &config.VarRule{Name: argument, RuleSet: config.RuleSet{}}
					cfg.Variables = append(cfg.Variables, rule)
					lastVarRule = rule
					lastTypeRule, lastFuncRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastVarRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a 'var' directive", fset.Position(comment.Pos()).Line)
					}
					lastVarRule.Disabled = argument == "true"
				} else if len(cmdParts) == 2 && lastVarRule != nil {
					// Generic sub-rule for var (e.g., :rename, :explicit)
					handleRule(&lastVarRule.RuleSet, lastVarRule.Name, cmdParts[1], argument)
				}

			case "const":
				if len(cmdParts) == 1 {
					rule := &config.ConstRule{Name: argument, RuleSet: config.RuleSet{}}
					cfg.Constants = append(cfg.Constants, rule)
					lastConstRule = rule
					lastTypeRule, lastFuncRule, lastVarRule, lastMemberRule = nil, nil, nil, nil
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastConstRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a 'const' directive", fset.Position(comment.Pos()).Line)
					}
					lastConstRule.Disabled = argument == "true"
				} else if len(cmdParts) == 2 && lastConstRule != nil {
					// Generic sub-rule for const (e.g., :rename, :explicit)
					handleRule(&lastConstRule.RuleSet, lastConstRule.Name, cmdParts[1], argument)
				}

			case "method", "field":
				if lastTypeRule == nil {
					return nil, fmt.Errorf("line %d: '%s' directive must follow a 'type' directive", fset.Position(comment.Pos()).Line, baseCmd)
				}
				if len(cmdParts) == 1 {
					member := &config.MemberRule{Name: argument, RuleSet: config.RuleSet{}}
					if baseCmd == "method" {
						lastTypeRule.Methods = append(lastTypeRule.Methods, member)
					} else {
						lastTypeRule.Fields = append(lastTypeRule.Fields, member)
					}
					lastMemberRule = member
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastMemberRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a member directive", fset.Position(comment.Pos()).Line)
					}
					lastMemberRule.Disabled = argument == "true"
				} else if len(cmdParts) == 2 && lastMemberRule != nil {
					// Generic sub-rule for method/field (e.g., :rename, :explicit)
					handleRule(&lastMemberRule.RuleSet, lastMemberRule.Name, cmdParts[1], argument)
				}
			}
		}
	}

	return cfg, nil
}

// getGlobalFuncRule finds or creates the global rule for functions.
// This is a temporary helper for the 'ignore' directive example.
// A full implementation would need similar helpers for all global rule types.
func getGlobalFuncRule(cfg *config.Config) *config.FuncRule {
	for _, r := range cfg.Functions {
		if r.Name == "*" {
			return r
		}
	}
	// Not found, create it
	globalRule := &config.FuncRule{Name: "*", RuleSet: config.RuleSet{}}
	cfg.Functions = append(cfg.Functions, globalRule)
	return globalRule
}

// handleRule applies a sub-rule to the appropriate ruleset.
func handleRule(ruleset *config.RuleSet, fromName, ruleName, argument string) {
	if ruleset == nil {
		return // Should not happen if logic is correct
	}

	switch ruleName {
	case "rename": // This is the old 'rename' sub-directive, now handled by explicit
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: fromName, To: argument})
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
