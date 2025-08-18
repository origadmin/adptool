package parser

import (
	"fmt"

	"github.com/origadmin/adptool/internal/config"
)

// TypeRule is a wrapper around config.TypeRule to implement the Container interface.
type TypeRule struct {
	*config.TypeRule
}

func (r *TypeRule) Type() RuleType {
	return RuleTypeType
}

func (r *TypeRule) ParseDirective(directive *Directive) error {
	if directive.BaseCmd != "type" {
		return NewParserErrorWithContext(directive, "TypeRule can only contain type directives")
	}
	if !directive.HasSub() {
		if directive.Argument == "" {
			return NewParserErrorWithContext(directive, "type directive requires an argument (name)")
		}
		r.TypeRule.Name = directive.Argument
		return nil
	}
	subDirective := directive.Sub()
	switch directive.BaseCmd {
	case "struct":
		r.TypeRule.Kind = "struct"
		r.TypeRule.Pattern = directive.Argument
		return nil
	case "rename":
		r.TypeRule.Explicit = append(r.TypeRule.Explicit, &config.ExplicitRule{
			From: r.TypeRule.Name,
			To:   directive.Argument,
		})
		return nil
	case "disabled":
		r.TypeRule.Disabled = directive.Argument == "true"
		return nil
	case "method":
		// todo
		return nil
	case "field":
		//todo
		return nil
	}
	// Delegate to the common RuleSet parser
	return parseRuleSetDirective(&r.RuleSet, subDirective)
}

func (r *TypeRule) AddPackage(pkg *PackageRule) error {
	return fmt.Errorf("TypeRule cannot contain a PackageRule")
}

func (r *TypeRule) AddTypeRule(rule *TypeRule) error {
	return fmt.Errorf("TypeRule cannot contain a TypeRule")
}

func (r *TypeRule) AddFuncRule(rule *FuncRule) error {
	return fmt.Errorf("TypeRule cannot contain a FuncRule")
}

func (r *TypeRule) AddVarRule(rule *VarRule) error {
	return fmt.Errorf("TypeRule cannot contain a VarRule")
}

func (r *TypeRule) AddConstRule(rule *ConstRule) error {
	return NewParserErrorWithContext(r, "TypeRule cannot contain a ConstRule")
}

func (r *TypeRule) AddMethodRule(rule *MethodRule) error {
	r.TypeRule.Methods = append(r.TypeRule.Methods, rule.MemberRule)
	return nil
}

func (r *TypeRule) AddFieldRule(rule *FieldRule) error {
	r.TypeRule.Fields = append(r.TypeRule.Fields, rule.MemberRule)
	return nil
}

func (r *TypeRule) Finalize(parent Container) error {
	if parent == nil {
		return fmt.Errorf("TypeRule cannot finalize without a parent container")
	}
	return parent.AddTypeRule(r)
}

func (r *TypeRule) AddRule(rule any) error {
	switch v := rule.(type) {
	case *MethodRule:
		return r.AddMethodRule(v)
	case *FieldRule:
		return r.AddFieldRule(v)
	default:
		return NewParserErrorWithContext(rule, "TypeRule cannot contain a rule of type %T", rule)
	}
}
