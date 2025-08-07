package config

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/spf13/viper"
)

// Config represents the user-facing configuration.
// It is compiled into a set of internal RenameRule objects.
type Config struct {
	// User-facing fields
	Prefix   string            `mapstructure:"prefix,omitempty"`
	Suffix   string            `mapstructure:"suffix,omitempty"`
	Explicit map[string]string `mapstructure:"explicit,omitempty"`
	Regex    []RegexRule       `mapstructure:"regex,omitempty"`
	Ignore   []string          `mapstructure:"ignore,omitempty"`
	Packages []PackageConfig   `mapstructure:"packages,omitempty"`

	// Internal-facing, compiled rules
	RenameRules []RenameRule `mapstructure:"-"` // Ignored by mapstructure
}

// PackageConfig represents user-facing config for a specific package.
type PackageConfig struct {
	Import          string            `mapstructure:"import"`
	Path            string            `mapstructure:"path,omitempty"`
	Alias           string            `mapstructure:"alias,omitempty"`
	Prefix          string            `mapstructure:"prefix,omitempty"`
	Suffix          string            `mapstructure:"suffix,omitempty"`
	Explicit        map[string]string `mapstructure:"explicit,omitempty"`
	Regex           []RegexRule       `mapstructure:"regex,omitempty"`
	Ignore          []string          `mapstructure:"ignore,omitempty"`
	InheritExplicit bool              `mapstructure:"inherit_explicit,omitempty"`
	InheritRegex    bool              `mapstructure:"inherit_regex,omitempty"`
	InheritIgnore   bool              `mapstructure:"inherit_ignore,omitempty"`

	// Internal-facing, compiled rules
	RenameRules []RenameRule `mapstructure:"-"`
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
		v.SetConfigName("adptool")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	if err := v.ReadInConfig(); err != nil {
		if fileConfigPath != "" || !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		log.Println("Project-level adptool.yaml not found, proceeding with empty config.")
	} else {
		log.Printf("Config loaded successfully. Settings: %+v", v.AllSettings())
	}

	// Unmarshal the user-facing configuration.
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Compile all user-facing rules into internal RenameRules.
	compileRules(cfg)

	log.Printf("Final compiled config: %+v", cfg)
	return cfg, nil
}

// compileRules processes the entire config tree and compiles the user-facing
// rules into the internal RenameRules slice for the generator to use.
func compileRules(cfg *Config) {
	cfg.RenameRules = compileRuleSet(cfg.Prefix, cfg.Suffix, cfg.Explicit, cfg.Regex)

	// Process each package
	for i := range cfg.Packages {
		pkgCfg := &cfg.Packages[i] // Get a pointer to modify the original struct in the slice

		// Determine effective Prefix: package-level overrides global
		effectivePrefix := pkgCfg.Prefix
		if effectivePrefix == "" { // If package-level prefix is not set, inherit from global
			effectivePrefix = cfg.Prefix
		}

		// Determine effective Suffix: package-level overrides global
		effectiveSuffix := pkgCfg.Suffix
		if effectiveSuffix == "" { // If package-level suffix is not set, inherit from global
			effectiveSuffix = cfg.Suffix
		}

		// Determine effective Explicit map
		effectiveExplicit := make(map[string]string)
		if pkgCfg.InheritExplicit {
			// Start with global explicit rules
			for k, v := range cfg.Explicit {
				effectiveExplicit[k] = v
			}
			// Merge package-specific explicit rules
			if pkgCfg.Explicit != nil { // Check if package-specific explicit is defined
				for k, v := range pkgCfg.Explicit {
					effectiveExplicit[k] = v // Package-level explicit overrides global if keys conflict
				}
			}
		} else {
			if pkgCfg.Explicit != nil { // If package-specific explicit is explicitly defined (even if empty), it overrides global
				effectiveExplicit = pkgCfg.Explicit
			} else { // If package-specific explicit is not defined (nil), inherit global
				effectiveExplicit = cfg.Explicit
			}
		}

		// Determine effective Regex rules
		effectiveRegex := []RegexRule{}
		if pkgCfg.InheritRegex {
			effectiveRegex = append(effectiveRegex, cfg.Regex...)
			if pkgCfg.Regex != nil { // Append package-specific regex if defined
				effectiveRegex = append(effectiveRegex, pkgCfg.Regex...)
			}
		} else {
			if pkgCfg.Regex != nil { // If package-specific regex is explicitly defined (even if empty), it overrides global
				effectiveRegex = pkgCfg.Regex
			} else { // If package-specific regex is not defined (nil), inherit global
				effectiveRegex = cfg.Regex
			}
		}

		// Determine effective Ignore rules
		effectiveIgnore := []string{}
		if pkgCfg.InheritIgnore {
			effectiveIgnore = append(effectiveIgnore, cfg.Ignore...)
			if pkgCfg.Ignore != nil { // Append package-specific ignore if defined
				effectiveIgnore = append(effectiveIgnore, pkgCfg.Ignore...)
			}
		} else {
			if pkgCfg.Ignore != nil { // If package-specific ignore is explicitly defined (even if empty), it overrides global
				effectiveIgnore = pkgCfg.Ignore
			} else { // If package-specific ignore is not defined (nil), inherit global
				effectiveIgnore = cfg.Ignore
			}
		}

		// Compile the effective rules for this package
		pkgCfg.RenameRules = compileRuleSet(effectivePrefix, effectiveSuffix, effectiveExplicit, effectiveRegex)
		pkgCfg.Ignore = effectiveIgnore // Update the ignore list for the package config
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
