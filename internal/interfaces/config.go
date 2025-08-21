package interfaces

import (
	"regexp"
)

// CompiledPackage holds the compiled information for a single source package.
type CompiledPackage struct {
	ImportPath  string
	ImportAlias string
	Types       interface{} // Types defined in this package
	Functions   interface{} // Functions defined in this package
	Variables   interface{} // Variables defined in this package
	Constants   interface{} // Constants defined in this package
}

// CompiledRenameRule represents a fully compiled and ready-to-apply renaming rule.
type CompiledRenameRule struct {
	Type          string         // e.g., "prefix", "suffix", "explicit", "regex"
	RuleType      RuleType       // The category of the rule (const, var, func, type)
	OriginalName  string         // The original name from the config rule (e.g., "*", "Worker")
	Value         string         // For prefix/suffix
	From          string         // For explicit
	To            string         // For explicit
	Pattern       string         // Original regex pattern string
	Replace       string         // Replacement string for regex
	CompiledRegex *regexp.Regexp // Pre-compiled regex for "regex" type rules
	Priority      int            // Priority of the rule
	IsWildcard    bool           // Indicates if the rule applies to all packages (wildcard)
}

// CompiledConfig holds all the compiled information needed for generation.
type CompiledConfig struct {
	PackageName string // The name of the package to be generated
	Packages    []*CompiledPackage

	// RulesByPackageAndType stores compiled rules, organized for efficient lookup.
	// Outer map key: PackageName (empty string for global/wildcard rules).
	// Inner map key: RuleType (e.g., RuleTypeType, RuleTypeFunc).
	// Value: A slice of CompiledRenameRule, sorted by Priority.
	RulesByPackageAndType map[string]map[RuleType][]CompiledRenameRule
}
