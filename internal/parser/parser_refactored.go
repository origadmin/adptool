package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// ContextEntry represents a single entry in the context stack.
type ContextEntry struct {
	Name       string
	IsExplicit bool // True if this context was explicitly started by a :context Directive
}

// Context manages the current parsing context
type Context struct {
	Directive         *Directive
	Config            *config.Config
	CurrentPackage    *config.Package
	CurrentTypeRule   *config.TypeRule
	CurrentFuncRule   *config.FuncRule
	CurrentVarRule    *config.VarRule
	CurrentConstRule  *config.ConstRule
	CurrentMemberRule *config.MemberRule
}

func (c *Context) CurrentRule() string {
	switch {
	case c.CurrentPackage != nil:
		return "package"
	case c.CurrentTypeRule != nil:
		return "type"
	case c.CurrentFuncRule != nil:
		return "function"
	case c.CurrentVarRule != nil:
		return "var"
	case c.CurrentMemberRule != nil:
		return "member"
	case c.CurrentConstRule != nil:
		return "const"
	default:
		return "unknown"
	}

}
func (c *Context) Reset() {
	c.CurrentTypeRule = nil
	c.CurrentFuncRule = nil
	c.CurrentVarRule = nil
	c.CurrentConstRule = nil
	c.CurrentMemberRule = nil
}

func NewContext() *Context {
	return &Context{
		Config: config.New(),
	}
}

// handleDefaultsDirective handles the parsing of defaults directives.
func handleDefaultsDirective(context *Context, cmdParts []string, argument string) error {
	if len(cmdParts) < 2 || cmdParts[1] != "mode" || len(cmdParts) < 3 {
		return fmt.Errorf("line %d: invalid defaults Directive format. Expected 'defaults:mode:<field> <value>'",
			context.Directive.Line)
	}
	if context.Config.Defaults == nil {
		context.Config.Defaults = &config.Defaults{}
	}
	if context.Config.Defaults.Mode == nil {
		context.Config.Defaults.Mode = &config.Mode{}
	}
	modeField := cmdParts[2]
	switch modeField {
	case "strategy":
		context.Config.Defaults.Mode.Strategy = argument
	case "prefix":
		context.Config.Defaults.Mode.Prefix = argument
	case "suffix":
		context.Config.Defaults.Mode.Suffix = argument
	case "explicit":
		context.Config.Defaults.Mode.Explicit = argument
	case "regex":
		context.Config.Defaults.Mode.Regex = argument
	case "ignores":
		context.Config.Defaults.Mode.Ignores = argument
	default:
		return fmt.Errorf("line %d: unknown defaults mode field '%s'", context.Directive.Line, modeField)
	}
	return nil
}

// handleVarsDirective handles the parsing of vars directives.
func handleVarsDirective(context *Context, cmdParts []string, argument string) error {
	if len(cmdParts) != 1 {
		return fmt.Errorf("line %d: invalid vars Directive format. Expected 'vars <name> <value>'", context.Directive.Line)
	}
	varNameParts := strings.SplitN(argument, " ", 2)
	if len(varNameParts) != 2 {
		return fmt.Errorf("line %d: invalid vars Directive argument. Expected 'name value'", context.Directive.Line)
	}
	entry := &config.PropsEntry{Name: varNameParts[0], Value: varNameParts[1]}
	if context.Config.Props == nil {
		context.Config.Props = make([]*config.PropsEntry, 0)
	}
	context.Config.Props = append(context.Config.Props, entry)
	return nil
}

