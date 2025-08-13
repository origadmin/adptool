package parser

// Rule defines the common behavior for all directive rules, allowing for polymorphic
// handling of sub-directives.
type Rule interface {
	// ApplySubDirective applies a sub-command (e.g., ":rename", ":disabled") to the rule.
	// Each rule type is responsible for handling the sub-directives that are valid for it.
	ApplySubDirective(ctx *Context, subCmds []string, argument string) error
}
