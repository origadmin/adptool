package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

func decodeTestDirective(directiveString string) Directive {
	if !strings.HasPrefix(directiveString, directivePrefix) {
		return Directive{}
	}

	rawDirective := strings.TrimPrefix(directiveString, directivePrefix)
	commentStart := strings.Index(rawDirective, "//")
	if commentStart != -1 {
		rawDirective = strings.TrimSpace(rawDirective[:commentStart])
	}

	return extractDirective(rawDirective, 0)
}

func TestRootConfigParseDirectiveDefaults(t *testing.T) {
	// Test cases for defaults directive
	tests := []struct {
		name            string
		directiveString string
		expectedMode    *config.Mode
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Basic defaults directive (no argument, no sub-command)",
			directiveString: "//go:adapter:default",
			expectedMode:    &config.Mode{},
			expectError:     true,
			errorContains:   "default directive requires an argument (key value)",
		},
		{
			name:            "Defaults with strategy (as sub-command)",
			directiveString: "//go:adapter:default:mode:strategy my-strategy",
			expectedMode:    &config.Mode{Strategy: "my-strategy"},
			expectError:     false,
		},
		{
			name:            "Defaults with prefix (as sub-command)",
			directiveString: "//go:adapter:default:mode:prefix my-prefix",
			expectedMode:    &config.Mode{Prefix: "my-prefix"},
			expectError:     false,
		},
		{
			name:            "Defaults with unknown sub-command",
			directiveString: "//go:adapter:default:unknown value",
			expectedMode:    nil,
			expectError:     true,
			errorContains:   "unrecognized directive 'unknown' for Defaults",
		},
		{
			name:            "Defaults with unknown mode sub-command",
			directiveString: "//go:adapter:default:mode:unknown value",
			expectedMode:    nil,
			expectError:     true,
			errorContains:   "unrecognized directive 'unknown' for mode",
		},
		{
			name:            "Defaults with JSON argument",
			directiveString: "//go:adapter:default:json {\"Mode\":{\"Strategy\":\"json-strategy\",\"Prefix\":\"json-prefix\"}}",
			expectedMode:    &config.Mode{Strategy: "json-strategy", Prefix: "json-prefix"},
			expectError:     false,
		},
		{
			name:            "Defaults with direct argument (should error)",
			directiveString: "//go:adapter:default some-value",
			expectedMode:    nil,
			expectError:     true,
			errorContains:   "default directive does not accept a direct argument unless it's a JSON block or has sub-commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := &RootConfig{Config: config.New()}
			dir := decodeTestDirective(tt.directiveString) // Use decodeTestDirective directly
			t.Logf("Decoded directive: %+v", dir)
			err := rc.ParseDirective(&dir) // Pass address of dir

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rc.Config.Defaults)
				assert.Equal(t, tt.expectedMode, rc.Config.Defaults.Mode)
			}
		})
	}
}

func TestRootConfigParseDirectiveMultiScope(t *testing.T) {
	tests := []struct {
		name             string
		directiveString  []string
		expectedDefaults *config.Defaults
		expectError      bool
		errorContains    string
	}{
		{
			name: "packages directive (should not be handled by ParseDirective)",
			directiveString: []string{
				"//go:adapter:default:mode:strategy my-strategy",
				"//go:adapter:default:mode:prefix my-prefix",
				"//go:adapter:default:json {\"Mode\":{\"Strategy\":\"json-strategy\",\"Prefix\":\"json-prefix\"}}",
			},
			expectedDefaults: &config.Defaults{
				Mode: &config.Mode{
					Strategy: "my-strategy",
					Prefix:   "my-prefix",
				},
			},
			expectError:   false,
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := &RootConfig{Config: config.New()}
			for _, directiveString := range tt.directiveString {
				dir := decodeTestDirective(directiveString)
				err := rc.ParseDirective(&dir)
				if tt.expectError {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorContains)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, rc.Config)
				}
			}
		})
	}
}

