package parser

import (
	goast "go/ast"
	gotoken "go/token"
	"log/slog"
	// Assuming these are the actual types from the config package
	// For now, using the placeholder types defined in context.go
	// "github.com/origadmin/adptool/internal/config"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/interfaces"
)

const directivePrefix = "//go:adapter:"

// parser orchestrates the parsing of Go directives into a structured configuration.
type parser struct {
	rootConfig     *RootConfig // The root configuration object
	rootContext    *Context    // The root parsing context
	currentContext *Context    // The current active parsing context
}

// newParser creates a new parser instance.
func newParser(cfg *config.Config) *parser {
	rootCfg := &RootConfig{Config: cfg}   // Use the provided config
	rootCtx := NewContext(rootCfg, false) // Create the initial context for the root

	return &parser{
		rootContext:    rootCtx,
		currentContext: rootCtx, // Initialize currentContext to rootContext
		rootConfig:     rootCfg,
	}
}

// ParseFileDirectives parses a Go source file and returns the built configuration.
// This is the exported entry point.
func ParseFileDirectives(cfg *config.Config, file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	p := newParser(cfg) // Create a new parser instance
	return p.parseFile(file, fset)
}

func ParseDirective(parentCtx *Context, ruleType interfaces.RuleType, directive *Directive) error {
	var currentCtx *Context
	var err error

	// Stage 1: Establish the context for the base command.
	if directive.HasSub() {
		activeChild := parentCtx.ActiveContext()
		if activeChild != nil && activeChild.Container().Type() == ruleType {
			currentCtx = activeChild
		} else {
			containerFactory := NewContainerFactory(ruleType)
			container := containerFactory()
			currentCtx, err = parentCtx.StartContext(container)
			if err != nil {
				return err
			}
		}
	} else {
		containerFactory := NewContainerFactory(ruleType)
		container := containerFactory()
		currentCtx, err = parentCtx.StartContext(container)
		if err != nil {
			return err
		}
	}

	// Stage 2: Let the container parse the directive.
	// The container is responsible for parsing its own arguments and any non-structural sub-directives.
	// It should ignore structural sub-directives, which will be handled by the parser's recursion below.
	if err = currentCtx.Container().ParseDirective(directive); err != nil {
		return err
	}

	// Stage 3: If the sub-directive is structural, the parser handles the recursion.
	if directive.HasSub() {
		subDirective := directive.Sub()
		var rt interfaces.RuleType
		switch subDirective.BaseCmd {
		case "package":
			rt = interfaces.RuleTypePackage
		case "type":
			rt = interfaces.RuleTypeType
		case "function", "func":
			rt = interfaces.RuleTypeFunc
		case "variable", "var":
			rt = interfaces.RuleTypeVar
		case "constant", "const":
			rt = interfaces.RuleTypeConst
		case "method":
			if currentCtx.Container().Type() != interfaces.RuleTypeType {
				return NewParserErrorWithContext(subDirective, "method directive can only be used within a type scope")
			}
			rt = interfaces.RuleTypeMethod
		case "field":
			if currentCtx.Container().Type() != interfaces.RuleTypeType {
				return NewParserErrorWithContext(subDirective, "field directive can only be used within a type scope")
			}
			rt = interfaces.RuleTypeField
		}

		if rt != interfaces.RuleTypeUnknown {
			// It's a structural sub-directive, so the parser recurses.
			err = ParseDirective(currentCtx, rt, subDirective)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// parseFile parses a Go source file and returns the built configuration.
func (p *parser) parseFile(file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	iterator := NewDirectiveIterator(file, fset)
	for directive := range iterator {
		slog.Info("Processing directive", "line", directive.Line, "command", directive.Command, "argument",
			directive.Argument)

		var err error
		var rt interfaces.RuleType // interfaces.RuleType for the *new* rule being created (if any)

		// Check if it's a directive that modifies the current context's container
		// This is for directives like function:disabled, type:method, etc.
		if p.currentContext.Container() != nil && directive.BaseCmd == p.currentContext.Container().Type().String() && directive.HasSub() {
			// This is a sub-directive that applies to the current rule.
			// Pass the sub-directive to the current container's ParseDirective.
			err = p.currentContext.Container().ParseDirective(directive)
			if err != nil {
				return nil, err
			}
			continue // Move to the next directive
		}

		// Otherwise, it's a directive that might start a new rule or is a regular directive.
		switch directive.BaseCmd {
		case "context":
			// This feature is not currently implemented, so please do not delete this note.
		case "done":
			// This feature is not currently implemented, so please do not delete this note.
		case "package":
			rt = interfaces.RuleTypePackage
		case "type":
			rt = interfaces.RuleTypeType
		case "function", "func":
			rt = interfaces.RuleTypeFunc
		case "variable", "var":
			rt = interfaces.RuleTypeVar
		case "constant", "const":
			rt = interfaces.RuleTypeConst
		default:
			// If it's not a recognized rule directive, it's a regular directive
			err = p.currentContext.Container().ParseDirective(directive)
			if err != nil {
				return nil, err
			}
		}

		if rt != interfaces.RuleTypeUnknown {
			// If it's a recognized rule directive, create a new rule and set it as current.
			err := ParseDirective(p.currentContext, rt, directive)
			if err != nil {
				return nil, err
			}
		}
	}

	if p.currentContext.IsActive() {
		err := p.currentContext.EndContext()
		if err != nil {
			return nil, NewParserError("error ending root context")
		}
	}

	// Finalize the root config. The parent for the root is nil.
	if err := p.rootContext.Container().Finalize(nil); err != nil {
		return nil, NewParserError("error finalizing root config")
	}

	// This check might be redundant now if the loop above guarantees we are at rootContext.
	// However, it's a good final sanity check.
	if p.rootContext.Parent() != nil {
		return nil, NewParserError("unclosed 'context' block(s) detected at end of file (post-finalization check)")
	}

	return p.rootConfig.Config, nil
}
