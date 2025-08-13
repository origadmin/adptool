package parser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// ConfigBuilder is responsible for constructing the config.Config object.
// It is the single stateful object for the parsing process, holding the config being built,
// the active rule, and the current directive for error reporting.
type ConfigBuilder struct {
	config         *config.Config
	activeRule     Rule
	activeMember   *config.MemberRule
	currentPackage *config.Package
	directive      *Directive // The directive currently being processed
}

// NewConfigBuilder creates a new builder.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: config.New(),
	}
}

// GetConfig returns the final, constructed config.
func (b *ConfigBuilder) GetConfig() *config.Config {
	return b.config
}

// SetCurrentDirective sets the directive currently being processed.
func (b *ConfigBuilder) SetCurrentDirective(d *Directive) {
	b.directive = d
}

// SetPackageScope sets the current package for subsequent rule additions.
func (b *ConfigBuilder) SetPackageScope(pkg *config.Package) {
	b.currentPackage = pkg
	b.config.Packages = append(b.config.Packages, pkg)
	b.activeRule = nil
	b.activeMember = nil
}

// EndPackageScope clears the current package scope.
func (b *ConfigBuilder) EndPackageScope() {
	b.currentPackage = nil
	b.activeRule = nil
	b.activeMember = nil
}

// AddTypeRule adds a new type rule to the correct scope (global or package).
func (b *ConfigBuilder) AddTypeRule(name string) {
	configRule := &config.TypeRule{Name: name, Kind: "type"}
	if b.currentPackage != nil {
		b.currentPackage.Types = append(b.currentPackage.Types, configRule)
	} else {
		b.config.Types = append(b.config.Types, configRule)
	}
	b.activeRule = &TypeRule{TypeRule: configRule}
	b.activeMember = nil
}

// AddFuncRule adds a new function rule to the correct scope.
func (b *ConfigBuilder) AddFuncRule(name string) {
	configRule := &config.FuncRule{Name: name}
	if b.currentPackage != nil {
		b.currentPackage.Functions = append(b.currentPackage.Functions, configRule)
	} else {
		b.config.Functions = append(b.config.Functions, configRule)
	}
	b.activeRule = &FuncRule{FuncRule: configRule}
	b.activeMember = nil
}

// AddVarRule adds a new variable rule to the correct scope.
func (b *ConfigBuilder) AddVarRule(name string) {
	configRule := &config.VarRule{Name: name}
	if b.currentPackage != nil {
		b.currentPackage.Variables = append(b.currentPackage.Variables, configRule)
	} else {
		b.config.Variables = append(b.config.Variables, configRule)
	}
	b.activeRule = &VarRule{VarRule: configRule}
	b.activeMember = nil
}

// AddConstRule adds a new constant rule to the correct scope.
func (b *ConfigBuilder) AddConstRule(name string) {
	configRule := &config.ConstRule{Name: name}
	if b.currentPackage != nil {
		b.currentPackage.Constants = append(b.currentPackage.Constants, configRule)
	} else {
		b.config.Constants = append(b.config.Constants, configRule)
	}
	b.activeRule = &ConstRule{ConstRule: configRule}
	b.activeMember = nil
}

// ApplySubDirective applies a sub-directive to the currently active rule.
func (b *ConfigBuilder) ApplySubDirective(subCmds []string, argument string) error {
	if b.activeRule == nil {
		return newDirectiveError(b.directive, "sub-directive ':%s' must follow a top-level rule directive", subCmds[0])
	}
	return b.activeRule.ApplySubDirective(b, subCmds, argument, b.directive)
}

// SetActiveMember sets the active member for sub-directive application.
func (b *ConfigBuilder) SetActiveMember(member *config.MemberRule) {
	b.activeMember = member
}

// ActiveMember returns the currently active member rule.
func (b *ConfigBuilder) ActiveMember() *config.MemberRule {
	return b.activeMember
}

// ApplyRuleToRuleSet is a method on the builder, so it can be called by Rule implementations.
func (b *ConfigBuilder) ApplyRuleToRuleSet(ruleset *config.RuleSet, fromName, ruleName, argument string) error {
	slog.Debug("Applying rule to ruleset", "fromName", fromName, "ruleName", ruleName, "argument", argument)
	if ruleset == nil {
		return fmt.Errorf("cannot apply rule to a nil ruleset")
	}

	switch ruleName {
	case "rename":
		ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: fromName, To: argument})
	case "explicit":
		explicitRules := strings.SplitN(argument, " ", 2)
		if len(explicitRules) == 2 {
			ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: explicitRules[0], To: explicitRules[1]})
		} else {
			return fmt.Errorf("explicit rule argument must be in 'from to' format")
		}
	case "explicit:json":
		var explicitRules []*config.ExplicitRule
		if err := json.Unmarshal([]byte(argument), &explicitRules); err == nil {
			ruleset.Explicit = explicitRules
		} else {
			return fmt.Errorf("invalid JSON for explicit:json: %w", err)
		}
	case "regex:json":
		var regexRules []*config.RegexRule
		if err := json.Unmarshal([]byte(argument), &regexRules); err == nil {
			ruleset.Regex = regexRules
		} else {
			return fmt.Errorf("invalid JSON for regex:json: %w", err)
		}
	case "strategy:json":
		var strategies []string
		if err := json.Unmarshal([]byte(argument), &strategies); err == nil {
			ruleset.Strategy = strategies
		} else {
			return fmt.Errorf("invalid JSON for strategy:json: %w", err)
		}
	case "ignores":
		ruleset.Ignores = append(ruleset.Ignores, argument)
	default:
		return fmt.Errorf("unknown rule name: %s", ruleName)
	}
	return nil
}
