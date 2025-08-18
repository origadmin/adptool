package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

// TestMethodRule_AddRuleErrors tests that MethodRule's Add*Rule methods return errors.
func TestMethodRule_AddRuleErrors(t *testing.T) {
	methodRule := &MethodRule{MemberRule: &config.MemberRule{Name: "MyMethod"}}

	tests := []struct {
		name        string
		addFunc     func() error
		expectedErr string
	}{
		{
			name:        "AddPackage should return error",
			addFunc:     func() error { return methodRule.AddPackage(&PackageRule{}) },
			expectedErr: "MethodRule cannot contain a PackageRule",
		},
		{
			name:        "AddTypeRule should return error",
			addFunc:     func() error { return methodRule.AddTypeRule(&TypeRule{}) },
			expectedErr: "MethodRule cannot contain a TypeRule",
		},
		{
			name:        "AddFuncRule should return error",
			addFunc:     func() error { return methodRule.AddFuncRule(&FuncRule{}) },
			expectedErr: "MethodRule cannot contain a FuncRule",
		},
		{
			name:        "AddVarRule should return error",
			addFunc:     func() error { return methodRule.AddVarRule(&VarRule{}) },
			expectedErr: "MethodRule cannot contain a VarRule",
		},
		{
			name:        "AddConstRule should return error",
			addFunc:     func() error { return methodRule.AddConstRule(&ConstRule{}) },
			expectedErr: "MethodRule cannot contain a ConstRule",
		},
		{
			name:        "AddMethodRule should return error",
			addFunc:     func() error { return methodRule.AddMethodRule(&MethodRule{}) },
			expectedErr: "MethodRule cannot contain a MethodRule",
		},
		{
			name:        "AddFieldRule should return error",
			addFunc:     func() error { return methodRule.AddFieldRule(&FieldRule{}) },
			expectedErr: "MethodRule cannot contain a FieldRule",
		},
		{
			name:        "AddRule (any) should return error",
			addFunc:     func() error { return methodRule.AddRule(struct{}{}) },
			expectedErr: "MethodRule cannot contain any child rules",
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

// TestMethodRule_Finalize tests the Finalize method of MethodRule.
func TestMethodRule_Finalize(t *testing.T) {
	methodRule := &MethodRule{MemberRule: &config.MemberRule{Name: "MyMethod"}}

	t.Run("Finalize with nil parent should return error", func(t *testing.T) {
		err := methodRule.Finalize(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MethodRule cannot finalize without a parent container")
	})

	t.Run("Finalize with valid parent should call AddMethodRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		mockParent.On("AddMethodRule", methodRule).Return(nil).Once()

		err := methodRule.Finalize(mockParent)
		assert.NoError(t, err)
		mockParent.AssertExpectations(t)
	})

	t.Run("Finalize with parent that errors on AddMethodRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		expectedErr := fmt.Errorf("parent cannot add method rule")
		mockParent.On("AddMethodRule", methodRule).Return(expectedErr).Once()

		err := methodRule.Finalize(mockParent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		mockParent.AssertExpectations(t)
	})
}

// TestMethodRule_ParseDirective tests the ParseDirective method of MethodRule.
func TestMethodRule_ParseDirective(t *testing.T) {
	tests := []struct {
		name            string
		directives      []string        // Sequence of directive strings to parse
		expectedRuleSet *config.RuleSet // The expected final state of MethodRule.RuleSet
		expectError     bool
		errorContains   string
	}{
		{
			name: "single strategy directive",
			directives: []string{
				"//go:adapter:method:strategy my-strategy",
			},
			expectedRuleSet: &config.RuleSet{
				Strategy: []string{"my-strategy"},
			},
			expectError: false,
		},
		{
			name: "accumulate multiple strategies",
			directives: []string{
				"//go:adapter:method:strategy s1",
				"//go:adapter:method:strategy s2",
			},
			expectedRuleSet: &config.RuleSet{
				Strategy: []string{"s1", "s2"},
			},
			expectError: false,
		},
		{
			name: "single prefix directive",
			directives: []string{
				"//go:adapter:method:prefix my-prefix",
			},
			expectedRuleSet: &config.RuleSet{
				Prefix: "my-prefix",
			},
			expectError: false,
		},
		{
			name: "override prefix directive",
			directives: []string{
				"//go:adapter:method:prefix p1",
				"//go:adapter:method:prefix p2",
			},
			expectedRuleSet: &config.RuleSet{
				Prefix: "p2",
			},
			expectError: false,
		},
		{
			name: "single ignores directive",
			directives: []string{
				"//go:adapter:method:ignores *.log",
			},
			expectedRuleSet: &config.RuleSet{
				Ignores: []string{"*.log"},
			},
			expectError: false,
		},
		{
			name: "accumulate multiple ignores",
			directives: []string{
				"//go:adapter:method:ignores *.log",
				"//go:adapter:method:ignores temp/",
			},
			expectedRuleSet: &config.RuleSet{
				Ignores: []string{"*.log", "temp/"},
			},
			expectError: false,
		},
		{
			name: "invalid directive",
			directives: []string{
				"//go:adapter:method:unknown-directive value",
			},
			expectedRuleSet: nil, // RuleSet won't be modified or will be default
			expectError:     true,
			errorContains:   "unrecognized directive 'unknown-directive'",
		},
		{
			name: "error in sequence should stop processing",
			directives: []string{
				"//go:adapter:method:prefix valid-prefix",
				"//go:adapter:method:strategy", // Invalid: strategy requires argument
				"//go:adapter:method:suffix should-not-be-processed",
			},
			expectedRuleSet: &config.RuleSet{ // State before error
				Prefix: "valid-prefix",
			},
			expectError:   true,
			errorContains: "strategy directive requires an argument",
		},
		{
			name: "directive with wrong base command should return error",
			directives: []string{
				"//go:adapter:type:strategy some-strategy", // BaseCmd is "type", not "method"
			},
			expectedRuleSet: nil,
			expectError:     true,
			errorContains:   "MethodRule can only contain method directives",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methodRule := &MethodRule{MemberRule: &config.MemberRule{}} // Fresh rule for each test
			initialRuleSet := methodRule.RuleSet                        // Capture initial state
			var actualErr error

			for i, dirString := range tt.directives {
				dir := decodeTestDirective(dirString) // Assuming decodeTestDirective is available
				var err error
				if tt.name == "directive with wrong base command should return error" {
					err = methodRule.ParseDirective(&dir)
				} else {
					if dir.BaseCmd != "method" {
						actualErr = fmt.Errorf("unexpected base command: %s", dir.BaseCmd)
						t.Logf("Error encountered at directive %d (%s): %v", i, dirString, actualErr)
						break // Stop processing directives on first error
					}
					if !dir.HasSub() {
						continue
					}
					err = methodRule.ParseDirective(&dir)
				}

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
				if tt.expectedRuleSet == nil {
					assert.Equal(t, initialRuleSet, methodRule.RuleSet)
				} else {
					assert.NotNil(t, methodRule.RuleSet)
					assert.Equal(t, tt.expectedRuleSet.Strategy, methodRule.RuleSet.Strategy)
					assert.Equal(t, tt.expectedRuleSet.Prefix, methodRule.RuleSet.Prefix)
					assert.Equal(t, tt.expectedRuleSet.Suffix, methodRule.RuleSet.Suffix)
					assert.Equal(t, tt.expectedRuleSet.Explicit, methodRule.RuleSet.Explicit)
					assert.Equal(t, tt.expectedRuleSet.Regex, methodRule.RuleSet.Regex)
					assert.Equal(t, tt.expectedRuleSet.Ignores, methodRule.RuleSet.Ignores)
					assert.Equal(t, tt.expectedRuleSet.Transforms, methodRule.RuleSet.Transforms)
				}
			}
		})
	}
}
