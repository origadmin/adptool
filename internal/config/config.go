package config

// New creates a new, fully initialized Config object.
func New() *Config {
	return &Config{
		Ignores:   make([]string, 0),
		Props:     make([]*PropsEntry, 0),
		Packages:  make([]*Package, 0),
		Types:     make([]*TypeRule, 0),
		Functions: make([]*FuncRule, 0),
		Variables: make([]*VarRule, 0),
		Constants: make([]*ConstRule, 0),
	}
}

// Config is the root of the .adptool.yaml configuration file.
type Config struct {
	Ignores   []string      `yaml:"ignores,omitempty" mapstructure:"ignores,omitempty" json:"ignores,omitempty" toml:"ignores,omitempty"`
	Defaults  *Defaults     `yaml:"defaults,omitempty" mapstructure:"defaults,omitempty" json:"defaults,omitempty" toml:"defaults,omitempty"`
	Props     []*PropsEntry `yaml:"props,omitempty" mapstructure:"props,omitempty" json:"props,omitempty" toml:"props,omitempty"`
	Packages  []*Package    `yaml:"packages,omitempty" mapstructure:"packages,omitempty" json:"packages,omitempty" toml:"packages,omitempty"`
	Types     []*TypeRule   `yaml:"types,omitempty" mapstructure:"types,omitempty" json:"types,omitempty" toml:"types,omitempty"`
	Functions []*FuncRule   `yaml:"functions,omitempty" mapstructure:"functions,omitempty" json:"functions,omitempty" toml:"functions,omitempty"`
	Variables []*VarRule    `yaml:"variables,omitempty" mapstructure:"variables,omitempty" json:"variables,omitempty" toml:"variables,omitempty"`
	Constants []*ConstRule  `yaml:"constants,omitempty" mapstructure:"constants,omitempty" json:"constants,omitempty" toml:"constants,omitempty"`
}

// PropsEntry defines a single variable entry in the config.
type PropsEntry struct {
	Name  string `yaml:"name" mapstructure:"name" json:"name" toml:"name"`
	Value string `yaml:"value" mapstructure:"value" json:"value" toml:"value"`
}

// TypeRule defines the set of rules for a single type declaration.
type TypeRule struct {
	Name     string        `yaml:"name" mapstructure:"name" json:"name" toml:"name"`
	Disabled bool          `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty" json:"disabled,omitempty" toml:"disabled,omitempty"`
	Kind     string        `yaml:"kind,omitempty" mapstructure:"kind,omitempty" json:"kind,omitempty" toml:"kind,omitempty"`
	Pattern  string        `yaml:"pattern,omitempty" mapstructure:"pattern,omitempty" json:"pattern,omitempty" toml:"pattern,omitempty"`
	Methods  []*MemberRule `yaml:"methods,omitempty" mapstructure:"methods,omitempty" json:"methods,omitempty" toml:"methods,omitempty"`
	Fields   []*MemberRule `yaml:"fields,omitempty" mapstructure:"fields,omitempty" json:"fields,omitempty" toml:"fields,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash" json:",inline" toml:",inline"`
}

// FuncRule defines the set of rules for a single function.
type FuncRule struct {
	Name     string `yaml:"name" mapstructure:"name" json:"name" toml:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty" json:"disabled,omitempty" toml:"disabled,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash" json:",inline" toml:",inline"`
}

// VarRule defines the set of rules for a single variable.
type VarRule struct {
	Name     string `yaml:"name" mapstructure:"name" json:"name" toml:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty" json:"disabled,omitempty" toml:"disabled,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash" json:",inline" toml:",inline"`
}

// ConstRule defines the set of rules for a single constant.
type ConstRule struct {
	Name     string `yaml:"name" mapstructure:"name" json:"name" toml:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty" json:"disabled,omitempty" toml:"disabled,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash" json:",inline" toml:",inline"`
}

// MemberRule defines the set of rules for a type member (method or field).
type MemberRule struct {
	Name     string `yaml:"name" mapstructure:"name" json:"name" toml:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty" json:"disabled,omitempty" toml:"disabled,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash" json:",inline" toml:",inline"`
}

// Transform defines the before and after template strings for renaming.
type Transform struct {
	Before string `yaml:"before,omitempty" mapstructure:"before,omitempty" json:"before,omitempty" toml:"before,omitempty"`
	After  string `yaml:"after,omitempty" mapstructure:"after,omitempty" json:"after,omitempty" toml:"after,omitempty"`
}

