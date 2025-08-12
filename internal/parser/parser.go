package parser

import (
	"encoding/json" // Added for JSON unmarshaling
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
	cfg := config.New()
	state := newParserState(cfg, fset, 0) // Line will be updated in the loop

	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			state.line = fset.Position(comment.Pos()).Line

			if !strings.HasPrefix(comment.Text, directivePrefix) {
				continue
			}

			rawDirective := strings.TrimPrefix(comment.Text, directivePrefix)
			// Clean rawDirective from any trailing comments
			commentStart := strings.Index(rawDirective, "//")
			if commentStart != -1 {
				rawDirective = strings.TrimSpace(rawDirective[:commentStart])
			}

			command, argument, baseCmd, cmdParts, isJsonArgument := parseDirective(rawDirective)

			// Helper to apply pending ignore arguments to a rule's RuleSet
			applyPendingIgnore := func(rs *config.RuleSet) {
				if len(state.pendingIgnoreArguments) > 0 {
					if rs.Ignores == nil {
						rs.Ignores = make([]string, 0)
					}
					rs.Ignores = append(rs.Ignores, state.pendingIgnoreArguments...)
					state.pendingIgnoreArguments = nil // Clear after applying
				}
			}

			switch baseCmd {
			case "ignore": // Handle global ignore directive
				if isJsonArgument {
					var patterns []string
					if err := json.Unmarshal([]byte(argument), &patterns); err != nil {
						return nil, fmt.Errorf("line %d: invalid JSON for ignore directive: %w", state.line, err)
					}
					state.cfg.Ignores = append(state.cfg.Ignores, patterns...)
				} else {
					state.cfg.Ignores = append(state.cfg.Ignores, argument)
				}
			case "defaults":
				if err := handleDefaultsDirective(state, cmdParts, argument); err != nil {
					return nil, err
				}
			case "props":
				if err := handleVarsDirective(state, cmdParts, argument); err != nil {
					return nil, err
				}
			case "package":
				if err := handlePackageDirective(state, cmdParts, argument, applyPendingIgnore); err != nil {
					return nil, err
				}
			case "type":
				if err := handleTypeDirective(state, cmdParts, argument, applyPendingIgnore); err != nil {
					return nil, err
				}
			case "func":
				if err := handleFuncDirective(state, cmdParts, argument, applyPendingIgnore); err != nil {
					return nil, err
				}
			case "var":
				if err := handleVarDirective(state, cmdParts, argument, applyPendingIgnore); err != nil {
					return nil, err
				}
			case "const":
				if err := handleConstDirective(state, cmdParts, argument, applyPendingIgnore); err != nil {
					return nil, err
				}
			case "method", "field":
				if err := handleMemberDirective(state, baseCmd, cmdParts, argument, applyPendingIgnore); err != nil {
					return nil, err
				}
			case "context", "done":
				// Context and done directives are handled by the loader, not the parser.
				// We will ignore them here.
			default:
				return nil, fmt.Errorf("line %d: unknown directive '%s'", state.line, command)
			}
		}
	}

	return cfg, nil
}
