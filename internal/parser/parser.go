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
	fmt.Println("Starting ParseFileDirectives...")
	cfg := config.New()

	// State trackers for the parser
	var lastTypeRule *config.TypeRule
	var lastFuncRule *config.FuncRule
	var lastVarRule *config.VarRule
	var lastConstRule *config.ConstRule
	var lastMemberRule *config.MemberRule
	var lastPackageRule *config.Package // New state variable for package

	// Temporary storage for pending ignore arguments
	var pendingIgnoreArguments []string

	// Context for package alias resolution (internal to parser)
	currentContext := &parsingContext{}

	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			fmt.Printf("Processing comment: %s\n", comment.Text)
			if !strings.HasPrefix(comment.Text, directivePrefix) {
				continue
			}

			rawDirective := strings.TrimPrefix(comment.Text, directivePrefix)
			commentIndex := strings.Index(rawDirective, "//")
			if commentIndex != -1 {
				rawDirective = rawDirective[:commentIndex]
			}
			parts := strings.SplitN(rawDirective, " ", 2)
			command := parts[0]
			argument := ""
			if len(parts) > 1 {
				// Strip inline comments from the argument
				argument = parts[1]
			}
			fmt.Printf("  rawDirective: %s, command: %s, argument: %s\n", rawDirective, command, argument)

			cmdParts := strings.Split(command, ":")
			baseCmd := cmdParts[0]
			fmt.Printf("  baseCmd: %s\n", baseCmd)

			// Handle top-level ignore directive
			if baseCmd == "ignore" {
				fmt.Printf("  Handling 'ignore' directive for argument: %s\n", argument)
				pendingIgnoreArguments = append(pendingIgnoreArguments, argument)
				fmt.Printf("  Added '%s' to pending ignore list. Current pending ignore list: %+v\n", argument, pendingIgnoreArguments)
				continue
			}

			// Helper to apply pending ignore arguments to a rule's RuleSet
			applyPendingIgnore := func(rs *config.RuleSet) {
				if len(pendingIgnoreArguments) > 0 {
					if rs.Ignore == nil {
						rs.Ignore = make([]string, 0)
					}
					rs.Ignore = append(rs.Ignore, pendingIgnoreArguments...)
					fmt.Printf("  Applied pending ignore arguments: %+v to RuleSet. Current RuleSet.Ignore: %+v\n", pendingIgnoreArguments, rs.Ignore)
					pendingIgnoreArguments = nil // Clear after applying
				}
			}

			switch baseCmd {
			case "defaults":
				fmt.Printf("  Processing 'defaults' directive. cmdParts: %+v\n", cmdParts)
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
				fmt.Printf("  Defaults Mode updated: %+v\n", cfg.Defaults.Mode)

			case "vars":
				fmt.Printf("  Processing 'vars' directive. argument: %s\n", argument)
				if len(cmdParts) != 1 {
					return nil, fmt.Errorf("line %d: invalid vars directive format. Expected 'vars <name> <value>'", fset.Position(comment.Pos()).Line)
				}
				varNameParts := strings.SplitN(argument, " ", 2)
				if len(varNameParts) != 2 {
					return nil, fmt.Errorf("line %d: invalid vars directive argument. Expected 'name value'", fset.Position(comment.Pos()).Line)
				}
				// Global vars are VarEntry, not VarRule
				entry := &config.VarEntry{Name: varNameParts[0], Value: varNameParts[1]}
				if cfg.Vars == nil {
					cfg.Vars = make([]*config.VarEntry, 0)
				}
				cfg.Vars = append(cfg.Vars, entry)
				fmt.Printf("  Added Var: Name=%s, Value=%s. Current Vars: %+v\n", varNameParts[0], varNameParts[1], cfg.Vars)

			case "package":
				fmt.Printf("  Processing 'package' directive. cmdParts: %+v, argument: %s\n", cmdParts, argument)
				if len(cmdParts) == 1 { // //go:adapter:package <import_path> [alias] or //go:adapter:package <import_path>
					pkgParts := strings.SplitN(argument, " ", 2)
					if len(pkgParts) == 2 { // Context-setting form: //go:adapter:package <import_path> <alias>
						currentContext.DefaultPackageImportPath = pkgParts[0]
						currentContext.DefaultPackageAlias = pkgParts[1]
						fmt.Printf("  Set package context: ImportPath=%s, Alias=%s\n", currentContext.DefaultPackageImportPath, currentContext.DefaultPackageAlias)
					} else if len(pkgParts) == 1 { // Config-adding form: //go:adapter:package <import_path>
						if cfg.Packages == nil {
							cfg.Packages = make([]*config.Package, 0)
						}
						pkg := &config.Package{Import: argument}
						cfg.Packages = append(cfg.Packages, pkg)
						lastPackageRule = pkg
						//applyPendingIgnore(&pkg.RuleSet) // Apply pending ignore to package rule
						fmt.Printf("  Added Package: Import=%s. Current Packages: %+v\n", argument, cfg.Packages)
						// Reset other last rules
						lastTypeRule, lastFuncRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil, nil
						fmt.Println("  Resetting last rules for new package.")
					} else {
						return nil, fmt.Errorf("line %d: invalid package directive format. Expected 'package <import_path>' or 'package <import_path> <alias>'", fset.Position(comment.Pos()).Line)
					}
				} else if len(cmdParts) > 1 && cmdParts[0] == "package" && lastPackageRule != nil {
					// Sub-directives for the lastPackageRule
					subCmd := cmdParts[1]
					fmt.Printf("  Processing package sub-directive: %s, argument: %s\n", subCmd, argument)
					switch subCmd {
					case "alias":
						lastPackageRule.Alias = argument
						fmt.Printf("  Package Alias updated: %s\n", lastPackageRule.Alias)
					case "path":
						lastPackageRule.Path = argument
						fmt.Printf("  Package Path updated: %s\n", lastPackageRule.Path)
					case "vars":
						varNameParts := strings.SplitN(argument, " ", 2)
						if len(varNameParts) != 2 {
							return nil, fmt.Errorf("line %d: invalid package vars directive argument. Expected 'name value'", fset.Position(comment.Pos()).Line)
						}
						// Package vars are VarEntry, not VarRule
						entry := &config.VarEntry{Name: varNameParts[0], Value: varNameParts[1]}
						if lastPackageRule.Vars == nil {
							lastPackageRule.Vars = make([]*config.VarEntry, 0)
						}
						lastPackageRule.Vars = append(lastPackageRule.Vars, entry)
						fmt.Printf("  Added Package Var: Name=%s, Value=%s. Current Package Vars: %+v\n", varNameParts[0], varNameParts[1], lastPackageRule.Vars)
					case "types":
						// Create TypeRule within package, set it as lastTypeRule
						rule := &config.TypeRule{Name: argument, RuleSet: config.RuleSet{}}
						if lastPackageRule.Types == nil {
							lastPackageRule.Types = make([]*config.TypeRule, 0)
						}
						lastPackageRule.Types = append(lastPackageRule.Types, rule)
						lastTypeRule = rule
						applyPendingIgnore(&rule.RuleSet) // Apply pending ignore to type rule
						fmt.Printf("  Added Package Type: Name=%s. Current Package Types: %+v\n", argument, lastPackageRule.Types)
						// Reset other last rules
						lastFuncRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
						fmt.Println("  Resetting last rules for new package type.")
					case "functions":
						rule := &config.FuncRule{Name: argument, RuleSet: config.RuleSet{}}
						if lastPackageRule.Functions == nil {
							lastPackageRule.Functions = make([]*config.FuncRule, 0)
						}
						lastPackageRule.Functions = append(lastPackageRule.Functions, rule)
						lastFuncRule = rule
						applyPendingIgnore(&rule.RuleSet) // Apply pending ignore to func rule
						fmt.Printf("  Added Package Function: Name=%s. Current Package Functions: %+v\n", argument, lastPackageRule.Functions)
						// Reset other last rules
						lastTypeRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
						fmt.Println("  Resetting last rules for new package function.")
					case "variables":
						rule := &config.VarRule{Name: argument, RuleSet: config.RuleSet{}}
						if lastPackageRule.Variables == nil {
							lastPackageRule.Variables = make([]*config.VarRule, 0)
						}
						lastPackageRule.Variables = append(lastPackageRule.Variables, rule)
						lastVarRule = rule
						applyPendingIgnore(&rule.RuleSet) // Apply pending ignore to var rule
						fmt.Printf("  Added Package Variable: Name=%s. Current Package Variables: %+v\n", argument, lastPackageRule.Variables)
						// Reset other last rules
						lastTypeRule, lastFuncRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
						fmt.Println("  Resetting last rules for new package variable.")
					case "constants":
						rule := &config.ConstRule{Name: argument, RuleSet: config.RuleSet{}}
						if lastPackageRule.Constants == nil {
							lastPackageRule.Constants = make([]*config.ConstRule, 0)
						}
						lastPackageRule.Constants = append(lastPackageRule.Constants, rule)
						lastConstRule = rule
						applyPendingIgnore(&rule.RuleSet) // Apply pending ignore to const rule
						fmt.Printf("  Added Package Constant: Name=%s. Current Package Constants: %+v\n", argument, lastPackageRule.Constants)
						// Reset other last rules
						lastTypeRule, lastVarRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
						fmt.Println("  Resetting last rules for new package constant.")
					default:
						// Handle sub-directives of types/functions/variables/constants within a package
						if len(cmdParts) == 3 && cmdParts[2] == "struct" && lastTypeRule != nil {
							lastTypeRule.Pattern = argument
							lastTypeRule.Kind = "struct"
							fmt.Printf("  Updated Package Type Pattern: %s, Kind: %s\n", lastTypeRule.Pattern, lastTypeRule.Kind)
						} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && lastTypeRule != nil {
							lastTypeRule.Disabled = argument == "true"
							fmt.Printf("  Updated Package Type Disabled: %t\n", lastTypeRule.Disabled)
						} else if len(cmdParts) == 3 && lastTypeRule != nil {
							// Generic sub-rule for type within package
							fmt.Printf("  Handling package type sub-rule: %s, argument: %s\n", cmdParts[2], argument)
							handleRule(&lastTypeRule.RuleSet, lastTypeRule.Name, cmdParts[2], argument)
						} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && lastFuncRule != nil {
							lastFuncRule.Disabled = argument == "true"
							fmt.Printf("  Updated Package Func Disabled: %t\n", lastFuncRule.Disabled)
						} else if len(cmdParts) == 3 && lastFuncRule != nil {
							// Generic sub-rule for func within package
							fmt.Printf("  Handling package func sub-rule: %s, argument: %s\n", cmdParts[2], argument)
							handleRule(&lastFuncRule.RuleSet, lastFuncRule.Name, cmdParts[2], argument)
						} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && lastVarRule != nil {
							lastVarRule.Disabled = argument == "true"
							fmt.Printf("  Updated Package Var Disabled: %t\n", lastVarRule.Disabled)
						} else if len(cmdParts) == 3 && lastVarRule != nil {
							// Generic sub-rule for var within package
							fmt.Printf("  Handling package var sub-rule: %s, argument: %s\n", cmdParts[2], argument)
							handleRule(&lastVarRule.RuleSet, lastVarRule.Name, cmdParts[2], argument)
						} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && lastConstRule != nil {
							lastConstRule.Disabled = argument == "true"
							fmt.Printf("  Updated Package Const Disabled: %t\n", lastConstRule.Disabled)
						} else if len(cmdParts) == 3 && lastConstRule != nil {
							// Generic sub-rule for const within package
							fmt.Printf("  Handling package const sub-rule: %s, argument: %s\n", cmdParts[2], argument)
							handleRule(&lastConstRule.RuleSet, lastConstRule.Name, cmdParts[2], argument)
						} else {
							return nil, fmt.Errorf("line %d: unknown package sub-directive '%s'", fset.Position(comment.Pos()).Line, command)
						}
					}
				}

			case "type":
				fmt.Printf("  Processing 'type' directive. cmdParts: %+v, argument: %s\n", cmdParts, argument)
				if len(cmdParts) == 1 {
					rule := &config.TypeRule{Name: argument, Kind: "type", RuleSet: config.RuleSet{}}
					cfg.Types = append(cfg.Types, rule)
					lastTypeRule = rule
					applyPendingIgnore(&rule.RuleSet) // Apply pending ignore to type rule
					lastFuncRule, lastVarRule, lastConstRule, lastMemberRule, lastPackageRule = nil, nil, nil, nil, nil
					fmt.Printf("  Added Type: Name=%s. Current Types: %+v\n", argument, cfg.Types)
					fmt.Println("  Resetting last rules for new type.")
				} else if len(cmdParts) == 2 && cmdParts[1] == "struct" {
					if lastTypeRule == nil {
						return nil, fmt.Errorf("line %d: 'type:struct' must follow a 'type' directive", fset.Position(comment.Pos()).Line)
					}
					lastTypeRule.Pattern = argument
					lastTypeRule.Kind = "struct" // Set Kind to struct
					fmt.Printf("  Updated Type Pattern: %s, Kind: %s\n", lastTypeRule.Pattern, lastTypeRule.Kind)
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastTypeRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a 'type' directive", fset.Position(comment.Pos()).Line)
					}
					lastTypeRule.Disabled = argument == "true"
					fmt.Printf("  Updated Type Disabled: %t\n", lastTypeRule.Disabled)
				} else if len(cmdParts) == 2 {
					// Generic sub-rule for type (e.g., :rename, :explicit)
					if lastTypeRule != nil {
						fmt.Printf("  Handling type sub-rule: %s, argument: %s\n", cmdParts[1], argument)
						handleRule(&lastTypeRule.RuleSet, lastTypeRule.Name, cmdParts[1], argument)
					}
				}

			case "func":
				fmt.Printf("  Processing 'func' directive. cmdParts: %+v, argument: %s\n", cmdParts, argument)
				if len(cmdParts) == 1 {
					rule := &config.FuncRule{Name: argument, RuleSet: config.RuleSet{}}
					cfg.Functions = append(cfg.Functions, rule)
					lastFuncRule = rule
					applyPendingIgnore(&rule.RuleSet) // Apply pending ignore to func rule
					lastTypeRule, lastVarRule, lastConstRule, lastMemberRule, lastPackageRule = nil, nil, nil, nil, nil
					fmt.Printf("  Added Function: Name=%s. Current Functions: %+v\n", argument, cfg.Functions)
					fmt.Println("  Resetting last rules for new function.")
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastFuncRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a 'func' directive", fset.Position(comment.Pos()).Line)
					}
					lastFuncRule.Disabled = argument == "true"
					fmt.Printf("  Updated Function Disabled: %t\n", lastFuncRule.Disabled)
				} else if len(cmdParts) == 2 && lastFuncRule != nil {
					// Generic sub-rule for func (e.g., :rename, :explicit)
					fmt.Printf("  Handling func sub-rule: %s, argument: %s\n", cmdParts[1], argument)
					handleRule(&lastFuncRule.RuleSet, lastFuncRule.Name, cmdParts[1], argument)
				}

			case "var":
				fmt.Printf("  Processing 'var' directive. cmdParts: %+v, argument: %s\n", cmdParts, argument)
				if len(cmdParts) == 1 {
					rule := &config.VarRule{Name: argument, RuleSet: config.RuleSet{}}
					cfg.Variables = append(cfg.Variables, rule)
					lastVarRule = rule
					applyPendingIgnore(&rule.RuleSet) // Apply pending ignore to var rule
					lastTypeRule, lastFuncRule, lastConstRule, lastMemberRule = nil, nil, nil, nil
					fmt.Printf("  Added Var: Name=%s. Current Vars: %+v\n", argument, cfg.Variables)
					fmt.Println("  Resetting last rules for new var.")
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastVarRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a 'var' directive", fset.Position(comment.Pos()).Line)
					}
					lastVarRule.Disabled = argument == "true"
					fmt.Printf("  Updated Var Disabled: %t\n", lastVarRule.Disabled)
				} else if len(cmdParts) == 2 && lastVarRule != nil {
					// Generic sub-rule for var (e.g., :rename, :explicit)
					fmt.Printf("  Handling var sub-rule: %s, argument: %s\n", cmdParts[1], argument)
					handleRule(&lastVarRule.RuleSet, lastVarRule.Name, cmdParts[1], argument)
				}

			case "const":
				fmt.Printf("  Processing 'const' directive. cmdParts: %+v, argument: %s\n", cmdParts, argument)
				if len(cmdParts) == 1 {
					rule := &config.ConstRule{Name: argument, RuleSet: config.RuleSet{}}
					cfg.Constants = append(cfg.Constants, rule)
					lastConstRule = rule
					applyPendingIgnore(&rule.RuleSet) // Apply pending ignore to const rule
					lastTypeRule, lastFuncRule, lastVarRule, lastMemberRule = nil, nil, nil, nil
					fmt.Printf("  Added Const: Name=%s. Current Constants: %+v\n", argument, cfg.Constants)
					fmt.Println("  Resetting last rules for new const.")
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastConstRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a 'const' directive", fset.Position(comment.Pos()).Line)
					}
					lastConstRule.Disabled = argument == "true"
					fmt.Printf("  Updated Const Disabled: %t\n", lastConstRule.Disabled)
				} else if len(cmdParts) == 2 && lastConstRule != nil {
					// Generic sub-rule for const (e.g., :rename, :explicit)
					fmt.Printf("  Handling const sub-rule: %s, argument: %s\n", cmdParts[1], argument)
					handleRule(&lastConstRule.RuleSet, lastConstRule.Name, cmdParts[1], argument)
				}

			case "method", "field":
				fmt.Printf("  Processing '%s' directive. cmdParts: %+v, argument: %s\n", baseCmd, cmdParts, argument)
				if lastTypeRule == nil {
					return nil, fmt.Errorf("line %d: '%s' directive must follow a 'type' directive", fset.Position(comment.Pos()).Line, baseCmd)
				}
				if len(cmdParts) == 1 {
					member := &config.MemberRule{Name: argument, RuleSet: config.RuleSet{}}
					if baseCmd == "method" {
						lastTypeRule.Methods = append(lastTypeRule.Methods, member)
						fmt.Printf("  Added Method: Name=%s. Current Methods: %+v\n", argument, lastTypeRule.Methods)
					} else {
						lastTypeRule.Fields = append(lastTypeRule.Fields, member)
						fmt.Printf("  Added Field: Name=%s. Current Fields: %+v\n", argument, lastTypeRule.Fields)
					}
					lastMemberRule = member
					applyPendingIgnore(&member.RuleSet) // Apply pending ignore to member rule
					fmt.Printf("  Added Member: Name=%s. Current Methods/Fields: %+v\n", argument, lastTypeRule.Methods)
					fmt.Println("  Resetting last rules for new member.")
				} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
					if lastMemberRule == nil {
						return nil, fmt.Errorf("line %d: ':disabled' must follow a member directive", fset.Position(comment.Pos()).Line)
					}
					lastMemberRule.Disabled = argument == "true"
					fmt.Printf("  Updated Member Disabled: %t\n", lastMemberRule.Disabled)
				} else if len(cmdParts) == 2 && lastMemberRule != nil {
					// Generic sub-rule for method/field (e.g., :rename, :explicit)
					fmt.Printf("  Handling member sub-rule: %s, argument: %s\n", cmdParts[1], argument)
					handleRule(&lastMemberRule.RuleSet, lastMemberRule.Name, cmdParts[1], argument)
				}
			case "context":
				fmt.Printf("  Processing 'context' directive. argument: %s\n", argument)
				// For now, context directives are handled by the loader, not the parser.
				// This parser only extracts directives from *this* file.
				// We will ignore this for now.
			case "done":
				fmt.Printf("  Processing 'done' directive. argument: %s\n", argument)
				// For now, done directives are handled by the loader, not the parser.
				// This parser only extracts directives from *this* file.
				// We will ignore this for now.
			}
		}
	}

	fmt.Println("Finished ParseFileDirectives.")
	return cfg, nil
}

