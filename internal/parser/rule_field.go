package parser

import (
	"fmt"
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
	return fmt.Errorf("FieldRule cannot contain any child rules")
}

func (f *FieldRule) AddPackage(pkg *PackageRule) error {
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

func (f *FieldRule) Finalize(parent Container) error {
	if parent == nil {
		return fmt.Errorf("FieldRule cannot finalize without a parent container")
	}
	return parent.AddFieldRule(f)
}
