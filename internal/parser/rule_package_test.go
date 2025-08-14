package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
)

// TestPackageRule_AddRuleErrors tests that PackageRule's Add*Rule methods return errors for unsupported types.
func TestPackageRule_AddRuleErrors(t *testing.T) {
	pkgRule := &PackageRule{Package: &config.Package{Import: "my/package"}}

	tests := []struct {
		name        string
		addFunc     func() error
		expectedErr string
	}{
		{
			name:        "AddPackage should return error",
			addFunc:     func() error { return pkgRule.AddPackage(&PackageRule{}) },
			expectedErr: "PackageRule cannot contain another PackageRule",
		},
		{
			name:        "AddMethodRule should return error",
			addFunc:     func() error { return pkgRule.AddMethodRule(&MethodRule{}) },
			expectedErr: "PackageRule cannot contain a MethodRule",
		},
		{
			name:        "AddFieldRule should return error",
			addFunc:     func() error { return pkgRule.AddFieldRule(&FieldRule{}) },
			expectedErr: "PackageRule cannot contain a FieldRule",
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

// TestPackageRule_AddSupportedRules tests adding supported rules to PackageRule.
func TestPackageRule_AddSupportedRules(t *testing.T) {
	t.Run("Add single TypeRule", func(t *testing.T) {
		pkgRule := &PackageRule{Package: &config.Package{Import: "my/package"}}
		typeRule := &TypeRule{TypeRule: &config.TypeRule{Name: "MyType"}}

		err := pkgRule.AddTypeRule(typeRule)
		assert.NoError(t, err)
		assert.Len(t, pkgRule.Package.Types, 1)
		assert.Equal(t, typeRule.TypeRule, pkgRule.Package.Types[0])
	})

	t.Run("Add single FuncRule", func(t *testing.T) {
		pkgRule := &PackageRule{Package: &config.Package{Import: "my/package"}}
		funcRule := &FuncRule{FuncRule: &config.FuncRule{Name: "MyFunc"}}

		err := pkgRule.AddFuncRule(funcRule)
		assert.NoError(t, err)
		assert.Len(t, pkgRule.Package.Functions, 1)
		assert.Equal(t, funcRule.FuncRule, pkgRule.Package.Functions[0])
	})

	t.Run("Add single VarRule", func(t *testing.T) {
		pkgRule := &PackageRule{Package: &config.Package{Import: "my/package"}}
		varRule := &VarRule{VarRule: &config.VarRule{Name: "MyVar"}}

		err := pkgRule.AddVarRule(varRule)
		assert.NoError(t, err)
		assert.Len(t, pkgRule.Package.Variables, 1)
		assert.Equal(t, varRule.VarRule, pkgRule.Package.Variables[0])
	})

	t.Run("Add single ConstRule", func(t *testing.T) {
		pkgRule := &PackageRule{Package: &config.Package{Import: "my/package"}}
		constRule := &ConstRule{ConstRule: &config.ConstRule{Name: "MyConst"}}

		err := pkgRule.AddConstRule(constRule)
		assert.NoError(t, err)
		assert.Len(t, pkgRule.Package.Constants, 1)
		assert.Equal(t, constRule.ConstRule, pkgRule.Package.Constants[0])
	})

	t.Run("Accumulate multiple rules of different types", func(t *testing.T) {
		pkgRule := &PackageRule{Package: &config.Package{Import: "my/package"}}

		type1 := &TypeRule{TypeRule: &config.TypeRule{Name: "Type1"}}
		func1 := &FuncRule{FuncRule: &config.FuncRule{Name: "Func1"}}
		var1 := &VarRule{VarRule: &config.VarRule{Name: "Var1"}}
		const1 := &ConstRule{ConstRule: &config.ConstRule{Name: "Const1"}}

		type2 := &TypeRule{TypeRule: &config.TypeRule{Name: "Type2"}}
		func2 := &FuncRule{FuncRule: &config.FuncRule{Name: "Func2"}}

		assert.NoError(t, pkgRule.AddTypeRule(type1))
		assert.NoError(t, pkgRule.AddFuncRule(func1))
		assert.NoError(t, pkgRule.AddVarRule(var1))
		assert.NoError(t, pkgRule.AddConstRule(const1))
		assert.NoError(t, pkgRule.AddTypeRule(type2))
		assert.NoError(t, pkgRule.AddFuncRule(func2))

		assert.Len(t, pkgRule.Package.Types, 2)
		assert.Equal(t, type1.TypeRule, pkgRule.Package.Types[0])
		assert.Equal(t, type2.TypeRule, pkgRule.Package.Types[1])

		assert.Len(t, pkgRule.Package.Functions, 2)
		assert.Equal(t, func1.FuncRule, pkgRule.Package.Functions[0])
		assert.Equal(t, func2.FuncRule, pkgRule.Package.Functions[1])

		assert.Len(t, pkgRule.Package.Variables, 1)
		assert.Equal(t, var1.VarRule, pkgRule.Package.Variables[0])

		assert.Len(t, pkgRule.Package.Constants, 1)
		assert.Equal(t, const1.ConstRule, pkgRule.Package.Constants[0])
	})
}

// TestPackageRule_Finalize tests the Finalize method of PackageRule.
func TestPackageRule_Finalize(t *testing.T) {
	pkgRule := &PackageRule{Package: &config.Package{Import: "my/package"}}

	t.Run("Finalize with nil parent should return error", func(t *testing.T) {
		err := pkgRule.Finalize(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PackageRule cannot finalize without a parent container")
	})

	t.Run("Finalize with valid parent should call AddPackage", func(t *testing.T) {
		mockParent := new(MockContainer)
		mockParent.On("AddPackage", pkgRule).Return(nil).Once()

		err := pkgRule.Finalize(mockParent)
		assert.NoError(t, err)
		mockParent.AssertExpectations(t)
	})

	t.Run("Finalize with parent that errors on AddPackage", func(t *testing.T) {
		mockParent := new(MockContainer)
		expectedErr := fmt.Errorf("parent cannot add package rule")
		mockParent.On("AddPackage", pkgRule).Return(expectedErr).Once()

		err := pkgRule.Finalize(mockParent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
		mockParent.AssertExpectations(t)
	})
}