// handlePackageDirective handles the parsing of package directives.
func handlePackageDirective(context *Context, cmdParts []string, argument string) error {
	if context.Directive.BaseCmd != "package" {
		return fmt.Errorf("line %d: invalid package Directive format. Expected 'package <import_path>' or 'package <import_path> <alias>'", context.Directive.Line)
	}
	if len(cmdParts) == 1 { // //go:adapter:package <import_path> [alias] or //go:adapter:package <import_path>
		pkgParts := strings.SplitN(argument, " ", 2)
		if len(pkgParts) == 2 { // Context-setting form: //go:adapter:package <import_path> <alias>
			//context := &Context{
			//	CurrentPackageImportPath: pkgParts[0],
			//	CurrentPackageAlias:      pkgParts[1],
			//}
			context.CurrentPackage = &config.Package{Import: pkgParts[0], Alias: pkgParts[1]}
		} else if len(pkgParts) == 1 { // Config-adding form: //go:adapter:package <import_path>
			if context.Config.Packages == nil {
				context.Config.Packages = make([]*config.Package, 0)
			}
			pkg := &config.Package{Import: argument}
			context.Config.Packages = append(context.Config.Packages, pkg)
			context.CurrentPackage = pkg
			//applyPendingIgnore(&pkg.RuleSet)
			// Reset other last rules
			//context.CurrentTypeRule, context.CurrentFuncRule, context.CurrentVarRule, context.CurrentConstRule, context.CurrentMemberRule = nil, nil, nil, nil, nil
		} else {
			return fmt.Errorf("line %d: invalid package Directive format. Expected 'package <import_path>' or 'package <import_path> <alias>'", context.Directive.Line)
		}
	} else if len(cmdParts) > 1 && context.CurrentPackage != nil {
		subCmd := cmdParts[1]
		switch subCmd {
		case "alias":
			context.CurrentPackage.Alias = argument
		case "path":
			context.CurrentPackage.Path = argument
		case "prop":
			varNameParts := strings.SplitN(argument, " ", 2)
			if len(varNameParts) != 2 {
				return fmt.Errorf("line %d: invalid package vars Directive argument. Expected 'name value'", context.Directive.Line)
			}
			entry := &config.PropsEntry{Name: varNameParts[0], Value: varNameParts[1]}
			if context.CurrentPackage.Props == nil {
				context.CurrentPackage.Props = make([]*config.PropsEntry, 0)
			}
			context.CurrentPackage.Props = append(context.CurrentPackage.Props, entry)
		case "type":
			rule := &config.TypeRule{Name: argument, RuleSet: config.RuleSet{}}
			if context.CurrentPackage.Types == nil {
				context.CurrentPackage.Types = make([]*config.TypeRule, 0)
			}
			context.CurrentPackage.Types = append(context.CurrentPackage.Types, rule)
			context.Reset()
			context.CurrentTypeRule = rule
		case "function":
			rule := &config.FuncRule{Name: argument, RuleSet: config.RuleSet{}}
			if context.CurrentPackage.Functions == nil {
				context.CurrentPackage.Functions = make([]*config.FuncRule, 0)
			}
			context.CurrentPackage.Functions = append(context.CurrentPackage.Functions, rule)
			context.Reset()
			context.CurrentFuncRule = rule
		case "variable":
			rule := &config.VarRule{Name: argument, RuleSet: config.RuleSet{}}
			if context.CurrentPackage.Variables == nil {
				context.CurrentPackage.Variables = make([]*config.VarRule, 0)
			}
			context.CurrentPackage.Variables = append(context.CurrentPackage.Variables, rule)
			context.Reset()
			context.CurrentVarRule = rule
		case "constant":
			rule := &config.ConstRule{Name: argument, RuleSet: config.RuleSet{}}
			if context.CurrentPackage.Constants == nil {
				context.CurrentPackage.Constants = make([]*config.ConstRule, 0)
			}
			context.CurrentPackage.Constants = append(context.CurrentPackage.Constants, rule)
			context.Reset()
			context.CurrentConstRule = rule
		default:
			// Handle sub-directives of types/functions/variables/constants within a package
			if len(cmdParts) == 3 && cmdParts[2] == "struct" && context.CurrentTypeRule != nil {
				context.CurrentTypeRule.Pattern = argument
				context.CurrentTypeRule.Kind = "struct"
			} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && context.CurrentTypeRule != nil {
				context.CurrentTypeRule.Disabled = argument == "true"
			} else if len(cmdParts) == 3 && context.CurrentTypeRule != nil {
				// Generic sub-rule for type within package
				handleRule(&context.CurrentTypeRule.RuleSet, context.CurrentTypeRule.Name, cmdParts[2], argument)
			} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && context.CurrentFuncRule != nil {
				context.CurrentFuncRule.Disabled = argument == "true"
			} else if len(cmdParts) == 3 && context.CurrentFuncRule != nil {
				// Generic sub-rule for func within package
				handleRule(&context.CurrentFuncRule.RuleSet, context.CurrentFuncRule.Name, cmdParts[2], argument)
			} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && context.CurrentVarRule != nil {
				context.CurrentVarRule.Disabled = argument == "true"
			} else if len(cmdParts) == 3 && context.CurrentVarRule != nil {
				// Generic sub-rule for var within package
				handleRule(&context.CurrentVarRule.RuleSet, context.CurrentVarRule.Name, cmdParts[2], argument)
			} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && context.CurrentConstRule != nil {
				context.CurrentConstRule.Disabled = argument == "true"
			} else if len(cmdParts) == 3 && context.CurrentConstRule != nil {
				// Generic sub-rule for const within package
				handleRule(&context.CurrentConstRule.RuleSet, context.CurrentConstRule.Name, cmdParts[2], argument)
			} else {
				return fmt.Errorf("line %d: unknown package sub-directive '%s'", context.Directive.Line, cmdParts[1])
			}
		}
	}
	return nil
}

