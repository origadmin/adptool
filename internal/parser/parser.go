package parser

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	gotoken "go/token"
	"log/slog"
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
		slog.Info("Processing directive", "line", pd.Line, "command", pd.Command, "argument", pd.Argument)
		switch pd.BaseCmd {
		case "ignore": // Handle local ignore Directive
			slog.Info("Handling 'ignore' directive")
			context.Config.Ignores = append(context.Config.Ignores, pd.Argument)
		case "ignores": // Handle global ignore Directive
			slog.Info("Handling 'ignores' directive")
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
			slog.Info("Handling 'default' directive")
			if err := handleDefaultDirective(context, pd.SubCmds, pd.Argument); err != nil {
				return nil, err
			}
		case "prop":
			slog.Info("Handling 'prop' directive")
			if err := handleVarsDirective(context, pd.SubCmds, pd.Argument); err != nil {
				return nil, err
			}
		case "package":
			slog.Info("Handling 'package' directive")
			if err := handlePackageDirective(context, pd.SubCmds, pd.Argument); err != nil {
				return nil, err
			}
		case "type":
			slog.Info("Handling 'type' directive")
			if err := handleTypeDirective(context, pd.SubCmds, pd.Argument); err != nil {
				return nil, err
			}
		case "func":
			slog.Info("Handling 'func' directive")
			if err := handleFuncDirective(context, pd.SubCmds, pd.Argument); err != nil {
				return nil, err
			}
		case "var":
			slog.Info("Handling 'var' directive")
			if err := handleVarDirective(context, pd.SubCmds, pd.Argument); err != nil {
				return nil, err
			}
		case "const":
			slog.Info("Handling 'const' directive")
			if err := handleConstDirective(context, pd.SubCmds, pd.Argument); err != nil {
				return nil, err
			}
		case "context", "done":
			slog.Info("Handling 'context' or 'done' directive")
			// Treat both 'context ""' and 'done' as scope-ending directives for backward compatibility with tests.
			if pd.BaseCmd == "done" || (pd.BaseCmd == "context" && pd.Argument == "") {
				context.EndPackageScope()
			}
		default:
			slog.Error("Unknown directive", "line", pd.Line, "command", pd.Command)
			return nil, fmt.Errorf("line %d: unknown Directive '%s'", pd.Line, pd.Command)
		}
	}

	return context.Config, nil
}
