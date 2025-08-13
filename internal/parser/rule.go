package parser

import "github.com/origadmin/adptool/internal/config"

// Rule defines the common behavior for all directive rules, allowing for polymorphic
// handling of sub-directives.
type Rule interface {
	// ApplySubDirective applies a sub-command (e.g., ":rename", ":disabled") to the rule.
	// It takes the builder to interact with the broader parsing state if necessary (e.g., to set an active member).
	ApplySubDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error
}

// --- Wrapper Implementations ---

// TypeRule is a wrapper around config.TypeRule to implement the Rule interface.
type TypeRule struct {
	*config.TypeRule
}

// ApplySubDirective applies a sub-command to the TypeRule.
func (r *TypeRule) ApplySubDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	if len(subCmds) == 0 {
		return newDirectiveError(d, "unexpected empty sub-commands for type rule")
	}

	subCmd := subCmds[0]
	switch subCmd {
	case "struct":
		r.Pattern = argument
		r.Kind = "struct"
	case "disabled":
		r.Disabled = argument == "true"
	case "method", "field":
		memberSubCmds := subCmds[1:]
		if len(memberSubCmds) == 0 {
			member := &config.MemberRule{Name: argument}
			if subCmd == "method" {
				r.Methods = append(r.Methods, member)
			} else {
				r.Fields = append(r.Fields, member)
			}
			builder.SetActiveMember(member)
		} else if len(memberSubCmds) == 1 {
			if builder.ActiveMember() == nil {
				return newDirectiveError(d, "':%s' must follow a 'method' or 'field' Directive", memberSubCmds[0])
			}
			if memberSubCmds[0] == "disabled" {
				builder.ActiveMember().Disabled = argument == "true"
			} else {
				return builder.ApplyRuleToRuleSet(&builder.ActiveMember().RuleSet, builder.ActiveMember().Name, memberSubCmds[0], argument)
			}
		} else {
			return newDirectiveError(d, "too many sub-commands for member Directive")
		}
	default:
		return builder.ApplyRuleToRuleSet(&r.RuleSet, r.Name, subCmd, argument)
	}
	return nil
}

// FuncRule is a wrapper around config.FuncRule to implement the Rule interface.
type FuncRule struct {
	*config.FuncRule
}

// ApplySubDirective applies a sub-command to the FuncRule.
func (r *FuncRule) ApplySubDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	return applySimpleRuleSubDirective(&r.RuleSet, &r.Name, &r.Disabled, builder, subCmds, argument, d)
}

// VarRule is a wrapper around config.VarRule to implement the Rule interface.
type VarRule struct {
	*config.VarRule
}

// ApplySubDirective applies a sub-command to the VarRule.
func (r *VarRule) ApplySubDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	return applySimpleRuleSubDirective(&r.RuleSet, &r.Name, &r.Disabled, builder, subCmds, argument, d)
}

// ConstRule is a wrapper around config.ConstRule to implement the Rule interface.
type ConstRule struct {
	*config.ConstRule
}

// ApplySubDirective applies a sub-command to the ConstRule.
func (r *ConstRule) ApplySubDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	return applySimpleRuleSubDirective(&r.RuleSet, &r.Name, &r.Disabled, builder, subCmds, argument, d)
}

// applySimpleRuleSubDirective is a helper for simple rules (Func, Var, Const) that only support
// :disabled and generic ruleset sub-directives.
func applySimpleRuleSubDirective(ruleset *config.RuleSet, name *string, disabled *bool, builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	if len(subCmds) == 0 {
		return newDirectiveError(d, "unexpected empty sub-commands for rule")
	}
	subCmd := subCmds[0]
	if subCmd == "disabled" {
		*disabled = (argument == "true")
	} else {
		return builder.ApplyRuleToRuleSet(ruleset, *name, subCmd, argument)
	}
	return nil
}
