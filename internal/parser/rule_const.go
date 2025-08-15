package parser

import (
	"github.com/origadmin/adptool/internal/config"
)

// ConstRule is a wrapper around config.ConstRule to implement the Container interface.
type ConstRule struct {
	*config.ConstRule
}

func (r *ConstRule) AddPackage(pkg *PackageRule) error {
	return NewParserError("ConstRule cannot contain a PackageRule")
}

func (r *ConstRule) ParseDirective(directive *Directive) error {
	if directive.BaseCmd != "const" {
		return newDirectiveError(directive, "ConstRule can only contain const directives")
	}
	if !directive.HasSub() {
		if directive.Argument == "" {
			return newDirectiveError(directive, "type directive requires an argument (name)")
		}
		r.ConstRule.Name = directive.Argument
		return nil
	}

	// Delegate to the common RuleSet parser
	return parseRuleSetDirective(&r.RuleSet, directive.Sub())
}

func (r *ConstRule) AddTypeRule(rule *TypeRule) error {
	return NewParserError("ConstRule cannot contain a TypeRule")
}

func (r *ConstRule) AddFuncRule(rule *FuncRule) error {
	return NewParserError("ConstRule cannot contain a FuncRule")
}

func (r *ConstRule) AddVarRule(rule *VarRule) error {
	return NewParserError("ConstRule cannot contain a VarRule")
}

func (r *ConstRule) AddConstRule(rule *ConstRule) error {
	return NewParserError("ConstRule cannot contain a ConstRule")
}

func (r *ConstRule) AddMethodRule(rule *MethodRule) error {
	return NewParserError("ConstRule cannot contain a MethodRule")
}

func (r *ConstRule) AddFieldRule(rule *FieldRule) error {
	return NewParserError("ConstRule cannot contain a FieldRule")
}

func (r *ConstRule) Finalize(parent Container) error {
	if parent == nil {
		return NewParserError("ConstRule cannot finalize without a parent container")
	}
	return parent.AddConstRule(r)
}

func (r *ConstRule) AddRule(rule any) error {
	return NewParserError("ConstRule cannot contain any child rules")
}
