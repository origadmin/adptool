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
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddPackage(pkg *PackageRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}

func (r *FuncRule) AddRule(rule any) error {
	return fmt.Errorf("FuncRule cannot contain any child rules")
}