func TestRootConfigParseDirectiveProps(t *testing.T) {
	// Test cases for props directive
	tests := []struct {
		name            string
		directiveString string
		expectedProps   []*config.PropsEntry
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Basic prop",
			directiveString: "//go:adapter:property key1 value1",
			expectedProps:   []*config.PropsEntry{{Name: "key1", Value: "value1"}},
			expectError:     false,
		},
		{
			name:            "Another prop",
			directiveString: "//go:adapter:property key2 value2",
			expectedProps:   []*config.PropsEntry{{Name: "key2", Value: "value2"}},
			expectError:     false,
		},
		{
			name:            "Invalid prop format",
			directiveString: "//go:adapter:property invalid_prop",
			expectedProps:   nil,
			expectError:     true,
			errorContains:   "invalid prop directive argument: argument must be in 'name value' format",
		},
		{
			name:            "Missing argument",
			directiveString: "//go:adapter:property",
			expectedProps:   nil,
			expectError:     true,
			errorContains:   "props directive requires an argument (key value)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := &RootConfig{Config: config.New()}
			dir := decodeTestDirective(tt.directiveString)
			err := rc.ParseDirective(&dir)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rc.Config.Props)
				if assert.Len(t, rc.Config.Props, len(tt.expectedProps)) {
					assert.Equal(t, tt.expectedProps[0].Name, rc.Config.Props[0].Name)
					assert.Equal(t, tt.expectedProps[0].Value, rc.Config.Props[0].Value)
				}
			}
		})
	}
}

func TestRootConfigParseDirectiveIgnores(t *testing.T) {
	// Test cases for ignores directive
	tests := []struct {
		name            string
		directiveString string
		expectedIgnores []string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Basic ignore",
			directiveString: "//go:adapter:ignores *.log",
			expectedIgnores: []string{"*.log"},
			expectError:     false,
		},
		{
			name:            "Another ignore",
			directiveString: "//go:adapter:ignores temp/",
			expectedIgnores: []string{"temp/"},
			expectError:     false,
		},
		{
			name:            "Missing argument",
			directiveString: "//go:adapter:ignores",
			expectedIgnores: nil,
			expectError:     true,
			errorContains:   "ignores directive requires an argument (pattern)",
		},
		{
			name:            "Multiple ignores",
			directiveString: "//go:adapter:ignores *.log temp/ *.tmp",
			expectedIgnores: []string{"*.log", "temp/", "*.tmp"},
			expectError:     false,
		},
		{
			name:            "Ignores with JSON argument",
			directiveString: "//go:adapter:ignores:json [\"pattern1\", \"pattern2\"]",
			expectedIgnores: []string{"pattern1", "pattern2"},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := &RootConfig{Config: config.New()}
			dir := decodeTestDirective(tt.directiveString)
			err := rc.ParseDirective(&dir)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rc.Config.Ignores)
				if assert.Len(t, rc.Config.Ignores, len(tt.expectedIgnores)) {
					assert.Equal(t, tt.expectedIgnores[0], rc.Config.Ignores[0])
				}
			}
		})
	}
}

func TestRootConfigParseDirectiveScopeErrors(t *testing.T) {
	// Test cases for scope errors
	tests := []struct {
		name            string
		directiveString string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "packages directive (should not be handled by ParseDirective)",
			directiveString: "//go:adapter:packages",
			expectError:     true,
			errorContains:   "directive 'packages' starts a new scope and should not be parsed by RootConfig.ParseDirective",
		},
		{
			name:            "types directive",
			directiveString: "//go:adapter:types",
			expectError:     true,
			errorContains:   "directive 'types' starts a new scope and should not be parsed by RootConfig.ParseDirective",
		},
		{
			name:            "functions directive",
			directiveString: "//go:adapter:functions",
			expectError:     true,
			errorContains:   "directive 'functions' starts a new scope and should not be parsed by RootConfig.ParseDirective",
		},
		{
			name:            "variables directive",
			directiveString: "//go:adapter:variables",
			expectError:     true,
			errorContains:   "directive 'variables' starts a new scope and should not be parsed by RootConfig.ParseDirective",
		},
		{
			name:            "constants directive",
			directiveString: "//go:adapter:constants",
			expectError:     true,
			errorContains:   "directive 'constants' starts a new scope and should not be parsed by RootConfig.ParseDirective",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := &RootConfig{Config: config.New()}
			dir := decodeTestDirective(tt.directiveString)
			err := rc.ParseDirective(&dir)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rc.Config)
			}
		})
	}
}
