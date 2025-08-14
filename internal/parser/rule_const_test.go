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
