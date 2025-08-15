package parser

import (
	"bytes"
	"fmt"
	"runtime"
)

// ParserError represents a structured error originating from the adptool parser.
// It includes a message, an optional wrapped error, the directive that caused it,
// and a stack trace for better debugging.
type parserError struct {
	Msg        string // Human-readable error message
	Context    any    // Additional context for the error
	StackTrace []byte // Captured stack trace
}

// Error implements the error interface for ParserError.
func (e *parserError) Error() string {
	var buf bytes.Buffer
	buf.WriteString(e.Msg) // Print Msg first

	if e.Context != nil {
		buf.WriteString(" [Context: ") // Separator and start of context block
		switch ctx := e.Context.(type) {
		case *Directive:
			buf.WriteString(fmt.Sprintf("Directive (line %d, original_cmd: %s, current_level_cmd: %s, argument: %s)",
				ctx.Line, ctx.Command, ctx.BaseCmd, ctx.Argument))
		// Add other context types here if needed in the future
		default:
			// Use %T to print the type of the context
			buf.WriteString(fmt.Sprintf("Type %T (%v)", ctx, ctx)) // Explicitly print type and value
		}
		buf.WriteString("]") // End of context block
	}

	// Add StackTrace
	if len(e.StackTrace) > 0 {
		buf.WriteString("\n--- Stack Trace ---\n")
		buf.Write(e.StackTrace)
		buf.WriteString("\n-------------------\n")
	}

	return buf.String()
}

// newDirectiveError is a helper to create a ParserError specifically for directive-related issues.
// This function will replace the existing newDirectiveError in parser.go.
func newDirectiveError(directive *Directive, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	stackBuf := make([]byte, 4096) // Capture stack here
	n := runtime.Stack(stackBuf, false)
	return &parserError{ // Directly create parserError
		Msg:        msg,
		Context:    directive, // Set Context to directive
		StackTrace: stackBuf[:n],
	}
}

// NewParserError creates a new parser error instance with a formatted message.
// It captures the current stack trace. This is for general parser errors
// not directly tied to a specific directive.
func NewParserError(format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false) // Capture stack trace, exclude goroutine info
	return &parserError{
		Msg:        msg,
		Context:    nil, // No context by default for general errors
		StackTrace: stackBuf[:n],
	}
}

// NewParserErrorWithContext creates a new parser error instance with a formatted message
// and an arbitrary context object. It captures the current stack trace.
func NewParserErrorWithContext(context any, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false) // Capture stack trace, exclude goroutine info
	return &parserError{
		Msg:        msg,
		Context:    context, // Set the provided context
		StackTrace: stackBuf[:n],
	}
}