// handleTypeDirective handles the parsing of type directives.
func handleTypeDirective(context *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		rule := &config.TypeRule{Name: argument, Kind: "type", RuleSet: config.RuleSet{}}
		context.Config.Types = append(context.Config.Types, rule) // Always add to global
		if context.CurrentPackage != nil {
			context.CurrentPackage.Types = append(context.CurrentPackage.Types, rule)
		}
		context.Reset()
		context.CurrentTypeRule = rule
	} else if len(subCmds) == 1 && subCmds[0] == "struct" {
		if context.CurrentTypeRule == nil {
			return fmt.Errorf("line %d: 'type:struct' must follow a 'type' Directive", context.Directive.Line)
		}
		context.CurrentTypeRule.Pattern = argument
		context.CurrentTypeRule.Kind = "struct"
	} else if len(subCmds) == 1 && subCmds[0] == "disabled" {
		if context.CurrentTypeRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a 'type' Directive", context.Directive.Line)
		}
		context.CurrentTypeRule.Disabled = argument == "true"
	} else if len(subCmds) == 1 {
		// Generic sub-rule for type (e.g., :rename, :explicit)
		if context.CurrentTypeRule != nil {
			handleRule(&context.CurrentTypeRule.RuleSet, context.CurrentTypeRule.Name, subCmds[1], argument)
		}
	}
	return nil
}

// handleFuncDirective handles the parsing of func directives.
func handleFuncDirective(context *Context, cmdParts []string, argument string) error {
	if len(cmdParts) == 1 {
		rule := &config.FuncRule{Name: argument, RuleSet: config.RuleSet{}}
		context.Config.Functions = append(context.Config.Functions, rule) // Always add to global
		if context.CurrentPackage != nil {
			context.CurrentPackage.Functions = append(context.CurrentPackage.Functions, rule)
		}
		context.Reset()
		context.CurrentFuncRule = rule
	} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
		if context.CurrentFuncRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a 'func' Directive", context.Directive.Line)
		}
		context.CurrentFuncRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && context.CurrentFuncRule != nil {
		// Generic sub-rule for func (e.g., :rename, :explicit)
		handleRule(&context.CurrentFuncRule.RuleSet, context.CurrentFuncRule.Name, cmdParts[1], argument)
	}
	return nil
}

