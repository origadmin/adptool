package parser

import (
	"fmt"
	"github.com/origadmin/adptool/internal/config"
)

// MethodRule is a wrapper around config.MemberRule to implement the Container interface.
type MethodRule struct {
	*config.MemberRule
}

func (m *MethodRule) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddRule(rule any) error {
	return fmt.Errorf("MethodRule cannot contain any child rules")
}

func (m *MethodRule) AddPackage(pkg *PackageRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (m *MethodRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}
