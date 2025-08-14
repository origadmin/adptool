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

// TestPackageRule_ParseDirective tests the ParseDirective method of PackageRule.
func TestPackageRule_ParseDirective(t *testing.T) {
	tests := []struct {
		name            string
		directives      []string        // Sequence of directive strings to parse
		expectedPackage *config.Package // The expected final state of PackageRule.Package
		expectError     bool
		errorContains   string
	}{
		{
			name: "single import directive",
			directives: []string{
				"//go:adapter:package:import github.com/my/package",
			},
			expectedPackage: &config.Package{
				Import: "github.com/my/package",
			},
			expectError: false,
		},
		{
			name: "override import directive",
			directives: []string{
				"//go:adapter:package:import old/package",
				"//go:adapter:package:import new/package",
			},
			expectedPackage: &config.Package{
				Import: "new/package",
			},
			expectError: false,
		},
		{
			name: "single path directive",
			directives: []string{
				"//go:adapter:package:path /path/to/package",
			},
			expectedPackage: &config.Package{
				Path: "/path/to/package",
			},
			expectError: false,
		},
		{
			name: "single alias directive",
			directives: []string{
				"//go:adapter:package:alias mypkg",
			},
			expectedPackage: &config.Package{
				Alias: "mypkg",
			},
			expectError: false,
		},
		{
			name: "single props directive",
			directives: []string{
				"//go:adapter:package:props key1=value1",
			},
			expectedPackage: &config.Package{
				Props: []*config.PropsEntry{{Name: "key1", Value: "value1"}},
			},
			expectError: false,
		},
		{
			name: "accumulate multiple props directives",
			directives: []string{
				"//go:adapter:package:props key1=value1",
				"//go:adapter:package:props key2=value2",
			},
			expectedPackage: &config.Package{
				Props: []*config.PropsEntry{
					{Name: "key1", Value: "value1"},
					{Name: "key2", Value: "value2"},
				},
			},
			expectError: false,
		},
		{
			name: "invalid props directive",
			directives: []string{
				"//go:adapter:package:props invalid_format",
			},
			expectedPackage: nil, // Expect partial or nil due to error
			expectError:     true,
			errorContains:   "invalid props directive argument",
		},
		{
			name: "scope-starting directive should error",
			directives: []string{
				"//go:adapter:package:types",
			},
			expectedPackage: nil, // Expect partial or nil due to error
			expectError:     true,
			errorContains:   "directive 'package:types' starts a new scope and should not be parsed by PackageRule.ParseDirective",
		},
		{
			name: "unrecognized directive should error",
			directives: []string{
				"//go:adapter:package:unknown-command value",
			},
			expectedPackage: nil, // Expect partial or nil due to error
			expectError:     true,
			errorContains:   "unrecognized directive 'package:unknown-command' for PackageRule",
		},
		{
			name: "error in sequence should stop processing",
			directives: []string{
				"//go:adapter:package:import good/import",
				"//go:adapter:package:props invalid_prop", // This should cause an error
				"//go:adapter:package:alias should-not-be-processed",
			},
			expectedPackage: &config.Package{ // State before error
				Import: "good/import",
			},
			expectError:   true,
			errorContains: "invalid props directive argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgRule := &PackageRule{Package: &config.Package{}} // Fresh rule for each test
			var actualErr error

			for i, dirString := range tt.directives {
				dir := decodeTestDirective(dirString) // Assuming decodeTestDirective is available
				if dir.BaseCmd != "package" {
					actualErr = fmt.Errorf("unexpected base command: %s", dir.BaseCmd)
					t.Logf("Error encountered at directive %d (%s): %v", i, dirString, actualErr)
					break // Stop processing directives on first error
				}
				if !dir.HasSub() {
					continue
				}
				err := pkgRule.ParseDirective(dir.Sub())
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
				assert.NotNil(t, pkgRule.Package)
				assert.Equal(t, tt.expectedPackage.Import, pkgRule.Package.Import)
				assert.Equal(t, tt.expectedPackage.Path, pkgRule.Package.Path)
				assert.Equal(t, tt.expectedPackage.Alias, pkgRule.Package.Alias)
				assert.ElementsMatch(t, tt.expectedPackage.Props, pkgRule.Package.Props)
				// Note: PackageRule.ParseDirective does not handle Types, Functions, Variables, Constants directly.
				// Those are handled by AddRule methods and the main parser loop.
			}
		})
	}
}
