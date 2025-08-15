package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// RuleType is an enum for different container rule types.
type RuleType int

// Enum for RuleType
const (
	RuleTypeUnknown RuleType = iota
	RuleTypeRoot
	RuleTypePackage
	RuleTypeType
	RuleTypeFunc
	RuleTypeVar
	RuleTypeConst
	RuleTypeMethod
	RuleTypeField
	// Add other rule types as needed
)

// init registers all the top-level container rules with the factory.
func init() {
	RegisterContainer(RuleTypeRoot, func() Container { return &RootConfig{Config: config.New()} })
	RegisterContainer(RuleTypePackage, func() Container { return &PackageRule{Package: &config.Package{}} })
	RegisterContainer(RuleTypeType, func() Container { return &TypeRule{TypeRule: &config.TypeRule{}} })
	RegisterContainer(RuleTypeFunc, func() Container { return &FuncRule{FuncRule: &config.FuncRule{}} })
	RegisterContainer(RuleTypeVar, func() Container { return &VarRule{VarRule: &config.VarRule{}} })
	RegisterContainer(RuleTypeConst, func() Container { return &ConstRule{ConstRule: &config.ConstRule{}} })
	RegisterContainer(RuleTypeMethod, func() Container { return &MethodRule{MemberRule: &config.MemberRule{}} })
	RegisterContainer(RuleTypeField, func() Container { return &FieldRule{MemberRule: &config.MemberRule{}} })
}

// NewContainerFactory resolves a command string (including abbreviations) and returns the
// corresponding RuleType constant.
func NewContainerFactory(ruleType RuleType) ContainerFactory {
	return func() Container {
		return NewContainer(ruleType)
	}
}

// parseRuleSetDirective handles directives that apply to a config.RuleSet.
func parseRuleSetDirective(rs *config.RuleSet, directive *Directive) error {
	if directive.ShouldUnmarshal() { // Handle JSON block for defaults
		err := json.Unmarshal([]byte(directive.Argument), rs)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON for RuleSet: %w", NewParserError("failed to unmarshal JSON for RuleSet"))
		}
		return nil
	}
	switch directive.BaseCmd {
	case "strategy":
		if directive.Argument == "" {
			return newDirectiveError(directive, "strategy directive requires an argument")
		}
		rs.Strategy = append(rs.Strategy, directive.Argument)
		return nil
	case "prefix":
		rs.Prefix = directive.Argument
		return nil
	case "prefix_mode":
		rs.PrefixMode = directive.Argument
		return nil
	case "suffix":
		rs.Suffix = directive.Argument
		return nil
	case "suffix_mode":
		rs.SuffixMode = directive.Argument
		return nil
	case "explicit":
		// Explicit rules are key=value pairs, need to parse directive.Argument
		if directive.Argument == "" {
			return newDirectiveError(directive, "explicit directive requires an argument (from=to)")
		}
		parts := strings.SplitN(directive.Argument, "=", 2)
		if len(parts) != 2 {
			return newDirectiveError(directive, "invalid explicit directive argument '%s', expected from=to", directive.Argument)
		}
		rs.Explicit = append(rs.Explicit, &config.ExplicitRule{
			From: parts[0],
			To:   parts[1],
		})
		return nil
	case "explicit_mode":
		rs.ExplicitMode = directive.Argument
		return nil
	case "regex":
		// Regex rules are pattern=replace pairs
		if directive.Argument == "" {
			return newDirectiveError(directive, "regex directive requires an argument (pattern=replace)")
		}
		parts := strings.SplitN(directive.Argument, "=", 2)
		if len(parts) != 2 {
			return newDirectiveError(directive, "invalid regex directive argument '%s', expected pattern=replace", directive.Argument)
		}
		rs.Regex = append(rs.Regex, &config.RegexRule{
			Pattern: parts[0],
			Replace: parts[1],
		})
		return nil
	case "regex_mode":
		rs.RegexMode = directive.Argument
		return nil
	case "ignore":
		if directive.Argument == "" {
			return newDirectiveError(directive, "ignore directive requires an argument (pattern)")
		}
		rs.Ignores = append(rs.Ignores, directive.Argument)
		return nil
	case "ignores":
		if directive.Argument == "" {
			return newDirectiveError(directive, "ignores directive requires an argument (pattern)")
		}
		ignores, err := handleIgnoreDirective(directive)
		if err != nil {
			return fmt.Errorf("failed to handle ignore directive: %w", newDirectiveError(directive, "failed to handle ignore directive"))
		}
		rs.Ignores = append(rs.Ignores, ignores...)
		return nil
	case "ignores_mode":
		rs.IgnoresMode = directive.Argument
		return nil
	case "transform":
		if !directive.HasSub() {
			return newDirectiveError(directive, "transform directive requires a sub-command")
		}
		sub := directive.Sub()
			if sub.ShouldUnmarshal() {
		err := json.Unmarshal([]byte(directive.Argument), rs.Transforms)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON for RuleSet.Transforms: %w", newDirectiveError(directive, "failed to unmarshal JSON for RuleSet.Transforms"))
		}
		return nil
	}
		switch sub.BaseCmd {
		case "before":
			rs.Transforms.Before = sub.Argument
		case "after":
			rs.Transforms.After = sub.Argument
		default:
			return newDirectiveError(sub, "unrecognized directive '%s' for RuleSet.Transforms", sub.BaseCmd)
		}
		return nil
	case "transform_before":
		rs.TransformBefore = directive.Argument
		rs.Transforms.Before = directive.Argument
		return nil
	case "transform_after":
		rs.TransformAfter = directive.Argument
		rs.Transforms.After = directive.Argument
		return nil
	default:
		return newDirectiveError(directive, "unrecognized directive '%s' for RuleSet", directive.BaseCmd)
	}
}
