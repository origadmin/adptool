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
	return fmt.Errorf("TypeRule cannot contain a PackageRule")
}

func (r *TypeRule) AddTypeRule(rule *TypeRule) error {
	return fmt.Errorf("TypeRule cannot contain a TypeRule")
}

func (r *TypeRule) AddFuncRule(rule *FuncRule) error {
	return fmt.Errorf("TypeRule cannot contain a FuncRule")
}

func (r *TypeRule) AddVarRule(rule *VarRule) error {
	return fmt.Errorf("TypeRule cannot contain a VarRule")
}

func (r *TypeRule) AddConstRule(rule *ConstRule) error {
	return fmt.Errorf("TypeRule cannot contain a ConstRule")
}

func (r *TypeRule) AddMethodRule(rule *MethodRule) error {
	r.TypeRule.Methods = append(r.TypeRule.Methods, rule.MemberRule)
	return nil
}

func (r *TypeRule) AddFieldRule(rule *FieldRule) error {
	r.TypeRule.Fields = append(r.TypeRule.Fields, rule.MemberRule)
	return nil
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
		return r.AddMethodRule(v)
	case *FieldRule:
		return r.AddFieldRule(v)
	default:
		return fmt.Errorf("TypeRule cannot contain a rule of type %T", rule)
	}
}
