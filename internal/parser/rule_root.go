package parser

import (
	"fmt"
	"github.com/origadmin/adptool/internal/config"
)

// RootConfig is a wrapper around config.Config to implement the Container interface.
type RootConfig struct {
	*config.Config
}

func (r *RootConfig) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddRule(rule any) error {
	switch r := rule.(type) {
	case *TypeRule:
		return r.AddTypeRule(r)
	case *FuncRule:
		return r.AddFuncRule(r)
	case *VarRule:
		return r.AddVarRule(r)
	case *ConstRule:
		return r.AddConstRule(r)
	default:
		return fmt.Errorf("RootConfig cannot contain a rule of type %T", rule)
	}
}

func (r *RootConfig) AddPackage(pkg *PackageRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (r *RootConfig) Finalize() error {
	//TODO implement me
	panic("implement me")
}
