package rules

import (
	"fmt"
	"regexp"

	"github.com/origadmin/adptool/internal/config" // Assuming config is accessible
)

// RenameRule defines a single renaming rule.
type RenameRule struct {
	Type    string // e.g., "prefix", "suffix", "explicit", "regex"
	Value   string // For prefix/suffix
	From    string // For explicit
	To      string // For explicit
	Pattern string // For regex
	Replace string // For regex
}

// ConvertRuleSetToRenameRules converts a config.RuleSet to a slice of RenameRule.
func ConvertRuleSetToRenameRules(rs *config.RuleSet) []RenameRule {
	var renameRules []RenameRule

	if rs.Prefix != "" {
		renameRules = append(renameRules, RenameRule{Type: "prefix", Value: rs.Prefix})
	}
	if rs.Suffix != "" {
		renameRules = append(renameRules, RenameRule{Type: "suffix", Value: rs.Suffix})
	}
	for _, explicit := range rs.Explicit {
		renameRules = append(renameRules, RenameRule{Type: "explicit", From: explicit.From, To: explicit.To})
	}
	for _, regex := range rs.Regex {
		renameRules = append(renameRules, RenameRule{Type: "regex", Pattern: regex.Pattern, Replace: regex.Replace})
	}
	// Add other rule types as needed

	return renameRules
}

// ApplyRules applies a set of rename rules to a given name and returns the result.
func ApplyRules(name string, rules []RenameRule) (string, error) {
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
