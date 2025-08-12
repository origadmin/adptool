package config

// New creates a new, fully initialized Config object.
func New() *Config {
	return &Config{
		Ignores: make([]string, 0),
		Props:   make([]*VarEntry, 0),
	}
}

// newRuleSet creates a new RuleSet with initialized slices.
func newRuleSet() RuleSet {
	return RuleSet{
		Strategy: make([]string, 0),
		Explicit: make([]*ExplicitRule, 0),
		Regex:    make([]*RegexRule, 0),
		Ignores:  make([]string, 0),
	}
}

// newTypeRuleSet creates a new TypeRuleSet with initialized nested RuleSets.
func newTypeRule() TypeRule {
	return TypeRule{
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
	Ignores   []string     `yaml:"ignores,omitempty" mapstructure:"ignores,omitempty"` // Global ignore patterns
	Defaults  *Defaults    `yaml:"defaults,omitempty" mapstructure:"defaults,omitempty"`
	Props     []*VarEntry  `yaml:"props,omitempty" mapstructure:"props,omitempty"` // Configuration properties
	Packages  []*Package   `yaml:"packages,omitempty" mapstructure:"packages,omitempty"`
	Types     []*TypeRule  `yaml:"types,omitempty" mapstructure:"types,omitempty"`
	Functions []*FuncRule  `yaml:"functions,omitempty" mapstructure:"functions,omitempty"`
	Variables []*VarRule   `yaml:"variables,omitempty" mapstructure:"variables,omitempty"`
	Constants []*ConstRule `yaml:"constants,omitempty" mapstructure:"constants,omitempty"`
}

// VarEntry defines a single variable entry in the config.
type VarEntry struct {
	Name  string `yaml:"name" mapstructure:"name"`
	Value string `yaml:"value" mapstructure:"value"`
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
	RuleSet  `yaml:",inline" mapstructure:",squash"` // Changed from *RuleSet to RuleSet
}

// FuncRule defines the set of rules for a single function.
type FuncRule struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash"` // Changed from *RuleSet to RuleSet
}

// VarRule defines the set of rules for a single variable.
type VarRule struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash"` // Changed from *RuleSet to RuleSet
}

// ConstRule defines the set of rules for a single constant.
type ConstRule struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash"` // Changed from *RuleSet to RuleSet
}

// MemberRule defines the set of rules for a type member (method or field).
type MemberRule struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Disabled bool   `yaml:"disabled,omitempty" mapstructure:"disabled,omitempty"`
	RuleSet  `yaml:",inline" mapstructure:",squash"` // Changed from *RuleSet to RuleSet
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
	Ignores         []string        `yaml:"ignores,omitempty" mapstructure:"ignores,omitempty"`
	IgnoresMode     string          `yaml:"ignores_mode,omitempty" mapstructure:"ignores_mode,omitempty"`
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
	Import    string       `yaml:"import" mapstructure:"import"`
	Path      string       `yaml:"path,omitempty" mapstructure:"path,omitempty"`
	Alias     string       `yaml:"alias,omitempty" mapstructure:"alias,omitempty"`
	Vars      []*VarEntry  `yaml:"vars,omitempty" mapstructure:"vars,omitempty"` // Changed to array of VarEntry
	Types     []*TypeRule  `yaml:"types,omitempty" mapstructure:"types,omitempty"`
	Functions []*FuncRule  `yaml:"functions,omitempty" mapstructure:"functions,omitempty"`
	Variables []*VarRule   `yaml:"variables,omitempty" mapstructure:"variables,omitempty"`
	Constants []*ConstRule `yaml:"constants,omitempty" mapstructure:"constants,omitempty"`
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
	Ignores  string `yaml:"ignores,omitempty" mapstructure:"ignores,omitempty"`
}

