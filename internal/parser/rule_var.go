package parser

import (
	"fmt"

	"github.com/origadmin/adptool/internal/config"
)

// VarRule is a wrapper around config.VarRule to implement the Container interface.
type VarRule struct {
	*config.VarRule
}

func (r *VarRule) AddPackage(pkg *PackageRule) error {
	return fmt.Errorf("VarRule cannot contain a PackageRule")
}

func (r *VarRule) ParseDirective(directive *Directive) error {
	// Delegate to the common RuleSet parser
	return parseRuleSetDirective(&r.RuleSet, directive)
}

func (r *VarRule) AddTypeRule(rule *TypeRule) error {
	return fmt.Errorf("VarRule cannot contain a TypeRule")
}

func (r *VarRule) AddFuncRule(rule *FuncRule) error {
	return fmt.Errorf("VarRule cannot contain a FuncRule")
}

func (r *VarRule) AddVarRule(rule *VarRule) error {
	return fmt.Errorf("VarRule cannot contain a VarRule")
}

func (r *VarRule) AddConstRule(rule *ConstRule) error {
	return fmt.Errorf("VarRule cannot contain a ConstRule")
}

func (r *VarRule) AddMethodRule(rule *MethodRule) error {
	return fmt.Errorf("VarRule cannot contain a MethodRule")
}

func (r *VarRule) AddFieldRule(rule *FieldRule) error {
	return fmt.Errorf("VarRule cannot contain a FieldRule")
}

func (r *VarRule) Finalize(parent Container) error {
	if parent == nil {
		return fmt.Errorf("VarRule cannot finalize without a parent container")
	}
	return parent.AddVarRule(r)
}

func (r *VarRule) AddRule(rule any) error {
	return fmt.Errorf("VarRule cannot contain any child rules")
}
