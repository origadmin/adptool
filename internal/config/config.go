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

	if err := v.ReadInConfig(); err != nil {
		if fileConfigPath != "" || !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		log.Println("Project-level .adptool.yaml not found, proceeding with empty config.") // Updated log message
	} else {
		log.Printf("Config loaded successfully. Settings: %+v", v.AllSettings())
	}

	// Unmarshal the user-facing configuration.
	cfg := &Config{}
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
	effective := pkgCfg // Start with package-specific config

	// Handle Prefix
	if effective.Prefix == "" { // If package-level prefix is not set, inherit from global
		effective.Prefix = globalCfg.Prefix
	}

	// Handle Suffix
	if effective.Suffix == "" { // If package-level suffix is not set, inherit from global
		effective.Suffix = globalCfg.Suffix
	}

	// Handle Explicit map
	if effective.InheritExplicit {
		mergedExplicit := make(map[string]string)
		for k, v := range globalCfg.Explicit {
			mergedExplicit[k] = v
		}
		for k, v := range pkgCfg.Explicit {
			mergedExplicit[k] = v
		}
		effective.Explicit = mergedExplicit
	} else {
		if pkgCfg.Explicit == nil {
			effective.Explicit = globalCfg.Explicit
		}
	}

	// Handle Regex rules
	if effective.InheritRegex {
		mergedRegex := make([]RegexRule, len(globalCfg.Regex))
		copy(mergedRegex, globalCfg.Regex)
		mergedRegex = append(mergedRegex, pkgCfg.Regex...)
		effective.Regex = mergedRegex
	} else {
		if pkgCfg.Regex == nil {
			effective.Regex = globalCfg.Regex
		}
	}

	// Handle Ignore rules
	if effective.InheritIgnore {
		mergedIgnore := make([]string, len(globalCfg.Ignore))
		copy(mergedIgnore, globalCfg.Ignore)
		mergedIgnore = append(mergedIgnore, pkgCfg.Ignore...)
		effective.Ignore = mergedIgnore
	} else {
		if pkgCfg.Ignore == nil {
			effective.Ignore = globalCfg.Ignore
		}
	}

	return effective
}

// compileRules processes the entire config tree and compiles the user-facing
// rules into the internal CompiledCategoryRules for the generator to use.
func compileRules(cfg *Config) {
	// Compile global rules for each category
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
