package parser

import (
	"strings"
)

// Directive represents a parsed adptool directive from a Go comment.
// It is immutable after creation.
type Directive struct {
	Line     int    // Line number in the source file.
	Command  string // The full command string (e.g., "type:struct"). Note: :json suffix is removed here.
	Argument string // The raw argument string.

	// Parsed components of the command.
	BaseCmd string   // The base command (e.g., "type").
	SubCmds []string // Sub-commands (e.g., ["struct"]).
	IsJSON  bool     // True if the original command had a ":json" suffix.
}

func (d Directive) Root() *Directive {
	newDirective := d                         // Copy the original directive
	cmdParts := strings.Split(d.Command, ":") // Use a local variable for cmdParts
	newDirective.BaseCmd = cmdParts[0]
	newDirective.SubCmds = cmdParts[1:]
	return &newDirective
}

// Sub creates a new Directive with the given sub-command appended to its BaseCmd.
// The new Directive's Command will be BaseCmd:subCmd, and SubCmds will be updated.
// The original Directive's Line, Argument, and IsJSON are preserved.
func (d Directive) Sub() (*Directive, bool) {
	if len(d.SubCmds) == 0 {
		return &Directive{}, false
	}
	newDirective := d // Copy the original directive
	newDirective.BaseCmd = d.SubCmds[0]
	newDirective.SubCmds = d.SubCmds[1:]
	return &newDirective, true
}

// parseDirective extracts command, argument, and their parsed components from a raw directive string.
func parseDirective(rawDirective string, line int) Directive {
	var directive Directive
	parts := strings.SplitN(rawDirective, " ", 2)
	directive.Line = line
	directive.Command = parts[0]
	directive.Argument = ""
	if len(parts) > 1 {
		directive.Argument = parts[1]
	}

	// Determine IsJSON and clean Command
	if strings.HasSuffix(directive.Command, ":json") {
		directive.IsJSON = true
		directive.Command = strings.TrimSuffix(directive.Command, ":json")
	}

	cmdParts := strings.Split(directive.Command, ":") // Use a local variable for cmdParts
	directive.BaseCmd = cmdParts[0]
	directive.SubCmds = cmdParts[1:]
	return directive
}
