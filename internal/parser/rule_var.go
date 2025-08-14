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
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}

func (r *VarRule) AddRule(rule any) error {
	return fmt.Errorf("VarRule cannot contain any child rules")
}
