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
	extractor := NewDirectiveIterator(file, fset)
	currentContext := p.rootContext
	var err error
	for directive := range extractor.Seq() {
		slog.Info("Processing directive", "line", directive.Line, "command", directive.Command, "argument", directive.Argument)

		var err error
		switch directive.Command {
		case "context":
			if currentContext.IsExplicit() && currentContext.Container() == nil {
				return nil, newDirectiveError(directive, "consecutive 'context' directives without intervening rules are not allowed")
			}
			currentContext.SetExplicit(true)
		case "done":
			if !currentContext.IsExplicit() {
				return nil, newDirectiveError(directive, "'done' directive without a matching explicit 'context'")
			}
			currentContext, err = currentContext.EndContext()
			if err != nil {
				return nil, fmt.Errorf("error ending context: %w", err)
			}
		// Directives that create new containers (scopes)
		case "packages":
			newContainer := NewContainer(RuleTypePackage)
			currentContext = currentContext.StartContext(newContainer)

		case "types":
			newContainer := NewContainer(RuleTypeType)
			currentContext = currentContext.StartContext(newContainer)

		case "functions":
			newContainer := NewContainer(RuleTypeFunc)
			currentContext = currentContext.StartContext(newContainer)

		case "variables":
			newContainer := NewContainer(RuleTypeVar)
			currentContext = currentContext.StartContext(newContainer)

		case "constants":
			newContainer := NewContainer(RuleTypeConst)
			currentContext = currentContext.StartContext(newContainer)
		default:
			if err = currentContext.Container().ParseDirective(directive); err != nil {
				return nil, err
			}
		}
	}

	// After processing all directives, ensure all contexts are properly ended.
	for currentContext != p.rootContext {
		slog.Info("Finalizing unclosed context at end of file", "container", currentContext.Container())
		currentContext, err = currentContext.EndContext()
		if err != nil {
			return nil, fmt.Errorf("error finalizing context: %w", err)
		}
	}

	// Finalize the root config. The parent for the root is nil.
	if err := p.rootContext.Container().Finalize(nil); err != nil {
		return nil, err
	}

	// This check might be redundant now if the loop above guarantees we are at rootContext.
	// However, it's a good final sanity check.
	if p.rootContext.Parent() != nil {
		return nil, fmt.Errorf("unclosed 'context' block(s) detected at end of file (post-finalization check)")
	}

	return p.rootConfig.Config, nil
}
