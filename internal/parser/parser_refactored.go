package parser

import (
	"encoding/json"
	"log/slog"
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
	slog.Debug("Entering CurrentRule")
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
	slog.Debug("Entering Reset")
	c.CurrentTypeRule = nil
	c.CurrentFuncRule = nil
	c.CurrentVarRule = nil
	c.CurrentConstRule = nil
	c.CurrentMemberRule = nil
}

func (c *Context) SubContext() *Context {
	slog.Debug("Entering SubContext")
	return &Context{
		Config: c.Config,
	}
}

func NewContext() *Context {
	slog.Debug("Entering NewContext")
	return &Context{
		Config: config.New(),
	}
}

// handleDefaultsDirective handles the parsing of defaults directives.
func handleDefaultsDirective(context *Context, cmdParts []string, argument string) error {
	slog.Debug("Entering handleDefaultsDirective", "cmdParts", cmdParts, "argument", argument)
	if len(cmdParts) < 2 || cmdParts[1] != "mode" || len(cmdParts) < 3 {
		return newDirectiveError(context.Directive, "invalid defaults Directive format. Expected 'defaults:mode:<field> <value>'")
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
		return newDirectiveError(context.Directive, "unknown defaults mode field '%s'", modeField)
	}
	return nil
}

// handleVarsDirective handles the parsing of vars directives.
func handleVarsDirective(context *Context, cmdParts []string, argument string) error {
	slog.Debug("Entering handleVarsDirective", "cmdParts", cmdParts, "argument", argument)
	if len(cmdParts) != 1 {
		return newDirectiveError(context.Directive, "invalid vars Directive format. Expected 'vars <name> <value>'")
	}
	name, value, err := parseNameValue(argument)
	if err != nil {
		return newDirectiveError(context.Directive, "invalid vars Directive argument: %v", err)
	}
	entry := &config.PropsEntry{Name: name, Value: value}
	if context.Config.Props == nil {
		context.Config.Props = make([]*config.PropsEntry, 0)
	}
	context.Config.Props = append(context.Config.Props, entry)
	return nil
}

