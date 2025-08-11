package config

// New creates a new, fully initialized Config object.
func New() *Config {
	return &Config{
		Defaults: &Defaults{Mode: &Mode{}},
		Vars:     make(map[string]string),
		Packages: make([]*Package, 0),
	}
}

// newRuleSet creates a new RuleSet with initialized slices.
func newRuleSet() *RuleSet {
	return &RuleSet{
		Strategy: make([]string, 0),
		Explicit: make([]*ExplicitRule, 0),
		Regex:    make([]*RegexRule, 0),
		Ignore:   make([]string, 0),
	}
}

// newTypeRuleSet creates a new TypeRuleSet with initialized nested RuleSets.
func newTypeRule() *TypeRule {
	return &TypeRule{
		Methods: newMemberRule(),
		Fields:  newMemberRule(),
		RuleSet: newRuleSet(),
	}
}

func newMemberRule() []*MemberRule {
	return []*MemberRule{
		{
			RuleSet: newRuleSet(),
		},
	}
}

// Config is the root of the .adptool.yaml configuration file.
type Config struct {
	Defaults  *Defaults         `yaml:"defaults,omitempty" mapstructure:"defaults,omitempty"`
	Vars      map[string]string `yaml:"vars,omitempty" mapstructure:"vars,omitempty"`
	Packages  []*Package        `yaml:"packages,omitempty" mapstructure:"packages,omitempty"`
	Types     []*TypeRule       `yaml:"types,omitempty" mapstructure:"types,omitempty"`
	Functions []*FuncRule       `yaml:"functions,omitempty" mapstructure:"functions,omitempty"`
	Variables []*VarRule        `yaml:"variables,omitempty" mapstructure:"variables,omitempty"`
	Constants []*ConstRule      `yaml:"constants,omitempty" mapstructure:"constants,omitempty"`
}

// --- Rule Structures ---

// TypeRule defines the set of rules for a single type declaration.
type TypeRule struct {
	Name     string        `yaml:"name" mapstructure:"name"`
	Disabled bool          `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	Kind     string        `yaml:"kind,omitempty" mapstructure:"kind,omitempty"`
	Pattern  string        `yaml:"pattern,omitempty" mapstructure:"pattern,omitempty"`
	Methods  []*MemberRule `yaml:"methods,omitempty" mapstructure:"methods,omitempty"`
	Fields   []*MemberRule `yaml:"fields,omitempty" mapstructure:"fields,omitempty"`
	*RuleSet `yaml:",inline" mapstructure:",squash"`
}

// FuncRule defines the set of rules for a single function.
type FuncRule struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	*RuleSet `yaml:",inline" mapstructure:",squash"`
}

// VarRule defines the set of rules for a single variable.
type VarRule struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	*RuleSet `yaml:",inline" mapstructure:",squash"`
}

// ConstRule defines the set of rules for a single constant.
type ConstRule struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	*RuleSet `yaml:",inline" mapstructure:",squash"`
}

// MemberRule defines the set of rules for a type member (method or field).
type MemberRule struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	*RuleSet `yaml:",inline" mapstructure:",squash"`
}

// RuleSet is the fundamental, reusable building block for defining transformation rules.
type RuleSet struct {
	Strategy        []string        `yaml:"strategy,omitempty" mapstructure:"strategy,omitempty"`
	Prefix          string          `yaml:"prefix,omitempty" mapstructure:"prefix,omitempty"`
	PrefixMode      string          `yaml:"prefix_mode,omitempty" mapstructure:"prefix_mode,omitempty"`
	Suffix          string          `yaml:"suffix,omitempty" mapstructure:"suffix,omitempty"`
	SuffixMode      string          `yaml:"suffix_mode,omitempty" mapstructure:"suffix_mode,omitempty"`
	Explicit        []*ExplicitRule `yaml:"explicit,omitempty" mapstructure:"explicit,omitempty"`
	ExplicitMode    string          `yaml:"explicit_mode,omitempty" mapstructure:"explicit_mode,omitempty"`
	Regex           []*RegexRule    `yaml:"regex,omitempty" mapstructure:"regex,omitempty"`
	RegexMode       string          `yaml:"regex_mode,omitempty" mapstructure:"regex_mode,omitempty"`
	Ignore          []string        `yaml:"ignore,omitempty" mapstructure:"ignore,omitempty"`
	IgnoreMode      string          `yaml:"ignore_mode,omitempty" mapstructure:"ignore_mode,omitempty"`
	TransformBefore string          `yaml:"transform_before,omitempty" mapstructure:"transform_before,omitempty"`
	TransformAfter  string          `yaml:"transform_after,omitempty" mapstructure:"transform_after,omitempty"`
}

// ExplicitRule defines a direct from/to renaming rule.
type ExplicitRule struct {
	From string `yaml:"from" mapstructure:"from"`
	To   string `yaml:"to" mapstructure:"to"`
}

// RegexRule defines a regex-based renaming rule.
type RegexRule struct {
	Pattern string `yaml:"pattern" mapstructure:"pattern"`
	Replace string `yaml:"replace" mapstructure:"replace"`
}

//// TypeRuleSet extends a RuleSet with nested rules for type members.
//type TypeRuleSet struct {
//	*RuleSet `yaml:",inline"`
//	Methods  *RuleSet `yaml:"methods,omitempty"`
//	Fields   *RuleSet `yaml:"fields,omitempty"`
//}

// --- Other Top-Level Structures ---

// Package defines rules and variables for a single package.
type Package struct {
	Import    string            `yaml:"import" mapstructure:"import"`
	Path      string            `yaml:"path,omitempty" mapstructure:"path,omitempty"`
	Alias     string            `yaml:"alias,omitempty" mapstructure:"alias,omitempty"`
	Vars      map[string]string `yaml:"vars,omitempty" mapstructure:"vars,omitempty"`
	Types     []*TypeRule       `yaml:"types,omitempty" mapstructure:"types,omitempty"`
	Functions []*FuncRule       `yaml:"functions,omitempty" mapstructure:"functions,omitempty"`
	Variables []*VarRule        `yaml:"variables,omitempty" mapstructure:"variables,omitempty"`
	Constants []*ConstRule      `yaml:"constants,omitempty" mapstructure:"constants,omitempty"`
}

// Defaults defines the global default behaviors for the entire system.
type Defaults struct {
	Mode *Mode `yaml:"mode,omitempty" mapstructure:"mode,omitempty"`
}

// Mode contains key-value pairs where the key is a rule type and the value is the default mode.
// This is for low-level rule-engine behavior, not to be confused with adaptation patterns.
type Mode struct {
	Strategy string `yaml:"strategy,omitempty" mapstructure:"strategy,omitempty"`
	Prefix   string `yaml:"prefix,omitempty" mapstructure:"prefix,omitempty"`
	Suffix   string `yaml:"suffix,omitempty" mapstructure:"suffix,omitempty"`
	Explicit string `yaml:"explicit,omitempty" mapstructure:"explicit,omitempty"`
	Regex    string `yaml:"regex,omitempty" mapstructure:"regex,omitempty"`
	Ignore   string `yaml:"ignore,omitempty" mapstructure:"ignore,omitempty"`
}