// RuleSet is the fundamental, reusable building block for defining transformation rules.
type RuleSet struct {
	Strategy        []string        `yaml:"strategy,omitempty" mapstructure:"strategy,omitempty" json:"strategy,omitempty" toml:"strategy,omitempty"`
	Prefix          string          `yaml:"prefix,omitempty" mapstructure:"prefix,omitempty" json:"prefix,omitempty" toml:"prefix,omitempty"`
	PrefixMode      string          `yaml:"prefix_mode,omitempty" mapstructure:"prefix_mode,omitempty" json:"prefix_mode,omitempty" toml:"prefix_mode,omitempty"`
	Suffix          string          `yaml:"suffix,omitempty" mapstructure:"suffix,omitempty" json:"suffix,omitempty" toml:"suffix,omitempty"`
	SuffixMode      string          `yaml:"suffix_mode,omitempty" mapstructure:"suffix_mode,omitempty" json:"suffix_mode,omitempty" toml:"suffix_mode,omitempty"`
	Explicit        []*ExplicitRule `yaml:"explicit,omitempty" mapstructure:"explicit,omitempty" json:"explicit,omitempty" toml:"explicit,omitempty"`
	ExplicitMode    string          `yaml:"explicit_mode,omitempty" mapstructure:"explicit_mode,omitempty" json:"explicit_mode,omitempty" toml:"explicit_mode,omitempty"`
	Regex           []*RegexRule    `yaml:"regex,omitempty" mapstructure:"regex,omitempty" json:"regex,omitempty" toml:"regex,omitempty"`
	RegexMode       string          `yaml:"regex_mode,omitempty" mapstructure:"regex_mode,omitempty" json:"regex_mode,omitempty" toml:"regex_mode,omitempty"`
	Ignores         []string        `yaml:"ignores,omitempty" mapstructure:"ignores,omitempty" json:"ignores,omitempty" toml:"ignores,omitempty"`
	IgnoresMode     string          `yaml:"ignores_mode,omitempty" mapstructure:"ignores_mode,omitempty" json:"ignores_mode,omitempty" toml:"ignores_mode,omitempty"`
	TransformBefore string          `yaml:"transform_before,omitempty" mapstructure:"transform_before,omitempty" json:"transform_before,omitempty" toml:"transform_before,omitempty"`
	TransformAfter  string          `yaml:"transform_after,omitempty" mapstructure:"transform_after,omitempty" json:"transform_after,omitempty" toml:"transform_after,omitempty"`
}

// ExplicitRule defines a direct from/to renaming rule.
type ExplicitRule struct {
	From string `yaml:"from" mapstructure:"from" json:"from" toml:"from"`
	To   string `yaml:"to" mapstructure:"to" json:"to" toml:"to"`
}

// RegexRule defines a regex-based renaming rule.
type RegexRule struct {
	Pattern string `yaml:"pattern" mapstructure:"pattern" json:"pattern" toml:"pattern"`
	Replace string `yaml:"replace" mapstructure:"replace" json:"replace" toml:"replace"`
}

// Package defines rules and variables for a single package.
type Package struct {
	Import    string        `yaml:"import" mapstructure:"import" json:"import" toml:"import"`
	Path      string        `yaml:"path,omitempty" mapstructure:"path,omitempty" json:"path,omitempty" toml:"path,omitempty"`
	Alias     string        `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty" toml:"alias,omitempty"`
	Props     []*PropsEntry `yaml:"props,omitempty" mapstructure:"props,omitempty" json:"props,omitempty" toml:"props,omitempty"`
	Types     []*TypeRule   `yaml:"types,omitempty" mapstructure:"types,omitempty" json:"types,omitempty" toml:"types,omitempty"`
	Functions []*FuncRule   `yaml:"functions,omitempty" mapstructure:"functions,omitempty" json:"functions,omitempty" toml:"functions,omitempty"`
	Variables []*VarRule    `yaml:"variables,omitempty" mapstructure:"variables,omitempty" json:"variables,omitempty" toml:"variables,omitempty"`
	Constants []*ConstRule  `yaml:"constants,omitempty" mapstructure:"constants,omitempty" json:"constants,omitempty" toml:"constants,omitempty"`
}

// Defaults defines the global default behaviors for the entire system.
type Defaults struct {
	Mode *Mode `yaml:"mode,omitempty" mapstructure:"mode,omitempty" json:"mode,omitempty" toml:"mode,omitempty"`
}

// Mode contains key-value pairs where the key is a rule type and the value is the default mode.
type Mode struct {
	Strategy string `yaml:"strategy,omitempty" mapstructure:"strategy,omitempty" json:"strategy,omitempty" toml:"strategy,omitempty"`
	Prefix   string `yaml:"prefix,omitempty" mapstructure:"prefix,omitempty" json:"prefix,omitempty" toml:"prefix,omitempty"`
	Suffix   string `yaml:"suffix,omitempty" mapstructure:"suffix,omitempty" json:"suffix,omitempty" toml:"suffix,omitempty"`
	Explicit string `yaml:"explicit,omitempty" mapstructure:"explicit,omitempty" json:"explicit,omitempty" toml:"explicit,omitempty"`
	Regex    string `yaml:"regex,omitempty" mapstructure:"regex,omitempty" json:"regex,omitempty" toml:"regex,omitempty"`
	Ignores  string `yaml:"ignores,omitempty" mapstructure:"ignores,omitempty" json:"ignores,omitempty" toml:"ignores,omitempty"`
}

// Merge combines two Config objects.
func Merge(base *Config, overlay *Config) (*Config, error) {
	// This is a placeholder for the actual merge logic.
	// A real implementation would be much more sophisticated.
	return base, nil
}
