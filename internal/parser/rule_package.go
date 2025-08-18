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

func (p *PackageRule) Type() RuleType {
	return RuleTypePackage
}

func (p *PackageRule) ParseDirective(directive *Directive) error {
	if directive.BaseCmd != "package" {
		return NewParserErrorWithContext(directive, "type directive requires a base command")
	}
	if directive.HasSub() {
		subDirective := directive.Sub()
		switch subDirective.BaseCmd {
		case "import":
			if subDirective.Argument == "" {
				return fmt.Errorf("import subDirective requires an argument (path)")
			}
			p.Package.Import = subDirective.Argument
			return nil
		case "path":
			p.Package.Path = subDirective.Argument
			return nil
		case "alias":
			p.Package.Alias = subDirective.Argument
			return nil
		case "property":
			if subDirective.Argument == "" {
				return fmt.Errorf("props directive requires an argument (key value)")
			}
			props, err := handlePropDirective(subDirective)
			if err != nil {
				return NewParserErrorWithContext(subDirective, "failed to handle property directive: %w", err)
			}
			p.Package.Props = append(p.Package.Props, props...)
			return nil
		case "types", "functions", "variables", "constants":
			return fmt.Errorf("directive '%s' starts a new scope and should not be parsed by PackageRule.ParseDirective", directive.Command)
		default:
			// Handle other potential directives that might be part of RuleSet if embedded directly
			// For now, return an error for unknown directives.
			return NewParserErrorWithContext(subDirective, "unrecognized directive '%s' for PackageRule", subDirective.Command)
		}
	} else {
		if directive.Argument == "" {
			return NewParserErrorWithContext(directive, "type directive requires an argument (name)")
		}
		args := strings.SplitN(directive.Argument, " ", 2)
		if len(args) == 1 {
			p.Package.Import = args[0]
		} else if len(args) > 1 {
			p.Package.Import = args[0]
			p.Package.Alias = args[1]
		} else {
		}

		return nil
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
	return fmt.Errorf("PackageRule cannot contain another PackageRule")
}

func (p *PackageRule) AddTypeRule(rule *TypeRule) error {
	return p.AddRule(rule)
}

func (p *PackageRule) AddFuncRule(rule *FuncRule) error {
	return p.AddRule(rule)
}

func (p *PackageRule) AddVarRule(rule *VarRule) error {
	return p.AddRule(rule)
}

func (p *PackageRule) AddConstRule(rule *ConstRule) error {
	return p.AddRule(rule)
}

func (p *PackageRule) AddMethodRule(rule *MethodRule) error {
	return fmt.Errorf("PackageRule cannot contain a MethodRule")
}

func (p *PackageRule) AddFieldRule(rule *FieldRule) error {
	return fmt.Errorf("PackageRule cannot contain a FieldRule")
}

func (p *PackageRule) Finalize(parent Container) error {
	if parent == nil {
		return NewParserErrorWithContext(p, "PackageRule cannot finalize without a parent container")
	}
	return parent.AddPackage(p)
}
