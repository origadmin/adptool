package config

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/spf13/viper"
)

// CategoryConfig defines renaming rules for a specific category (e.g., types, functions, methods).
type CategoryConfig struct {
	Prefix   string            `mapstructure:"prefix,omitempty"`
	Suffix   string            `mapstructure:"suffix,omitempty"`
	Explicit map[string]string `mapstructure:"explicit,omitempty"`
	Regex    []RegexRule       `mapstructure:"regex,omitempty"`
	Ignore   []string          `mapstructure:"ignore,omitempty"`

	// Inheritance flags for global rules
	InheritPrefix   bool `mapstructure:"inherit_prefix,omitempty"`
	InheritSuffix   bool `mapstructure:"inherit_suffix,omitempty"`
	InheritExplicit bool `mapstructure:"inherit_explicit,omitempty"`
	InheritRegex    bool `mapstructure:"inherit_regex,omitempty"`
	InheritIgnore   bool `mapstructure:"inherit_ignore,omitempty"`
}

// CompiledCategoryRules holds the compiled and effective rules for a specific category.
type CompiledCategoryRules struct {
	Rules  []RenameRule
	Ignore []string
}

// Config represents the user-facing configuration.
// It is compiled into a set of internal RenameRule objects.
type Config struct {
	Prefix   string            `mapstructure:"prefix,omitempty"`
	Suffix   string            `mapstructure:"suffix,omitempty"`
	Explicit map[string]string `mapstructure:"explicit,omitempty"`
	Regex    []RegexRule       `mapstructure:"regex,omitempty"`
	Ignore   []string          `mapstructure:"ignore,omitempty"`

	// Inheritance flags for global rules
	InheritPrefix   bool `mapstructure:"inherit_prefix,omitempty"`
	InheritSuffix   bool `mapstructure:"inherit_suffix,omitempty"`
	InheritExplicit bool `mapstructure:"inherit_explicit,omitempty"`
	InheritRegex    bool `mapstructure:"inherit_regex,omitempty"`
	InheritIgnore   bool `mapstructure:"inherit_ignore,omitempty"`

	// User-facing fields
	Types     CategoryConfig `mapstructure:"types,omitempty"`
	Functions CategoryConfig `mapstructure:"functions,omitempty"`
	Methods   CategoryConfig `mapstructure:"methods,omitempty"`

	// Internal-facing, compiled rules
	CompiledTypes     CompiledCategoryRules `mapstructure:"-"`
	CompiledFunctions CompiledCategoryRules `mapstructure:"-"`
	CompiledMethods   CompiledCategoryRules `mapstructure:"-"`

	// Package-specific configs
	Packages []PackageConfig `mapstructure:"packages,omitempty"`
}

// PackageConfig represents user-facing config for a specific package.
type PackageConfig struct {
	Import string `mapstructure:"import"`
	Path   string `mapstructure:"path,omitempty"`
	Alias  string `mapstructure:"alias,omitempty"`

	Prefix   string            `mapstructure:"prefix,omitempty"`
	Suffix   string            `mapstructure:"suffix,omitempty"`
	Explicit map[string]string `mapstructure:"explicit,omitempty"`
	Regex    []RegexRule       `mapstructure:"regex,omitempty"`
	Ignore   []string          `mapstructure:"ignore,omitempty"`

	// Inheritance flags for global rules
	InheritPrefix   bool `mapstructure:"inherit_prefix,omitempty"`
	InheritSuffix   bool `mapstructure:"inherit_suffix,omitempty"`
	InheritExplicit bool `mapstructure:"inherit_explicit,omitempty"`
	InheritRegex    bool `mapstructure:"inherit_regex,omitempty"`
	InheritIgnore   bool `mapstructure:"inherit_ignore,omitempty"`

	// User-facing fields
	Types     CategoryConfig `mapstructure:"types,omitempty"`
	Functions CategoryConfig `mapstructure:"functions,omitempty"`
	Methods   CategoryConfig `mapstructure:"methods,omitempty"`

	// Internal-facing, compiled rules
	CompiledTypes     CompiledCategoryRules `mapstructure:"-"`
	CompiledFunctions CompiledCategoryRules `mapstructure:"-"`
	CompiledMethods   CompiledCategoryRules `mapstructure:"-"`
}

