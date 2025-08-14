package parser

import (
	"fmt"
	goast "go/ast"
	gotoken "go/token"
	"log/slog"
	// Assuming these are the actual types from the config package
	// For now, using the placeholder types defined in context.go
	// "github.com/origadmin/adptool/internal/config"

	"github.com/origadmin/adptool/internal/config"
)

const directivePrefix = "//go:adapter:"

// parser orchestrates the parsing of Go directives into a structured configuration.
type parser struct {
	rootConfig  *RootConfig // The root configuration object
	rootContext *Context    // The current active parsing context (head of the linked list)
}

// newParser creates a new parser instance.
func newParser() *parser {
	rootCfg := NewRootConfig()            // Use constructor
	rootCtx := NewContext(rootCfg, false) // Create the initial context for the root

	return &parser{
		rootContext: rootCtx,
		rootConfig:  rootCfg,
	}
}

func NewRootConfig() *RootConfig {
	return &RootConfig{
		Config: config.New(),
	}
}

// ParseFileDirectives parses a Go source file and returns the built configuration.
// This is the exported entry point.
func ParseFileDirectives(file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	p := newParser() // Create a new parser instance
	return p.parseFile(file, fset)
}

// parseFile parses a Go source file and returns the built configuration.
func (p *parser) parseFile(file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	extractor := NewDirectiveExtractor(file, fset)
	var directive *Directive

	currentContext := p.rootContext
	for directive = range extractor.Seq() {
		slog.Info("Processing directive", "line", directive.Line, "command", directive.Command, "argument", directive.Argument)
		// Not a container-creating command. Handle context, done, sub-directives.
		var err error
		switch directive.BaseCmd {
		case "context":
			if currentContext.IsExplicit() && p.rootContext.Container() == nil { // Simplified check for empty explicit context
				return nil, newDirectiveError(directive, "consecutive 'context' directives without intervening rules are not allowed")
			}
			currentContext.SetExplicit(true)
		case "done":
			if !currentContext.IsExplicit() {
				return nil, newDirectiveError(directive, "'done' directive without a matching explicit 'context'")
			}
			currentContext.SetExplicit(false)
		default:
			// Delegate to the current context's container for any other commands.
			if err = p.rootContext.Container().ParseDirective(directive); err != nil {
				return nil, err
			}
		}
	}
	// Finalize the root config
	if err := currentContext.Container().Finalize(); err != nil {
		return nil, err
	}

	// Check for unclosed explicit contexts
	if currentContext.Parent() != nil { // Check if we returned to the original root context
		return nil, fmt.Errorf("unclosed 'context' block(s) detected at end of file")
	}

	return p.rootConfig.Config, nil
}
