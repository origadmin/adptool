package parser

import (
	"github.com/origadmin/adptool/internal/config"
)

// FieldRule is a wrapper around config.MemberRule to implement the Container interface.
type FieldRule struct {
	*config.MemberRule
}

func (f *FieldRule) Type() RuleType {
	return RuleTypeField
}

func (f *FieldRule) ParseDirective(directive *Directive) error {
	if directive.BaseCmd != "field" {
		return NewParserErrorWithContext(directive, "FieldRule can only contain field directives")
	}

	if !directive.HasSub() {
		if directive.Argument == "" {
			return NewParserErrorWithContext(directive, "field directive requires an argument (name)")
		}
		f.MemberRule.Name = directive.Argument
		return nil
	}

	subDirective := directive.Sub()
	switch subDirective.BaseCmd {
	// Add field-specific cases here in the future (e.g., "type", "tag")
	}

	// Delegate to the common RuleSet parser for generic rules
	return parseRuleSetDirective(&f.RuleSet, subDirective)
}

func (f *FieldRule) AddRule(rule any) error {
	return NewParserError("FieldRule cannot contain any child rules")
}

func (f *FieldRule) AddPackage(pkg *PackageRule) error {
	return NewParserError("FieldRule cannot contain a PackageRule")
}

func (f *FieldRule) AddTypeRule(rule *TypeRule) error {
	return NewParserError("FieldRule cannot contain a TypeRule")
}

func (f *FieldRule) AddFuncRule(rule *FuncRule) error {
	return NewParserError("FieldRule cannot contain a FuncRule")
}

func (f *FieldRule) AddVarRule(rule *VarRule) error {
	return NewParserError("FieldRule cannot contain a VarRule")
}

func (f *FieldRule) AddConstRule(rule *ConstRule) error {
	return NewParserError("FieldRule cannot contain a ConstRule")
}

func (f *FieldRule) AddMethodRule(rule *MethodRule) error {
	return NewParserError("FieldRule cannot contain a MethodRule")
}

func (f *FieldRule) AddFieldRule(rule *FieldRule) error {
	return NewParserError("FieldRule cannot contain a FieldRule")
}

func (f *FieldRule) Finalize(parent Container) error {
	if parent == nil {
		return NewParserErrorWithContext(f, "FieldRule cannot finalize without a parent container")
	}
	return parent.AddFieldRule(f)
}
