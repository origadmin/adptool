package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// parseRuleSetDirective handles directives that apply to a config.RuleSet.
func parseRuleSetDirective(rs *config.RuleSet, directive *Directive) error {
	switch directive.Command {
	case "strategy":
		if directive.Argument == "" {
			return fmt.Errorf("strategy directive requires an argument")
		}
		rs.Strategy = append(rs.Strategy, directive.Argument)
	case "prefix":
		rs.Prefix = directive.Argument
	case "prefix_mode":
		rs.PrefixMode = directive.Argument
	case "suffix":
		rs.Suffix = directive.Argument
	case "suffix_mode":
		rs.SuffixMode = directive.Argument
	case "explicit":
		// Explicit rules are key=value pairs, need to parse directive.Argument
		if directive.Argument == "" {
			return fmt.Errorf("explicit directive requires an argument (from=to)")
		}
		parts := strings.SplitN(directive.Argument, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid explicit directive argument '%s', expected from=to", directive.Argument)
		}
		rs.Explicit = append(rs.Explicit, &config.ExplicitRule{
			From: parts[0],
			To:   parts[1],
		})
	case "explicit_mode":
		rs.ExplicitMode = directive.Argument
	case "regex":
		// Regex rules are pattern=replace pairs
		if directive.Argument == "" {
			return fmt.Errorf("regex directive requires an argument (pattern=replace)")
		}
		parts := strings.SplitN(directive.Argument, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid regex directive argument '%s', expected pattern=replace", directive.Argument)
		}
		rs.Regex = append(rs.Regex, &config.RegexRule{
			Pattern: parts[0],
			Replace: parts[1],
		})
	case "regex_mode":
		rs.RegexMode = directive.Argument
	case "ignores":
		if directive.Argument == "" {
			return fmt.Errorf("ignores directive requires an argument (pattern)")
		}
		rs.Ignores = append(rs.Ignores, directive.Argument)
	case "ignores_mode":
		rs.IgnoresMode = directive.Argument
	case "transform_before":
		rs.TransformBefore = directive.Argument
	case "transform_after":
		rs.TransformAfter = directive.Argument
	default:
		return fmt.Errorf("unrecognized directive '%s' for RuleSet", directive.Command)
	}
	return nil
}

// RootConfig is a wrapper around config.Config to implement the Container interface.
type RootConfig struct {
	*config.Config
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
			return err
		}
		r.Config.Ignores = append(r.Config.Ignores, ignores...)
		return nil
	case "property":
		if directive.Argument == "" {
			return fmt.Errorf("props directive requires an argument (key value)")
		}
		props, err := handlePropDirective(directive)
		if err != nil {
			return err
		}
		r.Config.Props = append(r.Config.Props, props...)
		return nil
	// Directives that start new containers (packages, types, funcs, vars, consts)
	// are handled by the parser's main loop (parseFile) via StartContext,
	// not by ParseDirective of the current container.
	case "packages", "types", "functions", "variables", "constants":
		return fmt.Errorf("directive '%s' starts a new scope and should not be parsed by RootConfig.ParseDirective", directive.Command)
	default:
		return fmt.Errorf("unrecognized directive '%s' for RootConfig", directive.Command)
	}
}

func (r *RootConfig) AddRule(rule any) error {
	switch v := rule.(type) {
	case *PackageRule:
		r.Config.Packages = append(r.Config.Packages, v.Package)
		return nil
	case *TypeRule:
		r.Config.Types = append(r.Config.Types, v.TypeRule)
		return nil
	case *FuncRule:
		r.Config.Functions = append(r.Config.Functions, v.FuncRule)
		return nil
	case *VarRule:
		r.Config.Variables = append(r.Config.Variables, v.VarRule)
		return nil
	case *ConstRule:
		r.Config.Constants = append(r.Config.Constants, v.ConstRule)
		return nil
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
	return errors.New("RootConfig cannot contain a FieldRule")
}

func (r *RootConfig) Finalize(parent Container) error {
	// RootConfig is the top-level container for a file, it has no parent in the parsing hierarchy.
	// Its finalization primarily involves ensuring its internal config.Config is complete.
	// The parent argument will be nil when called from EndContext for the root.
	return nil
}
