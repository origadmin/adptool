package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

func TestRootConfigParseDirectiveDefaults(t *testing.T) {
	// Setup
	rc := &RootConfig{Config: config.New()}

	// Test case 1: Basic defaults directive (no argument, no sub-command)
	dir := &Directive{BaseCmd: "default", Command: "default", Argument: ""}
	err := rc.ParseDirective(dir)
	assert.NoError(t, err)
	assert.NotNil(t, rc.Config.Defaults)
	assert.NotNil(t, rc.Config.Defaults.Mode)

	// Test case 2: Defaults with strategy (as sub-command)
	dir = &Directive{BaseCmd: "default", Command: "default:strategy", Argument: "my-strategy", SubCmds: []string{"strategy"}}
	err = rc.ParseDirective(dir)
	assert.NoError(t, err)
	assert.Equal(t, "my-strategy", rc.Config.Defaults.Mode.Strategy)

	// Test case 3: Defaults with prefix (as sub-command)
	dir = &Directive{BaseCmd: "default", Command: "default:prefix", Argument: "my-prefix", SubCmds: []string{"prefix"}}
	err = rc.ParseDirective(dir)
	assert.NoError(t, err)
	assert.Equal(t, "my-prefix", rc.Config.Defaults.Mode.Prefix)

	// Test case 4: Defaults with unknown sub-command
	dir = &Directive{BaseCmd: "default", Command: "default:unknown", Argument: "value", SubCmds: []string{"unknown"}}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unrecognized directive 'unknown' for Defaults")

	// Test case 5: Defaults with JSON argument
	jsonArg := `{"Mode":{"Strategy":"json-strategy","Prefix":"json-prefix"}}`
	dir = &Directive{BaseCmd: "default", Command: "default", Argument: jsonArg, IsJSON: true}
	err = rc.ParseDirective(dir)
	assert.NoError(t, err)
	assert.Equal(t, "json-strategy", rc.Config.Defaults.Mode.Strategy)
	assert.Equal(t, "json-prefix", rc.Config.Defaults.Mode.Prefix)

	// Test case 6: Defaults with direct argument (should error)
	dir = &Directive{BaseCmd: "default", Command: "default", Argument: "some-value"}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default directive does not accept a direct argument unless it's a JSON block or has sub-commands")
}

func TestRootConfigParseDirectiveProps(t *testing.T) {
	// Setup
	rc := &RootConfig{Config: config.New()}

	// Test case 1: Basic prop
	dir := &Directive{BaseCmd: "property", Command: "property", Argument: "key1=value1"}
	err := rc.ParseDirective(dir)
	assert.NoError(t, err)
	assert.Len(t, rc.Config.Props, 1)
	assert.Equal(t, "key1", rc.Config.Props[0].Name)
	assert.Equal(t, "value1", rc.Config.Props[0].Value)

	// Test case 2: Another prop
	dir = &Directive{BaseCmd: "property", Command: "property", Argument: "key2=value2"}
	err = rc.ParseDirective(dir)
	assert.NoError(t, err)
	assert.Len(t, rc.Config.Props, 2)
	assert.Equal(t, "key2", rc.Config.Props[1].Name)
	assert.Equal(t, "value2", rc.Config.Props[1].Value)

	// Test case 3: Invalid prop format
	dir = &Directive{BaseCmd: "property", Command: "property", Argument: "invalid_prop"}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid props directive argument 'invalid_prop', expected key=value")

	// Test case 4: Missing argument
	dir = &Directive{BaseCmd: "property", Command: "property", Argument: ""}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "props directive requires an argument (key=value)")
}

func TestRootConfigParseDirectiveIgnores(t *testing.T) {
	// Setup
	rc := &RootConfig{Config: config.New()}

	// Test case 1: Basic ignore
	dir := &Directive{Command: "ignores", Argument: "*.log"}
	err := rc.ParseDirective(dir)
	assert.NoError(t, err)
	assert.Len(t, rc.Config.Ignores, 1)
	assert.Equal(t, "*.log", rc.Config.Ignores[0])

	// Test case 2: Another ignore
	dir = &Directive{Command: "ignores", Argument: "temp/"}
	err = rc.ParseDirective(dir)
	assert.NoError(t, err)
	assert.Len(t, rc.Config.Ignores, 2)
	assert.Equal(t, "temp/", rc.Config.Ignores[1])

	// Test case 3: Missing argument
	dir = &Directive{Command: "ignores", Argument: ""}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ignores directive requires an argument (pattern)")
}

func TestRootConfigParseDirectiveScopeErrors(t *testing.T) {
	// Setup
	rc := &RootConfig{Config: config.New()}

	// Test case 1: packages directive (should not be handled by ParseDirective)
	dir := &Directive{Command: "packages", Argument: ""}
	err := rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directive 'packages' starts a new scope and should not be parsed by RootConfig.ParseDirective")

	// Test case 2: types directive
	dir = &Directive{Command: "types", Argument: ""}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directive 'types' starts a new scope and should not be parsed by RootConfig.ParseDirective")

	// Add more cases for functions, variables, constants
	dir = &Directive{Command: "functions", Argument: ""}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directive 'functions' starts a new scope and should not be parsed by RootConfig.ParseDirective")

	dir = &Directive{Command: "variables", Argument: ""}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directive 'variables' starts a new scope and should not be parsed by RootConfig.ParseDirective")

	dir = &Directive{Command: "constants", Argument: ""}
	err = rc.ParseDirective(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directive 'constants' starts a new scope and should not be parsed by RootConfig.ParseDirective")
}