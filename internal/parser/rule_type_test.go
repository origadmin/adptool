package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

// TestTypeRule_AddRuleErrors tests that TypeRule's Add*Rule methods return errors for unsupported types.
func TestTypeRule_AddRuleErrors(t *testing.T) {
	typeRule := &TypeRule{TypeRule: &config.TypeRule{Name: "MyType"}}

	tests := []struct {
		name        string
		addFunc     func() error
		expectedErr string
	}{
		{
			name:        "AddPackage should return error",
			addFunc:     func() error { return typeRule.AddPackage(&PackageRule{}) },
			expectedErr: "TypeRule cannot contain a PackageRule",
		},
		{
			name:        "AddTypeRule should return error",
			addFunc:     func() error { return typeRule.AddTypeRule(&TypeRule{}) },
			expectedErr: "TypeRule cannot contain a TypeRule",
		},
		{
			name:        "AddFuncRule should return error",
			addFunc:     func() error { return typeRule.AddFuncRule(&FuncRule{}) },
			expectedErr: "TypeRule cannot contain a FuncRule",
		},
		{
			name:        "AddVarRule should return error",
			addFunc:     func() error { return typeRule.AddVarRule(&VarRule{}) },
			expectedErr: "TypeRule cannot contain a VarRule",
		},
		{
			name:        "AddConstRule should return error",
			addFunc:     func() error { return typeRule.AddConstRule(&ConstRule{}) },
			expectedErr: "TypeRule cannot contain a ConstRule",
		},
		{
			name:        "AddRule (unsupported) should return error",
			addFunc:     func() error { return typeRule.AddRule(struct{}{}) },
			expectedErr: "TypeRule cannot contain a rule of type struct {}",
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

// TestTypeRule_AddSupportedRules tests adding supported rules to TypeRule.
func TestTypeRule_AddSupportedRules(t *testing.T) {
	t.Run("Add single MethodRule", func(t *testing.T) {
		typeRule := &TypeRule{TypeRule: &config.TypeRule{Name: "MyType"}}
		methodRule := &MethodRule{MemberRule: &config.MemberRule{Name: "MyMethod"}}
		
		err := typeRule.AddMethodRule(methodRule)
		assert.NoError(t, err)
		assert.Len(t, typeRule.TypeRule.Methods, 1)
		assert.Equal(t, methodRule.MemberRule, typeRule.TypeRule.Methods[0])
	})

	t.Run("Add single FieldRule", func(t *testing.T) {
		typeRule := &TypeRule{TypeRule: &config.TypeRule{Name: "MyType"}}
		fieldRule := &FieldRule{MemberRule: &config.MemberRule{Name: "MyField"}}
		
		err := typeRule.AddFieldRule(fieldRule)
		assert.NoError(t, err)
		assert.Len(t, typeRule.TypeRule.Fields, 1)
		assert.Equal(t, fieldRule.MemberRule, typeRule.TypeRule.Fields[0])
	})

	t.Run("Accumulate multiple Method and Field rules", func(t *testing.T) {
		typeRule := &TypeRule{TypeRule: &config.TypeRule{Name: "MyType"}}

		method1 := &MethodRule{MemberRule: &config.MemberRule{Name: "Method1"}}
		field1 := &FieldRule{MemberRule: &config.MemberRule{Name: "Field1"}}
		method2 := &MethodRule{MemberRule: &config.MemberRule{Name: "Method2"}}
		field2 := &FieldRule{MemberRule: &config.MemberRule{Name: "Field2"}}

		assert.NoError(t, typeRule.AddMethodRule(method1))
		assert.NoError(t, typeRule.AddFieldRule(field1))
		assert.NoError(t, typeRule.AddMethodRule(method2))
		assert.NoError(t, typeRule.AddFieldRule(field2))

		assert.Len(t, typeRule.TypeRule.Methods, 2)
		assert.Equal(t, method1.MemberRule, typeRule.TypeRule.Methods[0])
		assert.Equal(t, method2.MemberRule, typeRule.TypeRule.Methods[1])

		assert.Len(t, typeRule.TypeRule.Fields, 2)
		assert.Equal(t, field1.MemberRule, typeRule.TypeRule.Fields[0])
		assert.Equal(t, field2.MemberRule, typeRule.TypeRule.Fields[1])
	})
}

// TestTypeRule_Finalize tests the Finalize method of TypeRule.
func TestTypeRule_Finalize(t *testing.T) {
	typeRule := &TypeRule{TypeRule: &config.TypeRule{Name: "MyType"}}

	t.Run("Finalize with nil parent should return error", func(t *testing.T) {
		err := typeRule.Finalize(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TypeRule cannot finalize without a parent container")
	})

	t.Run("Finalize with valid parent should call AddTypeRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		mockParent.On("AddTypeRule", typeRule).Return(nil).Once()

		err := typeRule.Finalize(mockParent)
		assert.NoError(t, err)
		mockParent.AssertExpectations(t)
	})

	t.Run("Finalize with parent that errors on AddTypeRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		expectedErr := fmt.Errorf("parent cannot add type rule")
		mockParent.On("AddTypeRule", typeRule).Return(expectedErr).Once()

		err := typeRule.Finalize(mockParent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		mockParent.AssertExpectations(t)
	})
}

// TestTypeRule_ParseDirective tests the ParseDirective method of TypeRule.
func TestTypeRule_ParseDirective(t *testing.T) {
	tests := []struct {
		name          string
		directives    []string // Sequence of directive strings to parse
		expectedRuleSet *config.RuleSet // The expected final state of TypeRule.RuleSet
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
			typeRule := &TypeRule{TypeRule: &config.TypeRule{}} // Fresh rule for each test
			var actualErr error

			for i, dirString := range tt.directives {
				dir := decodeTestDirective(dirString) // Assuming decodeTestDirective is available
				err := typeRule.ParseDirective(&dir)
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
				assert.NotNil(t, typeRule.RuleSet)
				assert.Equal(t, tt.expectedRuleSet.Strategy, typeRule.RuleSet.Strategy)
				assert.Equal(t, tt.expectedRuleSet.Prefix, typeRule.RuleSet.Prefix)
				assert.Equal(t, tt.expectedRuleSet.Suffix, typeRule.RuleSet.Suffix)
				assert.Equal(t, tt.expectedRuleSet.Explicit, typeRule.RuleSet.Explicit)
				assert.Equal(t, tt.expectedRuleSet.Regex, typeRule.RuleSet.Regex)
				assert.Equal(t, tt.expectedRuleSet.Ignores, typeRule.RuleSet.Ignores)
				assert.Equal(t, tt.expectedRuleSet.Transforms, typeRule.RuleSet.Transforms)
				// Add assertions for other RuleSet fields as needed
			}
		})
	}
}