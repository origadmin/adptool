package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

// TestVarRule_AddRuleErrors tests that VarRule's Add*Rule methods return errors.
func TestVarRule_AddRuleErrors(t *testing.T) {
	varRule := &VarRule{VarRule: &config.VarRule{Name: "MyVar"}}

	tests := []struct {
		name        string
		addFunc     func() error
		expectedErr string
	}{
		{
			name:        "AddPackage should return error",
			addFunc:     func() error { return varRule.AddPackage(&PackageRule{}) },
			expectedErr: "VarRule cannot contain a PackageRule",
		},
		{
			name:        "AddTypeRule should return error",
			addFunc:     func() error { return varRule.AddTypeRule(&TypeRule{}) },
			expectedErr: "VarRule cannot contain a TypeRule",
		},
		{
			name:        "AddFuncRule should return error",
			addFunc:     func() error { return varRule.AddFuncRule(&FuncRule{}) },
			expectedErr: "VarRule cannot contain a FuncRule",
		},
		{
			name:        "AddVarRule should return error",
			addFunc:     func() error { return varRule.AddVarRule(&VarRule{}) },
			expectedErr: "VarRule cannot contain a VarRule",
		},
		{
			name:        "AddConstRule should return error",
			addFunc:     func() error { return varRule.AddConstRule(&ConstRule{}) },
			expectedErr: "VarRule cannot contain a ConstRule",
		},
		{
			name:        "AddMethodRule should return error",
			addFunc:     func() error { return varRule.AddMethodRule(&MethodRule{}) },
			expectedErr: "VarRule cannot contain a MethodRule",
		},
		{
			name:        "AddFieldRule should return error",
			addFunc:     func() error { return varRule.AddFieldRule(&FieldRule{}) },
			expectedErr: "VarRule cannot contain a FieldRule",
		},
		{
			name:        "AddRule (any) should return error",
			addFunc:     func() error { return varRule.AddRule(struct{}{}) },
			expectedErr: "VarRule cannot contain any child rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.addFunc()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// TestVarRule_Finalize tests the Finalize method of VarRule.
func TestVarRule_Finalize(t *testing.T) {
	varRule := &VarRule{VarRule: &config.VarRule{Name: "MyVar"}}

	t.Run("Finalize with nil parent should return error", func(t *testing.T) {
		err := varRule.Finalize(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "VarRule cannot finalize without a parent container")
	})

	t.Run("Finalize with valid parent should call AddVarRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		mockParent.On("AddVarRule", varRule).Return(nil).Once()

		err := varRule.Finalize(mockParent)
		assert.NoError(t, err)
		mockParent.AssertExpectations(t)
	})

	t.Run("Finalize with parent that errors on AddVarRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		expectedErr := fmt.Errorf("parent cannot add var rule")
		mockParent.On("AddVarRule", varRule).Return(expectedErr).Once()

		err := varRule.Finalize(mockParent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		mockParent.AssertExpectations(t)
	})
}

// TestVarRule_ParseDirective tests the ParseDirective method of VarRule.
func TestVarRule_ParseDirective(t *testing.T) {
	tests := []struct {
		name          string
		directives    []string // Sequence of directive strings to parse
		expectedRuleSet *config.RuleSet // The expected final state of VarRule.RuleSet
		expectError   bool
		errorContains string
	}{
		{
			name: "single strategy directive",
			directives: []string{
				"//go:adapter:strategy my-strategy",
			},
			expectedRuleSet: &config.RuleSet{
				Strategy: []string{"my-strategy"},
			},
			expectError: false,
		},
		{
			name: "accumulate multiple strategies",
			directives: []string{
				"//go:adapter:strategy s1",
				"//go:adapter:strategy s2",
			},
			expectedRuleSet: &config.RuleSet{
				Strategy: []string{"s1", "s2"},
			},
			expectError: false,
		},
		{
			name: "single prefix directive",
			directives: []string{
				"//go:adapter:prefix my-prefix",
			},
			expectedRuleSet: &config.RuleSet{
				Prefix: "my-prefix",
			},
			expectError: false,
		},
		{
			name: "override prefix directive",
			directives: []string{
				"//go:adapter:prefix p1",
				"//go:adapter:prefix p2",
			},
			expectedRuleSet: &config.RuleSet{
				Prefix: "p2",
			},
			expectError: false,
		},
		{
			name: "single ignores directive",
			directives: []string{
				"//go:adapter:ignores *.log",
			},
			expectedRuleSet: &config.RuleSet{
				Ignores: []string{"*.log"},
			},
			expectError: false,
		},
		{
			name: "accumulate multiple ignores",
			directives: []string{
				"//go:adapter:ignores *.log",
				"//go:adapter:ignores temp/",
			},
			expectedRuleSet: &config.RuleSet{
				Ignores: []string{"*.log", "temp/"},
			},
			expectError: false,
		},
		{
			name: "invalid directive",
			directives: []string{
				"//go:adapter:unknown-directive value",
			},
			expectedRuleSet: nil, // RuleSet won't be modified or will be default
			expectError: true,
			errorContains: "unrecognized directive 'unknown-directive'",
		},
		{
			name: "error in sequence should stop processing",
			directives: []string{
				"//go:adapter:prefix valid-prefix",
				"//go:adapter:strategy", // Invalid: strategy requires argument
				"//go:adapter:suffix should-not-be-processed",
			},
			expectedRuleSet: &config.RuleSet{ // State before error
				Prefix: "valid-prefix",
			},
			expectError: true,
			errorContains: "strategy directive requires an argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			varRule := &VarRule{VarRule: &config.VarRule{}} // Fresh rule for each test
			var actualErr error

			for i, dirString := range tt.directives {
				dir := decodeTestDirective(dirString) // Assuming decodeTestDirective is available
				err := varRule.ParseDirective(&dir)
				if err != nil {
					actualErr = err
					t.Logf("Error encountered at directive %d (%s): %v", i, dirString, err)
					break // Stop processing directives on first error
				}
			}

			if tt.expectError {
				assert.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errorContains)
			} else {
				assert.NoError(t, actualErr)
				assert.NotNil(t, varRule.RuleSet)
				assert.Equal(t, tt.expectedRuleSet.Strategy, varRule.RuleSet.Strategy)
				assert.Equal(t, tt.expectedRuleSet.Prefix, varRule.RuleSet.Prefix)
				assert.Equal(t, tt.expectedRuleSet.Suffix, varRule.RuleSet.Suffix)
				assert.Equal(t, tt.expectedRuleSet.Explicit, varRule.RuleSet.Explicit)
				assert.Equal(t, tt.expectedRuleSet.Regex, varRule.RuleSet.Regex)
				assert.Equal(t, tt.expectedRuleSet.Ignores, varRule.RuleSet.Ignores)
				assert.Equal(t, tt.expectedRuleSet.Transforms, varRule.RuleSet.Transforms)
				// Add assertions for other RuleSet fields as needed
			}
		})
	}
}