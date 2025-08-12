package parser

import (
	"encoding/json"
	"fmt"
	gotoken "go/token"
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// parsingContext holds the current default package information for directive resolution.
type parsingContext struct {
	DefaultPackageImportPath string
	DefaultPackageAlias      string
}

// state holds the mutable state during parsing.
type parserState struct {
	cfg  *config.Config
	fset *gotoken.FileSet
	line int

	lastTypeRule    *config.TypeRule
	lastFuncRule    *config.FuncRule
	lastVarRule     *config.VarRule
	lastConstRule   *config.ConstRule
	lastMemberRule  *config.MemberRule
	lastPackageRule *config.Package

	pendingIgnoreArguments []string
	currentContext         *parsingContext
}

func newParserState(cfg *config.Config, fset *gotoken.FileSet, line int) *parserState {
	return &parserState{
		cfg:            cfg,
		fset:           fset,
		line:           line,
		currentContext: &parsingContext{},
	}
}

// parseDirective extracts command, argument, base command, and command parts from a raw directive string.
func parseDirective(rawDirective string) (command, argument, baseCmd string, cmdParts []string) {
	parts := strings.SplitN(rawDirective, " ", 2)
	command = parts[0]
	argument = ""
	if len(parts) > 1 {
		// Strip inline comments from the argument
		argument = parts[1]
	}
	cmdParts = strings.Split(command, ":")
	baseCmd = cmdParts[0]
	return
}

// handleDefaultsDirective handles the parsing of defaults directives.
func handleDefaultsDirective(state *parserState, cmdParts []string, argument string) error {
	if len(cmdParts) < 2 || cmdParts[1] != "mode" || len(cmdParts) < 3 {
		return fmt.Errorf("line %d: invalid defaults directive format. Expected 'defaults:mode:<field> <value>'", state.line)
	}
	if state.cfg.Defaults == nil {
		state.cfg.Defaults = &config.Defaults{}
	}
	if state.cfg.Defaults.Mode == nil {
		state.cfg.Defaults.Mode = &config.Mode{}
	}
	modeField := cmdParts[2]
	switch modeField {
	case "strategy":
		state.cfg.Defaults.Mode.Strategy = argument
	case "prefix":
		state.cfg.Defaults.Mode.Prefix = argument
	case "suffix":
		state.cfg.Defaults.Mode.Suffix = argument
	case "explicit":
		state.cfg.Defaults.Mode.Explicit = argument
	case "regex":
		state.cfg.Defaults.Mode.Regex = argument
	case "ignores":
		state.cfg.Defaults.Mode.Ignores = argument
	default:
		return fmt.Errorf("line %d: unknown defaults mode field '%s'", state.line, modeField)
	}
	return nil
}

// handleVarsDirective handles the parsing of vars directives.
func handleVarsDirective(state *parserState, cmdParts []string, argument string) error {
	if len(cmdParts) != 1 {
		return fmt.Errorf("line %d: invalid vars directive format. Expected 'vars <name> <value>'", state.line)
	}
	varNameParts := strings.SplitN(argument, " ", 2)
	if len(varNameParts) != 2 {
		return fmt.Errorf("line %d: invalid vars directive argument. Expected 'name value'", state.line)
	}
	entry := &config.VarEntry{Name: varNameParts[0], Value: varNameParts[1]}
	if state.cfg.Props == nil {
		state.cfg.Props = make([]*config.VarEntry, 0)
	}
	state.cfg.Props = append(state.cfg.Props, entry)
	return nil
}

// handlePackageDirective handles the parsing of package directives.
func handlePackageDirective(state *parserState, cmdParts []string, argument string, applyPendingIgnore func(rs *config.RuleSet)) error {
	if len(cmdParts) == 1 { // //go:adapter:package <import_path> [alias] or //go:adapter:package <import_path>
		pkgParts := strings.SplitN(argument, " ", 2)
		if len(pkgParts) == 2 { // Context-setting form: //go:adapter:package <import_path> <alias>
			state.currentContext.DefaultPackageImportPath = pkgParts[0]
			state.currentContext.DefaultPackageAlias = pkgParts[1]
		} else if len(pkgParts) == 1 { // Config-adding form: //go:adapter:package <import_path>
			if state.cfg.Packages == nil {
				state.cfg.Packages = make([]*config.Package, 0)
			}
			pkg := &config.Package{Import: argument}
			state.cfg.Packages = append(state.cfg.Packages, pkg)
			state.lastPackageRule = pkg
			//applyPendingIgnore(&pkg.RuleSet)
			// Reset other last rules
			state.lastTypeRule, state.lastFuncRule, state.lastVarRule, state.lastConstRule, state.lastMemberRule = nil, nil, nil, nil, nil
		} else {
			return fmt.Errorf("line %d: invalid package directive format. Expected 'package <import_path>' or 'package <import_path> <alias>'", state.line)
		}
	} else if len(cmdParts) > 1 && cmdParts[0] == "package" && state.lastPackageRule != nil {
		subCmd := cmdParts[1]
		switch subCmd {
		case "alias":
			state.lastPackageRule.Alias = argument
		case "path":
			state.lastPackageRule.Path = argument
		case "vars":
			varNameParts := strings.SplitN(argument, " ", 2)
			if len(varNameParts) != 2 {
				return fmt.Errorf("line %d: invalid package vars directive argument. Expected 'name value'", state.line)
			}
			entry := &config.VarEntry{Name: varNameParts[0], Value: varNameParts[1]}
			if state.lastPackageRule.Vars == nil {
				state.lastPackageRule.Vars = make([]*config.VarEntry, 0)
			}
			state.lastPackageRule.Vars = append(state.lastPackageRule.Vars, entry)
		case "types":
			rule := &config.TypeRule{Name: argument, RuleSet: config.RuleSet{}}
			if state.lastPackageRule.Types == nil {
				state.lastPackageRule.Types = make([]*config.TypeRule, 0)
			}
			state.lastPackageRule.Types = append(state.lastPackageRule.Types, rule)
			state.lastTypeRule = rule
			applyPendingIgnore(&rule.RuleSet)
			state.lastFuncRule, state.lastVarRule, state.lastConstRule, state.lastMemberRule = nil, nil, nil, nil
		case "functions":
			rule := &config.FuncRule{Name: argument, RuleSet: config.RuleSet{}}
			if state.lastPackageRule.Functions == nil {
				state.lastPackageRule.Functions = make([]*config.FuncRule, 0)
			}
			state.lastPackageRule.Functions = append(state.lastPackageRule.Functions, rule)
			state.lastFuncRule = rule
			applyPendingIgnore(&rule.RuleSet)
			state.lastTypeRule, state.lastVarRule, state.lastConstRule, state.lastMemberRule = nil, nil, nil, nil
		case "variables":
			rule := &config.VarRule{Name: argument, RuleSet: config.RuleSet{}}
			if state.lastPackageRule.Variables == nil {
				state.lastPackageRule.Variables = make([]*config.VarRule, 0)
			}
			state.lastPackageRule.Variables = append(state.lastPackageRule.Variables, rule)
			state.lastVarRule = rule
			applyPendingIgnore(&rule.RuleSet)
			state.lastTypeRule, state.lastFuncRule, state.lastConstRule, state.lastMemberRule = nil, nil, nil, nil
		case "constants":
			rule := &config.ConstRule{Name: argument, RuleSet: config.RuleSet{}}
			if state.lastPackageRule.Constants == nil {
				state.lastPackageRule.Constants = make([]*config.ConstRule, 0)
			}
			state.lastPackageRule.Constants = append(state.lastPackageRule.Constants, rule)
			state.lastConstRule = rule
			applyPendingIgnore(&rule.RuleSet)
			state.lastTypeRule, state.lastVarRule, state.lastConstRule, state.lastMemberRule = nil, nil, nil, nil
		default:
			// Handle sub-directives of types/functions/variables/constants within a package
			if len(cmdParts) == 3 && cmdParts[2] == "struct" && state.lastTypeRule != nil {
				state.lastTypeRule.Pattern = argument
				state.lastTypeRule.Kind = "struct"
			} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && state.lastTypeRule != nil {
				state.lastTypeRule.Disabled = argument == "true"
			} else if len(cmdParts) == 3 && state.lastTypeRule != nil {
				// Generic sub-rule for type within package
				handleRule(&state.lastTypeRule.RuleSet, state.lastTypeRule.Name, cmdParts[2], argument)
			} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && state.lastFuncRule != nil {
				state.lastFuncRule.Disabled = argument == "true"
			} else if len(cmdParts) == 3 && state.lastFuncRule != nil {
				// Generic sub-rule for func within package
				handleRule(&state.lastFuncRule.RuleSet, state.lastFuncRule.Name, cmdParts[2], argument)
			} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && state.lastVarRule != nil {
				state.lastVarRule.Disabled = argument == "true"
			} else if len(cmdParts) == 3 && state.lastVarRule != nil {
				// Generic sub-rule for var within package
				handleRule(&state.lastVarRule.RuleSet, state.lastVarRule.Name, cmdParts[2], argument)
			} else if len(cmdParts) == 3 && cmdParts[2] == "disabled" && state.lastConstRule != nil {
				state.lastConstRule.Disabled = argument == "true"
			} else if len(cmdParts) == 3 && state.lastConstRule != nil {
				// Generic sub-rule for const within package
				handleRule(&state.lastConstRule.RuleSet, state.lastConstRule.Name, cmdParts[2], argument)
			} else {
				return fmt.Errorf("line %d: unknown package sub-directive '%s'", state.line, cmdParts[2])
			}
		}
	}
	return nil
}

// handleTypeDirective handles the parsing of type directives.
func handleTypeDirective(state *parserState, cmdParts []string, argument string, applyPendingIgnore func(rs *config.RuleSet)) error {
	if len(cmdParts) == 1 {
		rule := &config.TypeRule{Name: argument, Kind: "type", RuleSet: config.RuleSet{}}
		state.cfg.Types = append(state.cfg.Types, rule)
		state.lastTypeRule = rule
		applyPendingIgnore(&rule.RuleSet)
		state.lastFuncRule, state.lastVarRule, state.lastConstRule, state.lastMemberRule, state.lastPackageRule = nil, nil, nil, nil, nil
	} else if len(cmdParts) == 2 && cmdParts[1] == "struct" {
		if state.lastTypeRule == nil {
			return fmt.Errorf("line %d: 'type:struct' must follow a 'type' directive", state.line)
		}
		state.lastTypeRule.Pattern = argument
		state.lastTypeRule.Kind = "struct"
	} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
		if state.lastTypeRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a 'type' directive", state.line)
		}
		state.lastTypeRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 {
		// Generic sub-rule for type (e.g., :rename, :explicit)
		if state.lastTypeRule != nil {
			handleRule(&state.lastTypeRule.RuleSet, state.lastTypeRule.Name, cmdParts[1], argument)
		}
	}
	return nil
}

