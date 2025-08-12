package parser

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	gotoken "go/token"

	"github.com/origadmin/adptool/internal/config"
)

const directivePrefix = "//go:adapter:"

// ParseFileDirectives parses a Go source file (provided as an AST) and builds a config.Config object
// containing only the adptool directives found in that file.
// It does not perform any merging with global configurations.
func ParseFileDirectives(file *goast.File, fset *gotoken.FileSet) (*config.Config, error) {
	cfg := config.New()
	state := newParserState(cfg, fset, 0) // Line will be updated by Directive

	extractor := NewDirectiveExtractor(file, fset)

	for pd := extractor.Next(); pd != nil; pd = extractor.Next() {
		state.line = pd.Line // Update state with the current line number from Directive

		switch pd.BaseCmd {
		case "ignore": // Handle global ignore Directive
			if pd.IsJSON {
				var patterns []string
				if err := json.Unmarshal([]byte(pd.Argument), &patterns); err != nil {
					return nil, err
				}
				state.cfg.Ignores = append(state.cfg.Ignores, patterns...)
			} else {
				state.cfg.Ignores = append(state.cfg.Ignores, pd.Argument)
			}
		case "defaults":
			if err := handleDefaultsDirective(state, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "props":
			if err := handleVarsDirective(state, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "package":
			if err := handlePackageDirective(state, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "type":
			if err := handleTypeDirective(state, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "func":
			if err := handleFuncDirective(state, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "var":
			if err := handleVarDirective(state, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "const":
			if err := handleConstDirective(state, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "method", "field":
			if err := handleMemberDirective(state, pd.BaseCmd, pd.CmdParts, pd.Argument); err != nil {
				return nil, err
			}
		case "context", "done":
			// Context and done directives are handled by the loader, not the parser.
			// We will ignore them here.
		default:
			return nil, fmt.Errorf("line %d: unknown Directive \'%s\'", pd.Line, pd.Command)
		}
	}

	return cfg, nil
}
