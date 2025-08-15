package parser

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/origadmin/adptool/internal/config"
)

// RootConfig is a wrapper around config.Config to implement the Container interface.
type RootConfig struct {
	*config.Config
}

func (r *RootConfig) Type() RuleType {
	return RuleTypeRoot
}

func (r *RootConfig) ParseDirective(directive *Directive) error {
	if r.Config.Defaults == nil {
		r.Config.Defaults = config.NewDefaults()
		r.Config.Props = []*config.PropsEntry{}
	}
	switch directive.BaseCmd {
	case "default":
		// If it's just "//go:adapter:default" with no argument and not JSON
		if directive.Argument == "" {
			return fmt.Errorf("default directive requires an argument (key value)")
		}
		if directive.ShouldUnmarshal() { // Handle JSON block for defaults
			err := json.Unmarshal([]byte(directive.Argument), r.Config.Defaults)
			if err != nil {
				return err
			}
			return nil
		}
		// If there are sub-commands (e.g., "default:strategy")
		if !directive.HasSub() { // Should not happen if len(SubCmds) > 0
			return fmt.Errorf("default directive does not accept a direct argument unless it's a JSON block or has sub-commands")
		}
		return handleDefaultDirective(r.Config.Defaults, directive.Sub())
	case "ignore":
		if directive.Argument == "" {
			return fmt.Errorf("ignore directive requires an argument (pattern)")
		}
		r.Config.Ignores = append(r.Config.Ignores, directive.Argument)
		return nil
	case "ignores":
		if directive.Argument == "" {
			return fmt.Errorf("ignores directive requires an argument (pattern)")
		}
		ignores, err := handleIgnoreDirective(directive)
		if err != nil {
			return NewParserErrorWithContext(directive, "failed to handle ignores directive: %w", err)
		}
		r.Config.Ignores = append(r.Config.Ignores, ignores...)
		return nil
	case "property":
		if directive.Argument == "" {
			return fmt.Errorf("props directive requires an argument (key value)")
		}
		props, err := handlePropDirective(directive)
		if err != nil {
			return NewParserErrorWithContext(directive, "failed to handle property directive: %w", err)
		}
		r.Config.Props = append(r.Config.Props, props...)
		return nil
	// Directives that start new containers (packages, types, funcs, vars, consts)
	// are handled by the parser's main loop (parseFile) via StartContext,
	// not by ParseDirective of the current container.
	case "packages", "types", "functions", "variables", "constants":
		return NewParserErrorWithContext(directive, "directive '%s' starts a new scope and should not be parsed by RootConfig.ParseDirective",
			directive.BaseCmd)
	default:
		return NewParserErrorWithContext(directive, "unrecognized directive '%s' for RootConfig", directive.BaseCmd)
	}
}

func (r *RootConfig) AddRule(rule any) error {
	switch v := rule.(type) {
	case *PackageRule:

		return r.AddPackage(v)
	case *TypeRule:
		return r.AddTypeRule(v)
	case *FuncRule:
		return r.AddFuncRule(v)
	case *VarRule:
		return r.AddVarRule(v)
	case *ConstRule:
		return r.AddConstRule(v)
	default:
		return fmt.Errorf("RootConfig cannot contain a rule of type %T", rule)
	}
}

func (r *RootConfig) AddPackage(pkg *PackageRule) error {
	r.Config.Packages = append(r.Config.Packages, pkg.Package)
	return nil
}

func (r *RootConfig) AddTypeRule(rule *TypeRule) error {
	r.Config.Types = append(r.Config.Types, rule.TypeRule)
	return nil
}

func (r *RootConfig) AddFuncRule(rule *FuncRule) error {
	r.Config.Functions = append(r.Config.Functions, rule.FuncRule)
	return nil
}

func (r *RootConfig) AddVarRule(rule *VarRule) error {
	r.Config.Variables = append(r.Config.Variables, rule.VarRule)
	return nil
}

func (r *RootConfig) AddConstRule(rule *ConstRule) error {
	r.Config.Constants = append(r.Config.Constants, rule.ConstRule)
	return nil
}

func (r *RootConfig) AddMethodRule(rule *MethodRule) error {
	return errors.New("RootConfig cannot contain a MethodRule")
}

func (r *RootConfig) AddFieldRule(rule *FieldRule) error {
	return NewParserErrorWithContext(r, "RootConfig cannot contain a FieldRule")
}

func (r *RootConfig) Finalize(parent Container) error {
	// RootConfig is the top-level container for a file, it has no parent in the parsing hierarchy.
	// Its finalization primarily involves ensuring its internal config.Config is complete.
	// The parent argument will be nil when called from EndContext for the root.
	return nil
}
