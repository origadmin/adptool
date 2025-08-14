package parser

import (
	"fmt"

	"github.com/origadmin/adptool/internal/config"
)

// TypeRule is a wrapper around config.TypeRule to implement the Container interface.
type TypeRule struct {
	*config.TypeRule
}

func (r *TypeRule) ParseDirective(directive *Directive) error {
	// Delegate to the common RuleSet parser
	return parseRuleSetDirective(&r.RuleSet, directive)
}

func (r *TypeRule) AddPackage(pkg *PackageRule) error {
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

func (r *TypeRule) Finalize(parent Container) error {
	if parent == nil {
		return fmt.Errorf("TypeRule cannot finalize without a parent container")
	}
	return parent.AddTypeRule(r)
}

func (r *TypeRule) AddRule(rule any) error {
	switch v := rule.(type) {
	case *MethodRule:
		r.TypeRule.Methods = append(r.TypeRule.Methods, v.MemberRule)
		return nil
	case *FieldRule:
		r.TypeRule.Fields = append(r.TypeRule.Fields, v.MemberRule)
		return nil
	default:
		return fmt.Errorf("TypeRule cannot contain a rule of type %T", rule)
	}
}
