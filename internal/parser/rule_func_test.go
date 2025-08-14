package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

// TestFuncRule_AddRuleErrors tests that FuncRule's Add*Rule methods return errors.
func TestFuncRule_AddRuleErrors(t *testing.T) {
	funcRule := &FuncRule{FuncRule: &config.FuncRule{Name: "MyFunc"}}

	tests := []struct {
		name        string
		addFunc     func() error
		expectedErr string
	}{
		{
			name:        "AddPackage should return error",
			addFunc:     func() error { return funcRule.AddPackage(&PackageRule{}) },
			expectedErr: "FuncRule cannot contain a PackageRule",
		},
		{
			name:        "AddTypeRule should return error",
			addFunc:     func() error { return funcRule.AddTypeRule(&TypeRule{}) },
			expectedErr: "FuncRule cannot contain a TypeRule",
		},
		{
			name:        "AddFuncRule should return error",
			addFunc:     func() error { return funcRule.AddFuncRule(&FuncRule{}) },
			expectedErr: "FuncRule cannot contain a FuncRule",
		},
		{
			name:        "AddVarRule should return error",
			addFunc:     func() error { return funcRule.AddVarRule(&VarRule{}) },
			expectedErr: "FuncRule cannot contain a VarRule",
		},
		{
			name:        "AddConstRule should return error",
			addFunc:     func() error { return funcRule.AddConstRule(&ConstRule{}) },
			expectedErr: "FuncRule cannot contain a ConstRule",
		},
		{
			name:        "AddMethodRule should return error",
			addFunc:     func() error { return funcRule.AddMethodRule(&MethodRule{}) },
			expectedErr: "FuncRule cannot contain a MethodRule",
		},
		{
			name:        "AddFieldRule should return error",
			addFunc:     func() error { return funcRule.AddFieldRule(&FieldRule{}) },
			expectedErr: "FuncRule cannot contain a FieldRule",
		},
		{
			name:        "AddRule (any) should return error",
			addFunc:     func() error { return funcRule.AddRule(struct{}{}) },
			expectedErr: "FuncRule cannot contain any child rules",
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

// TestFuncRule_Finalize tests the Finalize method of FuncRule.
func TestFuncRule_Finalize(t *testing.T) {
	funcRule := &FuncRule{FuncRule: &config.FuncRule{Name: "MyFunc"}}

	t.Run("Finalize with nil parent should return error", func(t *testing.T) {
		err := funcRule.Finalize(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FuncRule cannot finalize without a parent container")
	})

	t.Run("Finalize with valid parent should call AddFuncRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		mockParent.On("AddFuncRule", funcRule).Return(nil).Once()

		err := funcRule.Finalize(mockParent)
		assert.NoError(t, err)
		mockParent.AssertExpectations(t)
	})

	t.Run("Finalize with parent that errors on AddFuncRule", func(t *testing.T) {
		mockParent := new(MockContainer)
		expectedErr := fmt.Errorf("parent cannot add func rule")
		mockParent.On("AddFuncRule", funcRule).Return(expectedErr).Once()

		err := funcRule.Finalize(mockParent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		mockParent.AssertExpectations(t)
	})
}
