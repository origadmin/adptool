package parser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// --- Default and Prop Handlers ---

func handleDefaultDirective(context *Context, subCmds []string, argument string) error {
	if len(subCmds) < 1 || subCmds[0] != "mode" || len(subCmds) < 2 {
		return newDirectiveError(context.Directive, "invalid defaults Directive format. Expected 'default:mode:<field> <value>'")
	}
	if context.Config.Defaults == nil {
		context.Config.Defaults = &config.Defaults{}
	}
	if context.Config.Defaults.Mode == nil {
		context.Config.Defaults.Mode = &config.Mode{}
	}
	modeField := subCmds[1]
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

func handleVarsDirective(context *Context, subCmds []string, argument string) error {
	if len(subCmds) != 0 {
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

// --- Core Handlers ---

func handlePackageDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		pkgParts := strings.SplitN(argument, " ", 2)
		var pkg *config.Package
		if len(pkgParts) == 2 {
			pkg = &config.Package{Import: pkgParts[0], Alias: pkgParts[1]}
		} else {
			pkg = &config.Package{Import: argument}
		}

		if ctx.Config.Packages == nil {
			ctx.Config.Packages = make([]*config.Package, 0)
		}
		ctx.Config.Packages = append(ctx.Config.Packages, pkg)
		ctx.CurrentPackage = pkg
		ctx.ResetActiveRule()
		return nil
	}

	if ctx.CurrentPackage == nil {
		return newDirectiveError(ctx.Directive, "'package:%s' must follow a 'package' Directive", subCmds[0])
	}

	switch subCmds[0] {
	case "alias":
		ctx.CurrentPackage.Alias = argument
	case "path":
		ctx.CurrentPackage.Path = argument
	case "prop":
		name, value, err := parseNameValue(argument)
		if err != nil {
			return newDirectiveError(ctx.Directive, "invalid package prop argument: %v", err)
		}
		if ctx.CurrentPackage.Props == nil {
			ctx.CurrentPackage.Props = make([]*config.PropsEntry, 0)
		}
		ctx.CurrentPackage.Props = append(ctx.CurrentPackage.Props, &config.PropsEntry{Name: name, Value: value})
	case "type":
		return handleTypeDirective(ctx, subCmds[1:], argument)
	case "function", "func":
		return handleFuncDirective(ctx, subCmds[1:], argument)
	case "variable", "var":
		return handleVarDirective(ctx, subCmds[1:], argument)
	case "constant", "const":
		return handleConstDirective(ctx, subCmds[1:], argument)
	default:
		return newDirectiveError(ctx.Directive, "unknown package sub-directive '%s'", subCmds[0])
	}
	return nil
}

func handleTypeDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		configRule := &config.TypeRule{Name: argument, Kind: "type"}
		rule := &TypeRule{TypeRule: configRule}
		ctx.AddTypeRule(configRule)
		ctx.SetActiveRule(rule)
		return nil
	}

	activeRule := ctx.ActiveRule()
	if activeRule == nil {
		return newDirectiveError(ctx.Directive, "':%s' must follow a 'type' Directive", subCmds[0])
	}

	return activeRule.ApplySubDirective(ctx, subCmds, argument)
}

func handleFuncDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		configRule := &config.FuncRule{Name: argument}
		rule := &FuncRule{FuncRule: configRule}
		ctx.AddFuncRule(configRule)
		ctx.SetActiveRule(rule)
		return nil
	}

	activeRule := ctx.ActiveRule()
	if activeRule == nil {
		return newDirectiveError(ctx.Directive, "':%s' must follow a 'func' Directive", subCmds[0])
	}

	return activeRule.ApplySubDirective(ctx, subCmds, argument)
}

func handleVarDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		configRule := &config.VarRule{Name: argument}
		rule := &VarRule{VarRule: configRule}
		ctx.AddVarRule(configRule)
		ctx.SetActiveRule(rule)
		return nil
	}

	activeRule := ctx.ActiveRule()
	if activeRule == nil {
		return newDirectiveError(ctx.Directive, "':%s' must follow a 'var' Directive", subCmds[0])
	}

	return activeRule.ApplySubDirective(ctx, subCmds, argument)
}

func handleConstDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		configRule := &config.ConstRule{Name: argument}
		rule := &ConstRule{ConstRule: configRule}
		ctx.AddConstRule(configRule)
		ctx.SetActiveRule(rule)
		return nil
	}

	activeRule := ctx.ActiveRule()
	if activeRule == nil {
		return newDirectiveError(ctx.Directive, "':%s' must follow a 'const' Directive", subCmds[0])
	}

	return activeRule.ApplySubDirective(ctx, subCmds, argument)
}

// applyRuleToRuleSet applies a sub-rule to the provided ruleset.
func applyRuleToRuleSet(ruleset *config.RuleSet, fromName, ruleName, argument string) error {
	slog.Debug("Applying rule to ruleset", "fromName", fromName, "ruleName", ruleName, "argument", argument)
	if ruleset == nil {
		return fmt.Errorf("cannot apply rule to a nil ruleset")
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
		} else {
			return fmt.Errorf("explicit rule argument must be in 'from to' format")
		}
	case "explicit:json":
		var explicitRules []*config.ExplicitRule
		if err := json.Unmarshal([]byte(argument), &explicitRules); err == nil {
			ruleset.Explicit = explicitRules
		} else {
			return fmt.Errorf("invalid JSON for explicit:json: %w", err)
		}
	case "regex:json":
		var regexRules []*config.RegexRule
		if err := json.Unmarshal([]byte(argument), &regexRules); err == nil {
			ruleset.Regex = regexRules
		} else {
			return fmt.Errorf("invalid JSON for regex:json: %w", err)
		}
	case "strategy:json":
		var strategies []string
		if err := json.Unmarshal([]byte(argument), &strategies); err == nil {
			ruleset.Strategy = strategies
		} else {
			return fmt.Errorf("invalid JSON for strategy:json: %w", err)
		}
	case "ignores":
		ruleset.Ignores = append(ruleset.Ignores, argument)
	default:
		return fmt.Errorf("unknown rule name: %s", ruleName)
	}
	return nil
}
