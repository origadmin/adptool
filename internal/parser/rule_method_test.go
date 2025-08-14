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
