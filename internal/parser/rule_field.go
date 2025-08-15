package parser

import (
	"github.com/origadmin/adptool/internal/config"
)

// FieldRule is a wrapper around config.MemberRule to implement the Container interface.
type FieldRule struct {
	*config.MemberRule
}

func (f *FieldRule) ParseDirective(directive *Directive) error {
	// Delegate to the common RuleSet parser
	return parseRuleSetDirective(&f.RuleSet, directive)
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
		return NewParserError("FieldRule cannot finalize without a parent container")
	}
	return parent.AddFieldRule(f)
}
