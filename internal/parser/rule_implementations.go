package parser

import (
	"github.com/origadmin/adptool/internal/config"
)

// TypeRule is a wrapper around config.TypeRule to implement the Rule interface.
// This avoids creating a circular dependency between the parser and config packages.
type TypeRule struct {
	*config.TypeRule
}

// ApplySubDirective applies a sub-command to the TypeRule.
func (r *TypeRule) ApplySubDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		return newDirectiveError(ctx.Directive, "unexpected empty sub-commands for type rule")
	}

	subCmd := subCmds[0]
	switch subCmd {
	case "struct":
		r.Pattern = argument
		r.Kind = "struct"
	case "disabled":
		r.Disabled = argument == "true"
	case "method", "field":
		// This logic is now self-contained within the TypeRule.
		memberSubCmds := subCmds[1:]
		if len(memberSubCmds) == 0 {
			member := &config.MemberRule{Name: argument}
			if subCmd == "method" {
				if r.Methods == nil {
					r.Methods = make([]*config.MemberRule, 0)
				}
				r.Methods = append(r.Methods, member)
			} else {
				if r.Fields == nil {
					r.Fields = make([]*config.MemberRule, 0)
				}
				r.Fields = append(r.Fields, member)
			}
			ctx.SetActiveMemberRule(member)
		} else if len(memberSubCmds) == 1 {
			if ctx.ActiveMemberRule() == nil {
				return newDirectiveError(ctx.Directive, "':%s' must follow a 'method' or 'field' Directive", memberSubCmds[0])
			}
			if memberSubCmds[0] == "disabled" {
				ctx.ActiveMemberRule().Disabled = argument == "true"
			} else {
				return applyRuleToRuleSet(&ctx.ActiveMemberRule().RuleSet, ctx.ActiveMemberRule().Name, memberSubCmds[0], argument)
			}
		} else {
			return newDirectiveError(ctx.Directive, "too many sub-commands for member Directive")
		}
	default:
		return applyRuleToRuleSet(&r.RuleSet, r.Name, subCmd, argument)
	}
	return nil
}

// FuncRule is a wrapper around config.FuncRule to implement the Rule interface.
type FuncRule struct {
	*config.FuncRule
}

// ApplySubDirective applies a sub-command to the FuncRule.
func (r *FuncRule) ApplySubDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		return newDirectiveError(ctx.Directive, "unexpected empty sub-commands for func rule")
	}
	subCmd := subCmds[0]
	if subCmd == "disabled" {
		r.Disabled = argument == "true"
	} else {
		return applyRuleToRuleSet(&r.RuleSet, r.Name, subCmd, argument)
	}
	return nil
}

// VarRule is a wrapper around config.VarRule to implement the Rule interface.
type VarRule struct {
	*config.VarRule
}

// ApplySubDirective applies a sub-command to the VarRule.
func (r *VarRule) ApplySubDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		return newDirectiveError(ctx.Directive, "unexpected empty sub-commands for var rule")
	}
	subCmd := subCmds[0]
	if subCmd == "disabled" {
		r.Disabled = argument == "true"
	} else {
		return applyRuleToRuleSet(&r.RuleSet, r.Name, subCmd, argument)
	}
	return nil
}

// ConstRule is a wrapper around config.ConstRule to implement the Rule interface.
type ConstRule struct {
	*config.ConstRule
}

// ApplySubDirective applies a sub-command to the ConstRule.
func (r *ConstRule) ApplySubDirective(ctx *Context, subCmds []string, argument string) error {
	if len(subCmds) == 0 {
		return newDirectiveError(ctx.Directive, "unexpected empty sub-commands for const rule")
	}
	subCmd := subCmds[0]
	if subCmd == "disabled" {
		r.Disabled = argument == "true"
	} else {
		return applyRuleToRuleSet(&r.RuleSet, r.Name, subCmd, argument)
	}
	return nil
}
