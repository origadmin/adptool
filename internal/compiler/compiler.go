package compiler

import (
	"fmt"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// CompiledConfig represents the fully resolved and flattened configuration after the compile stage.
// It's optimized for quick lookup during the apply stage.
type CompiledConfig struct {
	// Global rules (after variable substitution)
	GlobalTypes     *config.TypeRuleSet
	GlobalFunctions *config.RuleSet
	GlobalVariables *config.RuleSet
	GlobalConstants *config.RuleSet

	// Package-specific overrides (after variable substitution, not merged with global)
	Packages map[string]*CompiledPackage

	// Global resolved variables
	ResolvedVars map[string]string

	// Default modes for rule merging
	DefaultModes *config.Mode
}

// CompiledPackage holds compiled rules and resolved variables for a specific package.
type CompiledPackage struct {
	Import string
	Alias  string

	// Package-specific rules (after variable substitution, not merged with global)
	Types     *config.TypeRuleSet
	Functions *config.RuleSet
	Variables *config.RuleSet
	Constants *config.RuleSet

	ResolvedVars map[string]string // Resolved for this package
}

// Compile processes the raw Config and produces a CompiledConfig.
// This stage resolves variables and merges hierarchical rules.
func Compile(cfg *config.Config) (*CompiledConfig, error) {
	compiled := &CompiledConfig{
		Packages: make(map[string]*CompiledPackage),
	}

	// Set default modes from config, or use hardcoded defaults if not specified
	compiled.DefaultModes = &config.Mode{
		Strategy: "replace",
		Prefix:   "append",
		Suffix:   "append",
		Explicit: "merge",
		Regex:    "merge",
		Ignore:   "merge",
	}
	// Override with user-defined defaults
	if cfg.Defaults != nil && cfg.Defaults.Mode != nil {
		if cfg.Defaults.Mode.Strategy != "" {
			compiled.DefaultModes.Strategy = cfg.Defaults.Mode.Strategy
		}
		if cfg.Defaults.Mode.Prefix != "" {
			compiled.DefaultModes.Prefix = cfg.Defaults.Mode.Prefix
		}
		if cfg.Defaults.Mode.Suffix != "" {
			compiled.DefaultModes.Suffix = cfg.Defaults.Mode.Suffix
		}
		if cfg.Defaults.Mode.Explicit != "" {
			compiled.DefaultModes.Explicit = cfg.Defaults.Mode.Explicit
		}
		if cfg.Defaults.Mode.Regex != "" {
			compiled.DefaultModes.Regex = cfg.Defaults.Mode.Regex
		}
		if cfg.Defaults.Mode.Ignore != "" {
			compiled.DefaultModes.Ignore = cfg.Defaults.Mode.Ignore
		}
	}

	// 1. Resolve global variables
	globalResolvedVars, err := resolveVars(cfg.Vars, nil, nil) // No parent vars, no stack
	if err != nil {
		return nil, fmt.Errorf("failed to resolve global variables: %w", err)
	}
	compiled.ResolvedVars = globalResolvedVars

	// 2. Compile global rules (only variable substitution, no merging here)
	compiled.GlobalTypes, err = compileTypeRuleSet(cfg.Types, globalResolvedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to compile global types rules: %w", err)
	}
	compiled.GlobalFunctions, err = compileRuleSet(cfg.Functions, globalResolvedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to compile global functions rules: %w", err)
	}
	compiled.GlobalVariables, err = compileRuleSet(cfg.Variables, globalResolvedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to compile global variables rules: %w", err)
	}
	compiled.GlobalConstants, err = compileRuleSet(cfg.Constants, globalResolvedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to compile global constants rules: %w", err)
	}

	// 3. Compile package-specific rules (only variable substitution, no merging with global here)
	for _, pkgCfg := range cfg.Packages {
		pkgCompiled := &CompiledPackage{
			Import: pkgCfg.Import,
			Alias:  pkgCfg.Alias,
		}

		// Resolve package variables (inherits from global)
		pkgResolvedVars, err := resolveVars(pkgCfg.Props, globalResolvedVars, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve variables for package %s: %w", pkgCfg.Import, err)
		}
		pkgCompiled.ResolvedVars = pkgResolvedVars

		// Compile package rules (only variable substitution)
		pkgCompiled.Types, err = compileTypeRuleSet(pkgCfg.Types, pkgResolvedVars)
		if err != nil {
			return nil, fmt.Errorf("failed to compile types rules for package %s: %w", pkgCfg.Import, err)
		}
		pkgCompiled.Functions, err = compileRuleSet(pkgCfg.Functions, pkgResolvedVars)
		if err != nil {
			return nil, fmt.Errorf("failed to compile functions rules for package %s: %w", pkgCfg.Import, err)
		}
		pkgCompiled.Variables, err = compileRuleSet(pkgCfg.Variables, pkgResolvedVars)
		if err != nil {
			return nil, fmt.Errorf("failed to compile variables rules for package %s: %w", pkgCfg.Import, err)
		}
		pkgCompiled.Constants, err = compileRuleSet(pkgCfg.Constants, pkgResolvedVars)
		if err != nil {
			return nil, fmt.Errorf("failed to compile constants rules for package %s: %w", pkgCfg.Import, err)
		}

		compiled.Packages[pkgCfg.Import] = pkgCompiled
	}

	return compiled, nil
}

// resolveVars resolves compile-time variables recursively.
// It detects circular dependencies and handles variable substitution.
func resolveVars(localVars map[string]string, parentResolvedVars map[string]string, resolutionStack map[string]bool) (map[string]string, error) {
	if resolutionStack == nil {
		resolutionStack = make(map[string]bool)
	}

	resolved := make(map[string]string)

	// Copy parent resolved vars first
	for k, v := range parentResolvedVars {
		resolved[k] = v
	}

	for k, v := range localVars {
		if resolutionStack[k] {
			return nil, fmt.Errorf("circular dependency detected for variable: %s", k)
		}
		resolutionStack[k] = true

		// Parse and substitute variables in the current value
		parsedValue, err := parseAndSubstitute(v, resolved, resolutionStack) // Pass 'resolved' for already resolved vars in current scope
		if err != nil {
			return nil, fmt.Errorf("failed to resolve variable %s: %w", k, err)
		}
		resolved[k] = parsedValue

		delete(resolutionStack, k)
	}

	return resolved, nil
}

// parseAndSubstitute parses a string for ${...} variables and substitutes them.
// It recursively calls resolveVars for nested variables.
func parseAndSubstitute(s string, resolvedVars map[string]string, resolutionStack map[string]bool) (string, error) {
	var builder strings.Builder
	start := 0

	for {
		idx := strings.Index(s[start:], "${")
		if idx == -1 {
			builder.WriteString(s[start:])
			break
		}

		builder.WriteString(s[start : start+idx])
		varStart := start + idx
		varEnd := strings.Index(s[varStart:], "}")
		if varEnd == -1 {
			return "", fmt.Errorf("unclosed variable placeholder in: %s", s[varStart:])
		}

		varName := s[varStart+2 : varStart+varEnd]

		if resolvedVal, ok := resolvedVars[varName]; ok {
			builder.WriteString(resolvedVal)
		} else {
			return "", fmt.Errorf("undefined variable: %s", varName)
		}

		start = varStart + varEnd + 1
	}

	return builder.String(), nil
}

// compileRuleSet performs variable substitution on a RuleSet.
func compileRuleSet(rs *config.RuleSet, resolvedVars map[string]string) (*config.RuleSet, error) {
	if rs == nil {
		return &config.RuleSet{
			Strategy: []string{},
			Explicit: []*config.ExplicitRule{},
			Regex:    []*config.RegexRule{},
			Ignores:  []string{},
		}, nil
	}

	// Create a copy to avoid modifying the original (if it's a parent RuleSet)
	copiedRs := &config.RuleSet{
		Strategy:     make([]string, len(rs.Strategy)),
		Prefix:       rs.Prefix,
		PrefixMode:   rs.PrefixMode,
		Suffix:       rs.Suffix,
		SuffixMode:   rs.SuffixMode,
		Explicit:     make([]*config.ExplicitRule, len(rs.Explicit)),
		ExplicitMode: rs.ExplicitMode,
		Regex:        make([]*config.RegexRule, len(rs.Regex)),
		RegexMode:    rs.RegexMode,
		Ignores:      make([]string, len(rs.Ignores)),
		IgnoresMode:  rs.IgnoresMode,
	}

	// Deep copy for slices and pointers
	copy(copiedRs.Strategy, rs.Strategy)
	copy(copiedRs.Ignores, rs.Ignores)

	for i, r := range rs.Explicit {
		copiedRs.Explicit[i] = &config.ExplicitRule{From: r.From, To: r.To}
	}
	for i, r := range rs.Regex {
		copiedRs.Regex[i] = &config.RegexRule{Pattern: r.Pattern, Replace: r.Replace}
	}

	// Handle the new nested Transform struct
	if rs.Transform != nil {
		copiedRs.Transform = &config.Transform{
			Before: rs.Transform.Before,
			After:  rs.Transform.After,
		}
	}

	// Substitute variables in string fields
	var err error
	copiedRs.Prefix, err = substituteString(copiedRs.Prefix, resolvedVars)
	if err != nil {
		return nil, err
	}
	copiedRs.Suffix, err = substituteString(copiedRs.Suffix, resolvedVars)
	if err != nil {
		return nil, err
	}

	// Substitute variables in the new Transform struct
	if copiedRs.Transform != nil {
		copiedRs.Transform.Before, err = substituteString(copiedRs.Transform.Before, resolvedVars)
		if err != nil {
			return nil, err
		}
		copiedRs.Transform.After, err = substituteString(copiedRs.Transform.After, resolvedVars)
		if err != nil {
			return nil, err
		}
	}

	// Substitute variables in ExplicitRule and RegexRule fields
	for _, r := range copiedRs.Explicit {
		r.From, err = substituteString(r.From, resolvedVars)
		if err != nil {
			return nil, err
		}
		r.To, err = substituteString(r.To, resolvedVars)
		if err != nil {
			return nil, err
		}
	}
	for _, r := range copiedRs.Regex {
		r.Pattern, err = substituteString(r.Pattern, resolvedVars)
		if err != nil {
			return nil, err
		}
		r.Replace, err = substituteString(r.Replace, resolvedVars)
		if err != nil {
			return nil, err
		}
	}

	return copiedRs, nil
}

// compileTypeRuleSet performs variable substitution on a TypeRuleSet.
func compileTypeRuleSet(trs *config.TypeRuleSet, resolvedVars map[string]string) (*config.TypeRuleSet, error) {
	if trs == nil {
		return config.NewTypeRuleSet(), nil // Return an initialized empty TypeRuleSet if nil
	}

	compiledTrs := config.NewTypeRuleSet()

	// Compile the base RuleSet part
	baseRuleSet, err := compileRuleSet(trs.RuleSet, resolvedVars)
	if err != nil {
		return nil, err
	}
	compiledTrs.RuleSet = baseRuleSet

	// Compile nested methods rules
	compiledTrs.Methods, err = compileRuleSet(trs.Methods, resolvedVars)
	if err != nil {
		return nil, err
	}

	// Compile nested fields rules
	compiledTrs.Fields, err = compileRuleSet(trs.Fields, resolvedVars)
	if err != nil {
		return nil, err
	}

	return compiledTrs, nil
}
