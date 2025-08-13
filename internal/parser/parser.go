package parser

import (
	"encoding/json"
	goast "go/ast"
	gotoken "go/token"
	"log/slog"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

const directivePrefix = "//go:adapter:"

// ParseFileDirectives parses a Go source file and builds a config.Config object.
func ParseFileDirectives(file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	extractor := NewDirectiveExtractor(file, fset)
	builder := NewConfigBuilder()

	for {
		d := extractor.Next()
		if d == nil {
			break
		}

		slog.Info("Processing directive", "line", d.Line, "command", d.Command, "argument", d.Argument)

		var err error
		switch d.BaseCmd {
		case "ignore":
			builder.config.Ignores = append(builder.config.Ignores, d.Argument)
		case "ignores":
			if d.IsJSON {
				var patterns []string
				if jsonErr := json.Unmarshal([]byte(d.Argument), &patterns); jsonErr != nil {
					err = jsonErr
				} else {
					builder.config.Ignores = append(builder.config.Ignores, patterns...)
				}
			} else {
				patterns := strings.Split(d.Argument, ",")
				for i := range patterns {
					patterns[i] = strings.TrimSpace(patterns[i])
				}
				builder.config.Ignores = append(builder.config.Ignores, patterns...)
			}
		case "default":
			err = handleDefaultDirective(builder, d)
		case "prop":
			err = handleVarsDirective(builder, d)
		case "package":
			err = handlePackageDirective(builder, d)
		case "type":
			err = handleTypeDirective(builder, d.SubCmds, d.Argument, d)
		case "func":
			err = handleFuncDirective(builder, d.SubCmds, d.Argument, d)
		case "var":
			err = handleVarDirective(builder, d.SubCmds, d.Argument, d)
		case "const":
			err = handleConstDirective(builder, d.SubCmds, d.Argument, d)
		case "done", "context": // Treat 'context ""' as done for backward compatibility
			if d.BaseCmd == "done" || (d.BaseCmd == "context" && d.Argument == "") {
				builder.EndPackageScope()
			}
		default:
			err = newDirectiveError(d, "unknown directive '%s'", d.Command)
		}

		if err != nil {
			return nil, err
		}
	}

	return builder.GetConfig(), nil
}
