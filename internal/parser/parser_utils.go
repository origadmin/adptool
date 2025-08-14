package parser

import (
	"fmt"
	"strings"
)

// newDirectiveError creates a formatted error with the directive's line number.
func newDirectiveError(directive *Directive, format string, args ...interface{}) error {
	return fmt.Errorf("command %s, line %d: %s", directive.Command, directive.Line, fmt.Sprintf(format, args...))
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
