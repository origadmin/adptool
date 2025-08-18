package parser

import (
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

func ParseDirective(parentCtx *Context, ruleType RuleType, directive *Directive) error {
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
		var rt RuleType
		switch directive.BaseCmd {
		case "context":
			// Explicit context handling is temporarily disabled. Do nothing.
		case "done":
			// Explicit context handling is temporarily disabled. Do nothing.
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
			slog.Info("Processing rule directive", "line", directive.Line, "command", directive.Command, "argument",
				directive.Argument)
			err = ParseDirective(p.rootContext, rt, directive)
			if err != nil {
				return nil, err
			}
		}
	}

	if p.rootContext.IsActive() {
		err := p.rootContext.EndContext()
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