// handlePackageDirective handles the parsing of package directives.
func handlePackageDirective(context *Context, subCmds []string, argument string) error {
	slog.Debug("Entering handlePackageDirective", "subCmds", subCmds, "argument", argument)
	switch {
	case len(subCmds) == 0:
		pkgParts := strings.SplitN(argument, " ", 2)
		if len(pkgParts) == 2 { // Context-setting form: //go:adapter:package <import_path> <alias>
			context.CurrentPackage = &config.Package{Import: pkgParts[0], Alias: pkgParts[1]}
		} else if len(pkgParts) == 1 { // Config-adding form: //go:adapter:package <import_path>
			if context.Config.Packages == nil {
				context.Config.Packages = make([]*config.Package, 0)
			}
			pkg := &config.Package{Import: argument}
			context.Config.Packages = append(context.Config.Packages, pkg)
			context.CurrentPackage = pkg
		} else {
			return newDirectiveError(context.Directive, "invalid package Directive format. Expected 'package <import_path>' or 'package <import_path> <alias>'")
		}
	case len(subCmds) == 1:
		subCmd := subCmds[0]
		switch subCmd {
		case "alias":
			context.CurrentPackage.Alias = argument
		case "path":
			context.CurrentPackage.Path = argument
		case "prop":
			name, value, err := parseNameValue(argument)
			if err != nil {
				return newDirectiveError(context.Directive, "invalid package vars Directive argument: %v", err)
			}
			entry := &config.PropsEntry{Name: name, Value: value}
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
			// This case is handled by the outer switch's default or subsequent cases
		}
	case len(subCmds) == 2:
		// Handle sub-directives of types/functions/variables/constants within a package
		if subCmds[1] == "struct" && context.CurrentTypeRule != nil {
			context.CurrentTypeRule.Pattern = argument
			context.CurrentTypeRule.Kind = "struct"
		} else if subCmds[1] == "disabled" && context.CurrentTypeRule != nil {
			context.CurrentTypeRule.Disabled = argument == "true"
		} else if context.CurrentTypeRule != nil {
			if err := applyRuleToRuleSet(&context.CurrentTypeRule.RuleSet, context.CurrentTypeRule.Name, subCmds[1], argument); err != nil {
				return newDirectiveError(context.Directive, "failed to apply rule to type '%s': %v", context.CurrentTypeRule.Name, err)
			}
		} else if subCmds[1] == "disabled" && context.CurrentFuncRule != nil {
			context.CurrentFuncRule.Disabled = argument == "true"
		} else if context.CurrentFuncRule != nil {
			if err := applyRuleToRuleSet(&context.CurrentFuncRule.RuleSet, context.CurrentFuncRule.Name, subCmds[1], argument); err != nil {
				return newDirectiveError(context.Directive, "failed to apply rule to function '%s': %v", context.CurrentFuncRule.Name, err)
			}
		} else if subCmds[1] == "disabled" && context.CurrentVarRule != nil {
			context.CurrentVarRule.Disabled = argument == "true"
		} else if context.CurrentVarRule != nil {
			if err := applyRuleToRuleSet(&context.CurrentVarRule.RuleSet, context.CurrentVarRule.Name, subCmds[1], argument); err != nil {
				return newDirectiveError(context.Directive, "failed to apply rule to variable '%s': %v", context.CurrentVarRule.Name, err)
			}
		} else if subCmds[1] == "disabled" && context.CurrentConstRule != nil {
			context.CurrentConstRule.Disabled = argument == "true"
		} else if context.CurrentConstRule != nil {
			if err := applyRuleToRuleSet(&context.CurrentConstRule.RuleSet, context.CurrentConstRule.Name, subCmds[1], argument); err != nil {
				return newDirectiveError(context.Directive, "failed to apply rule to constant '%s': %v", context.CurrentConstRule.Name, err)
			}
		} else {
			return newDirectiveError(context.Directive, "unknown package sub-directive '%s'", subCmds[1])
		}
	case len(subCmds) > 2:
		if subCmds[1] == "method" {
			if context.CurrentTypeRule.Methods == nil {
				context.CurrentTypeRule.Methods = make([]*config.MemberRule, 0)
			}
			// Pass the correct subCmds slice for handleMemberDirective
			return handleMemberDirective(context, subCmds[1], subCmds[2:], argument)
		}
		return newDirectiveError(context.Directive, "unknown package sub-directive '%s'", subCmds[0])
	default:
		return newDirectiveError(context.Directive, "unknown package directive '%s'", subCmds[0])
	}
	return nil
}

// handleTypeDirective handles the parsing of type directives.
func handleTypeDirective(context *Context, subCmds []string, argument string) error {
	slog.Debug("Entering handleTypeDirective", "subCmds", subCmds, "argument", argument)
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
			return newDirectiveError(context.Directive, "'type:struct' must follow a 'type' Directive")
		}
		context.CurrentTypeRule.Pattern = argument
		context.CurrentTypeRule.Kind = "struct"
	} else if len(subCmds) == 1 && subCmds[0] == "disabled" {
		if context.CurrentTypeRule == nil {
			return newDirectiveError(context.Directive, ":disabled' must follow a 'type' Directive")
		}
		context.CurrentTypeRule.Disabled = argument == "true"
	} else if len(subCmds) == 1 {
		// Generic sub-rule for type (e.g., :rename, :explicit)
		if context.CurrentTypeRule != nil {
			if err := applyRuleToRuleSet(&context.CurrentTypeRule.RuleSet, context.CurrentTypeRule.Name, subCmds[0], argument); err != nil {
				return newDirectiveError(context.Directive, "failed to apply rule to type '%s': %v", context.CurrentTypeRule.Name, err)
			}
		}
	}
	return nil
}

