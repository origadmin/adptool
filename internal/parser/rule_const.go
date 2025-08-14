package parser

import (
	"fmt"

	"github.com/origadmin/adptool/internal/config"
)

// ConstRule is a wrapper around config.ConstRule to implement the Container interface.
type ConstRule struct {
	*config.ConstRule
}

func (r *ConstRule) AddPackage(pkg *PackageRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}

func (r *ConstRule) AddRule(rule any) error {
	return fmt.Errorf("ConstRule cannot contain any child rules")
}
