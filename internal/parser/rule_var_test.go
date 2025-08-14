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