// handleFuncDirective handles the parsing of func directives.
func handleFuncDirective(context *Context, cmdParts []string, argument string) error {
	slog.Debug("Entering handleFuncDirective", "cmdParts", cmdParts, "argument", argument)
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
			return newDirectiveError(context.Directive, ":disabled' must follow a 'func' Directive")
		}
		context.CurrentFuncRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && context.CurrentFuncRule != nil {
		// Generic sub-rule for func (e.g., :rename, :explicit)
		if err := applyRuleToRuleSet(&context.CurrentFuncRule.RuleSet, context.CurrentFuncRule.Name, cmdParts[1], argument); err != nil {
			return newDirectiveError(context.Directive, "failed to apply rule to function '%s': %v", context.CurrentFuncRule.Name, err)
		}
	}
	return nil
}

// handleVarDirective handles the parsing of var directives.
func handleVarDirective(context *Context, cmdParts []string, argument string) error {
	slog.Debug("Entering handleVarDirective", "cmdParts", cmdParts, "argument", argument)
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
			return newDirectiveError(context.Directive, ":disabled' must follow a 'var' Directive")
		}
		context.CurrentVarRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && context.CurrentVarRule != nil {
		// Generic sub-rule for var (e.g., :rename, :explicit)
		if err := applyRuleToRuleSet(&context.CurrentVarRule.RuleSet, context.CurrentVarRule.Name, cmdParts[1], argument); err != nil {
			return newDirectiveError(context.Directive, "failed to apply rule to variable '%s': %v", context.CurrentVarRule.Name, err)
		}
	}
	return nil
}

// handleConstDirective handles the parsing of const directives.
func handleConstDirective(context *Context, cmdParts []string, argument string) error {
	slog.Debug("Entering handleConstDirective", "cmdParts", cmdParts, "argument", argument)
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
			return newDirectiveError(context.Directive, ":disabled' must follow a 'const' Directive")
		}
		context.CurrentConstRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && cmdParts[1] == "ignores" {
		if context.CurrentConstRule == nil {
			return newDirectiveError(context.Directive, ":ignores' must follow a 'const' Directive")
		}
		context.CurrentConstRule.RuleSet.Ignores = append(context.CurrentConstRule.RuleSet.Ignores, argument)
	} else if len(cmdParts) == 2 && context.CurrentConstRule != nil {
		// Generic sub-rule for const (e.g., :rename, :explicit)
		if err := applyRuleToRuleSet(&context.CurrentConstRule.RuleSet, context.CurrentConstRule.Name, cmdParts[1], argument); err != nil {
			return newDirectiveError(context.Directive, "failed to apply rule to constant '%s': %v", context.CurrentConstRule.Name, err)
		}
	}
	return nil
}

// handleMemberDirective handles the parsing of method and field directives.
func handleMemberDirective(context *Context, baseCmd string, subCmds []string, argument string) error {
	slog.Debug("Entering handleMemberDirective", "baseCmd", baseCmd, "subCmds", subCmds, "argument", argument)
	if context.CurrentTypeRule == nil {
		return newDirectiveError(context.Directive, "'%s' Directive must follow a 'type' Directive", baseCmd)
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
			return newDirectiveError(context.Directive, ":disabled' must follow a member Directive")
		}
		context.CurrentMemberRule.Disabled = argument == "true"
	} else if len(subCmds) == 1 && context.CurrentMemberRule != nil {
		// Generic sub-rule for method/field (e.g., :rename, :explicit)
		if err := applyRuleToRuleSet(&context.CurrentMemberRule.RuleSet, context.CurrentMemberRule.Name, subCmds[0], argument); err != nil {
			return newDirectiveError(context.Directive, "failed to apply rule to member '%s': %v", context.CurrentMemberRule.Name, err)
		}
	}
	return nil
}

// handleRule applies a sub-rule to the appropriate ruleset.
func handleRule(ruleset *config.RuleSet, fromName, ruleName, argument string) {
	slog.Debug("Entering handleRule", "fromName", fromName, "ruleName", ruleName, "argument", argument)
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
