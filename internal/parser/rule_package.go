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
	// DEBUG: Log the BaseCmd right before the check
	// t.Logf("DEBUG: Inside PackageRule.ParseDirective, directive.BaseCmd: %s", directive.BaseCmd)
	if directive.BaseCmd != "package" {
		return NewParserErrorWithContext(directive, "PackageRule can only contain package directives")
	}

	// Handles: //go:adapter:package <import_path> [alias]
	if !directive.HasSub() {
		if directive.Argument == "" {
			return NewParserErrorWithContext(directive, "package directive requires an argument (import path)")
		}
		args := strings.SplitN(directive.Argument, " ", 2)
		if len(args) >= 1 {
			p.Package.Import = args[0]
		}
		if len(args) >= 2 {
			p.Package.Alias = args[1]
		}
		return nil
	}

	// Handle sub-directives, e.g., "go:adapter:package:alias mypkg"
	subDirective := directive.Sub()
	switch subDirective.BaseCmd {
	case "import":
		p.Package.Import = subDirective.Argument
		return nil
	case "alias":
		p.Package.Alias = subDirective.Argument
		return nil
	case "path":
		p.Package.Path = subDirective.Argument
		return nil
	case "property":
		props, err := handlePropDirective(subDirective)
		if err != nil {
			return NewParserErrorWithContext(subDirective, "failed to handle property directive: %w", err)
		}
		p.Package.Props = append(p.Package.Props, props...)
		return nil
	default:
		// This allows structural directives like 'type' or 'function' to be ignored here
		// as they are handled by the main parser's recursion.
		return nil
	}
}

func (p *PackageRule) AddRule(rule any) error {
	switch v := rule.(type) {
	case *TypeRule:
		return p.AddTypeRule(v)
	case *FuncRule:
		return p.AddFuncRule(v)
	case *VarRule:
		return p.AddVarRule(v)
	case *ConstRule:
		return p.AddConstRule(v)
	default:
		return fmt.Errorf("PackageRule cannot contain a rule of type %T", rule)
	}
}

func (p *PackageRule) AddPackage(pkg *PackageRule) error {
	return fmt.Errorf("PackageRule cannot contain another PackageRule")
}

func (p *PackageRule) AddTypeRule(rule *TypeRule) error {
	p.Package.Types = append(p.Package.Types, rule.TypeRule)
	return nil
}

func (p *PackageRule) AddFuncRule(rule *FuncRule) error {
	p.Package.Functions = append(p.Package.Functions, rule.FuncRule)
	return nil
}

func (p *PackageRule) AddVarRule(rule *VarRule) error {
	p.Package.Variables = append(p.Package.Variables, rule.VarRule)
	return nil
}

func (p *PackageRule) AddConstRule(rule *ConstRule) error {
	p.Package.Constants = append(p.Package.Constants, rule.ConstRule)
	return nil
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
