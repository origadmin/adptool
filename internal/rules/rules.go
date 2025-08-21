package rules

import (
	"fmt"

	"github.com/origadmin/adptool/internal/interfaces"
)

// ApplyRules applies a set of compiled rename rules to a given name and returns the result.
func ApplyRules(name string, rules []interfaces.CompiledRenameRule) (string, error) {
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
			// Use the pre-compiled regex
			if rule.CompiledRegex == nil {
				return "", fmt.Errorf("regex rule '%s' has no compiled regex", rule.Pattern)
			}
			currentName = rule.CompiledRegex.ReplaceAllString(currentName, rule.Replace)
		}
	}
	return currentName, nil
}