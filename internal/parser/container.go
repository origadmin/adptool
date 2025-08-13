package parser

import "github.com/origadmin/adptool/internal/config"

// Container defines the interface for any object that can hold parsed rules
// and participate in the hierarchical configuration structure.
type Container interface {
	// ParseDirective applies a sub-command (e.g., ":rename", ":disabled") to the rule.
	// It takes the builder to interact with the broader parsing state if necessary (e.g., to set an active member).
	ParseDirective(directive *Directive) error

	// AddPackage adds a nested package configuration to this container.
	AddPackage(pkg Container) error

	// AddTypeRule adds a TypeRule to this container.
	AddTypeRule(rule *TypeRule) error
	// AddFuncRule adds a FuncRule to this container.
	AddFuncRule(rule *FuncRule) error
	// AddVarRule adds a VarRule to this container.
	AddVarRule(rule *VarRule) error
	// AddConstRule adds a ConstRule to this container.
	AddConstRule(rule *ConstRule) error
	AddMethodRule(rule *MethodRule) error
	AddFieldRule(rule *FieldRule) error
	// Finalize performs any post-processing or validation for this container
	// after all its direct rules have been added.
	Finalize() error
}

// --- Wrapper Implementations ---

// RootConfig is a wrapper around config.Config to implement the Container interface.
type RootConfig struct {
	*config.Config
}

func (r *RootConfig) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddPackage(pkg Container) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) Finalize() error {
	//TODO implement me
	panic("implement me")
}

// PackageConfig is a wrapper around config.Package to implement the Container interface.
type PackageConfig struct {
	*config.Package
}

func (p PackageConfig) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (p PackageConfig) AddPackage(pkg Container) error {
	//TODO implement me
	panic("implement me")
}

func (p PackageConfig) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (p PackageConfig) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (p PackageConfig) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (p PackageConfig) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (p PackageConfig) Finalize() error {
	//TODO implement me
	panic("implement me")
}

// TypeRule is a wrapper around config.TypeRule to implement the Rule interface.
type TypeRule struct {
	*config.TypeRule
}

func (r *TypeRule) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (r *TypeRule) AddPackage(pkg Container) error {
	//TODO implement me
	panic("implement me")
}

func (r *TypeRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *TypeRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *TypeRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *TypeRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *TypeRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *TypeRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *TypeRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}

// ApplySubDirective applies a sub-command to the TypeRule.
func (r *TypeRule) ApplySubDirective(subCmds []string, argument string, d *Directive) error {
	//if len(subCmds) == 0 {
	//	return newDirectiveError(d, "unexpected empty sub-commands for type rule")
	//}
	//
	//subCmd := subCmds[0]
	//switch subCmd {
	//case "struct":
	//	r.Pattern = argument
	//	r.Kind = "struct"
	//case "disabled":
	//	r.Disabled = argument == "true"
	//case "method", "field":
	//	memberSubCmds := subCmds[1:]
	//	if len(memberSubCmds) == 0 {
	//		member := &config.MemberRule{Name: argument}
	//		if subCmd == "method" {
	//			r.Methods = append(r.Methods, member)
	//		} else {
	//			r.Fields = append(r.Fields, member)
	//		}
	//		builder.SetActiveMember(member)
	//	} else if len(memberSubCmds) == 1 {
	//		if builder.ActiveMember() == nil {
	//			return newDirectiveError(d, "':%s' must follow a 'method' or 'field' Directive", memberSubCmds[0])
	//		}
	//		if memberSubCmds[0] == "disabled" {
	//			builder.ActiveMember().Disabled = argument == "true"
	//		} else {
	//			return builder.ApplyRuleToRuleSet(&builder.ActiveMember().RuleSet, builder.ActiveMember().Name, memberSubCmds[0], argument)
	//		}
	//	} else {
	//		return newDirectiveError(d, "too many sub-commands for member Directive")
	//	}
	//default:
	//	return builder.ApplyRuleToRuleSet(&r.RuleSet, r.Name, subCmd, argument)
	//}
	return nil
}

// FuncRule is a wrapper around config.FuncRule to implement the Rule interface.
type FuncRule struct {
	*config.FuncRule
}

func (r *FuncRule) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddPackage(pkg Container) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}

// ApplySubDirective applies a sub-command to the FuncRule.
func (r *FuncRule) ApplySubDirective(subCmds []string, argument string, d *Directive) error {
	//return applySimpleRuleSubDirective(&r.RuleSet, &r.Name, &r.Disabled, subCmds, argument, d)
	return nil
}

// VarRule is a wrapper around config.VarRule to implement the Rule interface.
type VarRule struct {
	*config.VarRule
}

// ApplySubDirective applies a sub-command to the VarRule.
func (r *VarRule) ApplySubDirective(subCmds []string, argument string, d *Directive) error {
	//return applySimpleRuleSubDirective(&r.RuleSet, &r.Name, &r.Disabled, builder, subCmds, argument, d)
	return nil
}

// ConstRule is a wrapper around config.ConstRule to implement the Rule interface.
type ConstRule struct {
	*config.ConstRule
}

// ApplySubDirective applies a sub-command to the ConstRule.
func (r *ConstRule) ApplySubDirective(subCmds []string, argument string, d *Directive) error {
	//return applySimpleRuleSubDirective(&r.RuleSet, &r.Name, &r.Disabled, builder, subCmds, argument, d)
	return nil
}

// applySimpleRuleSubDirective is a helper for simple rules (Func, Var, Const) that only support
// :disabled and generic ruleset sub-directives.
func applySimpleRuleSubDirective(ruleset *config.RuleSet, name *string, disabled *bool, subCmds []string, argument string, d *Directive) error {
	if len(subCmds) == 0 {
		return newDirectiveError(d, "unexpected empty sub-commands for rule")
	}
	subCmd := subCmds[0]
	if subCmd == "disabled" {
		*disabled = argument == "true"
	} else {
		//return builder.ApplyRuleToRuleSet(ruleset, *name, subCmd, argument)
	}
	return nil
}

type MethodRule struct {
	*config.MemberRule
}

func (m *MethodRule) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddPackage(pkg Container) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}

type FieldRule struct {
	*config.MemberRule
}

func (f *FieldRule) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldRule) AddPackage(pkg Container) error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (f *FieldRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}