// handleVarDirective handles the parsing of var directives.
func handleVarDirective(context *Context, cmdParts []string, argument string) error {
	if len(cmdParts) == 1 {
		rule := &config.VarRule{Name: argument, RuleSet: config.RuleSet{}}
		context.Config.Variables = append(context.Config.Variables, rule) // Always add to global
		if context.CurrentPackage != nil {
			context.CurrentPackage.Variables = append(context.CurrentPackage.Variables, rule)
		}
		context.Reset()
		context.CurrentVarRule = rule
	} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
		if context.CurrentVarRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a 'var' Directive", context.Directive.Line)
		}
		context.CurrentVarRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && context.CurrentVarRule != nil {
		// Generic sub-rule for var (e.g., :rename, :explicit)
		handleRule(&context.CurrentVarRule.RuleSet, context.CurrentVarRule.Name, cmdParts[1], argument)
	}
	return nil
}

// handleConstDirective handles the parsing of const directives.
func handleConstDirective(context *Context, cmdParts []string, argument string) error {
	if len(cmdParts) == 1 {
		rule := &config.ConstRule{Name: argument, RuleSet: config.RuleSet{}}
		context.Config.Constants = append(context.Config.Constants, rule) // Always add to global
		if context.CurrentPackage != nil {
			context.CurrentPackage.Constants = append(context.CurrentPackage.Constants, rule)
		}
		context.Reset()
		context.CurrentConstRule = rule
	} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
		if context.CurrentConstRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a 'const' Directive", context.Directive.Line)
		}
		context.CurrentConstRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && cmdParts[1] == "ignores" {
		if context.CurrentConstRule == nil {
			return fmt.Errorf("line %d: ':ignores' must follow a 'const' Directive", context.Directive.Line)
		}
		context.CurrentConstRule.RuleSet.Ignores = append(context.CurrentConstRule.RuleSet.Ignores, argument)
	} else if len(cmdParts) == 2 && context.CurrentConstRule != nil {
		// Generic sub-rule for const (e.g., :rename, :explicit)
		handleRule(&context.CurrentConstRule.RuleSet, context.CurrentConstRule.Name, cmdParts[1], argument)
	}
	return nil
}

// handleMemberDirective handles the parsing of method and field directives.
func handleMemberDirective(context *Context, baseCmd string, subCmds []string, argument string) error {
	if context.CurrentTypeRule == nil {
		return fmt.Errorf("line %d: '%s' Directive must follow a 'type' Directive", context.Directive.Line, baseCmd)
	}
	if len(subCmds) == 0 {
		member := &config.MemberRule{Name: argument, RuleSet: config.RuleSet{}}
		if baseCmd == "method" {
			context.CurrentTypeRule.Methods = append(context.CurrentTypeRule.Methods, member)
		} else {
			context.CurrentTypeRule.Fields = append(context.CurrentTypeRule.Fields, member)
		}
		context.CurrentMemberRule = member
	} else if len(subCmds) == 1 && subCmds[0] == "disabled" {
		if context.CurrentMemberRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a member Directive", context.Directive.Line)
		}
		context.CurrentMemberRule.Disabled = argument == "true"
	} else if len(subCmds) == 1 && context.CurrentMemberRule != nil {
		// Generic sub-rule for method/field (e.g., :rename, :explicit)
		handleRule(&context.CurrentMemberRule.RuleSet, context.CurrentMemberRule.Name, subCmds[0], argument)
	}
	return nil
}

// handleRule applies a sub-rule to the appropriate ruleset.
func handleRule(ruleset *config.RuleSet, fromName, ruleName, argument string) {
	if ruleset == nil {
		return
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
	case "explicit:json":
		var explicitRules []*config.ExplicitRule
		if err := json.Unmarshal([]byte(argument), &explicitRules); err == nil {
			ruleset.Explicit = explicitRules
		}
	case "regex:json":
		var regexRules []*config.RegexRule
		if err := json.Unmarshal([]byte(argument), &regexRules); err == nil {
			ruleset.Regex = regexRules
		}
	case "strategy:json":
		var strategies []string
		if err := json.Unmarshal([]byte(argument), &strategies); err == nil {
			ruleset.Strategy = strategies
		}
	}
}
