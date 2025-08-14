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
	//TODO implement me
	panic("implement me")
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

func (r *TypeRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}

func (p *TypeRule) AddRule(rule any) error {
	switch r := rule.(type) {
	case *MethodRule:
		return p.AddMethodRule(r)
	case *FieldRule:
		return p.AddFieldRule(r)
	default:
		return fmt.Errorf("TypeRule cannot contain a rule of type %T", rule)
	}
}
