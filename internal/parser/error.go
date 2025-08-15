package parser

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
)

// ParserError represents a structured error originating from the adptool parser.
// It includes a message, an optional wrapped error, the directive that caused it,
// and a stack trace for better debugging.
type parserError struct {
	msg        string // Human-readable error message
	context    any    // Additional context for the error
	cause      error  // Wrapped error
	stackTrace []byte // Captured stack trace
}

// Error implements the error interface for ParserError.
func (e *parserError) Error() string {
	return e.msg
}

func (e *parserError) Unwrap() error {
	return e.cause
}

func (e *parserError) Cause() error {
	return e.cause
}

// String implements the fmt.Stringer interface, providing a detailed error message for debugging.
func (e *parserError) String() string {
	var buf bytes.Buffer
	buf.WriteString(e.msg) // Print msg first

	if e.context != nil {
		buf.WriteString(" [Context: ") // Separator and start of context block
		switch ctx := e.context.(type) {
		case *Directive:
			buf.WriteString(fmt.Sprintf("Directive (line %d, original_cmd: %s, current_level_cmd: %s, argument: %s)",
				ctx.Line, ctx.Command, ctx.BaseCmd, ctx.Argument))
		// Add other context types here if needed in the future
		case error:
			buf.WriteString(fmt.Sprintf("Error: %v", ctx))
		default:
			// Use %T to print the type of the context
			buf.WriteString(fmt.Sprintf("Type %T (%v)", ctx, ctx)) // Explicitly print type and value
		}
		buf.WriteString("]") // End of context block
	}

	// Add stackTrace
	if len(e.stackTrace) > 0 {
		buf.WriteString("\n--- Stack Trace ---\n")
		buf.Write(e.stackTrace)
		buf.WriteString("\n-------------------\n")
	}

	return buf.String()
}

// Is implements the errors.Is interface, comparing parserError instances by their message.
func (e *parserError) Is(target error) bool {
	var pe *parserError
	if errors.As(target, &pe) {
		return e.msg == pe.msg
	}
	return false
}



// NewParserError creates a new parser error instance with a formatted message.
// It captures the current stack trace. This is for general parser errors
// not directly tied to a specific directive.
func NewParserError(format string, args ...any) error {
	baseError := fmt.Errorf(format, args...)
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false) // Capture stack trace, exclude goroutine info
	return &parserError{
		msg:        baseError.Error(),
		context:    nil, // No context by default for general errors
		cause:      errors.Unwrap(baseError),
		stackTrace: stackBuf[:n],
	}
}
func NewParserErrorWithCause(cause error, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false) // Capture stack trace, exclude goroutine info
	return &parserError{
		msg:        msg,
		context:    nil, // No context by default for general errors
		cause:      cause,
		stackTrace: stackBuf[:n],
	}
}

func NewParserErrorWithCauseAndContext(cause error, context any, format string, args ...any) error {
	baseError := fmt.Sprintf(format, args...)

	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false) // Capture stack trace, exclude goroutine info
	return &parserError{
		msg:        baseError,
		context:    context, // Set the provided context
		cause:      cause,   // Set the wrapped error if provided
		stackTrace: stackBuf[:n],
	}
}

// NewParserErrorWithContext creates a new parser error instance with a formatted message
// and an arbitrary context object. It captures the current stack trace.
func NewParserErrorWithContext(context any, format string, args ...any) error {
	baseError := fmt.Errorf(format, args...)

	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false) // Capture stack trace, exclude goroutine info
	return &parserError{
		msg:        baseError.Error(),
		context:    context,                  // Set the provided context
		cause:      errors.Unwrap(baseError), // Set the wrapped error if provided
		stackTrace: stackBuf[:n],
	}
}
