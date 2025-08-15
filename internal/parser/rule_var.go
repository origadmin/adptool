package parser

import (
	"github.com/origadmin/adptool/internal/config"
)

// VarRule is a wrapper around config.VarRule to implement the Container interface.
type VarRule struct {
	*config.VarRule
}

func (r *VarRule) Type() RuleType {
	return RuleTypeVar
}

func (r *VarRule) ParseDirective(directive *Directive) error {
	if directive.BaseCmd != "var" && directive.BaseCmd != "variable" {
		return NewParserErrorWithContext(directive, "VarRule can only contain var directives")
	}
	if !directive.HasSub() {
		if directive.Argument == "" {
			return NewParserErrorWithContext(directive, "var directive requires an argument (name)")
		}
		r.VarRule.Name = directive.Argument
		return nil
	}

	// Delegate to the common RuleSet parser
	return parseRuleSetDirective(&r.RuleSet, directive.Sub())
}

func (r *VarRule) AddPackage(pkg *PackageRule) error {
	return NewParserErrorWithContext(r, "VarRule cannot contain a PackageRule")
}

func (r *VarRule) AddTypeRule(rule *TypeRule) error {
	return NewParserErrorWithContext(r, "VarRule cannot contain a TypeRule")
}

func (r *VarRule) AddFuncRule(rule *FuncRule) error {
	return NewParserErrorWithContext(r, "VarRule cannot contain a FuncRule")
}

func (r *VarRule) AddVarRule(rule *VarRule) error {
	return NewParserErrorWithContext(r, "VarRule cannot contain a VarRule")
}

func (r *VarRule) AddConstRule(rule *ConstRule) error {
	return NewParserErrorWithContext(r, "VarRule cannot contain a ConstRule")
}

func (r *VarRule) AddMethodRule(rule *MethodRule) error {
	return NewParserErrorWithContext(r, "VarRule cannot contain a MethodRule")
}

func (r *VarRule) AddFieldRule(rule *FieldRule) error {
	return NewParserErrorWithContext(r, "VarRule cannot contain a FieldRule")
}

func (r *VarRule) Finalize(parent Container) error {
	if parent == nil {
		return NewParserErrorWithContext(r, "VarRule cannot finalize without a parent container")
	}
	return parent.AddVarRule(r)
}

func (r *VarRule) AddRule(rule any) error {
	return NewParserErrorWithContext(r, "VarRule cannot contain any child rules")
}
