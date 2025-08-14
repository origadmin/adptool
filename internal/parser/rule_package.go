package parser

import (
	"fmt"
	"github.com/origadmin/adptool/internal/config"
)

// PackageRule is a wrapper around config.Package to implement the Container interface.
// (Previously PackageConfig)
type PackageRule struct {
	*config.Package
}

func (p *PackageRule) ParseDirective(directive *Directive) error {
	//TODO implement me
	panic("implement me")
}

func (p *PackageRule) AddRule(rule any) error {
	switch r := rule.(type) {
	case *TypeRule:
		return p.AddTypeRule(r)
	case *FuncRule:
		return p.AddFuncRule(r)
	case *VarRule:
		return p.AddVarRule(r)
	case *ConstRule:
		return p.AddConstRule(r)
	default:
		return fmt.Errorf("PackageRule cannot contain a rule of type %T", rule)
	}
}

func (p *PackageRule) AddPackage(pkg *PackageRule) error {
	//TODO implement me
	panic("implement me")
}

func (p *PackageRule) AddTypeRule(rule *TypeRule) error {
	//TODO implement me
	panic("implement me")
}

func (p *PackageRule) AddFuncRule(rule *FuncRule) error {
	//TODO implement me
	panic("implement me")
}

func (p *PackageRule) AddVarRule(rule *VarRule) error {
	//TODO implement me
	panic("implement me")
}

func (p *PackageRule) AddConstRule(rule *ConstRule) error {
	//TODO implement me
	panic("implement me")
}

func (p *PackageRule) Finalize() error {
	//TODO implement me
	panic("implement me")
}
