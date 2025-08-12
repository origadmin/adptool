package parser

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	gotoken "go/token"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

const directivePrefix = "//go:adapter:"

// ParseFileDirectives parses a Go source file (provided as an AST) and builds a config.Config object
// containing only the adptool directives found in that file.
// It does not perform any merging with global configurations.
func ParseFileDirectives(file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	extractor := NewDirectiveExtractor(file, fset)
	context := NewContext()
	for {
		context.Directive = extractor.Next()
		if context.Directive == nil {
			break
		}
		pd := context.Directive
		switch pd.BaseCmd {
		case "ignore": // Handle local ignore Directive
			context.Config.Ignores = append(context.Config.Ignores, pd.Argument)
		case "ignores": // Handle global ignore Directive
			if pd.IsJSON {
				var patterns []string
				if err := json.Unmarshal([]byte(pd.Argument), &patterns); err != nil {
					return nil, err
				}
				context.Config.Ignores = append(context.Config.Ignores, patterns...)
			} else {
				patterns := strings.Split(pd.Argument, ",")
				for i := range patterns {
					patterns[i] = strings.TrimSpace(patterns[i])
				}
				context.Config.Ignores = append(context.Config.Ignores, patterns...)
			}
		case "default":
			if err := handleDefaultsDirective(context, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "prop":
			if err := handleVarsDirective(context, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "package":
			if err := handlePackageDirective(context, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "type":
			if err := handleTypeDirective(context, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "func":
			if err := handleFuncDirective(context, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "var":
			if err := handleVarDirective(context, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "const":
			if err := handleConstDirective(context, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "method", "field":
			if err := handleMemberDirective(context, pd.BaseCmd, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "context", "done":
			// Context and done directives are handled by the loader, not the parser.
			// We will ignore them here.
		default:
			return nil, fmt.Errorf("line %d: unknown Directive '%s'", pd.Line, pd.Command)
		}
	}

	return context.Config, nil
}
