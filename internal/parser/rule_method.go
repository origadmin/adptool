package parser

import (
	"github.com/origadmin/adptool/internal/config"
)

// MethodRule is a wrapper around config.MemberRule to implement the Container interface.
type MethodRule struct {
	*config.MemberRule
}

func (m *MethodRule) Type() RuleType {
	return RuleTypeMethod
}

func (m *MethodRule) ParseDirective(directive *Directive) error {
	if directive.BaseCmd != "method" {
		return NewParserErrorWithContext(directive, "MethodRule can only contain method directives")
	}
	// Delegate to the common RuleSet parser
	return parseRuleSetDirective(&m.RuleSet, directive)
}

func (m *MethodRule) AddRule(rule any) error {
	return NewParserError("MethodRule cannot contain any child rules")
}

func (m *MethodRule) AddPackage(pkg *PackageRule) error {
	return NewParserError("MethodRule cannot contain a PackageRule")
}

func (m *MethodRule) AddTypeRule(rule *TypeRule) error {
	return NewParserError("MethodRule cannot contain a TypeRule")
}

func (m *MethodRule) AddFuncRule(rule *FuncRule) error {
	return NewParserError("MethodRule cannot contain a FuncRule")
}

func (m *MethodRule) AddVarRule(rule *VarRule) error {
	return NewParserError("MethodRule cannot contain a VarRule")
}

func (m *MethodRule) AddConstRule(rule *ConstRule) error {
	return NewParserError("MethodRule cannot contain a ConstRule")
}

func (m *MethodRule) AddMethodRule(rule *MethodRule) error {
	return NewParserError("MethodRule cannot contain a MethodRule")
}

func (m *MethodRule) AddFieldRule(rule *FieldRule) error {
	return NewParserError("MethodRule cannot contain a FieldRule")
}

func (m *MethodRule) Finalize(parent Container) error {
	if parent == nil {
		return NewParserErrorWithContext(m, "MethodRule cannot finalize without a parent container")
	}
	return parent.AddMethodRule(m)
}