// RegexRule is a user-facing struct for regex renaming.
type RegexRule struct {
	Pattern string `mapstructure:"pattern"`
	Replace string `mapstructure:"replace"`
}

// RenameRule is the internal, unified representation of any renaming rule.
// This is not intended to be used directly by end-users in config files.
type RenameRule struct {
	Type    string // "explicit", "prefix", "suffix", "regex"
	From    string // For "explicit"
	To      string // For "explicit"
	Value   string // For "prefix", "suffix"
	Pattern string // For "regex"
	Replace string // For "regex"
}

func DefaultConfig() *Config {
	return &Config{
		Explicit:        make(map[string]string),
		Regex:           []RegexRule{},
		Ignore:          []string{},
		InheritPrefix:   false,
		InheritSuffix:   false,
		InheritExplicit: false,
		InheritRegex:    false,
		InheritIgnore:   false,
		Types: CategoryConfig{
			Explicit: make(map[string]string),
			Regex:    []RegexRule{},
			Ignore:   []string{},
		},
		Functions: CategoryConfig{
			Explicit: make(map[string]string),
			Regex:    []RegexRule{},
			Ignore:   []string{},
		},
		Methods: CategoryConfig{
			Explicit: make(map[string]string),
			Regex:    []RegexRule{},
			Ignore:   []string{},
		},
		CompiledTypes: CompiledCategoryRules{
			Rules:  nil,
			Ignore: nil,
		},
		CompiledFunctions: CompiledCategoryRules{
			Rules:  nil,
			Ignore: nil,
		},
		CompiledMethods: CompiledCategoryRules{
			Rules:  nil,
			Ignore: nil,
		},
		Packages: nil,
	}
}

