package config

// Config is the root of the .adptool.yaml configuration file.
type Config struct {
	Defaults  *Defaults         `yaml:"defaults,omitempty"`
	Vars      map[string]string `yaml:"vars,omitempty"`
	Types     *TypeRuleSet      `yaml:"types,omitempty"`
	Functions *RuleSet          `yaml:"functions,omitempty"`
	Variables *RuleSet          `yaml:"variables,omitempty"`
	Constants *RuleSet          `yaml:"constants,omitempty"`
	Packages  []*Package        `yaml:"packages,omitempty"`
}

// Defaults defines the global default behaviors for the entire system.
type Defaults struct {
	Mode *Mode `yaml:"mode,omitempty"`
}

// Mode contains key-value pairs where the key is a rule type and the value is the default mode.
type Mode struct {
	Strategy string `yaml:"strategy,omitempty"`
	Prefix   string `yaml:"prefix,omitempty"`
	Suffix   string `yaml:"suffix,omitempty"`
	Explicit string `yaml:"explicit,omitempty"`
	Regex    string `yaml:"regex,omitempty"`
	Ignore   string `yaml:"ignore,omitempty"`
}

// RuleSet is the fundamental building block for defining rules.
type RuleSet struct {
	Strategy        []string        `yaml:"strategy,omitempty"`
	Prefix          string          `yaml:"prefix,omitempty"`
	PrefixMode      string          `yaml:"prefix_mode,omitempty"`
	Suffix          string          `yaml:"suffix,omitempty"`
	SuffixMode      string          `yaml:"suffix_mode,omitempty"`
	Explicit        []*ExplicitRule `yaml:"explicit,omitempty"`
	ExplicitMode    string          `yaml:"explicit_mode,omitempty"`
	Regex           []*RegexRule    `yaml:"regex,omitempty"`
	RegexMode       string          `yaml:"regex_mode,omitempty"`
	Ignore          []string        `yaml:"ignore,omitempty"`
	IgnoreMode      string          `yaml:"ignore_mode,omitempty"`
	TransformBefore string          `yaml:"transform_before,omitempty"`
	TransformAfter  string          `yaml:"transform_after,omitempty"`
}

// TypeRuleSet extends a RuleSet with nested rules for type members.
type TypeRuleSet struct {
	*RuleSet `yaml:",inline"`
	Methods  *RuleSet `yaml:"methods,omitempty"`
	Fields   *RuleSet `yaml:"fields,omitempty"`
}

// Package defines rules and variables for a single package.
type Package struct {
	Import    string            `yaml:"import"`
	Path      string            `yaml:"path,omitempty"`
	Alias     string            `yaml:"alias,omitempty"`
	Vars      map[string]string `yaml:"vars,omitempty"`
	Types     *TypeRuleSet      `yaml:"types,omitempty"`
	Functions *RuleSet          `yaml:"functions,omitempty"`
	Variables *RuleSet          `yaml:"variables,omitempty"`
	Constants *RuleSet          `yaml:"constants,omitempty"`
}

// ExplicitRule defines a direct from/to renaming rule.
type ExplicitRule struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// RegexRule defines a regex-based renaming rule.
type RegexRule struct {
	Pattern string `yaml:"pattern"`
	Replace string `yaml:"replace"`
}
