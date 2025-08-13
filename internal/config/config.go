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

// Config is the root of the .adptool.onfiguration file.
type Config struct {
	Ignores   []string      `mapstructure:"ignores,omitempty"`
	Defaults  *Defaults     `mapstructure:"defaults,omitempty"`
	Props     []*PropsEntry `mapstructure:"props,omitempty"`
	Packages  []*Package    `mapstructure:"packages,omitempty"`
	Types     []*TypeRule   `mapstructure:"types,omitempty"`
	Functions []*FuncRule   `mapstructure:"functions,omitempty"`
	Variables []*VarRule    `mapstructure:"variables,omitempty"`
	Constants []*ConstRule  `mapstructure:"constants,omitempty"`
}

// PropsEntry defines a single variable entry in the config.
type PropsEntry struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
}

// TypeRule defines the set of rules for a single type declaration.
type TypeRule struct {
	Name     string        `mapstructure:"name"`
	Disabled bool          `mapstructure:"disabled,omitempty"`
	Kind     string        `mapstructure:"kind,omitempty"`
	Pattern  string        `mapstructure:"pattern,omitempty"`
	Methods  []*MemberRule `mapstructure:"methods,omitempty"`
	Fields   []*MemberRule `mapstructure:"fields,omitempty"`
	RuleSet  `mapstructure:",squash"`
}

// FuncRule defines the set of rules for a single function.
type FuncRule struct {
	Name     string `mapstructure:"name"`
	Disabled bool   `mapstructure:"disabled,omitempty"`
	RuleSet  `mapstructure:",squash"`
}

// VarRule defines the set of rules for a single variable.
type VarRule struct {
	Name     string `mapstructure:"name"`
	Disabled bool   `mapstructure:"disabled,omitempty"`
	RuleSet  `mapstructure:",squash"`
}

// ConstRule defines the set of rules for a single constant.
type ConstRule struct {
	Name     string `mapstructure:"name"`
	Disabled bool   `mapstructure:"disabled,omitempty"`
	RuleSet  `mapstructure:",squash"`
}

// MemberRule defines the set of rules for a type member (method or field).
type MemberRule struct {
	Name     string `mapstructure:"name"`
	Disabled bool   `mapstructure:"disabled,omitempty"`
	RuleSet  `mapstructure:",squash"`
}

// RuleSet is the fundamental, reusable building block for defining transformation rules.
type RuleSet struct {
	Strategy        []string        `mapstructure:"strategy,omitempty"`
	Prefix          string          `mapstructure:"prefix,omitempty"`
	PrefixMode      string          `mapstructure:"prefix_mode,omitempty"`
	Suffix          string          `mapstructure:"suffix,omitempty"`
	SuffixMode      string          `mapstructure:"suffix_mode,omitempty"`
	Explicit        []*ExplicitRule `mapstructure:"explicit,omitempty"`
	ExplicitMode    string          `mapstructure:"explicit_mode,omitempty"`
	Regex           []*RegexRule    `mapstructure:"regex,omitempty"`
	RegexMode       string          `mapstructure:"regex_mode,omitempty"`
	Ignores         []string        `mapstructure:"ignores,omitempty"`
	IgnoresMode     string          `mapstructure:"ignores_mode,omitempty"`
	TransformBefore string          `mapstructure:"transform_before,omitempty"`
	TransformAfter  string          `mapstructure:"transform_after,omitempty"`
}

// ExplicitRule defines a direct from/to renaming rule.
type ExplicitRule struct {
	From string `mapstructure:"from"`
	To   string `mapstructure:"to"`
}

// RegexRule defines a regex-based renaming rule.
type RegexRule struct {
	Pattern string `mapstructure:"pattern"`
	Replace string `mapstructure:"replace"`
}

// Package defines rules and variables for a single package.
type Package struct {
	Import    string        `mapstructure:"import"`
	Path      string        `mapstructure:"path,omitempty"`
	Alias     string        `mapstructure:"alias,omitempty"`
	Props     []*PropsEntry `mapstructure:"props,omitempty"`
	Types     []*TypeRule   `mapstructure:"types,omitempty"`
	Functions []*FuncRule   `mapstructure:"functions,omitempty"`
	Variables []*VarRule    `mapstructure:"variables,omitempty"`
	Constants []*ConstRule  `mapstructure:"constants,omitempty"`
}

// Defaults defines the global default behaviors for the entire system.
type Defaults struct {
	Mode *Mode `mapstructure:"mode,omitempty"`
}

// Mode contains key-value pairs where the key is a rule type and the value is the default mode.
type Mode struct {
	Strategy string `mapstructure:"strategy,omitempty"`
	Prefix   string `mapstructure:"prefix,omitempty"`
	Suffix   string `mapstructure:"suffix,omitempty"`
	Explicit string `mapstructure:"explicit,omitempty"`
	Regex    string `mapstructure:"regex,omitempty"`
	Ignores  string `mapstructure:"ignores,omitempty"`
}

// Merge combines two Config objects.
func Merge(base *Config, overlay *Config) (*Config, error) {
	// This is a placeholder for the actual merge logic.
	// A real implementation would be much more sophisticated.
	return base, nil
}
