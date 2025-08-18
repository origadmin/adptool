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

func ParseDirective(context *Context, ruleType RuleType, directive *Directive) error {
	// Check for an unclosed explicit context. This is a hard error.
	if activeCtx := context.ActiveContext(); activeCtx != nil && activeCtx.IsExplicit() {
		return NewParserErrorWithContext(directive, "cannot start a new rule context; an explicit context is currently active and must be closed with a 'done' directive first")
	}

	// Start a new context. This will now also end any previous implicit context.
	currentContext, err := context.StartOrActiveContext(ruleType)
	if err != nil {
		return err
	}

	if directive.HasSub() {
		subDirective := directive.Sub()
		var rt RuleType
		switch subDirective.BaseCmd {
		case "package":
			rt = RuleTypePackage
		case "type":
			rt = RuleTypeType
		case "function", "func":
			rt = RuleTypeFunc
		case "variable", "var":
			rt = RuleTypeVar
		case "constant", "const":
			rt = RuleTypeConst
		}
		if rt != RuleTypeUnknown {
			err = ParseDirective(currentContext, rt, subDirective)
			if err != nil {
				return err
			}
		}
	} else {
		err = currentContext.Container().ParseDirective(directive)
		if err != nil {
			return err
		}
	}
	return nil
}

// parseFile parses a Go source file and returns the built configuration.
func (p *parser) parseFile(file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	iterator := NewDirectiveIterator(file, fset)
	for directive := range iterator {
		slog.Info("Processing directive", "line", directive.Line, "command", directive.BaseCmd, "argument",
			directive.Argument)

		var err error
		var rt RuleType
		switch directive.BaseCmd {
		case "context":
			if p.rootContext.IsExplicit() {
				return nil, NewParserErrorWithContext(directive, "consecutive or nested 'context' directives are not allowed")
			}
			p.rootContext.SetExplicit(true)
		case "done":
			if !p.rootContext.IsExplicit() {
				return nil, NewParserErrorWithContext(directive, "'done' directive without a matching explicit 'context'")
			}
			fmt.Println("Done processing directive")
			err = p.rootContext.EndContext()
			if err != nil {
				return nil, NewParserErrorWithContext(directive, "error ending context")
			}
		case "package":
			rt = RuleTypePackage
		case "type":
			rt = RuleTypeType
		case "function", "func":
			rt = RuleTypeFunc
		case "variable", "var":
			rt = RuleTypeVar
		case "constant", "const":
			rt = RuleTypeConst
		default:
			// If it's not a recognized directive, it's a regular directive
			// go:adapter:default:xxx
			// go:adapter:ignore xxx
			// go:adapter:ignores xxx
			// go:adapter:property xxx
			slog.Info("Processing regular directive", "line", directive.Line, "command", directive.BaseCmd, "argument",
				directive.Argument)
			if err = p.rootContext.Container().ParseDirective(directive); err != nil {
				return nil, err
			}
		}
		if rt != RuleTypeUnknown {
			// If it's a recognized directive, it's a rule directive
			// go:adapter:package xxx
			// go:adapter:type xxx
			// go:adapter:function xxx
			// go:adapter:variable xxx
			// go:adapter:constant xxx
			slog.Info("Processing rule directive", "line", directive.Line, "command", directive.BaseCmd, "argument",
				directive.Argument)
			err := ParseDirective(p.rootContext, rt, directive)
			if err != nil {
				return nil, err
			}
		}
	}

	if p.rootContext.IsExplicit() {
		return nil, NewParserError("unclosed 'context' block(s) detected at end of file")
	} else {
		if p.rootContext.IsActive() {
			err := p.rootContext.EndContext()
			if err != nil {
				return nil, NewParserError("error ending root context")
			}
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
