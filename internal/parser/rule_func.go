package parser

import (
	"fmt"

	"github.com/origadmin/adptool/internal/config"
)

// FuncRule is a wrapper around config.FuncRule to implement the Container interface.
type FuncRule struct {
	*config.FuncRule
}

func (r *FuncRule) ParseDirective(directive *Directive) error {
	if directive.BaseCmd != "func" {
		return fmt.Errorf("FuncRule can only contain func directives")
	}
	if !directive.HasSub() {
		if directive.Argument == "" {
			return fmt.Errorf("type directive requires an argument (name)")
		}
		r.FuncRule.Name = directive.Argument
		return nil
	}

	// Delegate to the common RuleSet parser
	return parseRuleSetDirective(&r.RuleSet, directive.Sub())
}

func (r *FuncRule) AddPackage(pkg *PackageRule) error {
	return fmt.Errorf("FuncRule cannot contain a PackageRule")
}

func (r *FuncRule) AddTypeRule(rule *TypeRule) error {
	return fmt.Errorf("FuncRule cannot contain a TypeRule")
}

func (r *FuncRule) AddFuncRule(rule *FuncRule) error {
	return fmt.Errorf("FuncRule cannot contain a FuncRule")
}

func (r *FuncRule) AddVarRule(rule *VarRule) error {
	return fmt.Errorf("FuncRule cannot contain a VarRule")
}

func (r *FuncRule) AddConstRule(rule *ConstRule) error {
	return fmt.Errorf("FuncRule cannot contain a ConstRule")
}

func (r *FuncRule) AddMethodRule(rule *MethodRule) error {
	return fmt.Errorf("FuncRule cannot contain a MethodRule")
}

func (r *FuncRule) AddFieldRule(rule *FieldRule) error {
	return fmt.Errorf("FuncRule cannot contain a FieldRule")
}

func (r *FuncRule) Finalize(parent Container) error {
	if parent == nil {
		return fmt.Errorf("FuncRule cannot finalize without a parent container")
	}
	return parent.AddFuncRule(r)
}

func (r *FuncRule) AddRule(rule any) error {
	return fmt.Errorf("FuncRule cannot contain any child rules")
}