// LoadConfig loads and compiles the user-facing configuration into a set of
// internal, prioritized RenameRule objects.
func LoadConfig(fileConfigPath string) (*Config, error) {
	v := viper.New()
	if fileConfigPath != "" {
		v.SetConfigFile(fileConfigPath)
	} else {
		v.SetConfigName(".adptool") // Changed from "adptool" to ".adptool"
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	// Unmarshal the user-facing configuration.
	// Initialize cfg with default values
	cfg := DefaultConfig()

	if err := v.ReadInConfig(); err != nil {
		if fileConfigPath != "" || !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		log.Println("Project-level .adptool.yaml not found, proceeding with default config.") // Updated log message
	} else {
		log.Printf("Config loaded successfully. Settings: %+v", v.AllSettings())
	}

	// Unmarshal the user-facing configuration.
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Compile all user-facing rules into internal CompiledCategoryRules.
	compileRules(cfg)

	log.Printf("Final compiled config: %+v", cfg)
	return cfg, nil
}

// mergeCategoryConfig merges a package-specific CategoryConfig with a global one,
// respecting inheritance flags.
func mergeCategoryConfig(globalCfg, pkgCfg CategoryConfig) CategoryConfig {
	effective := CategoryConfig{ // Initialize with zero values
		Explicit: make(map[string]string),
		Regex:    []RegexRule{},
		Ignore:   []string{},
	}

	// Handle Prefix
	if pkgCfg.Prefix != "" { // If package-level prefix is explicitly set
		effective.Prefix = pkgCfg.Prefix
	} else if pkgCfg.InheritPrefix { // If package-level prefix is empty and inheritance is true
		effective.Prefix = globalCfg.Prefix
	} else { // If package-level prefix is empty and inheritance is false (explicit clear)
		effective.Prefix = ""
	}

	// Handle Suffix
	if pkgCfg.Suffix != "" { // If package-level suffix is explicitly set
		effective.Suffix = pkgCfg.Suffix
	} else if pkgCfg.InheritSuffix { // If package-level suffix is empty and inheritance is true
		effective.Suffix = globalCfg.Suffix
	} else { // If package-level suffix is empty and inheritance is false (explicit clear)
		effective.Suffix = ""
	}

	// Handle Explicit map
	if pkgCfg.Explicit != nil && len(pkgCfg.Explicit) > 0 { // If package-level explicit is explicitly set and not empty
		if pkgCfg.InheritExplicit { // If inheriting, merge
			for k, v := range globalCfg.Explicit {
				effective.Explicit[k] = v
			}
			for k, v := range pkgCfg.Explicit {
				effective.Explicit[k] = v
			}
		} else { // If not inheriting, use package-level explicit only
			effective.Explicit = pkgCfg.Explicit
		}
	} else if pkgCfg.InheritExplicit { // If package-level explicit is empty/nil and inheritance is true
		effective.Explicit = globalCfg.Explicit
	} else { // If package-level explicit is empty/nil and inheritance is false (explicit clear)
		effective.Explicit = make(map[string]string) // Ensure it's an empty map
	}

	// Handle Regex rules
	if pkgCfg.Regex != nil && len(pkgCfg.Regex) > 0 { // If package-level regex is explicitly set and not empty
		if pkgCfg.InheritRegex { // If inheriting, append
			effective.Regex = append(effective.Regex, globalCfg.Regex...)
			effective.Regex = append(effective.Regex, pkgCfg.Regex...)
		} else { // If not inheriting, use package-level regex only
			effective.Regex = pkgCfg.Regex
		}
	} else if pkgCfg.InheritRegex { // If package-level regex is empty/nil and inheritance is true
		effective.Regex = globalCfg.Regex
	} else { // If package-level regex is empty/nil and inheritance is false (explicit clear)
		effective.Regex = []RegexRule{} // Ensure it's an empty slice
	}

	// Handle Ignore rules
	if pkgCfg.Ignore != nil && len(pkgCfg.Ignore) > 0 { // If package-level ignore is explicitly set and not empty
		if pkgCfg.InheritIgnore { // If inheriting, append
			effective.Ignore = append(effective.Ignore, globalCfg.Ignore...)
			effective.Ignore = append(effective.Ignore, pkgCfg.Ignore...)
		} else { // If not inheriting, use package-level ignore only
			effective.Ignore = pkgCfg.Ignore
		}
	} else if pkgCfg.InheritIgnore { // If package-level ignore is empty/nil and inheritance is true
		effective.Ignore = globalCfg.Ignore
	} else { // If package-level ignore is empty/nil and inheritance is false (explicit clear)
		effective.Ignore = []string{} // Ensure it's an empty slice
	}

	// Copy inheritance flags from pkgCfg to effective
	effective.InheritPrefix = pkgCfg.InheritPrefix
	effective.InheritSuffix = pkgCfg.InheritSuffix
	effective.InheritExplicit = pkgCfg.InheritExplicit
	effective.InheritRegex = pkgCfg.InheritRegex
	effective.InheritIgnore = pkgCfg.InheritIgnore

	return effective
}

// compileRules processes the entire config tree and compiles the user-facing
// rules into the internal CompiledCategoryRules for the generator to use.
func compileRules(cfg *Config) {
	// Create a temporary CategoryConfig for the top-level global rules
	globalCategoryConfig := CategoryConfig{
		Prefix:   cfg.Prefix,
		Suffix:   cfg.Suffix,
		Explicit: cfg.Explicit,
		Regex:    cfg.Regex,
		Ignore:   cfg.Ignore,
		// Top-level global rules don't inherit from anything above them,
		// so their Inherit flags are effectively always false for their own compilation.
		// However, they act as 'globalCfg' for the next level (Types, Functions, Methods).
	}

	// Compile global rules for each category, merging with top-level global rules
	// The 'globalCategoryConfig' acts as the 'globalCfg' for these category-level compilations.
	cfg.Types = mergeCategoryConfig(globalCategoryConfig, cfg.Types)
	cfg.Functions = mergeCategoryConfig(globalCategoryConfig, cfg.Functions)
	cfg.Methods = mergeCategoryConfig(globalCategoryConfig, cfg.Methods)

	cfg.CompiledTypes = CompiledCategoryRules{
		Rules:  compileRuleSet(cfg.Types.Prefix, cfg.Types.Suffix, cfg.Types.Explicit, cfg.Types.Regex),
		Ignore: cfg.Types.Ignore,
	}
	cfg.CompiledFunctions = CompiledCategoryRules{
		Rules:  compileRuleSet(cfg.Functions.Prefix, cfg.Functions.Suffix, cfg.Functions.Explicit, cfg.Functions.Regex),
		Ignore: cfg.Functions.Ignore,
	}
	cfg.CompiledMethods = CompiledCategoryRules{
		Rules:  compileRuleSet(cfg.Methods.Prefix, cfg.Methods.Suffix, cfg.Methods.Explicit, cfg.Methods.Regex),
		Ignore: cfg.Methods.Ignore,
	}

	// Process each package
	for i := range cfg.Packages {
		pkgCfg := &cfg.Packages[i] // Get a pointer to modify the original struct in the slice

		// Merge package-specific category configs with global ones to get effective configs
		// The 'cfg.Types', 'cfg.Functions', 'cfg.Methods' now hold the effective global category rules
		effectiveTypes := mergeCategoryConfig(cfg.Types, pkgCfg.Types)
		effectiveFunctions := mergeCategoryConfig(cfg.Functions, pkgCfg.Functions)
		effectiveMethods := mergeCategoryConfig(cfg.Methods, pkgCfg.Methods)

		// Compile rules for this package based on effective configs
		pkgCfg.CompiledTypes = CompiledCategoryRules{
			Rules:  compileRuleSet(effectiveTypes.Prefix, effectiveTypes.Suffix, effectiveTypes.Explicit, effectiveTypes.Regex),
			Ignore: effectiveTypes.Ignore,
		}
		pkgCfg.CompiledFunctions = CompiledCategoryRules{
			Rules:  compileRuleSet(effectiveFunctions.Prefix, effectiveFunctions.Suffix, effectiveFunctions.Explicit, effectiveFunctions.Regex),
			Ignore: effectiveFunctions.Ignore,
		}
		pkgCfg.CompiledMethods = CompiledCategoryRules{
			Rules:  compileRuleSet(effectiveMethods.Prefix, effectiveMethods.Suffix, effectiveMethods.Explicit, effectiveMethods.Regex),
			Ignore: effectiveMethods.Ignore,
		}
	}
}

// compileRuleSet is the core compilation logic.
// It takes user-facing rules and creates a prioritized, internal list of RenameRule objects.
func compileRuleSet(prefix, suffix string, explicit map[string]string, regexRules []RegexRule) []RenameRule {
	var compiledRules []RenameRule

	// Priority 1: Explicit
	if len(explicit) > 0 {
		keys := make([]string, 0, len(explicit))
		for k := range explicit {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Ensure deterministic order
		for _, from := range keys {
			compiledRules = append(compiledRules, RenameRule{Type: "explicit", From: from, To: explicit[from]})
		}
	}

	// Priority 2: Prefix
	if prefix != "" {
		compiledRules = append(compiledRules, RenameRule{Type: "prefix", Value: prefix})
	}

	// Priority 3: Suffix
	if suffix != "" {
		compiledRules = append(compiledRules, RenameRule{Type: "suffix", Value: suffix})
	}

	// Priority 4: Regex
	for _, r := range regexRules {
		compiledRules = append(compiledRules, RenameRule{Type: "regex", Pattern: r.Pattern, Replace: r.Replace})
	}

	return compiledRules
}
