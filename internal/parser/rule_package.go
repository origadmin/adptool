package parser

import (
	"fmt"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// PackageRule is a wrapper around config.Package to implement the Container interface.
// (Previously PackageConfig)
type PackageRule struct {
	*config.Package
}

func (p *PackageRule) ParseDirective(directive *Directive) error {
	switch directive.Command {
	case "import":
		if directive.Argument == "" {
			return fmt.Errorf("import directive requires an argument (path)")
		}
		p.Package.Import = directive.Argument
		return nil
	case "path":
		p.Package.Path = directive.Argument
		return nil
	case "alias":
		p.Package.Alias = directive.Argument
		return nil
	case "props":
		if directive.Argument == "" {
			return fmt.Errorf("props directive requires an argument (key=value)")
		}
		parts := strings.SplitN(directive.Argument, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid props directive argument '%s', expected key=value", directive.Argument)
		}
		p.Package.Props = append(p.Package.Props, &config.PropsEntry{
			Name:  parts[0],
			Value: parts[1],
		})
		return nil
	// Directives that start new containers (types, funcs, vars, consts)
	// are handled by the parser's main loop (parseFile) via StartContext,
	// not by ParseDirective of the current container.
	case "types", "functions", "variables", "constants":
		return fmt.Errorf("directive '%s' starts a new scope and should not be parsed by PackageRule.ParseDirective", directive.Command)
	default:
		// Handle other potential directives that might be part of RuleSet if embedded directly
		// For now, return an error for unknown directives.
		return fmt.Errorf("unrecognized directive '%s' for PackageRule", directive.Command)
	}
}

func (p *PackageRule) AddRule(rule any) error {
	switch v := rule.(type) {
	case *TypeRule:
		p.Package.Types = append(p.Package.Types, v.TypeRule)
		return nil
	case *FuncRule:
		p.Package.Functions = append(p.Package.Functions, v.FuncRule)
		return nil
	case *VarRule:
		p.Package.Variables = append(p.Package.Variables, v.VarRule)
		return nil
	case *ConstRule:
		p.Package.Constants = append(p.Package.Constants, v.ConstRule)
		return nil
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

func (p *PackageRule) AddMethodRule(rule *MethodRule) error {
	//TODO implement me
	panic("implement me")
}

func (p *PackageRule) AddFieldRule(rule *FieldRule) error {
	//TODO implement me
	panic("implement me")
}

func (p *PackageRule) Finalize(parent Container) error {
	if parent == nil {
		return fmt.Errorf("PackageRule cannot finalize without a parent container")
	}
	return parent.AddPackage(p)
}
