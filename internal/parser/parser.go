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
	var lastPackageRule *config.Package // New state variable for package

	// Context for package alias resolution (internal to parser)
	currentContext := &parsingContext{}

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
				// Strip inline comments from the argument
				argWithComment := parts[1]
				commentIndex := strings.Index(argWithComment, "//")
				if commentIndex != -1 {
					argument = strings.TrimSpace(argWithComment[:commentIndex])
				} else {
					argument = strings.TrimSpace(argWithComment)
				}
			}

			cmdParts := strings.Split(command, ":")
			baseCmd := cmdParts[0]

			// Handle top-level ignore directive
			if baseCmd == "ignore" {
				// This parser only extracts directives from *this* file.
				// The top-level 'ignore' directive is for global rules.
				// We need to add this to a global list in the config.Config object.
				// This requires identifying the type of the argument (e.g., func, type, const).
				// For now, we'll assume it's a constant for testing purposes.
				globalConstRule := getGlobalFuncRule(cfg)
				if globalConstRule.RuleSet.Ignore == nil {
					globalConstRule.RuleSet.Ignore = make([]string, 0)
				}
				globalConstRule.RuleSet.Ignore = append(globalConstRule.RuleSet.Ignore, argument)
				continue
			}

			switch baseCmd {
			case "defaults":
				if len(cmdParts) < 2 || cmdParts[1] != "mode" || len(cmdParts) < 3 {
					return nil, fmt.Errorf("line %d: invalid defaults directive format. Expected 'defaults:mode:<field> <value>'", fset.Position(comment.Pos()).Line)
				}
				if cfg.Defaults == nil {
					cfg.Defaults = &config.Defaults{}
				}
				if cfg.Defaults.Mode == nil {
					cfg.Defaults.Mode = &config.Mode{}
				}
				// Handle defaults:mode:<field> <value>
				modeField := cmdParts[2]
				switch modeField {
				case "strategy":
					cfg.Defaults.Mode.Strategy = argument
				case "prefix":
					cfg.Defaults.Mode.Prefix = argument
				case "suffix":
					cfg.Defaults.Mode.Suffix = argument
				case "explicit":
					cfg.Defaults.Mode.Explicit = argument
				case "regex":
					cfg.Defaults.Mode.Regex = argument
				case "ignore":
					cfg.Defaults.Mode.Ignore = argument
				default:
					return nil, fmt.Errorf("line %d: unknown defaults mode field '%s'", fset.Position(comment.Pos()).Line, modeField)
				}

			case "vars":
				if len(cmdParts) != 1 {
					return nil, fmt.Errorf("line %d: invalid vars directive format. Expected 'vars <name> <value>'", fset.Position(comment.Pos()).Line)
				}
				varNameParts := strings.SplitN(argument, " ", 2)
				if len(varNameParts) != 2 {
					return nil, fmt.Errorf("line %d: invalid vars directive argument. Expected 'name value'", fset.Position(comment.Pos()).Line)
				}
				if cfg.Vars == nil {
					cfg.Vars = make([]*config.VarEntry, 0)
				}
				cfg.Vars = append(cfg.Vars, &config.VarEntry{Name: varNameParts[0], Value: varNameParts[1]})

			case "package":
				if len(cmdParts) == 1 { // //go:adapter:package <import_path> [alias] or //go:adapter:package <import_path>
					pkgParts := strings.SplitN(argument, " ", 2)
					if len(pkgParts) == 2 { // Context-setting form: //go:adapter:package <import_path> <alias>
						currentContext.DefaultPackageImportPath = pkgParts[0]
						currentContext.DefaultPackageAlias = pkgParts[1]
					} else if len(pkgParts) == 1 { // Config-adding form: //go:adapter:package <import_path>
						if cfg.Packages == nil {
							cfg.Packages = make([]*config.Package, 0)
						}
						pkg := &config.Package{Import: argument}
						cfg.Packages = append(cfg.Packages, pkg)
						lastPackageRule = pkg
						// Reset other last rules
						lastTypeRule, lastFuncRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil, nil
					} else {
						return nil, fmt.Errorf("line %d: invalid package directive format. Expected 'package <import_path>' or 'package <import_path> <alias>'", fset.Position(comment.Pos()).Line)
					}
				} else if len(cmdParts) > 1 && cmdParts[0] == "package" && lastPackageRule != nil {
					// Sub-directives for the lastPackageRule
					subCmd := cmdParts[1]
					switch subCmd {
					case "alias":
						lastPackageRule.Alias = argument
					case "path":
						lastPackageRule.Path = argument
					case "vars":
						varNameParts := strings.SplitN(argument, " ", 2)
						if len(varNameParts) != 2 {
							return nil, fmt.Errorf("line %d: invalid package vars directive argument. Expected 'name value'", fset.Position(comment.Pos()).Line)
						}
						if lastPackageRule.Vars == nil {
							lastPackageRule.Vars = make([]*config.VarEntry, 0)
						}
						lastPackageRule.Vars = append(lastPackageRule.Vars, &config.VarEntry{Name: varNameParts[0], Value: varNameParts[1]})
					case "types":
						// Create TypeRule within package, set it as lastTypeRule
						rule := &config.TypeRule{Name: argument, RuleSet: config.RuleSet{}}
						if lastPackageRule.Types == nil {
							lastPackageRule.Types = make([]*config.TypeRule, 0)
						}
						lastPackageRule.Types = append(lastPackageRule.Types, rule)
						lastTypeRule = rule
						// Reset other last rules
						lastFuncRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
					case "functions":
						rule := &config.FuncRule{Name: argument, RuleSet: config.RuleSet{}}
						if lastPackageRule.Functions == nil {
							lastPackageRule.Functions = make([]*config.FuncRule, 0)
						}
						lastPackageRule.Functions = append(lastPackageRule.Functions, rule)
						lastFuncRule = rule
						// Reset other last rules
						lastTypeRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
					case "variables":
						rule := &config.VarRule{Name: argument, RuleSet: config.RuleSet{}}
						if lastPackageRule.Variables == nil {
							lastPackageRule.Variables = make([]*config.VarRule, 0)
						}
						lastPackageRule.Variables = append(lastPackageRule.Variables, rule)
						lastVarRule = rule
						// Reset other last rules
						lastTypeRule, lastFuncRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
					case "constants":
						rule := &config.ConstRule{Name: argument, RuleSet: config.RuleSet{}}
						if lastPackageRule.Constants == nil {
							lastPackageRule.Constants = make([]*config.ConstRule, 0)
						}
						lastPackageRule.Constants = append(lastPackageRule.Constants, rule)
						lastConstRule = rule
						// Reset other last rules
						lastTypeRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
					default:
						return nil, fmt.Errorf("line %d: unknown package sub-directive '%s'", fset.Position(comment.Pos()).Line, command)
					}
				}

			case "type":
				if len(cmdParts) == 1 {
					rule := &config.TypeRule{Name: argument, Kind: "type", RuleSet: config.RuleSet{}}
					cfg.Types = append(cfg.Types, rule)
					lastTypeRule = rule
					lastFuncRule, lastVarRule, lastConstRule, lastMemberRule, lastPackageRule = nil, nil, nil, nil, nil
				} else if len(cmdParts) == 2 && cmdParts[1] == "struct" {
					if lastTypeRule == nil {
						return nil, fmt.Errorf("line %d: 'type:struct' must follow a 'type' directive", fset.Position(comment.Pos()).Line)
					}
					lastTypeRule.Pattern = argument
					lastTypeRule.Kind = "struct" // Set Kind to struct
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
					lastTypeRule, lastVarRule, lastConstRule, lastMemberRule, lastPackageRule = nil, nil, nil, nil, nil
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

// parsingContext holds the current default package information for directive resolution.
type parsingContext struct {
	DefaultPackageImportPath string
	DefaultPackageAlias      string
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
	case "rename":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: fromName, To: argument})
	case "explicit":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		explicitRules := strings.SplitN(argument, " ", 2)
		if len(explicitRules) == 2 {
			ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: explicitRules[0], To: explicitRules[1]})
		}
	case "explicit.json":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		var explicitRules []*config.ExplicitRule
		if err := json.Unmarshal([]byte(argument), &ruleset.Explicit); err == nil {
			ruleset.Explicit = explicitRules
		}
	}
}