// handleFuncDirective handles the parsing of func directives.
func handleFuncDirective(state *parserState, cmdParts []string, argument string, applyPendingIgnore func(rs *config.RuleSet)) error {
	if len(cmdParts) == 1 {
		rule := &config.FuncRule{Name: argument, RuleSet: config.RuleSet{}}
		state.cfg.Functions = append(state.cfg.Functions, rule)
		state.lastFuncRule = rule
		applyPendingIgnore(&rule.RuleSet)
		state.lastTypeRule, state.lastVarRule, state.lastConstRule, state.lastMemberRule, state.lastPackageRule = nil, nil, nil, nil, nil
	} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
		if state.lastFuncRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a 'func' directive", state.line)
		}
		state.lastFuncRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && state.lastFuncRule != nil {
		// Generic sub-rule for func (e.g., :rename, :explicit)
		handleRule(&state.lastFuncRule.RuleSet, state.lastFuncRule.Name, cmdParts[1], argument)
	}
	return nil
}

// handleVarDirective handles the parsing of var directives.
func handleVarDirective(state *parserState, cmdParts []string, argument string, applyPendingIgnore func(rs *config.RuleSet)) error {
	if len(cmdParts) == 1 {
		rule := &config.VarRule{Name: argument, RuleSet: config.RuleSet{}}
		state.cfg.Variables = append(state.cfg.Variables, rule)
		state.lastVarRule = rule
		applyPendingIgnore(&rule.RuleSet)
		state.lastTypeRule, state.lastFuncRule, state.lastConstRule, state.lastMemberRule = nil, nil, nil, nil
	} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
		if state.lastVarRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a 'var' directive", state.line)
		}
		state.lastVarRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && state.lastVarRule != nil {
		// Generic sub-rule for var (e.g., :rename, :explicit)
		handleRule(&state.lastVarRule.RuleSet, state.lastVarRule.Name, cmdParts[1], argument)
	}
	return nil
}