// Merge combines two Config objects based on adptool's specific precedence rules.
// The overlay Config's rules will take precedence over the base Config's rules.
// This function is crucial for applying file-specific directives over project-wide defaults.
func Merge(base *Config, overlay *Config) (*Config, error) {
	if base == nil && overlay == nil {
		return nil, nil
	}
	if base == nil {
		return overlay, nil
	}
	if overlay == nil {
		return base, nil
	}

	// Create a new Config to hold the merged result
	merged := &Config{}
	*merged = *base // Start with a shallow copy of the base

	// Merge simple fields (maps, slices, pointers)
	// Props: overlay takes precedence
	if overlay.Props != nil {
		if merged.Props == nil {
			merged.Props = make([]*VarEntry, 0)
		}
		for _, v := range overlay.Props {
			merged.Props = append(merged.Props, v)
		}
	}

	// Ignores: overlay appends to base
	if overlay.Ignores != nil {
		merged.Ignores = append(merged.Ignores, overlay.Ignores...)
	}

	// Defaults: overlay takes precedence for individual fields
	// This is a simplified merge, a full merge would iterate over Mode fields
	if overlay.Defaults != nil && overlay.Defaults.Mode != nil {
		if merged.Defaults == nil {
			merged.Defaults = &Defaults{}
		}
		if merged.Defaults.Mode == nil {
			merged.Defaults.Mode = &Mode{}
		}
		// Shallow copy Mode for now, a deeper merge would be field by field
		*merged.Defaults.Mode = *overlay.Defaults.Mode
	}

	// Packages: overlay appends to base, or replaces if overlay has a matching import
	// This is a simplified merge, a full merge would need to match by Import path
	if overlay.Packages != nil {
		merged.Packages = append(merged.Packages, overlay.Packages...)
	}

	// Merge rule lists (Types, Functions, Variables, Constants)
	// This is the core of the custom merge logic.
	// Rules are merged based on their 'Name' field.
	// Overlay rules with matching names replace base rules.
	// Overlay rules with new names are appended.

	merged.Types = mergeTypeRules(base.Types, overlay.Types)
	merged.Functions = mergeFuncRules(base.Functions, overlay.Functions)
	merged.Variables = mergeVarRules(base.Variables, overlay.Variables)
	merged.Constants = mergeConstRules(base.Constants, overlay.Constants)

	return merged, nil
}

// mergeTypeRules merges two lists of TypeRule objects.
func mergeTypeRules(base, overlay []*TypeRule) []*TypeRule {
	return mergeRules(base, overlay, func(r *TypeRule) string { return r.Name }) // Use Name for matching
}

// mergeFuncRules merges two lists of FuncRule objects.
func mergeFuncRules(base, overlay []*FuncRule) []*FuncRule {
	return mergeRules(base, overlay, func(r *FuncRule) string { return r.Name }) // Use Name for matching
}

// mergeVarRules merges two lists of VarRule objects.
func mergeVarRules(base, overlay []*VarRule) []*VarRule {
	return mergeRules(base, overlay, func(r *VarRule) string { return r.Name }) // Use Name for matching
}

// mergeConstRules merges two lists of ConstRule objects.
func mergeConstRules(base, overlay []*ConstRule) []*ConstRule {
	return mergeRules(base, overlay, func(r *ConstRule) string { return r.Name }) // Use Name for matching
}

// mergeRules is a generic helper to merge lists of rules based on a key extractor.
// It assumes the rule objects embed a RuleSet and have a Name field.
func mergeRules[T interface{ TypeRule | FuncRule | VarRule | ConstRule }](base, overlay []*T, keyExtractor func(*T) string) []*T {
	if len(overlay) == 0 {
		return base
	}
	if len(base) == 0 {
		return overlay
	}

	// Create a map for faster lookups of base rules
	baseMap := make(map[string]*T)
	for _, r := range base {
		baseMap[keyExtractor(r)] = r
	}

	mergedList := make([]*T, 0, len(base)+len(overlay))

	// Add base rules, merging with overlay if a match exists
	for _, rBase := range base {
		k := keyExtractor(rBase)
		if rOverlay, found := baseMap[k]; found {
			// Match found, merge rOverlay into rBase
			// This is where the actual field-by-field merge of RuleSet happens
			// For now, we'll just take the overlay's RuleSet if it exists, otherwise base's
			// A more sophisticated merge would merge individual fields of RuleSet
			// This part needs to be refined to merge individual fields of RuleSet
			// For now, if a rule exists in overlay, it completely replaces the base rule.
			// This is the simplest merge strategy.
			mergedList = append(mergedList, rOverlay)
			delete(baseMap, k) // Mark as processed
		} else {
			mergedList = append(mergedList, rBase)
		}
	}

	// Add remaining overlay rules (those not in base)
	for _, rOverlay := range overlay {
		if _, found := baseMap[keyExtractor(rOverlay)]; !found {
			mergedList = append(mergedList, rOverlay)
		}
	}

	return mergedList
}
