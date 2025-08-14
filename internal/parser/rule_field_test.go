package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

// TestFieldRule_AddRuleErrors tests that FieldRule's Add*Rule methods return errors.
func TestFieldRule_AddRuleErrors(t *testing.T) {
	fieldRule := &FieldRule{MemberRule: &config.MemberRule{Name: "MyField"}}

	tests := []struct {
		name        string
		addFunc     func() error
		expectedErr string
	}{
		{
			name:        "AddPackage should return error",
			addFunc:     func() error { return fieldRule.AddPackage(&PackageRule{}) },
			expectedErr: "FieldRule cannot contain a PackageRule",
		},
		{
			name:        "AddTypeRule should return error",
			addFunc:     func() error { return fieldRule.AddTypeRule(&TypeRule{}) },
			expectedErr: "FieldRule cannot contain a TypeRule",
		},
		{
			name:        "AddFuncRule should return error",
			addFunc:     func() error { return fieldRule.AddFuncRule(&FuncRule{}) },
			expectedErr: "FieldRule cannot contain a FuncRule",
		},
		{
			name:        "AddVarRule should return error",
			addFunc:     func() error { return fieldRule.AddVarRule(&VarRule{}) },
			expectedErr: "FieldRule cannot contain a VarRule",
		},
		{
			name:        "AddConstRule should return error",
			addFunc:     func() error { return fieldRule.AddConstRule(&ConstRule{}) },
			expectedErr: "FieldRule cannot contain a ConstRule",
		},
		{
			name:        "AddMethodRule should return error",
			addFunc:     func() error { return fieldRule.AddMethodRule(&MethodRule{}) },
			expectedErr: "FieldRule cannot contain a MethodRule",
		},
		{
			name:        "AddFieldRule should return error",
			addFunc:     func() error { return fieldRule.AddFieldRule(&FieldRule{}) },
			expectedErr: "FieldRule cannot contain a FieldRule",
		},
		{
			name:        "AddRule (any) should return error",
			addFunc:     func() error { return fieldRule.AddRule(struct{}{}) },
			expectedErr: "FieldRule cannot contain any child rules",
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

// TestFieldRule_Finalize tests the Finalize method of FieldRule.
func TestFieldRule_Finalize(t *testing.T) {
	fieldRule := &FieldRule{MemberRule: &config.MemberRule{Name: "MyField"}}

	t.Run("Finalize with nil parent should return error", func(t *testing.T) {
		err := fieldRule.Finalize(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FieldRule cannot finalize without a parent container")
	})

	t.Run("Finalize with valid parent should call AddFieldRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		mockParent.On("AddFieldRule", fieldRule).Return(nil).Once()

		err := fieldRule.Finalize(mockParent)
		assert.NoError(t, err)
		mockParent.AssertExpectations(t)
	})

	t.Run("Finalize with parent that errors on AddFieldRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		expectedErr := fmt.Errorf("parent cannot add field rule")
		mockParent.On("AddFieldRule", fieldRule).Return(expectedErr).Once()

		err := fieldRule.Finalize(mockParent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		mockParent.AssertExpectations(t)
	})
}