// handleConstDirective handles the parsing of const directives.
func handleConstDirective(state *parserState, cmdParts []string, argument string, applyPendingIgnore func(rs *config.RuleSet)) error {
	if len(cmdParts) == 1 {
		rule := &config.ConstRule{Name: argument, RuleSet: config.RuleSet{}}
		state.cfg.Constants = append(state.cfg.Constants, rule)
		state.lastConstRule = rule
		applyPendingIgnore(&rule.RuleSet)
		state.lastTypeRule, state.lastFuncRule, state.lastVarRule, state.lastMemberRule = nil, nil, nil, nil
	} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
		if state.lastConstRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a 'const' directive", state.line)
		}
		state.lastConstRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && state.lastConstRule != nil {
		// Generic sub-rule for const (e.g., :rename, :explicit)
		handleRule(&state.lastConstRule.RuleSet, state.lastConstRule.Name, cmdParts[1], argument)
	}
	return nil
}

// handleMemberDirective handles the parsing of method and field directives.
func handleMemberDirective(state *parserState, baseCmd string, cmdParts []string, argument string, applyPendingIgnore func(rs *config.RuleSet)) error {
	if state.lastTypeRule == nil {
		return fmt.Errorf("line %d: '%s' directive must follow a 'type' directive", state.line, baseCmd)
	}
	if len(cmdParts) == 1 {
		member := &config.MemberRule{Name: argument, RuleSet: config.RuleSet{}}
		if baseCmd == "method" {
			state.lastTypeRule.Methods = append(state.lastTypeRule.Methods, member)
		} else {
			state.lastTypeRule.Fields = append(state.lastTypeRule.Fields, member)
		}
		state.lastMemberRule = member
		applyPendingIgnore(&member.RuleSet)
	} else if len(cmdParts) == 2 && cmdParts[1] == "disabled" {
		if state.lastMemberRule == nil {
			return fmt.Errorf("line %d: ':disabled' must follow a member directive", state.line)
		}
		state.lastMemberRule.Disabled = argument == "true"
	} else if len(cmdParts) == 2 && state.lastMemberRule != nil {
		// Generic sub-rule for method/field (e.g., :rename, :explicit)
		handleRule(&state.lastMemberRule.RuleSet, state.lastMemberRule.Name, cmdParts[1], argument)
	}
	return nil
}

// handleRule applies a sub-rule to the appropriate ruleset.
func handleRule(ruleset *config.RuleSet, fromName, ruleName, argument string) {
	if ruleset == nil {
		return
	}

	switch ruleName {
	case "rename":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: fromName, To: argument})
	case "explicit":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		explicitRules := strings.SplitN(argument, " ", 2)
		if len(explicitRules) == 2 {
			ruleset.Explicit = append(ruleset.Explicit, &config.ExplicitRule{From: explicitRules[0], To: explicitRules[1]})
		}
	case "explicit.json":
		if ruleset.Explicit == nil {
			ruleset.Explicit = make([]*config.ExplicitRule, 0)
		}
		var explicitRules []*config.ExplicitRule
		if err := json.Unmarshal([]byte(argument), &ruleset.Explicit); err == nil {
			ruleset.Explicit = explicitRules
		}
	}
}