// parsingContext holds the current default package information for directive resolution.
type parsingContext struct {
	DefaultPackageImportPath string
	DefaultPackageAlias      string
}

// handleRule applies a sub-rule to the appropriate ruleset.
func handleRule(ruleset *config.RuleSet, fromName, ruleName, argument string) {
	fmt.Printf("  handleRule called: ruleName=%s, fromName=%s, argument=%s\n", ruleName, fromName, argument)
	if ruleset == nil {
		fmt.Println("  ruleset is nil in handleRule. Skipping.")
		return // Should not happen if logic is correct
	}

	switch ruleName {
	case "rename":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: fromName, To: argument})
		fmt.Printf("  Added Explicit Rule (rename): From=%s, To=%s. Current Explicit: %+v\n", fromName, argument, ruleset.Explicit)
	case "explicit":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		explicitRules := strings.SplitN(argument, " ", 2)
		if len(explicitRules) == 2 {
			ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: explicitRules[0], To: explicitRules[1]})
			fmt.Printf("  Added Explicit Rule: From=%s, To=%s. Current Explicit: %+v\n", explicitRules[0], explicitRules[1], ruleset.Explicit)
		}
	case "explicit.json":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		var explicitRules []*config.ExplicitRule
		if err := json.Unmarshal([]byte(argument), &ruleset.Explicit); err == nil {
			ruleset.Explicit = explicitRules
			fmt.Printf("  Added Explicit Rules from JSON. Current Explicit: %+v\n", ruleset.Explicit)
		}
	}
}
