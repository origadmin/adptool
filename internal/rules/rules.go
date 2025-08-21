package rules

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/interfaces"
)

// ApplyRules applies a set of rename rules to a given name and returns the result.
func ApplyRules(name string, rules []interfaces.RenameRule) (string, error) {
	currentName := name
	for _, rule := range rules {
		switch rule.Type {
		case "explicit":
			if name == rule.From {
				return rule.To, nil // Explicit rule is final
			}
		case "prefix":
			currentName = rule.Value + currentName
		case "suffix":
			currentName = currentName + rule.Value
		case "regex":
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return "", fmt.Errorf("invalid regex pattern '%s': %w", rule.Pattern, err)
			}
			currentName = re.ReplaceAllString(currentName, rule.Replace)
		}
	}
	return currentName, nil
}

// ConvertRuleSetToRenameRules converts a RuleSet to a slice of RenameRule.
func ConvertRuleSetToRenameRules(rs *config.RuleSet) []interfaces.RenameRule {
	var renameRules []interfaces.RenameRule

	if rs == nil {
		return renameRules
	}

	// 1. Process explicit rules if present (highest priority, implies override)
	if len(rs.Explicit) > 0 {
		explicitRules := make([]*config.ExplicitRule, len(rs.Explicit))
		copy(explicitRules, rs.Explicit)
		sort.Slice(explicitRules, func(i, j int) bool {
			return explicitRules[i].From < explicitRules[j].From
		})
		for _, explicit := range explicitRules {
			renameRules = append(renameRules, interfaces.RenameRule{Type: "explicit", From: explicit.From, To: explicit.To})
		}
		return renameRules // If explicit rules are present, only they are processed
	}

	// 2. Else, process regex rules if present (next priority, implies override)
	if len(rs.Regex) > 0 {
		regexRules := make([]*config.RegexRule, len(rs.Regex))
		copy(regexRules, rs.Regex)
		sort.Slice(regexRules, func(i, j int) bool {
			return regexRules[i].Pattern < regexRules[j].Pattern
		})
		for _, regex := range regexRules {
			renameRules = append(renameRules, interfaces.RenameRule{Type: "regex", Pattern: regex.Pattern, Replace: regex.Replace})
		}
		return renameRules // If regex rules are present (and explicit were not), only they are processed
	}

	// 3. Else, process prefix and suffix rules (lowest priority)
	if rs.Prefix != "" {
		renameRules = append(renameRules, interfaces.RenameRule{Type: "prefix", Value: rs.Prefix})
	}

	if rs.Suffix != "" {
		renameRules = append(renameRules, interfaces.RenameRule{Type: "suffix", Value: rs.Suffix})
	}

	return renameRules
}