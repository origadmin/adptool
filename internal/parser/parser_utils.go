package parser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// newDirectiveError creates a formatted error with the directive's line number.
func newDirectiveError(directive *Directive, format string, args ...interface{}) error {
	return fmt.Errorf("line %d: %s", directive.Line, fmt.Sprintf(format, args...))
}

// parseNameValue parses an argument string into a name and value.
// Expected format: "name value"
func parseNameValue(argument string) (name, value string, err error) {
	parts := strings.SplitN(argument, " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("argument must be in 'name value' format")
	}
	return parts[0], parts[1], nil
}

// parseFromTo parses an argument string into 'from' and 'to' values.
// Expected format: "from to"
func parseFromTo(argument string) (from, to string, err error) {
	parts := strings.SplitN(argument, " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("argument must be in 'from to' format")
	}
	return parts[0], parts[1], nil
}

// ensureSliceInitialized ensures that a slice is not nil before appending.
// This is a generic helper, but for specific types, direct initialization might be clearer.
// For now, I'll keep it simple and use specific initializations in the handlers.
// If needed, this can be refactored later.

// applyRuleToRuleSet applies a sub-rule to the appropriate ruleset.
func applyRuleToRuleSet(ruleset *config.RuleSet, fromName, ruleName, argument string) error {
	slog.Debug("Entering applyRuleToRuleSet", "fromName", fromName, "ruleName", ruleName, "argument", argument)
	if ruleset == nil {
		return fmt.Errorf("ruleset is nil")
	}

	switch ruleName {
	case "rename":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: fromName, To: argument})
	case "explicit":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		explicitRules := strings.SplitN(argument, " ", 2)
		if len(explicitRules) == 2 {
			ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: explicitRules[0], To: explicitRules[1]})
		} else {
			return fmt.Errorf("explicit rule argument must be in 'from to' format")
		}
	case "explicit:json":
		var explicitRules []*config.ExplicitRule
		if err := json.Unmarshal([]byte(argument), &explicitRules); err == nil {
			ruleset.Explicit = explicitRules
		} else {
			return fmt.Errorf("invalid JSON for explicit:json: %w", err)
		}
	case "regex:json":
		var regexRules []*config.RegexRule
		if err := json.Unmarshal([]byte(argument), &regexRules); err == nil {
			ruleset.Regex = regexRules
		} else {
			return fmt.Errorf("invalid JSON for regex:json: %w", err)
		}
	case "strategy:json":
		var strategies []string
		if err := json.Unmarshal([]byte(argument), &strategies); err == nil {
			ruleset.Strategy = strategies
		} else {
			return fmt.Errorf("invalid JSON for strategy:json: %w", err)
		}
	case "ignores":
		ruleset.Ignores = append(ruleset.Ignores, argument)
	default:
		return fmt.Errorf("unknown rule name: %s", ruleName)
	}
	return nil
}
