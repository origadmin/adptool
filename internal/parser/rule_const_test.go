package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

// TestConstRule_AddRuleErrors tests that ConstRule's Add*Rule methods return errors.
func TestConstRule_AddRuleErrors(t *testing.T) {
	constRule := &ConstRule{ConstRule: &config.ConstRule{Name: "MyConst"}}

	tests := []struct {
		name        string
		addFunc     func() error
		expectedErr string
	}{
		{
			name:        "AddPackage should return error",
			addFunc:     func() error { return constRule.AddPackage(&PackageRule{}) },
			expectedErr: "ConstRule cannot contain a PackageRule",
		},
		{
			name:        "AddTypeRule should return error",
			addFunc:     func() error { return constRule.AddTypeRule(&TypeRule{}) },
			expectedErr: "ConstRule cannot contain a TypeRule",
		},
		{
			name:        "AddFuncRule should return error",
			addFunc:     func() error { return constRule.AddFuncRule(&FuncRule{}) },
			expectedErr: "ConstRule cannot contain a FuncRule",
		},
		{
			name:        "AddVarRule should return error",
			addFunc:     func() error { return constRule.AddVarRule(&VarRule{}) },
			expectedErr: "ConstRule cannot contain a VarRule",
		},
		{
			name:        "AddConstRule should return error",
			addFunc:     func() error { return constRule.AddConstRule(&ConstRule{}) },
			expectedErr: "ConstRule cannot contain a ConstRule",
		},
		{
			name:        "AddMethodRule should return error",
			addFunc:     func() error { return constRule.AddMethodRule(&MethodRule{}) },
			expectedErr: "ConstRule cannot contain a MethodRule",
		},
		{
			name:        "AddFieldRule should return error",
			addFunc:     func() error { return constRule.AddFieldRule(&FieldRule{}) },
			expectedErr: "ConstRule cannot contain a FieldRule",
		},
		{
			name:        "AddRule (any) should return error",
			addFunc:     func() error { return constRule.AddRule(struct{}{}) },
			expectedErr: "ConstRule cannot contain any child rules",
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

// TestConstRule_Finalize tests the Finalize method of ConstRule.
func TestConstRule_Finalize(t *testing.T) {
	constRule := &ConstRule{ConstRule: &config.ConstRule{Name: "MyConst"}}

	t.Run("Finalize with nil parent should return error", func(t *testing.T) {
		err := constRule.Finalize(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ConstRule cannot finalize without a parent container")
	})

	t.Run("Finalize with valid parent should call AddConstRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		mockParent.On("AddConstRule", constRule).Return(nil).Once()

		err := constRule.Finalize(mockParent)
		assert.NoError(t, err)
		mockParent.AssertExpectations(t)
	})

	t.Run("Finalize with parent that errors on AddConstRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		expectedErr := fmt.Errorf("parent cannot add const rule")
		mockParent.On("AddConstRule", constRule).Return(expectedErr).Once()

		err := constRule.Finalize(mockParent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		mockParent.AssertExpectations(t)
	})
}

// TestConstRule_ParseDirective tests the ParseDirective method of ConstRule.
func TestConstRule_ParseDirective(t *testing.T) {
	tests := []struct {
		name            string
		directives      []string        // Sequence of directive strings to parse
		expectedRuleSet *config.RuleSet // The expected final state of ConstRule.RuleSet
		expectError     bool
		errorContains   string
	}{
		{
			name: "single strategy directive",
			directives: []string{
				"//go:adapter:const:strategy my-strategy",
			},
			expectedRuleSet: &config.RuleSet{
				Strategy: []string{"my-strategy"},
			},
			expectError: false,
		},
		{
			name: "accumulate multiple strategies",
			directives: []string{
				"//go:adapter:const:strategy s1",
				"//go:adapter:const:strategy s2",
			},
			expectedRuleSet: &config.RuleSet{
				Strategy: []string{"s1", "s2"},
			},
			expectError: false,
		},
		{
			name: "single prefix directive",
			directives: []string{
				"//go:adapter:const:prefix my-prefix",
			},
			expectedRuleSet: &config.RuleSet{
				Prefix: "my-prefix",
			},
			expectError: false,
		},
		{
			name: "override prefix directive",
			directives: []string{
				"//go:adapter:const:prefix p1",
				"//go:adapter:const:prefix p2",
			},
			expectedRuleSet: &config.RuleSet{
				Prefix: "p2",
			},
			expectError: false,
		},
		{
			name: "single ignores directive",
			directives: []string{
				"//go:adapter:const:ignores *.log",
			},
			expectedRuleSet: &config.RuleSet{
				Ignores: []string{"*.log"},
			},
			expectError: false,
		},
		{
			name: "accumulate multiple ignores",
			directives: []string{
				"//go:adapter:const:ignores *.log",
				"//go:adapter:const:ignores temp/",
			},
			expectedRuleSet: &config.RuleSet{
				Ignores: []string{"*.log", "temp/"},
			},
			expectError: false,
		},
		{
			name: "invalid directive",
			directives: []string{
				"//go:adapter:const:unknown-directive value",
			},
			expectedRuleSet: nil, // RuleSet won't be modified or will be default
			expectError:     true,
			errorContains:   "unrecognized directive 'const:unknown-directive' for RuleSet",
		},
		{
			name: "error in sequence should stop processing",
			directives: []string{
				"//go:adapter:const:prefix valid-prefix",
				"//go:adapter:const:strategy", // Invalid: strategy requires argument
				"//go:adapter:const:suffix should-not-be-processed",
			},
			expectedRuleSet: &config.RuleSet{ // State before error
				Prefix: "valid-prefix",
			},
			expectError:   true,
			errorContains: "strategy directive requires an argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constRule := &ConstRule{ConstRule: &config.ConstRule{}} // Fresh rule for each test
			var actualErr error

			for i, dirString := range tt.directives {
				dir := decodeTestDirective(dirString) // Assuming decodeTestDirective is available
				if dir.BaseCmd != "const" {
					actualErr = fmt.Errorf("unexpected base command: %s", dir.BaseCmd)
					t.Logf("Error encountered at directive %d (%s): %v", i, dirString, actualErr)
					break // Stop processing directives on first error
				}
				if !dir.HasSub() {
					continue
				}
				err := constRule.ParseDirective(dir.Sub())
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
				assert.NotNil(t, constRule.RuleSet)
				assert.Equal(t, tt.expectedRuleSet.Strategy, constRule.RuleSet.Strategy)
				assert.Equal(t, tt.expectedRuleSet.Prefix, constRule.RuleSet.Prefix)
				assert.Equal(t, tt.expectedRuleSet.Suffix, constRule.RuleSet.Suffix)
				assert.Equal(t, tt.expectedRuleSet.Explicit, constRule.RuleSet.Explicit)
				assert.Equal(t, tt.expectedRuleSet.Regex, constRule.RuleSet.Regex)
				assert.Equal(t, tt.expectedRuleSet.Ignores, constRule.RuleSet.Ignores)
				assert.Equal(t, tt.expectedRuleSet.Transforms, constRule.RuleSet.Transforms)
				// Add assertions for other RuleSet fields as needed
			}
		})
	}
}
